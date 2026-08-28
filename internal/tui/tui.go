package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvandenberg/printer-ledger/internal/app"
)

const currency = "€"

// globalHelp lists the keys the shell owns. A tab appends its own keys in front
// of it.
const globalHelp = "tab/shift-tab switch tabs · q quit"

var (
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("62")).Padding(0, 1)
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 1)
	titleStyle       = lipgloss.NewStyle().Bold(true)
	helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).PaddingTop(1)
)

type Model struct {
	tabs   []namedTab
	active int
	width  int
}

func Run(a *app.App) error {
	spools, err := newSpoolsModel(a)
	if err != nil {
		return err
	}
	m := Model{tabs: []namedTab{
		{name: "Spools", model: spools},
		newPlaceholder("Designs"),
		newPlaceholder("Prints"),
		newPlaceholder("Sales"),
		newPlaceholder("Report"),
		newPlaceholder("Settings"),
	}}
	_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		// Checked before the capture branch: a capturing tab swallows every
		// other key, and the operator must always be able to quit.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if !m.current().capturesInput() {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "tab":
				m.active = (m.active + 1) % len(m.tabs)
				return m, nil
			case "shift+tab":
				m.active = (m.active + len(m.tabs) - 1) % len(m.tabs)
				return m, nil
			}
		}

		return m.routeKey(msg)
	}
	return m, nil
}

// routeKey hands the key to the active tab and puts the tab it returns back in
// place. The slice is cloned so the returned Model does not share its backing
// array with the caller's.
func (m Model) routeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	updated, cmd := m.current().update(msg)
	tabs := make([]namedTab, len(m.tabs))
	copy(tabs, m.tabs)
	tabs[m.active].model = updated
	m.tabs = tabs
	return m, cmd
}

func (m Model) current() tabModel { return m.tabs[m.active].model }

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.tabBar())
	b.WriteString("\n\n")
	b.WriteString(m.current().view())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(m.current().help()))
	return b.String()
}

func (m Model) tabBar() string {
	rendered := make([]string, 0, len(m.tabs))
	for i, t := range m.tabs {
		style := inactiveTabStyle
		if i == m.active {
			style = activeTabStyle
		}
		rendered = append(rendered, style.Render(t.name))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}
