# Code Review & Cleanup Plan

**Date**: 2025-11-24
**Scope**: Task-manager program + API integration
**Goal**: Create reusable patterns for future "bolt-on" programs

---

## 📊 Current State Analysis

### Codebase Size
```
Task Manager Program:
- 8 Go files
- 2,141 total lines of code
- Largest file: db/sqlite.go (819 lines)
- 4 executables: project, checklist, journal, query

API Integration:
- 3 program files
- 542 total lines
- program.go: 99 lines (interface/registry)
- echo.go: 72 lines (example)
- task_manager.go: 371 lines (our wrapper)
```

### Architecture Review ✅

**Strengths:**
- Clean separation: CLI tools + API wrapper
- Interface-based design (programs.Program)
- Consistent exit codes
- Good error handling
- JSON output for LLM consumption

**Areas for Improvement:**
- Repetitive parameter extraction in task_manager.go
- No helper functions for common patterns
- Missing inline documentation
- No program development guide

---

## 🔍 Detailed Code Review

### 1. Task Manager Program (PROGRAMS/task-manager)

#### db/sqlite.go (819 lines) - Status: ✅ GOOD
**Strengths:**
- Well-structured with clear method names
- Good separation of concerns
- Comprehensive CRUD operations
- Proper error handling

**Recommendations:**
- ✅ Keep as-is - well organized
- Consider: Add more inline comments for complex queries
- Consider: Extract common SQL patterns if we build more programs

#### cmd/{project,checklist,journal,query}/main.go - Status: ⚠️ NEEDS DRY
**Issues:**
- Repetitive flag parsing boilerplate
- Similar error handling patterns in all 4 files
- Database initialization duplicated

**Fix:** Create shared helper functions

#### config/config.go (55 lines) - Status: ✅ PERFECT
**Why it's good:**
- Simple, focused
- Reusable across programs
- Good environment variable support

**Action:** Use as template for future programs

#### pkg/app/app.go (25 lines) - Status: ✅ PERFECT
**Why it's good:**
- Single responsibility
- DRY database initialization
- Clean interface

**Action:** Pattern to reuse

---

### 2. API Integration (ORCHESTRATOR/agent-gateway/programs)

#### programs/program.go (99 lines) - Status: ✅ EXCELLENT
**Strengths:**
- Clean interface design
- Generic registry pattern
- Good validation
- Extensible

**Action:** Document as THE pattern for future programs

#### programs/echo.go (72 lines) - Status: ✅ PERFECT TEMPLATE
**Why it's good:**
- Minimal, clear example
- Shows all required methods
- Good parameter handling

**Action:** Use as boilerplate for new programs

#### programs/task_manager.go (371 lines) - Status: ⚠️ NEEDS REFACTORING
**Issues Identified:**

1. **Repetitive Parameter Extraction** (Lines 320-380)
```go
// This pattern repeats 14 times:
func getStringParam(params map[string]interface{}, name string) string {
	if val, ok := params[name]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
```

2. **No Command Registry** - Giant switch statement (130 lines)
3. **Context normalization** - Added ad-hoc (should be centralized)
4. **Error messages** - Repetitive patterns

**Recommended Refactoring:**
- Extract parameter helpers to programs/helpers.go
- Create command builder pattern
- Add struct-based command definitions

---

## 🎯 Cleanup Tasks

### High Priority

#### 1. Extract Common Helpers ⚠️
**File to create:** `ORCHESTRATOR/agent-gateway/programs/helpers.go`
```go
package programs

import (
	"fmt"
	"strings"
)

// Parameter extraction helpers (DRY up task_manager.go)
type ParamExtractor struct {
	params map[string]interface{}
}

func NewParamExtractor(params map[string]interface{}) *ParamExtractor {
	return &ParamExtractor{params: params}
}

func (e *ParamExtractor) String(name string) string {
	// ... implementation
}

func (e *ParamExtractor) StringLower(name string) string {
	return strings.ToLower(e.String(name))
}

func (e *ParamExtractor) Int(name string) int {
	// ... implementation
}

func (e *ParamExtractor) Bool(name string) bool {
	// ... implementation
}

func (e *ParamExtractor) Require(name string) error {
	// ... validates required param exists
}
```

#### 2. Create Program Builder Pattern ⚠️
**File to create:** `ORCHESTRATOR/agent-gateway/programs/command_builder.go`
```go
package programs

import (
	"context"
	"os/exec"
)

type CommandBuilder struct {
	bin  string
	args []string
}

func NewCommand(bin string) *CommandBuilder {
	return &CommandBuilder{bin: bin, args: []string{}}
}

func (b *CommandBuilder) Arg(arg string) *CommandBuilder {
	b.args = append(b.args, arg)
	return b
}

func (b *CommandBuilder) Flag(name, value string) *CommandBuilder {
	b.args = append(b.args, "--"+name, value)
	return b
}

func (b *CommandBuilder) Build(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, b.bin, b.args...)
}
```

#### 3. Reduce task_manager.go Complexity ⚠️
**Goal:** Reduce from 371 lines to ~200 lines
**Method:** Use helpers and command builder

**Before:**
```go
case "add-project":
	contextName := strings.ToLower(getStringParam(params, "context"))
	projectName := getStringParam(params, "project_name")
	if contextName == "" || projectName == "" {
		return &ExecutionResult{Success: false, Error: "..."}, fmt.Errorf("...")
	}
	cmdArgs = []string{"add", "--name", projectName, "--context", contextName}
	if goal := getStringParam(params, "goal"); goal != "" {
		cmdArgs = append(cmdArgs, "--goal", goal)
	}
	cmd = exec.CommandContext(ctx, "task_project_run", cmdArgs...)
```

**After:**
```go
case "add-project":
	params := NewParamExtractor(params)
	if err := params.Require("context", "project_name"); err != nil {
		return nil, err
	}

	cmd := NewCommand("task_project_run").
		Arg("add").
		Flag("name", params.String("project_name")).
		Flag("context", params.StringLower("context"))

	if goal := params.String("goal"); goal != "" {
		cmd.Flag("goal", goal)
	}

	return cmd.Build(ctx)
```

---

### Medium Priority

#### 4. Add Inline Documentation 📝
**Files needing docs:**
- `programs/task_manager.go` - Add command descriptions
- `cmd/*/main.go` - Document flag usage
- `db/sqlite.go` - Complex query explanations

#### 5. Create Shared Constants ⚠️
**File:** `PROGRAMS/task-manager/pkg/constants/constants.go`
```go
package constants

const (
	DefaultDBPath = "~/.local/share/task-manager/tasks.db"
	DefaultConfigDir = "~/.config/task-manager/"
)

// Exit codes (already in exitcodes package - good!)

// Status values
const (
	StatusActive   = "active"
	StatusPaused   = "paused"
	StatusBlocked  = "blocked"
	StatusComplete = "complete"
)
```

#### 6. Standardize Script Templates ⚠️
**Files:** `install.sh`, `update.sh`, `uninstall.sh`
- All 3 scripts have similar patterns
- Extract common functions to `scripts/common.sh`

---

### Low Priority

#### 7. Add Unit Tests 🧪
**Test files to create:**
- `db/sqlite_test.go`
- `config/config_test.go`
- `programs/task_manager_test.go`

#### 8. Add Integration Tests 🧪
**File:** `PROGRAMS/task-manager/test_integration.sh`
- Test full workflow: create context → project → checklist → query

---

## 📋 Program Development Blueprint

### Template: New Bolt-On Program

When creating a new program (e.g., `habit-tracker`, `finance-goals`), follow this pattern:

#### Step 1: Program Structure
```
PROGRAMS/{program-name}/
├── cmd/
│   ├── {command1}/main.go
│   ├── {command2}/main.go
│   └── {command3}/main.go
├── config/
│   └── config.go              ← Copy from task-manager
├── db/
│   └── sqlite.go              ← Your data model
├── pkg/
│   ├── app/app.go             ← Copy from task-manager
│   └── exitcodes/exitcodes.go ← Copy from task-manager
├── go.mod
├── go.sum
├── install.sh                  ← Template below
├── uninstall.sh                ← Template below
├── update.sh                   ← Template below
└── README.md
```

#### Step 2: API Wrapper
```
ORCHESTRATOR/agent-gateway/programs/
└── {program_name}.go
```

**Template:**
```go
package programs

import (
	"context"
	"os/exec"
)

type {Name}Program struct{}

func New{Name}Program() *{Name}Program {
	return &{Name}Program{}
}

func (p *{Name}Program) Info() ProgramInfo {
	return ProgramInfo{
		ID:          "{program-id}",
		Name:        "{Program Name}",
		Description: "What this program does...",
		Category:    "category",
		Parameters:  []Parameter{
			// Define parameters
		},
	}
}

func (p *{Name}Program) Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error) {
	extractor := NewParamExtractor(params)
	command := extractor.String("command")

	switch command {
	case "command1":
		// Use CommandBuilder pattern
	case "command2":
		// ...
	}
}
```

#### Step 3: Register in main.go
```go
if err := programsRegistry.Register(programs.New{Name}Program()); err != nil {
	log.Printf("Warning: Failed to register {program}: %v", err)
} else {
	log.Println("Registered program: {program}")
}
```

---

## 🚀 Refactoring Plan

### Phase 1: Extract Helpers (Today)
1. Create `programs/helpers.go`
2. Create `programs/command_builder.go`
3. Update imports in `task_manager.go`
4. Test compilation

**Estimated Time:** 2 hours
**Risk:** Low (additive changes)

### Phase 2: Refactor task_manager.go (Today)
1. Replace parameter extraction with ParamExtractor
2. Replace cmd building with CommandBuilder
3. Test all commands still work
4. Update tests

**Estimated Time:** 3 hours
**Risk:** Medium (changing working code)

### Phase 3: Documentation (Today)
1. Add inline comments to complex code
2. Create PROGRAM_DEVELOPMENT_GUIDE.md
3. Document testing procedures
4. Add examples

**Estimated Time:** 2 hours
**Risk:** None (documentation only)

### Phase 4: Script Templates (Optional)
1. Extract common bash functions
2. Create script templates
3. Update existing scripts

**Estimated Time:** 1 hour
**Risk:** Low

---

## 📚 Documentation To Create

### 1. PROGRAM_DEVELOPMENT_GUIDE.md
**Location:** `/DOCS/`
**Contents:**
- Step-by-step guide to create new programs
- Code templates and examples
- Testing checklist
- Integration steps
- Best practices

### 2. API_PROGRAMS_ARCHITECTURE.md
**Location:** `/ORCHESTRATOR/agent-gateway/DOCS/`
**Contents:**
- How the programs registry works
- Interface requirements
- LLM context injection
- Error handling patterns
- Security considerations

### 3. Inline Documentation
**Files needing godoc comments:**
- All exported types and functions
- Complex algorithms
- Non-obvious design decisions

---

## ✅ Success Criteria

### Code Quality
- [ ] No code duplication (DRY)
- [ ] All functions < 50 lines
- [ ] Clear, descriptive names
- [ ] Inline documentation for complex logic
- [ ] Consistent error handling

### Maintainability
- [ ] New programs can be added in < 1 day
- [ ] Clear patterns to follow
- [ ] Comprehensive documentation
- [ ] Test coverage for critical paths

### Architecture
- [ ] Programs are loosely coupled
- [ ] Shared logic is extracted
- [ ] Clear interfaces
- [ ] Easy to extend

---

## 🎯 Today's Goals

1. ✅ Complete code review (this document)
2. ⏳ Create helper utilities
3. ⏳ Refactor task_manager.go
4. ⏳ Create program development guide
5. ⏳ Add comprehensive documentation

**Total Estimated Time:** 6-8 hours
**Priority:** High (foundation for scaling)

---

## 📝 Notes

### Design Principles
1. **DRY**: Don't Repeat Yourself
2. **KISS**: Keep It Simple, Stupid
3. **YAGNI**: You Aren't Gonna Need It (don't over-engineer)
4. **Separation of Concerns**: Each component has one job
5. **Interface Segregation**: Small, focused interfaces

### Future Considerations
- Consider gRPC for program communication (if performance matters)
- Consider plugin architecture (if we get 10+ programs)
- Consider centralized logging
- Consider metrics/monitoring
- Consider rate limiting per program

---

## 🔄 Review Schedule

- **Weekly**: Review new code for DRY violations
- **Monthly**: Check for outdated patterns
- **Quarterly**: Major refactoring if needed

---

**Next Steps:** Begin Phase 1 - Extract Helpers
