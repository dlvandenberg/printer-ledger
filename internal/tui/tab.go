package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// tabModel is what the shell knows about a tab. The shell routes keys to
// whichever tab is active without asking which one that is; a tab that ignores
// a key returns itself.
type tabModel interface {
	Update(msg tea.KeyMsg) (tabModel, tea.Cmd)
	View() string
	Help() string
	// CapturesInput true means the tab receives every key except the shell's quit.
	CapturesInput() bool
}

type namedTab struct {
	name  string
	model tabModel
}

var _ tabModel = placeholderTab{}

type placeholderTab struct {
	title string
}

func newPlaceholder(name string) namedTab {
	return namedTab{name: name, model: placeholderTab{title: name}}
}

func (t placeholderTab) Update(tea.KeyMsg) (tabModel, tea.Cmd) { return t, nil }

func (t placeholderTab) View() string {
	return titleStyle.Render(t.title) + "\n" + placeholderStyle.Render("Nothing here yet.")
}

func (t placeholderTab) Help() string { return globalHelp }

func (t placeholderTab) CapturesInput() bool { return false }
