# DBC Technical Documentation

## Technical Overview

- This document describes the current factual implementation state and helps readers navigate the repository as it exists today.
- Canonical architecture guidance for future changes lives in `docs/clean-architecture-ddd.md`, and operational agent rules live in `AGENTS.md`.
- DBC is a terminal SQLite data browser/editor written in Go.
- Runtime behavior is split into read operations through application ports and staged write operations applied only on explicit save.
- The current package layout is organized around Clean Architecture / DDD-style boundaries with inward dependency direction.
- Current implementation target is macOS/Linux runtime with SQLite as the only engine.

## Architecture and Boundaries

### Current Dependency Shape

- The current dependency shape is `interfaces -> application -> domain` and `infrastructure -> application/domain`.
- Current imports keep the domain layer isolated from `application`, `interfaces`, and `infrastructure`.
- Current imports keep the application layer isolated from `interfaces` and `infrastructure`.
- Current interface adapters do not import infrastructure adapters directly.

### Boundary Contracts

- Database access currently goes through `internal/application/port.Engine`.
- Use cases currently orchestrate behavior against ports and stay independent from SQLite-specific details.
- Runtime dirty-navigation orchestration now lives in application use cases; the TUI adapter renders prompts, keeps only interaction-local continuation metadata, and executes the adapter-side next action returned by application.
- Infrastructure packages currently implement boundary ports (`Engine`, `ConfigStore`, `DatabaseConnectionChecker`).

## Components and Responsibilities

- `cmd/dbc`: process entrypoint, startup argument handling, runtime/selector orchestration.
- `internal/domain/model`: domain value objects, entities, and error contracts.
- `internal/domain/service`: pure domain helpers (table sorting, typed value parsing, input spec inference).
- `internal/application/usecase`: read/write orchestration, config management, staging policy, runtime navigation workflow, and shared runtime/startup database-target resolution.
- `internal/application/port`: application boundary interfaces for infrastructure implementations.
- `internal/application/dto`: adapter-facing data contracts exchanged between use cases and interfaces.
- `internal/interfaces/tui`: public TUI adapter facade plus runtime UI model/router, runtime write-side staging-state ownership, runtime database-selector popup hosting, prompt rendering, and runtime-session entrypoints.
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
- Guarantee: direct-launch path resolves configured identity through the same application-level SQLite target resolver used by runtime reopen requests.
- Guarantee: direct-launch failure exits non-zero without selector fallback.
- Guarantee: startup selector flow is separate from runtime database reopening; startup selects only the initial database, while runtime `:config` / `:c` uses a selector popup surface and runtime `:edit` / `:e` uses the command spotlight surface to request a selected database, then `cmd/dbc` reopens that database in a fresh runtime instance.
- Guarantee: the startup selector host stays outside the runtime overlay presenter and therefore does not render the runtime backdrop treatment used by runtime popups and spotlight overlays.
- Guarantee: runtime database-selector popup dismissal is in-model only, so popup close automatically restores the prior runtime view without per-popup resume snapshots.
- Enforced in: `cmd/dbc/startup_runtime.go`, `cmd/dbc/startup_runtime_selection.go`.

### Runtime Navigation Orchestration

- Guarantee: dirty table switch, runtime database reload/open, and runtime quit are planned in one application-level navigation workflow instead of being classified directly in the TUI adapter.
- Guarantee: runtime dirty confirm popups render application-provided decision IDs and labels; the TUI does not keep separate dirty-flow enums for table switch, database transition, or quit semantics.
- Guarantee: successful save after a dirty database navigation request resumes the pending application action, while save failure keeps that pending action and restores submitted `:edit` command input when needed.
- Guarantee: adapter-side execution remains limited to switching to a resolved table name, quitting runtime, or exiting with `OpenDatabaseNext`.
- Enforced in: `internal/application/usecase/runtime_navigation_workflow.go`, `internal/application/usecase/runtime_database_target_resolver.go`, `internal/interfaces/tui/model_runtime_confirm_popup.go`, `internal/interfaces/tui/model_runtime_database_transition.go`, `internal/interfaces/tui/model_runtime_update.go`.

### JSON Config Persistence

- Guarantee: DBC reads and writes only JSON config at the active config path.
- Guarantee: config reads and writes are bounded to `1 MiB`; oversized input or oversized serialized output fail explicitly with `ErrConfigTooLarge`.
- Guarantee: trimmed-empty config content is treated as an empty config state before JSON decoding.
- Guarantee: unknown JSON fields are rejected during config decode.
- Enforced in: `internal/infrastructure/config/config.go`.

### Staged Write Model

- Guarantee: insert/edit/delete actions are staged in memory and persisted only after an explicit save command (`:w`, `:write`, or dirty-flow save choice).
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

### SQLite Schema Introspection

- Guarantee: schema reads keep using SQLite PRAGMA introspection and expose per-column metadata for name, type, nullability, primary-key membership, single-column uniqueness, default value, autoincrement, and foreign-key references.
- Guarantee: single-column `UNIQUE` is derived from `PRAGMA index_list(...)` plus `PRAGMA index_info(...)`; composite unique memberships are not surfaced as per-column `UNIQUE`, and primary-key columns are not duplicated as `UNIQUE`.
- Guarantee: foreign-key references are derived from `PRAGMA foreign_key_list(...)` and mapped per source column, including composite foreign keys.
- Enforced in: `internal/infrastructure/engine/sqlite_record_materialization.go`, `internal/infrastructure/engine/sqlite_engine.go`.

### Input Normalization and Typed Parsing

- Guarantee: staged values are parsed by column type and nullability before persistence payload generation.
- Guarantee: boolean and enum-like schema types expose select-style input metadata.
- Enforced in: `internal/domain/service/value_parser.go`, `internal/application/usecase/get_schema.go`, `internal/application/usecase/staged_changes_translator.go`.

### Centralized Runtime and Selector Command Registry

- Guarantee: key bindings, exact command aliases, parameterized runtime commands, and help text are maintained in one shared primitives registry surface.
- Guarantee: runtime command-entry availability and runtime save availability are gated by one shared non-blocking runtime-context rule inside the TUI adapter, so `:`, `:w`, and context help stay aligned across `Tables`, `Schema`, `Records`, and record-detail contexts, and all of them are disabled while `saveInFlight` is active.
- Guarantee: the shared registry is split by concern-specific files for keys, runtime commands, runtime help/status text, and selector help/status text, while remaining the central definition point for runtime/selector input semantics.
- Guarantee: command parsing trims optional `:`, matches command keywords case-insensitively, and returns explicit validation errors for recognized malformed commands.
- Enforced in: `internal/interfaces/tui/internal/primitives/input_registry_keys.go`, `internal/interfaces/tui/internal/primitives/input_registry_runtime_commands.go`, `internal/interfaces/tui/internal/primitives/input_registry_runtime_text.go`, `internal/interfaces/tui/internal/primitives/input_registry_selector_text.go`, `internal/interfaces/tui/model_runtime_help_command.go`, `internal/interfaces/tui/model_runtime_key_dispatch.go`, `internal/interfaces/tui/model_runtime_command_context.go`, `internal/interfaces/tui/model_staging_save_flow.go`.

### Shared Runtime Overlay Presentation

- Guarantee: runtime help, confirm, edit, filter, sort, runtime `:config` selector, and command spotlight overlays are resolved through one shared runtime presenter instead of per-overlay host logic.
- Guarantee: the command spotlight stays editable-only; accepted `:edit` requests exit the current runtime instead of entering an in-runtime pending mode.
- Guarantee: when any runtime overlay is active, the runtime layout is still rendered first and then composed behind the overlay through one centered overlay path.
- Guarantee: runtime overlay priority stays centralized in the runtime view path instead of being redefined inside individual popup renderers.
- Enforced in: `internal/interfaces/tui/view.go`, `internal/interfaces/tui/view_popups.go`, `internal/interfaces/tui/view_command_spotlight.go`.

### Terminal-Native TUI Styling

- Guarantee: runtime and selector models resolve their styling profile once at construction time and render deterministically from that profile.
- Guarantee: TUI emphasis uses only ANSI SGR attributes (`bold`, `faint`, `underline`, `reverse`, `strike`) on the terminal's current foreground/background theme; the application does not define its own color palette.
- Guarantee: the current renderer assigns every user-visible TUI text a semantic role before final ANSI rendering; plain text remains only for structural layout elements without standalone semantic meaning.
- Guarantee: semantic roles remain the central meaning-to-style mapping across normal and backdrop rendering profiles.
- Guarantee: the active semantic-role catalog is `Body`, `Muted`, `Title`, `Header`, `Summary`, `Label`, `Dirty`, `Error`, `Selected`, `Deleted`, and `SelectedDeleted`.
- Guarantee: regular content, helper text, structural headers, dirty/error states, and selected/delete-marked rows each map through those semantic roles instead of ad hoc local ANSI overrides.
- Guarantee: runtime background rendering accepts an explicit render-style profile, so the shared runtime overlay presenter can render the background in a backdrop variant while keeping overlay content in the normal profile.
- Guarantee: the backdrop variant stays terminal-theme-driven and uses only subdued ANSI attributes; it does not introduce an application-defined palette or colorscheme detection.
- Guarantee: setting `NO_COLOR` or running with `TERM=dumb` disables ANSI styling and falls back to plain text rendering without changing the shared overlay composition path.
- Enforced in: `internal/interfaces/tui/internal/primitives/render_style.go`, `internal/interfaces/tui/internal/primitives/view_layout.go`, `internal/interfaces/tui/internal/primitives/popup_component.go`, `internal/interfaces/tui/view*.go`, `internal/interfaces/tui/internal/selector/view.go`.

## Data and Interface Contracts

### Runtime Composition Entry Contract

- Process entrypoint: `cmd/dbc/main.go`.
- Runtime boundary: `internal/interfaces/tui.Run(...)`.
- Runtime startup opens one initial `RuntimeRunDeps` bundle, including the runtime save workflow, runtime navigation workflow, and shared database-target resolver, and `internal/interfaces/tui.Run(...)` returns a `RuntimeExitResult` that either quits normally or requests reopening a selected database.
- Runtime-initiated reopen requests carry a fully resolved `DatabaseOption`, including whether the target is config-backed or CLI-scoped.

### Configuration Contract

- Active config path: `~/.config/dbc/config.json`.
- Persisted config entries: top-level `databases` array with required fields `name` and `db_path`.
- Unknown JSON fields are rejected (`DisallowUnknownFields`).
- Missing file, trimmed-empty file, and empty `databases` list are valid startup states and route to mandatory first-entry setup.
- Save behavior is atomic (`CreateTemp` + `Rename`) and creates config directory with `0700`.

### Application Port Contracts

- `Engine`: list tables, read schema, read records (with optional filter/sort), list operators, apply table changes, and return the total applied-row count for that save operation.
- Read-record responses carry render-facing `Values` separately from persisted-row identity data, so browse placeholders do not change write identity.
- Read-record responses also carry per-cell browse-edit safety metadata; the application-layer persisted-record access resolver consumes that metadata to decide whether edit may start from the current browse value.
- `ConfigStore`: list/create/update/delete config entries and expose active config path.
- `DatabaseConnectionChecker`: validate candidate DB path before persisting selector add/edit changes.

### Schema Read Contract

- Schema read contracts (`model.Column` and `dto.SchemaColumn`) carry `Name`, `Type`, `Nullable`, `PrimaryKey`, `Unique`, `DefaultValue`, `AutoIncrement`, and zero or more foreign-key references; `dto.SchemaColumn` additionally carries ordered precomputed `MetadataBadges` for adapter rendering.
- Foreign-key references are value objects with `Table` and optional `Column`; when SQLite does not expose a referenced column name, the contract keeps `Column` empty instead of synthesizing one.
- The application `GetSchema` use case preserves schema meaning while mapping the richer engine/domain contract into TUI-facing DTOs and remains the central mapping point for metadata-badge projection used by `Schema` and record-detail rendering.

### Record Identity and Change Payload Contract

- Persisted-row updates/deletes require non-empty identity keys.
- Identity keys are derived from primary-key columns and carried as typed staged values (`Text`, `IsNull`, optional `Raw`).
- Read-path record contracts (`model.Record` and `dto.RecordRow`) may provide precomputed row key + identity, and the application persisted-record access resolver prefers that precomputed identity over reparsing rendered values.
- If any primary-key component exceeds the browse materialization safety cap, the read contract marks row identity unavailable instead of materializing an oversized key; edit/delete then stay blocked for that row.
- The application persisted-record access resolver owns persisted-row access semantics for delete and edit-start (`ResolveForDelete` / `ResolveForEdit`) and returns a boundary DTO containing `RowKey` plus typed `Identity`.
- Application/domain write contracts do not depend on SQLite `rowid`.

### Records Page Contract

- Read contract returns `Rows`, `TotalCount`, and `HasMore`.
- Browse materialization is bounded to `256 KiB` per cell on read paths.
- Oversized non-BLOB cells render as `<truncated N bytes>`.
- `BLOB` cells render as size placeholders (`<blob N bytes>` / `<blob truncated N bytes>`) instead of raw binary/text coercions.
- Per-cell browse-edit safety metadata marks synthetic placeholders as not editable-from-browse; `ResolveForEdit` rejects such cells for browse-started edit, while the TUI keeps the only local exception for reopening the popup from an already staged value.
- Materialized display aliases stay internal to the projection, while `ORDER BY` continues to target the raw table columns so sort semantics remain identical to stored SQLite values.
- `HasMore` is computed via look-ahead (`LIMIT limit+1`).
- Runtime records page limit defaults to `20`, but page loads and total-page calculations use the effective runtime-local limit from runtime state.
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
- Runtime `:config` / `:c` opens an in-runtime database-selector popup over the current layout; `Esc` closes only that popup and leaves runtime state untouched.
- Active runtime popups and the command spotlight keep the runtime layout visible underneath through the shared runtime backdrop presenter; the startup selector remains the only intentional selector exception to that rule.
- Runtime `:w` / `:write` starts save immediately when dirty and surfaces a non-error no-op status when no staged changes exist; runtime `:wq` starts save immediately when dirty and quits only after a successful save.
- Runtime `:config` / `:c` and `:edit` / `:e` resolve the requested target through the shared application-level target resolver and runtime navigation workflow, then exit the current runtime with a structured reopen result; if dirty state exists, save/discard/cancel is resolved before exit.
- `cmd/dbc` consumes runtime reopen results, tracks CLI-scoped targets, attempts the requested reopen through the normal startup/runtime loop, and falls back to the fullscreen selector with status + preferred connection string when that reopen fails.
- Successful runtime-initiated reopen always starts a fresh per-database runtime model in a safe base state (`FocusTables` + `ViewSchema`), does not restore prior browse-state snapshots, and resets runtime-local record-limit overrides to the default `20`.
- Runtime save is input-blocking inside the TUI adapter: user key input is ignored until the async `saveChangesMsg` response arrives, while terminal resize remains handled through `WindowSizeMsg`.
- Runtime record reload path ignores stale async responses using request ID checks, including after record-limit changes.
- Runtime/selector rendering assumes terminal support for UTF-8 box and marker glyphs plus standard ANSI SGR text attributes; `NO_COLOR` or `TERM=dumb` forces unstyled rendering.
- Correct runtime layout visibility requires terminal height of at least `10` rows. This preserves the fixed 3-row status box and leaves at least 5 content rows available in both main panels; lower heights are outside the supported rendering contract.

## Technical Decisions and Tradeoffs

### SQLite-First with Engine Port

- Current decision: one production engine (SQLite) sits behind the application engine port.
- Rationale: low infrastructure complexity now, explicit seam for later engine extensions.
- Where: `internal/application/port/engine.go`, `internal/infrastructure/engine/sqlite_engine.go`.

### Staged Edits Before Save

- Current decision: write operations are staged first and persisted only on explicit save.
- Rationale: explicit write boundary and recoverable in-session edit workflow.
- Where: `internal/interfaces/tui/model_staging_*.go`, `internal/application/usecase/save_table_changes.go`.

### Primary-Key Identity for Persisted Writes

- Current decision: persisted update/delete operations use primary-key-derived identities.
- Rationale: stable app/domain contract without SQLite `rowid` coupling.
- Where: `internal/application/usecase/staged_changes_translator.go`, `internal/domain/model/update.go`.

### Strict Startup Contract with Explicit Direct-Launch Failure

- Current decision: startup keeps strict argument validation and fail-fast direct-launch behavior.
- Rationale: deterministic automation behavior and clearer startup failure semantics.
- Where: `cmd/dbc/main.go`, `cmd/dbc/startup_runtime.go`.

### Application-Owned Runtime Navigation Planning

- Current decision: dirty runtime navigation planning and SQLite database-target classification live in application use cases and are reused by both runtime reopen and direct-launch identity resolution.
- Rationale: keep TUI limited to adapter concerns, remove duplicated SQLite identity logic, and centralize save/discard/cancel continuation rules in one application seam.
- Where: `internal/application/usecase/runtime_navigation_workflow.go`, `internal/application/usecase/runtime_database_target_resolver.go`, `cmd/dbc/startup_runtime_selection.go`, `internal/interfaces/tui/model_runtime_database_transition.go`.

### JSON-Only Config

- Current decision: runtime supports only `config.json`.
- Rationale: keep config handling isolated in one adapter and use only Go stdlib JSON handling.
- Where: `internal/infrastructure/config/config.go`.

### Centralized Keybinding and Command Registry

- Current decision: shortcut bindings, command aliases, parameterized runtime commands, and help/status hints stay centralized in `internal/interfaces/tui/internal/primitives`, split by concern rather than reintroduced through adapter-local bridge files.
- Rationale: prevent drift between key handlers, command parsing, and rendered guidance while keeping the shared primitives surface discoverable after registry decomposition.
- Where: `internal/interfaces/tui/internal/primitives/input_registry_keys.go`, `internal/interfaces/tui/internal/primitives/input_registry_runtime_commands.go`, `internal/interfaces/tui/internal/primitives/input_registry_runtime_text.go`, `internal/interfaces/tui/internal/primitives/input_registry_selector_text.go`.

### Session-Scoped Runtime Overrides

- Current decision: runtime-only overrides, including the records page limit, stay inside one runtime instance instead of being persisted to config or carried across runtime-initiated reopen.
- Rationale: preserve temporary local behavior without mutating the config contract and keep reopened runtimes deterministic.
- Where: `internal/interfaces/tui/app.go`, `internal/interfaces/tui/model.go`, `internal/interfaces/tui/model_runtime_record_limit.go`, `internal/interfaces/tui/runtime_session.go`.

### Selector-First Decomposition Inside The TUI Adapter

- Current decision: `internal/interfaces/tui` remains the public facade/runtime package, selector workflow plus low-level terminal UI primitives stay isolated in internal subpackages, runtime write-side state stays behind `Model.staging`, runtime overlay dispatch stays centralized in one runtime presenter, and one selector controller is reused across a startup host and a runtime popup host.
- Rationale: reduce mixed-context hotspots while preserving the adapter boundary, stable top-level runtime entry points used by `cmd/dbc`, a predictable change location for overlay composition, and the startup/runtime selector split without reintroducing runtime restart/resume logic for selector dismissal.
- Where: `internal/interfaces/tui/model.go`, `internal/interfaces/tui/model_runtime_key_dispatch.go`, `internal/interfaces/tui/model_runtime_filter_sort.go`, `internal/interfaces/tui/model_runtime_edit_popup.go`, `internal/interfaces/tui/model_runtime_help_command.go`, `internal/interfaces/tui/model_runtime_record_detail.go`, `internal/interfaces/tui/model_runtime_confirm_popup.go`, `internal/interfaces/tui/model_staging_state.go`, `internal/interfaces/tui/model_staging_*.go`, `internal/interfaces/tui/selector.go`, `internal/interfaces/tui/internal/selector/*.go`, `internal/interfaces/tui/internal/primitives/*.go`.

### Terminal-Theme-Driven TUI Styling

- Current decision: TUI styling stays limited to terminal-native ANSI attributes and the terminal's active foreground/background theme instead of app-defined colors, including for shared runtime backdrop rendering, with semantic text roles as the only style-selection input.
- Rationale: inherit the user's terminal colorscheme automatically while preserving a predictable monochrome fallback path, one overlay composition architecture for both styled and unstyled terminals, and one central mapping from text meaning to presentation.
- Where: `internal/interfaces/tui/internal/primitives/render_style.go`, `internal/interfaces/tui/internal/primitives/view_layout.go`, `internal/interfaces/tui/internal/primitives/popup_component.go`, `internal/interfaces/tui/view*.go`, `internal/interfaces/tui/internal/selector/view.go`.

### Guarded Dynamic Query Composition

- Current decision: dynamic SQL paths keep operator allowlists, schema-validated sort columns, and identifier quoting.
- Rationale: reduce malformed query risk and unsafe string composition.
- Where: `internal/infrastructure/engine/sqlite_filter.go`, `internal/infrastructure/engine/sqlite_operator.go`, `internal/infrastructure/engine/sqlite_sort.go`.

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
