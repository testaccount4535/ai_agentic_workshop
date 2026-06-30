#!/usr/bin/env bash
# Summarize a bbolt ride-hailing database: info, integrity, stats, buckets, and
# every record dumped as JSON.
# Usage: inspect.sh [path-to-db]   (default: rides.db)
set -euo pipefail

DB="${1:-rides.db}"

if ! command -v bbolt >/dev/null 2>&1; then
  echo "bbolt CLI not found. Install with: go install go.etcd.io/bbolt/cmd/bbolt@latest" >&2
  exit 1
fi

if [[ ! -f "$DB" ]]; then
  echo "database not found: $DB" >&2
  exit 1
fi

pretty() { if command -v jq >/dev/null 2>&1; then jq .; else cat; fi; }

echo "# Database: $DB"
echo
echo "## Info"
bbolt info "$DB"
echo
echo "## Integrity"
bbolt check "$DB"
echo
echo "## Stats"
bbolt stats "$DB"
echo
echo "## Buckets and records"
bbolt buckets "$DB" | while IFS= read -r bucket; do
  [[ -z "$bucket" ]] && continue
  echo
  echo "### $bucket"
  keys=$(bbolt keys "$DB" "$bucket")
  if [[ -z "$keys" ]]; then
    echo "(empty)"
    continue
  fi
  while IFS= read -r key; do
    [[ -z "$key" ]] && continue
    echo "- $key:"
    bbolt get "$DB" "$bucket" "$key" | pretty
  done <<< "$keys"
done
