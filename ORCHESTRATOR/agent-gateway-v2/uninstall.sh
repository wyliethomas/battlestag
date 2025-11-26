#!/bin/bash

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Find BATTLESTAG-BOT root by going up from script location
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BIN_DIR="$REPO_ROOT/bin"
CONFIG_DIR="${HOME}/.config/agent-gateway"

echo "=== Agent Gateway Uninstallation ===="
echo ""

# Check if gateway is running
if pgrep -f "agent-gateway" > /dev/null; then
    echo "⚠️  Gateway is currently running"
    read -p "Stop the gateway before uninstalling? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Stopping gateway..."
        pkill -SIGTERM -f "agent-gateway" || true
        sleep 2
        # Force kill if still running
        pkill -9 -f "agent-gateway" 2>/dev/null || true
        echo "✓ Gateway stopped"
    else
        echo "⚠️  Gateway is still running. It will continue using the installed binary."
    fi
fi

# Remove binary
if [ -f "$BIN_DIR/agent-gateway" ]; then
    echo "Removing gateway binary..."
    rm "$BIN_DIR/agent-gateway"
    echo "✓ Removed: $BIN_DIR/agent-gateway"
else
    echo "Gateway binary not found (already removed?)"
fi

# Ask about config directory
if [ -d "$CONFIG_DIR" ]; then
    echo ""
    read -p "Remove configuration directory? ($CONFIG_DIR) (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Removing config directory..."
        rm -rf "$CONFIG_DIR"
        echo "✓ Removed: $CONFIG_DIR"
    else
        echo "Keeping config directory: $CONFIG_DIR"
    fi
fi

echo ""
echo "=== Uninstallation Complete! ==="
echo ""
echo "Note: This script only removes the installed binary and config."
echo "The source code in $SCRIPT_DIR is preserved."
echo ""
echo "To completely remove the source code:"
echo "  rm -rf $SCRIPT_DIR"
echo ""
