package domain

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type SpoolRepository interface {
	CreateSpool(ctx context.Context, s Spool) (Spool, error)
	UpdateSpool(ctx context.Context, s Spool) (Spool, error)
	SpoolLedgers(ctx context.Context) ([]SpoolLedger, error)
	SpoolLedger(ctx context.Context, id int64) (SpoolLedger, error)
	CreateSpoolAdjustment(ctx context.Context, a SpoolAdjustment) (SpoolAdjustment, error)
	SpoolAdjustments(ctx context.Context, spoolID int64) ([]SpoolAdjustment, error)
}

type DesignRepository interface {
	CreateDesign(ctx context.Context, d Design) (Design, error)
	UpdateDesign(ctx context.Context, d Design) (Design, error)
	Designs(ctx context.Context) ([]Design, error)
	Design(ctx context.Context, id int64) (Design, error)
}

type PrintRepository interface {
	CreatePrint(ctx context.Context, p Print) (Print, error)
	Prints(ctx context.Context) ([]Print, error)
	Print(ctx context.Context, id int64) (Print, error)
}

type SettingsRepository interface {
	Settings(ctx context.Context) (Settings, error)
	SaveSettings(ctx context.Context, s Settings) error
}

type Database interface {
	SpoolRepository
	SettingsRepository
	DesignRepository
	PrintRepository
	InTx(ctx context.Context, fn func(Database) error) error
}
