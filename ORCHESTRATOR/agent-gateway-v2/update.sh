#!/bin/bash

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Find BATTLESTAG-BOT root by going up from script location
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BIN_DIR="$REPO_ROOT/bin"
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

# Config update function
update_config() {
    echo ""
    echo "=== Configuration Update ==="
    echo ""

    read -p "Do you want to update the configuration? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Keeping existing configuration"
        return
    fi

    # Load existing config values
    local CURRENT_OLLAMA_URL=""
    local CURRENT_OLLAMA_MODEL=""
    local CURRENT_PORT=""

    if [ -f "$CONFIG_DIR/config.yaml" ]; then
        CURRENT_OLLAMA_URL=$(grep -A1 "^ollama:" "$CONFIG_DIR/config.yaml" | grep "url:" | awk '{print $2}' || echo "")
        CURRENT_OLLAMA_MODEL=$(grep -A2 "^ollama:" "$CONFIG_DIR/config.yaml" | grep "model:" | awk '{print $2}' || echo "")
        CURRENT_PORT=$(grep -A1 "^server:" "$CONFIG_DIR/config.yaml" | grep "port:" | awk '{print $2}' || echo "")
    fi

    # Set defaults from current config or fallback
    local DEFAULT_OLLAMA_URL="${CURRENT_OLLAMA_URL:-http://localhost:11434}"
    local DEFAULT_OLLAMA_MODEL="${CURRENT_OLLAMA_MODEL:-llama3.1:8b}"
    local DEFAULT_PORT="${CURRENT_PORT:-8080}"

    echo ""
    echo "Current configuration:"
    echo "  Ollama URL: $DEFAULT_OLLAMA_URL"
    echo "  Ollama Model: $DEFAULT_OLLAMA_MODEL"
    echo "  Server Port: $DEFAULT_PORT"
    echo ""
    echo "Press Enter to keep current value, or type new value:"
    echo ""

    # Prompt for values
    read -p "Ollama server URL [$DEFAULT_OLLAMA_URL]: " OLLAMA_URL
    OLLAMA_URL="${OLLAMA_URL:-$DEFAULT_OLLAMA_URL}"

    read -p "Ollama model [$DEFAULT_OLLAMA_MODEL]: " OLLAMA_MODEL
    OLLAMA_MODEL="${OLLAMA_MODEL:-$DEFAULT_OLLAMA_MODEL}"

    read -p "Server port [$DEFAULT_PORT]: " PORT
    PORT="${PORT:-$DEFAULT_PORT}"

    # Write config file
    cat > "$CONFIG_DIR/config.yaml" <<EOF
# Agent Gateway Configuration File
# Updated by update.sh

# Ollama server settings
ollama:
  url: $OLLAMA_URL
  model: $OLLAMA_MODEL

# HTTP server settings
server:
  port: $PORT

# Program execution settings
execution:
  timeout: 30s

# Programs directory (relative to gateway binary)
programs:
  directory: ./programs
EOF

    echo ""
    echo "✓ Config file updated: $CONFIG_DIR/config.yaml"
    echo ""
}

# Update configuration
update_config

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
