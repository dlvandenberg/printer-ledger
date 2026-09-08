# 02: Type the weight the slicer reports

**What to build:** the operator types the estimate Bambu Studio shows, digit for digit. `85.59` is
accepted wherever a weight is entered — a design's estimated grams, a spool's initial and tare
weight, a reweigh's measured weight, a print's usage row — and reaches the ledger as 85.59g rather
than 86g. A third decimal, `85.594`, is refused as malformed: grams reject over-precision the way
money and percentages already do, rather than silently swallowing a digit.

One parser, one rule, every weight field. A kitchen scale reads whole grams and will keep producing
`1000`, which is fine and needs no special case; the domain has no business asserting what
resolution the operator's scale has, and a per-field decimal ban would be a validation rule with no
invariant behind it.

Weights display trimmed, the way percentages and gram prices already do: `85.59g`, `85.5g`, `85g` —
never `85.00g`. Formatting and parsing stay a lossless round trip. Because the estimate is now
exact, a multi-copy print prefills the true multiple of the design's estimate instead of the
multiple of a rounded one: twenty copies of an 85.59g design prefills 1711.8g, not 1720g.

The unit decision is recorded by amending ADR-0002 in place — grams-as-hundredths extends the same
integer-units decision that already covers money and minutes to a third unit, and splitting it into
its own ADR would fragment the one place to look for "what is an integer here, and why". The
conventions doc currently forbids editing an accepted ADR at all; it gains a carve-out saying that
extending an accepted decision is an amendment while reversing one is a supersede. The conventions
doc also uses grams as its worked example of a named unit type, quoting the old malformed-weight
message, so that example moves with the change. `CONTEXT.md` glossary entries for the weight terms
say that entry accepts two decimals.

The minimum stays "more than 0g", so 0.01g is a legal weight. The domain has no opinion on what is
too small to print, and any floor above zero would be a number with nothing to defend it.

**Blocked by:** 01 (Grams are stored as hundredths of a gram)

**Status:** resolved

- [x] `85.59` is accepted for a design estimate, a spool's initial and tare weight, a reweigh and a
      usage row, and the value that comes back is 85.59g
- [x] `85.594` is rejected on the field that was typed, with the single malformed-weight sentinel
- [x] A comma decimal separator and a `g` suffix are accepted, as for the other units
- [x] Weights display with trailing zeros trimmed, and every displayed weight parses back to itself
- [x] A print recorded for twenty copies of an 85.59g design prefills 1711.8g
- [x] A weight of 0.01g is legal; zero and negative weights are still refused with their existing
      messages
- [x] A quote and a print cost built from a fractional estimate are correct to the cent, rounding
      once at the end
- [x] ADR-0002 covers grams, the conventions doc distinguishes amending an ADR from superseding one
      and its unit example matches the code, and the glossary says entry takes two decimals
- [x] `make lint` and `make test` pass
