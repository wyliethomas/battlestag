#!/bin/bash
set -e

echo "=== Financial Liability Tracker - Uninstallation ==="
echo

# Configuration
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="financial-liability-tracker"
CONFIG_DIR="$HOME/.local/share/financial-liability-tracker"
DEFAULT_CONFIG_DIR="$HOME/.config/financial-liability-tracker"

# Remove binary
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo "Removing $BINARY_NAME from $INSTALL_DIR..."
    rm "$INSTALL_DIR/$BINARY_NAME"
    echo "✓ Binary removed"
else
    echo "Binary not found at $INSTALL_DIR/$BINARY_NAME"
fi

# Ask about PATH removal
echo
read -p "Remove PATH entry from shell configuration? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Detect shell config file
    SHELL_CONFIG=""
    if [ -n "$ZSH_VERSION" ]; then
        SHELL_CONFIG="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        if [ -f "$HOME/.bashrc" ]; then
            SHELL_CONFIG="$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            SHELL_CONFIG="$HOME/.bash_profile"
        fi
    fi

    if [ -n "$SHELL_CONFIG" ] && [ -f "$SHELL_CONFIG" ]; then
        # Check if our marker exists
        if grep -q "# Added by financial-liability-tracker installer" "$SHELL_CONFIG"; then
            # Create backup
            cp "$SHELL_CONFIG" "$SHELL_CONFIG.backup.$(date +%Y%m%d_%H%M%S)"

            # Remove the marker and the PATH line after it
            sed -i '/# Added by financial-liability-tracker installer/,+1d' "$SHELL_CONFIG"

            echo "✓ PATH entry removed from $SHELL_CONFIG"
            echo "  (Backup created)"
        else
            echo "  No installer-added PATH entry found in $SHELL_CONFIG"
        fi
    else
        echo "  Could not detect shell configuration file"
    fi
fi

# Ask about data removal
echo
echo "Data locations:"
echo "  Config: $DEFAULT_CONFIG_DIR"
echo "  Database: $CONFIG_DIR"
echo
read -p "Remove all data (configuration and database)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -d "$DEFAULT_CONFIG_DIR" ]; then
        echo "Removing configuration directory..."
        rm -rf "$DEFAULT_CONFIG_DIR"
        echo "✓ Configuration removed"
    fi

    if [ -d "$CONFIG_DIR" ]; then
        echo "Removing data directory..."
        rm -rf "$CONFIG_DIR"
        echo "✓ Data removed"
    fi

    echo
    echo "All data has been removed."
else
    echo
    echo "Data preserved at:"
    [ -d "$DEFAULT_CONFIG_DIR" ] && echo "  $DEFAULT_CONFIG_DIR"
    [ -d "$CONFIG_DIR" ] && echo "  $CONFIG_DIR"
fi

echo
echo "✓ Uninstallation complete!"
echo
