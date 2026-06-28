#!/bin/bash
# Denies reading .env files (.env, .env.local, .env.production, etc).
# Wired to the beforeReadFile / beforeTabFileRead hook events.
set -euo pipefail

input=$(cat)

# The path field name can vary by event; try the known candidates.
path=$(printf '%s' "$input" | jq -r '.file_path // .path // .filePath // .arguments.path // empty')

if [[ -n "$path" ]]; then
  base=$(basename "$path")
  if [[ "$base" =~ ^\.env(\..*)?$ ]]; then
    echo '{
      "permission": "deny",
      "user_message": "Reading .env files is blocked by a project hook.",
      "agent_message": "Access to .env files is denied by project policy. Do not read environment files; ask the user for any required values instead."
    }'
    exit 0
  fi
fi

echo '{ "permission": "allow" }'
exit 0
