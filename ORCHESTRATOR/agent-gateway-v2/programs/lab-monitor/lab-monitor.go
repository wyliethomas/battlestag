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
	Command    string `json:"command"`
	ServerID   int    `json:"server_id,omitempty"`
	ServerName string `json:"server_name,omitempty"`
	Name       string `json:"name,omitempty"`
	IP         string `json:"ip,omitempty"`
	SSHUser    string `json:"ssh_user,omitempty"`
	SSHPort    int    `json:"ssh_port,omitempty"`
	Notes      string `json:"notes,omitempty"`
	Status     string `json:"status,omitempty"`
	Limit      int    `json:"limit,omitempty"`
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

	// Build and execute command
	var cmd *exec.Cmd
	var cmdArgs []string

	switch input.Command {
	// ==================== Server Management ====================
	case "add-server":
		if input.Name == "" || input.IP == "" {
			fmt.Fprintf(os.Stderr, "Error: name and ip required for add-server\n")
			os.Exit(1)
		}
		cmdArgs = []string{"add", "--name", input.Name, "--ip", input.IP}
		if input.SSHUser != "" {
			cmdArgs = append(cmdArgs, "--ssh-user", input.SSHUser)
		}
		if input.SSHPort > 0 {
			cmdArgs = append(cmdArgs, "--ssh-port", fmt.Sprintf("%d", input.SSHPort))
		}
		if input.Notes != "" {
			cmdArgs = append(cmdArgs, "--notes", input.Notes)
		}
		cmd = exec.Command("lab_server", cmdArgs...)

	case "list-servers":
		cmdArgs = []string{"list"}
		if input.Status != "" {
			cmdArgs = append(cmdArgs, "--status", input.Status)
		}
		cmd = exec.Command("lab_server", cmdArgs...)

	case "update-server":
		if input.ServerID == 0 {
			fmt.Fprintf(os.Stderr, "Error: server_id required for update-server\n")
			os.Exit(1)
		}
		cmdArgs = []string{"update", "--id", fmt.Sprintf("%d", input.ServerID)}
		if input.Status != "" {
			cmdArgs = append(cmdArgs, "--status", input.Status)
		}
		if input.Notes != "" {
			cmdArgs = append(cmdArgs, "--notes", input.Notes)
		}
		if len(cmdArgs) == 3 {
			fmt.Fprintf(os.Stderr, "Error: status or notes required for update-server\n")
			os.Exit(1)
		}
		cmd = exec.Command("lab_server", cmdArgs...)

	case "remove-server":
		if input.ServerID == 0 {
			fmt.Fprintf(os.Stderr, "Error: server_id required for remove-server\n")
			os.Exit(1)
		}
		cmd = exec.Command("lab_server", "remove", "--id", fmt.Sprintf("%d", input.ServerID))

	// ==================== Health Check Commands ====================
	case "check-server":
		if input.ServerID == 0 && input.ServerName == "" {
			fmt.Fprintf(os.Stderr, "Error: either server_id or server_name required for check-server\n")
			os.Exit(1)
		}
		if input.ServerID != 0 && input.ServerName != "" {
			fmt.Fprintf(os.Stderr, "Error: cannot specify both server_id and server_name\n")
			os.Exit(1)
		}

		if input.ServerID != 0 {
			cmd = exec.Command("lab_health", "check-server", "--id", fmt.Sprintf("%d", input.ServerID))
		} else {
			cmd = exec.Command("lab_health", "check-server", "--name", input.ServerName)
		}

	case "check-all":
		cmd = exec.Command("lab_health", "check-all")

	case "health-history":
		if input.ServerID == 0 && input.ServerName == "" {
			fmt.Fprintf(os.Stderr, "Error: either server_id or server_name required for health-history\n")
			os.Exit(1)
		}
		if input.ServerID != 0 && input.ServerName != "" {
			fmt.Fprintf(os.Stderr, "Error: cannot specify both server_id and server_name\n")
			os.Exit(1)
		}

		cmdArgs = []string{"history"}
		if input.ServerID != 0 {
			cmdArgs = append(cmdArgs, "--server", fmt.Sprintf("%d", input.ServerID))
		} else {
			cmdArgs = append(cmdArgs, "--name", input.ServerName)
		}
		if input.Limit > 0 {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", input.Limit))
		}
		cmd = exec.Command("lab_health", cmdArgs...)

	// ==================== Query Commands ====================
	case "query-overview":
		cmd = exec.Command("lab_query", "overview")

	case "query-server-status":
		if input.ServerID == 0 {
			fmt.Fprintf(os.Stderr, "Error: server_id required for query-server-status\n")
			os.Exit(1)
		}
		cmd = exec.Command("lab_query", "server-status", "--id", fmt.Sprintf("%d", input.ServerID))

	case "query-offline":
		cmd = exec.Command("lab_query", "offline")

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", input.Command)
		os.Exit(1)
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// For query commands, format output for readability
	if strings.HasPrefix(input.Command, "query-") {
		// Output is likely JSON, pretty-print it
		var jsonData interface{}
		if json.Unmarshal(output, &jsonData) == nil {
			if formatted, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
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
