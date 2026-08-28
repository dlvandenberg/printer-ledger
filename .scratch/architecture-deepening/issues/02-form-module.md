# 02: One form module, the Spool form as its first adapter

**What to build:** The Add Spool form behaves exactly as it does today — tab and shift-tab move
between fields and wrap around, left and right cycle the Filament Type, enter submits, esc cancels,
errors render beside the offending row, typed values survive a rejected submit — but it is now
declared as a list of fields rather than hand-written. The form module owns focus, rendering,
choice fields and error placement; a tab supplies field keys, labels, placeholders and kinds, and
reads back what was typed.

Today the Spool form hand-rolls focus arithmetic and special-cases the Filament Type chooser as
"index 0, not in the input slice", which every later form would copy. Five more forms are specified
— Settings, spool re-weigh, Design, Print, Sale. This is the change that stops the copying.

Choice fields and text fields must be uniform in the field list, so the index-0 special case
disappears rather than moving.

**Blocked by:** 01 — the intake command changes what the form hands back, so building the interface
first means building it twice.

**Status:** ready-for-agent

- [ ] A form is declared as field specs; the module owns focus, wrapping, layout and error display
- [ ] Choice fields and text fields sit in the same field list with no positional special case
- [ ] The Spool form is expressed through the module with no behaviour change the operator can see
- [ ] Errors keyed by field render beside the right row
- [ ] The form module reports what was typed per field; it does not parse or validate
- [ ] Adding a spool from the Spools tab still works end to end
