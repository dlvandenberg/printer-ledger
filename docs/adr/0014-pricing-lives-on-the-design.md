# ADR-0014: Pricing lives on the Design; a Print shows cost only

**Status:** Accepted — 2026-08-26

## Context

Suggested Price was originally shown in two places: on the Design from slicer estimates
(deliberately stable) and on the Print from actuals. Two numbers computed from different inputs,
built to disagree, for a single physical object.

The moment a price is actually needed is the one in the problem statement: someone asks what a
print costs *before* it exists. Once the objects are in the drawer, the useful number is what they
cost, so you know your floor.

## Decision

Suggested Price appears on the **Design** only, computed from its per-copy estimates. A `Print`
shows the cost breakdown and `costPerCopy`, and no price.

Margin is `Design.marginPct`: non-null, seeded from `Settings.defaultMarginPct` when the Design is
created, edited freely thereafter. The nullable `marginOverridePct` and its read-time fallback are
gone.

## Considered Options

- **Price on the Print, on a recovery basis** (`jobCost / survivors`, so a print that goes badly
  suggests charging more). Rejected: the price would move every time a copy was scrapped, after it
  had already been quoted, and the Design's stable quote would contradict it.
- **Drop margin entirely** in favour of a typed `listPriceCents` per Design, with the app reporting
  the implied margin. Rejected: it removes the app's ability to price a design that has never been
  priced, which is half the problem the app exists to solve.

## Consequences

- One price per Design, one meaning, no reconciliation.
- A print that overran its estimates does not suggest recovering the overrun. The Print screen
  shows cost against the Design's price, and the operator decides.
- Raising `Settings.defaultMarginPct` re-prices nothing that already exists. Existing Designs keep
  their margin until individually edited — predictable, but a global raise is manual.
- `costPerCopy` divides across all copies including scrapped ones, unchanged.
