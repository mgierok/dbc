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

- DBC reads zero or more database entries from the active config file at `~/.config/dbc/config.json`.
- Empty config state (`missing file`, `empty file`, or `{"databases":[]}`) opens mandatory first-entry setup before normal browsing. Malformed config state (for example invalid JSON or invalid entry structure) stops startup with an explicit error.
- Mandatory first-entry setup requires at least one valid entry before continue and allows optional additional entries; `Esc` from the forced setup form exits the application. In startup selector browse mode, `Esc` exits startup.
- Each configured entry requires `name` and `db_path`. The selector shows the active config file path, keeps config-backed entries in configuration order, and uses source markers `⚙` for config-backed entries and `⌨` for session-scoped direct-launch entries.
- If a direct-launch path does not match an existing configured SQLite path, returning to the selector during the same app session shows it as a session-scoped `⌨` entry appended after config-backed entries. If the path matches an existing configured entry, DBC reuses that config-backed entry instead of showing a duplicate session entry.
- The selector main view shows database options, selector status, and a right-aligned help hint: `Context help: ?`.
- Selector help opened with `?` is context-sensitive to the current selector mode (`browse`, `add/edit form`, or `delete confirmation`). Overflowing help can be scrolled with `j/k`, `Ctrl+f`/`Ctrl+b`, and `g`/`G`, and closes with `Esc`.
- In-selector management supports add, edit, and delete. Add creates config-backed entries. Edit and delete apply only to config-backed `⚙` entries; session-scoped `⌨` entries remain selectable but cannot be edited or deleted. Add and edit require non-empty `name` and `db_path`, validate the target before save, and keep the user in the form on validation failure. Delete requires explicit confirmation and may remove the last config-backed entry, leaving an empty selector state. Active editable fields show a visible caret `|`.
- If a selected config-backed entry cannot be opened, DBC keeps the selector active and surfaces the connection error in selector status. From startup selection, the user must pick another reachable entry or edit the failing entry. From runtime `:config` / `:c` re-entry, the retry selector can also be dismissed with browse-mode `Esc` to resume the previously active database session.
- Informational aliases `-h` / `--help` and `-v` / `--version` short-circuit startup and cannot be combined with direct launch. `--version` prints one stdout token: a short commit hash when revision metadata exists, otherwise `dev`.
- Direct-launch aliases `-d <db_path>` and `--database <db_path>` validate connectivity before runtime start. Success opens the main view directly; failure prints startup guidance and exits non-zero without falling back to the selector.
- Invalid usage and argument-validation failures exit with code `2` and guidance (`Error`, `Hint`, `Usage`). Startup runtime failures exit with code `1`.
- During an active session, `:` opens command entry from non-popup runtime views, including tables, schema, records, and record detail. Popup overlays keep their own local controls and do not open command entry on `:`. `:config` / `:c` returns to selector/management, where browse-mode `Esc` closes the selector and resumes the current runtime database; `:help` / `:h` opens runtime context help, `:w` / `:write` opens save confirmation for staged changes, `:wq` opens the same save confirmation when staged changes exist and otherwise exits immediately, `:quit` / `:q` exits the application when no staged changes exist, and `:set limit=<n>` sets the persisted-record page limit for the current app session only.
- Runtime help is context-sensitive, lists only controls available where it was opened, stays open until `Esc`, and supports scrolling when content exceeds the visible area. Re-running `:help` / `:h` while help is already open leaves it open.
- Unsupported runtime commands keep the session active and surface an unknown-command status.
- `:set limit=<n>` accepts only whole-number values in the range `1..1000`. Invalid `:set limit` input keeps the previous limit unchanged and surfaces an explicit validation error.
- A successful `:set limit=<n>` replaces any earlier session value, is never persisted into config, survives `:config` round-trips in the same app session, and resets persisted-record pagination to page `1`. If Records view is already open, records reload immediately with the new limit.
- If `:config` / `:c` is invoked while staged changes exist, DBC requires an explicit `save`, `discard`, or `cancel` decision before navigation.

### Main Layout and Focus Model

- The runtime layout is permanently two-panel in the current product state.
- The left panel shows tables in an independent framed box titled `Tables`. The right panel shows schema or records for the selected table in its own framed box with a context title.
- `Enter` from the left panel opens the selected table in right-panel Records view.
- `Esc` from a neutral right-panel state returns focus to left-panel table selection and forces the right panel back to Table Discovery (Schema view).
- In nested right-panel contexts, `Esc` exits the local context first before any panel transition.
- The active panel is visually indicated.

### Table Discovery and Schema View

- Table discovery excludes internal SQLite system tables and lists visible tables in alphabetical order.
- Schema view shows column name and type for the selected table.
- If schema data is not yet available, DBC shows an empty-state message.

### Records View and Navigation

- Records view shows table data for the selected table in persisted-record pages that default to `20` rows and can be overridden for the current app session with `:set limit=<n>`. `Ctrl+f` and `Ctrl+b` move between pages, and page navigation is bounded to the available range.
- Pending insert rows marked with `✚` stay pinned at the top of the records list and do not count toward persisted-record page size.
- Row selection is visible in the focused records panel. Field focus mode supports cell-level navigation inside the records grid.
- Cell content in the records grid is width-constrained and may be truncated in the list view.
- Opening single-record detail renders the effective row state, including staged insert or edit values, as stacked field blocks: each field shows a `column (type)` header followed by wrapped value lines. Edited fields show `✱` on the field header. Detail content is wrapped instead of truncated, supports scrolling, and closes with `Esc`.
- Records view supports a guided sort flow that selects one column and one direction (`ASC` or `DESC`).
- Exactly one sort can be active per selected table. Re-running sort replaces the current sort, and switching tables resets sort state.
- Pending insert rows stay at the top even when sort is active.
- The records header marks the sorted column with `↑` for `ASC` and `↓` for `DESC`.

### Filtering

- Records view supports a guided filter flow that selects a column, an operator, and a value only when the chosen operator requires one.
- Supported operator labels are `Equals`, `Not Equals`, `Less Than`, `Less Or Equal`, `Greater Than`, `Greater Or Equal`, `Like`, `Is Null`, and `Is Not Null` (corresponding to SQL `=`, `!=`, `<`, `<=`, `>`, `>=`, `LIKE`, `IS NULL`, and `IS NOT NULL`).
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
- Save is a confirmed action triggered via `:w` / `:write` that applies staged insert, update, and delete changes as a single save operation for the current table. `:wq` reuses the same confirmation and save behavior, then exits immediately only after a successful save.
- On save success, staged state is cleared, the status line reports the number of saved affected rows, and records reload for the current table with the active filter and sort still applied.
- On save failure, staged state is retained and the error is shown in the status line.
- Attempting to switch tables with unsaved changes opens a `Switch Table` decision popup that warns about unsaved-change loss using the affected-row count and requires an explicit discard decision before the table switch proceeds.
- Invoking `:config` / `:c` with unsaved changes blocks navigation until the user explicitly chooses `save`, `discard`, or `cancel`.
- Invoking `:quit` / `:q` with unsaved changes opens a `Quit` decision popup that warns about unsaved-change loss using the affected-row count and requires an explicit `discard and quit` or `continue editing` decision before exit proceeds.

### Visual State Communication

- The product mode indicator shows `READ-ONLY` when no staged changes exist and `WRITE (dirty: N)` when staged changes are present, where `N` is the number of unique affected rows in the current table.
- Records use visual row markers: `✚` for pending insert, `✖` for pending delete, and `✱` for edited rows. Row-state summaries in record detail use `ℹ`.
- Visual emphasis uses terminal-native text attributes instead of application-defined colors: selected items use reverse video, titles and status labels use emphasis, secondary hints are visually subdued, and error messages are emphasized with underline.
- The status bar is rendered in its own 3-row framed box. Runtime and selector popups use titled framed windows with padded content rows and a minimum height of `40%` of terminal height.
- The status bar always communicates current mode, current table, active filter summary, active sort summary, right-aligned `Context help: ?`, and runtime status or error messages. In Records view it additionally shows persisted-record summary (`Records: current/total`) and pagination summary (`Page: current/total`).
- Every active editable text field in the product shows a visible caret `|`.
- If `NO_COLOR` is set or the terminal reports `TERM=dumb`, DBC falls back to unstyled monochrome rendering.

## Constraints and Non-Goals

Current user-visible constraints:

- Only SQLite is supported.
- Editing and deleting persisted records requires a primary key in the table.
- Only one active filter is supported per table.
- Only one active sort is supported per table.
- Runtime page-limit overrides via `:set limit=<n>` are limited to the range `1..1000` and apply only to the current app session.
- There is no shortcut that switches from Records view back to Schema view while keeping right-panel focus.
- There is no dedicated clear-filter command; filter state is replaced by applying a new filter or cleared by switching tables.
- There is no dedicated clear-sort command; sort state is replaced by applying a new sort or cleared by switching tables.
- Write behavior is intentionally conservative: edits are staged first, save requires explicit confirmation, dirty state stays visible, `:wq` does not bypass save confirmation, and unsaved table-switch, `:config` navigation, or `:quit` exit always requires an explicit decision.

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
| Open command entry in non-popup runtime views | `:` |

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
| Toggle auto-increment fields in pending insert | `Ctrl+a` |

### Commands, Selector, and Popup Controls

| Context | Controls |
| --- | --- |
| Runtime commands | `:config` / `:c`, `:help` / `:h`, `:w` / `:write`, `:wq`, `:quit` / `:q`, `:set limit=<n>` |
| Startup selector navigation | `j/k`, arrow keys, `g/G`, `Home`/`End`, `Ctrl+f`/`Ctrl+b`, `PgDown`/`PgUp` |
| Startup selector browse mode | `Enter` select, `a` add, `e` edit selected config-backed entry, `d` delete selected config-backed entry, `Esc` quit |
| Runtime selector browse mode (from `:config` / `:c`) | `Enter` select, `a` add, `e` edit selected config-backed entry, `d` delete selected config-backed entry, `Esc` close |
| Selector form | `Tab` / `Shift+Tab` switch field, `Ctrl+u` clear field, `Backspace` / `Ctrl+h` delete character, `Enter` save, `Esc` cancel (`Esc` exits app during mandatory first-entry setup) |
| Filter popup | `j/k` select, `Enter` confirm step, `Esc` close; value-entry step also supports typing, `left/right`, and `Backspace` |
| Sort popup | `j/k` select, `Enter` confirm step, `Esc` close |
| Edit popup | `Enter` confirm, `Esc` cancel, `Ctrl+n` set `NULL` when field is nullable; text entry supports typing, `left/right`, and `Backspace`, while select-style fields use `j/k` |
| Command input | Type command text, `left/right` move caret, `Backspace` delete, `Enter` run, `Esc` cancel |
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
- Dirty State: Session state in which staged changes exist and the mode indicator shows an unsaved affected-row count.
