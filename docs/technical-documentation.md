# DBC Technical Documentation

## Technical Overview

- DBC is a terminal SQLite data browser/editor written in Go.
- Runtime behavior is split into read operations through application ports and staged write operations applied only on explicit save.
- The implementation follows Clean Architecture / DDD-style boundaries with inward dependency direction.
- Current implementation target is macOS/Linux runtime with SQLite as the only engine.

## Architecture and Boundaries

### Dependency Invariants

- Allowed dependency direction: `interfaces -> application -> domain` and `infrastructure -> application/domain`.
- Domain layer MUST NOT import `application`, `interfaces`, or `infrastructure`.
- Application layer MUST NOT import `interfaces` or `infrastructure`.
- Interface adapters MUST NOT import infrastructure adapters directly.

### Boundary Contracts

- Database access is behind `internal/application/port.Engine`.
- Use cases orchestrate behavior against ports and stay independent from SQLite-specific details.
- TUI remains an adapter layer and delegates business behavior to use cases.
- Infrastructure packages implement boundary ports (`Engine`, `ConfigStore`, `DatabaseConnectionChecker`).

## Components and Responsibilities

- `cmd/dbc`: process entrypoint, startup argument handling, runtime/selector orchestration.
- `internal/domain/model`: domain value objects, entities, and error contracts.
- `internal/domain/service`: pure domain helpers (table sorting, typed value parsing, input spec inference).
- `internal/application/usecase`: read/write orchestration, config management, staging policy, and dirty-navigation policy.
- `internal/application/port`: application boundary interfaces for infrastructure implementations.
- `internal/application/dto`: adapter-facing data contracts exchanged between use cases and interfaces.
- `internal/interfaces/tui`: public TUI adapter facade plus runtime UI model and runtime-session entrypoints.
- `internal/interfaces/tui/internal/selector`: selector-specific Bubble Tea model, selector view/state transitions, and selector option normalization.
- `internal/interfaces/tui/internal/primitives`: terminal UI primitives shared by runtime and selector, including key/help registry, popup/layout rendering, iconography, and style helpers.
- `internal/infrastructure/config`: TOML config loading/validation/persistence adapter.
- `internal/infrastructure/engine`: SQLite adapter for reads/writes/filter/sort and connectivity checks.

## Core Technical Mechanisms

### Startup Dispatch and Failure Classification

- Guarantee: informational startup flags (`-h`/`--help`, `-v`/`--version`) short-circuit runtime startup.
- Guarantee: usage/argument errors map to exit code `2`; startup/runtime operational failures map to exit code `1`.
- Enforced in: `cmd/dbc/main.go`.

### Selector-First vs Direct-Launch Startup

- Guarantee: selector-first startup is default; `-d`/`--database` enables direct-launch path.
- Guarantee: direct-launch path resolves configured identity first via normalized SQLite path matching.
- Guarantee: direct-launch failure exits non-zero without selector fallback.
- Enforced in: `cmd/dbc/startup_runtime.go`, `cmd/dbc/startup_runtime_selection.go`.

### Staged Write Model

- Guarantee: insert/edit/delete actions are staged in memory and persisted only after explicit save confirmation.
- Guarantee: dirty-change counting and initial insert defaults are delegated to application staging policy.
- Enforced in: `internal/interfaces/tui/model_staging_*.go`, `internal/application/usecase/staging_policy.go`.

### Transactional Save Semantics

- Guarantee: one save applies inserts, updates, and deletes in one transaction for the selected table.
- Guarantee: updates targeting rows also staged for delete are skipped.
- Enforced in: `internal/application/usecase/save_table_changes.go`, `internal/infrastructure/engine/sqlite_update.go`.

### Query-Safety Constraints for Dynamic SQL

- Guarantee: runtime values are bound using placeholders.
- Guarantee: dynamic identifiers are quoted through `quoteIdentifier`.
- Guarantee: filter operators come from an allowlist and sort columns are validated against table schema.
- Enforced in: `internal/infrastructure/engine/sqlite_filter.go`, `internal/infrastructure/engine/sqlite_operator.go`, `internal/infrastructure/engine/sqlite_sort.go`, `internal/infrastructure/engine/sqlite_engine.go`.

### Input Normalization and Typed Parsing

- Guarantee: staged values are parsed by column type and nullability before persistence payload generation.
- Guarantee: boolean and enum-like schema types expose select-style input metadata.
- Enforced in: `internal/domain/service/value_parser.go`, `internal/application/usecase/get_schema.go`, `internal/application/usecase/staged_changes_translator.go`.

### Centralized Runtime and Selector Command Registry

- Guarantee: key bindings, command aliases, and help text are maintained in a shared registry.
- Guarantee: command parsing trims optional `:` and matches aliases case-insensitively.
- Enforced in: `internal/interfaces/tui/internal/primitives/input_registry.go`.

### Terminal-Native TUI Styling

- Guarantee: runtime and selector models resolve their styling profile once at construction time and render deterministically from that profile.
- Guarantee: TUI emphasis uses only ANSI SGR attributes (`bold`, `faint`, `underline`, `reverse`) on the terminal's current foreground/background theme; the application does not define its own color palette.
- Guarantee: setting `NO_COLOR` or running with `TERM=dumb` disables ANSI styling and falls back to plain text rendering.
- Enforced in: `internal/interfaces/tui/internal/primitives/render_style.go`, `internal/interfaces/tui/internal/primitives/view_layout.go`, `internal/interfaces/tui/internal/primitives/popup_component.go`, `internal/interfaces/tui/view*.go`, `internal/interfaces/tui/internal/selector/view.go`.

## Data and Interface Contracts

### Runtime Composition Entry Contract

- Process entrypoint: `cmd/dbc/main.go`.
- Runtime boundary: `internal/interfaces/tui.Run(...)`.
- Selector-return signal: `internal/interfaces/tui.ErrOpenConfigSelector`.

### Configuration Contract

- Active config path: `~/.config/dbc/config.toml`.
- Persisted config entries: `[[databases]]` with required fields `name` and `db_path`.
- Unknown TOML fields are rejected (`DisallowUnknownFields`).
- Missing file and empty `databases` list are valid startup states and route to mandatory first-entry setup.
- Save behavior is atomic (`CreateTemp` + `Rename`) and creates config directory with `0700`.

### Application Port Contracts

- `Engine`: list tables, read schema, read records (with optional filter/sort), list operators, apply table changes.
- `ConfigStore`: list/create/update/delete config entries and expose active config path.
- `DatabaseConnectionChecker`: validate candidate DB path before persisting selector add/edit changes.

### Record Identity and Change Payload Contract

- Persisted-row updates/deletes require non-empty identity keys.
- Identity keys are derived from primary-key columns and carried as typed staged values (`Text`, `IsNull`, optional `Raw`).
- Application/domain write contracts do not depend on SQLite `rowid`.

### Records Page Contract

- Read contract returns `Rows`, `TotalCount`, and `HasMore`.
- `HasMore` is computed via look-ahead (`LIMIT limit+1`).
- Runtime page size is fixed at `20`.

### Selector Launch Contract

- Selector launch state supports preferred connection string reselection and additional session-scoped options.
- Session-scoped options are CLI-origin and are not persisted into config.
- Selector edit/delete operations are allowed only for config-backed entries.

## Runtime and Operational Considerations

- Supported startup operating systems: `darwin`, `linux`.
- Startup usage failures emit deterministic `Error`/`Hint`/`Usage` stderr output.
- SQLite open contract requires existing file path (missing files and directory paths fail fast).
- Startup database open and config add/edit validation share the same open/ping helper.
- Runtime closes active DB handle on session end; close failures are logged.
- Runtime record reload path ignores stale async responses using request ID checks.
- Runtime/selector rendering assumes terminal support for UTF-8 box and marker glyphs plus standard ANSI SGR text attributes; `NO_COLOR` or `TERM=dumb` forces unstyled rendering.

## Technical Decisions and Tradeoffs

### SQLite-First with Engine Port

- Decision: keep one production engine (SQLite) behind application engine port.
- Rationale: low infrastructure complexity now, explicit seam for later engine extensions.
- Where: `internal/application/port/engine.go`, `internal/infrastructure/engine/sqlite_engine.go`.

### Staged Edits Before Save

- Decision: write operations are staged first and persisted only on explicit save.
- Rationale: explicit write boundary and recoverable in-session edit workflow.
- Where: `internal/interfaces/tui/model_staging_*.go`, `internal/application/usecase/save_table_changes.go`.

### Primary-Key Identity for Persisted Writes

- Decision: persisted update/delete operations use primary-key-derived identities.
- Rationale: stable app/domain contract without SQLite `rowid` coupling.
- Where: `internal/application/usecase/staged_changes_translator.go`, `internal/domain/model/update.go`.

### Strict Startup Contract with Explicit Direct-Launch Failure

- Decision: keep strict argument validation and fail-fast direct-launch behavior.
- Rationale: deterministic automation behavior and clearer startup failure semantics.
- Where: `cmd/dbc/main.go`, `cmd/dbc/startup_runtime.go`.

### Centralized Keybinding and Command Registry

- Decision: keep shortcut bindings, command aliases, and help/status hints in one registry.
- Rationale: prevent drift between key handlers and rendered guidance.
- Where: `internal/interfaces/tui/internal/primitives/input_registry.go`.

### Selector-First Decomposition Inside The TUI Adapter

- Decision: keep `internal/interfaces/tui` as the public facade/runtime package and isolate selector workflow plus low-level terminal UI primitives in internal subpackages.
- Rationale: reduce mixed-context hotspots while preserving the adapter boundary used by `cmd/dbc`.
- Where: `internal/interfaces/tui/selector.go`, `internal/interfaces/tui/internal/selector/*.go`, `internal/interfaces/tui/internal/primitives/*.go`.

### Terminal-Theme-Driven TUI Styling

- Decision: keep TUI styling limited to terminal-native ANSI attributes and the terminal's active foreground/background theme instead of app-defined colors.
- Rationale: inherit the user's terminal colorscheme automatically while preserving a predictable monochrome fallback path.
- Where: `internal/interfaces/tui/internal/primitives/render_style.go`, `internal/interfaces/tui/internal/primitives/view_layout.go`, `internal/interfaces/tui/internal/primitives/popup_component.go`, `internal/interfaces/tui/view*.go`, `internal/interfaces/tui/internal/selector/view.go`.

### Guarded Dynamic Query Composition

- Decision: keep operator allowlist, schema-validated sort columns, and identifier quoting for dynamic SQL paths.
- Rationale: reduce malformed query risk and unsafe string composition.
- Where: `internal/infrastructure/engine/sqlite_filter.go`, `internal/infrastructure/engine/sqlite_operator.go`, `internal/infrastructure/engine/sqlite_sort.go`.

### Explicit Infrastructure Test Naming

- Decision: use `*_unit_test.go` for pure logic tests and `*_integration_test.go` for filesystem-backed or real-SQLite-backed adapter tests; both run in `go test ./...`.
- Rationale: make test level obvious without splitting the default test lane.
- Where: `internal/infrastructure/{config,engine}`.

## Technology Stack and Versions

Version source: `go.mod`.

- Go language version: `1.25.0`.
- Go toolchain: `go1.25.5`.
- Direct runtime dependencies:
  - `github.com/charmbracelet/bubbletea v1.3.10`
  - `github.com/pelletier/go-toml/v2 v2.2.4`
  - `modernc.org/sqlite v1.42.2`

## Technical Constraints and Risks

- Only SQLite engine is implemented.
- Runtime/config validation cannot bootstrap a missing database file; the selected path must already exist and be reachable.
- Editing/deleting persisted rows requires primary keys.
- Runtime supports one active filter and one active sort for the selected table.
- Records reload performs `COUNT(*)` on each fetch; large tables can increase read latency.
- Selector updates/deletes are index-based and can be sensitive to concurrent external config edits.
- Shortcut/command changes must keep input registry and help/status rendering synchronized.
- Dynamic SQL paths must preserve identifier quoting, operator allowlists, and value placeholders.

## Deep-Dive References

- `docs/clean-architecture-ddd.md`
- `docs/test-driven-development.md`
- `docs/cli-parameter-and-output-standards.md`
