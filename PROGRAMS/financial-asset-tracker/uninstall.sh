#!/bin/bash
set -e

echo "=== Financial Asset Tracker Uninstallation ==="
echo

# Find existing wrapper to get directories
WRAPPER_PATHS=(
    "$HOME/.local/bin/financial-asset-tracker-run"
    "/usr/local/bin/financial-asset-tracker-run"
)

CONFIG_DIR=""
DB_PATH=""
for wrapper in "${WRAPPER_PATHS[@]}"; do
    if [ -f "$wrapper" ]; then
        CONFIG_DIR=$(grep "^CONFIG_DIR=" "$wrapper" | cut -d'"' -f2)
        INSTALL_DIR=$(dirname "$wrapper")

        # Get DB_PATH from config if it exists
        if [ -f "$CONFIG_DIR/.env" ]; then
            DB_PATH=$(grep "^DB_PATH=" "$CONFIG_DIR/.env" | cut -d'=' -f2)
        fi
        break
    fi
done

if [ -z "$CONFIG_DIR" ]; then
    echo "Error: Could not find existing installation."
    exit 1
fi

echo "Found installation:"
echo "  Install directory: $INSTALL_DIR"
echo "  Config directory:  $CONFIG_DIR"
echo "  Database path:     $DB_PATH"
echo

# Ask for confirmation
read -p "Remove wrapper scripts? [y/N]: " REMOVE_WRAPPERS
if [[ "$REMOVE_WRAPPERS" =~ ^[Yy]$ ]]; then
    rm -f "$INSTALL_DIR/financial-asset-tracker-run"
    rm -f "$INSTALL_DIR/financial-asset-tracker-query-run"
    echo "✓ Wrapper scripts removed"
fi

read -p "Remove configuration directory? [y/N]: " REMOVE_CONFIG
if [[ "$REMOVE_CONFIG" =~ ^[Yy]$ ]]; then
    rm -rf "$CONFIG_DIR"
    echo "✓ Configuration directory removed"
fi

if [ -n "$DB_PATH" ] && [ -f "$DB_PATH" ]; then
    read -p "Remove database file (THIS DELETES ALL DATA)? [y/N]: " REMOVE_DB
    if [[ "$REMOVE_DB" =~ ^[Yy]$ ]]; then
        rm -f "$DB_PATH"
        rmdir "$(dirname "$DB_PATH")" 2>/dev/null || true
        echo "✓ Database removed"
    fi
fi

# Offer to remove PATH entry
read -p "Remove PATH entry from shell config? [y/N]: " REMOVE_PATH
if [[ "$REMOVE_PATH" =~ ^[Yy]$ ]]; then
    # Detect shell config file
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
        # Fallback
        if [ -f "$HOME/.bashrc" ]; then
            SHELL_CONFIG="$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            SHELL_CONFIG="$HOME/.bash_profile"
        elif [ -f "$HOME/.profile" ]; then
            SHELL_CONFIG="$HOME/.profile"
        fi
    fi

    if [ -n "$SHELL_CONFIG" ] && [ -f "$SHELL_CONFIG" ]; then
        # Check if our comment exists
        if grep -q "# Added by financial-asset-tracker installer" "$SHELL_CONFIG"; then
            # Create backup
            cp "$SHELL_CONFIG" "${SHELL_CONFIG}.backup"

            # Remove the lines (comment and export)
            sed -i '/# Added by financial-asset-tracker installer/,+1d' "$SHELL_CONFIG"

            echo "✓ PATH entry removed from $SHELL_CONFIG"
            echo "  (Backup saved to ${SHELL_CONFIG}.backup)"
            echo ""
            echo "⚠️  Run 'source $SHELL_CONFIG' or restart your shell to apply changes."
        else
            echo "  No PATH entry found in $SHELL_CONFIG"
        fi
    else
        echo "  Could not find shell configuration file"
    fi
fi

echo
echo "Uninstallation complete."
echo "The source code and built executables in $(pwd) have not been removed."
echo
