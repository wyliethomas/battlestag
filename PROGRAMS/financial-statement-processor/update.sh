#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}  Financial Statement Processor - Update${NC}"
echo -e "${GREEN}====================================================${NC}"
echo ""

# Check if we're in the right directory
if [ ! -f "schema.sql" ]; then
    echo -e "${RED}Error: schema.sql not found. Please run this script from the financial-statement-processor directory.${NC}"
    exit 1
fi

# Try to detect where it was installed by looking for the wrapper script
INSTALL_DIR=""
WRAPPER_PATH=""
if [ -f "/usr/local/bin/financial-statement-processor-run" ]; then
    INSTALL_DIR="/usr/local/bin"
    WRAPPER_PATH="/usr/local/bin/financial-statement-processor-run"
elif [ -f "$HOME/.local/bin/financial-statement-processor-run" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    WRAPPER_PATH="$HOME/.local/bin/financial-statement-processor-run"
fi

if [ -z "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Could not auto-detect installation directory.${NC}"
    echo -e "${YELLOW}Where are the binaries currently installed?${NC}"
    read -p "Install path [/usr/local/bin]: " INSTALL_DIR
    INSTALL_DIR=${INSTALL_DIR:-/usr/local/bin}
    WRAPPER_PATH="$INSTALL_DIR/financial-statement-processor-run"
fi

echo "Detected installation: $INSTALL_DIR"

# Extract CONFIG_DIR from existing wrapper script
CONFIG_DIR=""
if [ -f "$WRAPPER_PATH" ]; then
    # Extract the path from the source line in the wrapper
    CONFIG_DIR=$(grep "source" "$WRAPPER_PATH" | sed 's/.*source "\(.*\)\/\.env".*/\1/')
    echo "Detected config directory: $CONFIG_DIR"
else
    echo -e "${YELLOW}Warning: Wrapper script not found at $WRAPPER_PATH${NC}"
    echo -e "${YELLOW}Where is the .env configuration stored?${NC}"
    read -p "Config directory [~/.config/financial-processor]: " CONFIG_DIR
    CONFIG_DIR=${CONFIG_DIR:-~/.config/financial-processor}
    CONFIG_DIR=$(eval echo "$CONFIG_DIR")  # Expand ~
fi

echo ""

if [ ! -f "$INSTALL_DIR/financial-statement-processor" ]; then
    echo -e "${RED}Error: financial-statement-processor not found in $INSTALL_DIR${NC}"
    echo "Run ./install.sh first to do initial installation."
    exit 1
fi

echo -e "${YELLOW}This will rebuild and update the binaries at:${NC}"
echo "  $INSTALL_DIR/financial-statement-processor"
echo "  $INSTALL_DIR/financial-statement-query"
echo ""
echo -e "${YELLOW}Your configuration and database will NOT be changed.${NC}"
echo ""
read -p "Proceed with update? (y/N): " CONFIRM

if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
    echo "Update cancelled."
    exit 0
fi

echo ""
echo -e "${GREEN}Installing Go dependencies...${NC}"
go mod tidy

echo -e "${GREEN}Building processor binary...${NC}"
go build -o financial-statement-processor ./cmd/processor

echo -e "${GREEN}Building query binary...${NC}"
go build -o financial-statement-query ./cmd/query

echo -e "${GREEN}Updating binaries...${NC}"
if [ "$INSTALL_DIR" = "/usr/local/bin" ] || [ "$INSTALL_DIR" = "/usr/bin" ]; then
    sudo cp financial-statement-processor "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/financial-statement-processor"
    sudo cp financial-statement-query "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/financial-statement-query"
else
    cp financial-statement-processor "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/financial-statement-processor"
    cp financial-statement-query "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/financial-statement-query"
fi

# Recreate wrapper scripts to ensure they have correct paths
echo -e "${GREEN}Recreating wrapper scripts...${NC}"

WRAPPER_PROCESSOR="$INSTALL_DIR/financial-statement-processor-run"
WRAPPER_QUERY="$INSTALL_DIR/financial-statement-query-run"

WRAPPER_CONTENT_PROCESSOR="#!/bin/bash
# Auto-generated wrapper script for financial-statement-processor
# Updated on $(date)
set -a
source \"$CONFIG_DIR/.env\"
set +a
exec $INSTALL_DIR/financial-statement-processor \"\$@\"
"

WRAPPER_CONTENT_QUERY="#!/bin/bash
# Auto-generated wrapper script for financial-statement-query
# Updated on $(date)
set -a
source \"$CONFIG_DIR/.env\"
set +a
exec $INSTALL_DIR/financial-statement-query \"\$@\"
"

if [ "$INSTALL_DIR" = "/usr/local/bin" ] || [ "$INSTALL_DIR" = "/usr/bin" ]; then
    echo "$WRAPPER_CONTENT_PROCESSOR" | sudo tee "$WRAPPER_PROCESSOR" > /dev/null
    sudo chmod +x "$WRAPPER_PROCESSOR"
    echo "$WRAPPER_CONTENT_QUERY" | sudo tee "$WRAPPER_QUERY" > /dev/null
    sudo chmod +x "$WRAPPER_QUERY"
else
    echo "$WRAPPER_CONTENT_PROCESSOR" > "$WRAPPER_PROCESSOR"
    chmod +x "$WRAPPER_PROCESSOR"
    echo "$WRAPPER_CONTENT_QUERY" > "$WRAPPER_QUERY"
    chmod +x "$WRAPPER_QUERY"
fi

echo "  Created: $WRAPPER_PROCESSOR"
echo "  Created: $WRAPPER_QUERY"

echo ""
echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}  Update Complete!${NC}"
echo -e "${GREEN}====================================================${NC}"
echo ""
echo "Binaries updated:"
echo "  $INSTALL_DIR/financial-statement-processor"
echo "  $INSTALL_DIR/financial-statement-query"
echo ""
echo "Wrapper scripts recreated:"
echo "  $WRAPPER_PROCESSOR"
echo "  $WRAPPER_QUERY"
echo ""
echo "Configuration directory: $CONFIG_DIR"
echo "  (Database and .env unchanged)"
echo ""
echo "Test the update:"
echo "  ${YELLOW}financial-statement-query-run --summary${NC}"
echo ""
