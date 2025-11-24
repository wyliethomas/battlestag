package main

import (
	"flag"
	"fmt"
	"os"

	"lab-monitor/pkg/app"
	"lab-monitor/pkg/exitcodes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [flags]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands: add, list, update, remove\n")
		os.Exit(exitcodes.ArgsError)
	}

	command := os.Args[1]

	switch command {
	case "add":
		handleAdd(os.Args[2:])
	case "list":
		handleList(os.Args[2:])
	case "update":
		handleUpdate(os.Args[2:])
	case "remove":
		handleRemove(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Commands: add, list, update, remove\n")
		os.Exit(exitcodes.ArgsError)
	}
}

func handleAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	name := fs.String("name", "", "Server name (required)")
	ip := fs.String("ip", "", "IP address (required)")
	sshUser := fs.String("ssh-user", "", "SSH username (required)")
	sshPort := fs.Int("ssh-port", 22, "SSH port (default: 22)")
	notes := fs.String("notes", "", "Optional notes")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.ArgsError)
	}

	// Validate required parameters
	if *name == "" || *ip == "" || *sshUser == "" {
		fmt.Fprintf(os.Stderr, "Error: --name, --ip, and --ssh-user are required\n")
		os.Exit(exitcodes.ArgsError)
	}

	// Initialize database
	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	// Add server
	id, err := database.AddServer(*name, *ip, *sshUser, *sshPort, *notes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Server '%s' added successfully with ID %d\n", *name, id)
	fmt.Printf("IP: %s, SSH: %s@%s:%d\n", *ip, *sshUser, *ip, *sshPort)
	if *notes != "" {
		fmt.Printf("Notes: %s\n", *notes)
	}

	os.Exit(exitcodes.Success)
}

func handleList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	statusFilter := fs.String("status", "", "Filter by status (online, offline, unknown)")

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

	// List servers
	servers, err := database.ListServers(*statusFilter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	if len(servers) == 0 {
		if *statusFilter != "" {
			fmt.Printf("No servers with status '%s'\n", *statusFilter)
		} else {
			fmt.Println("No servers configured")
		}
		os.Exit(exitcodes.Success)
	}

	fmt.Println("Servers:")
	for _, server := range servers {
		fmt.Printf("  [%d] %s (%s) - %s\n", server.ID, server.Name, server.IPAddress, server.Status)
		fmt.Printf("      SSH: %s@%s:%d\n", server.SSHUser, server.IPAddress, server.SSHPort)
		if server.LastSeen != nil {
			fmt.Printf("      Last seen: %s\n", server.LastSeen.Format("2006-01-02 15:04"))
		} else {
			fmt.Printf("      Last seen: never\n")
		}
		if server.Notes != "" {
			fmt.Printf("      Notes: %s\n", server.Notes)
		}
	}

	os.Exit(exitcodes.Success)
}

func handleUpdate(args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	id := fs.Int64("id", 0, "Server ID (required)")
	status := fs.String("status", "", "New status (online, offline, unknown)")
	notes := fs.String("notes", "", "New notes")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.ArgsError)
	}

	// Validate required parameters
	if *id == 0 {
		fmt.Fprintf(os.Stderr, "Error: --id is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	if *status == "" && *notes == "" {
		fmt.Fprintf(os.Stderr, "Error: at least one of --status or --notes must be provided\n")
		os.Exit(exitcodes.ArgsError)
	}

	// Validate status value
	if *status != "" && *status != "online" && *status != "offline" && *status != "unknown" {
		fmt.Fprintf(os.Stderr, "Error: status must be 'online', 'offline', or 'unknown'\n")
		os.Exit(exitcodes.ArgsError)
	}

	// Initialize database
	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	// Get server to verify it exists
	server, err := database.GetServer(*id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.NotFound)
	}

	// Update fields
	if *status != "" {
		if err := database.UpdateServerStatus(*id, *status); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating status: %v\n", err)
			os.Exit(exitcodes.DBError)
		}
		fmt.Printf("Updated status to '%s'\n", *status)
	}

	if *notes != "" {
		if err := database.UpdateServerNotes(*id, *notes); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating notes: %v\n", err)
			os.Exit(exitcodes.DBError)
		}
		fmt.Printf("Updated notes\n")
	}

	fmt.Printf("Server '%s' updated successfully\n", server.Name)
	os.Exit(exitcodes.Success)
}

func handleRemove(args []string) {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	id := fs.Int64("id", 0, "Server ID (required)")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.ArgsError)
	}

	// Validate required parameters
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

	// Get server to show info before deleting
	server, err := database.GetServer(*id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.NotFound)
	}

	// Remove server
	if err := database.RemoveServer(*id); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Server '%s' (%s) removed successfully\n", server.Name, server.IPAddress)
	os.Exit(exitcodes.Success)
}
