package tui

import (
	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func newSpoolForm() *form {
	return newForm("Add spool", []fieldSpec{
		choiceSpec{Key: domain.FieldFilamentType, Label: "Filament type", Choices: valueChoices(domain.FilamentTypeNames())},
		textSpec{Key: domain.FieldBrand, Label: "Brand", Placeholder: "Bambu"},
		textSpec{Key: domain.FieldColor, Label: "Colour", Placeholder: "Black"},
		textSpec{Key: domain.FieldInitialGrams, Label: "Filament grams", Placeholder: "1000"},
		textSpec{Key: domain.FieldTareGrams, Label: "Empty spool grams", Placeholder: "210"},
		textSpec{Key: domain.FieldPurchaseCost, Label: "Purchase price in EUR", Placeholder: "22.00"},
		textSpec{
			Key:         domain.FieldPurchaseDate,
			Label:       "Purchase date",
			Placeholder: unit.DateLayout,
			Prefill:     unit.FormatDate(unit.Today()),
		},
	})
}

func addSpoolCmd(f *form) app.AddSpoolCmd {
	return app.AddSpoolCmd{
		FilamentType: f.Value(domain.FieldFilamentType),
		Brand:        f.Value(domain.FieldBrand),
		Color:        f.Value(domain.FieldColor),
		InitialGrams: f.Value(domain.FieldInitialGrams),
		TareGrams:    f.Value(domain.FieldTareGrams),
		PurchaseCost: f.Value(domain.FieldPurchaseCost),
		PurchaseDate: f.Value(domain.FieldPurchaseDate),
	}
}

func newReweighForm() *form {
	return newForm("Re-weigh spool", []fieldSpec{
		textSpec{Key: domain.FieldMeasuredGrams, Label: "Weight on the scale", Placeholder: "610"},
		textSpec{
			Key:         domain.FieldAdjustedOn,
			Label:       dateLabel,
			Placeholder: unit.DateLayout,
			Prefill:     unit.FormatDate(unit.Today()),
		},
		textSpec{Key: domain.FieldNote, Label: "Note", Placeholder: "purge tower"},
	})
}

func reweighSpoolCmd(f *form, spoolID int64) app.ReweighSpoolCmd {
	return app.ReweighSpoolCmd{
		SpoolID:       spoolID,
		MeasuredGrams: f.Value(domain.FieldMeasuredGrams),
		AdjustedOn:    f.Value(domain.FieldAdjustedOn),
		Note:          f.Value(domain.FieldNote),
	}
}
