# BATTLESTAG Platform - Capability Hierarchy Architecture

**Date**: 2025-12-18
**Version**: 2.0
**Status**: Active Design

---

## 🎯 Core Principles

1. **Files Over Database**: All state in flat files (JSON, CSV, plain text)
2. **Ruthlessly Reduce Context**: Each layer only knows what it needs
3. **Composability**: Small, testable units that combine elegantly
4. **LLM-Friendly**: 8B models see minimal context per operation
5. **Client Portability**: Easy to deploy, customize, and maintain

---

## 🏗️ Three-Layer Capability Hierarchy

### Overview

```
┌─────────────────────────────────────────────────────┐
│ MODULES (Business Logic)                            │
│ - Domain-specific workflows                         │
│ - Complete use cases                                │
│ - Composes utilities for end-to-end solutions       │
└─────────────────────┬───────────────────────────────┘
                      │ calls
┌─────────────────────┴───────────────────────────────┐
│ UTILITIES (Composed Operations)                     │
│ - Higher-level operations                           │
│ - Combines primitives with logic                    │
│ - Format-aware (JSON, logs, templates)              │
└─────────────────────┬───────────────────────────────┘
                      │ calls
┌─────────────────────┴───────────────────────────────┐
│ PRIMITIVES (Foundation)                             │
│ - Pure bash operations                              │
│ - Filesystem & system calls                         │
│ - Zero dependencies                                 │
└─────────────────────────────────────────────────────┘
```

---

## 📦 Layer 1: PRIMITIVES

**Purpose**: Core filesystem and system operations - the atoms of the system

**Characteristics**:
- Pure bash - no external dependencies
- Single responsibility
- Idempotent where possible
- Error codes only (no output formatting)
- Fast execution (< 100ms typical)

### File Operations

```bash
# Read/Write
file.read <path>              # Cat a file to stdout
file.write <path> <content>   # Write/overwrite file
file.append <path> <content>  # Append to file

# Metadata
file.exists <path>            # Check existence (exit 0/1)
file.size <path>              # Get file size in bytes
file.modified <path>          # Last modification timestamp

# Operations
file.delete <path>            # Remove file
file.copy <src> <dst>         # Copy file
file.move <src> <dst>         # Move/rename file
file.permissions <path> <mode> # chmod
```

### Directory Operations

```bash
# CRUD
dir.create <path>             # mkdir -p
dir.list <path>               # ls (files only)
dir.tree <path> <depth>       # Recursive list
dir.exists <path>             # Check existence (exit 0/1)
dir.delete <path>             # rm -rf (careful!)

# Navigation
dir.current                   # pwd
dir.parent <path>             # Get parent directory
dir.basename <path>           # Get directory name
```

### Search Operations

```bash
# Content search
search.grep <pattern> <path>  # grep -r
search.count <pattern> <path> # Count matches

# File search
search.find <pattern> <path>  # find by name
search.locate <name>          # Quick locate (if available)
search.which <command>        # Find executable

# Context search
search.before <pattern> <path> <lines>  # grep -B
search.after <pattern> <path> <lines>   # grep -A
search.around <pattern> <path> <lines>  # grep -C
```

### System Operations

```bash
# Environment
sys.exec <command>            # Execute arbitrary command
sys.env <var>                 # Get environment variable
sys.setenv <var> <value>      # Set environment variable

# Info
sys.hostname                  # Get hostname
sys.user                      # Get current user
sys.date                      # Current date/time
sys.uptime                    # System uptime

# Process
sys.pid                       # Current process ID
sys.ppid                      # Parent process ID
sys.kill <pid> <signal>       # Kill process
```

---

## 🔧 Layer 2: UTILITIES

**Purpose**: Higher-level operations that combine primitives with business logic

**Characteristics**:
- Composes multiple primitives
- Format-aware (JSON, CSV, logs)
- Error handling and retries
- Returns structured output
- May maintain state files

### File Utilities

```bash
# Safe operations
util.backup <file>                    # Create timestamped backup
util.safe_write <file> <content>      # Backup then write
util.restore <file> <timestamp>       # Restore from backup

# Directory management
util.ensure_dir <path>                # Create if not exists
util.clean_dir <path> <days>          # Remove old files
util.archive <dir> <output>           # Tar/gzip directory

# Templates
util.template <tmpl> <out> <vars>     # Replace {{vars}} in template
util.config <file> <key> <value>      # Update config file (KEY=VALUE format)
```

### JSON Utilities

```bash
# Read/Write (using jq)
util.json.read <file> <path>          # Extract value: .foo.bar
util.json.write <file> <path> <value> # Update field
util.json.delete <file> <path>        # Remove field
util.json.merge <file1> <file2>       # Merge two JSON files

# Validation
util.json.validate <file>             # Check if valid JSON
util.json.schema <file> <schema>      # Validate against schema

# Query
util.json.query <file> <jq_expr>      # Complex jq queries
util.json.keys <file> <path>          # List keys at path
```

### Log Utilities

```bash
# Logging
util.log.info <message>               # Timestamped info log
util.log.error <message>              # Timestamped error log
util.log.debug <message>              # Debug log (if enabled)
util.log.warn <message>               # Warning log

# Management
util.log.rotate <logfile>             # Rotate logs (keep last 10)
util.log.clean <logdir> <days>        # Remove old logs
util.log.tail <logfile> <lines>       # Show recent entries
util.log.grep <logfile> <pattern>     # Search logs
```

### LLM Utilities

```bash
# Basic queries
util.llm.query <prompt>               # Send prompt to LM Studio
util.llm.streaming <prompt>           # Streaming response

# Context-aware
util.llm.with_context <file> <prompt> # Query with file context
util.llm.multi_file <files> <prompt>  # Query with multiple files

# Specialized
util.llm.summarize <file>             # Summarize content
util.llm.translate <text> <lang>      # Translate text
util.llm.extract <file> <query>       # Extract specific info
```

### Data Utilities

```bash
# CSV operations
util.csv.read <file> <row> <col>      # Get cell value
util.csv.write <file> <row> <data>    # Append row
util.csv.query <file> <col> <value>   # Find rows

# Text processing
util.text.trim <text>                 # Remove whitespace
util.text.upper <text>                # Uppercase
util.text.lower <text>                # Lowercase
util.text.replace <text> <old> <new>  # Replace all

# Hashing/Encoding
util.hash.md5 <file>                  # MD5 checksum
util.hash.sha256 <file>               # SHA256 checksum
util.encode.base64 <text>             # Base64 encode
util.decode.base64 <text>             # Base64 decode
```

---

## 🎯 Layer 3: MODULES

**Purpose**: Domain-specific capabilities that compose utilities for complete workflows

**Characteristics**:
- Complete use cases
- Multiple utilities orchestrated
- LLM integration for intelligence
- Error recovery and rollback
- User-facing results

### AI Development Module

**Location**: `~/scripts/modules/ai/`

```bash
# Code analysis
module.ai.code_review <file>
  → util.llm.with_context <file> "Review this code..."
  → util.log.info "Starting code review..."
  → Parse response, create report
  → util.json.write report.json

module.ai.explain <file>
  → util.llm.with_context <file> "Explain this code..."
  → Format as markdown
  → util.file.write explanation.md

# Code generation
module.ai.generate_tests <file>
  → util.llm.with_context <file> "Generate unit tests..."
  → Extract test code from response
  → util.file.write test_file
  → util.log.info "Tests generated"

module.ai.refactor <file> <goal>
  → util.backup <file>
  → util.llm.with_context <file> "Refactor to: <goal>..."
  → Parse refactored code
  → util.safe_write <file> <new_code>
  → util.log.info "Refactored successfully"

# Documentation
module.ai.doc_gen <file>
  → util.llm.with_context <file> "Generate documentation..."
  → Format as markdown/docstrings
  → util.template docs.tmpl
  → util.file.write docs/

module.ai.security_scan <file>
  → util.llm.with_context <file> "Security audit..."
  → Parse vulnerabilities
  → util.json.write security-report.json
  → util.log.warn for each issue
```

### Finance Module

**Location**: `~/scripts/modules/finance/`

```bash
# Transaction processing
module.finance.analyze_transaction <file>
  → util.csv.read <file> (parse transactions)
  → util.llm.query "Categorize these transactions..."
  → util.json.write categorized.json
  → util.log.info "Processed N transactions"

module.finance.budget_check <category>
  → util.json.read budget.json .categories.<category>
  → util.csv.query transactions.csv "category" <category>
  → Calculate totals
  → Compare to budget
  → util.log.warn if over budget

# Reporting
module.finance.report <month>
  → util.csv.query transactions.csv "month" <month>
  → Calculate summaries
  → util.llm.query "Analyze spending patterns..."
  → util.template report.tmpl
  → util.file.write reports/month.md
```

### Health Module

**Location**: `~/scripts/modules/health/`

```bash
# Data logging
module.health.log_metric <type> <value>
  → util.json.read health-data.json
  → Append new metric with timestamp
  → util.json.write health-data.json
  → util.log.info "Logged <type>: <value>"

module.health.analyze_trends <metric>
  → util.json.query health-data.json (last 30 days)
  → util.llm.query "Analyze health trends..."
  → Generate insights
  → util.file.write insights.md

# Reminders
module.health.reminder <condition>
  → util.json.read health-data.json
  → Evaluate condition (e.g., "water < 8 glasses")
  → util.log.warn if condition met
  → Send notification
```

### Business Module

**Location**: `~/scripts/modules/business/`

```bash
# Project management
module.business.project_status <project>
  → util.json.read projects/<project>/status.json
  → util.csv.read projects/<project>/tasks.csv
  → Calculate completion percentage
  → util.llm.query "Summarize project status..."
  → Format report
  → util.file.write status-report.md

module.business.time_entry <project> <hours> <note>
  → util.json.read projects/<project>/time.json
  → Append time entry with timestamp
  → util.json.write
  → util.log.info "Logged <hours>h for <project>"

# Invoicing
module.business.invoice_generate <client> <month>
  → util.csv.query time-entries.csv "client" <client> "month" <month>
  → Calculate totals
  → util.template invoice.tmpl
  → util.file.write invoices/<client>-<month>.pdf
```

---

## 📁 Directory Structure

```
~/scripts/
├── primitives/
│   ├── file.sh              # All file.* functions
│   ├── dir.sh               # All dir.* functions
│   ├── search.sh            # All search.* functions
│   └── sys.sh               # All sys.* functions
│
├── utilities/
│   ├── file.sh              # util.backup, util.safe_write, etc.
│   ├── json.sh              # util.json.* functions
│   ├── log.sh               # util.log.* functions
│   ├── llm.sh               # util.llm.* functions
│   └── data.sh              # util.csv.*, util.text.*, etc.
│
└── modules/
    ├── ai/
    │   ├── code_review.sh
    │   ├── generate_tests.sh
    │   ├── explain.sh
    │   ├── refactor.sh
    │   ├── doc_gen.sh
    │   └── security_scan.sh
    │
    ├── finance/
    │   ├── analyze_transaction.sh
    │   ├── budget_check.sh
    │   └── report.sh
    │
    ├── health/
    │   ├── log_metric.sh
    │   ├── analyze_trends.sh
    │   └── reminder.sh
    │
    └── business/
        ├── project_status.sh
        ├── time_entry.sh
        └── invoice_generate.sh
```

---

## 🔄 Data Flow Example

**Use Case**: AI Code Review

```
User: "ai:code_review src/auth.rb"
  ↓
TUI → agent-gateway → /api/modules/ai/code_review
  ↓
module.ai.code_review.sh invoked with args: ["src/auth.rb"]
  ↓
Calls: util.backup src/auth.rb
  → Calls: file.copy src/auth.rb → src/auth.rb.backup.1234567890
  ↓
Calls: util.llm.with_context src/auth.rb "Review this Ruby code..."
  → Calls: file.read src/auth.rb
  → Calls: util.llm.query "<file content> Review this code..."
  → LM Studio API call
  → Returns: "Issues found: 1. SQL injection risk on line 45..."
  ↓
Parse LLM response
  ↓
Calls: util.json.write review-report.json
  → Calls: file.write review-report.json
  ↓
Calls: util.log.info "Code review complete"
  → Calls: file.append ~/.local/share/battlestag/logs/ai.log
  ↓
Return structured response to API
  ↓
API → TUI (displays review results)
```

**Context Used**:
- Module sees: filepath, LLM response
- Utilities see: file contents, JSON structures
- Primitives see: raw bytes, paths
- LLM sees: only the file content + prompt (< 8k tokens)

---

## 🎨 Integration with Existing Architecture

### TUI (Keep 100%)

**Current**: Command palette with categories
**Enhancement**: Auto-discover modules at startup

```go
// ui/commands.go (extend)
func loadDynamicCommands(apiClient *client.Client) {
    modules := apiClient.GetModules() // GET /api/modules

    for _, module := range modules {
        commands = append(commands, Command{
            Name: module.Name,
            Key: module.Key,
            Description: module.Description,
        })
    }
}
```

### Agent Gateway (Keep 100%)

**Current**: Programs registry
**Enhancement**: ModuleProgram type

```go
// programs/module_program.go (NEW)
type ModuleProgram struct {
    category    string
    name        string
    scriptPath  string
    description string
}

func (p *ModuleProgram) Execute(args map[string]interface{}) (*Result, error) {
    // Build command: bash ~/scripts/modules/<category>/<name>.sh <args>
    // Execute via CommandBuilder
    // Return structured result
}

// Auto-register on startup
func RegisterModulesFromDirectory(registry *Registry, baseDir string) {
    // Scan ~/scripts/modules/
    // Register each .sh file as ModuleProgram
}
```

### API Endpoints (NEW)

```
GET  /api/modules                    # List all available modules
GET  /api/modules/:category          # List modules in category
POST /api/modules/:category/:name    # Execute module
GET  /api/primitives                 # List available primitives (for debugging)
GET  /api/utilities                  # List available utilities (for debugging)
```

---

## 🚀 Implementation Phases

### Phase 1: Primitives Foundation (Week 1)
- [ ] Create `~/scripts/primitives/` structure
- [ ] Implement file.sh operations
- [ ] Implement dir.sh operations
- [ ] Implement search.sh operations
- [ ] Implement sys.sh operations
- [ ] Write test suite for primitives
- [ ] Document each primitive function

### Phase 2: Utilities Layer (Week 2)
- [ ] Create `~/scripts/utilities/` structure
- [ ] Implement file utilities (backup, safe_write, etc.)
- [ ] Implement JSON utilities (using jq)
- [ ] Implement log utilities
- [ ] Implement LLM utilities (wrapper for LM Studio API)
- [ ] Write test suite for utilities
- [ ] Document each utility function

### Phase 3: Modules Layer (Week 3)
- [ ] Create `~/scripts/modules/` structure
- [ ] Implement AI module (code_review, generate_tests)
- [ ] Implement Finance module (analyze_transaction)
- [ ] Implement Health module (log_metric)
- [ ] Implement Business module (project_status)
- [ ] Write test suite for modules
- [ ] Document each module

### Phase 4: API Integration (Week 4)
- [ ] Create ModuleProgram type in agent-gateway
- [ ] Implement module auto-discovery
- [ ] Add /api/modules endpoints
- [ ] Update TUI command loading
- [ ] End-to-end testing
- [ ] Performance optimization

---

## 💡 Key Benefits

### For Development
- **Clear boundaries**: Each layer has single responsibility
- **Easy testing**: Small units, easy to mock
- **Debuggable**: Can test each layer independently
- **Maintainable**: Changes isolated to one layer

### For LLM Integration
- **Minimal context**: 8B models only see what they need
- **Fast responses**: Less context = faster inference
- **Reliable**: Deterministic primitives + smart modules
- **Scalable**: Add modules without changing infrastructure

### For Client Deployments
- **Portable**: Pure bash, minimal dependencies
- **Customizable**: Mix and match modules per client
- **Secure**: All data stays local (files only)
- **Cost-effective**: No cloud API costs

### For Users
- **Fast**: Primitives execute in milliseconds
- **Predictable**: Same input = same output
- **Transparent**: Can see exactly what each layer does
- **Flexible**: Easy to add custom modules

---

## 📊 Comparison to Old Architecture

| Aspect | Old Approach | New Approach |
|--------|--------------|--------------|
| Script organization | Flat, monolithic | 3-layer hierarchy |
| Code reuse | Copy-paste | Compose from primitives |
| Testing | Hard to test big scripts | Easy to test small units |
| LLM context | Large, unfocused | Minimal, targeted |
| Client customization | Edit scripts | Add/remove modules |
| Debugging | Log everything, grep | Test each layer |
| Performance | Variable | Predictable (primitives fast) |
| Dependencies | Scattered | Explicit per layer |

---

## 🎯 Success Criteria

### Primitives Layer
- ✅ All primitives execute in < 100ms
- ✅ 100% test coverage
- ✅ Zero external dependencies
- ✅ Error codes only (exit 0/1)
- ✅ Idempotent operations

### Utilities Layer
- ✅ All utilities compose primitives only
- ✅ Structured output (JSON/text)
- ✅ Error handling and retries
- ✅ 90%+ test coverage
- ✅ Format-aware (JSON, CSV, logs)

### Modules Layer
- ✅ Complete workflows
- ✅ LLM integration working
- ✅ Error recovery mechanisms
- ✅ User-friendly output
- ✅ Logging and monitoring

### Integration
- ✅ TUI auto-discovers modules
- ✅ End-to-end workflow < 5 seconds
- ✅ API handles all module types
- ✅ Deployment script works on fresh Ubuntu
- ✅ Client documentation complete

---

## 🔮 Future Enhancements

### Phase 5: Advanced Features
- [ ] Module dependency management
- [ ] Parallel execution of independent modules
- [ ] Module marketplace (share modules)
- [ ] Visual module builder (no-code)
- [ ] Remote module execution (distributed)

### Phase 6: Intelligence Layer
- [ ] Learning from module execution patterns
- [ ] Auto-suggest next modules
- [ ] Predictive context loading
- [ ] Self-optimizing workflows

### Phase 7: Enterprise Features
- [ ] RBAC for modules
- [ ] Audit logging
- [ ] Compliance reporting
- [ ] Multi-tenant support

---

## 📝 Notes

**Why This Architecture Works**:
1. **Files over database**: All state in flat files, easy to backup/version
2. **Ruthless context reduction**: Each layer is a firewall against complexity
3. **Composability**: Mix and match primitives → utilities → modules
4. **LLM-friendly**: 8B models thrive with focused context
5. **Future-proof**: Easy to add layers or modify without breaking existing

**Trade-offs**:
- More files, but easier to understand
- More function calls, but clearer flow
- More structure, but better maintainability

**This is the foundation for the next 5 years of development.**

---

*Last Updated: 2025-12-18*
*Next Review: After Phase 1 completion*
