package tui

import (
	"fmt"
	"strings"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const (
	printDetailWidth = 10
	printDetailLine  = "  %-12s %s\n"
	printUsageRow    = "  %-6s  %-10s  %-16s  %10s  %10s\n" // Type, Brand, Colour, Grams, c/g
)

// printDetailView shows the rate snapshot, which is the only place those three
// numbers are visible: they are what makes a past print's frozen cost checkable.
func printDetailView(p app.PrintView) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", titleStyle.Render(fmt.Sprintf("%s  %s", p.DesignName, unit.FormatDate(p.Date))))
	fmt.Fprintf(&b, printDetailLine, "Copies", column(unit.FormatCopies(p.Quantity), printDetailWidth))
	fmt.Fprintf(&b, printDetailLine, "Time", column(unit.FormatMinutes(p.Minutes), printDetailWidth))
	fmt.Fprintf(&b, printDetailLine, "Filament", column(unit.FormatGrams(p.UsedGrams), printDetailWidth))

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Stock"))
	b.WriteString("\n")
	fmt.Fprintf(&b, printDetailLine, "quantity", column(unit.FormatCopies(p.Quantity), printDetailWidth))
	fmt.Fprintf(&b, printDetailLine, "sold", column(unit.FormatCopies(p.SoldCount), printDetailWidth))
	fmt.Fprintf(&b, printDetailLine, "gifted", column(unit.FormatCopies(p.GiftedCount), printDetailWidth))
	fmt.Fprintf(&b, printDetailLine, "kept", column(unit.FormatCopies(p.KeptCount), printDetailWidth))
	fmt.Fprintf(&b, printDetailLine, "scrapped", column(unit.FormatCopies(p.ScrappedCount), printDetailWidth))
	fmt.Fprintf(&b, printDetailLine, "available", column(unit.FormatCopies(p.AvailableCopies), printDetailWidth))

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
	fmt.Fprintf(&b, printDetailLine, "kwh price", column(money(p.KwhPrice), printDetailWidth))
	fmt.Fprintf(&b, printDetailLine, "kwh per hour", column(unit.FormatKwhPerHour(p.KwhPerHour), printDetailWidth))
	fmt.Fprintf(&b, printDetailLine, "machine rate", column(money(p.MachineRate), printDetailWidth))
	return b.String()
}
