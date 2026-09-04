package tui

import (
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const usageFieldsPerRow = 2

func newPrintForm(designs []app.DesignView, spools []app.SpoolView, draft app.PrintDraftView) *form {
	specs := []fieldSpec{
		{Key: domain.FieldDesignID, Label: "Design", Choices: designNames(designs)},
		{Key: domain.FieldQuantity, Label: "Copies", Placeholder: "1", Prefill: unit.FormatCopies(draft.Quantity)},
		{Key: domain.FieldPrintDate, Label: "Date", Placeholder: unit.DateLayout, Prefill: unit.FormatDate(unit.Today())},
		{Key: domain.FieldMinutes, Label: "Actual time", Placeholder: "5:30", Prefill: unit.FormatHHmm(draft.Minutes)},
	}
	return newForm("Record print", append(specs, usageRowSpecs(0, spools, unit.FormatGrams(draft.Grams), "")...))
}

// usageRowSpecs is one Filament Usage row. A swap row carries neither prefill
// nor placeholder: how a job split across two Spools is not guessable from the
// estimate, and it opens on the Spool the row above it names.
func usageRowSpecs(row int, spools []app.SpoolView, grams, spool string) []fieldSpec {
	gramsLabel, spoolLabel, placeholder := "Actual grams", "Spool", "120"
	if row > 0 {
		gramsLabel = fmt.Sprintf("Grams %d", row+1)
		spoolLabel = fmt.Sprintf("Spool %d", row+1)
		placeholder = ""
	}
	return []fieldSpec{
		{Key: domain.FieldUsageGrams(row), Label: gramsLabel, Placeholder: placeholder, Prefill: grams},
		{Key: domain.FieldUsageSpool(row), Label: spoolLabel, Choices: spoolLabels(spools), Choice: spool},
	}
}

func recordPrintCmd(f *form, designs []app.DesignView, spools []app.SpoolView, usageRows int) app.RecordPrintCmd {
	usages := make([]app.FilamentUsageCmd, 0, usageRows)
	for row := range usageRows {
		usages = append(usages, app.FilamentUsageCmd{
			SpoolID: spoolIDAt(spools, f.ChoiceIndex(domain.FieldUsageSpool(row))),
			Grams:   f.Value(domain.FieldUsageGrams(row)),
		})
	}

	return app.RecordPrintCmd{
		DesignID: designIDAt(designs, f.ChoiceIndex(domain.FieldDesignID)),
		Date:     f.Value(domain.FieldPrintDate),
		Quantity: f.Value(domain.FieldQuantity),
		Minutes:  f.Value(domain.FieldMinutes),
		Usages:   usages,
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
		labels = append(labels, fmt.Sprintf("%s %s %s %s", spool.FilamentType, spool.Brand, spool.Color, unit.FormatGrams(spool.RemainingGrams)))
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
