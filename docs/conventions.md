# Conventions

How code in this repo is written. `CLAUDE.md` says which layer owns what; this says what the code
in those layers looks like. `CONTEXT.md` owns the vocabulary — read it before naming anything.

Rule of thumb: a new file should be indistinguishable from the existing files in its package. When
this document and the surrounding code disagree, the code wins and this document is stale — fix it.

## Naming

- Domain terms come from the `CONTEXT.md` glossary, spelled exactly: `Spool`, `SpoolLedger`,
  `FilamentType`, `FilamentUsage`, `SpoolAdjustment`. Never a glossary synonym (`roll`, `filament`
  for a spool).
- Money-valued fields end in `Cents` only where the unit is otherwise ambiguous — the domain uses
  the `Cents` type instead (`PurchaseCost Cents`), the SQL column spells it out
  (`purchase_cost_cents`). Weights: `Grams` type, `initial_grams` column.
- Go identifiers are Go-cased; SQL is `snake_case`; form/validation field keys are `camelCase`
  string constants in the domain (`domain.FieldPurchaseCost = "purchaseCost"`).
- The operator is "the operator", never "the user". One person uses this app.
- No stutter (`spool.SpoolID`), no `Get` prefix on accessors: `Remaining()`, `State()`, `ValueOf()`.
- Test helpers are named for what they produce: `newApp`, `plaSpool`, `fieldError`, `date`.

## Types over primitives

Every quantity with a unit gets a named type in `internal/domain`: `Cents`, `Grams`,
`FilamentType`, `SpoolState`. Each such type lives in its own file with the same three things:

```go
type Grams int64

var ErrMalformedGrams = errors.New("not a whole number of grams")

func ParseGrams(s string) (Grams, error)
func FormatGrams(g Grams) string
```

- `Parse*` is total: it accepts what the operator plausibly types (surrounding space, a `g` suffix,
  a `€` prefix, `,` for `.`, any case) and returns the type's single `ErrMalformed*` sentinel for
  everything else. Parse errors never wrap `strconv`'s message — the sentinel text is what the
  operator reads.
- `Format*` is the only way a value becomes text. The TUI never does its own `%.2f`.
- Enums are string-typed constants with a `Types()` slice, a `Valid()` method that ranges over it,
  and `Names()` for anything presenting the choice. Adding a member means touching one slice.
- The malformed-value message is built from that slice
  (`"must be one of " + strings.Join(FilamentTypeNames(), ", ")`) so it cannot drift.

## Validation

Two kinds of failure, one error type.

- **Parse failure** — the text isn't a value of the type. Raised by `Parse*` in the domain, keyed
  onto a field by the use case.
- **Invariant failure** — the values are well-formed but the aggregate rejects them. Raised by the
  `New*` constructor in the domain (`NewSpool`), which is the only way to build a valid aggregate.

Both land in `*domain.ValidationError`, a field-keyed list. Never a bare `fmt.Errorf` for something
the operator can fix by typing something else.

- Constructors trim and normalise first, then collect *every* violation before returning — one
  round trip shows the operator all their mistakes.
- Use cases parse each field independently, collect the parse errors, then call the constructor and
  `MergeMissing` its invariant errors. First message per field wins: a parse failure is more useful
  than the invariant the resulting zero value then trips.
- Invariant messages read as a predicate on the field: `"is required"`, `"cannot be negative"`,
  `"must be more than 0g"`. No field name inside the message — the key carries it.
- Field keys are the `domain.Field*` constants, shared by the domain, the use case and the form, so
  an error always lands on the row that caused it.

## `internal/app` use cases

One exported method per use case, `func (a *App) Verb(ctx context.Context, ...)`.

- Input is a `VerbCmd` struct of **raw strings** — exactly what the operator typed. Parsing is the
  use case's job, so every input path is reachable from an application-layer test.
- Output is a `NounView` struct of domain types, flattened for display: every number the TUI shows
  is a field on a view (`RemainingGrams`, `RemainingValue`, `State`), computed here or in the
  domain, never in the TUI.
- Parsing lives in an unexported `parseVerb(cmd)` returning the domain aggregate; view assembly in
  an unexported `toNounView(...)`. The exported method stays a readable sequence of steps.
- Anything that writes wraps its reads and writes in one `a.db.InTx(ctx, func(tx domain.Database)
  error { ... })` and only assigns to the return value inside. Read-only use cases skip `InTx`.
- Use cases return their result by re-reading it through the same query the list uses, so what the
  caller gets back is what a later query reports.

## `internal/domain`

- Pure Go. `context`, `errors`, `fmt`, `strconv`, `strings`, `time` and nothing else.
- The domain declares the interfaces it needs (`SpoolRepository`, `Database`) in
  `repository.go`; `internal/store` implements them. `Database` composes the per-aggregate
  repositories and adds `InTx`.
- Aggregate methods are value receivers on plain structs and are pure: no clock, no IO. Time comes
  in as a parameter or from `domain.Today()`.
- Derived state is a method on a read model, not a stored field: `SpoolLedger` holds the raw
  totals, `Remaining()`/`RemainingValue()`/`State()` derive from them (ADR-0003).
- Cost math truncates deliberately and rounds once at the end (ADR-0002). Never a float for money.

## `internal/store`

- One file per aggregate (`spools.go`), plus `store.go` (handle, `Open`, `InTx`) and
  `migrations.go`.
- `var _ domain.SpoolRepository = &Store{}` at the top of each aggregate file, so a drifting
  interface fails to compile.
- Every query goes through `s.q()`, which returns the transaction when inside one and the pool
  otherwise. A method never knows whether it is in a transaction.
- SQL is a raw string literal, keywords upper-case, columns aligned in `CREATE TABLE`, `?`
  placeholders only. A query shared by the list and the by-id read is a package-level const with
  the `WHERE`/`ORDER BY` appended at the call site.
- Every error is wrapped with the operation as the operator would name it:
  `fmt.Errorf("list spools: %w", err)`, `fmt.Errorf("spool %d: %w", id, domain.ErrNotFound)`.
  `sql.ErrNoRows` is translated to `domain.ErrNotFound` — `database/sql` types never escape.
- Scanning goes through one `scanRow`-style function taking a `scanner` interface, so the row and
  the rows path cannot diverge. Domain types are converted explicitly at the boundary
  (`string(sp.FilamentType)`, `domain.FormatDate`, `domain.ParseDate`).
- Migrations are an append-only ordered `[]migration{name, sql}` with `NNNN_slug` names, applied
  once each inside a transaction and recorded in `schema_migrations`. Never edit an applied
  migration.

## `internal/tui`

A renderer. It reads views and sends commands; it computes nothing.

- The shell (`tui.go`) holds `[]namedTab` and an active index and routes keys to the active tab
  through the `tabModel` interface. Adding a tab is one slice entry.
- A tab is a bubbletea value model: `Update(tea.KeyMsg) (tabModel, tea.Cmd)`, `View`, `Help`,
  `CapturesInput`. It returns itself for keys it ignores. `CapturesInput` true means the shell
  hands over every key except its own quit.
- Forms are the exception to value semantics: `*form` is mutable open state, and the owning tab
  holds exactly one pointer to it from open to close. Focus and choice moves would be lost if a
  form were copied along with the value model.
- A tab declares a form as `[]fieldSpec` (key, label, choices, placeholder, prefill) and reads it
  back with `f.Value(key)`. The form owns focus, layout, choice cycling and error placement; the
  tab owns which fields exist. A `*_form.go` file is the adapter between one form and one
  `app.*Cmd`.
- On a failed save the tab hands the `*domain.ValidationError` to `f.SetErrors` and leaves the form
  open; the form renders each message next to its row.
- Every style is a package-level `lipgloss` var in `tui.go`; no inline styling at a call site.
- Help lines are ` · `-joined `key action` pairs. The shell owns `globalHelp`; a tab prepends its
  keys; an open form replaces it entirely, since none of the shell's keys are live.

## Comments

Default is no comment. Code that needs one is usually code that should be clearer instead — rename
the thing, split the function, introduce the type. Reach for a comment only when the reader, having
understood the code perfectly, would still change it wrongly.

That leaves three cases:

- **A non-obvious decision, stated once.** Why a form is a pointer while tabs are values. Why
  `MergeMissing` prefers the parse error. Why the sqlite driver import is blank. Put it at the one
  place that would break, not on every participant.
- **A pointer to a decision recorded elsewhere:** `(ADR-0009)`. Cite, never re-argue.
- **A contract another package fills in** (`tabModel`, `fieldSpec`): one doc comment on the type
  saying what the implementor owns. Per-member comments only for a member whose purpose the name
  cannot carry.

Never: restating a signature, narrating control flow, section banners, `// TODO` (that's an issue
under `.scratch/`), a comment on a getter, commented-out code, or a comment that repeats what the
domain type already says.

Prose belongs in `CONTEXT.md` and `docs/adr/`, not in a header block.

## Tests

`internal/app` is the only seam. Details in `CLAUDE.md`; the shape:

- Package `app_test`, external, driven against a real SQLite database in `t.TempDir()`.
- Shared helpers live in `harness_test.go`: `newApp(t)`, a valid-input builder per aggregate
  (`plaSpool()`), `fieldError(t, err, field)`, `ctx()`. A test varies one field of the builder
  rather than spelling out a whole command.
- Table-driven where a case is one input and one expectation: `tests := []struct{ name ... }`,
  `for _, tc := range tests { t.Run(tc.name, ...) }`.
- Failure messages are `got`/`want`: `t.Errorf("PurchaseCost = %d, want 2200", got.PurchaseCost)`.
  `t.Fatalf` when continuing is pointless, `t.Error` when the test can keep checking.
- Names say the behaviour, in glossary words: `TestAddSpoolThenList`,
  `TestAddSpoolParsesFilamentType`, `TestAddSpoolRejectsMalformedMoney`.
- Assert only on what a use case returns or what a later query reports. Never a struct field
  reached into, never SQL, never rendered output, never a tolerance on money.

## Commits

Conventional Commits, imperative, lower-case subject, no trailing period. The subject states the
change in domain words: `refactor: form mutates through one pointer, owns its help`,
`feat: spools tab tracer bullet, add and list`. Body only when the "why" isn't obvious from the
subject. One behavioural change per commit; a rename is its own commit.

## Docs

- A decision with a live alternative becomes an ADR: `docs/adr/NNNN-slug.md`, sections
  **Status** (`Accepted — YYYY-MM-DD`), **Context**, **Decision**, **Consequences**, prose wrapped
  at 100 columns. ADRs are immutable once accepted — supersede, don't edit.
- A new or sharpened term becomes a `CONTEXT.md` glossary entry, with the synonyms to avoid.
- Work in progress lives in `.scratch/<feature-slug>/` (see `docs/agents/issue-tracker.md`), not in
  code comments and not in `docs/`.
