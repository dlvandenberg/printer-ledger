package tui

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

const (
	designRow = "%s%-16s  %-12s  %-16s  %-15s  %-7s  %6s\n"
)

var _ tabModel = designsModel{}

type designsModel struct {
	app     *app.App
	rows    []app.DesignView
	editing *app.DesignView
	quote   *app.DesignQuoteView
	cursor  int
	form    *form
	confirm *confirm
	loadErr error
}

func newDesignsModel(a *app.App) (designsModel, error) {
	m := designsModel{
		app: a,
	}
	if err := m.reload(); err != nil {
		return m, err
	}
	return m, nil
}

func (m *designsModel) reload() error {
	rows, err := m.app.ListDesigns(context.Background())
	if err != nil {
		return err
	}
	m.rows = rows
	if m.cursor >= len(rows) {
		m.cursor = max(0, len(rows)-1)
	}
	return nil
}

func (m designsModel) Help() string {
	if m.form != nil {
		return m.form.Help()
	}
	if m.confirm != nil {
		return m.confirm.Help()
	}
	if m.quote != nil {
		return fmt.Sprintf("%s edit · %s refresh · %s back · %s", KeyE, KeyR, KeyEsc, globalHelp)
	}
	var rowHelp string
	if len(m.rows) > 0 {
		rowHelp = fmt.Sprintf(" · %s edit · %s quote · %s delete", KeyE, KeyEnter, KeyD)
	}
	return fmt.Sprintf("%s add%s · %s/%s move · %s", KeyA, rowHelp, KeyDown, KeyUp, globalHelp)
}

func (m designsModel) View() string {
	if m.form != nil {
		return m.failure() + m.form.View()
	}
	if m.quote != nil {
		return m.failure() + designQuoteView(*m.quote)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Designs"))
	b.WriteString("\n\n")
	b.WriteString(m.failure())

	if len(m.rows) == 0 {
		b.WriteString(placeholderStyle.Render(fmt.Sprintf("No designs yet. Press %s to add one.", KeyA)))
		return b.String()
	}

	fmt.Fprintf(&b, designRow, "  ",
		"NAME", "DEFAULT TYPE", "ESTIMATED GRAMS", "ESTIMATED TIME", "MARGIN", "PRINTS")

	for i, row := range m.rows {
		marker := "  "
		if i == m.cursor {
			marker = "> "
		}
		fmt.Fprintf(&b, designRow,
			marker, truncate(row.Name, 16), row.DefaultFilamentType, unit.FormatGrams(row.EstimatedGrams), unit.FormatMinutes(row.EstimatedMinutes), unit.FormatPercent(row.MarginPct), strconv.FormatInt(row.PrintCount, 10))
	}

	if m.confirm != nil {
		b.WriteString("\n")
		b.WriteString(m.confirm.View())
	}

	return b.String()
}

func (m designsModel) CapturesInput() bool {
	return m.form != nil || m.confirm != nil
}

func (m designsModel) Update(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	if m.form != nil {
		return m.updateForm(msg)
	}
	if m.confirm != nil {
		return m.updateConfirm(msg)
	}
	if m.quote != nil {
		return m.updateQuote(msg)
	}

	switch msg.String() {
	case KeyA:
		settings, err := m.app.Settings(context.Background())
		if err != nil {
			m.loadErr = err
			return m, nil
		}
		m.form = newDesignForm(settings.DefaultMargin)
		m.editing = nil
	case KeyUp, KeyK:
		if m.cursor > 0 {
			m.cursor--
		}
	case KeyE:
		if len(m.rows) > 0 {
			row := m.rows[m.cursor]
			m.editing = &row
			m.form = editDesignForm(row)
		}
	case KeyEnter:
		if len(m.rows) > 0 {
			m.openQuote(m.rows[m.cursor].ID)
		}
	case KeyD:
		if len(m.rows) > 0 {
			m.askDelete(m.rows[m.cursor])
		}
	case KeyDown, KeyJ:
		if m.cursor < len(m.rows)-1 {
			m.cursor++
		}
	}
	return m, nil
}

// askDelete prompts even where the delete will be refused: whether a Design
// has been printed is the domain's answer, not the view's (ADR-0011).
func (m *designsModel) askDelete(design app.DesignView) {
	m.loadErr = nil
	m.confirm = newConfirm(
		fmt.Sprintf("Delete the design %s?", design.Name),
		func() error { return m.app.DeleteDesign(context.Background(), design.ID) })
}

func (m designsModel) updateConfirm(msg tea.KeyMsg) (tabModel, tea.Cmd) {
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

func (m designsModel) updateQuote(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	switch msg.String() {
	case KeyEsc:
		m.quote = nil
	case KeyE:
		row := m.quote.Design
		m.editing = &row
		m.form = editDesignForm(row)
	case KeyR:
		m.openQuote(m.quote.Design.ID)
	}
	return m, nil
}

func (m *designsModel) openQuote(id int64) {
	quote, err := m.app.DesignQuote(context.Background(), id)
	if err != nil {
		m.loadErr = err
		return
	}
	m.quote = &quote
}

func (m designsModel) updateForm(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	if cmd, handled := m.form.Update(msg); handled {
		return m, cmd
	}

	switch msg.String() {
	case KeyEsc:
		m.form = nil
		m.editing = nil
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
		m.editing = nil
		if err := m.reload(); err != nil {
			m.loadErr = err
		}
		if m.quote != nil {
			m.openQuote(m.quote.Design.ID)
		}
		return m, nil
	}
	return m, nil
}

func (m designsModel) submitForm() error {
	if m.editing != nil {
		_, err := m.app.EditDesign(context.Background(), editDesignCmd(m.form, m.editing.ID))
		return err
	}
	_, err := m.app.AddDesign(context.Background(), addDesignCmd(m.form))
	return err
}

func (m designsModel) failure() string {
	if m.loadErr == nil {
		return ""
	}
	return errorStyle.Render(m.loadErr.Error()) + "\n\n"
}

func (m designsModel) Refresh() tabModel {
	if err := m.reload(); err != nil {
		m.loadErr = err
	}
	return m
}
