package app

import "github.com/dlvandenberg/printer-ledger/internal/domain"

type App struct {
	db domain.Database
}

func New(db domain.Database) *App { return &App{db: db} }
