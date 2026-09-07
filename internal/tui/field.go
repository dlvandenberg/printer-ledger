package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	fieldWidth     = 24
	fieldCharLimit = 32

	// A choice row may spread over twice the column a text row gets before it
	// is worth collapsing (ADR-0024): the options are the row's whole content,
	// where a text row's column is only where its answer is typed.
	inlineChoiceWidth = 2 * fieldWidth
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
	Key      string
	Label    string
	Choices  []choice
	Selected choice
}

// choice is one option on a choice row: what the operator reads, and what it
// names. A row over records names them by ID; a row over values, such as
// Filament Type, names its own label and leaves ID zero.
type choice struct {
	Label string
	ID    int64
}

func choicesOf[T any](items []T, of func(T) choice) []choice {
	choices := make([]choice, 0, len(items))
	for _, item := range items {
		choices = append(choices, of(item))
	}
	return choices
}

// choiceFor is the choice naming one record, or the zero choice when the list
// no longer offers it — a row prefilled with that opens on its first option.
func choiceFor(choices []choice, id int64) choice {
	for _, c := range choices {
		if c.ID == id {
			return c
		}
	}
	return choice{}
}

func valueChoice(label string) choice { return choice{Label: label} }

func valueChoices(labels []string) []choice { return choicesOf(labels, valueChoice) }

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
	for i, c := range s.Choices {
		if c == s.Selected {
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

func (f *choiceField) value() string { return f.selection().Label }

func (f *choiceField) selection() choice {
	if len(f.spec.Choices) == 0 {
		return choice{}
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
	return newPicker(f.spec.Label, f.labels(), f.choice)
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
	for i, c := range f.spec.Choices {
		style := inactiveTabStyle
		if i == f.choice {
			style = activeTabStyle
		}
		rendered = append(rendered, style.Render(c.Label))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (f *choiceField) labels() []string {
	labels := make([]string, 0, len(f.spec.Choices))
	for _, c := range f.spec.Choices {
		labels = append(labels, c.Label)
	}
	return labels
}
