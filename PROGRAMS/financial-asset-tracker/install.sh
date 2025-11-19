#!/bin/bash
set -e

echo "=== Financial Asset Tracker Installation ==="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.19 or higher."
    exit 1
fi

# Get installation directory
DEFAULT_INSTALL_DIR="$HOME/.local/bin"
read -p "Installation directory [$DEFAULT_INSTALL_DIR]: " INSTALL_DIR
INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"

# Create installation directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Get config directory
DEFAULT_CONFIG_DIR="$HOME/.config/financial-financial-asset-tracker"
read -p "Configuration directory [$DEFAULT_CONFIG_DIR]: " CONFIG_DIR
CONFIG_DIR="${CONFIG_DIR:-$DEFAULT_CONFIG_DIR}"

# Create config directory
mkdir -p "$CONFIG_DIR"

# Get database path
DEFAULT_DB_PATH="$HOME/.local/share/financial-asset-tracker/assets.db"
read -p "Database file path [$DEFAULT_DB_PATH]: " DB_PATH
DB_PATH="${DB_PATH:-$DEFAULT_DB_PATH}"

# Create database directory
DB_DIR=$(dirname "$DB_PATH")
mkdir -p "$DB_DIR"

echo
echo "Installing financial-asset-tracker..."
echo "  Install directory: $INSTALL_DIR"
echo "  Config directory:  $CONFIG_DIR"
echo "  Database path:     $DB_PATH"
echo

# Build executables
echo "Building executables..."
go build -o financial-asset-tracker ./cmd/tracker
go build -o financial-asset-tracker-query ./cmd/query

# Create .env file
cat > "$CONFIG_DIR/.env" <<EOF
# Asset Tracker Configuration
DB_PATH=$DB_PATH
EOF

# Create wrapper scripts
echo "Creating wrapper scripts..."

# Tracker wrapper
cat > "$INSTALL_DIR/financial-asset-tracker-run" <<EOF
#!/bin/bash
# Auto-generated wrapper script for financial-asset-tracker
CONFIG_DIR="$CONFIG_DIR"

# Load environment
if [ -f "\$CONFIG_DIR/.env" ]; then
    set -a
    source "\$CONFIG_DIR/.env"
    set +a
fi

# Run the executable
exec "$PWD/financial-asset-tracker" "\$@"
EOF

# Query wrapper
cat > "$INSTALL_DIR/financial-asset-tracker-query-run" <<EOF
#!/bin/bash
# Auto-generated wrapper script for financial-asset-tracker-query
CONFIG_DIR="$CONFIG_DIR"

# Load environment
if [ -f "\$CONFIG_DIR/.env" ]; then
    set -a
    source "\$CONFIG_DIR/.env"
    set +a
fi

# Run the executable
exec "$PWD/financial-asset-tracker-query" "\$@"
EOF

# Make wrappers executable
chmod +x "$INSTALL_DIR/financial-asset-tracker-run"
chmod +x "$INSTALL_DIR/financial-asset-tracker-query-run"

echo
echo "✓ Installation complete!"
echo
echo "Wrapper scripts installed:"
echo "  $INSTALL_DIR/financial-asset-tracker-run"
echo "  $INSTALL_DIR/financial-asset-tracker-query-run"
echo
echo "Configuration file:"
echo "  $CONFIG_DIR/.env"
echo
echo "Make sure $INSTALL_DIR is in your PATH."
echo
echo "Quick start:"
echo "  financial-asset-tracker-run add --name \"My Asset\" --category other --current-value 1000"
echo "  financial-asset-tracker-query-run --pretty"
echo
