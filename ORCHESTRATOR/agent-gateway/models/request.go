package models

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Request body structs for financial operations

// AddAssetRequest represents a request to add a new asset
type AddAssetRequest struct {
	Name          string   `json:"name"`
	Category      string   `json:"category"`
	CurrentValue  float64  `json:"current_value"`
	PurchasePrice *float64 `json:"purchase_price,omitempty"`
	PurchaseDate  *string  `json:"purchase_date,omitempty"`
	Notes         string   `json:"notes,omitempty"`
}

// UpdateAssetRequest represents a request to update an asset's value
type UpdateAssetRequest struct {
	CurrentValue float64 `json:"current_value"`
	Notes        string  `json:"notes,omitempty"`
}

// AddLiabilityRequest represents a request to add a new liability
type AddLiabilityRequest struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Balance        float64  `json:"balance"`
	OriginalAmount *float64 `json:"original_amount,omitempty"`
	CreditLimit    *float64 `json:"credit_limit,omitempty"`
	InterestRate   *float64 `json:"interest_rate,omitempty"`
	MinimumPayment *float64 `json:"minimum_payment,omitempty"`
	CreditorName   string   `json:"creditor_name,omitempty"`
	AccountLast4   string   `json:"account_last4,omitempty"`
	OpenedDate     *string  `json:"opened_date,omitempty"`
	Notes          string   `json:"notes,omitempty"`
}

// UpdateLiabilityRequest represents a request to update a liability's balance
type UpdateLiabilityRequest struct {
	Balance float64 `json:"balance"`
	Notes   string  `json:"notes,omitempty"`
}

// Validation functions

// ValidateDate validates a date string in YYYY-MM-DD format
func ValidateDate(dateStr string) error {
	if dateStr == "" {
		return nil // Optional dates are allowed
	}
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	return nil
}

// ValidatePositiveFloat validates that a number is positive
func ValidatePositiveFloat(value float64, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}

// ValidateNonEmpty validates that a string is not empty
func ValidateNonEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateLiabilityType validates that the liability type is valid
func ValidateLiabilityType(liabilityType string) error {
	validTypes := map[string]bool{
		"credit-card":   true,
		"auto-loan":     true,
		"mortgage":      true,
		"student-loan":  true,
		"personal-loan": true,
		"medical-debt":  true,
	}
	if !validTypes[liabilityType] {
		return fmt.Errorf("invalid liability type, must be one of: credit-card, auto-loan, mortgage, student-loan, personal-loan, medical-debt")
	}
	return nil
}

// Request parsing helpers

// ParseJSONBody parses JSON from request body into a struct
func ParseJSONBody(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// GetQueryParam gets a query parameter with a default value
func GetQueryParam(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetQueryParamInt gets an integer query parameter with a default value
func GetQueryParamInt(r *http.Request, key string, defaultValue, maxValue int) (int, error) {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: must be an integer", key)
	}

	if value <= 0 {
		return 0, fmt.Errorf("invalid %s: must be positive", key)
	}

	if maxValue > 0 && value > maxValue {
		return 0, fmt.Errorf("invalid %s: maximum value is %d", key, maxValue)
	}

	return value, nil
}

// GetQueryParamBool gets a boolean query parameter with a default value
func GetQueryParamBool(r *http.Request, key string, defaultValue bool) bool {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// File upload helpers

const maxUploadSize = 10 * 1024 * 1024 // 10MB

// SaveUploadedFile saves an uploaded file to a temporary location
func SaveUploadedFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validate file size
	if header.Size > maxUploadSize {
		return "", fmt.Errorf("file too large: maximum size is 10MB")
	}

	// Validate file type (PDF only)
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		return "", fmt.Errorf("invalid file type: only PDF files are allowed")
	}

	// Create temp directory if it doesn't exist
	tempDir := filepath.Join(os.TempDir(), "agent-gateway-uploads")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate unique filename
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d_%s", timestamp, filepath.Base(header.Filename))
	filePath := filepath.Join(tempDir, filename)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath) // Clean up on error
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filePath, nil
}

// CleanupTempFile removes a temporary file
func CleanupTempFile(filePath string) {
	if filePath != "" {
		os.Remove(filePath)
	}
}

// Validate request structs

// Validate validates an AddAssetRequest
func (r *AddAssetRequest) Validate() error {
	if err := ValidateNonEmpty(r.Name, "name"); err != nil {
		return err
	}
	if err := ValidateNonEmpty(r.Category, "category"); err != nil {
		return err
	}
	if err := ValidatePositiveFloat(r.CurrentValue, "current_value"); err != nil {
		return err
	}
	if r.PurchaseDate != nil {
		if err := ValidateDate(*r.PurchaseDate); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates an UpdateAssetRequest
func (r *UpdateAssetRequest) Validate() error {
	if err := ValidatePositiveFloat(r.CurrentValue, "current_value"); err != nil {
		return err
	}
	return nil
}

// Validate validates an AddLiabilityRequest
func (r *AddLiabilityRequest) Validate() error {
	if err := ValidateNonEmpty(r.Name, "name"); err != nil {
		return err
	}
	if err := ValidateNonEmpty(r.Type, "type"); err != nil {
		return err
	}
	if err := ValidateLiabilityType(r.Type); err != nil {
		return err
	}
	if err := ValidatePositiveFloat(r.Balance, "balance"); err != nil {
		return err
	}
	if r.OpenedDate != nil {
		if err := ValidateDate(*r.OpenedDate); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates an UpdateLiabilityRequest
func (r *UpdateLiabilityRequest) Validate() error {
	if err := ValidatePositiveFloat(r.Balance, "balance"); err != nil {
		return err
	}
	return nil
}
