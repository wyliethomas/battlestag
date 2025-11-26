# BATTLESTAG Military Commands Reference

**Quick Reference Guide for Military Command System**

---

## Quick Start

### The Basic Rule
- **Use military keywords** → Run a program
- **No military keywords** → Just chat

### Examples
```
"SITREP on teton" → Lab Monitor runs health check on teton
"I need some name ideas" → Pure chat, no programs
```

---

## Lab Monitor Commands

**Domain:** Server infrastructure and health monitoring

### SITREP (Situation Report)
**Purpose:** Get current status and health information

**Usage:**
```
"SITREP" → Health check all servers, show overview
"SITREP on teton" → Health check specific server
"Give me a SITREP" → Same as SITREP
"I need a SITREP on badlands" → Check badlands
```

**What it does:**
- Runs ping and SSH health checks
- Shows response times
- Updates server status (online/offline/unknown)
- Returns current state

**Maps to:**
- Broad: `lab_health check-all` + `lab_query overview`
- Specific: `lab_health check-server --name [server]`

---

### CASREP (Casualty Report)
**Purpose:** Show which servers are offline/unreachable

**Usage:**
```
"CASREP" → List all offline servers
"What's the CASREP?" → Same
"Give me the CASREP" → Same
```

**What it does:**
- Queries for servers with "offline" status
- Shows when they were last seen
- Helps identify problems quickly

**Maps to:**
- `lab_query offline`

**Note:** CASREP is "all servers only" - doesn't make sense for a single server

---

### BDA (Battle Damage Assessment)
**Purpose:** View health check history for a specific server

**Usage:**
```
"BDA on glacier" → Show health history for glacier
"I need a BDA on teton" → Show health history for teton
"BDA on yellowstone, last 20" → Show last 20 checks
```

**What it does:**
- Shows recent health check results
- Includes timestamps, check types (ping/ssh)
- Shows response times and errors
- Helps diagnose intermittent issues

**Maps to:**
- `lab_health history --name [server] --limit [N]`

**Note:** BDA requires a specific server name

---

## Task Manager Commands

**Domain:** Project and task management

### OPSUM (Operations Summary)
**Purpose:** List projects and get overview

**Usage:**
```
"OPSUM" → List all projects across all contexts
"OPSUM on property" → List only property projects
"Give me an OPSUM" → List all projects
"OPSUM for house" → List house projects
```

**What it does:**
- Lists projects with status
- Shows goal, checklist completion
- Can filter by context (property, house, pxp, personal)

**Maps to:**
- Broad: `task_project list`
- Context: `task_project list --context [context]`

---

### TASKORD (Task Order)
**Purpose:** Get detailed status of a specific project

**Usage:**
```
"TASKORD on Property Renovation" → Show project details
"I need a TASKORD for Solar Array" → Show project details
"TASKORD property renovation" → Show project details
```

**What it does:**
- Shows project details (goal, status, dates)
- Lists all checklist items with completion
- Shows recent journal entries
- Complete project snapshot

**Maps to:**
- `task_query status --name "[project_name]"`

**Note:** TASKORD requires a specific project name

---

### FRAGO (Fragmentary Order)
**Purpose:** Add new tasks, projects, or checklist items

**Usage:**
```
"FRAGO: add Tent Pad to property" → Create new project
"FRAGO: add checklist items to Solar Array: install panels, wire batteries" → Add checklist
"FRAGO: journal entry for Workshop - completed framing" → Add journal
```

**What it does:**
- Creates new projects
- Adds checklist items
- Adds journal entries
- Flexible creation command

**Maps to:**
- Add project: `task_project add --name [name] --context [context]`
- Add checklist: `task_checklist add --project [id] --items [items]`
- Add journal: `task_journal add --project [id] --entry [entry]`

---

### AAR (After Action Review)
**Purpose:** View journal entries for a project

**Usage:**
```
"AAR on Solar Array" → Show recent journal entries
"I need an AAR for Property Renovation" → Show journal
"AAR for Workshop, last 5" → Show last 5 entries
```

**What it does:**
- Shows chronological journal entries
- Includes dates and content
- Helps review project history

**Maps to:**
- `task_journal list --project [id] --limit [N]`

**Note:** AAR requires a specific project name

---

## Server Names

Your lab servers (case-insensitive):

| Server Name | IP Address | Notes |
|-------------|------------|-------|
| yellowstone | 192.168.1.39 | Local mini |
| yosemite | 192.168.1.83 | Local mini |
| badlands | 192.168.1.241 | Local mini |
| joshuatree | 192.168.1.159 | Local mini |
| sequoia | 192.168.1.92 | Local mini |
| arches | 192.168.1.123 | Local mini |
| mammoth | 192.168.1.180 | Local mini |
| glacier | 192.168.1.77 | Local mini |
| teton | 192.168.1.232 | AI Lab |

---

## Context Names

Project contexts (case-insensitive):
- **property** - Property/land projects
- **house** - House/building projects
- **pxp** - PXP-related work
- **personal** - Personal projects

---

## Reserved Commands

Commands reserved for future programs:

| Term | Full Name | Reserved For |
|------|-----------|--------------|
| **QRF** | Quick Reaction Force | Future emergency response program |

---

## Tips and Tricks

### Mixing Natural Language
You can use military keywords in natural sentences:
```
"Can you give me a SITREP on teton?"
"I'd like to see the CASREP"
"Show me an OPSUM for the property context"
```

### Case Insensitive
All keywords work in any case:
```
"SITREP" = "sitrep" = "Sitrep"
"teton" = "TETON" = "Teton"
```

### Scoping
Some commands work both broad and specific:
```
"SITREP" → All servers
"SITREP on glacier" → Just glacier

"OPSUM" → All projects
"OPSUM on property" → Just property projects
```

### Default Behavior
When you DON'T use military keywords, you get pure chat:
```
"I need some ideas for a web agency name"
→ Battlestag responds naturally, no programs called
```

---

## Common Patterns

### Morning Check-In
```
"SITREP" → See all server status
"CASREP" → Any problems?
"OPSUM on property" → What projects am I working on?
```

### Troubleshooting
```
"SITREP on badlands" → Is it responding?
"BDA on badlands" → When did it go down?
```

### Project Management
```
"OPSUM on house" → What house projects do I have?
"TASKORD on Workshop" → What's the status?
"AAR on Workshop" → What's been happening?
"FRAGO: add new checklist items..." → Add tasks
```

---

## Getting Help

### In the TUI
Press **F1** or type **/help** to see this reference

### Learning the Commands
1. Start with just SITREP and OPSUM
2. Add CASREP when you need to check offline servers
3. Learn the others as you need them
4. Keep this reference handy

### Fallback
Don't remember a command? Use natural language - it still works!
```
"show me the property projects" → Still works
"check health on teton" → Still works
```

---

## Command Comparison

### Before vs After

**Before (Natural Language):**
```
"Show me all servers" → LLM might call lab-monitor, might not
"Give me ideas for names" → LLM might call task-manager incorrectly
```

**After (Military Commands):**
```
"SITREP" → Always calls lab-monitor
"Give me ideas for names" → Never calls programs, just chat
"OPSUM" → Always calls task-manager
```

**Benefit:** Predictable, explicit, no guessing

---

## Vocabulary Origins

Understanding the terms helps remember them:

| Term | Origin | Why It Fits |
|------|--------|-------------|
| SITREP | Military situation report | Server status = situation |
| CASREP | Navy casualty report | Offline servers = casualties |
| BDA | Military damage assessment | Health history = assessing damage |
| OPSUM | Military operations summary | Projects = operations |
| TASKORD | Military task order | Project details = task assignment |
| FRAGO | Military fragmentary order | Quick updates/additions |
| AAR | Military after action review | Journal = reviewing actions |

---

**Version:** 1.0
**Last Updated:** 2025-11-25
**See Also:** MILITARY_COMMAND_SYSTEM_PLAN.md
