package app

import (
	"context"
	"errors"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

type UpdateSettingsCmd struct {
	KwhPrice            string
	MachineHourlyRate   string
	PrinterPurchaseCost string
	DefaultMargin       string
	MinMargin           string
	PowerRates          map[domain.FilamentType]string
}

type SettingsView struct {
	KwhPrice            unit.Cents
	MachineHourlyRate   unit.Cents
	PrinterPurchaseCost unit.Cents
	DefaultMargin       unit.Percent
	MinMargin           unit.Percent
	PowerRates          []PowerRateView
}

type PowerRateView struct {
	FilamentType domain.FilamentType
	KwhPerHour   unit.KwhPerHour
	Measured     bool
}

func (a *App) Settings(ctx context.Context) (SettingsView, error) {
	settings, err := a.db.Settings(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	return toSettingsView(settings), nil
}

func (a *App) UpdateSettings(ctx context.Context, cmd UpdateSettingsCmd) (SettingsView, error) {
	var view SettingsView
	err := a.db.InTx(ctx, func(tx domain.Database) error {
		current, err := tx.Settings(ctx)
		if err != nil {
			return err
		}
		updated, err := parseUpdateSettings(cmd, current)
		if err != nil {
			return err
		}
		if err := tx.SaveSettings(ctx, updated); err != nil {
			return err
		}
		stored, err := tx.Settings(ctx)
		if err != nil {
			return err
		}
		view = toSettingsView(stored)
		return nil
	})
	if err != nil {
		return SettingsView{}, err
	}
	return view, nil
}

func parseUpdateSettings(cmd UpdateSettingsCmd, current domain.Settings) (domain.Settings, error) {
	errs := &domain.ValidationError{}
	settings := current

	if cents, err := unit.ParseCents(cmd.KwhPrice); err != nil {
		errs.Add(domain.FieldKwhPrice, err.Error())
	} else {
		settings.KwhPrice = cents
	}

	if cents, err := unit.ParseCents(cmd.MachineHourlyRate); err != nil {
		errs.Add(domain.FieldMachineHourlyRate, err.Error())
	} else {
		settings.MachineHourlyRate = cents
	}

	if cents, err := unit.ParseCents(cmd.PrinterPurchaseCost); err != nil {
		errs.Add(domain.FieldPrinterPurchaseCost, err.Error())
	} else {
		settings.PrinterPurchaseCost = cents
	}

	if pct, err := unit.ParsePercent(cmd.DefaultMargin); err != nil {
		errs.Add(domain.FieldDefaultMargin, err.Error())
	} else {
		settings.DefaultMargin = pct
	}

	if pct, err := unit.ParsePercent(cmd.MinMargin); err != nil {
		errs.Add(domain.FieldMinMargin, err.Error())
	} else {
		settings.MinMargin = pct
	}

	for _, filamentType := range domain.FilamentTypes() {
		if rate, err := unit.ParseKwhPerHour(cmd.PowerRates[filamentType]); err != nil {
			errs.Add(domain.FieldPowerRate(filamentType), err.Error())
		} else {
			settings = settings.WithPowerRate(filamentType, rate)
		}
	}

	valid, err := domain.NewSettings(settings)
	if err != nil {
		var invariants *domain.ValidationError
		if !errors.As(err, &invariants) {
			return domain.Settings{}, err
		}
		errs.MergeMissing(invariants)
	}
	if err := errs.OrNil(); err != nil {
		return domain.Settings{}, err
	}
	return valid, nil
}

func toSettingsView(s domain.Settings) SettingsView {
	rates := make([]PowerRateView, 0, len(s.PowerRates))
	for _, filamentType := range domain.FilamentTypes() {
		rate := s.PowerRate(filamentType)
		rates = append(rates, PowerRateView{
			FilamentType: filamentType,
			KwhPerHour:   rate.KwhPerHour,
			Measured:     rate.Measured,
		})
	}
	return SettingsView{
		KwhPrice:            s.KwhPrice,
		MachineHourlyRate:   s.MachineHourlyRate,
		PrinterPurchaseCost: s.PrinterPurchaseCost,
		DefaultMargin:       s.DefaultMargin,
		MinMargin:           s.MinMargin,
		PowerRates:          rates,
	}
}
