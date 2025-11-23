package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PromptModel represents the always-visible input prompt
type PromptModel struct {
	textInput     textinput.Model
	width         int
	showingCommands bool // true when "/" is typed
}

// NewPromptModel creates a new prompt model
func NewPromptModel() PromptModel {
	ti := textinput.New()
	ti.Placeholder = "Type a message or / for commands..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 80

	return PromptModel{
		textInput: ti,
	}
}

// Init initializes the prompt
func (m PromptModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles prompt messages
func (m PromptModel) Update(msg tea.Msg) (PromptModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		// Adjust textinput width based on terminal width
		// Leave space for prompt indicator and padding
		m.textInput.Width = msg.Width - 10
		return m, nil

	case tea.KeyMsg:
		// Detect "/" to show command palette
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '/' {
			// If input is empty and user types "/", show command palette
			if m.textInput.Value() == "" {
				m.showingCommands = true
				return m, func() tea.Msg {
					return ShowCommandPaletteMsg{}
				}
			}
		}

		// Handle enter key - submit message
		if msg.Type == tea.KeyEnter {
			value := strings.TrimSpace(m.textInput.Value())
			if value != "" {
				// Check if it's a command (starts with /)
				if strings.HasPrefix(value, "/") {
					// It's a command
					cmd := strings.TrimPrefix(value, "/")
					m.textInput.SetValue("")
					return m, func() tea.Msg {
						return CommandSubmitMsg{command: cmd}
					}
				} else {
					// It's a chat message
					msg := value
					m.textInput.SetValue("")
					return m, func() tea.Msg {
						return ChatSubmitMsg{message: msg}
					}
				}
			}
		}

		// Handle escape - clear input
		if msg.Type == tea.KeyEsc {
			m.textInput.SetValue("")
			m.showingCommands = false
			return m, nil
		}
	}

	// Update the textinput
	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

// View renders the prompt
func (m PromptModel) View() string {
	// Prompt indicator
	indicator := promptStyle.Render(">")

	// Input field
	input := m.textInput.View()

	// Combine
	promptLine := lipgloss.JoinHorizontal(lipgloss.Left, indicator, " ", input)

	// Wrap in a box style
	return promptBoxStyle.Width(m.width).Render(promptLine)
}

// Focus sets focus on the prompt
func (m *PromptModel) Focus() tea.Cmd {
	return m.textInput.Focus()
}

// Blur removes focus from the prompt
func (m *PromptModel) Blur() {
	m.textInput.Blur()
}

// SetValue sets the input value
func (m *PromptModel) SetValue(value string) {
	m.textInput.SetValue(value)
}

// Value returns the current input value
func (m *PromptModel) Value() string {
	return m.textInput.Value()
}

// Custom messages

// ShowCommandPaletteMsg indicates command palette should be shown
type ShowCommandPaletteMsg struct{}

// CommandSubmitMsg indicates a command was submitted
type CommandSubmitMsg struct {
	command string
}

// ChatSubmitMsg indicates a chat message was submitted
type ChatSubmitMsg struct {
	message string
}

// Styles for the prompt
var (
	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")). // Cyan
			Bold(true)

	promptBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
)
