#!/bin/bash

cd "$(dirname "$0")"

# Set default values if not already set
if [ -z "$AGENT_GATEWAY_URL" ]; then
    export AGENT_GATEWAY_URL="http://192.168.1.140:8080"
    echo "Using default AGENT_GATEWAY_URL: $AGENT_GATEWAY_URL"
fi

if [ -z "$AGENT_GATEWAY_API_KEY" ]; then
    export AGENT_GATEWAY_API_KEY="test-api-key-12345"
    echo "Using default AGENT_GATEWAY_API_KEY"
fi

echo "Starting BATTLESTAG TUI..."
echo

# Check if binary exists
if [ ! -f "./battlestag-tui" ]; then
    echo "Binary not found. Building..."
    ./build.sh
    if [ $? -ne 0 ]; then
        echo "Build failed"
        exit 1
    fi
fi

# Run the TUI
./battlestag-tui
