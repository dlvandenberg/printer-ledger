package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	inlineChoiceWidth = 48
	pickerWindow      = 10
)

// picker is the open filter over one choice field (ADR-0024). It keeps no
// selection of its own beyond a cursor: what it commits is an index into the
// Choices it was opened on, so two rows that read the same still name
// different records.
type picker struct {
	title   string
	choices []string
	filter  string
	cursor  int // into matches, not into choices
}

func newPicker(label string, choices []string, current int) *picker {
	return &picker{title: "Select " + label, choices: choices, cursor: current}
}

func matchChoices(choices []string, filter string) []int {
	needle := strings.ToLower(filter)
	matches := make([]int, 0, len(choices))
	for i, choice := range choices {
		if strings.Contains(strings.ToLower(choice), needle) {
			matches = append(matches, i)
		}
	}
	return matches
}

// Answer returns the picker still open, or nil once it is done. A picker
// closed with esc picks nothing and the field keeps what it had.
func (p *picker) Answer(msg tea.KeyMsg) (open *picker, choice int, picked bool) {
	matches := p.matches()

	switch msg.String() {
	case KeyEsc:
		return nil, 0, false
	case KeyEnter:
		if len(matches) == 0 {
			return p, 0, false
		}
		return nil, matches[p.cursor], true
	case KeyUp, KeyCtrlP:
		p.move(matches, -1)
		return p, 0, false
	case KeyDown, KeyCtrlN:
		p.move(matches, 1)
		return p, 0, false
	}

	before := p.filter
	switch {
	case msg.Alt:
	case msg.Type == tea.KeyBackspace:
		p.filter = dropLastRune(p.filter)
	case msg.Type == tea.KeyRunes, msg.Type == tea.KeySpace:
		p.filter += msg.String()
	}
	if p.filter != before {
		p.cursor = 0
	}
	return p, 0, false
}

func (p *picker) matches() []int { return matchChoices(p.choices, p.filter) }

func (p *picker) move(matches []int, step int) {
	if len(matches) == 0 {
		return
	}
	p.cursor = wrap(p.cursor, step, len(matches))
}

func (p *picker) Help() string {
	return fmt.Sprintf("%s/%s (or %s/%s) move · type to filter · %s select · %s back · %s quit",
		KeyUp, KeyDown, KeyCtrlP, KeyCtrlN, KeyEnter, KeyEsc, KeyCtrlC)
}

func (p *picker) View() string {
	matches := p.matches()

	var b strings.Builder
	b.WriteString(titleStyle.Render(p.title))
	b.WriteString("\n\n> " + p.filter + "\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf("%d of %d", len(matches), len(p.choices))))
	b.WriteString("\n\n")

	if len(matches) == 0 {
		b.WriteString(helpStyle.Render("nothing matches — backspace to widen"))
		return b.String() + "\n"
	}

	start := p.window(len(matches))
	for offset, index := range matches[start:min(start+pickerWindow, len(matches))] {
		style, marker := plainStyle, "  "
		if start+offset == p.cursor {
			style, marker = activeTabStyle, "› "
		}
		b.WriteString(marker + style.Render(p.choices[index]) + "\n")
	}
	return b.String()
}

// window is the first row on screen: the cursor stays inside it, and a list
// shorter than the window never scrolls.
func (p *picker) window(count int) int {
	if p.cursor < pickerWindow {
		return 0
	}
	return min(p.cursor-pickerWindow+1, max(0, count-pickerWindow))
}
