package app

import (
	"context"
	"time"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
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

type EditSpoolCmd struct {
	SpoolID      int64
	FilamentType string
	Brand        string
	Color        string
	InitialGrams string
	TareGrams    string
	PurchaseCost string
	PurchaseDate string
}

type ReweighSpoolCmd struct {
	SpoolID       int64
	MeasuredGrams string
	AdjustedOn    string
	Note          string
}

type SpoolView struct {
	ID             int64
	FilamentType   domain.FilamentType
	Brand          string
	Color          string
	InitialGrams   unit.Grams
	TareGrams      unit.Grams
	PurchaseCost   unit.Cents
	PurchaseDate   time.Time
	UsedGrams      unit.Grams
	AdjustedGrams  unit.Grams
	RemainingGrams unit.Grams
	RemainingValue unit.Cents
	State          domain.SpoolState
}

type SpoolDetailView struct {
	Spool       SpoolView
	Adjustments []SpoolAdjustmentView
}

type SpoolAdjustmentView struct {
	ID               int64
	MeasuredGrams    unit.Grams
	DerivedRemaining unit.Grams
	DeltaGrams       unit.Grams
	AdjustedOn       time.Time
	Note             string
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

// EditSpool corrects a Spool that was entered wrongly. Prints already recorded
// keep the gram price they froze, so a corrected price applies to later prints
// only (ADR-0004).
func (a *App) EditSpool(ctx context.Context, cmd EditSpoolCmd) (SpoolView, error) {
	spool, errs := parseEditSpool(cmd)

	var view SpoolView
	err := a.db.InTx(ctx, func(tx domain.Database) error {
		ledger, err := tx.SpoolLedger(ctx, cmd.SpoolID)
		if err != nil {
			return err
		}
		edited, err := validate(spool, errs, func(s domain.Spool) (domain.Spool, error) {
			return domain.EditedSpool(ledger, s)
		})
		if err != nil {
			return err
		}
		if _, err := tx.UpdateSpool(ctx, edited); err != nil {
			return err
		}
		reread, err := tx.SpoolLedger(ctx, cmd.SpoolID)
		if err != nil {
			return err
		}
		view = toSpoolView(reread)
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

func (a *App) ListActiveSpools(ctx context.Context) ([]SpoolView, error) {
	spools, err := a.ListSpools(ctx)
	if err != nil {
		return nil, err
	}
	active := make([]SpoolView, 0, len(spools))
	for _, spool := range spools {
		if spool.State == domain.SpoolActive {
			active = append(active, spool)
		}
	}
	return active, nil
}

func (a *App) SpoolDetail(ctx context.Context, id int64) (SpoolDetailView, error) {
	ledger, err := a.db.SpoolLedger(ctx, id)
	if err != nil {
		return SpoolDetailView{}, err
	}
	adjustments, err := a.db.SpoolAdjustments(ctx, id)
	if err != nil {
		return SpoolDetailView{}, err
	}
	return toSpoolDetailView(ledger, adjustments), nil
}

func (a *App) ReweighSpool(ctx context.Context, cmd ReweighSpoolCmd) (SpoolDetailView, error) {
	measured, adjustedOn, err := parseReweighSpool(cmd)
	if err != nil {
		return SpoolDetailView{}, err
	}

	var view SpoolDetailView
	err = a.db.InTx(ctx, func(tx domain.Database) error {
		ledger, err := tx.SpoolLedger(ctx, cmd.SpoolID)
		if err != nil {
			return err
		}
		adjustment, err := domain.NewSpoolAdjustment(ledger, measured, adjustedOn, cmd.Note)
		if err != nil {
			return err
		}
		if _, err := tx.CreateSpoolAdjustment(ctx, adjustment); err != nil {
			return err
		}
		reread, err := tx.SpoolLedger(ctx, cmd.SpoolID)
		if err != nil {
			return err
		}
		adjustments, err := tx.SpoolAdjustments(ctx, cmd.SpoolID)
		if err != nil {
			return err
		}
		view = toSpoolDetailView(reread, adjustments)
		return nil
	})
	if err != nil {
		return SpoolDetailView{}, err
	}
	return view, nil
}

func parseReweighSpool(cmd ReweighSpoolCmd) (unit.Grams, time.Time, error) {
	errs := &domain.ValidationError{}
	var (
		measured   unit.Grams
		adjustedOn time.Time
	)

	measured = parseField(errs, domain.FieldMeasuredGrams, cmd.MeasuredGrams, unit.ParseGrams)
	adjustedOn = parseField(errs, domain.FieldAdjustedOn, cmd.AdjustedOn, unit.ParseDate)

	if err := errs.OrNil(); err != nil {
		return 0, time.Time{}, err
	}
	return measured, adjustedOn, nil
}

func parseAddSpool(cmd AddSpoolCmd) (domain.Spool, error) {
	errs := &domain.ValidationError{}
	spool := domain.Spool{
		Brand:        cmd.Brand,
		Color:        cmd.Color,
		FilamentType: parseField(errs, domain.FieldFilamentType, cmd.FilamentType, domain.ParseFilamentType),
		InitialGrams: parseField(errs, domain.FieldInitialGrams, cmd.InitialGrams, unit.ParseGrams),
		TareGrams:    parseField(errs, domain.FieldTareGrams, cmd.TareGrams, unit.ParseGrams),
		PurchaseCost: parseField(errs, domain.FieldPurchaseCost, cmd.PurchaseCost, unit.ParseCents),
		PurchaseDate: parseField(errs, domain.FieldPurchaseDate, cmd.PurchaseDate, unit.ParseDate),
	}

	return validate(spool, errs, domain.NewSpool)
}

// parseEditSpool returns the parse errors rather than the aggregate: the
// invariants need the Spool's ledger, which only the transaction can read.
func parseEditSpool(cmd EditSpoolCmd) (domain.Spool, *domain.ValidationError) {
	errs := &domain.ValidationError{}
	spool := domain.Spool{
		ID:           cmd.SpoolID,
		Brand:        cmd.Brand,
		Color:        cmd.Color,
		FilamentType: parseField(errs, domain.FieldFilamentType, cmd.FilamentType, domain.ParseFilamentType),
		InitialGrams: parseField(errs, domain.FieldInitialGrams, cmd.InitialGrams, unit.ParseGrams),
		TareGrams:    parseField(errs, domain.FieldTareGrams, cmd.TareGrams, unit.ParseGrams),
		PurchaseCost: parseField(errs, domain.FieldPurchaseCost, cmd.PurchaseCost, unit.ParseCents),
		PurchaseDate: parseField(errs, domain.FieldPurchaseDate, cmd.PurchaseDate, unit.ParseDate),
	}

	return spool, errs
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
		UsedGrams:      l.UsedGrams,
		AdjustedGrams:  l.AdjustedGrams,
		RemainingGrams: l.Remaining(),
		RemainingValue: l.RemainingValue(),
		State:          l.State(),
	}
}

func toSpoolDetailView(l domain.SpoolLedger, adjustments []domain.SpoolAdjustment) SpoolDetailView {
	views := make([]SpoolAdjustmentView, 0, len(adjustments))
	for _, a := range adjustments {
		views = append(views, SpoolAdjustmentView{
			ID:               a.ID,
			MeasuredGrams:    a.MeasuredGrams,
			DerivedRemaining: a.DerivedRemaining(l.Spool.TareGrams),
			DeltaGrams:       a.DeltaGrams,
			AdjustedOn:       a.AdjustedOn,
			Note:             a.Note,
		})
	}
	return SpoolDetailView{Spool: toSpoolView(l), Adjustments: views}
}
