---
name: inspect-db
description: Inspect the ride-hailing bbolt database (rides.db) with the bbolt command-line tool — list buckets, count and dump ride_starts/ride_ends records as JSON, show stats, and verify integrity. Use when asked to inspect, debug, or show the contents of rides.db or the bbolt store.
disable-model-invocation: true
---

# Inspect rides.db (bbolt)

`rides.db` is a bbolt key/value database written by this project's `internal/store`.
Records are JSON values keyed by ride ID, in two buckets:

- `ride_starts` — `model.RideStart` records
- `ride_ends` — `model.RideEnd` records

## Prerequisites

The `bbolt` CLI must be installed:

```bash
command -v bbolt || go install go.etcd.io/bbolt/cmd/bbolt@latest
```

Stop the server first: bbolt takes an exclusive file lock, so these commands block
while the application holds the database open.

## Quick summary

Run the helper to print a full report (info, integrity, stats, and every record as JSON):

```bash
.cursor/skills/inspect-db/scripts/inspect.sh rides.db
```

## Manual commands

Every command takes the database path as the first argument.

```bash
bbolt info rides.db                        # page size / basic info
bbolt check rides.db                       # verify integrity (prints OK)
bbolt buckets rides.db                     # list buckets
bbolt stats rides.db                       # aggregate key/page statistics
bbolt inspect rides.db                     # bucket/key tree structure
bbolt keys rides.db ride_starts            # list ride IDs in a bucket
bbolt get  rides.db ride_starts <ride-id>  # print one record (JSON)
```

Values are JSON — pipe through `jq` for readability:

```bash
bbolt get rides.db ride_starts ride-1 | jq .
```

## Notes

- An empty bucket prints nothing for `keys`; that is normal.
- A bucket may be absent in older database files (e.g. `ride_ends` was added later);
  `bbolt buckets` shows what actually exists.
- Never run `bbolt surgery` against a live database. Operate on a copy:
  `bbolt compact -o copy.db rides.db`.
