#!/bin/bash
set -e

echo "=== Financial Liability Tracker - Update ==="
echo

# Configuration
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="financial-liability-tracker"

# Check if this is a git repository
if [ -d .git ]; then
    echo "Pulling latest changes..."
    git pull
    echo
fi

# Update dependencies
echo "Updating Go dependencies..."
go mod tidy
echo

# Build the binary
echo "Building $BINARY_NAME..."
go build -o "$BINARY_NAME" ./cmd/tracker

# Install the binary
echo "Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
cp "$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo
echo "✓ Update complete!"
echo
echo "Run '$BINARY_NAME help' to see available commands"
echo
