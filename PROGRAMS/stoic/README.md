# Stoic Thought Generator

A simple Go application that generates daily stoic thoughts by querying a local Ollama LLM server.

## Features

- Generates one unique stoic thought per day
- Stores thoughts in SQLite database with timestamps
- Retrieves existing thoughts instead of regenerating
- Configurable via environment variables
- Retry logic for LLM requests
- Clean CLI interface with helpful flags

## Prerequisites

1. **Go** (version 1.16 or higher)
2. **Ollama** running locally with the dolphin3 model

### Installing Ollama

If you don't have Ollama installed:

```bash
# Linux
curl -fsSL https://ollama.com/install.sh | sh

# macOS
brew install ollama

# Start Ollama service
ollama serve

# Pull the dolphin3 model (or dolphin-mistral)
ollama pull dolphin3
```

## Installation

1. Clone or navigate to the project directory:

```bash
cd stoic-thought
```

2. Initialize Go module and install dependencies:

```bash
go mod init stoic-thought
go get github.com/mattn/go-sqlite3
go get github.com/joho/godotenv
go mod tidy
```

3. Build the application:

```bash
go build -o stoic-thought
```

## Usage

### Basic Usage

Generate or display today's thought:

```bash
./stoic-thought
```

Example output:
```
📜 Stoic Thought for 2025-11-16:

"You have power over your mind, not outside events.
Realize this, and you will find strength."

(Generated at 2025-11-16 08:23:15)
```

### CLI Flags

**Show thought from a specific date:**

```bash
./stoic-thought --date 2025-11-15
```

**Force regenerate today's thought:**

```bash
./stoic-thought --regenerate
```

Note: Cannot use `--regenerate` with `--date` flag.

## Configuration

Configure the application using environment variables. You can set these in two ways:

### Option 1: Using a .env File (Recommended for Local Development)

Create a `.env` file in the project directory:

```bash
# Copy the example file
cp .env.example .env

# Edit .env with your settings
OLLAMA_HOST=http://192.168.1.100:11434
OLLAMA_MODEL=dolphin3
DB_PATH=./stoic_thoughts.db
```

The application will automatically load the `.env` file if it exists.

### Option 2: Using Shell Environment Variables

Set environment variables directly:

```bash
export OLLAMA_HOST="http://localhost:11434"
export OLLAMA_MODEL="dolphin-mistral"
export DB_PATH="/home/user/.stoic/thoughts.db"

./stoic-thought
```

**Note:** Shell environment variables take precedence over `.env` file values.

### Configuration Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama API endpoint (use remote IP for network LLM servers) |
| `OLLAMA_MODEL` | `dolphin3` | LLM model to use |
| `DB_PATH` | `./stoic_thoughts.db` | SQLite database path |

## Project Structure

```
stoic-thought/
├── main.go              # Entry point, CLI handling
├── llm/
│   ├── client.go        # Ollama HTTP client
│   └── client_test.go   # LLM client tests
├── db/
│   ├── sqlite.go        # Database operations
│   └── sqlite_test.go   # Database tests
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
└── README.md            # This file
```

## Running Tests

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Test specific package:

```bash
go test ./db
go test ./llm
```

## Database Schema

The application uses SQLite with the following schema:

```sql
CREATE TABLE IF NOT EXISTS thoughts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT UNIQUE NOT NULL,
    thought TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Error Handling

The application handles common errors gracefully:

- **Ollama server unreachable**: Health check before generation with automatic retry (3 attempts with 20-second delays for server wake-up)
- **LLM request failures**: Automatic retry (3 attempts with 20-second delay)
- **Database errors**: Detailed error messages for debugging
- **Invalid date format**: Validation with helpful error message

## Automating Daily Thoughts

### Using systemd (Linux)

Create a systemd service and timer to run the application daily.

**Service file** (`~/.config/systemd/user/stoic-thought.service`):

```ini
[Unit]
Description=Daily Stoic Thought Generator
After=network.target

[Service]
Type=oneshot
ExecStart=/path/to/stoic-thought
Environment="OLLAMA_HOST=http://localhost:11434"
Environment="OLLAMA_MODEL=dolphin3"
Environment="DB_PATH=/home/user/.stoic/thoughts.db"

[Install]
WantedBy=default.target
```

**Timer file** (`~/.config/systemd/user/stoic-thought.timer`):

```ini
[Unit]
Description=Run Stoic Thought Generator Daily
Requires=stoic-thought.service

[Timer]
OnCalendar=daily
OnCalendar=08:00
Persistent=true

[Install]
WantedBy=timers.target
```

Enable and start the timer:

```bash
systemctl --user daemon-reload
systemctl --user enable stoic-thought.timer
systemctl --user start stoic-thought.timer

# Check timer status
systemctl --user list-timers
```

### Using cron

Add to crontab:

```bash
crontab -e

# Add this line to run daily at 8 AM
0 8 * * * /path/to/stoic-thought >> /path/to/stoic.log 2>&1
```

## Deploying to a Remote Server

The project includes deployment scripts to easily deploy the application to a remote Linux server.

### Prerequisites

- SSH access to the remote server with key-based authentication
- The remote server must have Ollama running with the dolphin3 model
- sudo privileges on the remote server (for installing binary to `/usr/local/bin`)

### Quick Deployment

1. Build the application locally:

```bash
go build -o stoic-thought
```

2. Deploy to remote server:

```bash
./deploy.sh user@remote-host
```

Or specify a custom Ollama host:

```bash
./deploy.sh user@192.168.1.100 http://localhost:11434
```

### What the Deployment Does

The `deploy.sh` script will:

1. Test SSH connectivity to the remote server
2. Copy the binary, systemd files, and installation script
3. Execute the remote installation which:
   - Installs the binary to `/usr/local/bin/stoic-thought`
   - Creates data directory at `/var/lib/stoic-thought`
   - Installs systemd service and timer files
   - Tests Ollama connectivity
   - Optionally enables and starts the timer

### Manual Deployment

If you prefer to deploy manually:

1. Copy files to the remote server:

```bash
scp stoic-thought user@remote-host:/tmp/
scp stoic-thought.service user@remote-host:/tmp/
scp stoic-thought.timer user@remote-host:/tmp/
scp remote-install.sh user@remote-host:/tmp/
```

2. SSH into the remote server and run the installation:

```bash
ssh user@remote-host
cd /tmp
chmod +x remote-install.sh
./remote-install.sh
```

### Managing the Remote Service

After deployment, manage the service on the remote server:

```bash
# Check timer status
ssh user@remote-host 'systemctl --user status stoic-thought.timer'

# View scheduled timers
ssh user@remote-host 'systemctl --user list-timers'

# View service logs
ssh user@remote-host 'journalctl --user -u stoic-thought.service'

# Run manually
ssh user@remote-host '/usr/local/bin/stoic-thought'

# Stop/disable the timer
ssh user@remote-host 'systemctl --user stop stoic-thought.timer'
ssh user@remote-host 'systemctl --user disable stoic-thought.timer'
```

### Customizing Deployment

The deployment script will prompt you for the Ollama host address during deployment, which is then configured in the systemd service file. You can also manually customize the configuration:

**Option 1: During Deployment**
```bash
# The script will prompt for the Ollama host
./deploy.sh user@remote-host

# Or pass it directly
./deploy.sh user@remote-host http://192.168.1.100:11434
```

**Option 2: Edit Service File Before Deploying**
```bash
# Edit the Environment variables in stoic-thought.service
Environment="OLLAMA_HOST=http://192.168.1.100:11434"
Environment="OLLAMA_MODEL=dolphin-mistral"
Environment="DB_PATH=/custom/path/thoughts.db"
```

**Note:** On remote servers, the systemd service file environment variables are used. The `.env` file is primarily for local development.

### Deployment Files

- `deploy.sh` - Main deployment script (run locally)
- `remote-install.sh` - Installation script (runs on remote server)
- `stoic-thought.service` - systemd service definition
- `stoic-thought.timer` - systemd timer definition

## Troubleshooting

**"Ollama server unreachable"**
- Ensure Ollama is running: `ollama serve`
- Check the host configuration matches your Ollama instance
- Verify the model is pulled: `ollama list`

**"failed after 3 attempts"**
- Check Ollama logs for errors
- Ensure the model name is correct
- Try a different model if dolphin3 is not available

**"Invalid date format"**
- Use YYYY-MM-DD format for the `--date` flag
- Example: `2025-11-16`

**Database locked errors**
- Ensure no other instance is running
- Check file permissions on the database file

## Development

To modify the LLM prompt, edit `llm/client.go`:

```go
prompt := "Generate a single stoic thought for the day. Keep it under 3 sentences. Focus on practical wisdom from Marcus Aurelius, Epictetus, or Seneca. Be direct and actionable."
```

To adjust retry behavior, modify constants in `llm/client.go`:

```go
const (
    defaultMaxRetries = 3
    defaultTimeout    = 30 * time.Second
    retryDelay        = 20 * time.Second  // Wait for Ollama server to wake up
)
```

## License

This is a proof-of-concept application. Use and modify as needed.

## Contributing

This is a template for single-purpose agents in the agent swarm pattern. Feel free to adapt for your own use cases.
