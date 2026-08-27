# ADR-0006: Machine overhead is an hourly cost; break-even is cash-based

**Status:** Accepted — 2026-08-25

## Context

The operator wants a suggested price whose margin "covers machine overhead", and separately wants
to know when the printer has paid for itself. Two ways to model overhead: fold it into a fat margin
percentage, or treat it as an explicit hourly cost.

## Decision

Both questions are answered, with two different numbers, both labelled:

- **Per-print cost** includes `machineHourlyRateCents × hours` — depreciation, maintenance, and a
  failure buffer, amortised per hour. Margin on top of that is therefore real earnings, and true
  break-even per print is visible.
- **Break-even** is cash-based and all-time:
  `revenue − (printerPurchaseCostCents + all spool spend + energy spend)`.
  It counts spools bought but not yet printed, because that cash is really gone.

## Consequences

- Two profit-ish numbers on screen that will not agree. Mitigated by labelling them explicitly;
  conflating them would hide either true unit economics or true cash position.
- `machineHourlyRateCents` is an operator estimate, so per-print cost is only as good as it.
- Break-even is pessimistic right after a filament order, and jumps on each sale. That is accurate:
  it reflects the bank account, not accrual accounting.
