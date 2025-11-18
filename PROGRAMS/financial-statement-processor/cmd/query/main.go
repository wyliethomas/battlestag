package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"financial-statement-processor/db"
	"financial-statement-processor/pkg/app"
	"financial-statement-processor/pkg/exitcodes"
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
	// Parse command-line arguments
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Query financial transactions from SQLite database.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  DB_PATH      SQLite database file path (default: ./transactions.db)\n\n")
		fmt.Fprintf(os.Stderr, "Exit Codes:\n")
		fmt.Fprintf(os.Stderr, "  0 - Success\n")
		fmt.Fprintf(os.Stderr, "  1 - Invalid arguments\n")
		fmt.Fprintf(os.Stderr, "  2 - Database error\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Query all transactions in October 2024\n")
		fmt.Fprintf(os.Stderr, "  %s --start-date 2024-10-01 --end-date 2024-10-31\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Query specific account\n")
		fmt.Fprintf(os.Stderr, "  %s --start-date 2024-10-01 --end-date 2024-10-31 --account 1234\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Query only debits (expenses) as CSV\n")
		fmt.Fprintf(os.Stderr, "  %s --start-date 2024-10-01 --end-date 2024-10-31 --type debit --csv\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Query with pretty-printed JSON\n")
		fmt.Fprintf(os.Stderr, "  %s --start-date 2024-10-01 --end-date 2024-10-31 --pretty\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Show account summary\n")
		fmt.Fprintf(os.Stderr, "  %s --summary --pretty\n", os.Args[0])
	}

	startDateStr := flag.String("start-date", "", "Start date (YYYY-MM-DD) - required")
	endDateStr := flag.String("end-date", "", "End date (YYYY-MM-DD) - required")
	account := flag.String("account", "", "Filter by account name or last 4 digits (optional)")
	transactionType := flag.String("type", "all", "Filter by transaction type: debit, credit, or all (default: all)")
	pretty := flag.Bool("pretty", false, "Pretty-print JSON output")
	csvOutput := flag.Bool("csv", false, "Output as CSV instead of JSON")
	summary := flag.Bool("summary", false, "Show account summary instead of transactions")
	flag.Parse()

	// Validate required flags
	if !*summary && (*startDateStr == "" || *endDateStr == "") {
		fmt.Fprintf(os.Stderr, "Error: --start-date and --end-date are required (unless using --summary)\n\n")
		flag.Usage()
		os.Exit(exitcodes.ArgsError)
	}

	// Validate transaction type
	if *transactionType != "all" && *transactionType != "debit" && *transactionType != "credit" {
		fmt.Fprintf(os.Stderr, "Error: --type must be 'debit', 'credit', or 'all' (got: %s)\n\n", *transactionType)
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

	// Handle summary mode
	if *summary {
		summaries, err := database.GetAccountSummary()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to retrieve account summary: %v\n", err)
			os.Exit(exitcodes.DBError)
		}

		output, err := formatOutput(summaries, *pretty)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
			os.Exit(exitcodes.DBError)
		}

		fmt.Println(output)
		os.Exit(exitcodes.Success)
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", *startDateStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid start date format (use YYYY-MM-DD): %v\n", err)
		os.Exit(exitcodes.ArgsError)
	}

	endDate, err := time.Parse("2006-01-02", *endDateStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid end date format (use YYYY-MM-DD): %v\n", err)
		os.Exit(exitcodes.ArgsError)
	}

	// Validate date range
	if endDate.Before(startDate) {
		fmt.Fprintf(os.Stderr, "End date must be after start date\n")
		os.Exit(exitcodes.ArgsError)
	}

	// Query transactions with type filter
	transactions, err := database.QueryTransactionsWithType(startDate, endDate, *account, *transactionType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to query transactions: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	// Handle CSV output
	if *csvOutput {
		if err := formatCSV(transactions); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format CSV output: %v\n", err)
			os.Exit(exitcodes.DBError)
		}
		os.Exit(exitcodes.Success)
	}

	// Organize transactions by account for JSON output
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

	// Format and output JSON
	output, err := formatOutput(result, *pretty)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
		os.Exit(exitcodes.DBError)
	}

	fmt.Println(output)
	os.Exit(exitcodes.Success)
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

// formatCSV writes transactions as CSV to stdout
func formatCSV(transactions []*db.Transaction) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Write header
	header := []string{
		"id",
		"account_name",
		"account_last4",
		"transaction_date",
		"post_date",
		"description",
		"amount",
		"transaction_type",
		"balance",
		"statement_date",
		"source_file",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, tx := range transactions {
		postDate := ""
		if tx.PostDate != nil {
			postDate = tx.PostDate.Format("2006-01-02")
		}

		balance := ""
		if tx.Balance != nil {
			balance = fmt.Sprintf("%.2f", *tx.Balance)
		}

		row := []string{
			fmt.Sprintf("%d", tx.ID),
			tx.AccountName,
			tx.AccountLast4,
			tx.TransactionDate.Format("2006-01-02"),
			postDate,
			tx.Description,
			fmt.Sprintf("%.2f", tx.Amount),
			tx.TransactionType,
			balance,
			tx.StatementDate.Format("2006-01-02"),
			tx.SourceFile,
		}

		if err := w.Write(row); err != nil {
			return err
		}
	}

	return nil
}
