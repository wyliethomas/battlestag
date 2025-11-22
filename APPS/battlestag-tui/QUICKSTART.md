# BATTLESTAG TUI - Quick Start Guide

## What is it?

A beautiful terminal user interface for managing your financial data through the BATTLESTAG Agent Gateway API.

## Prerequisites

1. Agent Gateway API running at http://192.168.1.140:8080
2. Network access to 192.168.1.140
3. Valid API key (default: test-api-key-12345)

## Installation & First Run

### Option 1: Using the run script (easiest)

```bash
cd /home/battlestag/Work/BATTLESTAG-BOT/APPS/battlestag-tui

# Run with defaults (no config needed!)
./run.sh

# Or override defaults if needed:
export AGENT_GATEWAY_URL="http://192.168.1.140:8080"
export AGENT_GATEWAY_API_KEY="test-api-key-12345"
./run.sh
```

### Option 2: Manual build

```bash
cd /home/battlestag/Work/BATTLESTAG-BOT/APPS/battlestag-tui

# Build
./build.sh

# Run (uses defaults automatically)
./battlestag-tui

# Or set custom values:
export AGENT_GATEWAY_URL="http://192.168.1.140:8080"
export AGENT_GATEWAY_API_KEY="test-api-key-12345"
./battlestag-tui
```

## First Time Usage

1. **Dashboard Screen (Tab 1)**
   - View your net worth overview
   - See today's stoic thought
   - Check system status

2. **Assets Screen (Tab 2)**
   - Browse your assets
   - Use arrow keys to navigate
   - Press `d` to delete
   - Press `r` to refresh

3. **Liabilities Screen (Tab 3)**
   - View credit cards, loans, etc.
   - Navigate and manage like assets

4. **Upload Screen (Tab 4)**
   - Type or paste PDF file path
   - Press Enter to upload
   - Statements are automatically processed

## Keyboard Cheat Sheet

```
Navigation:
  1, 2, 3, 4  - Switch between screens
  ↑/↓ or k/j  - Navigate lists
  r           - Refresh data
  d           - Delete selected item
  ?           - Help screen
  q           - Quit

Upload screen:
  Type path   - Enter file path
  Enter       - Upload file
  Ctrl+U      - Clear input
```

## Troubleshooting

### "Disconnected" status in header

Check:
1. Is the Agent Gateway running on the remote server?
   ```bash
   curl http://192.168.1.140:8080/api/health
   ```

2. Can you reach the server?
   ```bash
   ping 192.168.1.140
   ```

3. Is the URL correct? (should be http://192.168.1.140:8080)
   ```bash
   echo $AGENT_GATEWAY_URL
   ```

4. Is the API key correct? (should be test-api-key-12345)
   ```bash
   echo $AGENT_GATEWAY_API_KEY
   ```

### Build errors

```bash
# Clean and rebuild
cd /home/battlestag/Work/BATTLESTAG-BOT/APPS/battlestag-tui
go clean -modcache
go mod tidy
./build.sh
```

## What's Next?

- Explore the dashboard to see your financial overview
- Upload bank statements to track transactions
- Add assets and liabilities manually (future feature)
- Check daily stoic thoughts for wisdom

## Getting Help

- Press `?` in the TUI for help
- Check README.md for detailed documentation
- Report issues at the project repository
