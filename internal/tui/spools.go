package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

var _ tabModel = spoolsModel{}

type spoolsModel struct {
	app     *app.App
	rows    []app.SpoolView
	cursor  int
	form    *form
	loadErr error
}

func newSpoolsModel(a *app.App) (spoolsModel, error) {
	m := spoolsModel{app: a}
	if err := m.reload(); err != nil {
		return m, err
	}
	return m, nil
}

func (m *spoolsModel) reload() error {
	rows, err := m.app.ListSpools(context.Background())
	if err != nil {
		return err
	}
	m.rows = rows
	if m.cursor >= len(rows) {
		m.cursor = max(0, len(rows)-1)
	}
	return nil
}

func (m spoolsModel) capturesInput() bool { return m.form != nil }

func (m spoolsModel) update(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	if m.form != nil {
		return m.updateForm(msg)
	}

	switch msg.String() {
	case "a":
		m.form = newSpoolForm()
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.rows)-1 {
			m.cursor++
		}
	}
	return m, nil
}

func (m spoolsModel) updateForm(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.form = nil
		return m, nil
	case "enter":
		if _, err := m.app.AddSpool(context.Background(), addSpoolCmd(m.form)); err != nil {
			var v *domain.ValidationError
			if errors.As(err, &v) {
				m.form.setErrors(v)
				return m, nil
			}
			m.loadErr = err
			return m, nil
		}
		m.form = nil
		if err := m.reload(); err != nil {
			m.loadErr = err
		}
		return m, nil
	}

	return m, m.form.update(msg)
}

func (m spoolsModel) view() string {
	if m.form != nil {
		return m.form.view()
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Spools"))
	b.WriteString("\n\n")

	if m.loadErr != nil {
		b.WriteString(errorStyle.Render(m.loadErr.Error()))
		b.WriteString("\n\n")
	}

	if len(m.rows) == 0 {
		b.WriteString(placeholderStyle.Render("No spools yet. Press a to add one."))
		return b.String()
	}

	fmt.Fprintf(&b, "  %-6s  %-14s  %-12s  %10s  %10s  %-7s\n",
		"TYPE", "BRAND", "COLOUR", "REMAINING", "VALUE", "STATE")
	for i, row := range m.rows {
		marker := "  "
		if i == m.cursor {
			marker = "> "
		}
		fmt.Fprintf(&b, "%s%-6s  %-14s  %-12s  %10s  %10s  %-7s\n",
			marker, row.FilamentType, truncate(row.Brand, 14), truncate(row.Color, 12),
			domain.FormatGrams(row.RemainingGrams),
			currency+domain.FormatCents(row.RemainingValue),
			row.State)
	}
	return b.String()
}

func (m spoolsModel) help() string {
	if m.form != nil {
		return m.form.help()
	}
	return "a add · up/down move · " + globalHelp
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
