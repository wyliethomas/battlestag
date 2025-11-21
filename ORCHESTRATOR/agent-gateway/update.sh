#!/bin/bash
set -e

echo "=== Agent Gateway - Update ==="
echo

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    echo "Error: Do not run this script as root/sudo"
    echo "The script will prompt for sudo when needed"
    exit 1
fi

# Configuration
INSTALL_DIR="/opt/agent-gateway"
CONFIG_DIR="/etc/agent-gateway"
SERVICE_NAME="agent-gateway"
BINARY_NAME="agent-gateway"

# Check if installed
if [ ! -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo "Error: agent-gateway is not installed"
    echo "Run ./install.sh first"
    exit 1
fi

# Check if this is a git repository
if [ -d .git ]; then
    echo "Pulling latest changes..."
    git pull
    echo
fi

# Update Go dependencies
echo "Updating Go dependencies..."
go mod tidy
echo

# Build new binary
echo "Building new binary..."
go build -o "$BINARY_NAME"

if [ ! -f "$BINARY_NAME" ]; then
    echo "Error: Build failed"
    exit 1
fi

echo "✓ Build successful"
echo

# Stop service
echo "Stopping service..."
sudo systemctl stop $SERVICE_NAME
echo "✓ Service stopped"
echo

# Backup old binary
BACKUP_NAME="$BINARY_NAME.backup.$(date +%Y%m%d_%H%M%S)"
echo "Backing up old binary to $INSTALL_DIR/$BACKUP_NAME..."
sudo cp "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/$BACKUP_NAME"
echo "✓ Backup created"
echo

# Install new binary
echo "Installing new binary..."
sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
echo "✓ Binary updated"
echo

# Check if config file needs updating
if [ -f "config.yaml" ] && [ -f "$CONFIG_DIR/config.yaml" ]; then
    # Compare configs (ignore comments and whitespace)
    if ! diff -q <(grep -v '^#' config.yaml | grep -v '^[[:space:]]*$') \
                  <(sudo cat $CONFIG_DIR/config.yaml | grep -v '^#' | grep -v '^[[:space:]]*$') > /dev/null 2>&1; then
        echo "Config file has changed"
        read -p "Update config file? (Will backup old one) [Y/n] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            CONFIG_BACKUP="config.yaml.backup.$(date +%Y%m%d_%H%M%S)"
            sudo cp "$CONFIG_DIR/config.yaml" "$CONFIG_DIR/$CONFIG_BACKUP"
            sudo cp config.yaml "$CONFIG_DIR/config.yaml"
            # Update paths for current user
            sudo sed -i "s|/home/battlestag|$HOME|g" "$CONFIG_DIR/config.yaml"
            echo "✓ Config updated (backup: $CONFIG_DIR/$CONFIG_BACKUP)"
        else
            echo "  Keeping existing config"
        fi
    else
        echo "✓ Config file unchanged"
    fi
fi

echo

# Restart service
echo "Starting service..."
sudo systemctl start $SERVICE_NAME
sleep 2

# Check status
if sudo systemctl is-active --quiet $SERVICE_NAME; then
    echo "✓ Service started successfully"
    echo
    sudo systemctl status $SERVICE_NAME --no-pager -l
else
    echo "✗ Service failed to start"
    echo
    echo "Restoring backup..."
    sudo systemctl stop $SERVICE_NAME
    sudo cp "$INSTALL_DIR/$BACKUP_NAME" "$INSTALL_DIR/$BINARY_NAME"
    sudo systemctl start $SERVICE_NAME
    echo
    echo "Check logs with:"
    echo "  sudo journalctl -u $SERVICE_NAME -n 50"
    exit 1
fi

echo
echo "✓ Update complete!"
echo
echo "View logs:"
echo "  sudo journalctl -u $SERVICE_NAME -f"
echo
