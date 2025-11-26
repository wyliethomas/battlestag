# Agent Gateway Rebuild Plan

**Date:** 2025-11-25
**Status:** Planning Phase
**Goal:** Rebuild agent-gateway as a "dumb orchestrator" using Ollama function calling

---

## Table of Contents

1. [Overview](#overview)
2. [Current vs New Architecture](#current-vs-new-architecture)
3. [Benefits of New Approach](#benefits-of-new-approach)
4. [Implementation Phases](#implementation-phases)
5. [Program Manifest Design](#program-manifest-design)
6. [Migration Strategy](#migration-strategy)
7. [Testing Plan](#testing-plan)
8. [Compatibility with Military Commands](#compatibility-with-military-commands)

---

## Overview

### The Vision

Build a new agent-gateway that:
- **Knows nothing about specific programs** - all program knowledge in manifests
- **Auto-discovers programs** - walk directory, load manifests, register automatically
- **Uses Ollama function calling** - native tool support instead of custom parsing
- **Fails gracefully** - hallucinated tools → conversational fallback
- **Zero config for new programs** - just add a folder with manifest.json

### Core Philosophy

**"Dumb is smart"** - The gateway should be as simple as possible. All intelligence lives in:
1. The LLM (Ollama) - decides when to use tools
2. The programs - know their own capabilities via manifests
3. The client (optional) - can add intent filtering if desired

---

## Current vs New Architecture

### Current System (Complex)

```
┌─────────────────────────────────────────────────────────────┐
│ main.go                                                     │
│  ├─ Manually registers each program (NewTaskManagerProgram) │
│  ├─ Hardcoded program logic in programs/*.go               │
│  ├─ Custom buildProgramsContext() in llm.go                │
│  └─ Custom EXECUTE_PROGRAM parsing                         │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ LLM Handler (llm.go)                                        │
│  ├─ Build context string with all programs                 │
│  ├─ Add 11+ rules to system prompt                         │
│  ├─ Custom parsing: look for EXECUTE_PROGRAM: markers      │
│  └─ Extract params from PARAMETERS: JSON line              │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ Program Wrappers (programs/*.go)                           │
│  ├─ Each program has custom Go wrapper                     │
│  ├─ Parameter extraction (getStringParam, getIntParam)     │
│  ├─ Command building (exec.CommandContext)                 │
│  └─ Output formatting                                      │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ CLI Programs                                                │
│  └─ Actual executables (lab_health, task_project, etc.)    │
└─────────────────────────────────────────────────────────────┘

**Problems:**
- Adding a program requires code changes in multiple places
- System prompts are fragile (LLM ignores rules)
- Custom parsing is error-prone
- No graceful fallback for hallucinations
```

### New System (Simple)

```
┌─────────────────────────────────────────────────────────────┐
│ main.go                                                     │
│  └─ registry := NewRegistry("./programs")                  │
│     (That's it - auto-discovers everything)                │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ Registry (registry.go)                                      │
│  ├─ Walk programs/ directory                               │
│  ├─ Load each manifest.json                                │
│  ├─ Build map: program_name → {manifest, executable_path}  │
│  └─ Provide GetToolDefinitions() for Ollama                │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ Gateway (gateway.go)                                        │
│  ├─ Get tools from registry                                │
│  ├─ Send to Ollama with native function calling            │
│  ├─ Parse Ollama's tool_calls response                     │
│  ├─ Validate tool call against manifest                    │
│  ├─ Execute if valid, fallback to chat if not              │
│  └─ Return response to client                              │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ Programs (self-contained)                                   │
│  programs/task-manager/                                     │
│    ├─ manifest.json  ← All program knowledge here          │
│    └─ task-manager   ← Executable                          │
│  programs/lab-monitor/                                      │
│    ├─ manifest.json                                         │
│    └─ lab-monitor                                           │
└─────────────────────────────────────────────────────────────┘

**Benefits:**
- Adding a program = add folder with manifest.json
- No code changes in gateway
- Native Ollama function calling (proven, reliable)
- Automatic graceful fallback for hallucinations
```

---

## Benefits of New Approach

### 1. Zero-Config Program Addition

**Before:**
```go
// Must modify main.go:
if err := registry.Register(programs.NewLabMonitorProgram()); err != nil {
    log.Printf("Failed to register...")
}

// Must create programs/lab_monitor.go:
type LabMonitorProgram struct{}
func (p *LabMonitorProgram) Info() ProgramInfo { ... }
func (p *LabMonitorProgram) Execute(...) { ... }
// 200+ lines of wrapper code
```

**After:**
```bash
# Just add a folder:
mkdir programs/lab-monitor
echo '{"name": "lab-monitor", "description": "...", ...}' > programs/lab-monitor/manifest.json
cp ~/lab_monitor programs/lab-monitor/lab-monitor
# Done! Gateway auto-discovers it on next restart
```

### 2. Ollama Native Function Calling

**Before:**
- Custom EXECUTE_PROGRAM parsing
- Fragile string matching
- 11+ system prompt rules that LLM ignores
- No standard tool format

**After:**
- Ollama's proven function calling implementation
- Standard OpenAI-compatible tool format
- LLM trained on function calling patterns
- Automatic hallucination detection

### 3. Graceful Hallucination Handling

**Before:**
```
User: "Come up with ideas"
LLM: EXECUTE_PROGRAM: task-manager
Gateway: *tries to execute*
TUI: *validates, finds it exists*
TUI: *executes program*
Result: Wrong program called, confusing output
```

**After:**
```
User: "Come up with ideas"
LLM: tool_call: "brainstorm-helper" (doesn't exist)
Gateway: *tool not in registry*
Gateway: *re-prompts LLM with no tools*
LLM: "Here are some ideas..."
Result: Natural conversation response ✓
```

### 4. Self-Documenting System

**Before:**
- Program capabilities scattered across code
- Must read Go source to understand what programs do
- `GET /programs/list` built manually

**After:**
- `GET /programs` shows all manifests (auto-generated)
- Each program's manifest is single source of truth
- Easy to audit what gateway can do

### 5. Program Portability

**Before:**
- Programs tightly coupled to gateway
- Can't move programs between systems
- Hard to test programs in isolation

**After:**
- Programs are self-contained folders
- Manifest describes complete interface
- Can copy entire folder to another system
- Easy to test: just run the executable

---

## Implementation Phases

### Phase 1: Core Infrastructure (4-6 hours)

**Goal:** Build the foundation without breaking existing system

#### Files to Create

**1. `gateway/types.go`**
```go
package gateway

import "encoding/json"

// Manifest represents a program's manifest.json
type Manifest struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Executable  string          `json:"executable"`  // optional, defaults to name
    Parameters  json.RawMessage `json:"parameters"`  // JSON Schema
}

// Program is a loaded program ready for execution
type Program struct {
    Manifest       Manifest
    ExecutablePath string
}

// OllamaTool is the Ollama function calling format
type OllamaTool struct {
    Type     string `json:"type"`
    Function struct {
        Name        string          `json:"name"`
        Description string          `json:"description"`
        Parameters  json.RawMessage `json:"parameters"`
    } `json:"function"`
}

// OllamaToolCall represents a tool call from Ollama
type OllamaToolCall struct {
    Function struct {
        Name      string          `json:"name"`
        Arguments json.RawMessage `json:"arguments"`
    } `json:"function"`
}
```

**2. `gateway/registry.go`**
```go
package gateway

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
)

type Registry struct {
    programs map[string]Program
}

func NewRegistry(programsDir string) (*Registry, error) {
    r := &Registry{
        programs: make(map[string]Program),
    }

    // Walk the programs directory
    entries, err := ioutil.ReadDir(programsDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read programs dir: %w", err)
    }

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        // Look for manifest.json in each subdirectory
        manifestPath := filepath.Join(programsDir, entry.Name(), "manifest.json")
        if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
            continue // Skip dirs without manifest
        }

        // Load and parse manifest
        data, err := ioutil.ReadFile(manifestPath)
        if err != nil {
            fmt.Printf("Warning: failed to read %s: %v\n", manifestPath, err)
            continue
        }

        var manifest Manifest
        if err := json.Unmarshal(data, &manifest); err != nil {
            fmt.Printf("Warning: invalid manifest %s: %v\n", manifestPath, err)
            continue
        }

        // Resolve executable path
        execName := manifest.Executable
        if execName == "" {
            execName = manifest.Name
        }
        execPath := filepath.Join(programsDir, entry.Name(), execName)

        // Add to registry
        r.programs[manifest.Name] = Program{
            Manifest:       manifest,
            ExecutablePath: execPath,
        }

        fmt.Printf("Loaded program: %s (%s)\n", manifest.Name, execPath)
    }

    return r, nil
}

func (r *Registry) Get(name string) (Program, bool) {
    prog, ok := r.programs[name]
    return prog, ok
}

func (r *Registry) GetToolDefinitions() []OllamaTool {
    tools := make([]OllamaTool, 0, len(r.programs))
    for _, prog := range r.programs {
        tool := OllamaTool{
            Type: "function",
        }
        tool.Function.Name = prog.Manifest.Name
        tool.Function.Description = prog.Manifest.Description
        tool.Function.Parameters = prog.Manifest.Parameters
        tools = append(tools, tool)
    }
    return tools
}

func (r *Registry) List() []Manifest {
    manifests := make([]Manifest, 0, len(r.programs))
    for _, prog := range r.programs {
        manifests = append(manifests, prog.Manifest)
    }
    return manifests
}
```

**3. `gateway/executor.go`**
```go
package gateway

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "time"
)

type ExecutionResult struct {
    Success bool   `json:"success"`
    Output  string `json:"output"`
    Error   string `json:"error,omitempty"`
}

func ExecuteProgram(ctx context.Context, program Program, arguments json.RawMessage) (*ExecutionResult, error) {
    // Create command with timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Convert arguments to command-line args or stdin
    // For simplicity, pass as JSON via stdin
    cmd := exec.CommandContext(ctx, program.ExecutablePath)
    cmd.Stdin = bytes.NewReader(arguments)

    // Capture output
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

**4. `llm/ollama.go`**
```go
package llm

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"

    "agent-gateway/gateway"
)

type Client struct {
    BaseURL string
    Model   string
    Client  *http.Client
}

type Message struct {
    Role      string                   `json:"role"`
    Content   string                   `json:"content"`
    ToolCalls []gateway.OllamaToolCall `json:"tool_calls,omitempty"`
}

type ChatRequest struct {
    Model    string               `json:"model"`
    Messages []Message            `json:"messages"`
    Tools    []gateway.OllamaTool `json:"tools,omitempty"`
    Stream   bool                 `json:"stream"`
}

type ChatResponse struct {
    Message Message `json:"message"`
}

func NewClient(baseURL, model string) *Client {
    return &Client{
        BaseURL: baseURL,
        Model:   model,
        Client:  &http.Client{},
    }
}

func (c *Client) Chat(ctx context.Context, messages []Message, tools []gateway.OllamaTool) (*ChatResponse, error) {
    req := ChatRequest{
        Model:    c.Model,
        Messages: messages,
        Tools:    tools,
        Stream:   false,
    }

    body, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/chat", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.Client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    respBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var chatResp ChatResponse
    if err := json.Unmarshal(respBody, &chatResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w\nBody: %s", err, string(respBody))
    }

    return &chatResp, nil
}
```

**5. `gateway/gateway.go`**
```go
package gateway

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "agent-gateway/llm"
)

type Gateway struct {
    registry *Registry
    llm      *llm.Client
}

func New(registry *Registry, llmClient *llm.Client) *Gateway {
    return &Gateway{
        registry: registry,
        llm:      llmClient,
    }
}

func (g *Gateway) ProcessPrompt(ctx context.Context, userPrompt string) (string, error) {
    // Build conversation with user message
    messages := []llm.Message{
        {Role: "user", Content: userPrompt},
    }

    // Get tools from registry
    tools := g.registry.GetToolDefinitions()

    // Send to LLM with tools
    resp, err := g.llm.Chat(ctx, messages, tools)
    if err != nil {
        return "", fmt.Errorf("LLM error: %w", err)
    }

    // Check if LLM wants to use a tool
    if len(resp.Message.ToolCalls) > 0 {
        toolCall := resp.Message.ToolCalls[0]

        // Validate tool exists
        program, ok := g.registry.Get(toolCall.Function.Name)
        if !ok {
            log.Printf("LLM hallucinated tool: %s - falling back to conversation", toolCall.Function.Name)
            return g.conversationalFallback(ctx, userPrompt)
        }

        // Execute the program
        log.Printf("Executing tool: %s with args: %s", toolCall.Function.Name, string(toolCall.Function.Arguments))
        result, err := ExecuteProgram(ctx, program, toolCall.Function.Arguments)
        if err != nil {
            return "", fmt.Errorf("program execution failed: %w", err)
        }

        if !result.Success {
            return "", fmt.Errorf("program error: %s", result.Error)
        }

        // Send program output back to LLM for natural language formatting
        messages = append(messages, llm.Message{
            Role:      "assistant",
            ToolCalls: resp.Message.ToolCalls,
        })
        messages = append(messages, llm.Message{
            Role:    "tool",
            Content: result.Output,
        })

        finalResp, err := g.llm.Chat(ctx, messages, nil) // No tools on final call
        if err != nil {
            return "", fmt.Errorf("LLM formatting error: %w", err)
        }

        return finalResp.Message.Content, nil
    }

    // No tool call - return conversational response
    return resp.Message.Content, nil
}

func (g *Gateway) conversationalFallback(ctx context.Context, userPrompt string) (string, error) {
    // Re-prompt LLM without tools for pure conversation
    messages := []llm.Message{
        {Role: "user", Content: userPrompt},
    }

    resp, err := g.llm.Chat(ctx, messages, nil) // No tools
    if err != nil {
        return "", err
    }

    return resp.Message.Content, nil
}
```

#### Testing Phase 1

1. Create test programs (echo, time)
2. Load registry and verify programs discovered
3. Test Ollama communication (with/without tools)
4. Test tool execution
5. Test hallucination fallback

**Deliverables:**
- [ ] Core gateway package working
- [ ] Registry auto-discovers programs
- [ ] Ollama integration with function calling
- [ ] Executor runs programs
- [ ] Unit tests for each component

---

### Phase 2: HTTP API (2-3 hours)

**Goal:** Expose gateway via REST API

#### Files to Create

**`api/handlers.go`**
```go
package api

import (
    "encoding/json"
    "net/http"

    "agent-gateway/gateway"
)

type Server struct {
    gateway  *gateway.Gateway
    registry *gateway.Registry
}

func NewServer(gw *gateway.Gateway, reg *gateway.Registry) *Server {
    return &Server{
        gateway:  gw,
        registry: reg,
    }
}

// POST /prompt
func (s *Server) HandlePrompt(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Prompt string `json:"prompt"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    response, err := s.gateway.ProcessPrompt(r.Context(), req.Prompt)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "response": response,
        "success":  true,
    })
}

// GET /programs
func (s *Server) HandleListPrograms(w http.ResponseWriter, r *http.Request) {
    programs := s.registry.List()
    json.NewEncoder(w).Encode(programs)
}

// GET /health
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":   "healthy",
        "programs": len(s.registry.List()),
    })
}
```

**`main.go`**
```go
package main

import (
    "flag"
    "log"
    "net/http"
    "os"

    "agent-gateway/api"
    "agent-gateway/gateway"
    "agent-gateway/llm"
)

func main() {
    programsDir := flag.String("programs", "./programs", "Programs directory")
    ollamaURL := flag.String("ollama", "http://localhost:11434", "Ollama URL")
    model := flag.String("model", "llama3.1:8b", "LLM model")
    port := flag.String("port", "8080", "HTTP port")
    flag.Parse()

    // Load programs registry
    log.Printf("Loading programs from %s...", *programsDir)
    registry, err := gateway.NewRegistry(*programsDir)
    if err != nil {
        log.Fatalf("Failed to load registry: %v", err)
    }
    log.Printf("Loaded %d programs", len(registry.List()))

    // Create LLM client
    llmClient := llm.NewClient(*ollamaURL, *model)

    // Create gateway
    gw := gateway.New(registry, llmClient)

    // Create HTTP server
    server := api.NewServer(gw, registry)

    http.HandleFunc("/prompt", server.HandlePrompt)
    http.HandleFunc("/programs", server.HandleListPrograms)
    http.HandleFunc("/health", server.HandleHealth)

    log.Printf("Starting server on :%s", *port)
    if err := http.ListenAndServe(":"+*port, nil); err != nil {
        log.Fatal(err)
    }
}
```

**Deliverables:**
- [ ] HTTP API endpoints working
- [ ] Integration tests with curl/Postman
- [ ] Error handling for all edge cases

---

### Phase 3: Program Migration (3-4 hours)

**Goal:** Convert existing programs to manifest format

#### Task Manager Migration

**`programs/task-manager/manifest.json`**
```json
{
  "name": "task-manager",
  "description": "Hierarchical project management system for tracking tasks across contexts (property, house, pxp, personal)",
  "executable": "task-manager",
  "parameters": {
    "type": "object",
    "properties": {
      "command": {
        "type": "string",
        "enum": [
          "list-projects",
          "add-project",
          "update-project",
          "list-checklist",
          "add-checklist",
          "check-item",
          "add-journal",
          "list-journal",
          "query-status",
          "query-overview"
        ],
        "description": "The operation to perform"
      },
      "context": {
        "type": "string",
        "enum": ["property", "house", "pxp", "personal"],
        "description": "The context area"
      },
      "project_name": {
        "type": "string",
        "description": "Name of the project"
      },
      "project_id": {
        "type": "integer",
        "description": "Project ID number"
      },
      "goal": {
        "type": "string",
        "description": "Project goal or description"
      }
    },
    "required": ["command"]
  }
}
```

**Create wrapper executable:**
`programs/task-manager/task-manager` (Go program or shell script that reads JSON from stdin and calls the actual CLIs)

#### Lab Monitor Migration

Similar process - create manifest and wrapper executable.

**Deliverables:**
- [ ] Task manager manifest + wrapper
- [ ] Lab monitor manifest + wrapper
- [ ] Test both programs through new gateway
- [ ] Compare behavior with old system

---

### Phase 4: Cutover (1-2 hours)

**Goal:** Replace old agent-gateway with new one

1. Backup old agent-gateway
2. Build new agent-gateway
3. Deploy to remote server
4. Update TUI to use new API format
5. Test thoroughly
6. Remove old code

**Deliverables:**
- [ ] New gateway deployed
- [ ] TUI working with new gateway
- [ ] Old code archived/removed

---

## Program Manifest Design

### JSON Schema Format

Programs use JSON Schema to describe their parameters. This is standard, well-documented, and supported by Ollama.

**Simple Example (no parameters):**
```json
{
  "name": "time",
  "description": "Get current date and time",
  "parameters": {
    "type": "object",
    "properties": {},
    "required": []
  }
}
```

**Complex Example (multiple parameters):**
```json
{
  "name": "task-manager",
  "description": "Manage projects and tasks",
  "parameters": {
    "type": "object",
    "properties": {
      "command": {
        "type": "string",
        "enum": ["list", "add", "update"],
        "description": "Operation to perform"
      },
      "project_name": {
        "type": "string",
        "description": "Name of the project"
      },
      "context": {
        "type": "string",
        "enum": ["property", "house"],
        "description": "Context area"
      }
    },
    "required": ["command"]
  }
}
```

### Executable Interface

Programs can receive arguments in multiple ways:

**Option 1: JSON via stdin** (recommended)
```bash
echo '{"command": "list", "context": "property"}' | ./task-manager
```

**Option 2: Command-line args**
```bash
./task-manager --command list --context property
```

**Option 3: Wrapper script**
Create a wrapper that converts JSON to your existing CLI format:
```bash
#!/bin/bash
# programs/task-manager/task-manager

# Read JSON from stdin
JSON=$(cat)

# Extract fields
COMMAND=$(echo "$JSON" | jq -r '.command')
CONTEXT=$(echo "$JSON" | jq -r '.context // empty')

# Call actual CLI
if [ -n "$CONTEXT" ]; then
  task_project "$COMMAND" --context "$CONTEXT"
else
  task_project "$COMMAND"
fi
```

---

## Migration Strategy

### Parallel Running

1. **Keep old gateway running** during migration
2. **Build new gateway** on port 8081
3. **Test new gateway** thoroughly
4. **Switch TUI** to new gateway
5. **Verify** everything works
6. **Decommission** old gateway

### Risk Mitigation

- Feature flag in TUI to switch between old/new gateway
- Keep old code in `agent-gateway-old/` directory
- Thorough testing before cutover
- Quick rollback procedure

### Testing Checklist

**Functional Tests:**
- [ ] Pure conversation (no tools)
- [ ] Task manager queries
- [ ] Lab monitor health checks
- [ ] Hallucinated tools (graceful fallback)
- [ ] Invalid parameters (validation)
- [ ] Program errors (proper error handling)

**Performance Tests:**
- [ ] Response time acceptable
- [ ] No memory leaks
- [ ] Concurrent requests handled

**Integration Tests:**
- [ ] TUI works end-to-end
- [ ] All existing features work
- [ ] No regressions

---

## Compatibility with Military Commands

The new gateway is **perfectly compatible** with the military command idea!

### How It Works Together

The military command system can be implemented as a **client-side filter** in the TUI:

```
User Input
    ↓
TUI detects military keywords
    ├─ "SITREP" found → set tools filter to ["lab-monitor"]
    ├─ "OPSUM" found → set tools filter to ["task-manager"]
    └─ No keywords → set tools filter to []
    ↓
TUI sends to gateway: {prompt, tools_filter}
    ↓
Gateway filters registry before building tool list
    ↓
Send filtered tools to Ollama
```

**Modifications needed:**

1. Add `tools_filter` parameter to `/prompt` endpoint:
```json
{
  "prompt": "SITREP on teton",
  "tools_filter": ["lab-monitor"]  // Only show this tool to LLM
}
```

2. Gateway uses filter:
```go
func (g *Gateway) ProcessPrompt(ctx context.Context, prompt string, toolsFilter []string) {
    tools := g.registry.GetToolDefinitions()

    // Apply filter if provided
    if len(toolsFilter) > 0 {
        tools = filterTools(tools, toolsFilter)
    }

    // Rest of logic...
}
```

**Benefits:**
- Gateway stays dumb (knows nothing about military commands)
- TUI handles the intent detection (can be swapped out)
- Easy to add other filtering strategies
- Works with or without military commands

---

## Timeline

| Phase | Description | Hours | Dependencies |
|-------|-------------|-------|--------------|
| Phase 1 | Core infrastructure | 4-6 | None |
| Phase 2 | HTTP API | 2-3 | Phase 1 |
| Phase 3 | Program migration | 3-4 | Phase 1, 2 |
| Phase 4 | Cutover & deployment | 1-2 | All |
| **Total** | | **10-15** | |

**Realistic timeline:** 2-3 days of focused work

---

## Success Criteria

1. ✅ Gateway discovers programs automatically
2. ✅ Adding new program requires zero code changes
3. ✅ Ollama function calling works reliably
4. ✅ Hallucinated tools fall back to conversation
5. ✅ All existing functionality preserved
6. ✅ Code is simpler and more maintainable
7. ✅ Compatible with military command filtering

---

## Open Questions

1. Should program wrappers be required, or support direct CLI execution?
2. How do we handle programs that need streaming responses?
3. Should manifests support versioning?
4. Do we want a manifest validator/linter?
5. How do we handle programs that need environment-specific config?

---

## Next Steps

1. **Review this plan** - discuss, refine, approve
2. **Set up development environment** - new directory for rebuild
3. **Implement Phase 1** - build core infrastructure
4. **Test thoroughly** - ensure foundation is solid
5. **Continue with Phase 2-4** - build on solid foundation

---

**Document Version:** 1.0
**Last Updated:** 2025-11-25
**Status:** Ready for review
