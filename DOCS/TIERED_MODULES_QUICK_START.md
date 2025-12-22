# Tiered Modules Quick Start Guide

**For when you want to jump right in tomorrow!**

---

## 🎯 What We're Building

A **3-tier AI module system** that gets smarter over time:

```
Tier 1 (Free) → Fast, good enough, unlimited usage
Tier 2 (Paid) → Smart, deep analysis, ~$0.02 per call
Tier 3 (Smart) → Tries free first, upgrades if needed

All outputs → Obsidian (human readable) → RAG (machine searchable)
```

**Result:** Your 8B model becomes as smart as Claude by learning from past interactions!

---

## 📋 Before You Start - Answer These

### 1. Claude API Setup
**Options:**
- [ ] Direct Anthropic API (most straightforward)
  - Get key: https://console.anthropic.com/
  - Cost: $3 per million input tokens
- [ ] OpenRouter (multiple providers, one key)
  - Get key: https://openrouter.ai/
  - Cost: Varies by provider
- [ ] Skip for now (build Tier 1 only, add Tier 2 later)

**Your choice:** ________________

### 2. Budget Comfort
**Monthly spending on AI:**
- [ ] $0 (local only, accept limitations)
- [ ] $5-10 (occasional wisdom extraction)
- [ ] $20-30 (regular use, multiple extractions/week)
- [ ] $50+ (heavy user, daily analyses)

**Your choice:** ________________

**Reality check:** At $0.02 per extract_wisdom call:
- $10/mo = 500 calls = 16/day
- $20/mo = 1000 calls = 33/day
- Most people use < 100/month = $2

### 3. Obsidian Vault
**Where to save module outputs:**

- [ ] Use existing vault: `~/obsidian_vault`
- [ ] Create new vault: `~/battlestag-knowledge`
- [ ] Other location: ________________

**Structure:**
```
[vault]/
├── wisdom/         # extract_wisdom outputs
├── code-reviews/   # code review outputs
├── analysis/       # project analyses
└── daily/          # daily logs
```

**Your choice:** ________________

### 4. First Modules Priority

**Which would you use most?**

Rank 1-5 (1 = build first):

- [ ] `extract_wisdom` - Rich insights from articles/videos (Tier 2)
- [ ] `improve_writing` - Grammar/style fixes (Tier 1)
- [ ] `code_review_smart` - Smart code review with RAG (Tier 1 + RAG)
- [ ] `summarize` - Quick summaries (Tier 1)
- [ ] `analyze_codebase` - Deep project analysis (Tier 2)

**Your top 3:** ________________

---

## 🚀 Phase 1 Kickoff (Week 1)

### Day 1: LLM Utilities (2-3 hours)

**Tasks:**
1. Set up Claude API credentials
2. Enhance `SCRIPTS/utilities/llm.sh`:
   ```bash
   util.llm.query_big() {
       # Call Claude API
   }
   ```
3. Test both small and big brain queries
4. Add cost tracking utility

**Outcome:** Can call both Hermes-3 and Claude from bash

---

### Day 2-3: Build First Modules (4-6 hours)

**Tier 1 Module: `improve_writing.sh`**
```bash
#!/bin/bash
# MODULE: improve_writing
# TIER: 1
# MODEL: hermes-3-8b
# DESCRIPTION: Fix grammar and style

source "$SCRIPT_DIR/utilities/llm.sh"

text=$(cat "$1")
improved=$(util.llm.query_small "Fix grammar and style: $text")
echo "$improved"
```

**Tier 2 Module: `extract_wisdom.sh`**
```bash
#!/bin/bash
# MODULE: extract_wisdom
# TIER: 2
# MODEL: claude-sonnet-4-5
# DESCRIPTION: Extract insights from content (Fabric pattern)

source "$SCRIPT_DIR/utilities/llm.sh"

content=$(cat "$1")
# Use Fabric's extract_wisdom prompt
wisdom=$(util.llm.query_big "$fabric_prompt\n\n$content")
echo "$wisdom"
```

**Outcome:** Two working modules, different tiers

---

### Day 4: Integration Testing (2-3 hours)

**Test via command line:**
```bash
cd SCRIPTS
./modules/tier1/text/improve_writing.sh sample.txt
./modules/tier2/wisdom/extract_wisdom.sh article.md
```

**Test via agent-gateway:**
```bash
cd ORCHESTRATOR/agent-gateway
go run main.go  # Should auto-discover new modules
curl http://localhost:8080/api/programs/list | jq
```

**Test via TUI:**
```bash
cd APPS/battlestag-tui
go run main.go
# Check command palette for new modules
```

**Outcome:** Modules work in all 3 interfaces

---

### Day 5: Refinement (2-3 hours)

**Add polish:**
- Error handling
- Cost tracking
- Better prompts
- Documentation
- Basic tests

**Outcome:** Production-ready Tier 1 & 2 system

---

## 📝 Week 1 Success Criteria

By end of Week 1, you should have:

- ✅ Two working modules (one Tier 1, one Tier 2)
- ✅ Both callable via TUI command palette
- ✅ Tier 1 costs $0, completes in < 5 seconds
- ✅ Tier 2 costs ~$0.02, produces rich output
- ✅ Cost tracking working
- ✅ Module metadata properly parsed
- ✅ Basic documentation

**If yes to all → Move to Phase 2 (Obsidian integration)**

**If no → Debug, polish, then proceed**

---

## 🎯 Week 2 Preview: Obsidian Integration

Once Phase 1 works, Week 2 adds:

1. `utilities/obsidian.sh` - Write outputs to vault
2. Automatic git commits
3. Frontmatter with metadata
4. Human-readable knowledge base

**Week 2 outcome:** Every module run creates a knowledge artifact

---

## 📚 Reference Documents

**Full Plan:**
- `DOCS/TIERED_MODULES_RAG_PLAN.md` - Complete architecture

**Previous Work:**
- `DOCS/CAPABILITY_HIERARCHY_ARCHITECTURE.md` - Primitives/Utilities foundation
- `.claude/PROJECT_PROGRESS.md` - Current status

**Fabric Research:**
- GitHub: https://github.com/danielmiessler/fabric
- 233 patterns to inspire from
- Your extract_wisdom is their most popular pattern!

---

## 💡 Key Reminders

**Architecture Principles:**
1. **Files over database** - Everything in Obsidian (human readable)
2. **Tier 1 for volume** - Free, fast, unlimited
3. **Tier 2 for value** - Paid, smart, creates knowledge assets
4. **RAG multiplies Tier 1** - Gets smarter over time
5. **Obsidian = insurance** - Can rebuild RAG anytime

**Common Pitfalls:**
- ❌ Don't build RAG first (no content yet!)
- ❌ Don't over-engineer Tier 3 routing initially
- ❌ Don't skip Obsidian integration (it's critical!)
- ✅ Do build incrementally
- ✅ Do test each piece thoroughly
- ✅ Do celebrate small wins!

---

## 🎉 The Flywheel

Once all phases complete:

```
1. Read article → extract_wisdom (Tier 2, $0.02)
   ↓
2. Rich insights saved to Obsidian + RAG
   ↓
3. Next code review uses those insights (Tier 1, $0.00)
   ↓
4. Review creates new patterns doc
   ↓
5. Back to step 2, system gets smarter!
```

**After 3 months:**
- 60+ wisdom documents
- 1000+ code reviews with context
- Your 8B model references your learnings
- Feels like a personalized AI that knows you

**This is the goal!** 🚀

---

## ❓ Stuck? Check These

**Can't call Claude API:**
- Verify API key is set correctly
- Check network connectivity
- Try curl test: `curl https://api.anthropic.com/v1/messages -H "x-api-key: $KEY"`

**Module not appearing in TUI:**
- Check metadata format (comments at top)
- Restart agent-gateway
- Check logs: `tail -f ~/.local/share/battlestag/logs/*.log`

**8B model not responding:**
- Check teton server: `curl http://192.168.1.232:1234/v1/models`
- Verify LM Studio running
- Check network (Tailscale up?)

**Cost tracking not working:**
- Check write permissions: `~/.local/share/battlestag/`
- Verify JSON file created
- Check util.llm.track_cost() function

---

## 🚦 Ready to Start?

**Complete the checklist above, then:**

```bash
cd BATTLESTAG-BOT
git pull  # Make sure you're current
cd SCRIPTS/utilities
# Start editing llm.sh to add util.llm.query_big()
```

**You got this!** The plan is solid, the research is done, now it's just execution.

---

*Created: 2025-12-20*
*For: Phase 1 kickoff tomorrow*
*Estimated time: 10-15 hours over 5 days*
