package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ProgramInfo represents a program from the API
type ProgramInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// ProgramsListResponse is the API response for /api/programs/list
type ProgramsListResponse struct {
	Success bool          `json:"success"`
	Data    []ProgramInfo `json:"data"`
}

// CommandType represents the type of command
type CommandType int

const (
	CategoryCommand CommandType = iota
	ActionCommand
)

// Command represents a command in the palette
type Command struct {
	Name        string      // Display name (e.g., "Finance", "Dashboard")
	Key         string      // Command key (e.g., "finance", "finance:dashboard")
	Description string      // Description shown in palette
	Type        CommandType // Category or Action
	Category    string      // Parent category (empty for top-level)
}

// CommandRegistry holds all available commands
type CommandRegistry struct {
	commands map[string]Command
	rootKeys []string // Top-level category keys
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	reg := &CommandRegistry{
		commands: make(map[string]Command),
		rootKeys: []string{},
	}

	reg.registerCommands()
	return reg
}

// registerCommands registers all available commands
func (r *CommandRegistry) registerCommands() {
	// Top-level categories
	categories := []Command{
		{
			Name:        "💰 Finance",
			Key:         "finance",
			Description: "Financial management and tracking",
			Type:        CategoryCommand,
		},
		{
			Name:        "📖 Stoic",
			Key:         "stoic",
			Description: "Stoic thoughts and wisdom",
			Type:        CategoryCommand,
		},
		{
			Name:        "💻 Tech",
			Key:         "tech",
			Description: "Tech tips and guides",
			Type:        CategoryCommand,
		},
		{
			Name:        "⚙️  System",
			Key:         "system",
			Description: "App settings and utilities",
			Type:        CategoryCommand,
		},
	}

	for _, cat := range categories {
		r.commands[cat.Key] = cat
		r.rootKeys = append(r.rootKeys, cat.Key)
	}

	// Finance sub-commands
	financeCommands := []Command{
		{
			Name:        "Dashboard",
			Key:         "finance:dashboard",
			Description: "View financial overview and net worth",
			Type:        ActionCommand,
			Category:    "finance",
		},
		{
			Name:        "Assets",
			Key:         "finance:assets",
			Description: "List all tracked assets",
			Type:        ActionCommand,
			Category:    "finance",
		},
		{
			Name:        "Liabilities",
			Key:         "finance:liabilities",
			Description: "List all liabilities",
			Type:        ActionCommand,
			Category:    "finance",
		},
		{
			Name:        "Upload Statement",
			Key:         "finance:upload",
			Description: "Upload and process bank statement",
			Type:        ActionCommand,
			Category:    "finance",
		},
		{
			Name:        "Query",
			Key:         "finance:query",
			Description: "Ask questions about your finances",
			Type:        ActionCommand,
			Category:    "finance",
		},
	}

	for _, cmd := range financeCommands {
		r.commands[cmd.Key] = cmd
	}

	// Stoic sub-commands
	stoicCommands := []Command{
		{
			Name:        "Today's Thought",
			Key:         "stoic:today",
			Description: "Get today's stoic thought",
			Type:        ActionCommand,
			Category:    "stoic",
		},
		{
			Name:        "Random Thought",
			Key:         "stoic:random",
			Description: "Get a random stoic thought",
			Type:        ActionCommand,
			Category:    "stoic",
		},
	}

	for _, cmd := range stoicCommands {
		r.commands[cmd.Key] = cmd
	}

	// Tech sub-commands
	techCommands := []Command{
		{
			Name:        "Random Tip",
			Key:         "tech:random",
			Description: "Get a random tech tip",
			Type:        ActionCommand,
			Category:    "tech",
		},
		{
			Name:        "Latest Tip",
			Key:         "tech:latest",
			Description: "Get the latest tech tip",
			Type:        ActionCommand,
			Category:    "tech",
		},
	}

	for _, cmd := range techCommands {
		r.commands[cmd.Key] = cmd
	}

	// System sub-commands
	systemCommands := []Command{
		{
			Name:        "Programs",
			Key:         "system:programs",
			Description: "List available programs",
			Type:        ActionCommand,
			Category:    "system",
		},
		{
			Name:        "Settings",
			Key:         "system:settings",
			Description: "Configure application settings",
			Type:        ActionCommand,
			Category:    "system",
		},
		{
			Name:        "Clear Chat",
			Key:         "system:clear",
			Description: "Clear chat history",
			Type:        ActionCommand,
			Category:    "system",
		},
		{
			Name:        "Help",
			Key:         "system:help",
			Description: "Show help and documentation",
			Type:        ActionCommand,
			Category:    "system",
		},
		{
			Name:        "Quit",
			Key:         "system:quit",
			Description: "Exit the application",
			Type:        ActionCommand,
			Category:    "system",
		},
	}

	for _, cmd := range systemCommands {
		r.commands[cmd.Key] = cmd
	}
}

// GetRootCommands returns top-level categories
func (r *CommandRegistry) GetRootCommands() []Command {
	var commands []Command
	for _, key := range r.rootKeys {
		commands = append(commands, r.commands[key])
	}
	return commands
}

// GetCategoryCommands returns sub-commands for a category
func (r *CommandRegistry) GetCategoryCommands(category string) []Command {
	var commands []Command
	for _, cmd := range r.commands {
		if cmd.Category == category {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// FilterCommands filters commands based on input text
// Supports both category navigation and direct typing
func (r *CommandRegistry) FilterCommands(input string) []Command {
	// Remove leading "/"
	input = strings.TrimPrefix(input, "/")

	// Empty input = show root categories
	if input == "" {
		return r.GetRootCommands()
	}

	// Check if input contains ":" (category:command format)
	if strings.Contains(input, ":") {
		parts := strings.Split(input, ":")
		category := parts[0]
		filter := ""
		if len(parts) > 1 {
			filter = strings.ToLower(parts[1])
		}

		// Get commands in this category
		categoryCommands := r.GetCategoryCommands(category)

		// Filter by name if there's text after ":"
		if filter != "" {
			var filtered []Command
			for _, cmd := range categoryCommands {
				if strings.Contains(strings.ToLower(cmd.Name), filter) ||
					strings.Contains(strings.ToLower(cmd.Key), filter) {
					filtered = append(filtered, cmd)
				}
			}
			return filtered
		}

		return categoryCommands
	}

	// No ":" - filter categories by name
	var filtered []Command
	inputLower := strings.ToLower(input)
	for _, key := range r.rootKeys {
		cmd := r.commands[key]
		if strings.Contains(strings.ToLower(cmd.Name), inputLower) ||
			strings.Contains(strings.ToLower(cmd.Key), inputLower) {
			filtered = append(filtered, cmd)
		}
	}

	return filtered
}

// GetCommand retrieves a command by key
func (r *CommandRegistry) GetCommand(key string) (Command, bool) {
	cmd, exists := r.commands[key]
	return cmd, exists
}

// LoadProgramsFromAPI queries the agent-gateway for available programs
// and registers them as commands dynamically
func (r *CommandRegistry) LoadProgramsFromAPI(apiURL, apiKey string) error {
	// Create request
	req, err := http.NewRequest("GET", apiURL+"/api/programs/list", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key header
	req.Header.Set("X-API-Key", apiKey)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch programs: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var programsResp ProgramsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&programsResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !programsResp.Success {
		return fmt.Errorf("API returned success=false")
	}

	// Track which categories we've added
	addedCategories := make(map[string]bool)

	// Register each program as a command
	for _, prog := range programsResp.Data {
		// Add category if not already present
		if !addedCategories[prog.Category] {
			categoryKey := prog.Category
			// Check if category already exists
			if _, exists := r.commands[categoryKey]; !exists {
				// Add new category
				categoryCmd := Command{
					Name:        "📦 " + strings.Title(prog.Category),
					Key:         categoryKey,
					Description: strings.Title(prog.Category) + " programs and modules",
					Type:        CategoryCommand,
				}
				r.commands[categoryKey] = categoryCmd
				r.rootKeys = append(r.rootKeys, categoryKey)
			}
			addedCategories[prog.Category] = true
		}

		// Register program as command
		commandKey := prog.Category + ":" + prog.ID
		cmd := Command{
			Name:        prog.Name,
			Key:         commandKey,
			Description: prog.Description,
			Type:        ActionCommand,
			Category:    prog.Category,
		}
		r.commands[commandKey] = cmd
	}

	return nil
}
