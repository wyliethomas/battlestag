# Tech Tip Generator

A simple Go application that generates daily tech tips by querying a local LLM server (Ollama). Part of the WYBOT agent swarm pattern.

## Features

- Generate one practical tech tip per day using a local Ollama server
- Automatic categorization (git, docker, shell, vim, ssh, python, go, linux, cli, etc.)
- SQLite database for persistent storage
- Show tips from previous dates
- Get a random tip from history
- Force regenerate today's tip
- Zero external dependencies except SQLite driver
- Clean separation between LLM client, database, and main logic

## Prerequisites

- Go 1.21 or later
- [Ollama](https://ollama.ai/) running locally
- Dolphin3 model (or any other Ollama model)

### Install Ollama and Model

```bash
# Install Ollama (if not already installed)
curl -fsSL https://ollama.ai/install.sh | sh

# Pull the Dolphin3 model
ollama pull dolphin3

# Verify Ollama is running
ollama list
```

## Installation

### Build from Source

```bash
# Clone/navigate to the tech-tip directory
cd tech-tip

# Install dependencies
go get github.com/mattn/go-sqlite3
go get github.com/joho/godotenv
go mod tidy

# Build the binary
go build -o tech-tip

# Optionally, install to your PATH
sudo cp tech-tip /usr/local/bin/
```

### Run Tests

```bash
# Run unit tests
go test -short ./...

# Run all tests including integration tests (requires Ollama running)
go test ./...
```

## Usage

### Basic Usage

```bash
# Generate/show today's tip
./tech-tip

# Force regenerate today's tip
./tech-tip --regenerate

# Show tip from a specific date
./tech-tip --date 2025-11-15

# Show a random tip from history
./tech-tip --random
```

### Example Output

```
💡 Tech Tip for 2025-11-16:

Use `git log --oneline --graph --all --decorate` to visualize your Git branch
history in the terminal. The --graph flag shows ASCII art tree, --all includes
all branches, and --decorate displays branch names and tags.

Category: git
(Generated at 2025-11-16 08:23:15)
```

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
DB_PATH=./tech_tips.db
```

The application will automatically load the `.env` file if it exists.

### Option 2: Using Shell Environment Variables

Set environment variables directly:

```bash
# Ollama server URL (default: http://localhost:11434)
export OLLAMA_HOST=http://localhost:11434

# Ollama model to use (default: dolphin3)
export OLLAMA_MODEL=dolphin3

# Database file path (default: ./tech_tips.db)
export DB_PATH=/var/lib/tech-tip/tech_tips.db
```

**Note:** Shell environment variables take precedence over `.env` file values.

### Configuration Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama API endpoint (use remote IP for network LLM servers) |
| `OLLAMA_MODEL` | `dolphin3` | LLM model to use |
| `DB_PATH` | `./tech_tips.db` | SQLite database path |

Example with custom configuration:

```bash
OLLAMA_HOST=http://192.168.1.100:11434 OLLAMA_MODEL=mistral ./tech-tip
```

## Systemd Integration (Optional)

Run the tech tip generator automatically every day using systemd user services.

### Setup

```bash
# Install systemd service files
./install-systemd.sh

# Enable and start the timer
systemctl --user enable tech-tip.timer
systemctl --user start tech-tip.timer

# Check timer status
systemctl --user list-timers tech-tip.timer

# View logs
journalctl --user -u tech-tip.service -f
```

The timer runs daily at 8:00 AM with a randomized delay of up to 5 minutes.

### Manual Management

```bash
# Run the service manually
systemctl --user start tech-tip.service

# Check service status
systemctl --user status tech-tip.service

# Disable the timer
systemctl --user disable tech-tip.timer
systemctl --user stop tech-tip.timer
```

## Remote Deployment

Deploy to a remote server using the included deployment script:

```bash
# Build the binary first
go build -o tech-tip

# Deploy to remote server
./deploy.sh user@remote-host

# Or specify custom Ollama host
./deploy.sh user@remote-host http://192.168.1.100:11434
```

The deployment script will:
1. Test SSH connection
2. Copy the binary and systemd files
3. Install to the specified directories
4. Set up systemd user service with your Ollama host configuration
5. Enable and start the timer

**Note:** On remote servers, the systemd service file environment variables are used. The `.env` file is primarily for local development.

## Project Structure

```
tech-tip/
├── main.go              # Entry point, CLI handling
├── llm/
│   ├── client.go        # Ollama HTTP client
│   └── client_test.go   # LLM client tests
├── db/
│   ├── sqlite.go        # Database operations
│   └── sqlite_test.go   # Database tests
├── go.mod               # Go module definition
├── go.sum               # Go dependencies
├── README.md            # This file
├── tech-tip             # Compiled binary
├── tech-tip.service     # Systemd service file
├── tech-tip.timer       # Systemd timer file
├── deploy.sh            # Deployment script
├── remote-install.sh    # Remote installation script
└── install-systemd.sh   # Local systemd installation script
```

## Database Schema

```sql
CREATE TABLE IF NOT EXISTS tips (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT UNIQUE NOT NULL,
    tip TEXT NOT NULL,
    category TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Error Handling

The application handles common errors gracefully:

- **Ollama server unreachable**: Health check before generation with automatic retry (3 attempts with 20-second delays for server wake-up)
- **Retry logic**: 3 attempts with 20-second delay between retries for both health checks and generation
- **Date validation**: Validates YYYY-MM-DD format for date inputs
- **Missing tips**: Clear error messages when querying non-existent dates
- **Database errors**: Detailed error messages for database operations

## Development

### Adding New Categories

Edit the `extractCategory` function in `llm/client.go`:

```go
categories := map[string][]string{
    "yourcategory": {`pattern1`, `pattern2`, `pattern3`},
}
```

### Customizing the Prompt

Edit the `GenerateTechTip` function in `llm/client.go` to customize what kind of tips are generated.

## Troubleshooting

### "Ollama server health check failed"

Make sure Ollama is running:

```bash
# Check if Ollama is running
ps aux | grep ollama

# Start Ollama if needed
ollama serve
```

### "failed to open database: unable to open database file"

Ensure the database directory exists and is writable:

```bash
# Create directory if using custom DB_PATH
mkdir -p /var/lib/tech-tip
chmod 755 /var/lib/tech-tip
```

### Integration tests failing

Integration tests require Ollama to be running with the specified model:

```bash
# Make sure Ollama is running
ollama serve

# Pull the required model
ollama pull dolphin3

# Run tests
go test ./...
```

## License

Part of the WYBOT project. See main project repository for license information.

## See Also

- [stoic-thought](../stoic/) - Daily stoic thought generator (sister application)
- [Ollama Documentation](https://github.com/ollama/ollama/blob/main/docs/api.md)
