#!/bin/bash

# Local systemd installation script for tech-tip
# Installs the service and timer to the user's systemd directory

set -e

SYSTEMD_USER_DIR="$HOME/.config/systemd/user"

echo "Installing tech-tip systemd service and timer..."
echo ""

# Create systemd user directory if it doesn't exist
mkdir -p "$SYSTEMD_USER_DIR"

# Copy service and timer files
echo "Copying service files to $SYSTEMD_USER_DIR"
cp tech-tip.service "$SYSTEMD_USER_DIR/"
cp tech-tip.timer "$SYSTEMD_USER_DIR/"

# Reload systemd daemon
echo "Reloading systemd daemon..."
systemctl --user daemon-reload

echo ""
echo "Installation complete!"
echo ""
echo "To enable and start the timer:"
echo "  systemctl --user enable tech-tip.timer"
echo "  systemctl --user start tech-tip.timer"
echo ""
echo "To check timer status:"
echo "  systemctl --user list-timers tech-tip.timer"
echo ""
echo "To view logs:"
echo "  journalctl --user -u tech-tip.service -f"
echo ""
