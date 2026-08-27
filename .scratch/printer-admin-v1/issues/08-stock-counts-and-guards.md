# 08: Stock counts, edit and delete guards

**What to build:** The operator says what became of a Print's copies. Some were given away, some
kept for themselves, some ruined. They record those counts on the Print and see how many are still
available to sell — what is actually in the drawer.

Stock is counted, not individually tracked (ADR-0013): copies of one Design from one Print are
interchangeable, so there is nothing to tell them apart. Available is quantity minus sold, gifted,
kept and scrapped, and the `available >= 0` invariant is the single rule that keeps the ledger
honest. It is the same rule that later makes overselling impossible in ticket 09, and the same rule
that stops quantity being edited below what has already been accounted for.

Gifted, kept and scrapped copies keep their share of job cost — the money was really spent — but
never contribute revenue. A partly failed plate is recorded by raising the scrapped count; there is
no separate failure concept.

Editing a Print in place recomputes spool remaining for free, since remaining is derived, and
re-validates the invariant (ADR-0011). Deletion is guarded: a Print with any accounted-for copy
cannot be deleted, and neither can a Spool that a Print has drawn from. Every delete is confirmed.

**Blocked by:** 06

**Status:** ready-for-agent

- [ ] Gifted, kept and scrapped counts are recorded per Print and persist
- [ ] Available is shown per Print and derives from quantity minus everything accounted for
- [ ] Accounting for more copies than the Print produced is rejected
- [ ] A Print can be edited in place; changing its grams recomputes the spool's remaining and re-validates the invariant
- [ ] Quantity cannot shrink below the copies already sold, gifted, kept or scrapped
- [ ] A Print with any accounted-for copy cannot be deleted, and the error says which copies are in the way
- [ ] A Spool that a Print has drawn from cannot be deleted
- [ ] Deletion is confirmed before it happens
- [ ] Per-copy cost divides across all copies including scrapped ones
