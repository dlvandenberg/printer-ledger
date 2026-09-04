package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const saleQuery = `
SELECT s.id, s.print_id, s.price_cents, s.date
FROM sales s`

var _ domain.SaleRepository = &Store{}

func (s *Store) CreateSale(ctx context.Context, sale domain.Sale) (domain.Sale, error) {
	res, err := s.q().ExecContext(ctx, `
INSERT INTO sales (print_id, price_cents, date)
VALUES (?, ?, ?)`,
		sale.PrintID, int64(sale.Price), unit.FormatDate(sale.Date))
	if err != nil {
		return domain.Sale{}, fmt.Errorf("record sale: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Sale{}, fmt.Errorf("record sale: %w", err)
	}
	sale.ID = id
	return sale, nil
}

func (s *Store) UpdateSale(ctx context.Context, sale domain.Sale) (domain.Sale, error) {
	res, err := s.q().ExecContext(ctx, `
UPDATE sales
   SET print_id    = ?,
       price_cents = ?,
       date        = ?
 WHERE id = ?`,
		sale.PrintID, int64(sale.Price), unit.FormatDate(sale.Date), sale.ID)
	if err != nil {
		return domain.Sale{}, fmt.Errorf("edit sale %d: %w", sale.ID, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return domain.Sale{}, fmt.Errorf("edit sale %d: %w", sale.ID, err)
	}
	if affected == 0 {
		return domain.Sale{}, fmt.Errorf("sale %d: %w", sale.ID, domain.ErrNotFound)
	}
	return sale, nil
}

func (s *Store) DeleteSale(ctx context.Context, id int64) error {
	res, err := s.q().ExecContext(ctx, `DELETE FROM sales WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete sale %d: %w", id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete sale %d: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("sale %d: %w", id, domain.ErrNotFound)
	}
	return nil
}

func (s *Store) Sales(ctx context.Context) ([]domain.Sale, error) {
	rows, err := s.q().QueryContext(ctx, saleQuery+` ORDER BY s.date DESC, s.id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list sales: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()

	var sales []domain.Sale
	for rows.Next() {
		sale, err := scanSale(rows)
		if err != nil {
			return nil, fmt.Errorf("list sales: %w", err)
		}
		sales = append(sales, sale)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list sales: %w", err)
	}
	return sales, nil
}

func (s *Store) Sale(ctx context.Context, id int64) (domain.Sale, error) {
	row := s.q().QueryRowContext(ctx, saleQuery+` WHERE s.id = ?`, id)
	sale, err := scanSale(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Sale{}, fmt.Errorf("sale %d: %w", id, domain.ErrNotFound)
	}
	if err != nil {
		return domain.Sale{}, fmt.Errorf("sale %d: %w", id, err)
	}
	return sale, nil
}

func scanSale(row scanner) (domain.Sale, error) {
	var (
		sale domain.Sale
		date string
	)
	if err := row.Scan(&sale.ID, &sale.PrintID, &sale.Price, &date); err != nil {
		return domain.Sale{}, err
	}
	var err error
	if sale.Date, err = unit.ParseDate(date); err != nil {
		return domain.Sale{}, err
	}
	return sale, nil
}
