package tui

import (
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

func money(c unit.Cents) string { return currency + unit.FormatCents(c) }

// amount pads before styling: a width verb counts the escape bytes of an
// already-styled string, so a bold figure would sit short of its column.
func amount(s string) string { return fmt.Sprintf("%12s", s) }

func truncate(s string, width int) string {
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return string(r[:width])
	}
	return string(r[:width-1]) + "…"
}

func dropLastRune(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	return string(r[:len(r)-1])
}
