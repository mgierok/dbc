# Database Commander (dbc)

Database Commander is a terminal-first application for browsing and managing database data with a keyboard-driven workflow.

## Why dbc?
- Stay in CLI while inspecting schema and records.
- Use vim-like navigation for fast movement across tables and rows.
- Stage writes before save for a safer edit workflow.

## Supported databases
- SQLite (only supported engine).

## Installation
Prerequisites:
- Go `1.25.0` or newer.

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
- `Ctrl+f` / `Ctrl+b`: page down/up
- `Ctrl+w h`, `Ctrl+w l`, `Ctrl+w w`: switch panel focus
- `q`, `Ctrl+c`: quit

Records and staged changes:
- `Enter`: open records view / enter field focus / open edit popup (context dependent)
- `Esc`: exit field focus or close popup
- `F`: open filter popup
- `i`: stage insert
- `d`: toggle delete marker or remove pending insert
- `u`: undo staged action
- `Ctrl+r`: redo staged action
- `w`: save staged changes
- `Ctrl+a`: toggle auto-increment fields for pending insert row

Command mode:
- `:`: open command entry
- `:config`: return to startup selector/config management from active session
- `Enter`: execute command
- `Esc`: cancel command entry

## License
Licensed under Apache License 2.0. See `LICENSE`.
