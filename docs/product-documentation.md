# DBC Product Documentation

## Document Control

| Field | Value |
| --- | --- |
| Document Name | DBC Product Documentation |
| Product Name | DBC (Database Commander) |
| Audience | Junior Product Manager and Junior Software Engineer |
| Purpose | Define the current product state from a business and user-value perspective |
| Status | Active |
| Last Updated | 2026-02-16 |
| Source of Truth Scope | Current product behavior and scope (product perspective) |
| Related technical doc | `docs/technical-documentation.md` |

## Table of Contents

1. [Product Summary](#1-product-summary)
2. [Positioning and Value Proposition](#2-positioning-and-value-proposition)
3. [Target Users and Core Jobs](#3-target-users-and-core-jobs)
4. [Current Product Scope](#4-current-product-scope)
5. [Product Experience Principles](#5-product-experience-principles)
6. [End-to-End User Journey](#6-end-to-end-user-journey)
7. [Functional Specification (Current State)](#7-functional-specification-current-state)
8. [Keyboard Interaction Model](#8-keyboard-interaction-model)
9. [Data Safety and Change Governance](#9-data-safety-and-change-governance)
10. [Known Constraints and Non-Goals](#10-known-constraints-and-non-goals)
11. [Glossary](#11-glossary)
12. [Cross-References to Technical Documentation](#12-cross-references-to-technical-documentation)

## 1. Product Summary

DBC is a terminal-first product for browsing and managing database content with a keyboard-centric workflow. The product is designed for users who need rapid data inspection and controlled data changes without leaving the command-line environment.

Canonical ownership note:

- This document is canonical for user-facing behavior, scope, and constraints.
- Implementation details and code-level decisions are canonical in `docs/technical-documentation.md`.

At present, DBC supports SQLite databases and delivers two major capability groups:

- Read and inspect: database selection, table discovery, schema viewing, record browsing, and filtering.
- Controlled write operations: staged insert/edit/delete workflows with explicit save confirmation.

The product interaction model is panel-based and strongly aligned with vim-like navigation patterns.

## 2. Positioning and Value Proposition

### Positioning

DBC positions itself as a productivity-focused database commander for terminal users who prioritize speed, keyboard flow, and safety.

### Core Value Proposition

- Faster table and data inspection than context-switching into GUI tools.
- Predictable keyboard workflow inspired by familiar terminal patterns.
- Safer change process through staged edits, explicit confirmation, and save-all-or-save-nothing behavior.

## 3. Target Users and Core Jobs

### Primary User Segments

- Developers: inspect schema and records during development/debugging workflows.
- DevOps/SRE practitioners: validate operational data quickly in terminal contexts.
- Technical analysts comfortable with terminal tools: browse and filter records efficiently.

### Core Jobs-to-be-Done

- Select a known database quickly from a configured list.
- Discover tables and understand schema with minimal navigation overhead.
- Browse records in large tables without manual pagination setup.
- Apply a focused filter to inspect a relevant subset of records.
- Stage data corrections safely before committing changes.
- Undo mistakes during a working session before writing to the database.

## 4. Current Product Scope

### In Scope (Current Product State)

- SQLite database support.
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

### Out of Scope (Current Product State)

- Non-SQLite database engines.
- Schema-altering operations (create/alter/drop tables, indexes, views, triggers).
- SQL console/REPL execution.
- Bulk import/export workflows.
- User and permission management.
- Password manager integration.

## 5. Product Experience Principles

- Keyboard-first by default: all primary actions are accessible from keyboard.
- Fast orientation: panel layout keeps navigation context visible.
- Safe-by-design editing: data changes are staged before save.
- Explicit commitment: save requires user confirmation.
- Visible state: status line communicates mode, view, selected table, filter, and key actions.
- Consistent interaction language: vim-like motions and commands are reused across key contexts.

## 6. End-to-End User Journey

### Step 1: Startup and Database Selection

- User launches DBC and sees a centered database selector.
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

## 7. Functional Specification (Current State)

### 7.1 Database Configuration and Access

- DBC reads database entries from OS-specific default config paths:
  - macOS and Linux: `~/.config/dbc/config.toml`
  - Windows: `%APPDATA%\dbc\config.toml`
- Configuration requires at least one database entry.
- Empty config state (`missing file`, `empty file`, or `databases = []`) is treated as "no configured databases", so DBC opens mandatory first-entry setup before normal browsing can start.
- Malformed config state (for example invalid TOML or invalid entry structure) stops startup with an explicit error.
- During mandatory setup, users add one required entry and can optionally add more entries before continuing.
- During mandatory setup, `Esc` cancels startup and exits the application.
- Each entry requires:
  - `name` (display name).
  - `db_path` (SQLite connection path/string).
- Startup selector displays entries in configuration order.
- Startup selector supports in-app add/edit/delete management:
  - Add and edit require non-empty `name` and `db_path`.
  - Add and edit execute connection validation before config save:
    - `db_path` must point to an existing SQLite database file.
    - if connection validation fails, entry is not saved and user stays in form with an error message.
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

### 7.2 Main Layout and Focus Model

- Layout is permanently two-panel in current state.
- Left panel: tables.
- Right panel: schema or records.
- Focus can be moved between panels.
- Active panel is visually indicated.

### 7.3 Table Discovery and Schema View

- Table list excludes internal SQLite system tables.
- Table list is alphabetically sorted for predictable scanning.
- Schema view displays column name and type for selected table.
- If no schema is available yet, product shows an empty-state message.

### 7.4 Records View and Navigation

- Records view shows table data for the selected table.
- Records are fetched and displayed incrementally for continuous browsing.
- Row selection is visible in the focused records panel.
- Field focus mode is supported for cell-level editing navigation.
- Record cell content is width-constrained in the UI (truncated when needed).

### 7.5 Filtering

- User can open a guided filter popup from main browsing contexts.
- Filter flow is step-based:
  - Select column.
  - Select operator.
  - Input value (only if operator requires one).
- Supported operators:
  - `=`, `!=`, `<`, `<=`, `>`, `>=`, `LIKE`, `IS NULL`, `IS NOT NULL`.
- Exactly one filter can be active at a time for the selected table.
- Filter is reset when switching to a different table.

### 7.6 Data Operations (Insert, Edit, Delete)

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
- Pending insert row behavior differs:
  - `d` removes pending insert immediately (instead of marking delete).

### 7.7 Staging, Undo/Redo, and Save

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

### 7.8 Visual State Communication

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

## 8. Keyboard Interaction Model

### 8.1 Global/Main Navigation

| Action | Shortcut |
| --- | --- |
| Quit | `q`, `Ctrl+c` |
| Move down/up | `j`, `k` |
| Move left/right (field focus in records) | `h`, `l` |
| Jump to top | `gg` |
| Jump to bottom | `G` |
| Page down/up | `Ctrl+f`, `Ctrl+b` |
| Switch panel focus | `Ctrl+w h`, `Ctrl+w l`, `Ctrl+w w` |

### 8.2 View and Record Actions

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

### 8.3 Popup Interactions

| Context | Key Behavior |
| --- | --- |
| Filter popup | `j/k` selection, `Enter` confirm step, `Esc` close |
| Edit popup | `Enter` confirm, `Esc` cancel, `Ctrl+n` set `NULL` (nullable fields) |
| Confirm popup (binary) | `Enter` or `y` confirm, `Esc` or `n` cancel |
| Dirty `:config` decision popup | `j/k` choose action, `Enter` or `y` select, `Esc` or `n` cancel |
| Command entry | `Enter` execute command, `Esc` cancel command |

### 8.4 Startup Selector

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

## 9. Data Safety and Change Governance

- Safe default state: product starts in read-only mode.
- Write actions are staged and reversible before save.
- Save requires explicit confirmation.
- Save applies as one unit for the current table.
- On execution failure, product preserves staged intent for user correction.
- Dirty-state visibility is always present in status line.
- Table switch with unsaved staged changes is guarded by discard confirmation.
- `:config` navigation with unsaved staged changes is guarded by explicit save/discard/cancel decision.

## 10. Known Constraints and Non-Goals

This section is the canonical product wording for user-visible constraints and non-goals.

### Current Constraints

- Only SQLite is supported.
- Editing/deleting existing records requires a primary key in the table.
- Only one active filter is supported at a time.
- There is no direct shortcut to switch from Records view back to Schema view after entering Records view.
- No dedicated command exists to clear filter directly (filter resets on table switch or is replaced by applying a new filter).
- Quit action does not prompt to preserve unsaved staged changes.

### Explicit Non-Goals in Current State

- Schema management and administrative operations.
- Multi-engine runtime usage.
- Advanced analytics/reporting and BI workflows.

## 11. Glossary

- Database Entry: named configuration item that points to a SQLite database path.
- Table: primary unit of navigation and data operations.
- Column: typed field in a table schema.
- Record: single row of table data.
- Schema View: right-panel mode showing table columns and types.
- Records View: right-panel mode showing table data rows.
- Filter: active condition applied to records of selected table.
- Staged Change: pending insert/edit/delete not yet saved to the database.
- Dirty State: staged changes exist, and write mode with an unsaved changes count is active.

## 12. Cross-References to Technical Documentation

- Runtime lifecycle implementation: `docs/technical-documentation.md#5-runtime-flow`
- Layer boundaries and dependency direction: `docs/technical-documentation.md#4-architecture-guidelines`
- Save transaction and persistence internals: `docs/technical-documentation.md#53-write-flow`
- Engine and adapter decisions: `docs/technical-documentation.md#6-technical-decisions`
- Test-layer coverage and technical constraints: `docs/technical-documentation.md#8-testing-strategy-and-coverage`, `docs/technical-documentation.md#10-common-technical-constraints`
