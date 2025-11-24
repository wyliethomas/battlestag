# API Programs Architecture

**Agent Gateway Programs Registry System**

This document explains how the agent-gateway API discovers, registers, and executes external programs, allowing the LLM to interact with them through natural language.

---

## Table of Contents

1. [Overview](#overview)
2. [Core Components](#core-components)
3. [Program Interface](#program-interface)
4. [Registry System](#registry-system)
5. [LLM Context Injection](#llm-context-injection)
6. [Parameter Extraction](#parameter-extraction)
7. [Command Execution](#command-execution)
8. [Error Handling](#error-handling)
9. [Security Considerations](#security-considerations)
10. [Best Practices](#best-practices)
11. [Integration Checklist](#integration-checklist)

---

## Overview

### Architecture Pattern

The agent-gateway uses a **programs registry** pattern that allows external programs to be:
- **Discovered** automatically by the LLM
- **Executed** via a consistent interface
- **Extended** without modifying core API code

```
┌─────────────────────────────────────────────────────┐
│  LLM (Natural Language)                             │
│  "Add a sawmill project to property"                │
└─────────────────┬───────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────────────────┐
│  Agent Gateway API                                  │
│  ┌────────────────────────────────────────────┐    │
│  │  Programs Registry                         │    │
│  │  - task-manager                            │    │
│  │  - echo                                    │    │
│  │  - future programs...                      │    │
│  └────────────────────────────────────────────┘    │
└─────────────────┬───────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────────────────┐
│  Program Wrapper (programs/task_manager.go)         │
│  - Info() → Metadata for LLM                        │
│  - Execute() → Translates params to CLI calls       │
└─────────────────┬───────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────────────────┐
│  CLI Executables (~/.local/bin)                     │
│  - task_project_run                                 │
│  - task_checklist_run                               │
│  - task_journal_run                                 │
│  - task_query_run                                   │
└─────────────────────────────────────────────────────┘
```

### Data Flow

1. **User** sends natural language request to LLM
2. **LLM** receives context about available programs (via buildProgramsContext)
3. **LLM** decides which program to use and extracts parameters
4. **API** routes request to registered program's Execute method
5. **Program Wrapper** translates parameters into CLI command
6. **CLI Executable** performs the actual work
7. **Output** flows back through the chain to the user

---

## Core Components

### File Structure

```
ORCHESTRATOR/agent-gateway/
├── main.go                  # Program registration
├── programs/
│   ├── program.go           # Interface & registry
│   ├── echo.go              # Example program
│   ├── task_manager.go      # Task manager wrapper
│   └── your_program.go      # Your new program
└── DOCS/
    └── API_PROGRAMS_ARCHITECTURE.md  # This file
```

### Key Files

#### programs/program.go (99 lines)
- **Program interface**: Defines contract for all programs
- **ProgramsRegistry**: Manages program registration and lookup
- **buildProgramsContext()**: Generates LLM context with available programs

#### programs/echo.go (72 lines)
- **Minimal example**: Shows simplest possible program
- **Template**: Copy this when creating new programs

#### programs/task_manager.go (371 lines)
- **Full-featured example**: Shows comprehensive parameter handling
- **Pattern to follow**: Command routing, validation, execution

---

## Program Interface

### Interface Definition

**Location**: `programs/program.go` lines 8-14

```go
type Program interface {
    Info() ProgramInfo
    Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error)
}
```

### Required Methods

#### 1. Info() ProgramInfo

Returns metadata about the program that is injected into the LLM's context.

**Structure**:
```go
type ProgramInfo struct {
    ID          string      // Unique identifier (e.g., "task-manager")
    Name        string      // Human-readable name
    Description string      // What the program does (LLM uses this!)
    Category    string      // Grouping (tasks, data, servers, files)
    Parameters  []Parameter // Available parameters
}
```

**Parameter Definition**:
```go
type Parameter struct {
    Name        string // Parameter key (e.g., "command", "project_id")
    Type        string // Data type (string, int, bool)
    Description string // What this parameter does (LLM uses this!)
    Required    bool   // Whether parameter must be provided
}
```

**Example**:
```go
func (p *TaskManagerProgram) Info() ProgramInfo {
    return ProgramInfo{
        ID:          "task-manager",
        Name:        "Task Manager",
        Description: "Hierarchical project management system for tracking long-running tasks across life domains (property, house, pxp, personal). Manages projects with journal logs and checklists. Use this for task tracking, project status, and intelligent work suggestions.",
        Category:    "tasks",
        Parameters: []Parameter{
            {
                Name:        "command",
                Type:        "string",
                Description: "Command to execute: 'list-contexts', 'add-project', 'query-overview', etc.",
                Required:    true,
            },
            {
                Name:        "context",
                Type:        "string",
                Description: "Context name (property, house, pxp, personal)",
                Required:    false,
            },
            // ... more parameters
        },
    }
}
```

**Best Practices for Info()**:
- ✅ **Detailed Description**: LLM uses this to decide when to use the program
- ✅ **Comprehensive Parameters**: Document every parameter with clear descriptions
- ✅ **Mark Required**: Set `Required: true` for mandatory parameters
- ✅ **List Valid Values**: Mention valid options in descriptions (e.g., "status (active, paused, blocked)")

#### 2. Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error)

Executes the program with given parameters.

**Input**:
- `ctx`: Context for cancellation and timeouts
- `params`: Map of parameter names to values (extracted by LLM)

**Output**:
```go
type ExecutionResult struct {
    Success bool   // Whether execution succeeded
    Output  string // Output from command (stdout)
    Error   string // Error message if failed
}
```

**Example Structure**:
```go
func (p *TaskManagerProgram) Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error) {
    // 1. Extract command parameter
    command, ok := params["command"].(string)
    if !ok {
        return &ExecutionResult{
            Success: false,
            Error:   "command parameter must be a string",
        }, fmt.Errorf("invalid command parameter")
    }

    // 2. Route to appropriate handler
    var cmd *exec.Cmd
    switch command {
    case "add-project":
        // Build command with parameters
        contextName := strings.ToLower(getStringParam(params, "context"))
        projectName := getStringParam(params, "project_name")

        // Validate required parameters
        if contextName == "" || projectName == "" {
            return &ExecutionResult{
                Success: false,
                Error:   "context and project_name required",
            }, fmt.Errorf("missing parameters")
        }

        // Build CLI command
        cmd = exec.CommandContext(ctx, "task_project_run", "add", "--name", projectName, "--context", contextName)

    case "list-projects":
        // Handle another command
        cmd = exec.CommandContext(ctx, "task_project_run", "list")

    default:
        return &ExecutionResult{
            Success: false,
            Error:   fmt.Sprintf("unknown command: %s", command),
        }, fmt.Errorf("unknown command")
    }

    // 3. Execute command
    output, err := cmd.CombinedOutput()
    if err != nil {
        return &ExecutionResult{
            Success: false,
            Output:  string(output),
            Error:   err.Error(),
        }, err
    }

    // 4. Return result
    return &ExecutionResult{
        Success: true,
        Output:  string(output),
    }, nil
}
```

---

## Registry System

### Registration Process

**Location**: `main.go` lines 102-106

Programs are registered at application startup:

```go
func main() {
    // ... setup code ...

    // Initialize programs registry
    programsRegistry := programs.NewProgramsRegistry()

    // Register each program
    if err := programsRegistry.Register(programs.NewTaskManagerProgram()); err != nil {
        log.Printf("Warning: Failed to register task-manager: %v", err)
    } else {
        log.Println("Registered program: task-manager")
    }

    if err := programsRegistry.Register(programs.NewEchoProgram()); err != nil {
        log.Printf("Warning: Failed to register echo: %v", err)
    } else {
        log.Println("Registered program: echo")
    }

    // ... more registrations ...
}
```

### Registry Methods

**Location**: `programs/program.go` lines 56-83

```go
type ProgramsRegistry struct {
    programs map[string]Program
    mu       sync.RWMutex
}

// Register adds a new program to the registry
func (r *ProgramsRegistry) Register(program Program) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    info := program.Info()
    if _, exists := r.programs[info.ID]; exists {
        return fmt.Errorf("program %s already registered", info.ID)
    }

    r.programs[info.ID] = program
    return nil
}

// Get retrieves a program by ID
func (r *ProgramsRegistry) Get(id string) (Program, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    program, exists := r.programs[id]
    return program, exists
}

// ListAll returns all registered programs
func (r *ProgramsRegistry) ListAll() []ProgramInfo {
    r.mu.RLock()
    defer r.mu.RUnlock()

    infos := make([]ProgramInfo, 0, len(r.programs))
    for _, program := range r.programs {
        infos = append(infos, program.Info())
    }
    return infos
}
```

**Key Features**:
- **Thread-safe**: Uses mutex for concurrent access
- **Duplicate prevention**: Returns error if program ID already exists
- **Dynamic lookup**: Programs can be retrieved by ID at runtime

---

## LLM Context Injection

### How It Works

The LLM receives information about available programs in every request, allowing it to decide which program to use.

**Location**: `programs/program.go` lines 85-99

```go
func (r *ProgramsRegistry) buildProgramsContext() string {
    programs := r.ListAll()
    if len(programs) == 0 {
        return "No programs available."
    }

    var sb strings.Builder
    sb.WriteString("Available programs:\n\n")

    for _, info := range programs {
        sb.WriteString(fmt.Sprintf("Program: %s\n", info.Name))
        sb.WriteString(fmt.Sprintf("ID: %s\n", info.ID))
        sb.WriteString(fmt.Sprintf("Description: %s\n", info.Description))
        sb.WriteString(fmt.Sprintf("Category: %s\n", info.Category))

        if len(info.Parameters) > 0 {
            sb.WriteString("Parameters:\n")
            for _, param := range info.Parameters {
                req := "optional"
                if param.Required {
                    req = "required"
                }
                sb.WriteString(fmt.Sprintf("  - %s (%s, %s): %s\n",
                    param.Name, param.Type, req, param.Description))
            }
        }
        sb.WriteString("\n")
    }

    return sb.String()
}
```

### Example Context Sent to LLM

```
Available programs:

Program: Task Manager
ID: task-manager
Description: Hierarchical project management system for tracking long-running tasks across life domains (property, house, pxp, personal). Manages projects with journal logs and checklists. Use this for task tracking, project status, and intelligent work suggestions.
Category: tasks
Parameters:
  - command (string, required): Command to execute: 'list-contexts', 'add-project', 'query-overview', etc.
  - context (string, optional): Context name (property, house, pxp, personal)
  - project_id (int, optional): Project ID
  - project_name (string, optional): Project name
  ...

Program: Echo
ID: echo
Description: Simple echo program that returns whatever you send it. Useful for testing.
Category: utility
Parameters:
  - message (string, required): The message to echo back
```

### LLM Decision Process

1. User: "Add a sawmill project to property"
2. LLM reads available programs context
3. LLM identifies "Task Manager" has relevant functionality
4. LLM extracts parameters:
   - `command`: "add-project"
   - `context`: "Property" (capitalized by LLM)
   - `project_name`: "Sawmill"
5. LLM makes API call to execute program

---

## Parameter Extraction

### Common Pattern

Programs receive parameters as `map[string]interface{}` from the LLM, requiring type-safe extraction.

**Location**: `programs/task_manager.go` lines 342-385

### Helper Functions

```go
// getStringParam extracts a string parameter from the params map.
// Returns empty string if parameter doesn't exist or isn't a string.
func getStringParam(params map[string]interface{}, name string) string {
    if val, ok := params[name]; ok {
        if str, ok := val.(string); ok {
            return str
        }
    }
    return ""
}

// getIntParam extracts an integer parameter from the params map.
// Handles float64 (JSON default), int, and string conversions.
// Returns 0 if parameter doesn't exist or can't be converted.
func getIntParam(params map[string]interface{}, name string) int {
    if val, ok := params[name]; ok {
        switch v := val.(type) {
        case float64:
            return int(v)
        case int:
            return v
        case string:
            var i int
            fmt.Sscanf(v, "%d", &i)
            return i
        }
    }
    return 0
}

// getBoolParam extracts a boolean parameter from the params map.
// Returns false if parameter doesn't exist or isn't a boolean.
func getBoolParam(params map[string]interface{}, name string) bool {
    if val, ok := params[name]; ok {
        if b, ok := val.(bool); ok {
            return b
        }
    }
    return false
}
```

### Usage Example

```go
func (p *TaskManagerProgram) Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error) {
    command := getStringParam(params, "command")

    switch command {
    case "add-checklist":
        projectID := getIntParam(params, "project_id")
        items := getStringParam(params, "items")
        pendingOnly := getBoolParam(params, "pending_only")

        if projectID == 0 || items == "" {
            return &ExecutionResult{
                Success: false,
                Error:   "project_id and items required",
            }, fmt.Errorf("missing parameters")
        }

        // Build command...
    }
}
```

### Important Considerations

1. **JSON Number Type**: JSON unmarshals all numbers as `float64`, so getIntParam must handle conversion
2. **Type Safety**: Always validate extracted values before use
3. **Default Values**: Helper functions return zero values if parameter missing
4. **Normalization**: Apply `strings.ToLower()`, `strings.TrimSpace()`, etc. as needed

### Case Sensitivity Pattern

**Critical for user input**: LLMs often capitalize proper nouns (e.g., "Property" instead of "property")

```go
// Always normalize case-sensitive values
contextName := strings.ToLower(getStringParam(params, "context"))
```

Applied to all context-related commands in task-manager:
- add-context
- add-project
- list-projects
- query-overview
- query-suggest
- query-context

---

## Command Execution

### Execution Pattern

**Location**: `programs/task_manager.go` lines 317-340

```go
// Build command
var cmd *exec.Cmd
switch command {
case "add-project":
    cmd = exec.CommandContext(ctx, "task_project_run", "add", "--name", name, "--context", context)
case "list-projects":
    cmd = exec.CommandContext(ctx, "task_project_run", "list")
}

// Execute command
output, err := cmd.CombinedOutput()
outputStr := string(output)

if err != nil {
    return &ExecutionResult{
        Success: false,
        Output:  outputStr,
        Error:   err.Error(),
    }, err
}

return &ExecutionResult{
    Success: true,
    Output:  outputStr,
}, nil
```

### Command Building Best Practices

**1. Use CommandContext for cancellation**:
```go
// Good: Respects context cancellation
cmd := exec.CommandContext(ctx, "task_project_run", "list")

// Bad: Ignores context
cmd := exec.Command("task_project_run", "list")
```

**2. Use CombinedOutput for stderr + stdout**:
```go
// Good: Captures both streams
output, err := cmd.CombinedOutput()

// Acceptable: Separate streams
stdout, _ := cmd.Output()
```

**3. Build arguments dynamically**:
```go
// Good: Flexible argument building
cmdArgs := []string{"add", "--name", name}
if goal != "" {
    cmdArgs = append(cmdArgs, "--goal", goal)
}
cmd := exec.CommandContext(ctx, "task_project_run", cmdArgs...)

// Avoid: Hardcoded optional arguments
cmd := exec.CommandContext(ctx, "task_project_run", "add", "--name", name, "--goal", "")
```

**4. Validate before execution**:
```go
// Good: Validate first
if projectID == 0 {
    return &ExecutionResult{Success: false, Error: "project_id required"}, fmt.Errorf("missing project_id")
}
cmd := exec.CommandContext(ctx, "task_checklist_run", "list", "--project", fmt.Sprintf("%d", projectID))

// Bad: Execute then fail
cmd := exec.CommandContext(ctx, "task_checklist_run", "list", "--project", fmt.Sprintf("%d", projectID))
output, err := cmd.CombinedOutput()
// CLI tool returns exit code 1 instead of clear API error
```

### JSON Output Formatting

For query commands that return JSON, pretty-print for readability:

```go
// Format JSON output
if strings.HasPrefix(command, "query-") {
    if formatted := formatJSON(outputStr); formatted != "" {
        outputStr = formatted
    }
}

func formatJSON(input string) string {
    var data interface{}
    if err := json.Unmarshal([]byte(input), &data); err != nil {
        return ""
    }
    formatted, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return ""
    }
    return string(formatted)
}
```

---

## Error Handling

### Error Levels

**1. Missing Required Parameters**:
```go
if projectID == 0 {
    return &ExecutionResult{
        Success: false,
        Error:   "project_id required for list-checklist",
    }, fmt.Errorf("missing project_id")
}
```

**2. CLI Execution Errors**:
```go
output, err := cmd.CombinedOutput()
if err != nil {
    return &ExecutionResult{
        Success: false,
        Output:  string(output),  // Include output (may contain error details)
        Error:   err.Error(),
    }, err
}
```

**3. Unknown Commands**:
```go
default:
    return &ExecutionResult{
        Success: false,
        Error:   fmt.Sprintf("unknown command: %s", command),
    }, fmt.Errorf("unknown command")
}
```

### Exit Codes from CLI Tools

CLI tools should use consistent exit codes:
- `0`: Success
- `1`: Invalid arguments
- `2`: Database error
- `3`: Resource not found

**Location**: `PROGRAMS/*/pkg/exitcodes/exitcodes.go`

```go
package exitcodes

const (
    Success  = 0 // Command succeeded
    ArgsError = 1 // Invalid arguments
    DBError  = 2 // Database error
    NotFound = 3 // Resource not found
)
```

The wrapper can interpret these exit codes:
```go
output, err := cmd.CombinedOutput()
if err != nil {
    if exitErr, ok := err.(*exec.ExitError); ok {
        switch exitErr.ExitCode() {
        case 1:
            return &ExecutionResult{Success: false, Error: "Invalid arguments"}, err
        case 2:
            return &ExecutionResult{Success: false, Error: "Database error"}, err
        case 3:
            return &ExecutionResult{Success: false, Error: "Not found"}, err
        }
    }
    return &ExecutionResult{Success: false, Error: err.Error()}, err
}
```

### Error Messages for LLM

Make error messages clear and actionable:

```go
// Good: Specific, actionable
"context and project_name required for add-project"

// Bad: Generic, unhelpful
"missing parameters"
```

The LLM reads these error messages and can retry with corrected parameters.

---

## Security Considerations

### 1. Input Validation

**Always validate and sanitize parameters** before passing to CLI:

```go
// Validate parameter format
if !isValidContextName(contextName) {
    return &ExecutionResult{
        Success: false,
        Error:   "invalid context name format",
    }, fmt.Errorf("invalid context")
}

func isValidContextName(name string) bool {
    // Only allow lowercase letters, numbers, hyphens
    matched, _ := regexp.MatchString("^[a-z0-9-]+$", name)
    return matched
}
```

### 2. Command Injection Prevention

**Never use shell execution** with user input:

```go
// DANGEROUS: Shell injection vulnerability
cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("task_project_run add --name %s", name))

// SAFE: Direct command execution
cmd := exec.CommandContext(ctx, "task_project_run", "add", "--name", name)
```

### 3. Path Traversal Prevention

**Validate file paths** if programs accept them:

```go
// Prevent directory traversal
if strings.Contains(filename, "..") {
    return &ExecutionResult{
        Success: false,
        Error:   "invalid filename",
    }, fmt.Errorf("path traversal attempt")
}
```

### 4. Resource Limits

**Use context timeouts** to prevent runaway processes:

```go
// Create timeout context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Execute with timeout
cmd := exec.CommandContext(ctx, "task_query_run", "overview")
output, err := cmd.CombinedOutput()
```

### 5. Least Privilege

**CLI executables should**:
- Run as the service user (not root)
- Only access their own data directories
- Use restrictive file permissions (0644 for data, 0755 for executables)

**Database files should**:
- Be owned by service user
- Have mode 0644 (rw-r--r--)
- Be in user-specific directories (`~/.local/share/`)

### 6. API Key Protection

**Never log API keys**:
```go
// Bad: Logs API key
log.Printf("Request with key: %s", apiKey)

// Good: Redact sensitive data
log.Printf("Request authenticated: %t", apiKey != "")
```

---

## Best Practices

### 1. Program Design

✅ **DO**:
- Keep programs focused on a single domain (task management, finance, etc.)
- Use descriptive program IDs (lowercase-with-hyphens)
- Provide comprehensive parameter descriptions for the LLM
- Return structured JSON for complex queries
- Handle optional parameters gracefully

❌ **DON'T**:
- Create monolithic programs that do everything
- Use cryptic parameter names ("ctx", "proj", "id")
- Assume LLM will provide parameters in a specific format
- Hard-code paths or configuration

### 2. Parameter Handling

✅ **DO**:
- Normalize case for user-facing values (`strings.ToLower()`)
- Trim whitespace (`strings.TrimSpace()`)
- Provide clear validation error messages
- Handle missing optional parameters with sensible defaults
- Support both ID and name lookups when practical

❌ **DON'T**:
- Assume parameters have specific types without checking
- Fail silently on invalid parameters
- Require exact case matching for user input

### 3. Command Execution

✅ **DO**:
- Use `exec.CommandContext()` for cancellation support
- Capture both stdout and stderr with `CombinedOutput()`
- Build arguments dynamically for optional parameters
- Validate all parameters before execution
- Return execution output even on errors (helps debugging)

❌ **DON'T**:
- Use shell execution (`sh -c`)
- Ignore context cancellation
- Hard-code all arguments upfront

### 4. Error Handling

✅ **DO**:
- Return descriptive error messages
- Include command output in error cases
- Use consistent exit codes in CLI tools
- Log errors for debugging (but not sensitive data)

❌ **DON'T**:
- Return generic "error occurred" messages
- Swallow errors silently
- Expose internal system details in errors

### 5. Documentation

✅ **DO**:
- Write detailed parameter descriptions (LLM reads these!)
- Document valid values and formats
- Provide examples in comments
- Explain complex logic with inline comments

❌ **DON'T**:
- Assume parameter names are self-explanatory
- Leave Info() descriptions vague
- Skip commenting complex switch cases

### 6. Testing

✅ **DO**:
- Test with various parameter combinations
- Test missing required parameters
- Test invalid parameter values
- Test concurrent execution
- Verify LLM can discover and use the program

❌ **DON'T**:
- Only test the happy path
- Skip validation testing
- Forget to test edge cases

---

## Integration Checklist

When integrating a new program into the agent-gateway:

### Phase 1: Program Wrapper

- [ ] Create `programs/{program_name}.go`
- [ ] Implement `Program` interface (Info, Execute)
- [ ] Define comprehensive ProgramInfo with all parameters
- [ ] Implement parameter extraction helpers (getStringParam, etc.)
- [ ] Add command routing logic (switch statement)
- [ ] Validate all required parameters
- [ ] Handle optional parameters with defaults
- [ ] Build CLI commands dynamically
- [ ] Execute with CommandContext
- [ ] Handle execution errors properly
- [ ] Format output (JSON pretty-printing if applicable)

### Phase 2: Registration

- [ ] Add import in `main.go`
- [ ] Register program in `main()` function
- [ ] Add registration error handling
- [ ] Add success log message

### Phase 3: Testing

- [ ] Build API: `go build -o agent-gateway .`
- [ ] Start service: `./agent-gateway` or `sudo systemctl restart agent-gateway`
- [ ] Verify registration in logs: "Registered program: {name}"
- [ ] Test `/api/programs/list` endpoint
- [ ] Verify program appears in LLM context
- [ ] Test each command via API
- [ ] Test missing parameter validation
- [ ] Test invalid parameter values
- [ ] Test concurrent requests
- [ ] Verify output format is correct

### Phase 4: Documentation

- [ ] Add inline comments to complex logic
- [ ] Document all exported functions
- [ ] Add examples in comments
- [ ] Update this architecture doc if introducing new patterns
- [ ] Create/update program-specific README

### Phase 5: Deployment

- [ ] Commit changes with clear message
- [ ] Push to repository
- [ ] Pull on server
- [ ] Rebuild API: `go build`
- [ ] Update service: `./update.sh`
- [ ] Verify in production logs
- [ ] Test via production API
- [ ] Monitor for errors

---

## Example: Minimal Program

Here's a complete minimal example for reference:

```go
package programs

import (
    "context"
    "fmt"
    "os/exec"
)

// ExampleProgram demonstrates the minimal Program implementation
type ExampleProgram struct{}

func NewExampleProgram() *ExampleProgram {
    return &ExampleProgram{}
}

// Info returns program metadata for LLM context
func (p *ExampleProgram) Info() ProgramInfo {
    return ProgramInfo{
        ID:          "example",
        Name:        "Example Program",
        Description: "A minimal example program showing the basic pattern. Use this to understand how programs work.",
        Category:    "utility",
        Parameters: []Parameter{
            {
                Name:        "command",
                Type:        "string",
                Description: "Command to execute: 'hello', 'goodbye'",
                Required:    true,
            },
            {
                Name:        "name",
                Type:        "string",
                Description: "Name to use in greeting",
                Required:    false,
            },
        },
    }
}

// Execute runs the program with given parameters
func (p *ExampleProgram) Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error) {
    // Extract command
    command := getStringParam(params, "command")
    if command == "" {
        return &ExecutionResult{
            Success: false,
            Error:   "command parameter required",
        }, fmt.Errorf("missing command")
    }

    // Route to handler
    var cmd *exec.Cmd
    switch command {
    case "hello":
        name := getStringParam(params, "name")
        if name == "" {
            name = "World"
        }
        cmd = exec.CommandContext(ctx, "example_hello", name)

    case "goodbye":
        name := getStringParam(params, "name")
        if name == "" {
            name = "World"
        }
        cmd = exec.CommandContext(ctx, "example_goodbye", name)

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

// Helper: Extract string parameter safely
func getStringParam(params map[string]interface{}, name string) string {
    if val, ok := params[name]; ok {
        if str, ok := val.(string); ok {
            return str
        }
    }
    return ""
}
```

**Registration in main.go**:
```go
if err := programsRegistry.Register(programs.NewExampleProgram()); err != nil {
    log.Printf("Warning: Failed to register example: %v", err)
} else {
    log.Println("Registered program: example")
}
```

---

## Troubleshooting

### Program Not Appearing in /api/programs/list

**Possible causes**:
1. Registration failed (check logs for "Failed to register")
2. Program not imported in main.go
3. API not restarted after code changes

**Solution**:
```bash
# Check logs
sudo journalctl -u agent-gateway -f

# Verify registration
grep "Registered program" /var/log/syslog

# Restart service
sudo systemctl restart agent-gateway
```

### Executables Not Found

**Error**: `exec: "program_name": executable file not found in $PATH`

**Causes**:
1. Executable not installed
2. PATH doesn't include installation directory
3. Systemd service has different PATH than interactive shell

**Solution**:
```bash
# Verify executable exists
which program_name

# Check systemd PATH
systemctl show agent-gateway | grep Environment

# Add to systemd service
sudo systemctl edit agent-gateway
# Add: Environment="PATH=/home/user/.local/bin:..."
sudo systemctl daemon-reload
sudo systemctl restart agent-gateway
```

### LLM Not Using Program

**Possible causes**:
1. Program description too vague
2. Parameters not well documented
3. LLM doesn't understand when to use it

**Solution**:
- Enhance Info() description with specific use cases
- Add examples in parameter descriptions
- Test with explicit requests: "Use the {program} program to..."

### Parameter Extraction Fails

**Causes**:
1. Type assertion without checking
2. Not handling JSON number type (float64)
3. Case sensitivity issues

**Solution**:
- Use helper functions (getStringParam, getIntParam, etc.)
- Always check type assertions with `, ok` pattern
- Normalize case with `strings.ToLower()` for user input

---

## Additional Resources

- **Program Development Guide**: `/DOCS/PROGRAM_DEVELOPMENT_GUIDE.md`
- **Code Review**: `/DOCS/CODE_REVIEW_AND_CLEANUP.md`
- **Example Programs**:
  - Simple: `programs/echo.go`
  - Complex: `programs/task_manager.go`
- **CLI Tool Pattern**: `PROGRAMS/task-manager/`

---

**Next Steps**: Ready to integrate a new program? Follow the [Integration Checklist](#integration-checklist) and refer to existing programs as templates.
