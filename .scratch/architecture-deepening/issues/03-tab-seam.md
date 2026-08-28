# 03: A seam at the tab

**What to build:** The shell keeps only what is genuinely global — quit, and cycling tabs — and
routes every other key, the body and the help line to whichever tab is active, without naming it.
The operator sees no change: six tabs, Spools working, five placeholders, an open form still
swallowing keys except ctrl+c.

Today the shell asks `active == tabSpools` in three separate methods and holds the Spools tab as a
named field. Every tab in tickets 02–10 of `printer-admin-v1` widens all three branches. One
interface — update, view, help, and whether the tab is currently capturing input — turns those
branches into a lookup, and makes the five placeholders trivial adapters rather than an else-clause.

Six adapters land here, so the seam is real rather than hypothetical.

**Blocked by:** None (can start immediately; independent of 01 and 02).

**Status:** ready-for-agent

- [ ] The shell routes update, view and help to the active tab without branching on tab identity
- [ ] Each of the six tabs is an adapter satisfying one interface; the placeholders are the smallest possible ones
- [ ] ctrl+c quits even while a form is capturing input
- [ ] q, tab and shift-tab behave as they do today when no form is open
- [ ] Adding a spool from the Spools tab still works end to end
