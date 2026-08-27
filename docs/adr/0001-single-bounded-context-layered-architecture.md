# ADR-0001: Single bounded context, four-layer architecture

**Status:** Accepted — 2026-08-25 (amended 2026-08-25: added `internal/app`)

## Context

printer-admin tracks filament, prints, and sales for one operator with one printer. A domain-driven
approach invites context boundaries (inventory vs. production vs. sales), but the whole model fits
in one head and every aggregate is read together on nearly every screen.

Originally this ADR named three layers. Choosing the test seam (see the v1 spec) showed that testing
`domain` and `store` separately verifies cost math and persistence apart but never wired together,
so a wiring bug passes both suites. A single use-case layer gives one seam that covers both.

## Decision

One bounded context. Four layers:

```
internal/domain   pure Go, zero dependencies. Aggregates, cost math, invariants.
                  Declares the repository interfaces it needs.
internal/store    SQLite implementation of those repositories.
internal/app      Use cases, one method each. Owns transaction boundaries.
                  This is the test seam.
internal/tui      bubbletea + bubbles + lipgloss. Thin renderer over internal/app.
```

Dependencies point inward: `tui → app → domain ← store`.

The TUI must be able to do nothing `internal/app` cannot. Any value the UI displays is returned by
a use case, never computed in the UI.

## Consequences

- One test seam: use cases driven against a real SQLite database in `t.TempDir()`. Cost math is
  verified end to end through persistence, not in isolation from it.
- No anti-corruption layers, no cross-context translation, no eventual consistency to reason about.
- The domain stays pure and could be tested directly, but routine tests go through `app` to keep the
  seam count at one.
- `internal/app` is a layer of pass-through methods for simple CRUD. Accepted: it is where
  transactions and invariant orchestration live, which CRUD-only cases do not show off.
- If the app ever grows a second printer owner or a webshop integration, the seams have to be cut
  then. Accepted: speculative boundaries cost more than the eventual refactor.
