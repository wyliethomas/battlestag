#!/usr/bin/env bash
# install.sh - Install BATTLESTAG to ~/.bstag
#
# This script installs the BATTLESTAG module system to the user's home directory.
# It can be run from the repo or via curl for remote installation.

set -euo pipefail

# Installation target
BSTAG_HOME="${HOME}/.bstag"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Parse arguments
UPDATE_MODE=false
FORCE_MODE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --update)
            UPDATE_MODE=true
            shift
            ;;
        --force)
            FORCE_MODE=true
            shift
            ;;
        -h|--help)
            cat << EOF
Usage: install.sh [OPTIONS]

Install BATTLESTAG module system to ~/.bstag

OPTIONS:
  --update    Update existing installation (preserves config and custom modules)
  --force     Force reinstall (overwrites everything)
  -h, --help  Show this help message

EXAMPLES:
  ./install.sh           # Fresh install
  ./install.sh --update  # Update existing install
  ./install.sh --force   # Force reinstall

EOF
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}" >&2
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   BATTLESTAG Module System Installer${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Check if already installed
if [[ -d "$BSTAG_HOME" ]] && [[ "$UPDATE_MODE" == "false" ]] && [[ "$FORCE_MODE" == "false" ]]; then
    echo -e "${YELLOW}BATTLESTAG is already installed at: $BSTAG_HOME${NC}"
    echo ""
    echo "Options:"
    echo "  ./install.sh --update   # Update installation"
    echo "  ./install.sh --force    # Force reinstall"
    echo ""
    exit 1
fi

if [[ "$UPDATE_MODE" == "true" ]]; then
    echo -e "${YELLOW}Updating existing installation...${NC}"
    echo ""
elif [[ "$FORCE_MODE" == "true" ]]; then
    echo -e "${YELLOW}Force reinstalling...${NC}"
    if [[ -d "$BSTAG_HOME" ]]; then
        echo "Backing up existing installation..."
        mv "$BSTAG_HOME" "${BSTAG_HOME}.backup.$(date +%s)"
        echo -e "${GREEN}✓ Backup created${NC}"
    fi
    echo ""
fi

# Create directory structure
echo "Step 1: Creating directory structure..."
mkdir -p "$BSTAG_HOME"/{primitives,utilities,bin,config,config/profiles}
mkdir -p "$BSTAG_HOME"/modules-{available,enabled,custom}
mkdir -p "$BSTAG_HOME"/modules-available/{tier1,tier2,tier3}
mkdir -p "$BSTAG_HOME"/modules-available/tier1/{text,code,data}
mkdir -p "$BSTAG_HOME"/modules-available/tier2/{wisdom,analysis,creative}
mkdir -p "$BSTAG_HOME"/modules-available/tier3
echo -e "${GREEN}✓ Directories created${NC}"

# Copy primitives
echo ""
echo "Step 2: Installing primitives..."
if [[ -d "$SCRIPT_DIR/primitives" ]]; then
    cp "$SCRIPT_DIR"/primitives/*.sh "$BSTAG_HOME"/primitives/ 2>/dev/null || true
    # Don't copy test files
    rm -f "$BSTAG_HOME"/primitives/test_*.sh 2>/dev/null || true
    prim_count=$(find "$BSTAG_HOME/primitives" -name "*.sh" -type f | wc -l)
    echo "  Installed $prim_count primitive modules"
fi
echo -e "${GREEN}✓ Primitives installed${NC}"

# Copy utilities
echo ""
echo "Step 3: Installing utilities..."
if [[ -d "$SCRIPT_DIR/utilities" ]]; then
    cp "$SCRIPT_DIR"/utilities/*.sh "$BSTAG_HOME"/utilities/ 2>/dev/null || true
    # Don't copy test files
    rm -f "$BSTAG_HOME"/utilities/test_*.sh 2>/dev/null || true
    util_count=$(find "$BSTAG_HOME/utilities" -name "*.sh" -type f | wc -l)
    echo "  Installed $util_count utility modules"
fi
echo -e "${GREEN}✓ Utilities installed${NC}"

# Copy modules
echo ""
echo "Step 4: Installing modules..."
if [[ -d "$SCRIPT_DIR/modules" ]]; then
    # Copy entire tier structure
    cp -r "$SCRIPT_DIR"/modules/tier* "$BSTAG_HOME"/modules-available/ 2>/dev/null || true

    # Count modules
    module_count=$(find "$BSTAG_HOME/modules-available" -name "*.sh" -type f | wc -l)
    echo "  Installed $module_count modules"

    # List them
    if [[ $module_count -gt 0 ]]; then
        find "$BSTAG_HOME/modules-available" -name "*.sh" -type f | \
            sed "s|$BSTAG_HOME/modules-available/||" | \
            sed 's|\.sh$||' | \
            sort | \
            sed 's/^/    /'
    fi
fi
echo -e "${GREEN}✓ Modules installed${NC}"

# Copy bin scripts
echo ""
echo "Step 5: Installing commands..."
if [[ -d "$SCRIPT_DIR/bin" ]] && [[ -n "$(ls -A "$SCRIPT_DIR/bin" 2>/dev/null)" ]]; then
    cp "$SCRIPT_DIR"/bin/* "$BSTAG_HOME"/bin/ 2>/dev/null || true
    chmod +x "$BSTAG_HOME"/bin/* 2>/dev/null || true
    bin_count=$(find "$BSTAG_HOME/bin" -type f | wc -l)
    echo "  Installed $bin_count commands"
else
    echo "  No commands to install (will be added in Phase 2)"
fi
echo -e "${GREEN}✓ Commands ready${NC}"

# Copy config files
echo ""
echo "Step 6: Setting up configuration..."
if [[ -d "$SCRIPT_DIR/config" ]]; then
    # Copy example configs (don't overwrite existing)
    if [[ ! -f "$BSTAG_HOME/config/llm.conf" ]] && [[ -f "$SCRIPT_DIR/config/llm.conf.example" ]]; then
        cp "$SCRIPT_DIR/config/llm.conf.example" "$BSTAG_HOME/config/"
        echo "  Installed llm.conf.example (copy to llm.conf and configure)"
    elif [[ -f "$BSTAG_HOME/config/llm.conf" ]]; then
        echo "  Preserved existing llm.conf"
    fi

    # Copy config README
    if [[ -f "$SCRIPT_DIR/config/README.md" ]]; then
        cp "$SCRIPT_DIR/config/README.md" "$BSTAG_HOME/config/"
        echo "  Installed config/README.md"
    fi
fi
echo -e "${GREEN}✓ Configuration files ready${NC}"

# Create README files (only if not updating)
if [[ "$UPDATE_MODE" == "false" ]] || [[ ! -f "$BSTAG_HOME/README.md" ]]; then
    echo ""
    echo "Step 7: Creating documentation..."

    cat > "$BSTAG_HOME/README.md" << 'EOF'
# BATTLESTAG Module System

Apache-style module management for AI-powered automation.

## Quick Start

```bash
# List available modules
bstag-module available

# Enable a module
bstag-enmod csv_insight

# Disable a module
bstag-dismod csv_insight

# List enabled modules
bstag-module list
```

## Directory Structure

- `primitives/` - Core bash functions
- `utilities/` - Composed operations (JSON, LLM, logs)
- `modules-available/` - All official modules (organized by tier)
- `modules-enabled/` - Symlinks to enabled modules
- `modules-custom/` - Your custom modules
- `bin/` - Module management scripts
- `config/` - Configuration files

## Tiers

- **Tier 1**: Local LLM (free, fast) - Simple tasks
- **Tier 2**: Cloud API (paid, quality) - Complex analysis
- **Tier 3**: Adaptive routing - Auto-selects tier

## Updating

```bash
cd /path/to/battlestag-bot-repo
git pull
./SCRIPTS/install.sh --update
```
EOF

    cat > "$BSTAG_HOME/modules-custom/README.md" << 'EOF'
# Custom Modules

This directory is for your own modules. Files here are preserved during updates.

## Module Template

```bash
#!/usr/bin/env bash
# MODULE: my_module
# NAME: My Custom Module
# CATEGORY: custom
# TIER: 1
# DESCRIPTION: Does something awesome

set -euo pipefail

BSTAG_HOME="${BSTAG_HOME:-$HOME/.bstag}"
source "$BSTAG_HOME/utilities/file.sh"
source "$BSTAG_HOME/utilities/llm.sh"

module.my_module() {
    echo "Success!"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    module.my_module "$@"
fi
```

## Enabling

```bash
bstag-enmod my_module
```
EOF

    cat > "$BSTAG_HOME/modules-custom/.gitignore" << 'EOF'
# Ignore all custom modules
*.sh
!README.md
EOF

    echo -e "${GREEN}✓ Documentation created${NC}"
else
    echo ""
    echo "Step 6: Skipping documentation (update mode)"
fi

# Add to PATH
echo ""
echo "Step 8: Configuring PATH..."

# Detect shell
SHELL_RC=""
if [[ -n "${BASH_VERSION:-}" ]]; then
    SHELL_RC="$HOME/.bashrc"
elif [[ -n "${ZSH_VERSION:-}" ]]; then
    SHELL_RC="$HOME/.zshrc"
else
    case "$SHELL" in
        */bash) SHELL_RC="$HOME/.bashrc" ;;
        */zsh) SHELL_RC="$HOME/.zshrc" ;;
    esac
fi

if [[ -n "$SHELL_RC" ]]; then
    if grep -q "/.bstag/bin" "$SHELL_RC" 2>/dev/null; then
        echo "  PATH already configured in $SHELL_RC"
    else
        echo "" >> "$SHELL_RC"
        echo "# BATTLESTAG module system" >> "$SHELL_RC"
        echo 'export PATH="$HOME/.bstag/bin:$PATH"' >> "$SHELL_RC"
        echo -e "${GREEN}✓ Added to PATH in $SHELL_RC${NC}"
        echo "  Reload with: source $SHELL_RC"
    fi
else
    echo -e "${YELLOW}  Could not detect shell${NC}"
    echo "  Manually add: export PATH=\"\$HOME/.bstag/bin:\$PATH\""
fi

# Summary
echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Installation Complete!${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""
echo "Location: $BSTAG_HOME"
echo ""
echo "Next steps:"
if [[ -n "$SHELL_RC" ]]; then
    echo "  1. source $SHELL_RC"
fi
echo "  2. bstag-module available  # (Phase 2)"
echo "  3. bstag-enmod csv_insight # (Phase 2)"
echo ""
