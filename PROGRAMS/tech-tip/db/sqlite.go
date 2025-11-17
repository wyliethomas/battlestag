package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Tip struct {
	ID        int
	Date      string
	Tip       string
	Category  string
	CreatedAt time.Time
}

type DB struct {
	conn *sql.DB
}

// New creates a new database connection and initializes the schema
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// initSchema creates the tips table if it doesn't exist
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tips (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT UNIQUE NOT NULL,
		tip TEXT NOT NULL,
		category TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// GetTipByDate retrieves a tip for a specific date (YYYY-MM-DD format)
func (db *DB) GetTipByDate(date string) (*Tip, error) {
	query := `SELECT id, date, tip, category, created_at FROM tips WHERE date = ?`

	var t Tip
	var createdAtStr string
	var category sql.NullString

	err := db.conn.QueryRow(query, date).Scan(&t.ID, &t.Date, &t.Tip, &category, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, nil // No tip found for this date
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query tip: %w", err)
	}

	// Handle nullable category
	if category.Valid {
		t.Category = category.String
	}

	// Parse the created_at timestamp
	t.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		// Fallback to current time if parsing fails
		t.CreatedAt = time.Now()
	}

	return &t, nil
}

// GetRandomTip retrieves a random tip from the database
func (db *DB) GetRandomTip() (*Tip, error) {
	query := `SELECT id, date, tip, category, created_at FROM tips ORDER BY RANDOM() LIMIT 1`

	var t Tip
	var createdAtStr string
	var category sql.NullString

	err := db.conn.QueryRow(query).Scan(&t.ID, &t.Date, &t.Tip, &category, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, nil // No tips in database
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query random tip: %w", err)
	}

	// Handle nullable category
	if category.Valid {
		t.Category = category.String
	}

	// Parse the created_at timestamp
	t.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		t.CreatedAt = time.Now()
	}

	return &t, nil
}

// SaveTip saves a new tip for a specific date
func (db *DB) SaveTip(date, tip, category string) error {
	query := `INSERT INTO tips (date, tip, category) VALUES (?, ?, ?)`

	_, err := db.conn.Exec(query, date, tip, category)
	if err != nil {
		return fmt.Errorf("failed to save tip: %w", err)
	}

	return nil
}

// DeleteTipByDate deletes a tip for a specific date (used for --regenerate)
func (db *DB) DeleteTipByDate(date string) error {
	query := `DELETE FROM tips WHERE date = ?`

	_, err := db.conn.Exec(query, date)
	if err != nil {
		return fmt.Errorf("failed to delete tip: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}
