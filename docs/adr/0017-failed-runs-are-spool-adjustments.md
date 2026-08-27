# ADR-0017: A run that yields nothing is a Spool Adjustment, not a Print

**Status:** Accepted — 2026-08-26

## Context

A print that fails at layer three consumes real filament and real machine time but produces no
objects. Under ADR-0005 failures were recorded by scrapping units, which for a zero-yield run
means creating N records for objects that never existed. Allowing `quantity: 0` instead makes
`unitCost = jobCost / 0` undefined and puts a zero-division path into every consumer of unit cost.

## Decision

A run that yields nothing is not recorded as a `Print`. The lost filament is recorded as a **Spool
Adjustment** — a re-weigh with a note. `Print.quantity` is therefore always at least 1.

A *partly* failed plate is still a Print: `scrappedCount` records the ruined copies.

## Considered Options

Allowing `quantity: 0` Prints, guarding the division and showing "—" for unit cost. Rejected for
the branch it adds everywhere unit cost is consumed, against a simplicity goal.

## Consequences

- The machine hours and energy of a failed run are never recorded. Break-even's energy spend
  undercounts by every failure.
- The app cannot report a failure rate. `machineHourlyRateCents` carries a failure buffer for
  exactly this reason, but that buffer is now a guess that can never be checked against reality.
  If that becomes intolerable, the cheap first step is a `reason` field on Spool Adjustment, which
  buys a failure count but still no hours.
- Nothing needs to divide by a zero quantity.
