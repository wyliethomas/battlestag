package programs

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// TaskManagerProgram provides access to the task management system
type TaskManagerProgram struct{}

// NewTaskManagerProgram creates a new task manager program
func NewTaskManagerProgram() *TaskManagerProgram {
	return &TaskManagerProgram{}
}

// Info returns the program's metadata
func (p *TaskManagerProgram) Info() ProgramInfo {
	return ProgramInfo{
		ID:          "task-manager",
		Name:        "Task Manager",
		Description: "Hierarchical project management system for tracking long-running tasks across life domains (property, house, pxp, personal). Manages projects with journal logs and checklists. Use this for task tracking, project status, and intelligent work suggestions.",
		Category:    "tasks",
		Parameters: []Parameter{
			{
				Name:        "command",
				Type:        "string",
				Description: "Command to execute: 'list-contexts', 'add-context', 'add-project', 'list-projects', 'update-project', 'add-checklist', 'list-checklist', 'check-item', 'uncheck-item', 'add-journal', 'list-journal', 'latest-journal', 'query-status', 'query-overview', 'query-suggest', 'query-neglected', 'query-closing', 'query-context'",
				Required:    true,
			},
			{
				Name:        "context",
				Type:        "string",
				Description: "Context name (property, house, pxp, personal) - used for add-context, add-project, list-projects, query-overview, query-suggest, query-context",
				Required:    false,
			},
			{
				Name:        "project_id",
				Type:        "int",
				Description: "Project ID - used for update-project, add-checklist, list-checklist, add-journal, list-journal, latest-journal, query-status",
				Required:    false,
			},
			{
				Name:        "project_name",
				Type:        "string",
				Description: "Project name - used for add-project or query-status (alternative to project_id)",
				Required:    false,
			},
			{
				Name:        "goal",
				Type:        "string",
				Description: "Project goal - used for add-project, update-project",
				Required:    false,
			},
			{
				Name:        "status",
				Type:        "string",
				Description: "Project status (active, paused, blocked, complete) - used for update-project, list-projects",
				Required:    false,
			},
			{
				Name:        "items",
				Type:        "string",
				Description: "Comma-separated checklist items - used for add-checklist",
				Required:    false,
			},
			{
				Name:        "item_id",
				Type:        "int",
				Description: "Checklist item ID - used for check-item, uncheck-item",
				Required:    false,
			},
			{
				Name:        "entry",
				Type:        "string",
				Description: "Journal entry content - used for add-journal",
				Required:    false,
			},
			{
				Name:        "limit",
				Type:        "int",
				Description: "Limit number of results - used for list-journal",
				Required:    false,
			},
			{
				Name:        "pending_only",
				Type:        "bool",
				Description: "Show only incomplete items - used for list-checklist",
				Required:    false,
			},
			{
				Name:        "time_available",
				Type:        "string",
				Description: "Time available (e.g., '3 days', 'weekend') - used for query-suggest",
				Required:    false,
			},
			{
				Name:        "days",
				Type:        "int",
				Description: "Number of days - used for query-neglected (default: 14)",
				Required:    false,
			},
			{
				Name:        "threshold",
				Type:        "int",
				Description: "Completion threshold percentage - used for query-closing (default: 70)",
				Required:    false,
			},
		},
	}
}

// Execute runs the task manager command by translating API parameters
// into calls to the installed task-manager executables (task_project_run, etc.)
//
// All context parameters are normalized to lowercase to handle LLM capitalization
// variations (e.g., "Property" → "property").
//
// Returns ExecutionResult with output from the command execution.
func (p *TaskManagerProgram) Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error) {
	command, ok := params["command"].(string)
	if !ok {
		return &ExecutionResult{
			Success: false,
			Error:   "command parameter must be a string",
		}, fmt.Errorf("invalid command parameter")
	}

	// Build the appropriate task_* command based on the command parameter
	var cmd *exec.Cmd
	var cmdArgs []string

	switch command {
	// ==================== Context Management ====================
	// Commands for managing top-level contexts (property, house, pxp, personal)

	case "list-contexts":
		cmd = exec.CommandContext(ctx, "task_project_run", "contexts", "--list")

	case "add-context":
		contextName := strings.ToLower(getStringParam(params, "context"))
		if contextName == "" {
			return &ExecutionResult{Success: false, Error: "context parameter required for add-context"}, fmt.Errorf("missing context")
		}
		cmd = exec.CommandContext(ctx, "task_project_run", "contexts", "--add", contextName)

	// ==================== Project Management ====================
	// Commands for creating, listing, and updating projects

	case "add-project":
		contextName := strings.ToLower(getStringParam(params, "context"))
		projectName := getStringParam(params, "project_name")
		if contextName == "" || projectName == "" {
			return &ExecutionResult{Success: false, Error: "context and project_name required for add-project"}, fmt.Errorf("missing parameters")
		}
		cmdArgs = []string{"add", "--name", projectName, "--context", contextName}
		if goal := getStringParam(params, "goal"); goal != "" {
			cmdArgs = append(cmdArgs, "--goal", goal)
		}
		cmd = exec.CommandContext(ctx, "task_project_run", cmdArgs...)

	case "list-projects":
		cmdArgs = []string{"list"}
		if contextName := strings.ToLower(getStringParam(params, "context")); contextName != "" {
			cmdArgs = append(cmdArgs, "--context", contextName)
		}
		if status := getStringParam(params, "status"); status != "" {
			cmdArgs = append(cmdArgs, "--status", status)
		}
		cmd = exec.CommandContext(ctx, "task_project_run", cmdArgs...)

	case "update-project":
		projectID := getIntParam(params, "project_id")
		if projectID == 0 {
			return &ExecutionResult{Success: false, Error: "project_id required for update-project"}, fmt.Errorf("missing project_id")
		}
		cmdArgs = []string{"update", "--id", fmt.Sprintf("%d", projectID)}
		if status := getStringParam(params, "status"); status != "" {
			cmdArgs = append(cmdArgs, "--status", status)
		}
		if goal := getStringParam(params, "goal"); goal != "" {
			cmdArgs = append(cmdArgs, "--goal", goal)
		}
		if len(cmdArgs) == 3 { // Only --id provided
			return &ExecutionResult{Success: false, Error: "status or goal required for update-project"}, fmt.Errorf("missing update parameters")
		}
		cmd = exec.CommandContext(ctx, "task_project_run", cmdArgs...)

	// ==================== Checklist Management ====================
	// Commands for managing granular checklist items within projects

	case "add-checklist":
		projectID := getIntParam(params, "project_id")
		items := getStringParam(params, "items")
		if projectID == 0 || items == "" {
			return &ExecutionResult{Success: false, Error: "project_id and items required for add-checklist"}, fmt.Errorf("missing parameters")
		}
		cmd = exec.CommandContext(ctx, "task_checklist_run", "add", "--project", fmt.Sprintf("%d", projectID), "--items", items)

	case "list-checklist":
		projectID := getIntParam(params, "project_id")
		if projectID == 0 {
			return &ExecutionResult{Success: false, Error: "project_id required for list-checklist"}, fmt.Errorf("missing project_id")
		}
		cmdArgs = []string{"list", "--project", fmt.Sprintf("%d", projectID)}
		if getBoolParam(params, "pending_only") {
			cmdArgs = append(cmdArgs, "--pending")
		}
		cmd = exec.CommandContext(ctx, "task_checklist_run", cmdArgs...)

	case "check-item":
		itemID := getIntParam(params, "item_id")
		if itemID == 0 {
			return &ExecutionResult{Success: false, Error: "item_id required for check-item"}, fmt.Errorf("missing item_id")
		}
		cmd = exec.CommandContext(ctx, "task_checklist_run", "check", "--id", fmt.Sprintf("%d", itemID))

	case "uncheck-item":
		itemID := getIntParam(params, "item_id")
		if itemID == 0 {
			return &ExecutionResult{Success: false, Error: "item_id required for uncheck-item"}, fmt.Errorf("missing item_id")
		}
		cmd = exec.CommandContext(ctx, "task_checklist_run", "uncheck", "--id", fmt.Sprintf("%d", itemID))

	// ==================== Journal Management ====================
	// Commands for adding and viewing narrative journal entries

	case "add-journal":
		projectID := getIntParam(params, "project_id")
		entry := getStringParam(params, "entry")
		if projectID == 0 || entry == "" {
			return &ExecutionResult{Success: false, Error: "project_id and entry required for add-journal"}, fmt.Errorf("missing parameters")
		}
		cmd = exec.CommandContext(ctx, "task_journal_run", "add", "--project", fmt.Sprintf("%d", projectID), "--entry", entry)

	case "list-journal":
		projectID := getIntParam(params, "project_id")
		if projectID == 0 {
			return &ExecutionResult{Success: false, Error: "project_id required for list-journal"}, fmt.Errorf("missing project_id")
		}
		cmdArgs = []string{"list", "--project", fmt.Sprintf("%d", projectID)}
		if limit := getIntParam(params, "limit"); limit > 0 {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
		}
		cmd = exec.CommandContext(ctx, "task_journal_run", cmdArgs...)

	case "latest-journal":
		projectID := getIntParam(params, "project_id")
		if projectID == 0 {
			return &ExecutionResult{Success: false, Error: "project_id required for latest-journal"}, fmt.Errorf("missing project_id")
		}
		cmd = exec.CommandContext(ctx, "task_journal_run", "latest", "--project", fmt.Sprintf("%d", projectID))

	// ==================== Query Commands ====================
	// Intelligent queries that return JSON output for LLM consumption
	// These provide structured data for AI-powered suggestions and analysis

	case "query-status":
		projectID := getIntParam(params, "project_id")
		projectName := getStringParam(params, "project_name")
		if projectID == 0 && projectName == "" {
			return &ExecutionResult{Success: false, Error: "project_id or project_name required for query-status"}, fmt.Errorf("missing project identifier")
		}
		if projectID > 0 {
			cmd = exec.CommandContext(ctx, "task_query_run", "status", "--project", fmt.Sprintf("%d", projectID))
		} else {
			cmd = exec.CommandContext(ctx, "task_query_run", "status", "--name", projectName)
		}

	case "query-overview":
		cmdArgs = []string{"overview"}
		if contextName := strings.ToLower(getStringParam(params, "context")); contextName != "" {
			cmdArgs = append(cmdArgs, "--context", contextName)
		}
		cmd = exec.CommandContext(ctx, "task_query_run", cmdArgs...)

	case "query-suggest":
		cmdArgs = []string{"suggest"}
		if contextName := strings.ToLower(getStringParam(params, "context")); contextName != "" {
			cmdArgs = append(cmdArgs, "--context", contextName)
		}
		if timeAvailable := getStringParam(params, "time_available"); timeAvailable != "" {
			cmdArgs = append(cmdArgs, "--time", timeAvailable)
		}
		cmd = exec.CommandContext(ctx, "task_query_run", cmdArgs...)

	case "query-neglected":
		days := getIntParam(params, "days")
		if days == 0 {
			days = 14
		}
		cmd = exec.CommandContext(ctx, "task_query_run", "neglected", "--days", fmt.Sprintf("%d", days))

	case "query-closing":
		threshold := getIntParam(params, "threshold")
		if threshold == 0 {
			threshold = 70
		}
		cmd = exec.CommandContext(ctx, "task_query_run", "closing", "--threshold", fmt.Sprintf("%d", threshold))

	case "query-context":
		contextName := strings.ToLower(getStringParam(params, "context"))
		if contextName == "" {
			return &ExecutionResult{Success: false, Error: "context required for query-context"}, fmt.Errorf("missing context")
		}
		cmd = exec.CommandContext(ctx, "task_query_run", "context", "--name", contextName)

	default:
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("unknown command: %s", command),
		}, fmt.Errorf("unknown command")
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &ExecutionResult{
			Success: false,
			Output:  outputStr,
			Error:   err.Error(),
		}, err
	}

	// For query commands, try to pretty-print JSON
	if strings.HasPrefix(command, "query-") {
		if formatted := formatJSON(outputStr); formatted != "" {
			outputStr = formatted
		}
	}

	return &ExecutionResult{
		Success: true,
		Output:  outputStr,
	}, nil
}

// ==================== Helper Functions ====================
// Parameter extraction helpers (getStringParam, getIntParam, getBoolParam)
// are now in programs/helpers.go for reuse across all programs.

// formatJSON attempts to parse and pretty-print JSON output.
// Used for query commands to make JSON output more readable.
// Returns empty string if input isn't valid JSON.
func formatJSON(input string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return ""
	}
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return ""
	}
	return string(formatted)
}
