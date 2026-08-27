# ADR-0005: Unit-level inventory, failures as scrapped units

**Status:** Superseded by ADR-0013 — 2026-08-26

## Context

A print job can yield several copies. Not every copy gets sold: some are gifts, some come out bad,
some are kept. A sale must not be recordable against a copy that does not exist or is already sold.
An early design tracked failures with a `goodQuantity` field on Print.

## Decision

A `Print` of quantity N creates N `Unit` records. Each Unit has status
`available | sold | gifted | scrapped`.

- A `Sale` may only reference an `available` Unit. Overselling is impossible by invariant, not by
  UI convention.
- A failed print is recorded by marking Units `scrapped`. No `goodQuantity`, no print-level
  status field — one concept instead of two.
- `gifted` and `scrapped` Units keep their share of job cost (the money was really spent) but never
  contribute revenue or units-sold. Reporting gives them their own line so profit stays explainable.
- Per-unit cost is `jobCost / quantity` across *all* units, including scrapped ones.

## Consequences

- The sale flow picks a real object, and its status transition is the ledger.
- "What's in the drawer" is a query, not a guess: unsold Units and their cost value.
- Rows scale with quantity printed. Irrelevant at hobby scale.

## Superseded

Once a `Print` was fixed to a single `Design`, a Unit row carried no fact distinguishing one copy
from another. ADR-0013 replaces the Unit table with counts on the Print. The parts of this ADR that
survive: failures are still not a print-level status field, gifted and scrapped still keep their
share of cost, and per-unit cost still divides across all copies.
