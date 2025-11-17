# Financial Statement Processor

A Go-based system for parsing bank statements (PDF and images) and storing transactions in PostgreSQL. Designed to work with the financial-document-watcher for automated statement processing.

## Features

- **Dual executables**: Processor for parsing statements, query tool for retrieving data
- **Multi-format support**: PDF and image files (via OCR)
- **PostgreSQL storage**: Robust relational database with proper indexing
- **Duplicate prevention**: Automatic detection and skipping of duplicate transactions
- **Proper exit codes**: Integration-friendly (0=success, 1=parse error, 2=db error)
- **JSON output**: Query results in structured JSON format
- **Comprehensive logging**: All operations logged to stdout for monitoring

## Components

### 1. financial-statement-processor
Parses bank statement files and extracts transactions.

**Extracts:**
- Account name and last 4 digits
- Transaction date and post date
- Description and amount
- Transaction type (debit/credit)
- Running balance
- Statement date

**Exit Codes:**
- `0` - Success
- `1` - Parse error (invalid file, parsing failed)
- `2` - Database error (connection or insertion failed)
- `3` - Configuration error (missing env vars)

### 2. financial-statement-query
Query and retrieve transaction data in JSON format.

**Features:**
- Date range filtering
- Account filtering
- Pretty-printed JSON
- Account summary view

## Installation

### Prerequisites

- Go 1.16 or higher
- PostgreSQL 12 or higher
- psql command-line tool (for schema setup)

### Quick Installation

```bash
cd financial-statement-processor
./install.sh
```

The installer will:
1. Prompt for installation paths
2. Collect database credentials
3. Build both executables
4. Install binaries to `/usr/local/bin`
5. Create configuration file with database settings
6. Optionally run database schema setup
7. Create wrapper scripts with environment variables

### Manual Installation

1. Initialize Go module:
```bash
go mod init financial-statement-processor
go mod tidy
```

2. Build executables:
```bash
go build -o financial-statement-processor ./cmd/processor
go build -o financial-statement-query ./cmd/query
```

3. Set up database:
```bash
createdb financial_data
psql -d financial_data -f schema.sql
```

4. Configure environment:
```bash
cp .env.example .env
# Edit .env with your database credentials
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL server host |
| `DB_PORT` | `5432` | PostgreSQL server port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | *(required)* | Database password |
| `DB_NAME` | `financial_data` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode (disable, require, verify-full) |

### Database Schema

The schema includes:
- **transactions** table: Stores all transaction records
- **processing_log** table: Tracks statement processing history
- **account_summary** view: Quick account statistics
- **monthly_summary** view: Monthly aggregates by account

See `schema.sql` for complete schema definition.

## Usage

### Processing Statements

Process a single statement file:

```bash
financial-statement-processor-run /path/to/statement.pdf
```

With custom database:

```bash
DB_HOST=192.168.1.100 DB_PASSWORD=secret financial-statement-processor /path/to/statement.pdf
```

Verbose output:

```bash
financial-statement-processor-run --verbose /path/to/statement.pdf
```

### Querying Transactions

Query by date range:

```bash
financial-statement-query-run --start-date 2024-10-01 --end-date 2024-10-31
```

Filter by account:

```bash
financial-statement-query-run \
  --start-date 2024-10-01 \
  --end-date 2024-10-31 \
  --account "Checking" \
  --pretty
```

Get account summary:

```bash
financial-statement-query-run --summary --pretty
```

### Example Output

```json
{
  "start_date": "2024-10-01",
  "end_date": "2024-10-31",
  "total_records": 45,
  "accounts": {
    "Checking Account (...1234)": [
      {
        "id": 1,
        "account_name": "Checking Account",
        "account_last4": "1234",
        "transaction_date": "2024-10-15",
        "post_date": "2024-10-16",
        "description": "PURCHASE AT GROCERY STORE",
        "amount": -52.34,
        "transaction_type": "debit",
        "balance": 1247.66,
        "statement_date": "2024-10-31",
        "source_file": "statement_oct_2024.pdf"
      }
    ]
  }
}
```

## Integration with Financial Document Watcher

Update your `watches.json` in the financial-document-watcher:

```json
{
  "watch_id": "bank_statements",
  "watch_path": "/home/user/Documents/financial/incoming/bank",
  "file_pattern": "*.pdf",
  "executable_path": "/usr/local/bin/financial-statement-processor-run",
  "processed_path": "/home/user/Documents/financial/processed/bank"
}
```

The watcher will:
1. Detect new PDF files
2. Execute the processor with the file path
3. Check exit code (0 = success)
4. Move file to processed folder on success
5. Leave file in place on failure for retry

## Implementing Parser Logic

The current implementation includes placeholder parsing functions. You need to add bank-specific parsing logic.

### PDF Parsing

Edit `parser/parser.go` and implement `parsePDF()`:

**Recommended libraries:**
- `github.com/ledongthuc/pdf` - Simple text extraction
- `github.com/pdfcpu/pdfcpu` - Advanced PDF operations
- `github.com/unidoc/unipdf` - Commercial-grade PDF parsing

**Example:**

```go
import "github.com/ledongthuc/pdf"

func parsePDF(filePath string) (*StatementData, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    // Get file info for size
    fileInfo, _ := f.Stat()

    // Read PDF
    pdfReader, err := pdf.NewReader(f, fileInfo.Size())
    if err != nil {
        return nil, err
    }

    // Extract text from all pages
    var fullText strings.Builder
    for pageNum := 1; pageNum <= pdfReader.NumPage(); pageNum++ {
        page := pdfReader.Page(pageNum)
        text, _ := page.GetPlainText(nil)
        fullText.WriteString(text)
    }

    // Parse the text
    return parseStatementText(fullText.String(), filePath)
}
```

### OCR for Images

Edit `parser/parser.go` and implement `parseImage()`:

**Requirements:**
1. Install Tesseract: `apt-get install tesseract-ocr`
2. Add library: `go get github.com/otiai10/gosseract/v2`

**Example:**

```go
import "github.com/otiai10/gosseract/v2"

func parseImage(filePath string) (*StatementData, error) {
    client := gosseract.NewClient()
    defer client.Close()

    client.SetImage(filePath)
    text, err := client.Text()
    if err != nil {
        return nil, err
    }

    return parseStatementText(text, filePath)
}
```

### Text Parsing

Implement `parseStatementText()` with bank-specific regex patterns:

```go
func parseStatementText(text, sourceFile string) (*StatementData, error) {
    // Extract account info
    accountPattern := regexp.MustCompile(`Account ending in (\d{4})`)
    accountMatch := accountPattern.FindStringSubmatch(text)

    // Extract statement date
    datePattern := regexp.MustCompile(`Statement Date:\s*(\d{1,2}/\d{1,2}/\d{4})`)
    dateMatch := datePattern.FindStringSubmatch(text)

    // Extract transactions
    // Pattern: DATE DESCRIPTION AMOUNT BALANCE
    txPattern := regexp.MustCompile(`(\d{1,2}/\d{1,2})\s+(.+?)\s+([\-\$\d,\.]+)\s+([\$\d,\.]+)`)
    txMatches := txPattern.FindAllStringSubmatch(text, -1)

    // Parse and build StatementData...
}
```

## Database Management

### View Processing Log

```sql
SELECT * FROM processing_log ORDER BY processed_at DESC LIMIT 10;
```

### Find Duplicates

```sql
SELECT account_last4, transaction_date, description, amount, COUNT(*)
FROM transactions
GROUP BY account_last4, transaction_date, description, amount
HAVING COUNT(*) > 1;
```

### Account Summary

```sql
SELECT * FROM account_summary;
```

### Monthly Totals

```sql
SELECT * FROM monthly_summary WHERE month >= '2024-01-01';
```

### Delete Test Data

```sql
DELETE FROM transactions WHERE source_file LIKE 'test%';
DELETE FROM processing_log WHERE source_file LIKE 'test%';
```

## Troubleshooting

### "DB_PASSWORD environment variable is required"
- Ensure DB_PASSWORD is set in your .env file
- Or use the wrapper scripts which load .env automatically

### "Failed to connect to database"
- Check PostgreSQL is running: `systemctl status postgresql`
- Verify database exists: `psql -l | grep financial_data`
- Test connection: `psql -h localhost -U postgres -d financial_data`

### "PDF parsing not yet implemented"
- Add PDF parsing library (see Implementing Parser Logic)
- Implement bank-specific parsing logic
- Test with sample statements

### "No transactions found in statement"
- Check parser regex patterns match your bank's format
- Print extracted text to debug: `fmt.Println(text)`
- Adjust patterns based on your bank's statement format

### Exit code 1 (parse error)
- File format not supported
- Parser couldn't extract required fields
- Check logs for specific error message

### Exit code 2 (database error)
- Database connection failed
- Transaction insertion failed
- Check PostgreSQL logs: `/var/log/postgresql/`

## Project Structure

```
financial-statement-processor/
├── cmd/
│   ├── processor/
│   │   └── main.go              # Processor executable
│   └── query/
│       └── main.go              # Query executable
├── db/
│   └── postgres.go              # Database operations
├── parser/
│   └── parser.go                # PDF/OCR parsing logic
├── config/
│   └── config.go                # Configuration management
├── schema.sql                   # Database schema
├── .env.example                 # Example environment file
├── install.sh                   # Installation script
├── uninstall.sh                 # Uninstallation script
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
└── README.md                    # This file
```

## Security Considerations

- **Password storage**: .env file contains database password (chmod 600)
- **SQL injection**: All queries use parameterized statements
- **File validation**: Validates file existence and type before processing
- **Connection pooling**: Limits max database connections
- **SSL mode**: Configure DB_SSLMODE for production environments

## Development

### Add Dependencies

```bash
go get github.com/lib/pq
go get github.com/joho/godotenv
# Add your PDF/OCR libraries here
go mod tidy
```

### Run Tests

```bash
# Test with a sample statement
./financial-statement-processor-run ./test_statements/sample.pdf

# Check database
psql -d financial_data -c "SELECT COUNT(*) FROM transactions;"

# Query results
./financial-statement-query-run --start-date 2024-01-01 --end-date 2024-12-31 --pretty
```

## Supported Statement Formats

The parser currently needs customization for your specific bank. Common formats:

- **PDF statements**: Text-based PDFs (most online banking statements)
- **Scanned PDFs**: Require OCR (slower, less accurate)
- **Images**: JPG, PNG, TIFF (require OCR)

**Note:** Each bank has different statement formats. You'll need to adjust regex patterns and parsing logic for your specific bank's layout.

## License

This is a personal financial tool. Use and modify as needed.

## Contributing

This is part of a personal agent swarm system. Feel free to adapt for your own use cases.
