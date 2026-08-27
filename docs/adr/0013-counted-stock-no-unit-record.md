# ADR-0013: Stock is counted on the Print; there is no Unit record

**Status:** Accepted — 2026-08-26. Supersedes ADR-0005.

## Context

ADR-0005 gave every copy of a print its own `Unit` record with a status, so that a Sale referenced
a real object and overselling was impossible by foreign key rather than by UI convention.

A later decision fixed a `Print` to exactly one `Design`. That made the Unit row `{id, printID,
status}` and nothing else: with one design and N copies off one plate, there is no fact that
distinguishes copy #2 from copy #3. ADR-0005 rejected a `goodQuantity` field for being a second
concept, but what it built instead was a row per object carrying no data about that object.

## Decision

No `Unit` table. A `Print` carries `giftedCount`, `keptCount` and `scrappedCount`; `soldCount`
derives from the Sales referencing that Print.

```
available = quantity − soldCount − giftedCount − keptCount − scrappedCount
```

**Invariant: `available >= 0`**, enforced in the domain.

A `Sale` references the Print it came from. `kept` is a new outcome ADR-0005 named in its context
paragraph but never modelled: copies that are yours, distinct from `gifted` copies that have left.

## Considered Options

Keeping the Unit table. Rejected because it costs an entity, a table and a status enum to
distinguish objects that are by definition interchangeable, and it forces a "which copy do I
delete?" question that has no meaningful answer when quantity shrinks.

## Consequences

- One invariant does two jobs. Overselling and "quantity cannot shrink below what is accounted
  for" are now the same check, not two guards that can disagree.
- Deleting a Sale restores availability by arithmetic rather than by a status transition.
- Per-object provenance is given up permanently. The app can never say which physical copy a
  buyer received, carry a serial number, or record that one specific copy had a layer shift.
  Retrofitting identity means backfilling rows that never existed.
- Rows no longer scale with quantity printed.
