# ADR-0002: Money as int64 cents, time as integer minutes, weight as centigrams

**Status:** Accepted — 2026-08-25 (amended 2026-09-04: weight)

## Context

Cost math multiplies and divides small amounts repeatedly: cost per gram (€22.00 / 1000g =
0.22c/g), energy (5.5h × 0.09 kWh/h × 28c), per-copy division. Floating point accumulates error
and produces amounts like €4.709999999. The slicer reports time as `5h 31m` and a filament
estimate as `85.59g`, so a weight rounded to the whole gram loses digits the operator typed and
throws off a twenty-copy plate by grams.

## Decision

- All money is `int64` in minor units (cents). No floats for money, anywhere, ever.
- Rounding happens once, at the end of a calculation — not per component.
- Durations are integer minutes. Input format is `HH:MM`, display format is `5h 31m`.
- Weight is `int64` in hundredths of a gram (centigrams), so `85.59g` survives entry, storage and
  display. A gram price stays hundredths of a cent per **whole** gram; the two places grams
  multiply a money rate carry the gram scale.
- `kwhPerHour` stays a float64: it is a measured physical rate, not money.

## Consequences

- No rounding drift; totals always equal the sum of their parts.
- `HH:MM` input matches what Bambu Studio shows, so no mental decimal conversion at entry time.
- Per-copy cost division truncates. Accepted: sub-cent per copy, and Suggested Price rounds up to
  €0.50 anyway (see ADR-0007).
- Currency is display-only; no multi-currency support and no FX.
- A third decimal of weight is malformed input, like a third decimal of money. A kitchen scale
  reading whole grams needs no special case.
