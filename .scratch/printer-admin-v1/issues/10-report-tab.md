# 10: Report tab

**What to build:** The operator finally finds out whether this makes money. The Report tab switches
between this month, last month, this year and all time, and for the chosen period shows revenue, the
cost of the copies actually sold in it, profit, margin percentage and copies sold.

Cost is matched to revenue, not to production (ADR-0006 as extended): a copy sold in March carries
the cost of the January Print it came from, so margin percentage genuinely answers "is my pricing
working". Production cost for the period is a separate, clearly labelled line, so a heavy printing
month is visible without being mistaken for a loss. Gifted, kept and scrapped copies get their own
line, which is what makes the gap between cost and revenue explainable rather than mysterious.
Designs are ranked by profit on the same matched basis.

Always visible regardless of period: break-even progress — revenue against the printer's purchase
price plus all filament ever bought plus energy spent — and a Inventory block showing unsold copies with
their cost value alongside spool grams with their value.

Break-even and period profit will disagree, and that is the point. One matches cost to what was
sold; the other counts every euro that has left the bank, including filament still on the shelf.
Both stay on screen and both are labelled, so the disagreement reads as two answers to two questions
rather than a bug.

No custom date ranges, no charts.

**Blocked by:** 09

**Status:** done

- [x] Four periods: this month, last month, this year, all time
- [x] Revenue, cost of copies sold, profit, margin percentage and copies sold for the period
- [x] A copy printed in one period and sold in another is costed into the period it sold in
- [x] Period boundaries are correct for a sale on the first and on the last day of a month
- [x] Production cost for the period on its own labelled line
- [x] Gifted, kept and scrapped copies on their own line
- [x] Designs ranked by profit on the matched basis
- [x] Break-even progress, counting the printer, all spool spend including filament not yet printed, and energy
- [x] Break-even and period profit are labelled as different measures
- [x] An Inventory block showing unsold copies and their cost value, and spool grams and their value
