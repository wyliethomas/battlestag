# Communication Observer

A modular Go application that monitors communication channels (Slack, Gmail, Trello), extracts action items using AI, and writes them to note-taking systems (Obsidian or Notion).

## Features

- **Multi-Platform Support**: Monitor Slack channels, Gmail labels, and Trello boards
- **AI-Powered Extraction**: Uses LLMs (Claude, GPT) to identify action items
- **Flexible Output**: Write to Obsidian vaults or Notion databases
- **Stateful Syncing**: Tracks last sync time per channel to avoid duplicates
- **Queue-Based Architecture**: Decouples data collection from processing
- **Self-Contained Executables**: Each component runs independently

## Architecture

The system consists of separate executables that communicate via a filesystem queue:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Slack     │────▶│             │     │             │
│ Connector   │     │   Message   │     │             │
└─────────────┘     │    Queue    │────▶│ Note Writer │────▶ Obsidian/Notion
┌─────────────┐     │             │     │             │
│   Gmail     │────▶│ (Filesystem)│     │ (LLM Extract│
│ Connector   │     │             │     │  + Write)   │
└─────────────┘     │             │     │             │
┌─────────────┐     │             │     │             │
│   Trello    │────▶│             │     │             │
│ Connector   │     └─────────────┘     └─────────────┘
└─────────────┘
```

**Design Principles**:
- Connectors only fetch and queue messages (no LLM calls)
- Note writer handles extraction and writing
- Each executable loads config, does its job, and exits
- State is persisted between runs

## Project Structure

```
com-observer/
├── cmd/
│   ├── slack/       # Slack message collector
│   ├── gmail/       # Gmail message collector
│   ├── trello/      # Trello action collector
│   └── notewriter/  # Task extractor and writer
├── internal/
│   ├── config/      # Configuration loading
│   ├── messages/    # Message types and queue
│   ├── extractor/   # LLM-based task extraction
│   ├── writers/     # Obsidian and Notion writers
│   └── state/       # Sync state management
├── config.yaml.example
├── go.mod
└── README.md
```

## Installation

### Prerequisites

- Go 1.21 or later
- API credentials for the services you want to use:
  - Slack: Bot token
  - Gmail: OAuth credentials
  - Trello: API key and token
  - LLM: Anthropic or OpenAI API key
  - Notion: Integration token (if using Notion)

### Build

```bash
# Build all executables
go build -o bin/slack ./cmd/slack
go build -o bin/gmail ./cmd/gmail
go build -o bin/trello ./cmd/trello
go build -o bin/notewriter ./cmd/notewriter
```

## Configuration

1. Copy the example configuration:
```bash
cp config.yaml.example config.yaml
```

2. Edit `config.yaml` with your credentials and preferences:

```yaml
user:
  name: "Your Name"
  slack_id: "U12345678"
  email: "you@example.com"

queue:
  type: "filesystem"
  path: "/tmp/com-observer/queue"

slack:
  token: "xoxb-your-bot-token"
  channels_include: []  # Empty = all channels
  channels_exclude: []
  watch_dms: true
  sync_interval: "5m"

gmail:
  credentials_file: "/path/to/gmail-credentials.json"
  labels: ["INBOX", "IMPORTANT"]
  sync_interval: "10m"

trello:
  api_key: "your-trello-api-key"
  token: "your-trello-token"
  boards: ["board-id-1", "board-id-2"]
  sync_interval: "15m"

notewriter:
  type: "obsidian"  # or "notion"
  obsidian:
    vault_path: "/path/to/obsidian/vault"
    tasks_file: "Tasks/Action Items.md"
  notion:
    token: "your-notion-integration-token"
    database_id: "your-database-id"

llm:
  provider: "anthropic"  # "anthropic", "openai", or "custom"
  api_key: "your-api-key"
  model: "claude-3-5-sonnet-20241022"
  endpoint: ""  # Optional for custom providers
```

## Usage

### Manual Execution

Run each connector to fetch new messages:

```bash
# Sync Slack messages
./bin/slack -config config.yaml -state-dir /tmp/com-observer/state

# Sync Gmail messages
./bin/gmail -config config.yaml -state-dir /tmp/com-observer/state

# Sync Trello actions
./bin/trello -config config.yaml -state-dir /tmp/com-observer/state

# Process queue and write tasks
./bin/notewriter -config config.yaml
```

### Scheduled Execution (Cron)

Add to your crontab for automated monitoring:

```bash
# Run Slack sync every 5 minutes
*/5 * * * * /path/to/bin/slack -config /path/to/config.yaml -state-dir /tmp/com-observer/state

# Run Gmail sync every 10 minutes
*/10 * * * * /path/to/bin/gmail -config /path/to/config.yaml -state-dir /tmp/com-observer/state

# Run Trello sync every 15 minutes
*/15 * * * * /path/to/bin/trello -config /path/to/config.yaml -state-dir /tmp/com-observer/state

# Process queue every minute
* * * * * /path/to/bin/notewriter -config /path/to/config.yaml
```

### Systemd Timers

For more robust scheduling, use systemd timers. See [systemd timer documentation](https://www.freedesktop.org/software/systemd/man/systemd.timer.html).

## API Credential Setup

### Slack

1. Create a Slack app at https://api.slack.com/apps
2. Add Bot Token Scopes:
   - `channels:history`
   - `channels:read`
   - `groups:history`
   - `groups:read`
   - `im:history`
   - `im:read`
3. Install app to workspace
4. Copy Bot User OAuth Token to `config.yaml`

### Gmail

1. Create a project in Google Cloud Console
2. Enable Gmail API
3. Create OAuth 2.0 credentials (Desktop app)
4. Download credentials JSON file
5. Reference file path in `config.yaml`
6. On first run, you'll be prompted to authorize

### Trello

1. Get API key from https://trello.com/app-key
2. Generate token using the link on that page
3. Add both to `config.yaml`

### Notion

1. Create an integration at https://www.notion.so/my-integrations
2. Share your database with the integration
3. Copy integration token and database ID to `config.yaml`

**Note**: The Notion database must have these properties:
- `Name` (Title)
- `Status` (Select)
- `Source` (Select)
- `Channel` (Rich Text)
- `Author` (Rich Text)
- `Date` (Date)
- `Link` (URL)

## How It Works

### 1. Message Collection

Each connector:
- Loads sync state from previous run
- Fetches new messages since last sync
- Converts to standard message format
- Pushes to filesystem queue
- Updates sync state

### 2. Task Extraction

The note writer:
- Pops messages from queue (batch of 20)
- Sends to LLM with context about user identity
- LLM analyzes conversations for action items
- Returns structured task data

### 3. Task Writing

The note writer:
- Checks if task already exists (by permalink)
- Writes new tasks to configured system:
  - **Obsidian**: Appends to markdown file
  - **Notion**: Creates database entry

## Example Output

### Obsidian Format

```markdown
- [ ] Review and approve the PR for authentication refactor
  - **Slack**: #engineering | @john | 2024-11-28
  - PR is ready for review, includes tests and documentation
  - [View](https://slack.com/archives/C123456/p1234567890)

- [ ] Update the API documentation with new endpoints
  - **Gmail**: Project Planning | john@company.com | 2024-11-28
  - Need to document the new v2 endpoints before launch
  - [View](https://mail.google.com/mail/u/0/#inbox/abc123)
```

### Notion Database

Creates entries with:
- **Name**: Task description
- **Status**: "To Do"
- **Source**: "slack", "gmail", or "trello"
- **Channel**: Where it came from
- **Author**: Who mentioned it
- **Date**: When it was mentioned
- **Link**: Permalink to source

## Troubleshooting

### No messages being collected

- Check API credentials are valid
- Verify channel/label IDs are correct
- Check state directory is writable
- Look at connector logs for errors

### Tasks not being extracted

- Verify LLM API key is valid
- Check queue directory has messages
- Ensure user name in config matches your identity in messages
- Review note writer logs

### Duplicate tasks appearing

- Check that permalink detection is working
- Verify state is being saved properly
- For Obsidian: ensure tasks file hasn't been manually modified

### Rate limiting

- Adjust sync intervals in config
- Add delays between API calls if needed
- Check API quota limits

## Development

### Running Tests

```bash
go test ./...
```

### Adding a New Connector

1. Create new directory in `cmd/`
2. Implement message fetching logic
3. Convert to standard `messages.Message` format
4. Push to queue
5. Manage state for incremental sync

### Adding a New Writer

1. Create new file in `internal/writers/`
2. Implement `Writer` interface:
   - `WriteTask(task messages.Task) error`
   - `TaskExists(permalink string) bool`
3. Update `NewWriter()` factory function
4. Add configuration struct to `config.go`

## License

MIT

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request
