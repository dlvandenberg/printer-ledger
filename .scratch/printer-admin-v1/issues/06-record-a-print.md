# 06: Record a print

**What to build:** The operator records what they actually printed. They pick a Design, say how many
copies came off the plate, and the form opens prefilled with the Design's per-copy estimates
multiplied by that quantity — so a normal plate is nearly keystroke-free. They overwrite grams and
time with the totals Bambu Studio or the printer screen actually reported, pick the spool the
filament came from, and watch the cost breakdown update live: filament, energy, overhead, job cost,
and cost per copy. The date defaults to today and can be backdated.

Cost is frozen at creation and never moves again (ADR-0004 as amended). The Print stores the
electricity price, the power rate and the machine rate in force at the time, and the usage row
stores the spool's price per gram. Editing Settings next year, or fixing a spool's purchase-price
typo, must leave this Print's cost untouched.

Recording a Print consumes filament. The spool's remaining drops, and asking for more grams than the
spool holds is **rejected** inline naming the shortfall, not warned about (ADR-0012) — which is what
forces a real mid-print swap to be recorded honestly in ticket 07.

A run that produced nothing is not a Print at all (ADR-0017): quantity is at least one, and the lost
filament goes in as a re-weigh from ticket 03.

The heaviest ticket in this set: it carries the cost engine and its form. Pin the reference case
exactly — a PLA spool at €22.00 per 1000g, a print of 120g over 5h30m at quantity 2, €0.28 per kWh,
0.09 kWh/h, €0.35/h machine — giving 264c + 14c + 193c = 471c job cost and 236c per copy.

**Blocked by:** 03, 04

**Status:** resolved

- [x] Record a Print against a Design with quantity, `HH:MM` actual time, and one filament usage row
- [x] The form prefills grams and minutes as the Design's per-copy estimates times the quantity
- [x] The cost breakdown and per-copy cost update live as fields change
- [x] The Print stores its own electricity price, power rate and machine rate at creation
- [x] The usage row stores the spool's price per gram at creation
- [x] Editing Settings afterwards does not change the Print's cost; a Print recorded later uses the new rates
- [x] Editing the source spool's purchase price afterwards does not change the Print's cost; a later Print uses the corrected price
- [x] The spool's remaining drops by the grams used
- [x] Requesting more grams than the spool holds is rejected inline, naming what is actually left
- [x] Quantity must be at least one
- [x] Prints can be backdated
- [x] The reference case above is pinned exactly by a test
- [x] The Print screen shows no suggested price

## Comments

Implemented. `unit.CentsPerGram` (hundredths of a cent per gram) holds the frozen gram price: a
whole cent cannot express €22.00 / 1000g, and the reference case needs 2.2c/g exactly. Filament cost
sums in hundredths and rounds once.

`app.EditSpool` was added here as a dependency: the checklist item about correcting a spool's
purchase price is unobservable at the use-case seam without an edit path, and no ticket owns spool
editing (spec stories 21 and 22 are unassigned). It has no TUI yet — that remains open work for a
ticket of its own.

The one-type rule (ADR-0012 as narrowed) is not implemented: with one usage row it is unreachable.
`Print.filamentType` takes the first row that names a spool on the shelf, which is where 07 hangs
its rejection.
