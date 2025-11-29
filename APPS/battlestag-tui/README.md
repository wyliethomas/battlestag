# Battlestag TUI

A Claude Code-style terminal user interface for managing personal finances with AI assistance.


## Features

- **AI-Powered Chat**: Conversational interface powered by local LLM (via Ollama)
- **Always-Visible Prompt**: Input line at the bottom for unified interaction
- **Command System**: Type "/" to access commands (command palette coming soon)
- **Financial Management**: Integration with Agent Gateway for assets, liabilities, and financial data
- **Global Command**: Accessible from any terminal after installation

## Architecture

The TUI follows a Claude Code-inspired design:

- **Chat Mode**: Default view showing AI conversation history
- **Prompt Line**: Always visible at bottom for input
- **No Tabs**: Command-driven navigation instead of numbered tabs
- **Unified Input**:
  - Regular text → sent to AI chat
  - Text starting with "/" → executed as command

## Installation

### Quick Install

```bash
./install.sh
```

This will:
1. Build the `battlestag` binary
2. Install to `~/.local/bin/battlestag`
3. Create config directory at `~/.config/battlestag-tui/`
4. Check if `~/.local/bin` is in your PATH

### Manual Installation

```bash
# Build
make build

# Install
make install

# Or build manually
go build -o battlestag .
```

### Uninstallation

```bash
./uninstall.sh
```

Or manually:
```bash
make uninstall
```

## Configuration

On first run, you'll be prompted to configure:
- **Agent Gateway URL**: The API endpoint (default: `http://192.168.1.140:8080`)
- **API Key**: Authentication key for the gateway

Configuration is saved to: `~/.config/battlestag-tui/config.yaml`

## Usage

### Starting the Application

```bash
battlestag
```

### Chat Interface

The chat interface is the default view. Simply type your message and press Enter:

```
> What's my current net worth?
> Show me my assets
> How do I upload a bank statement?
```

The AI will respond and may suggest relevant commands.

### Commands

Type "/" followed by a command:

```
/dashboard     - View financial overview
/assets        - List all assets
/liabilities   - List all liabilities
/upload        - Upload bank statement
/settings      - Configure application
```

**Note**: Command palette is coming soon. For now, type the full command after "/"

### Keyboard Shortcuts

- **Enter** - Submit message/command
- **/** - Start typing a command
- **Ctrl+L** - Clear chat history
- **Ctrl+C** or **Ctrl+D** - Quit application
- **Esc** - Clear current input

## Development

### Project Structure

```
battlestag-tui/
├── client/          # API client for Agent Gateway
│   ├── api.go       # REST API methods
│   └── llm.go       # LLM chat client
├── config/          # Configuration management
│   └── config.go
├── models/          # Data models
│   └── models.go
├── ui/              # User interface components
│   ├── app.go       # Main application logic
│   ├── chat.go      # Chat conversation view
│   ├── prompt.go    # Input prompt component
│   ├── onboarding.go # First-run setup
│   └── styles.go    # Shared styles
├── main.go          # Entry point
├── Makefile         # Build automation
├── install.sh       # Installation script
└── uninstall.sh     # Uninstallation script
```

### Building

```bash
# Build binary
make build

# Format code
make fmt

# Run tests
make test

# Vet code
make vet

# Run without installing
make run
```

### Module Structure

The project uses the simplified module name `battlestag`:

```go
module battlestag

import (
    "battlestag/client"
    "battlestag/config"
    "battlestag/models"
    "battlestag/ui"
)
```

### Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## Requirements

- Go 1.21 or higher (upgraded to 1.23 with latest dependencies)
- Agent Gateway running and accessible
- Ollama server with llama3.2:8b model (for AI chat)

## Troubleshooting

### Binary not found after installation

Make sure `~/.local/bin` is in your PATH:

```bash
# For bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# For zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### Can't connect to Agent Gateway

1. Check if the gateway is running
2. Verify the URL in settings during onboarding
3. Test connection: `curl http://192.168.1.140:8080/api/health`

### LLM not responding

1. Check if Ollama server is running on 192.168.1.232
2. Verify LLM endpoint in Agent Gateway config
3. Test Ollama: `curl http://192.168.1.232:11434/api/tags`

### Build Errors

```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

## Roadmap

- [x] Phase 1: LLM Chat Integration
- [x] Phase 2: Global Installation
- [x] Phase 3: Claude Code-style Interface
- [ ] Phase 4: Command Palette with Fuzzy Search
- [ ] Phase 5: Command Execution (dashboard, assets, etc.)
- [ ] Phase 6: Tab Autocomplete
- [ ] Phase 7: Enhanced UI Polish

## Migration from Old Version

The new version has been completely rewritten with:
- Removed tab-based navigation (no more numbered tabs 1-5)
- Chat-first interface instead of dashboard
- Unified input via prompt line
- Command-driven instead of screen-based navigation

Old keyboard shortcuts (1-5 for tabs) have been replaced with "/" commands.

## License

Part of the BATTLESTAG-BOT project.
