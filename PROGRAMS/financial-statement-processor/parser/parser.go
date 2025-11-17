package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"financial-statement-processor/db"
)

// StatementData represents parsed statement information
type StatementData struct {
	AccountName   string
	AccountLast4  string
	StatementDate time.Time
	Transactions  []*db.Transaction
}

// ParseFile parses a bank statement file (PDF or image)
func ParseFile(filePath string) (*StatementData, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return parsePDF(filePath)
	case ".jpg", ".jpeg", ".png", ".tiff", ".tif":
		return parseImage(filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s (supported: PDF, JPG, PNG, TIFF)", ext)
	}
}

// parsePDF extracts transaction data from a PDF bank statement
func parsePDF(filePath string) (*StatementData, error) {
	// TODO: Implement PDF parsing
	// Recommended libraries:
	// - github.com/ledongthuc/pdf for simple text extraction
	// - github.com/pdfcpu/pdfcpu for more advanced PDF operations
	// - github.com/unidoc/unipdf for commercial-grade PDF parsing

	// Example structure:
	// 1. Extract text from PDF
	// 2. Identify account information (name, last 4 digits)
	// 3. Find statement date
	// 4. Parse transaction lines using regex patterns
	// 5. Extract: date, description, amount, type (debit/credit), balance

	// For now, return a placeholder error
	return nil, fmt.Errorf("PDF parsing not yet implemented - please add PDF library and parsing logic")

	// Example implementation outline:
	/*
		// Open PDF file
		f, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("open PDF: %w", err)
		}
		defer f.Close()

		// Read PDF content
		pdfReader, err := pdf.NewReader(f, fileInfo.Size())
		if err != nil {
			return nil, fmt.Errorf("read PDF: %w", err)
		}

		// Extract text from all pages
		var fullText strings.Builder
		for pageNum := 1; pageNum <= pdfReader.NumPage(); pageNum++ {
			page := pdfReader.Page(pageNum)
			text, err := page.GetPlainText(nil)
			if err != nil {
				return nil, fmt.Errorf("extract text from page %d: %w", pageNum, err)
			}
			fullText.WriteString(text)
		}

		// Parse the extracted text
		return parseStatementText(fullText.String(), filePath)
	*/
}

// parseImage extracts transaction data from an image using OCR
func parseImage(filePath string) (*StatementData, error) {
	// TODO: Implement OCR parsing
	// Recommended approach:
	// - Use gosseract (Go bindings for Tesseract OCR)
	// - Install Tesseract: apt-get install tesseract-ocr
	// - Library: github.com/otiai10/gosseract/v2

	// For now, return a placeholder error
	return nil, fmt.Errorf("OCR parsing not yet implemented - please install Tesseract and add gosseract library")

	// Example implementation outline:
	/*
		client := gosseract.NewClient()
		defer client.Close()

		// Set image file
		client.SetImage(filePath)

		// Get text from image
		text, err := client.Text()
		if err != nil {
			return nil, fmt.Errorf("OCR failed: %w", err)
		}

		// Parse the extracted text
		return parseStatementText(text, filePath)
	*/
}

// parseStatementText parses extracted text to identify transactions
func parseStatementText(text, sourceFile string) (*StatementData, error) {
	// TODO: Implement text parsing logic
	// This is highly bank-specific and will need customization

	// Common patterns to extract:
	// 1. Account name: often appears as "Account Holder: JOHN DOE" or similar
	// 2. Account number: last 4 digits, often "Account ending in 1234"
	// 3. Statement date: "Statement Date: 10/31/2024" or similar
	// 4. Transaction lines: typically formatted as:
	//    DATE    DESCRIPTION                      AMOUNT    BALANCE
	//    10/15   PURCHASE AT STORE NAME          -50.00    1,234.56

	// Example regex patterns (adjust for your bank's format):
	// accountPattern := regexp.MustCompile(`Account ending in (\d{4})`)
	// statementDatePattern := regexp.MustCompile(`Statement Date:\s*(\d{1,2}/\d{1,2}/\d{4})`)
	// transactionPattern := regexp.MustCompile(`(\d{1,2}/\d{1,2})\s+([A-Z0-9\s\-\.]+?)\s+([\-\$\d,\.]+)\s+([\$\d,\.]+)`)

	return nil, fmt.Errorf("text parsing not implemented - add bank-specific parsing logic")
}

// Example helper function for parsing dates
func parseDate(dateStr string) (time.Time, error) {
	// Common date formats in bank statements
	formats := []string{
		"01/02/2006",
		"1/2/2006",
		"01/02/06",
		"2006-01-02",
		"Jan 02, 2006",
		"January 02, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// Example helper function for parsing amounts
func parseAmount(amountStr string) (float64, error) {
	// Remove currency symbols and commas
	cleaned := strings.ReplaceAll(amountStr, "$", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.TrimSpace(cleaned)

	var amount float64
	_, err := fmt.Sscanf(cleaned, "%f", &amount)
	if err != nil {
		return 0, fmt.Errorf("parse amount: %w", err)
	}

	return amount, nil
}

// DetermineTransactionType determines if a transaction is debit or credit
func DetermineTransactionType(amount float64, amountStr string) string {
	// Check if amount string contains negative sign or parentheses
	if strings.Contains(amountStr, "-") || strings.Contains(amountStr, "(") {
		return "debit"
	}

	// Some banks use positive for credits, negative for debits
	if amount < 0 {
		return "debit"
	}

	return "credit"
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
		return fmt.Errorf("account last 4 must be exactly 4 digits")
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
			return fmt.Errorf("transaction %d: type must be 'debit' or 'credit'", i)
		}
	}

	return nil
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
