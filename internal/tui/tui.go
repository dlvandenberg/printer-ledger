package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvandenberg/printer-ledger/internal/app"
)

const currency = "€"

const globalHelp = "tab/shift-tab switch tabs · q quit"

var (
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("62")).Padding(0, 1)
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 1)
	titleStyle       = lipgloss.NewStyle().Bold(true)
	helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).PaddingTop(1)
)

func wrap(i, delta, n int) int { return ((i+delta)%n + n) % n }

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
	settings, err := newSettingsModel(a)
	if err != nil {
		return err
	}
	m := Model{tabs: []namedTab{
		{name: "Spools", model: spools},
		newPlaceholder("Designs"),
		newPlaceholder("Prints"),
		newPlaceholder("Sales"),
		newPlaceholder("Report"),
		{name: "Settings", model: settings},
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

		if !m.current().CapturesInput() {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "tab":
				m.active = wrap(m.active, 1, len(m.tabs))
				return m, nil
			case "shift+tab":
				m.active = wrap(m.active, -1, len(m.tabs))
				return m, nil
			}
		}

		return m.routeKey(msg)
	}
	return m, nil
}

// The returned tab is written into the shared backing array, so it sticks even
// though bubbletea keeps only the Model returned from Update.
func (m Model) routeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	updated, cmd := m.current().Update(msg)
	m.tabs[m.active].model = updated
	return m, cmd
}

func (m Model) current() tabModel { return m.tabs[m.active].model }

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.tabBar())
	b.WriteString("\n\n")
	b.WriteString(m.current().View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(m.current().Help()))
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
