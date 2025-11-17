package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"agent-gateway/config"
	"agent-gateway/models"
)

// Manager manages connections to all agent databases
type Manager struct {
	stoicDB *sql.DB
	techDB  *sql.DB
	config  *config.Config
}

// NewManager creates a new database manager
func NewManager(cfg *config.Config) (*Manager, error) {
	mgr := &Manager{
		config: cfg,
	}

	// Connect to stoic database
	if cfg.Agents.Stoic.DBPath != "" {
		db, err := sql.Open("sqlite3", cfg.Agents.Stoic.DBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open stoic database: %w", err)
		}
		mgr.stoicDB = db
	}

	// Connect to tech database
	if cfg.Agents.Tech.DBPath != "" {
		db, err := sql.Open("sqlite3", cfg.Agents.Tech.DBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open tech database: %w", err)
		}
		mgr.techDB = db
	}

	return mgr, nil
}

// Close closes all database connections
func (m *Manager) Close() error {
	if m.stoicDB != nil {
		m.stoicDB.Close()
	}
	if m.techDB != nil {
		m.techDB.Close()
	}
	return nil
}

// Stoic Database Operations

// GetStoicToday retrieves today's stoic thought
func (m *Manager) GetStoicToday() (*models.StoicThought, error) {
	today := time.Now().Format("2006-01-02")
	return m.GetStoicByDate(today)
}

// GetStoicByDate retrieves a stoic thought for a specific date
func (m *Manager) GetStoicByDate(date string) (*models.StoicThought, error) {
	if m.stoicDB == nil {
		return nil, fmt.Errorf("stoic database not available")
	}

	var thought models.StoicThought
	query := `SELECT id, date, thought, source, created_at FROM stoic_thoughts WHERE date = ? LIMIT 1`

	err := m.stoicDB.QueryRow(query, date).Scan(
		&thought.ID,
		&thought.Date,
		&thought.Thought,
		&thought.Source,
		&thought.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no thought found for date %s", date)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query stoic thought: %w", err)
	}

	return &thought, nil
}

// GetStoicRandom retrieves a random stoic thought
func (m *Manager) GetStoicRandom() (*models.StoicThought, error) {
	if m.stoicDB == nil {
		return nil, fmt.Errorf("stoic database not available")
	}

	var thought models.StoicThought
	query := `SELECT id, date, thought, source, created_at FROM stoic_thoughts ORDER BY RANDOM() LIMIT 1`

	err := m.stoicDB.QueryRow(query).Scan(
		&thought.ID,
		&thought.Date,
		&thought.Thought,
		&thought.Source,
		&thought.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query random stoic thought: %w", err)
	}

	return &thought, nil
}

// GetStoicLatest retrieves the latest N stoic thoughts
func (m *Manager) GetStoicLatest(n int) ([]models.StoicThought, error) {
	if m.stoicDB == nil {
		return nil, fmt.Errorf("stoic database not available")
	}

	query := `SELECT id, date, thought, source, created_at FROM stoic_thoughts ORDER BY date DESC LIMIT ?`
	rows, err := m.stoicDB.Query(query, n)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest stoic thoughts: %w", err)
	}
	defer rows.Close()

	var thoughts []models.StoicThought
	for rows.Next() {
		var thought models.StoicThought
		if err := rows.Scan(&thought.ID, &thought.Date, &thought.Thought, &thought.Source, &thought.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan stoic thought: %w", err)
		}
		thoughts = append(thoughts, thought)
	}

	return thoughts, nil
}

// GetStoicAll retrieves all stoic thoughts with pagination
func (m *Manager) GetStoicAll(page, pageSize int) (*models.PaginatedResponse, error) {
	if m.stoicDB == nil {
		return nil, fmt.Errorf("stoic database not available")
	}

	// Get total count
	var totalCount int
	err := m.stoicDB.QueryRow(`SELECT COUNT(*) FROM stoic_thoughts`).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count stoic thoughts: %w", err)
	}

	// Calculate pagination
	offset := (page - 1) * pageSize
	totalPages := (totalCount + pageSize - 1) / pageSize

	// Get paginated data
	query := `SELECT id, date, thought, source, created_at FROM stoic_thoughts ORDER BY date DESC LIMIT ? OFFSET ?`
	rows, err := m.stoicDB.Query(query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query stoic thoughts: %w", err)
	}
	defer rows.Close()

	var thoughts []models.StoicThought
	for rows.Next() {
		var thought models.StoicThought
		if err := rows.Scan(&thought.ID, &thought.Date, &thought.Thought, &thought.Source, &thought.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan stoic thought: %w", err)
		}
		thoughts = append(thoughts, thought)
	}

	return &models.PaginatedResponse{
		Data:       thoughts,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

// Tech Database Operations

// GetTechToday retrieves today's tech tip
func (m *Manager) GetTechToday() (*models.TechTip, error) {
	today := time.Now().Format("2006-01-02")
	return m.GetTechByDate(today)
}

// GetTechByDate retrieves a tech tip for a specific date
func (m *Manager) GetTechByDate(date string) (*models.TechTip, error) {
	if m.techDB == nil {
		return nil, fmt.Errorf("tech database not available")
	}

	var tip models.TechTip
	query := `SELECT id, date, title, tip, category, created_at FROM tech_tips WHERE date = ? LIMIT 1`

	err := m.techDB.QueryRow(query, date).Scan(
		&tip.ID,
		&tip.Date,
		&tip.Title,
		&tip.Tip,
		&tip.Category,
		&tip.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no tip found for date %s", date)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query tech tip: %w", err)
	}

	return &tip, nil
}

// GetTechRandom retrieves a random tech tip
func (m *Manager) GetTechRandom() (*models.TechTip, error) {
	if m.techDB == nil {
		return nil, fmt.Errorf("tech database not available")
	}

	var tip models.TechTip
	query := `SELECT id, date, title, tip, category, created_at FROM tech_tips ORDER BY RANDOM() LIMIT 1`

	err := m.techDB.QueryRow(query).Scan(
		&tip.ID,
		&tip.Date,
		&tip.Title,
		&tip.Tip,
		&tip.Category,
		&tip.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query random tech tip: %w", err)
	}

	return &tip, nil
}

// GetTechLatest retrieves the latest N tech tips
func (m *Manager) GetTechLatest(n int) ([]models.TechTip, error) {
	if m.techDB == nil {
		return nil, fmt.Errorf("tech database not available")
	}

	query := `SELECT id, date, title, tip, category, created_at FROM tech_tips ORDER BY date DESC LIMIT ?`
	rows, err := m.techDB.Query(query, n)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest tech tips: %w", err)
	}
	defer rows.Close()

	var tips []models.TechTip
	for rows.Next() {
		var tip models.TechTip
		if err := rows.Scan(&tip.ID, &tip.Date, &tip.Title, &tip.Tip, &tip.Category, &tip.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tech tip: %w", err)
		}
		tips = append(tips, tip)
	}

	return tips, nil
}

// GetTechByCategory retrieves tech tips filtered by category
func (m *Manager) GetTechByCategory(category string) ([]models.TechTip, error) {
	if m.techDB == nil {
		return nil, fmt.Errorf("tech database not available")
	}

	query := `SELECT id, date, title, tip, category, created_at FROM tech_tips WHERE category = ? ORDER BY date DESC`
	rows, err := m.techDB.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query tech tips by category: %w", err)
	}
	defer rows.Close()

	var tips []models.TechTip
	for rows.Next() {
		var tip models.TechTip
		if err := rows.Scan(&tip.ID, &tip.Date, &tip.Title, &tip.Tip, &tip.Category, &tip.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tech tip: %w", err)
		}
		tips = append(tips, tip)
	}

	return tips, nil
}

// GetTechAll retrieves all tech tips with pagination
func (m *Manager) GetTechAll(page, pageSize int) (*models.PaginatedResponse, error) {
	if m.techDB == nil {
		return nil, fmt.Errorf("tech database not available")
	}

	// Get total count
	var totalCount int
	err := m.techDB.QueryRow(`SELECT COUNT(*) FROM tech_tips`).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count tech tips: %w", err)
	}

	// Calculate pagination
	offset := (page - 1) * pageSize
	totalPages := (totalCount + pageSize - 1) / pageSize

	// Get paginated data
	query := `SELECT id, date, title, tip, category, created_at FROM tech_tips ORDER BY date DESC LIMIT ? OFFSET ?`
	rows, err := m.techDB.Query(query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query tech tips: %w", err)
	}
	defer rows.Close()

	var tips []models.TechTip
	for rows.Next() {
		var tip models.TechTip
		if err := rows.Scan(&tip.ID, &tip.Date, &tip.Title, &tip.Tip, &tip.Category, &tip.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tech tip: %w", err)
		}
		tips = append(tips, tip)
	}

	return &models.PaginatedResponse{
		Data:       tips,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

// Health and Stats Operations

// CheckHealth checks the health of all agent databases
func (m *Manager) CheckHealth() models.HealthStatus {
	status := models.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Agents:    make(map[string]models.AgentInfo),
	}

	// Check stoic DB
	stoicAvailable := false
	stoicError := ""
	if m.stoicDB != nil {
		if err := m.stoicDB.Ping(); err != nil {
			stoicError = err.Error()
			status.Status = "degraded"
		} else {
			stoicAvailable = true
		}
	} else {
		stoicError = "database not configured"
		status.Status = "degraded"
	}
	status.Agents["stoic"] = models.AgentInfo{
		Available: stoicAvailable,
		DBPath:    m.config.Agents.Stoic.DBPath,
		Error:     stoicError,
	}

	// Check tech DB
	techAvailable := false
	techError := ""
	if m.techDB != nil {
		if err := m.techDB.Ping(); err != nil {
			techError = err.Error()
			status.Status = "degraded"
		} else {
			techAvailable = true
		}
	} else {
		techError = "database not configured"
		status.Status = "degraded"
	}
	status.Agents["tech"] = models.AgentInfo{
		Available: techAvailable,
		DBPath:    m.config.Agents.Tech.DBPath,
		Error:     techError,
	}

	return status
}

// GetStats retrieves system statistics
func (m *Manager) GetStats(startTime time.Time) (*models.Stats, error) {
	stats := &models.Stats{
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).String(),
	}

	// Count stoic thoughts
	if m.stoicDB != nil {
		var count int
		var lastDate string
		m.stoicDB.QueryRow(`SELECT COUNT(*) FROM stoic_thoughts`).Scan(&count)
		m.stoicDB.QueryRow(`SELECT date FROM stoic_thoughts ORDER BY date DESC LIMIT 1`).Scan(&lastDate)
		stats.TotalThoughts = count
		stats.LastThoughtDate = lastDate
	}

	// Count tech tips
	if m.techDB != nil {
		var count int
		var lastDate string
		m.techDB.QueryRow(`SELECT COUNT(*) FROM tech_tips`).Scan(&count)
		m.techDB.QueryRow(`SELECT date FROM tech_tips ORDER BY date DESC LIMIT 1`).Scan(&lastDate)
		stats.TotalTips = count
		stats.LastTipDate = lastDate
	}

	return stats, nil
}
