#!/bin/bash
# Uninstall com-observer program

set -e

TARGET_DIR="${BATTLESTAG_TARGET_DIR:-$HOME/battlestag}"
PROGRAM_DIR="$TARGET_DIR/programs/com-observer"

echo "Uninstalling com-observer"
echo "========================="
echo ""

# Check if installed
if [ ! -d "$PROGRAM_DIR" ]; then
    echo "✗ com-observer not installed"
    exit 0
fi

# Warning
echo "⚠ This will remove com-observer and all its data"
echo "  Installation directory: $PROGRAM_DIR"
echo ""
read -p "Continue? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Uninstall cancelled."
    exit 0
fi

# Remove cron jobs
if command -v crontab &> /dev/null; then
    TEMP_CRON=$(mktemp)
    crontab -l > "$TEMP_CRON" 2>/dev/null || true
    if grep -q "com-observer" "$TEMP_CRON"; then
        echo "Removing cron jobs..."
        grep -v "com-observer" "$TEMP_CRON" > "$TEMP_CRON.tmp" || true
        mv "$TEMP_CRON.tmp" "$TEMP_CRON"
        crontab "$TEMP_CRON"
        echo "✓ Cron jobs removed"
    fi
    rm "$TEMP_CRON"
fi

# Remove installation directory
echo "Removing installation directory..."
rm -rf "$PROGRAM_DIR"
echo "✓ Directory removed"

echo ""
echo "✓ com-observer uninstalled successfully"
