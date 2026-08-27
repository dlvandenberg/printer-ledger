package app

import (
	"context"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

type AddSpoolCmd struct {
	FilamentType domain.FilamentType
	Brand        string
	Color        string
	InitialGrams domain.Grams
	TareGrams    domain.Grams
	PurchaseCost domain.Cents
	PurchaseDate time.Time
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
	spool, err := domain.NewSpool(domain.Spool{
		FilamentType: cmd.FilamentType,
		Brand:        cmd.Brand,
		Color:        cmd.Color,
		InitialGrams: cmd.InitialGrams,
		TareGrams:    cmd.TareGrams,
		PurchaseCost: cmd.PurchaseCost,
		PurchaseDate: cmd.PurchaseDate,
	})
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
