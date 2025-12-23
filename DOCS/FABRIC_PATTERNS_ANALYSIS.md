# Fabric Patterns Analysis for BATTLESTAG Modules

**Total Patterns**: 233
**Source**: https://github.com/danielmiessler/fabric/tree/main/data/patterns
**Date**: 2025-12-22

---

## Tier Classification Strategy

### Tier 1 (Local LLM - hermes3:8b)
**Best for**: Simple, structured prompts with clear outputs
- Pattern recognition tasks
- Basic extraction/formatting
- Simple transformations
- Quick analysis with clear criteria
- Cost: $0.00 | Privacy: 100% local

### Tier 2 (Cloud LLM - Claude/GPT/Gemini)
**Best for**: Deep reasoning, nuance, synthesis
- Complex analysis requiring judgment
- Creative writing with tone/style
- Multi-dimensional evaluation
- Strategic thinking
- Cost: $0.01-0.10 | Privacy: Cloud provider

### Tier 3 (Adaptive Routing)
**Best for**: Try Tier 1 first, escalate if needed
- Quality-sensitive tasks with fallback
- Cost optimization
- Unknown complexity

---

## High-Priority Patterns by Category

### 📝 CONTENT ANALYSIS (Tier 2 Heavy)

**Tier 2 - Deep Analysis:**
- ✅ **extract_wisdom** - ALREADY BUILT! ✓
- **extract_insights** - Find profound insights
- **extract_article_wisdom** - Article-specific wisdom
- **analyze_paper** - Academic paper analysis
- **analyze_claims** - Fact-check and validate claims
- **analyze_debate** - Debate analysis and scoring
- **analyze_presentation** - Presentation effectiveness
- **rate_content** - Quality rating with criteria
- **get_wow_per_minute** - Content value density

**Tier 1 - Simple Extraction:**
- **extract_ideas** - List key ideas
- **extract_questions** - Pull out questions
- **extract_references** - Extract citations/links
- **extract_quotes** - Find quotable sections
- **create_tags** - Generate topic tags
- **extract_sponsors** - List sponsors mentioned

### ✍️ WRITING & IMPROVEMENT (Mixed Tiers)

**Tier 1 - Grammar/Formatting:**
- ✅ **improve_writing** - ALREADY BUILT! ✓
- **fix_typos** - Simple typo correction
- **clean_text** - Remove formatting issues
- **humanize** - Make AI text more natural

**Tier 2 - Style/Creative:**
- **improve_academic_writing** - Academic style refinement
- **write_essay** - Essay composition
- **write_micro_essay** - Short-form essays
- **create_better_frame** - Reframe arguments
- **enrich_blog_post** - Enhance blog content

### 📊 SUMMARIZATION (Tiered by Complexity)

**Tier 1 - Basic Summaries:**
- **summarize** - General summarization
- **summarize_micro** - Ultra-short summaries
- **create_micro_summary** - Condensed versions
- **create_5_sentence_summary** - 5-sentence limit

**Tier 2 - Complex Summaries:**
- **summarize_paper** - Academic paper summaries
- **summarize_lecture** - Lecture notes
- **summarize_meeting** - Meeting minutes
- **youtube_summary** - Video transcript summaries
- **extract_main_idea** - Core concept extraction

### 💻 CODE & DEVELOPMENT (Tier 1 Dominant)

**Tier 1 - Code Operations:**
- **explain_code** - Code explanation
- **explain_project** - Project overview
- **analyze_logs** - Log file analysis
- **summarize_git_changes** - Git diff summary
- **review_code** - Code review checklist
- **create_git_diff_commit** - Commit message generation

**Tier 2 - Complex Code:**
- **create_coding_project** - Full project scaffolding
- **create_design_document** - Technical design docs
- **create_prd** - Product requirements
- **refine_design_document** - Design doc improvement

### 🎨 CREATIVE (Tier 2 Heavy)

**Tier 2 - Creative Generation:**
- **create_art_prompt** - AI art prompts
- **create_story_explanation** - Story narratives
- **create_mnemonic_phrases** - Memory aids
- **create_aphorisms** - Wisdom statements
- **create_keynote** - Presentation creation
- **tweet** - Social media posts

### 🔒 SECURITY & THREAT ANALYSIS (Mixed)

**Tier 1 - Structured Analysis:**
- **analyze_threat_report** - Threat report parsing
- **analyze_incident** - Incident breakdown
- **extract_domains** - Domain extraction
- **extract_poc** - Extract proof-of-concept code

**Tier 2 - Strategic Security:**
- **analyze_malware** - Malware behavior analysis
- **create_threat_scenarios** - Threat modeling
- **create_stride_threat_model** - STRIDE analysis
- **create_security_update** - Security bulletins

### 📈 BUSINESS & STRATEGY (Tier 2 Heavy)

**Tier 2 - Business Analysis:**
- **analyze_product_feedback** - Customer feedback synthesis
- **analyze_sales_call** - Sales call analysis
- **create_user_story** - User story creation
- **create_business_ideas** - Business idea generation
- **analyze_proposition** - Value proposition analysis
- **prepare_7s_strategy** - McKinsey 7S framework

### 🧠 LEARNING & EDUCATION (Mixed)

**Tier 1 - Study Aids:**
- **create_quiz** - Quiz generation
- **to_flashcards** / **create_flash_cards** - Flashcard creation
- **create_reading_plan** - Reading schedule
- **explain_terms** - Terminology explanations

**Tier 2 - Deep Learning:**
- **create_conceptmap** - Concept mapping
- **create_mermaid_visualization** - Diagram generation
- **extract_algorithm_update_recommendations** - Learning path
- **label_and_rate** - Educational assessment

### 🎯 DECISION MAKING (Tier 2)

**Tier 2 - Analysis & Judgment:**
- **analyze_risk** - Risk assessment
- **compare_and_contrast** - Comparison analysis
- **find_logical_fallacies** - Logic checking
- **rate_ai_response** - AI output evaluation
- **judge_output** - Quality judgment

---

## Recommended Implementation Priority

### Phase 1: Foundation Modules (Week 1)
**Tier 1 (Local - Privacy First):**
1. ✅ improve_writing - DONE
2. ✅ csv_insight - DONE
3. **summarize** - Basic summarization
4. **explain_code** - Code explanation
5. **create_tags** - Topic tagging
6. **extract_ideas** - Idea extraction

**Tier 2 (Cloud - Quality First):**
1. ✅ extract_wisdom - DONE
2. **analyze_claims** - Fact-checking
3. **summarize_paper** - Academic summaries
4. **improve_academic_writing** - Academic polish

### Phase 2: Content Processing (Week 2)
**Tier 1:**
- **clean_text** - Text cleanup
- **fix_typos** - Typo correction
- **extract_references** - Citation extraction
- **summarize_git_changes** - Git summaries

**Tier 2:**
- **analyze_paper** - Deep paper analysis
- **extract_insights** - Insight extraction
- **rate_content** - Content quality rating
- **create_better_frame** - Argument reframing

### Phase 3: Creative & Business (Week 3)
**Tier 2 (Most need creative thinking):**
- **create_art_prompt** - Art generation prompts
- **write_essay** - Essay writing
- **analyze_product_feedback** - Customer insights
- **create_user_story** - User story creation
- **tweet** - Social media content

### Phase 4: Development Tools (Week 4)
**Tier 1:**
- **review_code** - Code review
- **analyze_logs** - Log analysis
- **explain_project** - Project docs
- **create_git_diff_commit** - Commit messages

**Tier 2:**
- **create_design_document** - Design docs
- **create_prd** - Product specs
- **create_coding_project** - Project scaffolding

### Phase 5: Specialized (Ongoing)
**Security (Mixed):**
- **analyze_malware** (T2)
- **create_threat_scenarios** (T2)
- **analyze_threat_report** (T1)

**Learning (Mixed):**
- **create_quiz** (T1)
- **create_flash_cards** (T1)
- **create_conceptmap** (T2)

---

## Module Template Structure

Each Fabric pattern becomes a BATTLESTAG module:

```
modules-available/
├── tier1/
│   ├── text/
│   │   ├── improve_writing.sh ✓
│   │   ├── summarize.sh
│   │   ├── clean_text.sh
│   │   └── fix_typos.sh
│   ├── code/
│   │   ├── explain_code.sh
│   │   ├── review_code.sh
│   │   └── summarize_git_changes.sh
│   └── data/
│       ├── csv_insight.sh ✓
│       ├── create_tags.sh
│       └── extract_ideas.sh
├── tier2/
│   ├── wisdom/
│   │   ├── extract_wisdom.sh ✓
│   │   ├── extract_insights.sh
│   │   └── analyze_claims.sh
│   ├── creative/
│   │   ├── write_essay.sh
│   │   ├── create_art_prompt.sh
│   │   └── tweet.sh
│   └── analysis/
│       ├── analyze_paper.sh
│       ├── rate_content.sh
│       └── summarize_paper.sh
└── tier3/
    ├── adaptive_summarize.sh
    ├── adaptive_writing.sh
    └── adaptive_analysis.sh
```

---

## Quick Win Patterns (Easiest to Adapt)

These patterns have simple, well-defined prompts:

1. **summarize** - "Summarize this in X sentences"
2. **create_tags** - "Generate topic tags"
3. **extract_ideas** - "List key ideas"
4. **explain_code** - "Explain this code"
5. **create_5_sentence_summary** - "5 sentence summary"
6. **extract_questions** - "List all questions"
7. **extract_references** - "Extract citations"
8. **clean_text** - "Remove formatting issues"
9. **fix_typos** - "Fix spelling errors"
10. **humanize** - "Make more natural"

---

## Implementation Notes

### For Each Pattern:
1. **Analyze the Fabric prompt** - What's the core task?
2. **Determine tier** - Simple/structured = T1, Complex/creative = T2
3. **Create module file** - Use existing templates
4. **Adapt prompt** - Optimize for our models
5. **Test with both tiers** - Validate quality
6. **Document cost** - Set user expectations

### Module Metadata Template:
```bash
# MODULE: pattern_name
# NAME: Pattern Display Name
# CATEGORY: text|code|data|wisdom|creative|analysis
# TIER: 1|2|3
# DESCRIPTION: What this pattern does
# PARAM: input_file:path:Description
# PARAM: output_file:path:Description (optional)
```

---

## Next Steps

1. **Pick 5-10 priority patterns** - User decides based on their needs
2. **Create modules** - Start with Tier 1 (free, fast validation)
3. **Test quality** - Compare T1 vs T2 on same pattern
4. **Document** - Usage examples, cost estimates
5. **Iterate** - Refine prompts based on results

**Question for User**: Which categories are most valuable for your workflow?
- Content analysis?
- Writing improvement?
- Code tools?
- Business analysis?
- Creative generation?
