# ADR-0025: Plain aggregate reads beside the ledger read models

**Status:** Accepted — 2026-09-09

## Context

ADR-0003 made a Spool's `remaining` derived rather than stored, and the **Spool Ledger** read model
exists so that the derivation has one definition: the same expression answers what the Spools list
displays and whether a Filament Usage row would overdraw the Spool. **Design Ledger** and **Print
Ledger** followed the same reasoning for `printCount` and `available`.

The stores then answered *every* read with a ledger, which put the event-table subqueries in front
of callers that never look at the derived number. The prints and sales lists load Designs to map a
`designID` to a name, and Spools to name the color a Print came out in; the report loads Designs
for its ranking; `RecordPrint`, `EditPrint`, `PrintDraft`, `PreviewSale` and `readSaleView` each
count a Design's Prints to put its name on a view. Every one of those counts and sums is computed
and dropped.

The obvious fix — a second query shape — is also how the one-definition guarantee erodes. A
`Designs` that returned a print count nobody could see would be worse than the waste it saves,
because two queries would then both claim to answer "has this been printed".

## Decision

A store answers a read in two shapes, and which one a use case asks for is decided by what it needs
rather than by what is convenient:

- The **ledger** read model wherever the derived number is the point: a list the operator reads, an
  invariant, a delete guard.
- A **plain** read — the aggregate and nothing derived — wherever only identity is needed: a
  Design's name, a Spool's brand and color.

The plain read is what makes the split safe. Because it carries no derived field at all, it cannot
disagree with the ledger; there is nothing on it to disagree with. `printCount` stays reachable
only through `DesignLedger` and `remaining` only through `SpoolLedger`, so each keeps exactly one
definition.

A plain read exists only once a caller needs it. There is no by-id plain Spool read, because every
single-Spool read in the app — the detail screen, the re-weigh, the edit invariant, the delete
guard — is about `remaining`.

## Consequences

- Identity lookups stop running correlated subqueries over `filament_usages`, `spool_adjustments`
  and `prints`. Nothing the operator sees changes, which is why the existing use-case tests are
  what say so.
- Two query shapes and two scanners per aggregate in `internal/store`, whose column lists must stay
  in step. The plain query is the shared prefix of the ledger query, so they are written once.
- Domain methods that name a Print's material (`FilamentType`, `SpoolColors`) take `[]Spool`, so a
  caller cannot pass a ledger by habit and quietly re-earn the cost. `SpoolsOf` converts at the two
  cost-math call sites that already hold ledgers.
- Choosing wrongly is now possible: a use case that reads plain where it needed derived loses the
  number rather than getting a stale one, so the failure is a compile error, not a wrong figure.
