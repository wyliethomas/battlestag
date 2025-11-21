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
DEFAULT_CONFIG_DIR="$HOME/.config/financial-asset-tracker"
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

# Install binaries
echo "Installing binaries..."
cp financial-asset-tracker "$INSTALL_DIR/"
cp financial-asset-tracker-query "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/financial-asset-tracker"
chmod +x "$INSTALL_DIR/financial-asset-tracker-query"

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
exec "$INSTALL_DIR/financial-asset-tracker" "\$@"
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
exec "$INSTALL_DIR/financial-asset-tracker-query" "\$@"
EOF

# Make wrappers executable
chmod +x "$INSTALL_DIR/financial-asset-tracker-run"
chmod +x "$INSTALL_DIR/financial-asset-tracker-query-run"

echo
echo "✓ Installation complete!"
echo
echo "Binaries installed:"
echo "  $INSTALL_DIR/financial-asset-tracker"
echo "  $INSTALL_DIR/financial-asset-tracker-query"
echo
echo "Wrapper scripts installed:"
echo "  $INSTALL_DIR/financial-asset-tracker-run"
echo "  $INSTALL_DIR/financial-asset-tracker-query-run"
echo
echo "Configuration file:"
echo "  $CONFIG_DIR/.env"
echo

# Check and update PATH
echo "Checking PATH configuration..."

# Detect shell and config file
SHELL_CONFIG=""
if [ -n "$BASH_VERSION" ]; then
    if [ -f "$HOME/.bashrc" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
    elif [ -f "$HOME/.bash_profile" ]; then
        SHELL_CONFIG="$HOME/.bash_profile"
    fi
elif [ -n "$ZSH_VERSION" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
else
    # Fallback: check for common files
    if [ -f "$HOME/.bashrc" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
    elif [ -f "$HOME/.bash_profile" ]; then
        SHELL_CONFIG="$HOME/.bash_profile"
    elif [ -f "$HOME/.profile" ]; then
        SHELL_CONFIG="$HOME/.profile"
    fi
fi

# Check if PATH contains the install directory
PATH_CONFIGURED=false
if [ -n "$PATH" ]; then
    if echo "$PATH" | grep -q "$INSTALL_DIR"; then
        PATH_CONFIGURED=true
    fi
fi

if [ "$PATH_CONFIGURED" = false ]; then
    if [ -n "$SHELL_CONFIG" ]; then
        echo "Adding $INSTALL_DIR to PATH in $SHELL_CONFIG"

        # Create backup
        cp "$SHELL_CONFIG" "${SHELL_CONFIG}.backup"

        # Add PATH export
        echo "" >> "$SHELL_CONFIG"
        echo "# Added by financial-asset-tracker installer" >> "$SHELL_CONFIG"
        echo "export PATH=\"\$HOME/.local/bin:\$PATH\"" >> "$SHELL_CONFIG"

        echo "✓ PATH updated in $SHELL_CONFIG"
        echo ""
        echo "⚠️  IMPORTANT: Run this command to apply changes:"
        echo "    source $SHELL_CONFIG"
        echo ""
        echo "Or log out and log back in."
    else
        echo "⚠️  Could not detect shell configuration file."
        echo "   Please manually add this to your shell config (~/.bashrc or ~/.zshrc):"
        echo "   export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
else
    echo "✓ $INSTALL_DIR is already in PATH"
fi

echo
echo "Quick start:"
echo "  financial-asset-tracker-run add --name \"My Asset\" --category other --current-value 1000"
echo "  financial-asset-tracker-query-run --pretty"
echo
