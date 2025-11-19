#!/bin/bash
set -e

echo "=== Asset Tracker Update ==="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed."
    exit 1
fi

# Find existing wrapper to get config directory
WRAPPER_PATHS=(
    "$HOME/.local/bin/financial-asset-tracker-run"
    "/usr/local/bin/financial-asset-tracker-run"
)

CONFIG_DIR=""
for wrapper in "${WRAPPER_PATHS[@]}"; do
    if [ -f "$wrapper" ]; then
        CONFIG_DIR=$(grep "^CONFIG_DIR=" "$wrapper" | cut -d'"' -f2)
        INSTALL_DIR=$(dirname "$wrapper")
        break
    fi
done

if [ -z "$CONFIG_DIR" ]; then
    echo "Error: Could not find existing installation."
    echo "Please run ./install.sh first."
    exit 1
fi

echo "Found existing installation:"
echo "  Install directory: $INSTALL_DIR"
echo "  Config directory:  $CONFIG_DIR"
echo

# Rebuild executables
echo "Rebuilding executables..."
go build -o financial-asset-tracker ./cmd/tracker
go build -o financial-asset-tracker-query ./cmd/query

# Recreate wrapper scripts (in case PWD changed)
echo "Recreating wrapper scripts..."

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

chmod +x "$INSTALL_DIR/financial-asset-tracker-run"
chmod +x "$INSTALL_DIR/financial-asset-tracker-query-run"

echo
echo "✓ Update complete!"
echo
