#!/bin/bash

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${HOME}/Work/BATTLESTAG-BOT/bin"
CONFIG_DIR="${HOME}/.config/agent-gateway"

echo "=== Agent Gateway Update ===="
echo ""

# Check if gateway is currently running
if pgrep -f "agent-gateway" > /dev/null; then
    echo "⚠️  Gateway is currently running"
    read -p "Do you want to stop it before updating? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Stopping gateway..."
        pkill -SIGTERM -f "agent-gateway" || true
        sleep 2
        echo "✓ Gateway stopped"
    else
        echo "Continuing with update (gateway will continue running with old version)"
    fi
fi

# Update Go dependencies
echo "Updating Go dependencies..."
cd "$SCRIPT_DIR"
go mod tidy
echo "✓ Dependencies updated"

# Rebuild the gateway
echo "Rebuilding agent-gateway..."
go build -o "$BIN_DIR/agent-gateway" main.go config.go
echo "✓ Gateway rebuilt: $BIN_DIR/agent-gateway"

# Rebuild programs
echo "Rebuilding programs..."
PROGRAMS_REBUILT=0

for program_dir in "$SCRIPT_DIR/programs"/*; do
    if [ -d "$program_dir" ]; then
        program_name=$(basename "$program_dir")

        # Check if there's a Go source file
        if ls "$program_dir"/*.go 1> /dev/null 2>&1; then
            echo "  - Rebuilding $program_name..."
            cd "$program_dir"
            go build -o "$program_name" *.go
            PROGRAMS_REBUILT=$((PROGRAMS_REBUILT + 1))
            cd "$SCRIPT_DIR"
        fi
    fi
done

echo "✓ Rebuilt $PROGRAMS_REBUILT programs"

# Check for config updates
if [ -f "$SCRIPT_DIR/config.example.yaml" ] && [ -f "$CONFIG_DIR/config.yaml" ]; then
    echo ""
    echo "Checking for config changes..."

    # Compare example config with user config (just inform, don't overwrite)
    echo "  Your config: $CONFIG_DIR/config.yaml"
    echo "  Example config: $SCRIPT_DIR/config.example.yaml"
    echo ""
    echo "  Review config.example.yaml for any new configuration options"
fi

echo ""
echo "=== Update Complete! ==="
echo ""
echo "Binary location: $BIN_DIR/agent-gateway"
echo "Config location: $CONFIG_DIR/config.yaml"
echo ""

# Ask if user wants to restart
read -p "Do you want to start the gateway now? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Starting gateway..."
    cd "$SCRIPT_DIR"
    nohup "$BIN_DIR/agent-gateway" --config "$CONFIG_DIR/config.yaml" > /tmp/agent-gateway.log 2>&1 &
    sleep 2
    if pgrep -f "agent-gateway" > /dev/null; then
        echo "✓ Gateway started (PID: $(pgrep -f 'agent-gateway'))"
        echo "  Logs: /tmp/agent-gateway.log"
    else
        echo "✗ Gateway failed to start. Check /tmp/agent-gateway.log"
    fi
fi
echo ""
