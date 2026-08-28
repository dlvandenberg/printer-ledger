package tui

import (
	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

func newSpoolForm() *form {
	return newForm("Add spool", []fieldSpec{
		{key: domain.FieldFilamentType, label: "Filament type", choices: domain.FilamentTypeNames()},
		{key: domain.FieldBrand, label: "Brand", placeholder: "Bambu"},
		{key: domain.FieldColor, label: "Colour", placeholder: "Black"},
		{key: domain.FieldInitialGrams, label: "Filament grams", placeholder: "1000"},
		{key: domain.FieldTareGrams, label: "Empty spool grams", placeholder: "210"},
		{key: domain.FieldPurchaseCost, label: "Purchase price in EUR", placeholder: "22.00"},
		{
			key:         domain.FieldPurchaseDate,
			label:       "Purchase date",
			placeholder: domain.DateLayout,
			value:       domain.FormatDate(domain.Today()),
		},
	})
}

func addSpoolCmd(f *form) app.AddSpoolCmd {
	return app.AddSpoolCmd{
		FilamentType: f.value(domain.FieldFilamentType),
		Brand:        f.value(domain.FieldBrand),
		Color:        f.value(domain.FieldColor),
		InitialGrams: f.value(domain.FieldInitialGrams),
		TareGrams:    f.value(domain.FieldTareGrams),
		PurchaseCost: f.value(domain.FieldPurchaseCost),
		PurchaseDate: f.value(domain.FieldPurchaseDate),
	}
}
