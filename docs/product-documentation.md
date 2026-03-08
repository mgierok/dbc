# DBC Product Documentation

## Product Overview

DBC (Database Commander) is a terminal-first product for browsing and managing SQLite data with a keyboard-centric workflow. The current product state is optimized for fast inspection, controlled edits, and minimal context switching for users who prefer to stay in the command line.

Current product value and scope:

- Primary users are developers, DevOps/SRE practitioners, and technical analysts who are comfortable with keyboard-first tooling.
- The product supports macOS and Linux.
- The core experience covers startup database selection or direct launch, two-panel schema and record browsing, guided filtering and sorting, single-record detail inspection, and staged insert/edit/delete operations with undo/redo and explicit save.
- The current scope is limited to SQLite data access and management; it does not aim to be a general SQL console, schema-management tool, or multi-engine database client.

## Functional Behavior

### Database Configuration and Access

- DBC reads zero or more database entries from the active system configuration location.
- Empty config state (`missing file`, `empty file`, or `databases = []`) opens mandatory first-entry setup before normal browsing. Malformed config state (for example invalid TOML or invalid entry structure) stops startup with an explicit error.
- Mandatory first-entry setup requires at least one valid entry before continue and allows optional additional entries; `Esc` cancels startup.
- Each configured entry requires `name` and `db_path`. The selector shows entries in configuration order, shows the active config file path, and uses source markers `⚙` for config-backed entries and `⌨` for session-scoped direct-launch entries.
- The selector main view shows database options and status without inline shortcut or legend rows and keeps a right-aligned help hint: `Context help: ?`.
- Selector help opened with `?` is context-sensitive to the current selector mode (`browse`, `add/edit form`, or `delete confirmation`). Overflowing help can be scrolled with `j/k`, `Ctrl+f`/`Ctrl+b`, and `g`/`G`, and closes with `Esc`.
- In-selector management supports add, edit, and delete. Add and edit require non-empty `name` and `db_path`, validate the target before save, and keep the user in the form on validation failure. Delete requires explicit confirmation and may remove the last config-backed entry, leaving an empty selector state. Active editable fields show a visible caret `|`.
- If a selected config-backed entry cannot be opened, DBC keeps the selector active, surfaces the startup connection error in selector status, and requires the user to pick another reachable entry or edit the failing entry.
- Informational aliases `-h` / `--help` and `-v` / `--version` short-circuit startup and cannot be combined with direct launch. `--version` prints one stdout token: a short commit hash when revision metadata exists, otherwise `dev`.
- Direct-launch aliases `-d <db_path>` and `--database <db_path>` validate connectivity before runtime start. Success opens the main view directly; failure prints startup guidance and exits non-zero without falling back to the selector.
- Invalid usage and argument-validation failures exit with code `2` and guidance (`Error`, `Hint`, `Usage`). Startup runtime failures exit with code `1`.
- During an active session, `:` opens command entry. `:config` / `:c` returns to selector/management, `:help` / `:h` opens runtime context help, and `:quit` / `:q` exits the application.
- Runtime help is context-sensitive, lists only keybindings available where it was opened, stays open until `Esc`, and supports scrolling when content exceeds the visible area. Re-running `:help` / `:h` while help is already open leaves it open.
- Unsupported runtime commands keep the session active and surface an unknown-command status.
- If `:config` / `:c` is invoked while staged changes exist, DBC requires an explicit `save`, `discard`, or `cancel` decision before navigation.

### Main Layout and Focus Model

- The runtime layout is permanently two-panel in the current product state.
- The left panel shows tables in an independent framed box titled `Tables`. The right panel shows schema or records for the selected table in its own framed box with a context title.
- `Enter` from the left panel opens the selected table in right-panel Records view.
- `Esc` from a neutral right-panel state returns focus to left-panel table selection and forces the right panel back to Table Discovery (Schema view).
- In nested right-panel contexts, `Esc` exits the local context first before any panel transition.
- The active panel is visually indicated.
- `Ctrl+w h`, `Ctrl+w l`, and `Ctrl+w w` do not trigger runtime panel transitions.

### Table Discovery and Schema View

- Table discovery excludes internal SQLite system tables and lists visible tables in alphabetical order.
- Schema view shows column name and type for the selected table.
- If schema data is not yet available, DBC shows an empty-state message.

### Records View and Navigation

- Records view shows table data for the selected table in fixed pages of `20` persisted records. `Ctrl+f` and `Ctrl+b` move between pages, and page navigation is bounded to the available range.
- Pending insert rows marked with `✚` stay pinned at the top of the records list and do not count toward persisted-record page size.
- Row selection is visible in the focused records panel. Field focus mode supports cell-level navigation inside the records grid.
- Cell content in the records grid is width-constrained and may be truncated in the list view.
- Opening single-record detail renders the effective row state, including staged insert or edit values, as a vertical `column -> value` view. Detail content is wrapped instead of truncated, supports scrolling, and closes with `Esc`.
- Records view supports a guided sort flow that selects one column and one direction (`ASC` or `DESC`).
- Exactly one sort can be active per selected table. Re-running sort replaces the current sort, and switching tables resets sort state.
- Pending insert rows stay at the top even when sort is active.
- The records header marks the sorted column with `↑` for `ASC` and `↓` for `DESC`.

### Filtering

- Records view supports a guided filter flow that selects a column, an operator, and a value only when the chosen operator requires one.
- Supported operators are `=`, `!=`, `<`, `<=`, `>`, `>=`, `LIKE`, `IS NULL`, and `IS NOT NULL`.
- Exactly one filter can be active per selected table. Applying a new filter replaces the current one, and switching tables resets filter state.

### Data Operations (Insert, Edit, Delete)

#### Insert

- Insert stages a new record at the top of the records view.
- Pending inserts prefill column defaults when present, use `NULL` for nullable columns, and leave required columns without defaults empty until the user fills them.
- Auto-increment fields are hidden by default for pending inserts and can be revealed for explicit value entry.

#### Edit

- Editing an existing persisted record requires a table with a primary key.
- Editing is performed from field focus through an edit popup that shows column identity, type, nullability, and value entry.
- Nullable fields can be explicitly set to `NULL`.
- Boolean and enum-like fields use option selection instead of unrestricted free-text entry.
- Validation happens on confirm. Invalid values keep the popup open and surface error feedback.

#### Delete

- Delete toggles a delete marker on the selected persisted record.
- For pending inserts, delete removes the staged row immediately instead of adding a delete marker.

### Staging, Undo/Redo, and Save

- All writes are staged first. The database remains unchanged until save succeeds.
- Undo and redo are available during the current app session for staged actions in the selected table.
- Save is a confirmed action that applies staged insert, update, and delete changes as a single save operation for the current table.
- On save success, staged state is cleared and records reload for the current table with the active filter still applied.
- On save failure, staged state is retained and the error is shown in the status line.
- Attempting to switch tables with unsaved changes opens a `Switch Table` decision popup that warns about unsaved-change loss and requires an explicit discard decision before the table switch proceeds.
- Invoking `:config` / `:c` with unsaved changes blocks navigation until the user explicitly chooses `save`, `discard`, or `cancel`.

### Visual State Communication

- The product mode indicator shows `READ-ONLY` when no staged changes exist and `WRITE (dirty: N)` when staged changes are present.
- Records use visual row markers: `✚` for pending insert, `✖` for pending delete, and `✱` for edited rows. Row-state summaries in record detail use `ℹ`.
- The status bar is rendered in its own 3-row framed box. Runtime and selector popups use titled framed windows with padded content rows and a minimum height of `40%` of terminal height.
- The status bar communicates current mode, current table, persisted-record summary (`Records: current/total`), pagination summary (`Page: current/total`), active filter summary, active sort summary, right-aligned `Context help: ?`, and runtime status or error messages.
- Every active editable text field in the product shows a visible caret `|`.

## Constraints and Non-Goals

Current user-visible constraints:

- Only SQLite is supported.
- Editing and deleting persisted records requires a primary key in the table.
- Only one active filter is supported per table.
- Only one active sort is supported per table.
- There is no shortcut that switches from Records view back to Schema view while keeping right-panel focus.
- There is no dedicated clear-filter command; filter state is replaced by applying a new filter or cleared by switching tables.
- Quit does not prompt to preserve unsaved staged changes.
- Write behavior is intentionally conservative: edits are staged first, save requires explicit confirmation, dirty state stays visible, and unsaved table-switch or `:config` navigation always requires an explicit decision.

Explicit non-goals in the current product state:

- Non-SQLite or multi-engine database support.
- Schema-altering operations such as create, alter, or drop for tables, indexes, views, or triggers.
- SQL console or REPL execution.
- Bulk import or export workflows.
- User and permission management.
- Password manager integration.
- Advanced analytics, reporting, or BI workflows.

## Interaction Model

DBC is keyboard-first by design and reuses a small set of stable navigation patterns across selector, tables, records, popups, and command entry.

### Global and Runtime Navigation

| Action | Shortcut |
| --- | --- |
| Move down/up | `j`, `k` |
| Move left/right in field focus | `h`, `l` |
| Jump to top/bottom | `gg`, `G` |
| Page down/up | `Ctrl+f`, `Ctrl+b` |
| Open context help for current state | `?` |
| Open selected table in records panel | `Enter` |
| Return to left panel from neutral right-panel state | `Esc` |
| Open command entry | `:` |

### Records and Data Actions

| Action | Shortcut |
| --- | --- |
| Enter field focus | `e` |
| Open guided filter | `Shift+F` |
| Open guided sort | `Shift+S` |
| Open selected row detail | `Enter` |
| Stage insert | `i` |
| Toggle delete marker / remove pending insert | `d` |
| Undo staged action | `u` |
| Redo staged action | `Ctrl+r` |
| Save staged changes | `w` |
| Toggle auto-increment fields in pending insert | `Ctrl+a` |

### Commands, Selector, and Popup Controls

| Context | Controls |
| --- | --- |
| Runtime commands | `:config` / `:c`, `:help` / `:h`, `:quit` / `:q` |
| Startup selector navigation | `j/k`, arrow keys, `g/G`, `Home`/`End`, `Ctrl+f`/`Ctrl+b`, `PgDown`/`PgUp` |
| Startup selector management | `a` add, `e` edit, `d` delete, `Enter` select |
| Selector form | `Tab` switch field, `Ctrl+u` clear field, `Enter` save, `Esc` cancel |
| Filter and sort popups | `j/k` select, `Enter` confirm step, `Esc` close |
| Edit popup | `Enter` confirm, `Esc` cancel, `Ctrl+n` set `NULL` when field is nullable |
| Confirm and dirty-decision popups | `j/k` choose action, `Enter` or `y` confirm, `Esc` or `n` cancel |
| Help and record-detail popups | `j/k` and `Ctrl+f`/`Ctrl+b` scroll, `Esc` close |

## Glossary

- Database Entry: Named configuration item that points to a SQLite database path.
- Schema View: Right-panel mode that shows column names and types for the selected table.
- Records View: Right-panel mode that shows table rows for the selected table.
- Field Focus: Cell-level navigation mode used to select a specific field before editing.
- Filter: The single active condition applied to the selected table's records.
- Sort: The single active column ordering (`ASC` or `DESC`) applied to the selected table's records.
- Staged Change: Pending insert, edit, or delete that has not yet been saved to the database.
- Dirty State: Session state in which staged changes exist and the mode indicator shows an unsaved change count.
