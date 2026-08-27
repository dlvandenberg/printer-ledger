# ADR-0012: A print holds multiple filament usage rows; overdraw is blocked

**Status:** Accepted — 2026-08-25. Narrowed 2026-08-26: multiple spools, one Filament Type.

## Context

A single print can draw from more than one spool: a colour swap, a multi-material job, or — most
commonly — a spool running out mid-print. If a print referenced exactly one spool, a mid-print swap
would have to be recorded as two separate prints, which is a lie about what was printed and breaks
per-copy cost.

Separately, the spool `remaining >= 0` invariant (ADR-0003) needs enforcing somewhere.

## Decision

A `Print` holds one or more `FilamentUsage{spoolID, grams, costPerGramCents}` rows.

Each row is validated against that spool's current remaining. Overdraw is **rejected** with an
inline error (`PLA Black: only 40g left`), not warned about. To record a print that outran a spool,
the operator splits it into two rows: 40g from the emptied spool, 80g from its replacement.

The form opens with one row, spool selected, grams prefilled from the design estimate. The extra
row is only needed in the rare swap case.

## Consequences

- Mid-print spool swaps are recorded explicitly and accurately, as one print.
- Spool remaining can never go negative, in the domain or the database.
- Slicer estimates and real consumption differ, so blocking overdraw can occasionally stand in the
  way. The escape hatch is a re-weigh adjustment (ADR-0003), which is the correct fix anyway.
- Multi-material prints are supported for free, within one Filament Type.

## Narrowing (2026-08-26)

All usage rows on a Print must be Spools of the **same Filament Type**. Energy is
`hours × kwhPerHour[filamentType] × kwhPriceCents`, and a Print spanning PLA and PETG has no
defined rate to look up. A grams-weighted blend was considered and rejected as machinery for a job
that does not happen here; a mixed-type print is rejected with an inline error instead. A two-colour
swap within one type — the common case this ADR was written for — is unaffected.

This ADR is also not an argument by analogy for multiple Designs on one Print. A plate mixing
designs is recorded as separate Prints; only the *inputs* are one-or-more, not the outputs.
