package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
	"github.com/dlvandenberg/printer-ledger/internal/domain/unit"
)

// One format per table, shared by its header and its rows, so the two cannot
// drift out of alignment.
const (
	printRow      = "%s%-10s  %-16s  %6s  %6s  %-9s  %7s  %10s  %10s\n"
	printCostLine = "  %-10s  %s\n"
)

var (
	errNoDesigns = errors.New("no designs yet: add one on the Designs tab first")
	errNoSpools  = errors.New("no spool with filament left: add or re-weigh one on the Spools tab first")
)

var _ tabModel = printsModel{}

type printsModel struct {
	app       *app.App
	rows      []app.PrintView
	cursor    int
	form      *form
	detail    *app.PrintView
	editing   *app.PrintView
	confirm   *confirm
	spools    []app.SpoolView
	preview   app.PrintPreviewView
	usageRows int
	drafted   string
	priced    string
	loadErr   error
}

func newPrintsModel(a *app.App) (printsModel, error) {
	m := printsModel{app: a}
	if err := m.reload(); err != nil {
		return m, err
	}
	return m, nil
}

func (m *printsModel) reload() error {
	rows, err := m.app.ListPrints(context.Background())
	if err != nil {
		return err
	}
	m.rows = rows
	if m.cursor >= len(rows) {
		m.cursor = max(0, len(rows)-1)
	}
	return nil
}

func (m printsModel) Help() string {
	if m.form != nil {
		return m.form.Help()
	}
	if m.confirm != nil {
		return m.confirm.Help()
	}
	if m.detail != nil {
		return fmt.Sprintf("%s edit · %s back · %s", KeyE, KeyEsc, globalHelp)
	}
	var rowHelp string
	if len(m.rows) > 0 {
		rowHelp = fmt.Sprintf(" · %s detail · %s edit · %s delete", KeyEnter, KeyE, KeyD)
	}
	return fmt.Sprintf("%s record%s · %s/%s move · %s", KeyA, rowHelp, KeyDown, KeyUp, globalHelp)
}

func (m printsModel) CapturesInput() bool { return m.form != nil || m.confirm != nil }

func (m printsModel) Update(msg tea.KeyMsg) (tabModel, tea.Cmd) {
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

func (m printsModel) updateDetail(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	switch msg.String() {
	case KeyEsc:
		m.detail = nil
	case KeyE:
		if err := m.openEditForm(*m.detail); err != nil {
			m.loadErr = err
		}
	}
	return m, nil
}

func (m *printsModel) openDetail(id int64) {
	detail, err := m.app.PrintDetail(context.Background(), id)
	if err != nil {
		m.loadErr = err
		return
	}
	m.detail = &detail
}

func (m *printsModel) openForm() error {
	designs, err := m.app.ListDesigns(context.Background())
	if err != nil {
		return err
	}
	if len(designs) == 0 {
		return errNoDesigns
	}
	spools, err := m.app.ListActiveSpools(context.Background())
	if err != nil {
		return err
	}
	if len(spools) == 0 {
		return errNoSpools
	}

	draft, err := m.app.PrintDraft(context.Background(), designs[0].ID, "1")
	if err != nil {
		return err
	}

	m.loadErr = nil
	m.editing = nil
	m.spools = spools
	m.form = newPrintForm(designs, spools, draft)
	m.reserveUsageChords()
	m.usageRows = 1
	m.drafted, m.priced = "", ""
	m.reprice()
	return nil
}

// openEditForm offers every Spool, empty ones included: a Print may already
// have drawn a Spool down to nothing, and its own row must still name it.
func (m *printsModel) openEditForm(print app.PrintView) error {
	designs, err := m.app.ListDesigns(context.Background())
	if err != nil {
		return err
	}
	spools, err := m.app.ListSpools(context.Background())
	if err != nil {
		return err
	}

	m.loadErr = nil
	m.editing = &print
	m.spools = spools
	m.form, m.usageRows = newEditPrintForm(designs, spools, print)
	m.reserveUsageChords()
	m.drafted = m.draftKey()
	m.priced = ""
	m.reprice()
	return nil
}

func (m *printsModel) askDelete(print app.PrintView) {
	m.loadErr = nil
	m.confirm = newConfirm(
		fmt.Sprintf("Delete the print of %s on %s?", print.DesignName, unit.FormatDate(print.Date)),
		func() error { return m.app.DeletePrint(context.Background(), print.ID) })
}

func (m printsModel) updateConfirm(msg tea.KeyMsg) (tabModel, tea.Cmd) {
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

func (m printsModel) updateForm(msg tea.KeyMsg) (tabModel, tea.Cmd) {
	if cmd, handled := m.form.Update(msg); handled {
		m.applyDraft()
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
	case KeyCtrlR:
		m.addUsageRow()
		return m, nil
	case KeyCtrlX:
		m.removeUsageRow()
		return m, nil
	}

	return m, nil
}

func (m *printsModel) submitForm() error {
	if m.editing != nil {
		edited, err := m.app.EditPrint(context.Background(), editPrintCmd(m.form, m.editing.ID, m.usageRows))
		if err != nil {
			return err
		}
		if m.detail != nil {
			m.detail = &edited
		}
		return nil
	}
	_, err := m.app.RecordPrint(context.Background(), recordPrintCmd(m.form, m.usageRows))
	return err
}

// addUsageRow and removeUsageRow record a mid-print spool swap as one Print
// (ADR-0012). A row change drops the errors, which are keyed by row. The new
// row opens on the Spool above it, so an untouched row cannot mix types.
func (m *printsModel) reserveUsageChords() {
	m.form.Reserve(KeyCtrlR, "add spool row")
	m.form.Reserve(KeyCtrlX, "remove spool row")
}

func (m *printsModel) addUsageRow() {
	above := m.form.Choice(domain.FieldUsageSpool(m.usageRows - 1))
	m.form.Append(usageRowSpecs(m.usageRows, choicesOf(m.spools, spoolChoice), "", above)...)
	m.usageRows++
	m.form.SetErrors(nil)
	m.reprice()
}

func (m *printsModel) removeUsageRow() {
	if m.usageRows < 2 {
		return
	}
	m.usageRows--
	m.form.DropLast(usageFieldsPerRow)
	m.form.SetErrors(nil)
	m.reprice()
}

// applyDraft re-prefills the estimates when the design or the quantity moves.
// A field the operator has typed into keeps what they typed, so a corrected
// quantity never overwrites an actual read off the printer. An edit form is
// never re-prefilled: every field on it is already an actual.
func (m *printsModel) applyDraft() {
	if m.editing != nil {
		return
	}
	key := m.draftKey()
	if key == m.drafted {
		return
	}
	m.drafted = key

	draft, err := m.app.PrintDraft(context.Background(),
		m.form.Choice(domain.FieldDesignID).ID, m.form.Value(domain.FieldQuantity))
	if err != nil {
		return
	}
	m.form.Prefill(domain.FieldMinutes, unit.FormatHHmm(draft.Minutes))
	m.form.Prefill(domain.FieldUsageGrams(0), unit.FormatGrams(draft.Grams))
}

// draftKey is what a draft was fetched for. It names the Design by id, so two
// Designs of one name are not one key.
func (m *printsModel) draftKey() string {
	return fmt.Sprintf("%d|%s", m.form.Choice(domain.FieldDesignID).ID, m.form.Value(domain.FieldQuantity))
}

func (m *printsModel) reprice() {
	cmd := recordPrintCmd(m.form, m.usageRows)
	key := costInputs(cmd)
	if key == m.priced {
		return
	}
	m.priced = key

	preview, err := m.previewOf(cmd)
	if err != nil {
		m.loadErr = err
		return
	}
	m.preview = preview
}

func (m printsModel) previewOf(cmd app.RecordPrintCmd) (app.PrintPreviewView, error) {
	if m.editing != nil {
		return m.app.PreviewEditPrint(context.Background(),
			app.EditPrintCmd{PrintID: m.editing.ID, RecordPrintCmd: cmd})
	}
	return m.app.PreviewPrint(context.Background(), cmd)
}

// costInputs names every field the breakdown reads, so a key that moves none of
// them re-renders the last breakdown instead of asking for it again.
func costInputs(cmd app.RecordPrintCmd) string {
	key := cmd.Quantity + "|" + cmd.Minutes
	for _, usage := range cmd.Usages {
		key += fmt.Sprintf("|%d:%s", usage.SpoolID, usage.Grams)
	}
	return key
}

func (m printsModel) View() string {
	if m.form != nil {
		if m.form.PickerOpen() {
			return m.failure() + m.form.View()
		}
		return m.failure() + m.form.View() + "\n" + costBreakdown(m.preview)
	}
	if m.detail != nil {
		return m.failure() + printDetailView(*m.detail)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Prints"))
	b.WriteString("\n\n")
	b.WriteString(m.failure())

	if len(m.rows) == 0 {
		b.WriteString(placeholderStyle.Render(fmt.Sprintf("No prints yet. Press %s to record one.", KeyA)))
		return b.String()
	}

	fmt.Fprintf(&b, printRow, "  ",
		"DATE", "DESIGN", "COPIES", "AVAIL", "TIME", "GRAMS", "JOB COST", "PER COPY")
	for i, row := range m.rows {
		marker := "  "
		if i == m.cursor {
			marker = "> "
		}
		fmt.Fprintf(&b, printRow, marker,
			unit.FormatDate(row.Date),
			truncate(row.DesignName, 16),
			unit.FormatCopies(row.Quantity),
			unit.FormatCopies(row.AvailableCopies),
			unit.FormatMinutes(row.Minutes),
			unit.FormatGrams(row.UsedGrams),
			currency+unit.FormatCents(row.Cost.JobCost),
			currency+unit.FormatCents(row.Cost.CostPerCopy))
	}
	if m.confirm != nil {
		b.WriteString("\n")
		b.WriteString(m.confirm.View())
	}
	return b.String()
}

// costBreakdown shows what the job cost and what one copy cost. A Print never
// shows a suggested price: the price was decided on the Design. An unmeasured
// power rate warns and never blocks: the figure is a placeholder, not an error.
func costBreakdown(preview app.PrintPreviewView) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Cost"))
	b.WriteString("\n")
	b.WriteString(costLines(preview.Cost))
	if preview.RateSeeded {
		b.WriteString("\n")
		b.WriteString(warningStyle.Render(fmt.Sprintf(
			"energy uses the %s rate the ledger seeded, not a measured one", preview.FilamentType)))
		b.WriteString("\n")
	}
	return b.String()
}

// costLines is the breakdown the form and the detail screen share, so the two
// cannot show one cost split two ways.
func costLines(cost app.PrintCostView) string {
	var b strings.Builder
	fmt.Fprintf(&b, printCostLine, "filament", costAmount(cost.Filament, plainStyle))
	fmt.Fprintf(&b, printCostLine, "energy", costAmount(cost.Energy, plainStyle))
	fmt.Fprintf(&b, printCostLine, "overhead", costAmount(cost.Overhead, plainStyle))
	fmt.Fprintf(&b, printCostLine, "job cost", costAmount(cost.JobCost, titleStyle))
	fmt.Fprintf(&b, printCostLine, "per copy", costAmount(cost.CostPerCopy, titleStyle))
	return b.String()
}

// costAmount pads before styling: a width verb counts the escape bytes of an
// already-styled string, so a bold amount would sit short of its column.
func costAmount(cents unit.Cents, style lipgloss.Style) string {
	return style.Render(fmt.Sprintf("%10s", currency+unit.FormatCents(cents)))
}

func (m printsModel) failure() string {
	if m.loadErr == nil {
		return ""
	}
	return errorStyle.Render(m.loadErr.Error()) + "\n\n"
}

func (m printsModel) Refresh() tabModel {
	if err := m.reload(); err != nil {
		m.loadErr = err
	}
	if m.detail != nil {
		m.openDetail(m.detail.ID)
	}
	return m
}
