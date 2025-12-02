#!/bin/bash
# Update BATTLESTAG Orchestrator

set -e

TARGET_DIR="${BATTLESTAG_TARGET_DIR:-$HOME/battlestag}"
SCRIPT_DIR="${BATTLESTAG_SCRIPT_DIR}"

echo "Updating Orchestrator"
echo "====================="
echo ""

# Check if installed
if [ ! -f "$TARGET_DIR/orchestrator/agent-gateway" ]; then
    echo "✗ Orchestrator not installed"
    echo "  Run: ./deploy.sh orchestrator install"
    exit 1
fi

# Stop service if running
if command -v systemctl &> /dev/null; then
    if systemctl is-active --quiet battlestag-orchestrator; then
        echo "Stopping orchestrator service..."
        sudo systemctl stop battlestag-orchestrator
        RESTART_SERVICE=true
    fi
fi

# Backup current binary
echo "Creating backup..."
cp "$TARGET_DIR/orchestrator/agent-gateway" "$TARGET_DIR/orchestrator/agent-gateway.backup"
echo "✓ Backup created"

# Update binary
PACKAGES_DIR="$SCRIPT_DIR/packages"
if [ -f "$PACKAGES_DIR/orchestrator.tar.gz" ]; then
    echo "Extracting new version..."
    # Extract only the binary, preserve config
    tar -xzf "$PACKAGES_DIR/orchestrator.tar.gz" -C "/tmp" agent-gateway
    mv "/tmp/agent-gateway" "$TARGET_DIR/orchestrator/agent-gateway"
    chmod +x "$TARGET_DIR/orchestrator/agent-gateway"
    echo "✓ Binary updated"
else
    echo "✗ Update package not found"
    echo "  Restoring backup..."
    mv "$TARGET_DIR/orchestrator/agent-gateway.backup" "$TARGET_DIR/orchestrator/agent-gateway"
    exit 1
fi

# Restart service if it was running
if [ "$RESTART_SERVICE" = true ]; then
    echo "Restarting orchestrator service..."
    sudo systemctl start battlestag-orchestrator
    echo "✓ Service restarted"
fi

echo ""
echo "✓ Orchestrator updated successfully!"
echo ""
echo "Backup available at: $TARGET_DIR/orchestrator/agent-gateway.backup"
