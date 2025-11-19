package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var schema = `-- Financial Liability Tracker Schema
-- SQLite

-- Liabilities table
CREATE TABLE IF NOT EXISTS liabilities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    liability_type TEXT NOT NULL CHECK (liability_type IN (
        'credit-card', 'auto-loan', 'mortgage',
        'student-loan', 'personal-loan', 'medical-debt'
    )),
    current_balance REAL NOT NULL,
    original_amount REAL,
    credit_limit REAL,  -- For credit cards
    interest_rate REAL,
    minimum_payment REAL,
    creditor_name TEXT,
    account_last4 TEXT,
    opened_date TEXT,  -- ISO8601 format: YYYY-MM-DD
    notes TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Balance history table
CREATE TABLE IF NOT EXISTS liability_balance_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    liability_id INTEGER NOT NULL REFERENCES liabilities(id) ON DELETE CASCADE,
    balance REAL NOT NULL,
    recorded_at TEXT DEFAULT CURRENT_TIMESTAMP,
    notes TEXT
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_liabilities_type ON liabilities(liability_type);
CREATE INDEX IF NOT EXISTS idx_liabilities_name ON liabilities(name);
CREATE INDEX IF NOT EXISTS idx_balance_history_liability_id ON liability_balance_history(liability_id);
CREATE INDEX IF NOT EXISTS idx_balance_history_recorded_at ON liability_balance_history(recorded_at);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_liabilities_updated_at
AFTER UPDATE ON liabilities
FOR EACH ROW
BEGIN
    UPDATE liabilities SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
`

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// Liability represents a financial liability
type Liability struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	LiabilityType  string     `json:"liability_type"`
	CurrentBalance float64    `json:"current_balance"`
	OriginalAmount *float64   `json:"original_amount,omitempty"`
	CreditLimit    *float64   `json:"credit_limit,omitempty"`
	InterestRate   *float64   `json:"interest_rate,omitempty"`
	MinimumPayment *float64   `json:"minimum_payment,omitempty"`
	CreditorName   string     `json:"creditor_name,omitempty"`
	AccountLast4   string     `json:"account_last4,omitempty"`
	OpenedDate     *time.Time `json:"opened_date,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// BalanceHistory represents a historical balance record
type BalanceHistory struct {
	ID          int       `json:"id"`
	LiabilityID int       `json:"liability_id"`
	Balance     float64   `json:"balance"`
	RecordedAt  time.Time `json:"recorded_at"`
	Notes       string    `json:"notes,omitempty"`
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	// Initialize schema
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("initialize schema: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// AddLiability adds a new liability to the database
func (db *DB) AddLiability(l *Liability) error {
	query := `
		INSERT INTO liabilities (
			name, liability_type, current_balance, original_amount,
			credit_limit, interest_rate, minimum_payment, creditor_name,
			account_last4, opened_date, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(
		query,
		l.Name,
		l.LiabilityType,
		l.CurrentBalance,
		l.OriginalAmount,
		l.CreditLimit,
		l.InterestRate,
		l.MinimumPayment,
		l.CreditorName,
		l.AccountLast4,
		l.OpenedDate,
		l.Notes,
	)

	if err != nil {
		return fmt.Errorf("insert liability: %w", err)
	}

	// Get the inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	l.ID = int(id)

	// Get timestamps as strings and parse
	var createdAt, updatedAt string
	err = db.conn.QueryRow(
		"SELECT created_at, updated_at FROM liabilities WHERE id = ?",
		l.ID,
	).Scan(&createdAt, &updatedAt)
	if err != nil {
		return fmt.Errorf("get timestamps: %w", err)
	}

	// Parse timestamps
	l.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	l.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	// Add initial balance history entry
	historyQuery := `
		INSERT INTO liability_balance_history (liability_id, balance, notes)
		VALUES (?, ?, ?)
	`
	_, err = db.conn.Exec(historyQuery, l.ID, l.CurrentBalance, "Initial balance")
	if err != nil {
		return fmt.Errorf("insert balance history: %w", err)
	}

	return nil
}

// UpdateLiability updates an existing liability
func (db *DB) UpdateLiability(name string, newBalance *float64, notes string) error {
	// Get current liability
	var currentBalance float64
	var liabilityID int
	err := db.conn.QueryRow(
		"SELECT id, current_balance FROM liabilities WHERE name = ?",
		name,
	).Scan(&liabilityID, &currentBalance)

	if err == sql.ErrNoRows {
		return fmt.Errorf("liability not found: %s", name)
	}
	if err != nil {
		return fmt.Errorf("query liability: %w", err)
	}

	// Update balance if provided
	if newBalance != nil && *newBalance != currentBalance {
		_, err = db.conn.Exec(
			"UPDATE liabilities SET current_balance = ? WHERE name = ?",
			*newBalance,
			name,
		)
		if err != nil {
			return fmt.Errorf("update liability: %w", err)
		}

		// Add balance history entry
		historyQuery := `
			INSERT INTO liability_balance_history (liability_id, balance, notes)
			VALUES (?, ?, ?)
		`
		_, err = db.conn.Exec(historyQuery, liabilityID, *newBalance, notes)
		if err != nil {
			return fmt.Errorf("insert balance history: %w", err)
		}
	}

	return nil
}

// DeleteLiability deletes a liability by name
func (db *DB) DeleteLiability(name string) error {
	result, err := db.conn.Exec("DELETE FROM liabilities WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("delete liability: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("liability not found: %s", name)
	}

	return nil
}

// GetLiability retrieves a liability by name
func (db *DB) GetLiability(name string) (*Liability, error) {
	query := `
		SELECT id, name, liability_type, current_balance, original_amount,
		       credit_limit, interest_rate, minimum_payment, creditor_name,
		       account_last4, opened_date, notes, created_at, updated_at
		FROM liabilities
		WHERE name = ?
	`

	l := &Liability{}
	var createdAt, updatedAt string
	var openedDate sql.NullString

	err := db.conn.QueryRow(query, name).Scan(
		&l.ID,
		&l.Name,
		&l.LiabilityType,
		&l.CurrentBalance,
		&l.OriginalAmount,
		&l.CreditLimit,
		&l.InterestRate,
		&l.MinimumPayment,
		&l.CreditorName,
		&l.AccountLast4,
		&openedDate,
		&l.Notes,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("liability not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("query liability: %w", err)
	}

	// Parse timestamps
	l.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	l.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	// Parse opened date if present
	if openedDate.Valid {
		parsed, err := time.Parse("2006-01-02", openedDate.String)
		if err == nil {
			l.OpenedDate = &parsed
		}
	}

	return l, nil
}

// ListLiabilities retrieves all liabilities, optionally filtered by type
func (db *DB) ListLiabilities(liabilityType string) ([]*Liability, error) {
	query := `
		SELECT id, name, liability_type, current_balance, original_amount,
		       credit_limit, interest_rate, minimum_payment, creditor_name,
		       account_last4, opened_date, notes, created_at, updated_at
		FROM liabilities
	`

	args := []interface{}{}
	if liabilityType != "" {
		query += " WHERE liability_type = ?"
		args = append(args, liabilityType)
	}

	query += " ORDER BY created_at DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query liabilities: %w", err)
	}
	defer rows.Close()

	var liabilities []*Liability
	for rows.Next() {
		l := &Liability{}
		var createdAt, updatedAt string
		var openedDate sql.NullString

		err := rows.Scan(
			&l.ID,
			&l.Name,
			&l.LiabilityType,
			&l.CurrentBalance,
			&l.OriginalAmount,
			&l.CreditLimit,
			&l.InterestRate,
			&l.MinimumPayment,
			&l.CreditorName,
			&l.AccountLast4,
			&openedDate,
			&l.Notes,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan liability: %w", err)
		}

		// Parse timestamps
		l.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		l.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		// Parse opened date if present
		if openedDate.Valid {
			parsed, err := time.Parse("2006-01-02", openedDate.String)
			if err == nil {
				l.OpenedDate = &parsed
			}
		}

		liabilities = append(liabilities, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate liabilities: %w", err)
	}

	return liabilities, nil
}

// GetBalanceHistory retrieves balance history for a liability
func (db *DB) GetBalanceHistory(name string) ([]*BalanceHistory, error) {
	// First get the liability ID
	var liabilityID int
	err := db.conn.QueryRow("SELECT id FROM liabilities WHERE name = ?", name).Scan(&liabilityID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("liability not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("query liability: %w", err)
	}

	query := `
		SELECT id, liability_id, balance, recorded_at, notes
		FROM liability_balance_history
		WHERE liability_id = ?
		ORDER BY recorded_at DESC
	`

	rows, err := db.conn.Query(query, liabilityID)
	if err != nil {
		return nil, fmt.Errorf("query balance history: %w", err)
	}
	defer rows.Close()

	var history []*BalanceHistory
	for rows.Next() {
		h := &BalanceHistory{}
		var recordedAt string

		err := rows.Scan(
			&h.ID,
			&h.LiabilityID,
			&h.Balance,
			&recordedAt,
			&h.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("scan balance history: %w", err)
		}

		// Parse timestamp
		h.RecordedAt, _ = time.Parse("2006-01-02 15:04:05", recordedAt)

		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate balance history: %w", err)
	}

	return history, nil
}

// GetTotalBalance calculates the sum of all current balances
func (db *DB) GetTotalBalance() (float64, error) {
	var total float64
	err := db.conn.QueryRow("SELECT COALESCE(SUM(current_balance), 0) FROM liabilities").Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("calculate total balance: %w", err)
	}
	return total, nil
}
