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
- `internal/interfaces/tui`: public TUI adapter facade plus runtime UI model/router, runtime write-side staging-state ownership, and runtime-session entrypoints.
- `internal/interfaces/tui/internal/selector`: selector-specific Bubble Tea model, selector view/state transitions, and selector option normalization.
- `internal/interfaces/tui/internal/primitives`: terminal UI primitives shared by runtime and selector, including key/help registry, popup/layout rendering, iconography, and style helpers.
- `internal/infrastructure/config`: JSON config loading/validation/persistence adapter.
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

### JSON Config Persistence

- Guarantee: DBC reads and writes only JSON config at the active config path.
- Guarantee: trimmed-empty config content is treated as an empty config state before JSON decoding.
- Guarantee: unknown JSON fields are rejected during config decode.
- Enforced in: `internal/infrastructure/config/config.go`.

### Staged Write Model

- Guarantee: insert/edit/delete actions are staged in memory and persisted only after explicit save confirmation.
- Guarantee: runtime write-side session state is owned as a database-scoped staging registry keyed by table name, while each table keeps isolated undo/redo history inside its own staging bucket.
- Guarantee: table switching resets read-side browsing state only and MUST NOT discard staged changes from other tables.
- Guarantee: dirty-change counting and initial insert defaults are delegated to application staging policy; dirty counts are based on affected rows, not edited cells.
- Enforced in: `internal/interfaces/tui/model_staging_state.go`, `internal/interfaces/tui/model_staging_*.go`, `internal/application/usecase/staging_policy.go`.

### Transactional Save Semantics

- Guarantee: one save applies inserts, updates, and deletes for all dirty tables in the current runtime database session in one SQLite transaction.
- Guarantee: updates targeting rows also staged for delete are skipped.
- Guarantee: save success clears the full database staging registry; save failure preserves all staged table state and blocks pending leave-runtime actions.
- Enforced in: `internal/application/usecase/save_database_changes.go`, `internal/interfaces/tui/model_staging_save_flow.go`, `internal/interfaces/tui/model_runtime_update.go`, `internal/infrastructure/engine/sqlite_update.go`.

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

- Guarantee: key bindings, exact command aliases, parameterized runtime commands, and help text are maintained in one shared primitives registry surface.
- Guarantee: the shared registry is split by concern-specific files for keys, runtime commands, runtime help/status text, and selector help/status text, but remains the single source of truth for runtime/selector input semantics.
- Guarantee: command parsing trims optional `:`, matches command keywords case-insensitively, and returns explicit validation errors for recognized malformed commands.
- Enforced in: `internal/interfaces/tui/internal/primitives/input_registry_keys.go`, `internal/interfaces/tui/internal/primitives/input_registry_runtime_commands.go`, `internal/interfaces/tui/internal/primitives/input_registry_runtime_text.go`, `internal/interfaces/tui/internal/primitives/input_registry_selector_text.go`, `internal/interfaces/tui/model_runtime_help_command.go`, `internal/interfaces/tui/model_runtime_key_dispatch.go`.

### Terminal-Native TUI Styling

- Guarantee: runtime and selector models resolve their styling profile once at construction time and render deterministically from that profile.
- Guarantee: TUI emphasis uses only ANSI SGR attributes (`bold`, `faint`, `underline`, `reverse`) on the terminal's current foreground/background theme; the application does not define its own color palette.
- Guarantee: setting `NO_COLOR` or running with `TERM=dumb` disables ANSI styling and falls back to plain text rendering.
- Enforced in: `internal/interfaces/tui/internal/primitives/render_style.go`, `internal/interfaces/tui/internal/primitives/view_layout.go`, `internal/interfaces/tui/internal/primitives/popup_component.go`, `internal/interfaces/tui/view*.go`, `internal/interfaces/tui/internal/selector/view.go`.

## Data and Interface Contracts

### Runtime Composition Entry Contract

- Process entrypoint: `cmd/dbc/main.go`.
- Runtime boundary: `internal/interfaces/tui.Run(...)`.
- Runtime startup re-entry reuses `internal/interfaces/tui.RuntimeSessionState` across repeated runtime launches in the same DBC process.
- Selector-return signal: `internal/interfaces/tui.ErrOpenConfigSelector`.

### Configuration Contract

- Active config path: `~/.config/dbc/config.json`.
- Persisted config entries: top-level `databases` array with required fields `name` and `db_path`.
- Unknown JSON fields are rejected (`DisallowUnknownFields`).
- Missing file, trimmed-empty file, and empty `databases` list are valid startup states and route to mandatory first-entry setup.
- Save behavior is atomic (`CreateTemp` + `Rename`) and creates config directory with `0700`.

### Application Port Contracts

- `Engine`: list tables, read schema, read records (with optional filter/sort), list operators, apply table changes, and apply database-wide named table-change batches atomically.
- `ConfigStore`: list/create/update/delete config entries and expose active config path.
- `DatabaseConnectionChecker`: validate candidate DB path before persisting selector add/edit changes.

### Record Identity and Change Payload Contract

- Persisted-row updates/deletes require non-empty identity keys.
- Identity keys are derived from primary-key columns and carried as typed staged values (`Text`, `IsNull`, optional `Raw`).
- Database-wide save payloads carry `table name + table changes` as a named batch contract between TUI/application and application/infrastructure layers.
- Application/domain write contracts do not depend on SQLite `rowid`.

### Records Page Contract

- Read contract returns `Rows`, `TotalCount`, and `HasMore`.
- `HasMore` is computed via look-ahead (`LIMIT limit+1`).
- Runtime records page limit defaults to `20`, but page loads and total-page calculations use the effective session-scoped limit from runtime state.
- Runtime command-driven page-limit overrides are accepted only in the bounded range `1..1000`.

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
- Runtime session state survives `:config` round-trips within the same DBC process and is not persisted into config.
- Dirty leave-runtime decisions are scoped to `:config` and `:quit`; dirty table switching is non-blocking because staged table state is database-scoped.
- Runtime record reload path ignores stale async responses using request ID checks, including after record-limit changes.
- Runtime/selector rendering assumes terminal support for UTF-8 box and marker glyphs plus standard ANSI SGR text attributes; `NO_COLOR` or `TERM=dumb` forces unstyled rendering.

## Technical Decisions and Tradeoffs

### SQLite-First with Engine Port

- Decision: keep one production engine (SQLite) behind application engine port.
- Rationale: low infrastructure complexity now, explicit seam for later engine extensions.
- Where: `internal/application/port/engine.go`, `internal/infrastructure/engine/sqlite_engine.go`.

### Staged Edits Before Save

- Decision: write operations are staged first in a database-scoped registry and persisted only on explicit save-all.
- Rationale: preserve recoverable in-session edits across table switching while keeping one explicit write boundary per opened database session.
- Where: `internal/interfaces/tui/model_staging_*.go`, `internal/application/usecase/save_database_changes.go`.

### Database-Wide Atomic Save Contract

- Decision: keep the application write contract database-scoped through named per-table batches and commit them in one SQLite transaction.
- Rationale: prevent partial persistence across dirty tables and keep the TUI independent from SQLite transaction orchestration details.
- Where: `internal/application/port/engine.go`, `internal/application/usecase/save_database_changes.go`, `internal/infrastructure/engine/sqlite_update.go`.

### Primary-Key Identity for Persisted Writes

- Decision: persisted update/delete operations use primary-key-derived identities.
- Rationale: stable app/domain contract without SQLite `rowid` coupling.
- Where: `internal/application/usecase/staged_changes_translator.go`, `internal/domain/model/update.go`.

### Strict Startup Contract with Explicit Direct-Launch Failure

- Decision: keep strict argument validation and fail-fast direct-launch behavior.
- Rationale: deterministic automation behavior and clearer startup failure semantics.
- Where: `cmd/dbc/main.go`, `cmd/dbc/startup_runtime.go`.

### JSON-Only Config

- Decision: support only `config.json` at runtime.
- Rationale: keep config handling isolated in one adapter and use only Go stdlib JSON handling.
- Where: `internal/infrastructure/config/config.go`.

### Centralized Keybinding and Command Registry

- Decision: keep shortcut bindings, command aliases, parameterized runtime commands, and help/status hints centralized in `internal/interfaces/tui/internal/primitives`, split by concern rather than reintroduced through adapter-local bridge files.
- Rationale: prevent drift between key handlers, command parsing, and rendered guidance while keeping the shared primitives surface discoverable after registry decomposition.
- Where: `internal/interfaces/tui/internal/primitives/input_registry_keys.go`, `internal/interfaces/tui/internal/primitives/input_registry_runtime_commands.go`, `internal/interfaces/tui/internal/primitives/input_registry_runtime_text.go`, `internal/interfaces/tui/internal/primitives/input_registry_selector_text.go`.

### Session-Scoped Runtime Overrides

- Decision: keep runtime-only overrides, including the records page limit, in runtime session state owned by startup/runtime orchestration instead of persisting them to config.
- Rationale: preserve temporary per-process behavior across `:config` round-trips without mutating the config file contract.
- Where: `cmd/dbc/startup_runtime.go`, `internal/interfaces/tui/app.go`, `internal/interfaces/tui/runtime_session.go`.

### Selector-First Decomposition Inside The TUI Adapter

- Decision: keep `internal/interfaces/tui` as the public facade/runtime package, isolate selector workflow plus low-level terminal UI primitives in internal subpackages, keep runtime write-side state behind `Model.staging`, and keep runtime overlay dispatch in one router with overlay workflows split into seam-specific runtime files.
- Rationale: reduce mixed-context hotspots while preserving the adapter boundary, stable top-level runtime entry points used by `cmd/dbc`, and a predictable change location for each overlay workflow.
- Where: `internal/interfaces/tui/model.go`, `internal/interfaces/tui/model_runtime_key_dispatch.go`, `internal/interfaces/tui/model_runtime_filter_sort.go`, `internal/interfaces/tui/model_runtime_edit_popup.go`, `internal/interfaces/tui/model_runtime_help_command.go`, `internal/interfaces/tui/model_runtime_record_detail.go`, `internal/interfaces/tui/model_runtime_confirm_popup.go`, `internal/interfaces/tui/model_staging_state.go`, `internal/interfaces/tui/model_staging_*.go`, `internal/interfaces/tui/selector.go`, `internal/interfaces/tui/internal/selector/*.go`, `internal/interfaces/tui/internal/primitives/*.go`.

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
  - `modernc.org/sqlite v1.42.2`

## Technical Constraints and Risks

- Only SQLite engine is implemented.
- Runtime/config validation cannot bootstrap a missing database file; the selected path must already exist and be reachable.
- Editing/deleting persisted rows requires primary keys.
- Runtime supports one active filter and one active sort for the selected table.
- Database-wide save requires cached schema for every dirty table staging bucket; losing that schema-to-staging association would make save payload construction fail.
- Records reload performs `COUNT(*)` on each fetch; large tables can increase read latency.
- Runtime-set records page limits are capped at `1000`; increasing or removing that cap requires revisiting engine-side slice preallocation in record loading.
- Selector updates/deletes are index-based and can be sensitive to concurrent external config edits.
- Shortcut/command changes must keep input registry and help/status rendering synchronized.
- Dynamic SQL paths must preserve identifier quoting, operator allowlists, and value placeholders.

## Deep-Dive References

- `docs/clean-architecture-ddd.md`
- `docs/test-driven-development.md`
- `docs/cli-parameter-and-output-standards.md`
