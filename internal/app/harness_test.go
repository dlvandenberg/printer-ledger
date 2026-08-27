package app_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/store"
)

func newApp(t *testing.T) *app.App {
	t.Helper()
	return newAppAt(t, filepath.Join(t.TempDir(), "test.db"))
}

func newAppAt(t *testing.T, path string) *app.App {
	t.Helper()
	st, err := store.Open(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return app.New(st)
}

func date(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := domain.ParseDate(s)
	if err != nil {
		t.Fatalf("parse date %q: %v", s, err)
	}
	return d
}

func plaSpool() app.AddSpoolCmd {
	return app.AddSpoolCmd{
		FilamentType: domain.PLA,
		Brand:        "Bambu",
		Color:        "Black",
		InitialGrams: 1000,
		TareGrams:    210,
		PurchaseCost: 2200,
		PurchaseDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	}
}

func fieldError(t *testing.T, err error, field string) string {
	t.Helper()
	var v *domain.ValidationError
	if !errors.As(err, &v) {
		t.Fatalf("expected a validation error, got %v", err)
	}
	return v.For(field)
}

func ctx() context.Context { return context.Background() }
