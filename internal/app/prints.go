package app

import (
	"context"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

type RecordPrintCmd struct {
	DesignID int64
	Date     string
	Quantity string
	Minutes  string
	Gifted   string
	Kept     string
	Scrapped string
	Usages   []FilamentUsageCmd
}

// EditPrintCmd is a RecordPrintCmd against a Print that already exists: an
// edit re-types every field the record form has.
type EditPrintCmd struct {
	RecordPrintCmd

	PrintID int64
}

type FilamentUsageCmd struct {
	SpoolID int64
	Grams   string
}

type PrintView struct {
	ID              int64
	DesignID        int64
	DesignName      string
	Date            time.Time
	Quantity        unit.Copies
	Minutes         unit.Minutes
	UsedGrams       unit.Grams
	SoldCount       unit.Copies
	GiftedCount     unit.Copies
	KeptCount       unit.Copies
	ScrappedCount   unit.Copies
	AvailableCopies unit.Copies
	KwhPrice        unit.Cents
	KwhPerHour      unit.KwhPerHour
	MachineRate     unit.Cents
	Cost            PrintCostView
	Usages          []FilamentUsageView
}

type PrintCostView struct {
	Filament    unit.Cents
	Energy      unit.Cents
	Overhead    unit.Cents
	JobCost     unit.Cents
	CostPerCopy unit.Cents
}

type FilamentUsageView struct {
	SpoolID      int64
	SpoolBrand   string
	SpoolColor   string
	FilamentType domain.FilamentType
	Grams        unit.Grams
	CostPerGram  unit.CentsPerGram
}

// PrintPreviewView costs a Print still being typed. RateSeeded says the form
// should warn: the energy figure came from a seeded rate rather than a
// smart-plug reading. It never blocks recording.
type PrintPreviewView struct {
	Cost         PrintCostView
	FilamentType domain.FilamentType
	RateSeeded   bool
}

// PrintDraftView is what the Print form opens on: the Design's per-copy
// estimates multiplied by the quantity, so a normal plate needs no typing.
type PrintDraftView struct {
	DesignID   int64
	DesignName string
	Quantity   unit.Copies
	Grams      unit.Grams
	Minutes    unit.Minutes
}

func (a *App) RecordPrint(ctx context.Context, cmd RecordPrintCmd) (PrintView, error) {
	var view PrintView
	err := a.db.InTx(ctx, func(tx domain.Database) error {
		if _, err := designOf(ctx, tx, cmd.DesignID); err != nil {
			return err
		}
		settings, err := tx.Settings(ctx)
		if err != nil {
			return err
		}
		ledgers, err := tx.SpoolLedgers(ctx)
		if err != nil {
			return err
		}

		recorded, err := parseRecordPrint(cmd, settings, ledgers)
		if err != nil {
			return err
		}

		created, err := tx.CreatePrint(ctx, recorded)
		if err != nil {
			return err
		}
		view, err = printViewOf(ctx, tx, created.ID)
		return err
	})
	if err != nil {
		return PrintView{}, err
	}
	return view, nil
}

// EditPrint corrects a Print in place (ADR-0011). Its cost does not move: the
// rate snapshot and the frozen gram prices are kept, while the Spool's
// remaining follows the edited grams, being derived (ADR-0003).
func (a *App) EditPrint(ctx context.Context, cmd EditPrintCmd) (PrintView, error) {
	var view PrintView
	err := a.db.InTx(ctx, func(tx domain.Database) error {
		stored, err := tx.PrintLedger(ctx, cmd.PrintID)
		if err != nil {
			return err
		}
		if _, err := designOf(ctx, tx, cmd.DesignID); err != nil {
			return err
		}
		ledgers, err := tx.SpoolLedgers(ctx)
		if err != nil {
			return err
		}

		edited, err := parseEditPrint(cmd, stored, ledgers)
		if err != nil {
			return err
		}

		if _, err := tx.UpdatePrint(ctx, edited); err != nil {
			return err
		}
		view, err = printViewOf(ctx, tx, cmd.PrintID)
		return err
	})
	if err != nil {
		return PrintView{}, err
	}
	return view, nil
}

func (a *App) DeletePrint(ctx context.Context, id int64) error {
	return a.db.InTx(ctx, func(tx domain.Database) error {
		ledger, err := tx.PrintLedger(ctx, id)
		if err != nil {
			return err
		}
		if err := ledger.DeleteBlocked(); err != nil {
			return err
		}
		return tx.DeletePrint(ctx, id)
	})
}

func (a *App) ListPrints(ctx context.Context) ([]PrintView, error) {
	ledgers, err := a.db.PrintLedgers(ctx)
	if err != nil {
		return nil, err
	}
	designs, err := designsByID(ctx, a.db)
	if err != nil {
		return nil, err
	}
	spools, err := a.db.SpoolLedgers(ctx)
	if err != nil {
		return nil, err
	}

	views := make([]PrintView, 0, len(ledgers))
	for _, ledger := range ledgers {
		views = append(views, toPrintView(ledger, designs[ledger.Print.DesignID], spools))
	}
	return views, nil
}

func (a *App) PrintDetail(ctx context.Context, id int64) (PrintView, error) {
	return printViewOf(ctx, a.db, id)
}

func (a *App) PrintDraft(ctx context.Context, designID int64, quantity string) (PrintDraftView, error) {
	design, err := designOf(ctx, a.db, designID)
	if err != nil {
		return PrintDraftView{}, err
	}
	copies, err := parseQuantity(quantity)
	if err != nil {
		return PrintDraftView{}, err
	}
	return PrintDraftView{
		DesignID:   design.ID,
		DesignName: design.Name,
		Quantity:   copies,
		Grams:      design.EstimatedGrams * unit.Grams(copies),
		Minutes:    design.EstimatedMinutes * unit.Minutes(copies),
	}, nil
}

// PreviewPrint costs a Print that has not been recorded, so the breakdown moves
// with the form. A field the operator has not finished typing costs nothing
// rather than failing, which is what keeps the breakdown on screen.
func (a *App) PreviewPrint(ctx context.Context, cmd RecordPrintCmd) (PrintPreviewView, error) {
	settings, err := a.db.Settings(ctx)
	if err != nil {
		return PrintPreviewView{}, err
	}
	ledgers, err := a.db.SpoolLedgers(ctx)
	if err != nil {
		return PrintPreviewView{}, err
	}

	drafted := draftPrint(cmd, &domain.ValidationError{}).WithRates(settings, ledgers)
	filamentType := drafted.FilamentType(ledgers)
	return PrintPreviewView{
		Cost:         toPrintCostView(drafted),
		FilamentType: filamentType,
		RateSeeded:   filamentType != "" && !settings.PowerRate(filamentType).Measured,
	}, nil
}

// PreviewEditPrint costs a Print being edited at the rates it was recorded
// with, so the breakdown on the edit form is the cost that will be stored.
func (a *App) PreviewEditPrint(ctx context.Context, cmd EditPrintCmd) (PrintPreviewView, error) {
	stored, err := a.db.PrintLedger(ctx, cmd.PrintID)
	if err != nil {
		return PrintPreviewView{}, err
	}
	ledgers, err := a.db.SpoolLedgers(ctx)
	if err != nil {
		return PrintPreviewView{}, err
	}

	drafted := draftPrint(cmd.RecordPrintCmd, &domain.ValidationError{}).WithFrozenRatesOf(stored.Print, ledgers)
	return PrintPreviewView{
		Cost:         toPrintCostView(drafted),
		FilamentType: drafted.FilamentType(ledgers),
	}, nil
}

func parseRecordPrint(cmd RecordPrintCmd, s domain.Settings, ledgers []domain.SpoolLedger) (domain.Print, error) {
	errs := &domain.ValidationError{}
	drafted := draftPrint(cmd, errs)

	return validate(drafted, errs, func(p domain.Print) (domain.Print, error) {
		return domain.NewPrint(p, s, ledgers)
	})
}

func parseEditPrint(cmd EditPrintCmd, stored domain.PrintLedger, ledgers []domain.SpoolLedger) (domain.Print, error) {
	errs := &domain.ValidationError{}
	drafted := draftPrint(cmd.RecordPrintCmd, errs)

	return validate(drafted, errs, func(p domain.Print) (domain.Print, error) {
		return domain.EditedPrint(stored, p, ledgers)
	})
}

func draftPrint(cmd RecordPrintCmd, errs *domain.ValidationError) domain.Print {
	usages := make([]domain.FilamentUsage, 0, len(cmd.Usages))
	for row, usage := range cmd.Usages {
		usages = append(usages, domain.FilamentUsage{
			SpoolID: usage.SpoolID,
			Grams:   parseField(errs, domain.FieldUsageGrams(row), usage.Grams, unit.ParseGrams),
		})
	}

	return domain.Print{
		DesignID:      cmd.DesignID,
		Date:          parseField(errs, domain.FieldPrintDate, cmd.Date, unit.ParseDate),
		Quantity:      parseField(errs, domain.FieldQuantity, cmd.Quantity, unit.ParseCopies),
		Minutes:       parseField(errs, domain.FieldMinutes, cmd.Minutes, unit.ParseTime),
		GiftedCount:   parseField(errs, domain.FieldGifted, cmd.Gifted, fallback(0, unit.ParseCopies)),
		KeptCount:     parseField(errs, domain.FieldKept, cmd.Kept, fallback(0, unit.ParseCopies)),
		ScrappedCount: parseField(errs, domain.FieldScrapped, cmd.Scrapped, fallback(0, unit.ParseCopies)),
		Usages:        usages,
	}
}

func parseQuantity(raw string) (unit.Copies, error) {
	errs := &domain.ValidationError{}
	copies := parseField(errs, domain.FieldQuantity, raw, unit.ParseCopies)
	if copies < 1 && errs.For(domain.FieldQuantity) == "" {
		errs.Add(domain.FieldQuantity, "must be at least 1")
	}
	if err := errs.OrNil(); err != nil {
		return 0, err
	}
	return copies, nil
}

// printViewOf is how a Print reads back, and every use case that returns one
// goes through it: what a caller gets is what a later query reports.
func printViewOf(ctx context.Context, db domain.Database, id int64) (PrintView, error) {
	ledger, err := db.PrintLedger(ctx, id)
	if err != nil {
		return PrintView{}, err
	}
	design, err := designOf(ctx, db, ledger.Print.DesignID)
	if err != nil {
		return PrintView{}, err
	}
	spools, err := db.SpoolLedgers(ctx)
	if err != nil {
		return PrintView{}, err
	}
	return toPrintView(ledger, design, spools), nil
}

func toPrintView(l domain.PrintLedger, design domain.Design, ledgers []domain.SpoolLedger) PrintView {
	p := l.Print
	usages := make([]FilamentUsageView, 0, len(p.Usages))
	for _, usage := range p.Usages {
		view := FilamentUsageView{
			SpoolID:     usage.SpoolID,
			Grams:       usage.Grams,
			CostPerGram: usage.CostPerGram,
		}
		for _, ledger := range ledgers {
			if ledger.Spool.ID != usage.SpoolID {
				continue
			}
			view.SpoolBrand = ledger.Spool.Brand
			view.SpoolColor = ledger.Spool.Color
			view.FilamentType = ledger.Spool.FilamentType
		}
		usages = append(usages, view)
	}

	return PrintView{
		ID:              p.ID,
		DesignID:        p.DesignID,
		DesignName:      design.Name,
		Date:            p.Date,
		Quantity:        p.Quantity,
		Minutes:         p.Minutes,
		UsedGrams:       p.UsedGrams(),
		SoldCount:       l.SoldCount,
		GiftedCount:     p.GiftedCount,
		KeptCount:       p.KeptCount,
		ScrappedCount:   p.ScrappedCount,
		AvailableCopies: l.Available(),
		KwhPrice:        p.KwhPrice,
		KwhPerHour:      p.KwhPerHour,
		MachineRate:     p.MachineRate,
		Cost:            toPrintCostView(p),
		Usages:          usages,
	}
}

func toPrintCostView(p domain.Print) PrintCostView {
	cost := p.Cost()
	return PrintCostView{
		Filament:    cost.Filament,
		Energy:      cost.Energy,
		Overhead:    cost.Overhead,
		JobCost:     cost.JobCost(),
		CostPerCopy: cost.CostPerCopy(p.Quantity),
	}
}
