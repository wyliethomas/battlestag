package ui

import (
	"fmt"
	"strings"
	"time"

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
	SettingsMode
	ProgramsMode
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
	settings   *SettingsModel
	programs   *ProgramsModel

	// State
	waitingForLLM bool
	connected     bool // Current connection status
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
		settings:   NewSettingsModel(cfg),
		programs:   NewProgramsModel(apiClient),
		connected:  true, // Start optimistic
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.prompt.Init(),
		tea.EnterAltScreen,
		doConnectionCheck(m.client), // Initial connection check
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate chat viewport height more precisely
		// Count actual lines used by fixed elements:
		// - Header: 3 lines (padding top, content, padding bottom+border)
		// - Footer: 4 lines (status line if present, help text, padding, border)
		// - Prompt: 3 lines (border, content with padding, bottom)
		// Total: ~10 lines for UI chrome
		reservedLines := 10

		if m.palette.IsVisible() {
			reservedLines += 12 // Additional space for palette
		}

		chatViewportHeight := msg.Height - reservedLines
		if chatViewportHeight < 5 {
			chatViewportHeight = 5 // Minimum viewport
		}

		m.chat.SetSize(msg.Width, chatViewportHeight)
		m.palette.SetWidth(msg.Width)
		return m, nil

	case ConfigSavedMsg:
		// Config was saved, return to chat mode
		if m.mode == OnboardingMode || m.mode == SettingsMode {
			m.mode = ChatMode
			m.statusMsg = "Settings saved successfully!"
			// Update client with new config
			m.client = client.NewClientWithConfig(m.config.AgentGateway.URL)
			// IMPORTANT: Also update chat model's client reference (preserving history)
			m.chat.UpdateClient(m.client)
			// Reinitialize programs model with new client
			m.programs = NewProgramsModel(m.client)
			// Trigger connection check
			return m, doConnectionCheck(m.client)
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

		// If LLM suggested a program, validate it exists before executing
		if msg.programID != "" {
			// Validate program exists
			programs, err := m.client.ListPrograms()
			if err != nil {
				// Can't validate, log and skip execution
				m.statusMsg = "Warning: Could not validate program"
				return m, nil
			}

			// Check if program exists in registry
			programExists := false
			for _, prog := range programs {
				if prog.Name == msg.programID {
					programExists = true
					break
				}
			}

			if programExists {
				// Program is valid, execute it
				m.statusMsg = fmt.Sprintf("Executing program: %s", msg.programID)
				return m, m.executeProgramCmd(msg.programID, msg.programParams)
			} else {
				// LLM hallucinated a program name - log and ignore
				fmt.Printf("[WARNING] LLM suggested non-existent program: '%s' - treating as normal chat\n", msg.programID)
				m.statusMsg = fmt.Sprintf("Note: LLM suggested non-existent program '%s', ignoring", msg.programID)
				// Just continue with the chat response, don't try to execute
			}
		}

		return m, nil

	case ProgramResponseMsg:
		// Received program execution output
		m.chat.AddProgramMessage(msg.programName, msg.output)
		m.statusMsg = ""
		return m, nil

	case ShowCommandPaletteMsg:
		// Show command palette
		m.palette.Show(msg.filter)
		// Recalculate chat viewport when palette opens
		chatHeight := m.height - 10 - 12 // UI chrome + palette
		if chatHeight < 5 {
			chatHeight = 5
		}
		m.chat.SetSize(m.width, chatHeight)
		return m, nil

	case PaletteCloseMsg:
		// Close command palette
		m.palette.Hide()
		// Recalculate chat viewport when palette closes
		chatHeight := m.height - 10 // Just UI chrome (header, footer, prompt)
		if chatHeight < 5 {
			chatHeight = 5
		}
		m.chat.SetSize(m.width, chatHeight)
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

	case ConnectionCheckMsg:
		// Update connection status
		m.connected = msg.connected
		// Schedule next check in 10 seconds
		return m, tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
			return doConnectionCheck(m.client)()
		})
	}

	// Route updates based on current mode
	switch m.mode {
	case OnboardingMode:
		if m.onboarding != nil {
			onboarding, cmd := m.onboarding.Update(msg)
			m.onboarding = &onboarding
			cmds = append(cmds, cmd)
		}

	case SettingsMode:
		if m.settings != nil {
			settings, cmd := m.settings.Update(msg)
			m.settings = &settings
			cmds = append(cmds, cmd)
		}
		// Allow ESC to return to chat
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				m.mode = ChatMode
			}
		}

	case ProgramsMode:
		if m.programs != nil {
			programs, cmd := m.programs.Update(msg)
			m.programs = &programs
			cmds = append(cmds, cmd)
		}
		// Allow ESC to return to chat
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				m.mode = ChatMode
			}
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

	// Show settings if in settings mode
	if m.mode == SettingsMode {
		if m.settings != nil {
			return m.settings.View()
		}
	}

	// Show programs if in programs mode
	if m.mode == ProgramsMode {
		if m.programs != nil {
			return m.programs.View()
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
	if !m.connected {
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
	// Return directly without extra padding/styling to maintain exact height
	return m.chat.View()
}

// renderFooter renders the application footer with help text
func (m Model) renderFooter() string {
	var help string

	if m.waitingForLLM {
		help = "Waiting for response..."
	} else {
		help = "[/] Commands • [↑↓] Scroll • [Ctrl+L] Clear • [Ctrl+C] Quit"
	}

	// Pad the help text to fill the width
	// Account for border and padding characters
	availableWidth := m.width
	helpLength := lipgloss.Width(help)

	// Add spacing to fill the line
	if helpLength < availableWidth {
		help = help + strings.Repeat(" ", availableWidth-helpLength)
	}

	// Build footer content
	var parts []string
	if m.statusMsg != "" {
		statusLength := lipgloss.Width(m.statusMsg)
		paddedStatus := m.statusMsg
		if statusLength < availableWidth {
			paddedStatus = paddedStatus + strings.Repeat(" ", availableWidth-statusLength)
		}
		parts = append(parts, paddedStatus)
	}
	parts = append(parts, help)

	content := strings.Join(parts, "\n")

	// Apply styling to the entire footer
	styledFooter := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // Bright black/grey (respects theme)
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(lipgloss.Color("8")). // Bright black/grey (respects theme)
		Padding(1, 0).
		Render(content)

	return styledFooter
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
			programID:         response.ProgramID,
			programParams:     response.ProgramParams,
		}
	}
}

// executeProgramCmd executes a program and displays the result
func (m *Model) executeProgramCmd(programID string, params map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		// First, get program info to get the display name
		programs, err := m.client.ListPrograms()
		if err != nil {
			return ErrorMsg{err: fmt.Errorf("failed to get program info: %w", err)}
		}

		// Find the program name
		programName := programID // Fallback to Name
		for _, prog := range programs {
			if prog.Name == programID {
				programName = prog.Name
				break
			}
		}

		// Execute the program
		result, err := m.client.ExecuteProgram(programID, params)
		if err != nil {
			return ErrorMsg{err: err}
		}

		if !result.Success {
			// Return as program error, not as assistant message
			return ProgramResponseMsg{
				programName: programName + " (Error)",
				output:      fmt.Sprintf("❌ Program execution failed: %s", result.Error),
			}
		}

		// Return program output with program name
		return ProgramResponseMsg{
			programName: programName,
			output:      result.Output,
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

  /system:programs        - List available programs
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
		m.mode = SettingsMode
		m.statusMsg = ""
		return nil

	case "system:programs":
		m.mode = ProgramsMode
		m.programs = NewProgramsModel(m.client) // Reinitialize to refresh
		m.statusMsg = ""
		return m.programs.Init()

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

		return ChatResponseMsg{message: message, suggestedCommands: nil}
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

		return ChatResponseMsg{message: message, suggestedCommands: nil}
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

		return ChatResponseMsg{message: message, suggestedCommands: nil}
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

// fetchPrograms fetches the list of available programs
func (m *Model) fetchPrograms() tea.Cmd {
	return func() tea.Msg {
		programs, err := m.client.ListPrograms()
		if err != nil {
			return ErrorMsg{err: err}
		}

		if len(programs) == 0 {
			return ChatResponseMsg{message: "No programs available.", suggestedCommands: nil}
		}

		message := "📦 Available Programs:\n\n"
		for _, program := range programs {
			message += fmt.Sprintf("• %s\n", program.Name)
			message += fmt.Sprintf("  %s\n", program.Description)
			message += "\n"
		}

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

// doConnectionCheck performs a lightweight health check on the API
func doConnectionCheck(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		// Try to list programs as a lightweight health check
		_, err := c.ListPrograms()
		return ConnectionCheckMsg{
			connected: err == nil,
		}
	}
}

// Custom messages

// ChatResponseMsg contains the LLM response
type ChatResponseMsg struct {
	message           string
	suggestedCommands []string
	programID         string
	programParams     map[string]interface{}
}

// ProgramResponseMsg contains program execution output
type ProgramResponseMsg struct {
	programName string
	output      string
}

// ErrorMsg represents an error message
type ErrorMsg struct {
	err error
}

// StatusMsg represents a status message
type StatusMsg struct {
	message string
}

// ConnectionCheckMsg represents the result of a connection health check
type ConnectionCheckMsg struct {
	connected bool
}

// Styles specific to app.go (non-duplicates)

var (
	footerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(lipgloss.Color("240"))
)
