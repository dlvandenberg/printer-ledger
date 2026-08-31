package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const spoolLedgerQuery = `
SELECT s.id, s.filament_type, s.brand, s.color, s.initial_grams, s.tare_grams,
       s.purchase_cost_cents, s.purchase_date,
       0 AS used_grams,
       COALESCE((SELECT SUM(a.delta_grams) FROM spool_adjustments a WHERE a.spool_id = s.id), 0) AS adjusted_grams
FROM spools s`

const spoolAdjustmentQuery = `
SELECT a.id, a.spool_id, a.measured_grams, a.delta_grams,
       a.adjusted_on, a.note
FROM spool_adjustments a`

var _ domain.SpoolRepository = &Store{}

func (s *Store) CreateSpool(ctx context.Context, sp domain.Spool) (domain.Spool, error) {
	res, err := s.q().ExecContext(ctx, `
INSERT INTO spools (filament_type, brand, color, initial_grams, tare_grams,
                    purchase_cost_cents, purchase_date)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		string(sp.FilamentType), sp.Brand, sp.Color, int64(sp.InitialGrams),
		int64(sp.TareGrams), int64(sp.PurchaseCost), unit.FormatDate(sp.PurchaseDate))
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
	//nolint:errcheck
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
	if ledger.Spool.PurchaseDate, err = unit.ParseDate(purchaseDate); err != nil {
		return domain.SpoolLedger{}, err
	}
	return ledger, nil
}

func (s *Store) CreateSpoolAdjustment(ctx context.Context, a domain.SpoolAdjustment) (domain.SpoolAdjustment, error) {
	res, err := s.q().ExecContext(ctx, `
INSERT INTO spool_adjustments (spool_id, measured_grams, delta_grams,
                               adjusted_on, note)
VALUES (?, ?, ?, ?, ?)`,
		a.SpoolID, int64(a.MeasuredGrams), int64(a.DeltaGrams),
		unit.FormatDate(a.AdjustedOn), a.Note)
	if err != nil {
		return domain.SpoolAdjustment{}, fmt.Errorf("re-weigh spool %d: %w", a.SpoolID, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.SpoolAdjustment{}, fmt.Errorf("re-weigh spool %d: %w", a.SpoolID, err)
	}
	a.ID = id
	return a, nil
}

func (s *Store) SpoolAdjustments(ctx context.Context, spoolID int64) ([]domain.SpoolAdjustment, error) {
	rows, err := s.q().QueryContext(ctx,
		// Insertion order, not date order: each delta was computed against the
		// remaining at the moment it was recorded, so only that order chains.
		spoolAdjustmentQuery+` WHERE a.spool_id = ? ORDER BY a.id`, spoolID)
	if err != nil {
		return nil, fmt.Errorf("list adjustments of spool %d: %w", spoolID, err)
	}
	//nolint:errcheck
	defer rows.Close()

	var adjustments []domain.SpoolAdjustment
	for rows.Next() {
		adjustment, err := scanSpoolAdjustment(rows)
		if err != nil {
			return nil, fmt.Errorf("list adjustments of spool %d: %w", spoolID, err)
		}
		adjustments = append(adjustments, adjustment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list adjustments of spool %d: %w", spoolID, err)
	}
	return adjustments, nil
}

func scanSpoolAdjustment(row scanner) (domain.SpoolAdjustment, error) {
	var (
		adjustment domain.SpoolAdjustment
		adjustedOn string
	)
	err := row.Scan(&adjustment.ID, &adjustment.SpoolID, &adjustment.MeasuredGrams, &adjustment.DeltaGrams, &adjustedOn, &adjustment.Note)
	if err != nil {
		return domain.SpoolAdjustment{}, err
	}
	if adjustment.AdjustedOn, err = unit.ParseDate(adjustedOn); err != nil {
		return domain.SpoolAdjustment{}, err
	}
	return adjustment, nil
}
