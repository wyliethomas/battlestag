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
