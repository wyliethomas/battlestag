#!/usr/bin/env bash
# check_dependencies.sh - Check for required dependencies

set -euo pipefail

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}╔════════════════════════════════════════════╗${NC}"
echo -e "${YELLOW}║   BATTLESTAG Dependency Check             ║${NC}"
echo -e "${YELLOW}╚════════════════════════════════════════════╝${NC}"
echo ""

MISSING=()
OPTIONAL_MISSING=()

# Required for primitives
echo -e "${YELLOW}Checking required dependencies...${NC}"

check_command() {
    local cmd="$1"
    local name="${2:-$cmd}"
    if command -v "$cmd" &>/dev/null; then
        echo -e "  ${GREEN}✓${NC} $name"
        return 0
    else
        echo -e "  ${RED}✗${NC} $name"
        return 1
    fi
}

# Core requirements
check_command "bash" "bash (4.0+)" || MISSING+=("bash")
check_command "cat" "coreutils" || MISSING+=("coreutils")
check_command "ls" || true
check_command "grep" || true
check_command "find" || true
check_command "chmod" || true

# Optional for utilities
echo ""
echo -e "${YELLOW}Checking optional dependencies...${NC}"

check_command "jq" "jq (for JSON utilities)" || OPTIONAL_MISSING+=("jq")
check_command "curl" "curl (for LLM integration)" || OPTIONAL_MISSING+=("curl")

# Check bash version
echo ""
echo -e "${YELLOW}Checking versions...${NC}"
BASH_VERSION_NUM="${BASH_VERSION%%[^0-9]*}"
if [[ "$BASH_VERSION_NUM" -ge 4 ]]; then
    echo -e "  ${GREEN}✓${NC} bash version: $BASH_VERSION"
else
    echo -e "  ${RED}✗${NC} bash version: $BASH_VERSION (need 4.0+)"
    MISSING+=("bash-4.0+")
fi

# Summary
echo ""
echo -e "${YELLOW}╔════════════════════════════════════════════╗${NC}"
echo -e "${YELLOW}║   Summary                                  ║${NC}"
echo -e "${YELLOW}╚════════════════════════════════════════════╝${NC}"

if [[ ${#MISSING[@]} -eq 0 ]]; then
    echo -e "${GREEN}✓ All required dependencies installed${NC}"
else
    echo -e "${RED}✗ Missing required dependencies:${NC}"
    for dep in "${MISSING[@]}"; do
        echo -e "  - $dep"
    done
fi

if [[ ${#OPTIONAL_MISSING[@]} -eq 0 ]]; then
    echo -e "${GREEN}✓ All optional dependencies installed${NC}"
else
    echo -e "${YELLOW}⚠ Missing optional dependencies:${NC}"
    for dep in "${OPTIONAL_MISSING[@]}"; do
        echo -e "  - $dep"
    done
fi

# Installation hints
if [[ ${#MISSING[@]} -gt 0 ]] || [[ ${#OPTIONAL_MISSING[@]} -gt 0 ]]; then
    echo ""
    echo -e "${YELLOW}Installation hints:${NC}"
    echo ""

    if [[ -f /etc/debian_version ]]; then
        echo -e "  ${GREEN}# Ubuntu/Debian${NC}"
        echo "  sudo apt-get update"
        [[ ${#MISSING[@]} -gt 0 ]] && echo "  sudo apt-get install bash coreutils"
        [[ " ${OPTIONAL_MISSING[@]} " =~ " jq " ]] && echo "  sudo apt-get install jq"
        [[ " ${OPTIONAL_MISSING[@]} " =~ " curl " ]] && echo "  sudo apt-get install curl"
    elif [[ -f /etc/redhat-release ]]; then
        echo -e "  ${GREEN}# RHEL/CentOS/Fedora${NC}"
        echo "  sudo yum update"
        [[ ${#MISSING[@]} -gt 0 ]] && echo "  sudo yum install bash coreutils"
        [[ " ${OPTIONAL_MISSING[@]} " =~ " jq " ]] && echo "  sudo yum install jq"
        [[ " ${OPTIONAL_MISSING[@]} " =~ " curl " ]] && echo "  sudo yum install curl"
    elif [[ -f /etc/arch-release ]]; then
        echo -e "  ${GREEN}# Arch Linux${NC}"
        [[ ${#MISSING[@]} -gt 0 ]] && echo "  sudo pacman -S bash coreutils"
        [[ " ${OPTIONAL_MISSING[@]} " =~ " jq " ]] && echo "  sudo pacman -S jq"
        [[ " ${OPTIONAL_MISSING[@]} " =~ " curl " ]] && echo "  sudo pacman -S curl"
    elif [[ "$(uname)" == "Darwin" ]]; then
        echo -e "  ${GREEN}# macOS${NC}"
        [[ " ${OPTIONAL_MISSING[@]} " =~ " jq " ]] && echo "  brew install jq"
        [[ " ${OPTIONAL_MISSING[@]} " =~ " curl " ]] && echo "  brew install curl"
    fi
    echo ""
fi

# Exit with error if required dependencies missing
if [[ ${#MISSING[@]} -gt 0 ]]; then
    exit 1
fi

exit 0
