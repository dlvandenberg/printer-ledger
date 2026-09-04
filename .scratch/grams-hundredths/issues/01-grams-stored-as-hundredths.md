# 01: Grams are stored as hundredths of a gram

**What to build:** nothing the operator can see. `unit.Grams` becomes an `int64` count of
hundredths of a gram — a centigram — so that a later ticket can accept the two decimals Bambu
Studio reports. Entry and display are unchanged by this ticket: `85.59` is still rejected with
"not a whole number of grams", a 1kg spool still reads `1000g`, and every quote, print cost, spool
value, reweigh and report comes out to the identical cent it does today.

`CentsPerGram` keeps its meaning: hundredths of a cent per **whole** gram, displayed as `2.2`. It
is frozen on every usage row, and a per-centigram rate would round 2.2c/g down to 2c/g and destroy
the third digit the type exists to hold. The two places where grams multiply a money rate carry the
extra factor instead: the spool's gram price scales its numerator, and a print's filament cost
divides by the gram scale as well as the gram-price scale. Rounding still happens once, at the end
of the sum, per ADR-0002.

Every other weight calculation is either pure gram arithmetic or a ratio with grams on both sides —
a spool's value and quoted value, the comparison that picks the priciest capable spool, remaining
value, inventory value, the per-spool overdraw check, the capable-spool stock test, the initial-grams
floor, derived remaining from a reweigh, and estimated grams times copies. Those are scale-invariant
and must not be touched. Read every one to confirm it, because `Grams` stays an `int64`: the compiler
will not flag a site that needed the factor and did not get it. A missed site shows up as a wrong
cent in a test, which is why this ticket changes no behaviour — the suite is the oracle.

The database has no production data. Rescale the columns in place in the ordered migration list
rather than adding a conversion migration, and rename them so the unit is unmissable in SQL and any
stale development database fails loudly on a missing column instead of silently reading 85g as
0.85g. The frozen gram-price column keeps its name and meaning. Delete the local development ledger
so it is recreated.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [x] `unit.Grams` counts hundredths of a gram, with a named scale constant beside the type
- [x] Parsing `"1000"` yields one thousand grams and formatting it returns `1000g`; a fractional
      weight is still rejected with the existing `ErrMalformedGrams` message
- [x] A design's suggested price, a print's filament cost and cost per copy, a spool's remaining
      value and the report's inventory value are unchanged to the cent for every existing test case
- [x] A spool's gram price still reports `2.2` for a €22.00 spool of 1000g
- [x] Weight columns are stored as centigrams and named for it; the frozen gram-price column is
      unchanged; no conversion migration exists
- [x] Spool remaining, the per-spool overdraw refusal and the capable-spool stock test behave as
      before, including at the boundary where remaining exactly equals the weight requested
- [x] A weight round-trips through the database and back to the same displayed text
- [x] `make lint` and `make test` pass
