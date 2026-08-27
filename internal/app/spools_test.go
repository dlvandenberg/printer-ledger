package app_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

func TestAddSpoolThenList(t *testing.T) {
	a := newApp(t)

	added, err := a.AddSpool(ctx(), plaSpool())
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	if added.ID == 0 {
		t.Error("expected the added spool to have an ID")
	}

	spools, err := a.ListSpools(ctx())
	if err != nil {
		t.Fatalf("ListSpools: %v", err)
	}
	if len(spools) != 1 {
		t.Fatalf("got %d spools, want 1", len(spools))
	}

	got := spools[0]
	if got.ID != added.ID {
		t.Errorf("ID = %d, want %d", got.ID, added.ID)
	}
	if got.FilamentType != domain.PLA {
		t.Errorf("FilamentType = %q, want PLA", got.FilamentType)
	}
	if got.Brand != "Bambu" || got.Color != "Black" {
		t.Errorf("Brand/Color = %q/%q, want Bambu/Black", got.Brand, got.Color)
	}
	if got.InitialGrams != 1000 {
		t.Errorf("InitialGrams = %d, want 1000", got.InitialGrams)
	}
	if got.TareGrams != 210 {
		t.Errorf("TareGrams = %d, want 210", got.TareGrams)
	}
	if got.PurchaseCost != 2200 {
		t.Errorf("PurchaseCost = %d, want 2200", got.PurchaseCost)
	}
	if !got.PurchaseDate.Equal(date(t, "2026-08-01")) {
		t.Errorf("PurchaseDate = %s, want 2026-08-01", domain.FormatDate(got.PurchaseDate))
	}
}

func TestSpoolRemainingWithNoEvents(t *testing.T) {
	a := newApp(t)

	if _, err := a.AddSpool(ctx(), plaSpool()); err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	spools, err := a.ListSpools(ctx())
	if err != nil {
		t.Fatalf("ListSpools: %v", err)
	}

	got := spools[0]
	if got.RemainingGrams != 1000 {
		t.Errorf("RemainingGrams = %d, want 1000", got.RemainingGrams)
	}
	if got.RemainingValue != 2200 {
		t.Errorf("RemainingValue = %d, want 2200", got.RemainingValue)
	}
	if got.State != domain.SpoolActive {
		t.Errorf("State = %q, want active", got.State)
	}
}

func TestSpoolRemainingValueIsPerSpoolPrice(t *testing.T) {
	a := newApp(t)

	cheap := plaSpool()
	cheap.Brand, cheap.PurchaseCost, cheap.InitialGrams = "Cheap", 1500, 1000

	dear := plaSpool()
	dear.Brand, dear.PurchaseCost, dear.InitialGrams = "Dear", 3000, 750

	for _, cmd := range []app.AddSpoolCmd{cheap, dear} {
		if _, err := a.AddSpool(ctx(), cmd); err != nil {
			t.Fatalf("AddSpool %s: %v", cmd.Brand, err)
		}
	}

	spools, err := a.ListSpools(ctx())
	if err != nil {
		t.Fatalf("ListSpools: %v", err)
	}

	byBrand := map[string]app.SpoolView{}
	for _, s := range spools {
		byBrand[s.Brand] = s
	}
	if got := byBrand["Cheap"].RemainingValue; got != 1500 {
		t.Errorf("Cheap RemainingValue = %d, want 1500", got)
	}
	if got := byBrand["Dear"].RemainingValue; got != 3000 {
		t.Errorf("Dear RemainingValue = %d, want 3000", got)
	}
}

func TestAddSpoolValidation(t *testing.T) {
	tests := []struct {
		name  string
		mut   func(*app.AddSpoolCmd)
		field string
	}{
		{"unknown filament type", func(c *app.AddSpoolCmd) { c.FilamentType = "ABS" }, domain.FieldFilamentType},
		{"empty filament type", func(c *app.AddSpoolCmd) { c.FilamentType = "" }, domain.FieldFilamentType},
		{"missing brand", func(c *app.AddSpoolCmd) { c.Brand = "  " }, domain.FieldBrand},
		{"missing color", func(c *app.AddSpoolCmd) { c.Color = "" }, domain.FieldColor},
		{"zero initial grams", func(c *app.AddSpoolCmd) { c.InitialGrams = 0 }, domain.FieldInitialGrams},
		{"negative initial grams", func(c *app.AddSpoolCmd) { c.InitialGrams = -1 }, domain.FieldInitialGrams},
		{"negative tare grams", func(c *app.AddSpoolCmd) { c.TareGrams = -1 }, domain.FieldTareGrams},
		{"negative purchase cost", func(c *app.AddSpoolCmd) { c.PurchaseCost = -1 }, domain.FieldPurchaseCost},
		{"missing purchase date", func(c *app.AddSpoolCmd) { c.PurchaseDate = time.Time{} }, domain.FieldPurchaseDate},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t)
			cmd := plaSpool()
			tc.mut(&cmd)

			if _, err := a.AddSpool(ctx(), cmd); err == nil {
				t.Fatal("expected AddSpool to be rejected")
			} else if msg := fieldError(t, err, tc.field); msg == "" {
				t.Errorf("no error reported against field %q: %v", tc.field, err)
			}

			spools, err := a.ListSpools(ctx())
			if err != nil {
				t.Fatalf("ListSpools: %v", err)
			}
			if len(spools) != 0 {
				t.Errorf("got %d spools, want none persisted", len(spools))
			}
		})
	}
}

func TestAddSpoolAcceptsFreeSpool(t *testing.T) {
	a := newApp(t)

	cmd := plaSpool()
	cmd.PurchaseCost = 0

	added, err := a.AddSpool(ctx(), cmd)
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	if added.RemainingValue != 0 {
		t.Errorf("RemainingValue = %d, want 0", added.RemainingValue)
	}
	if added.RemainingGrams != 1000 {
		t.Errorf("RemainingGrams = %d, want 1000", added.RemainingGrams)
	}
}

func TestAddSpoolTrimsIdentification(t *testing.T) {
	a := newApp(t)

	cmd := plaSpool()
	cmd.Brand, cmd.Color = "  Bambu  ", "  Black  "

	added, err := a.AddSpool(ctx(), cmd)
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	if added.Brand != "Bambu" || added.Color != "Black" {
		t.Errorf("Brand/Color = %q/%q, want Bambu/Black", added.Brand, added.Color)
	}
}

func TestSeparateDatabasesAreSeparateLedgers(t *testing.T) {
	dir := t.TempDir()

	first := newAppAt(t, filepath.Join(dir, "one.db"))
	if _, err := first.AddSpool(ctx(), plaSpool()); err != nil {
		t.Fatalf("AddSpool: %v", err)
	}

	second := newAppAt(t, filepath.Join(dir, "two.db"))
	spools, err := second.ListSpools(ctx())
	if err != nil {
		t.Fatalf("ListSpools: %v", err)
	}
	if len(spools) != 0 {
		t.Errorf("got %d spools in the second ledger, want 0", len(spools))
	}
}

func TestReopeningAnExistingDatabaseIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reopen.db")

	first := newAppAt(t, path)
	added, err := first.AddSpool(ctx(), plaSpool())
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}

	second := newAppAt(t, path)
	spools, err := second.ListSpools(ctx())
	if err != nil {
		t.Fatalf("ListSpools after reopen: %v", err)
	}
	if len(spools) != 1 {
		t.Fatalf("got %d spools after reopen, want 1", len(spools))
	}
	if spools[0].ID != added.ID {
		t.Errorf("ID = %d, want %d", spools[0].ID, added.ID)
	}
}
