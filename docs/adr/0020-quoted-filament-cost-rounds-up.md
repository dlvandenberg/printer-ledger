# ADR-0020: Quoted filament cost rounds up, inventory value truncates

**Status:** Accepted — 2026-09-03

## Context

`costPerGram` is a ratio — `purchaseCostCents / initialGrams` — and rarely divides evenly. A spool
at €25.00 per 1000g is 2.5c/g, so 81 grams of it cost 202.5c. That figure has to land on a whole
cent somewhere.

Until now one method answered every question about what some grams of a Spool are worth:
`Spool.ValueOf`, which truncates. It is used for two different purposes:

- `SpoolLedger.remainingValue` — what the filament still on the shelf is worth, an **Inventory**
  figure that is reported and summed.
- the `filament` line of an **Estimated Cost** — what one copy of a **Design** would cost, which is
  then marked up and spoken aloud as a **Suggested Price**.

Truncating suits the first and works against the second. ADR-0015 picks the *most expensive*
Capable Spool and ADR-0007 rounds the price *up*, both because the cost of erring low is a price
already promised to a customer. A truncating filament line quietly pulls in the other direction.

The error is at most one cent per copy and the €0.50 round-up usually absorbs it — but not always.
An 80g design at 2.5c/g, 4h, 0% margin totals exactly 350c and prices at exactly €3.50, while the
true cost is 350.5c. On a 50c boundary the cent survives all the way to the price.

## Decision

The two questions get two methods.

- `Spool.ValueOf` keeps truncating. Inventory should never claim value that is not there, and these
  figures are summed across every spool, where a systematic round-up would accumulate.
- `Spool.QuotedValueOf` rounds up, and is what the `filament` line of an Estimated Cost uses. A
  quote never sits below what the filament actually cost.

Both stay exact when the ratio divides evenly, so every worked example in `CONTEXT.md` is
unchanged: 60g at 2.2c/g is 132c either way.

## Considered Options

- **One method, truncating** (the status quo). Rejected: it makes the one number that becomes a
  price the only one in the quote rounded in the unsafe direction.
- **One method, rounding up.** Rejected: it would inflate `remainingValue` and the Inventory block,
  by up to a cent per spool, on figures that are summed rather than quoted.
- **One method, rounding half-up.** Rejected: still below true cost half the time, and it answers
  neither question well — a compromise where the two callers genuinely want opposite things.
- **Store `costPerGram` at higher precision on the Spool.** Rejected: a second stored number that
  can disagree with `purchaseCostCents / initialGrams`, to fix a one-cent problem.

## Consequences

- A Design's `filament` line, and therefore its Suggested Price, may be one cent above the exact
  product. That is the safe direction, consistent with ADR-0007 and ADR-0015.
- Inventory value and a quote may differ by a cent for the same grams of the same spool. They are
  answering different questions, and only one of them becomes a price.
- A **Print**'s cost is unaffected: it multiplies actual grams by the `costPerGramCents` frozen on
  each **Filament Usage** row, and is history rather than a quote.
