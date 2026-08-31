package tui

import (
	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func newSpoolForm() *form {
	return newForm("Add spool", []fieldSpec{
		{Key: domain.FieldFilamentType, Label: "Filament type", Choices: domain.FilamentTypeNames()},
		{Key: domain.FieldBrand, Label: "Brand", Placeholder: "Bambu"},
		{Key: domain.FieldColor, Label: "Colour", Placeholder: "Black"},
		{Key: domain.FieldInitialGrams, Label: "Filament grams", Placeholder: "1000"},
		{Key: domain.FieldTareGrams, Label: "Empty spool grams", Placeholder: "210"},
		{Key: domain.FieldPurchaseCost, Label: "Purchase price in EUR", Placeholder: "22.00"},
		{
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
		{Key: domain.FieldMeasuredGrams, Label: "Weight on the scale", Placeholder: "610"},
		{
			Key:         domain.FieldAdjustedOn,
			Label:       "Date",
			Placeholder: unit.DateLayout,
			Prefill:     unit.FormatDate(unit.Today()),
		},
		{Key: domain.FieldNote, Label: "Note", Placeholder: "purge tower"},
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
