# Agent Gateway v2

A manifest-driven REST API gateway for orchestrating LLM function calling with local programs.

## Features

- 🔌 **Auto-Discovery**: Programs automatically discovered from `programs/` directory
- 📄 **Manifest-Driven**: Each program self-describes via `manifest.json`
- 🎯 **Zero-Config Addition**: Add new programs by just dropping a folder with manifest
- 🤖 **Native Ollama Function Calling**: Uses Ollama's built-in tool support
- 🛡️ **Graceful Fallback**: Automatically handles hallucinated tool calls
- ⚙️ **Flexible Configuration**: YAML config file with command-line overrides

## Quick Start

### Installation

```bash
./install.sh
```

This will:
- Build the gateway and all programs
- Install binary to `~/Work/BATTLESTAG-BOT/bin`
- Create config at `~/.config/agent-gateway/config.yaml`

### Configuration

Edit `~/.config/agent-gateway/config.yaml`:

```yaml
ollama:
  url: http://192.168.1.232:11434
  model: llama3.1:8b

server:
  port: 8080

execution:
  timeout: 30s

programs:
  directory: ./programs
```

### Running

```bash
# With config file
agent-gateway --config ~/.config/agent-gateway/config.yaml

# With command-line overrides
agent-gateway --ollama-url http://your-ollama:11434 --ollama-model llama3.1:8b

# See all options
agent-gateway --help
```

## Management Scripts

### Install
```bash
./install.sh
```
Builds and installs the gateway and all programs.

### Update
```bash
./update.sh
```
Rebuilds gateway and programs. Optionally stops/restarts the service.

### Uninstall
```bash
./uninstall.sh
```
Removes installed binaries. Optionally removes config.

## API Endpoints

### POST /chat
Send chat messages with automatic tool execution.

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "What time is it?"}
    ]
  }'
```

### GET /health
Check gateway status.

```bash
curl http://localhost:8080/health
```

### GET /programs
List all available programs.

```bash
curl http://localhost:8080/programs
```

## Adding New Programs

1. Create a program directory: `programs/my-program/`
2. Add `manifest.json`:
   ```json
   {
     "name": "my-program",
     "description": "What this program does",
     "parameters": {
       "type": "object",
       "properties": {
         "param1": {
           "type": "string",
           "description": "First parameter"
         }
       },
       "required": ["param1"]
     }
   }
   ```
3. Create executable that reads JSON from stdin
4. Gateway auto-discovers it on next start!

## Architecture

```
agent-gateway-v2/
├── main.go              # Entry point
├── config.go            # Configuration loader
├── gateway/             # Core orchestration
│   ├── types.go         # Data structures
│   ├── registry.go      # Program discovery
│   ├── executor.go      # Program execution
│   └── gateway.go       # Orchestration logic
├── llm/                 # LLM integration
│   ├── types.go         # Ollama types
│   └── ollama.go        # Ollama client
├── api/                 # HTTP layer
│   └── handlers.go      # REST endpoints
└── programs/            # Executable programs
    ├── echo/
    ├── time/
    ├── lab-monitor/
    └── task-manager/
```

## Available Programs

- **echo**: Simple echo test program
- **time**: Returns current date and time
- **lab-monitor**: Home lab server monitoring (13 commands)
- **task-manager**: Project management system (18 commands)

## Development

### Build
```bash
go build -o agent-gateway main.go config.go
```

### Run Tests
```bash
go test ./...
```

### Add Debug Logging
The gateway includes debug output for troubleshooting tool execution.

## Requirements

- Go 1.21+
- Ollama server with function calling support
- Recommended models: llama3.1:8b, llama3.2:latest

## License

Part of the BATTLESTAG-BOT project.
