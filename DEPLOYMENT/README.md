# BATTLESTAG Deployment System

Professional deployment tooling for BATTLESTAG orchestrator and programs. Supports both offline (USB) and remote (SSH) installation.

## Overview

This deployment system provides:
- **Offline Installation**: Deploy from USB in air-gapped data centers
- **Remote Installation**: Deploy via SSH to remote servers
- **Automated Setup**: Guided configuration and service installation
- **Update Management**: Safe updates with automatic backups
- **Clean Uninstall**: Complete removal of components

## Quick Start

### Prepare USB Drive (One-Time Setup)

```bash
# From development machine
cd DEPLOYMENT
./build-packages.sh

# Copy entire DEPLOYMENT directory to USB
cp -r . /media/usb/battlestag-deployment/
```

### Deploy in Data Center (Offline Mode)

```bash
# Insert USB on target server
cd /media/usb/battlestag-deployment

# Install orchestrator
./deploy.sh orchestrator install

# Install programs
./deploy.sh program install com-observer
```

### Deploy Remotely (Online Mode)

```bash
# From your local machine
cd DEPLOYMENT

# Install orchestrator on remote server
./deploy.sh --remote user@server orchestrator install

# Install program remotely
./deploy.sh --remote user@server program install com-observer
```

## Components

### Orchestrator (Agent Gateway)

The central orchestrator manages programs and provides API endpoints.

**Install:**
```bash
./deploy.sh orchestrator install
```

**Update:**
```bash
./deploy.sh orchestrator update
```

**Uninstall:**
```bash
./deploy.sh orchestrator uninstall
```

### Programs

Individual functional programs that integrate with the orchestrator.

**Available Programs:**
- `com-observer` - Monitors Slack/Gmail/Trello and extracts tasks
- `lab-monitor` - Server monitoring and health checks
- `task-manager` - Task and project management

**Install:**
```bash
./deploy.sh program install <program-name>
```

**Update:**
```bash
./deploy.sh program update <program-name>
```

**Uninstall:**
```bash
./deploy.sh program uninstall <program-name>
```

## Installation Modes

### USB/Offline Mode

**Use Case:** Air-gapped environments, data centers without internet access

**Prerequisites:**
- USB drive with DEPLOYMENT directory
- Pre-built packages (run `./build-packages.sh` before copying to USB)

**Workflow:**
1. Prepare USB with `./build-packages.sh`
2. Insert USB on target server
3. Run deployment scripts from USB
4. All binaries install from pre-built packages

**Advantages:**
- No internet required
- Faster installation
- Consistent versions

### Remote/SSH Mode

**Use Case:** Remote servers with SSH access

**Prerequisites:**
- SSH access to target server
- SSH key authentication (recommended)

**Workflow:**
1. Run deploy script with `--remote` flag
2. Script packages files and transfers via SCP
3. Executes installation on remote server
4. Cleans up temporary files

**Advantages:**
- Deploy from anywhere
- Centralized control
- Multiple server deployments

## Directory Structure

```
DEPLOYMENT/
├── deploy.sh                 # Master deployment script
├── build-packages.sh         # Package builder for offline mode
├── README.md                 # This file
├── packages/                 # Pre-built binaries (created by build-packages.sh)
│   ├── orchestrator.tar.gz
│   ├── com-observer.tar.gz
│   ├── lab-monitor.tar.gz
│   └── task-manager.tar.gz
├── scripts/
│   ├── orchestrator/
│   │   ├── install.sh
│   │   ├── update.sh
│   │   └── uninstall.sh
│   └── programs/
│       ├── com-observer/
│       │   ├── install.sh
│       │   ├── update.sh
│       │   └── uninstall.sh
│       ├── lab-monitor/
│       └── task-manager/
└── configs/                  # Configuration templates
```

## Installation Locations

Default installation directory: `~/battlestag`

**Orchestrator:**
```
~/battlestag/orchestrator/
├── agent-gateway             # Binary
├── config.yaml               # Configuration
├── programs/                 # Built-in programs
└── README.md
```

**Programs:**
```
~/battlestag/programs/<program-name>/
├── bin/                      # Executables
├── config.yaml               # Configuration
├── logs/                     # Log files
├── state/                    # State persistence
└── queue/                    # Message queue (if applicable)
```

**Customize with:**
```bash
./deploy.sh --target-dir /opt/battlestag orchestrator install
```

## Configuration

All components use interactive configuration during installation:

**Orchestrator:**
- Server port
- Authentication settings
- Program directories

**com-observer:**
- User identity (name, Slack ID, email)
- Slack bot token
- LLM provider (Ollama/Claude/GPT)
- Note writer (Obsidian/Notion)
- Vault/database paths

**Manual Configuration:**
```bash
# Edit after installation
nano ~/battlestag/orchestrator/config.yaml
nano ~/battlestag/programs/com-observer/config.yaml
```

## Automation

### Systemd (Linux)

Orchestrator installs as systemd service (optional):

```bash
sudo systemctl start battlestag-orchestrator
sudo systemctl enable battlestag-orchestrator
sudo systemctl status battlestag-orchestrator
```

### Cron (Programs)

Programs like com-observer install cron jobs:

```bash
# View cron jobs
crontab -l | grep com-observer

# Edit manually
crontab -e
```

### launchd (macOS)

For Mac servers, see program-specific documentation.

## Updates

### Update Workflow

1. **Prepare new packages:**
   ```bash
   cd DEPLOYMENT
   ./build-packages.sh
   ```

2. **Update remotely:**
   ```bash
   ./deploy.sh --remote user@server orchestrator update
   ./deploy.sh --remote user@server program update com-observer
   ```

3. **Or update locally:**
   ```bash
   ./deploy.sh orchestrator update
   ```

**Updates:**
- Preserve existing configurations
- Create automatic backups
- Restart services if running

## Troubleshooting

### Installation Fails

**Check prerequisites:**
```bash
# Go version
go version  # Should be 1.21+

# Disk space
df -h ~/battlestag

# Permissions
ls -ld ~/battlestag
```

### Remote Deployment Fails

**Check SSH access:**
```bash
ssh user@server "echo Connection OK"
```

**Check SSH key:**
```bash
ssh-add -l  # Should show your key
```

### Service Won't Start

**Check logs:**
```bash
# Systemd
sudo journalctl -u battlestag-orchestrator -f

# Programs
tail -f ~/battlestag/programs/com-observer/logs/*.log
```

**Check configuration:**
```bash
# Validate YAML
python3 -c "import yaml; yaml.safe_load(open('config.yaml'))"
```

### Package Not Found (Offline Mode)

**Rebuild packages:**
```bash
cd DEPLOYMENT
./build-packages.sh
```

**Verify packages:**
```bash
ls -lh packages/
```

## Security

### Best Practices

1. **Protect configuration files:**
   ```bash
   chmod 600 ~/battlestag/*/config.yaml
   ```

2. **Use SSH keys for remote deployment:**
   ```bash
   ssh-keygen -t ed25519
   ssh-copy-id user@server
   ```

3. **Restrict systemd service:**
   ```bash
   # Service runs as your user, not root
   # Limited to specific working directory
   ```

4. **Keep USB secure:**
   - Encrypt USB drive for sensitive deployments
   - Don't commit API tokens to packages
   - Use environment variables for secrets

### API Tokens

**Never include in packages:**
- Slack bot tokens
- LLM API keys
- Database credentials

**Configure during installation:**
- Interactive prompts
- Environment variables
- Secure config files

## Advanced Usage

### Custom Installation Directory

```bash
./deploy.sh --target-dir /opt/my-deployment orchestrator install
```

### Multiple Remote Servers

```bash
# Deploy to multiple servers
for server in server1 server2 server3; do
    ./deploy.sh --remote user@$server orchestrator install
done
```

### Automated Deployment Scripts

```bash
#!/bin/bash
# deploy-all.sh
./deploy.sh --remote prod@server1 orchestrator install
./deploy.sh --remote prod@server1 program install com-observer
./deploy.sh --remote prod@server1 program install lab-monitor
```

## Client Deployment Checklist

- [ ] Prepare USB drive with `./build-packages.sh`
- [ ] Test installation on staging server
- [ ] Document client-specific configuration
- [ ] Verify all API credentials
- [ ] Test automation (cron/systemd)
- [ ] Create backup of configuration
- [ ] Deploy to production
- [ ] Monitor logs for 24 hours
- [ ] Document for client handoff

## Support

For issues or questions:
1. Check logs: `~/battlestag/*/logs/`
2. Verify configuration: `cat ~/battlestag/*/config.yaml`
3. Review README files in each component
4. Check GitHub issues/documentation

## Version History

- **1.0.0** - Initial deployment system
  - Offline USB deployment
  - Remote SSH deployment
  - Orchestrator management
  - Program installation
  - Automated configuration
