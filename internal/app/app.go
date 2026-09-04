package app

import "github.com/dlvandenberg/printer-ledger/internal/domain"

// Where the ledger lives is not a domain concept, so it is not on domain.Database:
// whatever opened the ledger answers for it. internal/store satisfies this.
type ledgerLocator interface {
	Location() string
}

type App struct {
	db domain.Database
}

func New(db domain.Database) *App { return &App{db: db} }

func (a *App) ledgerFile() string {
	locator, ok := a.db.(ledgerLocator)
	if !ok {
		return ""
	}
	return locator.Location()
}
