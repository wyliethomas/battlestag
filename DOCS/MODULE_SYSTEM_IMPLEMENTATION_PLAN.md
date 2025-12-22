# Apache-Style Module System Implementation Plan

**Date**: 2025-12-22
**Status**: Planning
**Pattern**: Apache a2enmod/a2dismod style

---

## 🎯 Vision

Build a module management system that:
- Scales to hundreds of modules
- Uses familiar Apache-style enable/disable pattern
- Supports user customization (modules-custom/)
- Zero changes to agent-gateway (just point scanner at different dir)
- Git-based distribution (pull updates like code)

---

## 📁 Target Directory Structure

```
~/battlestag/                              # New unified location
├── modules-available/                     # Git repository (all official modules)
│   ├── .git/
│   ├── README.md
│   ├── tier1/
│   │   ├── text/
│   │   │   ├── improve_writing.sh
│   │   │   ├── format_code.sh
│   │   │   └── fix_grammar.sh
│   │   ├── code/
│   │   │   ├── extract_todos.sh
│   │   │   └── simple_explain.sh
│   │   └── data/
│   │       └── csv_insight.sh           # Move existing module here
│   ├── tier2/
│   │   ├── wisdom/
│   │   │   ├── extract_wisdom.sh
│   │   │   └── summarize_research.sh
│   │   ├── analysis/
│   │   │   ├── security_audit.sh
│   │   │   └── architecture_review.sh
│   │   └── creative/
│   │       └── write_essay.sh
│   └── tier3/
│       ├── code_review_auto.sh
│       └── summarize_adaptive.sh
│
├── modules-enabled/                       # Symlinks (agent-gateway scans this)
│   ├── csv_insight.sh -> ../modules-available/tier1/data/csv_insight.sh
│   ├── improve_writing.sh -> ../modules-available/tier1/text/improve_writing.sh
│   └── my_custom.sh -> ../modules-custom/my_custom.sh
│
├── modules-custom/                        # User's own modules (git ignored)
│   ├── README.md                         # Instructions for creating custom modules
│   ├── .gitignore                        # Ignore all *.sh in this dir
│   └── my_custom.sh                      # User-created module
│
├── primitives/                            # Existing (move from SCRIPTS/)
│   ├── file.sh
│   ├── dir.sh
│   ├── search.sh
│   └── sys.sh
│
├── utilities/                             # Existing (move from SCRIPTS/)
│   ├── file.sh
│   ├── json.sh
│   ├── log.sh
│   ├── llm.sh
│   └── data.sh
│
├── bin/                                   # Module management scripts
│   ├── bstag-enmod                       # Enable module
│   ├── bstag-dismod                      # Disable module
│   ├── bstag-module                      # Main module command
│   └── bstag-module-scaffold             # Create custom module template
│
└── config/
    ├── modules.state                     # Optional: track enabled modules
    └── profiles/                         # Optional: saved module configs
        ├── default.conf
        └── work.conf
```

---

## 🏗️ Implementation Phases

### **Phase 1: Directory Migration & Structure** (Week 1, Day 1-2)

**Goal**: Set up new directory structure and migrate existing modules

**Tasks**:
1. Create `~/battlestag/` directory structure
2. Move `SCRIPTS/primitives/` → `~/battlestag/primitives/`
3. Move `SCRIPTS/utilities/` → `~/battlestag/utilities/`
4. Create `~/battlestag/modules-available/` with tier structure
5. Move existing `csv_insight.sh` to `modules-available/tier1/data/`
6. Create empty `modules-enabled/` and `modules-custom/` directories
7. Update all script source paths to use new locations

**Deliverables**:
- [ ] New directory structure created
- [ ] All scripts moved and paths updated
- [ ] Existing tests still pass (run test suites)
- [ ] Document migration in README

**Migration Script**: `SCRIPTS/migrate-to-battlestag.sh`
```bash
#!/usr/bin/env bash
# Migrates SCRIPTS/ to ~/battlestag/ structure
# Preserves existing functionality while creating new layout
```

---

### **Phase 2: Core Module Management Scripts** (Week 1, Day 3-4)

**Goal**: Build `bstag-enmod`, `bstag-dismod`, and `bstag-module` commands

**Tasks**:
1. **Create `bstag-enmod`**:
   - Search for module in modules-available/
   - Check if already enabled
   - Create symlink in modules-enabled/
   - Validate module metadata
   - Show success message with restart instructions

2. **Create `bstag-dismod`**:
   - Check if module is enabled
   - Remove symlink from modules-enabled/
   - Show success message

3. **Create `bstag-module`** (main command):
   - Subcommands: list, available, enable, disable, info, update
   - `bstag module list` - Show enabled modules
   - `bstag module available` - Show all available modules
   - `bstag module info <name>` - Show module metadata
   - `bstag module enable <name>` - Wrapper for bstag-enmod
   - `bstag module disable <name>` - Wrapper for bstag-dismod
   - `bstag module update` - Git pull modules-available/

4. **Create helper functions**:
   - `find_module()` - Search for module by name
   - `parse_module_metadata()` - Extract MODULE, TIER, etc.
   - `validate_module()` - Check required metadata present
   - `list_dependencies()` - Check REQUIRES field (future)

**Deliverables**:
- [ ] `~/battlestag/bin/bstag-enmod` (executable)
- [ ] `~/battlestag/bin/bstag-dismod` (executable)
- [ ] `~/battlestag/bin/bstag-module` (main CLI)
- [ ] Help text for all commands
- [ ] Basic error handling

**Example Usage**:
```bash
# Enable a module
bstag-enmod csv_insight
# or
bstag module enable csv_insight

# Disable a module
bstag-dismod csv_insight
# or
bstag module disable csv_insight

# List enabled
bstag module list

# List available
bstag module available

# Get module info
bstag module info extract_wisdom
# → Name: Extract Wisdom
# → Tier: 2
# → Category: wisdom
# → Description: Extract insights from content
# → Status: Not enabled
```

---

### **Phase 3: Agent-Gateway Integration** (Week 1, Day 5)

**Goal**: Update agent-gateway to scan `modules-enabled/` instead of `SCRIPTS/modules/`

**Tasks**:
1. Update `main.go`:
   ```go
   // Before:
   modulesDir := "./SCRIPTS/modules"

   // After:
   modulesDir := filepath.Join(os.Getenv("HOME"), "battlestag", "modules-enabled")
   ```

2. Update module scanner to handle flat directory (no subdirs)
   - modules-enabled/ is flat (all symlinks at root level)
   - Scanner should still parse metadata from linked files

3. Test module discovery:
   - Enable csv_insight module
   - Start agent-gateway
   - Verify module appears in /api/programs/list
   - Execute module via API

4. Update TUI to refresh module list:
   - TUI should query API for available programs
   - Display enabled modules in command palette

**Deliverables**:
- [ ] Updated `ORCHESTRATOR/agent-gateway/main.go`
- [ ] Updated module scanner (if needed)
- [ ] Test: Module shows in API
- [ ] Test: Module executes via API
- [ ] Test: Module appears in TUI

---

### **Phase 4: Custom Module Support** (Week 1, Day 6-7)

**Goal**: Allow users to create and manage custom modules

**Tasks**:
1. **Create `bstag-module-scaffold`**:
   - Generates module template in `modules-custom/`
   - Prompts for: name, tier, category, description
   - Creates boilerplate with proper metadata
   - Opens in $EDITOR

2. **Create `bstag module create <name>`**:
   - Wrapper for bstag-module-scaffold
   - Auto-enables module after creation

3. **Create `bstag module customize <name>`**:
   - Copies module from modules-available/ to modules-custom/
   - Disables original symlink
   - Creates new symlink to custom version
   - Opens in $EDITOR
   - Preserves metadata comments

4. **Create README in modules-custom/**:
   - Explain how custom modules work
   - Show example module structure
   - Document how to source primitives/utilities

**Deliverables**:
- [ ] `~/battlestag/bin/bstag-module-scaffold`
- [ ] `bstag module create` subcommand
- [ ] `bstag module customize` subcommand
- [ ] `~/battlestag/modules-custom/README.md`
- [ ] Example custom module

**Example Workflow**:
```bash
# Create new custom module
bstag module create my_analyzer
# → Created ~/battlestag/modules-custom/my_analyzer.sh
# → Enabled module
# → Opening in editor...

# Customize existing module
bstag module customize extract_wisdom
# → Copied to ~/battlestag/modules-custom/extract_wisdom.sh
# → Switched to custom version
# → Opening in editor...
```

---

### **Phase 5: Module Repository Setup** (Week 2, Day 1-2)

**Goal**: Create official module repository on GitHub

**Tasks**:
1. Create new GitHub repo: `battlestag/battlestag-modules`
2. Set up directory structure:
   ```
   battlestag-modules/
   ├── README.md
   ├── CONTRIBUTING.md
   ├── tier1/
   ├── tier2/
   ├── tier3/
   └── templates/
       └── module-template.sh
   ```

3. Migrate existing modules to repo
4. Document module metadata format
5. Create contribution guidelines
6. Add GitHub Actions for module validation:
   - Check metadata present
   - Run shellcheck
   - Validate tier classification

**Deliverables**:
- [ ] GitHub repo created and public
- [ ] README with installation instructions
- [ ] CONTRIBUTING guide for module authors
- [ ] Module template
- [ ] CI/CD for validation

**Installation for Users**:
```bash
# Clone modules repository
cd ~/battlestag
git clone https://github.com/battlestag/battlestag-modules.git modules-available

# Update modules
cd ~/battlestag/modules-available
git pull
```

---

### **Phase 6: First Tier 1 & Tier 2 Modules** (Week 2, Day 3-5)

**Goal**: Build first working tiered modules

**Tasks**:
1. **Update `utilities/llm.sh`**:
   - Add `util.llm.query_small()` - Uses LM Studio (existing)
   - Add `util.llm.query_big()` - Uses Claude API (NEW)
   - Add config for API keys (`~/.config/battlestag/llm.conf`)

2. **Build Tier 1: `improve_writing.sh`**:
   - Simple text improvement
   - Uses local LLM (free)
   - Fast execution (< 5 seconds)
   - Input: text file
   - Output: improved text

3. **Build Tier 2: `extract_wisdom.sh`**:
   - Fabric-inspired pattern
   - Uses Claude API (paid)
   - Rich output (insights, quotes, ideas)
   - Input: article/transcript
   - Output: structured markdown

4. **Test both modules**:
   - Enable via bstag-enmod
   - Execute via agent-gateway API
   - Execute via TUI
   - Verify tier routing works

**Deliverables**:
- [ ] Enhanced `utilities/llm.sh` with big brain support
- [ ] `modules-available/tier1/text/improve_writing.sh`
- [ ] `modules-available/tier2/wisdom/extract_wisdom.sh`
- [ ] Test suite for both modules
- [ ] Documentation with examples

---

### **Phase 7: Documentation & Polish** (Week 2, Day 6-7)

**Goal**: Complete documentation and user-facing polish

**Tasks**:
1. **Create comprehensive README**:
   - Overview of module system
   - Installation instructions
   - Common commands cheatsheet
   - Tier explanation (when to use 1 vs 2)
   - Custom module guide

2. **Create video demo** (optional):
   - Show enabling/disabling modules
   - Create custom module
   - Execute via TUI

3. **Update project documentation**:
   - Update CLAUDE.md with new structure
   - Update TIERED_MODULES_RAG_PLAN.md status
   - Update PROJECT_PROGRESS.md

4. **Add shell completions** (nice-to-have):
   - Bash completion for bstag-module commands
   - Auto-complete module names

**Deliverables**:
- [ ] `~/battlestag/README.md` (comprehensive)
- [ ] Updated project docs
- [ ] Bash completions (optional)
- [ ] Video demo (optional)

---

## 📊 Success Criteria

### Phase 1-2 (Core System):
- ✅ Can enable/disable modules via CLI
- ✅ Symlinks work correctly
- ✅ Existing csv_insight module works in new location
- ✅ All tests pass after migration

### Phase 3-4 (Integration):
- ✅ Agent-gateway discovers enabled modules
- ✅ TUI shows enabled modules
- ✅ Can execute modules via API
- ✅ Can create custom modules

### Phase 5-6 (Modules):
- ✅ Module repository on GitHub
- ✅ Tier 1 module works (local LLM)
- ✅ Tier 2 module works (Claude API)
- ✅ Can update modules via git pull

### Phase 7 (Polish):
- ✅ Documentation complete
- ✅ User can follow README to install
- ✅ Contribution guide available

---

## 🚀 Quick Start (After Implementation)

**For Users**:
```bash
# Install battlestag module system
cd ~
git clone https://github.com/battlestag/battlestag.git
cd battlestag
./install.sh

# Install official modules
git clone https://github.com/battlestag/battlestag-modules.git modules-available

# Enable some modules
bstag module enable improve_writing
bstag module enable csv_insight

# Start using
bstag-tui
# → Command: ai:improve_writing my-doc.txt
```

**For Module Authors**:
```bash
# Create custom module
bstag module create my_module

# Edit
vim ~/battlestag/modules-custom/my_module.sh

# Test
bstag module enable my_module
bstag-tui
```

---

## 📋 Open Questions

### Before Starting:

1. **Unified location**: Move everything to `~/battlestag/` or keep SCRIPTS/ in repo?
   - **Proposal**: Keep SCRIPTS/ in repo for development, install to ~/battlestag/ for deployment
   - Deploy script copies SCRIPTS/ → ~/battlestag/

2. **PATH setup**: Should we add `~/battlestag/bin` to PATH?
   - **Proposal**: Yes, install script adds to ~/.bashrc

3. **Existing modules**: What to do with csv_insight?
   - **Proposal**: Move to modules-available/tier1/data/, keep as example

4. **Claude API key**: Where to store?
   - **Proposal**: `~/.config/battlestag/llm.conf` with encryption option

5. **Module versioning**: Do we need it in Phase 1?
   - **Proposal**: No, just use git for now, add later if needed

---

## 🎯 Next Session

**Tonight**: Review plan, answer open questions
**Tomorrow**: Start Phase 1 - Directory migration

**Key Decision Points**:
1. Confirm directory structure (~/battlestag/ vs SCRIPTS/)
2. Decide on Claude API setup (key storage, provider)
3. Priority modules to build first

---

*This plan gets us from current state to production-ready Apache-style module system in ~2 weeks.*
