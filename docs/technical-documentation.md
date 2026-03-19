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
- `internal/interfaces/tui`: public TUI adapter facade plus runtime UI model/router, runtime write-side staging-state ownership, runtime database-selector popup hosting, and runtime-session entrypoints.
- `internal/interfaces/tui/internal/selector`: shared database-selector controller plus selector host rendering reused by startup selection and the runtime database-selector popup.
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
- Guarantee: startup selector flow is separate from runtime database switching; startup selects only the initial database, while runtime `:config` / `:c` opens a runtime overlay popup and keeps the same runtime program alive across successful database switches.
- Guarantee: the startup selector host stays outside the runtime overlay presenter and therefore does not render the runtime backdrop treatment used by runtime popups and spotlight overlays.
- Guarantee: runtime database-selector popup dismissal is in-model only, so popup close automatically restores the prior runtime view without per-popup resume snapshots.
- Enforced in: `cmd/dbc/startup_runtime.go`, `cmd/dbc/startup_runtime_selection.go`.

### JSON Config Persistence

- Guarantee: DBC reads and writes only JSON config at the active config path.
- Guarantee: trimmed-empty config content is treated as an empty config state before JSON decoding.
- Guarantee: unknown JSON fields are rejected during config decode.
- Enforced in: `internal/infrastructure/config/config.go`.

### Staged Write Model

- Guarantee: insert/edit/delete actions are staged in memory and persisted only after explicit save confirmation.
- Guarantee: runtime write-side session state and undo/redo history stay behind `Model.staging` so staging mutations remain local to the TUI write workflow.
- Guarantee: dirty-row counting and initial insert defaults are delegated to application staging policy.
- Guarantee: the dirty count represents unique affected rows in the current table: each pending insert counts once, each persisted row with staged edits counts once regardless of edited columns, and pending deletes are deduplicated against the same persisted row already staged for update.
- Enforced in: `internal/interfaces/tui/model_staging_state.go`, `internal/interfaces/tui/model_staging_*.go`, `internal/application/usecase/staging_policy.go`.

### Transactional Save Semantics

- Guarantee: one save applies inserts, updates, and deletes in one transaction for the selected table.
- Guarantee: the save path returns the database operation's actual applied-row total aggregated across insert, update, and delete statements in that transaction.
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

- Guarantee: key bindings, exact command aliases, parameterized runtime commands, and help text are maintained in one shared primitives registry surface.
- Guarantee: runtime command-entry availability and runtime save availability are gated by one shared non-blocking runtime-context rule inside the TUI adapter, so `:`, `:w`, and context help stay aligned across `Tables`, `Schema`, `Records`, and record-detail contexts, and all of them are disabled while `saveInFlight` is active.
- Guarantee: the shared registry is split by concern-specific files for keys, runtime commands, runtime help/status text, and selector help/status text, but remains the single source of truth for runtime/selector input semantics.
- Guarantee: command parsing trims optional `:`, matches command keywords case-insensitively, and returns explicit validation errors for recognized malformed commands.
- Enforced in: `internal/interfaces/tui/internal/primitives/input_registry_keys.go`, `internal/interfaces/tui/internal/primitives/input_registry_runtime_commands.go`, `internal/interfaces/tui/internal/primitives/input_registry_runtime_text.go`, `internal/interfaces/tui/internal/primitives/input_registry_selector_text.go`, `internal/interfaces/tui/model_runtime_help_command.go`, `internal/interfaces/tui/model_runtime_key_dispatch.go`, `internal/interfaces/tui/model_runtime_command_context.go`, `internal/interfaces/tui/model_staging_save_flow.go`.

### Shared Runtime Overlay Presentation

- Guarantee: runtime help, confirm, edit, filter, sort, runtime `:config` selector, and command spotlight overlays are resolved through one shared runtime presenter instead of per-overlay host logic.
- Guarantee: when any runtime overlay is active, the runtime layout is still rendered first and then composed behind the overlay through one centered overlay path.
- Guarantee: runtime overlay priority stays centralized in the runtime view path instead of being redefined inside individual popup renderers.
- Enforced in: `internal/interfaces/tui/view.go`, `internal/interfaces/tui/view_popups.go`, `internal/interfaces/tui/view_command_spotlight.go`.

### Terminal-Native TUI Styling

- Guarantee: runtime and selector models resolve their styling profile once at construction time and render deterministically from that profile.
- Guarantee: TUI emphasis uses only ANSI SGR attributes (`bold`, `faint`, `underline`, `reverse`, `strike`) on the terminal's current foreground/background theme; the application does not define its own color palette.
- Guarantee: every user-visible TUI text MUST be assigned a semantic role before final ANSI rendering; plain text MAY remain only for structural layout elements without standalone semantic meaning.
- Guarantee: semantic roles are the single source of truth for TUI visual meaning across normal and backdrop rendering profiles.
- Guarantee: the active semantic-role catalog is `Body`, `Muted`, `Title`, `Header`, `Summary`, `Label`, `Dirty`, `Error`, `Selected`, `Deleted`, and `SelectedDeleted`.
- Guarantee: `Body` is the default user-content role for regular values, rows, selector content, popup rows, and command input content.
- Guarantee: `Muted` is reserved for secondary helper text such as contextual hints and scroll indicators.
- Guarantee: `Title` is reserved for panel and popup titles; `Header` is reserved for structural headers inside content such as column and field names.
- Guarantee: `Summary` is reserved for short contextual state summaries; `Label` is reserved for label-value prefixes such as `Table:` or `Path:`.
- Guarantee: `Dirty` is reserved for explicit unsaved-change tokens; `Error` is reserved for validation and runtime error content.
- Guarantee: `Selected`, `Deleted`, and `SelectedDeleted` are reserved for interactive stateful content and MUST be preferred over local ANSI overrides for selected or delete-marked rows.
- Guarantee: role selection MUST prefer an existing semantic role first; a new role MAY be added only when a text category has a stable distinct meaning and requires a different style mapping in at least one active render profile.
- Guarantee: runtime background rendering accepts an explicit render-style profile, so the shared runtime overlay presenter can render the background in a backdrop variant while keeping overlay content in the normal profile.
- Guarantee: the backdrop variant stays terminal-theme-driven and uses only subdued ANSI attributes; it does not introduce an application-defined palette or colorscheme detection.
- Guarantee: setting `NO_COLOR` or running with `TERM=dumb` disables ANSI styling and falls back to plain text rendering without changing the shared overlay composition path.
- Enforced in: `internal/interfaces/tui/internal/primitives/render_style.go`, `internal/interfaces/tui/internal/primitives/view_layout.go`, `internal/interfaces/tui/internal/primitives/popup_component.go`, `internal/interfaces/tui/view*.go`, `internal/interfaces/tui/internal/selector/view.go`.

## Data and Interface Contracts

### Runtime Composition Entry Contract

- Process entrypoint: `cmd/dbc/main.go`.
- Runtime boundary: `internal/interfaces/tui.Run(...)`.
- Runtime startup opens one initial `RuntimeRunDeps` bundle and passes a runtime database-switch callback implemented in `cmd/dbc`.
- `internal/interfaces/tui.RuntimeSessionState` survives in-process runtime database switches, while per-database runtime model state is recreated from a fresh dependency bundle.

### Configuration Contract

- Active config path: `~/.config/dbc/config.json`.
- Persisted config entries: top-level `databases` array with required fields `name` and `db_path`.
- Unknown JSON fields are rejected (`DisallowUnknownFields`).
- Missing file, trimmed-empty file, and empty `databases` list are valid startup states and route to mandatory first-entry setup.
- Save behavior is atomic (`CreateTemp` + `Rename`) and creates config directory with `0700`.

### Application Port Contracts

- `Engine`: list tables, read schema, read records (with optional filter/sort), list operators, apply table changes, and return the total applied-row count for that save operation.
- `ConfigStore`: list/create/update/delete config entries and expose active config path.
- `DatabaseConnectionChecker`: validate candidate DB path before persisting selector add/edit changes.

### Record Identity and Change Payload Contract

- Persisted-row updates/deletes require non-empty identity keys.
- Identity keys are derived from primary-key columns and carried as typed staged values (`Text`, `IsNull`, optional `Raw`).
- Application/domain write contracts do not depend on SQLite `rowid`.

### Records Page Contract

- Read contract returns `Rows`, `TotalCount`, and `HasMore`.
- `HasMore` is computed via look-ahead (`LIMIT limit+1`).
- Runtime records page limit defaults to `20`, but page loads and total-page calculations use the effective session-scoped limit from runtime state.
- Runtime command-driven page-limit overrides are accepted only in the bounded range `1..1000`.

### Selector Launch Contract

- Selector launch state supports preferred connection string reselection and additional session-scoped options.
- Startup selector API remains `SelectDatabaseWithState(...)` and is startup-only; runtime popup selection does not route through that startup contract.
- Session-scoped options are CLI-origin and are not persisted into config.
- Selector edit/delete operations are allowed only for config-backed entries.

## Runtime and Operational Considerations

- Supported startup operating systems: `darwin`, `linux`.
- Startup usage failures emit deterministic `Error`/`Hint`/`Usage` stderr output.
- SQLite open contract requires existing file path (missing files and directory paths fail fast).
- Startup database open and config add/edit validation share the same open/ping helper.
- Runtime closes active DB handle on session end; close failures are logged.
- Runtime session state survives in-process runtime database switches within the same DBC process and is not persisted into config.
- Runtime `:config` / `:c` opens an in-runtime database-selector popup over the current layout; `Esc` closes only that popup and leaves runtime state untouched.
- Active runtime popups and the command spotlight keep the runtime layout visible underneath through the shared runtime backdrop presenter; the startup selector remains the only intentional selector exception to that rule.
- Runtime database switching is async inside the TUI adapter: failed switch preparation keeps the current session active and leaves the selector popup open with selector status, while successful switch preparation atomically swaps runtime dependencies in-process and closes the previous DB handle only after the replacement bundle is ready.
- Runtime save is input-blocking inside the TUI adapter: user key input is ignored until the async `saveChangesMsg` response arrives, while terminal resize remains handled through `WindowSizeMsg`.
- Runtime record reload path ignores stale async responses using request ID checks, including after record-limit changes.
- Runtime/selector rendering assumes terminal support for UTF-8 box and marker glyphs plus standard ANSI SGR text attributes; `NO_COLOR` or `TERM=dumb` forces unstyled rendering.
- Correct runtime layout visibility requires terminal height of at least `10` rows. This preserves the fixed 3-row status box and leaves at least 5 content rows available in both main panels; lower heights are outside the supported rendering contract.

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
- Rationale: preserve temporary per-process behavior across runtime database switches without mutating the config file contract.
- Where: `cmd/dbc/startup_runtime.go`, `internal/interfaces/tui/app.go`, `internal/interfaces/tui/runtime_session.go`.

### Selector-First Decomposition Inside The TUI Adapter

- Decision: keep `internal/interfaces/tui` as the public facade/runtime package, isolate selector workflow plus low-level terminal UI primitives in internal subpackages, keep runtime write-side state behind `Model.staging`, centralize runtime overlay dispatch in one runtime presenter, and reuse one selector controller across a startup host and a runtime popup host.
- Rationale: reduce mixed-context hotspots while preserving the adapter boundary, stable top-level runtime entry points used by `cmd/dbc`, a predictable change location for overlay composition, and the startup/runtime selector split without reintroducing runtime restart/resume logic for selector dismissal.
- Where: `internal/interfaces/tui/model.go`, `internal/interfaces/tui/model_runtime_key_dispatch.go`, `internal/interfaces/tui/model_runtime_filter_sort.go`, `internal/interfaces/tui/model_runtime_edit_popup.go`, `internal/interfaces/tui/model_runtime_help_command.go`, `internal/interfaces/tui/model_runtime_record_detail.go`, `internal/interfaces/tui/model_runtime_confirm_popup.go`, `internal/interfaces/tui/model_staging_state.go`, `internal/interfaces/tui/model_staging_*.go`, `internal/interfaces/tui/selector.go`, `internal/interfaces/tui/internal/selector/*.go`, `internal/interfaces/tui/internal/primitives/*.go`.

### Terminal-Theme-Driven TUI Styling

- Decision: keep TUI styling limited to terminal-native ANSI attributes and the terminal's active foreground/background theme instead of app-defined colors, including for shared runtime backdrop rendering, with semantic text roles as the only style-selection input.
- Rationale: inherit the user's terminal colorscheme automatically while preserving a predictable monochrome fallback path, one overlay composition architecture for both styled and unstyled terminals, and one central mapping from text meaning to presentation.
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
- Records reload performs `COUNT(*)` on each fetch; large tables can increase read latency.
- Runtime-set records page limits are capped at `1000`; increasing or removing that cap requires revisiting engine-side slice preallocation in record loading.
- Selector updates/deletes are index-based and can be sensitive to concurrent external config edits.
- Shortcut/command changes must keep input registry, help/status rendering, and adapter-side runtime context gating synchronized.
- Dynamic SQL paths must preserve identifier quoting, operator allowlists, and value placeholders.

## Deep-Dive References

- `docs/clean-architecture-ddd.md`
- `docs/test-driven-development.md`
- `docs/cli-parameter-and-output-standards.md`
