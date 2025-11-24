package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxVisibleCommands = 6

// PaletteModel represents the command palette
type PaletteModel struct {
	registry       *CommandRegistry
	commands       []Command // Filtered commands to display
	selected       int       // Selected index
	scrollOffset   int       // Scroll offset for long lists
	currentPath    string    // Current navigation path (e.g., "finance")
	inputFilter    string    // Current input filter
	width          int
	visible        bool
}

// NewPaletteModel creates a new command palette
func NewPaletteModel() *PaletteModel {
	registry := NewCommandRegistry()
	return &PaletteModel{
		registry: registry,
		commands: registry.GetRootCommands(),
		selected: 0,
		visible:  false,
	}
}

// Show shows the palette with optional initial filter
func (m *PaletteModel) Show(filter string) {
	m.visible = true
	m.inputFilter = filter
	m.selected = 0
	m.scrollOffset = 0
	m.updateCommands()
}

// Hide hides the palette
func (m *PaletteModel) Hide() {
	m.visible = false
	m.currentPath = ""
	m.inputFilter = ""
	m.selected = 0
	m.scrollOffset = 0
}

// IsVisible returns whether the palette is visible
func (m *PaletteModel) IsVisible() bool {
	return m.visible
}

// SetFilter updates the filter and refreshes commands
func (m *PaletteModel) SetFilter(filter string) {
	m.inputFilter = filter
	m.selected = 0
	m.scrollOffset = 0
	m.updateCommands()
}

// updateCommands refreshes the command list based on current filter/path
func (m *PaletteModel) updateCommands() {
	m.commands = m.registry.FilterCommands(m.inputFilter)

	// Ensure selected index is valid
	if m.selected >= len(m.commands) {
		m.selected = len(m.commands) - 1
		if m.selected < 0 {
			m.selected = 0
		}
	}

	m.adjustScroll()
}

// adjustScroll ensures the selected item is visible
func (m *PaletteModel) adjustScroll() {
	if len(m.commands) == 0 {
		m.scrollOffset = 0
		return
	}

	// If selected is below visible area, scroll down
	if m.selected >= m.scrollOffset+maxVisibleCommands {
		m.scrollOffset = m.selected - maxVisibleCommands + 1
	}

	// If selected is above visible area, scroll up
	if m.selected < m.scrollOffset {
		m.scrollOffset = m.selected
	}

	// Ensure scroll offset doesn't go negative
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// Update handles palette messages
func (m *PaletteModel) Update(msg tea.Msg) tea.Cmd {
	if !m.visible {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.selected--
			if m.selected < 0 {
				m.selected = len(m.commands) - 1
				m.scrollOffset = max(0, len(m.commands)-maxVisibleCommands)
			} else {
				m.adjustScroll()
			}

		case "down", "j":
			m.selected++
			if m.selected >= len(m.commands) {
				m.selected = 0
				m.scrollOffset = 0
			} else {
				m.adjustScroll()
			}

		case "enter":
			if len(m.commands) > 0 && m.selected < len(m.commands) {
				selectedCmd := m.commands[m.selected]

				if selectedCmd.Type == CategoryCommand {
					// Drill into category
					m.currentPath = selectedCmd.Key
					m.inputFilter = "/" + selectedCmd.Key + ":"
					m.updateCommands()

					return func() tea.Msg {
						return PaletteFilterUpdateMsg{filter: m.inputFilter}
					}
				} else {
					// Execute action command
					m.Hide()
					return func() tea.Msg {
						return CommandExecuteMsg{command: selectedCmd.Key}
					}
				}
			}

		case "esc":
			// Go back one level or close
			if m.currentPath != "" {
				// Go back to root
				m.currentPath = ""
				m.inputFilter = "/"
				m.updateCommands()

				return func() tea.Msg {
					return PaletteFilterUpdateMsg{filter: m.inputFilter}
				}
			} else {
				// Close palette
				m.Hide()
				return func() tea.Msg {
					return PaletteCloseMsg{}
				}
			}
		}
	}

	return nil
}

// View renders the command palette
func (m *PaletteModel) View() string {
	if !m.visible {
		return ""
	}

	if len(m.commands) == 0 {
		return m.renderEmpty()
	}

	return m.renderCommands()
}

// renderEmpty renders the empty state
func (m *PaletteModel) renderEmpty() string {
	empty := paletteEmptyStyle.Render("No commands found")
	return paletteBoxStyle.Width(m.width - 4).Render(empty)
}

// renderCommands renders the command list
func (m *PaletteModel) renderCommands() string {
	var items []string

	// Determine visible range
	start := m.scrollOffset
	end := min(start+maxVisibleCommands, len(m.commands))

	// Show scroll indicator at top if there are items above
	if start > 0 {
		items = append(items, paletteScrollIndicator.Render("▲ More above..."))
	}

	// Render visible commands
	for i := start; i < end; i++ {
		cmd := m.commands[i]
		isSelected := i == m.selected

		var itemStr string
		if cmd.Type == CategoryCommand {
			// Category - show with arrow
			itemStr = fmt.Sprintf("%s →", cmd.Name)
		} else {
			// Action command - show name and description
			itemStr = fmt.Sprintf("%-20s %s", cmd.Name, paletteDescStyle.Render(cmd.Description))
		}

		if isSelected {
			items = append(items, paletteSelectedStyle.Render(itemStr))
		} else {
			items = append(items, paletteItemStyle.Render(itemStr))
		}
	}

	// Show scroll indicator at bottom if there are items below
	if end < len(m.commands) {
		items = append(items, paletteScrollIndicator.Render("▼ More below..."))
	}

	// Add help text
	helpText := paletteHelpStyle.Render("[↑↓] Navigate • [Enter] Select • [Esc] Back/Close")
	items = append(items, "", helpText)

	content := strings.Join(items, "\n")
	return paletteBoxStyle.Width(m.width - 4).Render(content)
}

// SetWidth sets the palette width
func (m *PaletteModel) SetWidth(width int) {
	m.width = width
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Custom messages

// PaletteFilterUpdateMsg indicates the filter should be updated in the prompt
type PaletteFilterUpdateMsg struct {
	filter string
}

// PaletteCloseMsg indicates the palette should close
type PaletteCloseMsg struct{}

// CommandExecuteMsg indicates a command should be executed
type CommandExecuteMsg struct {
	command string
}

// Palette styles

var (
	paletteBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")). // Purple
			Padding(1, 2).
			Background(lipgloss.Color("235")) // Dark background

	paletteItemStyle = lipgloss.NewStyle().
				Padding(0, 1)

	paletteSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(lipgloss.Color("63")). // Purple
				Foreground(lipgloss.Color("15")).  // White
				Bold(true)

	paletteDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")) // Gray

	paletteScrollIndicator = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")). // Gray
				Italic(true).
				Align(lipgloss.Center)

	paletteEmptyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")). // Gray
				Italic(true).
				Padding(2, 0)

	paletteHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")). // Gray
				Align(lipgloss.Center)
)
