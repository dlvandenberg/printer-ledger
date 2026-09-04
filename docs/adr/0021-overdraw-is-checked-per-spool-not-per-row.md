# ADR-0021: Overdraw is checked per Spool, not per Filament Usage row

**Status:** Accepted — 2026-09-04

## Context

ADR-0012 gave a Print one or more Filament Usage rows and said each row is validated against that
Spool's remaining. With one row per Spool that is the same rule stated two ways. It stops being so
once rows can repeat a Spool, which they can: a two-colour job that swaps back to the first Spool
is two rows naming it.

Checked literally per row, a Spool with 100g left accepts a row of 60g and a second row of 60g —
each passes alone — and the Print takes the Spool to −20g. That breaks the `remaining >= 0`
invariant of ADR-0003 in the one place ADR-0012 exists to defend it.

## Decision

Grams are summed **per Spool** across the Print's rows, and each sum is compared against that
Spool's remaining. Two rows naming one Spool are legal; together they may not exceed what it holds.

The overdraw is reported once per Spool, on the row whose grams cross what is left, keyed to that
row so the message sits where the operator is typing. A third row on an already-overdrawn Spool
adds no second message.

The alternative — rejecting a duplicate Spool on one Print, which would make the per-row check
correct again — was considered and rejected. A swap back to the first Spool is a real job, and
forbidding it would push the operator into merging rows by hand, which is exactly the arithmetic
the Print is supposed to do.

## Consequences

- `remaining >= 0` holds for every combination of rows, not only for rows naming distinct Spools.
- The reported shortfall is the Spool's remaining, not what is left after earlier rows. The row key
  names the row; the number names the Spool.
- A Print may name one Spool on several rows, so nothing downstream may assume rows are unique by
  Spool.
