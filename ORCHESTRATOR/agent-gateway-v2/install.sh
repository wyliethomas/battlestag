#!/bin/bash

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Find BATTLESTAG-BOT root by going up from script location
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BIN_DIR="$REPO_ROOT/bin"
CONFIG_DIR="${HOME}/.config/agent-gateway"

echo "=== Agent Gateway Installation ===="
echo ""

# Create bin directory if it doesn't exist
if [ ! -d "$BIN_DIR" ]; then
    echo "Creating bin directory: $BIN_DIR"
    mkdir -p "$BIN_DIR"
fi

# Create config directory if it doesn't exist
if [ ! -d "$CONFIG_DIR" ]; then
    echo "Creating config directory: $CONFIG_DIR"
    mkdir -p "$CONFIG_DIR"
fi

# Check if config file exists
if [ -f "$CONFIG_DIR/config.yaml" ]; then
    echo "✓ Config file already exists: $CONFIG_DIR/config.yaml"
else
    echo "Creating default config file..."
    cp "$SCRIPT_DIR/config.example.yaml" "$CONFIG_DIR/config.yaml"
    echo "✓ Created config file: $CONFIG_DIR/config.yaml"
    echo ""
    echo "IMPORTANT: Edit $CONFIG_DIR/config.yaml to configure:"
    echo "  - Ollama server URL"
    echo "  - Ollama model"
    echo "  - Server port"
    echo ""
fi

# Install Go dependencies
echo "Installing Go dependencies..."
cd "$SCRIPT_DIR"
go mod tidy
echo "✓ Dependencies installed"

# Build the gateway
echo "Building agent-gateway..."
go build -o "$BIN_DIR/agent-gateway" main.go config.go
echo "✓ Gateway built: $BIN_DIR/agent-gateway"

# Build programs
echo "Building programs..."
PROGRAMS_BUILT=0

for program_dir in "$SCRIPT_DIR/programs"/*; do
    if [ -d "$program_dir" ]; then
        program_name=$(basename "$program_dir")

        # Skip test programs that are already built
        if [ -f "$program_dir/$program_name" ]; then
            echo "  - $program_name (already built)"
            PROGRAMS_BUILT=$((PROGRAMS_BUILT + 1))
            continue
        fi

        # Check if there's a Go source file
        if ls "$program_dir"/*.go 1> /dev/null 2>&1; then
            echo "  - Building $program_name..."
            cd "$program_dir"
            go build -o "$program_name" *.go
            PROGRAMS_BUILT=$((PROGRAMS_BUILT + 1))
            cd "$SCRIPT_DIR"
        fi
    fi
done

echo "✓ Built $PROGRAMS_BUILT programs"

# Check if bin directory is in PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo ""
    echo "⚠️  WARNING: $BIN_DIR is not in your PATH"
    echo ""
    echo "Add this line to your ~/.bashrc or ~/.zshrc:"
    echo "  export PATH=\"$BIN_DIR:\$PATH\""
    echo ""
fi

echo ""
echo "=== Installation Complete! ==="
echo ""
echo "Config file: $CONFIG_DIR/config.yaml"
echo "Binary location: $BIN_DIR/agent-gateway"
echo ""
echo "To start the gateway:"
echo "  agent-gateway --config $CONFIG_DIR/config.yaml"
echo ""
echo "Or with command-line overrides:"
echo "  agent-gateway --ollama-url http://your-ollama:11434 --ollama-model llama3.1:8b"
echo ""
