#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}  Financial Document Watcher - Update${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    echo -e "${RED}Error: main.go not found. Please run this script from the financial-document-watcher directory.${NC}"
    exit 1
fi

# Try to detect where it was installed by looking for the wrapper script
INSTALL_DIR=""
CONFIG_DIR=""

if [ -f "/usr/local/bin/financial-document-watcher-run" ]; then
    INSTALL_DIR="/usr/local/bin"
    # Extract config dir from wrapper script
    CONFIG_DIR=$(grep "config" /usr/local/bin/financial-document-watcher-run | grep -oP '(?<=-config ")[^"]+' | head -1 | xargs dirname)
elif [ -f "$HOME/.local/bin/financial-document-watcher-run" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    CONFIG_DIR=$(grep "config" "$HOME/.local/bin/financial-document-watcher-run" | grep -oP '(?<=-config ")[^"]+' | head -1 | xargs dirname)
fi

if [ -z "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Could not auto-detect installation directory.${NC}"
    echo -e "${YELLOW}Where is financial-document-watcher currently installed?${NC}"
    read -p "Install path [/usr/local/bin]: " INSTALL_DIR
    INSTALL_DIR=${INSTALL_DIR:-/usr/local/bin}
fi

if [ -z "$CONFIG_DIR" ]; then
    echo -e "${YELLOW}Could not auto-detect config directory.${NC}"
    read -p "Config directory [~/.config/financial-watcher]: " CONFIG_DIR
    CONFIG_DIR=${CONFIG_DIR:-~/.config/financial-watcher}
    CONFIG_DIR=$(eval echo "$CONFIG_DIR")
fi

echo "Detected installation:"
echo "  Binary: $INSTALL_DIR"
echo "  Config: $CONFIG_DIR"
echo ""

if [ ! -f "$INSTALL_DIR/financial-document-watcher" ]; then
    echo -e "${RED}Error: financial-document-watcher not found in $INSTALL_DIR${NC}"
    echo "Run ./install.sh first to do initial installation."
    exit 1
fi

# Check what needs updating
UPDATE_BINARY=false
UPDATE_CONFIG=false

echo -e "${YELLOW}What would you like to update?${NC}"
echo ""
read -p "Update binary? (Y/n): " UPDATE_BINARY_INPUT
UPDATE_BINARY_INPUT=${UPDATE_BINARY_INPUT:-y}
if [[ $UPDATE_BINARY_INPUT =~ ^[Yy]$ ]]; then
    UPDATE_BINARY=true
fi

# Check if watches.json exists in repo
if [ -f "watches.json" ]; then
    read -p "Update watches.json config? (Y/n): " UPDATE_CONFIG_INPUT
    UPDATE_CONFIG_INPUT=${UPDATE_CONFIG_INPUT:-y}
    if [[ $UPDATE_CONFIG_INPUT =~ ^[Yy]$ ]]; then
        UPDATE_CONFIG=true
    fi
else
    echo -e "${YELLOW}Note: watches.json not found in repo (only watches.json.example exists)${NC}"
    echo "      Config will not be updated."
fi

if [ "$UPDATE_BINARY" = false ] && [ "$UPDATE_CONFIG" = false ]; then
    echo "Nothing to update. Exiting."
    exit 0
fi

echo ""
echo -e "${GREEN}Summary:${NC}"
[ "$UPDATE_BINARY" = true ] && echo "  ✓ Update binary: $INSTALL_DIR/financial-document-watcher"
[ "$UPDATE_CONFIG" = true ] && echo "  ✓ Update config: $CONFIG_DIR/watches.json"
echo ""
read -p "Proceed with update? (y/N): " CONFIRM

if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
    echo "Update cancelled."
    exit 0
fi

echo ""

# Update binary
if [ "$UPDATE_BINARY" = true ]; then
    echo -e "${GREEN}Installing Go dependencies...${NC}"
    go mod tidy

    echo -e "${GREEN}Building binary...${NC}"
    go build -o financial-document-watcher

    echo -e "${GREEN}Updating binary...${NC}"
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] || [ "$INSTALL_DIR" = "/usr/bin" ]; then
        sudo cp financial-document-watcher "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/financial-document-watcher"
    else
        cp financial-document-watcher "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/financial-document-watcher"
    fi
    echo "  ✓ Binary updated"
fi

# Update config
if [ "$UPDATE_CONFIG" = true ]; then
    echo ""
    echo -e "${GREEN}Updating configuration...${NC}"

    # Backup existing config if it exists
    if [ -f "$CONFIG_DIR/watches.json" ]; then
        BACKUP_FILE="$CONFIG_DIR/watches.json.backup.$(date +%Y%m%d-%H%M%S)"
        cp "$CONFIG_DIR/watches.json" "$BACKUP_FILE"
        echo "  Backed up existing config to: $BACKUP_FILE"
    fi

    # Copy new config
    cp watches.json "$CONFIG_DIR/watches.json"
    echo "  ✓ Config updated: $CONFIG_DIR/watches.json"
fi

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}  Update Complete!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""

if [ "$UPDATE_BINARY" = true ]; then
    echo "Binary updated: $INSTALL_DIR/financial-document-watcher"
fi

if [ "$UPDATE_CONFIG" = true ]; then
    echo "Config updated: $CONFIG_DIR/watches.json"
fi

echo ""
echo "Test the update:"
echo "  ${YELLOW}financial-document-watcher-run --dry-run${NC}"
echo ""
