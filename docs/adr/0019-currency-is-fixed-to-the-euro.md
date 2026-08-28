# ADR-0019: Currency is fixed to the euro, not a setting

**Status:** Accepted — 2026-08-28. Supersedes the `currency` entry in ADR-0008.

## Context

ADR-0008 listed `currency` among the global rates a singleton `Settings` aggregate would hold,
editable on the Settings tab. Building the tab showed the setting has no work to do: the ledger
records euros, spools are bought in euros, sales are priced in euros, and nothing converts between
currencies (ADR-0002 already rules out FX). A configurable symbol would only change the character
printed in front of an amount, while inviting the reading that the amounts themselves change with
it.

The live alternative was to keep it: an operator trading in another currency could relabel the whole
ledger in one place.

## Decision

The euro is fixed. `Settings` holds six values — electricity price per kWh, a kWh/hour power rate
per filament type, machine hourly rate, printer purchase cost, default margin, minimum margin — and
no currency. The `€` stays a `internal/tui` constant, alongside the other formatting the renderer
owns.

## Consequences

- Nothing to validate, store, migrate, or explain for a value that never varies.
- An operator on another currency edits one constant and rebuilds. Acceptable for a single-operator
  binary; the numbers are already euro-denominated, so a symbol swap alone would misrepresent them.
- Money formatting stays entirely in the renderer. The domain keeps returning `Cents`, which carries
  no currency at all.
