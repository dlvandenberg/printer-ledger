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
	spec    fieldSpec
	input   textinput.Model
	choice  int
	touched bool
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
		fields = append(fields, newField(spec))
	}

	f := &form{title: title, fields: fields}
	f.applyFocus()
	return f
}

func newField(spec fieldSpec) field {
	f := field{spec: spec}
	if !f.isChoice() {
		in := textinput.New()
		in.Placeholder = spec.Placeholder
		in.SetValue(spec.Prefill)
		in.CharLimit = fieldCharLimit
		in.Width = fieldWidth
		f.input = in
		return f
	}
	for i, choice := range spec.Choices {
		if choice == spec.Choice {
			f.choice = i
		}
	}
	return f
}

// Append and DropLast let a tab grow and shrink a group of trailing rows — one
// Filament Usage row per Spool a Print drew from (ADR-0012). The tab owns how
// many rows there are; the form still owns focus and layout.
func (f *form) Append(specs ...fieldSpec) {
	for _, spec := range specs {
		f.fields = append(f.fields, newField(spec))
	}
	f.applyFocus()
}

func (f *form) DropLast(count int) {
	f.fields = f.fields[:max(1, len(f.fields)-count)]
	f.focus = min(f.focus, len(f.fields)-1)
	f.applyFocus()
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

	before := f.fields[f.focus].input.Value()
	var cmd tea.Cmd
	f.fields[f.focus].input, cmd = f.fields[f.focus].input.Update(msg)
	f.fields[f.focus].touched = f.fields[f.focus].touched || f.fields[f.focus].input.Value() != before
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

// Prefill writes a derived value into a field the operator has not typed into,
// so a changed quantity moves an estimate but never overwrites an actual.
func (f *form) Prefill(key, value string) {
	for i := range f.fields {
		if f.fields[i].spec.Key == key && !f.fields[i].isChoice() && !f.fields[i].touched {
			f.fields[i].input.SetValue(value)
			// SetValue only clamps the cursor, so without this a backspace
			// deletes from wherever the last, shorter, prefill ended.
			f.fields[i].input.CursorEnd()
		}
	}
}

// ChoiceIndex locates a choice by position rather than by label, so two rows
// that read the same still name different records.
func (f *form) ChoiceIndex(key string) int {
	for _, fld := range f.fields {
		if fld.spec.Key == key {
			return fld.choice
		}
	}
	return -1
}

func (f *form) SetErrors(errs *domain.ValidationError) { f.errs = errs }

func (f *form) Help() string {
	keys := []string{fmt.Sprintf("%s/%s next/prev field", KeyTab, KeyShiftTab)}
	if f.hasChoiceField() {
		keys = append(keys, fmt.Sprintf("%s/%s (or %s/%s) change choice", KeyLeft, KeyRight, KeyH, KeyL))
	}
	keys = append(keys, fmt.Sprintf("%s save · %s cancel · %s quit", KeyEnter, KeyEsc, KeyCtrlC))
	return strings.Join(keys, " · ")
}

func (f *form) hasChoiceField() bool {
	for _, fld := range f.fields {
		if fld.isChoice() {
			return true
		}
	}
	return false
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
