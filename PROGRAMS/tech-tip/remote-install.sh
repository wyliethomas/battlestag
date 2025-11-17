#!/bin/bash

# Remote installation script for tech-tip
# This script is executed on the remote server to install and configure the application
# Usage: ./remote-install.sh [install_dir] [data_dir] [ollama_host] [ollama_model]

set -e

# Parse arguments or use defaults
INSTALL_DIR="${1:-/usr/local/bin}"
DATA_DIR="${2:-/var/lib/tech-tip}"
OLLAMA_HOST="${3:-http://localhost:11434}"
OLLAMA_MODEL="${4:-dolphin3}"

BINARY_NAME="tech-tip"
SYSTEMD_USER_DIR="$HOME/.config/systemd/user"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Tech Tip Generator - Remote Installation ===${NC}"
echo ""
echo "Installation configuration:"
echo "  Binary:       $INSTALL_DIR/$BINARY_NAME"
echo "  Data:         $DATA_DIR"
echo "  Ollama host:  $OLLAMA_HOST"
echo "  Ollama model: $OLLAMA_MODEL"
echo ""

# Check if running as root (we don't want that for user systemd services)
if [ "$EUID" -eq 0 ]; then
    echo -e "${RED}Error: Do not run this script as root${NC}"
    echo "This script installs a user systemd service"
    exit 1
fi

# Create data directory
echo -e "${YELLOW}Creating data directory...${NC}"
mkdir -p "$DATA_DIR"
echo "Data directory: $DATA_DIR"

# Install binary
echo -e "${YELLOW}Installing binary to $INSTALL_DIR...${NC}"
if [ -f "$BINARY_NAME" ]; then
    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        if [ -w "$(dirname "$INSTALL_DIR")" ]; then
            mkdir -p "$INSTALL_DIR"
        else
            sudo mkdir -p "$INSTALL_DIR"
        fi
    fi

    # Install binary (use sudo if needed)
    if [ -w "$INSTALL_DIR" ]; then
        install -m 755 "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo install -m 755 "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    fi
    echo "Binary installed: $INSTALL_DIR/$BINARY_NAME"
else
    echo -e "${RED}Error: Binary $BINARY_NAME not found${NC}"
    exit 1
fi

# Create systemd user directory
echo -e "${YELLOW}Setting up systemd service...${NC}"
mkdir -p "$SYSTEMD_USER_DIR"

# Install systemd service and timer
if [ -f "tech-tip.service" ] && [ -f "tech-tip.timer" ]; then
    # Update service file with custom paths and settings
    sed -e "s|ExecStart=.*|ExecStart=$INSTALL_DIR/$BINARY_NAME|g" \
        -e "s|Environment=\"OLLAMA_HOST=.*\"|Environment=\"OLLAMA_HOST=$OLLAMA_HOST\"|g" \
        -e "s|Environment=\"OLLAMA_MODEL=.*\"|Environment=\"OLLAMA_MODEL=$OLLAMA_MODEL\"|g" \
        -e "s|Environment=\"DB_PATH=.*\"|Environment=\"DB_PATH=$DATA_DIR/tech_tips.db\"|g" \
        tech-tip.service > "$SYSTEMD_USER_DIR/tech-tip.service"

    cp tech-tip.timer "$SYSTEMD_USER_DIR/"
    echo "Systemd files installed to $SYSTEMD_USER_DIR"
else
    echo -e "${RED}Error: systemd service or timer files not found${NC}"
    exit 1
fi

# Reload systemd
echo -e "${YELLOW}Reloading systemd...${NC}"
systemctl --user daemon-reload

# Test Ollama connection
echo -e "${YELLOW}Testing Ollama connection...${NC}"
if curl -s "$OLLAMA_HOST/api/tags" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Ollama is reachable at $OLLAMA_HOST${NC}"
else
    echo -e "${RED}✗ Warning: Cannot reach Ollama at $OLLAMA_HOST${NC}"
    echo "Make sure Ollama is running before starting the service"
fi

# Test binary execution
echo -e "${YELLOW}Testing binary...${NC}"
if "$INSTALL_DIR/$BINARY_NAME" --help > /dev/null 2>&1 || true; then
    echo -e "${GREEN}✓ Binary is executable${NC}"
else
    echo -e "${RED}✗ Warning: Binary test failed${NC}"
fi

echo ""
echo -e "${GREEN}=== Installation Complete ===${NC}"
echo ""
echo "Next steps:"
echo "1. Enable the timer:"
echo "   systemctl --user enable tech-tip.timer"
echo ""
echo "2. Start the timer:"
echo "   systemctl --user start tech-tip.timer"
echo ""
echo "3. Check timer status:"
echo "   systemctl --user list-timers tech-tip.timer"
echo ""
echo "4. Run manually to test:"
echo "   $INSTALL_DIR/$BINARY_NAME"
echo ""
echo "5. View logs:"
echo "   journalctl --user -u tech-tip.service"
echo ""

# Ask if user wants to enable and start now
read -p "Enable and start the timer now? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    systemctl --user enable tech-tip.timer
    systemctl --user start tech-tip.timer
    echo -e "${GREEN}✓ Timer enabled and started${NC}"
    echo ""
    systemctl --user list-timers tech-tip.timer
fi
