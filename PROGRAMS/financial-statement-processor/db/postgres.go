package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps the PostgreSQL database connection
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
	ID                  int64
	SourceFile          string
	StatementDate       *time.Time
	AccountName         string
	TransactionsInserted int
	TransactionsSkipped int
	Status              string
	ErrorMessage        string
	ProcessedAt         time.Time
}

// New creates a new database connection
func New(connectionString string) (*DB, error) {
	conn, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// InsertTransaction inserts a transaction into the database
// Returns true if inserted, false if skipped (duplicate)
func (db *DB) InsertTransaction(tx *Transaction) (bool, error) {
	query := `
		INSERT INTO transactions (
			account_name, account_last4, transaction_date, post_date,
			description, amount, transaction_type, balance,
			statement_date, source_file
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (account_last4, transaction_date, description, amount) DO NOTHING
		RETURNING id
	`

	var id int64
	err := db.conn.QueryRow(
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
	).Scan(&id)

	if err == sql.ErrNoRows {
		// Duplicate, skipped
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("insert transaction: %w", err)
	}

	return true, nil
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
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (account_last4, transaction_date, description, amount) DO NOTHING
		RETURNING id
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, t := range transactions {
		var id int64
		err := stmt.QueryRow(
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
		).Scan(&id)

		if err == sql.ErrNoRows {
			skipped++
		} else if err != nil {
			return inserted, skipped, fmt.Errorf("insert transaction: %w", err)
		} else {
			inserted++
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
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
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

// QueryTransactions retrieves transactions within a date range
func (db *DB) QueryTransactions(startDate, endDate time.Time, accountFilter string) ([]*Transaction, error) {
	query := `
		SELECT
			id, account_name, account_last4, transaction_date, post_date,
			description, amount, transaction_type, balance,
			statement_date, source_file, created_at, updated_at
		FROM transactions
		WHERE transaction_date >= $1 AND transaction_date <= $2
	`

	args := []interface{}{startDate, endDate}

	if accountFilter != "" {
		query += " AND (account_name ILIKE $3 OR account_last4 = $3)"
		args = append(args, accountFilter)
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
	query := `SELECT * FROM account_summary`

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
			"account_name":       accountName,
			"account_last4":      accountLast4,
			"transaction_count":  transactionCount,
			"total_debits":       totalDebits,
			"total_credits":      totalCredits,
			"first_transaction":  firstTx,
			"last_transaction":   lastTx,
			"latest_statement":   latestStmt,
		})
	}

	return summaries, nil
}
