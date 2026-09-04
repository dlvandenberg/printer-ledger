# 11: Delete an unprinted design

**What to build:** The operator adds a Design, decides it was a mistake, and gets rid of it. They
press `d` on the Designs list, confirm, and it is gone.

A Design that has actually been printed refuses. Its Prints are a cost history and deleting the
Design would leave that history describing nothing, so the delete is blocked and says what is in the
way — `3 prints`. The count of copies is deliberately not named: it is a number the guard does not
act on, and naming it invites "so delete the copies first", which is not a thing.

The operator should not have to press `d` to find that out. The Designs list gains a trailing
`PRINTS` column, after `MARGIN`, so the answer is on screen before they try. The identity columns
stay adjacent; a count of prints is history about a Design rather than part of what it is.

"Has this been printed" gets one definition rather than one per caller — a **Design Ledger** read
model, the same shape the **Spool Ledger** and **Print Ledger** already have, so the number the list
displays and the number that blocks the delete cannot disagree. The guard lives in the domain: the
TUI must not decide whether a delete is allowed, and the foreign key must not be the thing that
says no.

The delete is confirmed like every other (ADR-0011), and prompts even when it will be refused — the
guard runs where it belongs and the error lands after.

No schema migration: the count is a join over tables that already exist.

**Blocked by:** None (can start immediately)

**Status:** done

- [x] A Design no Print references can be deleted from the Designs list and is gone from it afterwards
- [x] A Design a Print references cannot be deleted, and the error names how many Prints are in the way
- [x] Deleting that Design's Print first, then the Design, succeeds
- [x] Neither a blocked nor a successful delete disturbs the Prints that reference other Designs
- [x] The Designs list reports a print count per Design, trailing the margin column
- [x] Deletion is confirmed before it happens
- [x] `CONTEXT.md` gains the Design Ledger entry and the Design delete rule
