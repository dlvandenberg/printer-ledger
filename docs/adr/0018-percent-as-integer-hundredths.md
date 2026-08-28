# ADR-0018: Margins are integer hundredths of a percent

**Status:** Accepted — 2026-08-28

## Context

Settings hold a default margin and a minimum margin, and every Design holds its own margin. A margin
multiplies a cost in minor units to produce a **Suggested Price** and to test a **Sale** price
against a floor, so its representation decides whether the money pipeline stays integral.

The alternatives were a `float64` fraction (`0.5`), whole percent points as an `int`, and hundredths
of a percent as an `int64`. The operator plausibly types `50`, `50%`, `12.5` or `12,5`.

## Decision

`domain.Percent` is an `int64` counting hundredths of a percent: `50%` is `5000`. `ParsePercent`
accepts a `%` suffix, a comma decimal separator and at most two decimal places, through the same
digit reader as `ParseCents`. `FormatPercent` trims trailing zeros, so `5000` reads back as `50%`.

`KwhPerHour` remains a `float64`, unchanged: it is a measured physical rate, not money (ADR-0002).

## Consequences

- `cost × percent / 10000` stays in `int64`, so ADR-0002's "round once, at the end" survives pricing.
- Half-percent margins are expressible; thousandths are not, and are rejected at parse time.
- Money and percentages accept and reject the same digits, because they share one parser.
- A stored margin is an exact integer, so a Design's price cannot drift with float repair.
