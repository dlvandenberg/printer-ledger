package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// confirm is an open delete prompt. Nothing is deleted without one (ADR-0011).
// The tab owns what a yes runs and what to reload afterwards; the prompt owns
// the question and the keys that answer it.
type confirm struct {
	question string
	run      func() error
}

func newConfirm(question string, run func() error) *confirm {
	return &confirm{question: question, run: run}
}

// Answer returns the prompt still to answer, or nil once it has been answered
// or cancelled, together with whether the delete ran and what it failed with.
func (c *confirm) Answer(msg tea.KeyMsg) (*confirm, bool, error) {
	switch msg.String() {
	case KeyY:
		return nil, true, c.run()
	case KeyN, KeyEsc:
		return nil, false, nil
	}
	return c, false, nil
}

func (c *confirm) View() string { return warningStyle.Render(c.question) }

func (c *confirm) Help() string {
	return fmt.Sprintf("%s delete · %s cancel · %s quit", KeyY, KeyN, KeyCtrlC)
}
