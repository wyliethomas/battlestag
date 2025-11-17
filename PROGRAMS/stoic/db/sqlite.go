package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Thought struct {
	ID        int
	Date      string
	Thought   string
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

// initSchema creates the thoughts table if it doesn't exist
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS thoughts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT UNIQUE NOT NULL,
		thought TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// GetThoughtByDate retrieves a thought for a specific date (YYYY-MM-DD format)
func (db *DB) GetThoughtByDate(date string) (*Thought, error) {
	query := `SELECT id, date, thought, created_at FROM thoughts WHERE date = ?`

	var t Thought
	var createdAtStr string

	err := db.conn.QueryRow(query, date).Scan(&t.ID, &t.Date, &t.Thought, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, nil // No thought found for this date
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query thought: %w", err)
	}

	// Parse the created_at timestamp
	t.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		// Fallback to current time if parsing fails
		t.CreatedAt = time.Now()
	}

	return &t, nil
}

// SaveThought saves a new thought for a specific date
func (db *DB) SaveThought(date, thought string) error {
	query := `INSERT INTO thoughts (date, thought) VALUES (?, ?)`

	_, err := db.conn.Exec(query, date, thought)
	if err != nil {
		return fmt.Errorf("failed to save thought: %w", err)
	}

	return nil
}

// DeleteThoughtByDate deletes a thought for a specific date (used for --regenerate)
func (db *DB) DeleteThoughtByDate(date string) error {
	query := `DELETE FROM thoughts WHERE date = ?`

	_, err := db.conn.Exec(query, date)
	if err != nil {
		return fmt.Errorf("failed to delete thought: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}
