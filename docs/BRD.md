# DBC - Business Requirements Document (BRD/PRD)

Status: draft
Version: 0.1
Owner: DBC Team
Date: YYYY-MM-DD

## 1. Purpose
This document defines the business scope, goals, and functional roadmap for the DBC application. It is a living document and will evolve over time.

## 2. Problem Statement
Users working with databases need a fast, terminal-first interface to browse and manage data. Existing tools are either GUI-only or too low-level for quick inspection and operations. DBC aims to combine the ergonomics of Midnight Commander with core vim-like keyboard navigation.

## 3. Product Vision
DBC is a terminal-based application for browsing and managing databases, optimized for fast, keyboard-driven workflows without leaving the CLI.
Tagline: Database Commander: inspect data at the speed of your terminal.

## 4. Business Goals
- Reduce time to inspect database structure and data.
- Lower the barrier to working with databases without GUI tools.
- Provide a consistent UX across multiple database engines over time.

## 5. Target Users and Needs
- Developers: quick data inspection and schema discovery.
- DevOps/SRE: validation of data and structure without GUI.
- Technical analysts: fast filtering and browsing of records.

## 6. Scope
### In Scope (MVP)
- Read-only browsing of SQLite databases.
- Terminal UI inspired by Midnight Commander.
- Basic vim-like keyboard shortcuts.

### Out of Scope (for now)
- Full user management for server-based engines.
- Advanced analytics or reporting.

## 7. Domain Glossary
- Database: a named config entry pointing to a SQLite file.
- Table: a collection of records; primary browsing unit.
- Column: a single field in a table with a data type.
- Record: a single row of data in a table.
- Schema: the database structure; tables and their columns.
- Panel: a UI region (e.g., table list left, data view right).
- Focus: the active panel receiving keyboard input.
- Read-only mode: a mode that guarantees no data changes.

## 8. Product Principles (UX)
- Panel-based layout and navigation similar to Midnight Commander.
- Keyboard-first interaction with familiar vim-like shortcuts.
- Safe default mode: read-only data access.
- Two-panel layout by default; single-panel views allowed when focused on one context (e.g., schema-only).
- Panel switching uses vim-style shortcuts.
- Read-only status is always visible in the status bar.
- Data browsing uses infinite scroll with vim-style motion keys, including page scrolling and jump-to-top/bottom.

## 9. Keyboard Shortcuts (Initial, vim-based)
```
h       move left
j       move down
k       move up
l       move right

gg      jump to the top
G       jump to the bottom

Ctrl+f  page down by one screen
Ctrl+b  page up by one screen

Enter   show records in the right panel
F       open filter popup for the current table

q       quit application
Ctrl+c  quit application
Esc     close filter popup

Ctrl+w h  focus left panel
Ctrl+w l  focus right panel
Ctrl+w w  cycle panel focus
```
Startup selector uses the same navigation keys; Enter confirms the database choice and Esc exits.

## 10. Business Requirements (non-implementation)
- Extensibility from day one to support additional database engines.
- Data safety: minimize accidental changes to data.
- Portability: cross-platform support (Go).

## 11. Success Metrics (examples)
- Time to locate a table and view data under 60 seconds for a new database.
- Low learning curve due to familiar shortcuts.
- Smooth browsing of large SQLite files without noticeable lag.

## 12. Risks and Assumptions
- Different database engines require a consistent interaction model.
- Users expect predictable keyboard behavior.
- Read-only mode must be safe even for large datasets.

## 13. Functional Roadmap (Checklist)
These checklists are part of the business documentation and can be updated in future iterations.

### Stage 0: Product Foundations
- [x] Naming, positioning, and core product narrative
- [x] Domain glossary and shared terminology
- [x] UX principles (layout, navigation, shortcuts)

### Stage 1: SQLite Browsing (MVP)
- [x] Startup database selector (from config list)
- [x] Table list and schema view (schema shown in right panel)
- [x] Record preview with infinite scroll (records shown in right panel)
- [x] Column filter popup with operator selection
- [x] Read-only mode by default
- [x] Fast keyboard navigation (vim-like)

#### Stage 1 Definition (Detailed)
- Scope: SQLite only, with architecture prepared for future engine adapters.
- Configuration: load databases from `~/.config/dbc/config.toml` using `[[databases]]` entries with `name` and `db_path`.
- Layout: left panel lists tables; right panel switches between schema and records for the selected table.
- Startup: show a centered database selector listing `Name | conn_string`; Enter selects and Esc exits.
- Default view: schema in the right panel after database selection.
- Schema view: shows the selected table's columns and data types in the right panel.
- Records view: read-only table data with infinite scroll for large datasets; values are truncated to panel width (no horizontal scroll in Stage 1).
- Table list: sorted alphabetically.
- Filtering: `F` opens a popup; step 1 selects a column; step 2 selects an operator; step 3 enters a value when required.
- Filtering: one active filter at a time; filters reset on table change.
- Operators: provided by the engine adapter; SQLite supports `=`, `!=`, `<`, `<=`, `>`, `>=`, `LIKE`, `IS NULL`, `IS NOT NULL`.
- Navigation: vim-style motion keys (`h/j/k/l`), page scrolling (`Ctrl+f/Ctrl+b`), and jump-to-top/bottom (`gg/G`).
- Panel focus: vim-style switching between panels (`Ctrl+w h`, `Ctrl+w l`, `Ctrl+w w`).
- Mode indicator: read-only status is always visible in the status bar.
- Status bar: shows contextual action shortcuts for the active panel (non-navigation).
- Exit: `q` or `Ctrl+C`.
- Exclusions: no data writes, no schema changes, no SQL REPL.

### Stage 2: SQLite Data Operations
- [ ] Insert, edit, delete records
- [ ] Transaction management for safe writes
- [ ] Session-level undo/redo

### Stage 3: SQLite Schema Management
- [ ] Create and modify tables
- [ ] Indexes, views, triggers
- [ ] Export/import (e.g., CSV, SQL)

### Stage 4: Server Engines (first integrations)
- [ ] Support for the first server engine (TBD)
- [ ] Connection and context management
- [ ] Read-only mode as safe default

### Stage 5: Multi-engine and Administration
- [ ] Additional engines (MySQL, MSSQL, others)
- [ ] User and permissions management (server engines)
- [ ] Full database management: backup, restore, migrations

## 14. Open Questions
- What criteria will drive selection of the next server engine after SQLite?
- Which administrative operations are critical for early iterations?
- Is integration with password managers required?
