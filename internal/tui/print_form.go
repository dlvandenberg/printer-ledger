package tui

import (
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func newPrintForm(designs []app.DesignView, spools []app.SpoolView, draft app.PrintDraftView) *form {
	return newForm("Record print", []fieldSpec{
		{Key: domain.FieldDesignID, Label: "Design", Choices: designNames(designs)},
		{Key: domain.FieldQuantity, Label: "Copies", Placeholder: "1", Prefill: unit.FormatCopies(draft.Quantity)},
		{Key: domain.FieldPrintDate, Label: "Date", Placeholder: unit.DateLayout, Prefill: unit.FormatDate(unit.Today())},
		{Key: domain.FieldMinutes, Label: "Actual time", Placeholder: "5:30", Prefill: unit.FormatHHmm(draft.Minutes)},
		{Key: domain.FieldUsageGrams(0), Label: "Actual grams", Placeholder: "120", Prefill: unit.FormatGrams(draft.Grams)},
		{Key: domain.FieldUsageSpool(0), Label: "Spool", Choices: spoolLabels(spools)},
	})
}

func recordPrintCmd(f *form, designs []app.DesignView, spools []app.SpoolView) app.RecordPrintCmd {
	return app.RecordPrintCmd{
		DesignID: designIDAt(designs, f.ChoiceIndex(domain.FieldDesignID)),
		Date:     f.Value(domain.FieldPrintDate),
		Quantity: f.Value(domain.FieldQuantity),
		Minutes:  f.Value(domain.FieldMinutes),
		Usages: []app.FilamentUsageCmd{{
			SpoolID: spoolIDAt(spools, f.ChoiceIndex(domain.FieldUsageSpool(0))),
			Grams:   f.Value(domain.FieldUsageGrams(0)),
		}},
	}
}

func designNames(designs []app.DesignView) []string {
	names := make([]string, 0, len(designs))
	for _, design := range designs {
		names = append(names, design.Name)
	}
	return names
}

func spoolLabels(spools []app.SpoolView) []string {
	labels := make([]string, 0, len(spools))
	for _, spool := range spools {
		labels = append(labels, fmt.Sprintf("%s %s %s", spool.FilamentType, spool.Brand, spool.Color))
	}
	return labels
}

func designIDAt(designs []app.DesignView, index int) int64 {
	if index < 0 || index >= len(designs) {
		return 0
	}
	return designs[index].ID
}

func spoolIDAt(spools []app.SpoolView, index int) int64 {
	if index < 0 || index >= len(spools) {
		return 0
	}
	return spools[index].ID
}
