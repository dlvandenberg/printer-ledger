package domain

import (
	"cmp"
	"slices"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const FieldReportPeriod = "period"

// ReportInput is everything the report is derived from. A Report is a read
// model, so it takes the ledgers whole rather than querying for the slice of
// each it needs.
type ReportInput struct {
	Period       Period
	Today        time.Time
	PrintLedgers []PrintLedger
	Sales        []Sale
	SpoolLedgers SpoolLedgers
	Designs      []Design
	Settings     Settings
}

type Report struct {
	Period Period
	Range  DateRange
	// Sold matches cost to revenue: the cost is the cost of the copies sold in
	// the period, whenever they were printed (ADR-0006).
	Sold SoldSummary
	// ProductionCost is what the Prints made in the period cost, whether or not
	// they sold. A separate measure, never folded into Sold.
	ProductionCost unit.Cents
	NotSold        NotSold
	Designs        []DesignProfit
	// BreakEven and Inventory are all-time and ignore the period.
	BreakEven BreakEven
	Inventory Inventory
}

type SoldSummary struct {
	Revenue    unit.Cents
	Cost       unit.Cents
	CopiesSold unit.Copies
}

func (s SoldSummary) Profit() unit.Cents { return s.Revenue - s.Cost }

// MarginPct is false where nothing sold: no revenue is not a margin of zero.
func (s SoldSummary) MarginPct() (unit.Percent, bool) {
	if s.Revenue == 0 {
		return 0, false
	}
	return unit.Percent(int64(s.Profit()) * unit.PercentScale / int64(s.Revenue)), true
}

// NotSold is the gifted, kept and scrapped copies of the Prints made in the
// period, with the share of job cost they carry. It is what makes the gap
// between production cost and revenue explainable (ADR-0023).
type NotSold struct {
	Gifted   unit.Copies
	Kept     unit.Copies
	Scrapped unit.Copies
	Cost     unit.Cents
}

func (n NotSold) Copies() unit.Copies { return n.Gifted + n.Kept + n.Scrapped }

func (n NotSold) including(p Print) NotSold {
	n.Gifted += p.GiftedCount
	n.Kept += p.KeptCount
	n.Scrapped += p.ScrappedCount
	n.Cost += unit.Cents(p.NotSoldCopies()) * p.CostPerCopy()
	return n
}

type DesignProfit struct {
	DesignID   int64
	DesignName string
	Sold       SoldSummary
}

// BreakEven is cash-based and all-time: it counts filament bought but not yet
// printed, because that money has really left the bank (ADR-0006).
type BreakEven struct {
	Revenue     unit.Cents
	PrinterCost unit.Cents
	SpoolSpend  unit.Cents
	EnergySpend unit.Cents
}

func (b BreakEven) Spend() unit.Cents { return b.PrinterCost + b.SpoolSpend + b.EnergySpend }

func (b BreakEven) Balance() unit.Cents { return b.Revenue - b.Spend() }

func (b BreakEven) Reached() bool { return b.Balance() >= 0 }

type Inventory struct {
	UnsoldCopies unit.Copies
	UnsoldValue  unit.Cents
	SpoolGrams   unit.Grams
	SpoolValue   unit.Cents
}

func (i Inventory) Value() unit.Cents { return i.UnsoldValue + i.SpoolValue }

func NewReport(in ReportInput) Report {
	report := Report{Period: in.Period, Range: in.Period.Range(in.Today)}
	report.Sold, report.Designs = soldInRange(in, report.Range)

	for _, ledger := range in.PrintLedgers {
		p := ledger.Print
		report.BreakEven.EnergySpend += p.Cost().Energy
		report.Inventory.UnsoldCopies += ledger.Available()
		report.Inventory.UnsoldValue += unit.Cents(ledger.Available()) * p.CostPerCopy()
		if !report.Range.Contains(p.Date) {
			continue
		}
		// Production and the copies that will never repay it are scoped by the
		// Print's date: a gifted copy has no sale to be matched to (ADR-0023).
		report.ProductionCost += p.Cost().JobCost()
		report.NotSold = report.NotSold.including(p)
	}

	for _, ledger := range in.SpoolLedgers {
		report.BreakEven.SpoolSpend += ledger.Spool.PurchaseCost
		report.Inventory.SpoolGrams += ledger.Remaining()
		report.Inventory.SpoolValue += ledger.RemainingValue()
	}

	report.BreakEven.PrinterCost = in.Settings.PrinterPurchaseCost
	for _, sale := range in.Sales {
		report.BreakEven.Revenue += sale.Price
	}
	return report
}

// soldInRange costs a Sale from the Print it names, so a copy printed in one
// period and sold in another is costed into the period it sold in.
func soldInRange(in ReportInput, r DateRange) (SoldSummary, []DesignProfit) {
	byPrint := make(map[int64]PrintLedger, len(in.PrintLedgers))
	for _, ledger := range in.PrintLedgers {
		byPrint[ledger.Print.ID] = ledger
	}

	var sold SoldSummary
	perDesign := map[int64]SoldSummary{}
	for _, sale := range in.Sales {
		ledger, ok := byPrint[sale.PrintID]
		if !ok || !r.Contains(sale.Date) {
			continue
		}
		cost := ledger.Print.CostPerCopy()
		sold.Revenue += sale.Price
		sold.Cost += cost
		sold.CopiesSold++

		design := perDesign[ledger.Print.DesignID]
		design.Revenue += sale.Price
		design.Cost += cost
		design.CopiesSold++
		perDesign[ledger.Print.DesignID] = design
	}
	return sold, rankedDesigns(perDesign, in.Designs)
}

func rankedDesigns(perDesign map[int64]SoldSummary, designs []Design) []DesignProfit {
	named := make(map[int64]string, len(designs))
	for _, design := range designs {
		named[design.ID] = design.Name
	}

	ranked := make([]DesignProfit, 0, len(perDesign))
	for id, sold := range perDesign {
		ranked = append(ranked, DesignProfit{DesignID: id, DesignName: named[id], Sold: sold})
	}
	slices.SortFunc(ranked, func(a, b DesignProfit) int {
		if c := cmp.Compare(b.Sold.Profit(), a.Sold.Profit()); c != 0 {
			return c
		}
		return cmp.Compare(a.DesignName, b.DesignName)
	})
	return ranked
}
