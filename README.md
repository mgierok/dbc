# dbc
Database Commander — a terminal-based application for browsing and managing databases.

## Setup
1) Copy the example config:
```
mkdir -p ~/.config/dbc
cp docs/config.example.toml ~/.config/dbc/config.toml
```
2) Edit `~/.config/dbc/config.toml` and add `[[databases]]` entries with `name` and `db_path` (connection string).

## Run
```
go run ./cmd/dbc
```

## Keyboard Shortcuts (Records View)
- `Enter` edit selected cell (or enter field focus)
- `i` insert pending row
- `d` toggle delete on persisted row / remove pending insert
- `u` undo last staged action
- `Ctrl+r` redo staged action
- `w` save staged changes
- `F` open filter popup
