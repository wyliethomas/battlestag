package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"financial-statement-processor/db"

	"github.com/ledongthuc/pdf"
)

// StatementData represents parsed statement information
type StatementData struct {
	AccountName   string
	AccountLast4  string
	StatementDate time.Time
	Transactions  []*db.Transaction
}

// LLMStatementResponse represents the JSON structure expected from the LLM
type LLMStatementResponse struct {
	AccountName   string           `json:"account_name"`
	AccountLast4  string           `json:"account_last4"`
	StatementDate string           `json:"statement_date"`
	Transactions  []LLMTransaction `json:"transactions"`
}

// LLMTransaction represents a transaction as returned by the LLM
type LLMTransaction struct {
	TransactionDate string   `json:"transaction_date"`
	PostDate        *string  `json:"post_date"`
	Description     string   `json:"description"`
	Amount          float64  `json:"amount"`
	TransactionType string   `json:"transaction_type"`
	Balance         *float64 `json:"balance"`
}

// ParseFile parses a bank statement file (PDF or image) using local LLM
func ParseFile(filePath, ollamaHost, ollamaModel string) (*StatementData, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return parsePDF(filePath, ollamaHost, ollamaModel)
	case ".jpg", ".jpeg", ".png", ".tiff", ".tif":
		return parseImage(filePath, ollamaHost, ollamaModel)
	default:
		return nil, fmt.Errorf("unsupported file type: %s (supported: PDF, JPG, PNG, TIFF)", ext)
	}
}

// parsePDF extracts transaction data from a PDF bank statement
func parsePDF(filePath, ollamaHost, ollamaModel string) (*StatementData, error) {
	// Extract text from PDF
	text, err := extractPDFText(filePath)
	if err != nil {
		return nil, fmt.Errorf("extract PDF text: %w", err)
	}

	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("PDF contains no text (may be a scanned image)")
	}

	// Parse using LLM
	return parseWithLLM(text, filePath, ollamaHost, ollamaModel)
}

// extractPDFText extracts all text content from a PDF file
func extractPDFText(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open PDF: %w", err)
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("stat PDF: %w", err)
	}

	pdfReader, err := pdf.NewReader(f, fileInfo.Size())
	if err != nil {
		return "", fmt.Errorf("create PDF reader: %w", err)
	}

	var fullText strings.Builder
	numPages := pdfReader.NumPage()

	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			// Continue with other pages if one fails
			continue
		}

		fullText.WriteString(text)
		fullText.WriteString("\n")
	}

	return fullText.String(), nil
}

// parseImage extracts transaction data from an image using OCR
func parseImage(filePath, ollamaHost, ollamaModel string) (*StatementData, error) {
	// TODO: Implement OCR parsing
	// For now, return error suggesting PDF conversion
	return nil, fmt.Errorf("image OCR not yet implemented - please convert to PDF or use tesseract manually")
}

// parseWithLLM sends extracted text to local LLM for structured parsing
func parseWithLLM(text, sourceFile, ollamaHost, ollamaModel string) (*StatementData, error) {
	// Create LLM client
	client := NewOllamaClient(ollamaHost, ollamaModel)

	// Health check
	if err := client.HealthCheck(); err != nil {
		return nil, fmt.Errorf("ollama health check failed: %w (make sure Ollama is running at %s)", err, ollamaHost)
	}

	// Send to LLM for parsing
	responseJSON, err := client.ParseStatementText(text)
	if err != nil {
		return nil, fmt.Errorf("LLM parsing failed: %w", err)
	}

	// Parse LLM response
	var llmResp LLMStatementResponse
	if err := json.Unmarshal([]byte(responseJSON), &llmResp); err != nil {
		return nil, fmt.Errorf("parse LLM response as JSON: %w\nResponse was: %s", err, responseJSON)
	}

	// Convert to StatementData
	statementDate, err := parseDate(llmResp.StatementDate)
	if err != nil {
		return nil, fmt.Errorf("parse statement date '%s': %w", llmResp.StatementDate, err)
	}

	data := &StatementData{
		AccountName:   llmResp.AccountName,
		AccountLast4:  llmResp.AccountLast4,
		StatementDate: statementDate,
		Transactions:  make([]*db.Transaction, 0, len(llmResp.Transactions)),
	}

	// Convert LLM transactions to db.Transaction
	for i, llmTx := range llmResp.Transactions {
		txDate, err := parseDate(llmTx.TransactionDate)
		if err != nil {
			return nil, fmt.Errorf("transaction %d: parse date '%s': %w", i, llmTx.TransactionDate, err)
		}

		tx := &db.Transaction{
			AccountName:     llmResp.AccountName,
			AccountLast4:    llmResp.AccountLast4,
			TransactionDate: txDate,
			Description:     llmTx.Description,
			Amount:          llmTx.Amount,
			TransactionType: llmTx.TransactionType,
			Balance:         llmTx.Balance,
			StatementDate:   statementDate,
			SourceFile:      filepath.Base(sourceFile),
		}

		// Parse post date if provided
		if llmTx.PostDate != nil && *llmTx.PostDate != "" {
			postDate, err := parseDate(*llmTx.PostDate)
			if err == nil {
				tx.PostDate = &postDate
			}
		}

		data.Transactions = append(data.Transactions, tx)
	}

	return data, nil
}

// parseDate attempts to parse a date string in common formats
func parseDate(dateStr string) (time.Time, error) {
	// Common date formats in bank statements
	formats := []string{
		"2006-01-02",                // YYYY-MM-DD (preferred)
		"01/02/2006",                // MM/DD/YYYY
		"1/2/2006",                  // M/D/YYYY
		"01/02/06",                  // MM/DD/YY
		"Jan 02, 2006",              // Mon DD, YYYY
		"January 02, 2006",          // Month DD, YYYY
		"2006-01-02T15:04:05Z07:00", // ISO 8601 with timezone
		"06-01-02",                  // YY-MM-DD (defensive fallback for LLM output like "25-10-01")
		"06/01/02",                  // YY/MM/DD (defensive fallback)
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ValidateStatementData validates parsed statement data
func ValidateStatementData(data *StatementData) error {
	if data.AccountName == "" {
		return fmt.Errorf("account name is required")
	}

	if data.AccountLast4 == "" {
		return fmt.Errorf("account last 4 digits is required")
	}

	if len(data.AccountLast4) != 4 {
		return fmt.Errorf("account last 4 must be exactly 4 digits, got: %s", data.AccountLast4)
	}

	if data.StatementDate.IsZero() {
		return fmt.Errorf("statement date is required")
	}

	if len(data.Transactions) == 0 {
		return fmt.Errorf("no transactions found in statement")
	}

	// Validate each transaction
	for i, tx := range data.Transactions {
		if tx.Description == "" {
			return fmt.Errorf("transaction %d: description is required", i)
		}
		if tx.TransactionDate.IsZero() {
			return fmt.Errorf("transaction %d: transaction date is required", i)
		}
		if tx.TransactionType != "debit" && tx.TransactionType != "credit" {
			return fmt.Errorf("transaction %d: type must be 'debit' or 'credit', got: %s", i, tx.TransactionType)
		}
	}

	return nil
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
