package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/battlestag/battlestag-tui/client"
	"github.com/battlestag/battlestag-tui/config"
	"github.com/battlestag/battlestag-tui/ui"
)

func main() {
	// Load or create configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Check if this is first run (no config file exists)
	showOnboarding := !config.ConfigExists()

	// Create API client from configuration
	apiClient := client.NewClientWithConfig(cfg.AgentGateway.URL, cfg.AgentGateway.APIKey)

	// Test connection before starting TUI (unless onboarding)
	if !showOnboarding {
		_, err := apiClient.Health()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not connect to API: %v\n", err)
			fmt.Fprintf(os.Stderr, "You can update settings from the Settings screen (press 5 or 's').\n\n")
		}
	}

	// Create the application model
	model := ui.NewModel(apiClient, cfg, showOnboarding)

	// Create the program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
