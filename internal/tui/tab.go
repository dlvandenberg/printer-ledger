package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// tabModel is what the shell knows about a tab: it handles a key, renders a
// body and a help line, and says whether it is currently capturing input. The
// shell routes to whichever tab is active without asking which one that is.
type tabModel interface {
	// update handles one key. A tab that ignores the key returns itself.
	update(msg tea.KeyMsg) (tabModel, tea.Cmd)
	view() string
	help() string
	// capturesInput reports that the tab is mid-interaction — an open form,
	// say — and must receive every key except the shell's quit.
	capturesInput() bool
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

func (t placeholderTab) update(tea.KeyMsg) (tabModel, tea.Cmd) { return t, nil }

func (t placeholderTab) view() string {
	return titleStyle.Render(t.title) + "\n" + placeholderStyle.Render("Nothing here yet.")
}

func (t placeholderTab) help() string { return globalHelp }

func (t placeholderTab) capturesInput() bool { return false }
