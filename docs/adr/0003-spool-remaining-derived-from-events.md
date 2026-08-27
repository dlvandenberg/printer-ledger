# ADR-0003: Spool remaining is derived, corrected by append-only adjustments

**Status:** Accepted — 2026-08-25

## Context

The operator needs to know how much filament a spool has left. Slicer estimates drift from reality
(purge towers, tangles, partial spools), so a counter alone goes stale. The physical correction
available is a kitchen scale, which reads filament *plus* spool.

## Decision

`remaining` is never stored. It is derived:

```
remaining = initialGrams − Σ(Filament Usage grams) + Σ(Spool Adjustment deltaGrams)
```

`Spool` stores `tareGrams`, the empty spool's weight. A **Spool Adjustment** records a re-weigh:
the operator enters `measuredGrams` from the scale, the domain derives
`derivedRemaining = measuredGrams − tareGrams` and stores the delta against the previously computed
remaining.

Spool Adjustments are append-only: never edited, never deleted. They are the correction mechanism,
so correcting them would need a correction mechanism of its own.

**Invariant:** `remaining >= 0`, enforced in the domain. See ADR-0012 for how prints respect it.

## Consequences

- Fully auditable: the spool detail screen explains its number (initial, printed, adjusted).
- Editing a print recomputes remaining for free — there is no counter to keep in sync (ADR-0011).
- Storing `tareGrams` means the operator types what the scale says, not what it says minus a number
  they had to remember.
- Remaining is a computed read, so it needs the usage rows loaded. Fine at this data scale.
