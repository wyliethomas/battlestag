package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// ProcessedFile represents a file that has been processed
type ProcessedFile struct {
	ID          int64
	WatchID     string
	FilePath    string
	ProcessedAt time.Time
}

// New creates a new database connection and initializes the schema
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
	CREATE TABLE IF NOT EXISTS processed_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		watch_id TEXT NOT NULL,
		file_path TEXT NOT NULL,
		processed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(watch_id, file_path)
	);

	CREATE INDEX IF NOT EXISTS idx_watch_id ON processed_files(watch_id);
	CREATE INDEX IF NOT EXISTS idx_file_path ON processed_files(file_path);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	return nil
}

// IsFileProcessed checks if a file has already been processed for a given watch
func (db *DB) IsFileProcessed(watchID, filePath string) (bool, error) {
	query := `SELECT COUNT(*) FROM processed_files WHERE watch_id = ? AND file_path = ?`

	var count int
	err := db.conn.QueryRow(query, watchID, filePath).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query processed file: %w", err)
	}

	return count > 0, nil
}

// RecordProcessedFile records that a file has been processed
func (db *DB) RecordProcessedFile(watchID, filePath string) error {
	query := `INSERT INTO processed_files (watch_id, file_path) VALUES (?, ?)`

	_, err := db.conn.Exec(query, watchID, filePath)
	if err != nil {
		return fmt.Errorf("insert processed file: %w", err)
	}

	return nil
}

// GetLastProcessedFile returns the most recently processed file for a watch
func (db *DB) GetLastProcessedFile(watchID string) (*ProcessedFile, error) {
	query := `
		SELECT id, watch_id, file_path, processed_at
		FROM processed_files
		WHERE watch_id = ?
		ORDER BY processed_at DESC
		LIMIT 1
	`

	var pf ProcessedFile
	err := db.conn.QueryRow(query, watchID).Scan(
		&pf.ID,
		&pf.WatchID,
		&pf.FilePath,
		&pf.ProcessedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("query last processed file: %w", err)
	}

	return &pf, nil
}

// GetProcessedFiles returns all processed files for a watch
func (db *DB) GetProcessedFiles(watchID string) ([]ProcessedFile, error) {
	query := `
		SELECT id, watch_id, file_path, processed_at
		FROM processed_files
		WHERE watch_id = ?
		ORDER BY processed_at DESC
	`

	rows, err := db.conn.Query(query, watchID)
	if err != nil {
		return nil, fmt.Errorf("query processed files: %w", err)
	}
	defer rows.Close()

	var files []ProcessedFile
	for rows.Next() {
		var pf ProcessedFile
		if err := rows.Scan(&pf.ID, &pf.WatchID, &pf.FilePath, &pf.ProcessedAt); err != nil {
			return nil, fmt.Errorf("scan processed file: %w", err)
		}
		files = append(files, pf)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate processed files: %w", err)
	}

	return files, nil
}

// DeleteProcessedFile removes a file from the processed records (for reprocessing)
func (db *DB) DeleteProcessedFile(watchID, filePath string) error {
	query := `DELETE FROM processed_files WHERE watch_id = ? AND file_path = ?`

	result, err := db.conn.Exec(query, watchID, filePath)
	if err != nil {
		return fmt.Errorf("delete processed file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found in processed records")
	}

	return nil
}
