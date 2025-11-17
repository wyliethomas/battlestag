#!/bin/bash

# Deployment script for stoic-thought to a remote server
# Usage: ./deploy.sh user@remote-host

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check arguments
if [ $# -lt 1 ]; then
    echo "Usage: $0 user@remote-host [ollama-host]"
    echo ""
    echo "Examples:"
    echo "  $0 user@192.168.1.100"
    echo "  $0 user@myserver.local http://localhost:11434"
    echo ""
    exit 1
fi

REMOTE_HOST="$1"
OLLAMA_HOST="${2:-http://localhost:11434}"
DEPLOY_DIR="/tmp/stoic-thought-deploy"

echo -e "${BLUE}=== Stoic Thought Generator - Deployment Script ===${NC}"
echo ""
echo "Target: $REMOTE_HOST"
echo ""

# Prompt for installation directories
echo -e "${YELLOW}Installation Configuration${NC}"
echo ""

read -p "Binary installation directory [/usr/local/bin]: " INSTALL_DIR
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

read -p "Data directory [/var/lib/stoic-thought]: " DATA_DIR
DATA_DIR="${DATA_DIR:-/var/lib/stoic-thought}"

read -p "Ollama host [$OLLAMA_HOST]: " OLLAMA_INPUT
OLLAMA_HOST="${OLLAMA_INPUT:-$OLLAMA_HOST}"

read -p "Ollama model [dolphin3]: " OLLAMA_MODEL
OLLAMA_MODEL="${OLLAMA_MODEL:-dolphin3}"

echo ""
echo -e "${BLUE}Deployment Settings:${NC}"
echo "  Target host:   $REMOTE_HOST"
echo "  Binary path:   $INSTALL_DIR/stoic-thought"
echo "  Data path:     $DATA_DIR"
echo "  Ollama host:   $OLLAMA_HOST"
echo "  Ollama model:  $OLLAMA_MODEL"
echo ""

read -p "Continue with deployment? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Deployment cancelled."
    exit 0
fi
echo ""

# Verify binary exists
if [ ! -f "stoic-thought" ]; then
    echo -e "${RED}Error: Binary 'stoic-thought' not found${NC}"
    echo "Please run 'go build -o stoic-thought' first"
    exit 1
fi

# Verify systemd files exist
if [ ! -f "stoic-thought.service" ] || [ ! -f "stoic-thought.timer" ]; then
    echo -e "${RED}Error: systemd files not found${NC}"
    exit 1
fi

# Verify remote-install.sh exists
if [ ! -f "remote-install.sh" ]; then
    echo -e "${RED}Error: remote-install.sh not found${NC}"
    exit 1
fi

# Test SSH connection
echo -e "${YELLOW}Testing SSH connection...${NC}"
if ! ssh -o ConnectTimeout=5 -o BatchMode=yes "$REMOTE_HOST" "echo 'SSH connection successful'" 2>/dev/null; then
    echo -e "${RED}Error: Cannot connect to $REMOTE_HOST${NC}"
    echo "Please ensure:"
    echo "  1. SSH is running on the remote host"
    echo "  2. You have SSH key-based authentication set up"
    echo "  3. The hostname/IP is correct"
    exit 1
fi
echo -e "${GREEN}✓ SSH connection successful${NC}"

# Create deployment directory on remote server
echo -e "${YELLOW}Creating deployment directory on remote server...${NC}"
ssh "$REMOTE_HOST" "mkdir -p $DEPLOY_DIR"

# Copy files to remote server
echo -e "${YELLOW}Copying files to remote server...${NC}"
FILES=(
    "stoic-thought"
    "stoic-thought.service"
    "stoic-thought.timer"
    "remote-install.sh"
)

for file in "${FILES[@]}"; do
    echo "  → $file"
    scp -q "$file" "$REMOTE_HOST:$DEPLOY_DIR/"
done
echo -e "${GREEN}✓ Files copied${NC}"

# Make remote-install.sh executable
ssh "$REMOTE_HOST" "chmod +x $DEPLOY_DIR/remote-install.sh"

# Execute remote installation
echo ""
echo -e "${BLUE}=== Running remote installation ===${NC}"
echo ""

ssh -t "$REMOTE_HOST" "cd $DEPLOY_DIR && ./remote-install.sh '$INSTALL_DIR' '$DATA_DIR' '$OLLAMA_HOST' '$OLLAMA_MODEL'"

# Cleanup
echo ""
echo -e "${YELLOW}Cleaning up deployment directory...${NC}"
ssh "$REMOTE_HOST" "rm -rf $DEPLOY_DIR"
echo -e "${GREEN}✓ Cleanup complete${NC}"

echo ""
echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo ""
echo "The stoic-thought generator has been deployed to $REMOTE_HOST"
echo ""
echo "Installation details:"
echo "  Binary:   $INSTALL_DIR/stoic-thought"
echo "  Database: $DATA_DIR/thoughts.db"
echo ""
echo "To manage the service on the remote server:"
echo "  ssh $REMOTE_HOST 'systemctl --user status stoic-thought.timer'"
echo "  ssh $REMOTE_HOST 'systemctl --user list-timers'"
echo "  ssh $REMOTE_HOST 'journalctl --user -u stoic-thought.service'"
echo ""
echo "To run manually:"
echo "  ssh $REMOTE_HOST '$INSTALL_DIR/stoic-thought'"
echo ""
