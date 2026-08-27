# ADR-0011: Edit records in place, except spool adjustments

**Status:** Accepted — 2026-08-25. Reworded 2026-08-26 for counted stock (ADR-0013).

## Context

Prints and sales are historical facts, which argues for append-only records corrected by
compensating events. But this is a single-operator tool with no accounting obligation, and typos
happen: `1200g` instead of `120g` wrecks a spool's remaining and every cost derived from it.

The alternative to editing was delete-and-recreate. Its failure mode is worse: a print with a sold
copy cannot be deleted at all, since that would erase the revenue — making a typo on a sold print
permanently unfixable.

## Decision

Spools, designs, prints, and sales are editable in place. Delete is available with a confirmation
prompt.

One guard, and it is the `available >= 0` invariant doing double duty (ADR-0013): a print's
`quantity` cannot shrink below the copies already sold, gifted, kept or scrapped, and a print with
any such copy cannot be deleted. Nor can a spool that a print has drawn from.

Spool adjustments remain append-only (ADR-0003) — they are corrections, and correcting a correction
is incoherent.

## Consequences

- Editing a print recomputes spool remaining for free, because remaining is derived (ADR-0003).
  The `remaining >= 0` invariant is re-validated on every edit.
- No audit trail of what a record used to say. Accepted: single operator, no compliance need.
- Update paths are code that delete-and-recreate would not need. Worth it; the quantity guard is
  free, being the same invariant that prevents overselling.
- Editing a spool's purchase price does *not* re-cost past prints: the gram price is frozen onto
  each usage row (ADR-0004 as amended).
