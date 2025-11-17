package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"financial-statement-processor/config"
	"financial-statement-processor/db"

	"github.com/joho/godotenv"
)

// QueryResult represents the output structure
type QueryResult struct {
	StartDate    string                       `json:"start_date"`
	EndDate      string                       `json:"end_date"`
	TotalRecords int                          `json:"total_records"`
	Accounts     map[string][]TransactionJSON `json:"accounts"`
}

// TransactionJSON represents a transaction in JSON format
type TransactionJSON struct {
	ID              int64    `json:"id"`
	AccountName     string   `json:"account_name"`
	AccountLast4    string   `json:"account_last4"`
	TransactionDate string   `json:"transaction_date"`
	PostDate        *string  `json:"post_date,omitempty"`
	Description     string   `json:"description"`
	Amount          float64  `json:"amount"`
	TransactionType string   `json:"transaction_type"`
	Balance         *float64 `json:"balance,omitempty"`
	StatementDate   string   `json:"statement_date"`
	SourceFile      string   `json:"source_file"`
}

func main() {
	// Load .env file if it exists (silently ignore if it doesn't)
	_ = godotenv.Load()

	// Parse command-line arguments
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Query financial transactions from SQLite database.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  DB_PATH      SQLite database file path (default: ./transactions.db)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Query all transactions in October 2024\n")
		fmt.Fprintf(os.Stderr, "  %s --start-date 2024-10-01 --end-date 2024-10-31\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Query specific account\n")
		fmt.Fprintf(os.Stderr, "  %s --start-date 2024-10-01 --end-date 2024-10-31 --account 1234\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Query with pretty-printed JSON\n")
		fmt.Fprintf(os.Stderr, "  %s --start-date 2024-10-01 --end-date 2024-10-31 --pretty\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Show account summary\n")
		fmt.Fprintf(os.Stderr, "  %s --summary --pretty\n", os.Args[0])
	}

	startDateStr := flag.String("start-date", "", "Start date (YYYY-MM-DD) - required")
	endDateStr := flag.String("end-date", "", "End date (YYYY-MM-DD) - required")
	account := flag.String("account", "", "Filter by account name or last 4 digits (optional)")
	pretty := flag.Bool("pretty", false, "Pretty-print JSON output")
	summary := flag.Bool("summary", false, "Show account summary instead of transactions")
	flag.Parse()

	// Validate required flags
	if !*summary && (*startDateStr == "" || *endDateStr == "") {
		fmt.Fprintf(os.Stderr, "Error: --start-date and --end-date are required (unless using --summary)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Load database configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Connect to database
	database, err := db.New(cfg.DatabasePath())
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Handle summary mode
	if *summary {
		summaries, err := database.GetAccountSummary()
		if err != nil {
			log.Fatalf("Failed to retrieve account summary: %v", err)
		}

		output, err := formatOutput(summaries, *pretty)
		if err != nil {
			log.Fatalf("Failed to format output: %v", err)
		}

		fmt.Println(output)
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", *startDateStr)
	if err != nil {
		log.Fatalf("Invalid start date format (use YYYY-MM-DD): %v", err)
	}

	endDate, err := time.Parse("2006-01-02", *endDateStr)
	if err != nil {
		log.Fatalf("Invalid end date format (use YYYY-MM-DD): %v", err)
	}

	// Validate date range
	if endDate.Before(startDate) {
		log.Fatalf("End date must be after start date")
	}

	// Query transactions
	transactions, err := database.QueryTransactions(startDate, endDate, *account)
	if err != nil {
		log.Fatalf("Failed to query transactions: %v", err)
	}

	// Organize transactions by account
	accountMap := make(map[string][]TransactionJSON)
	for _, tx := range transactions {
		accountKey := fmt.Sprintf("%s (...%s)", tx.AccountName, tx.AccountLast4)

		// Convert transaction to JSON format
		txJSON := TransactionJSON{
			ID:              tx.ID,
			AccountName:     tx.AccountName,
			AccountLast4:    tx.AccountLast4,
			TransactionDate: tx.TransactionDate.Format("2006-01-02"),
			Description:     tx.Description,
			Amount:          tx.Amount,
			TransactionType: tx.TransactionType,
			StatementDate:   tx.StatementDate.Format("2006-01-02"),
			SourceFile:      tx.SourceFile,
		}

		if tx.PostDate != nil {
			postDateStr := tx.PostDate.Format("2006-01-02")
			txJSON.PostDate = &postDateStr
		}

		if tx.Balance != nil {
			txJSON.Balance = tx.Balance
		}

		accountMap[accountKey] = append(accountMap[accountKey], txJSON)
	}

	// Build result
	result := QueryResult{
		StartDate:    startDate.Format("2006-01-02"),
		EndDate:      endDate.Format("2006-01-02"),
		TotalRecords: len(transactions),
		Accounts:     accountMap,
	}

	// Format and output
	output, err := formatOutput(result, *pretty)
	if err != nil {
		log.Fatalf("Failed to format output: %v", err)
	}

	fmt.Println(output)
}

// formatOutput formats data as JSON
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
