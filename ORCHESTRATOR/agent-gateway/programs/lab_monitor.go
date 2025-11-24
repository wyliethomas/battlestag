package programs

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// LabMonitorProgram provides access to the home lab monitoring system
type LabMonitorProgram struct{}

// NewLabMonitorProgram creates a new lab monitor program
func NewLabMonitorProgram() *LabMonitorProgram {
	return &LabMonitorProgram{}
}

// Info returns the program's metadata
func (p *LabMonitorProgram) Info() ProgramInfo {
	return ProgramInfo{
		ID:          "lab-monitor",
		Name:        "Lab Monitor",
		Description: "Home lab server monitoring and health check system. Manages servers with ping and SSH health checks. Use this for server status, uptime monitoring, and connectivity verification.",
		Category:    "infrastructure",
		Parameters: []Parameter{
			{
				Name:        "command",
				Type:        "string",
				Description: "Command to execute: 'add-server', 'list-servers', 'update-server', 'remove-server', 'check-server', 'check-all', 'health-history', 'query-overview', 'query-server-status', 'query-offline'",
				Required:    true,
			},
			{
				Name:        "server_id",
				Type:        "int",
				Description: "Server ID - used for update-server, remove-server, check-server, health-history, query-server-status",
				Required:    false,
			},
			{
				Name:        "name",
				Type:        "string",
				Description: "Server name - used for add-server",
				Required:    false,
			},
			{
				Name:        "ip",
				Type:        "string",
				Description: "Server IP address - used for add-server",
				Required:    false,
			},
			{
				Name:        "ssh_user",
				Type:        "string",
				Description: "SSH username - used for add-server (default: current user)",
				Required:    false,
			},
			{
				Name:        "ssh_port",
				Type:        "int",
				Description: "SSH port - used for add-server (default: 22)",
				Required:    false,
			},
			{
				Name:        "notes",
				Type:        "string",
				Description: "Server notes - used for add-server, update-server",
				Required:    false,
			},
			{
				Name:        "status",
				Type:        "string",
				Description: "Server status filter (online, offline, unknown) - used for list-servers",
				Required:    false,
			},
			{
				Name:        "limit",
				Type:        "int",
				Description: "Limit number of health check results - used for health-history (default: 10)",
				Required:    false,
			},
		},
	}
}

// Execute runs the lab monitor command by translating API parameters
// into calls to the installed lab-monitor executables (lab_server, lab_health, lab_query)
//
// Returns ExecutionResult with output from the command execution.
func (p *LabMonitorProgram) Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error) {
	command, ok := params["command"].(string)
	if !ok {
		return &ExecutionResult{
			Success: false,
			Error:   "command parameter must be a string",
		}, fmt.Errorf("invalid command parameter")
	}

	// Build the appropriate lab_* command based on the command parameter
	var cmd *exec.Cmd
	var cmdArgs []string

	switch command {
	// ==================== Server Management ====================
	// Commands for managing lab servers (add, list, update, remove)

	case "add-server":
		name := getStringParam(params, "name")
		ip := getStringParam(params, "ip")
		if name == "" || ip == "" {
			return &ExecutionResult{Success: false, Error: "name and ip required for add-server"}, fmt.Errorf("missing parameters")
		}
		cmdArgs = []string{"add", "--name", name, "--ip", ip}
		if sshUser := getStringParam(params, "ssh_user"); sshUser != "" {
			cmdArgs = append(cmdArgs, "--ssh-user", sshUser)
		}
		if sshPort := getIntParam(params, "ssh_port"); sshPort > 0 {
			cmdArgs = append(cmdArgs, "--ssh-port", fmt.Sprintf("%d", sshPort))
		}
		if notes := getStringParam(params, "notes"); notes != "" {
			cmdArgs = append(cmdArgs, "--notes", notes)
		}
		cmd = exec.CommandContext(ctx, "lab_server", cmdArgs...)

	case "list-servers":
		cmdArgs = []string{"list"}
		if status := getStringParam(params, "status"); status != "" {
			cmdArgs = append(cmdArgs, "--status", status)
		}
		cmd = exec.CommandContext(ctx, "lab_server", cmdArgs...)

	case "update-server":
		serverID := getIntParam(params, "server_id")
		if serverID == 0 {
			return &ExecutionResult{Success: false, Error: "server_id required for update-server"}, fmt.Errorf("missing server_id")
		}
		cmdArgs = []string{"update", "--id", fmt.Sprintf("%d", serverID)}
		if status := getStringParam(params, "status"); status != "" {
			cmdArgs = append(cmdArgs, "--status", status)
		}
		if notes := getStringParam(params, "notes"); notes != "" {
			cmdArgs = append(cmdArgs, "--notes", notes)
		}
		if len(cmdArgs) == 3 { // Only --id provided
			return &ExecutionResult{Success: false, Error: "status or notes required for update-server"}, fmt.Errorf("missing update parameters")
		}
		cmd = exec.CommandContext(ctx, "lab_server", cmdArgs...)

	case "remove-server":
		serverID := getIntParam(params, "server_id")
		if serverID == 0 {
			return &ExecutionResult{Success: false, Error: "server_id required for remove-server"}, fmt.Errorf("missing server_id")
		}
		cmd = exec.CommandContext(ctx, "lab_server", "remove", "--id", fmt.Sprintf("%d", serverID))

	// ==================== Health Check Commands ====================
	// Commands for running health checks and viewing history

	case "check-server":
		serverID := getIntParam(params, "server_id")
		if serverID == 0 {
			return &ExecutionResult{Success: false, Error: "server_id required for check-server"}, fmt.Errorf("missing server_id")
		}
		cmd = exec.CommandContext(ctx, "lab_health", "check-server", "--id", fmt.Sprintf("%d", serverID))

	case "check-all":
		cmd = exec.CommandContext(ctx, "lab_health", "check-all")

	case "health-history":
		serverID := getIntParam(params, "server_id")
		if serverID == 0 {
			return &ExecutionResult{Success: false, Error: "server_id required for health-history"}, fmt.Errorf("missing server_id")
		}
		cmdArgs = []string{"history", "--server", fmt.Sprintf("%d", serverID)}
		if limit := getIntParam(params, "limit"); limit > 0 {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
		}
		cmd = exec.CommandContext(ctx, "lab_health", cmdArgs...)

	// ==================== Query Commands ====================
	// Intelligent queries that return JSON output for LLM consumption
	// These provide structured data for AI-powered monitoring and analysis

	case "query-overview":
		cmd = exec.CommandContext(ctx, "lab_query", "overview")

	case "query-server-status":
		serverID := getIntParam(params, "server_id")
		if serverID == 0 {
			return &ExecutionResult{Success: false, Error: "server_id required for query-server-status"}, fmt.Errorf("missing server_id")
		}
		cmd = exec.CommandContext(ctx, "lab_query", "server-status", "--id", fmt.Sprintf("%d", serverID))

	case "query-offline":
		cmd = exec.CommandContext(ctx, "lab_query", "offline")

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
		if formatted := formatJSONLabMonitor(outputStr); formatted != "" {
			outputStr = formatted
		}
	}

	return &ExecutionResult{
		Success: true,
		Output:  outputStr,
	}, nil
}

// formatJSONLabMonitor attempts to parse and pretty-print JSON output.
// Used for query commands to make JSON output more readable.
// Returns empty string if input isn't valid JSON.
func formatJSONLabMonitor(input string) string {
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
