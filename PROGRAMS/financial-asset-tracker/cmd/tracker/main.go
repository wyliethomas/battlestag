package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"financial-asset-tracker/db"
	"financial-asset-tracker/pkg/app"
	"financial-asset-tracker/pkg/exitcodes"
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
	case "update":
		handleUpdate(args)
	case "remove":
		handleRemove(args)
	case "restore":
		handleRestore(args)
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
	fmt.Fprintf(os.Stderr, `Usage: financial-asset-tracker <command> [options]

Commands:
  add      Add a new asset
  update   Update an asset's value
  remove   Remove an asset (soft delete)
  restore  Restore a removed asset
  help     Show this help message

Examples:
  # Add a vehicle
  financial-asset-tracker add --name "2019 Honda Civic" --category vehicle --purchase-price 25000 --purchase-date 2019-06-15 --current-value 18000

  # Add property without purchase info
  financial-asset-tracker add --name "Main Residence" --category property --current-value 450000

  # Update asset value
  financial-asset-tracker update --id 1 --value 17500

  # Update with notes
  financial-asset-tracker update --id 1 --value 17500 --notes "Post-accident valuation"

  # Remove asset
  financial-asset-tracker remove --id 1

  # Restore asset
  financial-asset-tracker restore --id 1

Environment Variables:
  DB_PATH    SQLite database file path (default: ~/.local/share/financial-asset-tracker/assets.db)

Exit Codes:
  0 - Success
  1 - Invalid arguments
  2 - Database error
  3 - Asset not found
`)
}

func handleAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	name := fs.String("name", "", "Asset name (required)")
	category := fs.String("category", "other", "Asset category (vehicle, property, investment, other)")
	purchasePrice := fs.Float64("purchase-price", 0, "Purchase price (optional)")
	purchaseDate := fs.String("purchase-date", "", "Purchase date YYYY-MM-DD (optional)")
	currentValue := fs.Float64("current-value", 0, "Current value (required)")
	notes := fs.String("notes", "", "Additional notes")

	fs.Parse(args)

	if *name == "" {
		fmt.Fprintf(os.Stderr, "Error: --name is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	if *currentValue == 0 {
		fmt.Fprintf(os.Stderr, "Error: --current-value is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	asset := &db.Asset{
		Name:         *name,
		Category:     *category,
		CurrentValue: *currentValue,
		Notes:        *notes,
	}

	if *purchasePrice > 0 {
		asset.PurchasePrice = purchasePrice
	}

	if *purchaseDate != "" {
		pd, err := time.Parse("2006-01-02", *purchaseDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid purchase date format (use YYYY-MM-DD): %v\n", err)
			os.Exit(exitcodes.ArgsError)
		}
		asset.PurchaseDate = &pd
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	id, err := database.AddAsset(asset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add asset: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Asset added successfully (ID: %d)\n", id)
	os.Exit(exitcodes.Success)
}

func handleUpdate(args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	id := fs.Int64("id", 0, "Asset ID (required)")
	value := fs.Float64("value", 0, "New value (required)")
	notes := fs.String("notes", "", "Update notes")

	fs.Parse(args)

	if *id == 0 {
		fmt.Fprintf(os.Stderr, "Error: --id is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	if *value == 0 {
		fmt.Fprintf(os.Stderr, "Error: --value is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	err = database.UpdateAssetValue(*id, *value, *notes)
	if err != nil {
		if err.Error() == "asset not found" {
			fmt.Fprintf(os.Stderr, "Error: Asset ID %d not found\n", *id)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to update asset: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Asset %d updated successfully\n", *id)
	os.Exit(exitcodes.Success)
}

func handleRemove(args []string) {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	id := fs.Int64("id", 0, "Asset ID (required)")
	dateStr := fs.String("date", "", "Removal date YYYY-MM-DD (optional, defaults to today)")

	fs.Parse(args)

	if *id == 0 {
		fmt.Fprintf(os.Stderr, "Error: --id is required\n")
		os.Exit(exitcodes.ArgsError)
	}

	var removeDate *time.Time
	if *dateStr != "" {
		rd, err := time.Parse("2006-01-02", *dateStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid date format (use YYYY-MM-DD): %v\n", err)
			os.Exit(exitcodes.ArgsError)
		}
		removeDate = &rd
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	err = database.RemoveAsset(*id, removeDate)
	if err != nil {
		if err.Error() == "asset not found" {
			fmt.Fprintf(os.Stderr, "Error: Asset ID %d not found\n", *id)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to remove asset: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Asset %d removed successfully\n", *id)
	os.Exit(exitcodes.Success)
}

func handleRestore(args []string) {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)
	id := fs.Int64("id", 0, "Asset ID (required)")

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

	err = database.RestoreAsset(*id)
	if err != nil {
		if err.Error() == "asset not found" {
			fmt.Fprintf(os.Stderr, "Error: Asset ID %d not found\n", *id)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, "Failed to restore asset: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Printf("Asset %d restored successfully\n", *id)
	os.Exit(exitcodes.Success)
}
