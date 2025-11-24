package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"lab-monitor/pkg/app"
	"lab-monitor/pkg/exitcodes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [flags]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands: overview, server-status, offline\n")
		os.Exit(exitcodes.ArgsError)
	}

	command := os.Args[1]

	switch command {
	case "overview":
		handleOverview(os.Args[2:])
	case "server-status":
		handleServerStatus(os.Args[2:])
	case "offline":
		handleOffline(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Commands: overview, server-status, offline\n")
		os.Exit(exitcodes.ArgsError)
	}
}

type OverviewResponse struct {
	TotalServers   int              `json:"total_servers"`
	OnlineServers  int              `json:"online_servers"`
	OfflineServers int              `json:"offline_servers"`
	Servers        []ServerSummary  `json:"servers"`
}

type ServerSummary struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IPAddress string `json:"ip_address"`
	Status    string `json:"status"`
	LastSeen  string `json:"last_seen,omitempty"`
}

type ServerStatusResponse struct {
	ID            int64         `json:"id"`
	Name          string        `json:"name"`
	IPAddress     string        `json:"ip_address"`
	SSHUser       string        `json:"ssh_user"`
	SSHPort       int           `json:"ssh_port"`
	Status        string        `json:"status"`
	LastSeen      string        `json:"last_seen,omitempty"`
	Notes         string        `json:"notes,omitempty"`
	RecentChecks  []CheckSummary `json:"recent_checks"`
}

type CheckSummary struct {
	CheckType      string `json:"check_type"`
	Status         string `json:"status"`
	ResponseTimeMS int    `json:"response_time_ms"`
	Timestamp      string `json:"timestamp"`
	ErrorMessage   string `json:"error_message,omitempty"`
}

func handleOverview(args []string) {
	fs := flag.NewFlagSet("overview", flag.ExitOnError)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.ArgsError)
	}

	// Initialize database
	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	// Get all servers
	servers, err := database.ListServers("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	response := OverviewResponse{
		TotalServers: len(servers),
		Servers:      make([]ServerSummary, 0, len(servers)),
	}

	for _, server := range servers {
		summary := ServerSummary{
			ID:        server.ID,
			Name:      server.Name,
			IPAddress: server.IPAddress,
			Status:    server.Status,
		}

		if server.LastSeen != nil {
			summary.LastSeen = server.LastSeen.Format("2006-01-02 15:04:05")
		}

		response.Servers = append(response.Servers, summary)

		// Count by status
		switch server.Status {
		case "online":
			response.OnlineServers++
		case "offline":
			response.OfflineServers++
		}
	}

	// Output JSON
	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleServerStatus(args []string) {
	fs := flag.NewFlagSet("server-status", flag.ExitOnError)
	id := fs.Int64("id", 0, "Server ID (required)")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.ArgsError)
	}

	if *id == 0 {
		fmt.Fprintf(os.Stderr, "Error: --id is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	// Initialize database
	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	// Get server
	server, err := database.GetServer(*id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.NotFound)
	}

	// Get recent health checks
	checks, err := database.GetHealthCheckHistory(*id, 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	response := ServerStatusResponse{
		ID:           server.ID,
		Name:         server.Name,
		IPAddress:    server.IPAddress,
		SSHUser:      server.SSHUser,
		SSHPort:      server.SSHPort,
		Status:       server.Status,
		Notes:        server.Notes,
		RecentChecks: make([]CheckSummary, 0, len(checks)),
	}

	if server.LastSeen != nil {
		response.LastSeen = server.LastSeen.Format("2006-01-02 15:04:05")
	}

	for _, check := range checks {
		summary := CheckSummary{
			CheckType:      check.CheckType,
			Status:         check.Status,
			ResponseTimeMS: check.ResponseTimeMS,
			Timestamp:      check.Timestamp.Format("2006-01-02 15:04:05"),
			ErrorMessage:   check.ErrorMessage,
		}
		response.RecentChecks = append(response.RecentChecks, summary)
	}

	// Output JSON
	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleOffline(args []string) {
	fs := flag.NewFlagSet("offline", flag.ExitOnError)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.ArgsError)
	}

	// Initialize database
	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	// Get offline servers
	servers, err := database.ListServers("offline")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	response := struct {
		OfflineServers []ServerSummary `json:"offline_servers"`
		Count          int             `json:"count"`
	}{
		OfflineServers: make([]ServerSummary, 0, len(servers)),
		Count:          len(servers),
	}

	for _, server := range servers {
		summary := ServerSummary{
			ID:        server.ID,
			Name:      server.Name,
			IPAddress: server.IPAddress,
			Status:    server.Status,
		}

		if server.LastSeen != nil {
			summary.LastSeen = server.LastSeen.Format("2006-01-02 15:04:05")
		}

		response.OfflineServers = append(response.OfflineServers, summary)
	}

	// Output JSON
	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}
