package store

import (
	"context"
	"fmt"
)

type migration struct {
	name string
	sql  string
}

var migrations = []migration{
	{
		name: "0001_spools",
		sql: `
CREATE TABLE spools (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    filament_type       TEXT    NOT NULL,
    brand               TEXT    NOT NULL,
    color               TEXT    NOT NULL,
    initial_grams       INTEGER NOT NULL,
    tare_grams          INTEGER NOT NULL,
    purchase_cost_cents INTEGER NOT NULL,
    purchase_date       TEXT    NOT NULL
);`,
	},
	{
		name: "0002_settings",
		sql: `
CREATE TABLE settings (
    id                          INTEGER PRIMARY KEY CHECK (id = 1),
    kwh_price_cents             INTEGER NOT NULL,
    machine_hourly_rate_cents   INTEGER NOT NULL,
    printer_purchase_cost_cents INTEGER NOT NULL,
    default_margin_hundredths   INTEGER NOT NULL,
    min_margin_hundredths       INTEGER NOT NULL
);

CREATE TABLE power_rates (
    filament_type TEXT    PRIMARY KEY,
    kwh_per_hour  REAL    NOT NULL,
    measured      INTEGER NOT NULL
);`,
	},
	{
		name: "0003_spool_adjustments",
		sql: `
CREATE TABLE spool_adjustments (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    spool_id                INTEGER NOT NULL REFERENCES spools(id),
    measured_grams          INTEGER NOT NULL,
    delta_grams             INTEGER NOT NULL,
    adjusted_on             TEXT    NOT NULL,
    note                    TEXT    NOT NULL
);`,
	},
	{
		name: "0004_designs",
		sql: `
CREATE TABLE designs (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    name                  TEXT NOT NULL,
    default_filament_type TEXT NOT NULL,
    estimated_grams       INTEGER NOT NULL,
    estimated_minutes     INTEGER NOT NULL,
    margin_hundredths     INTEGER NOT NULL
);`,
	},
	{
		name: "0005_prints",
		sql: `
CREATE TABLE prints (
    id                       INTEGER PRIMARY KEY AUTOINCREMENT,
    design_id                INTEGER NOT NULL REFERENCES designs(id),
    date                     TEXT    NOT NULL,
    quantity                 INTEGER NOT NULL,
    minutes                  INTEGER NOT NULL,
    kwh_price_cents          INTEGER NOT NULL,
    kwh_per_hour             REAL    NOT NULL,
    machine_rate_cents       INTEGER NOT NULL
);

CREATE TABLE filament_usages (
    id                       INTEGER PRIMARY KEY AUTOINCREMENT,
    print_id                 INTEGER NOT NULL REFERENCES prints(id) ON DELETE CASCADE,
    spool_id                 INTEGER NOT NULL REFERENCES spools(id),
    grams                    INTEGER NOT NULL,
    cost_per_gram_hundredths INTEGER NOT NULL
);

CREATE INDEX filament_usages_spool ON filament_usages(spool_id);
CREATE INDEX filament_usages_print ON filament_usages(print_id);`,
	},
	{
		name: "0006_print_stock_counts",
		sql: `
ALTER TABLE prints ADD COLUMN gifted_count   INTEGER NOT NULL DEFAULT 0;
ALTER TABLE prints ADD COLUMN kept_count     INTEGER NOT NULL DEFAULT 0;
ALTER TABLE prints ADD COLUMN scrapped_count INTEGER NOT NULL DEFAULT 0;`,
	},
}

func (s *Store) migrate(ctx context.Context) error {
	q := s.q()
	if _, err := q.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    name       TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := map[string]bool{}
	rows, err := q.QueryContext(ctx, `SELECT name FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("read schema_migrations: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("read schema_migrations: %w", err)
		}
		applied[name] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read schema_migrations: %w", err)
	}

	for _, m := range migrations {
		if applied[m.name] {
			continue
		}
		err := s.inTx(ctx, func(tx *Store) error {
			if _, err := tx.q().ExecContext(ctx, m.sql); err != nil {
				return err
			}
			_, err := tx.q().ExecContext(ctx, `INSERT INTO schema_migrations (name) VALUES (?)`, m.name)
			return err
		})
		if err != nil {
			return fmt.Errorf("migration %s: %w", m.name, err)
		}
	}
	return nil
}
