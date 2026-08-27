# 03: Spool detail and re-weigh

**What to build:** The operator opens a spool and sees how its remaining figure was reached —
initial grams, what prints have consumed, and every adjustment — so the number is trustworthy rather
than mysterious. From that screen they put the spool on a kitchen scale, type the total measured
weight, and the app derives true remaining from the spool's tare weight without any mental
arithmetic. A note records why it drifted: purge tower, a tangle, a partial spool, or a run that failed
and produced nothing.

Adjustments are append-only: never edited, never deleted (ADR-0003, ADR-0011). They are the
correction mechanism, and per ADR-0017 they are also how a zero-yield run is recorded, since such a
run is never a Print.

The `remaining >= 0` invariant lives in the domain, so the database can never hold a negative spool.
A spool at zero becomes `empty` and drops out of the way when picking filament, but stays visible in
history so past prints still explain their costs.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] Spool detail shows initial grams, consumption, each adjustment with its date and note, and the resulting remaining
- [ ] Re-weighing takes the total weight from the scale and derives remaining using the spool's tare
- [ ] An adjustment records measured grams, derived remaining, the delta against the previous figure, a date and a note
- [ ] Adjustments cannot be edited or deleted
- [ ] A sequence of several adjustments produces the correct remaining
- [ ] An adjustment that would drive remaining below zero is rejected
- [ ] A spool at zero remaining shows as empty and is excluded from filament pickers, but stays in the list and in history
