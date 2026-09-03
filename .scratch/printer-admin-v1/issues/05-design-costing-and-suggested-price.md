# 05: Design costing and suggested price

**What to build:** The operator opens a Design and sees what one copy would cost in each Filament
Type they stock — filament, energy and overhead — and one headline suggested price to quote someone
before the thing has ever been printed.

Each cost row is priced against the **most expensive Capable Spool** of that type, meaning a spool
of that type with enough remaining to actually complete the job, and the row names the spool it
used (ADR-0015). Quoting off the cheap spool and then printing off the expensive one loses money on
a price already spoken aloud, so the direction of error is deliberately safe.

Fallbacks matter as much as the happy path: no Capable Spool of that type falls back to the most
expensive spool of that type ever bought, clearly labelled a reference price rather than a live one;
a type never bought produces no row at all.

The headline suggested price comes from the Design's default Filament Type row, with the Design's
own margin, rounded **up** to the nearest €0.50 (ADR-0007). Other rows show cost only, so the effect
of switching material is visible without implying a second price.

This is a live quote read from current Settings — it moves when rates change, when estimated grams
change, and when the spool it quotes against runs out. That asymmetry against a Print's frozen cost
is deliberate.

Pin the reference case exactly: a PLA spool at €22.00 per 1000g, a design of 60g and 2h45m per copy,
€0.28 per kWh, 0.09 kWh/h for PLA, €0.35/h machine rate and 50% margin gives 132c + 7c + 96c = 235c
and a suggested price of €4.00.

**Blocked by:** 03, 04

**Status:** done

- [x] A Design shows one cost row per Filament Type, broken into filament, energy and overhead
- [x] Each row uses the most expensive spool of that type with remaining at or above the Design's estimated grams, and names that spool
- [x] A spool that is active but too small for the job is excluded from the choice
- [x] With no capable spool, the row falls back to the most expensive spool of that type ever bought and is labelled a reference price
- [x] A Filament Type never purchased produces no row
- [x] One headline suggested price, from the default Filament Type row, using the Design's own margin
- [x] Suggested price rounds up to the nearest €0.50, leaving exact multiples unchanged
- [x] The reference case above is pinned exactly by a test
