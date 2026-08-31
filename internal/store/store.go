package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	_ "modernc.org/sqlite" // pure-Go driver: CGO_ENABLED=0 still builds (ADR-0009)
)

const DefaultPath = "./printer-ledger.db"

type queryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Store struct {
	db *sql.DB
	tx *sql.Tx // non-nil on the Store handed to an InTx callback
}

func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	db.SetMaxOpenConns(1)

	if err := db.PingContext(context.Background()); err != nil {
		//nolint:errcheck
		db.Close()
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		//nolint:errcheck
		db.Close()
		return nil, err
	}
	if err := s.seedSettings(context.Background()); err != nil {
		//nolint:errcheck
		db.Close()
		return nil, err
	}
	return s, nil
}

func MustOpen(path string) *Store {
	s, err := Open(path)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) InTx(ctx context.Context, fn func(domain.Database) error) error {
	return s.inTx(ctx, func(tx *Store) error { return fn(tx) })
}

func (s *Store) inTx(ctx context.Context, fn func(*Store) error) error {
	if s.tx != nil {
		return fn(s)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	//nolint:errcheck
	defer tx.Rollback()
	if err := fn(&Store{db: s.db, tx: tx}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) q() queryer {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}
