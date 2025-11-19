package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"financial-liability-tracker/db"
	"financial-liability-tracker/pkg/app"
	"financial-liability-tracker/pkg/exitcodes"
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
	case "delete":
		handleDelete(args)
	case "list":
		handleList(args)
	case "get":
		handleGet(args)
	case "total":
		handleTotal(args)
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
	usage := `Financial Liability Tracker

Usage: financial-liability-tracker <command> [options]

Commands:
  add      Add a new liability
  update   Update liability balance
  delete   Delete a liability
  list     List all liabilities
  get      Get liability details
  total    Calculate total of all balances
  help     Show this help message

Examples:
  # Add a credit card
  financial-liability-tracker add --type credit-card --name "Chase Sapphire" \\
    --balance 2500 --limit 10000 --rate 18.99 --min-payment 50

  # Add an auto loan
  financial-liability-tracker add --type auto-loan --name "Honda Civic Loan" \\
    --balance 15000 --original 25000 --rate 4.5 --min-payment 350

  # Update balance
  financial-liability-tracker update chase-sapphire --balance 2100

  # List all liabilities
  financial-liability-tracker list

  # List by type
  financial-liability-tracker list --type credit-card

  # Get liability with history
  financial-liability-tracker get chase-sapphire --history

  # Get total balance
  financial-liability-tracker total

Environment Variables:
  POSTGRES_HOST      PostgreSQL host (default: localhost)
  POSTGRES_PORT      PostgreSQL port (default: 5432)
  POSTGRES_USER      PostgreSQL user (default: postgres)
  POSTGRES_PASSWORD  PostgreSQL password (required)
  POSTGRES_DB        PostgreSQL database (default: financial_tracker)
  POSTGRES_SSLMODE   SSL mode (default: disable)

Exit Codes:
  0 - Success
  1 - Invalid arguments
  2 - Database error
  3 - Not found
`
	fmt.Fprint(os.Stderr, usage)
}

func handleAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	name := fs.String("name", "", "Liability name (required, unique identifier)")
	liabilityType := fs.String("type", "", "Liability type (required: credit-card, auto-loan, mortgage, student-loan, personal-loan, medical-debt)")
	balance := fs.Float64("balance", 0, "Current balance (required)")
	original := fs.Float64("original", 0, "Original amount (optional)")
	limit := fs.Float64("limit", 0, "Credit limit for credit cards (optional)")
	rate := fs.Float64("rate", 0, "Interest rate percentage (optional)")
	minPayment := fs.Float64("min-payment", 0, "Minimum payment (optional)")
	creditor := fs.String("creditor", "", "Creditor name (optional)")
	last4 := fs.String("last4", "", "Last 4 digits of account (optional)")
	opened := fs.String("opened", "", "Opened date YYYY-MM-DD (optional)")
	notes := fs.String("notes", "", "Additional notes (optional)")

	fs.Parse(args)

	if *name == "" || *liabilityType == "" || *balance == 0 {
		fmt.Fprintf(os.Stderr, `{"error": "name, type, and balance are required"}`+"\n")
		os.Exit(exitcodes.ArgsError)
	}

	// Validate liability type
	validTypes := map[string]bool{
		"credit-card":   true,
		"auto-loan":     true,
		"mortgage":      true,
		"student-loan":  true,
		"personal-loan": true,
		"medical-debt":  true,
	}
	if !validTypes[*liabilityType] {
		fmt.Fprintf(os.Stderr, `{"error": "invalid liability type: %s"}`+"\n", *liabilityType)
		os.Exit(exitcodes.ArgsError)
	}

	liability := &db.Liability{
		Name:           *name,
		LiabilityType:  *liabilityType,
		CurrentBalance: *balance,
		CreditorName:   *creditor,
		AccountLast4:   *last4,
		Notes:          *notes,
	}

	if *original > 0 {
		liability.OriginalAmount = original
	}
	if *limit > 0 {
		liability.CreditLimit = limit
	}
	if *rate > 0 {
		liability.InterestRate = rate
	}
	if *minPayment > 0 {
		liability.MinimumPayment = minPayment
	}
	if *opened != "" {
		openedDate, err := time.Parse("2006-01-02", *opened)
		if err != nil {
			fmt.Fprintf(os.Stderr, `{"error": "invalid date format (use YYYY-MM-DD): %v"}`+"\n", err)
			os.Exit(exitcodes.ArgsError)
		}
		liability.OpenedDate = &openedDate
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error": "database initialization failed: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	if err := database.AddLiability(liability); err != nil {
		fmt.Fprintf(os.Stderr, `{"error": "failed to add liability: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}

	output, _ := json.Marshal(map[string]interface{}{
		"success":    true,
		"liability":  liability,
	})
	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleUpdate(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, `{"error": "liability name is required"}`+"\n")
		os.Exit(exitcodes.ArgsError)
	}

	name := args[0]
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	balance := fs.Float64("balance", 0, "New balance")
	notes := fs.String("notes", "", "Update notes")

	fs.Parse(args[1:])

	if *balance == 0 {
		fmt.Fprintf(os.Stderr, `{"error": "balance is required"}`+"\n")
		os.Exit(exitcodes.ArgsError)
	}

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error": "database initialization failed: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	if err := database.UpdateLiability(name, balance, *notes); err != nil {
		if err.Error() == fmt.Sprintf("liability not found: %s", name) {
			fmt.Fprintf(os.Stderr, `{"error": "liability not found: %s"}`+"\n", name)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, `{"error": "failed to update liability: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}

	output, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Liability '%s' updated successfully", name),
	})
	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleDelete(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, `{"error": "liability name is required"}`+"\n")
		os.Exit(exitcodes.ArgsError)
	}

	name := args[0]

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error": "database initialization failed: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	if err := database.DeleteLiability(name); err != nil {
		if err.Error() == fmt.Sprintf("liability not found: %s", name) {
			fmt.Fprintf(os.Stderr, `{"error": "liability not found: %s"}`+"\n", name)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, `{"error": "failed to delete liability: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}

	output, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Liability '%s' deleted successfully", name),
	})
	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	liabilityType := fs.String("type", "", "Filter by liability type")
	fs.Parse(args)

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error": "database initialization failed: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	liabilities, err := database.ListLiabilities(*liabilityType)
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error": "failed to list liabilities: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}

	output, _ := json.Marshal(map[string]interface{}{
		"liabilities": liabilities,
		"count":       len(liabilities),
	})
	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleGet(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, `{"error": "liability name is required"}`+"\n")
		os.Exit(exitcodes.ArgsError)
	}

	name := args[0]
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	showHistory := fs.Bool("history", false, "Include balance history")
	fs.Parse(args[1:])

	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error": "database initialization failed: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	liability, err := database.GetLiability(name)
	if err != nil {
		if err.Error() == fmt.Sprintf("liability not found: %s", name) {
			fmt.Fprintf(os.Stderr, `{"error": "liability not found: %s"}`+"\n", name)
			os.Exit(exitcodes.NotFound)
		}
		fmt.Fprintf(os.Stderr, `{"error": "failed to get liability: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}

	result := map[string]interface{}{
		"liability": liability,
	}

	if *showHistory {
		history, err := database.GetBalanceHistory(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, `{"error": "failed to get balance history: %v"}`+"\n", err)
			os.Exit(exitcodes.DBError)
		}
		result["balance_history"] = history
	}

	output, _ := json.Marshal(result)
	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}

func handleTotal(args []string) {
	database, err := app.InitDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error": "database initialization failed: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}
	defer database.Close()

	total, err := database.GetTotalBalance()
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error": "failed to calculate total: %v"}`+"\n", err)
		os.Exit(exitcodes.DBError)
	}

	output, _ := json.Marshal(map[string]interface{}{
		"total_balance": total,
	})
	fmt.Println(string(output))
	os.Exit(exitcodes.Success)
}
