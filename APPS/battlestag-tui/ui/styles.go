package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors - Using ANSI colors to respect terminal theme
	colorPrimary   = lipgloss.Color("6")   // Cyan - for primary highlights
	colorSuccess   = lipgloss.Color("2")   // Green - for success/positive
	colorDanger    = lipgloss.Color("1")   // Red - for errors/danger
	colorWarning   = lipgloss.Color("3")   // Yellow - for warnings
	colorMuted     = lipgloss.Color("8")   // Bright black/grey - for muted text
	colorBorder    = lipgloss.Color("8")   // Bright black/grey - for borders
	colorHighlight = lipgloss.Color("11")  // Bright yellow - for highlights

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1).
			MarginBottom(1)

	// Status styles
	statusBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorBorder).
			Padding(1, 1)

	statusConnected = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	statusDisconnected = lipgloss.NewStyle().
				Foreground(colorDanger).
				Bold(true)

	// Navigation styles
	tabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorMuted)

	activeTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorPrimary).
			Bold(true).
			Underline(true)

	// Table styles
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(colorBorder)

	tableRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	tableSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(colorPrimary).
				Foreground(lipgloss.Color("#FFFFFF"))

	// Value styles
	positiveStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	negativeStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	neutralStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2).
			MarginBottom(1)

	// Help text
	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(1, 0)

	// Error message
	errorStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true).
			Padding(1, 2)

	// Success message
	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true).
			Padding(1, 2)

	// Info message
	infoStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Padding(1, 2)

	// Form styles
	labelStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginRight(1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1)
)

// FormatCurrency formats a float as currency
func FormatCurrency(amount float64) string {
	if amount >= 0 {
		return positiveStyle.Render(fmt.Sprintf("$%.2f", amount))
	}
	return negativeStyle.Render(fmt.Sprintf("-$%.2f", -amount))
}

// FormatPercent formats a float as percentage with color
func FormatPercent(value float64) string {
	if value > 0 {
		return positiveStyle.Render(fmt.Sprintf("+%.2f%%", value))
	} else if value < 0 {
		return negativeStyle.Render(fmt.Sprintf("%.2f%%", value))
	}
	return neutralStyle.Render(fmt.Sprintf("%.2f%%", value))
}
