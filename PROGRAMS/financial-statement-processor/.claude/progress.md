# Financial Statement Processor - Development Progress

## Session Summary (2025-11-17)

Successfully built and debugged a complete financial document processing system with LLM-based PDF parsing.

---

## ✅ Completed Features

### 1. Core System Architecture
- **Dual executables**: `financial-statement-processor` and `financial-statement-query`
- **SQLite database**: Single-file storage at `~/.local/share/financial-processor/transactions.db`
- **LLM-based parsing**: Uses local Ollama (dolphin3:8b) for privacy-preserving multi-bank support
- **Integration with watcher**: Works with `financial-document-watcher` for automated processing

### 2. PDF Parser Implementation
- **Text extraction**: Extracts text from all PDF pages using `github.com/ledongthuc/pdf`
- **Page-by-page LLM processing**: Processes each page independently to avoid context window limits
- **Transaction deduplication**: Removes duplicate entries when merging multi-page results
- **Verbose logging**: Comprehensive logging for debugging and monitoring

**Key Files:**
- `parser/parser.go` - Main parsing logic
- `parser/llm.go` - Ollama client and prompt engineering

### 3. Bug Fixes & Enhancements

#### Date Parsing Fix
**Problem**: LLM returned date as "25-10-01" (YY-MM-DD) which parser couldn't parse
**Solution**:
- Added 2-digit year formats to parseDate() (`"06-01-02"`, `"06/01/02"`)
- Made LLM prompt more explicit about requiring YYYY-MM-DD format
- Defensive coding approach: prompt improvement + fallback parsing

**Files Changed**: `parser/llm.go`, `parser/parser.go`

#### .env Loading Fix
**Problem**: After running `update.sh`, .env settings not loaded (using localhost instead of remote Ollama)
**Solution**:
- Enhanced `update.sh` to recreate wrapper scripts (not just binaries)
- Auto-detect CONFIG_DIR from existing wrapper script
- Wrapper scripts source .env before calling binary

**Files Changed**: `update.sh`

#### Multi-Page PDF Processing
**Problem**: Only first page transactions extracted (4 out of 5 pages missed)
**Root Cause**: All pages sent to LLM at once exceeded context window
**Solution**:
- Created `extractPDFTextByPage()` - extracts each page separately
- Created `parseMultiPageWithLLM()` - processes each page independently
- Created `deduplicateTransactions()` - merges results and removes duplicates
- Added comprehensive logging for each page

**Files Changed**: `parser/parser.go`

**Test Result**: Successfully extracted all transactions from 5-page PDF

### 4. Query Tool Enhancements

Added features to make query tool callable by other executables:

#### Exit Codes
- `0` - Success
- `1` - Invalid arguments (bad dates, invalid type)
- `2` - Database error

#### CSV Output Format
- New `--csv` flag outputs comma-separated values
- Header row with column names
- Properly quoted fields
- Easy parsing by other programs

#### Transaction Type Filtering
- New `--type` flag: `debit`, `credit`, or `all` (default)
- Enables expense-only or income-only queries
- Integrated with database layer

**Files Changed**:
- `cmd/query/main.go` - CLI enhancements
- `db/sqlite.go` - Added `QueryTransactionsWithType()` method

---

## 🧪 Testing Results

### Integration Test (Successful)
```
financial-document-watcher-run
  → Detected: GetDocument.pdf (5 pages)
  → Processor extracted all pages
  → LLM parsed each page independently
  → Total transactions: All extracted successfully
  → File moved to processed folder
  → Exit code: 0
```

### Query Tool Test Examples
```bash
# JSON output (default)
financial-statement-query-run --start-date 2025-10-01 --end-date 2025-10-31

# CSV output
financial-statement-query-run --start-date 2025-10-01 --end-date 2025-10-31 --csv

# Filter by type
financial-statement-query-run --start-date 2025-10-01 --end-date 2025-10-31 --type debit --csv
```

---

## 📋 Architecture Overview

```
financial-document-watcher (cron/systemd timer)
    ↓ watches folders
    ↓ finds new PDF
    ↓ calls ↓
financial-statement-processor
    ↓ extracts PDF text (page-by-page)
    ↓ sends to local Ollama LLM
    ↓ parses transactions
    ↓ deduplicates
    ↓ stores in SQLite
    ↓ exit code 0
    ↓
financial-document-watcher
    ↓ moves file to processed/

financial-statement-query (called by other tools/agents)
    ↓ queries SQLite with filters
    ↓ outputs JSON or CSV
    ↓ exit code 0/1/2
```

---

## 🔧 Current Configuration

**Environment Variables** (`.env`):
```
DB_PATH=~/.local/share/financial-processor/transactions.db
OLLAMA_HOST=http://192.168.1.232:11434
OLLAMA_MODEL=dolphin3:8b
```

**Wrapper Scripts**:
- `/usr/local/bin/financial-statement-processor-run`
- `/usr/local/bin/financial-statement-query-run`

**Database**: SQLite at `~/.local/share/financial-processor/transactions.db`

---

## 📝 Key Learnings

1. **LLM Context Windows**: Multi-page documents must be chunked to avoid hitting token limits
2. **Wrapper Scripts**: Essential for environment variable loading - must be recreated during updates
3. **Exit Codes**: Critical for inter-executable communication in swarm architecture
4. **Deduplication**: LLMs may extract overlapping content from pages - dedup before insertion
5. **Defensive Parsing**: Combine strict prompts with fallback format support

---

## 🎯 Swarm Architecture Pattern

This project established a reusable pattern for building agent swarms:

**Individual Executables** (Lightweight & Focused):
- Single responsibility
- Standard interfaces (exit codes, params, output formats)
- Stateless (no daemon processes)
- Composable (can pipe/chain together)
- Easy updates (`git pull && ./update.sh`)
- Isolated data (own database/config)
- Agent-ready (orchestrated by agent-gateway)

**Agent Gateway** (Only Service to Maintain):
- Calls executables on-demand with params
- Parses stdout + exit codes
- No inter-process communication complexity

---

## 🚀 Next Steps / Future Enhancements

### Potential Improvements (Not Critical)
- [ ] Add OCR support for scanned PDFs (tesseract integration)
- [ ] Add `--limit` and `--offset` to query tool for pagination
- [ ] Add `--fields` flag to select specific columns in output
- [ ] Implement retry logic if LLM extraction seems incomplete
- [ ] Add validation warning if transaction count seems low

### Other Executables to Build (Examples)
- Expense analyzer (reads transactions, generates reports)
- Budget tracker (compares spending to limits)
- Anomaly detector (flags unusual transactions)
- Receipt processor (extracts data from receipt images)

---

## 📚 Documentation

- **Main README**: Comprehensive guide at `README.md` (506 lines)
- **Installation**: `./install.sh` with interactive prompts
- **Updates**: `./update.sh` rebuilds and recreates wrappers
- **Uninstallation**: `./uninstall.sh` removes everything

---

## ✨ Success Metrics

- ✅ End-to-end automation working
- ✅ Multi-page PDF support (tested with 5 pages)
- ✅ Privacy-preserving (local LLM, no external APIs)
- ✅ Query tool ready for other executables
- ✅ Proper error handling with exit codes
- ✅ Multiple output formats (JSON, CSV)
- ✅ Zero manual intervention required

**Status**: Production-ready for single-user home network deployment

---

*Last Updated: 2025-11-17*
*Architecture: Swarm - Lightweight, Stateless, Composable*
