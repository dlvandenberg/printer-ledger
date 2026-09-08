package app

import (
	"context"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

type ReportCmd struct {
	Period string
}

type ReportView struct {
	Period domain.Period
	From   time.Time
	// Through is the last date in the period, not the first date after it, so
	// the TUI prints the range rather than deriving it.
	Through time.Time
	// Bounded is false for all time, which has neither end.
	Bounded bool

	Revenue    unit.Cents
	CostOfSold unit.Cents
	Profit     unit.Cents
	MarginPct  unit.Percent
	HasMargin  bool
	CopiesSold unit.Copies

	ProductionCost unit.Cents
	NotSold        NotSoldView

	Designs   []DesignProfitView
	BreakEven BreakEvenView
	Inventory InventoryView
}

type NotSoldView struct {
	GiftedCount   unit.Copies
	KeptCount     unit.Copies
	ScrappedCount unit.Copies
	Copies        unit.Copies
	Cost          unit.Cents
}

type DesignProfitView struct {
	DesignID   int64
	DesignName string
	Revenue    unit.Cents
	Cost       unit.Cents
	Profit     unit.Cents
	MarginPct  unit.Percent
	HasMargin  bool
	CopiesSold unit.Copies
}

type BreakEvenView struct {
	Revenue     unit.Cents
	PrinterCost unit.Cents
	SpoolSpend  unit.Cents
	EnergySpend unit.Cents
	Spend       unit.Cents
	Balance     unit.Cents
	Reached     bool
}

type InventoryView struct {
	UnsoldCopies unit.Copies
	UnsoldValue  unit.Cents
	SpoolGrams   unit.Grams
	SpoolValue   unit.Cents
	Value        unit.Cents
}

// Report answers two questions at once and labels them apart: profit for the
// period matches cost to what sold, while break-even counts every euro spent
// (ADR-0006).
func (a *App) Report(ctx context.Context, cmd ReportCmd) (ReportView, error) {
	period, err := parseReportPeriod(cmd)
	if err != nil {
		return ReportView{}, err
	}

	settings, err := a.db.Settings(ctx)
	if err != nil {
		return ReportView{}, err
	}
	prints, err := a.db.PrintLedgers(ctx)
	if err != nil {
		return ReportView{}, err
	}
	sales, err := a.db.Sales(ctx)
	if err != nil {
		return ReportView{}, err
	}
	spools, err := a.db.SpoolLedgers(ctx)
	if err != nil {
		return ReportView{}, err
	}
	designs, err := a.db.Designs(ctx)
	if err != nil {
		return ReportView{}, err
	}

	return toReportView(domain.NewReport(domain.ReportInput{
		Period:   period,
		Today:    unit.Today(),
		Prints:   prints,
		Sales:    sales,
		Spools:   spools,
		Designs:  designs,
		Settings: settings,
	})), nil
}

func (a *App) ReportPeriods() []domain.Period { return domain.Periods() }

func parseReportPeriod(cmd ReportCmd) (domain.Period, error) {
	errs := &domain.ValidationError{}
	period := parseField(errs, domain.FieldReportPeriod, cmd.Period, domain.ParsePeriod)
	if err := errs.OrNil(); err != nil {
		return "", err
	}
	return period, nil
}

func toReportView(r domain.Report) ReportView {
	margin, hasMargin := r.Sold.MarginPct()
	return ReportView{
		Period:         r.Period,
		From:           r.Range.From,
		Through:        r.Range.Through(),
		Bounded:        r.Range.Bounded(),
		Revenue:        r.Sold.Revenue,
		CostOfSold:     r.Sold.Cost,
		Profit:         r.Sold.Profit(),
		MarginPct:      margin,
		HasMargin:      hasMargin,
		CopiesSold:     r.Sold.CopiesSold,
		ProductionCost: r.ProductionCost,
		NotSold: NotSoldView{
			GiftedCount:   r.NotSold.Gifted,
			KeptCount:     r.NotSold.Kept,
			ScrappedCount: r.NotSold.Scrapped,
			Copies:        r.NotSold.Copies(),
			Cost:          r.NotSold.Cost,
		},
		Designs: toDesignProfitViews(r.Designs),
		BreakEven: BreakEvenView{
			Revenue:     r.BreakEven.Revenue,
			PrinterCost: r.BreakEven.PrinterCost,
			SpoolSpend:  r.BreakEven.SpoolSpend,
			EnergySpend: r.BreakEven.EnergySpend,
			Spend:       r.BreakEven.Spend(),
			Balance:     r.BreakEven.Balance(),
			Reached:     r.BreakEven.Reached(),
		},
		Inventory: InventoryView{
			UnsoldCopies: r.Inventory.UnsoldCopies,
			UnsoldValue:  r.Inventory.UnsoldValue,
			SpoolGrams:   r.Inventory.SpoolGrams,
			SpoolValue:   r.Inventory.SpoolValue,
			Value:        r.Inventory.Value(),
		},
	}
}

func toDesignProfitViews(ranked []domain.DesignProfit) []DesignProfitView {
	views := make([]DesignProfitView, 0, len(ranked))
	for _, d := range ranked {
		margin, hasMargin := d.Sold.MarginPct()
		views = append(views, DesignProfitView{
			DesignID:   d.DesignID,
			DesignName: d.DesignName,
			Revenue:    d.Sold.Revenue,
			Cost:       d.Sold.Cost,
			Profit:     d.Sold.Profit(),
			MarginPct:  margin,
			HasMargin:  hasMargin,
			CopiesSold: d.Sold.CopiesSold,
		})
	}
	return views
}
