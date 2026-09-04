# ADR-0009: SQLite via modernc.org/sqlite, local database file

**Status:** Accepted — 2026-08-25. Amended 2026-09-04: the default path moved out of the working
directory.

## Context

Persistence for a single-user local TUI. Options: JSON/YAML file, embedded key-value store, or
SQLite. Reporting needs aggregation (sums grouped by period, designs ranked by profit) and derived
spool remaining needs joins across usage rows.

The database first defaulted to `./printer-ledger.db`. That is correct while the binary is run from
the repository with `go run`, and wrong the moment it is installed: with the binary on `PATH`, the
ledger resolves against whatever directory the operator happened to be standing in. Running the same
command from two places silently opens two different ledgers, and a fresh empty one looks identical
to lost data. Options for a stable default:

- `os.UserConfigDir()`, which on darwin is `~/Library/Application Support/printer-ledger`. Native,
  but the path contains a space, which is hostile to a `-db` argument or a backup script.
- `~/.printer-ledger/`. One rule everywhere, no environment variable to read.
- The XDG Base Directory layout: `$XDG_DATA_HOME`, falling back to `~/.local/share`.

Keeping the working directory as a *fallback* — use `./printer-ledger.db` when it exists, otherwise
the stable default — was considered and rejected: it reproduces the same bug more quietly, because a
stray database file anywhere on disk would capture the next run.

## Decision

SQLite through `modernc.org/sqlite` — a pure-Go implementation, so no cgo and no C toolchain for
cross-compilation.

The database defaults to `$XDG_DATA_HOME/printer-ledger/ledger.db`, falling back to
`~/.local/share/printer-ledger/ledger.db`. A non-absolute `$XDG_DATA_HOME` is ignored, per the XDG
spec, so a relative value cannot reintroduce a working-directory-relative ledger.

`store.DefaultPath()` answers where the ledger defaults to; `main.go` owns precedence:

1. the `-db` flag — a relative value resolves against the working directory, because the operator
   named that file explicitly. A leading `~` is not expanded: that is the shell's job, and the open
   fails loudly if the shell did not do it.
2. `PRINTER_LEDGER_DB`
3. `store.DefaultPath()`

`store.Open` creates the parent directory (0700) if it is absent, and the run that creates the
database file prints one line to stderr naming it. The Settings tab shows the ledger file as
read-only text — it is not a **Settings** value, so it reaches the tab from whatever opened the
ledger rather than from a row.

There is no automatic migration of a database left at the old default. The pre-existing
`./printer-ledger.db` was removed by hand.

`.gitignore` excludes `*.db`, `*.db-shm`, `*.db-wal`.

## Consequences

- Aggregation and joins are the database's job, not hand-rolled Go loops over a decoded JSON blob.
- Transactions give atomic multi-row writes (a print plus its usage rows plus its units).
- `CGO_ENABLED=0` builds work; single static binary.
- `modernc.org/sqlite` is slower than the cgo `mattn/go-sqlite3`. Irrelevant at this data volume.
- Schema migrations become a real concern as the app evolves. Handled with a simple ordered
  migration list in `internal/store`.
- The installed binary opens one ledger from every working directory, at a path with no spaces, so
  `-db` arguments and backup scripts stay quotable.
- `$XDG_DATA_HOME` gives a one-variable way to point a whole shell at a scratch ledger, and
  `PRINTER_LEDGER_DB` a way to point at a single file.
- The default is environment-dependent, so `DefaultPath()` is a function that can fail, and it gets
  a test in `internal/store` driven with `t.Setenv` — a second, deliberately narrow test seam beside
  `internal/app`. `make test` and `make lint` widened to cover it.
