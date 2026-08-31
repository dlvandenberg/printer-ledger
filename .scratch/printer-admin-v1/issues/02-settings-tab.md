# 02: Settings tab

**What to build:** The operator opens the Settings tab on a brand-new empty database and configures
the rates that drive every cost in the app: electricity price per kWh, a kWh-per-print-hour figure
for each Filament Type, machine hourly rate, printer purchase cost, default margin and minimum
margin.

Every per-type power rate is seeded with a plausible default at first launch, so energy is never
silently costed at zero before anything has been measured. Each rate is flagged `default` or
`measured`, and the flag is visible: editing a rate marks it measured. Ticket 07 uses that flag to
warn on the print form; nothing else reads it yet.

Settings live in the database, not a config file (ADR-0008). They must be reachable and editable
before any spool, design or print exists.

**Blocked by:** 01

**Status:** resolved

- [x] All six settings are editable and persist across a restart
- [x] A fresh database is seeded with sensible defaults, including one power rate per Filament Type
- [x] Each power rate shows whether it is a seeded default or an operator measurement
- [x] Editing a power rate marks it as measured
- [x] The Settings tab is reachable and usable on a completely empty database
- [x] Validation errors render inline; rejected submits keep what was typed

## Comments

### Decisions taken while implementing

- **`Percent` is hundredths of a percent** (`50%` → `5000`), an integer for the same reason money is
  (ADR-0002): margin multiplies a cost, so a float here would put a float in the cost math.
  `KwhPerHour` stays a `float64` — a physical rate, not money.
- **Money and percentage share one digit parser** (`parseHundredths`), so `22.000` and `12.755` are
  rejected by the same rule and `,` works as a decimal separator in both.
- **Seeded rates:** PLA 0.09, PLA+ 0.10, PETG 0.12 kWh/h, electricity €0.28, machine €0.35/h,
  default margin 50%, minimum margin 15%. Printer purchase cost seeds to **0**: it is
  the one figure the operator knows exactly and an invented number would silently distort
  Break-Even (ticket 10) while looking configured.
- **`measured` flips when the value moves.** `Settings.WithPowerRate` marks a rate measured only if
  the submitted rate differs from the stored one, so saving the Settings form without touching a
  rate leaves it flagged `default`, and a measured rate resubmitted unchanged stays measured.
- **Seeding runs on every open**, not inside the migration, so the defaults have one source
  (`domain.DefaultSettings`) rather than a copy in SQL, and a ledger created before a Filament Type
  existed gains that type's rate instead of costing its energy at zero (ADR-0010).
- **Settings are edited as one form** (`e`), not field by field, so a rejected submit keeps every
  typed value and reports every violation at once.

### Review follow-ups applied

- **`default` / `measured` is a rendering** of `PowerRate.measured`, not a domain type. A
  `RateSource` enum alongside the bool held one concept twice, and unlike `SpoolState` nothing
  computes with the two words — they only get printed. The edit form shows the flag in each rate's
  label, so it is visible at the moment the operator decides whether to overwrite.
- **ADR-0018** records `Percent` as integer hundredths; the `CONTEXT.md` Settings entry now matches
  the field names the code ships, and gains **Power Rate** and **Percent** entries.

### Left as they are

- **A rate resubmitted unchanged stays `default`.** Every rate is submitted on every save, so
  flipping on submit would mark all three measured the first time the operator changes the
  electricity price. The cost is that measuring a rate and confirming the seeded figure exactly
  leaves it flagged `default`.
- **`printerPurchaseCost` seeds to 0**, the one setting that ships unconfigured. Break-Even
  (ticket 10) should surface that rather than the Settings tab inventing a price.
- **`power_rates` rows are never deleted.** Retiring a Filament Type would leave an orphan row that
  `NewSettings` ignores.

### Currency dropped from the ticket

The operator ruled out the display-currency setting after the first pass: the ledger is euros in the
domain and in the database, and `€` is a TUI constant (as ticket 01 had it). Story 8 of the spec and
the `currency` line in ADR-0008's context no longer describe the build. Nothing converts, so there
was nothing for the setting to do but relabel amounts it could not restate.

The `settings` table therefore has six columns, not seven, and migration `0002_settings` was
corrected in place rather than followed by a drop — it had not shipped past this branch.
