#!/bin/bash
# BATTLESTAG Deployment System
# Supports both local (USB) and remote (SSH) deployment

set -e

VERSION="1.0.0"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
print_header() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC}  BATTLESTAG Deployment System v${VERSION}                   ${BLUE}║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Usage information
usage() {
    cat << EOF
Usage: $0 [OPTIONS] COMPONENT ACTION

COMPONENTS:
    orchestrator    Agent Gateway orchestrator
    program         Individual programs (com-observer, lab-monitor, etc.)

ACTIONS:
    install         Install component
    update          Update existing installation
    uninstall       Remove component
    status          Check installation status

OPTIONS:
    --remote USER@HOST      Deploy to remote server via SSH
    --target-dir DIR        Installation directory (default: ~/battlestag)
    --help                  Show this help message

EXAMPLES:
    # Local installation from USB
    ./deploy.sh orchestrator install

    # Remote installation via SSH
    ./deploy.sh --remote user@server orchestrator install

    # Install a program
    ./deploy.sh program install com-observer

    # Update orchestrator remotely
    ./deploy.sh --remote user@server orchestrator update

OFFLINE DEPLOYMENT:
    Copy entire DEPLOYMENT directory to USB drive
    Mount USB on target server
    Run deploy.sh from USB location

EOF
}

# Parse command line arguments
REMOTE=""
TARGET_DIR="$HOME/battlestag"
COMPONENT=""
ACTION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --remote)
            REMOTE="$2"
            shift 2
            ;;
        --target-dir)
            TARGET_DIR="$2"
            shift 2
            ;;
        --help)
            usage
            exit 0
            ;;
        orchestrator|program)
            COMPONENT="$1"
            shift
            ;;
        install|update|uninstall|status)
            ACTION="$1"
            shift
            ;;
        *)
            if [[ "$COMPONENT" == "program" ]] && [[ -z "$PROGRAM_NAME" ]]; then
                PROGRAM_NAME="$1"
                shift
            else
                print_error "Unknown argument: $1"
                usage
                exit 1
            fi
            ;;
    esac
done

# Validate arguments
if [[ -z "$COMPONENT" ]] || [[ -z "$ACTION" ]]; then
    print_error "Missing required arguments"
    usage
    exit 1
fi

if [[ "$COMPONENT" == "program" ]] && [[ -z "$PROGRAM_NAME" ]]; then
    print_error "Program name required for 'program' component"
    echo ""
    echo "Available programs:"
    ls -1 "$SCRIPT_DIR/scripts/programs" 2>/dev/null || echo "  (none found)"
    exit 1
fi

# Remote deployment
if [[ -n "$REMOTE" ]]; then
    print_header
    print_info "Remote deployment mode"
    print_info "Target: $REMOTE"
    print_info "Creating deployment package..."

    # Create temporary deployment package
    TEMP_PKG=$(mktemp -d)
    cp -r "$SCRIPT_DIR"/* "$TEMP_PKG/"

    # Create tarball
    TARBALL="/tmp/battlestag-deploy-$$.tar.gz"
    tar -czf "$TARBALL" -C "$TEMP_PKG" .

    print_info "Transferring to remote server..."
    ssh "$REMOTE" "mkdir -p /tmp/battlestag-deploy-$$"
    scp -q "$TARBALL" "$REMOTE:/tmp/battlestag-deploy-$$/deploy.tar.gz"

    print_info "Executing remote installation..."
    ssh "$REMOTE" "cd /tmp/battlestag-deploy-$$ && tar -xzf deploy.tar.gz && ./deploy.sh --target-dir $TARGET_DIR $COMPONENT $ACTION $PROGRAM_NAME"

    # Cleanup
    ssh "$REMOTE" "rm -rf /tmp/battlestag-deploy-$$"
    rm -rf "$TEMP_PKG" "$TARBALL"

    print_success "Remote deployment complete"
    exit 0
fi

# Local deployment
print_header
print_info "Local deployment mode"
print_info "Component: $COMPONENT"
print_info "Action: $ACTION"
print_info "Target directory: $TARGET_DIR"
echo ""

# Route to appropriate installer
case "$COMPONENT" in
    orchestrator)
        INSTALLER="$SCRIPT_DIR/scripts/orchestrator/${ACTION}.sh"
        ;;
    program)
        INSTALLER="$SCRIPT_DIR/scripts/programs/${PROGRAM_NAME}/${ACTION}.sh"
        ;;
esac

if [[ ! -f "$INSTALLER" ]]; then
    print_error "Installer not found: $INSTALLER"
    exit 1
fi

# Execute installer
export BATTLESTAG_TARGET_DIR="$TARGET_DIR"
export BATTLESTAG_SCRIPT_DIR="$SCRIPT_DIR"
export BATTLESTAG_PROGRAM_NAME="$PROGRAM_NAME"

bash "$INSTALLER"

print_success "Deployment complete!"
