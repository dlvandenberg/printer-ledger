package app

import (
	"context"

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
	LedgerFile          string
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
	return a.toSettingsView(settings), nil
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
		view = a.toSettingsView(stored)
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
	settings.KwhPrice = parseField(errs, domain.FieldKwhPrice, cmd.KwhPrice, unit.ParseCents)
	settings.MachineHourlyRate = parseField(errs, domain.FieldMachineHourlyRate, cmd.MachineHourlyRate, unit.ParseCents)
	settings.PrinterPurchaseCost = parseField(errs, domain.FieldPrinterPurchaseCost, cmd.PrinterPurchaseCost, unit.ParseCents)
	settings.DefaultMargin = parseField(errs, domain.FieldDefaultMargin, cmd.DefaultMargin, unit.ParsePercent)
	settings.MinMargin = parseField(errs, domain.FieldMinMargin, cmd.MinMargin, unit.ParsePercent)

	for _, filamentType := range domain.FilamentTypes() {
		if rate, err := unit.ParseKwhPerHour(cmd.PowerRates[filamentType]); err != nil {
			errs.Add(domain.FieldPowerRate(filamentType), err.Error())
		} else {
			settings = settings.WithPowerRate(filamentType, rate)
		}
	}

	return validate(settings, errs, domain.NewSettings)
}

func (a *App) toSettingsView(s domain.Settings) SettingsView {
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
		LedgerFile:          a.ledgerFile(),
	}
}
