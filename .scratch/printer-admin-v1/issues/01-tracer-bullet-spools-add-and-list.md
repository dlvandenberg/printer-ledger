# 01: Tracer bullet — skeleton, Spools tab, add and list

**What to build:** The operator launches the binary, lands on a six-tab shell, adds a Spool with its
Filament Type, brand, colour, initial grams, tare grams, purchase price and purchase date, and sees
it in a list showing remaining grams and remaining value. Pointing the binary at a different
database file with `-db` gives a separate, empty ledger.

This is the first code in the repo. It sets the conventions every later ticket copies: the shape of
the application-layer use cases, the ordered migration list, and the test harness that drives those
use cases against a real SQLite database in a temp directory. Be fussy here.

One design question to settle deliberately and record in the ticket comments once decided: do
application-layer use cases return domain aggregates or their own result types? It leaks into every
screen and every test, so pick it now rather than at call site thirty.

Remaining is derived, not stored, per ADR-0003. With no prints and no adjustments yet it equals
initial grams — build the derivation properly anyway, because tickets 03 and 06 extend the same
expression.

The other five tabs are placeholders. `CGO_ENABLED=0` must produce a working static binary.

**Blocked by:** None (can start immediately).

**Status:** resolved

- [x] A Go module with a `main` that opens or creates the database and starts the TUI
- [x] `-db` flag overrides the default database path
- [x] Ordered migration list; opening a fresh database succeeds and reopening an existing one is idempotent
- [x] Six tabs render, `tab` / `shift-tab` switch between them, five are placeholders
- [x] Add a Spool from the Spools tab with `a`; validation errors render inline beside the offending field and typed values survive a rejected submit
- [x] Spool list shows remaining grams and remaining value in money
- [x] Money is `int64` minor units end to end, per ADR-0002
- [x] Domain layer has no dependency on the store or the TUI
- [x] Tests drive the application layer against a real SQLite database in a temp directory
- [x] `CGO_ENABLED=0 go build` produces a single static binary
- [x] ADR-0001 amended to record the application layer as a fourth layer and the test seam

## Comments

### Decision: use cases return their own result types, not domain aggregates

Settled as the ticket asked, because it leaks into every screen and every test.

`internal/app` returns view types — `app.SpoolView`, not `domain.Spool`. Reasons:

- The Spools list must show remaining grams and remaining value, both derived. Returning a
  `domain.Spool` would leave the TUI holding a spool and a derivation to run, which ADR-0001 forbids:
  any number the UI shows is returned by a use case. `SpoolView` arrives with the numbers already in
  it.
- The test seam asserts on returned values. A view type is exactly the set of facts a screen is
  promised, so a test that pins `RemainingValue == 2200` is pinning a contract rather than a struct
  layout.
- Aggregates stay free to change shape. `Spool` gained `SpoolLedger` around it without the TUI or any
  test noticing, and the same will hold when Print grows its rate snapshot.

Cost: one mapping function per aggregate (`toSpoolView`), and view types will duplicate aggregate
fields. Accepted — the duplication is mechanical and the alternative leaks derivation into the
renderer.

Commands mirror it on the way in: `app.AddSpoolCmd` rather than a half-built `domain.Spool`.

### Other decisions taken while implementing

- **`domain` is flat**, not a package per aggregate. Spool, Design and Print are mutually
  referential — Capable Spool needs a Design, Filament Usage needs a Spool and a Print — so
  subpackages would hit an import cycle by ticket 05.
- **`SpoolLedger`** carries the two event sums so ADR-0003's expression lives in the domain while the
  store fetches the sums. Both are literal `0` in the SQL until tickets 03 and 06 fill them in.
  Glossary entry added to `CONTEXT.md`.
- **Field-name constants** (`domain.FieldBrand`, …) key validation errors, so the form renders
  `errs.For(field)` beside the right row and the two sides cannot drift.
- **Brand and colour are required**; **purchase cost of 0 is allowed** — a gifted spool genuinely
  costs nothing per gram. Confirmed with the operator.
- **`costPerGram` is never materialised.** It is fractional (€22.00/1000g = 2.2c/g), so
  `Spool.ValueOf` computes `grams × purchaseCost / initialGrams` in `int64` and rounds once.
- **Parsing lives in `domain`** (`ParseCents`, `ParseGrams`, `ParseDate`). Pure, and it keeps the TUI
  from owning input formats. Covered by a small `domain` test because it is unreachable through the
  app seam and the TUI is untested.
- **Currency symbol is a TUI constant.** It stays one: ticket 02 dropped the setting and fixed the
  euro (ADR-0019).
