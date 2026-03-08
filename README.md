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
- Terminal with UTF-8 Unicode rendering support for box-drawing glyphs and standard ANSI SGR text attributes used by the TUI. Set `NO_COLOR` to force monochrome output.

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

## License
Licensed under Apache License 2.0. See `LICENSE`.
