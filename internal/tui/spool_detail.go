package tui

import (
	"fmt"
	"strings"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

func spoolDetailView(d app.SpoolDetailView) string {
	spool := d.Spool

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", titleStyle.Render(fmt.Sprintf("%s %s %s", spool.FilamentType, spool.Brand, spool.Color)))
	fmt.Fprintf(&b, "  %-12s %10s\n", "Initial", domain.FormatGrams(spool.InitialGrams))
	fmt.Fprintf(&b, "  %-12s %10s\n", "Printed", domain.FormatGrams(-spool.UsedGrams))
	fmt.Fprintf(&b, "  %-12s %10s\n", "Adjusted", domain.FormatGrams(spool.AdjustedGrams))
	fmt.Fprintf(&b, "  %-12s %10s  %s  %s\n", "Remaining", domain.FormatGrams(spool.RemainingGrams),
		currency+domain.FormatCents(spool.RemainingValue), spool.State)

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Adjustments"))
	b.WriteString("\n")
	if len(d.Adjustments) == 0 {
		b.WriteString(placeholderStyle.Render("None yet. Press r to re-weigh."))
		return b.String()
	}

	fmt.Fprintf(&b, "  %-10s  %10s  %10s  %8s  %s\n", "DATE", "MEASURED", "REMAINING", "DELTA", "NOTE")
	for _, a := range d.Adjustments {
		fmt.Fprintf(&b, "  %-10s  %10s  %10s  %8s  %s\n",
			domain.FormatDate(a.AdjustedOn), domain.FormatGrams(a.MeasuredGrams),
			domain.FormatGrams(a.DerivedRemaining), domain.FormatGrams(a.DeltaGrams), a.Note)
	}
	return b.String()
}
