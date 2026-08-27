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

type Database interface {
	SpoolRepository
	InTx(ctx context.Context, fn func(Database) error) error
}
