# 07: Multi-spool usage and the one-type rule

**What to build:** A spool runs out halfway through a job. The operator records that as **one**
Print with two filament usage rows — 40g from the spool that emptied, 80g from its replacement —
rather than two fictional prints that would split the copies and wreck the per-copy cost. The same
mechanism covers a two-colour swap. While picking, each spool's remaining grams are visible, so it
is obvious whether one will make it through the job.

Mixing Filament Types on a single Print is **rejected** (ADR-0012 as narrowed). Energy is a single
rate lookup, and a Print spanning PLA and PETG has no defined rate; the error should say so plainly.

Also here: when the Print's Filament Type still carries a seeded rather than measured power rate,
the form warns that the energy figure is a placeholder. It warns, it does not block.

**Blocked by:** 06

**Status:** ready-for-agent

- [ ] A Print accepts two or more filament usage rows, and filament cost sums across them
- [ ] Rows can be added and removed on the form; it opens with one row
- [ ] The spool picker shows each spool's remaining grams
- [ ] A total that overdraws one spool but splits correctly across two is accepted
- [ ] Two spools of different Filament Types on one Print is rejected with a clear inline error
- [ ] Each row is validated against its own spool's remaining
- [ ] A warning shows when the Print's Filament Type has a seeded rather than measured power rate, without blocking submission
