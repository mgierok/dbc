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

Canonical ownership note:

- This document is canonical for user-visible behavior, product scope, and user-facing constraints.

## 3. Available Capabilities

### In Scope (Current State)

- SQLite database support.
- Supported operating systems: macOS, Linux, and Windows.
- Startup informational flags via `-h` / `--help` and `-v` / `--version`.
- Optional direct CLI launch for known SQLite targets via `-d` / `--database`.
- Multi-database startup selector from local configuration.
- In-selector configuration management for startup databases:
  - Add database entry.
  - Edit database entry.
  - Delete database entry with explicit confirmation.
  - Display active config file path.
- In-session command entry supports `:config` to return to selector/management without restarting.
- Two-panel browsing experience:
  - Left panel: table list.
  - Right panel: schema view or records view for selected table.
- Schema inspection for selected table columns.
- Record browsing with continuous scrolling behavior.
- Single active filter per selected table.
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
- Startup requires at least one configured database entry.
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
  - If validation fails, entry is not saved and user stays in form with an error message.
  - Delete requires explicit confirmation.
  - Add/edit form shows a visible caret (`|`) in the active editable field.
- If user selects an existing entry that cannot be opened at startup (invalid path/connection string), DBC shows startup connection error in selector status and keeps selector active.
- After such startup connection error, user must select another reachable entry or edit the selected entry before main view can open.
- Startup selector displays active configuration file path.
- During active database session, users can open command entry with `:` and execute `:config` to return to selector/management.
- If `:config` is invoked while staged changes exist, product requires explicit dirty-state decision before navigation:
  - `save`: persist staged changes, then open selector/management only on success.
  - `discard`: clear staged changes, then open selector/management.
  - `cancel`: keep current session context and preserve staged changes.
- Startup CLI behavior in this section follows `docs/cli-parameter-and-output-standards.md` for help/discoverability, argument-validation feedback, and exit-code mapping.

### 4.2 Main Layout and Focus Model

- Layout is permanently two-panel in current state.
- Left panel shows tables.
- Right panel shows schema or records for the selected table.
- Focus can be moved between panels.
- Active panel is visually indicated.

### 4.3 Table Discovery and Schema View

- Table list excludes internal SQLite system tables.
- Table list is alphabetically sorted for predictable scanning.
- Schema view displays column name and type for selected table.
- If no schema is available yet, product shows an empty-state message.

### 4.4 Records View and Navigation

- Records view shows table data for the selected table.
- Records are fetched and displayed incrementally for continuous browsing.
- Row selection is visible in the focused records panel.
- Field focus mode is supported for cell-level editing navigation.
- Record cell content is width-constrained in the UI (truncated when needed).

### 4.5 Filtering

- User can open a guided filter popup from main browsing contexts.
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
- User enters field focus, then opens edit popup for selected cell.
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
  - User presses save.
  - Product asks for confirmation.
  - Product applies staged insert/update/delete changes as one save operation.
- On save success:
  - Staged state is cleared.
  - Records reload for current table and active filter.
- On save failure:
  - Staged state is retained.
  - Error is surfaced in status line.
- If user attempts to switch tables with unsaved changes, product requests discard confirmation.
- If user invokes `:config` with unsaved changes, product blocks navigation until one explicit decision is selected: `save`, `discard`, or `cancel`.

### 4.8 Visual State Communication

- Product mode indicator:
  - `READ-ONLY` when no staged changes.
  - `WRITE (dirty: N)` when staged changes exist.
- Visual row markers:
  - `[INS]` pending insert.
  - `[DEL]` pending delete.
- Edited cell indicator:
  - `*` marker on edited values.
- Status line communicates:
  - Current mode.
  - Current view (Schema or Records).
  - Current table.
  - Active filter summary.
  - Contextual shortcut hints.
  - Runtime status/error messages.
- Every active editable text field in the app displays a visible caret (`|`) at the insertion point.

## 5. User Flows

### Step 1: Startup and Database Selection

- User launches DBC in one of three startup modes:
  - Informational mode: `-h` / `--help` or `-v` / `--version`; startup returns informational output and exits without opening selector or database.
  - Default: centered database selector.
  - Direct launch: pass `-d` / `--database` with a SQLite path to bypass selector on successful validation.
- Selector displays active config file path.
- Each database option is presented as `Name | Connection String`.
- User can manage entries in place (`add`, `edit`, `delete` with confirmation).
- User can confirm selection or cancel startup.

### Step 2: Table Discovery and Schema Orientation

- After database selection, user enters the main two-panel view.
- Left panel lists available tables in alphabetical order.
- Right panel defaults to schema view for the selected table.

### Step 3: Record Exploration

- User switches to records view for the selected table.
- Record list loads progressively as the user navigates deeper.
- User can navigate rows and jump/page efficiently.

### Step 4: Focused Inspection with Filters

- User opens filter flow and selects:
  - Column.
  - Operator.
  - Value (when required by operator).
- Product applies one active filter for the currently selected table.

### Step 5: Controlled Data Change Workflow

- User stages inserts, edits, and deletes in records view.
- Product visually marks pending changes.
- User can undo/redo staged actions before save.
- Save action requests confirmation, then applies all staged changes together.

## 6. Interaction Model

### 6.1 Product Interaction Principles

- Keyboard-first by default: all primary actions are accessible from keyboard.
- Fast orientation: panel layout keeps navigation context visible.
- Safe-by-design editing: data changes are staged before save.
- Explicit commitment: save requires user confirmation.
- Visible state: status line communicates mode, view, selected table, filter, and key actions.
- Consistent interaction language: vim-like motions and commands are reused across key contexts.

### 6.2 Global/Main Navigation

| Action | Shortcut |
| --- | --- |
| Quit | `q`, `Ctrl+c` |
| Move down/up | `j`, `k` |
| Move left/right (field focus in records) | `h`, `l` |
| Jump to top | `gg` |
| Jump to bottom | `G` |
| Page down/up | `Ctrl+f`, `Ctrl+b` |
| Switch panel focus | `Ctrl+w h`, `Ctrl+w l`, `Ctrl+w w` |

### 6.3 View and Record Actions

| Action | Shortcut |
| --- | --- |
| Open records view / enter field focus / open edit popup | `Enter` (context dependent) |
| Exit field focus | `Esc` |
| Open command entry | `:` |
| Open filter popup | `F` |
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
| Edit popup | `Enter` confirm, `Esc` cancel, `Ctrl+n` set `NULL` (nullable fields) |
| Confirm popup (binary) | `Enter` or `y` confirm, `Esc` or `n` cancel |
| Dirty `:config` decision popup | `j/k` choose action, `Enter` or `y` select, `Esc` or `n` cancel |
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
- There is no direct shortcut to switch from Records view back to Schema view after entering Records view.
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
- Table switch with unsaved staged changes is guarded by discard confirmation.
- `:config` navigation with unsaved staged changes is guarded by explicit save/discard/cancel decision.

## 9. Glossary

- Database Entry: named configuration item that points to a SQLite database path.
- Table: primary unit of navigation and data operations.
- Column: typed field in a table schema.
- Record: single row of table data.
- Schema View: right-panel mode showing table columns and types.
- Records View: right-panel mode showing table data rows.
- Filter: active condition applied to records of selected table.
- Staged Change: pending insert/edit/delete not yet saved to the database.
- Dirty State: staged changes exist, and write mode with an unsaved changes count is active.
