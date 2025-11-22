package models

import "time"

// APIResponse is the standard response wrapper from the API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Time    time.Time `json:"time"`
}

// StatsResponse represents system statistics
type StatsResponse struct {
	Uptime          string `json:"uptime"`
	RequestCount    int    `json:"request_count"`
	ActiveAgents    int    `json:"active_agents"`
	MemoryUsage     string `json:"memory_usage"`
}

// StoicThought represents a stoic philosophy thought
type StoicThought struct {
	ID        int       `json:"id"`
	Date      string    `json:"date"`
	Thought   string    `json:"thought"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

// TechTip represents a technology tip
type TechTip struct {
	ID        int       `json:"id"`
	Date      string    `json:"date"`
	Title     string    `json:"title"`
	Tip       string    `json:"tip"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

// Asset represents a financial asset
type Asset struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	CurrentValue  float64   `json:"current_value"`
	PurchasePrice float64   `json:"purchase_price"`
	PurchaseDate  string    `json:"purchase_date"`
	Notes         string    `json:"notes,omitempty"`
	IsRemoved     bool      `json:"is_removed"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Liability represents a financial liability
type Liability struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	LiabilityType  string    `json:"liability_type"`
	CurrentBalance float64   `json:"current_balance"`
	CreditLimit    float64   `json:"credit_limit,omitempty"`
	InterestRate   float64   `json:"interest_rate,omitempty"`
	MinimumPayment float64   `json:"minimum_payment,omitempty"`
	CreditorName   string    `json:"creditor_name,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// FinancialOverview represents the complete financial snapshot
type FinancialOverview struct {
	TotalAssets      float64   `json:"total_assets"`
	TotalLiabilities float64   `json:"total_liabilities"`
	NetWorth         float64   `json:"net_worth"`
	AssetCount       int       `json:"asset_count"`
	LiabilityCount   int       `json:"liability_count"`
	Timestamp        time.Time `json:"timestamp"`
}

// Transaction represents a bank transaction
type Transaction struct {
	ID          int       `json:"id"`
	Date        string    `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Merchant    string    `json:"merchant,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// AssetSummary represents asset summary by category
type AssetSummary struct {
	Category     string  `json:"category"`
	TotalValue   float64 `json:"total_value"`
	AssetCount   int     `json:"asset_count"`
}

// LiabilitySummary represents liability summary by type
type LiabilitySummary struct {
	LiabilityType  string  `json:"liability_type"`
	TotalBalance   float64 `json:"total_balance"`
	LiabilityCount int     `json:"liability_count"`
}
