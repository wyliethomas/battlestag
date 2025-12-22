#!/usr/bin/env bash
# deploy.sh - Deploy BATTLESTAG scripts to remote server
# Usage: ./deploy.sh user@server [install-path]

set -euo pipefail

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Check arguments
if [[ $# -lt 1 ]]; then
    echo "Usage: $0 user@server [install-path]"
    echo ""
    echo "Examples:"
    echo "  $0 user@server                    # Install to ~/.local/share/battlestag"
    echo "  $0 user@server /opt/battlestag    # Install to custom path"
    exit 1
fi

SERVER="$1"
REMOTE_PATH="${2:-}"

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${YELLOW}╔════════════════════════════════════════════╗${NC}"
echo -e "${YELLOW}║   BATTLESTAG Scripts Deployment           ║${NC}"
echo -e "${YELLOW}╚════════════════════════════════════════════╝${NC}"

# Test SSH connection
echo -e "${YELLOW}Testing connection to $SERVER...${NC}"
if ! ssh -o ConnectTimeout=5 "$SERVER" "echo 'Connected'" >/dev/null 2>&1; then
    echo -e "${RED}✗ Cannot connect to $SERVER${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Connection successful${NC}"

# Create temp directory for deployment
DEPLOY_TMP="/tmp/battlestag-deploy-$$"
mkdir -p "$DEPLOY_TMP"

# Copy scripts to temp
echo -e "${YELLOW}Preparing deployment package...${NC}"
cp -r "$SCRIPT_DIR" "$DEPLOY_TMP/scripts"

# Create remote installation script
cat > "$DEPLOY_TMP/install_remote.sh" <<'INSTALL_SCRIPT'
#!/usr/bin/env bash
set -euo pipefail

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}Installing BATTLESTAG scripts...${NC}"

# Check for required tools
for cmd in bash find chmod; do
    if ! command -v $cmd &>/dev/null; then
        echo -e "${RED}✗ Required command not found: $cmd${NC}"
        exit 1
    fi
done

# Install
cd /tmp/battlestag-deploy-*
./scripts/install.sh

# Cleanup
cd /
rm -rf /tmp/battlestag-deploy-*

echo -e "${GREEN}✓ Remote installation complete${NC}"
INSTALL_SCRIPT

chmod +x "$DEPLOY_TMP/install_remote.sh"

# Upload to server
echo -e "${YELLOW}Uploading to $SERVER...${NC}"
if ! scp -r "$DEPLOY_TMP" "$SERVER:/tmp/" >/dev/null 2>&1; then
    echo -e "${RED}✗ Upload failed${NC}"
    rm -rf "$DEPLOY_TMP"
    exit 1
fi
echo -e "${GREEN}✓ Upload complete${NC}"

# Run remote installation
echo -e "${YELLOW}Running remote installation...${NC}"
if ssh "$SERVER" "bash /tmp/$(basename "$DEPLOY_TMP")/install_remote.sh"; then
    echo -e "${GREEN}✓ Installation successful${NC}"
else
    echo -e "${RED}✗ Installation failed${NC}"
    rm -rf "$DEPLOY_TMP"
    exit 1
fi

# Cleanup local temp
rm -rf "$DEPLOY_TMP"

echo -e "\n${GREEN}╔════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   Deployment Complete!                     ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Test the installation:${NC}"
echo -e "  ssh $SERVER 'file.sh write /tmp/test.txt \"Hello from $SERVER\"'"
echo -e "  ssh $SERVER 'sys.sh hostname'"
echo ""
