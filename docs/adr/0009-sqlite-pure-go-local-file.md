# ADR-0009: SQLite via modernc.org/sqlite, local database file

**Status:** Accepted — 2026-08-25

## Context

Persistence for a single-user local TUI. Options: JSON/YAML file, embedded key-value store, or
SQLite. Reporting needs aggregation (sums grouped by period, designs ranked by profit) and derived
spool remaining needs joins across usage rows.

## Decision

SQLite through `modernc.org/sqlite` — a pure-Go implementation, so no cgo and no C toolchain for
cross-compilation.

Database path defaults to `./printer-ledger.db`, overridable with a `-db` flag. `.gitignore` already
excludes `*.db`, `*.db-shm`, `*.db-wal`.

## Consequences

- Aggregation and joins are the database's job, not hand-rolled Go loops over a decoded JSON blob.
- Transactions give atomic multi-row writes (a print plus its usage rows plus its units).
- `CGO_ENABLED=0` builds work; single static binary.
- `modernc.org/sqlite` is slower than the cgo `mattn/go-sqlite3`. Irrelevant at this data volume.
- Schema migrations become a real concern as the app evolves. Handled with a simple ordered
  migration list in `internal/store`.
