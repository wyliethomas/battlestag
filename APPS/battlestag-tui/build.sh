#!/bin/bash

cd "$(dirname "$0")"

echo "Building BATTLESTAG TUI..."
echo

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

if [ $? -ne 0 ]; then
    echo "Error: Failed to download dependencies"
    exit 1
fi

# Build the application
echo "Building application..."
go build -o battlestag-tui .

if [ $? -ne 0 ]; then
    echo "Error: Build failed"
    exit 1
fi

echo
echo "Build successful!"
echo "Run with: ./run.sh (uses defaults) or ./battlestag-tui"
echo
echo "Environment variables (optional, defaults are set):"
echo "  export AGENT_GATEWAY_URL=\"http://192.168.1.140:8080\""
echo "  export AGENT_GATEWAY_API_KEY=\"test-api-key-12345\""
