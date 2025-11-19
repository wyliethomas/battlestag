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

// Asset represents a tracked asset
type Asset struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Category      string     `json:"category"`
	PurchasePrice *float64   `json:"purchase_price,omitempty"`
	PurchaseDate  *time.Time `json:"purchase_date,omitempty"`
	CurrentValue  float64    `json:"current_value"`
	DateAdded     time.Time  `json:"date_added"`
	LastUpdated   time.Time  `json:"last_updated"`
	IsRemoved     bool       `json:"is_removed"`
	RemovedDate   *time.Time `json:"removed_date,omitempty"`
	Notes         string     `json:"notes,omitempty"`
}

// ValueHistory represents a historical value record
type ValueHistory struct {
	ID           int64     `json:"id"`
	AssetID      int64     `json:"asset_id"`
	Value        float64   `json:"value"`
	RecordedDate time.Time `json:"recorded_date"`
	Notes        string    `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
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
	-- Assets table
	CREATE TABLE IF NOT EXISTS assets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		category TEXT NOT NULL,
		purchase_price REAL,
		purchase_date DATE,
		current_value REAL NOT NULL,
		date_added DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_removed BOOLEAN DEFAULT 0,
		removed_date DATE,
		notes TEXT
	);

	-- Value history table
	CREATE TABLE IF NOT EXISTS asset_value_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		asset_id INTEGER NOT NULL,
		value REAL NOT NULL,
		recorded_date DATE NOT NULL,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (asset_id) REFERENCES assets(id)
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_assets_category ON assets(category);
	CREATE INDEX IF NOT EXISTS idx_assets_is_removed ON assets(is_removed);
	CREATE INDEX IF NOT EXISTS idx_assets_last_updated ON assets(last_updated);
	CREATE INDEX IF NOT EXISTS idx_value_history_asset_id ON asset_value_history(asset_id);
	CREATE INDEX IF NOT EXISTS idx_value_history_recorded_date ON asset_value_history(recorded_date);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	return nil
}

// AddAsset inserts a new asset into the database
func (db *DB) AddAsset(asset *Asset) (int64, error) {
	now := time.Now()
	asset.DateAdded = now
	asset.LastUpdated = now

	query := `
		INSERT INTO assets (
			name, category, purchase_price, purchase_date,
			current_value, date_added, last_updated, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(
		query,
		asset.Name,
		asset.Category,
		asset.PurchasePrice,
		asset.PurchaseDate,
		asset.CurrentValue,
		asset.DateAdded,
		asset.LastUpdated,
		asset.Notes,
	)

	if err != nil {
		return 0, fmt.Errorf("insert asset: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	// Create initial value history entry
	history := &ValueHistory{
		AssetID:      id,
		Value:        asset.CurrentValue,
		RecordedDate: now,
		Notes:        "Initial value",
	}

	if err := db.AddValueHistory(history); err != nil {
		return id, fmt.Errorf("create initial history: %w", err)
	}

	return id, nil
}

// UpdateAssetValue updates an asset's current value and creates history entry
func (db *DB) UpdateAssetValue(id int64, value float64, notes string) error {
	// Check if asset exists and is not removed
	var isRemoved bool
	err := db.conn.QueryRow("SELECT is_removed FROM assets WHERE id = ?", id).Scan(&isRemoved)
	if err == sql.ErrNoRows {
		return fmt.Errorf("asset not found")
	}
	if err != nil {
		return fmt.Errorf("check asset: %w", err)
	}
	if isRemoved {
		return fmt.Errorf("cannot update removed asset")
	}

	now := time.Now()

	// Update asset
	query := `UPDATE assets SET current_value = ?, last_updated = ? WHERE id = ?`
	_, err = db.conn.Exec(query, value, now, id)
	if err != nil {
		return fmt.Errorf("update asset: %w", err)
	}

	// Create history entry
	history := &ValueHistory{
		AssetID:      id,
		Value:        value,
		RecordedDate: now,
		Notes:        notes,
	}

	if err := db.AddValueHistory(history); err != nil {
		return fmt.Errorf("create history entry: %w", err)
	}

	return nil
}

// RemoveAsset soft-deletes an asset
func (db *DB) RemoveAsset(id int64, removeDate *time.Time) error {
	if removeDate == nil {
		now := time.Now()
		removeDate = &now
	}

	query := `UPDATE assets SET is_removed = 1, removed_date = ? WHERE id = ?`
	result, err := db.conn.Exec(query, removeDate, id)
	if err != nil {
		return fmt.Errorf("remove asset: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("asset not found")
	}

	return nil
}

// RestoreAsset restores a soft-deleted asset
func (db *DB) RestoreAsset(id int64) error {
	query := `UPDATE assets SET is_removed = 0, removed_date = NULL WHERE id = ?`
	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("restore asset: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("asset not found")
	}

	return nil
}

// GetAsset retrieves an asset by ID
func (db *DB) GetAsset(id int64) (*Asset, error) {
	query := `
		SELECT id, name, category, purchase_price, purchase_date,
		       current_value, date_added, last_updated,
		       is_removed, removed_date, notes
		FROM assets WHERE id = ?
	`

	asset := &Asset{}
	err := db.conn.QueryRow(query, id).Scan(
		&asset.ID,
		&asset.Name,
		&asset.Category,
		&asset.PurchasePrice,
		&asset.PurchaseDate,
		&asset.CurrentValue,
		&asset.DateAdded,
		&asset.LastUpdated,
		&asset.IsRemoved,
		&asset.RemovedDate,
		&asset.Notes,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("asset not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query asset: %w", err)
	}

	return asset, nil
}

// ListAssets retrieves all assets with optional filters
func (db *DB) ListAssets(includeRemoved bool, category string) ([]*Asset, error) {
	query := `
		SELECT id, name, category, purchase_price, purchase_date,
		       current_value, date_added, last_updated,
		       is_removed, removed_date, notes
		FROM assets
		WHERE 1=1
	`

	args := []interface{}{}

	if !includeRemoved {
		query += " AND is_removed = 0"
	}

	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}

	query += " ORDER BY last_updated DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query assets: %w", err)
	}
	defer rows.Close()

	var assets []*Asset
	for rows.Next() {
		asset := &Asset{}
		err := rows.Scan(
			&asset.ID,
			&asset.Name,
			&asset.Category,
			&asset.PurchasePrice,
			&asset.PurchaseDate,
			&asset.CurrentValue,
			&asset.DateAdded,
			&asset.LastUpdated,
			&asset.IsRemoved,
			&asset.RemovedDate,
			&asset.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		assets = append(assets, asset)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assets: %w", err)
	}

	return assets, nil
}

// GetValueHistory retrieves value history for an asset
func (db *DB) GetValueHistory(assetID int64) ([]*ValueHistory, error) {
	query := `
		SELECT id, asset_id, value, recorded_date, notes, created_at
		FROM asset_value_history
		WHERE asset_id = ?
		ORDER BY recorded_date DESC
	`

	rows, err := db.conn.Query(query, assetID)
	if err != nil {
		return nil, fmt.Errorf("query value history: %w", err)
	}
	defer rows.Close()

	var history []*ValueHistory
	for rows.Next() {
		vh := &ValueHistory{}
		err := rows.Scan(
			&vh.ID,
			&vh.AssetID,
			&vh.Value,
			&vh.RecordedDate,
			&vh.Notes,
			&vh.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan value history: %w", err)
		}
		history = append(history, vh)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate value history: %w", err)
	}

	return history, nil
}

// AddValueHistory adds a value history entry
func (db *DB) AddValueHistory(history *ValueHistory) error {
	query := `
		INSERT INTO asset_value_history (asset_id, value, recorded_date, notes)
		VALUES (?, ?, ?, ?)
	`

	_, err := db.conn.Exec(
		query,
		history.AssetID,
		history.Value,
		history.RecordedDate,
		history.Notes,
	)

	if err != nil {
		return fmt.Errorf("insert value history: %w", err)
	}

	return nil
}

// GetSummary returns summary statistics
func (db *DB) GetSummary() (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Total active assets value
	var totalValue float64
	err := db.conn.QueryRow(`
		SELECT COALESCE(SUM(current_value), 0)
		FROM assets
		WHERE is_removed = 0
	`).Scan(&totalValue)
	if err != nil {
		return nil, fmt.Errorf("get total value: %w", err)
	}
	summary["total_value"] = totalValue

	// Count by category
	rows, err := db.conn.Query(`
		SELECT category, COUNT(*), SUM(current_value)
		FROM assets
		WHERE is_removed = 0
		GROUP BY category
		ORDER BY category
	`)
	if err != nil {
		return nil, fmt.Errorf("get category breakdown: %w", err)
	}
	defer rows.Close()

	categories := make([]map[string]interface{}, 0)
	for rows.Next() {
		var category string
		var count int
		var value float64
		if err := rows.Scan(&category, &count, &value); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, map[string]interface{}{
			"category": category,
			"count":    count,
			"value":    value,
		})
	}
	summary["categories"] = categories

	// Total count
	var totalCount int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM assets WHERE is_removed = 0").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("get total count: %w", err)
	}
	summary["total_count"] = totalCount

	return summary, nil
}
