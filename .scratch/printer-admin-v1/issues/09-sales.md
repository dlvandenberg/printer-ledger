# 09: Sales

**What to build:** The operator sells something. They pick the Print the copy came from, see that
copy's real cost and the Design's suggested price side by side, and type what they actually charged.
Recording the sale decrements that Print's available count. Only Prints with copies still available
are offered, so the same object cannot be sold twice.

When the price falls below cost plus the minimum margin from Settings, the form **warns** — inline,
impossible to miss, and requiring no extra keystroke to proceed (ADR-0016). It never refuses. A
refusal would not prevent the bad sale, which already happened; it would only keep it out of the
ledger, and then the reporting would lie. Haggling and discounts get recorded truthfully.

Sales can be backdated, so a weekend market's takings go in on Monday. A sale can be edited to fix a
mistyped price, and deleting one returns its copy to available.

A Sale carries the Print it came from, the price and the date. Nothing else — no buyer, no shipping,
no fees.

**Blocked by:** 05, 08

**Status:** done

- [x] Record a Sale by picking a Print with copies available, entering a price and a date
- [x] The form shows the copy's actual cost and the Design's suggested price while entering
- [x] Recording a Sale decrements the Print's available count
- [x] Prints with no copies available are not offered
- [x] Selling the last available copy and then selling again from that Print fails
- [x] A price below cost plus the minimum margin warns inline and still records
- [x] Sales can be backdated, edited and deleted
- [x] Deleting a Sale returns its copy to available
