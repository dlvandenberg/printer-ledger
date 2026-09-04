package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

var _ tabModel = settingsModel{}

type settingsModel struct {
	app     *app.App
	view    app.SettingsView
	form    *form
	lastErr error
}

// A failed read leaves the tab open carrying the error: the operator has to be
// able to reach Settings on a ledger that cannot answer for itself.
func newSettingsModel(a *app.App) settingsModel {
	m := settingsModel{app: a}
	m.lastErr = m.reload()
	return m
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
	if msg.String() == KeyE {
		m.form = newSettingsForm(m.view)
	}
	return m, nil
}

func (m settingsModel) updateForm(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	switch msg.String() {
	case KeyEsc:
		m.form = nil
		return m, nil
	case KeyEnter:
		if _, err := m.app.UpdateSettings(context.Background(), updateSettingsCmd(m.form)); err != nil {
			var v *domain.ValidationError
			if errors.As(err, &v) {
				m.form.SetErrors(v)
				return m, nil
			}
			m.lastErr = err
			return m, nil
		}
		m.form = nil
		m.lastErr = m.reload()
		return m, nil
	}

	return m, m.form.Update(msg)
}

func (m settingsModel) View() string {
	var b strings.Builder
	if m.lastErr != nil {
		b.WriteString(errorStyle.Render(m.lastErr.Error()))
		b.WriteString("\n\n")
	}
	if m.form != nil {
		b.WriteString(m.form.View())
		return b.String()
	}

	b.WriteString(titleStyle.Render("Settings"))
	b.WriteString("\n\n")

	money := func(c unit.Cents) string { return currency + unit.FormatCents(c) }
	rows := [][2]string{
		{"Electricity per kWh", money(m.view.KwhPrice)},
		{"Machine per hour", money(m.view.MachineHourlyRate)},
		{"Printer purchase cost", money(m.view.PrinterPurchaseCost)},
		{"Default margin", unit.FormatPercent(m.view.DefaultMargin)},
		{"Minimum margin", unit.FormatPercent(m.view.MinMargin)},
	}
	for _, row := range rows {
		fmt.Fprintf(&b, "  %-*s%s\n", labelWidth, row[0], row[1])
	}

	b.WriteString("\n")
	fmt.Fprintf(&b, "  %-6s  %8s  %-8s\n", "TYPE", "KWH/H", "SOURCE")
	for _, rate := range m.view.PowerRates {
		fmt.Fprintf(&b, "  %-6s  %8s  %-8s\n",
			rate.FilamentType, unit.FormatKwhPerHour(rate.KwhPerHour), rateSource(rate.Measured))
	}
	return b.String()
}

func rateSource(measured bool) string {
	if measured {
		return "measured"
	}
	return "default"
}

func (m settingsModel) Help() string {
	if m.form != nil {
		return m.form.Help()
	}
	return fmt.Sprintf("%s edit · %s", KeyE, globalHelp)
}

func (m settingsModel) Refresh() tabModel {
	if err := m.reload(); err != nil {
		m.lastErr = err
	}
	return m
}
