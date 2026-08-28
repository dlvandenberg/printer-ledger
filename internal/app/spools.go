package app

import (
	"context"
	"errors"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

// AddSpoolCmd carries what the operator typed. Parsing it is the use case's
// job, so every input path is reachable from an application-layer test.
type AddSpoolCmd struct {
	FilamentType string
	Brand        string
	Color        string
	InitialGrams string
	TareGrams    string
	PurchaseCost string
	PurchaseDate string
}

type SpoolView struct {
	ID             int64
	FilamentType   domain.FilamentType
	Brand          string
	Color          string
	InitialGrams   domain.Grams
	TareGrams      domain.Grams
	PurchaseCost   domain.Cents
	PurchaseDate   time.Time
	RemainingGrams domain.Grams
	RemainingValue domain.Cents
	State          domain.SpoolState
}

func (a *App) AddSpool(ctx context.Context, cmd AddSpoolCmd) (SpoolView, error) {
	spool, err := parseAddSpool(cmd)
	if err != nil {
		return SpoolView{}, err
	}

	var view SpoolView
	err = a.db.InTx(ctx, func(tx domain.Database) error {
		created, err := tx.CreateSpool(ctx, spool)
		if err != nil {
			return err
		}
		ledger, err := tx.SpoolLedger(ctx, created.ID)
		if err != nil {
			return err
		}
		view = toSpoolView(ledger)
		return nil
	})
	if err != nil {
		return SpoolView{}, err
	}
	return view, nil
}

func (a *App) ListSpools(ctx context.Context) ([]SpoolView, error) {
	ledgers, err := a.db.SpoolLedgers(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]SpoolView, 0, len(ledgers))
	for _, ledger := range ledgers {
		views = append(views, toSpoolView(ledger))
	}
	return views, nil
}

func parseAddSpool(cmd AddSpoolCmd) (domain.Spool, error) {
	errs := &domain.ValidationError{}
	spool := domain.Spool{
		Brand: cmd.Brand,
		Color: cmd.Color,
	}

	if filamentType, err := domain.ParseFilamentType(cmd.FilamentType); err != nil {
		errs.Add(domain.FieldFilamentType, err.Error())
	} else {
		spool.FilamentType = filamentType
	}

	if grams, err := domain.ParseGrams(cmd.InitialGrams); err != nil {
		errs.Add(domain.FieldInitialGrams, err.Error())
	} else {
		spool.InitialGrams = grams
	}

	if grams, err := domain.ParseGrams(cmd.TareGrams); err != nil {
		errs.Add(domain.FieldTareGrams, err.Error())
	} else {
		spool.TareGrams = grams
	}

	if cents, err := domain.ParseCents(cmd.PurchaseCost); err != nil {
		errs.Add(domain.FieldPurchaseCost, err.Error())
	} else {
		spool.PurchaseCost = cents
	}

	if date, err := domain.ParseDate(cmd.PurchaseDate); err != nil {
		errs.Add(domain.FieldPurchaseDate, err.Error())
	} else {
		spool.PurchaseDate = date
	}

	valid, err := domain.NewSpool(spool)
	if err != nil {
		var invariants *domain.ValidationError
		if !errors.As(err, &invariants) {
			return domain.Spool{}, err
		}
		errs.MergeMissing(invariants)
	}
	if err := errs.OrNil(); err != nil {
		return domain.Spool{}, err
	}
	return valid, nil
}

func toSpoolView(l domain.SpoolLedger) SpoolView {
	return SpoolView{
		ID:             l.Spool.ID,
		FilamentType:   l.Spool.FilamentType,
		Brand:          l.Spool.Brand,
		Color:          l.Spool.Color,
		InitialGrams:   l.Spool.InitialGrams,
		TareGrams:      l.Spool.TareGrams,
		PurchaseCost:   l.Spool.PurchaseCost,
		PurchaseDate:   l.Spool.PurchaseDate,
		RemainingGrams: l.Remaining(),
		RemainingValue: l.RemainingValue(),
		State:          l.State(),
	}
}
