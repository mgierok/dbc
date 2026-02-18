#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" -ne 1 ]]; then
  echo "usage: bash scripts/cleanup-temp-environment.sh <TMP_ROOT>" >&2
  exit 1
fi

TMP_ROOT="$1"
if [[ -z "$TMP_ROOT" || "$TMP_ROOT" == "/" ]]; then
  echo "error: invalid TMP_ROOT value: $TMP_ROOT" >&2
  exit 1
fi

rm -rf "$TMP_ROOT"
