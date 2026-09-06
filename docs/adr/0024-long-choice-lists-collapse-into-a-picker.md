# ADR-0024: Long choice lists collapse into a filterable picker

**Status:** Accepted — 2026-09-05 (amended 2026-09-06: the threshold is width, not a count)

## Context

A form choice row renders every option side by side on one line and `left`/`right` steps between
them. That was written for Filament Type, which has three options and always will.

The same row type also names Designs on the Print form, Spools on each Filament Usage row, and
Prints on the Sale form. Those grow with the ledger. At a few dozen Designs the line no longer fits
the terminal, and picking the twentieth option costs twenty keypresses with no way to search.

Their labels are long as well as many. A Spool reads `PETG Bambu Black 640g` and a sellable Print
carries a Design, a material, a date, a cost and a count, so two options already run past the
terminal — long before any count of them looks large.

## Decision

A choice row whose options do not fit the width a row is allowed renders **collapsed** — the
current choice alone, truncated to that same width — and `enter` opens a **picker** over the form:
a title, a filter line, a `n of m` count, and a ten-row window of matches. Filtering is a
case-insensitive substring match against the whole rendered label, so `petg` narrows a Spool list
by type and `black` by colour without the picker knowing what a Spool is. `up`/`down` and
`ctrl+p`/`ctrl+n` move, `enter` commits, `esc` closes keeping the previous choice. Options that fit
keep the inline row; a collapsed row still cycles with `left`/`right`.

What is measured at runtime is the options: their rendered width side by side, rather than a count
declared per field, so a ledger with two Designs behaves as it always did and grows into the picker
on its own. What they are measured against is fixed — twice the column a text row is given, 48
terminal columns — and does not follow the terminal. The form is a fixed-column layout already: a
label column and a field column, both constants. A budget read from the terminal would make whether
a row collapses depend on how the window was dragged, and the form has no width to read anyway,
since nothing propagates `WindowSizeMsg` past the top-level model.

A count threshold was the first rule tried and it measures the wrong thing: it leaves two Spool
labels — 52 columns — on one inline row, and collapses five Designs named `Vase` and `Cat` that had
room to spare.

The picker lives in `internal/tui` alongside the form, and commits an index into the Choices it
was opened on — the same index `ChoiceIndex` already returns, so two rows that read the same still
name different records.

`ctrl+n` was the Print form's "add Filament Usage row". It becomes `ctrl+r`. A tab could instead
have guarded `ctrl+n` on whether the picker was open, but one chord meaning two things depending
on invisible state is a trap the next form to grow a long list would spring.

An inline scrolling list — the field expanding into a windowed list in place — needs no modal
state, but the rows below it move every time focus lands on a choice and there is still nowhere to
type a filter. A `Searchable` flag on the spec is explicit, but it is a decision every call site
must remember and none can get right, since both the length of a list and the length of its labels
are data, not code. Filtering the Print form's Spool list to Capable Spools would cut the list at
the source, but Capable is defined
against a Design's full estimate (ADR-0015) and a swap row may need only 40g, so it would hide
Spools that are genuinely usable.

## Consequences

- A tab acts on what `form.Update` refuses, rather than asking first: the form consumes every key
  while the picker is open and the `enter` that opens it, and hands back `esc` and the chords the
  tab reserved with `form.Reserve`. A tab that forgets to route through the form stops reaching the
  fields at all, which is visible immediately.
- The Prints and Sales tabs hide their cost breakdown and price guide while the picker is open —
  both describe a state the operator is halfway through changing.
- Whether a row collapses changes as the ledger does: renaming a Design to something long can
  collapse a row that was inline, which is the rule working rather than a surprise.
- The picker is a renderer over strings the form already holds; no use case changes, and there is
  nothing here for a test at the `internal/app` seam to see.
