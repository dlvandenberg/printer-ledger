package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const printQuery = `
SELECT p.id, p.design_id, p.date, p.quantity, p.minutes,
       p.kwh_price_cents, p.kwh_per_hour, p.machine_rate_cents
FROM prints p`

const filamentUsageQuery = `
SELECT u.id, u.print_id, u.spool_id, u.grams, u.cost_per_gram_hundredths
FROM filament_usages u`

var _ domain.PrintRepository = &Store{}

func (s *Store) CreatePrint(ctx context.Context, p domain.Print) (domain.Print, error) {
	res, err := s.q().ExecContext(ctx, `
INSERT INTO prints (design_id, date, quantity, minutes, kwh_price_cents,
                    kwh_per_hour, machine_rate_cents)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.DesignID, unit.FormatDate(p.Date), p.Quantity, int64(p.Minutes),
		int64(p.KwhPrice), float64(p.KwhPerHour), int64(p.MachineRate))
	if err != nil {
		return domain.Print{}, fmt.Errorf("record print: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Print{}, fmt.Errorf("record print: %w", err)
	}
	p.ID = id

	usages := make([]domain.FilamentUsage, 0, len(p.Usages))
	for _, usage := range p.Usages {
		usage.ID, err = s.createFilamentUsage(ctx, p.ID, usage)
		if err != nil {
			return domain.Print{}, err
		}
		usages = append(usages, usage)
	}
	p.Usages = usages
	return p, nil
}

func (s *Store) createFilamentUsage(ctx context.Context, printID int64, u domain.FilamentUsage) (int64, error) {
	res, err := s.q().ExecContext(ctx, `
INSERT INTO filament_usages (print_id, spool_id, grams, cost_per_gram_hundredths)
VALUES (?, ?, ?, ?)`,
		printID, u.SpoolID, int64(u.Grams), int64(u.CostPerGram))
	if err != nil {
		return 0, fmt.Errorf("record filament usage of print %d: %w", printID, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("record filament usage of print %d: %w", printID, err)
	}
	return id, nil
}

func (s *Store) Prints(ctx context.Context) ([]domain.Print, error) {
	rows, err := s.q().QueryContext(ctx, printQuery+` ORDER BY p.date DESC, p.id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list prints: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()

	var prints []domain.Print
	at := map[int64]int{}
	for rows.Next() {
		p, err := scanPrint(rows)
		if err != nil {
			return nil, fmt.Errorf("list prints: %w", err)
		}
		at[p.ID] = len(prints)
		prints = append(prints, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list prints: %w", err)
	}

	usages, err := s.q().QueryContext(ctx, filamentUsageQuery+` ORDER BY u.print_id, u.id`)
	if err != nil {
		return nil, fmt.Errorf("list prints: %w", err)
	}
	//nolint:errcheck
	defer usages.Close()

	for usages.Next() {
		usage, printID, err := scanFilamentUsage(usages)
		if err != nil {
			return nil, fmt.Errorf("list prints: %w", err)
		}
		if i, ok := at[printID]; ok {
			prints[i].Usages = append(prints[i].Usages, usage)
		}
	}
	if err := usages.Err(); err != nil {
		return nil, fmt.Errorf("list prints: %w", err)
	}
	return prints, nil
}

func (s *Store) Print(ctx context.Context, id int64) (domain.Print, error) {
	row := s.q().QueryRowContext(ctx, printQuery+` WHERE p.id = ?`, id)
	p, err := scanPrint(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Print{}, fmt.Errorf("print %d: %w", id, domain.ErrNotFound)
	}
	if err != nil {
		return domain.Print{}, fmt.Errorf("print %d: %w", id, err)
	}

	rows, err := s.q().QueryContext(ctx, filamentUsageQuery+` WHERE u.print_id = ? ORDER BY u.id`, id)
	if err != nil {
		return domain.Print{}, fmt.Errorf("print %d: %w", id, err)
	}
	//nolint:errcheck
	defer rows.Close()

	for rows.Next() {
		usage, _, err := scanFilamentUsage(rows)
		if err != nil {
			return domain.Print{}, fmt.Errorf("print %d: %w", id, err)
		}
		p.Usages = append(p.Usages, usage)
	}
	if err := rows.Err(); err != nil {
		return domain.Print{}, fmt.Errorf("print %d: %w", id, err)
	}
	return p, nil
}

func scanPrint(row scanner) (domain.Print, error) {
	var (
		p    domain.Print
		date string
	)
	err := row.Scan(&p.ID, &p.DesignID, &date, &p.Quantity, &p.Minutes,
		&p.KwhPrice, &p.KwhPerHour, &p.MachineRate)
	if err != nil {
		return domain.Print{}, err
	}
	if p.Date, err = unit.ParseDate(date); err != nil {
		return domain.Print{}, err
	}
	return p, nil
}

func scanFilamentUsage(row scanner) (domain.FilamentUsage, int64, error) {
	var (
		usage   domain.FilamentUsage
		printID int64
	)
	err := row.Scan(&usage.ID, &printID, &usage.SpoolID, &usage.Grams, &usage.CostPerGram)
	if err != nil {
		return domain.FilamentUsage{}, 0, err
	}
	return usage, printID, nil
}
