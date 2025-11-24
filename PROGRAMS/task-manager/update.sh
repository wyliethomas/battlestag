#!/bin/bash
set -e

echo "=== Task Manager Update ==="
echo

INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="${CONFIG_DIR:-$HOME/.config/task-manager}"

# Check if task-manager is installed
if [ ! -f "$INSTALL_DIR/task_project" ]; then
    echo "Error: Task Manager is not installed in $INSTALL_DIR"
    echo "Please run install.sh first"
    exit 1
fi

echo "Updating task-manager..."
echo "  Install directory: $INSTALL_DIR"
echo

# Build executables
echo "Building executables..."
go build -o task_project ./cmd/project
go build -o task_checklist ./cmd/checklist
go build -o task_journal ./cmd/journal
go build -o task_query ./cmd/query

# Install binaries
echo "Installing binaries..."
cp task_project "$INSTALL_DIR/"
cp task_checklist "$INSTALL_DIR/"
cp task_journal "$INSTALL_DIR/"
cp task_query "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/task_project"
chmod +x "$INSTALL_DIR/task_checklist"
chmod +x "$INSTALL_DIR/task_journal"
chmod +x "$INSTALL_DIR/task_query"

echo
echo "✓ Update complete!"
echo
echo "Binaries updated:"
echo "  $INSTALL_DIR/task_project"
echo "  $INSTALL_DIR/task_checklist"
echo "  $INSTALL_DIR/task_journal"
echo "  $INSTALL_DIR/task_query"
echo
echo "Configuration and data preserved:"
echo "  $CONFIG_DIR/.env"
echo
