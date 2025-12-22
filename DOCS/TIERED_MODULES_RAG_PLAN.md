# Tiered Modules + RAG Architecture Plan

**Date**: 2025-12-20
**Status**: Planning
**Goal**: Build tiered AI modules (patterns) with Obsidian + RAG integration

---

## 🎯 Vision

Create a **knowledge accumulation system** where:
1. **Tier 2 modules** (big brain/Claude) extract high-value insights
2. **Outputs saved to Obsidian** (version controlled, human-readable)
3. **Ingested into RAG** (vector database for semantic retrieval)
4. **Tier 1 modules** (small brain/Hermes-3) use RAG for enhanced context
5. **System gets smarter over time** without retraining

---

## 🏗️ Three-Tier Module System

### **Tier 1: Small Brain (Hermes-3-8B)**

**Characteristics:**
- Uses local LLM (teton:1234)
- Free, unlimited usage
- Fast execution (< 5 seconds)
- Good for pattern matching, formatting, simple extraction
- **Can be RAG-enhanced** for better context

**Use Cases:**
- Text formatting/improvement
- Simple code formatting
- Basic data extraction
- Todo extraction from code
- File summarization

**Model:** `hermes-3-llama-3.1-8b` (local)
**Cost:** $0.00 per call

### **Tier 2: Big Brain (Claude Sonnet 4.5)**

**Characteristics:**
- Uses cloud API (Anthropic)
- Paid, but affordable (~$0.01-0.05 per call)
- Slower execution (5-15 seconds)
- Excellent for synthesis, deep analysis, creative tasks
- **Outputs enrich the knowledge base**

**Use Cases:**
- extract_wisdom (Fabric pattern)
- Deep code review with architecture analysis
- Project/codebase analysis
- Security audits
- Creative writing

**Model:** `claude-sonnet-4-5` (Anthropic API)
**Cost:** ~$0.02 per call (varies by content length)

### **Tier 3: Adaptive/Hybrid**

**Characteristics:**
- Tries Tier 1 first
- Falls back to Tier 2 if:
  - Confidence score is low
  - Task complexity detected
  - User requests "thorough" mode
- Smart routing based on content

**Use Cases:**
- Code review (simple → tier 1, complex → tier 2)
- Summarization (short → tier 1, long/nuanced → tier 2)
- Analysis tasks with variable complexity

**Model:** Auto-selected
**Cost:** $0.00 - $0.02 depending on routing

---

## 📁 Module Structure

### **Directory Layout**

```
SCRIPTS/modules/
├── tier1/                          # Small brain modules
│   ├── text/
│   │   ├── improve_writing.sh
│   │   ├── fix_grammar.sh
│   │   └── format_markdown.sh
│   ├── code/
│   │   ├── format_code.sh
│   │   ├── extract_todos.sh
│   │   └── simple_explain.sh
│   └── data/
│       ├── parse_csv.sh
│       └── extract_numbers.sh
│
├── tier2/                          # Big brain modules
│   ├── wisdom/
│   │   ├── extract_wisdom.sh      # Fabric-style extraction
│   │   └── summarize_research.sh
│   ├── analysis/
│   │   ├── analyze_codebase.sh
│   │   ├── security_audit.sh
│   │   └── architecture_review.sh
│   └── creative/
│       ├── write_essay.sh
│       └── create_presentation.sh
│
└── tier3/                          # Adaptive modules
    ├── code_review_auto.sh
    ├── summarize_adaptive.sh
    └── explain_adaptive.sh
```

### **Module Metadata Format**

Each module includes metadata in comments:

```bash
#!/bin/bash
# MODULE: extract_wisdom
# NAME: Extract Wisdom
# CATEGORY: wisdom
# TIER: 2
# MODEL: claude-sonnet-4-5
# COST_ESTIMATE: 0.02
# DESCRIPTION: Extract insights, ideas, quotes, habits from content (Fabric pattern)
# PARAM: file:path:Path to content file (article, transcript, etc.)
# OUTPUT: markdown:obsidian:RAG-enabled
# TAGS: analysis, synthesis, learning

# ... module implementation ...
```

**New Fields:**
- `TIER` - 1, 2, or 3
- `MODEL` - Specific model to use
- `COST_ESTIMATE` - Estimated $ per call
- `OUTPUT` - Output format and destinations
- `TAGS` - For categorization and search

---

## 🔄 Data Flow Pipeline

### **The Complete Workflow**

```
┌─────────────────────────────────────────────────┐
│ 1. USER REQUEST                                 │
│    TUI: "ai:extract_wisdom article.md"          │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ 2. AGENT-GATEWAY                                │
│    - Finds extract_wisdom module                │
│    - Reads metadata: TIER=2, MODEL=claude       │
│    - Executes module with parameters            │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ 3. MODULE EXECUTION                             │
│    - Reads input file                           │
│    - Calls util.llm.query_big with Fabric prompt│
│    - Gets structured output (IDEAS, INSIGHTS, etc)│
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ 4. OBSIDIAN INTEGRATION                         │
│    util.obsidian.write:                         │
│    - Add frontmatter (date, tags, metadata)     │
│    - Write to ~/obsidian_vault/wisdom/          │
│    - Git commit with descriptive message        │
│    - Return file path                           │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ 5. RAG INGESTION (Async)                        │
│    util.rag.ingest:                             │
│    - Chunk markdown intelligently               │
│    - Generate embeddings (local model)          │
│    - Store in ChromaDB vector database          │
│    - Index for semantic search                  │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ 6. KNOWLEDGE BASE ENRICHED                      │
│    - New wisdom available for retrieval         │
│    - Future Tier 1 queries can use this context │
│    - System is now smarter                      │
└─────────────────────────────────────────────────┘
```

### **RAG-Enhanced Tier 1 Workflow**

```
┌─────────────────────────────────────────────────┐
│ 1. USER REQUEST                                 │
│    TUI: "ai:review_code auth.rb"                │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ 2. MODULE: code_review_smart.sh (Tier 1)        │
│    - Detects context needed: "auth security"    │
│    - Calls util.rag.retrieve                    │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ 3. RAG RETRIEVAL                                │
│    Query: "authentication security ruby"        │
│    Returns:                                     │
│    - Wisdom doc: JWT best practices             │
│    - Past review: Timing attack warning         │
│    - Architecture: Auth layer conventions       │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ 4. ENHANCED PROMPT                              │
│    Combines:                                    │
│    - Retrieved context (3K tokens)              │
│    - Current file (2K tokens)                   │
│    - Review instructions (0.5K tokens)          │
│    Total: 5.5K tokens (fits in 8K context!)     │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ 5. TIER 1 LLM (Hermes-3-8B)                     │
│    - Processes enhanced prompt                  │
│    - Makes connections from retrieved docs      │
│    - Generates context-aware review             │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ 6. SMART OUTPUT                                 │
│    Review references:                           │
│    - Project conventions from past reviews      │
│    - Security wisdom from ingested articles     │
│    - Architecture patterns from docs            │
│    Cost: $0.00 (local model)                    │
└─────────────────────────────────────────────────┘
```

---

## 🛠️ Implementation Phases

### **Phase 1: Tiered Module Foundation** (Week 1)

**Goal:** Prove tiered modules work, no RAG yet

**Tasks:**
1. ✅ Update module metadata parser to recognize TIER, MODEL, COST_ESTIMATE
2. ✅ Enhance utilities/llm.sh with tier support:
   - `util.llm.query_small()` - Uses teton (existing)
   - `util.llm.query_big()` - Uses Claude API (NEW)
3. ✅ Build first Tier 1 module: `improve_writing.sh`
   - Simple, fast, proves Tier 1 works
4. ✅ Build first Tier 2 module: `extract_wisdom.sh`
   - Adapt Fabric pattern
   - Proves big brain integration
5. ✅ Test both modules end-to-end via TUI

**Success Criteria:**
- ✅ Tier 1 module executes in < 5 seconds, costs $0
- ✅ Tier 2 module produces rich output, costs ~$0.02
- ✅ Agent-gateway correctly routes by tier
- ✅ TUI displays both modules in palette

**Deliverables:**
- `utilities/llm.sh` (enhanced)
- `modules/tier1/text/improve_writing.sh`
- `modules/tier2/wisdom/extract_wisdom.sh`
- Updated `programs/module_scanner.go` (tier metadata)

---

### **Phase 2: Obsidian Integration** (Week 2)

**Goal:** Module outputs flow to Obsidian vault

**Tasks:**
1. ✅ Create `utilities/obsidian.sh`:
   - `util.obsidian.write()` - Write with frontmatter
   - `util.obsidian.commit()` - Git commit helper
2. ✅ Configure vault location and structure:
   ```
   ~/obsidian_vault/
   ├── wisdom/           # extract_wisdom outputs
   ├── code-reviews/     # code review outputs
   ├── analysis/         # project analysis outputs
   └── daily/            # daily notes, logs
   ```
3. ✅ Update extract_wisdom.sh to use obsidian utility
4. ✅ Test: Run module → Check file in vault → Verify git commit
5. ✅ Add frontmatter templates for different output types

**Success Criteria:**
- ✅ Module output appears in Obsidian with proper formatting
- ✅ Frontmatter includes all relevant metadata
- ✅ Git commits are descriptive and automatic
- ✅ Files are human-readable and editable
- ✅ Can manually browse/edit in Obsidian app

**Deliverables:**
- `utilities/obsidian.sh`
- Updated `extract_wisdom.sh` with obsidian integration
- Obsidian vault structure documented
- Frontmatter templates

---

### **Phase 3: RAG Infrastructure** (Week 3)

**Goal:** Ingest Obsidian vault into searchable vector database

**Tasks:**
1. ✅ Choose and install ChromaDB:
   ```bash
   pip3 install chromadb sentence-transformers
   ```
2. ✅ Create `utilities/rag.sh`:
   - `util.rag.ingest <file>` - Add file to vector DB
   - `util.rag.retrieve <query> [top_k]` - Semantic search
   - `util.rag.status` - Show DB stats
3. ✅ Create Python RAG scripts:
   - `lib/rag_ingest.py` - Chunking and embedding
   - `lib/rag_retrieve.py` - Query and results
   - `lib/rag_chunk.py` - Intelligent markdown chunking
4. ✅ Test embedding model locally:
   - `all-MiniLM-L6-v2` (22MB, 384 dims, fast)
5. ✅ Bulk ingest existing Obsidian vault:
   ```bash
   util.rag.bulk_ingest ~/obsidian_vault
   ```
6. ✅ Test retrieval quality with sample queries

**Success Criteria:**
- ✅ Can ingest markdown files (< 2 seconds per file)
- ✅ Semantic search returns relevant chunks
- ✅ Vector DB is persistent across restarts
- ✅ Embedding model runs on CPU (no GPU needed)
- ✅ Can handle 1000+ documents

**Deliverables:**
- `utilities/rag.sh`
- `lib/rag_ingest.py`
- `lib/rag_retrieve.py`
- `lib/rag_chunk.py`
- ChromaDB instance at `~/.local/share/battlestag/rag_db`

---

### **Phase 4: RAG-Enhanced Tier 1 Modules** (Week 4)

**Goal:** Tier 1 modules use RAG for context

**Tasks:**
1. ✅ Update `utilities/llm.sh`:
   - `util.llm.query_with_rag()` - Retrieve + prompt
2. ✅ Build RAG-enhanced module: `code_review_smart.sh`
   - Tier 1 module
   - Retrieves past reviews and wisdom
   - Costs $0 but performs like Tier 2
3. ✅ Test comparison:
   - Same code review without RAG vs with RAG
   - Measure quality improvement
4. ✅ Build second RAG module: `explain_code_smart.sh`
5. ✅ Optimize retrieval strategy:
   - Tune top_k (how many chunks)
   - Test different query formulations
   - Add metadata filtering

**Success Criteria:**
- ✅ RAG-enhanced Tier 1 output quality approaches Tier 2
- ✅ Still executes in < 10 seconds
- ✅ Context retrieval is relevant (80%+ useful)
- ✅ System cost remains low ($0 for Tier 1)

**Deliverables:**
- Enhanced `utilities/llm.sh` with RAG support
- `modules/tier1/code/code_review_smart.sh`
- `modules/tier1/code/explain_code_smart.sh`
- Retrieval optimization documentation

---

### **Phase 5: Tier 3 Adaptive Routing** (Week 5)

**Goal:** Smart routing between Tier 1 and Tier 2

**Tasks:**
1. ✅ Design routing logic:
   ```bash
   # utilities/llm.sh
   util.llm.query_adaptive() {
       # Try Tier 1 with RAG first
       result=$(util.llm.query_with_rag "$query" "$prompt")

       # Check confidence/quality
       if needs_tier2; then
           # Fall back to Claude
           result=$(util.llm.query_big "$prompt")
       fi

       echo "$result"
   }
   ```
2. ✅ Implement confidence detection:
   - Check for hedging words ("might", "possibly")
   - Detect error messages or failures
   - Allow user override (--thorough flag)
3. ✅ Build Tier 3 module: `code_review_auto.sh`
4. ✅ Add cost tracking:
   - Log when Tier 2 is used
   - Monthly budget alerts
5. ✅ Test routing effectiveness

**Success Criteria:**
- ✅ Routes to Tier 2 only when necessary (< 20% of calls)
- ✅ User can override routing decision
- ✅ Cost savings vs always-Tier-2 approach
- ✅ Quality maintained (no degradation)

**Deliverables:**
- Adaptive routing in `utilities/llm.sh`
- `modules/tier3/code_review_auto.sh`
- Cost tracking system
- Routing analytics

---

## 📊 Module Catalog (Planned)

### **Tier 1 Modules** (Build First)

| Module | Description | Input | Output |
|--------|-------------|-------|--------|
| `improve_writing.sh` | Fix grammar, style | Text file | Improved text |
| `format_code.sh` | Code formatting | Source file | Formatted code |
| `extract_todos.sh` | Pull TODO comments | Code file | Todo list |
| `simple_summarize.sh` | Basic summary | Text file | 3-sentence summary |
| `code_review_smart.sh` | RAG-enhanced review | Source file | Review + suggestions |

### **Tier 2 Modules** (High Value)

| Module | Description | Input | Output | Est. Cost |
|--------|-------------|-------|--------|-----------|
| `extract_wisdom.sh` | Fabric pattern | Article/transcript | Rich insights | $0.02 |
| `analyze_codebase.sh` | Deep analysis | Directory | Architecture doc | $0.05 |
| `security_audit.sh` | Vulnerability scan | Source files | Security report | $0.03 |
| `write_essay.sh` | Creative writing | Topic + outline | Essay | $0.04 |

### **Tier 3 Modules** (Smart Routing)

| Module | Description | Routing Logic |
|--------|-------------|---------------|
| `code_review_auto.sh` | Adaptive review | Simple → T1, Complex → T2 |
| `summarize_adaptive.sh` | Smart summary | Short → T1, Long/nuanced → T2 |
| `explain_adaptive.sh` | Code explanation | Standard code → T1, Novel patterns → T2 |

---

## 🎯 Success Metrics

### **System Performance**

**Tier 1:**
- Response time: < 5 seconds
- Cost: $0.00 per call
- Quality: Good enough for 80% of tasks
- Improved by RAG: 2-3x better with context

**Tier 2:**
- Response time: 5-15 seconds
- Cost: $0.01-0.05 per call
- Quality: Excellent, publication-ready
- ROI: Creates knowledge assets

**RAG System:**
- Ingestion: < 2 seconds per file
- Retrieval: < 0.5 seconds
- Relevance: 80%+ useful chunks
- Scalability: 1000+ documents

### **Knowledge Accumulation**

**Month 1:**
- 20 wisdom extractions → 20 knowledge assets
- 100 code reviews reference past reviews
- System learns project conventions

**Month 3:**
- 60 wisdom extractions
- 500 code reviews with RAG
- Personal knowledge base established

**Month 6:**
- 120+ knowledge assets
- Tier 1 quality approaches Tier 2 (with RAG)
- System feels personalized and intelligent

---

## 💰 Cost Analysis

### **Monthly Budget Scenarios**

**Conservative ($10/month):**
- ~500 Tier 2 calls
- ~10-20 wisdom extractions per month
- Unlimited Tier 1 with RAG
- **Use case:** Personal learning, occasional deep analysis

**Moderate ($25/month):**
- ~1,250 Tier 2 calls
- ~50 wisdom extractions
- Daily deep analyses
- **Use case:** Active development, research

**Aggressive ($50/month):**
- ~2,500 Tier 2 calls
- ~100 wisdom extractions
- Heavy creative work
- **Use case:** Professional tool, client work

**Reality:** Most usage will be Tier 1 (free) with occasional Tier 2 for high-value tasks.

---

## 🔧 Technical Stack

### **Languages & Tools**

| Component | Technology | Why |
|-----------|-----------|-----|
| Modules | Bash | Shell integration, pipe-friendly |
| Primitives/Utilities | Bash | Composability, portability |
| RAG Scripts | Python 3 | ML libraries, ChromaDB support |
| Vector DB | ChromaDB | Simple, local, persistent |
| Embeddings | sentence-transformers | Free, local, quality |
| Embedding Model | all-MiniLM-L6-v2 | 22MB, fast, 384 dims |
| Tier 1 LLM | Hermes-3-8B (local) | Free, private, good quality |
| Tier 2 LLM | Claude Sonnet 4.5 | Best reasoning, synthesis |
| Version Control | Git | Obsidian vault tracking |
| API Gateway | Go (existing) | Fast, typed, proven |
| TUI | Go + Bubbletea | Professional, tested |

### **Dependencies**

**System:**
- Python 3.8+
- pip3
- Git
- curl/jq (existing)

**Python Packages:**
```bash
pip3 install chromadb sentence-transformers anthropic
```

**Disk Space:**
- ChromaDB: ~100MB per 1000 docs
- Embedding model: 22MB
- Obsidian vault: ~10-50MB
- Total: < 500MB

---

## 🚀 Getting Started (Phase 1)

### **Week 1 Focus: Prove Tiered Modules**

**Day 1-2: LLM Utilities**
- Enhance `utilities/llm.sh`
- Add `util.llm.query_big()` for Claude
- Test both small and big brain queries
- Document API key setup

**Day 3-4: First Modules**
- Build `improve_writing.sh` (Tier 1)
- Build `extract_wisdom.sh` (Tier 2)
- Test via command line
- Verify outputs

**Day 5: Integration**
- Update module scanner for tier metadata
- Test via agent-gateway API
- Test via TUI
- Celebrate first working tiered system! 🎉

**Day 6-7: Refinement**
- Add cost tracking
- Improve error handling
- Write tests
- Document patterns

---

## 📝 Open Questions

### **For Discussion:**

1. **Claude API Access:**
   - Do you have an Anthropic API key?
   - Or should we use OpenRouter/other proxy?
   - Budget comfort level?

2. **Obsidian Vault Location:**
   - Existing vault or new one?
   - Directory structure preferences?
   - Git workflow (auto-commit vs manual)?

3. **RAG Scope:**
   - Just module outputs or entire vault?
   - Include existing notes?
   - Filtering strategy (tags, folders)?

4. **Priority Modules:**
   - Which Fabric patterns excite you most?
   - Which would you use daily?
   - What's missing from Fabric you'd add?

5. **Tier 3 Routing:**
   - User control (manual flag) or automatic?
   - Confidence threshold tuning?
   - Allow tier override per invocation?

---

## 🎯 Next Session Plan

**Tonight's Outcome:** Solid plan documented ✅

**Tomorrow's Focus:**
1. Answer open questions above
2. Set up Claude API (or choose alternative)
3. Start Phase 1, Day 1: Enhance LLM utilities
4. Build first Tier 2 module (extract_wisdom)
5. Test end-to-end

**This Week's Goal:**
- Working tiered module system
- At least 2 modules (one Tier 1, one Tier 2)
- Proven via TUI

---

## 💡 Key Insights

**Why This Approach Wins:**

1. **Obsidian = Insurance**
   - If RAG breaks, rebuild from vault
   - Human-readable, editable
   - Version controlled
   - Multi-tool compatible (not locked to RAG)

2. **Tiered = Economics**
   - Free tier handles volume
   - Paid tier adds quality
   - Adaptive tier optimizes cost/quality
   - Scales with usage patterns

3. **RAG = Multiplier**
   - Tier 1 quality improves over time
   - Knowledge compounds
   - System learns without retraining
   - 8B model punches above weight

4. **Fabric-Inspired = Validated**
   - 233 patterns prove concept works
   - Can adapt best patterns
   - Community-tested prompts
   - Clear template to follow

**This is the right architecture.** 🎯

---

*Last Updated: 2025-12-20*
*Next Review: After Phase 1 completion*
