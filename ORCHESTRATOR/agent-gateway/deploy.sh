#!/bin/bash

# Agent Gateway Deployment Script
# This script builds and deploys the Agent Gateway API as a systemd service

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="agent-gateway"
BINARY_NAME="agent-gateway"
INSTALL_DIR="/opt/agent-gateway"
CONFIG_DIR="/etc/agent-gateway"
LOG_DIR="/var/log/agent-gateway"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

# Functions
print_step() {
    echo -e "${BLUE}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

check_dependencies() {
    print_step "Checking dependencies..."

    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
        exit 1
    fi
    print_success "Go is installed: $(go version)"
}

build_application() {
    print_step "Building application..."

    cd "$SCRIPT_DIR"

    # Clean previous builds
    rm -f "$BINARY_NAME"

    # Download dependencies
    go mod download

    # Build
    go build -ldflags="-s -w" -o "$BINARY_NAME" .

    if [[ ! -f "$BINARY_NAME" ]]; then
        print_error "Build failed"
        exit 1
    fi

    print_success "Build complete"
}

stop_service() {
    print_step "Stopping service if running..."

    if systemctl is-active --quiet "$SERVICE_NAME"; then
        systemctl stop "$SERVICE_NAME"
        print_success "Service stopped"
    else
        print_warning "Service was not running"
    fi
}

create_directories() {
    print_step "Creating directories..."

    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOG_DIR"

    # Set ownership to the user who will run the service
    chown -R battlestag:battlestag "$INSTALL_DIR"
    chown -R battlestag:battlestag "$CONFIG_DIR"
    chown -R battlestag:battlestag "$LOG_DIR"

    print_success "Directories created"
}

install_binary() {
    print_step "Installing binary..."

    cp "$SCRIPT_DIR/$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    print_success "Binary installed to $INSTALL_DIR/$BINARY_NAME"
}

install_config() {
    print_step "Installing configuration..."

    if [[ -f "$CONFIG_DIR/config.yaml" ]]; then
        print_warning "Configuration already exists at $CONFIG_DIR/config.yaml"
        print_warning "Backing up to $CONFIG_DIR/config.yaml.bak"
        cp "$CONFIG_DIR/config.yaml" "$CONFIG_DIR/config.yaml.bak"
    fi

    if [[ -f "$SCRIPT_DIR/config.yaml" ]]; then
        cp "$SCRIPT_DIR/config.yaml" "$CONFIG_DIR/"
        print_success "Configuration copied from existing config.yaml"
    elif [[ -f "$SCRIPT_DIR/config.yaml.example" ]]; then
        cp "$SCRIPT_DIR/config.yaml.example" "$CONFIG_DIR/config.yaml"
        print_warning "Copied example config - YOU MUST EDIT IT!"
        print_warning "Edit: $CONFIG_DIR/config.yaml"
        NEED_CONFIG_EDIT=1
    else
        print_error "No configuration file found"
        exit 1
    fi

    # Update log paths in config to use system log directory
    sed -i "s|logs/access.log|$LOG_DIR/access.log|g" "$CONFIG_DIR/config.yaml"
    sed -i "s|logs/error.log|$LOG_DIR/error.log|g" "$CONFIG_DIR/config.yaml"

    chown battlestag:battlestag "$CONFIG_DIR/config.yaml"
    chmod 600 "$CONFIG_DIR/config.yaml"

    print_success "Configuration installed to $CONFIG_DIR/config.yaml"
}

install_service() {
    print_step "Installing systemd service..."

    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Agent Gateway API Server
After=network.target
Documentation=https://github.com/battlestag/agent-gateway

[Service]
Type=simple
User=battlestag
Group=battlestag
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME --config $CONFIG_DIR/config.yaml

# Restart policy
Restart=always
RestartSec=5

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

    chmod 644 "$SERVICE_FILE"
    systemctl daemon-reload

    print_success "Service installed to $SERVICE_FILE"
}

enable_service() {
    print_step "Enabling service..."

    systemctl enable "$SERVICE_NAME"

    print_success "Service enabled to start on boot"
}

start_service() {
    print_step "Starting service..."

    systemctl start "$SERVICE_NAME"
    sleep 2

    if systemctl is-active --quiet "$SERVICE_NAME"; then
        print_success "Service started successfully"
    else
        print_error "Service failed to start"
        print_error "Check logs with: sudo journalctl -u $SERVICE_NAME -n 50"
        exit 1
    fi
}

show_status() {
    print_step "Service status:"
    systemctl status "$SERVICE_NAME" --no-pager || true
    echo ""
}

print_summary() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${GREEN}Deployment Complete!${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Installation Details:"
    echo "  • Binary:        $INSTALL_DIR/$BINARY_NAME"
    echo "  • Config:        $CONFIG_DIR/config.yaml"
    echo "  • Logs:          $LOG_DIR/"
    echo "  • Service:       $SERVICE_FILE"
    echo ""
    echo "Useful Commands:"
    echo "  • Check status:  sudo systemctl status $SERVICE_NAME"
    echo "  • View logs:     sudo journalctl -u $SERVICE_NAME -f"
    echo "  • Restart:       sudo systemctl restart $SERVICE_NAME"
    echo "  • Stop:          sudo systemctl stop $SERVICE_NAME"
    echo "  • Disable:       sudo systemctl disable $SERVICE_NAME"
    echo ""
    echo "API Endpoints:"
    echo "  • Health:        curl http://localhost:8080/api/health"
    echo "  • Stats:         curl -H \"X-API-Key: your-key\" http://localhost:8080/api/stats"
    echo "  • Stoic Today:   curl -H \"X-API-Key: your-key\" http://localhost:8080/api/stoic/today"
    echo "  • Tech Random:   curl -H \"X-API-Key: your-key\" http://localhost:8080/api/tech/random"
    echo ""

    if [[ $NEED_CONFIG_EDIT -eq 1 ]]; then
        echo -e "${YELLOW}⚠ IMPORTANT:${NC}"
        echo "  You MUST edit the configuration file before the service will work:"
        echo "  sudo nano $CONFIG_DIR/config.yaml"
        echo ""
        echo "  Update these settings:"
        echo "    - Database paths for stoic and tech agents"
        echo "    - API key (generate with: openssl rand -hex 32)"
        echo ""
        echo "  Then restart: sudo systemctl restart $SERVICE_NAME"
        echo ""
    fi

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Main deployment flow
main() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Agent Gateway Deployment"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    check_root
    check_dependencies
    build_application
    stop_service
    create_directories
    install_binary
    install_config
    install_service
    enable_service
    start_service
    show_status
    print_summary
}

# Handle command line arguments
case "${1:-deploy}" in
    deploy)
        main
        ;;
    uninstall)
        print_step "Uninstalling Agent Gateway..."
        check_root

        # Stop and disable service
        if systemctl is-active --quiet "$SERVICE_NAME"; then
            systemctl stop "$SERVICE_NAME"
        fi
        systemctl disable "$SERVICE_NAME" 2>/dev/null || true

        # Remove files
        rm -f "$SERVICE_FILE"
        systemctl daemon-reload

        print_warning "Keeping data directories (config, logs)"
        print_warning "To remove completely, run: sudo rm -rf $INSTALL_DIR $CONFIG_DIR $LOG_DIR"

        print_success "Service uninstalled"
        ;;
    restart)
        print_step "Restarting service..."
        check_root
        systemctl restart "$SERVICE_NAME"
        sleep 2
        systemctl status "$SERVICE_NAME" --no-pager
        ;;
    status)
        systemctl status "$SERVICE_NAME"
        ;;
    logs)
        journalctl -u "$SERVICE_NAME" -f
        ;;
    build-only)
        build_application
        print_success "Build complete: $SCRIPT_DIR/$BINARY_NAME"
        ;;
    help|--help|-h)
        echo "Agent Gateway Deployment Script"
        echo ""
        echo "Usage: sudo ./deploy.sh [command]"
        echo ""
        echo "Commands:"
        echo "  deploy       - Build and deploy the service (default)"
        echo "  uninstall    - Stop and remove the service"
        echo "  restart      - Restart the service"
        echo "  status       - Show service status"
        echo "  logs         - Follow service logs"
        echo "  build-only   - Only build the binary"
        echo "  help         - Show this help message"
        echo ""
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Run './deploy.sh help' for usage information"
        exit 1
        ;;
esac
