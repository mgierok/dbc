# DBC Product Documentation

## 1. Table of Contents

1. [Product Overview](#2-product-overview)
2. [Available Capabilities](#3-available-capabilities)
3. [Functional Behavior](#4-functional-behavior)
4. [User Flows](#5-user-flows)
5. [Interaction Model](#6-interaction-model)
6. [Constraints and Non-Goals](#7-constraints-and-non-goals)
7. [Safety and Governance](#8-safety-and-governance)
8. [Glossary](#9-glossary)

## 2. Product Overview

DBC (Database Commander) is a terminal-first product for browsing and managing database content with a keyboard-centric workflow. The current product state focuses on fast data inspection and controlled data changes without leaving the command-line environment.

Primary user segments:

- Developers inspecting schema and records during development/debugging workflows.
- DevOps/SRE practitioners validating operational data in terminal contexts.
- Technical analysts comfortable with keyboard-first tooling.

Core user value in current state:

- Faster table and data inspection than GUI context switching.
- Predictable keyboard workflow inspired by vim-like interaction patterns.
- Safer change process through staged edits, explicit confirmation, and transactional save behavior.

## 3. Available Capabilities

### In Scope (Current State)

- SQLite database support.
- Supported operating systems: macOS and Linux.
- Startup informational flags via `-h` / `--help` and `-v` / `--version`.
- Optional direct CLI launch for known SQLite targets via `-d` / `--database`.
- Multi-database startup selector from local configuration.
- In-selector configuration management for startup databases:
  - Add database entry.
  - Edit database entry.
  - Delete database entry with explicit confirmation.
  - Display active config file path.
- In-session command entry supports:
  - `:config` / `:c` to return to selector/management without restarting.
  - `:help` / `:h` to open runtime help popup reference during active session.
  - `:quit` / `:q` to exit runtime session.
- Two-panel browsing experience:
  - Left panel: table list.
  - Right panel: schema view or records view for selected table.
- Schema inspection for selected table columns.
- Record browsing with fixed-page pagination (20 records per page).
- Single-record detail inspection in right panel (`Enter`) with vertical field layout.
- Single active filter per selected table.
- Single active sort (one column + direction) per selected table.
- Staged data operations for current table:
  - Insert record.
  - Edit record fields.
  - Mark record for deletion.
  - Undo/redo staged actions.
- Explicit save confirmation and single-step save execution.
- Clear status visibility for read-only vs write mode (with unsaved changes count).

### Out of Scope (Current State)

- Non-SQLite database engines.
- Schema-altering operations (create/alter/drop tables, indexes, views, triggers).
- SQL console/REPL execution.
- Bulk import/export workflows.
- User and permission management.
- Password manager integration.

## 4. Functional Behavior

### 4.1 Database Configuration and Access

- DBC reads database entries from the active system configuration location.
- Startup supports zero or more configured database entries.
- Empty config state (`missing file`, `empty file`, or `databases = []`) opens mandatory first-entry setup before normal browsing.
- Malformed config state (for example invalid TOML or invalid entry structure) stops startup with an explicit error.
- Mandatory first-entry setup allows adding one required entry (with optional additional entries before continue); `Esc` cancels startup.
- DBC supports direct-launch aliases:
  - `-d <db_path>`
  - `--database <db_path>`
- DBC supports startup informational aliases:
  - `-h`
  - `--help`
  - `-v`
  - `--version`
- `--version` / `-v` prints one stdout token: short commit hash when revision metadata exists, otherwise `dev`.
- Informational aliases short-circuit startup and cannot be combined with direct-launch aliases.
- Invalid startup usage/argument-validation failures exit with code `2` and guidance (`Error`, `Hint`, `Usage`); startup runtime failures exit with code `1`.
- Direct launch validates target connectivity before runtime start:
  - success opens the main view directly (selector is skipped),
  - failure surfaces startup error guidance and exits non-zero (no selector fallback).
- Without direct launch, selector-first startup remains the default path.
- Each entry requires:
  - `name` (display name).
  - `db_path` (SQLite connection path/string).
- Startup selector displays entries in configuration order.
- Startup selector shows source marker for each option: `⚙` for config-backed entries and `⌨` for direct-launch session-scoped entry (in-memory only for current process, no auto-persistence).
- Startup selector supports in-app add/edit/delete management:
  - Add and edit require non-empty `name` and `db_path`.
  - Add and edit validate database target before config save.
  - If validation fails, entry is not saved and the user stays in the form with an error message.
  - Delete requires explicit confirmation.
  - Delete can remove the last config-backed entry; selector then stays open with an empty list state until a user adds an entry or exits startup.
  - Add/edit form shows a visible caret (`|`) in the active editable field.
- If a user selects an existing entry that cannot be opened at startup (invalid path/connection string), DBC shows startup connection error in selector status and keeps selector active.
- After such startup connection error, a user must select another reachable entry or edit the selected entry before the main view can open.
- Startup selector displays active configuration file path.
- During active database session, users can open command entry with `:` and execute `:config` / `:c` to return to selector/management.
- During active database session, users can press `?` to open runtime context help popup for the active panel/state.
- During active database session, users can execute `:help` / `:h` as an alias that also opens runtime context help popup.
- During active database session, users can execute `:quit` / `:q` to exit the application.
- Re-running `:help` or `:h` while help popup is already open keeps popup open.
- Help popup content lists only keybindings available in the context where help was opened (active panel/state).
- When help content exceeds visible popup height, users can scroll to reach final entries.
- Help popup closes only with `Esc`; unrelated keys keep it open.
- Unsupported runtime commands continue to show unknown-command status while session remains active.
- If `:config` or `:c` is invoked while staged changes exist, product requires explicit dirty-state decision before navigation:
  - `save`: persist staged changes, then open selector/management only on success.
  - `discard`: clear staged changes, then open selector/management.
  - `cancel`: keep current session context and preserve staged changes.
- Startup CLI behavior is consistent and user-visible across startup paths for help/discoverability, argument-validation feedback, and exit-code mapping.

### 4.2 Main Layout and Focus Model

- Layout is permanently two-panel in current state.
- Left panel shows tables.
- Right panel shows schema or records for the selected table.
- Left-panel `Enter` transitions to right panel in Records view.
- Neutral right-panel `Esc` returns focus to left-panel table selection and forces right panel to Table Discovery (Schema view).
- In nested right-panel contexts, `Esc` exits local context first before any panel transition.
- Active panel is visually indicated.

### 4.3 Table Discovery and Schema View

- Table list excludes internal SQLite system tables.
- Table list is alphabetically sorted for predictable scanning.
- Schema view displays column name and type for selected table.
- If no schema is available yet, product shows an empty-state message.

### 4.4 Records View and Navigation

- Records view shows table data for the selected table.
- Records are fetched and displayed in fixed pages of 20 persisted records.
- Users switch persisted-record pages with `Ctrl+f` (next page) and `Ctrl+b` (previous page).
- Page navigation is bounded to the available page range.
- Pending insert rows (`✚`) remain rendered at the top and are outside persisted-record page-size counting.
- Row selection is visible in the focused records panel.
- Field focus mode is supported for cell-level editing navigation.
- Record cell content is width-constrained in the UI (truncated when needed).
- Users can open a single-record detail view in records context with `Enter`.
- Single-record detail view behavior:
  - Renders selected row vertically as `column -> value`.
  - Uses effective row state (includes staged insert/edit values).
  - Does not truncate field content; values are wrapped and scrollable.
  - Closes with `Esc` and returns to records list context.
- Users can open a guided sort popup in records view with `Shift+S`.
- Sort flow is step-based:
  - Select one column.
  - Select direction (`ASC` or `DESC`).
- Exactly one sort can be active at a time for the selected table.
- Re-running sort replaces the previously active sort.
- Sort is reset when switching to a different table.
- Pending insert rows (`✚`) stay at the top of records view and are not reordered by sort.
- Records header shows active sort indicator on the sorted column:
  - `↑` for `ASC`.
  - `↓` for `DESC`.

### 4.5 Filtering

- Users can open a guided filter popup in records view with `Shift+F`.
- Filter flow is step-based:
  - Select column.
  - Select operator.
  - Input value (only if operator requires one).
- Supported operators:
  - `=`, `!=`, `<`, `<=`, `>`, `>=`, `LIKE`, `IS NULL`, `IS NOT NULL`.
- Exactly one filter can be active at a time for the selected table.
- Filter is reset when switching to a different table.

### 4.6 Data Operations (Insert, Edit, Delete)

#### Insert

- `i` stages a new record at the top of records view.
- Prefill behavior:
  - Column default value is prefilled when present.
  - Nullable columns default to `NULL`.
  - Required columns without default start as empty and must be completed before save.
- Auto-increment fields are hidden by default for pending inserts.
- `Ctrl+a` toggles visibility of auto-increment fields for explicit value entry.

#### Edit

- Existing record edits require a table with a primary key.
- Users enter field focus, then open the edit popup for the selected cell.
- Edit popup shows:
  - Column identity and type.
  - Nullable/not-null indicator.
  - Value entry control.
- Nullable fields can be explicitly set to `NULL` via shortcut.
- Boolean and enum-like fields use option selection instead of free-text typing.
- Validation occurs on confirm; invalid values remain in popup with error feedback.

#### Delete

- `d` toggles delete marker on existing selected record.
- For pending inserts, `d` removes the pending insert immediately instead of adding delete marker.

### 4.7 Staging, Undo/Redo, and Save

- All writes are staged first; database is unchanged until save.
- Undo/redo is available during the current app session for staged actions in the selected table.
- Save flow:
  - Users press save.
  - Product asks for confirmation.
  - Product applies staged insert/update/delete changes as one save operation.
- On save success:
  - Staged state is cleared.
  - Records reload for current table and active filter.
- On save failure:
  - Staged state is retained.
  - Error is surfaced in status line.
- If a user attempts to switch tables with unsaved changes, product opens a `Switch Table` decision popup with warning text: `Switching tables will cause loss of unsaved data (N changes). Are you sure you want to discard unsaved data?`.
- The table-switch decision popup exposes explicit actions:
  - `(y) Yes, discard changes and switch table`.
  - `(n) No, continue editing`.
- If a user invokes `:config` or `:c` with unsaved changes, product blocks navigation until one explicit decision is selected: `save`, `discard`, or `cancel`.

### 4.8 Visual State Communication

- Product mode indicator:
  - `READ-ONLY` when no staged changes.
  - `WRITE (dirty: N)` when staged changes exist.
- Visual row markers:
  - `✚` pending insert.
  - `✖` pending delete.
- Edited row indicator:
  - `✎` marker on edited rows in Records view.
- Record detail information indicator:
  - `ℹ` marker on row-state summary lines.
- Status line communicates:
  - Current mode.
  - Current view (Schema or Records).
  - Current table.
  - Persisted-record count summary for current page and filtered total (`Records: current/total`).
  - Persisted-record pagination summary (`Page: current/total`).
  - Active filter summary.
  - Active sort summary.
  - Right-aligned context-help hint (`Context help: ?`).
  - Runtime status/error messages.
- Every active editable text field in the app displays a visible caret (`|`) at the insertion point.

## 5. User Flows

### Step 1: Startup and Database Selection

- Users launch DBC in one of three startup modes:
  - Informational mode: `-h` / `--help` or `-v` / `--version`; startup returns informational output and exits without opening selector or database.
  - Default: centered database selector.
  - Direct launch: pass `-d` / `--database` with a SQLite path to bypass selector on successful validation.
- Selector displays active config file path.
- Each database option is presented as `Name | Connection String`.
- Users can manage entries in place (`add`, `edit`, `delete` with confirmation).
- Users can confirm selection or cancel startup.

### Step 2: Table Discovery and Schema Orientation

- After database selection, users enter the main two-panel view.
- Left panel lists available tables in alphabetical order.
- Right panel defaults to schema view for the selected table.

### Step 3: Record Exploration

- Users switch to records view for the selected table.
- Record list loads one fixed page at a time (`20` persisted records per page).
- Users move between pages with `Ctrl+f` and `Ctrl+b`.
- Users can navigate rows and jump/page efficiently.

### Step 4: Focused Inspection with Filters, Sorting, and Row Detail

- Users open the filter flow in records view and select:
  - Column.
  - Operator.
  - Value (when required by operator).
- Product applies one active filter for the currently selected table.
- Users open the sort flow and select:
  - One column.
  - `ASC` or `DESC`.
- Product applies one active sort for the currently selected table.
- Users can open single-record detail with `Enter` to inspect full field values without truncation.
- Users can leave single-record detail with `Esc` and continue records browsing.

### Step 5: Controlled Data Change Workflow

- Users stage inserts, edits, and deletes in records view.
- Product visually marks pending changes.
- Users can undo/redo staged actions before save.
- Save action requests confirmation, then applies all staged changes together.

## 6. Interaction Model

### 6.1 Product Interaction Principles

- Keyboard-first by default: all primary actions are accessible from keyboard.
- Fast orientation: panel layout keeps navigation context visible.
- Safe-by-design editing: data changes are staged before save.
- Explicit commitment: save requires user confirmation.
- Visible state: status line communicates mode, view, selected table, persisted-record count, pagination, filter, sort, runtime status, and right-aligned context-help access.
- Consistent interaction language: vim-like motions and commands are reused across key contexts.

### 6.2 Global/Main Navigation

| Action | Shortcut |
| --- | --- |
| Move down/up | `j`, `k` |
| Move left/right (field focus in records) | `h`, `l` |
| Jump to top | `gg` |
| Jump to bottom | `G` |
| Page down/up (records: next/previous persisted-record page) | `Ctrl+f`, `Ctrl+b` |
| Open runtime context help popup for current panel/state | `?` |
| Open selected table in records panel | `Enter` (from left panel) |
| Return to tables panel and force Table Discovery in right panel | `Esc` (from neutral right panel) |

### 6.3 View and Record Actions

| Action | Shortcut |
| --- | --- |
| Enter field focus / open edit popup | `e` (in records context) |
| Exit nested right-panel context | `Esc` |
| Open command entry | `:` |
| Open selector/config management (command mode) | `:config`, `:c` |
| Open runtime context help popup (command-mode alias) | `:help`, `:h` |
| Exit runtime session (command mode) | `:quit`, `:q` |
| Open filter popup | `Shift+F` |
| Open sort popup | `Shift+S` |
| Open selected row detail view | `Enter` (in records context) |
| Next/previous persisted-record page | `Ctrl+f`, `Ctrl+b` (in records context) |
| Stage insert | `i` |
| Toggle delete marker / remove pending insert | `d` |
| Undo staged action | `u` |
| Redo staged action | `Ctrl+r` |
| Save staged changes | `w` |
| Toggle auto-increment fields (pending insert row) | `Ctrl+a` |

### 6.4 Popup Interactions

| Context | Key Behavior |
| --- | --- |
| Filter popup | `j/k` selection, `Enter` confirm step, `Esc` close |
| Sort popup | `j/k` selection, `Enter` confirm step, `Esc` close |
| Edit popup | `Enter` confirm, `Esc` cancel, `Ctrl+n` set `NULL` (nullable fields) |
| Confirm popup (binary) | `Enter` or `y` confirm, `Esc` or `n` cancel |
| Dirty table-switch decision popup | `j/k` choose action, `Enter` or `y` select, `Esc` or `n` cancel |
| Dirty `:config` / `:c` decision popup | `j/k` choose action, `Enter` or `y` select, `Esc` or `n` cancel |
| Help popup | Shows keybindings for the context captured on open; `j/k` and `Ctrl+f`/`Ctrl+b` scroll, `Esc` close |
| Single-record detail view | `j/k` and `Ctrl+f`/`Ctrl+b` scroll, `Esc` close |
| Command entry | `Enter` execute command, `Esc` cancel command |

### 6.5 Startup Selector

| Action | Shortcut |
| --- | --- |
| Select database | `Enter` |
| Cancel startup | `Esc`, `q`, `Ctrl+c` |
| Move selection | `j/k` and arrow keys |
| Jump to top/bottom | `g`/`G` and `Home`/`End` |
| Page navigation | `Ctrl+f`/`Ctrl+b` and `PgDown`/`PgUp` |
| Add entry | `a` |
| Edit selected entry | `e` |
| Delete selected entry (confirm required) | `d`, then `Enter` |
| Selector form interaction | `Tab` switch field, `Ctrl+u` clear field, `Enter` save, `Esc` cancel (or exit startup during mandatory first setup) |

## 7. Constraints and Non-Goals

Current user-visible constraints:

- Only SQLite is supported.
- Editing/deleting existing records requires a primary key in the table.
- Only one active filter is supported at a time.
- Only one active sort is supported at a time.
- There is no shortcut to switch from Records view back to Schema view while keeping right-panel focus.
- No dedicated command exists to clear filter directly (filter resets on table switch or is replaced by applying a new filter).
- Quit action does not prompt to preserve unsaved staged changes.

Explicit non-goals in current state:

- Schema management and administrative operations.
- Multi-engine runtime usage.
- Advanced analytics/reporting and BI workflows.

For capability boundaries and scope classification, see Section 3.

## 8. Safety and Governance

- Safe default state: product starts in read-only mode.
- Write actions are staged and reversible before save.
- Save requires explicit confirmation.
- Save applies as one unit for the current table.
- On execution failure, product preserves staged intent for user correction.
- Dirty-state visibility is always present in status line.
- Table switch with unsaved staged changes is guarded by explicit yes/no discard decision popup with unsaved-change count warning.
- `:config` / `:c` navigation with unsaved staged changes is guarded by explicit save/discard/cancel decision.

## 9. Glossary

- Database Entry: named configuration item that points to a SQLite database path.
- Table: primary unit of navigation and data operations.
- Column: typed field in a table schema.
- Record: single row of table data.
- Schema View: right-panel mode showing table columns and types.
- Records View: right-panel mode showing table data rows.
- Filter: active condition applied to records of selected table.
- Sort: active single-column ordering (`ASC`/`DESC`) applied to records of selected table.
- Staged Change: pending insert/edit/delete not yet saved to the database.
- Dirty State: staged changes exist, and write mode with an unsaved changes count is active.
