#!/bin/bash
# Build deployment packages for offline installation

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PACKAGES_DIR="$SCRIPT_DIR/packages"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

echo "╔════════════════════════════════════════════════════════════╗"
echo "║  BATTLESTAG Package Builder                                ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Create packages directory
mkdir -p "$PACKAGES_DIR"

# Build Orchestrator
echo "Building orchestrator..."
cd "$ROOT_DIR/ORCHESTRATOR/agent-gateway-v2"
go build -o agent-gateway ./main.go
tar -czf "$PACKAGES_DIR/orchestrator.tar.gz" \
    agent-gateway \
    config.yaml.example \
    README.md \
    programs/
echo "✓ orchestrator.tar.gz created"

# Build Programs
echo ""
echo "Building programs..."

# com-observer
if [ -d "$ROOT_DIR/PROGRAMS/com-observer" ]; then
    echo "  Building com-observer..."
    cd "$ROOT_DIR/PROGRAMS/com-observer"
    make build
    tar -czf "$PACKAGES_DIR/com-observer.tar.gz" \
        bin/ \
        config.yaml.example \
        README.md \
        Makefile \
        go.mod \
        go.sum \
        cmd/ \
        internal/
    echo "  ✓ com-observer.tar.gz created"
fi

# lab-monitor
if [ -d "$ROOT_DIR/PROGRAMS/lab-monitor" ]; then
    echo "  Building lab-monitor..."
    cd "$ROOT_DIR/PROGRAMS/lab-monitor"
    go build -o lab-monitor .
    tar -czf "$PACKAGES_DIR/lab-monitor.tar.gz" \
        lab-monitor \
        manifest.json \
        README.md
    echo "  ✓ lab-monitor.tar.gz created"
fi

# task-manager
if [ -d "$ROOT_DIR/PROGRAMS/task-manager" ]; then
    echo "  Building task-manager..."
    cd "$ROOT_DIR/PROGRAMS/task-manager"
    make build
    tar -czf "$PACKAGES_DIR/task-manager.tar.gz" \
        task_* \
        manifest.json \
        README.md
    echo "  ✓ task-manager.tar.gz created"
fi

echo ""
echo "Package summary:"
ls -lh "$PACKAGES_DIR"/*.tar.gz 2>/dev/null || echo "  No packages created"

echo ""
echo "✓ Package building complete!"
echo ""
echo "To prepare USB drive:"
echo "  1. Copy entire DEPLOYMENT directory to USB"
echo "  2. USB is ready for offline deployment"
