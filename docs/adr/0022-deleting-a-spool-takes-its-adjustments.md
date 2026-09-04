# ADR-0022: Deleting a Spool takes its Adjustments with it

**Status:** Accepted — 2026-09-04. Narrows ADR-0003.

## Context

ADR-0003 makes Spool Adjustments append-only: never edited, never deleted, because they *are* the
correction mechanism and correcting a correction is incoherent. ADR-0011 then made Spools
deletable, guarded only against a Spool a Print has drawn from. The two meet on a Spool that was
re-weighed but never printed from — a mistyped Spool the operator wants gone.

An Adjustment is a delta against one Spool's remaining. Without that Spool it is not a record of
anything: there is no remaining for it to correct and nothing that reads it.

## Decision

Deleting a Spool deletes its Adjustments in the same transaction. ADR-0003's append-only rule is
about the operator's lifetime with a live Spool: no Adjustment can be edited, and none can be
deleted on its own.

The delete guard stays as ADR-0011 wrote it — a Spool a Print has drawn from cannot be deleted —
and a re-weigh history does not add a second guard.

## Considered Options

Blocking deletion of any Spool that carries an Adjustment. Rejected: a re-weigh is the most likely
thing the operator does before noticing the Spool was entered wrongly, so this would make exactly
the Spools that need deleting undeletable, with no compensating record kept — the Spool would stay
on the shelf reading a wrong weight.

Keeping orphaned Adjustments. Rejected: rows referencing a Spool that no longer exists, which no
screen can render and no query needs.

## Consequences

- A re-weigh history is lost with its Spool. Accepted: deletion is confirmed first (ADR-0011), and
  the guard still protects every Spool whose grams are inside a Print's cost.
- `remaining >= 0` is unaffected: the Spool the invariant is about is gone.
- The `spools` foreign key on `spool_adjustments` needs no cascade at the schema level; the delete
  is explicit in the store, in one transaction.
