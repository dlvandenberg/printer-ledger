package tui

import (
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func newSettingsForm(s app.SettingsView) *form {
	specs := []fieldSpec{
		{Key: domain.FieldKwhPrice, Label: "Electricity per kWh", Placeholder: "0.28", Prefill: unit.FormatCents(s.KwhPrice)},
	}
	for _, rate := range s.PowerRates {
		specs = append(specs, fieldSpec{
			Key:         domain.FieldPowerRate(rate.FilamentType),
			Label:       fmt.Sprintf("%s kWh/h (%s)", rate.FilamentType, rateSource(rate.Measured)),
			Placeholder: "0.09",
			Prefill:     unit.FormatKwhPerHour(rate.KwhPerHour),
		})
	}
	specs = append(specs,
		fieldSpec{Key: domain.FieldMachineHourlyRate, Label: "Machine per hour", Placeholder: "0.35", Prefill: unit.FormatCents(s.MachineHourlyRate)},
		fieldSpec{Key: domain.FieldPrinterPurchaseCost, Label: "Printer purchase cost", Placeholder: "399.00", Prefill: unit.FormatCents(s.PrinterPurchaseCost)},
		fieldSpec{Key: domain.FieldDefaultMargin, Label: "Default margin", Placeholder: "50%", Prefill: unit.FormatPercent(s.DefaultMargin)},
		fieldSpec{Key: domain.FieldMinMargin, Label: "Minimum margin", Placeholder: "15%", Prefill: unit.FormatPercent(s.MinMargin)},
	)
	return newForm("Edit settings", specs)
}

func updateSettingsCmd(f *form) app.UpdateSettingsCmd {
	rates := map[domain.FilamentType]string{}
	for _, filamentType := range domain.FilamentTypes() {
		rates[filamentType] = f.Value(domain.FieldPowerRate(filamentType))
	}
	return app.UpdateSettingsCmd{
		KwhPrice:            f.Value(domain.FieldKwhPrice),
		MachineHourlyRate:   f.Value(domain.FieldMachineHourlyRate),
		PrinterPurchaseCost: f.Value(domain.FieldPrinterPurchaseCost),
		DefaultMargin:       f.Value(domain.FieldDefaultMargin),
		MinMargin:           f.Value(domain.FieldMinMargin),
		PowerRates:          rates,
	}
}
