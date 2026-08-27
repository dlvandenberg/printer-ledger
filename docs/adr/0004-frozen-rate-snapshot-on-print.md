# ADR-0004: Prints freeze the rates used to cost them

**Status:** Accepted — 2026-08-25. Amended 2026-08-26: the spool gram price freezes too.

## Context

Electricity prices change. The machine hourly rate gets revised as maintenance costs become clear.
If cost were always computed from current Settings, every past print's cost — and therefore every
historical report — would silently change whenever a setting was edited.

## Decision

Each `Print` stores a snapshot of the rates in force when it was created: `kwhPriceCents`,
`kwhPerHour`, `machineRateCents`. Cost is computed from the print's own snapshot, never from
current Settings.

**Amendment (2026-08-26).** The original decision exempted spool cost per gram on the grounds
that a spool's purchase price is immutable history. It is not: ADR-0011 makes a Spool editable so
that a mistyped purchase price can be fixed. Filament is the largest cost line — 264c of the 471c
reference case — so a typo corrected in month nine would silently rewrite eight months of profit,
margin and break-even.

Each `FilamentUsage` row therefore stores `costPerGramCents`, copied from the Spool at creation. A
Print's cost is fully frozen: rates from its snapshot, gram prices from its usage rows. Correcting
a spool's price applies to prints recorded afterwards only, which is what the story asking for the
correction actually wanted.

## Consequences

- Historical reports are stable. Changing a setting affects new prints only.
- Reports can show genuine cost trends over time (electricity got more expensive).
- Settings edits do not retroactively fix a mis-measured rate. Correcting an old print means
  editing that print. Accepted — rare, and ADR-0011 allows it.
- Three redundant-looking columns per print row. Accepted: that redundancy *is* the history.
