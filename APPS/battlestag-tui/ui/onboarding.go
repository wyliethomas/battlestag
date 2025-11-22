package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/battlestag/battlestag-tui/client"
	"github.com/battlestag/battlestag-tui/config"
)

// OnboardingModel represents the onboarding screen state
type OnboardingModel struct {
	config       *config.Config
	focusedField int // 0=URL, 1=APIKey, 2=TestButton, 3=SaveButton
	testing      bool
	testResult   string
	testError    error
	saved        bool
}

// NewOnboardingModel creates a new onboarding model
func NewOnboardingModel() *OnboardingModel {
	return &OnboardingModel{
		config:       config.DefaultConfig(),
		focusedField: 0,
	}
}

// Init initializes the onboarding screen
func (m *OnboardingModel) Init() tea.Cmd {
	return nil
}

// Update handles onboarding messages
func (m OnboardingModel) Update(msg tea.Msg) (OnboardingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ConnectionTestMsg:
		m.testing = false
		if msg.err != nil {
			m.testError = msg.err
			m.testResult = ""
		} else {
			m.testResult = msg.status
			m.testError = nil
		}
		return m, nil

	case ConfigSavedMsg:
		m.saved = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.focusedField = (m.focusedField + 1) % 4
		case "shift+tab", "up":
			m.focusedField = (m.focusedField - 1 + 4) % 4
		case "enter":
			switch m.focusedField {
			case 2: // Test Connection
				m.testing = true
				m.testResult = ""
				m.testError = nil
				return m, m.testConnection()
			case 3: // Save & Continue
				return m, m.saveConfig()
			}
		case "backspace":
			if m.focusedField == 0 && len(m.config.AgentGateway.URL) > 0 {
				m.config.AgentGateway.URL = m.config.AgentGateway.URL[:len(m.config.AgentGateway.URL)-1]
			} else if m.focusedField == 1 && len(m.config.AgentGateway.APIKey) > 0 {
				m.config.AgentGateway.APIKey = m.config.AgentGateway.APIKey[:len(m.config.AgentGateway.APIKey)-1]
			}
		case "ctrl+u": // Clear field
			if m.focusedField == 0 {
				m.config.AgentGateway.URL = ""
			} else if m.focusedField == 1 {
				m.config.AgentGateway.APIKey = ""
			}
		default:
			// Type into fields
			if len(msg.String()) == 1 {
				if m.focusedField == 0 {
					m.config.AgentGateway.URL += msg.String()
				} else if m.focusedField == 1 {
					m.config.AgentGateway.APIKey += msg.String()
				}
			}
		}
	}

	return m, nil
}

// testConnection tests the connection to the Agent Gateway
func (m *OnboardingModel) testConnection() tea.Cmd {
	return func() tea.Msg {
		c := client.NewClientWithConfig(m.config.AgentGateway.URL, m.config.AgentGateway.APIKey)
		health, err := c.Health()
		if err != nil {
			return ConnectionTestMsg{err: err}
		}
		return ConnectionTestMsg{status: health.Status}
	}
}

// saveConfig saves the configuration
func (m *OnboardingModel) saveConfig() tea.Cmd {
	return func() tea.Msg {
		if err := config.Save(m.config); err != nil {
			return ErrorMsg{err: err}
		}
		return ConfigSavedMsg{}
	}
}

// View renders the onboarding screen
func (m *OnboardingModel) View() string {
	title := titleStyle.Render("🚀 Welcome to BATTLESTAG TUI")

	intro := `
First-time setup! Please configure your Agent Gateway connection.

You can change these settings later from the Settings screen.
`

	// URL field
	urlLabel := labelStyle.Render("Agent Gateway URL:")
	urlValue := m.config.AgentGateway.URL
	if m.focusedField == 0 {
		urlValue += "█" // Cursor
	}
	urlInput := focusedInputStyle.Render(urlValue)
	if m.focusedField != 0 {
		urlInput = inputStyle.Render(urlValue)
	}
	urlField := fmt.Sprintf("%s\n%s", urlLabel, urlInput)

	// API Key field
	keyLabel := labelStyle.Render("API Key:")
	keyValue := m.config.AgentGateway.APIKey
	if m.focusedField == 1 {
		keyValue += "█" // Cursor
	}
	keyInput := focusedInputStyle.Render(keyValue)
	if m.focusedField != 1 {
		keyInput = inputStyle.Render(keyValue)
	}
	keyField := fmt.Sprintf("%s\n%s", keyLabel, keyInput)

	// Buttons
	testButton := "[Test Connection]"
	saveButton := "[Save & Continue]"

	if m.focusedField == 2 {
		testButton = activeTabStyle.Render(testButton)
	} else {
		testButton = tabStyle.Render(testButton)
	}

	if m.focusedField == 3 {
		saveButton = activeTabStyle.Render(saveButton)
	} else {
		saveButton = tabStyle.Render(saveButton)
	}

	buttons := testButton + "  " + saveButton

	// Status messages
	var statusMsg string
	if m.testing {
		statusMsg = infoStyle.Render("\n\nTesting connection...")
	} else if m.testError != nil {
		statusMsg = errorStyle.Render(fmt.Sprintf("\n\nConnection failed: %v", m.testError))
	} else if m.testResult != "" {
		statusMsg = successStyle.Render(fmt.Sprintf("\n\nConnection successful! Status: %s", m.testResult))
	}

	if m.saved {
		statusMsg += successStyle.Render("\n\nConfiguration saved! Starting application...")
	}

	// Help text
	help := helpStyle.Render("\n\n[Tab/↑↓] Navigate • [Enter] Select • [Ctrl+U] Clear field • [Backspace] Delete")

	content := title + intro + "\n\n" + urlField + "\n\n" + keyField + "\n\n" + buttons + statusMsg + help

	return boxStyle.Render(content)
}

// ConnectionTestMsg indicates connection test result
type ConnectionTestMsg struct {
	status string
	err    error
}

// ConfigSavedMsg indicates config was saved
type ConfigSavedMsg struct{}
