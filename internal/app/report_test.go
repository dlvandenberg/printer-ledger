package app_test

import (
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

// soldLastMonthsPrint prints the reference plate of two copies last month and
// sells one of them this month, so cost and revenue fall in different periods.
func soldLastMonthsPrint(t *testing.T, a *app.App) app.PrintView {
	t.Helper()
	spool := addedSpool(t, a, plaSpool())
	design := addedDesign(t, a, quotedDesign())

	cmd := printOf(design.ID, spool.ID)
	cmd.Date = dayOfMonth(-1, 10)
	print := recordedPrint(t, a, cmd)

	sale := saleOf(print.ID)
	sale.Date = dayOfMonth(0, 3)
	recordedSale(t, a, sale)
	return print
}

func TestReportCostsACopyIntoThePeriodItSoldIn(t *testing.T) {
	a := newApp(t)
	soldLastMonthsPrint(t, a)

	sold := reportOf(t, a, domain.ThisMonth)
	if sold.Revenue != 500 {
		t.Errorf("this month Revenue = %d, want 500", sold.Revenue)
	}
	if sold.CostOfSold != 235 {
		t.Errorf("this month CostOfSold = %d, want 235", sold.CostOfSold)
	}
	if sold.CopiesSold != 1 {
		t.Errorf("this month CopiesSold = %d, want 1", sold.CopiesSold)
	}
	if sold.ProductionCost != 0 {
		t.Errorf("this month ProductionCost = %d, want 0", sold.ProductionCost)
	}

	printed := reportOf(t, a, domain.LastMonth)
	if printed.Revenue != 0 {
		t.Errorf("last month Revenue = %d, want 0", printed.Revenue)
	}
	if printed.CostOfSold != 0 {
		t.Errorf("last month CostOfSold = %d, want 0", printed.CostOfSold)
	}
	if printed.ProductionCost != 471 {
		t.Errorf("last month ProductionCost = %d, want 471", printed.ProductionCost)
	}
}

func TestReportPeriodRunsToItsLastDay(t *testing.T) {
	a := newApp(t)
	report := reportOf(t, a, domain.ThisMonth)

	if !report.Bounded {
		t.Fatal("this month reports as unbounded, want bounded")
	}
	if want := dayOfMonth(0, 1); unit.FormatDate(report.From) != want {
		t.Errorf("From = %s, want %s", unit.FormatDate(report.From), want)
	}
	if want := lastDayOfMonth(0); unit.FormatDate(report.Through) != want {
		t.Errorf("Through = %s, want %s", unit.FormatDate(report.Through), want)
	}
}

func TestReportProfitAndMargin(t *testing.T) {
	a := newApp(t)
	soldLastMonthsPrint(t, a)

	report := reportOf(t, a, domain.ThisMonth)
	if report.Profit != 265 {
		t.Errorf("Profit = %d, want 265", report.Profit)
	}
	if !report.HasMargin {
		t.Fatal("HasMargin = false, want true")
	}
	if report.MarginPct != 5300 {
		t.Errorf("MarginPct = %d, want 5300", report.MarginPct)
	}
}

func TestReportWithoutSalesHasNoMargin(t *testing.T) {
	a := newApp(t)
	stockedPrint(t, a)

	report := reportOf(t, a, domain.AllTime)
	if report.HasMargin {
		t.Errorf("HasMargin = true, want false")
	}
	if report.MarginPct != 0 {
		t.Errorf("MarginPct = %d, want 0", report.MarginPct)
	}
}

func TestReportPeriodBoundariesIncludeFirstAndLastDay(t *testing.T) {
	tests := []struct {
		name      string
		saleDate  string
		thisMonth unit.Copies
		lastMonth unit.Copies
	}{
		{"first day of this month", dayOfMonth(0, 1), 1, 0},
		{"last day of last month", lastDayOfMonth(-1), 0, 1},
		{"first day of last month", dayOfMonth(-1, 1), 0, 1},
		{"last day of this month", lastDayOfMonth(0), 1, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t)
			spool := addedSpool(t, a, plaSpool())
			design := addedDesign(t, a, quotedDesign())

			cmd := printOf(design.ID, spool.ID)
			cmd.Date = dayOfMonth(-1, 1)
			print := recordedPrint(t, a, cmd)

			sale := saleOf(print.ID)
			sale.Date = tc.saleDate
			recordedSale(t, a, sale)

			if got := reportOf(t, a, domain.ThisMonth).CopiesSold; got != tc.thisMonth {
				t.Errorf("this month CopiesSold = %d, want %d", got, tc.thisMonth)
			}
			if got := reportOf(t, a, domain.LastMonth).CopiesSold; got != tc.lastMonth {
				t.Errorf("last month CopiesSold = %d, want %d", got, tc.lastMonth)
			}
		})
	}
}

func TestReportCountsGiftedKeptAndScrappedOnTheirOwnLine(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, plaSpool())
	design := addedDesign(t, a, quotedDesign())

	cmd := printOf(design.ID, spool.ID)
	cmd.Date = dayOfMonth(0, 4)
	cmd.Quantity = "4"
	cmd.Gifted, cmd.Kept, cmd.Scrapped = "1", "1", "1"
	recordedPrint(t, a, cmd)

	report := reportOf(t, a, domain.ThisMonth)
	if report.NotSold.GiftedCount != 1 || report.NotSold.KeptCount != 1 || report.NotSold.ScrappedCount != 1 {
		t.Errorf("NotSold counts = %d/%d/%d, want 1/1/1",
			report.NotSold.GiftedCount, report.NotSold.KeptCount, report.NotSold.ScrappedCount)
	}
	if report.NotSold.Copies != 3 {
		t.Errorf("NotSold.Copies = %d, want 3", report.NotSold.Copies)
	}
	if report.NotSold.Cost != 351 {
		t.Errorf("NotSold.Cost = %d, want 351", report.NotSold.Cost)
	}
	if report.CostOfSold != 0 {
		t.Errorf("CostOfSold = %d, want 0", report.CostOfSold)
	}
}

func TestReportRanksDesignsByProfit(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, plaSpool())
	planter := addedDesign(t, a, quotedDesign())
	clip := addedDesign(t, a, plaDesign())

	printed := printOf(planter.ID, spool.ID)
	printed.Date = dayOfMonth(0, 2)
	planterPrint := recordedPrint(t, a, printed)

	printed = printOf(clip.ID, spool.ID)
	printed.Date = dayOfMonth(0, 2)
	clipPrint := recordedPrint(t, a, printed)

	cheap := saleOf(planterPrint.ID)
	cheap.Price = "3.00"
	cheap.Date = dayOfMonth(0, 6)
	recordedSale(t, a, cheap)

	rich := saleOf(clipPrint.ID)
	rich.Price = "9.00"
	rich.Date = dayOfMonth(0, 6)
	recordedSale(t, a, rich)

	report := reportOf(t, a, domain.ThisMonth)
	if len(report.Designs) != 2 {
		t.Fatalf("ranked designs = %d, want 2", len(report.Designs))
	}
	if report.Designs[0].DesignName != clip.Name {
		t.Errorf("first ranked = %q, want %q", report.Designs[0].DesignName, clip.Name)
	}
	if got := designProfit(t, report, planter.Name); got.Profit != 65 {
		t.Errorf("%s Profit = %d, want 65", planter.Name, got.Profit)
	}
	if got := designProfit(t, report, clip.Name); got.Profit != 665 {
		t.Errorf("%s Profit = %d, want 665", clip.Name, got.Profit)
	}
}

func TestReportRanksDesignsOnTheSelectedPeriodOnly(t *testing.T) {
	a := newApp(t)
	soldLastMonthsPrint(t, a)

	if got := reportOf(t, a, domain.LastMonth).Designs; len(got) != 0 {
		t.Errorf("last month ranked designs = %d, want 0", len(got))
	}
	if got := reportOf(t, a, domain.ThisMonth).Designs; len(got) != 1 {
		t.Errorf("this month ranked designs = %d, want 1", len(got))
	}
}

func TestReportBreakEvenCountsSpoolSpendNotYetPrinted(t *testing.T) {
	a := newApp(t)
	soldLastMonthsPrint(t, a)
	addedSpool(t, a, spoolPriced(domain.PETG, "1000", "30.00"))

	settings := settingsUpdate()
	settings.PrinterPurchaseCost = "400.00"
	if _, err := a.UpdateSettings(ctx(), settings); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	report := reportOf(t, a, domain.ThisMonth)
	cash := report.BreakEven
	if cash.PrinterCost != 40000 {
		t.Errorf("PrinterCost = %d, want 40000", cash.PrinterCost)
	}
	if cash.SpoolSpend != 5200 {
		t.Errorf("SpoolSpend = %d, want 5200", cash.SpoolSpend)
	}
	if cash.EnergySpend != 14 {
		t.Errorf("EnergySpend = %d, want 14", cash.EnergySpend)
	}
	if cash.Revenue != 500 {
		t.Errorf("BreakEven Revenue = %d, want 500", cash.Revenue)
	}
	if cash.Spend != 45214 {
		t.Errorf("Spend = %d, want 45214", cash.Spend)
	}
	if cash.Balance != -44714 {
		t.Errorf("Balance = %d, want -44714", cash.Balance)
	}
	if cash.Reached {
		t.Error("Reached = true, want false")
	}
}

func TestReportBreakEvenIgnoresThePeriod(t *testing.T) {
	a := newApp(t)
	soldLastMonthsPrint(t, a)

	month := reportOf(t, a, domain.ThisMonth).BreakEven
	last := reportOf(t, a, domain.LastMonth).BreakEven
	if month != last {
		t.Errorf("break-even moved with the period: %+v and %+v", month, last)
	}
	if last.Revenue != 500 {
		t.Errorf("last month BreakEven Revenue = %d, want 500", last.Revenue)
	}
}

func TestReportInventoryValuesUnsoldCopiesAndSpools(t *testing.T) {
	a := newApp(t)
	soldLastMonthsPrint(t, a)

	inventory := reportOf(t, a, domain.ThisMonth).Inventory
	if inventory.UnsoldCopies != 1 {
		t.Errorf("UnsoldCopies = %d, want 1", inventory.UnsoldCopies)
	}
	if inventory.UnsoldValue != 235 {
		t.Errorf("UnsoldValue = %d, want 235", inventory.UnsoldValue)
	}
	if inventory.SpoolGrams != 880 {
		t.Errorf("SpoolGrams = %d, want 880", inventory.SpoolGrams)
	}
	if inventory.SpoolValue != 1936 {
		t.Errorf("SpoolValue = %d, want 1936", inventory.SpoolValue)
	}
	if inventory.Value != 2171 {
		t.Errorf("Inventory Value = %d, want 2171", inventory.Value)
	}
}

func TestReportAllTimeCountsEveryPeriod(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, plaSpool())
	design := addedDesign(t, a, quotedDesign())

	cmd := printOf(design.ID, spool.ID)
	cmd.Date = "2020-01-31"
	print := recordedPrint(t, a, cmd)

	sale := saleOf(print.ID)
	sale.Date = "2020-02-01"
	recordedSale(t, a, sale)

	report := reportOf(t, a, domain.AllTime)
	if report.Bounded {
		t.Error("all time reports as bounded, want unbounded")
	}
	if report.ProductionCost != 471 {
		t.Errorf("ProductionCost = %d, want 471", report.ProductionCost)
	}
	if report.CopiesSold != 1 {
		t.Errorf("CopiesSold = %d, want 1", report.CopiesSold)
	}
	if got := reportOf(t, a, domain.ThisMonth).CopiesSold; got != 0 {
		t.Errorf("this month CopiesSold = %d, want 0", got)
	}
}

func TestReportThisYearSpansTheWholeYear(t *testing.T) {
	a := newApp(t)
	spool := addedSpool(t, a, plaSpool())
	design := addedDesign(t, a, quotedDesign())

	printed := printOf(design.ID, spool.ID)
	printed.Date = dayOfMonth(0, 2)
	thisYear := recordedPrint(t, a, printed)

	printed = printOf(design.ID, spool.ID)
	printed.Date = "2020-12-31"
	lastYear := recordedPrint(t, a, printed)

	sale := saleOf(thisYear.ID)
	sale.Date = dayOfMonth(0, 5)
	recordedSale(t, a, sale)

	sale = saleOf(lastYear.ID)
	sale.Date = "2020-12-31"
	recordedSale(t, a, sale)

	report := reportOf(t, a, domain.ThisYear)
	if report.CopiesSold != 1 {
		t.Errorf("this year CopiesSold = %d, want 1", report.CopiesSold)
	}
	if report.ProductionCost != 471 {
		t.Errorf("this year ProductionCost = %d, want 471", report.ProductionCost)
	}
	if got := reportOf(t, a, domain.AllTime).CopiesSold; got != 2 {
		t.Errorf("all time CopiesSold = %d, want 2", got)
	}
}

func TestReportRejectsUnknownPeriod(t *testing.T) {
	a := newApp(t)

	_, err := a.Report(ctx(), app.ReportCmd{Period: "last week"})
	if got := fieldError(t, err, domain.FieldReportPeriod); got != domain.ErrMalformedPeriod.Error() {
		t.Errorf("period error = %q, want %q", got, domain.ErrMalformedPeriod.Error())
	}
}
