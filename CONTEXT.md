# printer-ledger — Domain Context

A single-user TUI ledger for a 3D printing side business. Tracks filament stock, prints, and
sales; computes what a print actually cost and what a design should sell for.

**Scope:** one operator, one printer (Bambu Lab A1), local in-person sales.
No shipping, no platform fees, no labour tracking, no multi-user, no networking.

**Bounded contexts:** one. The whole app is a single context; no seams worth the cost.

## Glossary

Use these terms exactly. Avoid the listed synonyms.

### Settings

Singleton aggregate holding the global rates that drive cost math.

| Field | Meaning |
|---|---|
| `kwhPrice` | Electricity price per kWh |
| `powerRates` | One **Power Rate** per **Filament Type** |
| `machineHourlyRate` | Machine overhead per print hour: depreciation + maintenance + failure buffer |
| `printerPurchaseCost` | What the printer cost. Used only by **Break-Even**. Seeds to zero — the operator is the only source |
| `defaultMargin` | Seed value copied into a new **Design**'s `margin` |
| `minMargin` | Floor below which a **Sale** price is flagged |

Edited in-app on the Settings tab, stored in the database. Not a config file — these are
historical inputs to reporting, so they belong with the data.

The ledger is denominated in euros throughout; the `€` the operator sees is a TUI constant, not a
setting. Currency is not configurable and there is no FX.

### Power Rate

The power draw of one **Filament Type**, in `kwhPerHour`, plus a `measured` flag. Avoid: "wattage",
"power setting".

Every Power Rate is seeded with a plausible default at first launch, so energy is never silently
costed at zero. It becomes `measured` once the operator replaces the seeded figure with a smart-plug
reading — that is, once the value moves. An unmeasured rate warns on the Print form. The TUI renders
the flag as `measured` or `default`; those two words are labels, not a stored value.

### Percent

A margin, held as hundredths of a percent (`50%` is `5000`) so that cost math stays integral
(ADR-0018). `kwhPerHour` is the one rate that stays a `float64`: it is a physical measurement, not
money.

### Filament Type

Enum: `PLA`, `PLA+`, `PETG`. The material class of a **Spool**.

Rates hang off Settings (`powerRates`), not off the enum, so adding a type is a settings
row rather than a code change.

Enclosed-chamber materials (ABS, ASA, PC, Nylon) are out of scope: the A1 is an open-frame
bedslinger. TPU is printable on the A1 (direct drive) but excluded from v1 for simplicity.

### Spool

A physically purchased spool of filament. Not a filament *kind* — two colors of the same PLA are
two Spools of one **Filament Type**.

| Field | Meaning |
|---|---|
| `filamentType` | **Filament Type** |
| `brand`, `color` | Identification |
| `initialGrams` | Filament grams when bought |
| `tareGrams` | Weight of the empty spool, used to interpret scale readings |
| `purchaseCostCents` | What was paid |
| `purchaseDate` | When |

Derived:

- `costPerGram = purchaseCostCents / initialGrams` — **per spool**, never a per-type average.
  A cheap spool genuinely makes a cheaper print.
- `remaining = initialGrams − Σ(Filament Usage grams) + Σ(Spool Adjustment deltas)`
- state: `active` while `remaining > 0`, `empty` at `remaining == 0`

**Invariant:** `remaining >= 0`. Enforced in the domain, so the database can never hold a
negative spool.

A Spool cannot be deleted while any **Filament Usage** references it.

Avoid: "filament" as a synonym for Spool; "roll".

### Capable Spool

A **Spool** that could actually supply a given **Design**: same **Filament Type**, and
`remaining >= Design.estimatedGrams`. Distinct from `active`, which only means `remaining > 0` —
a spool with 3g left is active but capable of almost nothing.

Capability is relative to a Design, not a property of the Spool, and it is the basis on which a
Design is quoted. See **Estimated Cost**.

### Spool Adjustment

An append-only correction event on a **Spool**, created by re-weighing. The operator puts the
spool on a scale and enters the total measured weight; the domain derives the true remaining
from the spool's `tareGrams`.

| Field | Meaning |
|---|---|
| `measuredGrams` | Total weight on the scale (filament + spool) |
| `derivedRemaining` | `measuredGrams − tareGrams` |
| `deltaGrams` | `derivedRemaining` minus the previously computed remaining |
| `date`, `note` | Provenance |

Adjustments are never edited or deleted — they *are* the correction mechanism. Everything else
in the app is edited in place.

A run that produced nothing — a print that failed at layer three — is recorded here, not as a
**Print**. The filament is genuinely gone; the machine hours are absorbed by
`machineHourlyRate`.

### Spool Ledger

A **Spool** together with the event totals its `remaining` derives from. A read model, not a stored
record — nothing in the database corresponds to it.

| Field | Meaning |
|---|---|
| `spool` | The **Spool** itself |
| `usedGrams` | `Σ` grams over every **Filament Usage** row referencing it |
| `adjustedGrams` | `Σ deltaGrams` over every **Spool Adjustment** on it |

Derived:

- `remaining = spool.initialGrams − usedGrams + adjustedGrams`
- `remainingValue = remaining × spool.costPerGram`
- state: `active` / `empty`

It exists so that the derivation in ADR-0003 lives in the domain while the two sums are fetched by
the store. The same expression answers two questions that must never disagree: what the Spools list
displays, and whether a **Filament Usage** row would overdraw the spool. `remaining >= 0` therefore
has one definition rather than one per caller.

Explaining a remaining figure to the operator — the initial / printed / adjusted breakdown on the
spool detail screen — needs the rows themselves, not these totals, and is a separate query.

### Design

A printable model, and the only place a price is decided.

| Field | Meaning |
|---|---|
| `name` | Identification |
| `estimatedGrams` | Slicer filament estimate **for one copy** |
| `estimatedMinutes` | Slicer time estimate **for one copy** |
| `defaultFilamentType` | The **Filament Type** this design is normally printed in |
| `marginPct` | Margin used for **Suggested Price**. Seeded from `Settings.defaultMargin` at creation, edited freely thereafter |

Estimates are always per single copy. A plate of four is recorded as a **Print** of
`quantity` 4 whose actuals are typed in from the slicer; the Design's numbers do not change.

`marginPct` is never null and has no fallback rule. Raising `Settings.defaultMargin` affects
Designs created afterwards, never existing ones.

### Estimated Cost

What a **Design** would cost to print, shown as one row per **Filament Type**:

```
filament = estimatedGrams × costPerGram of the most expensive Capable Spool of that type
energy   = estimatedHours × kwhPerHour[type] × kwhPrice
overhead = estimatedHours × machineHourlyRate
```

`filament` rounds **up** to the whole cent, where a **Spool**'s inventory value truncates
(ADR-0020) — the same safe direction as **Suggested Price**.

The gram price is the **most expensive Capable Spool**, and the row names it. Quoting off the
cheap spool and then printing off the expensive one loses money on a price already spoken aloud;
the direction of error is deliberately safe, as in **Suggested Price**.

Fallbacks: no Capable Spool of that type → the most expensive Spool of that type ever bought,
labelled a reference price rather than a live one. That type never bought → no row.

Estimated Cost reads **current** Settings. It is a live quote, not history, and it moves when
rates, stock or `estimatedGrams` change.

### Suggested Price

`ceilTo50c( Estimated Cost × (1 + Design.marginPct) )`, from the `defaultFilamentType` row.

Rounded **up** to the nearest €0.50 — always up, never nearest. Shown **only on the Design**.
Other **Filament Type** rows show cost alone, so the effect of a material switch is visible
without implying a second price.

A **Print** shows cost and per-copy cost, never a suggested price. Once objects exist the
question is what they cost you, and the price was already decided on the Design.

### Design Quote

A **Design**'s **Estimated Cost** in every **Filament Type** stocked, together with the one
**Suggested Price** and which row it came from. What you would say out loud to a customer.

A read model, not a stored record. It reads current Settings and current stock, so it moves; a
**Print**'s cost does not.

### Print

One machine run. Produces `quantity` copies of one **Design**.

| Field | Meaning |
|---|---|
| `designID` | The **Design** printed |
| `date` | When (defaults to today, editable — backdating supported) |
| `quantity` | How many copies came off the plate. At least 1 |
| `minutes` | Actual print time for the whole job. Entered `HH:MM`, stored as minutes, shown `5h 31m` |
| `usages` | One or more **Filament Usage** rows, all of one **Filament Type** |
| `giftedCount`, `keptCount`, `scrappedCount` | See **Stock** |
| rate snapshot | `kwhPriceCents`, `kwhPerHour`, `machineRateCents` frozen at creation |

Actuals are typed in from Bambu Studio or the printer screen and cover the whole job. The form
prefills `estimatedGrams × quantity` and `estimatedMinutes × quantity` from the **Design**.

All **Filament Usage** rows share one **Filament Type**, so the energy rate is unambiguous. A
Print's type need not match the Design's `defaultFilamentType` — printing a PETG design in PLA is
normal, and that is what the other **Estimated Cost** rows are for.

The rate snapshot, together with the per-gram price frozen on each usage row, means no later edit
to Settings or to a Spool ever rewrites the cost of a past Print.

A run that yielded nothing is not a Print. See **Spool Adjustment**.

Avoid: "job", "plate".

### Filament Usage

A child of **Print**: `{ spoolID, grams, costPerGramCents }`. A Print holds one or more.

`costPerGramCents` is copied from the **Spool** at creation and never changes, so correcting a
spool's purchase price affects later prints only.

Multiple rows exist so a mid-print spool swap is one Print, not two: 40g from the spool that ran
out plus 80g from its replacement. Each row is validated against that Spool's `remaining`, which
is what keeps the `remaining >= 0` invariant true and forces swaps to be recorded explicitly.

All rows on one Print must be Spools of the same **Filament Type**.

### Copy

One physical printed object, produced by a **Print**. A Print of `quantity` 4 produces four Copies.

Copies are **counted, never individually recorded**: all Copies of one Design from one Print are
interchangeable, so there is no fact that distinguishes one from another and nothing to identify.
What is tracked is how many are in each state — see **Stock**.

Avoid: "Unit" (there is no such record), "item", "piece".

### Stock

What became of a **Print**'s **Copies**. Scoped to one Print — the whole-business view of
everything owned, filament included, is the **Inventory** block on the report.

```
soldCount  = number of Sales referencing this Print   (derived)
available  = quantity − soldCount − giftedCount − keptCount − scrappedCount
```

**Invariant:** `available >= 0`. This single rule makes overselling impossible, and is also what
stops `quantity` being edited below what has already been accounted for.

- `gifted` left the building; `kept` is yours and might still be sold later; `scrapped` is ruined.
- All three keep their share of job cost — the money was really spent — but never contribute
  revenue or copies-sold. Reporting shows them on their own line.
- A partly failed plate is recorded by raising `scrappedCount`. There is no separate
  failure concept and no print-level status field.

A **Print** cannot be deleted while any copy is sold, gifted, kept or scrapped.

### Inventory

Everything owned right now, across both kinds of thing: unsold **Copies** with their cost value,
and **Spool** remaining grams with their value. Broader than **Stock**, which is per-Print.

Always on screen in the report, independent of the selected period.

### Sale

`{ printID, priceCents, date }`. The Print identifies which batch the copy came from, and
selling from it is what decrements `available`.

Warns — never blocks — when `priceCents < costPerCopy × (1 + Settings.minMargin)`. The ledger's
job is to record what was actually charged, haggling included; a refused sale would simply go
unrecorded.

Deleting a Sale returns its copy to `available`.

### Cost math

All money is `int64` minor units (cents). Never floats for money. Round only at the end.

```
filament  = Σ over usages: grams × usage.costPerGramCents
energy    = hours × kwhPerHour × kwhPriceCents
overhead  = hours × machineHourlyRateCents
jobCost   = filament + energy + overhead
costPerCopy  = jobCost / quantity
```

`costPerCopy` divides across **all** copies including scrapped ones: every copy cost the same to
make, and the scrap shows up as its own reporting line rather than by inflating the survivors.

Worked example — spool PLA €22.00 / 1000g; print 120g, 5h30m, quantity 2;
kWh €0.28, PLA 0.09 kWh/h, machine €0.35/h:

```
filament  120 × 2.2c/g              = 264c
energy    5.5 × 0.09 × 28c          =  14c
overhead  5.5 × 35c                 = 193c
jobCost                             = 471c   (€4.71)
costPerCopy  471 / 2                   = 236c   (€2.36)
```

The same design quoted on the Design screen — 60g and 2h45m per copy, margin 50%:

```
filament  60 × 2.2c/g               = 132c
energy    2.75 × 0.09 × 28c         =   7c
overhead  2.75 × 35c                =  96c
estimated                           = 235c   (€2.35)
suggested 235 × 1.5 = 353c          → €4.00
```

### Profit vs Break-Even

Two different questions, deliberately both reported and clearly labelled:

- **Per-period profit** matches cost to revenue: cost is the `costPerCopy` of copies *sold* in the
  period, so margin % answers "is my pricing working".
- **Break-Even** is cash-based and all-time:
  `revenue − (printerPurchaseCost + all spool spend + energy spend)`.
  It counts spools bought but not yet printed, because that cash is really gone.

Per-period profit and the change in the bank balance are therefore different numbers. Break-Even
is where the cash view lives.

## Reporting

Periods: this month, last month, this year, all time.

Per period: revenue, cost of copies sold, profit, margin %, copies sold, a separate production
cost line for Prints made in the period, a gifted/kept/scrapped line, and Designs ranked by
profit on the same matched basis.

Always visible: **Break-Even** progress, plus a Inventory block — unsold copies and their cost value,
Spool remaining grams and value.

No custom date ranges, no charts.

## Architecture

```
internal/domain   pure Go, no dependencies. Cost math and invariants.
                  Owns the repository interfaces it needs.
internal/store    SQLite (modernc.org/sqlite — pure Go, no cgo). Implements the repositories.
internal/app      One method per use case. Owns transaction boundaries. The test seam.
internal/tui      bubbletea + bubbles + lipgloss. A renderer over internal/app.
```

Dependencies point inward: `tui → app → domain ← store`. See ADR-0001.

The domain is testable without a database or a terminal; that is the point of the split. Tests
nonetheless drive `internal/app` against a real SQLite database, so cost math is verified wired to
persistence rather than apart from it.

Tabs: `Spools │ Designs │ Prints │ Sales │ Report │ Settings`.
Keys: `a` add, `e` edit, `enter` detail, `d` delete (with confirm), `tab`/`shift-tab` switch.

Database: `./printer-ledger.db`, overridable with `-db`.
