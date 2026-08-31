package tui

import (
	"fmt"
	"strings"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func spoolDetailView(d app.SpoolDetailView) string {
	spool := d.Spool

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", titleStyle.Render(fmt.Sprintf("%s %s %s", spool.FilamentType, spool.Brand, spool.Color)))
	fmt.Fprintf(&b, "  %-12s %10s\n", "Initial", unit.FormatGrams(spool.InitialGrams))
	fmt.Fprintf(&b, "  %-12s %10s\n", "Printed", unit.FormatGrams(-spool.UsedGrams))
	fmt.Fprintf(&b, "  %-12s %10s\n", "Adjusted", unit.FormatGrams(spool.AdjustedGrams))
	fmt.Fprintf(&b, "  %-12s %10s  %s  %s\n", "Remaining", unit.FormatGrams(spool.RemainingGrams),
		currency+unit.FormatCents(spool.RemainingValue), spool.State)

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Adjustments"))
	b.WriteString("\n")
	if len(d.Adjustments) == 0 {
		b.WriteString(placeholderStyle.Render(fmt.Sprintf("None yet. Press %s to re-weigh.", KeyR)))
		return b.String()
	}

	fmt.Fprintf(&b, "  %-10s  %10s  %10s  %8s  %s\n", "DATE", "MEASURED", "REMAINING", "DELTA", "NOTE")
	for _, a := range d.Adjustments {
		fmt.Fprintf(&b, "  %-10s  %10s  %10s  %8s  %s\n",
			unit.FormatDate(a.AdjustedOn), unit.FormatGrams(a.MeasuredGrams),
			unit.FormatGrams(a.DerivedRemaining), unit.FormatGrams(a.DeltaGrams), a.Note)
	}
	return b.String()
}
