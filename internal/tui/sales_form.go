package tui

import (
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

// newSaleForm asks for what was actually charged, not what was suggested: the
// price row opens empty, and the suggested price is shown beside the cost
// rather than prefilled (ADR-0016).
func newSaleForm(prints []app.SellablePrintView) *form {
	return newForm("Record sale", []fieldSpec{
		choiceSpec{Key: domain.FieldSalePrint, Label: "Print", Choices: choicesOf(prints, printChoice)},
		textSpec{Key: domain.FieldSalePrice, Label: "Price", Placeholder: "4.00"},
		textSpec{Key: domain.FieldSaleDate, Label: dateLabel, Placeholder: unit.DateLayout, Prefill: unit.FormatDate(unit.Today())},
	})
}

func newEditSaleForm(prints []app.SellablePrintView, sale app.SaleView) *form {
	choices := choicesOf(prints, printChoice)
	return newForm("Edit sale", []fieldSpec{
		choiceSpec{Key: domain.FieldSalePrint, Label: "Print", Choices: choices, Selected: choiceFor(choices, sale.PrintID)},
		textSpec{Key: domain.FieldSalePrice, Label: "Price", Prefill: unit.FormatCents(sale.Price)},
		textSpec{Key: domain.FieldSaleDate, Label: dateLabel, Placeholder: unit.DateLayout, Prefill: unit.FormatDate(sale.Date)},
	})
}

func recordSaleCmd(f *form) app.RecordSaleCmd {
	return app.RecordSaleCmd{
		PrintID: f.Choice(domain.FieldSalePrint).ID,
		Price:   f.Value(domain.FieldSalePrice),
		Date:    f.Value(domain.FieldSaleDate),
	}
}

func editSaleCmd(f *form, saleID int64) app.EditSaleCmd {
	return app.EditSaleCmd{
		SaleID:        saleID,
		RecordSaleCmd: recordSaleCmd(f),
	}
}

// printChoice leads with what the copies are — the Design and the material —
// and follows with the facts about them, so two Prints of one Design in
// different filament are told apart at the point the Sale names one. Prints
// alike in all of these are interchangeable: nothing else distinguishes them.
func printChoice(print app.SellablePrintView) choice {
	label := fmt.Sprintf("%s %s %s %s %s left",
		truncate(print.DesignName, 16),
		truncate(material(print.Material), materialWidth),
		unit.FormatDate(print.Date),
		money(print.CostPerCopy),
		unit.FormatCopies(print.AvailableCopies))
	return choice{Label: label, ID: print.PrintID}
}
