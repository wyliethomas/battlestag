#!/bin/bash
# Uninstall BATTLESTAG Orchestrator

set -e

TARGET_DIR="${BATTLESTAG_TARGET_DIR:-$HOME/battlestag}"

echo "Uninstalling Orchestrator"
echo "========================="
echo ""

# Check if installed
if [ ! -d "$TARGET_DIR/orchestrator" ]; then
    echo "✗ Orchestrator not installed"
    exit 0
fi

# Warning
echo "⚠ This will remove the orchestrator and all its data"
echo "  Installation directory: $TARGET_DIR/orchestrator"
echo ""
read -p "Continue? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Uninstall cancelled."
    exit 0
fi

# Stop and remove service
if command -v systemctl &> /dev/null; then
    if systemctl list-unit-files | grep -q battlestag-orchestrator; then
        echo "Stopping and removing service..."
        sudo systemctl stop battlestag-orchestrator 2>/dev/null || true
        sudo systemctl disable battlestag-orchestrator 2>/dev/null || true
        sudo rm -f /etc/systemd/system/battlestag-orchestrator.service
        sudo systemctl daemon-reload
        echo "✓ Service removed"
    fi
fi

# Remove installation directory
echo "Removing installation directory..."
rm -rf "$TARGET_DIR/orchestrator"
echo "✓ Directory removed"

echo ""
echo "✓ Orchestrator uninstalled successfully"
