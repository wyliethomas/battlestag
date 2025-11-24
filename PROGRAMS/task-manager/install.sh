#!/bin/bash
set -e

echo "=== Task Manager Installation ==="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.19 or higher."
    exit 1
fi

# Get installation directory
DEFAULT_INSTALL_DIR="$HOME/.local/bin"
read -p "Installation directory [$DEFAULT_INSTALL_DIR]: " INSTALL_DIR
INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"

# Create installation directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Get config directory
DEFAULT_CONFIG_DIR="$HOME/.config/task-manager"
read -p "Configuration directory [$DEFAULT_CONFIG_DIR]: " CONFIG_DIR
CONFIG_DIR="${CONFIG_DIR:-$DEFAULT_CONFIG_DIR}"

# Create config directory
mkdir -p "$CONFIG_DIR"

# Get database path
DEFAULT_DB_PATH="$HOME/.local/share/task-manager/tasks.db"
read -p "Database file path [$DEFAULT_DB_PATH]: " DB_PATH
DB_PATH="${DB_PATH:-$DEFAULT_DB_PATH}"

# Create database directory
DB_DIR=$(dirname "$DB_PATH")
mkdir -p "$DB_DIR"

echo
echo "Installing task-manager..."
echo "  Install directory: $INSTALL_DIR"
echo "  Config directory:  $CONFIG_DIR"
echo "  Database path:     $DB_PATH"
echo

# Build executables
echo "Building executables..."
go build -o task_project ./cmd/project
go build -o task_checklist ./cmd/checklist
go build -o task_journal ./cmd/journal
go build -o task_query ./cmd/query

# Install binaries
echo "Installing binaries..."
cp task_project "$INSTALL_DIR/"
cp task_checklist "$INSTALL_DIR/"
cp task_journal "$INSTALL_DIR/"
cp task_query "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/task_project"
chmod +x "$INSTALL_DIR/task_checklist"
chmod +x "$INSTALL_DIR/task_journal"
chmod +x "$INSTALL_DIR/task_query"

# Create .env file
cat > "$CONFIG_DIR/.env" <<EOF
# Task Manager Configuration
DB_PATH=$DB_PATH
EOF

# Create wrapper scripts
echo "Creating wrapper scripts..."

# Project wrapper
cat > "$INSTALL_DIR/task_project_run" <<EOF
#!/bin/bash
# Auto-generated wrapper script for task_project
CONFIG_DIR="$CONFIG_DIR"

# Load environment
if [ -f "\$CONFIG_DIR/.env" ]; then
    set -a
    source "\$CONFIG_DIR/.env"
    set +a
fi

# Run the executable
exec "$INSTALL_DIR/task_project" "\$@"
EOF

# Checklist wrapper
cat > "$INSTALL_DIR/task_checklist_run" <<EOF
#!/bin/bash
# Auto-generated wrapper script for task_checklist
CONFIG_DIR="$CONFIG_DIR"

# Load environment
if [ -f "\$CONFIG_DIR/.env" ]; then
    set -a
    source "\$CONFIG_DIR/.env"
    set +a
fi

# Run the executable
exec "$INSTALL_DIR/task_checklist" "\$@"
EOF

# Journal wrapper
cat > "$INSTALL_DIR/task_journal_run" <<EOF
#!/bin/bash
# Auto-generated wrapper script for task_journal
CONFIG_DIR="$CONFIG_DIR"

# Load environment
if [ -f "\$CONFIG_DIR/.env" ]; then
    set -a
    source "\$CONFIG_DIR/.env"
    set +a
fi

# Run the executable
exec "$INSTALL_DIR/task_journal" "\$@"
EOF

# Query wrapper
cat > "$INSTALL_DIR/task_query_run" <<EOF
#!/bin/bash
# Auto-generated wrapper script for task_query
CONFIG_DIR="$CONFIG_DIR"

# Load environment
if [ -f "\$CONFIG_DIR/.env" ]; then
    set -a
    source "\$CONFIG_DIR/.env"
    set +a
fi

# Run the executable
exec "$INSTALL_DIR/task_query" "\$@"
EOF

# Make wrappers executable
chmod +x "$INSTALL_DIR/task_project_run"
chmod +x "$INSTALL_DIR/task_checklist_run"
chmod +x "$INSTALL_DIR/task_journal_run"
chmod +x "$INSTALL_DIR/task_query_run"

echo
echo "Seeding default contexts..."

# Temporarily export DB_PATH for seeding
export DB_PATH="$DB_PATH"

# Seed default contexts
for context in "property" "house" "pxp" "personal"; do
    if "$INSTALL_DIR/task_project" contexts --add "$context" &>/dev/null; then
        echo "  ✓ Added context: $context"
    else
        echo "  - Context '$context' already exists or failed to add"
    fi
done

echo
echo "✓ Installation complete!"
echo
echo "Binaries installed:"
echo "  $INSTALL_DIR/task_project"
echo "  $INSTALL_DIR/task_checklist"
echo "  $INSTALL_DIR/task_journal"
echo "  $INSTALL_DIR/task_query"
echo
echo "Wrapper scripts installed:"
echo "  $INSTALL_DIR/task_project_run"
echo "  $INSTALL_DIR/task_checklist_run"
echo "  $INSTALL_DIR/task_journal_run"
echo "  $INSTALL_DIR/task_query_run"
echo
echo "Configuration file:"
echo "  $CONFIG_DIR/.env"
echo

# Check and update PATH
echo "Checking PATH configuration..."

# Detect shell and config file
SHELL_CONFIG=""
if [ -n "$BASH_VERSION" ]; then
    if [ -f "$HOME/.bashrc" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
    elif [ -f "$HOME/.bash_profile" ]; then
        SHELL_CONFIG="$HOME/.bash_profile"
    fi
elif [ -n "$ZSH_VERSION" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
else
    # Fallback: check for common files
    if [ -f "$HOME/.bashrc" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
    elif [ -f "$HOME/.bash_profile" ]; then
        SHELL_CONFIG="$HOME/.bash_profile"
    elif [ -f "$HOME/.profile" ]; then
        SHELL_CONFIG="$HOME/.profile"
    fi
fi

# Check if PATH contains the install directory
PATH_CONFIGURED=false
if [ -n "$PATH" ]; then
    if echo "$PATH" | grep -q "$INSTALL_DIR"; then
        PATH_CONFIGURED=true
    fi
fi

if [ "$PATH_CONFIGURED" = false ]; then
    if [ -n "$SHELL_CONFIG" ]; then
        echo "Adding $INSTALL_DIR to PATH in $SHELL_CONFIG"

        # Create backup
        cp "$SHELL_CONFIG" "${SHELL_CONFIG}.backup"

        # Add PATH export
        echo "" >> "$SHELL_CONFIG"
        echo "# Added by task-manager installer" >> "$SHELL_CONFIG"
        echo "export PATH=\"\$HOME/.local/bin:\$PATH\"" >> "$SHELL_CONFIG"

        echo "✓ PATH updated in $SHELL_CONFIG"
        echo ""
        echo "⚠️  IMPORTANT: Run this command to apply changes:"
        echo "    source $SHELL_CONFIG"
        echo ""
        echo "Or log out and log back in."
    else
        echo "⚠️  Could not detect shell configuration file."
        echo "   Please manually add this to your shell config (~/.bashrc or ~/.zshrc):"
        echo "   export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
else
    echo "✓ $INSTALL_DIR is already in PATH"
fi

echo
echo "Quick start:"
echo "  task_project_run contexts --list"
echo "  task_project_run add --name \"My Project\" --context property --goal \"Complete the project\""
echo "  task_checklist_run add --project 1 --items \"Task 1, Task 2, Task 3\""
echo "  task_journal_run add --project 1 --entry \"Started work today\""
echo "  task_query_run status --project 1"
echo
