# Program Development Guide

**How to Build "Bolt-On" Programs for the Battlestag Ecosystem**

This guide walks you through creating new programs that integrate with the Battlestag Agent Gateway API and LLM system.

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Step-by-Step Tutorial](#step-by-step-tutorial)
4. [Best Practices](#best-practices)
5. [Testing](#testing)
6. [Deployment](#deployment)
7. [Examples](#examples)

---

## Overview

### What is a "Bolt-On" Program?

A bolt-on program is a self-contained CLI application that:
- Performs a specific function (task management, finance tracking, habit tracking, etc.)
- Exposes its functionality through command-line executables
- Integrates with the API through a thin wrapper
- Can be discovered and used by the LLM automatically

### Benefits of This Architecture

✅ **Separation of Concerns**: CLI logic separate from API logic
✅ **Reusable**: Programs can be used standalone or via API
✅ **Scalable**: Add new programs without changing core API
✅ **Testable**: Each program can be tested independently
✅ **LLM-Friendly**: Automatic discovery and natural language interface

---

## Architecture

### The Two-Part Pattern

```
┌─────────────────────────────────────────┐
│  1. STANDALONE CLI PROGRAM              │
│  (PROGRAMS/{program-name}/)             │
│                                         │
│  ┌──────────┐  ┌──────────┐           │
│  │ cmd/     │  │ db/      │           │
│  │ main.go  │  │ sqlite.go│           │
│  └──────────┘  └──────────┘           │
│                                         │
│  Creates executables in ~/.local/bin   │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│  2. API WRAPPER                         │
│  (ORCHESTRATOR/agent-gateway/programs/) │
│                                         │
│  Implements Program interface           │
│  Calls CLI executables                  │
│  Exposes to LLM                         │
└─────────────────────────────────────────┘
```

### Data Flow

```
User (natural language)
  ↓
LLM (understands intent)
  ↓
API Wrapper (translates params)
  ↓
CLI Executable (does the work)
  ↓
Output (returned to user)
```

---

## Step-by-Step Tutorial

### Phase 1: Create the CLI Program

#### 1.1 Set Up Directory Structure

```bash
mkdir -p PROGRAMS/your-program/{cmd,config,db,pkg/{app,exitcodes}}
cd PROGRAMS/your-program
```

**Standard structure:**
```
your-program/
├── cmd/
│   ├── command1/
│   │   └── main.go
│   ├── command2/
│   │   └── main.go
│   └── command3/
│       └── main.go
├── config/
│   └── config.go
├── db/
│   └── sqlite.go
├── pkg/
│   ├── app/
│   │   └── app.go
│   ├── exitcodes/
│   │   └── exitcodes.go
│   └── constants/
│       └── constants.go
├── go.mod
├── install.sh
├── uninstall.sh
├── update.sh
└── README.md
```

#### 1.2 Initialize Go Module

```bash
go mod init your-program
```

#### 1.3 Copy Boilerplate Files

**Copy these from task-manager as templates:**

```bash
# Config (handles DB path, env vars)
cp ../task-manager/config/config.go ./config/

# App init (database initialization helper)
cp ../task-manager/pkg/app/app.go ./pkg/app/

# Exit codes (standard error codes)
cp ../task-manager/pkg/exitcodes/exitcodes.go ./pkg/exitcodes/
```

**Modify config.go:**
```go
const (
    defaultDBPath = "~/.local/share/your-program/data.db"
)
```

#### 1.4 Create Database Layer

**File:** `db/sqlite.go`

```go
package db

import (
    "database/sql"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
)

type DB struct {
    conn *sql.DB
}

func New(dbPath string) (*DB, error) {
    conn, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }

    if err := conn.Ping(); err != nil {
        conn.Close()
        return nil, fmt.Errorf("ping database: %w", err)
    }

    // Enable foreign keys
    _, err = conn.Exec("PRAGMA foreign_keys = ON")
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("enable foreign keys: %w", err)
    }

    db := &DB{conn: conn}

    if err := db.initSchema(); err != nil {
        conn.Close()
        return nil, fmt.Errorf("initialize schema: %w", err)
    }

    return db, nil
}

func (db *DB) Close() error {
    return db.conn.Close()
}

func (db *DB) initSchema() error {
    schema := `
    -- Your tables here
    CREATE TABLE IF NOT EXISTS your_table (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    `

    _, err := db.conn.Exec(schema)
    return err
}

// Add your CRUD methods here
```

#### 1.5 Create Command Executables

**File:** `cmd/list/main.go`

```go
package main

import (
    "flag"
    "fmt"
    "os"

    "your-program/pkg/app"
    "your-program/pkg/exitcodes"
)

func main() {
    // Parse flags
    fs := flag.NewFlagSet("list", flag.ExitOnError)
    // Add your flags here

    if err := fs.Parse(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(exitcodes.ArgsError)
    }

    // Initialize database
    database, err := app.InitDatabase()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
        os.Exit(exitcodes.DBError)
    }
    defer database.Close()

    // Your logic here

    os.Exit(exitcodes.Success)
}
```

#### 1.6 Create Installation Script

**File:** `install.sh`

```bash
#!/bin/bash
set -e

echo "=== Your Program - Installation ==="

# Configuration
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="${CONFIG_DIR:-$HOME/.config/your-program}"
DB_PATH="${DB_PATH:-$HOME/.local/share/your-program/data.db}"

# Create directories
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$(dirname "$DB_PATH")"

# Build executables
echo "Building..."
go build -o your_command1 ./cmd/command1
go build -o your_command2 ./cmd/command2

# Install
echo "Installing..."
cp your_command1 "$INSTALL_DIR/"
cp your_command2 "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR"/your_command*

# Create config
cat > "$CONFIG_DIR/.env" <<EOF
DB_PATH=$DB_PATH
EOF

echo "✓ Installation complete!"
```

Make it executable:
```bash
chmod +x install.sh
```

#### 1.7 Test CLI Program

```bash
# Install
./install.sh

# Test commands
your_command1 --help
your_command2 list
```

---

### Phase 2: Create API Wrapper

#### 2.1 Create Program File

**File:** `ORCHESTRATOR/agent-gateway/programs/your_program.go`

```go
package programs

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
)

// YourProgram provides API access to your CLI program
type YourProgram struct{}

func NewYourProgram() *YourProgram {
    return &YourProgram{}
}

// Info returns program metadata shown to the LLM
func (p *YourProgram) Info() ProgramInfo {
    return ProgramInfo{
        ID:          "your-program",
        Name:        "Your Program",
        Description: "What your program does - be descriptive for LLM",
        Category:    "category", // data, tasks, servers, files
        Parameters: []Parameter{
            {
                Name:        "command",
                Type:        "string",
                Description: "Command: 'list', 'add', 'update', 'delete'",
                Required:    true,
            },
            {
                Name:        "name",
                Type:        "string",
                Description: "Item name",
                Required:    false,
            },
            // Add more parameters
        },
    }
}

// Execute translates API calls to CLI commands
func (p *YourProgram) Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error) {
    // Extract command
    command := getStringParam(params, "command")

    var cmd *exec.Cmd

    switch command {
    case "list":
        cmd = exec.CommandContext(ctx, "your_command1", "list")

    case "add":
        name := getStringParam(params, "name")
        if name == "" {
            return &ExecutionResult{
                Success: false,
                Error:   "name required",
            }, fmt.Errorf("missing name")
        }
        cmd = exec.CommandContext(ctx, "your_command2", "add", "--name", name)

    default:
        return &ExecutionResult{
            Success: false,
            Error:   fmt.Sprintf("unknown command: %s", command),
        }, fmt.Errorf("unknown command")
    }

    // Execute
    output, err := cmd.CombinedOutput()
    if err != nil {
        return &ExecutionResult{
            Success: false,
            Output:  string(output),
            Error:   err.Error(),
        }, err
    }

    return &ExecutionResult{
        Success: true,
        Output:  string(output),
    }, nil
}

// Helper functions (copy from task_manager.go)
func getStringParam(params map[string]interface{}, name string) string {
    if val, ok := params[name]; ok {
        if str, ok := val.(string); ok {
            return str
        }
    }
    return ""
}
```

#### 2.2 Register in main.go

**File:** `ORCHESTRATOR/agent-gateway/main.go`

Add after other program registrations:

```go
if err := programsRegistry.Register(programs.NewYourProgram()); err != nil {
    log.Printf("Warning: Failed to register your-program: %v", err)
} else {
    log.Println("Registered program: your-program")
}
```

#### 2.3 Build and Deploy

```bash
cd ORCHESTRATOR/agent-gateway
go build -o agent-gateway .

# On server
./update.sh
```

---

## Best Practices

### CLI Program

1. **Use standard exit codes** (0=success, 1=args, 2=db, 3=not found)
2. **Print to stdout for success**, stderr for errors
3. **Use flags for all parameters** (`--name`, `--id`, etc.)
4. **Support `--help` flag** on all commands
5. **Keep executables focused** - one command per executable
6. **Use wrapper scripts** to load environment variables

### API Wrapper

1. **Normalize user input** (lowercase contexts, trim whitespace)
2. **Validate required parameters** before calling CLI
3. **Provide clear error messages** for the LLM
4. **Document parameters well** - LLM uses descriptions
5. **Use switch statements** for command routing
6. **Format JSON output** for query commands

### Database

1. **Use SQLite** for simplicity
2. **Enable foreign keys** (`PRAGMA foreign_keys = ON`)
3. **Create indexes** for frequently queried columns
4. **Use transactions** for multi-step operations
5. **Handle migrations** if you update schema

### Documentation

1. **Comprehensive README** with examples
2. **Inline comments** for complex logic
3. **Document all exported functions** (godoc)
4. **Troubleshooting section** in README
5. **Architecture diagram** if complex

---

## Testing

### Testing Checklist

**CLI Program:**
- [ ] All commands execute without errors
- [ ] Help messages display correctly
- [ ] Invalid arguments return exit code 1
- [ ] Database errors return exit code 2
- [ ] Not found errors return exit code 3
- [ ] Output is clean and parseable

**API Wrapper:**
- [ ] Program appears in `/api/programs/list`
- [ ] LLM can discover program in context
- [ ] All commands execute via API
- [ ] Error messages are clear
- [ ] JSON output (if any) is valid

**Integration:**
- [ ] Natural language requests work
- [ ] Parameter extraction handles variations
- [ ] Context normalization works (if applicable)
- [ ] Concurrent requests don't conflict

### Test Script Template

```bash
#!/bin/bash
# test.sh

export DB_PATH=/tmp/test-your-program.db
rm -f $DB_PATH

echo "=== Testing Your Program ==="

# Test 1: Create item
your_command add --name "Test"
if [ $? -ne 0 ]; then
    echo "FAIL: Add command"
    exit 1
fi

# Test 2: List items
your_command list | grep "Test"
if [ $? -ne 0 ]; then
    echo "FAIL: List command"
    exit 1
fi

echo "✓ All tests passed"
```

---

## Deployment

### Checklist

**Before deploying:**
- [ ] Code reviewed and commented
- [ ] Tests passing
- [ ] README updated
- [ ] Install/uninstall scripts tested
- [ ] API wrapper tested locally
- [ ] Git committed with clear message

**Deployment steps:**
1. Push code to repository
2. On server: `git pull`
3. Install CLI program: `./install.sh`
4. Build API: `go build`
5. Update service: `./update.sh`
6. Verify in logs: "Registered program: your-program"
7. Test via API: `/api/programs/list`

### Systemd Considerations

If executables are in `~/.local/bin`, ensure systemd service has PATH:

```bash
sudo systemctl edit agent-gateway
```

Add:
```ini
[Service]
Environment="PATH=/home/user/.local/bin:/usr/local/bin:/usr/bin:/bin"
```

---

## Examples

### Example: Simple Note-Taking Program

```
notes/
├── cmd/
│   ├── add/main.go      → notes_add
│   ├── list/main.go     → notes_list
│   └── search/main.go   → notes_search
└── db/sqlite.go

API Wrapper:
- Command: "add" → notes_add --text "..."
- Command: "list" → notes_list --limit 10
- Command: "search" → notes_search --query "..."
```

### Example: Habit Tracker

```
habits/
├── cmd/
│   ├── track/main.go    → habits_track
│   ├── stats/main.go    → habits_stats
│   └── streak/main.go   → habits_streak
└── db/sqlite.go

API Wrapper:
- Command: "track" → habits_track --habit "exercise" --date "2024-11-24"
- Command: "stats" → habits_stats --habit "exercise" --days 30
- Command: "streak" → habits_streak --habit "exercise"
```

### Example: Finance Goals

```
finance-goals/
├── cmd/
│   ├── goal/main.go     → goal_manage
│   ├── progress/main.go → goal_progress
│   └── report/main.go   → goal_report
└── db/sqlite.go

API Wrapper:
- Command: "create" → goal_manage create --name "..." --target 10000
- Command: "update" → goal_manage update --id 1 --current 2500
- Command: "report" → goal_report --format json
```

---

## Quick Reference

### File Templates Location

All templates are in the task-manager program:
- Config: `PROGRAMS/task-manager/config/config.go`
- App Init: `PROGRAMS/task-manager/pkg/app/app.go`
- Exit Codes: `PROGRAMS/task-manager/pkg/exitcodes/exitcodes.go`
- API Wrapper: `ORCHESTRATOR/agent-gateway/programs/task_manager.go`

### Common Commands

```bash
# Create program structure
mkdir -p PROGRAMS/your-program/{cmd,config,db,pkg/{app,exitcodes}}

# Initialize Go module
cd PROGRAMS/your-program && go mod init your-program

# Build API
cd ORCHESTRATOR/agent-gateway && go build

# Test API endpoint
curl -H "X-API-Key: key" http://localhost:8080/api/programs/list
```

---

## Getting Help

- Review existing programs in `PROGRAMS/`
- Check API patterns in `ORCHESTRATOR/agent-gateway/programs/`
- See `CODE_REVIEW_AND_CLEANUP.md` for best practices
- Test thoroughly before deploying

---

**Next:** Ready to build your first program? Start with Phase 1!
