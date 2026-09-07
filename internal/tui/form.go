package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

const dateLabel = "Date"

const labelWidth = 24

var (
	labelStyle        = lipgloss.NewStyle().Width(labelWidth)
	focusedLabelStyle = labelStyle.Bold(true).Foreground(lipgloss.Color("62"))
)

// form is mutable open state, always held as a *form and mutated in place. Tab
// models are values that bubbletea replaces on every key, so a form copied
// along with one would lose focus and choice moves; the owning tab holds the
// only pointer, from open until close.
type form struct {
	title  string
	fields []field
	focus  int
	errs   *domain.ValidationError
	pick   *picker
	chords []chord
}

// chord is a key the owning tab acts on itself. The form hands it back rather
// than feeding it to the focused field, and shows it in the help line.
type chord struct {
	key   string
	label string
}

func newForm(title string, specs []fieldSpec) *form {
	fields := make([]field, 0, len(specs))
	for _, spec := range specs {
		fields = append(fields, spec.newField())
	}

	f := &form{title: title, fields: fields}
	f.applyFocus()
	return f
}

// Append and DropLast let a tab grow and shrink a group of trailing rows — one
// Filament Usage row per Spool a Print drew from (ADR-0012). The tab owns how
// many rows there are; the form still owns focus and layout.
func (f *form) Append(specs ...fieldSpec) {
	for _, spec := range specs {
		f.fields = append(f.fields, spec.newField())
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
		f.fields[i].setFocus(f.focus == i)
	}
}

// Reserve names a chord the owning tab acts on itself, so a tab never has to
// ask what the form wants before handing it a key.
func (f *form) Reserve(key, label string) {
	f.chords = append(f.chords, chord{key: key, label: label})
}

func (f *form) PickerOpen() bool { return f.pick != nil }

func (f *form) reserves(key string) bool {
	for _, c := range f.chords {
		if c.key == key {
			return true
		}
	}
	return false
}

func (f *form) focused() field { return f.fields[f.focus] }

func (f *form) focusedChoice() (*choiceField, bool) {
	fld, ok := f.focused().(*choiceField)
	return fld, ok
}

func (f *form) openPicker() {
	if fld, ok := f.focusedChoice(); ok {
		f.pick = fld.picker()
	}
}

func (f *form) updatePicker(msg tea.KeyMsg) {
	open, choice, picked := f.pick.Answer(msg)
	f.pick = open
	if fld, ok := f.focusedChoice(); ok && picked {
		fld.commit(choice)
	}
}

// Update reports whether the form consumed the key. The picker owns every key
// while it is open and the enter that opens it (ADR-0024); esc and a Reserve-d
// chord always come back, so the tab acts on what the form refuses rather than
// asking first.
func (f *form) Update(msg tea.KeyMsg) (tea.Cmd, bool) {
	if f.pick != nil {
		f.updatePicker(msg)
		return nil, true
	}

	switch msg.String() {
	case KeyEsc:
		return nil, false
	case KeyEnter:
		fld, ok := f.focusedChoice()
		if !ok || !fld.collapsed() {
			return nil, false
		}
		f.openPicker()
		return nil, true
	case KeyTab, KeyDown:
		f.focus = wrap(f.focus, 1, len(f.fields))
		f.applyFocus()
		return nil, true
	case KeyShiftTab, KeyUp:
		f.focus = wrap(f.focus, -1, len(f.fields))
		f.applyFocus()
		return nil, true
	}

	if f.reserves(msg.String()) {
		return nil, false
	}
	return f.focused().update(msg)
}

func (f *form) Value(key string) string {
	for _, fld := range f.fields {
		if fld.key() == key {
			return fld.value()
		}
	}
	return ""
}

func (f *form) Prefill(key, value string) {
	for _, fld := range f.fields {
		text, ok := fld.(*textField)
		if ok && text.key() == key {
			text.prefill(value)
		}
	}
}

// Choice is what a choice row names, so a caller writes an id straight into a
// command rather than indexing back into the list the row was built from.
func (f *form) Choice(key string) choice {
	for _, fld := range f.fields {
		c, ok := fld.(*choiceField)
		if ok && c.key() == key {
			return c.selection()
		}
	}
	return choice{}
}

func (f *form) SetErrors(errs *domain.ValidationError) { f.errs = errs }

func (f *form) Help() string {
	if f.pick != nil {
		return f.pick.Help()
	}
	keys := []string{fmt.Sprintf("%s/%s next/prev field", KeyTab, KeyShiftTab)}
	if f.hasChoiceField() {
		keys = append(keys, fmt.Sprintf("%s/%s (or %s/%s) change choice", KeyLeft, KeyRight, KeyH, KeyL))
	}
	if f.hasCollapsedField() {
		keys = append(keys, fmt.Sprintf("%s pick from list", KeyEnter))
	}
	keys = append(keys, fmt.Sprintf("%s save · %s cancel · %s quit", KeyEnter, KeyEsc, KeyCtrlC))
	for _, c := range f.chords {
		keys = append(keys, c.key+" "+c.label)
	}
	return strings.Join(keys, " · ")
}

func (f *form) hasChoiceField() bool {
	for _, fld := range f.fields {
		if _, ok := fld.(*choiceField); ok {
			return true
		}
	}
	return false
}

func (f *form) hasCollapsedField() bool {
	for _, fld := range f.fields {
		if choice, ok := fld.(*choiceField); ok && choice.collapsed() {
			return true
		}
	}
	return false
}

func (f *form) View() string {
	if f.pick != nil {
		return f.pick.View()
	}

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
	line := style.Render(fld.label()) + fld.control()
	if msg := f.errs.For(fld.key()); msg != "" {
		line += "  " + errorStyle.Render("← "+msg)
	}
	return line + "\n"
}
