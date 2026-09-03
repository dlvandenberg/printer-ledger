package tui

import (
	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func newDesignForm(defaultMargin unit.Percent) *form {
	return newForm("Add design", []fieldSpec{
		{Key: domain.FieldName, Label: "Name", Placeholder: "Design name"},
		{Key: domain.FieldDefaultFilamentType, Label: "Default Filament", Choices: domain.FilamentTypeNames()},
		{Key: domain.FieldEstimatedGrams, Label: "Estimated grams", Placeholder: "60"},
		{Key: domain.FieldEstimatedMinutes, Label: "Estimated time", Placeholder: "0:50"},
		{Key: domain.FieldMarginPct, Label: "Margin", Placeholder: "60%", Prefill: unit.FormatPercent(defaultMargin)},
	})
}

func editDesignForm(d app.DesignView) *form {
	return newForm("Edit design", []fieldSpec{
		{Key: domain.FieldName, Label: "Name", Placeholder: "Design name", Prefill: d.Name},
		{Key: domain.FieldDefaultFilamentType, Label: "Default Filament", Choices: domain.FilamentTypeNames(), Choice: string(d.DefaultFilamentType)},
		{Key: domain.FieldEstimatedGrams, Label: "Estimated grams", Placeholder: "60", Prefill: unit.FormatGrams(d.EstimatedGrams)},
		{Key: domain.FieldEstimatedMinutes, Label: "Estimated time", Placeholder: "0:50", Prefill: unit.FormatHHmm(d.EstimatedMinutes)},
		{Key: domain.FieldMarginPct, Label: "Margin", Placeholder: "60%", Prefill: unit.FormatPercent(d.MarginPct)},
	})
}

func addDesignCmd(f *form) app.AddDesignCmd {
	return app.AddDesignCmd{
		Name:                f.Value(domain.FieldName),
		DefaultFilamentType: f.Value(domain.FieldDefaultFilamentType),
		EstimatedGrams:      f.Value(domain.FieldEstimatedGrams),
		EstimatedMinutes:    f.Value(domain.FieldEstimatedMinutes),
		MarginPct:           f.Value(domain.FieldMarginPct),
	}
}

func editDesignCmd(f *form, id int64) app.EditDesignCmd {
	return app.EditDesignCmd{
		DesignID:            id,
		Name:                f.Value(domain.FieldName),
		DefaultFilamentType: f.Value(domain.FieldDefaultFilamentType),
		EstimatedGrams:      f.Value(domain.FieldEstimatedGrams),
		EstimatedMinutes:    f.Value(domain.FieldEstimatedMinutes),
		MarginPct:           f.Value(domain.FieldMarginPct),
	}
}
