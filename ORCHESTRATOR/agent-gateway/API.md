# Agent Gateway API Documentation

Version: 1.0.0

## Authentication

All endpoints (except `/api/health`) require API key authentication via the `X-API-Key` header.

```bash
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/endpoint
```

## Base URL

```
http://localhost:8080/api
```

## Response Format

All responses follow a standard format:

**Success:**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error:**
```json
{
  "success": false,
  "error": "Error message"
}
```

---

## Financial Statement Processor

Process PDF bank statements and query transactions.

### Process PDF Statement

Upload and process a PDF bank statement.

**Endpoint:** `POST /api/financial-statement/process`

**Content-Type:** `multipart/form-data`

**Form Fields:**
- `file` (required): PDF file to process
- `account_type` (optional): Account type (checking/savings/credit)
- `force` (optional): Set to "true" to reprocess existing statements

**Example:**
```bash
curl -X POST \
  -H "X-API-Key: your-api-key" \
  -F "file=@statement.pdf" \
  -F "account_type=checking" \
  http://localhost:8080/api/financial-statement/process
```

**Response:**
```json
{
  "success": true,
  "data": {
    "transactions_processed": 45,
    "errors": [],
    "file_name": "statement.pdf"
  }
}
```

### Query Transactions

Query transactions with filters.

**Endpoint:** `GET /api/financial-statement/transactions`

**Query Parameters:**
- `start_date` (optional): Start date (YYYY-MM-DD)
- `end_date` (optional): End date (YYYY-MM-DD)
- `type` (optional): Transaction type (debit/credit)
- `merchant` (optional): Filter by merchant name

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8080/api/financial-statement/transactions?start_date=2024-01-01&type=debit"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "transactions": [
      {
        "id": 1,
        "date": "2024-01-15",
        "description": "Amazon Purchase",
        "amount": 45.99,
        "transaction_type": "debit",
        "category": "shopping",
        "merchant": "Amazon",
        "created_at": "2024-01-16T10:30:00Z",
        "updated_at": "2024-01-16T10:30:00Z"
      }
    ],
    "count": 1
  }
}
```

### Transaction Summary

Get aggregated transaction statistics.

**Endpoint:** `GET /api/financial-statement/summary`

**Query Parameters:**
- `start_date` (optional): Start date (defaults to 30 days ago)
- `end_date` (optional): End date (defaults to today)

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8080/api/financial-statement/summary?start_date=2024-01-01&end_date=2024-12-31"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "total_amount": 15234.56,
    "total_count": 127,
    "count_by_type": {
      "debit": 98,
      "credit": 29
    },
    "amount_by_type": {
      "debit": 12890.34,
      "credit": 2344.22
    },
    "start_date": "2024-01-01",
    "end_date": "2024-12-31"
  }
}
```

---

## Financial Asset Tracker

Manage financial assets with value history tracking.

### Add Asset

Add a new asset to track.

**Endpoint:** `POST /api/financial-asset`

**Request Body:**
```json
{
  "name": "tesla-stock",
  "category": "stocks",
  "current_value": 25000.00,
  "purchase_price": 20000.00,
  "purchase_date": "2023-06-15",
  "notes": "100 shares"
}
```

**Example:**
```bash
curl -X POST \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"name":"tesla-stock","category":"stocks","current_value":25000}' \
  http://localhost:8080/api/financial-asset
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "tesla-stock",
    "category": "stocks",
    "current_value": 25000.00,
    "purchase_price": 20000.00,
    "purchase_date": "2023-06-15",
    "notes": "100 shares",
    "is_removed": false,
    "created_at": "2024-11-19T10:30:00Z",
    "updated_at": "2024-11-19T10:30:00Z",
    "last_updated": "2024-11-19T10:30:00Z"
  }
}
```

### List Assets

Get all assets with optional filters.

**Endpoint:** `GET /api/financial-asset`

**Query Parameters:**
- `include_removed` (optional): Include soft-deleted assets (true/false, default: false)
- `category` (optional): Filter by category

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8080/api/financial-asset?category=stocks"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "assets": [ ... ],
    "count": 5
  }
}
```

### Get Asset

Get details for a specific asset.

**Endpoint:** `GET /api/financial-asset/{name}`

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/financial-asset/tesla-stock
```

### Update Asset Value

Update an asset's current value.

**Endpoint:** `PUT /api/financial-asset/{name}`

**Request Body:**
```json
{
  "current_value": 27500.00,
  "notes": "Price increase"
}
```

**Example:**
```bash
curl -X PUT \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"current_value":27500,"notes":"Price increase"}' \
  http://localhost:8080/api/financial-asset/tesla-stock
```

### Remove Asset

Soft-delete an asset (can be restored).

**Endpoint:** `DELETE /api/financial-asset/{name}`

**Example:**
```bash
curl -X DELETE \
  -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/financial-asset/tesla-stock
```

### Restore Asset

Restore a previously removed asset.

**Endpoint:** `POST /api/financial-asset/{name}/restore`

**Example:**
```bash
curl -X POST \
  -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/financial-asset/tesla-stock/restore
```

### Asset Summary

Get portfolio summary with category breakdowns.

**Endpoint:** `GET /api/financial-asset/summary`

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/financial-asset/summary
```

**Response:**
```json
{
  "success": true,
  "data": {
    "total_value": 125000.00,
    "total_count": 8,
    "count_by_category": {
      "stocks": 5,
      "crypto": 2,
      "real-estate": 1
    },
    "value_by_category": {
      "stocks": 75000.00,
      "crypto": 15000.00,
      "real-estate": 35000.00
    }
  }
}
```

---

## Financial Liability Tracker

Manage financial liabilities (credit cards, loans, mortgages).

### Add Liability

Add a new liability.

**Endpoint:** `POST /api/financial-liability`

**Request Body:**
```json
{
  "name": "chase-sapphire",
  "type": "credit-card",
  "balance": 2500.00,
  "original_amount": 5000.00,
  "credit_limit": 10000.00,
  "interest_rate": 19.99,
  "minimum_payment": 75.00,
  "creditor_name": "Chase Bank",
  "account_last4": "1234",
  "opened_date": "2022-01-15",
  "notes": "Rewards card"
}
```

**Valid liability types:**
- `credit-card`
- `auto-loan`
- `mortgage`
- `student-loan`
- `personal-loan`
- `medical-debt`

**Example:**
```bash
curl -X POST \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"name":"chase-sapphire","type":"credit-card","balance":2500}' \
  http://localhost:8080/api/financial-liability
```

### List Liabilities

Get all liabilities with optional type filter.

**Endpoint:** `GET /api/financial-liability`

**Query Parameters:**
- `type` (optional): Filter by liability type

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8080/api/financial-liability?type=credit-card"
```

### Get Liability

Get details for a specific liability.

**Endpoint:** `GET /api/financial-liability/{name}`

**Query Parameters:**
- `include_history` (optional): Include balance history (true/false)

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8080/api/financial-liability/chase-sapphire?include_history=true"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "liability": { ... },
    "balance_history": [
      {
        "id": 1,
        "liability_id": 1,
        "balance": 2500.00,
        "recorded_at": "2024-11-19T10:30:00Z",
        "notes": "Initial balance"
      }
    ]
  }
}
```

### Update Liability

Update a liability's balance.

**Endpoint:** `PUT /api/financial-liability/{name}`

**Request Body:**
```json
{
  "balance": 2200.00,
  "notes": "Made payment"
}
```

**Example:**
```bash
curl -X PUT \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"balance":2200,"notes":"Made payment"}' \
  http://localhost:8080/api/financial-liability/chase-sapphire
```

### Delete Liability

Permanently delete a liability.

**Endpoint:** `DELETE /api/financial-liability/{name}`

**Example:**
```bash
curl -X DELETE \
  -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/financial-liability/chase-sapphire
```

### Total Liabilities

Get total balance across all liabilities.

**Endpoint:** `GET /api/financial-liability/total`

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/financial-liability/total
```

**Response:**
```json
{
  "success": true,
  "data": {
    "total_balance": 45000.00
  }
}
```

### Liability Summary

Get summary grouped by liability type.

**Endpoint:** `GET /api/financial-liability/summary`

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/financial-liability/summary
```

**Response:**
```json
{
  "success": true,
  "data": {
    "total_balance": 45000.00,
    "total_count": 5,
    "count_by_type": {
      "credit-card": 2,
      "auto-loan": 1,
      "mortgage": 1,
      "student-loan": 1
    },
    "balance_by_type": {
      "credit-card": 5000.00,
      "auto-loan": 15000.00,
      "mortgage": 200000.00,
      "student-loan": 25000.00
    }
  }
}
```

---

## Financial Overview

Get comprehensive financial snapshots.

### Net Worth

Calculate net worth (assets - liabilities).

**Endpoint:** `GET /api/financial/net-worth`

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/financial/net-worth
```

**Response:**
```json
{
  "success": true,
  "data": {
    "net_worth": 80000.00,
    "total_assets": 125000.00,
    "total_liabilities": 45000.00,
    "timestamp": "2024-11-19T10:30:00Z"
  }
}
```

### Financial Summary

Get complete financial overview.

**Endpoint:** `GET /api/financial/summary`

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/financial/summary
```

**Response:**
```json
{
  "success": true,
  "data": {
    "total_assets": 125000.00,
    "total_liabilities": 45000.00,
    "net_worth": 80000.00,
    "asset_count": 8,
    "liability_count": 5,
    "timestamp": "2024-11-19T10:30:00Z"
  }
}
```

---

## Meta Endpoints

### Health Check

Check API and database health (no authentication required).

**Endpoint:** `GET /api/health`

**Example:**
```bash
curl http://localhost:8080/api/health
```

### Statistics

Get system statistics (requires authentication).

**Endpoint:** `GET /api/stats`

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/stats
```

---

## Error Codes

- `200` - Success
- `400` - Bad Request (invalid input, validation error)
- `401` - Unauthorized (missing or invalid API key)
- `404` - Not Found (resource doesn't exist)
- `500` - Internal Server Error

## Rate Limiting

Currently no rate limiting is implemented.

## CORS

CORS is enabled for all origins with the following methods:
- GET, POST, PUT, DELETE, PATCH, OPTIONS

## Notes

- All dates should be in `YYYY-MM-DD` format
- All timestamps are in RFC3339 format
- File uploads are limited to 10MB
- Only PDF files are accepted for statement processing
