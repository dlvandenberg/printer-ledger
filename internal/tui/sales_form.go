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
		choiceSpec{Key: domain.FieldSalePrint, Label: "Print", Choices: printLabels(prints)},
		textSpec{Key: domain.FieldSalePrice, Label: "Price", Placeholder: "4.00"},
		textSpec{Key: domain.FieldSaleDate, Label: dateLabel, Placeholder: unit.DateLayout, Prefill: unit.FormatDate(unit.Today())},
	})
}

func newEditSaleForm(prints []app.SellablePrintView, sale app.SaleView) *form {
	return newForm("Edit sale", []fieldSpec{
		choiceSpec{Key: domain.FieldSalePrint, Label: "Print", Choices: printLabels(prints), Choice: printLabelOf(prints, sale.PrintID)},
		textSpec{Key: domain.FieldSalePrice, Label: "Price", Prefill: unit.FormatCents(sale.Price)},
		textSpec{Key: domain.FieldSaleDate, Label: dateLabel, Placeholder: unit.DateLayout, Prefill: unit.FormatDate(sale.Date)},
	})
}

func recordSaleCmd(f *form, prints []app.SellablePrintView) app.RecordSaleCmd {
	return app.RecordSaleCmd{
		PrintID: printIDAt(prints, f.ChoiceIndex(domain.FieldSalePrint)),
		Price:   f.Value(domain.FieldSalePrice),
		Date:    f.Value(domain.FieldSaleDate),
	}
}

func editSaleCmd(f *form, saleID int64, prints []app.SellablePrintView) app.EditSaleCmd {
	return app.EditSaleCmd{
		SaleID:        saleID,
		RecordSaleCmd: recordSaleCmd(f, prints),
	}
}

func printLabels(prints []app.SellablePrintView) []string {
	labels := make([]string, 0, len(prints))
	for _, print := range prints {
		labels = append(labels, printLabel(print))
	}
	return labels
}

// printLabel leads with what the copies are — the Design and the material —
// and follows with the facts about them, so two Prints of one Design in
// different filament are told apart at the point the Sale names one. Prints
// alike in all of these are interchangeable: nothing else distinguishes them.
func printLabel(print app.SellablePrintView) string {
	return fmt.Sprintf("%s %s %s %s %s left",
		truncate(print.DesignName, 16),
		truncate(material(print.Material), materialWidth),
		unit.FormatDate(print.Date),
		money(print.CostPerCopy),
		unit.FormatCopies(print.AvailableCopies))
}

func printLabelOf(prints []app.SellablePrintView, id int64) string {
	for _, print := range prints {
		if print.PrintID == id {
			return printLabel(print)
		}
	}
	return ""
}

func printIDAt(prints []app.SellablePrintView, index int) int64 {
	if index < 0 || index >= len(prints) {
		return 0
	}
	return prints[index].PrintID
}
