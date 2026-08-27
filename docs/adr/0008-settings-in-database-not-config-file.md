# ADR-0008: Settings live in the database, edited in the TUI

**Status:** Accepted — 2026-08-25

## Context

Global rates are needed: electricity price, measured kWh/hour per filament type, machine hourly
rate, printer purchase cost, default margin, currency. These could be a hand-edited config file
(TOML/YAML) or rows in the database.

## Decision

A singleton `Settings` aggregate stored in the database, edited on a Settings tab in the TUI.
No config file.

## Consequences

- One source of truth. No "did I edit the file or the database" ambiguity, no file/DB drift.
- Settings are inputs to reporting and to print cost snapshots (ADR-0004), so they belong with the
  data they explain. Backing up the `.db` backs up everything.
- Bootstrapping needs sensible defaults on first run, and the Settings screen has to be reachable
  before any other data exists.
- No editing settings without launching the app. Fine for a single-operator TUI.
