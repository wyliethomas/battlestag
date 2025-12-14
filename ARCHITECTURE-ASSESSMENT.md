# BATTLESTAG Platform Architecture Assessment

**Date**: 2025-12-13
**Purpose**: Evaluate current architecture for AI-powered development assistant expansion

---

## 📊 Current State Analysis

### 1. battlestag-tui (Bubbletea TUI)

**What You Have** ✅:
- Clean MVC architecture using Bubbletea framework
- ~2,700 lines of well-organized Go code
- Multiple view modes: Chat, Command Palette, Settings, Programs, Onboarding
- Command system with categories (`finance:*`, `stoic:*`, `tech:*`, `system:*`)
- API client for agent-gateway communication
- Configuration management
- Connection health checks

**Strengths**:
- ✅ Professional, polished interface
- ✅ Modular component structure
- ✅ Command palette pattern (already supports deep navigation)
- ✅ Settings and onboarding flows
- ✅ Good separation of concerns

**What to Keep**:
- ✅ **KEEP EVERYTHING** - This is solid foundation
- Command registry system in `ui/commands.go`
- Chat interface with LLM integration
- Settings management
- Connection handling

**What Needs Extension**:
- ⚠️ Command registry needs to scale to hundreds of commands
- ⚠️ Command palette needs better filtering/search for deep hierarchies
- ⚠️ Need file picker/browser for AI commands that work with files

---

### 2. agent-gateway (REST API)

**What You Have** ✅:
- Go/Gorilla Mux REST API (~9K LOC in agent-gateway binary)
- API key authentication
- CORS support
- Logging (access + error logs)
- Health checks
- Systemd service integration
- Database manager for SQLite backends
- Executor for running external programs
- Programs registry system (extensible!)
- LLM integration (Ollama)

**Current Handlers**:
- `handlers/meta.go` - Health/stats
- `handlers/stoic.go` - Stoic thoughts
- `handlers/tech.go` - Tech tips
- `handlers/financial_*.go` - Finance features (4 handlers)
- `handlers/llm.go` - LLM chat with command suggestions
- `handlers/programs.go` - Extensible programs

**Strengths**:
- ✅ Production-ready (systemd service, logging, graceful shutdown)
- ✅ **Programs registry is brilliant** - exactly what we need!
- ✅ CommandBuilder pattern for external programs
- ✅ Already has LLM integration
- ✅ Clean handler separation

**What to Keep**:
- ✅ **KEEP EVERYTHING** - This is your solid API foundation
- Programs registry pattern (`programs/program.go`)
- CommandBuilder for shell command execution
- Handler structure
- Middleware (auth, logging, CORS)

**What Needs Extension**:
- ⚠️ Add new programs for AI dev tools
- ⚠️ May need category/namespace support in registry
- ⚠️ Need script execution wrapper program

---

### 3. Current Integration Points

**TUI → API Flow**:
```
User input in TUI
  ↓
Command detected (e.g., "finance:dashboard")
  ↓
TUI calls client.Client methods
  ↓
HTTP request to agent-gateway (http://battlestag:8080)
  ↓
Handler processes request
  ↓
May call external program via executor
  ↓
JSON response back to TUI
  ↓
TUI displays result
```

**Programs System**:
```
programs.Registry
  ├─ task-manager (Go program calling bash scripts)
  ├─ lab-monitor (Go program calling bash scripts)
  └─ [Future: AI tools as programs]
```

---

## 🎯 Proposed Architecture for AI Dev Tools

### Scripts Folder Structure

```
~/scripts/
├── ai/                    # AI-powered development tools
│   ├── code-review
│   ├── test-gen
│   ├── doc-gen
│   ├── security-scan
│   ├── refactor
│   └── explain
│
├── finance/               # Financial automation
│   ├── statement-parser
│   ├── budget-analyzer
│   └── tax-calculator
│
├── health/                # Health & wellness
│   ├── meal-planner
│   ├── workout-tracker
│   └── sleep-analyzer
│
├── automotive/            # Car maintenance
│   ├── maintenance-reminder
│   ├── fuel-tracker
│   └── repair-log
│
├── business/              # Business tools
│   ├── invoice-gen
│   ├── proposal-writer
│   └── meeting-notes
│
└── lib/                   # Shared libraries
    ├── llm-helper.sh      # LLM API wrapper
    └── common.sh          # Shared utilities
```

### API Gateway Extension

**New Program Type**: `ScriptProgram`

```go
// programs/script_program.go
type ScriptProgram struct {
    category string
    name     string
    scriptPath string
    description string
}

// Auto-register all scripts in ~/scripts/
func RegisterScriptsFromDirectory(registry *Registry, baseDir string)
```

**New Handler**: `handlers/ai.go`

```go
// Unified AI handler that routes to appropriate scripts
func HandleAICommand(w http.ResponseWriter, r *http.Request) {
    // Parse request
    // Call script via CommandBuilder
    // Stream results back
}
```

### TUI Command Structure

**New Command Category**:
```go
{
    Name: "🤖 AI Dev",
    Key:  "ai",
    Description: "AI-powered development tools",
}

// Sub-commands auto-generated from ~/scripts/ai/
ai:review
ai:test
ai:doc
ai:security
ai:refactor
ai:explain
```

**Auto-discovery**:
- TUI calls `/api/programs/list` on startup
- API scans `~/scripts/` and returns available commands
- TUI dynamically builds command registry

---

## ✅ KEEP vs ❌ THROW AWAY

### KEEP ✅

**TUI (100%)**:
- ✅ All UI components
- ✅ Command palette system
- ✅ Chat interface
- ✅ Settings management
- ✅ Client/API integration
- ✅ Bubbletea framework choice

**API Gateway (100%)**:
- ✅ All handlers
- ✅ Programs registry
- ✅ CommandBuilder
- ✅ Middleware
- ✅ Database manager
- ✅ Executor
- ✅ LLM client

**Why Keep Everything?**
- Architecture is sound
- Code quality is good
- Already handles extensibility (programs registry)
- Just needs new programs registered

### THROW AWAY ❌

**Nothing!**

The architecture is solid. We only need to **extend**, not replace.

---

## 🚀 Implementation Plan

### Phase 1: Scripts Foundation
1. Create `~/scripts/` directory structure
2. Move existing LLM helpers to `~/scripts/lib/`
3. Create AI dev tools in `~/scripts/ai/`
4. Make all scripts follow consistent interface

### Phase 2: API Extension
1. Create `programs/script_program.go` - generic script wrapper
2. Add script auto-discovery on startup
3. New endpoint: `GET /api/scripts/categories` (list all categories)
4. New endpoint: `GET /api/scripts/:category` (list scripts in category)
5. New endpoint: `POST /api/scripts/:category/:name` (execute script)

### Phase 3: TUI Enhancement
1. Extend command registry to support dynamic commands
2. Add file picker component for AI commands
3. Enhance command palette with better search/filtering
4. Add progress indicators for long-running AI tasks

### Phase 4: Client Packaging
1. Create deployment script
2. Docker containers for easy deployment
3. Configuration templates
4. Client customization guide

---

## 📝 Command Navigation Design Ideas

### Current: 2-Level Hierarchy
```
finance → dashboard, assets, liabilities, upload, query
stoic   → today, random
tech    → random, latest
system  → programs, settings, clear, help, quit
```

### Proposed: Multi-Level with Hundreds of Commands

**Option A: Breadcrumb Navigation**
```
┌─ Command Palette ─────────────────────────┐
│ ai > code >                                │
├───────────────────────────────────────────┤
│ ▸ review          Review code for bugs    │
│ ▸ test            Generate unit tests      │
│ ▸ security        Security audit           │
│ ▸ refactor        Improve code quality     │
│ ▸ doc             Generate documentation   │
└───────────────────────────────────────────┘
```

**Option B: Fuzzy Search**
```
┌─ Command Palette ─────────────────────────┐
│ Search: rev                                │
├───────────────────────────────────────────┤
│ ▸ ai:code-review       Review code         │
│ ▸ finance:review-asset Asset review        │
│ ▸ git:revert          Revert changes       │
└───────────────────────────────────────────┘
```

**Option C: Tree View (Expandable)**
```
┌─ Command Palette ─────────────────────────┐
│ ▾ ai                                       │
│   ▾ code                                   │
│     ▸ review                               │
│     ▸ test                                 │
│   ▾ docs                                   │
│     ▸ generate                             │
│ ▾ finance                                  │
│   ▸ dashboard                              │
└───────────────────────────────────────────┘
```

**Option D: Recent + Favorites + Search**
```
┌─ Command Palette ─────────────────────────┐
│ ⭐ Favorites                               │
│ ▸ ai:code-review                           │
│ ▸ finance:dashboard                        │
│                                            │
│ 🕐 Recent                                  │
│ ▸ ai:test                                  │
│ ▸ stoic:today                              │
│                                            │
│ 🔍 Search: _                               │
└───────────────────────────────────────────┘
```

**Recommendation**: **Hybrid Approach**
- Fuzzy search (primary)
- Recent commands (quick access)
- Breadcrumb navigation (when drilling down)
- Favorites (power users)

---

## 🎨 File Operations for AI Commands

Many AI commands need file input. Need UI for this:

**File Picker Component**:
```
┌─ Select File ─────────────────────────────┐
│ 📁 ~/projects/myapp/                       │
├───────────────────────────────────────────┤
│ 📁 src/                                    │
│ 📁 tests/                                  │
│ 📄 main.rb                                 │
│ 📄 config.yml                              │
│ 📄 README.md                               │
│                                            │
│ Selected: main.rb                          │
│ [Cancel] [Confirm]                         │
└───────────────────────────────────────────┘
```

**Or Prompt for Path**:
```
┌─ AI: Code Review ─────────────────────────┐
│ File path: src/auth.rb_                    │
│                                            │
│ Or paste code:                             │
│ ┌─────────────────────────────────────┐   │
│ │                                     │   │
│ │                                     │   │
│ └─────────────────────────────────────┘   │
│ [Cancel] [Review]                          │
└───────────────────────────────────────────┘
```

---

## 💡 Key Insights

1. **Your architecture is excellent** - Don't rebuild, extend
2. **Programs registry is the key** - Use it for script management
3. **TUI command system scales** - Just needs better search/navigation
4. **Script-based approach** - Perfect for client customization
5. **Modular design** - Easy to add categories for clients

---

## 🎯 Next Steps

1. **Design final scripts structure** - Get your approval
2. **Build script_program.go** - Generic script wrapper
3. **Enhance command palette** - Better navigation for deep hierarchies
4. **Add file picker** - For AI commands
5. **Test end-to-end** - TUI → API → Script → LLM → Results

**Ready to proceed?**
