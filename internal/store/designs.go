package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

const designQuery = `
SELECT d.id, d.name, d.default_filament_type, d.estimated_grams, d.estimated_minutes,
       d.margin_hundredths
FROM designs d`

var _ domain.DesignRepository = &Store{}

func (s *Store) CreateDesign(ctx context.Context, d domain.Design) (domain.Design, error) {
	res, err := s.q().ExecContext(ctx, `
INSERT INTO designs (name, default_filament_type, estimated_grams, estimated_minutes,
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
       estimated_grams       = ?,
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
	rows, err := s.q().QueryContext(ctx, designQuery+` ORDER BY d.name COLLATE NOCASE, d.id`)
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
