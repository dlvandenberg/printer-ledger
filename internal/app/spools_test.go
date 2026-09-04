package app_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
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
		t.Errorf("PurchaseDate = %s, want 2026-08-01", unit.FormatDate(got.PurchaseDate))
	}
}

func TestAddSpoolParsesFilamentType(t *testing.T) {
	tests := []struct {
		name  string
		typed string
		want  domain.FilamentType
	}{
		{"exact", "PETG", domain.PETG},
		{"lowercase", "pla+", domain.PLAPlus},
		{"padded", "  pla  ", domain.PLA},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t)
			cmd := plaSpool()
			cmd.FilamentType = tc.typed

			added, err := a.AddSpool(ctx(), cmd)
			if err != nil {
				t.Fatalf("AddSpool: %v", err)
			}
			if added.FilamentType != tc.want {
				t.Errorf("FilamentType = %q, want %q", added.FilamentType, tc.want)
			}
		})
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
	cheap.Brand, cheap.PurchaseCost, cheap.InitialGrams = "Cheap", "15.00", "1000"

	dear := plaSpool()
	dear.Brand, dear.PurchaseCost, dear.InitialGrams = "Dear", "30.00", "750"

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
		{"zero initial grams", func(c *app.AddSpoolCmd) { c.InitialGrams = "0" }, domain.FieldInitialGrams},
		{"negative initial grams", func(c *app.AddSpoolCmd) { c.InitialGrams = "-1" }, domain.FieldInitialGrams},
		{"negative tare grams", func(c *app.AddSpoolCmd) { c.TareGrams = "-1" }, domain.FieldTareGrams},
		{"negative purchase cost", func(c *app.AddSpoolCmd) { c.PurchaseCost = "-0.01" }, domain.FieldPurchaseCost},
		{"empty purchase date", func(c *app.AddSpoolCmd) { c.PurchaseDate = "" }, domain.FieldPurchaseDate},
		{"malformed purchase cost", func(c *app.AddSpoolCmd) { c.PurchaseCost = "22.000" }, domain.FieldPurchaseCost},
		{"empty purchase cost", func(c *app.AddSpoolCmd) { c.PurchaseCost = "" }, domain.FieldPurchaseCost},
		{"fractional initial grams", func(c *app.AddSpoolCmd) { c.InitialGrams = "1.5g" }, domain.FieldInitialGrams},
		{"malformed tare grams", func(c *app.AddSpoolCmd) { c.TareGrams = "heavy" }, domain.FieldTareGrams},
		{"out of range purchase date", func(c *app.AddSpoolCmd) { c.PurchaseDate = "2026-13-01" }, domain.FieldPurchaseDate},
		{"malformed purchase date", func(c *app.AddSpoolCmd) { c.PurchaseDate = "01-08-2026" }, domain.FieldPurchaseDate},
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
	cmd.PurchaseCost = "0"

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

func TestAddSpoolAcceptsCommaDecimalPrice(t *testing.T) {
	a := newApp(t)

	cmd := plaSpool()
	cmd.PurchaseCost = "22,00"

	added, err := a.AddSpool(ctx(), cmd)
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	if added.PurchaseCost != 2200 {
		t.Errorf("PurchaseCost = %d, want 2200", added.PurchaseCost)
	}
}

func TestAddSpoolReportsParseFailuresAndInvariantsTogether(t *testing.T) {
	a := newApp(t)

	cmd := plaSpool()
	cmd.PurchaseCost = "22.000"
	cmd.Brand = "  "

	_, err := a.AddSpool(ctx(), cmd)
	if err == nil {
		t.Fatal("expected AddSpool to be rejected")
	}
	if msg := fieldError(t, err, domain.FieldPurchaseCost); msg == "" {
		t.Error("no error reported against the purchase cost")
	}
	if msg := fieldError(t, err, domain.FieldBrand); msg == "" {
		t.Error("no error reported against the brand")
	}
}

func TestAddSpoolReportsTheParseFailureNotTheInvariantItTrips(t *testing.T) {
	a := newApp(t)

	cmd := plaSpool()
	cmd.InitialGrams = "1.5g"

	_, err := a.AddSpool(ctx(), cmd)
	if err == nil {
		t.Fatal("expected AddSpool to be rejected")
	}
	if msg := fieldError(t, err, domain.FieldInitialGrams); msg != unit.ErrMalformedGrams.Error() {
		t.Errorf("initialGrams error = %q, want %q", msg, unit.ErrMalformedGrams)
	}
}

func TestReweighSpoolDerivesRemainingFromTare(t *testing.T) {
	a := newApp(t)

	spool, err := a.AddSpool(ctx(), plaSpool())
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}

	detail, err := a.ReweighSpool(ctx(), reweigh(spool.ID, "610"))
	if err != nil {
		t.Fatalf("ReweighSpool: %v", err)
	}
	if detail.Spool.RemainingGrams != 400 {
		t.Errorf("RemainingGrams = %d, want 400", detail.Spool.RemainingGrams)
	}
	if len(detail.Adjustments) != 1 {
		t.Fatalf("got %d adjustments, want 1", len(detail.Adjustments))
	}

	got := detail.Adjustments[0]
	if got.MeasuredGrams != 610 {
		t.Errorf("MeasuredGrams = %d, want 610", got.MeasuredGrams)
	}
	if got.DerivedRemaining != 400 {
		t.Errorf("DerivedRemaining = %d, want 400", got.DerivedRemaining)
	}
	if got.DeltaGrams != -600 {
		t.Errorf("DeltaGrams = %d, want -600", got.DeltaGrams)
	}
	if !got.AdjustedOn.Equal(date(t, "2026-08-20")) {
		t.Errorf("AdjustedOn = %s, want 2026-08-20", unit.FormatDate(got.AdjustedOn))
	}
	if got.Note != "purge tower" {
		t.Errorf("Note = %q, want %q", got.Note, "purge tower")
	}
}

func TestSpoolDetailExplainsItsRemaining(t *testing.T) {
	a := newApp(t)

	spool, err := a.AddSpool(ctx(), plaSpool())
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	if _, err := a.ReweighSpool(ctx(), reweigh(spool.ID, "610")); err != nil {
		t.Fatalf("ReweighSpool: %v", err)
	}

	detail, err := a.SpoolDetail(ctx(), spool.ID)
	if err != nil {
		t.Fatalf("SpoolDetail: %v", err)
	}
	if detail.Spool.InitialGrams != 1000 {
		t.Errorf("InitialGrams = %d, want 1000", detail.Spool.InitialGrams)
	}
	if detail.Spool.UsedGrams != 0 {
		t.Errorf("UsedGrams = %d, want 0", detail.Spool.UsedGrams)
	}
	if detail.Spool.AdjustedGrams != -600 {
		t.Errorf("AdjustedGrams = %d, want -600", detail.Spool.AdjustedGrams)
	}
	if detail.Spool.RemainingGrams != 400 {
		t.Errorf("RemainingGrams = %d, want 400", detail.Spool.RemainingGrams)
	}
	if detail.Spool.RemainingValue != 880 {
		t.Errorf("RemainingValue = %d, want 880", detail.Spool.RemainingValue)
	}
	if len(detail.Adjustments) != 1 {
		t.Errorf("got %d adjustments, want 1", len(detail.Adjustments))
	}
}

func TestSpoolDetailReportsAnUnknownSpool(t *testing.T) {
	a := newApp(t)

	if _, err := a.SpoolDetail(ctx(), 404); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("SpoolDetail error = %v, want ErrNotFound", err)
	}
}

func TestSuccessiveReweighsProduceTheCorrectRemaining(t *testing.T) {
	a := newApp(t)

	spool, err := a.AddSpool(ctx(), plaSpool())
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}

	weights := []struct {
		measured  string
		remaining unit.Grams
		delta     unit.Grams
	}{
		{"810", 600, -400},
		{"560", 350, -250},
		{"460", 250, -100},
	}
	for _, w := range weights {
		detail, err := a.ReweighSpool(ctx(), reweigh(spool.ID, w.measured))
		if err != nil {
			t.Fatalf("ReweighSpool %s: %v", w.measured, err)
		}
		if detail.Spool.RemainingGrams != w.remaining {
			t.Errorf("after %sg RemainingGrams = %d, want %d", w.measured, detail.Spool.RemainingGrams, w.remaining)
		}
	}

	detail, err := a.SpoolDetail(ctx(), spool.ID)
	if err != nil {
		t.Fatalf("SpoolDetail: %v", err)
	}
	if len(detail.Adjustments) != len(weights) {
		t.Fatalf("got %d adjustments, want %d", len(detail.Adjustments), len(weights))
	}
	for i, w := range weights {
		if got := detail.Adjustments[i].DeltaGrams; got != w.delta {
			t.Errorf("adjustment %d DeltaGrams = %d, want %d", i, got, w.delta)
		}
	}
	if detail.Spool.RemainingGrams != 250 {
		t.Errorf("RemainingGrams = %d, want 250", detail.Spool.RemainingGrams)
	}
}

func TestReweighBelowTheEmptySpoolWeightIsRejected(t *testing.T) {
	a := newApp(t)

	spool, err := a.AddSpool(ctx(), plaSpool())
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}

	if _, err := a.ReweighSpool(ctx(), reweigh(spool.ID, "209")); err == nil {
		t.Fatal("expected ReweighSpool to be rejected")
	} else if msg := fieldError(t, err, domain.FieldMeasuredGrams); msg == "" {
		t.Errorf("no error reported against the measured grams: %v", err)
	}

	detail, err := a.SpoolDetail(ctx(), spool.ID)
	if err != nil {
		t.Fatalf("SpoolDetail: %v", err)
	}
	if len(detail.Adjustments) != 0 {
		t.Errorf("got %d adjustments, want none persisted", len(detail.Adjustments))
	}
	if detail.Spool.RemainingGrams != 1000 {
		t.Errorf("RemainingGrams = %d, want 1000", detail.Spool.RemainingGrams)
	}
}

func TestReweighSpoolValidation(t *testing.T) {
	tests := []struct {
		name  string
		mut   func(*app.ReweighSpoolCmd)
		field string
	}{
		{"malformed measured grams", func(c *app.ReweighSpoolCmd) { c.MeasuredGrams = "half" }, domain.FieldMeasuredGrams},
		{"fractional measured grams", func(c *app.ReweighSpoolCmd) { c.MeasuredGrams = "610.5" }, domain.FieldMeasuredGrams},
		{"empty measured grams", func(c *app.ReweighSpoolCmd) { c.MeasuredGrams = "" }, domain.FieldMeasuredGrams},
		{"negative measured grams", func(c *app.ReweighSpoolCmd) { c.MeasuredGrams = "-1" }, domain.FieldMeasuredGrams},
		{"empty date", func(c *app.ReweighSpoolCmd) { c.AdjustedOn = "" }, domain.FieldAdjustedOn},
		{"malformed date", func(c *app.ReweighSpoolCmd) { c.AdjustedOn = "20-08-2026" }, domain.FieldAdjustedOn},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t)
			spool, err := a.AddSpool(ctx(), plaSpool())
			if err != nil {
				t.Fatalf("AddSpool: %v", err)
			}
			cmd := reweigh(spool.ID, "610")
			tc.mut(&cmd)

			if _, err := a.ReweighSpool(ctx(), cmd); err == nil {
				t.Fatal("expected ReweighSpool to be rejected")
			} else if msg := fieldError(t, err, tc.field); msg == "" {
				t.Errorf("no error reported against field %q: %v", tc.field, err)
			}

			detail, err := a.SpoolDetail(ctx(), spool.ID)
			if err != nil {
				t.Fatalf("SpoolDetail: %v", err)
			}
			if len(detail.Adjustments) != 0 {
				t.Errorf("got %d adjustments, want none persisted", len(detail.Adjustments))
			}
		})
	}
}

func TestReweighAnUnknownSpoolIsReported(t *testing.T) {
	a := newApp(t)

	if _, err := a.ReweighSpool(ctx(), reweigh(404, "610")); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ReweighSpool error = %v, want ErrNotFound", err)
	}
}

func TestReweighAcceptsAnEmptyNote(t *testing.T) {
	a := newApp(t)

	spool, err := a.AddSpool(ctx(), plaSpool())
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	cmd := reweigh(spool.ID, "610")
	cmd.Note = "  "

	detail, err := a.ReweighSpool(ctx(), cmd)
	if err != nil {
		t.Fatalf("ReweighSpool: %v", err)
	}
	if got := detail.Adjustments[0].Note; got != "" {
		t.Errorf("Note = %q, want empty", got)
	}
}

func TestAReweighToTheTareEmptiesTheSpool(t *testing.T) {
	a := newApp(t)

	empty, err := a.AddSpool(ctx(), plaSpool())
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}
	full := plaSpool()
	full.Color = "Grey"
	if _, err := a.AddSpool(ctx(), full); err != nil {
		t.Fatalf("AddSpool: %v", err)
	}

	detail, err := a.ReweighSpool(ctx(), reweigh(empty.ID, "210"))
	if err != nil {
		t.Fatalf("ReweighSpool: %v", err)
	}
	if detail.Spool.RemainingGrams != 0 {
		t.Errorf("RemainingGrams = %d, want 0", detail.Spool.RemainingGrams)
	}
	if detail.Spool.State != domain.SpoolEmpty {
		t.Errorf("State = %q, want empty", detail.Spool.State)
	}

	all, err := a.ListSpools(ctx())
	if err != nil {
		t.Fatalf("ListSpools: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("got %d spools in the list, want both", len(all))
	}

	active, err := a.ListActiveSpools(ctx())
	if err != nil {
		t.Fatalf("ListActiveSpools: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("got %d active spools, want 1", len(active))
	}
	if active[0].Color != "Grey" {
		t.Errorf("active spool colour = %q, want Grey", active[0].Color)
	}
}

func TestSpoolDetailListsAdjustmentsAsRecorded(t *testing.T) {
	a := newApp(t)

	spool, err := a.AddSpool(ctx(), plaSpool())
	if err != nil {
		t.Fatalf("AddSpool: %v", err)
	}

	first := reweigh(spool.ID, "810")
	first.AdjustedOn = "2026-08-20"
	second := reweigh(spool.ID, "560")
	second.AdjustedOn = "2026-08-10"
	for _, cmd := range []app.ReweighSpoolCmd{first, second} {
		if _, err := a.ReweighSpool(ctx(), cmd); err != nil {
			t.Fatalf("ReweighSpool %s: %v", cmd.MeasuredGrams, err)
		}
	}

	detail, err := a.SpoolDetail(ctx(), spool.ID)
	if err != nil {
		t.Fatalf("SpoolDetail: %v", err)
	}
	if len(detail.Adjustments) != 2 {
		t.Fatalf("got %d adjustments, want 2", len(detail.Adjustments))
	}
	if got := detail.Adjustments[0].MeasuredGrams; got != 810 {
		t.Errorf("first MeasuredGrams = %d, want 810", got)
	}
	if got := detail.Adjustments[1].DeltaGrams; got != -250 {
		t.Errorf("second DeltaGrams = %d, want -250", got)
	}
	if got := detail.Adjustments[1].DerivedRemaining; got != detail.Spool.RemainingGrams {
		t.Errorf("last DerivedRemaining = %d, want the spool's remaining %d", got, detail.Spool.RemainingGrams)
	}
}

func TestDeleteSpoolBlockedWhenAPrintDrewFromIt(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())
	recordedPrint(t, a, printOf(design.ID, spool.ID))

	err := a.DeleteSpool(ctx(), spool.ID)
	if !errors.Is(err, domain.ErrSpoolDrawnFrom) {
		t.Fatalf("DeleteSpool error = %v, want ErrSpoolDrawnFrom", err)
	}

	spools, err := a.ListSpools(ctx())
	if err != nil {
		t.Fatalf("ListSpools: %v", err)
	}
	if len(spools) != 1 {
		t.Errorf("ListSpools returned %d spools, want 1", len(spools))
	}
}

func TestDeleteSpoolTakesItsAdjustments(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	if _, err := a.ReweighSpool(ctx(), reweigh(spool.ID, "1100")); err != nil {
		t.Fatalf("ReweighSpool: %v", err)
	}

	if err := a.DeleteSpool(ctx(), spool.ID); err != nil {
		t.Fatalf("DeleteSpool: %v", err)
	}

	spools, err := a.ListSpools(ctx())
	if err != nil {
		t.Fatalf("ListSpools: %v", err)
	}
	if len(spools) != 0 {
		t.Errorf("ListSpools returned %d spools, want 0", len(spools))
	}
}
