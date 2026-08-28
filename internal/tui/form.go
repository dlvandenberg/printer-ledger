package tui

import (
	"fmt"
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
	// Key is the field name errors are keyed by and values are read back by.
	// Use the domain.Field* constants so a use case's ValidationError lands on
	// the right row.
	Key   string
	Label string
	// Choices non-empty makes this a choice field, cycled with left/right.
	// Empty makes it a text field.
	Choices []string
	// Placeholder and Prefill apply to text fields. A choice field starts on its
	// first choice.
	Placeholder string
	Prefill     string
}

type field struct {
	spec   fieldSpec
	input  textinput.Model
	choice int
}

func (f field) isChoice() bool { return len(f.spec.Choices) > 0 }

// form is mutable open state, always held as a *form. One tab owns it from the
// moment it opens the form until it drops the pointer, and never copies it out.
// Tab models are values that bubbletea replaces on every key, so a form that
// were copied with them would lose focus and choice moves; keeping one pointer
// is what makes those moves stick.
type form struct {
	title  string
	fields []field
	focus  int
	errs   *domain.ValidationError
}

func newForm(title string, specs []fieldSpec) *form {
	fields := make([]field, 0, len(specs))
	for _, spec := range specs {
		f := field{spec: spec}
		if !f.isChoice() {
			in := textinput.New()
			in.Placeholder = spec.Placeholder
			in.SetValue(spec.Prefill)
			in.CharLimit = fieldCharLimit
			in.Width = fieldWidth
			f.input = in
		}
		fields = append(fields, f)
	}

	f := &form{title: title, fields: fields}
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

// Update handles one key by mutating the form in place. A form is open state,
// not a snapshot: the tab that opened it holds the only pointer and hands that
// same pointer on until it closes the form.
func (f *form) Update(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "tab", "down":
		f.focus = wrap(f.focus, 1, len(f.fields))
		f.applyFocus()
		return nil
	case "shift+tab", "up":
		f.focus = wrap(f.focus, -1, len(f.fields))
		f.applyFocus()
		return nil
	case "left":
		if f.fields[f.focus].isChoice() {
			f.cycle(-1)
			return nil
		}
	case "right":
		if f.fields[f.focus].isChoice() {
			f.cycle(1)
			return nil
		}
	}

	if f.fields[f.focus].isChoice() {
		return nil
	}

	var cmd tea.Cmd
	f.fields[f.focus].input, cmd = f.fields[f.focus].input.Update(msg)
	return cmd
}

func (f *form) cycle(step int) {
	current := &f.fields[f.focus]
	current.choice = wrap(current.choice, step, len(current.spec.Choices))
}

// Value reports what the operator typed or chose for a field. The form does no
// parsing and no validation; a use case decides whether the text is acceptable.
func (f *form) Value(key string) string {
	for _, fld := range f.fields {
		if fld.spec.Key != key {
			continue
		}
		if fld.isChoice() {
			return fld.spec.Choices[fld.choice]
		}
		return fld.input.Value()
	}
	return ""
}

func (f *form) SetErrors(errs *domain.ValidationError) { f.errs = errs }

// Help lists the keys the form owns, naming each choice field so the operator
// knows what left/right moves. The tab appends nothing to it: an open form
// captures every key except the shell's quit, so none of the shell's keys are
// live.
func (f *form) Help() string {
	keys := []string{"tab/shift-tab next field"}
	for _, fld := range f.fields {
		if fld.isChoice() {
			keys = append(keys, "left/right "+strings.ToLower(fld.spec.Label))
		}
	}
	keys = append(keys, "enter save", "esc cancel", "ctrl+c quit")
	return strings.Join(keys, " · ")
}

func (f *form) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(f.title))
	b.WriteString("\n\n")
	for i, fld := range f.fields {
		b.WriteString(f.row(i, fld))
	}
	return b.String()
}

func (f *form) row(index int, fld field) string {
	style := labelStyle
	if f.focus == index {
		style = focusedLabelStyle
	}
	line := style.Render(fld.spec.Label) + f.control(fld)
	if msg := f.errs.For(fld.spec.Key); msg != "" {
		line += "  " + errorStyle.Render("← "+msg)
	}
	return line + "\n"
}

func (f *form) control(fld field) string {
	if !fld.isChoice() {
		return fld.input.View()
	}
	rendered := make([]string, 0, len(fld.spec.Choices))
	for i, choice := range fld.spec.Choices {
		style := inactiveTabStyle
		if i == fld.choice {
			style = activeTabStyle
		}
		rendered = append(rendered, style.Render(choice))
	}
	return fmt.Sprintf("%-*s", fieldWidth, lipgloss.JoinHorizontal(lipgloss.Top, rendered...))
}
