#!/bin/bash
set -e

echo "=== Agent Gateway - Uninstallation ==="
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
LOG_DIR="/var/log/agent-gateway"
SERVICE_NAME="agent-gateway"

# Check if service exists
if [ ! -f "/etc/systemd/system/$SERVICE_NAME.service" ]; then
    echo "Service not found. Already uninstalled?"
    exit 0
fi

# Stop service
echo "Stopping service..."
if sudo systemctl is-active --quiet $SERVICE_NAME; then
    sudo systemctl stop $SERVICE_NAME
    echo "✓ Service stopped"
else
    echo "  Service not running"
fi

# Disable service
echo "Disabling service..."
sudo systemctl disable $SERVICE_NAME
echo "✓ Service disabled"
echo

# Remove service file
echo "Removing service file..."
sudo rm -f /etc/systemd/system/$SERVICE_NAME.service
sudo systemctl daemon-reload
echo "✓ Service file removed"
echo

# Ask about removing files
echo "Data locations:"
echo "  Binary:  $INSTALL_DIR"
echo "  Config:  $CONFIG_DIR"
echo "  Logs:    $LOG_DIR"
echo

read -p "Remove all data (binary, config, logs)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Removing installation directory..."
    sudo rm -rf "$INSTALL_DIR"
    echo "✓ Binary removed"

    echo "Removing config directory..."
    sudo rm -rf "$CONFIG_DIR"
    echo "✓ Config removed"

    echo "Removing log directory..."
    sudo rm -rf "$LOG_DIR"
    echo "✓ Logs removed"

    echo
    echo "All data has been removed."
else
    echo
    echo "Data preserved at:"
    [ -d "$INSTALL_DIR" ] && echo "  $INSTALL_DIR"
    [ -d "$CONFIG_DIR" ] && echo "  $CONFIG_DIR"
    [ -d "$LOG_DIR" ] && echo "  $LOG_DIR"
    echo
    echo "Remove manually if needed:"
    echo "  sudo rm -rf $INSTALL_DIR"
    echo "  sudo rm -rf $CONFIG_DIR"
    echo "  sudo rm -rf $LOG_DIR"
fi

echo
echo "✓ Uninstallation complete!"
echo
