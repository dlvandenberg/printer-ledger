package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const (
	materialColorLimit = 2
	materialWidth      = 16
	amountWidth        = 12
	detailValueWidth   = 10
	detailLineFormat   = "  %-12s %s\n"
)

func money(c unit.Cents) string { return currency + unit.FormatCents(c) }

// amount pads before styling: a width verb counts the escape bytes of an
// already-styled string, so a bold figure would sit short of its column.
func amount(s string) string { return column(s, amountWidth) }

// detailLine is one labelled figure of a detail screen, so the label column and
// the figure column are set in one place.
func detailLine(b *strings.Builder, label, value string) {
	fmt.Fprintf(b, detailLineFormat, label, column(value, detailValueWidth))
}

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

// pad fills a styled string out to a column. A width verb would count the
// escape bytes instead and leave the string as it found it.
func pad(s string, width int) string {
	if gap := width - lipgloss.Width(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}

// column right-aligns on display width: a width verb counts the bytes of the
// € prefix instead and leaves the figure two columns short.
func column(s string, width int) string {
	if gap := width - lipgloss.Width(s); gap > 0 {
		return strings.Repeat(" ", gap) + s
	}
	return s
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

func wrap(i, delta, n int) int { return ((i+delta)%n + n) % n }
