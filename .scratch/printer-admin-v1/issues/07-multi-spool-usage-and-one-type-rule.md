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

**Status:** resolved

- [x] A Print accepts two or more filament usage rows, and filament cost sums across them
- [x] Rows can be added and removed on the form; it opens with one row
- [x] The spool picker shows each spool's remaining grams
- [x] A total that overdraws one spool but splits correctly across two is accepted
- [x] Two spools of different Filament Types on one Print is rejected with a clear inline error
- [x] Each row is validated against its own spool's remaining
- [x] A warning shows when the Print's Filament Type has a seeded rather than measured power rate, without blocking submission

## Comments

Implemented. Overdraw is checked **per Spool, not per row** — grams are summed by `spoolID` before
the comparison, so two rows naming one Spool cannot each pass and together take it negative. That
contradicted ADR-0012's wording and CONTEXT.md, so it is recorded as ADR-0021 and the glossary
entry is sharpened. The shortfall is reported once per Spool, on the row that crosses it.

The one-type rule reuses `Print.FilamentType(ledgers)`, now exported: the first row naming a Spool
on the shelf sets the Print's type and every later row is validated against it, so the type the
energy rate is looked up by and the type the rule enforces cannot disagree.

`PreviewPrint` returns `PrintPreviewView{Cost, FilamentType, RateMeasured}` rather than a bare
`PrintCostView`. The seeded-rate warning needs a flag from a use case; `measured`/`default` stay
TUI labels.

Rows are added with `ctrl+n` and removed with `ctrl+x`, intercepted by the tab before the form sees
the key, since a printable key would land in the focused input. `form.Append`/`form.Truncate` grow
and shrink a group of trailing rows; the tab owns the row count, the form still owns focus.
