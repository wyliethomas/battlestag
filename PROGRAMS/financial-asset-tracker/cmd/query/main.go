package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"financial-asset-tracker/db"
	"financial-asset-tracker/pkg/app"
	"financial-asset-tracker/pkg/exitcodes"
)

func main() {
	// Parse command-line arguments
	all := flag.Bool("all", false, "Include removed assets")
	category := flag.String("category", "", "Filter by category")
	history := flag.Bool("history", false, "Show value history for an asset")
	assetID := flag.Int64("id", 0, "Asset ID (required with --history)")
	summary := flag.Bool("summary", false, "Show summary statistics")
	staleDays := flag.Int("stale-days", 0, "Highlight assets not updated in X days (0 = no check)")
	csvOutput := flag.Bool("csv", false, "Output as CSV instead of JSON")
	pretty := flag.Bool("pretty", false, "Pretty-print JSON output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: financial-asset-tracker-query [OPTIONS]

Query and display asset information.

Options:
  --all              Include removed assets (default: active only)
  --category string  Filter by category (vehicle, property, investment, other)
  --history          Show value history for an asset (requires --id)
  --id int           Asset ID (required with --history)
  --summary          Show summary statistics
  --stale-days int   Highlight assets not updated in X days (default: 0 = no check)
  --csv              Output as CSV instead of JSON
  --pretty           Pretty-print JSON output

Environment Variables:
  DB_PATH    SQLite database file path (default: ~/.local/share/financial-asset-tracker/assets.db)

Exit Codes:
  0 - Success
  1 - Invalid arguments
  2 - Database error
  3 - Asset not found

Examples:
  # List all active assets
  financial-asset-tracker-query

  # List all assets including removed
  financial-asset-tracker-query --all

  # Filter by category
  financial-asset-tracker-query --category vehicle

  # Show assets with stale value warning (not updated in 30 days)
  financial-asset-tracker-query --stale-days 30

  # Show value history for asset ID 1
  financial-asset-tracker-query --history --id 1

  # Show summary statistics
  financial-asset-tracker-query --summary --pretty

  # Export to CSV
  financial-asset-tracker-query --csv
`)
	}

	flag.Parse()

	// Validate flags
	if *history && *assetID == 0 {
		fmt.Fprintf(os.Stderr, "Error: --id is required when using --history\n\n")
		flag.Usage()
		os.Exit(exitcodes.ArgsError)
	}

	// Initialize database
	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	// Handle different query modes
	if *summary {
		handleSummary(database, *pretty)
	} else if *history {
		handleHistory(database, *assetID, *pretty)
	} else {
		handleList(database, *all, *category, *staleDays, *csvOutput, *pretty)
	}
}

func handleList(database *db.DB, all bool, category string, staleDays int, csvOutput, pretty bool) {
	assets, err := database.ListAssets(all, category)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list assets: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	// Check for stale assets if requested
	staleAssets := make(map[int64]int)
	if staleDays > 0 {
		threshold := time.Now().AddDate(0, 0, -staleDays)
		for _, asset := range assets {
			if !asset.IsRemoved && asset.LastUpdated.Before(threshold) {
				daysSince := int(time.Since(asset.LastUpdated).Hours() / 24)
				staleAssets[asset.ID] = daysSince
			}
		}
	}

	if csvOutput {
		formatCSV(assets, staleAssets)
		os.Exit(exitcodes.Success)
	}

	// Build JSON output
	result := map[string]interface{}{
		"total_count": len(assets),
		"assets":      assets,
	}

	if len(staleAssets) > 0 {
		warnings := make([]string, 0)
		for id, days := range staleAssets {
			for _, asset := range assets {
				if asset.ID == id {
					warnings = append(warnings, fmt.Sprintf("Asset #%d '%s' not updated in %d days", id, asset.Name, days))
				}
			}
		}
		result["warnings"] = warnings
	}

	output, err := formatOutput(result, pretty)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(output)
	os.Exit(exitcodes.Success)
}

func handleHistory(database *db.DB, assetID int64, pretty bool) {
	// Get asset info
	asset, err := database.GetAsset(assetID)
	if err != nil {
		if err.Error() == "asset not found" {
			fmt.Fprintf(os.Stderr, "Error: Asset ID %d not found\n", assetID)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to get asset: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	// Get value history
	history, err := database.GetValueHistory(assetID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get value history: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	result := map[string]interface{}{
		"asset":   asset,
		"history": history,
	}

	output, err := formatOutput(result, pretty)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(output)
	os.Exit(exitcodes.Success)
}

func handleSummary(database *db.DB, pretty bool) {
	summary, err := database.GetSummary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get summary: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	output, err := formatOutput(summary, pretty)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(output)
	os.Exit(exitcodes.Success)
}

func formatOutput(data interface{}, pretty bool) (string, error) {
	var jsonData []byte
	var err error

	if pretty {
		jsonData, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonData, err = json.Marshal(data)
	}

	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func formatCSV(assets []*db.Asset, staleAssets map[int64]int) {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Write header
	header := []string{
		"id",
		"name",
		"category",
		"purchase_price",
		"purchase_date",
		"current_value",
		"date_added",
		"last_updated",
		"is_removed",
		"removed_date",
		"notes",
		"stale_warning",
	}
	w.Write(header)

	// Write data rows
	for _, asset := range assets {
		purchasePrice := ""
		if asset.PurchasePrice != nil {
			purchasePrice = fmt.Sprintf("%.2f", *asset.PurchasePrice)
		}

		purchaseDate := ""
		if asset.PurchaseDate != nil {
			purchaseDate = asset.PurchaseDate.Format("2006-01-02")
		}

		removedDate := ""
		if asset.RemovedDate != nil {
			removedDate = asset.RemovedDate.Format("2006-01-02")
		}

		staleWarning := ""
		if days, ok := staleAssets[asset.ID]; ok {
			staleWarning = fmt.Sprintf("Not updated in %d days", days)
		}

		row := []string{
			fmt.Sprintf("%d", asset.ID),
			asset.Name,
			asset.Category,
			purchasePrice,
			purchaseDate,
			fmt.Sprintf("%.2f", asset.CurrentValue),
			asset.DateAdded.Format("2006-01-02"),
			asset.LastUpdated.Format("2006-01-02"),
			fmt.Sprintf("%t", asset.IsRemoved),
			removedDate,
			asset.Notes,
			staleWarning,
		}

		w.Write(row)
	}
}
