package app

import (
	"context"
	"errors"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

type RecordSaleCmd struct {
	PrintID int64
	Price   string
	Date    string
}

// EditSaleCmd is a RecordSaleCmd against a Sale that already exists: an edit
// re-types every field the record form has.
type EditSaleCmd struct {
	RecordSaleCmd

	SaleID int64
}

type SaleView struct {
	ID          int64
	PrintID     int64
	DesignName  string
	Material    MaterialView
	Date        time.Time
	Price       unit.Cents
	CostPerCopy unit.Cents
}

// MaterialView is what a Print's copies came out in: two Prints of one Design
// are interchangeable to the ledger but not to a buyer, who is looking at the
// color. A Print that swapped spools mid-run names every color it used, most
// filament first.
type MaterialView struct {
	FilamentType domain.FilamentType
	Colors       []string
}

func materialOf(print domain.Print, spools []domain.SpoolLedger) MaterialView {
	return MaterialView{
		FilamentType: print.FilamentType(spools),
		Colors:       print.SpoolColors(spools),
	}
}

// SellablePrintView is a Print a copy can still be sold from, with the two
// numbers the operator decides a price against: what the copy cost and what the
// Design suggests charging.
type SellablePrintView struct {
	PrintID           int64
	DesignID          int64
	DesignName        string
	Material          MaterialView
	Date              time.Time
	AvailableCopies   unit.Copies
	CostPerCopy       unit.Cents
	SuggestedPrice    unit.Cents
	HasSuggestedPrice bool
}

// SalePreviewView is what the Sale form shows while the price is being typed.
// BelowFloor is the warning of ADR-0016; it never stops the Sale being
// recorded.
type SalePreviewView struct {
	CostPerCopy       unit.Cents
	SuggestedPrice    unit.Cents
	HasSuggestedPrice bool
	MinimumPrice      unit.Cents
	BelowFloor        bool
}

func (a *App) RecordSale(ctx context.Context, cmd RecordSaleCmd) (SaleView, error) {
	var view SaleView
	err := a.db.InTx(ctx, func(tx domain.Database) error {
		ledger, err := salePrintLedger(ctx, tx, cmd.PrintID)
		if err != nil {
			return err
		}

		sale, err := parseRecordSale(cmd, ledger)
		if err != nil {
			return err
		}

		created, err := tx.CreateSale(ctx, sale)
		if err != nil {
			return err
		}
		stored, err := tx.Sale(ctx, created.ID)
		if err != nil {
			return err
		}
		view, err = readSaleView(ctx, tx, stored)
		return err
	})
	if err != nil {
		return SaleView{}, err
	}
	return view, nil
}

// EditSale corrects a Sale in place (ADR-0011). A mistyped price is the reason
// it exists; moving the Sale to another Print returns the copy to the one it
// came from, availability being derived (ADR-0013).
func (a *App) EditSale(ctx context.Context, cmd EditSaleCmd) (SaleView, error) {
	var view SaleView
	err := a.db.InTx(ctx, func(tx domain.Database) error {
		stored, err := tx.Sale(ctx, cmd.SaleID)
		if err != nil {
			return err
		}
		ledger, err := salePrintLedger(ctx, tx, cmd.PrintID)
		if err != nil {
			return err
		}

		edited, err := parseEditSale(cmd, stored, ledger)
		if err != nil {
			return err
		}

		if _, err := tx.UpdateSale(ctx, edited); err != nil {
			return err
		}
		reread, err := tx.Sale(ctx, cmd.SaleID)
		if err != nil {
			return err
		}
		view, err = readSaleView(ctx, tx, reread)
		return err
	})
	if err != nil {
		return SaleView{}, err
	}
	return view, nil
}

// DeleteSale returns the copy to its Print, which needs no write of its own:
// available derives from the Sales that remain (ADR-0013).
func (a *App) DeleteSale(ctx context.Context, id int64) error {
	return a.db.InTx(ctx, func(tx domain.Database) error {
		return tx.DeleteSale(ctx, id)
	})
}

func (a *App) ListSales(ctx context.Context) ([]SaleView, error) {
	sales, err := a.db.Sales(ctx)
	if err != nil {
		return nil, err
	}
	ledgers, err := a.db.PrintLedgers(ctx)
	if err != nil {
		return nil, err
	}
	byDesign, err := designsByID(ctx, a.db)
	if err != nil {
		return nil, err
	}
	spools, err := a.db.SpoolLedgers(ctx)
	if err != nil {
		return nil, err
	}

	byPrint := make(map[int64]domain.PrintLedger, len(ledgers))
	for _, ledger := range ledgers {
		byPrint[ledger.Print.ID] = ledger
	}

	views := make([]SaleView, 0, len(sales))
	for _, sale := range sales {
		ledger := byPrint[sale.PrintID]
		views = append(views, toSaleView(sale, ledger, byDesign[ledger.Print.DesignID], spools))
	}
	return views, nil
}

// SellablePrints offers the Prints a copy is still available from, so the same
// object cannot be sold twice.
func (a *App) SellablePrints(ctx context.Context) ([]SellablePrintView, error) {
	return a.sellablePrints(ctx, domain.Sale{})
}

// SellablePrintsForSale is what a Sale being edited may name. The copy that
// Sale already holds returns to its Print before availability is counted, so an
// edit keeps naming the Print it sold from even when that copy was the last.
func (a *App) SellablePrintsForSale(ctx context.Context, saleID int64) ([]SellablePrintView, error) {
	released, err := a.db.Sale(ctx, saleID)
	if err != nil {
		return nil, err
	}
	return a.sellablePrints(ctx, released)
}

func (a *App) sellablePrints(ctx context.Context, released domain.Sale) ([]SellablePrintView, error) {
	settings, err := a.db.Settings(ctx)
	if err != nil {
		return nil, err
	}
	ledgers, err := a.db.PrintLedgers(ctx)
	if err != nil {
		return nil, err
	}
	byID, err := designsByID(ctx, a.db)
	if err != nil {
		return nil, err
	}
	spools, err := a.db.SpoolLedgers(ctx)
	if err != nil {
		return nil, err
	}

	suggested := map[int64]unit.Cents{}
	hasSuggested := map[int64]bool{}
	for _, design := range byID {
		price, ok := domain.NewDesignQuote(design, settings, spools).SuggestedPrice()
		suggested[design.ID], hasSuggested[design.ID] = price, ok
	}

	views := make([]SellablePrintView, 0, len(ledgers))
	for _, ledger := range ledgers {
		if released.PrintID == ledger.Print.ID {
			ledger.SoldCount--
		}
		if ledger.Available() < 1 {
			continue
		}
		design := byID[ledger.Print.DesignID]
		views = append(views, SellablePrintView{
			PrintID:           ledger.Print.ID,
			DesignID:          design.ID,
			DesignName:        design.Name,
			Material:          materialOf(ledger.Print, spools),
			Date:              ledger.Print.Date,
			AvailableCopies:   ledger.Available(),
			CostPerCopy:       ledger.Print.CostPerCopy(),
			SuggestedPrice:    suggested[design.ID],
			HasSuggestedPrice: hasSuggested[design.ID],
		})
	}
	return views, nil
}

// PreviewSale is the cost, the suggested price and the floor warning of a Sale
// still being typed. A Print not yet picked shows nothing rather than failing,
// which is what keeps the panel on screen.
func (a *App) PreviewSale(ctx context.Context, cmd RecordSaleCmd) (SalePreviewView, error) {
	settings, err := a.db.Settings(ctx)
	if err != nil {
		return SalePreviewView{}, err
	}
	ledger, err := a.db.PrintLedger(ctx, cmd.PrintID)
	if errors.Is(err, domain.ErrNotFound) {
		return SalePreviewView{}, nil
	}
	if err != nil {
		return SalePreviewView{}, err
	}
	design, err := designOf(ctx, a.db, ledger.Print.DesignID)
	if err != nil {
		return SalePreviewView{}, err
	}
	spools, err := a.db.SpoolLedgers(ctx)
	if err != nil {
		return SalePreviewView{}, err
	}

	cost := ledger.Print.CostPerCopy()
	price, priced := draftPrice(cmd.Price)
	suggested, hasSuggested := domain.NewDesignQuote(design, settings, spools).SuggestedPrice()
	return SalePreviewView{
		CostPerCopy:       cost,
		SuggestedPrice:    suggested,
		HasSuggestedPrice: hasSuggested,
		MinimumPrice:      domain.MinimumPrice(cost, settings.MinMargin),
		BelowFloor:        priced && domain.BelowFloor(price, cost, settings.MinMargin),
	}, nil
}

func parseRecordSale(cmd RecordSaleCmd, ledger domain.PrintLedger) (domain.Sale, error) {
	errs := &domain.ValidationError{}
	drafted := draftSale(cmd, errs)

	return validate(drafted, errs, func(s domain.Sale) (domain.Sale, error) {
		return domain.NewSale(s, ledger)
	})
}

func parseEditSale(cmd EditSaleCmd, stored domain.Sale, ledger domain.PrintLedger) (domain.Sale, error) {
	errs := &domain.ValidationError{}
	drafted := draftSale(cmd.RecordSaleCmd, errs)

	return validate(drafted, errs, func(s domain.Sale) (domain.Sale, error) {
		return domain.EditedSale(stored, s, ledger)
	})
}

func draftSale(cmd RecordSaleCmd, errs *domain.ValidationError) domain.Sale {
	return domain.Sale{
		PrintID: cmd.PrintID,
		Price:   parseField(errs, domain.FieldSalePrice, cmd.Price, unit.ParseCents),
		Date:    parseField(errs, domain.FieldSaleDate, cmd.Date, unit.ParseDate),
	}
}

// draftPrice reports whether there is a price to warn about at all: an empty or
// half-typed field is not yet a sale below the floor.
func draftPrice(raw string) (unit.Cents, bool) {
	price, err := unit.ParseCents(raw)
	if err != nil {
		return 0, false
	}
	return price, true
}

// salePrintLedger reads the Print a Sale names. Naming no Print at all is a
// validation failure on the choice rather than a missing record.
func salePrintLedger(ctx context.Context, tx domain.Database, printID int64) (domain.PrintLedger, error) {
	if printID == 0 {
		return domain.PrintLedger{}, nil
	}
	return tx.PrintLedger(ctx, printID)
}

func readSaleView(ctx context.Context, tx domain.Database, sale domain.Sale) (SaleView, error) {
	ledger, err := tx.PrintLedger(ctx, sale.PrintID)
	if err != nil {
		return SaleView{}, err
	}
	design, err := designOf(ctx, tx, ledger.Print.DesignID)
	if err != nil {
		return SaleView{}, err
	}
	spools, err := tx.SpoolLedgers(ctx)
	if err != nil {
		return SaleView{}, err
	}
	return toSaleView(sale, ledger, design, spools), nil
}

// toSaleView carries the frozen cost of the copy but no floor verdict: the
// floor moves with Settings.minMargin, and a warning is a thing said while the
// price is being typed, not a judgement re-passed on a past sale (ADR-0016).
func toSaleView(sale domain.Sale, ledger domain.PrintLedger, design domain.Design, spools []domain.SpoolLedger) SaleView {
	return SaleView{
		ID:          sale.ID,
		PrintID:     sale.PrintID,
		DesignName:  design.Name,
		Material:    materialOf(ledger.Print, spools),
		Date:        sale.Date,
		Price:       sale.Price,
		CostPerCopy: ledger.Print.CostPerCopy(),
	}
}
