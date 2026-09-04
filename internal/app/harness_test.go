package app_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
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
	d, err := unit.ParseDate(s)
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
		rates[rate.FilamentType] = unit.FormatKwhPerHour(rate.KwhPerHour)
	}
	return app.UpdateSettingsCmd{
		KwhPrice:            unit.FormatCents(seeded.KwhPrice),
		MachineHourlyRate:   unit.FormatCents(seeded.MachineHourlyRate),
		PrinterPurchaseCost: unit.FormatCents(seeded.PrinterPurchaseCost),
		DefaultMargin:       unit.FormatPercent(seeded.DefaultMargin),
		MinMargin:           unit.FormatPercent(seeded.MinMargin),
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

func reweigh(spoolID int64, measured string) app.ReweighSpoolCmd {
	return app.ReweighSpoolCmd{
		SpoolID:       spoolID,
		MeasuredGrams: measured,
		AdjustedOn:    "2026-08-20",
		Note:          "purge tower",
	}
}

func plaDesign() app.AddDesignCmd {
	return app.AddDesignCmd{
		Name:                "Cable clip",
		EstimatedGrams:      "48",
		EstimatedMinutes:    "5:31",
		DefaultFilamentType: domain.PLA.String(),
		MarginPct:           "",
	}
}

func editOf(d app.DesignView) app.EditDesignCmd {
	return app.EditDesignCmd{
		DesignID:            d.ID,
		Name:                d.Name,
		EstimatedGrams:      unit.FormatGrams(d.EstimatedGrams),
		EstimatedMinutes:    unit.FormatHHmm(d.EstimatedMinutes),
		DefaultFilamentType: d.DefaultFilamentType.String(),
		MarginPct:           unit.FormatPercent(d.MarginPct),
	}
}

func designNamed(t *testing.T, designs []app.DesignView, name string) app.DesignView {
	t.Helper()
	for _, d := range designs {
		if d.Name == name {
			return d
		}
	}
	t.Fatalf("no design named %q", name)
	return app.DesignView{}
}

// quotedDesign is the reference case pinned in CONTEXT.md: 60g and 2h45m per
// copy at the seeded rates.
func quotedDesign() app.AddDesignCmd {
	cmd := plaDesign()
	cmd.Name = "Planter"
	cmd.EstimatedGrams = "60"
	cmd.EstimatedMinutes = "2:45"
	cmd.MarginPct = "50"
	return cmd
}

func spoolPriced(ft domain.FilamentType, grams, cost string) app.AddSpoolCmd {
	cmd := plaSpool()
	cmd.FilamentType = ft.String()
	cmd.InitialGrams = grams
	cmd.PurchaseCost = cost
	return cmd
}

func quoteRow(t *testing.T, q app.DesignQuoteView, ft domain.FilamentType) app.EstimatedCostView {
	t.Helper()
	for _, cost := range q.Costs {
		if cost.FilamentType == ft {
			return cost
		}
	}
	t.Fatalf("no cost row for %s", ft)
	return app.EstimatedCostView{}
}

func hasQuoteRow(q app.DesignQuoteView, ft domain.FilamentType) bool {
	for _, cost := range q.Costs {
		if cost.FilamentType == ft {
			return true
		}
	}
	return false
}

func addedSpool(t *testing.T, a *app.App, cmd app.AddSpoolCmd) app.SpoolView {
	t.Helper()
	spool, err := a.AddSpool(ctx(), cmd)
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	return spool
}

func addedDesign(t *testing.T, a *app.App, cmd app.AddDesignCmd) app.DesignView {
	t.Helper()
	design, err := a.AddDesign(ctx(), cmd)
	if err != nil {
		t.Fatalf("AddDesign: %v", err)
	}
	return design
}

func editOfSpool(s app.SpoolView) app.EditSpoolCmd {
	return app.EditSpoolCmd{
		SpoolID:      s.ID,
		FilamentType: s.FilamentType.String(),
		Brand:        s.Brand,
		Color:        s.Color,
		InitialGrams: unit.FormatGrams(s.InitialGrams),
		TareGrams:    unit.FormatGrams(s.TareGrams),
		PurchaseCost: unit.FormatCents(s.PurchaseCost),
		PurchaseDate: unit.FormatDate(s.PurchaseDate),
	}
}

// printOf is the reference case of CONTEXT.md: 120g over 5h30m at quantity 2.
func printOf(designID, spoolID int64) app.RecordPrintCmd {
	return app.RecordPrintCmd{
		DesignID: designID,
		Date:     "2026-08-20",
		Quantity: "2",
		Minutes:  "5:30",
		Usages:   []app.FilamentUsageCmd{{SpoolID: spoolID, Grams: "120"}},
	}
}

func printByID(t *testing.T, a *app.App, id int64) app.PrintView {
	t.Helper()
	prints, err := a.ListPrints(ctx())
	if err != nil {
		t.Fatalf("ListPrints: %v", err)
	}
	for _, p := range prints {
		if p.ID == id {
			return p
		}
	}
	t.Fatalf("no print %d", id)
	return app.PrintView{}
}

func remainingOf(t *testing.T, a *app.App, spoolID int64) unit.Grams {
	t.Helper()
	detail, err := a.SpoolDetail(ctx(), spoolID)
	if err != nil {
		t.Fatalf("SpoolDetail: %v", err)
	}
	return detail.Spool.RemainingGrams
}

func printOfRows(designID int64, rows ...app.FilamentUsageCmd) app.RecordPrintCmd {
	cmd := printOf(designID, 0)
	cmd.Usages = rows
	return cmd
}

func usage(spoolID int64, grams string) app.FilamentUsageCmd {
	return app.FilamentUsageCmd{SpoolID: spoolID, Grams: grams}
}
