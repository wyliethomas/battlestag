#!/bin/bash

# Battlestag TUI Installation Script

set -e

BINARY_NAME="battlestag"
INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/battlestag-tui"

echo "=========================================="
echo "  Battlestag TUI Installation"
echo "=========================================="
echo ""

# Check if ~/.local/bin exists
if [ ! -d "$INSTALL_DIR" ]; then
    echo "Creating installation directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
fi

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "⚠️  WARNING: $INSTALL_DIR is not in your PATH"
    echo ""
    echo "To add it to your PATH, add this line to your shell profile:"
    echo ""

    # Detect shell and provide appropriate instructions
    if [ -n "$BASH_VERSION" ]; then
        echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
        echo "  source ~/.bashrc"
    elif [ -n "$ZSH_VERSION" ]; then
        echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
        echo "  source ~/.zshrc"
    else
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi
    echo ""
    read -p "Continue with installation anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 1
    fi
fi

# Build and install using Makefile
echo ""
echo "Building and installing $BINARY_NAME..."
make install

echo ""
echo "=========================================="
echo "  Installation Complete!"
echo "=========================================="
echo ""
echo "Binary installed to: $INSTALL_DIR/$BINARY_NAME"
echo "Config directory: $CONFIG_DIR"
echo ""

# Check if binary is accessible
if command -v $BINARY_NAME &> /dev/null; then
    echo "✓ $BINARY_NAME is ready to use!"
    echo ""
    echo "Run '$BINARY_NAME' from any terminal to get started."
else
    echo "⚠️  $BINARY_NAME is not yet accessible in your current shell."
    echo ""
    echo "You may need to:"
    echo "  1. Add $INSTALL_DIR to your PATH (see instructions above)"
    echo "  2. Restart your terminal or run: source ~/.bashrc (or ~/.zshrc)"
    echo ""
    echo "After that, run '$BINARY_NAME' to get started."
fi

echo ""
echo "On first run, you'll be prompted to configure your"
echo "Agent Gateway connection settings."
echo ""
