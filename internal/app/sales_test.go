package app_test

import (
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func TestRecordSaleThenList(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)

	recorded := recordedSale(t, a, saleOf(print.ID))

	sales, err := a.ListSales(ctx())
	if err != nil {
		t.Fatalf("ListSales: %v", err)
	}
	if len(sales) != 1 {
		t.Fatalf("len(sales) = %d, want 1", len(sales))
	}
	sale := sales[0]
	if sale.ID != recorded.ID {
		t.Errorf("ID = %d, want %d", sale.ID, recorded.ID)
	}
	if sale.Price != 500 {
		t.Errorf("Price = %d, want 500", sale.Price)
	}
	if sale.DesignName != "Planter" {
		t.Errorf("DesignName = %q, want %q", sale.DesignName, "Planter")
	}
	if got := unit.FormatDate(sale.Date); got != "2026-08-22" {
		t.Errorf("Date = %s, want 2026-08-22", got)
	}
	if sale.CostPerCopy != 235 {
		t.Errorf("CostPerCopy = %d, want 235", sale.CostPerCopy)
	}
}

func TestSellablePrintShowsCostAndSuggestedPrice(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)

	offered, ok := sellablePrint(t, a, print.ID)
	if !ok {
		t.Fatalf("print %d is not offered", print.ID)
	}
	if offered.CostPerCopy != 235 {
		t.Errorf("CostPerCopy = %d, want 235", offered.CostPerCopy)
	}
	if !offered.HasSuggestedPrice {
		t.Fatal("HasSuggestedPrice = false, want true")
	}
	if offered.SuggestedPrice != 400 {
		t.Errorf("SuggestedPrice = %d, want 400", offered.SuggestedPrice)
	}
	if offered.DesignName != "Planter" {
		t.Errorf("DesignName = %q, want %q", offered.DesignName, "Planter")
	}
	if offered.AvailableCopies != 2 {
		t.Errorf("AvailableCopies = %d, want 2", offered.AvailableCopies)
	}
}

func TestPreviewSaleShowsCostAndSuggestedPrice(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)

	preview, err := a.PreviewSale(ctx(), saleOf(print.ID))
	if err != nil {
		t.Fatalf("PreviewSale: %v", err)
	}
	if preview.CostPerCopy != 235 {
		t.Errorf("CostPerCopy = %d, want 235", preview.CostPerCopy)
	}
	if preview.SuggestedPrice != 400 {
		t.Errorf("SuggestedPrice = %d, want 400", preview.SuggestedPrice)
	}
	if preview.MinimumPrice != 271 {
		t.Errorf("MinimumPrice = %d, want 271", preview.MinimumPrice)
	}
	if preview.BelowFloor {
		t.Error("BelowFloor = true, want false at €5.00")
	}
}

func TestRecordSaleDecrementsAvailableCopies(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)

	recordedSale(t, a, saleOf(print.ID))

	sold := printByID(t, a, print.ID)
	if sold.SoldCount != 1 {
		t.Errorf("SoldCount = %d, want 1", sold.SoldCount)
	}
	if sold.AvailableCopies != 1 {
		t.Errorf("AvailableCopies = %d, want 1", sold.AvailableCopies)
	}
}

func TestSellablePrintsOmitsPrintWithNoCopiesAvailable(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)

	recordedSale(t, a, saleOf(print.ID))
	if _, ok := sellablePrint(t, a, print.ID); !ok {
		t.Fatal("print with one copy left is not offered")
	}

	recordedSale(t, a, saleOf(print.ID))
	if _, ok := sellablePrint(t, a, print.ID); ok {
		t.Error("print with no copies available is still offered")
	}
}

func TestRecordSaleOfLastCopyThenAgainFails(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)
	recordedSale(t, a, saleOf(print.ID))
	recordedSale(t, a, saleOf(print.ID))

	_, err := a.RecordSale(ctx(), saleOf(print.ID))
	if err == nil {
		t.Fatal("RecordSale of a third copy succeeded, want a validation error")
	}
	if msg := fieldError(t, err, domain.FieldSalePrint); msg != "has no copies available" {
		t.Errorf("printID error = %q, want %q", msg, "has no copies available")
	}
	if got := availableOf(t, a, print.ID); got != 0 {
		t.Errorf("AvailableCopies = %d, want 0", got)
	}
}

func TestRecordSaleBelowMinMarginWarnsAndRecords(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)

	cmd := saleOf(print.ID)
	cmd.Price = "2.50"

	preview, err := a.PreviewSale(ctx(), cmd)
	if err != nil {
		t.Fatalf("PreviewSale: %v", err)
	}
	if !preview.BelowFloor {
		t.Error("BelowFloor = false, want true at €2.50 on a €2.35 copy")
	}
	if preview.MinimumPrice != 271 {
		t.Errorf("MinimumPrice = %d, want 271", preview.MinimumPrice)
	}

	recordedSale(t, a, cmd)

	sales, err := a.ListSales(ctx())
	if err != nil {
		t.Fatalf("ListSales: %v", err)
	}
	if len(sales) != 1 {
		t.Fatalf("len(sales) = %d, want 1: a sale below the floor is still recorded", len(sales))
	}
	if sales[0].Price != 250 {
		t.Errorf("Price = %d, want 250", sales[0].Price)
	}
}

func TestPreviewSaleWarnsBelowMinMargin(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)

	tests := []struct {
		name  string
		price string
		want  bool
	}{
		{"below the floor", "2.50", true},
		{"a cent under the floor", "2.70", true},
		{"exactly the floor", "2.71", false},
		{"the suggested price", "4.00", false},
		{"nothing typed yet", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := saleOf(print.ID)
			cmd.Price = tc.price
			preview, err := a.PreviewSale(ctx(), cmd)
			if err != nil {
				t.Fatalf("PreviewSale: %v", err)
			}
			if preview.BelowFloor != tc.want {
				t.Errorf("BelowFloor = %v, want %v", preview.BelowFloor, tc.want)
			}
		})
	}
}

func TestRecordSaleBackdated(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)

	cmd := saleOf(print.ID)
	cmd.Date = "2026-07-05"
	recordedSale(t, a, cmd)

	sales, err := a.ListSales(ctx())
	if err != nil {
		t.Fatalf("ListSales: %v", err)
	}
	if got := unit.FormatDate(sales[0].Date); got != "2026-07-05" {
		t.Errorf("Date = %s, want 2026-07-05", got)
	}
}

func TestRecordSaleRejectsMalformedInput(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)

	tests := []struct {
		name   string
		field  string
		mutate func(*app.RecordSaleCmd)
		want   string
	}{
		{"price", domain.FieldSalePrice, func(c *app.RecordSaleCmd) { c.Price = "a fiver" }, unit.ErrMalformedMoney.Error()},
		{"date", domain.FieldSaleDate, func(c *app.RecordSaleCmd) { c.Date = "yesterday" }, unit.ErrMalformedDate.Error()},
		{"negative price", domain.FieldSalePrice, func(c *app.RecordSaleCmd) { c.Price = "-5.00" }, "cannot be negative"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := saleOf(print.ID)
			tc.mutate(&cmd)
			_, err := a.RecordSale(ctx(), cmd)
			if err == nil {
				t.Fatal("RecordSale succeeded, want a validation error")
			}
			if msg := fieldError(t, err, tc.field); msg != tc.want {
				t.Errorf("%s error = %q, want %q", tc.field, msg, tc.want)
			}
		})
	}
}

func TestEditSaleCorrectsThePrice(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)
	sale := recordedSale(t, a, saleOf(print.ID))

	cmd := editOfSale(sale)
	cmd.Price = "6.50"
	if _, err := a.EditSale(ctx(), cmd); err != nil {
		t.Fatalf("EditSale: %v", err)
	}

	sales, err := a.ListSales(ctx())
	if err != nil {
		t.Fatalf("ListSales: %v", err)
	}
	if len(sales) != 1 {
		t.Fatalf("len(sales) = %d, want 1", len(sales))
	}
	if sales[0].Price != 650 {
		t.Errorf("Price = %d, want 650", sales[0].Price)
	}
	if got := availableOf(t, a, print.ID); got != 1 {
		t.Errorf("AvailableCopies = %d, want 1: an edit moves no copies", got)
	}
}

func TestEditSaleOfTheLastCopyKeepsItsOwnPrint(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)
	recordedSale(t, a, saleOf(print.ID))
	last := recordedSale(t, a, saleOf(print.ID))

	offered, err := a.SellablePrintsForSale(ctx(), last.ID)
	if err != nil {
		t.Fatalf("SellablePrints: %v", err)
	}
	if len(offered) != 1 || offered[0].PrintID != print.ID {
		t.Fatalf("SellablePrints while editing = %v, want the print the sale came from", offered)
	}

	cmd := editOfSale(last)
	cmd.Price = "7.00"
	edited, err := a.EditSale(ctx(), cmd)
	if err != nil {
		t.Fatalf("EditSale of the last copy sold: %v", err)
	}
	if edited.Price != 700 {
		t.Errorf("Price = %d, want 700", edited.Price)
	}
}

func TestDeleteSaleReturnsItsCopyToAvailable(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)
	sale := recordedSale(t, a, saleOf(print.ID))

	if err := a.DeleteSale(ctx(), sale.ID); err != nil {
		t.Fatalf("DeleteSale: %v", err)
	}

	sales, err := a.ListSales(ctx())
	if err != nil {
		t.Fatalf("ListSales: %v", err)
	}
	if len(sales) != 0 {
		t.Errorf("len(sales) = %d, want 0", len(sales))
	}
	returned := printByID(t, a, print.ID)
	if returned.SoldCount != 0 {
		t.Errorf("SoldCount = %d, want 0", returned.SoldCount)
	}
	if returned.AvailableCopies != 2 {
		t.Errorf("AvailableCopies = %d, want 2", returned.AvailableCopies)
	}
}

func TestEditSaleCorrectsTheDate(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)
	sale := recordedSale(t, a, saleOf(print.ID))

	cmd := editOfSale(sale)
	cmd.Date = "2026-07-05"
	if _, err := a.EditSale(ctx(), cmd); err != nil {
		t.Fatalf("EditSale: %v", err)
	}

	sales, err := a.ListSales(ctx())
	if err != nil {
		t.Fatalf("ListSales: %v", err)
	}
	if got := unit.FormatDate(sales[0].Date); got != "2026-07-05" {
		t.Errorf("Date = %s, want 2026-07-05", got)
	}
}

func TestEditSaleMovesTheCopyToAnotherPrint(t *testing.T) {
	a := newApp(t)
	first := stockedPrint(t, a)
	second := recordedPrint(t, a, printOf(first.DesignID, first.Usages[0].SpoolID))
	sale := recordedSale(t, a, saleOf(first.ID))

	cmd := editOfSale(sale)
	cmd.PrintID = second.ID
	if _, err := a.EditSale(ctx(), cmd); err != nil {
		t.Fatalf("EditSale onto another print: %v", err)
	}

	if got := availableOf(t, a, first.ID); got != 2 {
		t.Errorf("AvailableCopies of the print sold from = %d, want 2", got)
	}
	if got := availableOf(t, a, second.ID); got != 1 {
		t.Errorf("AvailableCopies of the print moved to = %d, want 1", got)
	}
}

func TestRecordSaleWithoutAPrintFails(t *testing.T) {
	a := newApp(t)
	stockedPrint(t, a)

	cmd := saleOf(0)
	_, err := a.RecordSale(ctx(), cmd)
	if err == nil {
		t.Fatal("RecordSale naming no print succeeded, want a validation error")
	}
	if msg := fieldError(t, err, domain.FieldSalePrint); msg != "is required" {
		t.Errorf("printID error = %q, want %q", msg, "is required")
	}
}

func TestSellablePrintNamesItsMaterial(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)

	offered, ok := sellablePrint(t, a, print.ID)
	if !ok {
		t.Fatalf("print %d is not offered", print.ID)
	}
	if offered.Material.FilamentType != domain.PLA {
		t.Errorf("FilamentType = %s, want %s", offered.Material.FilamentType, domain.PLA)
	}
	if len(offered.Material.Colors) != 1 || offered.Material.Colors[0] != "Black" {
		t.Errorf("Colors = %v, want [Black]", offered.Material.Colors)
	}
}

func TestSellablePrintOfASwapNamesBothColorsMostFilamentFirst(t *testing.T) {
	a := newApp(t)
	ranOut := addedSpool(t, a, spoolColored("Red"))
	replacement := addedSpool(t, a, spoolColored("Blue"))
	design := addedDesign(t, a, quotedDesign())
	print := recordedPrint(t, a, printOfRows(design.ID, usage(ranOut.ID, "40"), usage(replacement.ID, "80")))

	offered, ok := sellablePrint(t, a, print.ID)
	if !ok {
		t.Fatalf("print %d is not offered", print.ID)
	}
	if len(offered.Material.Colors) != 2 || offered.Material.Colors[0] != "Blue" || offered.Material.Colors[1] != "Red" {
		t.Errorf("Colors = %v, want [Blue Red]", offered.Material.Colors)
	}
}

// Equal grams leave nothing to order the colors by, so the older Spool goes
// first: the label must read the same on every reload.
func TestSellablePrintOrdersEqualGramsByOldestSpool(t *testing.T) {
	a := newApp(t)
	first := addedSpool(t, a, spoolColored("Red"))
	second := addedSpool(t, a, spoolColored("Blue"))
	design := addedDesign(t, a, quotedDesign())
	print := recordedPrint(t, a, printOfRows(design.ID, usage(second.ID, "60"), usage(first.ID, "60")))

	offered, ok := sellablePrint(t, a, print.ID)
	if !ok {
		t.Fatalf("print %d is not offered", print.ID)
	}
	if len(offered.Material.Colors) != 2 || offered.Material.Colors[0] != "Red" || offered.Material.Colors[1] != "Blue" {
		t.Errorf("Colors = %v, want [Red Blue]", offered.Material.Colors)
	}
}

func TestSaleNamesTheMaterialItWasPrintedIn(t *testing.T) {
	a := newApp(t)
	print := stockedPrint(t, a)
	recordSale := recordedSale(t, a, saleOf(print.ID))

	sales, err := a.ListSales(ctx())
	if err != nil {
		t.Fatalf("ListSales: %v", err)
	}
	if len(sales) != 1 || sales[0].ID != recordSale.ID {
		t.Fatalf("sales = %v, want the recorded sale", sales)
	}
	if sales[0].Material.FilamentType != domain.PLA {
		t.Errorf("FilamentType = %s, want %s", sales[0].Material.FilamentType, domain.PLA)
	}
	if len(sales[0].Material.Colors) != 1 || sales[0].Material.Colors[0] != "Black" {
		t.Errorf("Colors = %v, want [Black]", sales[0].Material.Colors)
	}
}
