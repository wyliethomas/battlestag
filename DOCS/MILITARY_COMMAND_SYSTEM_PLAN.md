# Military Command System - Implementation Plan

**Date:** 2025-11-25
**Status:** Planning Phase
**Priority:** High
**Complexity:** High (Invasive architectural change)

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Problem Statement](#problem-statement)
3. [Current State Analysis](#current-state-analysis)
4. [Proposed Solution](#proposed-solution)
5. [Implementation Phases](#implementation-phases)
6. [Testing Strategy](#testing-strategy)
7. [Rollback Plan](#rollback-plan)
8. [Success Criteria](#success-criteria)
9. [Future Considerations](#future-considerations)

---

## Executive Summary

### The Problem
The LLM is inconsistently deciding when to use programs, leading to:
- Programs being called for general conversation (brainstorming triggers task-manager)
- Incorrect program selection (echo was called instead of just answering)
- System prompt rules being ignored despite explicit instructions

### The Solution
Implement a **Military Command System** that uses distinctive keywords to explicitly signal intent:
- **No military keywords** → Pure chat, no programs shown to LLM
- **Military keywords present** → Filtered program list shown, clear intent

### Key Benefit
**Explicit > Implicit** - User controls program execution through clear, unambiguous language.

---

## Problem Statement

### Current Issues

#### Issue 1: LLM Over-Eagerness
**Symptom:** LLM calls programs when it shouldn't
**Example:** User asks "Come up with name ideas" → LLM calls task-manager
**Root Cause:** LLM sees programs in registry, tries to find a reason to use one

#### Issue 2: System Prompt Fragility
**Symptom:** Despite explicit rules ("DO NOT use programs for brainstorming"), LLM ignores them
**Example:** Added 11 rules, LLM still calls programs for creative tasks
**Root Cause:** System prompts are suggestions, not guarantees

#### Issue 3: Ambiguous Intent
**Symptom:** System can't distinguish "data query" from "general conversation"
**Example:** "Show me the projects" (needs program) vs "What projects should I start?" (brainstorming)
**Root Cause:** No deterministic intent classification

#### Issue 4: Parameter Confusion
**Symptom:** LLM alternates between server_id and server_name despite clear descriptions
**Example:** Sometimes uses server_name correctly, sometimes falls back to server_id
**Root Cause:** LLM making probabilistic choices

---

## Current State Analysis

### Architecture Flow (Current)

```
User Input
  ↓
TUI → POST /api/llm/chat
  ↓
Agent Gateway (llm.go)
  ├─ buildProgramsContext() → Shows ALL programs to LLM
  ├─ Sends to Ollama with programs list
  ↓
LLM decides (probabilistically):
  - Use program? (yes/no)
  - Which program? (task-manager, lab-monitor)
  - Which parameters? (server_id vs server_name)
  ↓
Returns: {message, program_id, program_params}
  ↓
TUI validates program exists
  ├─ If exists → Execute
  └─ If not exists → Ignore, show chat only
```

### Pain Points

1. **No control over program exposure** - All programs always shown
2. **Reactive validation only** - Can only validate after LLM decides
3. **No intent pre-filtering** - Can't prevent wrong decisions
4. **LLM has too much choice** - Increases probability of wrong choice

---

## Proposed Solution

### Military Command System

#### Core Concept
Use distinctive military terminology as **intent markers** to control program exposure.

#### Design Principles

1. **Explicit Intent**
   - Military keywords = "I want a program"
   - No military keywords = "Just chat"

2. **Domain Separation**
   - Each domain has unique terms (no overlap)
   - Lab Monitor: SITREP, CASREP, BDA
   - Task Manager: OPSUM, TASKORD, FRAGO, AAR

3. **Context Awareness**
   - Terms work with/without scoping
   - "SITREP" = all servers
   - "SITREP on teton" = specific server

4. **Backwards Compatibility**
   - Natural language still works as fallback
   - Gradual learning curve
   - User can mix styles

### Military Command Vocabulary

#### Lab Monitor Commands

| Term | Full Name | Scope | Action | Example |
|------|-----------|-------|--------|---------|
| **SITREP** | Situation Report | All or Specific | Health check + status | "SITREP" or "SITREP on teton" |
| **CASREP** | Casualty Report | All only | Show offline servers | "What's the CASREP?" |
| **BDA** | Battle Damage Assessment | Specific only | Health check history | "BDA on glacier" |

#### Task Manager Commands

| Term | Full Name | Scope | Action | Example |
|------|-----------|-------|--------|---------|
| **OPSUM** | Operations Summary | All or Context | List projects | "OPSUM" or "OPSUM on property" |
| **TASKORD** | Task Order | Specific only | Project details | "TASKORD on Property Renovation" |
| **FRAGO** | Fragmentary Order | N/A | Add task/project | "FRAGO: add Tent Pad to property" |
| **AAR** | After Action Review | Specific only | Journal entries | "AAR on Solar Array" |

#### Reserved for Future

| Term | Full Name | Reserved For |
|------|-----------|--------------|
| **QRF** | Quick Reaction Force | Future program (emergency response?) |

---

## Implementation Phases

### Phase 1: Intent Detection Layer
**Goal:** Add keyword detection before LLM processing
**Files Modified:** `handlers/llm.go`
**Risk:** Low (additive only)

#### Tasks

1. **Create Intent Detector**
   ```go
   // handlers/intent.go (NEW FILE)
   type Intent struct {
       Type     string   // "lab-monitor", "task-manager", "chat"
       Keywords []string // Detected military terms
       Context  string   // Extracted server/project names
   }

   func DetectIntent(message string) *Intent
   ```

2. **Integrate into LLM Handler**
   ```go
   // In llm.go Chat() method:
   intent := DetectIntent(req.Message)

   if intent.Type == "chat" {
       // Don't show ANY programs
       programsContext = ""
   } else if intent.Type == "lab-monitor" {
       // Show ONLY lab-monitor
       programsContext = buildLabMonitorContext()
   } else if intent.Type == "task-manager" {
       // Show ONLY task-manager
       programsContext = buildTaskManagerContext()
   }
   ```

3. **Keyword Detection Logic**
   - Case-insensitive matching
   - Whole word matching (avoid false positives)
   - Support variations ("SITREP", "sitrep", "situation report")

4. **Testing**
   - Unit tests for intent detection
   - Test all military keywords
   - Test edge cases (mixed keywords, typos)

**Deliverables:**
- [ ] intent.go file with detector
- [ ] Modified llm.go with integration
- [ ] Unit tests for intent detection
- [ ] Manual testing via TUI

**Estimated Time:** 2-3 hours

---

### Phase 2: Context Extraction
**Goal:** Parse server names, project names from messages
**Files Modified:** `handlers/intent.go`
**Risk:** Low (no existing behavior changed)

#### Tasks

1. **Server Name Extraction**
   - Known server list: yellowstone, yosemite, badlands, joshuatree, sequoia, arches, mammoth, glacier, teton
   - Case-insensitive matching
   - Handle variations ("teton" vs "Teton")

2. **Context/Project Name Extraction**
   - Known contexts: property, house, pxp, personal
   - Extract quoted project names: "Property Renovation"
   - Handle natural language: "on property" → context="property"

3. **Enhanced Intent Object**
   ```go
   type Intent struct {
       Type        string   // "lab-monitor", "task-manager", "chat"
       Keywords    []string // ["SITREP"]
       ServerName  string   // "teton" (if detected)
       Context     string   // "property" (if detected)
       ProjectName string   // "Property Renovation" (if detected)
       Scope       string   // "all" or "specific"
   }
   ```

**Deliverables:**
- [ ] Enhanced intent detection with name extraction
- [ ] Tests for extraction logic
- [ ] Documentation of extraction rules

**Estimated Time:** 2-3 hours

---

### Phase 3: Enhanced System Prompts
**Goal:** Update LLM prompts to work with military commands
**Files Modified:** `handlers/llm.go`
**Risk:** Medium (changes LLM behavior)

#### Tasks

1. **Military Command Documentation**
   - When lab-monitor shown: explain SITREP, CASREP, BDA
   - When task-manager shown: explain OPSUM, TASKORD, FRAGO, AAR
   - Include examples of how to use each term

2. **Simplified Rules**
   - Since programs are pre-filtered, fewer rules needed
   - Focus on execution, not "when to use"

3. **Context Hints**
   - If server_name detected: "User is asking about server: {name}"
   - If context detected: "User is asking about context: {context}"

**Deliverables:**
- [ ] Updated system prompts
- [ ] Testing with various commands
- [ ] Prompt refinement based on results

**Estimated Time:** 2-3 hours

---

### Phase 4: TUI Help System
**Goal:** In-TUI cheat sheet for military commands
**Files Modified:** `ui/app.go`, `ui/help.go` (new)
**Risk:** Low (UI addition only)

#### Tasks

1. **Help Panel Component**
   - Bubble Tea component showing command reference
   - Filterable/searchable
   - Organized by domain

2. **Hotkey Integration**
   - F1 or Ctrl+H to toggle help
   - /help command alternative
   - Persistent across sessions (remember if user had it open)

3. **Dynamic Content**
   - Show all available military commands
   - Show available servers
   - Show available contexts

**Deliverables:**
- [ ] Help panel component
- [ ] Hotkey integration
- [ ] Command reference display
- [ ] User testing

**Estimated Time:** 3-4 hours

---

### Phase 5: Logging and Monitoring
**Goal:** Track intent detection accuracy
**Files Modified:** `handlers/llm.go`, logging infrastructure
**Risk:** Low (observability only)

#### Tasks

1. **Intent Logging**
   ```
   [INTENT] Detected: lab-monitor, Keywords: [SITREP], Server: teton
   [INTENT] Detected: chat, Keywords: [], Server: (none)
   ```

2. **Metrics Tracking**
   - How often each keyword is used
   - Accuracy of name extraction
   - Fallback to natural language frequency

3. **Debug Mode**
   - Show intent detection in TUI status bar
   - Optional verbose logging

**Deliverables:**
- [ ] Intent logging
- [ ] Debug mode in TUI
- [ ] Metrics collection

**Estimated Time:** 1-2 hours

---

## Testing Strategy

### Unit Tests

**Intent Detection (`handlers/intent_test.go`):**
```go
func TestDetectIntent_SITREP(t *testing.T) {
    intent := DetectIntent("SITREP on teton")
    assert.Equal(t, "lab-monitor", intent.Type)
    assert.Contains(t, intent.Keywords, "SITREP")
    assert.Equal(t, "teton", intent.ServerName)
    assert.Equal(t, "specific", intent.Scope)
}

func TestDetectIntent_Chat(t *testing.T) {
    intent := DetectIntent("I need some name ideas")
    assert.Equal(t, "chat", intent.Type)
    assert.Empty(t, intent.Keywords)
}
```

### Integration Tests

**End-to-End via TUI:**
1. Pure chat (no keywords) → No programs shown, natural response
2. SITREP → lab-monitor overview
3. SITREP on [server] → lab-monitor specific server
4. OPSUM → task-manager list projects
5. Mixed keywords → First keyword wins (or error)
6. Typos/variations → Graceful fallback

### Manual Testing Checklist

- [ ] Each military keyword works correctly
- [ ] Server name extraction accurate
- [ ] Context extraction accurate
- [ ] No keywords = pure chat (most important!)
- [ ] Help system accessible and clear
- [ ] Natural language fallback still works
- [ ] Edge cases (typos, mixed case, etc.)

---

## Rollback Plan

### Safety Measures

1. **Feature Flag**
   ```go
   // config.yaml
   experimental:
     military_commands: false  // Set to true to enable
   ```

2. **Gradual Rollout**
   - Phase 1: Test locally only
   - Phase 2: Deploy to remote but disabled
   - Phase 3: Enable on remote
   - Phase 4: Make default behavior

3. **Quick Rollback**
   ```go
   // If feature flag is false, use old behavior:
   if !config.Experimental.MilitaryCommands {
       // Original buildProgramsContext() - show all programs
   }
   ```

### Rollback Triggers

- Intent detection has >20% false positive rate
- User confusion/frustration
- Performance degradation
- LLM behavior worse than before

---

## Success Criteria

### Must Have (Phase 1-3)

1. ✅ **No false program calls**
   - User asks for name ideas → pure chat, no programs
   - Success rate: 95%+

2. ✅ **Accurate intent detection**
   - Military keyword → correct program shown
   - Success rate: 98%+

3. ✅ **Server/context extraction works**
   - "SITREP on teton" → correctly extracts "teton"
   - Success rate: 95%+

4. ✅ **Backwards compatibility**
   - Natural language still works for those who prefer it
   - No regression in existing behavior

### Nice to Have (Phase 4-5)

1. ✅ **Help system is useful**
   - User can find commands quickly
   - Reduces need to remember all terms

2. ✅ **Metrics show adoption**
   - User starts using military commands
   - Reduction in LLM confusion

---

## Future Considerations

### Additional Commands

As new programs are added:
- Define new military terms for each domain
- Update help system automatically
- Maintain clear separation between domains

### Voice Interface

Military commands would work VERY well with voice:
- Distinctive, easy to recognize
- Short, efficient
- No ambiguity

### Team Usage

If system ever becomes multi-user:
- Military commands provide consistency
- New users learn from help system
- Standard operating procedures (SOPs)

### Expansion Ideas

- **COMMS CHECK** - Test connectivity to services
- **EXFIL** - Export/backup data
- **INTEL** - Search/query across all programs
- **MISSION** - Complex multi-program workflows

---

## Risk Assessment

| Risk | Severity | Likelihood | Mitigation |
|------|----------|------------|------------|
| Intent detection too strict | Medium | Medium | Maintain natural language fallback |
| User forgets commands | Low | High | Built-in help system (F1) |
| Keyword conflicts with normal speech | Low | Low | Choose distinctive military terms |
| Implementation bugs | Medium | Medium | Comprehensive testing, feature flag |
| LLM still misbehaves | Low | Low | Programs pre-filtered, less choice |
| Performance impact | Low | Low | Simple regex matching is fast |

---

## Timeline Estimate

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 1: Intent Detection | 2-3 hours | None |
| Phase 2: Context Extraction | 2-3 hours | Phase 1 |
| Phase 3: System Prompts | 2-3 hours | Phase 1, 2 |
| Phase 4: TUI Help System | 3-4 hours | None (parallel) |
| Phase 5: Logging | 1-2 hours | Phase 1 |
| **Total** | **10-15 hours** | |

**Recommended approach:** Implement Phase 1-3 first, test thoroughly, then add Phase 4-5.

---

## Open Questions

1. Should we support command aliases? (e.g., "STAT" as alias for "SITREP")
2. How do we handle multiple keywords in one message? (e.g., "SITREP on teton and OPSUM on property")
3. Should help system be always visible, toggle, or command-only?
4. Do we want autocomplete for military commands in TUI?
5. Should we log all commands for later analysis?

---

## Appendices

### Appendix A: Full Command Reference

See separate file: `MILITARY_COMMANDS_REFERENCE.md`

### Appendix B: Code Examples

See implementation phases above.

### Appendix C: LLM Prompt Templates

To be created during Phase 3.

---

**Document Version:** 1.0
**Last Updated:** 2025-11-25
**Next Review:** After Phase 1 completion
