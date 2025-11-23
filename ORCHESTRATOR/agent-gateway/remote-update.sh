#!/bin/bash

# Remote Update Script for Agent Gateway
# Run this from your local machine to update the remote server at 192.168.1.140

set -e

REMOTE_HOST="192.168.1.140"
REMOTE_USER="battlestag"  # Change if needed
REMOTE_PATH="~/Work/BATTLESTAG-BOT/ORCHESTRATOR/agent-gateway"
LOCAL_PATH="/home/battlestag/Work/BATTLESTAG-BOT/ORCHESTRATOR/agent-gateway"

echo "=========================================="
echo "  Agent Gateway - Remote Update"
echo "  Target: $REMOTE_USER@$REMOTE_HOST"
echo "=========================================="
echo ""

# Check if we can reach the remote server
echo "Testing connection to $REMOTE_HOST..."
if ! ssh -o ConnectTimeout=5 $REMOTE_USER@$REMOTE_HOST "echo 'Connection OK'" > /dev/null 2>&1; then
    echo "✗ Cannot connect to $REMOTE_USER@$REMOTE_HOST"
    echo ""
    echo "Please ensure:"
    echo "  1. Server is reachable"
    echo "  2. SSH is configured (try: ssh $REMOTE_USER@$REMOTE_HOST)"
    exit 1
fi
echo "✓ Connection successful"
echo ""

# Sync local changes to remote
echo "Syncing code to remote server..."
rsync -avz --exclude '.git' --exclude 'agent-gateway' \
    "$LOCAL_PATH/" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"

if [ $? -eq 0 ]; then
    echo "✓ Code synced"
else
    echo "✗ Sync failed"
    exit 1
fi
echo ""

# Run update.sh on remote server
echo "Running update.sh on remote server..."
echo ""
ssh -t $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_PATH && ./update.sh"

echo ""
echo "=========================================="
echo "  Remote Update Complete!"
echo "=========================================="
echo ""
echo "The remote server has been updated."
echo ""
echo "Check remote status:"
echo "  ssh $REMOTE_USER@$REMOTE_HOST 'sudo systemctl status agent-gateway'"
echo ""
echo "View remote logs:"
echo "  ssh $REMOTE_USER@$REMOTE_HOST 'sudo journalctl -u agent-gateway -f'"
echo ""
