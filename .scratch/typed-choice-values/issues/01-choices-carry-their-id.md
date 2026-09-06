# 01: A choice carries the record it names

**What to build:** A `choiceSpec` offers choices that carry the id of the record each one names, so a form hands back an id rather than a position into a list the caller must still hold. Selecting a Design, a Spool or a Print on a form produces the same command the use case receives today, with no index arithmetic between the picker and the command.

Today `choiceSpec.Choices` is `[]string` and identity comes back out through `form.ChoiceIndex(key)`, which the caller feeds to one of three near-identical helpers — `designIDAt`, `spoolIDAt`, `printIDAt` in `internal/tui` — each re-indexing the same slice of views the form was built from. The position is the only link between the label the operator picked and the record it meant, so every call site keeps the view slice alive purely to resolve it, and a slice that is reordered or refetched between build and submit resolves to the wrong record.

Filament Type rows have no id: their choice is the label itself, which the use case parses. Whatever shape a choice takes must carry those as well, without inventing an id for them.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] A choice offered to a form carries both what is rendered and what it names
- [ ] The form returns the named value directly; `ChoiceIndex` and the three `*IDAt` helpers in `internal/tui` are gone
- [ ] Filament Type rows, which name a value rather than a record, go through the same shape
- [ ] The picker still commits the choice it was opened on, so two rows that render alike still name different records (ADR-0024)
- [ ] Recording and editing a Print and a Sale write the same ids as before, including a Print whose Filament Usage rows name two different Spools
- [ ] Prefilled edit forms open on the record the row already names
- [ ] `make lint` and `make test` pass

## Comments

Split out of the `fieldSpec` variant refactor (commits `0bc7070`, `d085f2d`) to keep that diff reviewable. The two spec types landed there; this is the second half.
