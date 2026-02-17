# Database Commander (dbc)

## dbc
Database Commander is a terminal-first application for browsing and managing database data with a keyboard-driven workflow.

## Why dbc?
- Stay in CLI while inspecting schema and records.
- Use vim-like navigation for fast movement across tables and rows.
- Stage writes before save for a safer edit workflow.

## Supported databases
- SQLite (current and only supported engine in current state).

## Installation
Prerequisites:
- Go `1.25.0` or newer.

Option 1: install from GitHub with `go install`:
```bash
go install github.com/mgierok/dbc/cmd/dbc@latest
```
After install, ensure your Go binary directory is in `PATH`:
```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Option 2: clone repository and install:
```bash
git clone https://github.com/mgierok/dbc.git
cd dbc
go install ./cmd/dbc
```

## Configuration
Default config path:
- macOS/Linux: `~/.config/dbc/config.toml`
- Windows: `%APPDATA%\dbc\config.toml`

Create config from example:
```bash
mkdir -p ~/.config/dbc
cp docs/config.example.toml ~/.config/dbc/config.toml
```

Minimal config structure:
```toml
[[databases]]
name = "local"
db_path = "/absolute/path/to/database.sqlite"
```

## Usage
Run selector-first startup:
```bash
dbc
```

Run direct launch (skip selector when validation succeeds):
```bash
dbc -d /absolute/path/to/database.sqlite
dbc --database /absolute/path/to/database.sqlite
```

Open selector again inside active session:
- Open command entry with `:`, then run `:config`.

## Keybindings
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

## License
Licensed under Apache License 2.0. See `LICENSE`.
