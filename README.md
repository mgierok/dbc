# Database Commander (dbc)

Database Commander is a terminal-first application for browsing and managing database data with a keyboard-driven workflow.

## Why dbc?
- Stay in CLI while inspecting schema and records.
- Use vim-like navigation for fast movement across tables and rows.
- Stage writes before save for a safer edit workflow.
- Keep navigation context clear with independent boxed panels and a framed status bar.

## Supported databases
- SQLite (only supported engine).

## Installation
Prerequisites:
- Go `1.25.0` or newer.
- Supported operating systems: macOS and Linux.
- Terminal with UTF-8 Unicode rendering support for box-drawing glyphs used by the TUI.

Clone repository and install:
```bash
git clone https://github.com/mgierok/dbc.git
cd dbc
go install ./cmd/dbc
```

## Usage
Run selector-first startup:
```bash
dbc
```
- Opens the startup selector.

Run direct launch:
```bash
dbc -d /absolute/path/to/database.sqlite
```

## Keyboard Controls and Commands
Global/main navigation:
- `j` / `k`: move down/up
- `h` / `l`: move left/right (field focus in records)
- `gg` / `G`: jump top/bottom
- `Ctrl+f` / `Ctrl+b`: page down/up (`Records`: next/previous persisted-record page)
- `?`: open context help popup for active runtime panel/state or current selector mode

Records and staged changes:
- `Enter`: open records view (from tables) / open selected record detail (in records)
- `e`: enter field focus / open edit popup (in records)
- `Esc`: return to tables (right panel neutral) / exit field focus / close popup
- `Shift+F`: open filter popup (records view)
- `Shift+S`: open sort popup (single column + `ASC`/`DESC`)
- `i`: stage insert
- `d`: toggle delete marker or remove pending insert
- `u`: undo staged action
- `Ctrl+r`: redo staged action
- `w`: save staged changes
- `Ctrl+a`: toggle auto-increment fields for pending insert row
- records header marks active sort column with `↑` (`ASC`) or `↓` (`DESC`)
- status bar in `Records` shows persisted-record summary and pagination (`Records: current/total | Page: current/total`)

Command mode:
- `:`: open command entry
- `:config` / `:c`: return to startup selector/config management from active session
- `:help` / `:h`: open runtime context help popup (command alias)
- `:quit` / `:q`: quit application from active runtime session
- `Enter`: execute command
- `Esc`: cancel command entry

Startup selector:
- selector main content shows options/status and keeps right-aligned `Context help: ?`
- selector main content does not show inline shortcuts or legend rows
- `?`: open selector context-help popup (mode-specific shortcuts for browse/add-edit/delete-confirm)
- `Esc`: close selector context-help popup

## License
Licensed under Apache License 2.0. See `LICENSE`.
