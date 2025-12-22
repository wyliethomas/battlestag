## Utilities Layer

**Higher-level operations that compose primitives**

## Overview

The utilities layer composes primitives to create intelligent, format-aware operations. Each utility:

- ✅ Composes multiple primitives
- ✅ Handles specific formats (JSON, CSV, logs)
- ✅ Provides error handling and retries
- ✅ Returns structured output
- ✅ Has comprehensive test coverage (100%)

## Modules

### file.sh - File Utilities

Higher-level file operations that compose primitives.

```bash
# Backup & Safety
util.backup <file>                        # Create timestamped backup
util.safe_write <file> <content>          # Backup then write
util.restore <file> <timestamp>           # Restore from backup

# Directory Management
util.ensure_dir <path>                    # Create if not exists
util.clean_old_files <dir> <days>        # Remove old files
util.archive <dir> <output.tar.gz>       # Create archive
util.unarchive <archive> <dest>          # Extract archive

# Templates & Config
util.template <tmpl> <out> <var=val>...  # Replace {{vars}}
util.config_get <file> <key>             # Get KEY=VALUE
util.config_set <file> <key> <value>     # Set KEY=VALUE

# Helpers
util.file_age <file>                     # Age in seconds
util.temp_file <prefix>                  # Create temp file
util.temp_dir <prefix>                   # Create temp directory
```

**Examples:**
```bash
# Safe file operations with backup
util.safe_write "config.json" "{\"foo\": \"bar\"}"

# Template processing
util.template "email.tmpl" "email.txt" "name=John" "email=john@example.com"

# Config management
util.config_set ".env" "DATABASE_URL" "postgres://localhost/db"
VALUE=$(util.config_get ".env" "DATABASE_URL")
```

### json.sh - JSON Utilities

JSON manipulation using jq. **Requires**: `jq`

```bash
# Read/Write
util.json.read <file> <path>              # Extract value (.foo.bar)
util.json.write <file> <path> <value>     # Update field (JSON value)
util.json.write_string <file> <path> <str> # Update field (string)
util.json.delete <file> <path>            # Remove field

# Validation & Query
util.json.validate <file>                 # Check valid JSON
util.json.query <file> <jq_expr>         # Complex jq queries
util.json.keys <file> <path>             # List keys

# Array Operations
util.json.array_length <file> <path>      # Get array length
util.json.array_append <file> <path> <val> # Append to array
util.json.array_append_string <file> <path> <str> # Append string

# Formatting
util.json.pretty <file>                   # Pretty-print (in-place)
util.json.compact <file>                  # Compact (in-place)
util.json.merge <file1> <file2> <out>    # Merge two JSON files
```

**Examples:**
```bash
# Read/write JSON
util.json.write_string "config.json" ".database.host" "localhost"
HOST=$(util.json.read "config.json" ".database.host")

# Array manipulation
util.json.array_append_string "users.json" ".users" "john@example.com"
COUNT=$(util.json.array_length "users.json" ".users")

# Query with jq
util.json.query "data.json" '.users[] | select(.age > 30) | .name'
```

### log.sh - Logging Utilities

Structured logging with timestamps and levels.

```bash
# Basic Logging
util.log.info <message> [logfile]         # Info level
util.log.error <message> [logfile]        # Error level
util.log.warn <message> [logfile]         # Warning level
util.log.debug <message> [logfile]        # Debug level (if DEBUG=1)

# Log Management
util.log.tail <logfile> [lines]           # Show recent entries
util.log.grep <logfile> <pattern>         # Search logs
util.log.rotate <logfile> [keep_count]   # Rotate logs
util.log.clean <logdir> <days>           # Remove old logs

# Log Analysis
util.log.size <logfile>                   # Human-readable size
util.log.stats <logfile>                  # Statistics
util.log.levels <logfile>                 # Count by level
util.log.today <logfile>                  # Today's entries
util.log.errors <logfile>                 # Only errors
```

**Examples:**
```bash
# Logging
util.log.info "Application started" "/var/log/app.log"
util.log.error "Database connection failed" "/var/log/app.log"

# Analysis
util.log.stats "/var/log/app.log"
util.log.errors "/var/log/app.log" | tail -10

# Maintenance
util.log.rotate "/var/log/app.log" 10
util.log.clean "/var/log" 30
```

**Log Format:**
```
[2025-12-18T16:00:00Z] [hostname] [user] [INFO] Application started
[2025-12-18T16:00:05Z] [hostname] [user] [ERROR] Database connection failed
```

### llm.sh - LLM API Utilities

LLM integration for AI-powered operations. **Requires**: `curl`

```bash
# Basic Queries
util.llm.query <prompt> [max_tokens]      # Simple query
util.llm.with_context <file> <prompt>     # Query with file context
util.llm.multi_file <files...> <prompt>   # Query with multiple files

# Specialized Operations
util.llm.summarize <file>                 # Summarize content
util.llm.explain <file>                   # Explain code
util.llm.review <file>                    # Code review

# Management
util.llm.test_connection                  # Test endpoint
util.llm.config                           # Show configuration
util.llm.models                           # List models
```

**Configuration** (environment variables):
```bash
export LLM_ENDPOINT="http://localhost:1234/v1"
export LLM_MODEL="hermes-3-llama-3.1-8b"
export LLM_TIMEOUT=30
export LLM_MAX_TOKENS=2000
export LLM_TEMPERATURE=0.7
```

**Examples:**
```bash
# Simple query
util.llm.query "What is the capital of France?"

# Code review
util.llm.review "src/auth.rb"

# With context
util.llm.with_context "README.md" "Summarize this project"

# Test connection
util.llm.test_connection
```

### data.sh - Data Processing Utilities

CSV, text processing, hashing, and encoding.

```bash
# CSV Operations
util.csv.read <file> <row> <col>          # Get cell (1-indexed)
util.csv.write <file> <data>              # Append row
util.csv.header <file>                    # Get header
util.csv.count <file>                     # Count rows
util.csv.query <file> <col_name> <value>  # Find rows
util.csv.column <file> <col_number>       # Extract column

# Text Operations
util.text.trim <text>                     # Remove whitespace
util.text.upper <text>                    # Uppercase
util.text.lower <text>                    # Lowercase
util.text.replace <text> <old> <new>      # Replace all
util.text.lines <text>                    # Count lines
util.text.words <text>                    # Count words
util.text.chars <text>                    # Count characters
util.text.substring <text> <start> <len>  # Extract substring
util.text.split <text> <delimiter>        # Split to lines
util.text.join <delim> <args...>          # Join with delimiter

# Hashing & Encoding
util.hash.md5 <file>                      # MD5 checksum
util.hash.sha256 <file>                   # SHA256 checksum
util.encode.base64 <text>                 # Base64 encode
util.decode.base64 <text>                 # Base64 decode
util.encode.url <text>                    # URL encode
util.decode.url <text>                    # URL decode

# Data Transformation
util.data.to_json csv <file>              # CSV to JSON
util.data.stats <file>                    # File statistics
```

**Examples:**
```bash
# CSV processing
util.csv.write "data.csv" "name,age,city"
util.csv.write "data.csv" "John,30,NYC"
NAME=$(util.csv.read "data.csv" 2 1)

# Text manipulation
UPPER=$(util.text.upper "hello world")
TRIMMED=$(util.text.trim "  text  ")

# Hashing
MD5=$(util.hash.md5 "file.txt")
SHA=$(util.hash.sha256 "file.txt")

# Encoding
ENCODED=$(util.encode.base64 "secret")
DECODED=$(util.decode.base64 "$ENCODED")
```

## Usage Patterns

### As Library (Source in Scripts)

```bash
#!/usr/bin/env bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UTILS_DIR="$SCRIPT_DIR/../../SCRIPTS/utilities"

source "$UTILS_DIR/file.sh"
source "$UTILS_DIR/json.sh"
source "$UTILS_DIR/log.sh"

# Now use utilities
util.log.info "Script started"
util.json.write_string "config.json" ".status" "running"
util.backup "important_file.txt"
```

### As CLI Tools

```bash
# Call utilities directly
./SCRIPTS/utilities/file.sh backup myfile.txt
./SCRIPTS/utilities/json.sh read config.json '.database.host'
./SCRIPTS/utilities/log.sh info "Message" /var/log/app.log
./SCRIPTS/utilities/llm.sh query "What is 2+2?"
./SCRIPTS/utilities/data.sh text.upper "hello"
```

### In Modules

Modules compose utilities for complete workflows:

```bash
# In modules/ai/code_review.sh
source "../../utilities/file.sh"
source "../../utilities/llm.sh"
source "../../utilities/log.sh"

code_review() {
    local file="$1"

    util.log.info "Starting code review: $file"
    util.backup "$file"

    local review
    review=$(util.llm.review "$file")

    util.log.info "Review complete"
    echo "$review"
}
```

## Testing

Run the comprehensive test suite:

```bash
cd SCRIPTS/utilities
./test_utilities.sh
```

**Test Coverage:**
- 29 tests across all 5 modules
- 100% success rate
- Tests all core functionality
- Validates error handling
- Verifies composition of primitives

## Dependencies

### Required
- bash 4.0+
- Primitives layer (file.sh, dir.sh, sys.sh, search.sh)

### Optional (per utility)
- **json.sh**: `jq` (JSON processing)
- **llm.sh**: `curl` (API calls)
- **data.sh**: `python3` (CSV to JSON conversion)

Check dependencies:
```bash
cd SCRIPTS
./check_dependencies.sh
```

## Performance

Utilities are composed operations, slightly slower than primitives:

| Utility | Typical Time | Notes |
|---------|-------------|-------|
| util.backup | 5-20ms | file.copy + timestamp |
| util.json.read | 5-15ms | jq overhead |
| util.log.info | 2-5ms | file.append + formatting |
| util.llm.query | 500-5000ms | Network + LLM inference |
| util.text.upper | < 1ms | Pure bash tr |

## Design Principles

### 1. Compose Primitives
Every utility uses primitives - no direct system calls:

```bash
util.backup() {
    local timestamp=$(sys.timestamp)  # primitive
    file.copy "$src" "${src}.backup.${timestamp}"  # primitive
}
```

### 2. Format Awareness
Utilities understand file formats:
- json.sh knows JSON structure
- csv.sh understands CSV format
- log.sh creates structured logs

### 3. Error Handling
Robust error handling with meaningful messages:

```bash
if ! file.exists "$file"; then
    echo "Error: File not found: $file" >&2
    return 1
fi
```

### 4. Composition-Friendly
Utilities work well together:

```bash
# Backup, modify, log
util.backup "$config"
util.json.write_string "$config" ".version" "2.0"
util.log.info "Config updated to v2.0"
```

## Common Patterns

### Safe File Modification

```bash
# Always backup before modifying
util.backup "important_file.txt"
util.safe_write "important_file.txt" "$new_content"
```

### Structured Logging

```bash
# Consistent logging throughout application
LOG_FILE="/var/log/myapp.log"
util.log.info "Operation started" "$LOG_FILE"
# ... do work ...
util.log.info "Operation completed" "$LOG_FILE"
```

### JSON Configuration Management

```bash
# Read current config
HOST=$(util.json.read "config.json" ".database.host")

# Update config
util.json.write_string "config.json" ".database.host" "new-host"

# Validate
util.json.validate "config.json"
```

### LLM-Powered Operations

```bash
# Code review workflow
util.log.info "Starting review"
util.backup "$code_file"
review=$(util.llm.review "$code_file")
echo "$review" > "${code_file}.review"
util.log.info "Review saved"
```

## Error Handling

All utilities follow consistent error handling:

```bash
# Success
util.json.read "file.json" ".foo"
echo $?  # 0

# Failure
util.json.read "missing.json" ".foo"
echo $?  # 1 (error message sent to stderr)
```

Wrap in conditionals:

```bash
if util.json.validate "config.json"; then
    echo "Valid JSON"
else
    echo "Invalid JSON"
fi
```

## Environment Variables

### Global Settings

```bash
# Logging
export LOG_DIR="~/.local/share/battlestag/logs"
export DEBUG=1  # Enable debug logging

# LLM
export LLM_ENDPOINT="http://localhost:1234/v1"
export LLM_MODEL="hermes-3-llama-3.1-8b"
export LLM_TIMEOUT=30
export LLM_MAX_TOKENS=2000
export LLM_TEMPERATURE=0.7
```

## Next Steps

The utilities layer is now complete. Next up:

1. **Modules Layer** - Domain-specific workflows (AI, Finance, Health, Business)
2. **API Integration** - Connect to agent-gateway
3. **TUI Integration** - Add command palette entries

## Contributing

When adding new utilities:

1. Source required primitives at top
2. Follow naming convention: `util.category.operation`
3. Add error handling (exit codes + stderr)
4. Write tests in test_utilities.sh
5. Update this README
6. Ensure 100% test pass rate

---

**Status**: ✅ Complete - All 29 tests passing
**Version**: 1.0
**Last Updated**: 2025-12-18
