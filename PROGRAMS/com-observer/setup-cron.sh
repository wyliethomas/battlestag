#!/bin/bash
# Setup script for com-observer cron jobs

set -e

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
USER_HOME="$HOME"

echo "Com-Observer Cron Setup"
echo "======================="
echo ""
echo "Installation directory: $SCRIPT_DIR"
echo "User home: $USER_HOME"
echo ""

# Create necessary directories
echo "Creating directories..."
mkdir -p "$SCRIPT_DIR/queue"
mkdir -p "$SCRIPT_DIR/state"
mkdir -p "$SCRIPT_DIR/logs"

# Check if config exists
if [ ! -f "$SCRIPT_DIR/config.yaml" ]; then
    echo "ERROR: config.yaml not found!"
    echo "Please create config.yaml before running this script."
    echo "See config.yaml.example for reference."
    exit 1
fi

# Check if binaries exist
if [ ! -f "$SCRIPT_DIR/bin/slack" ]; then
    echo "ERROR: Binaries not found!"
    echo "Please run 'make build' first."
    exit 1
fi

echo ""
echo "Ready to set up cron jobs."
echo ""
echo "The following cron jobs will be added:"
echo "  - Slack sync: every 5 minutes"
echo "  - Note writer: every minute"
echo ""
read -p "Continue? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Create temporary cron file
TEMP_CRON=$(mktemp)

# Get existing crontab (if any)
crontab -l > "$TEMP_CRON" 2>/dev/null || true

# Check if our jobs already exist
if grep -q "com-observer" "$TEMP_CRON"; then
    echo "Removing existing com-observer cron jobs..."
    grep -v "com-observer" "$TEMP_CRON" > "$TEMP_CRON.tmp" || true
    mv "$TEMP_CRON.tmp" "$TEMP_CRON"
fi

# Add our cron jobs
echo "" >> "$TEMP_CRON"
echo "# com-observer - Slack message monitoring" >> "$TEMP_CRON"
echo "*/5 * * * * cd $SCRIPT_DIR && ./bin/slack -config config.yaml -state-dir $SCRIPT_DIR/state >> $SCRIPT_DIR/logs/slack.log 2>&1" >> "$TEMP_CRON"
echo "" >> "$TEMP_CRON"
echo "# com-observer - Task extraction and writing" >> "$TEMP_CRON"
echo "* * * * * cd $SCRIPT_DIR && ./bin/notewriter -config config.yaml >> $SCRIPT_DIR/logs/notewriter.log 2>&1" >> "$TEMP_CRON"

# Install new crontab
crontab "$TEMP_CRON"
rm "$TEMP_CRON"

echo ""
echo "✓ Cron jobs installed successfully!"
echo ""
echo "Monitoring schedule:"
echo "  - Slack: Every 5 minutes"
echo "  - Processing: Every minute"
echo ""
echo "Logs location:"
echo "  - Slack: $SCRIPT_DIR/logs/slack.log"
echo "  - Tasks: $SCRIPT_DIR/logs/notewriter.log"
echo ""
echo "To view logs:"
echo "  tail -f $SCRIPT_DIR/logs/slack.log"
echo "  tail -f $SCRIPT_DIR/logs/notewriter.log"
echo ""
echo "To remove cron jobs:"
echo "  crontab -e  # Then delete the com-observer lines"
echo ""
echo "First sync will run in the next minute."
