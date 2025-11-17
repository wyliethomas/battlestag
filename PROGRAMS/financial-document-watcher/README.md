# Financial Document Watcher

A Go CLI tool that monitors folders for financial documents and automatically triggers processing executables when files appear.

## Features

- Monitor multiple folders with different file patterns
- Execute custom processor binaries for each file type
- Automatic file movement to processed folders on success
- SQLite-based tracking to prevent duplicate processing
- Comprehensive logging for cron job integration
- Dry-run mode for testing configurations
- Retry-friendly: failed files remain in place for next run

## Use Cases

- Automatically process bank statements as they're downloaded
- OCR and categorize receipts
- Extract data from invoices
- Archive and index financial documents
- Trigger workflows based on file appearance

## Installation

### Prerequisites

- Go 1.16 or higher
- SQLite3 (usually pre-installed on Linux/macOS)

### Build from Source

1. Navigate to the project directory:

```bash
cd financial-document-watcher
```

2. Initialize Go module and install dependencies:

```bash
go mod init financial-document-watcher
go get github.com/mattn/go-sqlite3
go mod tidy
```

3. Build the static binary:

```bash
# Build for current platform
go build -o financial-document-watcher

# Build static binary for Linux (for deployment)
CGO_ENABLED=1 go build -ldflags '-extldflags "-static"' -o financial-document-watcher
```

4. Install to system (optional):

```bash
sudo cp financial-document-watcher /usr/local/bin/
sudo chmod +x /usr/local/bin/financial-document-watcher
```

## Configuration

### Create watches.json

Create a `watches.json` file with your watch configurations:

```bash
cp watches.json.example watches.json
```

Edit the file to configure your watches:

```json
[
  {
    "watch_id": "bank_statements",
    "watch_path": "/home/user/Documents/financial/incoming/bank",
    "file_pattern": "*.pdf",
    "executable_path": "/usr/local/bin/process-bank-statement",
    "processed_path": "/home/user/Documents/financial/processed/bank"
  }
]
```

### Configuration Fields

| Field | Description |
|-------|-------------|
| `watch_id` | Unique identifier for this watch (used in database tracking) |
| `watch_path` | Directory to monitor for files |
| `file_pattern` | Glob pattern to match files (e.g., `*.pdf`, `invoice_*.pdf`) |
| `executable_path` | Full path to the processor executable |
| `processed_path` | Directory where successfully processed files are moved |

### File Patterns

The `file_pattern` field uses standard glob patterns:

- `*.pdf` - All PDF files
- `invoice_*.pdf` - PDFs starting with "invoice_"
- `statement_[0-9]*.pdf` - PDFs starting with "statement_" followed by digits
- `*.{pdf,jpg,png}` - PDFs, JPGs, and PNGs (Note: requires shell expansion)

## Usage

### Basic Usage

Run the watcher once (typically via cron):

```bash
./financial-document-watcher
```

### Command-Line Options

```bash
# Use custom config file
./financial-document-watcher -config /path/to/watches.json

# Use custom database location
./financial-document-watcher -db /var/lib/watcher/watcher.db

# Dry-run mode (show what would be processed)
./financial-document-watcher -dry-run

# Combine options
./financial-document-watcher -config ./my-watches.json -dry-run
```

### Testing Your Configuration

Before setting up cron, test your configuration:

```bash
# Dry-run to see what would be processed
./financial-document-watcher -dry-run

# Run once manually
./financial-document-watcher

# Check the logs
echo $?  # Should be 0 on success
```

## Processor Executables

The watcher executes your processor binaries with the file path as the first argument:

```bash
/usr/local/bin/process-bank-statement "/path/to/file.pdf"
```

### Processor Requirements

Your processor executable must:

1. Accept file path as first command-line argument
2. Exit with code 0 on success
3. Exit with non-zero code on failure
4. Write errors to stderr (captured in logs)

### Example Processor (Bash)

```bash
#!/bin/bash
# /usr/local/bin/process-bank-statement

FILE_PATH="$1"

if [ ! -f "$FILE_PATH" ]; then
    echo "Error: File not found: $FILE_PATH" >&2
    exit 1
fi

# Process the file (example: extract text and save metadata)
pdftotext "$FILE_PATH" - | extract-transactions > /tmp/transactions.json

if [ $? -eq 0 ]; then
    echo "Successfully processed: $FILE_PATH"
    exit 0
else
    echo "Error: Failed to extract transactions" >&2
    exit 1
fi
```

### Example Processor (Python)

```python
#!/usr/bin/env python3
# /usr/local/bin/process-invoice

import sys
import os

def process_invoice(file_path):
    if not os.path.exists(file_path):
        print(f"Error: File not found: {file_path}", file=sys.stderr)
        return 1

    try:
        # Your processing logic here
        print(f"Processing: {file_path}")
        # ... extract invoice data ...
        print("Success")
        return 0
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: process-invoice <file_path>", file=sys.stderr)
        sys.exit(1)

    sys.exit(process_invoice(sys.argv[1]))
```

## Automation with Cron

The watcher is designed to be run periodically via cron.

### Setting Up Cron

1. Edit your crontab:

```bash
crontab -e
```

2. Add a cron entry:

```bash
# Run daily at 9 AM
0 9 * * * /usr/local/bin/financial-document-watcher -config /home/user/.config/watcher/watches.json >> /var/log/financial-watcher.log 2>&1

# Run every hour
0 * * * * /usr/local/bin/financial-document-watcher >> /var/log/financial-watcher.log 2>&1

# Run every 15 minutes during business hours (9 AM - 5 PM)
*/15 9-17 * * * /usr/local/bin/financial-document-watcher >> /var/log/financial-watcher.log 2>&1

# Run Monday-Friday at 9 AM and 5 PM
0 9,17 * * 1-5 /usr/local/bin/financial-document-watcher >> /var/log/financial-watcher.log 2>&1
```

### Cron Schedule Examples

| Schedule | Cron Expression | Use Case |
|----------|----------------|----------|
| Daily at 9 AM | `0 9 * * *` | Process overnight downloads |
| Every hour | `0 * * * *` | Frequent monitoring |
| Every 30 min | `*/30 * * * *` | Near real-time processing |
| Weekdays 9 AM | `0 9 * * 1-5` | Business day processing |
| 1st of month | `0 0 1 * *` | Monthly reports |

### Logging Best Practices

Create a dedicated log file with rotation:

```bash
# Create log directory
sudo mkdir -p /var/log/financial-watcher
sudo chown $USER:$USER /var/log/financial-watcher

# Cron entry with timestamped logging
0 9 * * * /usr/local/bin/financial-document-watcher >> /var/log/financial-watcher/watcher-$(date +\%Y-\%m).log 2>&1
```

Or use logger to send to syslog:

```bash
0 9 * * * /usr/local/bin/financial-document-watcher 2>&1 | logger -t financial-watcher
```

### Viewing Logs

```bash
# View today's log
tail -f /var/log/financial-watcher.log

# View last 100 lines
tail -n 100 /var/log/financial-watcher.log

# Search for errors
grep ERROR /var/log/financial-watcher.log

# View syslog entries
journalctl -t financial-watcher
```

## Database

The watcher uses SQLite to track processed files and prevent duplicate processing.

### Schema

```sql
CREATE TABLE processed_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    watch_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    processed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(watch_id, file_path)
);
```

### Inspecting the Database

```bash
# Open the database
sqlite3 watcher.db

# View all processed files
SELECT * FROM processed_files ORDER BY processed_at DESC;

# View files for specific watch
SELECT * FROM processed_files WHERE watch_id = 'bank_statements';

# Count processed files by watch
SELECT watch_id, COUNT(*) as count FROM processed_files GROUP BY watch_id;

# Clear all records (for testing)
DELETE FROM processed_files;
```

### Reprocessing Files

If you need to reprocess a file:

1. Remove it from the database:

```sql
DELETE FROM processed_files WHERE file_path = '/path/to/file.pdf';
```

2. Move the file back to the watch directory:

```bash
mv /path/to/processed/file.pdf /path/to/watch/
```

3. Run the watcher manually or wait for next cron run:

```bash
./financial-document-watcher
```

## Error Handling

### File Processing Failures

When a processor executable fails (non-zero exit code):

- Error is logged with exit code and output
- File is **left in place** in the watch directory
- File will be retried on next watcher run
- No database record is created

This retry-friendly behavior ensures failed files aren't lost.

### Common Issues

**"Watch path does not exist"**
- Ensure the watch_path directory exists before running
- Create directories: `mkdir -p /path/to/watch/dir`

**"Failed to execute: executable file not found"**
- Verify the executable_path is correct
- Ensure the processor has execute permissions: `chmod +x /path/to/processor`

**"Failed to move file to processed"**
- Check that processed_path is writable
- Ensure parent directory exists

**Duplicate processing**
- Database may be locked or corrupted
- Check database permissions
- Ensure only one watcher instance runs at a time

## Project Structure

```
financial-document-watcher/
├── main.go                    # Entry point and core logic
├── db/
│   └── sqlite.go              # Database operations
├── watches.json.example       # Example configuration
├── watches.json               # Your configuration (gitignored)
├── watcher.db                 # SQLite database (gitignored)
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
└── README.md                  # This file
```

## Deployment to Remote Server

Deploy the watcher to your home server:

1. Build the static binary:

```bash
CGO_ENABLED=1 GOOS=linux go build -ldflags '-extldflags "-static"' -o financial-document-watcher
```

2. Copy to server:

```bash
scp financial-document-watcher user@server:/tmp/
scp watches.json user@server:/tmp/
```

3. SSH and install:

```bash
ssh user@server
sudo mv /tmp/financial-document-watcher /usr/local/bin/
sudo chmod +x /usr/local/bin/financial-document-watcher

mkdir -p ~/.config/financial-watcher
mv /tmp/watches.json ~/.config/financial-watcher/
```

4. Set up cron on the remote server:

```bash
crontab -e

# Add:
0 9 * * * /usr/local/bin/financial-document-watcher -config ~/.config/financial-watcher/watches.json >> ~/.local/log/financial-watcher.log 2>&1
```

## Example Workflow

1. **Download financial documents** to watch directories
   - Browser downloads to `~/Documents/financial/incoming/bank/`

2. **Cron runs watcher** at scheduled time
   - Checks all watch directories for matching files

3. **Process new files**
   - Executes processor for each new file
   - Processor extracts data, updates database, etc.

4. **Move processed files**
   - On success, file moved to processed directory
   - Recorded in SQLite to prevent reprocessing

5. **Retry failures**
   - Failed files stay in watch directory
   - Next cron run will retry

## Security Considerations

- **File permissions**: Ensure watch and processed directories are only accessible by your user
- **Processor validation**: Only run trusted processor executables
- **Path sanitization**: The watcher passes file paths directly to processors
- **Database security**: Keep watcher.db readable only by your user
- **Logging**: Be mindful of sensitive data in logs

## Troubleshooting

### Debug Mode

Add verbose logging by modifying the code or use dry-run mode:

```bash
./financial-document-watcher -dry-run
```

### Test Individual Watches

Edit watches.json to include only one watch and test:

```bash
./financial-document-watcher -config test-watches.json
```

### Verify Processor Execution

Test your processor manually:

```bash
/usr/local/bin/process-bank-statement /path/to/test-file.pdf
echo $?  # Should print 0 on success
```

### Check Cron Execution

```bash
# View cron logs
grep CRON /var/log/syslog

# Or check user cron logs
journalctl -u cron | grep financial-watcher
```

## Advanced Configuration

### Environment Variables

Set environment variables for processors via cron:

```bash
0 9 * * * export API_KEY=secret; /usr/local/bin/financial-document-watcher >> /var/log/watcher.log 2>&1
```

### File Locking

For high-frequency monitoring, consider adding file locking to prevent concurrent runs:

```bash
0 * * * * flock -n /tmp/watcher.lock /usr/local/bin/financial-document-watcher
```

### Email Notifications on Errors

Send email when errors occur:

```bash
0 9 * * * /usr/local/bin/financial-document-watcher 2>&1 | grep -i error && echo "Watcher errors detected" | mail -s "Watcher Alert" you@example.com
```

## License

This is a personal utility program. Use and modify as needed.

## Contributing

This program is part of a personal agent swarm system. Feel free to adapt for your own use cases.
