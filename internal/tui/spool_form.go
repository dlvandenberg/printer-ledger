package tui

import (
	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
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
			Placeholder: domain.DateLayout,
			Prefill:     domain.FormatDate(domain.Today()),
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
