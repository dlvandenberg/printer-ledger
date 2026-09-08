# 12: Print detail screen

**What to build:** The operator presses `enter` on a row of the Prints list and sees the whole
print: what it was, what became of its copies, what it cost, and what it was costed at.

Five things are on screen. The Design, date, quantity and time. The **Stock** breakdown — quantity,
sold, gifted, kept, scrapped, available — which the list only summarises as one available figure.
The cost split into filament, energy and overhead, with the job cost and the per-copy cost. The
**Filament Usage** rows, each naming its Spool and **Filament Type**, the grams drawn and the gram
price frozen on that row, so a mid-print spool swap is legible as the two draws it really was. And
the rate snapshot — `kwhPrice`, `kwhPerHour`, `machineRate` — frozen when the Print was recorded.

The snapshot is the reason the screen is worth having. It is the only place those three numbers are
ever visible, and seeing them is what turns "a past print's cost never moves" from a claim into
something the operator can check.

Usage rows show grams and the frozen gram price, not a per-row cost. Filament cost sums in
hundredths of a cent across the rows and rounds once at the end, so per-row rounded amounts would
not reliably add up to the filament total shown above them.

Every number comes from a use case; the screen computes nothing. Nothing on it is detail-only, so it
reads the same view the list rows are built from.

From the detail, `e` edits the Print and returns to a re-read detail. `esc` goes back to the list.
There is no refresh key: a Print's cost is frozen and nothing but an edit to the Print itself can
move what is shown.

**Blocked by:** None (can start immediately)

**Status:** resolved

- [x] `enter` on a Prints row opens that Print's detail, `esc` returns to the list
- [x] The detail shows the stock counts, the cost breakdown, the usage rows and the frozen rate snapshot
- [x] A print drawing from two Spools shows one usage row per draw, each naming its Spool, type, grams and frozen gram price
- [x] Asking for a Print that does not exist reports not found rather than an empty screen
- [x] `e` from the detail edits the Print and comes back to the updated detail
- [x] `CONTEXT.md` gains the Print Ledger entry

## Comments

The **Print Ledger** entry was already in `CONTEXT.md`, added by `80e2927` alongside this issue, and
the **Print** entry already records the rate snapshot. Nothing needed changing for this screen.
