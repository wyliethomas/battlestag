package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"battlestag/client"
	"battlestag/config"
)

// SettingsModel represents the settings screen state
type SettingsModel struct {
	config       *config.Config
	originalConfig *config.Config
	focusedField int // 0=URL, 1=TestButton, 2=SaveButton, 3=ResetButton
	testing      bool
	testResult   string
	testError    error
	saveStatus   string
}

// NewSettingsModel creates a new settings model
func NewSettingsModel(cfg *config.Config) *SettingsModel {
	// Make a copy of the config
	configCopy := *cfg
	return &SettingsModel{
		config:       &configCopy,
		originalConfig: cfg,
		focusedField: 0,
	}
}

// Update handles settings messages
func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
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
		m.saveStatus = "Configuration saved successfully!"
		// Update the original config
		*m.originalConfig = *m.config
		return m, nil

	case tea.KeyMsg:
		m.saveStatus = "" // Clear save status on any key press

		switch msg.String() {
		case "tab", "down":
			m.focusedField = (m.focusedField + 1) % 4
		case "shift+tab", "up":
			m.focusedField = (m.focusedField - 1 + 4) % 4
		case "enter":
			switch m.focusedField {
			case 1: // Test Connection
				m.testing = true
				m.testResult = ""
				m.testError = nil
				return m, m.testConnection()
			case 2: // Save
				return m, m.saveConfig()
			case 3: // Reset to Defaults
				m.config = config.DefaultConfig()
				m.saveStatus = "Reset to defaults (not saved yet)"
			}
		case "backspace":
			if m.focusedField == 0 && len(m.config.AgentGateway.URL) > 0 {
				m.config.AgentGateway.URL = m.config.AgentGateway.URL[:len(m.config.AgentGateway.URL)-1]
			}
		case "ctrl+u": // Clear field
			if m.focusedField == 0 {
				m.config.AgentGateway.URL = ""
			}
		case "esc":
			// Revert changes
			*m.config = *m.originalConfig
			m.saveStatus = "Changes discarded"
		default:
			// Type into fields
			if len(msg.String()) == 1 {
				if m.focusedField == 0 {
					m.config.AgentGateway.URL += msg.String()
				}
			}
		}
	}

	return m, nil
}

// testConnection tests the connection to the Agent Gateway
func (m *SettingsModel) testConnection() tea.Cmd {
	return func() tea.Msg {
		c := client.NewClientWithConfig(m.config.AgentGateway.URL)
		health, err := c.Health()
		if err != nil {
			return ConnectionTestMsg{err: err}
		}
		return ConnectionTestMsg{status: health.Status}
	}
}

// saveConfig saves the configuration
func (m *SettingsModel) saveConfig() tea.Cmd {
	return func() tea.Msg {
		if err := config.Save(m.config); err != nil {
			return ErrorMsg{err: err}
		}
		return ConfigSavedMsg{}
	}
}

// View renders the settings screen
func (m *SettingsModel) View() string {
	title := titleStyle.Render("⚙️  Settings")

	intro := `
Configure your Agent Gateway connection settings.
Changes are saved to ~/.config/battlestag-tui/config.yaml
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

	// Buttons
	testButton := "[Test Connection]"
	saveButton := "[Save]"
	resetButton := "[Reset to Defaults]"

	if m.focusedField == 1 {
		testButton = activeTabStyle.Render(testButton)
	} else {
		testButton = tabStyle.Render(testButton)
	}

	if m.focusedField == 2 {
		saveButton = activeTabStyle.Render(saveButton)
	} else {
		saveButton = tabStyle.Render(saveButton)
	}

	if m.focusedField == 3 {
		resetButton = activeTabStyle.Render(resetButton)
	} else {
		resetButton = tabStyle.Render(resetButton)
	}

	buttons := testButton + "  " + saveButton + "  " + resetButton

	// Status messages
	var statusMsg string
	if m.testing {
		statusMsg = infoStyle.Render("\n\nTesting connection...")
	} else if m.testError != nil {
		statusMsg = errorStyle.Render(fmt.Sprintf("\n\nConnection failed: %v", m.testError))
	} else if m.testResult != "" {
		statusMsg = successStyle.Render(fmt.Sprintf("\n\nConnection successful! Status: %s", m.testResult))
	}

	if m.saveStatus != "" {
		statusMsg += successStyle.Render("\n\n" + m.saveStatus)
	}

	// Help text
	help := helpStyle.Render("\n\n[Tab/↑↓] Navigate • [Enter] Select • [Ctrl+U] Clear • [Esc] Discard changes")

	content := title + intro + "\n\n" + urlField + "\n\n" + buttons + statusMsg + help

	return boxStyle.Render(content)
}
