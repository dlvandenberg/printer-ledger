# 01: Spool intake takes raw text, the use case parses it

**What to build:** The operator types a purchase price, filament grams, tare grams and a purchase
date into the Add Spool form and gets the same inline field errors as today — `22,00` accepted,
`22.000` rejected against the price row, `1.5g` rejected against the grams row, `2026-13-01`
rejected against the date row, typed values surviving a rejected submit. What changes is where the
decision is made: the TUI hands over what was typed, and the Spool intake use case parses it,
merges parse failures with the Spool's own invariants into one validation error keyed by field, and
returns either that or the created Spool.

The parse helpers stay in the domain — `ParseCents`, `ParseGrams`, `ParseDate` are pure and belong
there. What moves is the caller. Today the TUI calls them and assembles its own validation error,
which is behaviour sitting past the one test seam, so no test covers a malformed price. After this
ticket every input path is reachable from an application-layer test.

Ticket 01 of `printer-admin-v1` recorded the opposite arrangement in its comments ("parsing lives in
`domain` … unreachable through the app seam and the TUI is untested"). That comment named the gap;
this ticket closes it. The decision that use cases take commands and return view types is unchanged
— only the field types on the command change.

This also restores the ADR-0001 rule that the TUI can do nothing the application layer cannot.
Today it can: it can reject input the use case would have accepted, and vice versa.

Do this before the Settings, Design, Print and Sale forms exist, or the same split has to be
unpicked five more times.

**Blocked by:** None (can start immediately).

**Status:** ready-for-agent

- [x] The Spool intake command carries what the operator typed, not pre-parsed domain values
- [x] Parse failures and Spool invariants arrive as one validation error keyed by the same field names the form renders against
- [x] Application-layer tests cover a malformed amount, malformed grams, a malformed date, and a value that parses but violates an invariant
- [x] A rejected submit persists nothing and keeps what was typed on screen
- [x] The TUI performs no parsing and constructs no validation error
- [x] Adding a spool from the Spools tab still works end to end
