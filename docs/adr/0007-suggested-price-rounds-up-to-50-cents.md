# ADR-0007: Suggested price rounds up to the nearest €0.50

**Status:** Accepted — 2026-08-25. Amended 2026-08-26: Design screen only (ADR-0014).

## Context

`estimatedCost × (1 + margin)` produces prices like €3.54. Local in-person cash sales want round,
memorable numbers.

## Decision

```
suggested = ceilTo50c( estimatedCost × (1 + margin) )
```

Always **up**, never nearest — rounding down would undercharge below the intended margin.
€3.54 → €4.00; €3.02 → €3.50; €3.50 → €3.50 (exact multiples unchanged).

Margin is `Design.marginPct`, which is non-null and seeded from `Settings.defaultMarginPct`
at creation (ADR-0014).

Suggested price appears on the **Design screen only**, computed from its per-copy estimates
(ADR-0014). A Print shows cost and per-unit cost, never a price.

## Consequences

- Prices are cash-friendly and never below target margin.
- Effective margin exceeds the configured percentage, by up to €0.50 per unit — proportionally
  large on cheap items (€0.90 → €1.00 is a 10% bump). Accepted: the direction of error is safe.
- Suggested price is advisory. The actual `Sale.priceCents` is whatever was charged.
