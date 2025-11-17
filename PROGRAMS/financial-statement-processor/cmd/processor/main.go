package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"financial-statement-processor/config"
	"financial-statement-processor/db"
	"financial-statement-processor/parser"

	"github.com/joho/godotenv"
)

const (
	ExitSuccess     = 0
	ExitParseError  = 1
	ExitDBError     = 2
	ExitConfigError = 3
)

func main() {
	// Load .env file if it exists (silently ignore if it doesn't)
	_ = godotenv.Load()

	// Parse command-line arguments
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <file_path>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Process a bank statement PDF or image file and store transactions in PostgreSQL.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  file_path    Path to the statement file (PDF, JPG, PNG, TIFF)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  DB_HOST      PostgreSQL host (default: localhost)\n")
		fmt.Fprintf(os.Stderr, "  DB_PORT      PostgreSQL port (default: 5432)\n")
		fmt.Fprintf(os.Stderr, "  DB_USER      PostgreSQL user (default: postgres)\n")
		fmt.Fprintf(os.Stderr, "  DB_PASSWORD  PostgreSQL password (required)\n")
		fmt.Fprintf(os.Stderr, "  DB_NAME      Database name (default: financial_data)\n")
		fmt.Fprintf(os.Stderr, "  DB_SSLMODE   SSL mode (default: disable)\n\n")
		fmt.Fprintf(os.Stderr, "Exit Codes:\n")
		fmt.Fprintf(os.Stderr, "  0 - Success\n")
		fmt.Fprintf(os.Stderr, "  1 - Parse error\n")
		fmt.Fprintf(os.Stderr, "  2 - Database error\n")
		fmt.Fprintf(os.Stderr, "  3 - Configuration error\n\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "  %s /path/to/statement.pdf\n", os.Args[0])
	}

	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: file path is required\n\n")
		flag.Usage()
		os.Exit(ExitConfigError)
	}

	filePath := flag.Arg(0)

	// Configure logging
	if !*verbose {
		log.SetFlags(0) // Remove timestamp for cleaner output
	}

	log.Printf("Financial Statement Processor")
	log.Printf("Processing file: %s", filePath)

	// Validate file exists
	if !parser.FileExists(filePath) {
		log.Printf("ERROR: File not found: %s", filePath)
		os.Exit(ExitParseError)
	}

	// Load database configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Printf("ERROR: Configuration error: %v", err)
		os.Exit(ExitConfigError)
	}

	// Connect to database
	log.Printf("Connecting to database...")
	database, err := db.New(cfg.ConnectionString())
	if err != nil {
		log.Printf("ERROR: Failed to connect to database: %v", err)
		os.Exit(ExitDBError)
	}
	defer database.Close()
	log.Printf("Database connected successfully")

	// Parse the statement file
	log.Printf("Parsing statement file...")
	statementData, err := parser.ParseFile(filePath)
	if err != nil {
		log.Printf("ERROR: Failed to parse file: %v", err)

		// Log parse failure
		logErr := database.LogProcessing(&db.ProcessingLog{
			SourceFile:   filepath.Base(filePath),
			Status:       "parse_error",
			ErrorMessage: err.Error(),
		})
		if logErr != nil {
			log.Printf("WARNING: Failed to log processing error: %v", logErr)
		}

		os.Exit(ExitParseError)
	}

	// Validate parsed data
	if err := parser.ValidateStatementData(statementData); err != nil {
		log.Printf("ERROR: Invalid statement data: %v", err)

		// Log validation failure
		logErr := database.LogProcessing(&db.ProcessingLog{
			SourceFile:   filepath.Base(filePath),
			Status:       "parse_error",
			ErrorMessage: fmt.Sprintf("Validation failed: %v", err),
		})
		if logErr != nil {
			log.Printf("WARNING: Failed to log processing error: %v", logErr)
		}

		os.Exit(ExitParseError)
	}

	log.Printf("Parsed statement successfully:")
	log.Printf("  Account: %s (...%s)", statementData.AccountName, statementData.AccountLast4)
	log.Printf("  Statement Date: %s", statementData.StatementDate.Format("2006-01-02"))
	log.Printf("  Transactions Found: %d", len(statementData.Transactions))

	// Insert transactions
	log.Printf("Inserting transactions into database...")
	inserted, skipped, err := database.InsertTransactions(statementData.Transactions)
	if err != nil {
		log.Printf("ERROR: Failed to insert transactions: %v", err)

		// Log database error
		logErr := database.LogProcessing(&db.ProcessingLog{
			SourceFile:          filepath.Base(filePath),
			StatementDate:       &statementData.StatementDate,
			AccountName:         statementData.AccountName,
			TransactionsInserted: inserted,
			TransactionsSkipped:  skipped,
			Status:              "db_error",
			ErrorMessage:        err.Error(),
		})
		if logErr != nil {
			log.Printf("WARNING: Failed to log processing error: %v", logErr)
		}

		os.Exit(ExitDBError)
	}

	log.Printf("Transactions inserted: %d", inserted)
	log.Printf("Transactions skipped (duplicates): %d", skipped)

	// Log successful processing
	err = database.LogProcessing(&db.ProcessingLog{
		SourceFile:           filepath.Base(filePath),
		StatementDate:        &statementData.StatementDate,
		AccountName:          statementData.AccountName,
		TransactionsInserted: inserted,
		TransactionsSkipped:  skipped,
		Status:               "success",
	})
	if err != nil {
		log.Printf("WARNING: Failed to log processing success: %v", err)
	}

	log.Printf("Processing completed successfully")
	os.Exit(ExitSuccess)
}
