package models

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response is the standard API response wrapper
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// StoicThought represents a stoic thought entry
type StoicThought struct {
	ID        int       `json:"id"`
	Date      string    `json:"date"`
	Thought   string    `json:"thought"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

// TechTip represents a tech tip entry
type TechTip struct {
	ID        int       `json:"id"`
	Date      string    `json:"date"`
	Title     string    `json:"title"`
	Tip       string    `json:"tip"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string               `json:"status"`
	Timestamp time.Time            `json:"timestamp"`
	Agents    map[string]AgentInfo `json:"agents"`
}

// AgentInfo contains information about an agent's status
type AgentInfo struct {
	Available bool   `json:"available"`
	DBPath    string `json:"db_path"`
	Error     string `json:"error,omitempty"`
}

// Stats represents system statistics
type Stats struct {
	TotalThoughts   int       `json:"total_thoughts"`
	TotalTips       int       `json:"total_tips"`
	LastThoughtDate string    `json:"last_thought_date,omitempty"`
	LastTipDate     string    `json:"last_tip_date,omitempty"`
	Uptime          string    `json:"uptime"`
	Timestamp       time.Time `json:"timestamp"`
}

// PaginatedResponse wraps paginated data
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int         `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}

// Transaction represents a financial transaction from statements
type Transaction struct {
	ID            int       `json:"id"`
	Date          string    `json:"date"`
	Description   string    `json:"description"`
	Amount        float64   `json:"amount"`
	TransactionType string  `json:"transaction_type"`
	Category      string    `json:"category"`
	Merchant      string    `json:"merchant,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TransactionSummary represents aggregated transaction data
type TransactionSummary struct {
	TotalAmount    float64          `json:"total_amount"`
	TotalCount     int              `json:"total_count"`
	CountByType    map[string]int   `json:"count_by_type"`
	AmountByType   map[string]float64 `json:"amount_by_type"`
	StartDate      string           `json:"start_date"`
	EndDate        string           `json:"end_date"`
}

// ProcessPDFResponse represents the response from processing a PDF
type ProcessPDFResponse struct {
	TransactionsProcessed int      `json:"transactions_processed"`
	Errors               []string  `json:"errors,omitempty"`
	FileName             string    `json:"file_name"`
}

// Asset represents a financial asset
type Asset struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	Category       string     `json:"category"`
	CurrentValue   float64    `json:"current_value"`
	PurchasePrice  *float64   `json:"purchase_price,omitempty"`
	PurchaseDate   *string    `json:"purchase_date,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	IsRemoved      bool       `json:"is_removed"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastUpdated    time.Time  `json:"last_updated"`
}

// ValueHistory represents historical value data for an asset
type ValueHistory struct {
	ID          int       `json:"id"`
	AssetID     int       `json:"asset_id"`
	Value       float64   `json:"value"`
	RecordedAt  time.Time `json:"recorded_at"`
	Notes       string    `json:"notes,omitempty"`
}

// AssetSummary represents aggregated asset data
type AssetSummary struct {
	TotalValue     float64            `json:"total_value"`
	TotalCount     int                `json:"total_count"`
	CountByCategory map[string]int    `json:"count_by_category"`
	ValueByCategory map[string]float64 `json:"value_by_category"`
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

// BalanceHistory represents historical balance data for a liability
type BalanceHistory struct {
	ID          int       `json:"id"`
	LiabilityID int       `json:"liability_id"`
	Balance     float64   `json:"balance"`
	RecordedAt  time.Time `json:"recorded_at"`
	Notes       string    `json:"notes,omitempty"`
}

// LiabilitySummary represents aggregated liability data
type LiabilitySummary struct {
	TotalBalance   float64            `json:"total_balance"`
	TotalCount     int                `json:"total_count"`
	CountByType    map[string]int     `json:"count_by_type"`
	BalanceByType  map[string]float64 `json:"balance_by_type"`
}

// FinancialOverview represents a complete financial snapshot
type FinancialOverview struct {
	TotalAssets      float64   `json:"total_assets"`
	TotalLiabilities float64   `json:"total_liabilities"`
	NetWorth         float64   `json:"net_worth"`
	AssetCount       int       `json:"asset_count"`
	LiabilityCount   int       `json:"liability_count"`
	Timestamp        time.Time `json:"timestamp"`
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// WriteSuccess writes a successful JSON response
func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// WriteError writes an error JSON response
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, Response{
		Success: false,
		Error:   message,
	})
}
