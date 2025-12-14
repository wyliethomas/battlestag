# Decisions Log - BATTLESTAG-BOT

## 2025-12-13: Use Local 8B LLM Instead of Heavy OpenCode Framework

**Decision:** Build lightweight script-based AI tools using local 8B models (Hermes-3) instead of using OpenCode's heavy tool-calling framework.

**Rationale:**
- OpenCode requires ~17k+ tokens just for system prompts and tool definitions
- Our 8B model (Hermes-3-Llama-3.1-8B) running on RTX 2060 Super (8GB VRAM) can only handle 8-16k context
- Direct API calls to LM Studio with minimal prompts work perfectly
- Script-based approach is more maintainable and customizable per client

**Alternatives Considered:**
1. Use cloud models (Claude/GPT-4) - Rejected due to privacy/cost concerns for client use
2. Upgrade to larger GPU - Too expensive, not scalable for clients
3. Use Ollama instead of LM Studio - Still evaluating, but both work

**Impact:**
- Need to build micro-service architecture with bash scripts
- TUI will use command-based interface (slash commands) instead of natural language tool selection
- Easier to customize for different clients (finance, healthcare, etc.)
- 100% private/secure - data never leaves local network

---

## 2025-12-13: Keep All Existing Architecture - Extend Don't Replace

**Decision:** Keep both battlestag-tui and agent-gateway as-is, only extending with new programs.

**Rationale:**
- TUI is well-architected with Bubbletea, ~2.7K LOC, professional interface
- agent-gateway is production-ready with programs registry system
- Programs registry is exactly what we need for AI tools
- No need to rebuild what already works

**Alternatives Considered:**
1. Rewrite TUI with different framework - Unnecessary, current works well
2. Replace agent-gateway with simpler API - Loses valuable features
3. Build separate AI service - Would duplicate functionality

**Impact:**
- Faster development - just add new programs
- Proven, stable foundation
- Easy to maintain
- Clients get battle-tested platform

---

## 2025-12-13: Modular Scripts Structure by Category

**Decision:** Organize scripts in category folders (ai/, finance/, health/, automotive/, business/) under ~/scripts/

**Rationale:**
- Different clients need different categories
- Easy to add/remove categories per deployment
- Clear organization for hundreds of potential commands
- Matches mental model of different domains

**Alternatives Considered:**
1. Flat structure with prefixes - Gets messy with many scripts
2. Single scripts directory - Hard to manage client customization
3. Database-driven commands - Over-engineered

**Impact:**
- Each client deployment can have custom category mix
- Easy to share scripts between clients (common lib/)
- Clear namespace for commands (ai:review, finance:analyze)
- Scales to hundreds of scripts per category

---
