package app_test

import (
	"errors"
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func TestRecordPrintWithStockCountsThenList(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	cmd := printOf(design.ID, spool.ID)
	cmd.Quantity = "4"
	cmd.Gifted = "1"
	cmd.Scrapped = "1"

	recorded, err := a.RecordPrint(ctx(), cmd)
	if err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}

	got := printByID(t, a, recorded.ID)
	if got.GiftedCount != 1 {
		t.Errorf("GiftedCount = %d, want 1", got.GiftedCount)
	}
	if got.KeptCount != 0 {
		t.Errorf("KeptCount = %d, want 0", got.KeptCount)
	}
	if got.ScrappedCount != 1 {
		t.Errorf("ScrappedCount = %d, want 1", got.ScrappedCount)
	}
	if got.SoldCount != 0 {
		t.Errorf("SoldCount = %d, want 0", got.SoldCount)
	}
	if got.AvailableCopies != 2 {
		t.Errorf("AvailableCopies = %d, want 2", got.AvailableCopies)
	}
}

func TestRecordPrintRejectsMoreCopiesAccountedForThanProduced(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	cmd := printOf(design.ID, spool.ID)
	cmd.Quantity = "2"
	cmd.Kept = "1"
	cmd.Scrapped = "2"

	_, err := a.RecordPrint(ctx(), cmd)
	if err == nil {
		t.Fatalf("RecordPrint accepted 3 accounted-for copies of a print of 2")
	}
	if got := fieldError(t, err, domain.FieldQuantity); got == "" {
		t.Errorf("no error on %s, got %v", domain.FieldQuantity, err)
	}
	if got := fieldError(t, err, domain.FieldScrapped); got == "" {
		t.Errorf("no error on %s, got %v", domain.FieldScrapped, err)
	}
}

func TestRecordPrintRejectsNegativeStockCount(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	cmd := printOf(design.ID, spool.ID)
	cmd.Scrapped = "-1"

	_, err := a.RecordPrint(ctx(), cmd)
	if err == nil {
		t.Fatalf("RecordPrint accepted a negative scrapped count")
	}
	if got := fieldError(t, err, domain.FieldScrapped); got == "" {
		t.Errorf("no error on %s, got %v", domain.FieldScrapped, err)
	}
}

// Per-copy cost divides across all copies including scrapped ones: every copy
// cost the same to make (ADR-0002).
func TestRecordPrintCostPerCopyIncludesScrappedCopies(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	cmd := printOf(design.ID, spool.ID)
	cmd.Scrapped = "1"

	view, err := a.RecordPrint(ctx(), cmd)
	if err != nil {
		t.Fatalf("RecordPrint: %v", err)
	}
	if view.Cost.CostPerCopy != 235 {
		t.Errorf("CostPerCopy = %d, want 235", view.Cost.CostPerCopy)
	}
}

func TestDeletePrintBlockedWhileACopyIsSold(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)
	sale := recordedSale(t, a, saleOf(print.ID))

	err := a.DeletePrint(ctx(), print.ID)
	if !errors.Is(err, domain.ErrPrintAccountedFor) {
		t.Fatalf("DeletePrint = %v, want ErrPrintAccountedFor", err)
	}

	if err := a.DeleteSale(ctx(), sale.ID); err != nil {
		t.Fatalf("DeleteSale: %v", err)
	}
	if err := a.DeletePrint(ctx(), print.ID); err != nil {
		t.Errorf("DeletePrint after the sale was deleted: %v", err)
	}
}

func TestEditPrintQuantityBelowSoldCountFails(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)
	recordedSale(t, a, saleOf(print.ID))
	recordedSale(t, a, saleOf(print.ID))

	cmd := editOfPrint(printByID(t, a, print.ID))
	cmd.Quantity = "1"
	_, err := a.EditPrint(ctx(), cmd)
	if err == nil {
		t.Fatal("EditPrint below the copies sold succeeded, want a validation error")
	}
	if msg := fieldError(t, err, domain.FieldQuantity); msg == "" {
		t.Error("no error on quantity, want one naming the copies accounted for")
	}
}

func TestRecordPrintRejectsNegativeStockCounts(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	tests := []struct {
		name  string
		count func(cmd *app.RecordPrintCmd)
		field string
	}{
		{name: "gifted", count: func(cmd *app.RecordPrintCmd) { cmd.Gifted = "-1" }, field: domain.FieldGifted},
		{name: "kept", count: func(cmd *app.RecordPrintCmd) { cmd.Kept = "-2" }, field: domain.FieldKept},
		{name: "scrapped", count: func(cmd *app.RecordPrintCmd) { cmd.Scrapped = "-3" }, field: domain.FieldScrapped},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := printOf(design.ID, spool.ID)
			cmd.Quantity = "4"
			tc.count(&cmd)

			_, err := a.RecordPrint(ctx(), cmd)
			want := unit.ErrMalformedCopies.Error()
			if got := fieldError(t, err, tc.field); got != want {
				t.Errorf("%s error = %q, want %q", tc.field, got, want)
			}
		})
	}
}

func TestRecordPrintKeepsCopiesAboveThe32BitRange(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, spoolPriced(domain.PLA, "1000", "22.00"))
	design := addedDesign(t, a, quotedDesign())

	const quantity = "3000000000"
	cmd := printOf(design.ID, spool.ID)
	cmd.Quantity = quantity
	cmd.Gifted = "2200000000"

	recorded := recordedPrint(t, a, cmd)

	got := printByID(t, a, recorded.ID)
	if formatted := unit.FormatCopies(got.Quantity); formatted != quantity {
		t.Errorf("Quantity = %s, want %s", formatted, quantity)
	}
	if got.GiftedCount != 2_200_000_000 {
		t.Errorf("GiftedCount = %d, want 2200000000", got.GiftedCount)
	}
	if got.AvailableCopies != 800_000_000 {
		t.Errorf("AvailableCopies = %d, want 800000000", got.AvailableCopies)
	}
}
