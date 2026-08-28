package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

const spoolLedgerQuery = `
SELECT s.id, s.filament_type, s.brand, s.color, s.initial_grams, s.tare_grams,
       s.purchase_cost_cents, s.purchase_date,
       0 AS used_grams,
       0 AS adjusted_grams
FROM spools s`

var _ domain.SpoolRepository = &Store{}

func (s *Store) CreateSpool(ctx context.Context, sp domain.Spool) (domain.Spool, error) {
	res, err := s.q().ExecContext(ctx, `
INSERT INTO spools (filament_type, brand, color, initial_grams, tare_grams,
                    purchase_cost_cents, purchase_date)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		string(sp.FilamentType), sp.Brand, sp.Color, int64(sp.InitialGrams),
		int64(sp.TareGrams), int64(sp.PurchaseCost), domain.FormatDate(sp.PurchaseDate))
	if err != nil {
		return domain.Spool{}, fmt.Errorf("create spool: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Spool{}, fmt.Errorf("create spool: %w", err)
	}
	sp.ID = id
	return sp, nil
}

func (s *Store) SpoolLedgers(ctx context.Context) ([]domain.SpoolLedger, error) {
	rows, err := s.q().QueryContext(ctx, spoolLedgerQuery+` ORDER BY s.purchase_date DESC, s.id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list spools: %w", err)
	}
	defer rows.Close()

	var ledgers []domain.SpoolLedger
	for rows.Next() {
		ledger, err := scanSpoolLedger(rows)
		if err != nil {
			return nil, fmt.Errorf("list spools: %w", err)
		}
		ledgers = append(ledgers, ledger)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list spools: %w", err)
	}
	return ledgers, nil
}

func (s *Store) SpoolLedger(ctx context.Context, id int64) (domain.SpoolLedger, error) {
	row := s.q().QueryRowContext(ctx, spoolLedgerQuery+` WHERE s.id = ?`, id)
	ledger, err := scanSpoolLedger(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.SpoolLedger{}, fmt.Errorf("spool %d: %w", id, domain.ErrNotFound)
	}
	if err != nil {
		return domain.SpoolLedger{}, fmt.Errorf("spool %d: %w", id, err)
	}
	return ledger, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanSpoolLedger(row scanner) (domain.SpoolLedger, error) {
	var (
		ledger       domain.SpoolLedger
		filamentType string
		purchaseDate string
	)
	err := row.Scan(&ledger.Spool.ID, &filamentType, &ledger.Spool.Brand, &ledger.Spool.Color,
		&ledger.Spool.InitialGrams, &ledger.Spool.TareGrams, &ledger.Spool.PurchaseCost,
		&purchaseDate, &ledger.UsedGrams, &ledger.AdjustedGrams)
	if err != nil {
		return domain.SpoolLedger{}, err
	}
	ledger.Spool.FilamentType = domain.FilamentType(filamentType)
	if ledger.Spool.PurchaseDate, err = domain.ParseDate(purchaseDate); err != nil {
		return domain.SpoolLedger{}, err
	}
	return ledger, nil
}
