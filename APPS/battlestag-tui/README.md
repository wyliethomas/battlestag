# BATTLESTAG TUI

A terminal user interface (TUI) client for the BATTLESTAG Agent Gateway API.

## Features

- **Dashboard**: View financial overview, net worth, and daily stoic thoughts
- **Asset Management**: Track and manage financial assets
- **Liability Management**: Monitor credit cards, loans, and other liabilities
- **PDF Upload**: Upload and process bank statements
- **Real-time Updates**: Refresh data on demand

## Installation

### Prerequisites

- Go 1.21 or higher
- Access to a running Agent Gateway API

### Build

```bash
cd /home/battlestag/Work/BATTLESTAG-BOT/APPS/battlestag-tui
go mod tidy
go build -o battlestag-tui
```

## Configuration

The TUI uses environment variables for configuration:

```bash
# Optional: API endpoint (default: http://192.168.1.140:8080)
export AGENT_GATEWAY_URL="http://192.168.1.140:8080"

# Optional: API authentication key (default: test-api-key-12345)
export AGENT_GATEWAY_API_KEY="test-api-key-12345"
```

**Note**: The defaults are already configured for the BATTLESTAG server at 192.168.1.140. You can run without setting these variables.

## Usage

### Running the TUI

```bash
./battlestag-tui
```

### Keyboard Shortcuts

#### Global Navigation
- `1` - Dashboard
- `2` - Assets
- `3` - Liabilities
- `4` - Upload
- `?` - Help
- `q` - Quit

#### List Navigation
- `↑/k` - Move up
- `↓/j` - Move down
- `r` - Refresh data
- `d` - Delete selected item
- `Esc` - Cancel/Go back

#### File Upload
- `Enter` - Submit file path
- `Ctrl+U` - Clear input
- `Backspace` - Delete character

## Screens

### Dashboard
Displays:
- Net worth and financial overview
- Total assets and liabilities
- Daily stoic thought
- System health status

### Assets
- View all tracked assets
- See current values and gains/losses
- Delete assets
- Summary totals

### Liabilities
- View all liabilities
- Monitor balances and credit utilization
- Delete liabilities
- Summary totals

### Upload
- Upload PDF bank statements
- Automatic processing and transaction extraction

## Development

### Project Structure

```
battlestag-tui/
├── main.go              # Application entry point
├── client/
│   └── api.go          # API client
├── models/
│   └── models.go       # Data structures
├── ui/
│   ├── app.go          # Main app model
│   ├── dashboard.go    # Dashboard view
│   ├── assets.go       # Assets view
│   ├── liabilities.go  # Liabilities view
│   ├── upload.go       # Upload view
│   └── styles.go       # UI styling
└── components/         # Reusable components (future)
```

### Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## Troubleshooting

### Connection Issues

If you see "Disconnected" status:
1. Verify the Agent Gateway is running on 192.168.1.140:
   ```bash
   curl http://192.168.1.140:8080/api/health
   ```
2. Check `AGENT_GATEWAY_URL` is correct (should be http://192.168.1.140:8080)
3. Verify `AGENT_GATEWAY_API_KEY` matches the server config (default: test-api-key-12345)
4. Ensure network connectivity to 192.168.1.140

### Build Errors

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod tidy
go mod download
```

## License

Part of the BATTLESTAG-BOT project.
