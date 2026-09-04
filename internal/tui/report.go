package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const (
	reportLine      = "  %-28s %s\n"
	reportCountLine = "  %-28s %s  %s\n"
	reportDesignRow = "  %-20s  %6s  %10s  %10s  %10s  %8s\n"
)

var _ tabModel = reportModel{}

type reportModel struct {
	app     *app.App
	periods []domain.Period
	period  int
	view    app.ReportView
	loadErr error
}

func newReportModel(a *app.App) (reportModel, error) {
	m := reportModel{app: a, periods: a.ReportPeriods()}
	if err := m.reload(); err != nil {
		return m, err
	}
	return m, nil
}

func (m *reportModel) reload() error {
	view, err := m.app.Report(context.Background(), app.ReportCmd{Period: m.periods[m.period].String()})
	if err != nil {
		return err
	}
	m.view = view
	return nil
}

func (m reportModel) CapturesInput() bool { return false }

func (m reportModel) Update(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	switch msg.String() {
	case KeyLeft, KeyH:
		m.selectPeriod(-1)
	case KeyRight, KeyL:
		m.selectPeriod(1)
	}
	return m, nil
}

func (m *reportModel) selectPeriod(delta int) {
	m.period = wrap(m.period, delta, len(m.periods))
	m.loadErr = m.reload()
}

func (m reportModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Report"))
	b.WriteString("\n\n")
	if m.loadErr != nil {
		b.WriteString(errorStyle.Render(m.loadErr.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString(m.periodBar())
	b.WriteString("\n\n")
	b.WriteString(m.sold())
	b.WriteString("\n")
	b.WriteString(m.produced())
	b.WriteString("\n")
	b.WriteString(m.designs())
	b.WriteString("\n")
	b.WriteString(m.breakEven())
	b.WriteString("\n")
	b.WriteString(m.inventory())
	return b.String()
}

func (m reportModel) periodBar() string {
	rendered := make([]string, 0, len(m.periods))
	for i, period := range m.periods {
		style := inactiveTabStyle
		if i == m.period {
			style = activeTabStyle
		}
		rendered = append(rendered, style.Render(period.String()))
	}
	return strings.Join(rendered, " ") + "  " + helpStyle.Render(m.dates())
}

func (m reportModel) dates() string {
	if !m.view.Bounded {
		return "everything recorded"
	}
	return fmt.Sprintf("%s to %s", unit.FormatDate(m.view.From), unit.FormatDate(m.view.Through))
}

// sold is the matched measure: the cost is the cost of the copies sold in the
// period, whenever they were printed (ADR-0006). The heading says so, because
// it will not agree with break-even.
func (m reportModel) sold() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Sold in this period"))
	b.WriteString(helpStyle.Render("  cost matched to what sold"))
	b.WriteString("\n")
	fmt.Fprintf(&b, reportLine, "Revenue", amount(money(m.view.Revenue)))
	fmt.Fprintf(&b, reportLine, "Cost of copies sold", amount(money(m.view.CostOfSold)))
	fmt.Fprintf(&b, reportLine, "Profit", titleStyle.Render(amount(money(m.view.Profit))))
	fmt.Fprintf(&b, reportLine, "Margin", amount(margin(m.view.MarginPct, m.view.HasMargin)))
	fmt.Fprintf(&b, reportLine, "Copies sold", amount(unit.FormatCopies(m.view.CopiesSold)))
	return b.String()
}

func (m reportModel) produced() string {
	notSold := m.view.NotSold
	var b strings.Builder
	b.WriteString(titleStyle.Render("Printed in this period"))
	b.WriteString(helpStyle.Render("  whether or not it sold"))
	b.WriteString("\n")
	fmt.Fprintf(&b, reportLine, "Production cost", amount(money(m.view.ProductionCost)))
	fmt.Fprintf(&b, reportCountLine, "Gifted, kept, scrapped", amount(money(notSold.Cost)),
		helpStyle.Render(fmt.Sprintf("%s copies (%s gifted, %s kept, %s scrapped)",
			unit.FormatCopies(notSold.Copies),
			unit.FormatCopies(notSold.GiftedCount),
			unit.FormatCopies(notSold.KeptCount),
			unit.FormatCopies(notSold.ScrappedCount))))
	return b.String()
}

func (m reportModel) designs() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Designs by profit"))
	b.WriteString("\n")
	if len(m.view.Designs) == 0 {
		b.WriteString(placeholderStyle.Render("Nothing sold in this period."))
		b.WriteString("\n")
		return b.String()
	}

	fmt.Fprintf(&b, reportDesignRow, "DESIGN", "SOLD", "REVENUE", "COST", "PROFIT", "MARGIN")
	for _, design := range m.view.Designs {
		fmt.Fprintf(&b, reportDesignRow,
			truncate(design.DesignName, 20),
			unit.FormatCopies(design.CopiesSold),
			money(design.Revenue),
			money(design.Cost),
			money(design.Profit),
			margin(design.MarginPct, design.HasMargin))
	}
	return b.String()
}

// breakEven is the cash measure, all time and independent of the period. It
// disagrees with profit above on purpose, so both headings name their basis
// (ADR-0006).
func (m reportModel) breakEven() string {
	cash := m.view.BreakEven
	var b strings.Builder
	b.WriteString(titleStyle.Render("Break-even"))
	b.WriteString(helpStyle.Render("  cash, all time, every euro spent"))
	b.WriteString("\n")
	fmt.Fprintf(&b, reportLine, "Revenue", amount(money(cash.Revenue)))
	fmt.Fprintf(&b, reportLine, "Printer", amount(money(-cash.PrinterCost)))
	fmt.Fprintf(&b, reportLine, "Filament bought", amount(money(-cash.SpoolSpend)))
	fmt.Fprintf(&b, reportLine, "Energy", amount(money(-cash.EnergySpend)))
	fmt.Fprintf(&b, reportCountLine, "Balance", titleStyle.Render(amount(money(cash.Balance))),
		helpStyle.Render(breakEvenState(cash)))
	return b.String()
}

func breakEvenState(cash app.BreakEvenView) string {
	if cash.Reached {
		return "broken even"
	}
	return money(-cash.Balance) + " to go"
}

func (m reportModel) inventory() string {
	inventory := m.view.Inventory
	var b strings.Builder
	b.WriteString(titleStyle.Render("Inventory"))
	b.WriteString(helpStyle.Render("  owned right now, all time"))
	b.WriteString("\n")
	fmt.Fprintf(&b, reportCountLine, "Unsold copies", amount(money(inventory.UnsoldValue)),
		helpStyle.Render(unit.FormatCopies(inventory.UnsoldCopies)+" copies"))
	fmt.Fprintf(&b, reportCountLine, "Filament on the shelf", amount(money(inventory.SpoolValue)),
		helpStyle.Render(unit.FormatGrams(inventory.SpoolGrams)))
	fmt.Fprintf(&b, reportLine, "Value", titleStyle.Render(amount(money(inventory.Value))))
	return b.String()
}

func margin(p unit.Percent, has bool) string {
	if !has {
		return "—"
	}
	return unit.FormatPercent(p)
}

func (m reportModel) Help() string {
	return fmt.Sprintf("%s/%s period · %s", KeyLeft, KeyRight, globalHelp)
}

func (m reportModel) Refresh() tabModel {
	m.loadErr = m.reload()
	return m
}
