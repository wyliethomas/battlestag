#!/bin/bash
# Install com-observer program

set -e

TARGET_DIR="${BATTLESTAG_TARGET_DIR:-$HOME/battlestag}"
SCRIPT_DIR="${BATTLESTAG_SCRIPT_DIR}"
PROGRAM_DIR="$TARGET_DIR/programs/com-observer"

echo "Installing com-observer"
echo "======================="
echo ""

# Check if already installed
if [ -d "$PROGRAM_DIR" ]; then
    echo "⚠ com-observer already installed at $PROGRAM_DIR"
    echo ""
    read -p "Reinstall? This will preserve your config. (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi
fi

# Create directories
echo "Creating installation directory..."
mkdir -p "$PROGRAM_DIR"
mkdir -p "$PROGRAM_DIR/queue"
mkdir -p "$PROGRAM_DIR/state"
mkdir -p "$PROGRAM_DIR/logs"

# Install from package or source
PACKAGES_DIR="$SCRIPT_DIR/packages"
if [ -f "$PACKAGES_DIR/com-observer.tar.gz" ]; then
    echo "✓ Using pre-built package (offline mode)"
    tar -xzf "$PACKAGES_DIR/com-observer.tar.gz" -C "$PROGRAM_DIR"
else
    echo "⚠ No pre-built package found, building from source..."

    ROOT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
    SOURCE_DIR="$ROOT_DIR/PROGRAMS/com-observer"

    if [ ! -d "$SOURCE_DIR" ]; then
        echo "✗ Source code not found"
        exit 1
    fi

    cd "$SOURCE_DIR"
    make build

    # Copy files
    cp -r bin "$PROGRAM_DIR/"
    cp -r cmd "$PROGRAM_DIR/"
    cp -r internal "$PROGRAM_DIR/"
    cp config.yaml.example "$PROGRAM_DIR/"
    cp README.md "$PROGRAM_DIR/" 2>/dev/null || true
    cp Makefile "$PROGRAM_DIR/"
    cp go.mod "$PROGRAM_DIR/"
    cp go.sum "$PROGRAM_DIR/" 2>/dev/null || true
fi

echo "✓ Files installed"

# Setup configuration
CONFIG_FILE="$PROGRAM_DIR/config.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo ""
    echo "Setting up configuration..."
    echo "This program monitors Slack/Gmail/Trello and extracts tasks to Obsidian/Notion"
    echo ""

    cp "$PROGRAM_DIR/config.yaml.example" "$CONFIG_FILE"

    # Prompt for configuration
    read -p "Your name: " USER_NAME
    read -p "Your Slack user ID: " SLACK_ID
    read -p "Your email: " USER_EMAIL

    echo ""
    echo "Slack Configuration:"
    read -p "Slack bot token (xoxb-...): " SLACK_TOKEN

    echo ""
    echo "LLM Configuration:"
    echo "1) Local Ollama"
    echo "2) Anthropic Claude"
    echo "3) OpenAI GPT"
    read -p "Select LLM provider (1-3): " LLM_CHOICE

    case $LLM_CHOICE in
        1)
            read -p "Ollama endpoint [http://localhost:11434/v1/chat/completions]: " OLLAMA_ENDPOINT
            OLLAMA_ENDPOINT=${OLLAMA_ENDPOINT:-http://localhost:11434/v1/chat/completions}
            read -p "Ollama model [llama3.1:8b]: " OLLAMA_MODEL
            OLLAMA_MODEL=${OLLAMA_MODEL:-llama3.1:8b}

            LLM_PROVIDER="openai"
            LLM_ENDPOINT="$OLLAMA_ENDPOINT"
            LLM_MODEL="$OLLAMA_MODEL"
            LLM_KEY="not-needed"
            ;;
        2)
            read -p "Anthropic API key: " ANTHROPIC_KEY
            LLM_PROVIDER="anthropic"
            LLM_ENDPOINT=""
            LLM_MODEL="claude-3-5-sonnet-20241022"
            LLM_KEY="$ANTHROPIC_KEY"
            ;;
        3)
            read -p "OpenAI API key: " OPENAI_KEY
            LLM_PROVIDER="openai"
            LLM_ENDPOINT=""
            LLM_MODEL="gpt-4"
            LLM_KEY="$OPENAI_KEY"
            ;;
    esac

    echo ""
    echo "Note Writer Configuration:"
    echo "1) Obsidian"
    echo "2) Notion"
    read -p "Select note writer (1-2): " WRITER_CHOICE

    case $WRITER_CHOICE in
        1)
            read -p "Obsidian vault path: " VAULT_PATH
            read -p "Tasks file [Tasks/Action Items.md]: " TASKS_FILE
            TASKS_FILE=${TASKS_FILE:-Tasks/Action Items.md}
            WRITER_TYPE="obsidian"
            ;;
        2)
            read -p "Notion integration token: " NOTION_TOKEN
            read -p "Notion database ID: " NOTION_DB
            WRITER_TYPE="notion"
            ;;
    esac

    # Update config file
    sed -i "s/YOUR_NAME_HERE/$USER_NAME/" "$CONFIG_FILE" || sed -i '' "s/YOUR_NAME_HERE/$USER_NAME/" "$CONFIG_FILE"
    sed -i "s/YOUR_SLACK_ID_HERE/$SLACK_ID/" "$CONFIG_FILE" || sed -i '' "s/YOUR_SLACK_ID_HERE/$SLACK_ID/" "$CONFIG_FILE"
    sed -i "s/YOUR_EMAIL_HERE/$USER_EMAIL/" "$CONFIG_FILE" || sed -i '' "s/YOUR_EMAIL_HERE/$USER_EMAIL/" "$CONFIG_FILE"
    sed -i "s|YOUR_SLACK_BOT_TOKEN_HERE|$SLACK_TOKEN|" "$CONFIG_FILE" || sed -i '' "s|YOUR_SLACK_BOT_TOKEN_HERE|$SLACK_TOKEN|" "$CONFIG_FILE"

    # Update LLM config
    sed -i "s/provider: \".*\"/provider: \"$LLM_PROVIDER\"/" "$CONFIG_FILE" || sed -i '' "s/provider: \".*\"/provider: \"$LLM_PROVIDER\"/" "$CONFIG_FILE"
    sed -i "s|YOUR_ANTHROPIC_API_KEY_HERE|$LLM_KEY|" "$CONFIG_FILE" || sed -i '' "s|YOUR_ANTHROPIC_API_KEY_HERE|$LLM_KEY|" "$CONFIG_FILE"
    sed -i "s/model: \".*\"/model: \"$LLM_MODEL\"/" "$CONFIG_FILE" || sed -i '' "s/model: \".*\"/model: \"$LLM_MODEL\"/" "$CONFIG_FILE"

    if [ -n "$LLM_ENDPOINT" ]; then
        sed -i "s|endpoint: \"\"|endpoint: \"$LLM_ENDPOINT\"|" "$CONFIG_FILE" || sed -i '' "s|endpoint: \"\"|endpoint: \"$LLM_ENDPOINT\"|" "$CONFIG_FILE"
    fi

    # Update writer config
    sed -i "s/type: \".*\"/type: \"$WRITER_TYPE\"/" "$CONFIG_FILE" || sed -i '' "s/type: \".*\"/type: \"$WRITER_TYPE\"/" "$CONFIG_FILE"

    if [ "$WRITER_TYPE" == "obsidian" ]; then
        sed -i "s|vault_path: \".*\"|vault_path: \"$VAULT_PATH\"|" "$CONFIG_FILE" || sed -i '' "s|vault_path: \".*\"|vault_path: \"$VAULT_PATH\"|" "$CONFIG_FILE"
        sed -i "s|tasks_file: \".*\"|tasks_file: \"$TASKS_FILE\"|" "$CONFIG_FILE" || sed -i '' "s|tasks_file: \".*\"|tasks_file: \"$TASKS_FILE\"|" "$CONFIG_FILE"
    fi

    # Update queue path
    sed -i "s|path: \"/tmp/com-observer/queue\"|path: \"$PROGRAM_DIR/queue\"|" "$CONFIG_FILE" || sed -i '' "s|path: \"/tmp/com-observer/queue\"|path: \"$PROGRAM_DIR/queue\"|" "$CONFIG_FILE"

    echo "✓ Configuration created"
else
    echo "✓ Existing configuration preserved"
fi

# Setup cron
echo ""
read -p "Setup automated monitoring with cron? (y/n) [y]: " SETUP_CRON
SETUP_CRON=${SETUP_CRON:-y}

if [[ $SETUP_CRON =~ ^[Yy]$ ]]; then
    # Create cron entries
    TEMP_CRON=$(mktemp)
    crontab -l > "$TEMP_CRON" 2>/dev/null || true

    # Remove existing com-observer entries
    grep -v "com-observer" "$TEMP_CRON" > "$TEMP_CRON.tmp" 2>/dev/null || true
    mv "$TEMP_CRON.tmp" "$TEMP_CRON"

    # Add new entries
    echo "" >> "$TEMP_CRON"
    echo "# com-observer - Slack message monitoring" >> "$TEMP_CRON"
    echo "*/5 * * * * cd $PROGRAM_DIR && ./bin/slack -config config.yaml -state-dir $PROGRAM_DIR/state >> $PROGRAM_DIR/logs/slack.log 2>&1" >> "$TEMP_CRON"
    echo "" >> "$TEMP_CRON"
    echo "# com-observer - Task extraction and writing" >> "$TEMP_CRON"
    echo "* * * * * cd $PROGRAM_DIR && ./bin/notewriter -config config.yaml >> $PROGRAM_DIR/logs/notewriter.log 2>&1" >> "$TEMP_CRON"

    crontab "$TEMP_CRON"
    rm "$TEMP_CRON"

    echo "✓ Cron jobs installed"
fi

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  com-observer Installation Complete!                       ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "Installation directory: $PROGRAM_DIR"
echo "Configuration file: $CONFIG_FILE"
echo ""
echo "Next steps:"
echo "  1. Test Slack sync: cd $PROGRAM_DIR && ./bin/slack -config config.yaml -state-dir state"
echo "  2. Process tasks: cd $PROGRAM_DIR && ./bin/notewriter -config config.yaml"
echo "  3. View logs: tail -f $PROGRAM_DIR/logs/*.log"
echo ""
if [[ $SETUP_CRON =~ ^[Yy]$ ]]; then
    echo "Automated monitoring is active:"
    echo "  - Slack sync: every 5 minutes"
    echo "  - Task processing: every minute"
    echo ""
fi
