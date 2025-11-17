#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${RED}================================================${NC}"
echo -e "${RED}  Financial Document Watcher - Uninstall${NC}"
echo -e "${RED}================================================${NC}"
echo ""

# Prompt for installation directory
echo -e "${YELLOW}Where was the binary installed?${NC}"
read -p "Install path [/usr/local/bin]: " INSTALL_DIR
INSTALL_DIR=${INSTALL_DIR:-/usr/local/bin}

# Prompt for config directory
echo ""
echo -e "${YELLOW}Where is the configuration stored?${NC}"
read -p "Config directory [~/.config/financial-watcher]: " CONFIG_DIR
CONFIG_DIR=${CONFIG_DIR:-~/.config/financial-watcher}
CONFIG_DIR=$(eval echo "$CONFIG_DIR")  # Expand ~

# Prompt for database directory
echo ""
echo -e "${YELLOW}Where is the database stored?${NC}"
read -p "Database directory [~/.local/share/financial-watcher]: " DB_DIR
DB_DIR=${DB_DIR:-~/.local/share/financial-watcher}
DB_DIR=$(eval echo "$DB_DIR")  # Expand ~

# Prompt for log directory
echo ""
echo -e "${YELLOW}Where are logs stored?${NC}"
read -p "Log directory [~/.local/log]: " LOG_DIR
LOG_DIR=${LOG_DIR:-~/.local/log}
LOG_DIR=$(eval echo "$LOG_DIR")  # Expand ~

echo ""
echo -e "${YELLOW}What to remove:${NC}"
echo "  Binary: $INSTALL_DIR/financial-document-watcher"
echo "  Wrapper: $INSTALL_DIR/financial-document-watcher-run"

FOUND_ITEMS=""

# Check what exists
if [ -f "$INSTALL_DIR/financial-document-watcher" ]; then
    FOUND_ITEMS="${FOUND_ITEMS}binary "
fi

if [ -f "$INSTALL_DIR/financial-document-watcher-run" ]; then
    FOUND_ITEMS="${FOUND_ITEMS}wrapper "
fi

if crontab -l 2>/dev/null | grep -q "financial-document-watcher\|financial-watcher"; then
    FOUND_ITEMS="${FOUND_ITEMS}cron "
    echo "  Cron job: Found (will be removed)"
else
    echo "  Cron job: Not found"
fi

if [ -d "$CONFIG_DIR" ]; then
    FOUND_ITEMS="${FOUND_ITEMS}config "
    echo "  Config: $CONFIG_DIR (will prompt)"
fi

if [ -d "$DB_DIR" ]; then
    FOUND_ITEMS="${FOUND_ITEMS}database "
    echo "  Database: $DB_DIR (will prompt)"
fi

if [ -f "$LOG_DIR/financial-watcher.log" ]; then
    FOUND_ITEMS="${FOUND_ITEMS}logs "
    echo "  Logs: $LOG_DIR/financial-watcher.log (will prompt)"
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

# Remove binary
if [ -f "$INSTALL_DIR/financial-document-watcher" ]; then
    echo -e "${GREEN}Removing binary...${NC}"
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] || [ "$INSTALL_DIR" = "/usr/bin" ]; then
        sudo rm -f "$INSTALL_DIR/financial-document-watcher"
    else
        rm -f "$INSTALL_DIR/financial-document-watcher"
    fi
    echo "  Removed: $INSTALL_DIR/financial-document-watcher"
fi

# Remove wrapper script
if [ -f "$INSTALL_DIR/financial-document-watcher-run" ]; then
    echo -e "${GREEN}Removing wrapper script...${NC}"
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] || [ "$INSTALL_DIR" = "/usr/bin" ]; then
        sudo rm -f "$INSTALL_DIR/financial-document-watcher-run"
    else
        rm -f "$INSTALL_DIR/financial-document-watcher-run"
    fi
    echo "  Removed: $INSTALL_DIR/financial-document-watcher-run"
fi

# Remove cron job
if crontab -l 2>/dev/null | grep -q "financial-document-watcher\|financial-watcher"; then
    echo -e "${GREEN}Removing cron job...${NC}"
    crontab -l 2>/dev/null | grep -v "financial-document-watcher\|financial-watcher" | crontab -
    echo "  Removed cron job"
fi

# Ask about config
if [ -d "$CONFIG_DIR" ]; then
    echo ""
    echo -e "${YELLOW}Configuration directory exists: $CONFIG_DIR${NC}"
    read -p "Remove configuration? (y/N): " REMOVE_CONFIG
    if [[ $REMOVE_CONFIG =~ ^[Yy]$ ]]; then
        rm -rf "$CONFIG_DIR"
        echo "  Removed: $CONFIG_DIR"
    else
        echo "  Kept: $CONFIG_DIR"
    fi
fi

# Ask about database
if [ -d "$DB_DIR" ]; then
    echo ""
    echo -e "${YELLOW}Database directory exists: $DB_DIR${NC}"
    read -p "Remove database? (y/N): " REMOVE_DB
    if [[ $REMOVE_DB =~ ^[Yy]$ ]]; then
        rm -rf "$DB_DIR"
        echo "  Removed: $DB_DIR"
    else
        echo "  Kept: $DB_DIR"
    fi
fi

# Ask about logs
if [ -f "$LOG_DIR/financial-watcher.log" ]; then
    echo ""
    echo -e "${YELLOW}Log file exists: $LOG_DIR/financial-watcher.log${NC}"
    read -p "Remove logs? (y/N): " REMOVE_LOGS
    if [[ $REMOVE_LOGS =~ ^[Yy]$ ]]; then
        rm -f "$LOG_DIR/financial-watcher.log"
        echo "  Removed: $LOG_DIR/financial-watcher.log"
    else
        echo "  Kept: $LOG_DIR/financial-watcher.log"
    fi
fi

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}  Uninstall Complete!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
