package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

const designQuery = `
SELECT d.id, d.name, d.default_filament_type, d.estimated_centigrams, d.estimated_minutes,
       d.margin_hundredths
FROM designs d`

const designLedgerQuery = `
SELECT d.id, d.name, d.default_filament_type, d.estimated_centigrams, d.estimated_minutes,
       d.margin_hundredths,
       (SELECT COUNT(*) FROM prints p WHERE p.design_id = d.id) AS print_count
FROM designs d`

const designOrder = ` ORDER BY d.name COLLATE NOCASE, d.id`

var _ domain.DesignRepository = &Store{}

func (s *Store) CreateDesign(ctx context.Context, d domain.Design) (domain.Design, error) {
	res, err := s.q().ExecContext(ctx, `
INSERT INTO designs (name, default_filament_type, estimated_centigrams, estimated_minutes,
                     margin_hundredths)
VALUES (?, ?, ?, ?, ?)`,
		d.Name, string(d.DefaultFilamentType), int64(d.EstimatedGrams),
		int64(d.EstimatedMinutes), int64(d.MarginPct))
	if err != nil {
		return domain.Design{}, fmt.Errorf("create design: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Design{}, fmt.Errorf("create design: %w", err)
	}
	d.ID = id
	return d, nil
}

func (s *Store) UpdateDesign(ctx context.Context, d domain.Design) (domain.Design, error) {
	res, err := s.q().ExecContext(ctx, `
UPDATE designs
   SET name                  = ?,
       default_filament_type = ?,
       estimated_centigrams  = ?,
       estimated_minutes     = ?,
       margin_hundredths     = ?
 WHERE id = ?`,
		d.Name, string(d.DefaultFilamentType), int64(d.EstimatedGrams),
		int64(d.EstimatedMinutes), int64(d.MarginPct), d.ID)
	if err != nil {
		return domain.Design{}, fmt.Errorf("edit design %d: %w", d.ID, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return domain.Design{}, fmt.Errorf("edit design %d: %w", d.ID, err)
	}
	if affected == 0 {
		return domain.Design{}, fmt.Errorf("design %d: %w", d.ID, domain.ErrNotFound)
	}
	return d, nil
}

func (s *Store) DeleteDesign(ctx context.Context, id int64) error {
	res, err := s.q().ExecContext(ctx, `DELETE FROM designs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete design %d: %w", id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete design %d: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("design %d: %w", id, domain.ErrNotFound)
	}
	return nil
}

func (s *Store) Design(ctx context.Context, id int64) (domain.Design, error) {
	row := s.q().QueryRowContext(ctx, designQuery+` WHERE d.id = ?`, id)
	design, err := scanDesign(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Design{}, fmt.Errorf("design %d: %w", id, domain.ErrNotFound)
	}
	if err != nil {
		return domain.Design{}, fmt.Errorf("design %d: %w", id, err)
	}
	return design, nil
}

func (s *Store) Designs(ctx context.Context) ([]domain.Design, error) {
	rows, err := s.q().QueryContext(ctx, designQuery+designOrder)
	if err != nil {
		return nil, fmt.Errorf("list designs: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()

	var designs []domain.Design
	for rows.Next() {
		design, err := scanDesign(rows)
		if err != nil {
			return nil, fmt.Errorf("list designs: %w", err)
		}
		designs = append(designs, design)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list designs: %w", err)
	}
	return designs, nil
}

func (s *Store) DesignLedger(ctx context.Context, id int64) (domain.DesignLedger, error) {
	row := s.q().QueryRowContext(ctx, designLedgerQuery+` WHERE d.id = ?`, id)
	ledger, err := scanDesignLedger(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DesignLedger{}, fmt.Errorf("design %d: %w", id, domain.ErrNotFound)
	}
	if err != nil {
		return domain.DesignLedger{}, fmt.Errorf("design %d: %w", id, err)
	}
	return ledger, nil
}

func (s *Store) DesignLedgers(ctx context.Context) ([]domain.DesignLedger, error) {
	rows, err := s.q().QueryContext(ctx, designLedgerQuery+designOrder)
	if err != nil {
		return nil, fmt.Errorf("list designs: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()

	var ledgers []domain.DesignLedger
	for rows.Next() {
		ledger, err := scanDesignLedger(rows)
		if err != nil {
			return nil, fmt.Errorf("list designs: %w", err)
		}
		ledgers = append(ledgers, ledger)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list designs: %w", err)
	}
	return ledgers, nil
}

func scanDesign(row scanner) (domain.Design, error) {
	var (
		design       domain.Design
		filamentType string
	)
	err := row.Scan(&design.ID, &design.Name, &filamentType, &design.EstimatedGrams,
		&design.EstimatedMinutes, &design.MarginPct)
	if err != nil {
		return domain.Design{}, err
	}
	design.DefaultFilamentType = domain.FilamentType(filamentType)
	return design, nil
}

func scanDesignLedger(row scanner) (domain.DesignLedger, error) {
	var (
		ledger       domain.DesignLedger
		design       domain.Design
		filamentType string
	)
	err := row.Scan(&design.ID, &design.Name, &filamentType, &design.EstimatedGrams,
		&design.EstimatedMinutes, &design.MarginPct, &ledger.PrintCount)
	if err != nil {
		return domain.DesignLedger{}, err
	}
	design.DefaultFilamentType = domain.FilamentType(filamentType)
	ledger.Design = design
	return ledger, nil
}
