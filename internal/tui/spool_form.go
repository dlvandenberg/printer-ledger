package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

var (
	labelStyle        = lipgloss.NewStyle().Width(24)
	focusedLabelStyle = labelStyle.Bold(true).Foreground(lipgloss.Color("62"))
)

type formInput struct {
	field string
	label string
	input textinput.Model
}

type spoolForm struct {
	typeIndex int
	inputs    []formInput
	focus     int
	errs      *domain.ValidationError
}

func newSpoolForm() spoolForm {
	specs := []struct {
		field       string
		label       string
		placeholder string
		value       string
	}{
		{domain.FieldBrand, "Brand", "Bambu", ""},
		{domain.FieldColor, "Colour", "Black", ""},
		{domain.FieldInitialGrams, "Filament grams", "1000", ""},
		{domain.FieldTareGrams, "Empty spool grams", "210", ""},
		{domain.FieldPurchaseCost, "Purchase price in EUR", "22.00", ""},
		{domain.FieldPurchaseDate, "Purchase date", domain.DateLayout, domain.FormatDate(domain.Today())},
	}

	inputs := make([]formInput, 0, len(specs))
	for _, spec := range specs {
		in := textinput.New()
		in.Placeholder = spec.placeholder
		in.SetValue(spec.value)
		in.CharLimit = 32
		in.Width = 24
		inputs = append(inputs, formInput{field: spec.field, label: spec.label, input: in})
	}

	form := spoolForm{inputs: inputs}
	form.applyFocus()
	return form
}

func (f spoolForm) fieldCount() int { return len(f.inputs) + 1 }

func (f *spoolForm) applyFocus() {
	for i := range f.inputs {
		if f.focus == i+1 {
			f.inputs[i].input.Focus()
		} else {
			f.inputs[i].input.Blur()
		}
	}
}

func (f spoolForm) update(msg tea.KeyMsg) (spoolForm, tea.Cmd) {
	switch msg.String() {
	case "tab", "down":
		f.focus = (f.focus + 1) % f.fieldCount()
		f.applyFocus()
		return f, nil
	case "shift+tab", "up":
		f.focus = (f.focus + f.fieldCount() - 1) % f.fieldCount()
		f.applyFocus()
		return f, nil
	case "left":
		if f.focus == 0 {
			types := domain.FilamentTypes()
			f.typeIndex = (f.typeIndex + len(types) - 1) % len(types)
			return f, nil
		}
	case "right":
		if f.focus == 0 {
			types := domain.FilamentTypes()
			f.typeIndex = (f.typeIndex + 1) % len(types)
			return f, nil
		}
	}

	if f.focus == 0 {
		return f, nil
	}

	var cmd tea.Cmd
	f.inputs[f.focus-1].input, cmd = f.inputs[f.focus-1].input.Update(msg)
	return f, cmd
}

func (f spoolForm) filamentType() domain.FilamentType {
	return domain.FilamentTypes()[f.typeIndex]
}

func (f spoolForm) value(field string) string {
	for _, in := range f.inputs {
		if in.field == field {
			return in.input.Value()
		}
	}
	return ""
}

func (f spoolForm) parse() (app.AddSpoolCmd, *domain.ValidationError) {
	errs := &domain.ValidationError{}
	cmd := app.AddSpoolCmd{
		FilamentType: f.filamentType(),
		Brand:        f.value(domain.FieldBrand),
		Color:        f.value(domain.FieldColor),
	}

	if grams, err := domain.ParseGrams(f.value(domain.FieldInitialGrams)); err != nil {
		errs.Add(domain.FieldInitialGrams, err.Error())
	} else {
		cmd.InitialGrams = grams
	}

	if grams, err := domain.ParseGrams(f.value(domain.FieldTareGrams)); err != nil {
		errs.Add(domain.FieldTareGrams, err.Error())
	} else {
		cmd.TareGrams = grams
	}

	if cents, err := domain.ParseCents(f.value(domain.FieldPurchaseCost)); err != nil {
		errs.Add(domain.FieldPurchaseCost, err.Error())
	} else {
		cmd.PurchaseCost = cents
	}

	if date, err := domain.ParseDate(f.value(domain.FieldPurchaseDate)); err != nil {
		errs.Add(domain.FieldPurchaseDate, err.Error())
	} else {
		cmd.PurchaseDate = date
	}

	return cmd, errs
}

func (f spoolForm) view() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Add spool"))
	b.WriteString("\n\n")
	b.WriteString(f.row(0, "Filament type", f.filamentTypeField(), domain.FieldFilamentType))
	for i, in := range f.inputs {
		b.WriteString(f.row(i+1, in.label, in.input.View(), in.field))
	}
	return b.String()
}

func (f spoolForm) row(index int, label, control, field string) string {
	style := labelStyle
	if f.focus == index {
		style = focusedLabelStyle
	}
	line := style.Render(label) + control
	if msg := f.errs.For(field); msg != "" {
		line += "  " + errorStyle.Render("← "+msg)
	}
	return line + "\n"
}

func (f spoolForm) filamentTypeField() string {
	types := domain.FilamentTypes()
	rendered := make([]string, 0, len(types))
	for i, t := range types {
		if i == f.typeIndex {
			rendered = append(rendered, activeTabStyle.Render(t.String()))
			continue
		}
		rendered = append(rendered, inactiveTabStyle.Render(t.String()))
	}
	return fmt.Sprintf("%-24s", lipgloss.JoinHorizontal(lipgloss.Top, rendered...))
}
