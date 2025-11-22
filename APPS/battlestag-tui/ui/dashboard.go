package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/battlestag/battlestag-tui/client"
	"github.com/battlestag/battlestag-tui/models"
)

// DashboardModel represents the dashboard screen state
type DashboardModel struct {
	client    *client.Client
	overview  *models.FinancialOverview
	thought   *models.StoicThought
	health    *models.HealthResponse
	loading   bool
	err       error
	lastUpdate time.Time
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel(apiClient *client.Client) *DashboardModel {
	return &DashboardModel{
		client:  apiClient,
		loading: true,
	}
}

// Init initializes the dashboard
func (m *DashboardModel) Init() tea.Cmd {
	return m.Refresh()
}

// Refresh fetches fresh data from the API
func (m *DashboardModel) Refresh() tea.Cmd {
	return func() tea.Msg {
		// Fetch health status
		health, err := m.client.Health()
		if err != nil {
			return ErrorMsg{err: err}
		}

		// Fetch financial overview
		overview, err := m.client.GetFinancialOverview()
		if err != nil {
			return ErrorMsg{err: err}
		}

		// Fetch today's stoic thought
		thought, err := m.client.GetTodayStoicThought()
		if err != nil {
			// Non-critical error, continue
			thought = nil
		}

		return DashboardDataMsg{
			health:   health,
			overview: overview,
			thought:  thought,
		}
	}
}

// Update handles dashboard messages
func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case DashboardDataMsg:
		m.health = msg.health
		m.overview = msg.overview
		m.thought = msg.thought
		m.loading = false
		m.lastUpdate = time.Now()
		m.err = nil
		return m, nil

	case ErrorMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.loading = true
			return m, m.Refresh()
		}
	}

	return m, nil
}

// View renders the dashboard
func (m *DashboardModel) View() string {
	if m.loading {
		return infoStyle.Render("Loading dashboard...")
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error loading dashboard: %v\n\nPress [r] to retry", m.err))
	}

	var sections []string

	// Financial Overview Section
	sections = append(sections, m.renderFinancialOverview())

	// Daily Insight Section
	if m.thought != nil {
		sections = append(sections, m.renderDailyInsight())
	}

	// System Status Section
	sections = append(sections, m.renderSystemStatus())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderFinancialOverview renders the financial summary
func (m *DashboardModel) renderFinancialOverview() string {
	if m.overview == nil {
		return ""
	}

	title := titleStyle.Render("💰 Financial Overview")

	// Calculate gain/loss
	netWorth := m.overview.NetWorth
	netWorthColor := positiveStyle
	if netWorth < 0 {
		netWorthColor = negativeStyle
	}

	content := fmt.Sprintf(`
Total Assets:      %s
Total Liabilities: %s
─────────────────────────────
Net Worth:         %s

Asset Count:       %d
Liability Count:   %d
`,
		FormatCurrency(m.overview.TotalAssets),
		FormatCurrency(m.overview.TotalLiabilities),
		netWorthColor.Render(fmt.Sprintf("$%.2f", netWorth)),
		m.overview.AssetCount,
		m.overview.LiabilityCount,
	)

	return boxStyle.Render(title + content)
}

// renderDailyInsight renders today's stoic thought
func (m *DashboardModel) renderDailyInsight() string {
	if m.thought == nil {
		return ""
	}

	title := titleStyle.Render("📖 Daily Stoic Thought")

	// Wrap the thought text
	wrappedThought := wordWrap(m.thought.Thought, 70)

	content := fmt.Sprintf(`
%s

— %s
`,
		wrappedThought,
		lipgloss.NewStyle().Italic(true).Foreground(colorMuted).Render(m.thought.Source),
	)

	return boxStyle.Render(title + content)
}

// renderSystemStatus renders system health status
func (m *DashboardModel) renderSystemStatus() string {
	title := titleStyle.Render("🔧 System Status")

	healthStatus := "Unknown"
	if m.health != nil {
		healthStatus = m.health.Status
	}

	lastUpdateStr := "Never"
	if !m.lastUpdate.IsZero() {
		lastUpdateStr = m.lastUpdate.Format("15:04:05")
	}

	content := fmt.Sprintf(`
API Status:    %s
Last Updated:  %s

Press [r] to refresh data
`,
		statusConnected.Render(healthStatus),
		lastUpdateStr,
	)

	return boxStyle.Render(title + content)
}

// wordWrap wraps text to a specified width
func wordWrap(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// DashboardDataMsg contains dashboard data
type DashboardDataMsg struct {
	health   *models.HealthResponse
	overview *models.FinancialOverview
	thought  *models.StoicThought
}
