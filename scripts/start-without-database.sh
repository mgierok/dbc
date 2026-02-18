#!/usr/bin/env bash
set -euo pipefail

TMP_ROOT="$(mktemp -d)"
TMP_HOME="$TMP_ROOT/home"
TMP_DB="$TMP_ROOT/test.db"
DBC_BIN="$TMP_ROOT/dbc"

mkdir -p "$TMP_HOME/.config/dbc"
cp scripts/test.db "$TMP_DB"
go build -o "$DBC_BIN" ./cmd/dbc

echo "TMP_ROOT=$TMP_ROOT"
echo "TMP_DB=$TMP_DB"
echo "cleanup: bash scripts/cleanup-temp-environment.sh \"$TMP_ROOT\""

HOME="$TMP_HOME" "$DBC_BIN"
