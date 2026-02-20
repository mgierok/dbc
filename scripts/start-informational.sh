#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" -ne 1 ]]; then
  echo "usage: bash scripts/start-informational.sh <help|version>" >&2
  exit 1
fi

MODE="$1"
FLAG=""
case "$MODE" in
  help)
    FLAG="--help"
    ;;
  version)
    FLAG="--version"
    ;;
  *)
    echo "error: unsupported mode '$MODE' (expected: help or version)" >&2
    exit 1
    ;;
esac

TMP_ROOT="$(mktemp -d)"
TMP_HOME="$TMP_ROOT/home"
DBC_BIN="$TMP_ROOT/dbc"

mkdir -p "$TMP_HOME/.config/dbc"
go build -o "$DBC_BIN" ./cmd/dbc

echo "TMP_ROOT=$TMP_ROOT"
echo "cleanup: bash scripts/cleanup-temp-environment.sh \"$TMP_ROOT\""

HOME="$TMP_HOME" "$DBC_BIN" "$FLAG"
