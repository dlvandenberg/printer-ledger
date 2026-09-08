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
	DeleteSpool(ctx context.Context, id int64) error
	CreateSpoolAdjustment(ctx context.Context, a SpoolAdjustment) (SpoolAdjustment, error)
	SpoolAdjustments(ctx context.Context, spoolID int64) ([]SpoolAdjustment, error)
}

// DesignRepository answers a read in two shapes. The ledger carries the print
// count, and is what a list the operator reads, an invariant and a delete guard
// ask for. The plain read carries the aggregate and nothing derived, so a caller
// that only needs a name cannot hold a second opinion on that count.
type DesignRepository interface {
	CreateDesign(ctx context.Context, d Design) (Design, error)
	UpdateDesign(ctx context.Context, d Design) (Design, error)
	DeleteDesign(ctx context.Context, id int64) error
	Designs(ctx context.Context) ([]Design, error)
	Design(ctx context.Context, id int64) (Design, error)
	DesignLedgers(ctx context.Context) ([]DesignLedger, error)
	DesignLedger(ctx context.Context, id int64) (DesignLedger, error)
}

type PrintRepository interface {
	CreatePrint(ctx context.Context, p Print) (Print, error)
	UpdatePrint(ctx context.Context, p Print) (Print, error)
	DeletePrint(ctx context.Context, id int64) error
	PrintLedgers(ctx context.Context) ([]PrintLedger, error)
	PrintLedger(ctx context.Context, id int64) (PrintLedger, error)
}

type SaleRepository interface {
	CreateSale(ctx context.Context, s Sale) (Sale, error)
	UpdateSale(ctx context.Context, s Sale) (Sale, error)
	DeleteSale(ctx context.Context, id int64) error
	Sales(ctx context.Context) ([]Sale, error)
	Sale(ctx context.Context, id int64) (Sale, error)
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
	SaleRepository
	InTx(ctx context.Context, fn func(Database) error) error
}
