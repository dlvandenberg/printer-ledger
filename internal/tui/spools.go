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

const (
	brandLength = 10
	colorLength = 16
	spoolsRow   = "%s%-6s  %-10s  %-16s  %10s  %10s  %-7s\n" // Type, Brand, Color, Remaining, Value, State
)

var _ tabModel = spoolsModel{}

type spoolsModel struct {
	app     *app.App
	rows    []app.SpoolView
	cursor  int
	detail  *app.SpoolDetailView
	form    *form
	confirm *confirm
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

func (m spoolsModel) CapturesInput() bool { return m.form != nil || m.confirm != nil }

func (m spoolsModel) Update(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	if m.form != nil {
		return m.updateForm(msg)
	}
	if m.confirm != nil {
		return m.updateConfirm(msg)
	}

	if m.detail != nil {
		return m.updateDetail(msg)
	}

	switch msg.String() {
	case KeyA:
		m.form = newSpoolForm()
	case KeyD:
		if len(m.rows) > 0 {
			m.askDelete(m.rows[m.cursor])
		}
	case KeyEnter:
		if len(m.rows) > 0 {
			m.openDetail(m.rows[m.cursor].ID)
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

func (m *spoolsModel) askDelete(spool app.SpoolView) {
	m.loadErr = nil
	m.confirm = newConfirm(
		fmt.Sprintf("Delete the %s spool %s %s?", spool.FilamentType, spool.Brand, spool.Color),
		func() error { return m.app.DeleteSpool(context.Background(), spool.ID) })
}

func (m spoolsModel) updateConfirm(msg tea.KeyMsg) (tabModel, tea.Cmd) {
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

func (m spoolsModel) updateDetail(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	switch msg.String() {
	case KeyEsc:
		m.detail = nil
	case KeyR:
		m.form = newReweighForm()
	}
	return m, nil
}

func (m *spoolsModel) openDetail(id int64) {
	detail, err := m.app.SpoolDetail(context.Background(), id)
	if err != nil {
		m.loadErr = err
		return
	}
	m.detail = &detail
}

func (m spoolsModel) updateForm(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	if cmd, handled := m.form.Update(msg); handled {
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

// submitForm reads the tab's own state to decide which use case the open form
// belongs to: a detail screen is open only for a re-weigh.
func (m *spoolsModel) submitForm() error {
	if m.detail == nil {
		_, err := m.app.AddSpool(context.Background(), addSpoolCmd(m.form))
		return err
	}
	detail, err := m.app.ReweighSpool(context.Background(), reweighSpoolCmd(m.form, m.detail.Spool.ID))
	if err != nil {
		return err
	}
	m.detail = &detail
	return nil
}

func (m spoolsModel) View() string {
	if m.form != nil {
		return m.failure() + m.form.View()
	}
	if m.detail != nil {
		return m.failure() + spoolDetailView(*m.detail)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Spools"))
	b.WriteString("\n\n")
	b.WriteString(m.failure())

	if len(m.rows) == 0 {
		b.WriteString(placeholderStyle.Render(fmt.Sprintf("No spools yet. Press %s to add one.", KeyA)))
		return b.String()
	}

	fmt.Fprintf(&b, spoolsRow, "  ",
		"TYPE", "BRAND", "COLOUR", "REMAINING", "VALUE", "STATE")
	for i, row := range m.rows {
		marker := "  "
		if i == m.cursor {
			marker = "> "
		}
		fmt.Fprintf(&b, spoolsRow,
			marker, row.FilamentType, truncate(row.Brand, brandLength), truncate(row.Color, colorLength),
			unit.FormatGrams(row.RemainingGrams),
			currency+unit.FormatCents(row.RemainingValue),
			row.State)
	}
	if m.confirm != nil {
		b.WriteString("\n")
		b.WriteString(m.confirm.View())
	}
	return b.String()
}

func (m spoolsModel) failure() string {
	if m.loadErr == nil {
		return ""
	}
	return errorStyle.Render(m.loadErr.Error()) + "\n\n"
}

func (m spoolsModel) Help() string {
	if m.form != nil {
		return m.form.Help()
	}
	if m.confirm != nil {
		return m.confirm.Help()
	}
	if m.detail != nil {
		return fmt.Sprintf("%s re-weigh · %s back · %s", KeyR, KeyEsc, globalHelp)
	}
	var rowHelp string
	if len(m.rows) > 0 {
		rowHelp = fmt.Sprintf(" · %s detail · %s delete", KeyEnter, KeyD)
	}
	return fmt.Sprintf("%s add%s · %s/%s move · %s", KeyA, rowHelp, KeyUp, KeyDown, globalHelp)
}

func (m spoolsModel) Refresh() tabModel {
	if err := m.reload(); err != nil {
		m.loadErr = err
	}
	if m.detail != nil {
		m.openDetail(m.detail.Spool.ID)
	}
	return m
}
