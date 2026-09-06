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
		choiceSpec{Key: domain.FieldDesignID, Label: "Design", Choices: choicesOf(designs, designChoice)},
		textSpec{Key: domain.FieldQuantity, Label: "Copies", Placeholder: "1", Prefill: unit.FormatCopies(draft.Quantity)},
		textSpec{Key: domain.FieldPrintDate, Label: dateLabel, Placeholder: unit.DateLayout, Prefill: unit.FormatDate(unit.Today())},
		textSpec{Key: domain.FieldMinutes, Label: "Actual time", Placeholder: "5:30", Prefill: unit.FormatHHmm(draft.Minutes)},
	}
	specs = append(specs, stockRowSpecs("", "", "")...)
	usage := usageRowSpecs(0, choicesOf(spools, spoolChoice), unit.FormatGrams(draft.Grams), choice{})
	return newForm("Record print", append(specs, usage...))
}

// newEditPrintForm opens on what the Print says, and reports how many Filament
// Usage rows it opened with, since the tab owns how many rows there are.
func newEditPrintForm(designs []app.DesignView, spools []app.SpoolView, print app.PrintView) (*form, int) {
	designChoices, spoolChoices := choicesOf(designs, designChoice), choicesOf(spools, spoolChoice)
	specs := []fieldSpec{
		choiceSpec{Key: domain.FieldDesignID, Label: "Design", Choices: designChoices, Selected: choiceFor(designChoices, print.DesignID)},
		textSpec{Key: domain.FieldQuantity, Label: "Copies", Prefill: unit.FormatCopies(print.Quantity)},
		textSpec{Key: domain.FieldPrintDate, Label: dateLabel, Placeholder: unit.DateLayout, Prefill: unit.FormatDate(print.Date)},
		textSpec{Key: domain.FieldMinutes, Label: "Actual time", Placeholder: "5:30", Prefill: unit.FormatHHmm(print.Minutes)},
	}
	specs = append(specs, stockRowSpecs(
		unit.FormatCopies(print.GiftedCount),
		unit.FormatCopies(print.KeptCount),
		unit.FormatCopies(print.ScrappedCount))...)

	for row, usage := range print.Usages {
		specs = append(specs, usageRowSpecs(row, spoolChoices,
			unit.FormatGrams(usage.Grams), choiceFor(spoolChoices, usage.SpoolID))...)
	}
	return newForm("Edit print", specs), len(print.Usages)
}

// stockRowSpecs is what became of the copies. Blank counts as 0, so a plate
// that all came off well needs no typing.
func stockRowSpecs(gifted, kept, scrapped string) []fieldSpec {
	return []fieldSpec{
		textSpec{Key: domain.FieldGifted, Label: "Gifted copies", Placeholder: "0", Prefill: gifted},
		textSpec{Key: domain.FieldKept, Label: "Kept copies", Placeholder: "0", Prefill: kept},
		textSpec{Key: domain.FieldScrapped, Label: "Scrapped copies", Placeholder: "0", Prefill: scrapped},
	}
}

// usageRowSpecs is one Filament Usage row. A swap row carries neither prefill
// nor placeholder: how a job split across two Spools is not guessable from the
// estimate, and it opens on the Spool the row above it names.
func usageRowSpecs(row int, spools []choice, grams string, spool choice) []fieldSpec {
	gramsLabel, spoolLabel, placeholder := "Actual grams", "Spool", "120"
	if row > 0 {
		gramsLabel = fmt.Sprintf("Grams %d", row+1)
		spoolLabel = fmt.Sprintf("Spool %d", row+1)
		placeholder = ""
	}
	return []fieldSpec{
		textSpec{Key: domain.FieldUsageGrams(row), Label: gramsLabel, Placeholder: placeholder, Prefill: grams},
		choiceSpec{Key: domain.FieldUsageSpool(row), Label: spoolLabel, Choices: spools, Selected: spool},
	}
}

func recordPrintCmd(f *form, usageRows int) app.RecordPrintCmd {
	usages := make([]app.FilamentUsageCmd, 0, usageRows)
	for row := range usageRows {
		usages = append(usages, app.FilamentUsageCmd{
			SpoolID: f.Choice(domain.FieldUsageSpool(row)).ID,
			Grams:   f.Value(domain.FieldUsageGrams(row)),
		})
	}

	return app.RecordPrintCmd{
		DesignID: f.Choice(domain.FieldDesignID).ID,
		Date:     f.Value(domain.FieldPrintDate),
		Quantity: f.Value(domain.FieldQuantity),
		Minutes:  f.Value(domain.FieldMinutes),
		Gifted:   f.Value(domain.FieldGifted),
		Kept:     f.Value(domain.FieldKept),
		Scrapped: f.Value(domain.FieldScrapped),
		Usages:   usages,
	}
}

func editPrintCmd(f *form, printID int64, usageRows int) app.EditPrintCmd {
	return app.EditPrintCmd{
		PrintID:        printID,
		RecordPrintCmd: recordPrintCmd(f, usageRows),
	}
}

func designChoice(design app.DesignView) choice {
	return choice{Label: design.Name, ID: design.ID}
}

func spoolChoice(spool app.SpoolView) choice {
	label := fmt.Sprintf("%s %s %s %s", spool.FilamentType, spool.Brand, spool.Color, unit.FormatGrams(spool.RemainingGrams))
	return choice{Label: label, ID: spool.ID}
}
