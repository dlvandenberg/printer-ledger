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

const saleRow = "%s%-10s  %-20s  %-16s  %10s  %10s\n"

var errNoSellablePrints = errors.New("no copies available: record a print on the Prints tab first")

var _ tabModel = salesModel{}

type salesModel struct {
	app     *app.App
	rows    []app.SaleView
	cursor  int
	form    *form
	editing *app.SaleView
	confirm *confirm
	prints  []app.SellablePrintView
	preview app.SalePreviewView
	priced  string
	loadErr error
}

func newSalesModel(a *app.App) (salesModel, error) {
	m := salesModel{app: a}
	if err := m.reload(); err != nil {
		return m, err
	}
	return m, nil
}

func (m *salesModel) reload() error {
	rows, err := m.app.ListSales(context.Background())
	if err != nil {
		return err
	}
	m.rows = rows
	if m.cursor >= len(rows) {
		m.cursor = max(0, len(rows)-1)
	}
	return nil
}

func (m salesModel) Help() string {
	if m.form != nil {
		return m.form.Help()
	}
	if m.confirm != nil {
		return m.confirm.Help()
	}
	var rowHelp string
	if len(m.rows) > 0 {
		rowHelp = fmt.Sprintf(" · %s edit · %s delete", KeyE, KeyD)
	}
	return fmt.Sprintf("%s record%s · %s/%s move · %s", KeyA, rowHelp, KeyDown, KeyUp, globalHelp)
}

func (m salesModel) CapturesInput() bool { return m.form != nil || m.confirm != nil }

func (m salesModel) Update(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	if m.form != nil {
		return m.updateForm(msg)
	}
	if m.confirm != nil {
		return m.updateConfirm(msg)
	}

	switch msg.String() {
	case KeyA:
		if err := m.openForm(); err != nil {
			m.loadErr = err
		}
	case KeyE:
		if len(m.rows) > 0 {
			if err := m.openEditForm(m.rows[m.cursor]); err != nil {
				m.loadErr = err
			}
		}
	case KeyD:
		if len(m.rows) > 0 {
			m.askDelete(m.rows[m.cursor])
		}
	case KeyUp, KeyK:
		if m.cursor > 0 {
			m.cursor--
		}
	case KeyDown, KeyJ:
		if m.cursor < len(m.rows)-1 {
			m.cursor++
		}
	}
	return m, nil
}

func (m *salesModel) openForm() error {
	prints, err := m.app.SellablePrints(context.Background())
	if err != nil {
		return err
	}
	if len(prints) == 0 {
		return errNoSellablePrints
	}

	m.loadErr = nil
	m.editing = nil
	m.prints = prints
	m.form = newSaleForm(prints)
	m.priced = ""
	m.reprice()
	return nil
}

// openEditForm releases the copy the Sale holds, so the Print it came from is
// offered even when that copy was the last one available.
func (m *salesModel) openEditForm(sale app.SaleView) error {
	prints, err := m.app.SellablePrintsForSale(context.Background(), sale.ID)
	if err != nil {
		return err
	}

	m.loadErr = nil
	m.editing = &sale
	m.prints = prints
	m.form = newEditSaleForm(prints, sale)
	m.priced = ""
	m.reprice()
	return nil
}

func (m *salesModel) askDelete(sale app.SaleView) {
	m.loadErr = nil
	m.confirm = newConfirm(
		fmt.Sprintf("Delete the sale of %s on %s? The copy returns to available.",
			sale.DesignName, unit.FormatDate(sale.Date)),
		func() error { return m.app.DeleteSale(context.Background(), sale.ID) })
}

func (m salesModel) updateConfirm(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	confirmed, deleted, err := m.confirm.Answer(msg)
	m.confirm = confirmed
	switch {
	case err != nil:
		m.loadErr = err
	case deleted:
		if err := m.reload(); err != nil {
			m.loadErr = err
		}
	}
	return m, nil
}

func (m salesModel) updateForm(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	if cmd, handled := m.form.Update(msg); handled {
		m.reprice()
		return m, cmd
	}

	switch msg.String() {
	case KeyEsc:
		m.form = nil
		return m, nil
	case KeyEnter:
		if err := m.submitForm(); err != nil {
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

	return m, nil
}

func (m salesModel) submitForm() error {
	if m.editing != nil {
		_, err := m.app.EditSale(context.Background(), editSaleCmd(m.form, m.editing.ID, m.prints))
		return err
	}
	_, err := m.app.RecordSale(context.Background(), recordSaleCmd(m.form, m.prints))
	return err
}

func (m *salesModel) reprice() {
	cmd := recordSaleCmd(m.form, m.prints)
	key := fmt.Sprintf("%d|%s", cmd.PrintID, cmd.Price)
	if key == m.priced {
		return
	}
	m.priced = key

	preview, err := m.app.PreviewSale(context.Background(), cmd)
	if err != nil {
		m.loadErr = err
		return
	}
	m.preview = preview
}

func (m salesModel) View() string {
	if m.form != nil {
		if m.form.PickerOpen() {
			return m.failure() + m.form.View()
		}
		return m.failure() + m.form.View() + "\n" + priceGuide(m.preview)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Sales"))
	b.WriteString("\n\n")
	b.WriteString(m.failure())

	if len(m.rows) == 0 {
		b.WriteString(placeholderStyle.Render(fmt.Sprintf("No sales yet. Press %s to record one.", KeyA)))
		return b.String()
	}

	fmt.Fprintf(&b, saleRow, "  ", "DATE", "DESIGN", "MATERIAL", "PRICE", "PER COPY")
	for i, row := range m.rows {
		marker := "  "
		if i == m.cursor {
			marker = "> "
		}
		fmt.Fprintf(&b, saleRow, marker,
			unit.FormatDate(row.Date),
			truncate(row.DesignName, 20),
			truncate(material(row.Material), materialWidth),
			currency+unit.FormatCents(row.Price),
			currency+unit.FormatCents(row.CostPerCopy))
	}
	if m.confirm != nil {
		b.WriteString("\n")
		b.WriteString(m.confirm.View())
	}
	return b.String()
}

// priceGuide puts what the copy cost beside what the Design suggests charging,
// which is the pair of numbers a price is decided against. A price under the
// floor warns and records anyway (ADR-0016).
func priceGuide(preview app.SalePreviewView) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Copy"))
	b.WriteString("\n")
	fmt.Fprintf(&b, printCostLine, "cost", costAmount(preview.CostPerCopy, plainStyle))
	if preview.HasSuggestedPrice {
		fmt.Fprintf(&b, printCostLine, "suggested", costAmount(preview.SuggestedPrice, titleStyle))
	}
	fmt.Fprintf(&b, printCostLine, "floor", costAmount(preview.MinimumPrice, plainStyle))
	if preview.BelowFloor {
		b.WriteString("\n")
		b.WriteString(warningStyle.Render(fmt.Sprintf(
			"below %s%s, the minimum margin over what this copy cost — recording it anyway",
			currency, unit.FormatCents(preview.MinimumPrice))))
		b.WriteString("\n")
	}
	return b.String()
}

func (m salesModel) failure() string {
	if m.loadErr == nil {
		return ""
	}
	return errorStyle.Render(m.loadErr.Error()) + "\n\n"
}

func (m salesModel) Refresh() tabModel {
	if err := m.reload(); err != nil {
		m.loadErr = err
	}
	return m
}
