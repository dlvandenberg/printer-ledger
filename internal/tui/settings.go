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

var _ tabModel = settingsModel{}

type settingsModel struct {
	app     *app.App
	view    app.SettingsView
	form    *form
	loadErr error
}

func newSettingsModel(a *app.App) (settingsModel, error) {
	m := settingsModel{app: a}
	if err := m.reload(); err != nil {
		return m, err
	}
	return m, nil
}

func (m *settingsModel) reload() error {
	view, err := m.app.Settings(context.Background())
	if err != nil {
		return err
	}
	m.view = view
	return nil
}

func (m settingsModel) CapturesInput() bool { return m.form != nil }

func (m settingsModel) Update(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	if m.form != nil {
		return m.updateForm(msg)
	}
	if msg.String() == "e" {
		m.form = newSettingsForm(m.view)
	}
	return m, nil
}

func (m settingsModel) updateForm(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.form = nil
		return m, nil
	case "enter":
		if _, err := m.app.UpdateSettings(context.Background(), updateSettingsCmd(m.form)); err != nil {
			var v *domain.ValidationError
			if errors.As(err, &v) {
				m.form.SetErrors(v)
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

	return m, m.form.Update(msg)
}

func (m settingsModel) View() string {
	if m.form != nil {
		return m.form.View()
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Settings"))
	b.WriteString("\n\n")

	if m.loadErr != nil {
		b.WriteString(errorStyle.Render(m.loadErr.Error()))
		b.WriteString("\n\n")
	}

	money := func(c domain.Cents) string { return m.view.Currency + domain.FormatCents(c) }
	rows := [][2]string{
		{"Electricity per kWh", money(m.view.KwhPrice)},
		{"Machine per hour", money(m.view.MachineHourlyRate)},
		{"Printer purchase cost", money(m.view.PrinterPurchaseCost)},
		{"Default margin", domain.FormatPercent(m.view.DefaultMargin)},
		{"Minimum margin", domain.FormatPercent(m.view.MinMargin)},
		{"Display currency", m.view.Currency},
	}
	for _, row := range rows {
		fmt.Fprintf(&b, "  %-*s%s\n", labelWidth, row[0], row[1])
	}

	b.WriteString("\n")
	fmt.Fprintf(&b, "  %-6s  %8s  %-8s\n", "TYPE", "KWH/H", "SOURCE")
	for _, rate := range m.view.PowerRates {
		fmt.Fprintf(&b, "  %-6s  %8s  %-8s\n",
			rate.FilamentType, domain.FormatKwhPerHour(rate.KwhPerHour), rate.Source)
	}
	return b.String()
}

func (m settingsModel) Help() string {
	if m.form != nil {
		return m.form.Help()
	}
	return "e edit · " + globalHelp
}
