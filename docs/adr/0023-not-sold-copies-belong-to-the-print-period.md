# ADR-0023: Gifted, kept and scrapped copies are reported in the period they were printed

**Status:** Accepted — 2026-09-04

## Context

The report matches cost to revenue: a copy sold in March carries the cost of the January Print it
came from, so margin percentage answers whether the pricing works (ADR-0006). Copies that were
gifted, kept or scrapped never sell, so there is no sale date to match them to, and they still have
to appear somewhere — they are what makes the gap between production cost and revenue explainable
rather than mysterious.

Two dates could scope them: the Print's date, or nothing at all (an all-time line like Break-Even).

## Decision

Gifted, kept and scrapped copies are counted in the period their **Print** falls in, alongside the
production cost line, and both lines are labelled as printed-in-this-period rather than sold-in-it.

## Consequences

- The two lines under one heading share a basis, so production cost and the copies that will never
  repay it are read together.
- A copy printed in January and gifted in March counts against January. There is no gifting date in
  the ledger, and adding one would record a fact the operator has no reason to type.
- The line moves when a Print is backdated or its counts are corrected, which is the same behaviour
  as the production cost line beside it.
