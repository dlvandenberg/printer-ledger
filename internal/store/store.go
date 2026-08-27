package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	_ "modernc.org/sqlite" // pure-Go driver: CGO_ENABLED=0 still builds (ADR-0009)
)

const DefaultPath = "./printer-ledger.db"

// queryer is the subset of *sql.DB and *sql.Tx the repository methods need, so
// one method body serves both the plain and the transactional Store.
type queryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Store owns the SQLite connection and implements the repository interfaces
// declared by internal/domain. Transaction boundaries belong to internal/app,
// which reaches them through InTx.
type Store struct {
	db *sql.DB
	tx *sql.Tx // non-nil on the Store handed to an InTx callback
}

// Open opens or creates the database at path and brings it up to the latest
// migration. Opening a fresh file and reopening an existing one are both valid.
func Open(path string) (*Store, error) {
	// Foreign keys are off by default in SQLite and must be asked for per
	// connection; busy_timeout keeps a stray second process from failing hard.
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	// One connection: a single-operator TUI has no concurrency to gain, and
	// SQLite writers serialise anyway.
	db.SetMaxOpenConns(1)

	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// MustOpen is Open for callers with nothing useful to do about a failure —
// main, and the test harness.
func MustOpen(path string) *Store {
	s, err := Open(path)
	if err != nil {
		panic(err)
	}
	return s
}

// Close releases the connection.
func (s *Store) Close() error { return s.db.Close() }

// InTx runs fn against a transaction-scoped Store and commits if it returns nil.
// Every use case that writes more than one row goes through here, so a print and
// its filament usage rows land together or not at all. Nesting is a no-op: the
// outermost InTx owns the boundary.
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
	if err := fn(&Store{db: s.db, tx: tx}); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// q is the handle repository methods run their statements against.
func (s *Store) q() queryer {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}
