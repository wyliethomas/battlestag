#!/bin/bash

echo "=== Task Manager Uninstallation ==="
echo

INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="${CONFIG_DIR:-$HOME/.config/task-manager}"
DATA_DIR="${DATA_DIR:-$HOME/.local/share/task-manager}"

echo "This will remove:"
echo "  - Binaries from $INSTALL_DIR"
echo "  - Configuration from $CONFIG_DIR"
echo
read -p "Do you want to also remove the database? (y/N): " REMOVE_DATA
echo

# Remove binaries
echo "Removing binaries..."
rm -f "$INSTALL_DIR/task_project"
rm -f "$INSTALL_DIR/task_checklist"
rm -f "$INSTALL_DIR/task_journal"
rm -f "$INSTALL_DIR/task_query"

# Remove wrappers
rm -f "$INSTALL_DIR/task_project_run"
rm -f "$INSTALL_DIR/task_checklist_run"
rm -f "$INSTALL_DIR/task_journal_run"
rm -f "$INSTALL_DIR/task_query_run"

echo "✓ Binaries removed"

# Remove config
if [ -d "$CONFIG_DIR" ]; then
    echo "Removing configuration..."
    rm -rf "$CONFIG_DIR"
    echo "✓ Configuration removed"
fi

# Remove data if requested
if [[ "$REMOVE_DATA" =~ ^[Yy]$ ]]; then
    if [ -d "$DATA_DIR" ]; then
        echo "Removing database..."
        rm -rf "$DATA_DIR"
        echo "✓ Database removed"
    fi
else
    echo "Database preserved at: $DATA_DIR"
fi

echo
echo "✓ Uninstallation complete!"
echo
