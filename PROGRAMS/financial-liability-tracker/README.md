# Financial Liability Tracker

A command-line tool for tracking and managing financial liabilities (credit cards, loans, mortgages, etc.) with balance history tracking. Uses SQLite for data storage and outputs JSON for easy integration with other tools.

## Features

- **Multiple Liability Types**: credit-card, auto-loan, mortgage, student-loan, personal-loan, medical-debt
- **Full Balance History**: Track every balance update with timestamps
- **CRUD Operations**: Complete create, read, update, delete functionality
- **JSON Output**: All commands output JSON for easy parsing
- **SQLite Backend**: Lightweight, local database storage
- **Exit Codes**: Proper exit codes for scripting and automation

## Quick Start

### 1. Prerequisites

- Go 1.19+ (for building)
- SQLite support (included with most systems)

### 2. Setup

```bash
# Clone/navigate to the directory
cd financial-liability-tracker

# Build the executable
go build -o financial-liability-tracker ./cmd/tracker

# Optional: Copy and configure environment variables if you want a custom database location
cp .env.example .env
# Edit .env to set DB_PATH if desired (defaults to ~/.local/share/financial-liability-tracker/liabilities.db)
```

### 3. Basic Usage

```bash
# Add a credit card
./financial-liability-tracker add --type credit-card --name "Chase Sapphire" \\
  --balance 2500 --limit 10000 --rate 18.99 --min-payment 50

# List all liabilities
./financial-liability-tracker list

# Update balance
./financial-liability-tracker update chase-sapphire --balance 2100

# Get total balance
./financial-liability-tracker total
```

## Commands

### add - Add a new liability

```bash
financial-liability-tracker add [flags]

Required flags:
  --name string       Unique liability name/identifier
  --type string       Liability type (see types below)
  --balance float     Current balance

Optional flags:
  --original float    Original amount borrowed
  --limit float       Credit limit (for credit cards)
  --rate float        Interest rate percentage
  --min-payment float Minimum payment amount
  --creditor string   Creditor/lender name
  --last4 string      Last 4 digits of account number
  --opened string     Date opened (YYYY-MM-DD)
  --notes string      Additional notes

Examples:
  # Credit card
  financial-liability-tracker add --type credit-card --name "Chase Sapphire" \\
    --balance 2500 --limit 10000 --rate 18.99 --min-payment 50 \\
    --creditor "Chase Bank" --last4 "1234"

  # Auto loan
  financial-liability-tracker add --type auto-loan --name "Honda Civic Loan" \\
    --balance 15000 --original 25000 --rate 4.5 --min-payment 350 \\
    --creditor "Honda Financial" --opened 2022-06-15

  # Mortgage
  financial-liability-tracker add --type mortgage --name "Main Residence" \\
    --balance 285000 --original 300000 --rate 3.25 --min-payment 1850 \\
    --creditor "Wells Fargo" --opened 2020-01-15
```

### update - Update liability balance

```bash
financial-liability-tracker update <name> [flags]

Arguments:
  name                Liability name

Flags:
  --balance float     New balance (required)
  --notes string      Update notes (optional)

Examples:
  # Simple balance update
  financial-liability-tracker update chase-sapphire --balance 2100

  # With notes
  financial-liability-tracker update chase-sapphire --balance 2100 \\
    --notes "Paid down after bonus"
```

### delete - Delete a liability

```bash
financial-liability-tracker delete <name>

Arguments:
  name                Liability name

Example:
  financial-liability-tracker delete chase-sapphire
```

### list - List all liabilities

```bash
financial-liability-tracker list [flags]

Flags:
  --type string       Filter by liability type (optional)

Examples:
  # List all
  financial-liability-tracker list

  # List only credit cards
  financial-liability-tracker list --type credit-card

  # List only loans
  financial-liability-tracker list --type auto-loan
```

### get - Get liability details

```bash
financial-liability-tracker get <name> [flags]

Arguments:
  name                Liability name

Flags:
  --history           Include balance history (optional)

Examples:
  # Basic details
  financial-liability-tracker get chase-sapphire

  # With balance history
  financial-liability-tracker get chase-sapphire --history
```

### total - Get total of all balances

```bash
financial-liability-tracker total

Example:
  financial-liability-tracker total
```

## Liability Types

- `credit-card` - Credit cards
- `auto-loan` - Auto/vehicle loans
- `mortgage` - Home mortgages
- `student-loan` - Student loans
- `personal-loan` - Personal loans
- `medical-debt` - Medical debt

## Output Format

All commands output JSON to stdout. Errors are output as JSON to stderr.

### Success Response Example

```json
{
  "success": true,
  "liability": {
    "id": 1,
    "name": "chase-sapphire",
    "liability_type": "credit-card",
    "current_balance": 2500,
    "credit_limit": 10000,
    "interest_rate": 18.99,
    "minimum_payment": 50,
    "creditor_name": "Chase Bank",
    "account_last4": "1234",
    "created_at": "2024-11-18T10:00:00Z",
    "updated_at": "2024-11-18T10:00:00Z"
  }
}
```

### List Response Example

```json
{
  "liabilities": [
    {
      "id": 1,
      "name": "chase-sapphire",
      "liability_type": "credit-card",
      "current_balance": 2500,
      "credit_limit": 10000,
      "interest_rate": 18.99,
      "minimum_payment": 50
    },
    {
      "id": 2,
      "name": "honda-civic-loan",
      "liability_type": "auto-loan",
      "current_balance": 15000,
      "original_amount": 25000,
      "interest_rate": 4.5,
      "minimum_payment": 350
    }
  ],
  "count": 2
}
```

### Balance History Example

```json
{
  "liability": {
    "id": 1,
    "name": "chase-sapphire",
    "current_balance": 2100
  },
  "balance_history": [
    {
      "id": 2,
      "liability_id": 1,
      "balance": 2100,
      "recorded_at": "2024-11-18T15:30:00Z",
      "notes": "Paid down after bonus"
    },
    {
      "id": 1,
      "liability_id": 1,
      "balance": 2500,
      "recorded_at": "2024-11-18T10:00:00Z",
      "notes": "Initial balance"
    }
  ]
}
```

### Total Balance Example

```json
{
  "total_balance": 287500.00
}
```

### Error Response Example

```json
{
  "error": "liability not found: chase-sapphire"
}
```

## Configuration

Configuration is done via environment variables. The database location can be customized:

```bash
# Database path (optional)
# If not specified, defaults to ~/.local/share/financial-liability-tracker/liabilities.db
DB_PATH=/path/to/your/liabilities.db
```

The database and schema are created automatically on first run.

## Database Schema

### liabilities table

| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER | Primary key (autoincrement) |
| name | TEXT | Unique liability identifier |
| liability_type | TEXT | Type of liability |
| current_balance | REAL | Current balance |
| original_amount | REAL | Original amount (nullable) |
| credit_limit | REAL | Credit limit for cards (nullable) |
| interest_rate | REAL | Interest rate % (nullable) |
| minimum_payment | REAL | Minimum payment (nullable) |
| creditor_name | TEXT | Creditor name (nullable) |
| account_last4 | TEXT | Last 4 of account (nullable) |
| opened_date | TEXT | Date opened ISO8601 (nullable) |
| notes | TEXT | Additional notes (nullable) |
| created_at | TEXT | Creation timestamp ISO8601 |
| updated_at | TEXT | Last update timestamp ISO8601 |

### liability_balance_history table

| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER | Primary key (autoincrement) |
| liability_id | INTEGER | Foreign key to liabilities |
| balance | REAL | Balance at this point |
| recorded_at | TEXT | When recorded ISO8601 |
| notes | TEXT | Update notes (nullable) |

## Exit Codes

- `0` - Success
- `1` - Invalid arguments
- `2` - Database error
- `3` - Not found

## Integration Examples

### Shell Script

```bash
#!/bin/bash
# Get total liability balance

RESULT=$(./financial-liability-tracker total)
TOTAL=$(echo "$RESULT" | jq -r '.total_balance')

echo "Total liabilities: \$$TOTAL"
```

### Python

```python
import subprocess
import json

# List all credit cards
result = subprocess.run(
    ['./financial-liability-tracker', 'list', '--type', 'credit-card'],
    capture_output=True,
    text=True
)

data = json.loads(result.stdout)
for liability in data['liabilities']:
    print(f"{liability['name']}: ${liability['current_balance']}")
```

### Agent Gateway Integration

```go
// In your agent's executable wrapper
cmd := exec.Command("financial-liability-tracker", "total")
output, err := cmd.Output()
if err != nil {
    return fmt.Errorf("failed to get total: %w", err)
}

var result map[string]interface{}
json.Unmarshal(output, &result)
total := result["total_balance"].(float64)
```

## Development

### Building

```bash
go build -o financial-liability-tracker ./cmd/tracker
```

### Running Tests

```bash
go test ./...
```

### Database Migrations

The `schema.sql` file contains the complete schema and is automatically applied when the database is created. You can also manually initialize a database:

```bash
sqlite3 /path/to/liabilities.db < schema.sql
```

## Troubleshooting

### "database initialization failed"

- Check that the database directory is writable
- Verify DB_PATH (if set) points to a valid location
- Database is created automatically on first run

### "liability not found"

- Use exact name as stored (names are case-sensitive)
- Use `list` command to see all liabilities

### "invalid liability type"

- Must be one of: credit-card, auto-loan, mortgage, student-loan, personal-loan, medical-debt
- Use hyphens, not spaces

## Privacy & Security

- **Local Database**: All data stored in local SQLite file
- **No External Calls**: Executable never connects to external services
- **File-based Storage**: Database is a single file for easy backup and portability
- **No Network Access**: SQLite runs entirely locally with no network exposure

## License

MIT License - See LICENSE file for details

## Future Enhancements

- Import from CSV
- Payment schedule calculator
- Interest calculation
- Payoff projection
- Export to various formats
- Liability consolidation analysis
