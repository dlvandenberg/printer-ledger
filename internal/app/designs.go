package app

import (
	"context"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

type AddDesignCmd struct {
	Name                string
	EstimatedGrams      string
	EstimatedMinutes    string
	DefaultFilamentType string
	MarginPct           string
}

type EditDesignCmd struct {
	DesignID            int64
	Name                string
	EstimatedGrams      string
	EstimatedMinutes    string
	DefaultFilamentType string
	MarginPct           string
}

type DesignView struct {
	ID                  int64
	Name                string
	DefaultFilamentType domain.FilamentType
	EstimatedGrams      unit.Grams
	EstimatedMinutes    unit.Minutes
	MarginPct           unit.Percent
}

type DesignQuoteView struct {
	Design            DesignView
	Costs             []EstimatedCostView
	SuggestedPrice    unit.Cents
	HasSuggestedPrice bool
	// SuggestedPriceReference is true when the price comes from a row that no
	// Capable Spool backs, so the operator is quoting filament not in stock.
	SuggestedPriceReference bool
}

type EstimatedCostView struct {
	FilamentType domain.FilamentType
	IsDefault    bool
	Filament     unit.Cents
	Energy       unit.Cents
	Overhead     unit.Cents
	Total        unit.Cents
	Reference    bool
	SpoolID      int64
	SpoolBrand   string
	SpoolColor   string
}

func (a *App) ListDesigns(ctx context.Context) ([]DesignView, error) {
	designs, err := a.db.Designs(ctx)
	if err != nil {
		return nil, err
	}

	views := make([]DesignView, 0, len(designs))
	for _, design := range designs {
		views = append(views, toDesignView(design))
	}

	return views, nil
}

func (a *App) DesignQuote(ctx context.Context, id int64) (DesignQuoteView, error) {
	design, err := a.db.Design(ctx, id)
	if err != nil {
		return DesignQuoteView{}, err
	}
	settings, err := a.db.Settings(ctx)
	if err != nil {
		return DesignQuoteView{}, err
	}
	ledgers, err := a.db.SpoolLedgers(ctx)
	if err != nil {
		return DesignQuoteView{}, err
	}

	return toDesignQuoteView(domain.NewDesignQuote(design, settings, ledgers)), nil
}

func (a *App) AddDesign(ctx context.Context, cmd AddDesignCmd) (DesignView, error) {
	var view DesignView
	err := a.db.InTx(ctx, func(tx domain.Database) error {
		settings, err := tx.Settings(ctx)
		if err != nil {
			return err
		}

		design, err := parseAddDesign(cmd, settings.DefaultMargin)
		if err != nil {
			return err
		}

		created, err := tx.CreateDesign(ctx, design)
		if err != nil {
			return err
		}
		stored, err := tx.Design(ctx, created.ID)
		if err != nil {
			return err
		}
		view = toDesignView(stored)
		return nil
	})

	if err != nil {
		return DesignView{}, err
	}

	return view, nil
}

func (a *App) EditDesign(ctx context.Context, cmd EditDesignCmd) (DesignView, error) {
	var view DesignView
	err := a.db.InTx(ctx, func(tx domain.Database) error {
		_, err := tx.Design(ctx, cmd.DesignID)
		if err != nil {
			return err
		}

		design, err := parseEditDesign(cmd)
		if err != nil {
			return err
		}

		updated, err := tx.UpdateDesign(ctx, design)
		if err != nil {
			return err
		}
		view = toDesignView(updated)
		return nil
	})

	if err != nil {
		return DesignView{}, err
	}

	return view, nil
}

func parseAddDesign(cmd AddDesignCmd, defaultMargin unit.Percent) (domain.Design, error) {
	errs := &domain.ValidationError{}
	design := domain.Design{
		Name:                cmd.Name,
		DefaultFilamentType: parseField(errs, domain.FieldDefaultFilamentType, cmd.DefaultFilamentType, domain.ParseFilamentType),
		EstimatedGrams:      parseField(errs, domain.FieldEstimatedGrams, cmd.EstimatedGrams, unit.ParseGrams),
		EstimatedMinutes:    parseField(errs, domain.FieldEstimatedMinutes, cmd.EstimatedMinutes, unit.ParseTime),
		MarginPct:           parseField(errs, domain.FieldMarginPct, cmd.MarginPct, fallback(defaultMargin, unit.ParsePercent)),
	}

	return validate(design, errs, domain.NewDesign)
}

func parseEditDesign(cmd EditDesignCmd) (domain.Design, error) {
	errs := &domain.ValidationError{}
	design := domain.Design{
		ID:                  cmd.DesignID,
		Name:                cmd.Name,
		DefaultFilamentType: parseField(errs, domain.FieldDefaultFilamentType, cmd.DefaultFilamentType, domain.ParseFilamentType),
		EstimatedGrams:      parseField(errs, domain.FieldEstimatedGrams, cmd.EstimatedGrams, unit.ParseGrams),
		EstimatedMinutes:    parseField(errs, domain.FieldEstimatedMinutes, cmd.EstimatedMinutes, unit.ParseTime),
		MarginPct:           parseField(errs, domain.FieldMarginPct, cmd.MarginPct, unit.ParsePercent),
	}

	return validate(design, errs, domain.NewDesign)
}

func toDesignView(d domain.Design) DesignView {
	return DesignView{
		ID:                  d.ID,
		Name:                d.Name,
		DefaultFilamentType: d.DefaultFilamentType,
		EstimatedGrams:      d.EstimatedGrams,
		EstimatedMinutes:    d.EstimatedMinutes,
		MarginPct:           d.MarginPct,
	}
}

func toDesignQuoteView(q domain.DesignQuote) DesignQuoteView {
	costs := make([]EstimatedCostView, 0, len(q.Costs))
	for _, cost := range q.Costs {
		costs = append(costs, EstimatedCostView{
			FilamentType: cost.FilamentType,
			IsDefault:    cost.FilamentType == q.Design.DefaultFilamentType,
			Filament:     cost.Filament,
			Energy:       cost.Energy,
			Overhead:     cost.Overhead,
			Total:        cost.Total(),
			Reference:    cost.Reference,
			SpoolID:      cost.Spool.ID,
			SpoolBrand:   cost.Spool.Brand,
			SpoolColor:   cost.Spool.Color,
		})
	}

	suggested, hasSuggested := q.SuggestedPrice()
	defaultCost, _ := q.DefaultCost()
	return DesignQuoteView{
		Design:                  toDesignView(q.Design),
		Costs:                   costs,
		SuggestedPrice:          suggested,
		HasSuggestedPrice:       hasSuggested,
		SuggestedPriceReference: hasSuggested && defaultCost.Reference,
	}
}
