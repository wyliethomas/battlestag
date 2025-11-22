package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/battlestag/battlestag-tui/client"
	"github.com/battlestag/battlestag-tui/config"
)

// Screen represents different views in the TUI
type Screen int

const (
	OnboardingScreen Screen = iota
	DashboardScreen
	AssetsScreen
	LiabilitiesScreen
	UploadScreen
	SettingsScreen
	HelpScreen
)

// Model represents the main application state
type Model struct {
	client        *client.Client
	config        *config.Config
	currentScreen Screen
	width         int
	height        int
	err           error
	statusMsg     string

	// Sub-models for different screens
	onboarding   *OnboardingModel
	dashboard    *DashboardModel
	assets       *AssetsModel
	liabilities  *LiabilitiesModel
	upload       *UploadModel
	settings     *SettingsModel
}

// NewModel creates a new application model
func NewModel(apiClient *client.Client, cfg *config.Config, showOnboarding bool) Model {
	startScreen := DashboardScreen
	var onboarding *OnboardingModel

	if showOnboarding {
		startScreen = OnboardingScreen
		onboarding = NewOnboardingModel()
	}

	return Model{
		client:        apiClient,
		config:        cfg,
		currentScreen: startScreen,
		onboarding:    onboarding,
		dashboard:     NewDashboardModel(apiClient),
		assets:        NewAssetsModel(apiClient),
		liabilities:   NewLiabilitiesModel(apiClient),
		upload:        NewUploadModel(apiClient),
		settings:      NewSettingsModel(cfg),
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.dashboard.Init(),
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
		return m, nil

	case ConfigSavedMsg:
		// Config was saved from onboarding, reload and start app
		if m.currentScreen == OnboardingScreen {
			m.currentScreen = DashboardScreen
			return m, m.dashboard.Refresh()
		}
		return m, nil

	case tea.KeyMsg:
		// Don't allow quitting or tab switching during onboarding
		if m.currentScreen == OnboardingScreen {
			break
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "1":
			m.currentScreen = DashboardScreen
			return m, m.dashboard.Refresh()

		case "2":
			m.currentScreen = AssetsScreen
			return m, m.assets.Refresh()

		case "3":
			m.currentScreen = LiabilitiesScreen
			return m, m.liabilities.Refresh()

		case "4":
			m.currentScreen = UploadScreen
			return m, nil

		case "5", "s":
			m.currentScreen = SettingsScreen
			return m, nil

		case "?":
			m.currentScreen = HelpScreen
			return m, nil
		}

	case ErrorMsg:
		m.err = msg.err
		return m, nil

	case StatusMsg:
		m.statusMsg = msg.message
		return m, nil
	}

	// Route updates to the current screen's model
	switch m.currentScreen {
	case OnboardingScreen:
		if m.onboarding != nil {
			onboarding, cmd := m.onboarding.Update(msg)
			m.onboarding = &onboarding
			cmds = append(cmds, cmd)
		}

	case DashboardScreen:
		dashboard, cmd := m.dashboard.Update(msg)
		m.dashboard = &dashboard
		cmds = append(cmds, cmd)

	case AssetsScreen:
		assets, cmd := m.assets.Update(msg)
		m.assets = &assets
		cmds = append(cmds, cmd)

	case LiabilitiesScreen:
		liabilities, cmd := m.liabilities.Update(msg)
		m.liabilities = &liabilities
		cmds = append(cmds, cmd)

	case UploadScreen:
		upload, cmd := m.upload.Update(msg)
		m.upload = &upload
		cmds = append(cmds, cmd)

	case SettingsScreen:
		settings, cmd := m.settings.Update(msg)
		m.settings = &settings
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the application
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string

	// Render the current screen
	switch m.currentScreen {
	case OnboardingScreen:
		if m.onboarding != nil {
			return m.onboarding.View() // Onboarding is full-screen
		}
	case DashboardScreen:
		content = m.dashboard.View()
	case AssetsScreen:
		content = m.assets.View()
	case LiabilitiesScreen:
		content = m.liabilities.View()
	case UploadScreen:
		content = m.upload.View()
	case SettingsScreen:
		content = m.settings.View()
	case HelpScreen:
		content = m.renderHelp()
	}

	// Build the full view
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		m.renderTabs(),
		content,
		m.renderFooter(),
	)
}

// renderHeader renders the application header
func (m Model) renderHeader() string {
	title := headerStyle.Render("BATTLESTAG TUI")

	// Check connection status
	status := "Checking..."
	statusStyle := statusConnected

	if m.err != nil {
		status = "Disconnected"
		statusStyle = statusDisconnected
	} else {
		status = "Connected"
	}

	statusText := statusStyle.Render(status)

	// Build header with title and status
	spacer := strings.Repeat(" ", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(statusText)-4))

	return statusBarStyle.Width(m.width).Render(
		lipgloss.JoinHorizontal(lipgloss.Top, title, spacer, statusText),
	)
}

// renderTabs renders the navigation tabs
func (m Model) renderTabs() string {
	tabs := []struct {
		name   string
		screen Screen
		key    string
	}{
		{"Dashboard", DashboardScreen, "1"},
		{"Assets", AssetsScreen, "2"},
		{"Liabilities", LiabilitiesScreen, "3"},
		{"Upload", UploadScreen, "4"},
		{"Settings", SettingsScreen, "5"},
	}

	var renderedTabs []string
	for _, tab := range tabs {
		style := tabStyle
		if tab.screen == m.currentScreen {
			style = activeTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(fmt.Sprintf("[%s] %s", tab.key, tab.name)))
	}

	return lipgloss.NewStyle().
		Padding(1, 0).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...))
}

// renderFooter renders the application footer
func (m Model) renderFooter() string {
	help := helpStyle.Render("[1-4] Switch tabs • [?] Help • [q] Quit")

	var statusMsg string
	if m.err != nil {
		statusMsg = errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	} else if m.statusMsg != "" {
		statusMsg = infoStyle.Render(m.statusMsg)
	}

	footer := lipgloss.JoinVertical(lipgloss.Left, statusMsg, help)
	return footer
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	help := `
BATTLESTAG TUI - Help

Navigation:
  [1] - Dashboard     View financial overview and daily insights
  [2] - Assets        Manage your financial assets
  [3] - Liabilities   Manage your liabilities
  [4] - Upload        Upload PDF bank statements
  [?] - Help          Show this help screen
  [q] - Quit          Exit the application

Screen-specific controls:
  [↑/↓] - Navigate lists
  [Enter] - Select item
  [n] - New item (Assets/Liabilities screens)
  [e] - Edit selected item
  [d] - Delete selected item
  [r] - Refresh data
  [Esc] - Cancel/Go back

Environment Variables:
  AGENT_GATEWAY_URL     - API endpoint (default: http://localhost:8080)
  AGENT_GATEWAY_API_KEY - API authentication key

Press any key to return to the previous screen.
`
	return boxStyle.Width(m.width - 4).Render(help)
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Custom messages
type ErrorMsg struct {
	err error
}

type StatusMsg struct {
	message string
}
