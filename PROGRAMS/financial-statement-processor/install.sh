#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}  Financial Statement Processor - Installation${NC}"
echo -e "${GREEN}====================================================${NC}"
echo ""

# Check if we're in the right directory
if [ ! -f "schema.sql" ]; then
    echo -e "${RED}Error: schema.sql not found. Please run this script from the financial-statement-processor directory.${NC}"
    exit 1
fi

# Prompt for installation directory
echo -e "${YELLOW}Where should the binaries be installed?${NC}"
read -p "Install path [/usr/local/bin]: " INSTALL_DIR
INSTALL_DIR=${INSTALL_DIR:-/usr/local/bin}

# Prompt for .env file location
echo ""
echo -e "${YELLOW}Where should the .env configuration be stored?${NC}"
read -p "Config directory [~/.config/financial-processor]: " CONFIG_DIR
CONFIG_DIR=${CONFIG_DIR:-~/.config/financial-processor}
CONFIG_DIR=$(eval echo "$CONFIG_DIR")  # Expand ~

# Prompt for database file location
echo ""
echo -e "${YELLOW}Where should the SQLite database be stored?${NC}"
read -p "Database path [~/.local/share/financial-processor/transactions.db]: " DB_PATH
DB_PATH=${DB_PATH:-~/.local/share/financial-processor/transactions.db}
DB_PATH=$(eval echo "$DB_PATH")  # Expand ~

echo ""
echo -e "${GREEN}Summary:${NC}"
echo "  Processor Binary: $INSTALL_DIR/financial-statement-processor"
echo "  Query Binary: $INSTALL_DIR/financial-statement-query"
echo "  Config: $CONFIG_DIR/.env"
echo "  Database: $DB_PATH"
echo ""
read -p "Proceed with installation? (y/N): " CONFIRM

if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
    echo "Installation cancelled."
    exit 0
fi

echo ""
echo -e "${GREEN}Installing Go dependencies...${NC}"
go mod tidy

echo -e "${GREEN}Building processor binary...${NC}"
go build -o financial-statement-processor ./cmd/processor

echo -e "${GREEN}Building query binary...${NC}"
go build -o financial-statement-query ./cmd/query

echo -e "${GREEN}Creating config directory...${NC}"
mkdir -p "$CONFIG_DIR"

# Create database directory
DB_DIR=$(dirname "$DB_PATH")
echo -e "${GREEN}Creating database directory...${NC}"
mkdir -p "$DB_DIR"

echo -e "${GREEN}Installing binaries...${NC}"
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

echo -e "${GREEN}Creating .env configuration...${NC}"
cat > "$CONFIG_DIR/.env" <<EOF
# Financial Statement Processor Configuration
# Generated on $(date)

# SQLite Database File Path
DB_PATH=$DB_PATH
EOF

chmod 600 "$CONFIG_DIR/.env"
echo "  Created: $CONFIG_DIR/.env"

# Create a wrapper script that loads the .env file
WRAPPER_PROCESSOR="$INSTALL_DIR/financial-statement-processor-run"
WRAPPER_QUERY="$INSTALL_DIR/financial-statement-query-run"

echo -e "${GREEN}Creating wrapper scripts...${NC}"

WRAPPER_CONTENT_PROCESSOR="#!/bin/bash
# Auto-generated wrapper script for financial-statement-processor
set -a
source \"$CONFIG_DIR/.env\"
set +a
exec $INSTALL_DIR/financial-statement-processor \"\$@\"
"

WRAPPER_CONTENT_QUERY="#!/bin/bash
# Auto-generated wrapper script for financial-statement-query
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

echo ""
echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}  Installation Complete!${NC}"
echo -e "${GREEN}====================================================${NC}"
echo ""
echo "Next steps:"
echo ""
echo "1. Test database initialization:"
echo "   ${YELLOW}$WRAPPER_QUERY --summary${NC}"
echo "   (This will create the database and tables automatically)"
echo ""
echo "2. Process a statement file:"
echo "   ${YELLOW}$WRAPPER_PROCESSOR /path/to/statement.pdf${NC}"
echo ""
echo "3. Query transactions:"
echo "   ${YELLOW}$WRAPPER_QUERY --start-date 2024-10-01 --end-date 2024-10-31 --pretty${NC}"
echo ""
echo "4. Update watches.json in financial-document-watcher to use:"
echo "   ${YELLOW}\"executable_path\": \"$WRAPPER_PROCESSOR\"${NC}"
echo ""
echo "Configuration stored in: $CONFIG_DIR/.env"
echo "Database will be created at: $DB_PATH"
echo ""
echo -e "${GREEN}NOTE: Database tables are created automatically on first use.${NC}"
echo "      No manual schema setup required!"
echo ""
