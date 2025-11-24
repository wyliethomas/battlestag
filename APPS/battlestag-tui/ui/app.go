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
	palette    *PaletteModel

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
		palette:    NewPaletteModel(),
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
		// Update palette width
		m.palette.SetWidth(msg.Width)
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

	case ShowCommandPaletteMsg:
		// Show command palette
		m.palette.Show(msg.filter)
		return m, nil

	case PaletteCloseMsg:
		// Close command palette
		m.palette.Hide()
		return m, nil

	case PromptFilterUpdateMsg:
		// Update palette filter from prompt typing
		if m.palette.IsVisible() {
			m.palette.SetFilter(msg.filter)
		}
		return m, nil

	case CommandExecuteMsg:
		// Execute a command
		return m, m.executeCommand(msg.command)

	case tea.KeyMsg:
		// If palette is open, let it handle navigation keys
		if m.palette.IsVisible() {
			// Route arrow keys and enter to palette
			switch msg.String() {
			case "up", "down", "k", "j", "enter", "esc":
				paletteCmd := m.palette.Update(msg)
				return m, paletteCmd
			}
		}

		// Global key handlers
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit

		case "ctrl+l":
			// Clear chat
			m.chat.Clear()
			m.statusMsg = "Chat cleared"
			return m, nil

		case "up":
			// Scroll chat up (when palette is not visible)
			if m.mode == ChatMode && !m.palette.IsVisible() {
				m.chat.ScrollUp(3)
				return m, nil
			}

		case "k":
			// Scroll chat up with k (only when input is empty)
			if m.mode == ChatMode && !m.palette.IsVisible() && m.prompt.Value() == "" {
				m.chat.ScrollUp(3)
				return m, nil
			}

		case "down":
			// Scroll chat down (when palette is not visible)
			if m.mode == ChatMode && !m.palette.IsVisible() {
				m.chat.ScrollDown(3)
				return m, nil
			}

		case "j":
			// Scroll chat down with j (only when input is empty)
			if m.mode == ChatMode && !m.palette.IsVisible() && m.prompt.Value() == "" {
				m.chat.ScrollDown(3)
				return m, nil
			}

		case "pgup":
			// Page up in chat
			if m.mode == ChatMode && !m.palette.IsVisible() {
				m.chat.ScrollUp(10)
				return m, nil
			}

		case "pgdown":
			// Page down in chat
			if m.mode == ChatMode && !m.palette.IsVisible() {
				m.chat.ScrollDown(10)
				return m, nil
			}

		case "home":
			// Scroll to top of chat
			if m.mode == ChatMode && !m.palette.IsVisible() {
				m.chat.ScrollUp(999999) // Scroll to max
				return m, nil
			}

		case "end":
			// Scroll to bottom of chat
			if m.mode == ChatMode && !m.palette.IsVisible() {
				m.chat.ScrollToBottom()
				return m, nil
			}
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
	var elements []string
	elements = append(elements, m.renderHeader())
	elements = append(elements, m.renderChatArea())
	elements = append(elements, m.renderFooter())

	// Add palette if visible (appears above prompt)
	if m.palette.IsVisible() {
		elements = append(elements, m.palette.View())
	}

	// Always show prompt at bottom
	elements = append(elements, m.prompt.View())

	return lipgloss.JoinVertical(lipgloss.Left, elements...)
}

// renderHeader renders the application header
func (m Model) renderHeader() string {
	title := headerStyle.Render("BATTLESTAG")

	// Connection status
	status := "Connected"
	statusStyle := statusConnected
	if m.err != nil {
		status = "Not Connected"
		statusStyle = statusDisconnected
	}
	statusText := statusStyle.Render(status)

	// Build header
	spacer := strings.Repeat(" ", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(statusText)-4))

	headerContent := lipgloss.JoinHorizontal(lipgloss.Top, title, spacer, statusText)

	return statusBarStyle.Width(m.width).Render(headerContent)
}

// renderChatArea renders the main chat conversation area
func (m Model) renderChatArea() string {
	// Chat model handles its own viewport and scrolling
	chatView := m.chat.View()

	// Just add padding, no height constraints here
	return lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 2).
		Render(chatView)
}

// renderFooter renders the application footer with help text
func (m Model) renderFooter() string {
	var help string

	if m.waitingForLLM {
		help = helpStyle.Render("Waiting for response...")
	} else {
		help = helpStyle.Render("[/] Commands • [↑↓] Scroll • [Ctrl+L] Clear • [Ctrl+C] Quit")
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

// executeCommand executes a command by key
func (m *Model) executeCommand(commandKey string) tea.Cmd {
	// Handle system commands
	switch commandKey {
	case "system:clear":
		m.chat.Clear()
		m.statusMsg = "Chat cleared"
		return nil

	case "system:quit":
		return tea.Quit

	case "system:help":
		helpText := `
Battlestag TUI Help

Commands:
  /finance:dashboard      - View financial overview
  /finance:assets         - List all assets
  /finance:liabilities    - List all liabilities
  /finance:upload         - Upload bank statement
  /finance:query          - Ask about finances

  /stoic:today            - Today's stoic thought
  /stoic:random           - Random stoic thought

  /tech:random            - Random tech tip
  /tech:latest            - Latest tech tip

  /system:settings        - Configure application
  /system:clear           - Clear chat history
  /system:help            - Show this help
  /system:quit            - Exit application

Keyboard Shortcuts:
  /                       - Open command palette
  ↑/↓ or k/j              - Scroll chat up/down
  PgUp/PgDown             - Page up/down in chat
  Home/End                - Jump to top/bottom of chat
  Ctrl+L                  - Clear chat
  Ctrl+C / Ctrl+D         - Quit
`
		m.chat.AddAssistantMessage(helpText, nil)
		return nil

	case "system:settings":
		m.statusMsg = "Settings not yet implemented"
		return nil

	// Finance commands - stub implementations
	case "finance:dashboard":
		m.statusMsg = "Fetching dashboard..."
		return m.fetchDashboard()

	case "finance:assets":
		m.statusMsg = "Fetching assets..."
		return m.fetchAssets()

	case "finance:liabilities":
		m.statusMsg = "Fetching liabilities..."
		return m.fetchLiabilities()

	case "finance:upload":
		m.chat.AddAssistantMessage("Upload feature coming soon! For now, you can upload statements via the API directly.", nil)
		return nil

	case "finance:query":
		m.chat.AddAssistantMessage("What would you like to know about your finances? Just ask me in the chat!", nil)
		return nil

	// Stoic commands
	case "stoic:today", "stoic:random":
		m.statusMsg = "Fetching stoic thought..."
		return m.fetchStoicThought(commandKey)

	// Tech commands
	case "tech:random", "tech:latest":
		m.statusMsg = "Tech tips not yet implemented"
		m.chat.AddAssistantMessage("Tech tips feature coming soon!", nil)
		return nil

	default:
		m.statusMsg = fmt.Sprintf("Unknown command: %s", commandKey)
		return nil
	}
}

// fetchDashboard fetches financial dashboard
func (m *Model) fetchDashboard() tea.Cmd {
	return func() tea.Msg {
		overview, err := m.client.GetFinancialOverview()
		if err != nil {
			return ErrorMsg{err: err}
		}

		message := fmt.Sprintf(`📊 Financial Overview

💰 Net Worth: $%.2f

Assets: $%.2f (%d items)
Liabilities: $%.2f (%d items)
`,
			overview.NetWorth,
			overview.TotalAssets, overview.AssetCount,
			overview.TotalLiabilities, overview.LiabilityCount,
		)

		return ChatResponseMsg{message: message, suggestedCommands: []string{"/finance:assets", "/finance:liabilities"}}
	}
}

// fetchAssets fetches asset list
func (m *Model) fetchAssets() tea.Cmd {
	return func() tea.Msg {
		assets, err := m.client.ListAssets()
		if err != nil {
			return ErrorMsg{err: err}
		}

		if len(assets) == 0 {
			return ChatResponseMsg{message: "You have no assets tracked yet.", suggestedCommands: nil}
		}

		message := "📊 Your Assets:\n\n"
		for _, asset := range assets {
			if !asset.IsRemoved {
				gain := asset.CurrentValue - asset.PurchasePrice
				message += fmt.Sprintf("• %s: $%.2f (gain: $%.2f)\n", asset.Name, asset.CurrentValue, gain)
			}
		}

		return ChatResponseMsg{message: message, suggestedCommands: []string{"/finance:dashboard"}}
	}
}

// fetchLiabilities fetches liability list
func (m *Model) fetchLiabilities() tea.Cmd {
	return func() tea.Msg {
		liabilities, err := m.client.ListLiabilities()
		if err != nil {
			return ErrorMsg{err: err}
		}

		if len(liabilities) == 0 {
			return ChatResponseMsg{message: "You have no liabilities tracked.", suggestedCommands: nil}
		}

		message := "💳 Your Liabilities:\n\n"
		for _, liability := range liabilities {
			message += fmt.Sprintf("• %s: $%.2f", liability.Name, liability.CurrentBalance)
			if liability.InterestRate > 0 {
				message += fmt.Sprintf(" (%.2f%% APR)", liability.InterestRate)
			}
			message += "\n"
		}

		return ChatResponseMsg{message: message, suggestedCommands: []string{"/finance:dashboard"}}
	}
}

// fetchStoicThought fetches a stoic thought
func (m *Model) fetchStoicThought(commandKey string) tea.Cmd {
	return func() tea.Msg {
		thought, err := m.client.GetTodayStoicThought()
		if err != nil {
			return ErrorMsg{err: err}
		}

		message := fmt.Sprintf("📖 Stoic Thought\n\n%s\n\n— %s", thought.Thought, thought.Source)
		return ChatResponseMsg{message: message, suggestedCommands: nil}
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
