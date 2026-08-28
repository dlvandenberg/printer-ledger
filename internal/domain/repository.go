package domain

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type SpoolRepository interface {
	CreateSpool(ctx context.Context, s Spool) (Spool, error)
	SpoolLedgers(ctx context.Context) ([]SpoolLedger, error)
	SpoolLedger(ctx context.Context, id int64) (SpoolLedger, error)
}

type SettingsRepository interface {
	Settings(ctx context.Context) (Settings, error)
	SaveSettings(ctx context.Context, s Settings) error
}

type Database interface {
	SpoolRepository
	SettingsRepository
	InTx(ctx context.Context, fn func(Database) error) error
}
