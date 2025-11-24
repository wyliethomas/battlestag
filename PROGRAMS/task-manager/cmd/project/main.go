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
	case "contexts":
		handleContexts(args)
	case "add":
		handleAdd(args)
	case "list":
		handleList(args)
	case "update":
		handleUpdate(args)
	case "archive":
		handleArchive(args)
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
	fmt.Fprintf(os.Stderr, `Usage: task_project <command> [options]

Commands:
  contexts    List or add contexts
  add         Add a new project
  list        List projects
  update      Update project status/goal
  archive     Mark project complete
  help        Show this help message

Examples:
  task_project contexts --list
  task_project contexts --add "property"
  task_project add --name "Sawmill" --context "property" --goal "Fully operational sawmill"
  task_project list
  task_project list --context "property"
  task_project list --status active
  task_project update --id 1 --status blocked
  task_project update --id 1 --goal "Updated goal description"
  task_project archive --id 1

Exit Codes:
  0 - Success
  1 - Invalid arguments
  2 - Database error
  3 - Not found
`)
}

func handleContexts(args []string) {
	fs := flag.NewFlagSet("contexts", flag.ExitOnError)
	list := fs.Bool("list", false, "List all contexts")
	add := fs.String("add", "", "Add a new context")

	fs.Parse(args)

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	if *add != "" {
		id, err := database.AddContext(*add)
		if err != nil {
			if err.Error() == "context already exists" {
				fmt.Fprintf(os.Stderr, "Error: Context '%s' already exists\n", *add)
				os.Exit(exitcodes.ArgsError)
			}
			fmt.Fprintf(os.Stderr, "Failed to add context: %v\n", err)
			os.Exit(exitcodes.DBError)
		}
		fmt.Printf("Context added successfully (ID: %d)\n", id)
		os.Exit(exitcodes.Success)
	}

	if *list {
		contexts, err := database.ListContexts()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list contexts: %v\n", err)
			os.Exit(exitcodes.DBError)
		}

		if len(contexts) == 0 {
			fmt.Println("No contexts found")
			os.Exit(exitcodes.Success)
		}

		fmt.Println("Contexts:")
		for _, ctx := range contexts {
			fmt.Printf("  [%d] %s\n", ctx.ID, ctx.Name)
		}
		os.Exit(exitcodes.Success)
	}

	fmt.Fprintf(os.Stderr, "Error: must specify --list or --add\n")
	os.Exit(exitcodes.ArgsError)
}

func handleAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	name := fs.String("name", "", "Project name (required)")
	context := fs.String("context", "", "Context name (required)")
	goal := fs.String("goal", "", "Project goal (optional)")

	fs.Parse(args)

	if *name == "" {
		fmt.Fprintf(os.Stderr, "Error: --name is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	if *context == "" {
		fmt.Fprintf(os.Stderr, "Error: --context is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	// Get context ID
	ctx, err := database.GetContextByName(*context)
	if err != nil {
		if err.Error() == "context not found" {
			fmt.Fprintf(os.Stderr, "Error: Context '%s' not found. Create it first with: task_project contexts --add \"%s\"\n", *context, *context)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to get context: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	id, err := database.AddProject(ctx.ID, *name, *goal)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add project: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Project added successfully (ID: %d)\n", id)
	os.Exit(exitcodes.Success)
}

func handleList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	context := fs.String("context", "", "Filter by context")
	status := fs.String("status", "", "Filter by status (active, paused, blocked, complete)")

	fs.Parse(args)

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	projects, err := database.ListProjects(*context, *status)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list projects: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found")
		os.Exit(exitcodes.Success)
	}

	fmt.Println("Projects:")
	for _, proj := range projects {
		fmt.Printf("  [%d] %s (%s) - %s\n", proj.ID, proj.Name, proj.ContextName, proj.Status)
		if proj.Goal != "" {
			fmt.Printf("      Goal: %s\n", proj.Goal)
		}
		fmt.Printf("      Last touched: %s\n", proj.LastTouched.Format("2006-01-02 15:04"))
	}
	os.Exit(exitcodes.Success)
}

func handleUpdate(args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	id := fs.Int64("id", 0, "Project ID (required)")
	status := fs.String("status", "", "New status (active, paused, blocked, complete)")
	goal := fs.String("goal", "", "New goal")

	fs.Parse(args)

	if *id == 0 {
		fmt.Fprintf(os.Stderr, "Error: --id is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	if *status == "" && *goal == "" {
		fmt.Fprintf(os.Stderr, "Error: must specify --status or --goal\n")
		os.Exit(exitcodes.ArgsError)
	}

	// Validate status if provided
	if *status != "" {
		validStatuses := map[string]bool{"active": true, "paused": true, "blocked": true, "complete": true}
		if !validStatuses[*status] {
			fmt.Fprintf(os.Stderr, "Error: invalid status '%s'. Must be one of: active, paused, blocked, complete\n", *status)
			os.Exit(exitcodes.ArgsError)
		}
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	var statusPtr, goalPtr *string
	if *status != "" {
		statusPtr = status
	}
	if *goal != "" {
		goalPtr = goal
	}

	err = database.UpdateProject(*id, statusPtr, goalPtr)
	if err != nil {
		if err.Error() == "project not found" {
			fmt.Fprintf(os.Stderr, "Error: Project ID %d not found\n", *id)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to update project: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Project %d updated successfully\n", *id)
	os.Exit(exitcodes.Success)
}

func handleArchive(args []string) {
	fs := flag.NewFlagSet("archive", flag.ExitOnError)
	id := fs.Int64("id", 0, "Project ID (required)")

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

	err = database.ArchiveProject(*id)
	if err != nil {
		if err.Error() == "project not found" {
			fmt.Fprintf(os.Stderr, "Error: Project ID %d not found\n", *id)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to archive project: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Project %d archived successfully\n", *id)
	os.Exit(exitcodes.Success)
}
