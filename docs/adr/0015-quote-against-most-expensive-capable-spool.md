# ADR-0015: A Design quotes against the most expensive Capable Spool

**Status:** Accepted — 2026-08-26

## Context

`costPerGram` is defined strictly per Spool, never as a per-type average, because a cheap spool
genuinely makes a cheaper print. But a Design shows an estimated cost per **Filament Type**, and
you may hold four PLA spools bought at four different prices. One gram price has to be picked, and
a Design references no Spool.

Separately, `active` (`remaining > 0`) is too loose to hang a quote on: a spool with 3g left is
active, and if it is also the most expensive it would set every quote for filament that cannot
print anything.

## Decision

Each Filament Type row on a Design is costed at the `costPerGram` of the **most expensive Capable
Spool** of that type, where capable means same type and `remaining >= Design.estimatedGrams`. The
row names the spool it used.

Fallbacks: no Capable Spool → the most expensive Spool of that type ever bought, labelled a
reference price. Type never bought → no row.

The headline Suggested Price comes from the `defaultFilamentType` row; other rows show cost only.

## Considered Options

- **Cheapest capable spool.** Rejected: quoting off the cheap spool and printing off the expensive
  one eats the difference on a price already spoken aloud.
- **Average of capable spools.** Rejected: an average is a price no spool actually has, and it is
  wrong in the unsafe direction half the time.
- **A nominal €/g per type in Settings.** Rejected: a second number to maintain that silently
  drifts from what was actually paid.
- **A fixed usable-grams floor** instead of capability. Rejected: a threshold to tune, where
  `remaining >= estimatedGrams` needs no tuning and asks the question that matters.

## Consequences

- The direction of error is safe, matching the round-up rule in ADR-0007.
- A quote moves when stock changes: buying an expensive spool, or emptying the one being quoted
  against, changes the number. This is intended — it is a live quote, not history.
- Editing a Design's `estimatedGrams` can change its quote, because the candidate spool set is
  defined relative to that figure.
- No manual `archived` flag on Spool. A spool that is gone gets re-weighed to zero.
