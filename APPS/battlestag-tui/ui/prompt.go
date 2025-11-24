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

	case PaletteFilterUpdateMsg:
		// Update input value from palette navigation
		m.textInput.SetValue(msg.filter)
		m.textInput.CursorEnd()
		return m, nil

	case tea.KeyMsg:
		// Don't process keys if palette is open - let app handle it
		// Except for text input which should update the filter
		if m.showingCommands {
			// Allow typing to continue filtering
			if msg.Type == tea.KeyRunes {
				// Update textinput
				m.textInput, cmd = m.textInput.Update(msg)
				// Send filter update
				return m, tea.Batch(cmd, func() tea.Msg {
					return PromptFilterUpdateMsg{filter: m.textInput.Value()}
				})
			}
			// Allow backspace to update filter
			if msg.Type == tea.KeyBackspace || msg.String() == "backspace" {
				m.textInput, cmd = m.textInput.Update(msg)
				newValue := m.textInput.Value()
				// If we deleted the "/", close palette
				if !strings.HasPrefix(newValue, "/") {
					m.showingCommands = false
					return m, tea.Batch(cmd, func() tea.Msg {
						return PaletteCloseMsg{}
					})
				}
				// Otherwise update filter
				return m, tea.Batch(cmd, func() tea.Msg {
					return PromptFilterUpdateMsg{filter: newValue}
				})
			}
			// Don't handle other keys when palette is open
			return m, nil
		}

		// Handle enter key - submit message
		if msg.Type == tea.KeyEnter {
			value := strings.TrimSpace(m.textInput.Value())
			if value != "" {
				// Check if it's a command (starts with /)
				if strings.HasPrefix(value, "/") {
					// It's a command - execute directly
					cmd := strings.TrimPrefix(value, "/")
					m.textInput.SetValue("")
					m.showingCommands = false
					return m, func() tea.Msg {
						return CommandExecuteMsg{command: cmd}
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

		// Detect "/" to show command palette
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '/' {
			// If input is empty and user types "/", show command palette
			if m.textInput.Value() == "" {
				m.showingCommands = true
				// First update the textinput to show "/"
				m.textInput, cmd = m.textInput.Update(msg)
				// Then show palette
				return m, tea.Batch(cmd, func() tea.Msg {
					return ShowCommandPaletteMsg{filter: "/"}
				})
			}
		}
	}

	// Update the textinput for normal typing
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
type ShowCommandPaletteMsg struct {
	filter string // Initial filter
}

// PromptFilterUpdateMsg indicates the prompt filter has changed (for palette)
type PromptFilterUpdateMsg struct {
	filter string
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
