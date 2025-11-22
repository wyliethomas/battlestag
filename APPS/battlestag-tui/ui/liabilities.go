package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/battlestag/battlestag-tui/client"
	"github.com/battlestag/battlestag-tui/models"
)

// LiabilitiesModel represents the liabilities management screen
type LiabilitiesModel struct {
	client      *client.Client
	liabilities []models.Liability
	selected    int
	loading     bool
	err         error
	showingForm bool
	formLiability models.Liability
}

// NewLiabilitiesModel creates a new liabilities model
func NewLiabilitiesModel(apiClient *client.Client) *LiabilitiesModel {
	return &LiabilitiesModel{
		client:  apiClient,
		loading: true,
	}
}

// Refresh fetches fresh liability data
func (m *LiabilitiesModel) Refresh() tea.Cmd {
	return func() tea.Msg {
		liabilities, err := m.client.ListLiabilities()
		if err != nil {
			return ErrorMsg{err: err}
		}
		return LiabilitiesDataMsg{liabilities: liabilities}
	}
}

// Update handles liability screen messages
func (m LiabilitiesModel) Update(msg tea.Msg) (LiabilitiesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case LiabilitiesDataMsg:
		m.liabilities = msg.liabilities
		m.loading = false
		m.err = nil
		return m, nil

	case ErrorMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		if m.showingForm {
			switch msg.String() {
			case "esc":
				m.showingForm = false
				return m, nil
			}
			return m, nil
		}

		// List navigation
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.liabilities)-1 {
				m.selected++
			}
		case "r":
			m.loading = true
			return m, m.Refresh()
		case "d":
			if len(m.liabilities) > 0 && m.selected < len(m.liabilities) {
				return m, m.deleteLiability()
			}
		}
	}

	return m, nil
}

// deleteLiability deletes the selected liability
func (m *LiabilitiesModel) deleteLiability() tea.Cmd {
	liability := m.liabilities[m.selected]
	return func() tea.Msg {
		err := m.client.DeleteLiability(liability.Name)
		if err != nil {
			return ErrorMsg{err: err}
		}
		return StatusMsg{message: fmt.Sprintf("Deleted liability: %s", liability.Name)}
	}
}

// View renders the liabilities screen
func (m *LiabilitiesModel) View() string {
	if m.loading {
		return infoStyle.Render("Loading liabilities...")
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error loading liabilities: %v\n\nPress [r] to retry", m.err))
	}

	title := titleStyle.Render("💳 Liability Management")

	if len(m.liabilities) == 0 {
		return boxStyle.Render(title + "\n\nNo liabilities found.\n\nPress [n] to add a new liability")
	}

	// Build table
	var rows []string

	// Header
	header := fmt.Sprintf("%-20s %-15s %15s %15s %10s",
		"Name", "Type", "Balance", "Credit Limit", "Rate")
	rows = append(rows, tableHeaderStyle.Render(header))

	// Data rows
	for i, liability := range m.liabilities {
		limitStr := "-"
		if liability.CreditLimit > 0 {
			limitStr = fmt.Sprintf("$%.2f", liability.CreditLimit)
		}

		rateStr := "-"
		if liability.InterestRate > 0 {
			rateStr = fmt.Sprintf("%.2f%%", liability.InterestRate)
		}

		row := fmt.Sprintf("%-20s %-15s %15s %15s %10s",
			truncate(liability.Name, 20),
			truncate(liability.LiabilityType, 15),
			fmt.Sprintf("$%.2f", liability.CurrentBalance),
			limitStr,
			rateStr,
		)

		if i == m.selected {
			rows = append(rows, tableSelectedStyle.Render(row))
		} else {
			rows = append(rows, tableRowStyle.Render(row))
		}
	}

	table := strings.Join(rows, "\n")

	// Calculate totals
	var totalBalance float64
	var totalLimit float64
	for _, liability := range m.liabilities {
		totalBalance += liability.CurrentBalance
		if liability.CreditLimit > 0 {
			totalLimit += liability.CreditLimit
		}
	}

	utilization := 0.0
	if totalLimit > 0 {
		utilization = (totalBalance / totalLimit) * 100
	}

	summary := fmt.Sprintf("\nTotal Balance: %s | Total Limit: %s | Utilization: %s",
		negativeStyle.Render(fmt.Sprintf("$%.2f", totalBalance)),
		fmt.Sprintf("$%.2f", totalLimit),
		FormatPercent(utilization),
	)

	help := helpStyle.Render("\n[↑/↓] Navigate • [d] Delete • [r] Refresh")

	content := title + "\n\n" + table + summary + help

	return boxStyle.Render(content)
}

// LiabilitiesDataMsg contains liability data
type LiabilitiesDataMsg struct {
	liabilities []models.Liability
}
