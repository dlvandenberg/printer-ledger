package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvandenberg/printer-ledger/internal/app"
)

const currency = "€"

type tab int

const (
	tabSpools tab = iota
	tabDesigns
	tabPrints
	tabSales
	tabReport
	tabSettings
)

var tabNames = map[tab]string{
	tabSpools:   "Spools",
	tabDesigns:  "Designs",
	tabPrints:   "Prints",
	tabSales:    "Sales",
	tabReport:   "Report",
	tabSettings: "Settings",
}

var tabOrder = []tab{tabSpools, tabDesigns, tabPrints, tabSales, tabReport, tabSettings}

var (
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("62")).Padding(0, 1)
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 1)
	titleStyle       = lipgloss.NewStyle().Bold(true)
	helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).PaddingTop(1)
)

type Model struct {
	active tab
	spools spoolsModel
	width  int
}

func Run(a *app.App) error {
	spools, err := newSpoolsModel(a)
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(Model{spools: spools}, tea.WithAltScreen()).Run()
	return err
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		// Checked before the capture branch: an open form swallows every other
		// key, and the operator must always be able to quit.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.active == tabSpools && m.spools.capturesInput() {
			spools, cmd := m.spools.update(msg)
			m.spools = spools
			return m, cmd
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "tab":
			m.active = tabOrder[(int(m.active)+1)%len(tabOrder)]
			return m, nil
		case "shift+tab":
			m.active = tabOrder[(int(m.active)+len(tabOrder)-1)%len(tabOrder)]
			return m, nil
		}

		if m.active == tabSpools {
			spools, cmd := m.spools.update(msg)
			m.spools = spools
			return m, cmd
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.tabBar())
	b.WriteString("\n\n")
	b.WriteString(m.body())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(m.help()))
	return b.String()
}

func (m Model) tabBar() string {
	rendered := make([]string, 0, len(tabOrder))
	for _, t := range tabOrder {
		style := inactiveTabStyle
		if t == m.active {
			style = activeTabStyle
		}
		rendered = append(rendered, style.Render(tabNames[t]))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m Model) body() string {
	if m.active == tabSpools {
		return m.spools.view()
	}
	return titleStyle.Render(tabNames[m.active]) + "\n" +
		placeholderStyle.Render("Nothing here yet.")
}

func (m Model) help() string {
	if m.active == tabSpools {
		return m.spools.help()
	}
	return "tab/shift-tab switch tabs · q quit"
}
