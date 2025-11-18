package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// Transaction represents a financial transaction
type Transaction struct {
	ID              int64
	AccountName     string
	AccountLast4    string
	TransactionDate time.Time
	PostDate        *time.Time
	Description     string
	Amount          float64
	TransactionType string // "debit" or "credit"
	Balance         *float64
	StatementDate   time.Time
	SourceFile      string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ProcessingLog represents a statement processing record
type ProcessingLog struct {
	ID                   int64
	SourceFile           string
	StatementDate        *time.Time
	AccountName          string
	TransactionsInserted int
	TransactionsSkipped  int
	Status               string
	ErrorMessage         string
	ProcessedAt          time.Time
}

// New creates a new database connection and initializes schema
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Enable foreign keys
	_, err = conn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("initialize schema: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// initSchema creates the necessary tables if they don't exist
func (db *DB) initSchema() error {
	schema := `
	-- Transactions table
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,

		-- Account information
		account_name TEXT NOT NULL,
		account_last4 TEXT NOT NULL,

		-- Transaction details
		transaction_date DATE NOT NULL,
		post_date DATE,
		description TEXT NOT NULL,
		amount REAL NOT NULL,
		transaction_type TEXT NOT NULL CHECK (transaction_type IN ('debit', 'credit')),
		balance REAL,

		-- Statement metadata
		statement_date DATE NOT NULL,
		source_file TEXT,

		-- Timestamps
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

		-- Prevent duplicate transactions
		UNIQUE(account_last4, transaction_date, description, amount)
	);

	-- Indexes for common queries
	CREATE INDEX IF NOT EXISTS idx_transactions_account
		ON transactions(account_name, account_last4);

	CREATE INDEX IF NOT EXISTS idx_transactions_date
		ON transactions(transaction_date);

	CREATE INDEX IF NOT EXISTS idx_transactions_post_date
		ON transactions(post_date);

	CREATE INDEX IF NOT EXISTS idx_transactions_statement_date
		ON transactions(statement_date);

	CREATE INDEX IF NOT EXISTS idx_transactions_type
		ON transactions(transaction_type);

	CREATE INDEX IF NOT EXISTS idx_transactions_account_date
		ON transactions(account_name, transaction_date DESC);

	-- Processing log table
	CREATE TABLE IF NOT EXISTS processing_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_file TEXT NOT NULL,
		statement_date DATE,
		account_name TEXT,
		transactions_inserted INTEGER DEFAULT 0,
		transactions_skipped INTEGER DEFAULT 0,
		status TEXT NOT NULL CHECK (status IN ('success', 'parse_error', 'db_error')),
		error_message TEXT,
		processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_processing_log_file
		ON processing_log(source_file);

	CREATE INDEX IF NOT EXISTS idx_processing_log_status
		ON processing_log(status);

	-- Trigger to update updated_at
	CREATE TRIGGER IF NOT EXISTS update_transactions_updated_at
		AFTER UPDATE ON transactions
		FOR EACH ROW
	BEGIN
		UPDATE transactions SET updated_at = CURRENT_TIMESTAMP
		WHERE id = NEW.id;
	END;
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	return nil
}

// InsertTransaction inserts a transaction into the database
// Returns true if inserted, false if skipped (duplicate)
func (db *DB) InsertTransaction(tx *Transaction) (bool, error) {
	query := `
		INSERT INTO transactions (
			account_name, account_last4, transaction_date, post_date,
			description, amount, transaction_type, balance,
			statement_date, source_file
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(
		query,
		tx.AccountName,
		tx.AccountLast4,
		tx.TransactionDate,
		tx.PostDate,
		tx.Description,
		tx.Amount,
		tx.TransactionType,
		tx.Balance,
		tx.StatementDate,
		tx.SourceFile,
	)

	if err != nil {
		// Check if it's a unique constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return false, nil
		}
		return false, fmt.Errorf("insert transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

// InsertTransactions inserts multiple transactions in a single transaction
func (db *DB) InsertTransactions(transactions []*Transaction) (inserted int, skipped int, err error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return 0, 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO transactions (
			account_name, account_last4, transaction_date, post_date,
			description, amount, transaction_type, balance,
			statement_date, source_file
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, t := range transactions {
		result, err := stmt.Exec(
			t.AccountName,
			t.AccountLast4,
			t.TransactionDate,
			t.PostDate,
			t.Description,
			t.Amount,
			t.TransactionType,
			t.Balance,
			t.StatementDate,
			t.SourceFile,
		)

		if err != nil {
			// Check if it's a unique constraint violation
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				skipped++
				continue
			}
			return inserted, skipped, fmt.Errorf("insert transaction: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return inserted, skipped, fmt.Errorf("get rows affected: %w", err)
		}

		if rowsAffected > 0 {
			inserted++
		} else {
			skipped++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("commit transaction: %w", err)
	}

	return inserted, skipped, nil
}

// LogProcessing logs the processing of a statement
func (db *DB) LogProcessing(log *ProcessingLog) error {
	query := `
		INSERT INTO processing_log (
			source_file, statement_date, account_name,
			transactions_inserted, transactions_skipped,
			status, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(
		query,
		log.SourceFile,
		log.StatementDate,
		log.AccountName,
		log.TransactionsInserted,
		log.TransactionsSkipped,
		log.Status,
		log.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("log processing: %w", err)
	}

	return nil
}

// QueryTransactionsWithType queries transactions with optional transaction type filter
func (db *DB) QueryTransactionsWithType(startDate, endDate time.Time, accountFilter, transactionType string) ([]*Transaction, error) {
	query := `
		SELECT
			id, account_name, account_last4, transaction_date, post_date,
			description, amount, transaction_type, balance,
			statement_date, source_file, created_at, updated_at
		FROM transactions
		WHERE transaction_date >= ? AND transaction_date <= ?
	`

	args := []interface{}{startDate, endDate}

	if accountFilter != "" {
		query += " AND (account_name LIKE ? OR account_last4 = ?)"
		likeFilter := "%" + accountFilter + "%"
		args = append(args, likeFilter, accountFilter)
	}

	if transactionType != "all" {
		query += " AND transaction_type = ?"
		args = append(args, transactionType)
	}

	query += " ORDER BY transaction_date DESC, id DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		tx := &Transaction{}
		err := rows.Scan(
			&tx.ID,
			&tx.AccountName,
			&tx.AccountLast4,
			&tx.TransactionDate,
			&tx.PostDate,
			&tx.Description,
			&tx.Amount,
			&tx.TransactionType,
			&tx.Balance,
			&tx.StatementDate,
			&tx.SourceFile,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}

	return transactions, nil
}

// GetAccountSummary retrieves account summary information
func (db *DB) GetAccountSummary() ([]map[string]interface{}, error) {
	query := `
		SELECT
			account_name,
			account_last4,
			COUNT(*) as transaction_count,
			SUM(CASE WHEN transaction_type = 'debit' THEN amount ELSE 0 END) as total_debits,
			SUM(CASE WHEN transaction_type = 'credit' THEN amount ELSE 0 END) as total_credits,
			MIN(transaction_date) as first_transaction,
			MAX(transaction_date) as last_transaction,
			MAX(statement_date) as latest_statement
		FROM transactions
		GROUP BY account_name, account_last4
		ORDER BY account_name
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query account summary: %w", err)
	}
	defer rows.Close()

	var summaries []map[string]interface{}
	for rows.Next() {
		var accountName, accountLast4 string
		var transactionCount int
		var totalDebits, totalCredits float64
		var firstTx, lastTx, latestStmt time.Time

		err := rows.Scan(
			&accountName,
			&accountLast4,
			&transactionCount,
			&totalDebits,
			&totalCredits,
			&firstTx,
			&lastTx,
			&latestStmt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan summary: %w", err)
		}

		summaries = append(summaries, map[string]interface{}{
			"account_name":      accountName,
			"account_last4":     accountLast4,
			"transaction_count": transactionCount,
			"total_debits":      totalDebits,
			"total_credits":     totalCredits,
			"first_transaction": firstTx,
			"last_transaction":  lastTx,
			"latest_statement":  latestStmt,
		})
	}

	return summaries, nil
}
