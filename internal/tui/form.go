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
// kinds; the form owns focus, layout, choice cycling and error placement. Key
// must be a domain.Field* constant, so a use case's ValidationError lands on
// the right row. A non-empty Choices makes the row a choice field; the rest are
// text fields, and choice and text rows sit in one list.
type fieldSpec struct {
	Key         string
	Label       string
	Choices     []string
	Placeholder string
	Prefill     string
	Choice      string
}

type field struct {
	spec   fieldSpec
	input  textinput.Model
	choice int
}

func (f field) isChoice() bool { return len(f.spec.Choices) > 0 }

// form is mutable open state, always held as a *form and mutated in place. Tab
// models are values that bubbletea replaces on every key, so a form copied
// along with one would lose focus and choice moves; the owning tab holds the
// only pointer, from open until close.
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
		} else {
			for i, choice := range spec.Choices {
				if choice == spec.Choice {
					f.choice = i
				}
			}
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

func (f *form) Update(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case KeyTab, KeyDown:
		f.focus = wrap(f.focus, 1, len(f.fields))
		f.applyFocus()
		return nil
	case KeyShiftTab, KeyUp:
		f.focus = wrap(f.focus, -1, len(f.fields))
		f.applyFocus()
		return nil
	case KeyLeft, KeyH:
		if f.fields[f.focus].isChoice() {
			f.cycle(-1)
			return nil
		}
	case KeyRight, KeyL:
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

func (f *form) Help() string {
	keys := []string{fmt.Sprintf("%s/%s next/prev field", KeyTab, KeyShiftTab)}
	for _, fld := range f.fields {
		if fld.isChoice() {
			keys = append(keys, fmt.Sprintf("%s/%s (or %s/%s) %s", KeyLeft, KeyRight, KeyH, KeyL, strings.ToLower(fld.spec.Label)))
		}
	}
	keys = append(keys, fmt.Sprintf("%s save · %s cancel · %s quit", KeyEnter, KeyEsc, KeyCtrlC))
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
