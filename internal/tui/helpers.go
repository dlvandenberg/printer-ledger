package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const (
	materialColorLimit = 2
	materialWidth      = 16
)

func money(c unit.Cents) string { return currency + unit.FormatCents(c) }

// amount pads before styling: a width verb counts the escape bytes of an
// already-styled string, so a bold figure would sit short of its column.
func amount(s string) string { return fmt.Sprintf("%12s", s) }

// material names a Print by what it was printed in, which is what tells two
// Prints of one Design apart. A swap of more than two spools would push the row
// off the screen, so the third color and beyond never appear; truncate then
// marks the label short.
func material(m app.MaterialView) string {
	shown := m.Colors
	if len(shown) > materialColorLimit {
		shown = slices.Clone(shown[:materialColorLimit])
	}
	return string(m.FilamentType) + " " + strings.Join(shown, "+")
}

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
