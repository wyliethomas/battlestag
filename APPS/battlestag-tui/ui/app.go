package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"battlestag/client"
	"battlestag/config"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	OnboardingMode ViewMode = iota
	ChatMode
	CommandPaletteMode
)

// Model represents the main application state
type Model struct {
	client    *client.Client
	config    *config.Config
	mode      ViewMode
	width     int
	height    int
	err       error
	statusMsg string

	// Core components
	onboarding *OnboardingModel
	chat       *ChatModel
	prompt     PromptModel

	// State
	waitingForLLM bool
}

// NewModel creates a new application model
func NewModel(apiClient *client.Client, cfg *config.Config, showOnboarding bool) Model {
	startMode := ChatMode
	var onboarding *OnboardingModel

	if showOnboarding {
		startMode = OnboardingMode
		onboarding = NewOnboardingModel()
	}

	return Model{
		client:     apiClient,
		config:     cfg,
		mode:       startMode,
		onboarding: onboarding,
		chat:       NewChatModel(apiClient),
		prompt:     NewPromptModel(),
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.prompt.Init(),
		tea.EnterAltScreen,
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update chat view size (leave room for header and prompt)
		m.chat.SetSize(msg.Width, msg.Height-10)
		return m, nil

	case ConfigSavedMsg:
		// Config was saved from onboarding, start chat mode
		if m.mode == OnboardingMode {
			m.mode = ChatMode
		}
		return m, nil

	case ChatSubmitMsg:
		// User submitted a chat message
		if !m.waitingForLLM {
			m.chat.AddUserMessage(msg.message)
			m.waitingForLLM = true
			m.chat.SetLoading(true)
			return m, m.sendChatMessage(msg.message)
		}
		return m, nil

	case ChatResponseMsg:
		// Received response from LLM
		m.waitingForLLM = false
		m.chat.SetLoading(false)
		m.chat.AddAssistantMessage(msg.message, msg.suggestedCommands)
		return m, nil

	case CommandSubmitMsg:
		// User submitted a command
		m.statusMsg = fmt.Sprintf("Command: %s (commands not yet implemented)", msg.command)
		// TODO: Implement command execution
		return m, nil

	case ShowCommandPaletteMsg:
		// Show command palette
		m.statusMsg = "Command palette coming soon! Type your command after /"
		// TODO: Implement command palette
		return m, nil

	case tea.KeyMsg:
		// Global key handlers
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit

		case "ctrl+l":
			// Clear chat
			m.chat.Clear()
			m.statusMsg = "Chat cleared"
			return m, nil
		}

	case ErrorMsg:
		m.err = msg.err
		m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		m.waitingForLLM = false
		m.chat.SetLoading(false)
		m.chat.AddErrorMessage(msg.err)
		return m, nil
	}

	// Route updates based on current mode
	switch m.mode {
	case OnboardingMode:
		if m.onboarding != nil {
			onboarding, cmd := m.onboarding.Update(msg)
			m.onboarding = &onboarding
			cmds = append(cmds, cmd)
		}

	case ChatMode:
		// Update prompt (always active in chat mode)
		var promptCmd tea.Cmd
		m.prompt, promptCmd = m.prompt.Update(msg)
		if promptCmd != nil {
			cmds = append(cmds, promptCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the application
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Show onboarding if in onboarding mode
	if m.mode == OnboardingMode {
		if m.onboarding != nil {
			return m.onboarding.View()
		}
	}

	// Build chat mode view
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		m.renderChatArea(),
		m.renderFooter(),
		m.prompt.View(),
	)
}

// renderHeader renders the application header
func (m Model) renderHeader() string {
	title := headerStyle.Render("BATTLESTAG")

	// Connection status
	status := "Connected"
	statusStyle := statusConnected
	if m.err != nil {
		status = "Error"
		statusStyle = statusDisconnected
	}
	statusText := statusStyle.Render(status)

	// Build header
	spacer := strings.Repeat(" ", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(statusText)-4))

	return statusBarStyle.Width(m.width).Render(
		lipgloss.JoinHorizontal(lipgloss.Top, title, spacer, statusText),
	)
}

// renderChatArea renders the main chat conversation area
func (m Model) renderChatArea() string {
	// Calculate available height for chat (subtract header, footer, prompt)
	chatHeight := m.height - 8

	chatView := m.chat.View()

	// Create scrollable area
	return lipgloss.NewStyle().
		Width(m.width).
		Height(chatHeight).
		Padding(1, 2).
		Render(chatView)
}

// renderFooter renders the application footer with help text
func (m Model) renderFooter() string {
	var help string

	if m.waitingForLLM {
		help = helpStyle.Render("Waiting for response...")
	} else {
		help = helpStyle.Render("[/] Commands • [Ctrl+L] Clear chat • [Ctrl+C] Quit")
	}

	var statusLine string
	if m.statusMsg != "" {
		statusLine = infoStyle.Render(m.statusMsg)
	}

	footer := lipgloss.JoinVertical(lipgloss.Left, statusLine, help)
	return footerStyle.Width(m.width).Render(footer)
}

// sendChatMessage sends a message to the LLM API
func (m *Model) sendChatMessage(message string) tea.Cmd {
	return func() tea.Msg {
		// Get conversation history
		history := m.chat.GetHistory()

		// Send to LLM
		response, err := m.client.Chat(message, history)
		if err != nil {
			return ErrorMsg{err: err}
		}

		return ChatResponseMsg{
			message:           response.Message,
			suggestedCommands: response.SuggestedCommands,
		}
	}
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Custom messages

// ChatResponseMsg contains the LLM response
type ChatResponseMsg struct {
	message           string
	suggestedCommands []string
}

// ErrorMsg represents an error message
type ErrorMsg struct {
	err error
}

// StatusMsg represents a status message
type StatusMsg struct {
	message string
}

// Styles specific to app.go (non-duplicates)

var (
	footerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
)
