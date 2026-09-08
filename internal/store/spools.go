package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const spoolQuery = `
SELECT s.id, s.filament_type, s.brand, s.color, s.initial_centigrams, s.tare_centigrams,
       s.purchase_cost_cents, s.purchase_date
FROM spools s`

const spoolLedgerQuery = `
SELECT s.id, s.filament_type, s.brand, s.color, s.initial_centigrams, s.tare_centigrams,
       s.purchase_cost_cents, s.purchase_date,
       COALESCE((SELECT SUM(u.centigrams) FROM filament_usages u WHERE u.spool_id = s.id), 0) AS used_centigrams,
       COALESCE((SELECT SUM(a.delta_centigrams) FROM spool_adjustments a WHERE a.spool_id = s.id), 0) AS adjusted_centigrams
FROM spools s`

const spoolOrder = ` ORDER BY s.purchase_date DESC, s.id DESC`

const spoolAdjustmentQuery = `
SELECT a.id, a.spool_id, a.measured_centigrams, a.delta_centigrams,
       a.adjusted_on, a.note
FROM spool_adjustments a`

var _ domain.SpoolRepository = &Store{}

func (s *Store) CreateSpool(ctx context.Context, sp domain.Spool) (domain.Spool, error) {
	res, err := s.q().ExecContext(ctx, `
INSERT INTO spools (filament_type, brand, color, initial_centigrams, tare_centigrams,
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

func (s *Store) UpdateSpool(ctx context.Context, sp domain.Spool) (domain.Spool, error) {
	res, err := s.q().ExecContext(ctx, `
UPDATE spools
   SET filament_type       = ?,
       brand               = ?,
       color               = ?,
       initial_centigrams  = ?,
       tare_centigrams     = ?,
       purchase_cost_cents = ?,
       purchase_date       = ?
 WHERE id = ?`,
		string(sp.FilamentType), sp.Brand, sp.Color, int64(sp.InitialGrams),
		int64(sp.TareGrams), int64(sp.PurchaseCost), unit.FormatDate(sp.PurchaseDate), sp.ID)
	if err != nil {
		return domain.Spool{}, fmt.Errorf("edit spool %d: %w", sp.ID, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return domain.Spool{}, fmt.Errorf("edit spool %d: %w", sp.ID, err)
	}
	if affected == 0 {
		return domain.Spool{}, fmt.Errorf("spool %d: %w", sp.ID, domain.ErrNotFound)
	}
	return sp, nil
}

// DeleteSpool takes the Spool's Adjustments with it: they are corrections to
// this Spool and mean nothing without it (ADR-0022).
func (s *Store) DeleteSpool(ctx context.Context, id int64) error {
	if _, err := s.q().ExecContext(ctx, `DELETE FROM spool_adjustments WHERE spool_id = ?`, id); err != nil {
		return fmt.Errorf("delete spool %d: %w", id, err)
	}
	res, err := s.q().ExecContext(ctx, `DELETE FROM spools WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete spool %d: %w", id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete spool %d: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("spool %d: %w", id, domain.ErrNotFound)
	}
	return nil
}

func (s *Store) Spools(ctx context.Context) ([]domain.Spool, error) {
	rows, err := s.q().QueryContext(ctx, spoolQuery+spoolOrder)
	if err != nil {
		return nil, fmt.Errorf("list spools: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()

	var spools []domain.Spool
	for rows.Next() {
		spool, err := scanSpool(rows)
		if err != nil {
			return nil, fmt.Errorf("list spools: %w", err)
		}
		spools = append(spools, spool)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list spools: %w", err)
	}
	return spools, nil
}

func (s *Store) SpoolLedgers(ctx context.Context) ([]domain.SpoolLedger, error) {
	rows, err := s.q().QueryContext(ctx, spoolLedgerQuery+spoolOrder)
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

func scanSpool(row scanner) (domain.Spool, error) {
	var (
		spool        domain.Spool
		filamentType string
		purchaseDate string
	)
	err := row.Scan(&spool.ID, &filamentType, &spool.Brand, &spool.Color,
		&spool.InitialGrams, &spool.TareGrams, &spool.PurchaseCost, &purchaseDate)
	if err != nil {
		return domain.Spool{}, err
	}
	spool.FilamentType = domain.FilamentType(filamentType)
	if spool.PurchaseDate, err = unit.ParseDate(purchaseDate); err != nil {
		return domain.Spool{}, err
	}
	return spool, nil
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
INSERT INTO spool_adjustments (spool_id, measured_centigrams, delta_centigrams,
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
