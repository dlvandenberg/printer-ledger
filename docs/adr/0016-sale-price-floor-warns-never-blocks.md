# ADR-0016: The sale-price floor warns, it never blocks

**Status:** Accepted — 2026-08-26

## Context

Selling below cost is the failure the app exists to prevent, so a sale price ought to clear cost
plus a small margin. But sales are local, in person and haggled, and the recorded price must be
whatever was actually charged.

## Decision

`Settings.minMarginPct` defines the floor. When a Sale is entered with
`priceCents < unitCost × (1 + minMarginPct)`, where `unitCost` is that of the **Print** the copy
came from, the form shows an inline warning. Recording proceeds with no extra keystroke.

## Consequences

- A refusal would not prevent a bad sale — it already happened — it would only keep it out of the
  ledger, and the reporting would then lie. Warning is the only behaviour compatible with a ledger.
- The Sale flow requires picking the source Print, not just the Design. This is needed anyway to
  decrement the right batch (ADR-0013).
- The floor compares against frozen cost (ADR-0004 as amended), so the warning cannot change
  retroactively.
