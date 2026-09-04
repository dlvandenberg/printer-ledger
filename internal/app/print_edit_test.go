package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

func TestEditPrintGramsRecomputesSpoolRemaining(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())
	print := recordedPrint(t, a, printOf(design.ID, spool.ID))

	if got := remainingOf(t, a, spool.ID); got != grams(880) {
		t.Fatalf("RemainingGrams = %d, want %d", got, grams(880))
	}

	cmd := editOfPrint(print)
	cmd.Usages = []app.FilamentUsageCmd{usage(spool.ID, "100")}
	if _, err := a.EditPrint(ctx(), cmd); err != nil {
		t.Fatalf("EditPrint: %v", err)
	}

	if got := remainingOf(t, a, spool.ID); got != grams(900) {
		t.Errorf("RemainingGrams = %d, want %d", got, grams(900))
	}
	if got := printByID(t, a, print.ID).UsedGrams; got != grams(100) {
		t.Errorf("UsedGrams = %d, want %d", got, grams(100))
	}
}

// An unchanged edit must pass its own overdraw check: the Spool's remaining
// already counts the grams this Print drew.
func TestEditPrintUnchangedKeepsSpoolRemaining(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())
	cmd := printOf(design.ID, spool.ID)
	cmd.Usages = []app.FilamentUsageCmd{usage(spool.ID, "1000")}
	print := recordedPrint(t, a, cmd)

	if _, err := a.EditPrint(ctx(), editOfPrint(print)); err != nil {
		t.Fatalf("EditPrint: %v", err)
	}
	if got := remainingOf(t, a, spool.ID); got != 0 {
		t.Errorf("RemainingGrams = %d, want 0", got)
	}
}

func TestEditPrintRejectsOverdrawingTheSpool(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())
	print := recordedPrint(t, a, printOf(design.ID, spool.ID))

	cmd := editOfPrint(print)
	cmd.Usages = []app.FilamentUsageCmd{usage(spool.ID, "1200")}

	_, err := a.EditPrint(ctx(), cmd)
	if err == nil {
		t.Fatalf("EditPrint accepted 1200g off a 1000g spool")
	}
	if got := fieldError(t, err, domain.FieldUsageGrams(0)); got == "" {
		t.Errorf("no error on %s, got %v", domain.FieldUsageGrams(0), err)
	}
	if got := remainingOf(t, a, spool.ID); got != grams(880) {
		t.Errorf("RemainingGrams = %d, want %d", got, grams(880))
	}
}

func TestEditPrintRejectsQuantityBelowAccountedForCopies(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())
	cmd := printOf(design.ID, spool.ID)
	cmd.Quantity = "4"
	cmd.Gifted = "2"
	cmd.Kept = "1"
	print := recordedPrint(t, a, cmd)

	edit := editOfPrint(print)
	edit.Quantity = "2"

	_, err := a.EditPrint(ctx(), edit)
	if err == nil {
		t.Fatalf("EditPrint shrank quantity to 2 below 3 accounted-for copies")
	}
	got := fieldError(t, err, domain.FieldQuantity)
	if !strings.Contains(got, "2 gifted") || !strings.Contains(got, "1 kept") {
		t.Errorf("%s error = %q, want it to name the gifted and kept copies", domain.FieldQuantity, got)
	}
	if quantity := printByID(t, a, print.ID).Quantity; quantity != 4 {
		t.Errorf("Quantity = %d, want 4", quantity)
	}
}

func TestEditPrintRecordsStockCounts(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())
	cmd := printOf(design.ID, spool.ID)
	cmd.Quantity = "4"
	print := recordedPrint(t, a, cmd)

	edit := editOfPrint(print)
	edit.Scrapped = "1"
	edit.Kept = "2"

	edited, err := a.EditPrint(ctx(), edit)
	if err != nil {
		t.Fatalf("EditPrint: %v", err)
	}
	if edited.AvailableCopies != 1 {
		t.Errorf("AvailableCopies = %d, want 1", edited.AvailableCopies)
	}
	if got := printByID(t, a, print.ID).ScrappedCount; got != 1 {
		t.Errorf("ScrappedCount = %d, want 1", got)
	}
}

// A past Print's cost never moves: the gram price is frozen on the usage row
// and the rates are snapshotted on the Print (ADR-0004).
func TestEditPrintKeepsItsFrozenCost(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())
	print := recordedPrint(t, a, printOf(design.ID, spool.ID))

	dearer := editOfSpool(spool)
	dearer.PurchaseCost = "44.00"
	if _, err := a.EditSpool(ctx(), dearer); err != nil {
		t.Fatalf("EditSpool: %v", err)
	}
	pricier := settingsUpdate()
	pricier.KwhPrice = "0.56"
	if _, err := a.UpdateSettings(ctx(), pricier); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	edited, err := a.EditPrint(ctx(), editOfPrint(print))
	if err != nil {
		t.Fatalf("EditPrint: %v", err)
	}
	if edited.Usages[0].CostPerGram != 220 {
		t.Errorf("CostPerGram = %d, want 220", edited.Usages[0].CostPerGram)
	}
	if edited.Cost.JobCost != 471 {
		t.Errorf("JobCost = %d, want 471", edited.Cost.JobCost)
	}
}

func TestDeletePrintReturnsFilamentToTheSpool(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())
	print := recordedPrint(t, a, printOf(design.ID, spool.ID))

	if err := a.DeletePrint(ctx(), print.ID); err != nil {
		t.Fatalf("DeletePrint: %v", err)
	}

	prints, err := a.ListPrints(ctx())
	if err != nil {
		t.Fatalf("ListPrints: %v", err)
	}
	if len(prints) != 0 {
		t.Errorf("ListPrints returned %d prints, want 0", len(prints))
	}
	if got := remainingOf(t, a, spool.ID); got != grams(1000) {
		t.Errorf("RemainingGrams = %d, want %d", got, grams(1000))
	}
}

func TestDeletePrintBlockedByAccountedForCopies(t *testing.T) {
	tests := []struct {
		name     string
		set      func(*app.RecordPrintCmd)
		inTheWay string
	}{
		{"gifted", func(c *app.RecordPrintCmd) { c.Gifted = "1" }, "1 gifted"},
		{"kept", func(c *app.RecordPrintCmd) { c.Kept = "2" }, "2 kept"},
		{"scrapped", func(c *app.RecordPrintCmd) { c.Scrapped = "1" }, "1 scrapped"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t)
			spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
			design := addedDesign(t, a, quotedDesign())
			cmd := printOf(design.ID, spool.ID)
			cmd.Quantity = "4"
			tc.set(&cmd)
			print := recordedPrint(t, a, cmd)

			err := a.DeletePrint(ctx(), print.ID)
			if !errors.Is(err, domain.ErrPrintAccountedFor) {
				t.Fatalf("DeletePrint error = %v, want ErrPrintAccountedFor", err)
			}
			if !strings.Contains(err.Error(), tc.inTheWay) {
				t.Errorf("DeletePrint error = %q, want it to name %q", err, tc.inTheWay)
			}
			if got := printByID(t, a, print.ID).ID; got != print.ID {
				t.Errorf("print %d is gone", print.ID)
			}
		})
	}
}

func TestPreviewEditPrintCostsAtTheFrozenRates(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())
	print := recordedPrint(t, a, printOf(design.ID, spool.ID))

	pricier := settingsUpdate()
	pricier.KwhPrice = "0.56"
	pricier.MachineHourlyRate = "0.70"
	if _, err := a.UpdateSettings(ctx(), pricier); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	preview, err := a.PreviewEditPrint(ctx(), editOfPrint(print))
	if err != nil {
		t.Fatalf("PreviewEditPrint: %v", err)
	}
	if preview.Cost.JobCost != 471 {
		t.Errorf("JobCost = %d, want 471", preview.Cost.JobCost)
	}
}

func TestEditPrintRoundTripsAFractionalUsageRow(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	cmd := printOf(design.ID, spool.ID)
	cmd.Usages[0].Grams = "85.59"
	recorded := recordedPrint(t, a, cmd)

	edited, err := a.EditPrint(ctx(), editOfPrint(recorded))
	if err != nil {
		t.Fatalf("EditPrint: %v", err)
	}
	if edited.UsedGrams != recorded.UsedGrams {
		t.Errorf("UsedGrams = %d, want %d", edited.UsedGrams, recorded.UsedGrams)
	}
}
