# printer-ledger

A single-binary Go TUI ledger for a one-person 3D printing side business. Read `CONTEXT.md` for the
domain glossary before naming anything.

## Architecture

Four layers. **Dependencies point inward only.** Rationale and alternatives: `docs/adr/0001`.

| Layer | Rule |
|---|---|
| `internal/domain` | Pure Go. No imports from the other three layers, no SQL, no bubbletea. Holds cost math and every invariant. Declares the repository interfaces it needs. |
| `internal/store` | SQLite via `modernc.org/sqlite`. Implements the domain's repository interfaces. Schema changes go in the ordered migration list. |
| `internal/app` | One method per use case. Owns transaction boundaries. The test seam. |
| `internal/tui` | bubbletea. A renderer. **No cost math, no invariants, no business rules.** |

Binding rules:

- The TUI must be able to do nothing `internal/app` cannot. Any number the UI shows is returned by a
  use case, never computed in the view.
- Invariants live in the domain, so the database cannot hold an invalid state. The two that matter:
  spool `remaining >= 0` and print `available >= 0`.
- Money is `int64` minor units everywhere. Never a float for money, never a tolerance in a money
  assertion. Round once, at the end. `kwhPerHour` stays `float64` — it is a physical rate, not money.
- Durations are integer minutes. Input `HH:MM`, display `5h 31m`.
- A print's cost is frozen at creation — rate snapshot on the print, gram price on each usage row.
  Nothing edited later may change a past print's cost.
- `CGO_ENABLED=0` must always produce a working static binary.

## Testing

One seam: the `internal/app` use-case API, driven against a real SQLite database in `t.TempDir()`.
Standard library `testing`, table-driven, no assertion framework, no mocks.

A test asserts on values a use case returns or on what a later query reports. It never reaches into
a struct field, never asserts on SQL, and never asserts on rendered terminal output. If moving the
cost calculation between packages breaks a test that still produces €4.71, the test is wrong.

## Agent skills

### Issue tracker

Local markdown: issues and specs live under `.scratch/<feature-slug>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

Canonical defaults: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: root `CONTEXT.md` + `docs/adr/`. See `docs/agents/domain.md`.
