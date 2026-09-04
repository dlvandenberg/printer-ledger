# 13: Plain list queries for name lookups

**What to build:** `Designs` and `Design` on the Design store, and `Spools` and `Spool` on the Spool
store, returning the aggregate alone, for the callers that only need identity.

Both stores currently answer every list read with a ledger. `SpoolLedgers` carries the used and
adjusted sums, and after issue 11 `DesignLedgers` carries the print count. Each sum is a correlated
subquery over the event tables, and the callers that want a name or a brand pay for them and throw
them away: `ListPrints` and `ListSales` load designs only to map `designID` to a name, and the
report loads designs for its per-design ranking. The single-row reads do the same: `RecordPrint`,
`EditPrint`, `PrintDraft`, `PreviewSale` and `readSaleView` each count a Design's Prints to put its
name on a view, which is why the by-id read needs the plain query as much as the list does.

The ledgers stay the answer wherever the derived number is the point — a list the operator reads,
an invariant, a delete guard. This ticket only gives the identity lookups a query that stops
counting.

The reason the ledgers are used everywhere today is that one definition of "what is a Spool's
remaining" beats one per caller, and a second query shape is how that guarantee erodes. So the
split has to be legible: a `Designs` that returned a print count nobody could see would be worse
than the waste it saves. The plain method returns the aggregate and nothing derived, which is what
makes it obviously not a second opinion on the derived number.

Nothing the operator sees changes. This is a read-path change behind the use cases, and the
existing tests are what say so.

**Blocked by:** 11

**Status:** ready-for-agent

- [ ] The Design store answers a plain `Designs` and the Spool store a plain `Spools`, neither running the event-table subqueries
- [ ] `ListPrints`, `ListSales` and the report read designs through the plain query
- [ ] The by-id identity reads — `RecordPrint`, `EditPrint`, `PrintDraft`, `PreviewSale`, `readSaleView` — do too, and the `designOf`/`designsByID`/`designList` shims in `internal/app/designs.go` are gone
- [ ] Every list the operator reads, every invariant and every delete guard still reads a ledger
- [ ] The existing use-case tests pass unchanged
- [ ] `prints(design_id)` is indexed, or the ticket records why the count does not need one
