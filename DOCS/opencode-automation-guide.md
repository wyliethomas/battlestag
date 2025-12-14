# OpenCode Automation Guide for Battlestag

## Overview
This guide documents how to use OpenCode for automated project creation and AI-driven development on the Battlestag server.

## Installation Summary

### Server (battlestag)
- **Location**: `~/.opencode/bin/opencode`
- **Config**: `~/.config/opencode/opencode.json`
- **Server Status**: Running on `http://battlestag:3000` (or `http://192.168.1.X:3000`)
- **Logs**: `/tmp/opencode-server.log`

### Client (laptop)
- **Location**: `~/.opencode/bin/opencode`
- **Config**: `~/.config/opencode/opencode.json`

### AI Models
- **Provider**: Anthropic at `anthropic/claude-sonnet-4-5`
- **Provider**: OpenAI at `openai/gpt-4o`
- **Provider**: Ollama at `http://192.168.1.232:11434`
- **Primary Model**: `ollama/llama3.1:8b-32k` (32k context for better tool calling)
- **Alt Models**:
  - `ollama/qwen2.5-coder:7b` (coding specialist)
  - `ollama/deepseek-r1:8b`
  - `ollama/llama3.1:8b` (standard)

## Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "ollama": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Ollama (battlestag)",
      "options": {
        "baseURL": "http://192.168.1.232:11434/v1"
      },
      "models": {
        "llama3.1:8b-32k": {
          "name": "Llama 3.1 8B (32k context)"
        },
        "qwen2.5-coder:7b": {
          "name": "Qwen 2.5 Coder 7B"
        }
      }
    }
  }
}
```

## Automation Approaches

### 1. CLI Automation with `opencode run`

**Best for**: Scripted automation, CI/CD pipelines, batch processing

```bash
# Basic usage
opencode run --model anthriopic/claude-sonnet-4-5 "Create a hello world Python script"

# With specific agent
opencode run --agent=build --model anthriopic/claude-sonnet-4-5 "Create a REST API with FastAPI"

# Continue from last session
opencode run --continue "Add error handling to the API"

# With specific prompt
opencode run -p "Fix all type errors" --model anthriopic/claude-sonnet-4-5
```

**Example automation script**:
```bash
#!/bin/bash
# Project generator script

PROJECT_NAME=$1
PROJECT_TYPE=$2

cd ~/projects
mkdir -p $PROJECT_NAME
cd $PROJECT_NAME

# Initialize project
opencode run --model ollama/qwen2.5-coder:7b --agent=build \
  "Create a new $PROJECT_TYPE project with:\n- proper directory structure\n- README.md\n- requirements.txt\n- basic tests\n- .gitignore"

# Add specific features
opencode run --continue \
  "Add logging and configuration management"

# Run tests
opencode run --continue \
  "Run all tests and fix any failures"
```

### 2. Server/Client Architecture

**Best for**: Interactive remote development, mobile/web clients

#### Start Server
```bash
# On battlestag server
cd /path/to/project
nohup opencode serve -p 3000 --hostname 0.0.0.0 --print-logs > /tmp/opencode-server.log 2>&1 &

# Check server status
curl http://battlestag:3000/config
```

#### Connect from Client
```bash
# From laptop
opencode attach http://battlestag:3000

# Or with specific directory
opencode attach http://battlestag:3000 --dir /path/to/project
```

#### Manage Server
```bash
# Check if running
ssh battlestag "pgrep -f 'opencode serve'"

# View logs
ssh battlestag "tail -f /tmp/opencode-server.log"

# Stop server
ssh battlestag "pkill -f 'opencode serve'"
```

### 3. REST API with Python SDK

**Best for**: Custom integrations, web services, advanced automation

#### Installation
```bash
pip install --pre opencode-ai
```

#### Basic Usage
```python
from opencode_ai import Opencode
import os

# Set API key (for cloud) or configure for local server
# os.environ['OPENCODE_API_KEY'] = 'your-key'

client = Opencode()

# List sessions
sessions = client.session.list()

# Create session and send message
# Note: SDK is primarily for cloud version
# For local server, you may need to use HTTP requests directly
```

#### Direct HTTP API (for local server)
```python
import requests

SERVER_URL = "http://battlestag:3000"

# Get configuration
config = requests.get(f"{SERVER_URL}/config").json()

# Create session
session_resp = requests.post(f"{SERVER_URL}/session", json={
    "projectId": "my-project",
    "directory": "/path/to/project"
})
session_id = session_resp.json()["id"]

# Send message
message_resp = requests.post(
    f"{SERVER_URL}/session/{session_id}/message",
    json={
        "content": "Create a Python script that...",
        "agent": "build",
        "model": "ollama/llama3.1:8b-32k"
    }
)

# Stream events (SSE)
import sseclient

events = sseclient.SSEClient(f"{SERVER_URL}/event")
for event in events:
    print(event.data)
```

### 4. Integration with Existing Tools

#### Git Workflow
```bash
# Create feature branch and implement
git checkout -b feature/new-api
opencode run --model ollama/qwen2.5-coder:7b \
  "Implement the user authentication API as specified in docs/api-spec.md"

# Review changes
git diff

# Commit if satisfied
git add .
git commit -m "Add user authentication API"
```

#### CI/CD Integration
```yaml
# .github/workflows/ai-review.yml
name: AI Code Review
on: [pull_request]
jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run OpenCode Review
        run: |
          curl -fsSL https://opencode.ai/install | bash
          opencode run --model ollama/qwen2.5-coder:7b \
            "Review this code for bugs, security issues, and best practices"
```

### 5. Project Generation Templates

Create reusable automation scripts:

```bash
# ~/scripts/generate-microservice.sh
#!/bin/bash
set -e

NAME=$1
PORT=$2

opencode run --model ollama/qwen2.5-coder:7b --agent=build << EOF
Create a new Go microservice called $NAME with:
1. Clean architecture (handler, service, repository layers)
2. REST API with chi router
3. PostgreSQL repository with migrations
4. Docker and docker-compose setup
5. Makefile with common tasks
6. Unit tests for each layer
7. README with setup instructions
8. Configuration via environment variables
9. Structured logging
10. Health check endpoint on port $PORT
EOF
```

## API Endpoints

### Server REST API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/config` | GET | Get server configuration |
| `/session` | POST | Create new session |
| `/session/:id` | GET | Get session details |
| `/session/:id/message` | POST | Send message to session |
| `/event` | GET | Server-Sent Events stream |
| `/models` | GET | List available models |

### Example API Calls

```bash
# Get configuration
curl http://battlestag:3000/config

# Create session
curl -X POST http://battlestag:3000/session \
  -H "Content-Type: application/json" \
  -d '{"directory": "/tmp/test"}'

# Send message
curl -X POST http://battlestag:3000/session/SESSION_ID/message \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Create a Python script",
    "agent": "build",
    "model": "ollama/llama3.1:8b-32k"
  }'
```

## Agents

### Build Agent
- **Mode**: Full access with file editing
- **Use for**: Creating, modifying, and building projects
- **Example**: `--agent=build`

### Plan Agent
- **Mode**: Read-only, exploration
- **Use for**: Code analysis, planning, understanding codebases
- **Example**: `--agent=plan`

## Tips for Automation

### 1. Use Appropriate Models
- **llama3.1:8b-32k**: General purpose, good for complex tasks
- **qwen2.5-coder:7b**: Best for coding tasks
- **deepseek-r1:8b**: Good for reasoning tasks

### 2. Session Management
```bash
# Continue previous session
opencode run --continue "Add more features"

# Use specific session
opencode run --session ses_abc123 "Continue work"

# Export session for analysis
opencode export ses_abc123 > session.json
```

### 3. Error Handling in Scripts
```bash
#!/bin/bash
set -e

if ! opencode run "$@"; then
    echo "OpenCode failed, check logs"
    ssh battlestag "tail -100 /tmp/opencode-server.log"
    exit 1
fi
```

### 4. Monitoring
```bash
# Watch server logs
ssh battlestag "tail -f /tmp/opencode-server.log"

# Check token usage
opencode stats
```

## Troubleshooting

### Server not accessible
```bash
# Check if server is running
ssh battlestag "pgrep -f 'opencode serve'"

# Check port
ssh battlestag "ss -tlnp | grep :3000"

# Restart server
ssh battlestag "pkill -f 'opencode serve'; cd /tmp && nohup ~/.opencode/bin/opencode serve -p 3000 --hostname 0.0.0.0 > /tmp/opencode-server.log 2>&1 &"
```

### Tool calling not working
- Ensure you're using the 32k context model: `ollama/llama3.1:8b-32k`
- Check Ollama server is accessible: `curl http://192.168.1.232:11434/api/tags`

### Model not found
```bash
# List available models
opencode models ollama

# Check Ollama models
curl http://192.168.1.232:11434/api/tags
```

## Next Steps

1. **Explore the API**: Read the docs to understand all available endpoints
2. **Create automation scripts**: Build scripts for common project types
3. **Test different models**: Compare performance of different models
4. **Build integrations**: Connect with your existing tools and workflows
5. **Monitor usage**: Track token usage and costs with `opencode stats`

## Resources

- OpenCode Docs: https://opencode.ai/docs
- GitHub: https://github.com/sst/opencode
- Python SDK: https://github.com/sst/opencode-sdk-python
- Go SDK: https://github.com/sst/opencode-sdk-go

---

*Last updated: 2025-12-03*
*Server: battlestag (192.168.1.X:3000)*
*Ollama: 192.168.1.232:11434*
