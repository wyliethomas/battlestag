#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${RED}====================================================${NC}"
echo -e "${RED}  Financial Statement Processor - Uninstall${NC}"
echo -e "${RED}====================================================${NC}"
echo ""

# Prompt for installation directory
echo -e "${YELLOW}Where were the binaries installed?${NC}"
read -p "Install path [/usr/local/bin]: " INSTALL_DIR
INSTALL_DIR=${INSTALL_DIR:-/usr/local/bin}

# Prompt for config directory
echo ""
echo -e "${YELLOW}Where is the configuration stored?${NC}"
read -p "Config directory [~/.config/financial-processor]: " CONFIG_DIR
CONFIG_DIR=${CONFIG_DIR:-~/.config/financial-processor}
CONFIG_DIR=$(eval echo "$CONFIG_DIR")  # Expand ~

echo ""
echo -e "${YELLOW}What to remove:${NC}"
echo "  Processor Binary: $INSTALL_DIR/financial-statement-processor"
echo "  Query Binary: $INSTALL_DIR/financial-statement-query"
echo "  Processor Wrapper: $INSTALL_DIR/financial-statement-processor-run"
echo "  Query Wrapper: $INSTALL_DIR/financial-statement-query-run"

FOUND_ITEMS=""

# Check what exists
if [ -f "$INSTALL_DIR/financial-statement-processor" ]; then
    FOUND_ITEMS="${FOUND_ITEMS}processor "
fi

if [ -f "$INSTALL_DIR/financial-statement-query" ]; then
    FOUND_ITEMS="${FOUND_ITEMS}query "
fi

if [ -f "$INSTALL_DIR/financial-statement-processor-run" ]; then
    FOUND_ITEMS="${FOUND_ITEMS}processor-wrapper "
fi

if [ -f "$INSTALL_DIR/financial-statement-query-run" ]; then
    FOUND_ITEMS="${FOUND_ITEMS}query-wrapper "
fi

if [ -d "$CONFIG_DIR" ]; then
    FOUND_ITEMS="${FOUND_ITEMS}config "
    echo "  Config: $CONFIG_DIR (will prompt)"
fi

if [ -z "$FOUND_ITEMS" ]; then
    echo ""
    echo -e "${YELLOW}Nothing found to uninstall. Already clean!${NC}"
    exit 0
fi

echo ""
read -p "Proceed with uninstall? (y/N): " CONFIRM

if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
    echo "Uninstall cancelled."
    exit 0
fi

echo ""

# Remove processor binary
if [ -f "$INSTALL_DIR/financial-statement-processor" ]; then
    echo -e "${GREEN}Removing processor binary...${NC}"
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] || [ "$INSTALL_DIR" = "/usr/bin" ]; then
        sudo rm -f "$INSTALL_DIR/financial-statement-processor"
    else
        rm -f "$INSTALL_DIR/financial-statement-processor"
    fi
    echo "  Removed: $INSTALL_DIR/financial-statement-processor"
fi

# Remove query binary
if [ -f "$INSTALL_DIR/financial-statement-query" ]; then
    echo -e "${GREEN}Removing query binary...${NC}"
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] || [ "$INSTALL_DIR" = "/usr/bin" ]; then
        sudo rm -f "$INSTALL_DIR/financial-statement-query"
    else
        rm -f "$INSTALL_DIR/financial-statement-query"
    fi
    echo "  Removed: $INSTALL_DIR/financial-statement-query"
fi

# Remove wrapper scripts
if [ -f "$INSTALL_DIR/financial-statement-processor-run" ]; then
    echo -e "${GREEN}Removing processor wrapper...${NC}"
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] || [ "$INSTALL_DIR" = "/usr/bin" ]; then
        sudo rm -f "$INSTALL_DIR/financial-statement-processor-run"
    else
        rm -f "$INSTALL_DIR/financial-statement-processor-run"
    fi
    echo "  Removed: $INSTALL_DIR/financial-statement-processor-run"
fi

if [ -f "$INSTALL_DIR/financial-statement-query-run" ]; then
    echo -e "${GREEN}Removing query wrapper...${NC}"
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] || [ "$INSTALL_DIR" = "/usr/bin" ]; then
        sudo rm -f "$INSTALL_DIR/financial-statement-query-run"
    else
        rm -f "$INSTALL_DIR/financial-statement-query-run"
    fi
    echo "  Removed: $INSTALL_DIR/financial-statement-query-run"
fi

# Ask about config
if [ -d "$CONFIG_DIR" ]; then
    echo ""
    echo -e "${YELLOW}Configuration directory exists: $CONFIG_DIR${NC}"
    echo "WARNING: This contains your database password!"
    read -p "Remove configuration? (y/N): " REMOVE_CONFIG
    if [[ $REMOVE_CONFIG =~ ^[Yy]$ ]]; then
        rm -rf "$CONFIG_DIR"
        echo "  Removed: $CONFIG_DIR"
    else
        echo "  Kept: $CONFIG_DIR"
    fi
fi

# Ask about database data
echo ""
echo -e "${YELLOW}Do you want to remove the database tables and data?${NC}"
echo "WARNING: This will delete all transaction data!"
read -p "Remove database tables? (y/N): " REMOVE_DB

if [[ $REMOVE_DB =~ ^[Yy]$ ]]; then
    echo ""
    echo "To remove database tables, run:"
    echo "  PGPASSWORD='your_password' psql -h host -p port -U user -d financial_data -c 'DROP TABLE IF EXISTS transactions CASCADE; DROP TABLE IF EXISTS processing_log CASCADE;'"
    echo ""
    echo "Or to drop the entire database:"
    echo "  PGPASSWORD='your_password' psql -h host -p port -U user -c 'DROP DATABASE IF EXISTS financial_data;'"
fi

echo ""
echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}  Uninstall Complete!${NC}"
echo -e "${GREEN}====================================================${NC}"
echo ""
