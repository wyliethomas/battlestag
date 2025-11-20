package executor

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"agent-gateway/models"
)

// Executor runs agent programs and parses their output
type Executor struct {
	stoicPath              string
	techPath               string
	financialStatementPath string
	financialAssetPath     string
	financialLiabilityPath string
}

// NewExecutor creates a new executor
func NewExecutor(stoicPath, techPath, financialStatementPath, financialAssetPath, financialLiabilityPath string) *Executor {
	return &Executor{
		stoicPath:              stoicPath,
		techPath:               techPath,
		financialStatementPath: financialStatementPath,
		financialAssetPath:     financialAssetPath,
		financialLiabilityPath: financialLiabilityPath,
	}
}

// StoicOutput represents the JSON output from stoic program
type StoicOutput struct {
	Date      string `json:"date"`
	Thought   string `json:"thought"`
	CreatedAt string `json:"created_at"`
}

// TechOutput represents the JSON output from tech program
type TechOutput struct {
	Date      string `json:"date"`
	Title     string `json:"title"`
	Tip       string `json:"tip"`
	Category  string `json:"category"`
	CreatedAt string `json:"created_at"`
}

// GetStoicToday gets today's stoic thought
func (e *Executor) GetStoicToday() (*models.StoicThought, error) {
	return e.GetStoicByDate("")
}

// GetStoicByDate gets a stoic thought for a specific date (empty string for today)
func (e *Executor) GetStoicByDate(date string) (*models.StoicThought, error) {
	args := []string{"--json"}
	if date != "" {
		args = append(args, "--date", date)
	}

	cmd := exec.Command(e.stoicPath, args...)
	// Set working directory to the program's directory so it can find its database
	cmd.Dir = filepath.Dir(e.stoicPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute stoic program: %w (output: %s)", err, string(output))
	}

	var stoicOut StoicOutput
	if err := json.Unmarshal(output, &stoicOut); err != nil {
		return nil, fmt.Errorf("failed to parse stoic output: %w (output: %s)", err, string(output))
	}

	createdAt, err := time.Parse(time.RFC3339, stoicOut.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at timestamp: %w", err)
	}

	// Clean up thought - remove surrounding quotes if present
	thought := strings.Trim(stoicOut.Thought, "\"")

	return &models.StoicThought{
		Date:      stoicOut.Date,
		Thought:   thought,
		CreatedAt: createdAt,
		Source:    "Stoic Agent", // Default source since program doesn't provide it
	}, nil
}

// GetTechToday gets today's tech tip
func (e *Executor) GetTechToday() (*models.TechTip, error) {
	return e.GetTechByDate("")
}

// GetTechByDate gets a tech tip for a specific date (empty string for today)
func (e *Executor) GetTechByDate(date string) (*models.TechTip, error) {
	args := []string{"--json"}
	if date != "" {
		args = append(args, "--date", date)
	}

	cmd := exec.Command(e.techPath, args...)
	// Set working directory to the program's directory so it can find its database
	cmd.Dir = filepath.Dir(e.techPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute tech program: %w (output: %s)", err, string(output))
	}

	var techOut TechOutput
	if err := json.Unmarshal(output, &techOut); err != nil {
		return nil, fmt.Errorf("failed to parse tech output: %w (output: %s)", err, string(output))
	}

	createdAt, err := time.Parse(time.RFC3339, techOut.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at timestamp: %w", err)
	}

	return &models.TechTip{
		Date:      techOut.Date,
		Title:     techOut.Title,
		Tip:       techOut.Tip,
		Category:  techOut.Category,
		CreatedAt: createdAt,
	}, nil
}

// Financial Statement Processor Methods

// FinancialStatementQueryOutput represents the JSON output from query command
type FinancialStatementQueryOutput struct {
	Transactions []models.Transaction `json:"transactions"`
	Count        int                  `json:"count"`
}

// ProcessPDF processes a PDF file and returns the result
func (e *Executor) ProcessPDF(pdfPath string, accountType string, force bool) (*models.ProcessPDFResponse, error) {
	args := []string{"process", pdfPath}
	if accountType != "" {
		args = append(args, "--account-type", accountType)
	}
	if force {
		args = append(args, "--force")
	}

	cmd := exec.Command(e.financialStatementPath, args...)
	cmd.Dir = filepath.Dir(e.financialStatementPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to process PDF: %w (output: %s)", err, string(output))
	}

	// Parse output - the processor outputs success/error info
	var result struct {
		Success               bool     `json:"success"`
		TransactionsProcessed int      `json:"transactions_processed"`
		Errors                []string `json:"errors,omitempty"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse process output: %w (output: %s)", err, string(output))
	}

	return &models.ProcessPDFResponse{
		TransactionsProcessed: result.TransactionsProcessed,
		Errors:                result.Errors,
		FileName:              filepath.Base(pdfPath),
	}, nil
}

// QueryTransactions queries transactions with filters
func (e *Executor) QueryTransactions(startDate, endDate, txnType, merchant string) ([]models.Transaction, error) {
	args := []string{"query"}
	if startDate != "" {
		args = append(args, "--start-date", startDate)
	}
	if endDate != "" {
		args = append(args, "--end-date", endDate)
	}
	if txnType != "" {
		args = append(args, "--type", txnType)
	}
	if merchant != "" {
		args = append(args, "--merchant", merchant)
	}

	cmd := exec.Command(e.financialStatementPath, args...)
	cmd.Dir = filepath.Dir(e.financialStatementPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w (output: %s)", err, string(output))
	}

	var queryOut FinancialStatementQueryOutput
	if err := json.Unmarshal(output, &queryOut); err != nil {
		return nil, fmt.Errorf("failed to parse query output: %w (output: %s)", err, string(output))
	}

	return queryOut.Transactions, nil
}

// Financial Asset Tracker Methods

// AssetListOutput represents the JSON output from list command
type AssetListOutput struct {
	Assets []models.Asset `json:"assets"`
	Count  int            `json:"count"`
}

// AddAsset adds a new asset
func (e *Executor) AddAsset(name, category string, currentValue float64, purchasePrice *float64, purchaseDate, notes string) (*models.Asset, error) {
	args := []string{"add", "--name", name, "--category", category, "--current-value", fmt.Sprintf("%.2f", currentValue)}
	if purchasePrice != nil {
		args = append(args, "--purchase-price", fmt.Sprintf("%.2f", *purchasePrice))
	}
	if purchaseDate != "" {
		args = append(args, "--purchase-date", purchaseDate)
	}
	if notes != "" {
		args = append(args, "--notes", notes)
	}

	cmd := exec.Command(e.financialAssetPath, args...)
	cmd.Dir = filepath.Dir(e.financialAssetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to add asset: %w (output: %s)", err, string(output))
	}

	var result struct {
		Success bool          `json:"success"`
		Asset   models.Asset  `json:"asset"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse add output: %w (output: %s)", err, string(output))
	}

	return &result.Asset, nil
}

// UpdateAssetValue updates an asset's value
func (e *Executor) UpdateAssetValue(name string, newValue float64, notes string) error {
	args := []string{"update", name, "--current-value", fmt.Sprintf("%.2f", newValue)}
	if notes != "" {
		args = append(args, "--notes", notes)
	}

	cmd := exec.Command(e.financialAssetPath, args...)
	cmd.Dir = filepath.Dir(e.financialAssetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update asset: %w (output: %s)", err, string(output))
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("failed to parse update output: %w (output: %s)", err, string(output))
	}

	if !result.Success {
		return fmt.Errorf("update failed")
	}

	return nil
}

// RemoveAsset soft-deletes an asset
func (e *Executor) RemoveAsset(name string) error {
	args := []string{"remove", name}

	cmd := exec.Command(e.financialAssetPath, args...)
	cmd.Dir = filepath.Dir(e.financialAssetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove asset: %w (output: %s)", err, string(output))
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("failed to parse remove output: %w (output: %s)", err, string(output))
	}

	return nil
}

// RestoreAsset restores a removed asset
func (e *Executor) RestoreAsset(name string) error {
	args := []string{"restore", name}

	cmd := exec.Command(e.financialAssetPath, args...)
	cmd.Dir = filepath.Dir(e.financialAssetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restore asset: %w (output: %s)", err, string(output))
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("failed to parse restore output: %w (output: %s)", err, string(output))
	}

	return nil
}

// ListAssets lists all assets
func (e *Executor) ListAssets(includeRemoved bool, category string) ([]models.Asset, error) {
	args := []string{"list"}
	if includeRemoved {
		args = append(args, "--include-removed")
	}
	if category != "" {
		args = append(args, "--category", category)
	}

	cmd := exec.Command(e.financialAssetPath, args...)
	cmd.Dir = filepath.Dir(e.financialAssetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w (output: %s)", err, string(output))
	}

	var listOut AssetListOutput
	if err := json.Unmarshal(output, &listOut); err != nil {
		return nil, fmt.Errorf("failed to parse list output: %w (output: %s)", err, string(output))
	}

	return listOut.Assets, nil
}

// GetAsset gets a specific asset
func (e *Executor) GetAsset(name string) (*models.Asset, error) {
	args := []string{"get", name}

	cmd := exec.Command(e.financialAssetPath, args...)
	cmd.Dir = filepath.Dir(e.financialAssetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w (output: %s)", err, string(output))
	}

	var result struct {
		Asset models.Asset `json:"asset"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse get output: %w (output: %s)", err, string(output))
	}

	return &result.Asset, nil
}

// Financial Liability Tracker Methods

// LiabilityListOutput represents the JSON output from list command
type LiabilityListOutput struct {
	Liabilities []models.Liability `json:"liabilities"`
	Count       int                `json:"count"`
}

// AddLiability adds a new liability
func (e *Executor) AddLiability(req *models.AddLiabilityRequest) (*models.Liability, error) {
	args := []string{"add", "--name", req.Name, "--type", req.Type, "--balance", fmt.Sprintf("%.2f", req.Balance)}
	if req.OriginalAmount != nil {
		args = append(args, "--original", fmt.Sprintf("%.2f", *req.OriginalAmount))
	}
	if req.CreditLimit != nil {
		args = append(args, "--limit", fmt.Sprintf("%.2f", *req.CreditLimit))
	}
	if req.InterestRate != nil {
		args = append(args, "--rate", fmt.Sprintf("%.2f", *req.InterestRate))
	}
	if req.MinimumPayment != nil {
		args = append(args, "--min-payment", fmt.Sprintf("%.2f", *req.MinimumPayment))
	}
	if req.CreditorName != "" {
		args = append(args, "--creditor", req.CreditorName)
	}
	if req.AccountLast4 != "" {
		args = append(args, "--last4", req.AccountLast4)
	}
	if req.OpenedDate != nil {
		args = append(args, "--opened", *req.OpenedDate)
	}
	if req.Notes != "" {
		args = append(args, "--notes", req.Notes)
	}

	cmd := exec.Command(e.financialLiabilityPath, args...)
	cmd.Dir = filepath.Dir(e.financialLiabilityPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to add liability: %w (output: %s)", err, string(output))
	}

	var result struct {
		Success    bool             `json:"success"`
		Liability  models.Liability `json:"liability"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse add output: %w (output: %s)", err, string(output))
	}

	return &result.Liability, nil
}

// UpdateLiability updates a liability's balance
func (e *Executor) UpdateLiability(name string, newBalance float64, notes string) error {
	args := []string{"update", name, "--balance", fmt.Sprintf("%.2f", newBalance)}
	if notes != "" {
		args = append(args, "--notes", notes)
	}

	cmd := exec.Command(e.financialLiabilityPath, args...)
	cmd.Dir = filepath.Dir(e.financialLiabilityPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update liability: %w (output: %s)", err, string(output))
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("failed to parse update output: %w (output: %s)", err, string(output))
	}

	return nil
}

// DeleteLiability deletes a liability
func (e *Executor) DeleteLiability(name string) error {
	args := []string{"delete", name}

	cmd := exec.Command(e.financialLiabilityPath, args...)
	cmd.Dir = filepath.Dir(e.financialLiabilityPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete liability: %w (output: %s)", err, string(output))
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("failed to parse delete output: %w (output: %s)", err, string(output))
	}

	return nil
}

// ListLiabilities lists all liabilities
func (e *Executor) ListLiabilities(liabilityType string) ([]models.Liability, error) {
	args := []string{"list"}
	if liabilityType != "" {
		args = append(args, "--type", liabilityType)
	}

	cmd := exec.Command(e.financialLiabilityPath, args...)
	cmd.Dir = filepath.Dir(e.financialLiabilityPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list liabilities: %w (output: %s)", err, string(output))
	}

	var listOut LiabilityListOutput
	if err := json.Unmarshal(output, &listOut); err != nil {
		return nil, fmt.Errorf("failed to parse list output: %w (output: %s)", err, string(output))
	}

	return listOut.Liabilities, nil
}

// GetLiability gets a specific liability
func (e *Executor) GetLiability(name string, includeHistory bool) (*models.Liability, []models.BalanceHistory, error) {
	args := []string{"get", name}
	if includeHistory {
		args = append(args, "--history")
	}

	cmd := exec.Command(e.financialLiabilityPath, args...)
	cmd.Dir = filepath.Dir(e.financialLiabilityPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get liability: %w (output: %s)", err, string(output))
	}

	var result struct {
		Liability      models.Liability        `json:"liability"`
		BalanceHistory []models.BalanceHistory `json:"balance_history,omitempty"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to parse get output: %w (output: %s)", err, string(output))
	}

	return &result.Liability, result.BalanceHistory, nil
}

// GetTotalLiabilities gets total liability balance
func (e *Executor) GetTotalLiabilities() (float64, error) {
	args := []string{"total"}

	cmd := exec.Command(e.financialLiabilityPath, args...)
	cmd.Dir = filepath.Dir(e.financialLiabilityPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to get total: %w (output: %s)", err, string(output))
	}

	var result struct {
		TotalBalance float64 `json:"total_balance"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return 0, fmt.Errorf("failed to parse total output: %w (output: %s)", err, string(output))
	}

	return result.TotalBalance, nil
}
