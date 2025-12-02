#!/bin/bash
# Update com-observer program

set -e

TARGET_DIR="${BATTLESTAG_TARGET_DIR:-$HOME/battlestag}"
SCRIPT_DIR="${BATTLESTAG_SCRIPT_DIR}"
PROGRAM_DIR="$TARGET_DIR/programs/com-observer"

echo "Updating com-observer"
echo "===================="
echo ""

# Check if installed
if [ ! -d "$PROGRAM_DIR" ]; then
    echo "✗ com-observer not installed"
    echo "  Run: ./deploy.sh program install com-observer"
    exit 1
fi

# Backup binaries
echo "Creating backup..."
cp -r "$PROGRAM_DIR/bin" "$PROGRAM_DIR/bin.backup"
echo "✓ Backup created"

# Update from package or source
PACKAGES_DIR="$SCRIPT_DIR/packages"
if [ -f "$PACKAGES_DIR/com-observer.tar.gz" ]; then
    echo "Extracting new version..."
    tar -xzf "$PACKAGES_DIR/com-observer.tar.gz" -C "/tmp"
    cp -r "/tmp/bin" "$PROGRAM_DIR/"
    rm -rf "/tmp/bin"
    echo "✓ Binaries updated"
else
    echo "✗ Update package not found"
    echo "  Restoring backup..."
    rm -rf "$PROGRAM_DIR/bin"
    mv "$PROGRAM_DIR/bin.backup" "$PROGRAM_DIR/bin"
    exit 1
fi

echo ""
echo "✓ com-observer updated successfully!"
echo "  Backup available at: $PROGRAM_DIR/bin.backup"
