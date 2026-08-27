# 04: Designs — add, edit, list

**What to build:** The operator adds a Design: a name, the slicer's filament-grams and print-time
estimates **for one copy**, the Filament Type it is normally printed in, and a margin. Print time is
typed as `HH:MM` so it can be copied straight from Bambu Studio, and displayed as `5h 31m`.
Estimates are editable after a re-slice.

Estimates are always per single copy — never a plate total. A plate of four is a Print of quantity
four, recorded in ticket 06; the Design's numbers do not change.

Margin is non-null and has no fallback rule (ADR-0014). A new Design copies `defaultMarginPct` from
Settings at creation and is freely edited afterwards, so raising the default later re-prices nothing
that already exists.

No costing and no suggested price yet — that is ticket 05. This ticket delivers the record and the
screen.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] Add, edit and list Designs, each with name, per-copy estimated grams, per-copy estimated minutes, default Filament Type and margin
- [ ] Print time is entered as `HH:MM`, stored as whole minutes, and displayed as `5h 31m`
- [ ] A new Design's margin is seeded from the Settings default without the operator typing it
- [ ] Changing the Settings default afterwards leaves existing Designs untouched
- [ ] A Design's margin can be edited independently
- [ ] Validation errors render inline; rejected submits keep what was typed
