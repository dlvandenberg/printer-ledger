package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type namedTab struct {
	name  string
	model tabModel
}

// tabModel is what the shell knows about a tab. The shell routes keys to
// whichever tab is active without asking which one that is; a tab that ignores
// a key returns itself.
type tabModel interface {
	Update(msg tea.KeyMsg) (tabModel, tea.Cmd)
	// Refresh re-reads what the tab shows. The shell calls it when the tab
	// becomes active, because a use case run on one tab can change what
	// another tab is already displaying.
	Refresh() tabModel
	View() string
	Help() string
	// CapturesInput true means the tab receives every key except the shell's quit.
	CapturesInput() bool
}
