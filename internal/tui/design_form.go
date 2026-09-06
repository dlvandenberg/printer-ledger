package tui

import (
	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func newDesignForm(defaultMargin unit.Percent) *form {
	return newForm("Add design", []fieldSpec{
		textSpec{Key: domain.FieldName, Label: "Name", Placeholder: "Design name"},
		choiceSpec{Key: domain.FieldDefaultFilamentType, Label: "Default Filament", Choices: domain.FilamentTypeNames()},
		textSpec{Key: domain.FieldEstimatedGrams, Label: "Estimated grams", Placeholder: "60"},
		textSpec{Key: domain.FieldEstimatedMinutes, Label: "Estimated time", Placeholder: "0:50"},
		textSpec{Key: domain.FieldMarginPct, Label: "Margin", Placeholder: "60%", Prefill: unit.FormatPercent(defaultMargin)},
	})
}

func editDesignForm(d app.DesignView) *form {
	return newForm("Edit design", []fieldSpec{
		textSpec{Key: domain.FieldName, Label: "Name", Placeholder: "Design name", Prefill: d.Name},
		choiceSpec{Key: domain.FieldDefaultFilamentType, Label: "Default Filament", Choices: domain.FilamentTypeNames(), Choice: string(d.DefaultFilamentType)},
		textSpec{Key: domain.FieldEstimatedGrams, Label: "Estimated grams", Placeholder: "60", Prefill: unit.FormatGrams(d.EstimatedGrams)},
		textSpec{Key: domain.FieldEstimatedMinutes, Label: "Estimated time", Placeholder: "0:50", Prefill: unit.FormatHHmm(d.EstimatedMinutes)},
		textSpec{Key: domain.FieldMarginPct, Label: "Margin", Placeholder: "60%", Prefill: unit.FormatPercent(d.MarginPct)},
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
