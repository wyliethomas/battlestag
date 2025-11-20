package handlers

import (
	"net/http"
	"time"

	"agent-gateway/db"
	"agent-gateway/executor"
	"agent-gateway/models"
)

// FinancialStatementHandler handles financial statement processor endpoints
type FinancialStatementHandler struct {
	executor  *executor.Executor
	dbManager *db.Manager
}

// NewFinancialStatementHandler creates a new financial statement handler
func NewFinancialStatementHandler(exec *executor.Executor, dbManager *db.Manager) *FinancialStatementHandler {
	return &FinancialStatementHandler{
		executor:  exec,
		dbManager: dbManager,
	}
}

// ProcessPDF processes an uploaded PDF statement
// POST /api/financial-statement/process
// Content-Type: multipart/form-data
// Form fields:
//   - file: PDF file (required)
//   - account_type: checking/savings/credit (optional)
//   - force: true/false (optional, reprocess existing)
func (h *FinancialStatementHandler) ProcessPDF(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		models.WriteError(w, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}

	// Get the file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, "file is required: "+err.Error())
		return
	}
	defer file.Close()

	// Save uploaded file to temp location
	tempPath, err := models.SaveUploadedFile(file, header)
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer models.CleanupTempFile(tempPath)

	// Get optional parameters
	accountType := r.FormValue("account_type")
	force := r.FormValue("force") == "true"

	// Process the PDF
	result, err := h.executor.ProcessPDF(tempPath, accountType, force)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, result)
}

// QueryTransactions queries transactions with filters
// GET /api/financial-statement/transactions?start_date=2024-01-01&end_date=2024-12-31&type=debit&merchant=amazon
func (h *FinancialStatementHandler) QueryTransactions(w http.ResponseWriter, r *http.Request) {
	startDate := models.GetQueryParam(r, "start_date", "")
	endDate := models.GetQueryParam(r, "end_date", "")
	txnType := models.GetQueryParam(r, "type", "")
	merchant := models.GetQueryParam(r, "merchant", "")

	// Validate dates if provided
	if startDate != "" {
		if err := models.ValidateDate(startDate); err != nil {
			models.WriteError(w, http.StatusBadRequest, "invalid start_date: "+err.Error())
			return
		}
	}
	if endDate != "" {
		if err := models.ValidateDate(endDate); err != nil {
			models.WriteError(w, http.StatusBadRequest, "invalid end_date: "+err.Error())
			return
		}
	}

	// Query transactions via executor
	transactions, err := h.executor.QueryTransactions(startDate, endDate, txnType, merchant)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, map[string]interface{}{
		"transactions": transactions,
		"count":        len(transactions),
	})
}

// GetSummary returns transaction summary statistics
// GET /api/financial-statement/summary?start_date=2024-01-01&end_date=2024-12-31&type=debit
func (h *FinancialStatementHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	startDate := models.GetQueryParam(r, "start_date", "")
	endDate := models.GetQueryParam(r, "end_date", "")

	// Set defaults if not provided (last 30 days)
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Validate dates
	if err := models.ValidateDate(startDate); err != nil {
		models.WriteError(w, http.StatusBadRequest, "invalid start_date: "+err.Error())
		return
	}
	if err := models.ValidateDate(endDate); err != nil {
		models.WriteError(w, http.StatusBadRequest, "invalid end_date: "+err.Error())
		return
	}

	// Get summary from database
	summary, err := h.dbManager.GetTransactionSummary(startDate, endDate)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, summary)
}
