# 02: Settings tab

**What to build:** The operator opens the Settings tab on a brand-new empty database and configures
the rates that drive every cost in the app: electricity price per kWh, a kWh-per-print-hour figure
for each Filament Type, machine hourly rate, printer purchase cost, default margin, minimum margin,
and display currency.

Every per-type power rate is seeded with a plausible default at first launch, so energy is never
silently costed at zero before anything has been measured. Each rate is flagged `default` or
`measured`, and the flag is visible: editing a rate marks it measured. Ticket 07 uses that flag to
warn on the print form; nothing else reads it yet.

Settings live in the database, not a config file (ADR-0008). They must be reachable and editable
before any spool, design or print exists.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] All seven settings are editable and persist across a restart
- [ ] A fresh database is seeded with sensible defaults, including one power rate per Filament Type
- [ ] Each power rate shows whether it is a seeded default or an operator measurement
- [ ] Editing a power rate marks it as measured
- [ ] The Settings tab is reachable and usable on a completely empty database
- [ ] Validation errors render inline; rejected submits keep what was typed
