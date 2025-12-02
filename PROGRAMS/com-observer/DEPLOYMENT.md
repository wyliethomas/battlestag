# Com-Observer Deployment Guide

## Overview
Deploy com-observer to a remote server with automated Slack monitoring via cron.

## Prerequisites on Remote Server
- Go 1.21+ installed
- Obsidian vault synced/accessible
- Ollama running (or accessible Ollama instance)
- SSH access to the server

## Step 1: Transfer Project

From your local machine:
```bash
# Package the project (already done)
# Transfer to remote
scp ~/com-observer-deploy.tar.gz user@remote-server:~/

# SSH to remote
ssh user@remote-server

# Extract
cd ~
mkdir -p com-observer
cd com-observer
tar -xzf ../com-observer-deploy.tar.gz
```

## Step 2: Build on Remote Server

```bash
cd ~/com-observer
go mod download
make build

# Verify binaries
ls -lh bin/
```

## Step 3: Configure for Remote

Create `config.yaml` on the remote server:

```yaml
user:
  name: "Wylie Thomas"
  slack_id: "U0NHL7JBF"
  email: "wylie@pxp200.com"

queue:
  type: "filesystem"
  path: "/home/YOUR_USER/com-observer/queue"

slack:
  token: "YOUR_SLACK_TOKEN_HERE"
  channels_include: []
  channels_exclude: []
  watch_dms: true
  sync_interval: "5m"

notewriter:
  type: "obsidian"
  obsidian:
    vault_path: "/path/to/obsidian/vault/on/remote"  # UPDATE THIS
    tasks_file: "Tasks/Action Items.md"

llm:
  provider: "openai"
  api_key: "not-needed"
  model: "llama3.1:8b"
  endpoint: "http://localhost:11434/v1/chat/completions"  # or remote Ollama IP
```

## Step 4: Test Manually

```bash
# Create directories
mkdir -p ~/com-observer/queue
mkdir -p ~/com-observer/state

# Test Slack sync
./bin/slack -config config.yaml -state-dir ~/com-observer/state

# Test notewriter
./bin/notewriter -config config.yaml
```

## Step 5: Set Up Cron Jobs

Edit crontab:
```bash
crontab -e
```

Add these lines:
```cron
# Sync Slack every 5 minutes
*/5 * * * * cd /home/YOUR_USER/com-observer && ./bin/slack -config config.yaml -state-dir /home/YOUR_USER/com-observer/state >> /home/YOUR_USER/com-observer/logs/slack.log 2>&1

# Process queue every minute
* * * * * cd /home/YOUR_USER/com-observer && ./bin/notewriter -config config.yaml >> /home/YOUR_USER/com-observer/logs/notewriter.log 2>&1

# Optional: Sync Gmail every 10 minutes (if configured)
# */10 * * * * cd /home/YOUR_USER/com-observer && ./bin/gmail -config config.yaml -state-dir /home/YOUR_USER/com-observer/state >> /home/YOUR_USER/com-observer/logs/gmail.log 2>&1

# Optional: Sync Trello every 15 minutes (if configured)
# */15 * * * * cd /home/YOUR_USER/com-observer && ./bin/trello -config config.yaml -state-dir /home/YOUR_USER/com-observer/state >> /home/YOUR_USER/com-observer/logs/trello.log 2>&1
```

Create logs directory:
```bash
mkdir -p ~/com-observer/logs
```

## Step 6: Monitor

Watch the logs:
```bash
# Watch Slack sync
tail -f ~/com-observer/logs/slack.log

# Watch task processing
tail -f ~/com-observer/logs/notewriter.log

# Check queue status
ls -lh ~/com-observer/queue/
```

## Mac Mini Specific Notes

### If using macOS on Mac Mini:

**Install Homebrew** (if not already):
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

**Install Go**:
```bash
brew install go
```

**Install Ollama**:
```bash
brew install ollama
ollama serve &
ollama pull llama3.1:8b
```

**Set up Obsidian Sync**:
- Option 1: Use Obsidian Sync service
- Option 2: Use git to sync vault
- Option 3: Use rsync/iCloud Drive

**Keep Mac Mini awake**:
```bash
# Prevent sleep while running
sudo pmset -c sleep 0
sudo pmset -c displaysleep 10
```

### Using launchd (macOS alternative to cron):

Create `~/Library/LaunchAgents/com.battlestag.com-observer.slack.plist`:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.battlestag.com-observer.slack</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/YOUR_USER/com-observer/bin/slack</string>
        <string>-config</string>
        <string>/Users/YOUR_USER/com-observer/config.yaml</string>
        <string>-state-dir</string>
        <string>/Users/YOUR_USER/com-observer/state</string>
    </array>
    <key>StartInterval</key>
    <integer>300</integer>
    <key>StandardOutPath</key>
    <string>/Users/YOUR_USER/com-observer/logs/slack.log</string>
    <key>StandardErrorPath</key>
    <string>/Users/YOUR_USER/com-observer/logs/slack.log</string>
</dict>
</plist>
```

Load it:
```bash
launchctl load ~/Library/LaunchAgents/com.battlestag.com-observer.slack.plist
```

## Troubleshooting

**Queue fills up but tasks not written:**
- Check Ollama is running: `curl http://localhost:11434/api/tags`
- Check notewriter logs

**Slack sync fails:**
- Verify token in config.yaml
- Check network connectivity to Slack API
- Ensure bot is still installed in workspace

**Obsidian file not updating:**
- Check vault path is correct
- Verify Tasks directory exists
- Check file permissions

**Tasks not syncing to other devices:**
- Ensure Obsidian Sync/git is working
- Check vault is being updated (ls -lh on the file)

## Security Notes

- Keep `config.yaml` secure (contains API tokens)
- Consider using environment variables for secrets
- Restrict file permissions: `chmod 600 config.yaml`
- Don't commit config.yaml to git (already in .gitignore)

## Maintenance

**View queue size:**
```bash
ls ~/com-observer/queue/ | wc -l
```

**Clear old logs:**
```bash
# Keep last 7 days
find ~/com-observer/logs -name "*.log" -mtime +7 -delete
```

**Update the application:**
```bash
cd ~/com-observer
git pull  # if using git
make build
# Cron will pick up new binaries automatically
```
