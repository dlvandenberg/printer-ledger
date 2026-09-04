package app_test

import (
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

// TestRecordPrintPinsTheReferenceCase is the case CONTEXT.md works through: a
// PLA spool at €22.00 per 1000g, 120g over 5h30m at quantity 2, €0.28 per kWh,
// 0.09 kWh/h and €0.35/h machine.
func TestRecordPrintPinsTheReferenceCase(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	view, err := a.RecordPrint(ctx(), printOf(design.ID, spool.ID))
	if err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}

	if view.Cost.Filament != 264 {
		t.Errorf("Filament = %d, want 264", view.Cost.Filament)
	}
	if view.Cost.Energy != 14 {
		t.Errorf("Energy = %d, want 14", view.Cost.Energy)
	}
	if view.Cost.Overhead != 193 {
		t.Errorf("Overhead = %d, want 193", view.Cost.Overhead)
	}
	if view.Cost.JobCost != 471 {
		t.Errorf("JobCost = %d, want 471", view.Cost.JobCost)
	}
	if view.Cost.CostPerCopy != 235 {
		t.Errorf("CostPerCopy = %d, want 235", view.Cost.CostPerCopy)
	}
}

func TestRecordPrintThenList(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	if _, err := a.RecordPrint(ctx(), printOf(design.ID, spool.ID)); err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}

	prints, err := a.ListPrints(ctx())
	if err != nil {
		t.Fatalf("ListPrints: %v", err)
	}
	if len(prints) != 1 {
		t.Fatalf("ListPrints returned %d prints, want 1", len(prints))
	}

	got := prints[0]
	if got.DesignName != "Planter" {
		t.Errorf("DesignName = %q, want \"Planter\"", got.DesignName)
	}
	if got.Quantity != 2 {
		t.Errorf("Quantity = %d, want 2", got.Quantity)
	}
	if got.Minutes != 330 {
		t.Errorf("Minutes = %d, want 330", got.Minutes)
	}
	if got.Cost.JobCost != 471 {
		t.Errorf("JobCost = %d, want 471", got.Cost.JobCost)
	}
	if len(got.Usages) != 1 {
		t.Fatalf("returned %d usage rows, want 1", len(got.Usages))
	}
	if got.Usages[0].Grams != 120 {
		t.Errorf("Grams = %d, want 120", got.Usages[0].Grams)
	}
	if got.Usages[0].CostPerGram != 220 {
		t.Errorf("CostPerGram = %d, want 220", got.Usages[0].CostPerGram)
	}
}

func TestPrintDraftPrefillsEstimatesTimesQuantity(t *testing.T) {
	a := newApp(t)
	design := addedDesign(t, a, quotedDesign())

	draft, err := a.PrintDraft(ctx(), design.ID, "2")
	if err != nil {
		t.Fatalf("PrintDraft: %v", err)
	}
	if draft.Grams != 120 {
		t.Errorf("Grams = %d, want 120", draft.Grams)
	}
	if draft.Minutes != 330 {
		t.Errorf("Minutes = %d, want 330", draft.Minutes)
	}
	if draft.DesignName != "Planter" {
		t.Errorf("DesignName = %q, want \"Planter\"", draft.DesignName)
	}
}

func TestPreviewPrintCostsWithoutRecording(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	preview, err := a.PreviewPrint(ctx(), printOf(design.ID, spool.ID))
	if err != nil {
		t.Fatalf("PreviewPrint: %v", err)
	}
	if preview.Cost.JobCost != 471 {
		t.Errorf("JobCost = %d, want 471", preview.Cost.JobCost)
	}
	if preview.Cost.CostPerCopy != 235 {
		t.Errorf("CostPerCopy = %d, want 235", preview.Cost.CostPerCopy)
	}

	prints, err := a.ListPrints(ctx())
	if err != nil {
		t.Fatalf("ListPrints: %v", err)
	}
	if len(prints) != 0 {
		t.Errorf("ListPrints returned %d prints, want 0", len(prints))
	}
	if remaining := remainingOf(t, a, spool.ID); remaining != 1000 {
		t.Errorf("RemainingGrams = %d, want 1000", remaining)
	}
}

func TestPreviewPrintCostsWhatItCanWhileTheFormIsHalfTyped(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	cmd := printOf(design.ID, spool.ID)
	cmd.Minutes = "5:"
	cmd.Usages[0].Grams = ""

	preview, err := a.PreviewPrint(ctx(), cmd)
	if err != nil {
		t.Fatalf("PreviewPrint: %v", err)
	}
	if preview.Cost.JobCost != 0 {
		t.Errorf("JobCost = %d, want 0", preview.Cost.JobCost)
	}
}

func TestRecordPrintDropsSpoolRemaining(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	if _, err := a.RecordPrint(ctx(), printOf(design.ID, spool.ID)); err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}

	detail, err := a.SpoolDetail(ctx(), spool.ID)
	if err != nil {
		t.Fatalf("SpoolDetail: %v", err)
	}
	if detail.Spool.UsedGrams != 120 {
		t.Errorf("UsedGrams = %d, want 120", detail.Spool.UsedGrams)
	}
	if detail.Spool.RemainingGrams != 880 {
		t.Errorf("RemainingGrams = %d, want 880", detail.Spool.RemainingGrams)
	}
}

func TestRecordPrintRejectsOverdrawNamingWhatIsLeft(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	if _, err := a.ReweighSpool(ctx(), reweigh(spool.ID, "250")); err != nil {
		t.Fatalf("ReweighSpool: %v", err)
	}

	_, err := a.RecordPrint(ctx(), printOf(design.ID, spool.ID))
	if got, want := fieldError(t, err, domain.FieldUsageGrams(0)), "only 40g left"; got != want {
		t.Errorf("usage grams error = %q, want %q", got, want)
	}
	if remaining := remainingOf(t, a, spool.ID); remaining != 40 {
		t.Errorf("RemainingGrams = %d, want 40", remaining)
	}
}

func TestRecordPrintRejectsQuantityBelowOne(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	tests := []struct {
		name     string
		quantity string
		want     string
	}{
		{name: "zero", quantity: "0", want: "must be at least 1"},
		{name: "negative", quantity: "-1", want: "must be at least 1"},
		{name: "malformed", quantity: "some", want: unit.ErrMalformedCopies.Error()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := printOf(design.ID, spool.ID)
			cmd.Quantity = tc.quantity

			_, err := a.RecordPrint(ctx(), cmd)
			if got := fieldError(t, err, domain.FieldQuantity); got != tc.want {
				t.Errorf("quantity error = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRecordPrintCanBeBackdated(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	cmd := printOf(design.ID, spool.ID)
	cmd.Date = "2026-07-15"

	view, err := a.RecordPrint(ctx(), cmd)
	if err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}
	if got := unit.FormatDate(view.Date); got != "2026-07-15" {
		t.Errorf("Date = %q, want \"2026-07-15\"", got)
	}
}

func TestRecordPrintKeepsItsCostWhenSettingsChange(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	recorded, err := a.RecordPrint(ctx(), printOf(design.ID, spool.ID))
	if err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}

	settings := settingsUpdate()
	settings.KwhPrice = "0.50"
	settings.MachineHourlyRate = "1.00"
	if _, err := a.UpdateSettings(ctx(), settings); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	if got := printByID(t, a, recorded.ID); got.Cost.JobCost != 471 {
		t.Errorf("JobCost of the recorded print = %d, want 471", got.Cost.JobCost)
	}

	later, err := a.RecordPrint(ctx(), printOf(design.ID, spool.ID))
	if err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}
	if later.Cost.JobCost != 839 {
		t.Errorf("JobCost of the later print = %d, want 839", later.Cost.JobCost)
	}
}

func TestEditSpoolRejectsAnInitialWeightBelowWhatPrintsTookOff(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	if _, err := a.RecordPrint(ctx(), printOf(design.ID, spool.ID)); err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}

	shrunk := editOfSpool(spool)
	shrunk.InitialGrams = "100"

	_, err := a.EditSpool(ctx(), shrunk)
	if err == nil {
		t.Fatal("expected EditSpool to be rejected")
	}
	if got, want := fieldError(t, err, domain.FieldInitialGrams), "cannot be less than the 120g already off the spool"; got != want {
		t.Errorf("initialGrams error = %q, want %q", got, want)
	}

	detail, err := a.SpoolDetail(ctx(), spool.ID)
	if err != nil {
		t.Fatalf("SpoolDetail: %v", err)
	}
	after := detail.Spool
	if after.InitialGrams != 1000 {
		t.Errorf("InitialGrams after the rejected edit = %d, want 1000", after.InitialGrams)
	}
	if after.RemainingGrams < 0 {
		t.Errorf("RemainingGrams = %d, want no negative remaining", after.RemainingGrams)
	}
}

func TestRecordPrintKeepsItsCostWhenTheSpoolPriceIsCorrected(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	recorded, err := a.RecordPrint(ctx(), printOf(design.ID, spool.ID))
	if err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}

	corrected := editOfSpool(spool)
	corrected.PurchaseCost = "44.00"
	if _, err := a.EditSpool(ctx(), corrected); err != nil {
		t.Fatalf("EditSpool: %v", err)
	}

	if got := printByID(t, a, recorded.ID); got.Cost.Filament != 264 {
		t.Errorf("Filament of the recorded print = %d, want 264", got.Cost.Filament)
	}

	later, err := a.RecordPrint(ctx(), printOf(design.ID, spool.ID))
	if err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}
	if later.Cost.Filament != 528 {
		t.Errorf("Filament of the later print = %d, want 528", later.Cost.Filament)
	}
}

func TestRecordPrintSumsFilamentAcrossUsageRows(t *testing.T) {
	a := newApp(t)
	emptied := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	replacement := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "30.00"))
	design := addedDesign(t, a, quotedDesign())

	view, err := a.RecordPrint(ctx(), printOfRows(design.ID,
		usage(emptied.ID, "40"), usage(replacement.ID, "80")))
	if err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}

	if len(view.Usages) != 2 {
		t.Fatalf("returned %d usage rows, want 2", len(view.Usages))
	}
	if view.UsedGrams != 120 {
		t.Errorf("UsedGrams = %d, want 120", view.UsedGrams)
	}
	if view.Cost.Filament != 328 {
		t.Errorf("Filament = %d, want 328", view.Cost.Filament)
	}
	if view.Cost.JobCost != 535 {
		t.Errorf("JobCost = %d, want 535", view.Cost.JobCost)
	}
	if got := remainingOf(t, a, emptied.ID); got != 960 {
		t.Errorf("emptied spool RemainingGrams = %d, want 960", got)
	}
	if got := remainingOf(t, a, replacement.ID); got != 920 {
		t.Errorf("replacement spool RemainingGrams = %d, want 920", got)
	}
}

func TestRecordPrintSplitsAcrossTwoSpoolsWhenOneWouldOverdraw(t *testing.T) {
	a := newApp(t)
	emptied := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	replacement := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	if _, err := a.ReweighSpool(ctx(), reweigh(emptied.ID, "250")); err != nil {
		t.Fatalf("ReweighSpool: %v", err)
	}

	if _, err := a.RecordPrint(ctx(), printOfRows(design.ID,
		usage(emptied.ID, "40"), usage(replacement.ID, "80"))); err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}

	if got := remainingOf(t, a, emptied.ID); got != 0 {
		t.Errorf("emptied spool RemainingGrams = %d, want 0", got)
	}
	if got := remainingOf(t, a, replacement.ID); got != 920 {
		t.Errorf("replacement spool RemainingGrams = %d, want 920", got)
	}
}

func TestRecordPrintValidatesEachRowAgainstItsOwnSpool(t *testing.T) {
	a := newApp(t)
	full := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	nearlyEmpty := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	if _, err := a.ReweighSpool(ctx(), reweigh(nearlyEmpty.ID, "250")); err != nil {
		t.Fatalf("ReweighSpool: %v", err)
	}

	_, err := a.RecordPrint(ctx(), printOfRows(design.ID,
		usage(full.ID, "40"), usage(nearlyEmpty.ID, "80")))
	if got, want := fieldError(t, err, domain.FieldUsageGrams(1)), "only 40g left"; got != want {
		t.Errorf("second row error = %q, want %q", got, want)
	}
	if got := fieldError(t, err, domain.FieldUsageGrams(0)); got != "" {
		t.Errorf("first row error = %q, want none", got)
	}
	if got := remainingOf(t, a, full.ID); got != 1000 {
		t.Errorf("full spool RemainingGrams = %d, want 1000", got)
	}
}

func TestRecordPrintRejectsTwoRowsOverdrawingOneSpool(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	if _, err := a.ReweighSpool(ctx(), reweigh(spool.ID, "310")); err != nil {
		t.Fatalf("ReweighSpool: %v", err)
	}

	_, err := a.RecordPrint(ctx(), printOfRows(design.ID,
		usage(spool.ID, "60"), usage(spool.ID, "60"), usage(spool.ID, "60")))
	if got, want := fieldError(t, err, domain.FieldUsageGrams(1)), "only 100g left"; got != want {
		t.Errorf("second row error = %q, want %q", got, want)
	}
	if got := fieldError(t, err, domain.FieldUsageGrams(2)); got != "" {
		t.Errorf("third row error = %q, want the overdraw reported once", got)
	}
	if got := remainingOf(t, a, spool.ID); got != 100 {
		t.Errorf("RemainingGrams = %d, want 100", got)
	}
}

func TestRecordPrintRejectsMixedFilamentTypes(t *testing.T) {
	a := newApp(t)
	pla := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	petg := addedSpool(t, a, spoolPriced(domain.PETG, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	_, err := a.RecordPrint(ctx(), printOfRows(design.ID,
		usage(pla.ID, "40"), usage(petg.ID, "80")))
	if got, want := fieldError(t, err, domain.FieldUsageSpool(1)), "cannot mix PETG with PLA on one print"; got != want {
		t.Errorf("second row error = %q, want %q", got, want)
	}

	prints, err := a.ListPrints(ctx())
	if err != nil {
		t.Fatalf("ListPrints: %v", err)
	}
	if len(prints) != 0 {
		t.Errorf("ListPrints returned %d prints, want 0", len(prints))
	}
}

func TestPreviewPrintFlagsASeededPowerRate(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	seeded, err := a.PreviewPrint(ctx(), printOf(design.ID, spool.ID))
	if err != nil {
		t.Fatalf("PreviewPrint: %v", err)
	}
	if seeded.FilamentType != domain.PLA {
		t.Errorf("FilamentType = %q, want PLA", seeded.FilamentType)
	}
	if !seeded.RateSeeded {
		t.Error("RateSeeded is false, want the seeded PLA rate flagged")
	}

	settings := settingsUpdate()
	settings.PowerRates[domain.PLA] = "0.11"
	if _, err := a.UpdateSettings(ctx(), settings); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	measured, err := a.PreviewPrint(ctx(), printOf(design.ID, spool.ID))
	if err != nil {
		t.Fatalf("PreviewPrint: %v", err)
	}
	if measured.RateSeeded {
		t.Error("RateSeeded is true, want the edited PLA rate flagged as measured")
	}
	if measured.Cost.Energy != 17 {
		t.Errorf("Energy = %d, want 17", measured.Cost.Energy)
	}
}

func TestPreviewPrintReportsNoRateForAnEmptyForm(t *testing.T) {
	a := newApp(t)
	design := addedDesign(t, a, quotedDesign())

	preview, err := a.PreviewPrint(ctx(), printOfRows(design.ID, usage(0, "")))
	if err != nil {
		t.Fatalf("PreviewPrint: %v", err)
	}
	if preview.FilamentType != "" {
		t.Errorf("FilamentType = %q, want empty", preview.FilamentType)
	}
	if preview.RateSeeded {
		t.Error("RateSeeded is true, want no warning with no spool picked")
	}
}
