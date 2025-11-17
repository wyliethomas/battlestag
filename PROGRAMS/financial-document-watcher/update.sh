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
if [ -f "/usr/local/bin/financial-document-watcher-run" ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ -f "$HOME/.local/bin/financial-document-watcher-run" ]; then
    INSTALL_DIR="$HOME/.local/bin"
fi

if [ -z "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Could not auto-detect installation directory.${NC}"
    echo -e "${YELLOW}Where is financial-document-watcher currently installed?${NC}"
    read -p "Install path [/usr/local/bin]: " INSTALL_DIR
    INSTALL_DIR=${INSTALL_DIR:-/usr/local/bin}
fi

echo "Detected installation: $INSTALL_DIR"
echo ""

if [ ! -f "$INSTALL_DIR/financial-document-watcher" ]; then
    echo -e "${RED}Error: financial-document-watcher not found in $INSTALL_DIR${NC}"
    echo "Run ./install.sh first to do initial installation."
    exit 1
fi

echo -e "${YELLOW}This will rebuild and update the binary at:${NC}"
echo "  $INSTALL_DIR/financial-document-watcher"
echo ""
echo -e "${YELLOW}Your configuration will NOT be changed.${NC}"
echo ""
read -p "Proceed with update? (y/N): " CONFIRM

if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
    echo "Update cancelled."
    exit 0
fi

echo ""
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

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}  Update Complete!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo "Binary updated: $INSTALL_DIR/financial-document-watcher"
echo "Configuration unchanged"
echo ""
echo "Test the update:"
echo "  ${YELLOW}financial-document-watcher-run --dry-run${NC}"
echo ""
