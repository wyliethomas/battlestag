package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

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
	case "check":
		handleCheck(args)
	case "uncheck":
		handleUncheck(args)
	case "remove":
		handleRemove(args)
	case "reorder":
		handleReorder(args)
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
	fmt.Fprintf(os.Stderr, `Usage: task_checklist <command> [options]

Commands:
  add         Add checklist item(s) to a project
  list        List checklist items for a project
  check       Mark item complete
  uncheck     Mark item incomplete
  remove      Remove a checklist item
  reorder     Change item sort order
  help        Show this help message

Examples:
  task_checklist add --project 1 --item "Level the carriage rails"
  task_checklist add --project 1 --items "Order new blade, Research shimming methods, Pick up lumber"
  task_checklist list --project 1
  task_checklist list --project 1 --pending
  task_checklist check --id 5
  task_checklist uncheck --id 5
  task_checklist remove --id 5
  task_checklist reorder --id 5 --position 1

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
	item := fs.String("item", "", "Single item to add")
	items := fs.String("items", "", "Comma-separated items to add")

	fs.Parse(args)

	if *project == 0 {
		fmt.Fprintf(os.Stderr, "Error: --project is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	if *item == "" && *items == "" {
		fmt.Fprintf(os.Stderr, "Error: must specify --item or --items\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	// Collect all items to add
	var itemsToAdd []string
	if *item != "" {
		itemsToAdd = append(itemsToAdd, *item)
	}
	if *items != "" {
		// Split by comma and trim whitespace
		for _, i := range strings.Split(*items, ",") {
			trimmed := strings.TrimSpace(i)
			if trimmed != "" {
				itemsToAdd = append(itemsToAdd, trimmed)
			}
		}
	}

	if len(itemsToAdd) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no items to add\n")
		os.Exit(exitcodes.ArgsError)
	}

	// Add all items
	addedCount := 0
	for _, desc := range itemsToAdd {
		id, err := database.AddChecklistItem(*project, desc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add item '%s': %v\n", desc, err)
			if addedCount == 0 {
				os.Exit(exitcodes.DBError)
			}
			continue
		}
		fmt.Printf("Added item %d: %s\n", id, desc)
		addedCount++
	}

	fmt.Printf("Successfully added %d item(s)\n", addedCount)
	os.Exit(exitcodes.Success)
}

func handleList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	project := fs.Int64("project", 0, "Project ID (required)")
	pending := fs.Bool("pending", false, "Show only incomplete items")

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

	items, err := database.ListChecklistItems(*project, *pending)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list items: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	if len(items) == 0 {
		if *pending {
			fmt.Println("No pending items")
		} else {
			fmt.Println("No checklist items found")
		}
		os.Exit(exitcodes.Success)
	}

	fmt.Println("Checklist items:")
	for _, item := range items {
		status := " "
		if item.Done {
			status = "✓"
		}
		fmt.Printf("  [%s] [%d] %s\n", status, item.ID, item.Description)
		if item.Done && item.CompletedAt != nil {
			fmt.Printf("      Completed: %s\n", item.CompletedAt.Format("2006-01-02 15:04"))
		}
	}

	// Print summary
	done := 0
	for _, item := range items {
		if item.Done {
			done++
		}
	}
	fmt.Printf("\nProgress: %d/%d complete\n", done, len(items))

	os.Exit(exitcodes.Success)
}

func handleCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	id := fs.Int64("id", 0, "Item ID (required)")

	fs.Parse(args)

	if *id == 0 {
		fmt.Fprintf(os.Stderr, "Error: --id is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	err = database.CheckItem(*id)
	if err != nil {
		if err.Error() == "item not found" {
			fmt.Fprintf(os.Stderr, "Error: Item ID %d not found\n", *id)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to check item: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Item %d marked as complete\n", *id)
	os.Exit(exitcodes.Success)
}

func handleUncheck(args []string) {
	fs := flag.NewFlagSet("uncheck", flag.ExitOnError)
	id := fs.Int64("id", 0, "Item ID (required)")

	fs.Parse(args)

	if *id == 0 {
		fmt.Fprintf(os.Stderr, "Error: --id is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	err = database.UncheckItem(*id)
	if err != nil {
		if err.Error() == "item not found" {
			fmt.Fprintf(os.Stderr, "Error: Item ID %d not found\n", *id)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to uncheck item: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Item %d marked as incomplete\n", *id)
	os.Exit(exitcodes.Success)
}

func handleRemove(args []string) {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	id := fs.Int64("id", 0, "Item ID (required)")

	fs.Parse(args)

	if *id == 0 {
		fmt.Fprintf(os.Stderr, "Error: --id is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	err = database.RemoveChecklistItem(*id)
	if err != nil {
		if err.Error() == "item not found" {
			fmt.Fprintf(os.Stderr, "Error: Item ID %d not found\n", *id)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to remove item: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Item %d removed successfully\n", *id)
	os.Exit(exitcodes.Success)
}

func handleReorder(args []string) {
	fs := flag.NewFlagSet("reorder", flag.ExitOnError)
	id := fs.Int64("id", 0, "Item ID (required)")
	position := fs.Int("position", 0, "New sort position (required)")

	fs.Parse(args)

	if *id == 0 {
		fmt.Fprintf(os.Stderr, "Error: --id is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	if *position < 0 {
		fmt.Fprintf(os.Stderr, "Error: --position must be >= 0\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	err = database.ReorderChecklistItem(*id, *position)
	if err != nil {
		if err.Error() == "item not found" {
			fmt.Fprintf(os.Stderr, "Error: Item ID %d not found\n", *id)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to reorder item: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Item %d reordered successfully\n", *id)
	os.Exit(exitcodes.Success)
}
