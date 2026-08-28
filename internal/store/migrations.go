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
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return fmt.Errorf("read schema_migrations: %w", err)
		}
		applied[name] = true
	}
	rows.Close()
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
