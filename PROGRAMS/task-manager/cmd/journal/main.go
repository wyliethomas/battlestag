package main

import (
	"flag"
	"fmt"
	"os"

	"task-manager/pkg/app"
	"task-manager/pkg/exitcodes"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(exitcodes.ArgsError)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "add":
		handleAdd(args)
	case "list":
		handleList(args)
	case "latest":
		handleLatest(args)
	case "help", "--help", "-h":
		printUsage()
		os.Exit(exitcodes.Success)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(exitcodes.ArgsError)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: task_journal <command> [options]

Commands:
  add         Add a journal entry
  list        List journal entries for a project
  latest      Show the most recent entry
  help        Show this help message

Examples:
  task_journal add --project 1 --entry "Spent 2 hours today. Got the blade mounted."
  task_journal list --project 1
  task_journal list --project 1 --limit 5
  task_journal latest --project 1

Exit Codes:
  0 - Success
  1 - Invalid arguments
  2 - Database error
  3 - Not found
`)
}

func handleAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	project := fs.Int64("project", 0, "Project ID (required)")
	entry := fs.String("entry", "", "Journal entry content (required)")

	fs.Parse(args)

	if *project == 0 {
		fmt.Fprintf(os.Stderr, "Error: --project is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	if *entry == "" {
		fmt.Fprintf(os.Stderr, "Error: --entry is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	id, err := database.AddJournalEntry(*project, *entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add journal entry: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Journal entry added successfully (ID: %d)\n", id)
	os.Exit(exitcodes.Success)
}

func handleList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	project := fs.Int64("project", 0, "Project ID (required)")
	limit := fs.Int("limit", 0, "Limit number of entries (0 = all)")

	fs.Parse(args)

	if *project == 0 {
		fmt.Fprintf(os.Stderr, "Error: --project is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	entries, err := database.ListJournalEntries(*project, *limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list journal entries: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	if len(entries) == 0 {
		fmt.Println("No journal entries found")
		os.Exit(exitcodes.Success)
	}

	fmt.Println("Journal entries:")
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		fmt.Printf("\n[%d] %s\n", entry.ID, entry.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Printf("%s\n", entry.Content)
		if i > 0 {
			fmt.Println("---")
		}
	}

	os.Exit(exitcodes.Success)
}

func handleLatest(args []string) {
	fs := flag.NewFlagSet("latest", flag.ExitOnError)
	project := fs.Int64("project", 0, "Project ID (required)")

	fs.Parse(args)

	if *project == 0 {
		fmt.Fprintf(os.Stderr, "Error: --project is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	entry, err := database.GetLatestJournalEntry(*project)
	if err != nil {
		if err.Error() == "no journal entries found" {
			fmt.Fprintf(os.Stderr, "No journal entries found for project %d\n", *project)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to get latest entry: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Latest entry [%d] - %s:\n", entry.ID, entry.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Printf("%s\n", entry.Content)

	os.Exit(exitcodes.Success)
}
