#!/bin/bash
set -e

echo "=== Lab Monitor - Installation ==="

# Configuration
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="${CONFIG_DIR:-$HOME/.config/lab-monitor}"
DB_PATH="${DB_PATH:-$HOME/.local/share/lab-monitor/servers.db}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create directories
echo "Creating directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$(dirname "$DB_PATH")"

# Check for dependencies
echo "Checking dependencies..."
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# Get dependencies
echo "Installing Go dependencies..."
go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto/ssh

# Build executables
echo "Building executables..."
go build -o lab_server ./cmd/server
go build -o lab_health ./cmd/health
go build -o lab_query ./cmd/query

# Install executables
echo "Installing to $INSTALL_DIR..."
cp lab_server "$INSTALL_DIR/"
cp lab_health "$INSTALL_DIR/"
cp lab_query "$INSTALL_DIR/"

# Make executable
chmod +x "$INSTALL_DIR"/lab_*

# Clean up build artifacts
rm -f lab_server lab_health lab_query

# Create config file
cat > "$CONFIG_DIR/.env" <<EOF
# Lab Monitor Configuration
DB_PATH=$DB_PATH
EOF

echo ""
echo -e "${GREEN}✓ Installation complete!${NC}"
echo ""
echo "Executables installed:"
echo "  - $INSTALL_DIR/lab_server"
echo "  - $INSTALL_DIR/lab_health"
echo "  - $INSTALL_DIR/lab_query"
echo ""
echo "Configuration: $CONFIG_DIR/.env"
echo "Database: $DB_PATH"
echo ""
echo -e "${BLUE}Quick Start:${NC}"
echo "  lab_server add --name 'myserver' --ip '192.168.1.100' --ssh-user 'admin'"
echo "  lab_health check-server --id 1"
echo "  lab_query overview"
echo ""

# Check if directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "Note: $INSTALL_DIR is not in your PATH"
    echo "Add this line to your ~/.bashrc or ~/.zshrc:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
fi
