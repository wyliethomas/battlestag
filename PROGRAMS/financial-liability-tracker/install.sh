#!/bin/bash
set -e

echo "=== Financial Liability Tracker - Installation ==="
echo

# Configuration
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="financial-liability-tracker"
CONFIG_DIR="$HOME/.local/share/financial-liability-tracker"
DEFAULT_CONFIG_DIR="$HOME/.config/financial-liability-tracker"

# Build the binary
echo "Building $BINARY_NAME..."
go build -o "$BINARY_NAME" ./cmd/tracker

# Create installation directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Install the binary
echo "Installing to $INSTALL_DIR..."
cp "$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Set up configuration directory
mkdir -p "$DEFAULT_CONFIG_DIR"

# Set up .env if it doesn't exist
if [ ! -f "$DEFAULT_CONFIG_DIR/.env" ]; then
    if [ -f ".env.example" ]; then
        echo "Creating default .env file..."
        cp .env.example "$DEFAULT_CONFIG_DIR/.env"
        echo "  Configuration file created at: $DEFAULT_CONFIG_DIR/.env"
        echo "  (Optional: Edit this file to customize database location)"
    fi
fi

# Check if PATH contains the installation directory
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo
    echo "Notice: $INSTALL_DIR is not in your PATH"

    # Detect shell config file
    SHELL_CONFIG=""
    if [ -n "$ZSH_VERSION" ]; then
        SHELL_CONFIG="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        if [ -f "$HOME/.bashrc" ]; then
            SHELL_CONFIG="$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            SHELL_CONFIG="$HOME/.bash_profile"
        fi
    fi

    if [ -n "$SHELL_CONFIG" ]; then
        echo "Detected shell configuration: $SHELL_CONFIG"
        read -p "Add $INSTALL_DIR to PATH in $SHELL_CONFIG? [y/N] " -n 1 -r
        echo

        if [[ $REPLY =~ ^[Yy]$ ]]; then
            # Create backup
            cp "$SHELL_CONFIG" "$SHELL_CONFIG.backup.$(date +%Y%m%d_%H%M%S)"

            # Add PATH export
            echo "" >> "$SHELL_CONFIG"
            echo "# Added by financial-liability-tracker installer" >> "$SHELL_CONFIG"
            echo "export PATH=\"\$HOME/.local/bin:\$PATH\"" >> "$SHELL_CONFIG"

            echo "✓ PATH updated in $SHELL_CONFIG"
            echo "  Run 'source $SHELL_CONFIG' or restart your terminal"
        fi
    else
        echo "  Manually add this to your shell configuration:"
        echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi
fi

echo
echo "✓ Installation complete!"
echo
echo "Usage:"
echo "  $BINARY_NAME add --type credit-card --name \"My Card\" --balance 1000"
echo "  $BINARY_NAME list"
echo "  $BINARY_NAME total"
echo "  $BINARY_NAME help"
echo
echo "Configuration:"
echo "  Config directory: $DEFAULT_CONFIG_DIR"
echo "  Data directory: $CONFIG_DIR (created on first use)"
echo
