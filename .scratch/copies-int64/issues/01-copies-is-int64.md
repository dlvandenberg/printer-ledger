# 01: Copies is an int64 and rejects non-positive input

**What to build:** `unit.Copies` is an `int64`, consistent with the other numeric units, and a copy count entered in a form can never be negative. A user who types `-2` copies, `-1` gifted, or `-3` scrapped gets the same "not a whole number of copies" field error they get for `abc`, instead of a negative count reaching the ledger. Counts round-trip through the database and the forms unchanged for every legal value, on a 32-bit build as well as a 64-bit one.

Zero stays a legal parse: gifted, kept and scrapped are genuinely `0` on most prints, and a user may type it explicitly. The "at least 1" rule belongs to Quantity alone and already lives in the use case and in the print's invariant — leave it there, do not move it into the parser.

**Blocked by:** None (can start immediately)

**Status:** ready-for-human

- [x] `unit.Copies` has an `int64` underlying type; parsing and formatting go through the 64-bit `strconv` calls, not `Atoi`/`Itoa`
- [x] A copy count above the 32-bit range parses and formats back to the same text
- [x] Negative copy text is rejected with `ErrMalformedCopies`, for Quantity and for the gifted/kept/scrapped fields
- [x] `"0"` still parses to zero, and creating a print with `0` gifted, kept and scrapped keeps working
- [x] A print created with quantity `0` or a negative quantity still fails on the Quantity field, with its existing message
- [x] Availability, sold/gifted/kept/scrapped accounting and cost-per-copy report the same values as before for existing prints
- [x] `make lint` and `make test` pass
