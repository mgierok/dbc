# dbc
Database Commander — a terminal-based application for browsing and managing databases.

## Setup
1) Copy the example config:
```
mkdir -p ~/.config/dbc
cp docs/config.example.toml ~/.config/dbc/config.toml
```
2) Edit `~/.config/dbc/config.toml` and set `db_path` to your SQLite file.

## Run
```
go run ./cmd/dbc
```
