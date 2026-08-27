# ADR-0010: Filament types limited to PLA, PLA+, PETG; rates in Settings

**Status:** Accepted — 2026-08-25

## Context

Cost math needs a power draw figure per material, obtained by measuring once per type with a smart
plug. The printer is a Bambu Lab A1: an open-frame bedslinger with no heated chamber.

## Decision

`FilamentType` is an enum: `PLA`, `PLA+`, `PETG`.

Power rates are **not** part of the enum — they live in `Settings.kwhPerHourByType`, so a new
material is a settings row rather than a code change.

Excluded: ABS, ASA, PC, and Nylon all want an enclosed, heated chamber that the A1 does not have.
TPU *is* printable on the A1 (direct-drive extruder, short filament path) but is excluded from v1
for simplicity, not capability.

## Consequences

- The enum matches what can actually be printed, so the type picker has no dead options.
- Adding TPU later means one enum value plus one measured rate.
- A material with no measured rate must fail loudly rather than silently costing energy at zero.
- Buying an enclosed printer reopens this ADR along with the single-printer assumption.
