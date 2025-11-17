# Agent Gateway API

A centralized HTTP API gateway providing unified access to all agent databases in your agent swarm. Built with Go for performance and reliability.

## Features

- REST API for multiple agent databases (Stoic Thoughts, Tech Tips, and more)
- Simple API key authentication
- CORS enabled for web access
- Comprehensive logging (stdout and file)
- Health checks and statistics
- Pagination support
- Graceful shutdown
- Systemd service integration

## Quick Start

### One-Command Deploy (Recommended)

```bash
# Deploy as systemd service (requires sudo)
sudo ./deploy.sh
```

This will:
- Build the application
- Install to `/opt/agent-gateway`
- Set up configuration in `/etc/agent-gateway`
- Create systemd service
- Start the service

After deployment, edit the config:
```bash
sudo nano /etc/agent-gateway/config.yaml
sudo systemctl restart agent-gateway
```

### Manual Installation

If you prefer manual setup:

#### 1. Build the Application

```bash
# Install dependencies
go mod download

# Build the binary
go build -o agent-gateway .
```

#### 2. Configure

```bash
# Copy the example configuration
cp config.yaml.example config.yaml

# Edit with your settings
nano config.yaml
```

Update the following in `config.yaml`:
- Set database paths for your agents
- Change the API key to a secure random string
- Adjust server host/port if needed

#### 3. Run

```bash
# Run with config file
./agent-gateway --config config.yaml

# Or use environment variables
API_KEY=your-secret-key PORT=8080 ./agent-gateway
```

## API Documentation

All endpoints (except `/api/health`) require authentication via the `X-API-Key` header.

### Authentication

```bash
curl -H "X-API-Key: your-secret-key" http://localhost:8080/api/stoic/today
```

### Response Format

#### Success Response
```json
{
  "success": true,
  "data": {
    "id": 1,
    "date": "2025-11-16",
    "thought": "You have power over your mind - not outside events...",
    "source": "Marcus Aurelius",
    "created_at": "2025-11-16T08:23:15Z"
  }
}
```

#### Error Response
```json
{
  "success": false,
  "error": "No thought found for date 2025-11-15"
}
```

### Stoic Thoughts Endpoints

#### Get Today's Thought
```bash
GET /api/stoic/today

curl -H "X-API-Key: your-key" http://localhost:8080/api/stoic/today
```

#### Get Thought by Date
```bash
GET /api/stoic/date/:date

curl -H "X-API-Key: your-key" http://localhost:8080/api/stoic/date/2025-11-15
```

#### Get Random Thought
```bash
GET /api/stoic/random

curl -H "X-API-Key: your-key" http://localhost:8080/api/stoic/random
```

#### Get Latest N Thoughts
```bash
GET /api/stoic/latest/:n

curl -H "X-API-Key: your-key" http://localhost:8080/api/stoic/latest/5
```

#### Get All Thoughts (Paginated)
```bash
GET /api/stoic/all?page=1&page_size=20

curl -H "X-API-Key: your-key" "http://localhost:8080/api/stoic/all?page=1&page_size=20"
```

### Tech Tips Endpoints

#### Get Today's Tip
```bash
GET /api/tech/today

curl -H "X-API-Key: your-key" http://localhost:8080/api/tech/today
```

#### Get Tip by Date
```bash
GET /api/tech/date/:date

curl -H "X-API-Key: your-key" http://localhost:8080/api/tech/date/2025-11-15
```

#### Get Random Tip
```bash
GET /api/tech/random

curl -H "X-API-Key: your-key" http://localhost:8080/api/tech/random
```

#### Get Latest N Tips
```bash
GET /api/tech/latest/:n

curl -H "X-API-Key: your-key" http://localhost:8080/api/tech/latest/10
```

#### Get Tips by Category
```bash
GET /api/tech/category/:category

curl -H "X-API-Key: your-key" http://localhost:8080/api/tech/category/git
curl -H "X-API-Key: your-key" http://localhost:8080/api/tech/category/docker
```

#### Get All Tips (Paginated)
```bash
GET /api/tech/all?page=1&page_size=20

curl -H "X-API-Key: your-key" "http://localhost:8080/api/tech/all?page=1&page_size=20"
```

### Meta Endpoints

#### Health Check (No Auth Required)
```bash
GET /api/health

curl http://localhost:8080/api/health
```

Response:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2025-11-16T10:30:00Z",
    "agents": {
      "stoic": {
        "available": true,
        "db_path": "/path/to/stoic_thoughts.db"
      },
      "tech": {
        "available": true,
        "db_path": "/path/to/tech_tips.db"
      }
    }
  }
}
```

#### System Statistics
```bash
GET /api/stats

curl -H "X-API-Key: your-key" http://localhost:8080/api/stats
```

Response:
```json
{
  "success": true,
  "data": {
    "total_thoughts": 42,
    "total_tips": 156,
    "last_thought_date": "2025-11-16",
    "last_tip_date": "2025-11-16",
    "uptime": "2h15m30s",
    "timestamp": "2025-11-16T10:30:00Z"
  }
}
```

## Configuration

### Configuration File (config.yaml)

```yaml
agents:
  stoic:
    db_path: "/path/to/stoic_thoughts.db"
  tech:
    db_path: "/path/to/tech_tips.db"

server:
  port: "8080"
  host: "0.0.0.0"

auth:
  api_key: "your-secret-key"

logging:
  access_log: "logs/access.log"
  error_log: "logs/error.log"
```

### Environment Variables

You can override configuration with environment variables:

- `PORT` - Server port (default: 8080)
- `HOST` - Server host (default: 0.0.0.0)
- `API_KEY` - API authentication key (required)
- `STOIC_DB_PATH` - Path to stoic thoughts database
- `TECH_DB_PATH` - Path to tech tips database

Example:
```bash
API_KEY=secret123 PORT=9000 ./agent-gateway
```

## Installation as System Service

### Automated Deployment (Recommended)

Use the included deployment script:

```bash
# Deploy everything automatically
sudo ./deploy.sh

# After deployment, edit config and restart
sudo nano /etc/agent-gateway/config.yaml
sudo systemctl restart agent-gateway
```

Other deployment commands:
```bash
sudo ./deploy.sh status      # Check service status
sudo ./deploy.sh logs        # Follow logs
sudo ./deploy.sh restart     # Restart service
sudo ./deploy.sh uninstall   # Remove service
sudo ./deploy.sh help        # Show all commands
```

### Manual Installation

If you prefer to install manually:

#### 1. Build the Binary
```bash
go build -o agent-gateway .
```

#### 2. Update Service File Paths
Edit `agent-gateway.service` and update paths:
- `WorkingDirectory`
- `ExecStart`
- User/Group

#### 3. Install and Enable Service
```bash
# Copy service file
sudo cp agent-gateway.service /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable agent-gateway

# Start the service
sudo systemctl start agent-gateway

# Check status
sudo systemctl status agent-gateway
```

#### 4. View Logs
```bash
# Follow service logs
sudo journalctl -u agent-gateway -f

# View last 50 lines
sudo journalctl -u agent-gateway -n 50
```

## Remote Access

Once deployed on your MacBook, you can access from anywhere on your network:

```bash
# From another machine
curl -H "X-API-Key: your-key" http://192.168.1.100:8080/api/stoic/today

# SSH into MacBook and query locally
ssh battlestag@macbook-ip
curl -H "X-API-Key: your-key" http://localhost:8080/api/tech/random
```

## CORS Support

CORS is enabled by default for all origins, allowing web applications to access the API:

```javascript
// Example JavaScript fetch
fetch('http://192.168.1.100:8080/api/stoic/today', {
  headers: {
    'X-API-Key': 'your-secret-key'
  }
})
  .then(response => response.json())
  .then(data => console.log(data));
```

## Development

### Run in Development Mode
```bash
# Run with live reload (requires air)
air

# Or run directly
go run main.go --config config.yaml
```

### Run Tests
```bash
go test ./...
```

### Build for Production
```bash
# Build with optimizations
go build -ldflags="-s -w" -o agent-gateway .

# Or use make
make build
```

## Project Structure

```
agent-gateway/
├── main.go                 # Server setup, routing, logging
├── config/
│   └── config.go          # Configuration loading
├── db/
│   └── agents.go          # Database connections for agents
├── handlers/
│   ├── stoic.go           # Stoic thought handlers
│   ├── tech.go            # Tech tip handlers
│   └── meta.go            # Health, stats handlers
├── middleware/
│   ├── auth.go            # API key authentication
│   ├── cors.go            # CORS support
│   └── logging.go         # Request logging
├── models/
│   └── response.go        # Response models
├── config.yaml            # Configuration file
├── config.yaml.example    # Example configuration
├── agent-gateway.service  # Systemd service file
└── README.md
```

## Adding New Agents

To add a new agent to the gateway:

1. Update `config/config.go` to add the new agent config:
```go
type AgentsConfig struct {
    Stoic    StoicConfig    `yaml:"stoic"`
    Tech     TechConfig     `yaml:"tech"`
    NewAgent NewAgentConfig `yaml:"new_agent"`  // Add this
}
```

2. Update `db/agents.go` to add database connection and methods
3. Create handler in `handlers/new_agent.go`
4. Register routes in `main.go`
5. Update `config.yaml` with new agent's database path

## Troubleshooting

### Service Won't Start
```bash
# Check service status
sudo systemctl status agent-gateway

# View detailed logs
sudo journalctl -u agent-gateway -n 100

# Check if port is already in use
sudo lsof -i :8080
```

### Database Connection Errors
- Verify database paths in config.yaml
- Check file permissions (databases must be readable)
- Ensure databases exist and have correct schema

### Authentication Failures
- Verify API key in config matches request header
- Check for typos in `X-API-Key` header name
- Ensure API key doesn't have leading/trailing spaces

## Security Notes

- Always use a strong, random API key
- Consider using HTTPS in production (reverse proxy with nginx/caddy)
- Restrict host to `127.0.0.1` if only local access is needed
- Keep databases read-only for the gateway user
- Regularly rotate API keys
- Monitor access logs for suspicious activity

## Performance

- Default timeouts: Read 15s, Write 15s, Idle 60s
- Pagination limits: Max 100 items per page
- Database connection pooling enabled
- Concurrent request handling via Go's net/http

## License

MIT License - See LICENSE file for details

## Support

For issues, questions, or contributions, please open an issue on GitHub.
