# ADR-0024: Long choice lists collapse into a filterable picker

**Status:** Accepted — 2026-09-05

## Context

A form choice row renders every option side by side on one line and `left`/`right` steps between
them. That was written for Filament Type, which has three options and always will.

The same row type also names Designs on the Print form, Spools on each Filament Usage row, and
Prints on the Sale form. Those grow with the ledger. At a few dozen Designs the line no longer fits
the terminal, and picking the twentieth option costs twenty keypresses with no way to search.

## Decision

A choice row with more than four options renders **collapsed** — the current choice alone,
truncated to the field width — and `enter` opens a **picker** over the form: a title, a filter
line, a `n of m` count, and a ten-row window of matches. Filtering is a case-insensitive substring
match against the whole rendered label, so `petg` narrows a Spool list by type and `black` by
colour without the picker knowing what a Spool is. `up`/`down` and `ctrl+p`/`ctrl+n` move,
`enter` commits, `esc` closes keeping the previous choice. Four or fewer options keep today's
inline row; a collapsed row still cycles with `left`/`right`.

The threshold is counted at runtime rather than declared per field, so a ledger with two Designs
behaves as it always did and grows into the picker on its own.

The picker lives in `internal/tui` alongside the form, and commits an index into the Choices it
was opened on — the same index `ChoiceIndex` already returns, so two rows that read the same still
name different records.

`ctrl+n` was the Print form's "add Filament Usage row". It becomes `ctrl+r`. A tab could instead
have guarded `ctrl+n` on whether the picker was open, but one chord meaning two things depending
on invisible state is a trap the next form to grow a long list would spring.

An inline scrolling list — the field expanding into a windowed list in place — needs no modal
state, but the rows below it move every time focus lands on a choice and there is still nowhere to
type a filter. A `Searchable` flag on `fieldSpec` is explicit, but it is a decision every call site
must remember and none can get right, since the list length is data, not code. Filtering the Print
form's Spool list to Capable Spools would cut the list at the source, but Capable is defined
against a Design's full estimate (ADR-0015) and a swap row may need only 40g, so it would hide
Spools that are genuinely usable.

## Consequences

- A tab acts on what `form.Update` refuses, rather than asking first: the form consumes every key
  while the picker is open and the `enter` that opens it, and hands back `esc` and the chords the
  tab reserved with `form.Reserve`. A tab that forgets to route through the form stops reaching the
  fields at all, which is visible immediately.
- The Prints and Sales tabs hide their cost breakdown and price guide while the picker is open —
  both describe a state the operator is halfway through changing.
- The picker is a renderer over strings the form already holds; no use case changes, and there is
  nothing here for a test at the `internal/app` seam to see.
