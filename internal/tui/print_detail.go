package tui

import (
	"fmt"
	"strings"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const printUsageRow = "  %-6s  %-10s  %-16s  %10s  %10s\n" // Type, Brand, Colour, Grams, c/g

// printDetailView shows the rate snapshot, which is the only place those three
// numbers are visible: they are what makes a past print's frozen cost checkable.
func printDetailView(p app.PrintView) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", titleStyle.Render(fmt.Sprintf("%s  %s", p.DesignName, unit.FormatDate(p.Date))))
	detailLine(&b, "copies", unit.FormatCopies(p.Quantity))
	detailLine(&b, "time", unit.FormatMinutes(p.Minutes))

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Stock"))
	b.WriteString("\n")
	detailLine(&b, "quantity", unit.FormatCopies(p.Quantity))
	detailLine(&b, "sold", unit.FormatCopies(p.SoldCount))
	detailLine(&b, "gifted", unit.FormatCopies(p.GiftedCount))
	detailLine(&b, "kept", unit.FormatCopies(p.KeptCount))
	detailLine(&b, "scrapped", unit.FormatCopies(p.ScrappedCount))
	detailLine(&b, "available", unit.FormatCopies(p.AvailableCopies))

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Cost"))
	b.WriteString("\n")
	b.WriteString(costLines(p.Cost))

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Filament Usage"))
	b.WriteString("\n")
	fmt.Fprintf(&b, printUsageRow, "TYPE", "BRAND", "COLOUR", "GRAMS", "C/G")
	for _, u := range p.Usages {
		fmt.Fprintf(&b, printUsageRow,
			u.FilamentType,
			truncate(u.SpoolBrand, brandLength),
			truncate(u.SpoolColor, colorLength),
			unit.FormatGrams(u.Grams),
			unit.FormatCentsPerGram(u.CostPerGram))
	}

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Rates when recorded"))
	b.WriteString("\n")
	detailLine(&b, "per kwh", money(p.KwhPrice))
	detailLine(&b, "kwh per hour", unit.FormatKwhPerHour(p.KwhPerHour))
	detailLine(&b, "machine per hour", money(p.MachineRate))
	return b.String()
}
