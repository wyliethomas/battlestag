# Quick Start Guide - USB Deployment

For data center deployments without internet access.

## Pre-Deployment (On Development Machine)

```bash
# 1. Build deployment packages
cd DEPLOYMENT
./build-packages.sh

# 2. Verify packages created
ls -lh packages/

# 3. Copy to USB drive
cp -r . /media/usb/battlestag/

# 4. Safely eject USB
sync
```

## On-Site Deployment (In Data Center)

### Step 1: Connect and Navigate

```bash
# Insert USB drive
# Mount if necessary
sudo mount /dev/sdb1 /mnt/usb

# Navigate to deployment directory
cd /mnt/usb/battlestag
```

### Step 2: Install Orchestrator

```bash
./deploy.sh orchestrator install
```

**You'll be prompted for:**
- Server port (default: 8080)
- Enable authentication (y/n)
- Create systemd service (y/n)

**Takes:** ~30 seconds

### Step 3: Start Orchestrator

```bash
# If using systemd
sudo systemctl start battlestag-orchestrator
sudo systemctl enable battlestag-orchestrator

# Or manually
cd ~/battlestag/orchestrator
./agent-gateway &
```

### Step 4: Install Programs

```bash
# Example: Install com-observer
./deploy.sh program install com-observer
```

**You'll be prompted for:**
- Your name
- Slack user ID
- Email address
- Slack bot token
- LLM provider (Ollama/Claude/OpenAI)
- Note writer (Obsidian/Notion)
- Setup cron automation (y/n)

**Takes:** ~1-2 minutes per program

### Step 5: Verify Installation

```bash
# Check orchestrator
curl http://localhost:8080/health

# Check program installation
ls -la ~/battlestag/programs/

# View logs
tail -f ~/battlestag/programs/com-observer/logs/*.log
```

## Common Scenarios

### Scenario 1: Minimal Setup (Just Orchestrator)

```bash
./deploy.sh orchestrator install
# Done! API available at :8080
```

**Use case:** API gateway only, programs installed later

### Scenario 2: Full Setup (Orchestrator + Programs)

```bash
# Install orchestrator
./deploy.sh orchestrator install

# Install monitoring
./deploy.sh program install lab-monitor

# Install task extraction
./deploy.sh program install com-observer

# Install task management
./deploy.sh program install task-manager
```

**Use case:** Complete personal AI agent setup

### Scenario 3: Client Handoff

```bash
# Install orchestrator
./deploy.sh orchestrator install

# Install client-specific programs
./deploy.sh program install <client-program>

# Configure for client environment
nano ~/battlestag/programs/<client-program>/config.yaml

# Start services
sudo systemctl start battlestag-orchestrator
```

**Use case:** Production client deployment

## Troubleshooting

### "Package not found" Error

**Cause:** Forgot to run `build-packages.sh`

**Fix:**
```bash
# On development machine
cd DEPLOYMENT
./build-packages.sh
# Re-copy to USB
```

### "Go not found" Error

**Cause:** Go not installed on target server

**Fix:**
```bash
# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### "Permission denied" Error

**Cause:** Scripts not executable

**Fix:**
```bash
chmod +x deploy.sh
chmod +x scripts/orchestrator/*.sh
chmod +x scripts/programs/*/*.sh
```

## Post-Deployment

### 1. Verify Everything Works

```bash
# Test orchestrator
curl http://localhost:8080/health

# Test program (example: com-observer)
cd ~/battlestag/programs/com-observer
./bin/slack -config config.yaml -state-dir state
```

### 2. Check Automation

```bash
# View cron jobs
crontab -l

# Check systemd services
sudo systemctl status battlestag-orchestrator
```

### 3. Monitor Logs

```bash
# Orchestrator logs
sudo journalctl -u battlestag-orchestrator -f

# Program logs
tail -f ~/battlestag/programs/*/logs/*.log
```

### 4. Document Configuration

```bash
# Save configuration for backup
cp ~/battlestag/orchestrator/config.yaml ~/battlestag-config-backup/
cp ~/battlestag/programs/*/config.yaml ~/battlestag-config-backup/
```

## Time Estimates

| Task | Time |
|------|------|
| USB Preparation | 5 min |
| Orchestrator Install | 30 sec |
| Program Install (each) | 1-2 min |
| Configuration | 5-10 min |
| Verification | 2 min |
| **Total (Full Setup)** | **15-20 min** |

## Checklist

Pre-deployment:
- [ ] Built packages with `./build-packages.sh`
- [ ] Verified packages exist in `packages/` directory
- [ ] Copied DEPLOYMENT to USB
- [ ] Tested USB is readable

On-site:
- [ ] Mounted USB drive
- [ ] Installed orchestrator
- [ ] Started orchestrator service
- [ ] Installed required programs
- [ ] Configured programs
- [ ] Verified all services running
- [ ] Checked logs for errors
- [ ] Documented configuration
- [ ] Created backup of configs

Post-deployment:
- [ ] Unmounted USB safely
- [ ] Monitored for 24 hours
- [ ] Verified automation (cron/systemd)
- [ ] Client handoff completed
