# Program Response Design Pattern

**Definitive guide for handling LLM-driven program execution**

Date: 2025-11-24
Status: Canonical Reference

---

## Table of Contents

1. [Overview](#overview)
2. [The Problem We Solved](#the-problem-we-solved)
3. [The Ideal Flow](#the-ideal-flow)
4. [Implementation Details](#implementation-details)
5. [Common Pitfalls](#common-pitfalls)
6. [Testing Checklist](#testing-checklist)
7. [Future Programs](#future-programs)

---

## Overview

This document defines the **canonical pattern** for how programs integrate with the LLM system, ensuring consistent behavior across all current and future programs.

### Key Principles

1. **Single Execution**: One user request = One LLM response = One program execution (maximum)
2. **Clear Attribution**: User knows who said what (LLM vs Program)
3. **Proper Error Handling**: Errors appear as program errors, not LLM messages
4. **No Hallucination**: LLM doesn't claim success until program confirms it

---

## The Problem We Solved

### What Was Happening

When a user requested: "Add a new project called Solar Array to property"

**Bad Flow**:
1. User sends message
2. LLM responds: "I've added a new project..." ← **Hallucination! Not added yet!**
3. Program executes with bad params
4. Program fails: "Error: missing parameters"
5. Error shows as SECOND LLM message ← **Wrong! Should be program error**
6. User sees TWO "Battlestag" messages (confusing)
7. Solar Array never actually created

### Root Causes Identified

1. **LLM Hallucination**: LLM said "I've added" before program executed
2. **Wrong Message Type**: Program failures returned as `ChatResponseMsg` (showed as "Battlestag")
3. **Unclear System Prompt**: LLM wasn't clear about format requirements
4. **Parameter Extraction**: Potential issues with JSON parsing

---

## The Ideal Flow

### Successful Execution

```
User: "Add Solar Array project to property"
  ↓
LLM receives request + programs context
  ↓
LLM generates:
  EXECUTE_PROGRAM: task-manager
  PARAMETERS: {"command": "add-project", "project_name": "Solar Array", "context": "property"}
  ---
  I'm adding a new project called "Solar Array" to your property context.
  ↓
API Parser extracts:
  - programID: "task-manager"
  - params: {command, project_name, context}
  - cleanMessage: "I'm adding..."
  ↓
TUI receives ChatResponseMsg:
  - Shows: "Battlestag • 15:26"
  - Message: "I'm adding..." ← Present tense, not claiming success yet
  - Has programID → Execute program
  ↓
TUI executes program with params
  ↓
Program succeeds, returns output
  ↓
TUI shows: "Task Manager • 15:26"
  - Output: "Project 'Solar Array' created with ID 4" ← Actual confirmation
  ↓
User sees:
  ✅ Battlestag said what it's doing
  ✅ Task Manager confirmed it worked
```

### Failed Execution

```
User: "Add Solar Array project to property"
  ↓
LLM generates program execution
  ↓
TUI shows: "Battlestag • 15:26"
  - "I'm adding a new project..." ← Intent, not claim of success
  ↓
TUI executes program
  ↓
Program fails: "missing parameters"
  ↓
TUI shows: "Task Manager (Error) • 15:26"
  - "❌ Program execution failed: missing parameters" ← Clear it's a program error
  ↓
User sees:
  ✅ Battlestag explained intent
  ✅ Task Manager showed the error (not Battlestag)
  ✅ Clear the program failed, not the LLM
```

---

## Implementation Details

### 1. System Prompt Format

**Location**: `/ORCHESTRATOR/agent-gateway/handlers/llm.go` lines 175-198

**Key Requirements**:
```
EXECUTE_PROGRAM: program_id
PARAMETERS: {"param1": "value1"}
---
Natural language explanation
```

**Critical Rules for LLM**:
1. **Only ONE** EXECUTE_PROGRAM block per response
2. **All required params** must be in PARAMETERS line
3. **Valid JSON** on a single line
4. **Explanation AFTER** the --- marker
5. **Present tense**, not past tense ("I'm adding" not "I've added")
6. **Don't claim success** until program confirms

### 2. Parameter Parsing

**Location**: `/ORCHESTRATOR/agent-gateway/handlers/llm.go` lines 196-240

**Parser Behavior**:
- Scans line-by-line
- Extracts first `EXECUTE_PROGRAM:` occurrence
- Extracts first `PARAMETERS:` occurrence
- Stops at first `---`
- Returns everything after `---` as cleanMessage
- **Only parses ONE program execution**

**Limitations**:
- Won't detect multiple EXECUTE_PROGRAM blocks
- Won't validate JSON structure (just attempts unmarshal)
- Won't validate that all required params are present

### 3. Error Handling in TUI

**Location**: `/APPS/battlestag-tui/ui/app.go` lines 441-447

**Fixed Pattern**:
```go
if !result.Success {
    // Return as program error, not as assistant message
    return ProgramResponseMsg{
        programName: programName + " (Error)",
        output:      fmt.Sprintf("❌ Program execution failed: %s", result.Error),
    }
}
```

**Why This Matters**:
- Program errors show as "Task Manager (Error)" not "Battlestag"
- User knows the program failed, not the LLM
- Error attribution is clear

### 4. Logging for Debugging

**Location**: `/ORCHESTRATOR/agent-gateway/handlers/programs.go` lines 72-89

**What We Log**:
```
[PROGRAM] Executing: task-manager with params: {...}
[PROGRAM] Success task-manager
  OR
[PROGRAM] Failed task-manager: missing parameters
  OR
[PROGRAM] Error executing task-manager: network error
```

**Location**: `/ORCHESTRATOR/agent-gateway/handlers/llm.go` lines 84-89

**What We Log**:
```
[LLM] Detected program: task-manager with params: {...}
[LLM] Clean message: I'm adding a new project...
```

**Purpose**:
- Trace execution flow
- Debug parameter issues
- Identify LLM prompt problems
- Monitor success/failure rates

---

## Common Pitfalls

### Pitfall 1: LLM Claims Success Before Execution

**Problem**: LLM says "I've added..." before program runs

**Symptom**: User sees success message, then error

**Fix**: System prompt rule #6: "Do NOT say you've completed the task until you get the actual program result"

**Example**:
```
❌ Bad:  "I've added Solar Array to your projects"
✅ Good: "I'm adding Solar Array to your projects"
```

### Pitfall 2: Errors Show as LLM Messages

**Problem**: Program failures returned as `ChatResponseMsg`

**Symptom**: Two "Battlestag" messages appear

**Fix**: Return `ProgramResponseMsg` for failures (lines 441-447 in app.go)

**Example**:
```
❌ Bad:  Battlestag: "Error: missing parameters"
✅ Good: Task Manager (Error): "❌ Program execution failed: missing parameters"
```

### Pitfall 3: Multiple Program Executions

**Problem**: LLM tries to execute program multiple times in one response

**Symptom**: Duplicate operations, confusing output

**Fix**: System prompt rule #1: "Use ONLY ONE EXECUTE_PROGRAM block per response"

**Parser Behavior**: Parser only extracts FIRST execution block

### Pitfall 4: Malformed JSON Parameters

**Problem**: LLM generates invalid JSON or multiline JSON

**Symptom**: Parameters not extracted, program gets empty params

**Fix**: System prompt rule #3: "The PARAMETERS must be valid JSON on a SINGLE line"

**Example**:
```
❌ Bad:  PARAMETERS: {
           "command": "add-project",
           "name": "Test"
         }
✅ Good: PARAMETERS: {"command": "add-project", "name": "Test"}
```

### Pitfall 5: Missing Required Parameters

**Problem**: LLM doesn't include all required parameters

**Symptom**: Program fails with validation error

**Fix**: System prompt rule #2: "ALL required parameters MUST be included"

**Program Validation**: All program wrappers should validate required params

**Example**:
```go
if contextName == "" || projectName == "" {
    return &ExecutionResult{
        Success: false,
        Error:   "context and project_name required for add-project",
    }, fmt.Errorf("missing parameters")
}
```

---

## Testing Checklist

### For Every New Program

- [ ] **System Prompt**
  - [ ] Program listed in buildProgramsContext()
  - [ ] All parameters documented with clear descriptions
  - [ ] Required parameters marked as `[REQUIRED]`

- [ ] **Parameter Extraction**
  - [ ] All parameter types handled (string, int, bool)
  - [ ] Required parameters validated
  - [ ] Clear error messages for missing params

- [ ] **Execution Flow**
  - [ ] Single user request = single program execution
  - [ ] Success returns clean output
  - [ ] Failure returns descriptive error

- [ ] **TUI Display**
  - [ ] LLM message appears as "Battlestag"
  - [ ] Program output appears as "Program Name"
  - [ ] Program errors appear as "Program Name (Error)"
  - [ ] No duplicate messages

- [ ] **LLM Behavior**
  - [ ] LLM uses present tense ("I'm doing" not "I've done")
  - [ ] LLM doesn't claim success prematurely
  - [ ] LLM generates valid JSON on one line
  - [ ] LLM includes all required parameters

### Test Scenarios

**Scenario 1: Successful Operation**
```
User: "Add project X to context Y"
Expected:
  1. Battlestag: "I'm adding project X..."
  2. Program Name: "Project created with ID 5"
```

**Scenario 2: Missing Parameters**
```
User: "Add project X" (missing context)
Expected:
  1. Battlestag: "I'm adding project X..."
  2. Program Name (Error): "❌ Program execution failed: context required"
```

**Scenario 3: Database Error**
```
User: "Add project X to Y"
Expected:
  1. Battlestag: "I'm adding project X..."
  2. Program Name (Error): "❌ Program execution failed: database error: ..."
```

**Scenario 4: Query Operation**
```
User: "Show all projects"
Expected:
  1. Battlestag: "Here are your projects..."
  2. Program Name: (JSON output or formatted list)
```

---

## Future Programs

### When Creating a New Program

Follow this pattern for all future programs:

#### 1. Program Wrapper Validation

```go
func (p *YourProgram) Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error) {
    command := getStringParam(params, "command")

    switch command {
    case "your-command":
        // Extract parameters
        requiredParam := getStringParam(params, "required_param")

        // Validate BEFORE execution
        if requiredParam == "" {
            return &ExecutionResult{
                Success: false,
                Error:   "required_param is required for your-command",
            }, fmt.Errorf("missing required parameter")
        }

        // Build and execute command
        cmd := exec.CommandContext(ctx, "your_executable", "--param", requiredParam)
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
}
```

#### 2. System Prompt Entry

Make descriptions clear and LLM-friendly:

```go
{
    Name:        "command",
    Type:        "string",
    Description: "Command to execute: 'list', 'add', 'update', 'delete'",
    Required:    true,
},
{
    Name:        "required_param",
    Type:        "string",
    Description: "The thing you need for this to work",
    Required:    true,
},
```

#### 3. CLI Tool Exit Codes

Use standard exit codes:

```go
const (
    Success  = 0 // Operation succeeded
    ArgsErr  = 1 // Invalid arguments
    DBError  = 2 // Database error
    NotFound = 3 // Resource not found
)
```

#### 4. Output Format

**For success**: Clean, parseable output
```
Project 'Workshop' created with ID 5
Goal: Organize tools
Last touched: 2025-11-24 15:30
```

**For errors**: Clear error message
```
Error: context 'invalid' does not exist
Available contexts: property, house, pxp, personal
```

**For queries**: Structured data (JSON or formatted list)
```json
{
  "projects": [
    {"id": 3, "name": "Sawmill", "status": "active", "progress": 65},
    {"id": 2, "name": "Workshop", "status": "active", "progress": 30}
  ]
}
```

---

## Summary

### The Golden Rules

1. ✅ **One request, one execution** - Never duplicate program calls
2. ✅ **Present tense, not past** - "I'm doing" not "I've done"
3. ✅ **Errors are program errors** - Not LLM messages
4. ✅ **Validate parameters** - Before execution, not after
5. ✅ **Log everything** - For debugging and monitoring
6. ✅ **Clear attribution** - User knows who said what

### The Pattern at a Glance

```
User Request
  ↓
LLM: "I'm doing X..."  (present tense intent)
  ↓
Execute Program
  ↓
If Success:
  Program: "X completed successfully" (confirmation)
If Failure:
  Program (Error): "❌ Failed because Y" (clear error)
```

### Files Modified for This Pattern

1. `/ORCHESTRATOR/agent-gateway/handlers/llm.go`
   - Improved system prompt (lines 175-198)
   - Added logging (lines 84-89)

2. `/ORCHESTRATOR/agent-gateway/handlers/programs.go`
   - Added logging (lines 72-89)

3. `/APPS/battlestag-tui/ui/app.go`
   - Fixed error handling (lines 441-447)
   - Changed `ChatResponseMsg` to `ProgramResponseMsg` for failures

---

**This pattern is now the standard for all programs in the Battlestag ecosystem.**

When building new programs, reference this document to ensure consistent, predictable behavior.
