#!/bin/bash

# Battlestag TUI Uninstallation Script

BINARY_NAME="battlestag"
INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/battlestag-tui"

echo "=========================================="
echo "  Battlestag TUI Uninstallation"
echo "=========================================="
echo ""

# Check if binary exists
if [ ! -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo "⚠️  $BINARY_NAME is not installed at $INSTALL_DIR"
    echo ""
    echo "Nothing to uninstall."
    exit 0
fi

# Remove the binary
echo "Removing binary from: $INSTALL_DIR/$BINARY_NAME"
rm -f "$INSTALL_DIR/$BINARY_NAME"
echo "✓ Binary removed"
echo ""

# Ask about config directory
if [ -d "$CONFIG_DIR" ]; then
    echo "Configuration directory found: $CONFIG_DIR"
    echo ""
    read -p "Do you want to remove configuration files? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Removing configuration directory..."
        rm -rf "$CONFIG_DIR"
        echo "✓ Configuration removed"
    else
        echo "Configuration files preserved at: $CONFIG_DIR"
        echo ""
        echo "You can remove them later with:"
        echo "  rm -rf $CONFIG_DIR"
    fi
else
    echo "No configuration directory found."
fi

echo ""
echo "=========================================="
echo "  Uninstallation Complete!"
echo "=========================================="
echo ""
echo "$BINARY_NAME has been removed from your system."
echo ""
