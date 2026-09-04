package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

// sold_count is 0 until Sales reference a Print; availability derives from it
// rather than being stored (ADR-0013).
const printLedgerQuery = `
SELECT p.id, p.design_id, p.date, p.quantity, p.minutes,
       p.gifted_count, p.kept_count, p.scrapped_count,
       p.kwh_price_cents, p.kwh_per_hour, p.machine_rate_cents,
       0 AS sold_count
FROM prints p`

const filamentUsageQuery = `
SELECT u.id, u.print_id, u.spool_id, u.grams, u.cost_per_gram_hundredths
FROM filament_usages u`

var _ domain.PrintRepository = &Store{}

func (s *Store) CreatePrint(ctx context.Context, p domain.Print) (domain.Print, error) {
	res, err := s.q().ExecContext(ctx, `
INSERT INTO prints (design_id, date, quantity, minutes, gifted_count, kept_count,
                    scrapped_count, kwh_price_cents, kwh_per_hour, machine_rate_cents)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.DesignID, unit.FormatDate(p.Date), p.Quantity, int64(p.Minutes),
		p.GiftedCount, p.KeptCount, p.ScrappedCount,
		int64(p.KwhPrice), float64(p.KwhPerHour), int64(p.MachineRate))
	if err != nil {
		return domain.Print{}, fmt.Errorf("record print: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Print{}, fmt.Errorf("record print: %w", err)
	}
	p.ID = id

	usages, err := s.replaceFilamentUsages(ctx, p)
	if err != nil {
		return domain.Print{}, err
	}
	p.Usages = usages
	return p, nil
}

// UpdatePrint rewrites the usage rows rather than diffing them: the Print it is
// handed already carries the price each row is costed at (ADR-0004).
func (s *Store) UpdatePrint(ctx context.Context, p domain.Print) (domain.Print, error) {
	res, err := s.q().ExecContext(ctx, `
UPDATE prints
   SET design_id          = ?,
       date               = ?,
       quantity           = ?,
       minutes            = ?,
       gifted_count       = ?,
       kept_count         = ?,
       scrapped_count     = ?
 WHERE id = ?`,
		p.DesignID, unit.FormatDate(p.Date), p.Quantity, int64(p.Minutes),
		p.GiftedCount, p.KeptCount, p.ScrappedCount, p.ID)
	if err != nil {
		return domain.Print{}, fmt.Errorf("edit print %d: %w", p.ID, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return domain.Print{}, fmt.Errorf("edit print %d: %w", p.ID, err)
	}
	if affected == 0 {
		return domain.Print{}, fmt.Errorf("print %d: %w", p.ID, domain.ErrNotFound)
	}

	if _, err := s.q().ExecContext(ctx, `DELETE FROM filament_usages WHERE print_id = ?`, p.ID); err != nil {
		return domain.Print{}, fmt.Errorf("edit print %d: %w", p.ID, err)
	}
	usages, err := s.replaceFilamentUsages(ctx, p)
	if err != nil {
		return domain.Print{}, err
	}
	p.Usages = usages
	return p, nil
}

func (s *Store) DeletePrint(ctx context.Context, id int64) error {
	res, err := s.q().ExecContext(ctx, `DELETE FROM prints WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete print %d: %w", id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete print %d: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("print %d: %w", id, domain.ErrNotFound)
	}
	return nil
}

func (s *Store) replaceFilamentUsages(ctx context.Context, p domain.Print) ([]domain.FilamentUsage, error) {
	usages := make([]domain.FilamentUsage, 0, len(p.Usages))
	for _, usage := range p.Usages {
		res, err := s.q().ExecContext(ctx, `
INSERT INTO filament_usages (print_id, spool_id, grams, cost_per_gram_hundredths)
VALUES (?, ?, ?, ?)`,
			p.ID, usage.SpoolID, int64(usage.Grams), int64(usage.CostPerGram))
		if err != nil {
			return nil, fmt.Errorf("record filament usage of print %d: %w", p.ID, err)
		}
		if usage.ID, err = res.LastInsertId(); err != nil {
			return nil, fmt.Errorf("record filament usage of print %d: %w", p.ID, err)
		}
		usages = append(usages, usage)
	}
	return usages, nil
}

func (s *Store) PrintLedgers(ctx context.Context) ([]domain.PrintLedger, error) {
	rows, err := s.q().QueryContext(ctx, printLedgerQuery+` ORDER BY p.date DESC, p.id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list prints: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()

	var ledgers []domain.PrintLedger
	at := map[int64]int{}
	for rows.Next() {
		ledger, err := scanPrintLedger(rows)
		if err != nil {
			return nil, fmt.Errorf("list prints: %w", err)
		}
		at[ledger.Print.ID] = len(ledgers)
		ledgers = append(ledgers, ledger)
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
			ledgers[i].Print.Usages = append(ledgers[i].Print.Usages, usage)
		}
	}
	if err := usages.Err(); err != nil {
		return nil, fmt.Errorf("list prints: %w", err)
	}
	return ledgers, nil
}

func (s *Store) PrintLedger(ctx context.Context, id int64) (domain.PrintLedger, error) {
	row := s.q().QueryRowContext(ctx, printLedgerQuery+` WHERE p.id = ?`, id)
	ledger, err := scanPrintLedger(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PrintLedger{}, fmt.Errorf("print %d: %w", id, domain.ErrNotFound)
	}
	if err != nil {
		return domain.PrintLedger{}, fmt.Errorf("print %d: %w", id, err)
	}

	rows, err := s.q().QueryContext(ctx, filamentUsageQuery+` WHERE u.print_id = ? ORDER BY u.id`, id)
	if err != nil {
		return domain.PrintLedger{}, fmt.Errorf("print %d: %w", id, err)
	}
	//nolint:errcheck
	defer rows.Close()

	for rows.Next() {
		usage, _, err := scanFilamentUsage(rows)
		if err != nil {
			return domain.PrintLedger{}, fmt.Errorf("print %d: %w", id, err)
		}
		ledger.Print.Usages = append(ledger.Print.Usages, usage)
	}
	if err := rows.Err(); err != nil {
		return domain.PrintLedger{}, fmt.Errorf("print %d: %w", id, err)
	}
	return ledger, nil
}

func scanPrintLedger(row scanner) (domain.PrintLedger, error) {
	var (
		ledger domain.PrintLedger
		date   string
	)
	err := row.Scan(&ledger.Print.ID, &ledger.Print.DesignID, &date, &ledger.Print.Quantity,
		&ledger.Print.Minutes, &ledger.Print.GiftedCount, &ledger.Print.KeptCount,
		&ledger.Print.ScrappedCount, &ledger.Print.KwhPrice, &ledger.Print.KwhPerHour,
		&ledger.Print.MachineRate, &ledger.SoldCount)
	if err != nil {
		return domain.PrintLedger{}, err
	}
	if ledger.Print.Date, err = unit.ParseDate(date); err != nil {
		return domain.PrintLedger{}, err
	}
	return ledger, nil
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
