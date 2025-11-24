package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// Server represents a server in the lab
type Server struct {
	ID         int64
	Name       string
	IPAddress  string
	SSHUser    string
	SSHPort    int
	Status     string
	LastSeen   *time.Time
	CreatedAt  time.Time
	Notes      string
}

// HealthCheck represents a health check record
type HealthCheck struct {
	ID             int64
	ServerID       int64
	CheckType      string
	Status         string
	ResponseTimeMS int
	ErrorMessage   string
	Timestamp      time.Time
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

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

// initSchema initializes the database schema
func (db *DB) initSchema() error {
	schema := `
	-- Servers table
	CREATE TABLE IF NOT EXISTS servers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		ip_address TEXT NOT NULL,
		ssh_user TEXT NOT NULL,
		ssh_port INTEGER DEFAULT 22,
		status TEXT DEFAULT 'unknown' CHECK(status IN ('online', 'offline', 'unknown')),
		last_seen DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		notes TEXT
	);

	-- Health check history
	CREATE TABLE IF NOT EXISTS health_checks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL,
		check_type TEXT NOT NULL CHECK(check_type IN ('ping', 'ssh')),
		status TEXT NOT NULL CHECK(status IN ('success', 'failure', 'timeout')),
		response_time_ms INTEGER,
		error_message TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
	);

	-- Indexes for performance
	CREATE INDEX IF NOT EXISTS idx_servers_status ON servers(status);
	CREATE INDEX IF NOT EXISTS idx_servers_last_seen ON servers(last_seen);
	CREATE INDEX IF NOT EXISTS idx_health_checks_server ON health_checks(server_id);
	CREATE INDEX IF NOT EXISTS idx_health_checks_timestamp ON health_checks(timestamp);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// ==================== Server CRUD Operations ====================

// AddServer adds a new server to the database
func (db *DB) AddServer(name, ipAddress, sshUser string, sshPort int, notes string) (int64, error) {
	result, err := db.conn.Exec(
		`INSERT INTO servers (name, ip_address, ssh_user, ssh_port, notes)
		 VALUES (?, ?, ?, ?, ?)`,
		name, ipAddress, sshUser, sshPort, notes,
	)
	if err != nil {
		return 0, fmt.Errorf("insert server: %w", err)
	}

	return result.LastInsertId()
}

// GetServer retrieves a server by ID
func (db *DB) GetServer(id int64) (*Server, error) {
	var server Server
	var lastSeen sql.NullTime

	err := db.conn.QueryRow(
		`SELECT id, name, ip_address, ssh_user, ssh_port, status, last_seen, created_at, notes
		 FROM servers WHERE id = ?`,
		id,
	).Scan(
		&server.ID, &server.Name, &server.IPAddress, &server.SSHUser, &server.SSHPort,
		&server.Status, &lastSeen, &server.CreatedAt, &server.Notes,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("server not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query server: %w", err)
	}

	if lastSeen.Valid {
		server.LastSeen = &lastSeen.Time
	}

	return &server, nil
}

// GetServerByName retrieves a server by name
func (db *DB) GetServerByName(name string) (*Server, error) {
	var server Server
	var lastSeen sql.NullTime

	err := db.conn.QueryRow(
		`SELECT id, name, ip_address, ssh_user, ssh_port, status, last_seen, created_at, notes
		 FROM servers WHERE name = ?`,
		name,
	).Scan(
		&server.ID, &server.Name, &server.IPAddress, &server.SSHUser, &server.SSHPort,
		&server.Status, &lastSeen, &server.CreatedAt, &server.Notes,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("server not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query server: %w", err)
	}

	if lastSeen.Valid {
		server.LastSeen = &lastSeen.Time
	}

	return &server, nil
}

// ListServers retrieves all servers
func (db *DB) ListServers(statusFilter string) ([]*Server, error) {
	query := `SELECT id, name, ip_address, ssh_user, ssh_port, status, last_seen, created_at, notes
	          FROM servers`

	var args []interface{}
	if statusFilter != "" {
		query += " WHERE status = ?"
		args = append(args, statusFilter)
	}

	query += " ORDER BY name ASC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query servers: %w", err)
	}
	defer rows.Close()

	var servers []*Server
	for rows.Next() {
		var server Server
		var lastSeen sql.NullTime

		if err := rows.Scan(
			&server.ID, &server.Name, &server.IPAddress, &server.SSHUser, &server.SSHPort,
			&server.Status, &lastSeen, &server.CreatedAt, &server.Notes,
		); err != nil {
			return nil, fmt.Errorf("scan server: %w", err)
		}

		if lastSeen.Valid {
			server.LastSeen = &lastSeen.Time
		}

		servers = append(servers, &server)
	}

	return servers, rows.Err()
}

// UpdateServerStatus updates a server's status and last_seen timestamp
func (db *DB) UpdateServerStatus(id int64, status string) error {
	result, err := db.conn.Exec(
		`UPDATE servers SET status = ?, last_seen = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update server status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("server not found")
	}

	return nil
}

// UpdateServerNotes updates a server's notes
func (db *DB) UpdateServerNotes(id int64, notes string) error {
	result, err := db.conn.Exec(
		`UPDATE servers SET notes = ? WHERE id = ?`,
		notes, id,
	)
	if err != nil {
		return fmt.Errorf("update server notes: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("server not found")
	}

	return nil
}

// RemoveServer deletes a server from the database
func (db *DB) RemoveServer(id int64) error {
	result, err := db.conn.Exec(`DELETE FROM servers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete server: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("server not found")
	}

	return nil
}

// ==================== Health Check Operations ====================

// AddHealthCheck records a health check result
func (db *DB) AddHealthCheck(serverID int64, checkType, status string, responseTimeMS int, errorMessage string) error {
	_, err := db.conn.Exec(
		`INSERT INTO health_checks (server_id, check_type, status, response_time_ms, error_message)
		 VALUES (?, ?, ?, ?, ?)`,
		serverID, checkType, status, responseTimeMS, errorMessage,
	)
	if err != nil {
		return fmt.Errorf("insert health check: %w", err)
	}

	return nil
}

// GetHealthCheckHistory retrieves health check history for a server
func (db *DB) GetHealthCheckHistory(serverID int64, limit int) ([]*HealthCheck, error) {
	query := `SELECT id, server_id, check_type, status, response_time_ms, error_message, timestamp
	          FROM health_checks
	          WHERE server_id = ?
	          ORDER BY timestamp DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.conn.Query(query, serverID)
	if err != nil {
		return nil, fmt.Errorf("query health checks: %w", err)
	}
	defer rows.Close()

	var checks []*HealthCheck
	for rows.Next() {
		var check HealthCheck
		var errorMsg sql.NullString

		if err := rows.Scan(
			&check.ID, &check.ServerID, &check.CheckType, &check.Status,
			&check.ResponseTimeMS, &errorMsg, &check.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scan health check: %w", err)
		}

		if errorMsg.Valid {
			check.ErrorMessage = errorMsg.String
		}

		checks = append(checks, &check)
	}

	return checks, rows.Err()
}

// CleanOldHealthChecks deletes health check records older than specified days
func (db *DB) CleanOldHealthChecks(days int) (int64, error) {
	result, err := db.conn.Exec(
		`DELETE FROM health_checks WHERE timestamp < datetime('now', ?)`,
		fmt.Sprintf("-%d days", days),
	)
	if err != nil {
		return 0, fmt.Errorf("delete old health checks: %w", err)
	}

	return result.RowsAffected()
}
