#!/bin/bash

# Battlestag TUI Update Script
# Rebuilds and reinstalls the binary without touching config

set -e

BINARY_NAME="battlestag"
INSTALL_DIR="$HOME/.local/bin"

echo "=========================================="
echo "  Battlestag TUI Update"
echo "=========================================="
echo ""

# Check if already installed
if [ ! -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo "⚠️  $BINARY_NAME is not currently installed."
    echo ""
    echo "Run ./install.sh to install for the first time."
    exit 1
fi

echo "Rebuilding $BINARY_NAME..."
make clean
make build

echo ""
echo "Updating binary at $INSTALL_DIR/$BINARY_NAME..."

# Check if binary is running
if pgrep -x "$BINARY_NAME" > /dev/null; then
    echo "⚠️  $BINARY_NAME is currently running"
    echo "Please close it first, then run this script again"
    exit 1
fi

cp $BINARY_NAME $INSTALL_DIR/
chmod +x $INSTALL_DIR/$BINARY_NAME

echo ""
echo "=========================================="
echo "  Update Complete!"
echo "=========================================="
echo ""
echo "Binary updated at: $INSTALL_DIR/$BINARY_NAME"
echo ""
echo "Your configuration at ~/.config/battlestag-tui/ was preserved."
echo ""
echo "Run '$BINARY_NAME' to use the updated version."
echo ""
