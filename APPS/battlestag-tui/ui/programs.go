package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"battlestag/client"
)

// ProgramsModel represents the programs list screen
type ProgramsModel struct {
	client   *client.Client
	programs []client.ProgramInfo
	loading  bool
	err      error
	width    int
	height   int
	scroll   int
}

// NewProgramsModel creates a new programs model
func NewProgramsModel(c *client.Client) *ProgramsModel {
	return &ProgramsModel{
		client:  c,
		loading: true,
	}
}

// Init initializes the programs screen
func (m *ProgramsModel) Init() tea.Cmd {
	return m.fetchPrograms()
}

// fetchPrograms fetches the program list from the gateway
func (m *ProgramsModel) fetchPrograms() tea.Cmd {
	return func() tea.Msg {
		programs, err := m.client.ListPrograms()
		if err != nil {
			return ProgramsFetchedMsg{err: err}
		}
		return ProgramsFetchedMsg{programs: programs}
	}
}

// Update handles programs screen messages
func (m ProgramsModel) Update(msg tea.Msg) (ProgramsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgramsFetchedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.programs = msg.programs
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.scroll > 0 {
				m.scroll--
			}
		case "down", "j":
			m.scroll++
		case "pgup":
			m.scroll -= 10
			if m.scroll < 0 {
				m.scroll = 0
			}
		case "pgdown":
			m.scroll += 10
		case "home":
			m.scroll = 0
		}
	}

	return m, nil
}

// View renders the programs screen
func (m *ProgramsModel) View() string {
	title := titleStyle.Render("📦 Available Programs")

	if m.loading {
		loading := loadingStyle.Render("Loading programs...")
		content := title + "\n\n" + loading
		return boxStyle.Render(content)
	}

	if m.err != nil {
		errorText := errorStyle.Render(fmt.Sprintf("Error loading programs: %v", m.err))
		content := title + "\n\n" + errorText
		help := helpStyle.Render("\n\n[Esc] Back to chat")
		return boxStyle.Render(content + help)
	}

	if len(m.programs) == 0 {
		empty := infoStyle.Render("No programs available")
		content := title + "\n\n" + empty
		help := helpStyle.Render("\n\n[Esc] Back to chat")
		return boxStyle.Render(content + help)
	}

	var programList []string
	programList = append(programList, title)
	programList = append(programList, "")

	for _, program := range m.programs {
		// Program name and description
		programHeader := lipgloss.NewStyle().
			Foreground(lipgloss.Color("63")).
			Bold(true).
			Render(fmt.Sprintf("• %s", program.Name))

		programDesc := lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Render(fmt.Sprintf("  %s", program.Description))

		programPath := lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")).
			Italic(true).
			Render(fmt.Sprintf("  Path: %s", program.Executable))

		programList = append(programList, programHeader)
		programList = append(programList, programDesc)
		programList = append(programList, programPath)
		programList = append(programList, "")
	}

	// Summary
	summary := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Render(fmt.Sprintf("\nTotal programs: %d", len(m.programs)))
	programList = append(programList, summary)

	// Help text
	help := helpStyle.Render("\n\n[↑↓/j/k] Scroll • [PgUp/PgDn] Page • [Home] Top • [Esc] Back to chat")
	programList = append(programList, help)

	content := strings.Join(programList, "\n")

	// Apply scroll
	lines := strings.Split(content, "\n")
	if m.scroll > 0 && m.scroll < len(lines) {
		visibleHeight := m.height - 4
		if visibleHeight < 10 {
			visibleHeight = 10
		}
		end := m.scroll + visibleHeight
		if end > len(lines) {
			end = len(lines)
		}
		content = strings.Join(lines[m.scroll:end], "\n")
	}

	return boxStyle.Width(m.width - 4).Render(content)
}

// ProgramsFetchedMsg indicates programs have been fetched
type ProgramsFetchedMsg struct {
	programs []client.ProgramInfo
	err      error
}
