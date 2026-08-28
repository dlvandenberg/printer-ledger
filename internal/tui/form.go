package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

const (
	labelWidth     = 24
	fieldWidth     = 24
	fieldCharLimit = 32
)

var (
	labelStyle        = lipgloss.NewStyle().Width(labelWidth)
	focusedLabelStyle = labelStyle.Bold(true).Foreground(lipgloss.Color("62"))
)

// fieldSpec declares one row of a form. A tab supplies the keys, labels and
// kinds; the form owns focus, layout, choice cycling and error placement.
// Choice and text fields sit in the same list, so neither has a position that
// means anything.
type fieldSpec struct {
	// key is the field name errors are keyed by and values are read back by.
	// Use the domain.Field* constants so a use case's ValidationError lands on
	// the right row.
	key   string
	label string
	// choices non-empty makes this a choice field, cycled with left/right.
	// Empty makes it a text field.
	choices []string
	// placeholder and value apply to text fields. A choice field starts on its
	// first choice.
	placeholder string
	value       string
}

type field struct {
	spec   fieldSpec
	input  textinput.Model
	choice int
}

func (f field) isChoice() bool { return len(f.spec.choices) > 0 }

type form struct {
	title  string
	fields []field
	focus  int
	errs   *domain.ValidationError
}

func newForm(title string, specs []fieldSpec) form {
	fields := make([]field, 0, len(specs))
	for _, spec := range specs {
		f := field{spec: spec}
		if !f.isChoice() {
			in := textinput.New()
			in.Placeholder = spec.placeholder
			in.SetValue(spec.value)
			in.CharLimit = fieldCharLimit
			in.Width = fieldWidth
			f.input = in
		}
		fields = append(fields, f)
	}

	f := form{title: title, fields: fields}
	f.applyFocus()
	return f
}

func (f *form) applyFocus() {
	for i := range f.fields {
		if f.fields[i].isChoice() {
			continue
		}
		if f.focus == i {
			f.fields[i].input.Focus()
		} else {
			f.fields[i].input.Blur()
		}
	}
}

func (f form) update(msg tea.KeyMsg) (form, tea.Cmd) {
	// A value receiver would otherwise still share its backing array with the
	// caller, so focus and choice moves would land on a form the caller has not
	// replaced yet.
	f.fields = slices.Clone(f.fields)

	switch msg.String() {
	case "tab", "down":
		f.focus = (f.focus + 1) % len(f.fields)
		f.applyFocus()
		return f, nil
	case "shift+tab", "up":
		f.focus = (f.focus + len(f.fields) - 1) % len(f.fields)
		f.applyFocus()
		return f, nil
	case "left":
		if f.fields[f.focus].isChoice() {
			f.cycle(-1)
			return f, nil
		}
	case "right":
		if f.fields[f.focus].isChoice() {
			f.cycle(1)
			return f, nil
		}
	}

	if f.fields[f.focus].isChoice() {
		return f, nil
	}

	var cmd tea.Cmd
	f.fields[f.focus].input, cmd = f.fields[f.focus].input.Update(msg)
	return f, cmd
}

func (f *form) cycle(step int) {
	current := &f.fields[f.focus]
	n := len(current.spec.choices)
	current.choice = (current.choice + step + n) % n
}

// value reports what the operator typed or chose for a field. The form does no
// parsing and no validation; a use case decides whether the text is acceptable.
func (f form) value(key string) string {
	for _, fld := range f.fields {
		if fld.spec.key != key {
			continue
		}
		if fld.isChoice() {
			return fld.spec.choices[fld.choice]
		}
		return fld.input.Value()
	}
	return ""
}

func (f *form) setErrors(errs *domain.ValidationError) { f.errs = errs }

func (f form) view() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(f.title))
	b.WriteString("\n\n")
	for i, fld := range f.fields {
		b.WriteString(f.row(i, fld))
	}
	return b.String()
}

func (f form) row(index int, fld field) string {
	style := labelStyle
	if f.focus == index {
		style = focusedLabelStyle
	}
	line := style.Render(fld.spec.label) + f.control(fld)
	if msg := f.errs.For(fld.spec.key); msg != "" {
		line += "  " + errorStyle.Render("← "+msg)
	}
	return line + "\n"
}

func (f form) control(fld field) string {
	if !fld.isChoice() {
		return fld.input.View()
	}
	rendered := make([]string, 0, len(fld.spec.choices))
	for i, choice := range fld.spec.choices {
		style := inactiveTabStyle
		if i == fld.choice {
			style = activeTabStyle
		}
		rendered = append(rendered, style.Render(choice))
	}
	return fmt.Sprintf("%-*s", fieldWidth, lipgloss.JoinHorizontal(lipgloss.Top, rendered...))
}
