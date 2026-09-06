package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

const dateLabel = "Date"

const (
	labelWidth     = 24
	fieldWidth     = 24
	fieldCharLimit = 32

	// A choice row may spread over twice the column a text row gets before it
	// is worth collapsing (ADR-0024): the options are the row's whole content,
	// where a text row's column is only where its answer is typed.
	inlineChoiceWidth = 2 * fieldWidth
)

var (
	labelStyle        = lipgloss.NewStyle().Width(labelWidth)
	focusedLabelStyle = labelStyle.Bold(true).Foreground(lipgloss.Color("62"))
)

// fieldSpec declares one row of a form. A tab supplies the keys, labels and
// kinds; the form owns focus, layout, choice cycling and error placement. Key
// must be a domain.Field* constant, so a use case's ValidationError lands on
// the right row. There is one spec type per kind of row, so a row cannot carry
// a placeholder and a choice list at once, and both kinds sit in one list.
type fieldSpec interface {
	newField() field
}

type textSpec struct {
	Key         string
	Label       string
	Placeholder string
	Prefill     string
}

type choiceSpec struct {
	Key     string
	Label   string
	Choices []string
	Choice  string
}

// field is one open row. The form holds pointers and mutates them in place;
// only the row itself knows what a key means once the form has taken the keys
// that move between rows.
type field interface {
	key() string
	label() string
	control() string
	value() string
	setFocus(bool)
	update(msg tea.KeyMsg) (tea.Cmd, bool)
}

type textField struct {
	spec    textSpec
	input   textinput.Model
	touched bool
}

type choiceField struct {
	spec   choiceSpec
	choice int
}

func (s textSpec) newField() field {
	in := textinput.New()
	in.Placeholder = s.Placeholder
	in.SetValue(s.Prefill)
	in.CharLimit = fieldCharLimit
	in.Width = fieldWidth
	return &textField{spec: s, input: in}
}

func (s choiceSpec) newField() field {
	f := &choiceField{spec: s}
	for i, choice := range s.Choices {
		if choice == s.Choice {
			f.choice = i
		}
	}
	return f
}

func (f *textField) key() string   { return f.spec.Key }
func (f *choiceField) key() string { return f.spec.Key }

func (f *textField) label() string   { return f.spec.Label }
func (f *choiceField) label() string { return f.spec.Label }

func (f *textField) value() string { return f.input.Value() }

func (f *choiceField) value() string {
	if len(f.spec.Choices) == 0 {
		return ""
	}
	return f.spec.Choices[f.choice]
}

func (f *textField) setFocus(on bool) {
	if on {
		f.input.Focus()
		return
	}
	f.input.Blur()
}

func (f *choiceField) setFocus(bool) {}

func (f *textField) update(msg tea.KeyMsg) (tea.Cmd, bool) {
	before := f.input.Value()
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	f.touched = f.touched || f.input.Value() != before
	return cmd, true
}

func (f *choiceField) update(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case KeyLeft, KeyH:
		f.cycle(-1)
	case KeyRight, KeyL:
		f.cycle(1)
	}
	return nil, true
}

func (f *choiceField) picker() *picker {
	return newPicker(f.spec.Label, f.spec.Choices, f.choice)
}

// commit takes what a picker chose. The picker was opened on the choices as
// they were, so a row that has since been rebuilt on a shorter list keeps what
// it had rather than naming a record off the end of it.
func (f *choiceField) commit(index int) {
	if index < 0 || index >= len(f.spec.Choices) {
		return
	}
	f.choice = index
}

func (f *choiceField) selected() int { return f.choice }

func (f *choiceField) cycle(step int) {
	if len(f.spec.Choices) == 0 {
		return
	}
	f.choice = wrap(f.choice, step, len(f.spec.Choices))
}

// prefill writes a derived value into a row the operator has not typed into,
// so a changed quantity moves an estimate but never overwrites an actual.
func (f *textField) prefill(value string) {
	if f.touched {
		return
	}
	f.input.SetValue(value)
	// SetValue only clamps the cursor, so without this a backspace deletes
	// from wherever the last, shorter, prefill ended.
	f.input.CursorEnd()
}

func (f *textField) control() string { return f.input.View() }

func (f *choiceField) control() string {
	if !f.collapsed() {
		return pad(f.inline(), fieldWidth)
	}
	// Less the two columns activeTabStyle pads with, so a long label still
	// ends inside the budget an inline row would have had.
	current := activeTabStyle.Render(truncate(f.value(), inlineChoiceWidth-2))
	return pad(current, fieldWidth)
}

// collapsed hands the picker a choice row too wide to sit on one line
// (ADR-0024). Two Spool labels already are.
func (f *choiceField) collapsed() bool {
	return lipgloss.Width(f.inline()) > inlineChoiceWidth
}

func (f *choiceField) inline() string {
	rendered := make([]string, 0, len(f.spec.Choices))
	for i, choice := range f.spec.Choices {
		style := inactiveTabStyle
		if i == f.choice {
			style = activeTabStyle
		}
		rendered = append(rendered, style.Render(choice))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

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

// ChoiceIndex locates a choice by position rather than by label, so two rows
// that read the same still name different records.
func (f *form) ChoiceIndex(key string) int {
	for _, fld := range f.fields {
		choice, ok := fld.(*choiceField)
		if ok && choice.key() == key {
			return choice.selected()
		}
	}
	return -1
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
