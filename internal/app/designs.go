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
	PrintCount          int64
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
	ledgers, err := a.db.DesignLedgers(ctx)
	if err != nil {
		return nil, err
	}

	views := make([]DesignView, 0, len(ledgers))
	for _, ledger := range ledgers {
		views = append(views, toDesignView(ledger))
	}

	return views, nil
}

func (a *App) DesignQuote(ctx context.Context, id int64) (DesignQuoteView, error) {
	ledger, err := a.db.DesignLedger(ctx, id)
	if err != nil {
		return DesignQuoteView{}, err
	}
	settings, err := a.db.Settings(ctx)
	if err != nil {
		return DesignQuoteView{}, err
	}
	spools, err := a.db.SpoolLedgers(ctx)
	if err != nil {
		return DesignQuoteView{}, err
	}

	return toDesignQuoteView(ledger, domain.NewDesignQuote(ledger.Design, settings, spools)), nil
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
		stored, err := tx.DesignLedger(ctx, created.ID)
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
		if _, err := tx.DesignLedger(ctx, cmd.DesignID); err != nil {
			return err
		}

		design, err := parseEditDesign(cmd)
		if err != nil {
			return err
		}

		if _, err := tx.UpdateDesign(ctx, design); err != nil {
			return err
		}
		stored, err := tx.DesignLedger(ctx, cmd.DesignID)
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

// DeleteDesign refuses a Design a Print references: the Print is a cost history
// and would be left describing nothing.
func (a *App) DeleteDesign(ctx context.Context, id int64) error {
	return a.db.InTx(ctx, func(tx domain.Database) error {
		ledger, err := tx.DesignLedger(ctx, id)
		if err != nil {
			return err
		}
		if err := ledger.DeleteBlocked(); err != nil {
			return err
		}
		return tx.DeleteDesign(ctx, id)
	})
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

func toDesignView(l domain.DesignLedger) DesignView {
	d := l.Design
	return DesignView{
		ID:                  d.ID,
		Name:                d.Name,
		DefaultFilamentType: d.DefaultFilamentType,
		EstimatedGrams:      d.EstimatedGrams,
		EstimatedMinutes:    d.EstimatedMinutes,
		MarginPct:           d.MarginPct,
		PrintCount:          l.PrintCount,
	}
}

func toDesignQuoteView(l domain.DesignLedger, q domain.DesignQuote) DesignQuoteView {
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
		Design:                  toDesignView(l),
		Costs:                   costs,
		SuggestedPrice:          suggested,
		HasSuggestedPrice:       hasSuggested,
		SuggestedPriceReference: hasSuggested && defaultCost.Reference,
	}
}

func designsByID(designs []domain.Design) map[int64]domain.Design {
	byID := make(map[int64]domain.Design, len(designs))
	for _, design := range designs {
		byID[design.ID] = design
	}
	return byID
}
