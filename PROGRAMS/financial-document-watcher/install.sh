#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}  Financial Document Watcher - Installation${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    echo -e "${RED}Error: main.go not found. Please run this script from the financial-document-watcher directory.${NC}"
    exit 1
fi

# Prompt for installation directory
echo -e "${YELLOW}Where should the binary be installed?${NC}"
read -p "Install path [/usr/local/bin]: " INSTALL_DIR
INSTALL_DIR=${INSTALL_DIR:-/usr/local/bin}

# Prompt for config directory
echo ""
echo -e "${YELLOW}Where should the configuration be stored?${NC}"
read -p "Config directory [~/.config/financial-watcher]: " CONFIG_DIR
CONFIG_DIR=${CONFIG_DIR:-~/.config/financial-watcher}
CONFIG_DIR=$(eval echo "$CONFIG_DIR")  # Expand ~

# Prompt for database directory
echo ""
echo -e "${YELLOW}Where should the database be stored?${NC}"
read -p "Database directory [~/.local/share/financial-watcher]: " DB_DIR
DB_DIR=${DB_DIR:-~/.local/share/financial-watcher}
DB_DIR=$(eval echo "$DB_DIR")  # Expand ~

# Prompt for log directory
echo ""
echo -e "${YELLOW}Where should logs be stored?${NC}"
read -p "Log directory [~/.local/log]: " LOG_DIR
LOG_DIR=${LOG_DIR:-~/.local/log}
LOG_DIR=$(eval echo "$LOG_DIR")  # Expand ~

# Prompt for cron schedule
echo ""
echo -e "${YELLOW}Cron schedule (leave empty to skip cron setup):${NC}"
echo "Examples:"
echo "  0 9 * * *       - Daily at 9 AM"
echo "  0 * * * *       - Every hour"
echo "  */30 * * * *    - Every 30 minutes"
read -p "Cron schedule: " CRON_SCHEDULE

echo ""
echo -e "${GREEN}Summary:${NC}"
echo "  Binary: $INSTALL_DIR/financial-document-watcher"
echo "  Config: $CONFIG_DIR/watches.json"
echo "  Database: $DB_DIR/watcher.db"
echo "  Logs: $LOG_DIR/financial-watcher.log"
if [ -n "$CRON_SCHEDULE" ]; then
    echo "  Cron: $CRON_SCHEDULE"
else
    echo "  Cron: Not configured"
fi
echo ""
read -p "Proceed with installation? (y/N): " CONFIRM

if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
    echo "Installation cancelled."
    exit 0
fi

echo ""
echo -e "${GREEN}Installing Go dependencies...${NC}"
go mod tidy

echo -e "${GREEN}Building binary...${NC}"
go build -o financial-document-watcher

echo -e "${GREEN}Creating directories...${NC}"
mkdir -p "$CONFIG_DIR"
mkdir -p "$DB_DIR"
mkdir -p "$LOG_DIR"

echo -e "${GREEN}Installing binary...${NC}"
if [ "$INSTALL_DIR" = "/usr/local/bin" ] || [ "$INSTALL_DIR" = "/usr/bin" ]; then
    sudo cp financial-document-watcher "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/financial-document-watcher"
else
    cp financial-document-watcher "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/financial-document-watcher"
fi

echo -e "${GREEN}Setting up configuration...${NC}"
if [ ! -f "$CONFIG_DIR/watches.json" ]; then
    cp watches.json.example "$CONFIG_DIR/watches.json"
    echo -e "${YELLOW}Created config file: $CONFIG_DIR/watches.json${NC}"
    echo -e "${YELLOW}IMPORTANT: Edit this file to configure your watch paths!${NC}"
else
    echo -e "${YELLOW}Config file already exists, skipping: $CONFIG_DIR/watches.json${NC}"
fi

# Create a wrapper script if using custom paths
if [ "$CONFIG_DIR" != "." ] || [ "$DB_DIR" != "." ]; then
    WRAPPER_SCRIPT="$INSTALL_DIR/financial-document-watcher-run"

    echo -e "${GREEN}Creating wrapper script...${NC}"

    WRAPPER_CONTENT="#!/bin/bash
# Auto-generated wrapper script for financial-document-watcher
exec $INSTALL_DIR/financial-document-watcher \\
    -config \"$CONFIG_DIR/watches.json\" \\
    -db \"$DB_DIR/watcher.db\" \\
    \"\$@\"
"

    if [ "$INSTALL_DIR" = "/usr/local/bin" ] || [ "$INSTALL_DIR" = "/usr/bin" ]; then
        echo "$WRAPPER_CONTENT" | sudo tee "$WRAPPER_SCRIPT" > /dev/null
        sudo chmod +x "$WRAPPER_SCRIPT"
    else
        echo "$WRAPPER_CONTENT" > "$WRAPPER_SCRIPT"
        chmod +x "$WRAPPER_SCRIPT"
    fi

    EXEC_COMMAND="$WRAPPER_SCRIPT"
else
    EXEC_COMMAND="$INSTALL_DIR/financial-document-watcher"
fi

# Set up cron if requested
if [ -n "$CRON_SCHEDULE" ]; then
    echo -e "${GREEN}Setting up cron job...${NC}"

    CRON_COMMAND="$CRON_SCHEDULE $EXEC_COMMAND >> $LOG_DIR/financial-watcher.log 2>&1"

    # Check if cron job already exists
    if crontab -l 2>/dev/null | grep -q "financial-document-watcher\|financial-watcher"; then
        echo -e "${YELLOW}Cron job already exists. Please manually update your crontab if needed.${NC}"
        echo "Suggested entry:"
        echo "  $CRON_COMMAND"
    else
        # Add to crontab
        (crontab -l 2>/dev/null; echo "$CRON_COMMAND") | crontab -
        echo -e "${GREEN}Cron job added successfully!${NC}"
    fi

    echo ""
    echo "View your crontab with: crontab -l"
    echo "Edit your crontab with: crontab -e"
fi

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}  Installation Complete!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo "Next steps:"
echo ""
echo "1. Edit your watches configuration:"
echo "   ${YELLOW}nano $CONFIG_DIR/watches.json${NC}"
echo ""
echo "2. Create your processor executables and update paths in watches.json"
echo ""
echo "3. Test the watcher:"
echo "   ${YELLOW}$EXEC_COMMAND -dry-run${NC}"
echo ""
echo "4. Run manually:"
echo "   ${YELLOW}$EXEC_COMMAND${NC}"
echo ""
echo "5. View logs:"
echo "   ${YELLOW}tail -f $LOG_DIR/financial-watcher.log${NC}"
echo ""
if [ -n "$CRON_SCHEDULE" ]; then
    echo "Cron is configured to run: $CRON_SCHEDULE"
    echo "Next cron run: $(date -d \"now + 1 hour\" '+%Y-%m-%d %H:00:00')"
fi
echo ""
