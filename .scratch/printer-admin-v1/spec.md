# Spec: printer-admin v1

Status: resolved

## Problem Statement

I sell 3D prints locally as a side business, and I have no idea whether it makes money.

Filament sits on shelves in half-used spools and I cannot tell how much is left on any of them
without guessing, so I over-order or start a print that runs out halfway. When someone asks what a
print costs, I make a number up: I know roughly what a spool cost and roughly how long the printer
ran, but not what _this_ object cost me once electricity and machine wear are counted. So I price by
feel, and I suspect some items I sell at a loss.

I also have no idea whether the printer has paid for itself, which designs are actually worth
printing, or how much of my "revenue" was really just filament I already bought.

## Solution

A single-binary Go terminal app — a workshop ledger for one operator and one printer (Bambu Lab A1).

I record what I buy (**Spools**), what I can print (**Designs**), what I actually printed
(**Prints**, which consume filament from specific Spools and produce a counted batch of copies), and
what I sold (**Sales**, each against one copy from one Print).

Pricing lives on the **Design**. A Design holds the slicer's estimates for _one copy_ and shows what
that copy would cost in each **Filament Type** I stock, plus a suggested price with my margin on
top. A **Print** shows what the job really cost and what each copy really cost — no second price.

From that the app tells me:

- how many grams each Spool has left, corrected by putting it on a kitchen scale
- what to charge for a design before I have ever printed it, costed against the most expensive
  spool that could actually do the job, so the quote is never optimistic
- what a Print really cost — filament at the gram price frozen when I recorded it, electricity from
  a measured per-material power draw, and machine overhead per hour
- when a sale price is about to fall below cost plus a small margin
- how the business is doing: revenue, cost of what I sold, profit, copies sold, best designs, and how
  far I am from breaking even on everything I have spent

Six tabs: `Spools │ Designs │ Prints │ Sales │ Report │ Settings`. No mouse, no network, no cloud.
One SQLite file I can back up.

## User Stories

### Settings

1. As the operator, I want to enter my electricity price per kWh, so that print costs reflect what I actually pay for power.
2. As the operator, I want a kWh-per-print-hour figure for each **Filament Type**, seeded with a plausible default at first launch, so that energy is never silently costed at zero before I have measured anything.
3. As the operator, I want to replace a seeded rate with my own smart-plug measurement and see it marked as measured, so that I can tell a real figure from a placeholder.
4. As the operator, I want to set a machine hourly rate covering depreciation, maintenance and a failure buffer, so that machine wear is priced into every print instead of quietly eating my margin.
5. As the operator, I want to record what the printer cost me, so that the app can tell me when it has paid for itself.
6. As the operator, I want to set a default margin percentage, so that new designs start from the profit I intend to make.
7. As the operator, I want to set a minimum margin percentage, so that the app can tell me when a sale price is too low.
8. ~~As the operator, I want to set my display currency, so that amounts read the way I think about money.~~ Dropped: the ledger is euros throughout and nothing converts, so the symbol is a renderer constant (ADR-0019).
9. As the operator, I want sensible defaults on first launch, so that I can start entering spools before I have finished configuring everything.
10. As the operator, I want to reach the Settings tab even when the database is empty, so that I can configure rates before recording any data.

### Spools

11. As the operator, I want to add a spool with its **Filament Type**, brand, colour, filament grams, empty-spool weight and purchase price, so that the app can compute what its filament costs per gram.
12. As the operator, I want each spool costed at its own purchase price, so that a print from a cheap spool is correctly shown as cheaper than the same print from an expensive one.
13. As the operator, I want to see the remaining grams for every spool in one list, so that I can tell at a glance what I can still print and what I need to order.
14. As the operator, I want to see the remaining value in money alongside the remaining grams, so that I know how much stock is sitting on my shelf.
15. As the operator, I want to open a spool and see how its remaining figure was reached — initial grams, grams consumed by prints, and adjustments — so that I trust the number instead of second-guessing it.
16. As the operator, I want to weigh a spool on my kitchen scale and enter the total weight, so that the app corrects its estimate using the spool's tare weight without me doing arithmetic.
17. As the operator, I want to add a note to a re-weigh, so that I can record why it drifted (purge tower, tangle, failed print, partial spool).
18. As the operator, I want a spool to be marked empty when it hits zero grams, so that it drops out of my way when I am picking filament for a print.
19. As the operator, I want empty spools to remain visible in history, so that past prints still explain their costs.
20. As the operator, I want spool remaining never to go negative, so that the number always means something physical.
21. As the operator, I want to correct a spool I entered wrongly, so that a typo in a purchase price does not poison every cost derived from it.
22. As the operator, I want correcting a spool's price to leave already-recorded prints untouched, so that fixing a typo does not rewrite last year's profit.
23. As the operator, I want to be stopped from deleting a spool that prints have drawn from, so that print history cannot lose its cost basis.

### Designs

24. As the operator, I want to add a design with its name and the slicer's filament-grams and print-time estimates **for one copy**, so that the app knows what a single object takes.
25. As the operator, I want to enter print time as `HH:MM`, so that I can copy what Bambu Studio shows me instead of converting to decimal hours.
26. As the operator, I want to set the **Filament Type** I normally print a design in, so that the app knows which cost row to headline.
27. As the operator, I want a design to show an estimated cost per **Filament Type** I stock, so that I can see what printing it in PLA versus PETG would cost me.
28. As the operator, I want each cost row priced against the most expensive spool that could actually complete the job, so that a quote is never optimistic about which spool I will end up using.
29. As the operator, I want each cost row to name the spool whose price it used, so that I understand why the number is what it is.
30. As the operator, I want a design with no spool in stock that could print it to fall back to a labelled reference price, so that I can still quote instead of seeing nothing.
31. As the operator, I want a suggested selling price on the design, so that I can quote someone a price before I have printed the thing.
32. As the operator, I want the suggested price computed from the design's estimates and its own margin, so that the number is stable and does not shift as individual prints come out better or worse.
33. As the operator, I want to edit the margin on a specific design, so that I can price a fiddly or premium item differently from my default.
34. As the operator, I want a new design to start at my default margin without me typing it, so that adding a design is quick.
35. As the operator, I want changing my default margin to leave existing designs alone, so that today's setting change does not silently re-price my whole catalogue.
36. As the operator, I want to edit a design's estimates, so that I can update them after re-slicing.

### Prints

37. As the operator, I want to record a print against a design, so that my print history is tied to what was printed.
38. As the operator, I want the print form prefilled with the design's estimates multiplied by the number of copies, so that recording a normal plate is nearly keystroke-free.
39. As the operator, I want to override grams and time with the totals the slicer or printer screen actually reported, so that cost reflects the real job rather than the estimate.
40. As the operator, I want to record how many copies came off the plate, so that per-copy cost is the job cost divided across them.
41. As the operator, I want to pick which specific spool the filament came from, so that cost uses that spool's real price per gram.
42. As the operator, I want to see each spool's remaining grams while picking, so that I can tell whether it will make it through the job.
43. As the operator, I want to split a print's filament across several spools, so that a mid-print spool swap is recorded as one print rather than two fictional ones.
44. As the operator, I want to be stopped from recording more grams from a spool than it holds, so that I am forced to record the swap that actually happened.
45. As the operator, I want a colour-swap print recorded as a single print, so that its cost and its copies stay together.
46. As the operator, I want to be stopped from mixing filament types on one print, so that the energy rate for the job is never ambiguous.
47. As the operator, I want to be warned when I record a print whose filament type still has a seeded rather than measured power rate, so that I know the energy figure is a placeholder.
48. As the operator, I want to see the computed cost breakdown — filament, energy, overhead — while filling in the form, so that I understand where the number comes from.
49. As the operator, I want to see the per-copy cost on the print, so that I know what the objects in front of me actually cost.
50. As the operator, I want a print's cost frozen at the rates and gram prices in force when I recorded it, so that changing my electricity price or fixing a spool typo next year does not rewrite last year's history.
51. As the operator, I want to backdate a print, so that I can catch up on entries I did not make at the time.
52. As the operator, I want to edit a print I entered wrongly, so that a mistyped gram figure can be fixed and spool remaining recomputed.
53. As the operator, I want to record a run that produced nothing as a spool re-weigh rather than a print, so that I am not inventing copies that never existed.

### Stock

54. As the operator, I want each print to track how many of its copies are still available, so that I know what is in the drawer to sell.
55. As the operator, I want to record how many copies I gave away, so that a gift does not look like unsold stock or lost revenue.
56. As the operator, I want to record how many copies I kept for myself, so that they leave my sellable stock without being logged as gifts.
57. As the operator, I want to record how many copies were ruined, so that a failed copy is not pretended to be sellable.
58. As the operator, I want the cost of gifted, kept and scrapped copies still counted as money spent, so that my profit figure is honest.
59. As the operator, I want gifted, kept and scrapped copies excluded from revenue and copies-sold, so that sales figures mean sales.
60. As the operator, I want the cost value of my unsold copies shown, so that I know how much money is tied up in finished stock.
61. As the operator, I want to be stopped from accounting for more copies than the print produced, so that the ledger cannot contradict itself.
62. As the operator, I want to be stopped from reducing a print's quantity below the copies already sold, gifted, kept or scrapped, so that the same rule protects me when I edit.
63. As the operator, I want to be stopped from deleting a print whose copies have been sold, gifted, kept or scrapped, so that deleting a row cannot erase past revenue.

### Sales

64. As the operator, I want to record a sale by picking the print a copy came from and entering what I charged, so that revenue is tied to a real batch with a real cost.
65. As the operator, I want only prints with copies still available offered for sale, so that I cannot accidentally sell the same object twice.
66. As the operator, I want to see the copy's cost and the design's suggested price while entering the sale, so that I notice immediately if I am about to sell too cheap.
67. As the operator, I want a warning — not a refusal — when the price is below cost plus my minimum margin, so that a haggled sale is still recorded truthfully.
68. As the operator, I want to record the actual price rather than the suggested one, so that discounts and haggling are captured truthfully.
69. As the operator, I want to backdate a sale, so that I can enter a weekend market's takings on Monday.
70. As the operator, I want to edit or delete a sale, so that a mistyped price can be corrected.
71. As the operator, I want deleting a sale to return its copy to available, so that the object goes back into stock.

### Reporting

72. As the operator, I want to switch the report between this month, last month, this year and all time, so that I can see both recent activity and the whole picture.
73. As the operator, I want to see revenue for the period, so that I know what came in.
74. As the operator, I want the period's cost to be the cost of the copies I actually sold in it, so that profit and margin tell me whether my pricing works.
75. As the operator, I want production cost for the period shown as a separate labelled line, so that a heavy printing month is visible without being mistaken for a loss.
76. As the operator, I want to see profit and margin percentage, so that I know whether the pricing is working.
77. As the operator, I want to see copies sold, so that I can judge volume separately from money.
78. As the operator, I want gifted, kept and scrapped copies shown on their own line, so that the gap between cost and revenue is explainable rather than mysterious.
79. As the operator, I want designs ranked by profit on the same matched basis, so that I know what is worth printing again and what to retire.
80. As the operator, I want break-even progress against everything I have spent — printer, all filament bought, electricity — so that I know whether this hobby has paid for itself in cash terms.
81. As the operator, I want break-even to count filament I bought but have not printed yet, so that it reflects my bank account rather than an accountant's view.
82. As the operator, I want break-even and period profit clearly labelled as different measures, so that I am not confused when they disagree.
83. As the operator, I want current stock summarised — unsold copies and their value, spool grams and their value — so that I can see what I own without visiting three tabs.

### Application-wide

84. As the operator, I want six tabs matching the six things I think about, so that I always know where I am.
85. As the operator, I want consistent keys for add, edit, detail and delete across every tab, so that I learn the app once.
86. As the operator, I want a confirmation before anything is deleted, so that a stray keystroke does not destroy history.
87. As the operator, I want validation errors shown inline next to the offending field, so that I can fix them without losing what I typed.
88. As the operator, I want all money handled exactly, so that totals always equal the sum of their parts.
89. As the operator, I want one SQLite file I can copy, so that backing up is a file copy.
90. As the operator, I want to point the app at a different database file, so that I can keep a scratch database while experimenting.
91. As the operator, I want a single static binary with no runtime dependencies, so that installing is copying one file.

## Implementation Decisions

Every decision below is recorded in `CONTEXT.md` (glossary) and `docs/adr/`. Read the referenced
ADR before touching the corresponding area.

### Layering

Four layers, dependencies pointing inward:

- `internal/domain` — pure Go, no dependencies. Aggregates, cost math, invariants. Declares the
  repository interfaces it needs.
- `internal/store` — SQLite implementation of those repositories (ADR-0009).
- `internal/app` — the application layer: one method per use case, owning transaction boundaries.
  **This is the test seam** (see Testing Decisions). It is an addition to the three layers named in
  ADR-0001, adopted so that cost math and persistence are verified wired together rather than
  apart; ADR-0001 must be amended to record it.
- `internal/tui` — bubbletea + bubbles + lipgloss. A thin renderer over `internal/app`. Contains no
  cost math and no invariants.

The TUI must be able to do nothing that `internal/app` cannot. Any calculation the UI needs is a
value returned by a use case, not something the UI computes.

### Aggregates and invariants

- **Settings** — singleton, database-backed, edited in the TUI, no config file (ADR-0008). Holds
  `kwhPriceCents`, `kwhPerHourByType`, `machineHourlyRateCents`, `printerPurchaseCostCents`,
  `defaultMargin`, `minMargin`. No currency — the euro is fixed (ADR-0019). Each
  `kwhPerHourByType` entry carries a `measured bool` and is seeded with a plausible default at first
  launch.
- **Filament Type** — enum `PLA`, `PLA+`, `PETG`. Power rates live in Settings, not on the enum, so
  a new material is a settings row rather than a code change (ADR-0010).
- **Spool** — filament type, brand, colour, `initialGrams`, `tareGrams`, `purchaseCostCents`,
  `purchaseDate`. `costPerGram` derives from that spool's own purchase price, never a per-type
  average. Undeletable while any Filament Usage references it.
- **Spool Adjustment** — append-only child of Spool. Records `measuredGrams` from the scale;
  `derivedRemaining = measuredGrams − tareGrams`; `deltaGrams` against the previously computed
  remaining; date and note. Never edited, never deleted (ADR-0003, ADR-0011). Also the way a
  zero-yield run is recorded (ADR-0017).
- **Spool remaining is derived, never stored** (ADR-0003):
  `remaining = initialGrams − Σ(usage grams) + Σ(adjustment deltas)`.
  **Invariant: `remaining >= 0`**, enforced in the domain so the database cannot hold a negative
  spool. State `active`/`empty` derives at zero.
- **Capable Spool** — same type as the Design and `remaining >= Design.estimatedGrams`. A property
  of the pair, not of the Spool. Basis for quoting (ADR-0015). Note this is a stricter test than
  `active`, which only means `remaining > 0`.
- **Design** — `name`, `estimatedGrams` and `estimatedMinutes` **per single copy**,
  `defaultFilamentType`, non-null `marginPct` seeded from `Settings.defaultMargin` at creation
  (ADR-0014). No nullable override, no read-time fallback. No source URL, no notes.
- **Print** — design ref, date, `quantity >= 1`, minutes for the whole job, one-or-more filament
  usage rows **all of one Filament Type** (ADR-0012 as narrowed), `giftedCount`, `keptCount`,
  `scrappedCount`, plus a frozen rate snapshot (`kwhPriceCents`, `kwhPerHour`, `machineRateCents`)
  captured at creation (ADR-0004). Undeletable while any copy is accounted for.
- **Filament Usage** — `{spoolID, grams, costPerGramCents}` child of Print, one or more per print.
  `costPerGramCents` is copied from the Spool at creation, so cost is fully frozen (ADR-0004 as
  amended). Each row is validated against that spool's remaining; **overdraw is rejected, not
  warned about**, which forces a mid-print swap to be recorded as two rows on one print (ADR-0012).
- **Stock** — counted on the Print, no Unit record (ADR-0013).
  `available = quantity − soldCount − giftedCount − keptCount − scrappedCount`, where `soldCount`
  derives from Sales. **Invariant: `available >= 0`** — this one rule makes overselling impossible
  and also guards quantity edits. A partly failed plate raises `scrappedCount`; a zero-yield run is
  not a Print at all.
- **Sale** — `{printID, priceCents, date}`. The Print identifies the batch and its frozen cost.
  Warns when `priceCents < costPerCopy × (1 + minMarginPct)`, never blocks (ADR-0016).

### Cost math

`int64` minor units throughout; no floats for money; round once at the end (ADR-0002). `kwhPerHour`
stays a float64 — it is a measured physical rate, not money. Durations are integer minutes; input
`HH:MM`, display `5h 31m`.

Actual cost of a Print:

```
filament  = Σ over usages: grams × usage.costPerGramCents
energy    = hours × kwhPerHour × kwhPriceCents        (from the snapshot)
overhead  = hours × machineHourlyRateCents            (from the snapshot)
jobCost   = filament + energy + overhead
costPerCopy  = jobCost / quantity
```

Estimated cost and price of a Design, one row per Filament Type, from **current** Settings:

```
filament  = estimatedGrams × costPerGram of the most expensive Capable Spool of that type
energy    = estimatedHours × kwhPerHour[type] × kwhPriceCents
overhead  = estimatedHours × machineHourlyRateCents
estimated = filament + energy + overhead
suggested = ceilTo50c( estimated × (1 + Design.marginPct) )     -- defaultFilamentType row only
```

Reference cases, which the test suite must pin exactly. Spool PLA €22.00 / 1000g; kWh €0.28,
PLA 0.09 kWh/h, machine €0.35/h, margin 50%.

Print — 120g, 5h30m, quantity 2:

```
filament  120 × 2.2c/g       = 264c
energy    5.5 × 0.09 × 28c   =  14c
overhead  5.5 × 35c          = 193c
jobCost                      = 471c   (€4.71)
costPerCopy  471 / 2            = 236c   (€2.36)
```

Design — 60g and 2h45m per copy:

```
filament  60 × 2.2c/g        = 132c
energy    2.75 × 0.09 × 28c  =   7c
overhead  2.75 × 35c         =  96c
estimated                    = 235c   (€2.35)
suggested 235 × 1.5 = 353c   → €4.00
```

Per-copy cost divides across **all** copies including scrapped ones: every copy cost the same to
make, and scrap gets its own reporting line rather than inflating the survivors. Suggested price
always rounds **up** to €0.50 — never nearest, never down (ADR-0007).

### Two profit measures, both reported

Deliberately different numbers, always labelled (ADR-0006):

- **Period profit** matches cost to revenue: revenue is Sales dated in the period, cost is the
  `costPerCopy` of the copies sold in it. Margin % therefore answers "is my pricing working".
  Production cost for the period is a separate labelled line so a heavy printing month is visible.
- **Break-even** is cash-based and all-time:
  `revenue − (printerPurchaseCostCents + all spool spend + energy spend)`. It counts spools bought
  but not yet printed, because that cash is really gone.

### Persistence

SQLite via `modernc.org/sqlite` — pure Go, no cgo, so `CGO_ENABLED=0` produces a static binary
(ADR-0009). Default path `./printer-admin.db`, overridable with `-db`. Schema evolution handled by
an ordered migration list in `internal/store`. A print plus its usage rows is written in one
transaction; the transaction boundary belongs to `internal/app`.

### Editing

Spools, designs, prints and sales are editable in place; delete is available with confirmation
(ADR-0011). Editing a print recomputes spool remaining for free, because remaining is derived, and
the `remaining >= 0` invariant is re-validated on every edit. Shrinking a print's quantity is
guarded by the same `available >= 0` invariant that prevents overselling. Deleting a sale returns
its copy to available. Deletion is blocked for a Print with any accounted-for copy, and for a Spool
any print has drawn from.

### TUI

Tabs `Spools │ Designs │ Prints │ Sales │ Report │ Settings`. Keys: `a` add, `e` edit, `enter`
detail, `d` delete with confirm, `tab`/`shift-tab` to switch. Validation errors render inline beside
the field, and the entered form state survives a rejected submit. Report periods: this month, last
month, this year, all time. The Design screen shows a cost row per Filament Type with the sourcing
spool named, and one headline suggested price. The Print form updates its cost breakdown live as
fields change and shows per-copy cost; it shows no suggested price.

## Testing Decisions

### What makes a good test here

A test drives a use case on `internal/app` and asserts on returned values or on what a subsequent
query reports. It never reaches into a struct field to check that some intermediate step happened,
never asserts on SQL, and never asserts on rendered terminal output. If a test would break when the
cost calculation is moved between packages while still producing €4.71, it is testing the wrong
thing.

Money assertions are exact integers. There are no tolerance windows — ADR-0002 exists precisely so
that `471` means `471`.

### The seam

One seam: the `internal/app` use-case API, exercised against a real SQLite database in
`t.TempDir()`. This deliberately covers domain plus store in one place, so the reference cost case
is verified end to end through persistence rather than in a pure unit test that a wiring bug could
bypass.

Shape:

```go
a := app.New(store.MustOpen(filepath.Join(t.TempDir(), "test.db")))
sp, _ := a.AddSpool(app.AddSpoolCmd{Type: domain.PLA, InitialGrams: 1000,
    TareGrams: 210, CostCents: 2200})
_, err := a.RecordPrint(app.RecordPrintCmd{DesignID: d.ID, Quantity: 2,
    Minutes: 330, Usages: []app.Usage{{SpoolID: sp.ID, Grams: 1200}}})
// err is ErrInsufficientFilament — only 1000g left
```

Test helpers construct a fixture with known Settings so cost assertions are stable.

### What is tested

- **Print cost** — the reference case above, exactly; per-copy division with scrapped copies
  included.
- **Design pricing** — the Design reference case exactly; the €0.50 round-up including the
  already-exact boundary; a design's own margin driving its price; a new design inheriting
  `defaultMargin`; changing `defaultMargin` leaving existing designs untouched.
- **Quote basis** — with three PLA spools in stock the most expensive _capable_ one is chosen; a
  spool too small for `estimatedGrams` is excluded even though it is active; with no capable spool
  the fallback reference price is used and flagged; a type never bought produces no row.
- **Spool remaining** — consumption by prints; re-weigh adjustment computed from tare; sequences of
  usages and adjustments; the `remaining >= 0` invariant.
- **Overdraw and type rules** — a single row exceeding remaining is rejected; the same total split
  correctly across two spools of one type succeeds; two spools of different types on one print is
  rejected.
- **Frozen cost** — a print's cost is unchanged after Settings are edited _and_ after the source
  spool's purchase price is edited; a print recorded afterwards uses the new rates and the new gram
  price.
- **Stock** — `available` arithmetic; accounting for more copies than were produced fails; selling
  the last available copy then selling again fails; deleting a sale restores availability; gifted,
  kept and scrapped copies keep cost but contribute no revenue.
- **Edit and delete guards** — shrinking quantity below accounted-for copies fails; deleting a print
  with a sold copy fails; deleting a spool with usage fails; editing a print's grams recomputes
  remaining and re-validates the invariant.
- **Reporting** — period boundaries (a sale on the first and last day of a month); revenue matched
  against the cost of copies sold, including a copy sold in a later period than it was printed;
  production cost line; design ranking; break-even including unprinted spool spend.
- **Migrations** — opening a fresh database succeeds; opening an existing one is idempotent.

### Prior art

None — this is the first code in the repo, so these tests _are_ the prior art and their shape sets
the convention. Standard library `testing` with table-driven cases, no assertion framework, no
mocks: the real SQLite store in a temp directory is cheaper and more honest than a fake.

The TUI is not tested automatically. It is a renderer over `internal/app`, driven manually.

## Out of Scope

- Multiple printers, multiple operators, and any notion of user accounts.
- Multiple designs on one plate. A mixed plate is recorded as separate Prints.
- Per-object identity: serial numbers, which physical copy a buyer received, per-copy defects
  (ADR-0013).
- Recording runs that produced nothing as Prints; they are spool re-weighs, so failure rate and
  failed machine hours are not reported (ADR-0017).
- Enclosed-chamber materials (ABS, ASA, PC, Nylon) — the A1 cannot print them. TPU is capable but
  excluded from v1 (ADR-0010).
- Shipping costs, platform fees, taxes, invoices, and buyer records. Sales carry only print, price
  and date.
- Labour time and non-filament consumables (packaging, magnets, inserts, glue).
- Custom report date ranges, charts, trend lines, and exports.
- Design source URLs, notes, tags, images, and any slicer or printer integration. Grams and time
  are typed in by hand.
- An audit trail of edits. Records are corrected in place (ADR-0011); only spool adjustments are
  append-only.
- Multi-currency and FX. The euro is fixed and the symbol is a renderer constant (ADR-0019).
- Networking, sync, cloud backup, and concurrent access to the database.
- Automated TUI tests.

## Further Notes

Period profit and break-even disagreeing is expected, not a bug: period profit matches the cost of
what was sold against what it sold for, while break-even counts every euro that has left the bank,
including the printer and filament still on the shelf. Both stay on screen, both labelled. Anyone
tempted to "fix" the discrepancy should read ADR-0006 first.

A Design's quote is _live_: it moves when rates change, when `estimatedGrams` changes, and when the
spool it was quoting against runs out. A Print's cost is _frozen_ and never moves. That asymmetry is
deliberate — one is a question about the future, the other is a record of the past.

`machineHourlyRateCents` is an operator estimate and now carries more weight than before: failed
runs are not recorded as Prints, so their machine hours exist only inside that buffer and can never
be checked against reality (ADR-0017). Suggested prices are advisory; the recorded sale price is
whatever was actually charged, and the minimum-margin check only ever warns (ADR-0016).

The `internal/app` layer amends ADR-0001. Blocking overdraw (ADR-0012) will occasionally stand in
the way when slicer estimates undershoot reality; the intended escape hatch is a re-weigh
adjustment, which is the correct fix anyway.

Buying an enclosed printer reopens ADR-0010 together with the single-printer assumption baked
through Settings.
