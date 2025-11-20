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
	stoicDB              *sql.DB
	techDB               *sql.DB
	financialStatementDB *sql.DB
	financialAssetDB     *sql.DB
	financialLiabilityDB *sql.DB
	config               *config.Config
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

	// Connect to financial statement database
	if cfg.Agents.FinancialStatement.DBPath != "" {
		db, err := sql.Open("sqlite3", cfg.Agents.FinancialStatement.DBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open financial statement database: %w", err)
		}
		mgr.financialStatementDB = db
	}

	// Connect to financial asset database
	if cfg.Agents.FinancialAsset.DBPath != "" {
		db, err := sql.Open("sqlite3", cfg.Agents.FinancialAsset.DBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open financial asset database: %w", err)
		}
		mgr.financialAssetDB = db
	}

	// Connect to financial liability database
	if cfg.Agents.FinancialLiability.DBPath != "" {
		db, err := sql.Open("sqlite3", cfg.Agents.FinancialLiability.DBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open financial liability database: %w", err)
		}
		mgr.financialLiabilityDB = db
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
	if m.financialStatementDB != nil {
		m.financialStatementDB.Close()
	}
	if m.financialAssetDB != nil {
		m.financialAssetDB.Close()
	}
	if m.financialLiabilityDB != nil {
		m.financialLiabilityDB.Close()
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

// Financial Database Operations

// GetTransactionSummary gets aggregated transaction data
func (m *Manager) GetTransactionSummary(startDate, endDate string) (*models.TransactionSummary, error) {
	if m.financialStatementDB == nil {
		return nil, fmt.Errorf("financial statement database not available")
	}

	summary := &models.TransactionSummary{
		StartDate:    startDate,
		EndDate:      endDate,
		CountByType:  make(map[string]int),
		AmountByType: make(map[string]float64),
	}

	// Get total amount and count
	query := `SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM transactions WHERE date >= ? AND date <= ?`
	err := m.financialStatementDB.QueryRow(query, startDate, endDate).Scan(&summary.TotalAmount, &summary.TotalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction totals: %w", err)
	}

	// Get counts by type
	rows, err := m.financialStatementDB.Query(
		`SELECT transaction_type, COUNT(*), SUM(amount) FROM transactions
		 WHERE date >= ? AND date <= ?
		 GROUP BY transaction_type`,
		startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var txnType string
		var count int
		var amount float64
		if err := rows.Scan(&txnType, &count, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan transaction breakdown: %w", err)
		}
		summary.CountByType[txnType] = count
		summary.AmountByType[txnType] = amount
	}

	return summary, nil
}

// GetAssetSummary gets aggregated asset data
func (m *Manager) GetAssetSummary() (*models.AssetSummary, error) {
	if m.financialAssetDB == nil {
		return nil, fmt.Errorf("financial asset database not available")
	}

	summary := &models.AssetSummary{
		CountByCategory: make(map[string]int),
		ValueByCategory: make(map[string]float64),
	}

	// Get total value and count (excluding removed assets)
	query := `SELECT COALESCE(SUM(current_value), 0), COUNT(*) FROM assets WHERE is_removed = 0`
	err := m.financialAssetDB.QueryRow(query).Scan(&summary.TotalValue, &summary.TotalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset totals: %w", err)
	}

	// Get breakdown by category
	rows, err := m.financialAssetDB.Query(
		`SELECT category, COUNT(*), SUM(current_value) FROM assets
		 WHERE is_removed = 0
		 GROUP BY category`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int
		var value float64
		if err := rows.Scan(&category, &count, &value); err != nil {
			return nil, fmt.Errorf("failed to scan asset breakdown: %w", err)
		}
		summary.CountByCategory[category] = count
		summary.ValueByCategory[category] = value
	}

	return summary, nil
}

// GetLiabilitySummary gets aggregated liability data
func (m *Manager) GetLiabilitySummary() (*models.LiabilitySummary, error) {
	if m.financialLiabilityDB == nil {
		return nil, fmt.Errorf("financial liability database not available")
	}

	summary := &models.LiabilitySummary{
		CountByType:   make(map[string]int),
		BalanceByType: make(map[string]float64),
	}

	// Get total balance and count
	query := `SELECT COALESCE(SUM(current_balance), 0), COUNT(*) FROM liabilities`
	err := m.financialLiabilityDB.QueryRow(query).Scan(&summary.TotalBalance, &summary.TotalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get liability totals: %w", err)
	}

	// Get breakdown by type
	rows, err := m.financialLiabilityDB.Query(
		`SELECT liability_type, COUNT(*), SUM(current_balance) FROM liabilities
		 GROUP BY liability_type`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get liability breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var liabilityType string
		var count int
		var balance float64
		if err := rows.Scan(&liabilityType, &count, &balance); err != nil {
			return nil, fmt.Errorf("failed to scan liability breakdown: %w", err)
		}
		summary.CountByType[liabilityType] = count
		summary.BalanceByType[liabilityType] = balance
	}

	return summary, nil
}

// GetFinancialOverview gets a complete financial snapshot
func (m *Manager) GetFinancialOverview() (*models.FinancialOverview, error) {
	overview := &models.FinancialOverview{
		Timestamp: time.Now(),
	}

	// Get asset totals
	if m.financialAssetDB != nil {
		query := `SELECT COALESCE(SUM(current_value), 0), COUNT(*) FROM assets WHERE is_removed = 0`
		m.financialAssetDB.QueryRow(query).Scan(&overview.TotalAssets, &overview.AssetCount)
	}

	// Get liability totals
	if m.financialLiabilityDB != nil {
		query := `SELECT COALESCE(SUM(current_balance), 0), COUNT(*) FROM liabilities`
		m.financialLiabilityDB.QueryRow(query).Scan(&overview.TotalLiabilities, &overview.LiabilityCount)
	}

	// Calculate net worth
	overview.NetWorth = overview.TotalAssets - overview.TotalLiabilities

	return overview, nil
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
