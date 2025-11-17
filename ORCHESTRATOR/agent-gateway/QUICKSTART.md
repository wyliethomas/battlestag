# Quick Start Guide

Get the Agent Gateway API up and running in 5 minutes.

## Prerequisites

- Go 1.19 or higher installed
- SQLite databases from your stoic and tech agents

## The Easy Way: One Command Deploy

```bash
cd /home/battlestag/Work/WYBOT/ORCHESTRATOR/agent-gateway

# Deploy as systemd service (does everything for you!)
sudo ./deploy.sh
```

This automatically:
- Builds the application
- Installs to `/opt/agent-gateway`
- Creates configuration at `/etc/agent-gateway/config.yaml`
- Sets up systemd service
- Starts the service

**After deployment:**
```bash
# Edit configuration
sudo nano /etc/agent-gateway/config.yaml

# Update database paths and API key, then:
sudo systemctl restart agent-gateway

# Check status
sudo systemctl status agent-gateway
```

**Jump to Step 4 to test!**

---

## The Manual Way: Step by Step

If you prefer manual setup or development:

### Step 1: Clone and Build

```bash
cd /home/battlestag/Work/WYBOT/ORCHESTRATOR/agent-gateway

# Download dependencies
go mod download

# Build the application
make build
```

### Step 2: Configure

```bash
# Copy the example configuration
cp config.yaml.example config.yaml

# Edit the configuration
nano config.yaml
```

Update these values in `config.yaml`:

```yaml
agents:
  stoic:
    db_path: "/home/battlestag/Work/WYBOT/stoic-agent/stoic_thoughts.db"
  tech:
    db_path: "/home/battlestag/Work/WYBOT/tech-agent/tech_tips.db"

auth:
  api_key: "change-this-to-a-secure-random-key"
```

**Important:** Generate a secure API key:
```bash
# Generate a random API key
openssl rand -hex 32
# Or use:
uuidgen
```

### Step 3: Run

```bash
# Start the server
./agent-gateway --config config.yaml
```

You should see:
```
Starting Agent Gateway v1.0.0
Configuration loaded from: config.yaml
Database connections initialized
Server listening on 0.0.0.0:8080
Press Ctrl+C to stop
```

## Step 4: Test

In another terminal:

```bash
# Set your API key
export API_KEY="your-api-key-from-config"

# Run the test suite
./test-api.sh

# Or test manually
curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/stoic/today
curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/tech/random
curl http://localhost:8080/api/health
```

## Step 5: Install as System Service (Optional)

**Easy way:**
```bash
# One command to deploy everything
sudo ./deploy.sh
```

**Manual way:**
```bash
# Update paths in agent-gateway.service if needed
nano agent-gateway.service

# Install the service
sudo cp agent-gateway.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable agent-gateway
sudo systemctl start agent-gateway

# Check status
sudo systemctl status agent-gateway
```

## Quick Commands

### Deployment Commands
```bash
# Deploy as service
sudo ./deploy.sh

# Check service status
sudo ./deploy.sh status

# Follow logs
sudo ./deploy.sh logs

# Restart service
sudo ./deploy.sh restart

# Uninstall service
sudo ./deploy.sh uninstall

# Build only (no install)
sudo ./deploy.sh build-only
```

### Development Commands
```bash
# Build
make build

# Run in development
make run

# Run tests
make test

# Clean build artifacts
make clean

# Install to /usr/local/bin
make install
```

### Service Management (if deployed)
```bash
# View logs
sudo journalctl -u agent-gateway -f

# Check status
sudo systemctl status agent-gateway

# Restart
sudo systemctl restart agent-gateway
```

## Common Issues

### "Failed to open database"
- Check that the database paths in config.yaml are correct
- Ensure the database files exist
- Verify file permissions (databases must be readable)

### "Port already in use"
- Change the port in config.yaml or use environment variable:
  ```bash
  PORT=9000 ./agent-gateway
  ```

### "Invalid API key"
- Ensure the X-API-Key header matches the key in config.yaml
- Check for leading/trailing spaces in the API key

## Next Steps

- Read the full [README.md](README.md) for complete API documentation
- Set up remote access from other machines on your network
- Create a web dashboard to visualize your agent data
- Add more agents to the gateway

## Remote Access Example

Once running on your MacBook, access from any device on your network:

```bash
# From desktop/laptop
curl -H "X-API-Key: your-key" http://192.168.1.100:8080/api/stoic/today

# From phone (using Termux or similar)
curl -H "X-API-Key: your-key" http://macbook-ip:8080/api/tech/random
```

## Environment Variable Configuration

Instead of config.yaml, you can use environment variables:

```bash
API_KEY="secret123" \
STOIC_DB_PATH="/path/to/stoic.db" \
TECH_DB_PATH="/path/to/tech.db" \
PORT="8080" \
./agent-gateway
```

This is useful for containerized deployments or CI/CD pipelines.
