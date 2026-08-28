# 02: One form module, the Spool form as its first adapter

**What to build:** The Add Spool form behaves exactly as it does today — tab and shift-tab move
between fields and wrap around, left and right cycle the Filament Type, enter submits, esc cancels,
errors render beside the offending row, typed values survive a rejected submit — but it is now
declared as a list of fields rather than hand-written. The form module owns focus, rendering,
choice fields and error placement; a tab supplies field keys, labels, placeholders and kinds, and
reads back what was typed.

Today the Spool form hand-rolls focus arithmetic and special-cases the Filament Type chooser as
"index 0, not in the input slice", which every later form would copy. Five more forms are specified
— Settings, spool re-weigh, Design, Print, Sale. This is the change that stops the copying.

Choice fields and text fields must be uniform in the field list, so the index-0 special case
disappears rather than moving.

**Blocked by:** 01 — the intake command changes what the form hands back, so building the interface
first means building it twice.

**Status:** resolved

- [x] A form is declared as field specs; the module owns focus, wrapping, layout and error display
- [x] Choice fields and text fields sit in the same field list with no positional special case
- [x] The Spool form is expressed through the module with no behaviour change the operator can see
- [x] Errors keyed by field render beside the right row
- [x] The form module reports what was typed per field; it does not parse or validate
- [x] Adding a spool from the Spools tab still works end to end

## Comments

**2026-08-28 — implemented and reviewed.**

`internal/tui/form.go` is the module: a `form` is `newForm(title, []fieldSpec)`, and it owns focus,
wrap-around, label styling, choice cycling, layout and error placement. A field is a choice field
when its spec carries `choices` and a text field otherwise, so both sit in one `fields` slice and
focus is a plain index into it — the old `len(inputs)+1` arithmetic and the "index 0 is the chooser,
not in the input slice" special case are both gone rather than moved.

`internal/tui/spool_form.go` shrank to seven field specs plus `addSpoolCmd(form)`, which reads
`form.value(key)` per field. The form parses nothing and constructs no validation error; errors
arrive from the use case via `setErrors` and land on the row whose spec key matches, so the
`domain.Field*` constants are the only thing shared between the two sides.

Verified by driving the real TUI in a pty against a temp database:

- happy path — `right` chose PLA+, tab through the rows, `22,00` accepted; persisted
  `('PLA+','Bambu','Black',1000,210,2200,'2026-08-28')` and the list rendered `1000g / €22.00`
- rejected submit — `1.5g` and `22.000` rendered both inline errors beside their own rows, kept the
  typed text on screen, and persisted nothing (`count(*) = 0`)
- wrap-around — `shift+tab` from the chooser landed on Purchase date, then Purchase price

No tests added: the spec's non-goal "Testing the TUI — the seam count stays at one" holds, and the
existing application-layer tests still pass. `CGO_ENABLED=0 go build` still produces a working
static binary.

Review findings actioned: dropped an unused choice-preselect path and an unreachable empty-fields
guard (speculative generality), stopped `form.update`'s value receiver sharing its backing array
with the caller, and stopped locals shadowing the `form` and `field` type names.

Not actioned: `spoolsModel.help()` still hardcodes "left/right filament type" although the form now
owns choice cycling. The help line is operator-visible text and this ticket forbids visible change;
it belongs to the tab, and 03 is where the help line moves behind the tab seam.

**2026-08-28 — review follow-ups: the form is pointer-mutating and owns its help line.**

A `form` is now open state held as a `*form`, never a value. `newForm` returns `*form`, every
method takes a pointer receiver, and `update` is `func (f *form) update(tea.KeyMsg) tea.Cmd` —
it mutates in place and no longer returns a form for the caller to assign back. The
`slices.Clone` that the previous comment describes is gone with it.

That clone was half a rule. `setErrors` already mutated through the shared pointer while `update`
defended against exactly that, so one struct carried two opposite disciplines and `spoolsModel`
was the place they met. The rule written on the struct now: one tab owns the pointer from the
moment it opens the form until it drops it, and never copies it out. Tab models are values that
bubbletea replaces on every key, so holding one pointer is what makes a focus or choice move
stick.

`Model.routeKey` lost its `make`+`copy` for the same reason — it defended the tab slice against a
caller holding the pre-update `Model`, which bubbletea never does. Both sites now say plainly that
the write is meant to land.

`form.help()` builds the help line from the field specs: the shared keys, one `left/right <label>`
per choice field, then save/cancel/quit. `spoolsModel.help()` delegates to it while a form is open.
The rendered text is byte-identical to the string it replaces, so the "no visible change" rule
still holds — and this closes the loose end the previous comment deferred to 03, which shipped
without touching it. The second adapter gets its help line for free.

Also: `wrap(i, delta, n)` in `internal/tui/tui.go` replaces the third copy of `(i ± 1 + n) % n`
(tab switching, field focus, choice cycling), and `var _ tabModel = spoolsModel{}` /
`placeholderTab{}` put the interface check at the implementer instead of at the shell's tab list.

Still not actioned: `form.value` returns `""` for a key no field declares, so a typo in a future
adapter submits an empty field rather than failing loudly. Worth a decision before the second
adapter, not worth guessing at now.
