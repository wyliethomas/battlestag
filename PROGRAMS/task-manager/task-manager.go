package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Input represents the JSON input structure
type Input struct {
	Command       string `json:"command"`
	Context       string `json:"context,omitempty"`
	ProjectID     int    `json:"project_id,omitempty"`
	ProjectName   string `json:"project_name,omitempty"`
	Goal          string `json:"goal,omitempty"`
	Status        string `json:"status,omitempty"`
	Items         string `json:"items,omitempty"`
	ItemID        int    `json:"item_id,omitempty"`
	Entry         string `json:"entry,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	PendingOnly   bool   `json:"pending_only,omitempty"`
	TimeAvailable string `json:"time_available,omitempty"`
	Days          int    `json:"days,omitempty"`
	Threshold     int    `json:"threshold,omitempty"`
}

func main() {
	// Read JSON from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Normalize context to lowercase
	input.Context = strings.ToLower(input.Context)

	// Build and execute command
	var cmd *exec.Cmd
	var cmdArgs []string

	switch input.Command {
	// ==================== Context Management ====================
	case "list-contexts":
		cmd = exec.Command("task_project_run", "contexts", "--list")

	case "add-context":
		if input.Context == "" {
			fmt.Fprintf(os.Stderr, "Error: context required for add-context\n")
			os.Exit(1)
		}
		cmd = exec.Command("task_project_run", "contexts", "--add", input.Context)

	// ==================== Project Management ====================
	case "add-project":
		if input.Context == "" || input.ProjectName == "" {
			fmt.Fprintf(os.Stderr, "Error: context and project_name required for add-project\n")
			os.Exit(1)
		}
		cmdArgs = []string{"add", "--name", input.ProjectName, "--context", input.Context}
		if input.Goal != "" {
			cmdArgs = append(cmdArgs, "--goal", input.Goal)
		}
		cmd = exec.Command("task_project_run", cmdArgs...)

	case "list-projects":
		cmdArgs = []string{"list"}
		if input.Context != "" {
			cmdArgs = append(cmdArgs, "--context", input.Context)
		}
		if input.Status != "" {
			cmdArgs = append(cmdArgs, "--status", input.Status)
		}
		cmd = exec.Command("task_project_run", cmdArgs...)

	case "update-project":
		if input.ProjectID == 0 {
			fmt.Fprintf(os.Stderr, "Error: project_id required for update-project\n")
			os.Exit(1)
		}
		cmdArgs = []string{"update", "--id", fmt.Sprintf("%d", input.ProjectID)}
		if input.Status != "" {
			cmdArgs = append(cmdArgs, "--status", input.Status)
		}
		if input.Goal != "" {
			cmdArgs = append(cmdArgs, "--goal", input.Goal)
		}
		if len(cmdArgs) == 3 {
			fmt.Fprintf(os.Stderr, "Error: status or goal required for update-project\n")
			os.Exit(1)
		}
		cmd = exec.Command("task_project_run", cmdArgs...)

	// ==================== Checklist Management ====================
	case "add-checklist":
		if input.ProjectID == 0 || input.Items == "" {
			fmt.Fprintf(os.Stderr, "Error: project_id and items required for add-checklist\n")
			os.Exit(1)
		}
		cmd = exec.Command("task_checklist_run", "add", "--project", fmt.Sprintf("%d", input.ProjectID), "--items", input.Items)

	case "list-checklist":
		if input.ProjectID == 0 {
			fmt.Fprintf(os.Stderr, "Error: project_id required for list-checklist\n")
			os.Exit(1)
		}
		cmdArgs = []string{"list", "--project", fmt.Sprintf("%d", input.ProjectID)}
		if input.PendingOnly {
			cmdArgs = append(cmdArgs, "--pending")
		}
		cmd = exec.Command("task_checklist_run", cmdArgs...)

	case "check-item":
		if input.ItemID == 0 {
			fmt.Fprintf(os.Stderr, "Error: item_id required for check-item\n")
			os.Exit(1)
		}
		cmd = exec.Command("task_checklist_run", "check", "--id", fmt.Sprintf("%d", input.ItemID))

	case "uncheck-item":
		if input.ItemID == 0 {
			fmt.Fprintf(os.Stderr, "Error: item_id required for uncheck-item\n")
			os.Exit(1)
		}
		cmd = exec.Command("task_checklist_run", "uncheck", "--id", fmt.Sprintf("%d", input.ItemID))

	// ==================== Journal Management ====================
	case "add-journal":
		if input.ProjectID == 0 || input.Entry == "" {
			fmt.Fprintf(os.Stderr, "Error: project_id and entry required for add-journal\n")
			os.Exit(1)
		}
		cmd = exec.Command("task_journal_run", "add", "--project", fmt.Sprintf("%d", input.ProjectID), "--entry", input.Entry)

	case "list-journal":
		if input.ProjectID == 0 {
			fmt.Fprintf(os.Stderr, "Error: project_id required for list-journal\n")
			os.Exit(1)
		}
		cmdArgs = []string{"list", "--project", fmt.Sprintf("%d", input.ProjectID)}
		if input.Limit > 0 {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", input.Limit))
		}
		cmd = exec.Command("task_journal_run", cmdArgs...)

	case "latest-journal":
		if input.ProjectID == 0 {
			fmt.Fprintf(os.Stderr, "Error: project_id required for latest-journal\n")
			os.Exit(1)
		}
		cmd = exec.Command("task_journal_run", "latest", "--project", fmt.Sprintf("%d", input.ProjectID))

	// ==================== Query Commands ====================
	case "query-status":
		if input.ProjectID == 0 && input.ProjectName == "" {
			fmt.Fprintf(os.Stderr, "Error: project_id or project_name required for query-status\n")
			os.Exit(1)
		}
		if input.ProjectID > 0 {
			cmd = exec.Command("task_query_run", "status", "--project", fmt.Sprintf("%d", input.ProjectID))
		} else {
			cmd = exec.Command("task_query_run", "status", "--name", input.ProjectName)
		}

	case "query-overview":
		cmdArgs = []string{"overview"}
		if input.Context != "" {
			cmdArgs = append(cmdArgs, "--context", input.Context)
		}
		cmd = exec.Command("task_query_run", cmdArgs...)

	case "query-suggest":
		cmdArgs = []string{"suggest"}
		if input.Context != "" {
			cmdArgs = append(cmdArgs, "--context", input.Context)
		}
		if input.TimeAvailable != "" {
			cmdArgs = append(cmdArgs, "--time", input.TimeAvailable)
		}
		cmd = exec.Command("task_query_run", cmdArgs...)

	case "query-neglected":
		days := input.Days
		if days == 0 {
			days = 14
		}
		cmd = exec.Command("task_query_run", "neglected", "--days", fmt.Sprintf("%d", days))

	case "query-closing":
		threshold := input.Threshold
		if threshold == 0 {
			threshold = 70
		}
		cmd = exec.Command("task_query_run", "closing", "--threshold", fmt.Sprintf("%d", threshold))

	case "query-context":
		if input.Context == "" {
			fmt.Fprintf(os.Stderr, "Error: context required for query-context\n")
			os.Exit(1)
		}
		cmd = exec.Command("task_query_run", "context", "--name", input.Context)

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", input.Command)
		os.Exit(1)
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// For query commands, try to pretty-print JSON
	if strings.HasPrefix(input.Command, "query-") {
		var jsonData interface{}
		if json.Unmarshal(output, &jsonData) == nil {
			if formatted, ferr := json.MarshalIndent(jsonData, "", "  "); ferr == nil {
				outputStr = string(formatted)
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Command failed: %v\n%s", err, outputStr)
		os.Exit(1)
	}

	fmt.Print(outputStr)
}
