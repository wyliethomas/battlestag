#!/bin/bash
# Install BATTLESTAG Orchestrator (Agent Gateway)

set -e

TARGET_DIR="${BATTLESTAG_TARGET_DIR:-$HOME/battlestag}"
SCRIPT_DIR="${BATTLESTAG_SCRIPT_DIR}"
PACKAGES_DIR="$SCRIPT_DIR/packages"

echo "Installing Orchestrator (Agent Gateway)"
echo "========================================"
echo ""

# Check if already installed
if [ -f "$TARGET_DIR/orchestrator/agent-gateway" ]; then
    echo "⚠ Orchestrator already installed at $TARGET_DIR/orchestrator"
    echo ""
    read -p "Reinstall? This will preserve your config. (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi
fi

# Create directory
echo "Creating installation directory..."
mkdir -p "$TARGET_DIR/orchestrator"

# Check for pre-built package (USB mode)
if [ -f "$PACKAGES_DIR/orchestrator.tar.gz" ]; then
    echo "✓ Using pre-built package (offline mode)"
    tar -xzf "$PACKAGES_DIR/orchestrator.tar.gz" -C "$TARGET_DIR/orchestrator"
else
    echo "⚠ No pre-built package found"
    echo "  Building from source..."

    # Check if source is available
    ROOT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
    if [ ! -d "$ROOT_DIR/ORCHESTRATOR/agent-gateway-v2" ]; then
        echo "✗ Source code not found"
        echo "  Please run build-packages.sh first or ensure source is available"
        exit 1
    fi

    # Build from source
    cd "$ROOT_DIR/ORCHESTRATOR/agent-gateway-v2"
    echo "  Building agent-gateway..."
    go build -o "$TARGET_DIR/orchestrator/agent-gateway" .

    # Copy files
    cp config.yaml.example "$TARGET_DIR/orchestrator/"
    cp README.md "$TARGET_DIR/orchestrator/" 2>/dev/null || true
    cp -r programs "$TARGET_DIR/orchestrator/" 2>/dev/null || true
fi

echo "✓ Files extracted"

# Setup configuration
CONFIG_FILE="$TARGET_DIR/orchestrator/config.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo ""
    echo "Setting up configuration..."
    # Try both config file naming conventions
    if [ -f "$TARGET_DIR/orchestrator/config.example.yaml" ]; then
        cp "$TARGET_DIR/orchestrator/config.example.yaml" "$CONFIG_FILE"
    elif [ -f "$TARGET_DIR/orchestrator/config.yaml.example" ]; then
        cp "$TARGET_DIR/orchestrator/config.yaml.example" "$CONFIG_FILE"
    else
        echo "⚠ No config example found, creating minimal config..."
        cat > "$CONFIG_FILE" << 'EOF'
server:
  port: 8080
  host: "0.0.0.0"
programs_dir: "./programs"
EOF
    fi

    # Prompt for basic config
    read -p "Server port [8080]: " PORT
    PORT=${PORT:-8080}

    read -p "Enable authentication? (y/n) [n]: " AUTH
    AUTH=${AUTH:-n}

    # Update config file
    sed -i "s/port: 8080/port: $PORT/" "$CONFIG_FILE" 2>/dev/null || \
        sed -i '' "s/port: 8080/port: $PORT/" "$CONFIG_FILE" 2>/dev/null || true

    echo "✓ Configuration created"
    echo "  Edit $CONFIG_FILE to customize"
else
    echo "✓ Existing configuration preserved"
fi

# Create systemd service (Linux only)
if command -v systemctl &> /dev/null; then
    echo ""
    read -p "Create systemd service? (y/n) [y]: " CREATE_SERVICE
    CREATE_SERVICE=${CREATE_SERVICE:-y}

    if [[ $CREATE_SERVICE =~ ^[Yy]$ ]]; then
        SERVICE_FILE="/etc/systemd/system/battlestag-orchestrator.service"

        cat > /tmp/battlestag-orchestrator.service << EOF
[Unit]
Description=BATTLESTAG Agent Gateway Orchestrator
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$TARGET_DIR/orchestrator
ExecStart=$TARGET_DIR/orchestrator/agent-gateway
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

        sudo mv /tmp/battlestag-orchestrator.service "$SERVICE_FILE"
        sudo systemctl daemon-reload

        echo "✓ Systemd service created"
        echo ""
        echo "To start the orchestrator:"
        echo "  sudo systemctl start battlestag-orchestrator"
        echo "  sudo systemctl enable battlestag-orchestrator  # Start on boot"
    fi
else
    echo ""
    echo "ℹ systemctl not found (macOS or non-systemd Linux)"
    echo "  Start manually: cd $TARGET_DIR/orchestrator && ./agent-gateway"
fi

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  Orchestrator Installation Complete!                       ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "Installation directory: $TARGET_DIR/orchestrator"
echo "Configuration file: $CONFIG_FILE"
echo ""
echo "Next steps:"
echo "  1. Edit configuration: nano $CONFIG_FILE"
echo "  2. Start orchestrator: cd $TARGET_DIR/orchestrator && ./agent-gateway"
echo "  3. Install programs: ./deploy.sh program install <program-name>"
echo ""
