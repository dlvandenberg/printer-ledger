package tui

import (
	"fmt"
	"strings"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func designQuoteView(q app.DesignQuoteView) string {
	design := q.Design

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", titleStyle.Render(design.Name))
	fmt.Fprintf(&b, "  %-12s %10s  %s  %s  %s\n", "Per copy", unit.FormatGrams(design.EstimatedGrams),
		unit.FormatMinutes(design.EstimatedMinutes), design.DefaultFilamentType,
		unit.FormatPercent(design.MarginPct)+" margin")

	if q.HasSuggestedPrice {
		note := ""
		if q.SuggestedPriceReference {
			note = "(reference — no spool in stock can print this)"
		}
		fmt.Fprintf(&b, "  %-12s %10s  %s\n", "Suggested",
			titleStyle.Render(currency+unit.FormatCents(q.SuggestedPrice)), note)
	}

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Estimated cost"))
	b.WriteString("\n")
	if len(q.Costs) == 0 {
		b.WriteString(placeholderStyle.Render("No spools bought yet, so nothing to price against."))
		return b.String()
	}

	fmt.Fprintf(&b, "  %-6s  %10s  %10s  %10s  %10s  %s\n",
		"TYPE", "FILAMENT", "ENERGY", "OVERHEAD", "TOTAL", "PRICED AGAINST")
	for _, cost := range q.Costs {
		marker := "  "
		if cost.IsDefault {
			marker = "> "
		}
		spool := truncate(fmt.Sprintf("%s %s", cost.SpoolBrand, cost.SpoolColor), 24)
		if cost.Reference {
			spool += " (reference)"
		}
		fmt.Fprintf(&b, "%s%-6s  %10s  %10s  %10s  %10s  %s\n",
			marker, cost.FilamentType,
			currency+unit.FormatCents(cost.Filament),
			currency+unit.FormatCents(cost.Energy),
			currency+unit.FormatCents(cost.Overhead),
			currency+unit.FormatCents(cost.Total),
			spool)
	}
	return b.String()
}
