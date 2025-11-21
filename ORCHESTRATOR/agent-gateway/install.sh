#!/bin/bash
set -e

echo "=== Agent Gateway - Installation ==="
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
BINARY_NAME="agent-gateway"

# Build the binary
echo "Building agent-gateway..."
go build -o "$BINARY_NAME"

if [ ! -f "$BINARY_NAME" ]; then
    echo "Error: Build failed"
    exit 1
fi

echo "✓ Build successful"
echo

# Create directories
echo "Creating installation directories..."
sudo mkdir -p "$INSTALL_DIR"
sudo mkdir -p "$CONFIG_DIR"
sudo mkdir -p "$LOG_DIR"
sudo chown $USER:$USER "$LOG_DIR"

# Install binary
echo "Installing binary to $INSTALL_DIR..."
sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "✓ Binary installed"
echo

# Check if config already exists
if [ -f "$CONFIG_DIR/config.yaml" ]; then
    echo "Config file already exists at $CONFIG_DIR/config.yaml"
    read -p "Overwrite with new config? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo cp config.yaml "$CONFIG_DIR/config.yaml"
        echo "✓ Config updated"
    else
        echo "  Keeping existing config"
    fi
else
    echo "Installing config file..."
    sudo cp config.yaml "$CONFIG_DIR/config.yaml"
    echo "✓ Config installed"
fi

# Update config with user's home directory
echo "Updating config paths for user: $USER"
sudo sed -i "s|/home/battlestag|$HOME|g" "$CONFIG_DIR/config.yaml"

echo

# Check if financial executables are installed
echo "Checking for financial executables..."
MISSING_EXECUTABLES=()

if [ ! -f "$HOME/.local/bin/financial-statement-processor" ]; then
    MISSING_EXECUTABLES+=("financial-statement-processor")
fi
if [ ! -f "$HOME/.local/bin/financial-asset-tracker" ]; then
    MISSING_EXECUTABLES+=("financial-asset-tracker")
fi
if [ ! -f "$HOME/.local/bin/financial-liability-tracker" ]; then
    MISSING_EXECUTABLES+=("financial-liability-tracker")
fi

if [ ${#MISSING_EXECUTABLES[@]} -gt 0 ]; then
    echo "⚠️  Warning: Some financial executables are not installed:"
    for exe in "${MISSING_EXECUTABLES[@]}"; do
        echo "  - $exe"
    done
    echo
    echo "The agent-gateway service will fail to start without these."
    echo "Install them first from their respective directories."
    echo
    read -p "Continue anyway? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled"
        exit 1
    fi
else
    echo "✓ All financial executables found"
fi

echo

# Create systemd service file
echo "Creating systemd service..."
sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null <<EOF
[Unit]
Description=Agent Gateway API Server
Documentation=https://github.com/battlestag/agent-gateway
After=network.target

[Service]
Type=simple
User=$USER
Group=$USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME --config $CONFIG_DIR/config.yaml
Restart=on-failure
RestartSec=5s

# Security settings
NoNewPrivileges=true
PrivateTmp=true

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

[Install]
WantedBy=multi-user.target
EOF

echo "✓ Service file created"
echo

# Reload systemd
echo "Reloading systemd..."
sudo systemctl daemon-reload

# Enable service
echo "Enabling service to start on boot..."
sudo systemctl enable $SERVICE_NAME

echo "✓ Service enabled"
echo

# Ask to start service now
read -p "Start agent-gateway service now? [Y/n] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
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
        echo "Check logs with:"
        echo "  sudo journalctl -u $SERVICE_NAME -n 50"
        exit 1
    fi
else
    echo "Service not started"
    echo "Start it later with: sudo systemctl start $SERVICE_NAME"
fi

echo
echo "✓ Installation complete!"
echo
echo "Service management:"
echo "  Start:   sudo systemctl start $SERVICE_NAME"
echo "  Stop:    sudo systemctl stop $SERVICE_NAME"
echo "  Restart: sudo systemctl restart $SERVICE_NAME"
echo "  Status:  sudo systemctl status $SERVICE_NAME"
echo "  Logs:    sudo journalctl -u $SERVICE_NAME -f"
echo
echo "Configuration:"
echo "  Config file: $CONFIG_DIR/config.yaml"
echo "  Edit and restart service to apply changes"
echo
echo "API Documentation:"
echo "  See API.md for endpoint documentation"
echo
echo "Test the API:"
echo "  curl http://localhost:8080/api/health"
echo
