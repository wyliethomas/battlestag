package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

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
	case "status":
		handleStatus(args)
	case "overview":
		handleOverview(args)
	case "suggest":
		handleSuggest(args)
	case "neglected":
		handleNeglected(args)
	case "closing":
		handleClosing(args)
	case "context":
		handleContext(args)
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
	fmt.Fprintf(os.Stderr, `Usage: task_query <command> [options]

Commands:
  status      Get status of a specific project
  overview    Get overview of all active projects
  suggest     Get suggestions for what to work on
  neglected   Show projects that haven't been touched recently
  closing     Show projects closest to completion
  context     Show all projects in a context with stats
  help        Show this help message

Examples:
  task_query status --project 1
  task_query status --name "sawmill"
  task_query overview
  task_query overview --context "property"
  task_query suggest --context "property" --time "3 days"
  task_query suggest --time "2 hours"
  task_query neglected --days 14
  task_query closing --threshold 70
  task_query context --name "property"

Exit Codes:
  0 - Success
  1 - Invalid arguments
  2 - Database error
  3 - Not found
`)
}

func handleStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	project := fs.Int64("project", 0, "Project ID")
	name := fs.String("name", "", "Project name")

	fs.Parse(args)

	if *project == 0 && *name == "" {
		fmt.Fprintf(os.Stderr, "Error: must specify --project or --name\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	var projectID int64
	if *name != "" {
		proj, err := database.GetProjectByName(*name)
		if err != nil {
			if err.Error() == "project not found" {
				fmt.Fprintf(os.Stderr, "Error: Project '%s' not found\n", *name)
				os.Exit(exitcodes.NotFound)
			}
			fmt.Fprintf(os.Stderr, "Failed to get project: %v\n", err)
			os.Exit(exitcodes.DBError)
		}
		projectID = proj.ID
	} else {
		projectID = *project
	}

	status, err := database.GetProjectStatus(projectID)
	if err != nil {
		if err.Error() == "project not found" {
			fmt.Fprintf(os.Stderr, "Error: Project ID %d not found\n", projectID)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to get project status: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleOverview(args []string) {
	fs := flag.NewFlagSet("overview", flag.ExitOnError)
	context := fs.String("context", "", "Filter by context")

	fs.Parse(args)

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	projects, err := database.ListProjects(*context, "active")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list projects: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	// Build overview with stats
	type ProjectOverview struct {
		ID               int64  `json:"id"`
		Name             string `json:"name"`
		Context          string `json:"context"`
		Status           string `json:"status"`
		Goal             string `json:"goal,omitempty"`
		ProgressPercent  int    `json:"progress_percent"`
		ChecklistTotal   int    `json:"checklist_total"`
		ChecklistDone    int    `json:"checklist_done"`
		DaysSinceTouched int    `json:"days_since_touched"`
	}

	overview := make([]ProjectOverview, 0, len(projects))
	for _, proj := range projects {
		total, done, _ := database.GetChecklistStats(proj.ID)
		progress := 0
		if total > 0 {
			progress = (done * 100) / total
		}

		daysSinceTouched := int(time.Since(proj.LastTouched).Hours() / 24)

		overview = append(overview, ProjectOverview{
			ID:               proj.ID,
			Name:             proj.Name,
			Context:          proj.ContextName,
			Status:           proj.Status,
			Goal:             proj.Goal,
			ProgressPercent:  progress,
			ChecklistTotal:   total,
			ChecklistDone:    done,
			DaysSinceTouched: daysSinceTouched,
		})
	}

	output, err := json.MarshalIndent(map[string]interface{}{
		"projects": overview,
		"total":    len(overview),
	}, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleSuggest(args []string) {
	fs := flag.NewFlagSet("suggest", flag.ExitOnError)
	context := fs.String("context", "", "Filter by context")
	timeAvailable := fs.String("time", "", "Time available (informational only)")

	fs.Parse(args)

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	projects, err := database.ListProjects(*context, "active")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list projects: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	// Build suggestions with scoring logic
	type Suggestion struct {
		ProjectID        int64  `json:"project_id"`
		Name             string `json:"name"`
		Context          string `json:"context"`
		Goal             string `json:"goal,omitempty"`
		ProgressPercent  int    `json:"progress_percent"`
		DaysSinceTouched int    `json:"days_since_touched"`
		Score            int    `json:"score"`
		Reason           string `json:"reason"`
	}

	suggestions := []Suggestion{}
	for _, proj := range projects {
		total, done, _ := database.GetChecklistStats(proj.ID)
		progress := 0
		if total > 0 {
			progress = (done * 100) / total
		}

		status, _ := database.GetProjectStatus(proj.ID)
		daysSince := status.DaysSinceTouched

		// Scoring logic:
		// - High completion momentum: +50 points if >70% complete
		// - Staleness: +1 point per day since last touched
		// - Active status: required (already filtered)
		score := 0
		reason := ""

		if progress >= 70 {
			score += 50
			reason = "High completion - close to done"
		} else if daysSince > 7 {
			score += daysSince
			reason = fmt.Sprintf("Neglected for %d days", daysSince)
		} else {
			score += daysSince / 2
			reason = "Recent activity"
		}

		suggestions = append(suggestions, Suggestion{
			ProjectID:        proj.ID,
			Name:             proj.Name,
			Context:          proj.ContextName,
			Goal:             proj.Goal,
			ProgressPercent:  progress,
			DaysSinceTouched: daysSince,
			Score:            score,
			Reason:           reason,
		})
	}

	// Sort by score (descending)
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Score > suggestions[i].Score {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	output, err := json.MarshalIndent(map[string]interface{}{
		"time_available": *timeAvailable,
		"suggestions":    suggestions,
	}, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleNeglected(args []string) {
	fs := flag.NewFlagSet("neglected", flag.ExitOnError)
	days := fs.Int("days", 14, "Number of days (default: 14)")

	fs.Parse(args)

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	projects, err := database.GetNeglectedProjects(*days)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get neglected projects: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	type NeglectedProject struct {
		ID               int64  `json:"id"`
		Name             string `json:"name"`
		Context          string `json:"context"`
		DaysSinceTouched int    `json:"days_since_touched"`
	}

	neglected := []NeglectedProject{}
	for _, proj := range projects {
		status, _ := database.GetProjectStatus(proj.ID)
		neglected = append(neglected, NeglectedProject{
			ID:               proj.ID,
			Name:             proj.Name,
			Context:          proj.ContextName,
			DaysSinceTouched: status.DaysSinceTouched,
		})
	}

	output, err := json.MarshalIndent(map[string]interface{}{
		"threshold_days": *days,
		"projects":       neglected,
		"total":          len(neglected),
	}, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleClosing(args []string) {
	fs := flag.NewFlagSet("closing", flag.ExitOnError)
	threshold := fs.Int("threshold", 70, "Completion threshold percentage (default: 70)")

	fs.Parse(args)

	if *threshold < 0 || *threshold > 100 {
		fmt.Fprintf(os.Stderr, "Error: threshold must be between 0 and 100\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	projects, err := database.GetProjectsNearCompletion(*threshold)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get projects near completion: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	type ClosingProject struct {
		ID              int64  `json:"id"`
		Name            string `json:"name"`
		Context         string `json:"context"`
		ProgressPercent int    `json:"progress_percent"`
		ChecklistTotal  int    `json:"checklist_total"`
		ChecklistDone   int    `json:"checklist_done"`
		Remaining       int    `json:"remaining"`
	}

	closing := []ClosingProject{}
	for _, status := range projects {
		closing = append(closing, ClosingProject{
			ID:              status.Project.ID,
			Name:            status.Project.Name,
			Context:         status.Project.ContextName,
			ProgressPercent: status.ProgressPercent,
			ChecklistTotal:  status.ChecklistTotal,
			ChecklistDone:   status.ChecklistDone,
			Remaining:       status.ChecklistTotal - status.ChecklistDone,
		})
	}

	output, err := json.MarshalIndent(map[string]interface{}{
		"threshold": *threshold,
		"projects":  closing,
		"total":     len(closing),
	}, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleContext(args []string) {
	fs := flag.NewFlagSet("context", flag.ExitOnError)
	name := fs.String("name", "", "Context name (required)")

	fs.Parse(args)

	if *name == "" {
		fmt.Fprintf(os.Stderr, "Error: --name is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	projects, err := database.GetContextProjects(*name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get context projects: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	output, err := json.MarshalIndent(map[string]interface{}{
		"context":  *name,
		"projects": projects,
		"total":    len(projects),
	}, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}
