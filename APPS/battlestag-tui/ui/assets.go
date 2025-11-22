package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/battlestag/battlestag-tui/client"
	"github.com/battlestag/battlestag-tui/models"
)

// AssetsModel represents the assets management screen
type AssetsModel struct {
	client      *client.Client
	assets      []models.Asset
	selected    int
	loading     bool
	err         error
	showingForm bool
	formAsset   models.Asset
}

// NewAssetsModel creates a new assets model
func NewAssetsModel(apiClient *client.Client) *AssetsModel {
	return &AssetsModel{
		client:  apiClient,
		loading: true,
	}
}

// Refresh fetches fresh asset data
func (m *AssetsModel) Refresh() tea.Cmd {
	return func() tea.Msg {
		assets, err := m.client.ListAssets()
		if err != nil {
			return ErrorMsg{err: err}
		}
		return AssetsDataMsg{assets: assets}
	}
}

// Update handles asset screen messages
func (m AssetsModel) Update(msg tea.Msg) (AssetsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case AssetsDataMsg:
		m.assets = msg.assets
		m.loading = false
		m.err = nil
		return m, nil

	case ErrorMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		if m.showingForm {
			// Handle form navigation (simplified for now)
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
			if m.selected < len(m.assets)-1 {
				m.selected++
			}
		case "r":
			m.loading = true
			return m, m.Refresh()
		case "d":
			if len(m.assets) > 0 && m.selected < len(m.assets) {
				return m, m.deleteAsset()
			}
		}
	}

	return m, nil
}

// deleteAsset deletes the selected asset
func (m *AssetsModel) deleteAsset() tea.Cmd {
	asset := m.assets[m.selected]
	return func() tea.Msg {
		err := m.client.DeleteAsset(asset.Name)
		if err != nil {
			return ErrorMsg{err: err}
		}
		return StatusMsg{message: fmt.Sprintf("Deleted asset: %s", asset.Name)}
	}
}

// View renders the assets screen
func (m *AssetsModel) View() string {
	if m.loading {
		return infoStyle.Render("Loading assets...")
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error loading assets: %v\n\nPress [r] to retry", m.err))
	}

	title := titleStyle.Render("📊 Asset Management")

	if len(m.assets) == 0 {
		return boxStyle.Render(title + "\n\nNo assets found.\n\nPress [n] to add a new asset")
	}

	// Build table
	var rows []string

	// Header
	header := fmt.Sprintf("%-20s %-15s %15s %15s %12s",
		"Name", "Category", "Current Value", "Purchase Price", "Gain/Loss")
	rows = append(rows, tableHeaderStyle.Render(header))

	// Data rows
	for i, asset := range m.assets {
		if asset.IsRemoved {
			continue
		}

		gain := asset.CurrentValue - asset.PurchasePrice
		gainStr := FormatCurrency(gain)

		row := fmt.Sprintf("%-20s %-15s %15s %15s %12s",
			truncate(asset.Name, 20),
			truncate(asset.Category, 15),
			fmt.Sprintf("$%.2f", asset.CurrentValue),
			fmt.Sprintf("$%.2f", asset.PurchasePrice),
			gainStr,
		)

		if i == m.selected {
			rows = append(rows, tableSelectedStyle.Render(row))
		} else {
			rows = append(rows, tableRowStyle.Render(row))
		}
	}

	table := strings.Join(rows, "\n")

	// Calculate totals
	var totalValue, totalCost float64
	for _, asset := range m.assets {
		if !asset.IsRemoved {
			totalValue += asset.CurrentValue
			totalCost += asset.PurchasePrice
		}
	}
	totalGain := totalValue - totalCost

	summary := fmt.Sprintf("\nTotal Value: %s | Total Cost: %s | Total Gain: %s",
		FormatCurrency(totalValue),
		FormatCurrency(totalCost),
		FormatCurrency(totalGain),
	)

	help := helpStyle.Render("\n[↑/↓] Navigate • [d] Delete • [r] Refresh")

	content := title + "\n\n" + table + summary + help

	return boxStyle.Render(content)
}

// truncate truncates a string to a max length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// AssetsDataMsg contains asset data
type AssetsDataMsg struct {
	assets []models.Asset
}
