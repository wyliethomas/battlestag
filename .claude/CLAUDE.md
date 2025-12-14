# Claude Context - BATTLESTAG-BOT

This file provides context for Claude Code sessions.

## Project Overview
BATTLESTAG-BOT is an AI-powered personal assistant platform with TUI interface. It integrates local LLMs (8B models like Hermes-3) for private, secure AI assistance across multiple domains (finance, health, development, business).

**Key Goal**: Build a platform that works with small local models for client deployments where privacy/compliance matters (HIPAA, GDPR, etc.)

## Tech Stack
- **TUI**: Go + Bubbletea framework (~2.7K LOC)
- **API**: Go + Gorilla Mux REST API (agent-gateway)
- **LLM**: LM Studio serving Hermes-3-Llama-3.1-8B (4.92GB, Q4_K_M)
- **Scripts**: Bash scripts for AI tools
- **Database**: SQLite for various agents (stoic thoughts, tech tips, finance)
- **Deployment**: Systemd services, Tailscale VPN
- **Infrastructure**: Home lab (battlestag + teton servers)

## Architecture
```
Laptop (TUI) → Tailscale → battlestag:8080 (agent-gateway)
                              ↓
                         ~/scripts/ (AI tools)
                              ↓
                         teton:1234 (LM Studio + Hermes-3)
```

## Key Patterns & Conventions

### Programs Registry Pattern
The agent-gateway uses a **programs registry** system for extensibility:
- Each program implements `Program` interface
- Programs registered at startup in `main.go`
- CommandBuilder pattern for shell command execution
- See: `ORCHESTRATOR/agent-gateway/programs/`

### TUI Command Structure
Commands use category:action format:
- `finance:dashboard`, `finance:assets`
- `stoic:today`, `stoic:random`
- `ai:review`, `ai:test` (planned)

Defined in: `APPS/battlestag-tui/ui/commands.go`

### API Handlers
Each domain has its own handler:
- `handlers/financial_*.go` - Finance features
- `handlers/stoic.go` - Stoic thoughts
- `handlers/llm.go` - LLM chat integration
- Future: `handlers/ai.go` - AI dev tools

### Scripts Organization
Scripts are organized by category in `~/scripts/`:
- `ai/` - AI development tools (code-review, test-gen, etc.)
- `finance/` - Financial automation
- `health/` - Health & wellness
- `automotive/` - Car maintenance
- `business/` - Business tools
- `lib/` - Shared utilities (llm-helper.sh, common.sh)

## Important Context

### LLM Constraints
- Using **8B models** (Hermes-3-Llama-3.1-8B) on 8GB GPU
- Context limit: **8-16k tokens** (not enough for OpenCode-style tool frameworks)
- Must use **minimal prompts** - direct API calls work best
- LM Studio endpoint: `http://192.168.1.232:1234/v1`

### Client Deployment Model
This platform will be sold to clients who need:
- **100% private AI** - data never leaves their infrastructure
- **Compliance-friendly** - HIPAA, GDPR, SOC2
- **Customizable** - different script categories per client
- **Cost-effective** - one-time cost, no per-token charges

### Network Setup
- **battlestag**: Main server (runs agent-gateway, scripts)
- **teton**: GPU server (RTX 2060 Super, runs LM Studio)
- Connected via local network + Tailscale VPN
- TUI on laptop connects to battlestag:8080

## Project Structure
```
BATTLESTAG-BOT/
├── APPS/
│   └── battlestag-tui/        # TUI application (Go + Bubbletea)
├── ORCHESTRATOR/
│   └── agent-gateway/         # REST API (Go + Gorilla Mux)
├── PROGRAMS/                  # Standalone programs
├── DOCS/                      # Documentation
├── DEPLOYMENT/                # Deployment scripts
├── CLAUDE.md                  # Main project context (root)
└── ARCHITECTURE-ASSESSMENT.md # Recent architecture review

~/scripts/                     # AI tools (not in repo)
├── ai/                        # AI dev tools
├── finance/                   # Financial tools
├── lib/                       # Shared utilities
└── lms-model-manager.sh       # LM Studio model management
```

## Current Phase
**Designing extensible AI tools architecture**
- Reviewed existing TUI and API gateway (both are solid - keep 100%)
- Identified programs registry as perfect extensibility point
- Designing scripts folder structure for client customization
- Planning TUI navigation for deep command hierarchies

## Recent Decisions (see .claude/DECISIONS.md)
1. Use local 8B LLM with lightweight scripts (not heavy frameworks)
2. Keep all existing architecture - extend don't replace
3. Modular scripts structure by category

## Helper Scripts
Scripts in `scripts/` are reusable helpers. **Before writing repetitive bash commands:**
1. Check if a script already exists in `scripts/`
2. If not, consider creating one for sequences you'll run again

This reduces permission prompts and ensures consistency.

## Tips for Working on This Project
- **TUI changes**: Edit `APPS/battlestag-tui/ui/*.go`
- **API changes**: Edit `ORCHESTRATOR/agent-gateway/handlers/*.go`
- **New programs**: Add to `ORCHESTRATOR/agent-gateway/programs/`
- **Testing API**: Use `test-api.sh` in agent-gateway directory
- **LLM testing**: Use `~/scripts/llm-helper.sh ask "question"`
