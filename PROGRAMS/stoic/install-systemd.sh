#!/bin/bash

# Installation script for stoic-thought systemd service and timer

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SYSTEMD_USER_DIR="$HOME/.config/systemd/user"

echo "Installing stoic-thought systemd service and timer..."

# Create systemd user directory if it doesn't exist
mkdir -p "$SYSTEMD_USER_DIR"

# Copy service and timer files
cp "$SCRIPT_DIR/stoic-thought.service" "$SYSTEMD_USER_DIR/"
cp "$SCRIPT_DIR/stoic-thought.timer" "$SYSTEMD_USER_DIR/"

echo "Files copied to $SYSTEMD_USER_DIR"

# Reload systemd
systemctl --user daemon-reload

echo ""
echo "Installation complete!"
echo ""
echo "To enable and start the timer:"
echo "  systemctl --user enable stoic-thought.timer"
echo "  systemctl --user start stoic-thought.timer"
echo ""
echo "To check timer status:"
echo "  systemctl --user list-timers"
echo ""
echo "To view logs:"
echo "  journalctl --user -u stoic-thought.service"
echo ""
echo "To manually run the service:"
echo "  systemctl --user start stoic-thought.service"
