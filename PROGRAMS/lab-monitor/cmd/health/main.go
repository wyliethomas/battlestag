package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"lab-monitor/pkg/app"
	"lab-monitor/pkg/exitcodes"
	"lab-monitor/pkg/network"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [flags]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands: check-server, check-all, history\n")
		os.Exit(exitcodes.ArgsError)
	}

	command := os.Args[1]

	switch command {
	case "check-server":
		handleCheckServer(os.Args[2:])
	case "check-all":
		handleCheckAll(os.Args[2:])
	case "history":
		handleHistory(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Commands: check-server, check-all, history\n")
		os.Exit(exitcodes.ArgsError)
	}
}

func handleCheckServer(args []string) {
	fs := flag.NewFlagSet("check-server", flag.ExitOnError)
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

	fmt.Printf("Checking server '%s' (%s)...\n", server.Name, server.IPAddress)

	ctx := context.Background()

	// Perform ping check
	fmt.Print("  Ping: ")
	pingResult := network.Ping(ctx, server.IPAddress, 5*time.Second)
	if pingResult.Success {
		fmt.Printf("✓ success (%dms)\n", pingResult.ResponseTime)
		database.AddHealthCheck(server.ID, "ping", "success", pingResult.ResponseTime, "")
	} else {
		fmt.Printf("✗ failed - %s\n", pingResult.ErrorMessage)
		database.AddHealthCheck(server.ID, "ping", "failure", 0, pingResult.ErrorMessage)
	}

	// Perform SSH check
	fmt.Print("  SSH:  ")
	sshResult := network.CheckSSH(ctx, server.IPAddress, server.SSHUser, server.SSHPort, 10*time.Second)
	if sshResult.Success {
		fmt.Printf("✓ success (%dms)\n", sshResult.ResponseTime)
		database.AddHealthCheck(server.ID, "ssh", "success", sshResult.ResponseTime, "")
	} else {
		fmt.Printf("✗ failed - %s\n", sshResult.ErrorMessage)
		database.AddHealthCheck(server.ID, "ssh", "failure", 0, sshResult.ErrorMessage)
	}

	// Update server status
	overallStatus := "offline"
	if pingResult.Success && sshResult.Success {
		overallStatus = "online"
	} else if pingResult.Success || sshResult.Success {
		overallStatus = "unknown" // Partial connectivity
	}

	database.UpdateServerStatus(server.ID, overallStatus)

	fmt.Printf("\nOverall status: %s\n", overallStatus)

	os.Exit(exitcodes.Success)
}

func handleCheckAll(args []string) {
	fs := flag.NewFlagSet("check-all", flag.ExitOnError)

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

	if len(servers) == 0 {
		fmt.Println("No servers configured")
		os.Exit(exitcodes.Success)
	}

	fmt.Printf("Checking %d server(s)...\n\n", len(servers))

	ctx := context.Background()
	onlineCount := 0

	for _, server := range servers {
		fmt.Printf("[%d] %s (%s)\n", server.ID, server.Name, server.IPAddress)

		// Ping check
		fmt.Print("    Ping: ")
		pingResult := network.Ping(ctx, server.IPAddress, 5*time.Second)
		if pingResult.Success {
			fmt.Printf("✓ (%dms) ", pingResult.ResponseTime)
			database.AddHealthCheck(server.ID, "ping", "success", pingResult.ResponseTime, "")
		} else {
			fmt.Printf("✗ %s ", pingResult.ErrorMessage)
			database.AddHealthCheck(server.ID, "ping", "failure", 0, pingResult.ErrorMessage)
		}

		// SSH check
		fmt.Print("SSH: ")
		sshResult := network.CheckSSH(ctx, server.IPAddress, server.SSHUser, server.SSHPort, 10*time.Second)
		if sshResult.Success {
			fmt.Printf("✓ (%dms)\n", sshResult.ResponseTime)
			database.AddHealthCheck(server.ID, "ssh", "success", sshResult.ResponseTime, "")
		} else {
			fmt.Printf("✗ %s\n", sshResult.ErrorMessage)
			database.AddHealthCheck(server.ID, "ssh", "failure", 0, sshResult.ErrorMessage)
		}

		// Update status
		status := "offline"
		if pingResult.Success && sshResult.Success {
			status = "online"
			onlineCount++
		} else if pingResult.Success || sshResult.Success {
			status = "unknown"
		}

		database.UpdateServerStatus(server.ID, status)
	}

	fmt.Printf("\nSummary: %d/%d servers online\n", onlineCount, len(servers))

	os.Exit(exitcodes.Success)
}

func handleHistory(args []string) {
	fs := flag.NewFlagSet("history", flag.ExitOnError)
	serverID := fs.Int64("server", 0, "Server ID (required)")
	limit := fs.Int("limit", 10, "Number of recent checks to show")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.ArgsError)
	}

	if *serverID == 0 {
		fmt.Fprintf(os.Stderr, "Error: --server is required\n")
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
	server, err := database.GetServer(*serverID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.NotFound)
	}

	// Get health check history
	checks, err := database.GetHealthCheckHistory(*serverID, *limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	if len(checks) == 0 {
		fmt.Printf("No health check history for server '%s'\n", server.Name)
		os.Exit(exitcodes.Success)
	}

	fmt.Printf("Health check history for '%s' (last %d checks):\n\n", server.Name, *limit)

	for _, check := range checks {
		status := "✓"
		if check.Status != "success" {
			status = "✗"
		}

		fmt.Printf("%s %s [%s] - %s", status, check.Timestamp.Format("2006-01-02 15:04:05"), check.CheckType, check.Status)

		if check.ResponseTimeMS > 0 {
			fmt.Printf(" (%dms)", check.ResponseTimeMS)
		}

		if check.ErrorMessage != "" {
			fmt.Printf(" - %s", check.ErrorMessage)
		}

		fmt.Println()
	}

	os.Exit(exitcodes.Success)
}
