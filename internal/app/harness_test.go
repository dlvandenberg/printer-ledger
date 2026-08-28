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
		FilamentType: domain.PLA.String(),
		Brand:        "Bambu",
		Color:        "Black",
		InitialGrams: "1000",
		TareGrams:    "210",
		PurchaseCost: "22.00",
		PurchaseDate: "2026-08-01",
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

// settingsUpdate resubmits the seeded defaults unchanged, so a test that varies
// one field is the only thing that edits a rate.
func settingsUpdate() app.UpdateSettingsCmd {
	seeded := domain.DefaultSettings()
	rates := map[domain.FilamentType]string{}
	for _, rate := range seeded.PowerRates {
		rates[rate.FilamentType] = domain.FormatKwhPerHour(rate.KwhPerHour)
	}
	return app.UpdateSettingsCmd{
		KwhPrice:            domain.FormatCents(seeded.KwhPrice),
		MachineHourlyRate:   domain.FormatCents(seeded.MachineHourlyRate),
		PrinterPurchaseCost: domain.FormatCents(seeded.PrinterPurchaseCost),
		DefaultMargin:       domain.FormatPercent(seeded.DefaultMargin),
		MinMargin:           domain.FormatPercent(seeded.MinMargin),
		Currency:            seeded.Currency,
		PowerRates:          rates,
	}
}

func powerRate(t *testing.T, s app.SettingsView, ft domain.FilamentType) app.PowerRateView {
	t.Helper()
	for _, rate := range s.PowerRates {
		if rate.FilamentType == ft {
			return rate
		}
	}
	t.Fatalf("no power rate for %s", ft)
	return app.PowerRateView{}
}
