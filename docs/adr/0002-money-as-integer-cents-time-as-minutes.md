# ADR-0002: Money as int64 cents, time as integer minutes

**Status:** Accepted — 2026-08-25

## Context

Cost math multiplies and divides small amounts repeatedly: cost per gram (€22.00 / 1000g =
0.22c/g), energy (5.5h × 0.09 kWh/h × 28c), per-copy division. Floating point accumulates error
and produces amounts like €4.709999999. The slicer reports time as `5h 31m`.

## Decision

- All money is `int64` in minor units (cents). No floats for money, anywhere, ever.
- Rounding happens once, at the end of a calculation — not per component.
- Durations are integer minutes. Input format is `HH:MM`, display format is `5h 31m`.
- `kwhPerHour` stays a float64: it is a measured physical rate, not money.

## Consequences

- No rounding drift; totals always equal the sum of their parts.
- `HH:MM` input matches what Bambu Studio shows, so no mental decimal conversion at entry time.
- Per-copy cost division truncates. Accepted: sub-cent per copy, and Suggested Price rounds up to
  €0.50 anyway (see ADR-0007).
- Currency is display-only; no multi-currency support and no FX.
