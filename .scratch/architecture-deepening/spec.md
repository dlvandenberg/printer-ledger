# Spec: architecture deepening before tickets 02–10

Status: resolved

## Problem Statement

`printer-admin-v1` is one ticket in. Ticket 01 built the Spools tab as a tracer bullet and, as it
was asked to, set the conventions the next nine tickets copy. Three of those conventions do not
survive being copied nine times.

**Input handling sits past the only test seam.** ADR-0001 says the test seam is the application
layer and that the TUI must be able to do nothing `internal/app` cannot. Today the Spool form calls
`ParseCents`, `ParseGrams` and `ParseDate` itself and assembles its own validation error, so the
TUI decides which text is acceptable. Nothing tests it — a malformed price is rejected by code no
test reaches. Ticket 01's own comments recorded the gap ("unreachable through the app seam and the
TUI is untested") without closing it.

**The form is shallow.** Focus arithmetic, wrap-around, label styling, error placement and a
positional special case for the Filament Type chooser are hand-written in the Spool form. Its
interface is nearly as wide as its implementation, and five more forms are specified: Settings,
spool re-weigh, Design, Print, Sale.

**Tabs are branched on by identity.** The shell asks `active == tabSpools` in three separate
methods and holds the Spools tab as a named field. Five more tabs widen all three branches.

None of this is wrong at one tab and one form. All of it multiplies.

## Solution

Three deepenings, taken before the feature tickets that would replicate the shapes. Nothing the
operator can see changes: six tabs, Spools add and list, inline field errors, typed values
surviving a rejected submit.

1. **Spool intake takes raw text.** The command carries what was typed; the use case parses it and
   merges parse failures with the Spool's invariants into one validation error keyed by field. The
   parse helpers stay pure in `domain`; only the caller moves. Every input path becomes reachable
   from an application-layer test.
2. **One form module.** A tab declares field specs; the module owns focus, choice fields, layout and
   error display. Choice and text fields sit in one uniform list, so the index-0 special case
   disappears rather than moving. One interface, six call sites.
3. **A seam at the tab.** One interface — update, view, help, capturing-input — with six adapters.
   The shell keeps quit and tab-cycling and routes everything else to `tabs[active]`.

## Non-goals

- **Changing what use cases return.** Ticket 01 settled that they return their own view types rather
  than domain aggregates, and recorded why. Not reopened here.
- **Reopening ADR-0001.** All three deepenings sit inside the four layers and the inward-only rule;
  deepening 1 restores a rule the code is currently breaking.
- **Narrowing or removing `domain.Database`.** The interface has one adapter and widens by a method
  per query, which is worth watching, but it costs nothing yet and contradicts ADR-0001. Revisit if
  it approaches the full schema.
- **Testing the TUI.** The seam count stays at one. These changes move behaviour to the existing
  seam rather than adding a second.
- **New behaviour of any kind.** A deepening that changes what the operator sees is out of scope.

## Sequencing

- **01** and **03** are independent and can start immediately.
- **02** is blocked by **01**: the intake command changes what the form hands back, so building the
  form interface first means building it twice.
- All three should land before `printer-admin-v1` ticket 02, which introduces the second form and
  the second real tab.

## Acceptance

- Adding a spool from the Spools tab works end to end, unchanged, after each ticket.
- Application-layer tests cover malformed money, malformed grams, a malformed date, and a value that
  parses but violates an invariant.
- The TUI performs no parsing, constructs no validation error, and branches on no tab identity.
- `CGO_ENABLED=0 go build` still produces a working static binary.
