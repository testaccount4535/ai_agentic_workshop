#!/bin/bash
# Runs `go fix` after a Go source file is edited by the agent or Tab.
# Wired to the afterFileEdit hook event. Fails open: a go fix problem never
# blocks the edit itself.
set -uo pipefail

input=$(cat)

path=$(printf '%s' "$input" | jq -r '.file_path // .path // .filePath // empty')

# Only act on Go source files.
case "$path" in
  *.go) ;;
  *) exit 0 ;;
esac

# Project hooks run from the repo root, so operate on the module.
if command -v go >/dev/null 2>&1; then
  go fix ./... >/dev/null 2>&1 || true
fi

exit 0
