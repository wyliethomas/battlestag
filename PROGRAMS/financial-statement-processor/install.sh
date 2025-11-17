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

# Prompt for database configuration
echo ""
echo -e "${YELLOW}PostgreSQL Database Configuration:${NC}"
read -p "DB Host [localhost]: " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "DB Port [5432]: " DB_PORT
DB_PORT=${DB_PORT:-5432}

read -p "DB User [postgres]: " DB_USER
DB_USER=${DB_USER:-postgres}

read -sp "DB Password (required): " DB_PASSWORD
echo ""

read -p "DB Name [financial_data]: " DB_NAME
DB_NAME=${DB_NAME:-financial_data}

read -p "DB SSL Mode [disable]: " DB_SSLMODE
DB_SSLMODE=${DB_SSLMODE:-disable}

if [ -z "$DB_PASSWORD" ]; then
    echo -e "${RED}Error: Database password is required${NC}"
    exit 1
fi

# Ask about database setup
echo ""
echo -e "${YELLOW}Do you want to run the database schema setup now?${NC}"
echo "This will create tables, indexes, and views in the $DB_NAME database."
read -p "Run schema setup? (y/N): " RUN_SCHEMA
RUN_SCHEMA=${RUN_SCHEMA:-n}

echo ""
echo -e "${GREEN}Summary:${NC}"
echo "  Processor Binary: $INSTALL_DIR/financial-statement-processor"
echo "  Query Binary: $INSTALL_DIR/financial-statement-query"
echo "  Config: $CONFIG_DIR/.env"
echo "  Database: $DB_HOST:$DB_PORT/$DB_NAME"
if [[ $RUN_SCHEMA =~ ^[Yy]$ ]]; then
    echo "  Schema: Will be installed"
else
    echo "  Schema: Skip (manual setup required)"
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

echo -e "${GREEN}Building processor binary...${NC}"
go build -o financial-statement-processor ./cmd/processor

echo -e "${GREEN}Building query binary...${NC}"
go build -o financial-statement-query ./cmd/query

echo -e "${GREEN}Creating config directory...${NC}"
mkdir -p "$CONFIG_DIR"

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

# PostgreSQL Database Configuration
DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_NAME=$DB_NAME
DB_SSLMODE=$DB_SSLMODE
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

# Run schema setup if requested
if [[ $RUN_SCHEMA =~ ^[Yy]$ ]]; then
    echo ""
    echo -e "${GREEN}Setting up database schema...${NC}"
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f schema.sql

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Database schema installed successfully!${NC}"
    else
        echo -e "${YELLOW}WARNING: Schema installation failed. You may need to run it manually:${NC}"
        echo "  PGPASSWORD='***' psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f schema.sql"
    fi
fi

echo ""
echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}  Installation Complete!${NC}"
echo -e "${GREEN}====================================================${NC}"
echo ""
echo "Next steps:"
echo ""
echo "1. Test database connection:"
echo "   ${YELLOW}$WRAPPER_QUERY --summary${NC}"
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
echo ""
if [[ ! $RUN_SCHEMA =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}NOTE: Database schema not installed. Run manually:${NC}"
    echo "  PGPASSWORD='$DB_PASSWORD' psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f schema.sql"
    echo ""
fi
