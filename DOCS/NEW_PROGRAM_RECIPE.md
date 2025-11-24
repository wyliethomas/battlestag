# New Program Recipe

**Quick-start guide for AI assistants building new Battlestag programs**

---

## Purpose

This is a **copy-paste recipe** for creating new programs in the Battlestag ecosystem. Follow this step-by-step, referencing the task-manager as the canonical example.

**Use this when**: Building any new "bolt-on" program (habits, finance-goals, notes, etc.)

---

## Prerequisites

Before starting, verify:
- [ ] You have access to `/home/battlestag/Work/BATTLESTAG-BOT/`
- [ ] Task-manager exists at `PROGRAMS/task-manager/`
- [ ] Agent-gateway exists at `ORCHESTRATOR/agent-gateway/`

---

## The Recipe

### Phase 1: CLI Program (30-60 minutes)

#### Step 1.1: Create Directory Structure

```bash
cd ~/Work/BATTLESTAG-BOT/PROGRAMS
mkdir -p NEW_PROGRAM_NAME/{cmd,config,db,pkg/{app,exitcodes,constants}}
cd NEW_PROGRAM_NAME
go mod init NEW_PROGRAM_NAME
```

Replace `NEW_PROGRAM_NAME` with your program name (e.g., `habit-tracker`, `finance-goals`)

#### Step 1.2: Copy Boilerplate Files

```bash
# From task-manager, copy these exactly
cp ../task-manager/config/config.go ./config/
cp ../task-manager/pkg/app/app.go ./pkg/app/
cp ../task-manager/pkg/exitcodes/exitcodes.go ./pkg/exitcodes/

# Edit config.go to change database path
# Change: defaultDBPath = "~/.local/share/NEW_PROGRAM_NAME/data.db"
```

#### Step 1.3: Create Database Layer

**File**: `db/sqlite.go`

**Template** (copy from task-manager, modify schema):

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
    -- YOUR SCHEMA HERE
    CREATE TABLE IF NOT EXISTS your_table (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_your_table_name ON your_table(name);
    `

    _, err := db.conn.Exec(schema)
    return err
}

// Add your CRUD methods here
// Examples: AddItem(), GetItem(), ListItems(), UpdateItem(), DeleteItem()
```

#### Step 1.4: Create Command Executables

**Pattern**: One executable per command group

**Example**: `cmd/manage/main.go`

```go
package main

import (
    "flag"
    "fmt"
    "os"

    "NEW_PROGRAM_NAME/pkg/app"
    "NEW_PROGRAM_NAME/pkg/exitcodes"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintf(os.Stderr, "Usage: %s <command> [flags]\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "Commands: add, list, update, delete\n")
        os.Exit(exitcodes.ArgsError)
    }

    command := os.Args[1]

    switch command {
    case "add":
        handleAdd(os.Args[2:])
    case "list":
        handleList(os.Args[2:])
    case "update":
        handleUpdate(os.Args[2:])
    case "delete":
        handleDelete(os.Args[2:])
    default:
        fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
        os.Exit(exitcodes.ArgsError)
    }
}

func handleAdd(args []string) {
    fs := flag.NewFlagSet("add", flag.ExitOnError)
    name := fs.String("name", "", "Name of the item")
    // Add more flags

    if err := fs.Parse(args); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(exitcodes.ArgsError)
    }

    if *name == "" {
        fmt.Fprintf(os.Stderr, "Error: --name is required\n")
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
    fmt.Printf("Added: %s\n", *name)
    os.Exit(exitcodes.Success)
}

// Implement handleList, handleUpdate, handleDelete...
```

#### Step 1.5: Create Install Script

**File**: `install.sh`

```bash
#!/bin/bash
set -e

echo "=== NEW_PROGRAM_NAME Installation ==="

# Configuration
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="${CONFIG_DIR:-$HOME/.config/NEW_PROGRAM_NAME}"
DB_PATH="${DB_PATH:-$HOME/.local/share/NEW_PROGRAM_NAME/data.db}"

# Create directories
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$(dirname "$DB_PATH")"

# Build executables
echo "Building..."
go build -o program_command1 ./cmd/command1
go build -o program_command2 ./cmd/command2

# Install
echo "Installing to $INSTALL_DIR..."
cp program_command1 "$INSTALL_DIR/"
cp program_command2 "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR"/program_*

# Create config
cat > "$CONFIG_DIR/.env" <<EOF
DB_PATH=$DB_PATH
EOF

echo "✓ Installation complete!"
echo ""
echo "Executables installed:"
echo "  - $INSTALL_DIR/program_command1"
echo "  - $INSTALL_DIR/program_command2"
echo ""
echo "Configuration: $CONFIG_DIR/.env"
echo "Database: $DB_PATH"
```

Make executable:
```bash
chmod +x install.sh
```

#### Step 1.6: Test CLI Program

```bash
# Install
./install.sh

# Test help
program_command1 --help

# Test basic operation
export DB_PATH=/tmp/test-program.db
program_command1 add --name "Test"
program_command1 list
rm /tmp/test-program.db
```

**Checkpoint**: CLI program must work standalone before API integration.

---

### Phase 2: API Integration (20-30 minutes)

#### Step 2.1: Create Program Wrapper

**File**: `ORCHESTRATOR/agent-gateway/programs/your_program.go`

**Template**:

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

// Info returns program metadata for LLM
func (p *YourProgram) Info() ProgramInfo {
    return ProgramInfo{
        ID:          "your-program",  // lowercase-with-hyphens
        Name:        "Your Program",  // Display name
        Description: "Detailed description of what this program does. Be specific - the LLM uses this to decide when to use the program.",
        Category:    "category",      // tasks, data, tracking, etc.
        Parameters: []Parameter{
            {
                Name:        "command",
                Type:        "string",
                Description: "Command to execute: 'list', 'add', 'update', 'delete'",
                Required:    true,
            },
            {
                Name:        "name",
                Type:        "string",
                Description: "Name of the item",
                Required:    false,  // Mark true if always required
            },
            // Add all your parameters
        },
    }
}

// Execute runs the program
func (p *YourProgram) Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error) {
    // Extract command
    command := getStringParam(params, "command")
    if command == "" {
        return &ExecutionResult{
            Success: false,
            Error:   "command parameter required",
        }, fmt.Errorf("missing command")
    }

    var cmd *exec.Cmd

    switch command {
    case "list":
        cmd = exec.CommandContext(ctx, "program_command1", "list")

    case "add":
        name := getStringParam(params, "name")
        if name == "" {
            return &ExecutionResult{
                Success: false,
                Error:   "name required for add command",
            }, fmt.Errorf("missing name")
        }
        cmd = exec.CommandContext(ctx, "program_command1", "add", "--name", name)

    case "update":
        // Similar pattern...
        cmd = exec.CommandContext(ctx, "program_command1", "update", /* ... */)

    case "delete":
        // Similar pattern...
        cmd = exec.CommandContext(ctx, "program_command1", "delete", /* ... */)

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
```

**Key Points**:
- Use `getStringParam()`, `getIntParam()`, `getBoolParam()` (already defined in helpers.go)
- Validate required parameters BEFORE executing
- Use `exec.CommandContext()` for cancellation support
- Return descriptive errors

#### Step 2.2: Register Program

**File**: `ORCHESTRATOR/agent-gateway/main.go`

Add after other registrations:

```go
// Register your-program
if err := programsRegistry.Register(programs.NewYourProgram()); err != nil {
    log.Printf("Warning: Failed to register your-program: %v", err)
} else {
    log.Println("Registered program: your-program")
}
```

#### Step 2.3: Build and Test API

```bash
cd ~/Work/BATTLESTAG-BOT/ORCHESTRATOR/agent-gateway
go build -o agent-gateway .

# Test programs list
curl -H "X-API-Key: your-key" http://localhost:8080/api/programs/list | jq

# Test execution
curl -X POST \
  -H "X-API-Key: your-key" \
  -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"list"}}' \
  http://localhost:8080/api/programs/execute | jq
```

---

### Phase 3: Deployment (10 minutes)

#### Step 3.1: Commit Changes

```bash
cd ~/Work/BATTLESTAG-BOT

git add PROGRAMS/NEW_PROGRAM_NAME
git add ORCHESTRATOR/agent-gateway/programs/your_program.go
git add ORCHESTRATOR/agent-gateway/main.go

git commit -m "Add NEW_PROGRAM_NAME program

- CLI program with add/list/update/delete commands
- SQLite database with schema
- API wrapper for LLM integration
- Installation script"

git push
```

#### Step 3.2: Deploy to Remote

```bash
# On remote server
ssh user@192.168.1.140
cd /path/to/BATTLESTAG-BOT

# Pull changes
git pull

# Install CLI program
cd PROGRAMS/NEW_PROGRAM_NAME
./install.sh

# Rebuild API
cd ../../ORCHESTRATOR/agent-gateway
go build -o agent-gateway .
sudo ./update.sh

# Verify registration
sudo journalctl -u agent-gateway -n 20 | grep "Registered program"
```

#### Step 3.3: Test in TUI

```bash
# On local machine
cd ~/Work/BATTLESTAG-BOT/APPS/battlestag-tui
./battlestag
```

Test prompts:
- "Show me all [items] in [your program]"
- "Add a new [item] called [name]"
- "Update [item] with [changes]"

**Expected behavior**:
1. Battlestag: "Let me get your [items]..." or "I'm adding..."
2. Your Program: [Actual output/confirmation]

---

## Validation Checklist

### CLI Program
- [ ] Compiles without errors
- [ ] All commands work standalone
- [ ] Help messages display correctly
- [ ] Exit codes correct (0=success, 1=args, 2=db, 3=notfound)
- [ ] Database schema initializes
- [ ] install.sh completes successfully

### API Wrapper
- [ ] Compiles without errors
- [ ] Appears in `/api/programs/list`
- [ ] All commands execute via API
- [ ] Required parameters validated
- [ ] Error messages clear

### LLM Integration
- [ ] TUI can discover program
- [ ] LLM understands when to use it
- [ ] LLM extracts parameters correctly
- [ ] Battlestag doesn't hallucinate data
- [ ] Program provides actual data/confirmation

---

## Reference Examples

**Use task-manager as your template**:

```
PROGRAMS/task-manager/
├── cmd/
│   ├── project/main.go       ← Pattern for command handling
│   ├── checklist/main.go     ← Pattern for subcommands
│   ├── journal/main.go       ← Pattern for logging
│   └── query/main.go         ← Pattern for JSON output
├── db/sqlite.go              ← Pattern for database layer
├── config/config.go          ← Copy this exactly
├── pkg/
│   ├── app/app.go           ← Copy this exactly
│   ├── exitcodes/exitcodes.go ← Copy this exactly
│   └── constants/constants.go ← Pattern for your constants
└── install.sh                ← Pattern for installation

ORCHESTRATOR/agent-gateway/programs/task_manager.go
                             ← Pattern for API wrapper
```

---

## Common Patterns

### Database CRUD

```go
// Add
func (db *DB) AddItem(name string) (int64, error) {
    result, err := db.conn.Exec(
        "INSERT INTO items (name) VALUES (?)",
        name,
    )
    if err != nil {
        return 0, err
    }
    return result.LastInsertId()
}

// Get
func (db *DB) GetItem(id int64) (*Item, error) {
    var item Item
    err := db.conn.QueryRow(
        "SELECT id, name, created_at FROM items WHERE id = ?",
        id,
    ).Scan(&item.ID, &item.Name, &item.CreatedAt)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("item not found")
    }
    if err != nil {
        return nil, err
    }
    return &item, nil
}

// List
func (db *DB) ListItems() ([]*Item, error) {
    rows, err := db.conn.Query(
        "SELECT id, name, created_at FROM items ORDER BY created_at DESC",
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []*Item
    for rows.Next() {
        var item Item
        if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt); err != nil {
            return nil, err
        }
        items = append(items, &item)
    }
    return items, rows.Err()
}

// Update
func (db *DB) UpdateItem(id int64, name string) error {
    result, err := db.conn.Exec(
        "UPDATE items SET name = ? WHERE id = ?",
        name, id,
    )
    if err != nil {
        return err
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return fmt.Errorf("item not found")
    }
    return nil
}

// Delete
func (db *DB) DeleteItem(id int64) error {
    result, err := db.conn.Exec("DELETE FROM items WHERE id = ?", id)
    if err != nil {
        return err
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return fmt.Errorf("item not found")
    }
    return nil
}
```

### Command Handler Pattern

```go
func handleCommand(args []string) {
    fs := flag.NewFlagSet("command", flag.ExitOnError)

    // Define flags
    flag1 := fs.String("flag1", "", "Description")
    flag2 := fs.Int("flag2", 0, "Description")

    // Parse
    if err := fs.Parse(args); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(exitcodes.ArgsError)
    }

    // Validate
    if *flag1 == "" {
        fmt.Fprintf(os.Stderr, "Error: --flag1 required\n")
        os.Exit(exitcodes.ArgsError)
    }

    // Initialize database
    database, err := app.InitDatabase()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
        os.Exit(exitcodes.DBError)
    }
    defer database.Close()

    // Execute operation
    result, err := database.DoSomething(*flag1, *flag2)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(exitcodes.DBError)
    }

    // Output result
    fmt.Printf("Success: %v\n", result)
    os.Exit(exitcodes.Success)
}
```

### API Parameter Validation

```go
case "your-command":
    // Extract parameters
    param1 := getStringParam(params, "param1")
    param2 := getIntParam(params, "param2")

    // Validate required parameters
    if param1 == "" {
        return &ExecutionResult{
            Success: false,
            Error:   "param1 required for your-command",
        }, fmt.Errorf("missing param1")
    }

    // Build command with all parameters
    cmdArgs := []string{"your-command", "--param1", param1}
    if param2 > 0 {
        cmdArgs = append(cmdArgs, "--param2", fmt.Sprintf("%d", param2))
    }

    cmd = exec.CommandContext(ctx, "your_executable", cmdArgs...)
```

---

## Pro Tips

### 1. Start Small
Build the minimal viable program first:
- One table
- Add and List commands only
- Basic API wrapper
Then iterate with more features.

### 2. Test Early, Test Often
- Test CLI standalone before API
- Test API before LLM integration
- Test each command individually

### 3. Follow Task-Manager Patterns
When in doubt, look at task-manager:
- Database patterns → `task-manager/db/sqlite.go`
- Command patterns → `task-manager/cmd/*/main.go`
- API patterns → `agent-gateway/programs/task_manager.go`

### 4. Use Standard Exit Codes
```go
exitcodes.Success   // 0 - Everything worked
exitcodes.ArgsError // 1 - Invalid arguments
exitcodes.DBError   // 2 - Database problem
exitcodes.NotFound  // 3 - Resource doesn't exist
```

### 5. Validate Before Execute
Always validate parameters in the API wrapper BEFORE calling the CLI:
```go
if requiredParam == "" {
    return &ExecutionResult{Success: false, Error: "param required"}, ...
}
```

### 6. Keep LLM Descriptions Clear
In `ProgramInfo`, write descriptions that help the LLM understand:
```go
Description: "Task management for long-running projects with journal logs and checklists. Use this when the user wants to track ongoing work, not for simple todos."
```

### 7. Log Everything During Development
Add logging to see what's happening:
```go
log.Printf("[YOUR_PROGRAM] Executing: %s with params: %+v", command, params)
```

---

## Troubleshooting

### CLI program works, API doesn't find executable
**Fix**: Check PATH in systemd service
```bash
sudo systemctl edit agent-gateway
# Add: Environment="PATH=/home/user/.local/bin:..."
```

### LLM doesn't use the program
**Fix**: Make description more specific in `Info()`

### Parameters not extracted correctly
**Fix**: Check parameter names match exactly between `Info()` and `Execute()`

### Database errors
**Fix**: Check database path and permissions
```bash
ls -la ~/.local/share/your-program/
chmod 644 ~/.local/share/your-program/data.db
```

---

## Quick Commands Reference

```bash
# Create new program structure
mkdir -p PROGRAMS/new-prog/{cmd,config,db,pkg/{app,exitcodes,constants}}

# Copy boilerplate
cp PROGRAMS/task-manager/config/config.go PROGRAMS/new-prog/config/
cp PROGRAMS/task-manager/pkg/app/app.go PROGRAMS/new-prog/pkg/app/
cp PROGRAMS/task-manager/pkg/exitcodes/exitcodes.go PROGRAMS/new-prog/pkg/exitcodes/

# Build CLI
cd PROGRAMS/new-prog && go build -o new_prog ./cmd/main

# Build API
cd ORCHESTRATOR/agent-gateway && go build -o agent-gateway .

# Test API
curl -H "X-API-Key: key" http://localhost:8080/api/programs/list | jq
```

---

## Success Criteria

Your new program is ready when:

✅ CLI works standalone
✅ API wrapper compiles and registers
✅ Program appears in `/api/programs/list`
✅ Can execute via curl
✅ TUI can discover and use it
✅ LLM provides intent, program provides data
✅ Errors show as program errors, not LLM confusion

---

**Estimated Time**: 1-2 hours for a basic program following this recipe

**Next Steps**: Once the program works, add more features iteratively, always testing after each addition.

---

## Additional Resources

- **Detailed Guide**: `PROGRAM_DEVELOPMENT_GUIDE.md`
- **Architecture**: `API_PROGRAMS_ARCHITECTURE.md`
- **Response Pattern**: `PROGRAM_RESPONSE_PATTERN.md`
- **Testing**: `TESTING_CHECKLIST.md`

---

**This recipe is designed to be given to any AI assistant in any future session to maintain consistency.**
