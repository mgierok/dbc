# DBC Technical Documentation

## 1. Table of Contents

1. [Technical Overview](#2-technical-overview)
2. [Architecture and Boundaries](#3-architecture-and-boundaries)
3. [Components and Responsibilities](#4-components-and-responsibilities)
4. [Core Technical Mechanisms](#5-core-technical-mechanisms)
5. [Data and Interface Contracts](#6-data-and-interface-contracts)
6. [Runtime and Operational Considerations](#7-runtime-and-operational-considerations)
7. [Technical Decisions and Tradeoffs](#8-technical-decisions-and-tradeoffs)
8. [Technology Stack and Versions](#9-technology-stack-and-versions)
9. [Technical Constraints and Risks](#10-technical-constraints-and-risks)
10. [Deep-Dive References](#11-deep-dive-references)

## 2. Technical Overview

DBC is a terminal application written in Go. It currently supports SQLite and follows Clean Architecture with DDD-style boundaries.

Core technical characteristics:

- Terminal UI adapter built on Bubble Tea.
- Layered architecture with inward dependency flow.
- Engine abstraction (`internal/application/port/engine.go`) to support future database engines.
- SQLite implementation in the infrastructure layer.
- Staged table changes (insert/update/delete) persisted in one transaction.

## 3. Architecture and Boundaries

This project follows architecture rules defined in:

- `docs/clean-architecture-ddd.md`

### 3.1 Dependency Direction

Allowed direction:

- `interfaces` -> `application` -> `domain`
- `infrastructure` -> `application` and `domain`

Not allowed:

- `domain` importing `application`, `interfaces`, or `infrastructure`
- `application` importing `interfaces` or `infrastructure`
- `interfaces` importing `infrastructure`

### 3.2 Key Technical Boundaries

- Database access is behind `application/port.Engine`.
- Use cases depend on port interfaces, not concrete database code.
- TUI does not access SQLite directly; it calls use cases.
- Infrastructure provides concrete adapters (`SQLiteEngine`, config loader).

## 4. Components and Responsibilities

Current source layout:

```text
cmd/
  dbc/
    main.go
internal/
  domain/
    model/
    service/
  application/
    dto/
    port/
    usecase/
  interfaces/
    tui/
  infrastructure/
    config/
    engine/
```

Package responsibilities:

- `cmd/dbc`: composition root and runtime wiring.
- `internal/domain/model`: domain entities/value structures and domain errors.
- `internal/domain/service`: domain-level helper logic (for example value parsing and table sorting).
- `internal/application/usecase`: use case orchestration.
- `internal/application/port`: interfaces that infrastructure implements.
- `internal/application/dto`: data structures exchanged with interface adapters.
- `internal/interfaces/tui`: terminal adapter (input, state, rendering), including shared popup rendering primitives for runtime overlays and centralized input/command registry definitions for runtime and selector contexts.
- `internal/infrastructure/config`: config file loading and validation.
- `internal/infrastructure/engine`: SQLite adapter implementation.

## 5. Core Technical Mechanisms

### 5.1 Startup Flow

1. `cmd/dbc/main.go` parses startup CLI arguments.
   - Startup validates operating system support before dispatch; unsupported systems return startup failure (`exit code 1`).
   - Supported direct-launch aliases: `-d <db_path>` and `--database <db_path>`.
   - Supported informational aliases: `-h` / `--help` and `-v` / `--version`.
   - Version informational rendering resolves `vcs.revision` from Go build metadata and emits a short hash token; when metadata is unavailable it emits `dev`.
   - `runStartupDispatch` short-circuits startup for informational aliases before config-path resolution or DB initialization.
   - Informational and direct-launch aliases are mutually exclusive in one startup invocation.
   - Invalid startup usage and argument-validation failures are classified as usage errors and mapped to exit code `2`.
   - Usage failures emit deterministic stderr guidance in `Error` -> `Hint` -> `Usage` format.
   - Non-usage startup/runtime failures are mapped to exit code `1`.
2. `cmd/dbc/main.go` resolves config path from user home:
   - `~/.config/dbc/config.toml`
3. Startup selector dependencies are created with config-management use cases:
   - list configured databases,
   - create/update/delete configured database entry,
   - resolve active config path.
4. If direct-launch argument is provided, startup attempts direct SQLite open/ping first (before selector UI):
   - startup first resolves direct-launch target against configured entries using normalized SQLite path identity,
   - when normalized match exists, startup reuses that configured entry identity (deterministic first match by config order),
   - success starts runtime immediately and skips selector for initial startup,
   - failure prints actionable failure message and exits non-zero (no selector fallback).
5. For selector-first path (no direct-launch argument), selector UI supports in-session config management (add/edit/delete with delete confirmation) and refreshes entries from config store after each successful mutation.
   Add/edit submit path performs use-case validation in this order: required fields -> SQLite connection check -> config store mutation.
   Active add/edit text input renders a caret (`|`) in the currently focused field.
6. Empty config state (`missing file`, `empty file`, or `databases = []`) starts selector in mandatory first-entry setup:
   - first valid add is required before continue,
   - users can optionally add more entries in the same setup context,
   - normal browsing cannot start until at least one entry exists.
   Pressing `Esc` in this setup cancels startup and exits application loop before DB open.
   Malformed config content (for example malformed TOML, unknown key shape, or invalid entry structure) is treated as startup error.
7. User confirms selected database from refreshed selector list.
8. Selected SQLite database is opened and pinged.
   Startup DB open and config-entry connection checks reuse the same infrastructure connection-open helper (`internal/infrastructure/engine.OpenSQLiteDatabase`) to keep validation behavior and errors consistent.
   If open/ping fails, startup loop reopens selector with status error and preferred selection set to failed connection string.
   Runtime does not start until user selects a reachable entry, updates configuration, or cancels database selection.
9. SQLite engine and runtime table/record use cases are created.
10. Bubble Tea application loop starts (`tui.Run`).
11. If runtime exits with `ErrOpenConfigSelector` (triggered by `:config` or `:c`), DB connection is closed and startup selector flow runs again without restarting process.
    `cmd/dbc/main.go` passes `SelectorLaunchState` (`PreferConnString` + session `AdditionalOptions`) built from in-memory startup context, and `internal/interfaces/tui/selector.go` merges those options after config-backed entries while keeping edit/delete mapped only to config indexes.
    Startup CLI contract details in this flow align with `docs/cli-parameter-and-output-standards.md` (input-error format, help discoverability, and exit-code mapping).

### 5.2 Main Read Flow

1. TUI model initializes by loading tables.
2. Selected table schema is loaded.
3. Records are loaded in pages (`offset`, `limit`) with optional filter and optional single-column sort.
4. Additional records load when selection approaches loaded tail.
5. Runtime panel transitions are handled in `internal/interfaces/tui/model.go`:
   - runtime key dispatch branches match physical keys via centralized binding definitions from `internal/interfaces/tui/input_registry.go` (`keyBindingID` + `keyMatches`),
   - `?` opens a runtime context-help popup that captures the currently active panel/state,
   - `Enter` from left-panel table focus calls `switchToRecords` (records view + right-panel focus).
   - `Esc` from neutral right-panel content focus returns focus to tables and forces `ViewSchema` (Table Discovery) in the right panel.
6. Dirty table-switch routing behavior:
   - when table selection changes while staged edits exist, TUI opens a modal decision popup,
   - popup summary includes current dirty-change count (`dirtyEditCount()`),
   - discard action clears staged state and switches to pending table selection,
   - cancel action keeps current table selection and preserves staged state.
7. Command entry (`:`) is handled inside the TUI model.
   - runtime command parsing delegates to `resolveRuntimeCommand` in `internal/interfaces/tui/input_registry.go`,
   - command aliases are defined once (`:config`/`:c`, `:help`/`:h`, `:q`/`:quit`) and matched case-insensitively after trimming optional `:` prefix.
8. `:config` / `:c` routing behavior:
   - if no staged changes: set selector-return signal and exit runtime loop,
   - if staged changes exist: open modal dirty-state decision popup (`Config`) with `save`, `discard`, `cancel`,
   - `save` executes save flow first and exits to selector only after successful save,
   - `discard` clears staged state and exits to selector immediately,
   - `cancel` keeps runtime session active with staged state unchanged.
9. `:help` / `:h` routing behavior:
   - command remains supported and opens the same runtime context-help popup used by `?`,
   - re-entering `:help` or `:h` while popup is already open keeps popup open (idempotent open),
   - popup title/content are generated from captured context metadata (`helpPopupContextTitle`, `helpPopupContentLines`) so shown keybindings stay aligned with the originating panel/state,
   - help popup maintains internal scroll offset for overflow content and supports keyboard scrolling (`j/k`, `down/up`, `Ctrl+f`/`Ctrl+b`, `g`/`G`, `home`/`end`),
   - popup closes on `Esc`; unrelated keys do not dismiss popup.
10. `:quit` / `:q` routing behavior:
   - command exits runtime loop immediately without save/discard confirmation.
11. Runtime records sort behavior:
   - records sort state is kept in runtime model as optional single-column sort (`column` + `ASC`/`DESC`),
   - sort apply triggers records reload from offset `0`,
   - sort state resets on table switch as part of table-context reset,
   - pending inserts remain rendered before persisted rows and are not SQL-sorted.
12. Runtime single-record detail behavior:
   - `Enter` opens selected-row detail state only when records panel context is active,
   - detail renderer uses effective row values (persisted values overridden by staged edits/inserts),
   - detail layout is vertical and wraps values without truncation,
   - detail supports keyboard scrolling (`j/k`, `Ctrl+f`/`Ctrl+b`, `g`/`G`) and closes with `Esc`,
   - detail state resets on table-context resets and records reload resets.
13. Runtime popup rendering standardization:
   - `internal/interfaces/tui/popup_component.go` provides shared frame rendering (`renderStandardizedPopup`) used by runtime help, filter, sort, edit, and confirm popup variants,
   - shared popup spec includes title/summary rows, optional selectable list rows, optional scroll window/indicator, and width clamping,
   - `View()` renders runtime popup overlays via `centerBoxLines` for help, filter, sort, edit, and confirm popup variants.
14. Runtime status-line composition uses split rendering (`renderStatusWithRightHint`) so the right-aligned `Context help: ?` hint remains visible under narrow widths.
15. Selector browse/form/delete context lines are generated by input-registry helper functions (`selectorContextLinesBrowseDefault`, `selectorContextLinesBrowseFirstSetup`, `selectorFormSwitchLine`, `selectorFormSubmitLine`, `selectorDeleteConfirmationLine`) to keep on-screen hints synchronized with selector key handling.
16. Unsupported runtime commands keep existing fallback:
   - status message shows unknown command text,
   - runtime session remains active.

### 5.3 Write Flow

1. TUI state accumulates insert/edit/delete staging operations.
2. On save confirmation, TUI builds `model.TableChanges`.
3. Use case validates payload (`SaveTableChanges`).
4. Engine applies all changes in one transaction:
   - inserts,
   - updates (skipping rows also marked for delete),
   - deletes.
5. On success:
   - default save action: staged state is cleared and records are reloaded,
   - save from dirty `:config` / `:c` decision: staged state is cleared and runtime exits to selector instead of reloading records.
6. On failure, rollback occurs and staged state remains.

### 5.4 Technical Interaction Patterns

Read interaction pattern:

- TUI requests table/schema/records through application use cases.
- Use cases call `application/port.Engine`.
- Infrastructure engine adapter maps calls to SQLite queries.
- Retrieved records are adapted into DTOs consumed by TUI state.

Write interaction pattern:

- TUI accumulates staged changes in session state.
- Save path builds `model.TableChanges`.
- `SaveTableChanges` validates payload and delegates persistence to the engine port.
- SQLite adapter executes inserts/updates/deletes in one transaction.

### 5.5 Testing Strategy and Coverage

Testing deep dive reference: `docs/test-driven-development.md`.

Current test layers:

- Domain service tests: `internal/domain/service/*_test.go`
- Application use case tests: `internal/application/usecase/*_test.go`
- Infrastructure adapter tests:
  - `internal/infrastructure/config/*_test.go`
  - `internal/infrastructure/engine/*_test.go`
- TUI behavior tests: `internal/interfaces/tui/*_test.go`

Coverage boundaries:

- Domain tests validate domain-level rules and helper behavior.
- Application tests validate use-case orchestration and port contracts.
- Infrastructure tests validate adapter behavior for config and SQLite integration.
- TUI tests validate user-visible state transitions and input handling, including standardized popup rendering and modal/inline popup placement.

Current implementation-level characteristics:

- Tests target behavior contracts instead of private implementation details.
- SQLite integration-like tests use in-memory databases.
- TUI registry tests (`internal/interfaces/tui/input_registry_test.go`) lock deterministic runtime command alias resolution and help-popup content generation.

## 6. Data and Interface Contracts

### 6.1 Runtime Entry Point and Composition Root

- Process entry point is `cmd/dbc/main.go`.
- `main.go` is the composition root for:
  - startup CLI argument parsing (`-d` / `--database`, `-h` / `--help`, `-v` / `--version`),
  - informational startup dispatch before runtime initialization,
  - config path resolution,
  - selector/config-management use case wiring,
  - database engine creation,
  - TUI runtime startup.

### 6.2 Runtime Configuration Contract

- Default config paths:
  - `~/.config/dbc/config.toml`
- Config contract supports zero or more `[[databases]]` entries.
- Each persisted `[[databases]]` entry requires:
  - `name`
  - `db_path`
- Empty config states (`missing file`, `empty file`, `databases = []`) are mapped to mandatory first-entry setup.
- Malformed config content is treated as startup error and blocks runtime initialization.

### 6.3 Application Port Contracts

- Engine contract is defined in `internal/application/port/engine.go`.
- `Engine.ListRecords` accepts optional filter and optional sort payload.
- Database connection validation boundary is defined in `internal/application/port/database_connection_checker.go`.
- Use cases call ports and remain independent from concrete SQLite implementation.
- Infrastructure adapters implement these contracts in:
  - `internal/infrastructure/engine/sqlite_engine.go`
  - `internal/infrastructure/engine/sqlite_sort.go`
  - `internal/infrastructure/engine/sqlite_connection_checker.go`

### 6.4 Selector Session Contracts

- Runtime-to-selector return signal is `ErrOpenConfigSelector`.
- Selector re-entry receives `SelectorLaunchState` with:
  - preferred connection string,
  - session-scoped additional startup options.
- Additional startup options are in-memory only for the active process and are merged after config-backed entries.

### 6.5 Input Registry Contracts

- `internal/interfaces/tui/input_registry.go` defines `keyBindingID` -> (`keys`, `label`) mappings used by both runtime model and selector model key dispatch.
- Runtime command contract is expressed as `runtimeCommandSpec` entries (`aliases`, `description`, `action`) and resolved through `resolveRuntimeCommand`.
- Runtime help/status hint copy uses registry-backed helper functions (including `runtimeStatusContextHelpHint`) instead of duplicated inline literals.

## 7. Runtime and Operational Considerations

### 7.1 Runtime Lifecycle Boundaries

- Startup supports two paths:
  - selector-first startup (no direct-launch argument),
  - direct-launch startup (`-d` / `--database`) with connectivity validation before runtime starts.
- Direct-launch validation failure exits non-zero without selector fallback.
- Runtime can return to selector without process restart via `ErrOpenConfigSelector`.
- Active DB connection is explicitly closed before selector re-entry.

### 7.2 Operational Error Handling Characteristics

- Startup usage/argument-validation failures terminate with exit code `2` and deterministic guidance output.
- Startup runtime/operational failures terminate with exit code `1`.
- Malformed config content blocks startup and surfaces explicit error.
- Invalid DB target chosen in selector keeps selector active with actionable status message.
- Save failures retain staged state to prevent unintended data-loss semantics.

## 8. Technical Decisions and Tradeoffs

### 8.1 SQLite-First with Engine Port

Decision:

- Keep current implementation focused on SQLite while coding to an engine interface.

Why:

- Reduces adapter and operational complexity in the current implementation.
- Preserves an extension seam for additional engines without rewriting use-case orchestration.

Where:

- Port: `internal/application/port/engine.go`
- SQLite adapter: `internal/infrastructure/engine/sqlite_engine.go`

### 8.2 Staged Changes Before Save

Decision:

- Do not write immediately on each edit; stage and save explicitly.

Why:

- Prevents implicit write-side effects during field editing.
- Keeps undo/redo state transitions local to the runtime session.
- Aligns persistence with explicit transaction boundaries.

Where:

- TUI staged state: `internal/interfaces/tui/model.go`
- Save use case: `internal/application/usecase/save_table_changes.go`

### 8.3 Transactional Save per Table

Decision:

- Apply inserts/updates/deletes for selected table inside one transaction.

Why:

- Prevent partial writes.
- Keep table changes consistent after failures.

Where:

- `internal/infrastructure/engine/sqlite_update.go`

### 8.4 Config Contract and Validation

Decision:

- Allow persisting empty database configuration (`databases = []`).
- Require `name` and `db_path` only for entries that exist in config.
- Require successful SQLite connection validation before persisting add/edit changes.

Why:

- Keep empty config behavior equivalent to missing config at startup.
- Allow deleting the last configured database without violating config persistence rules.
- Keep startup errors explicit and actionable.
- Prevent saving unreachable or non-existent database targets in selector configuration.

Where:

- `internal/infrastructure/config/config.go`
- `internal/infrastructure/config/config_test.go`
- `internal/application/port/database_connection_checker.go`
- `internal/application/usecase/config_management.go`
- `internal/infrastructure/engine/sqlite_connection_checker.go`

### 8.5 Operator Allowlist for Filters

Decision:

- Filter operators are limited to a known, validated set.

Why:

- Reduces unsafe SQL clause composition risk.
- Constrains query-builder behavior to a validated operator set.

Where:

- Operators: `internal/infrastructure/engine/sqlite_operator.go`
- Clause builder: `internal/infrastructure/engine/sqlite_filter.go`

### 8.6 Centralized TUI Input and Command Registry

Decision:

- Keep runtime and selector shortcuts, command aliases, and related help/status copy in one shared registry.

Why:

- Eliminates drift between key handlers and rendered shortcut/help text.
- Keeps runtime command parsing deterministic and testable as one contract surface.
- Lowers maintenance overhead when adding or changing shortcuts.

Where:

- Registry definitions: `internal/interfaces/tui/input_registry.go`
- Runtime key/command routing consumers: `internal/interfaces/tui/model.go`
- Selector key/context consumers: `internal/interfaces/tui/selector.go`
- Help/status rendering consumers: `internal/interfaces/tui/view.go`
- Contract tests: `internal/interfaces/tui/input_registry_test.go`

## 9. Technology Stack and Versions

Version source: `go.mod`.

### 9.1 Language and Toolchain

- Go language version: `1.25.0`
- Go toolchain: `go1.25.5`

### 9.2 Direct Dependencies

| Dependency | Version | Purpose |
| --- | --- | --- |
| `github.com/charmbracelet/bubbletea` | `v1.3.10` | Terminal UI framework |
| `modernc.org/sqlite` | `v1.42.2` | SQLite driver |
| `github.com/pelletier/go-toml/v2` | `v2.2.4` | TOML config parsing |

### 9.3 Notes on Indirect Dependencies

- Additional packages in `go.mod` are transitive dependencies and should not be edited manually unless intentionally upgrading dependencies.

### 9.4 Linting Tooling

- Static analysis is configured in `.golangci.yml`.
- The project uses deterministic `golangci-lint` configuration (`linters.default: standard` with explicitly enabled additional linters).
- `revive` comment-enforcement rules are intentionally disabled:
  - `package-comments`
  - `exported`
- `//nolint` usage is restricted:
  - explanation is required,
  - specific linter name is required.

## 10. Technical Constraints and Risks

Current technical constraints:

- Only SQLite engine is implemented.
- Table edit/delete for persisted rows depends on primary key identity.
- Filter pipeline applies one active predicate for the selected table at a time.
- Missing or empty config file is tolerated at startup and mapped to mandatory first-entry setup.
- Invalid config content still stops startup with explicit error.
- Config persistence allows `databases = []`; deleting the last configured entry is persisted as an empty config state.

Operational risks and tradeoffs to verify during changes:

- Direct-launch and selector paths must keep path normalization behavior consistent.
- Save flow must preserve transactional all-or-nothing behavior for one table.
- Selector re-entry flow must close active DB connections before reopening selector.
- Shortcut/command changes must update registry definitions (keys + labels + command specs) to avoid stale UI hints or missing aliases.

## 11. Deep-Dive References

- Architecture and DDD details: `docs/clean-architecture-ddd.md`
- TDD principles: `docs/test-driven-development.md`
