# Primitives Layer

**Pure bash operations - the foundation of the capability hierarchy**

## Overview

The primitives layer provides atomic, single-purpose operations that form the foundation for all higher-level functionality. Each primitive:

- ✅ Executes in < 100ms
- ✅ Has zero external dependencies (pure bash)
- ✅ Returns exit code 0 for success, 1 for error
- ✅ Outputs data to stdout, errors to stderr
- ✅ Is idempotent where possible
- ✅ Has 100% test coverage

## Modules

### file.sh - File Operations

Basic file manipulation primitives.

```bash
# Read/Write
file.read <path>                    # Read file to stdout
file.write <path> <content>         # Write/overwrite file
file.append <path> <content>        # Append to file

# Metadata
file.exists <path>                  # Check if file exists (exit 0/1)
file.size <path>                    # Get size in bytes
file.modified <path>                # Get modification timestamp

# Operations
file.delete <path>                  # Delete file
file.copy <src> <dst>              # Copy file
file.move <src> <dst>              # Move/rename file
file.permissions <path> <mode>      # Set permissions (chmod)
```

**Examples:**
```bash
# Write and read
file.write "/tmp/test.txt" "Hello World"
file.read "/tmp/test.txt"

# Check existence
if file.exists "/tmp/test.txt"; then
    echo "File exists"
fi

# Copy and move
file.copy "/tmp/test.txt" "/tmp/test_backup.txt"
file.move "/tmp/test_backup.txt" "/tmp/test_final.txt"
```

### dir.sh - Directory Operations

Directory management primitives.

```bash
# CRUD
dir.create <path>                   # mkdir -p
dir.list <path>                     # List files (one per line)
dir.tree <path> <depth>            # Recursive list
dir.exists <path>                   # Check if directory exists
dir.delete <path>                   # rm -rf (with safety checks)

# Navigation
dir.current                         # pwd
dir.parent <path>                   # Get parent directory
dir.basename <path>                 # Get directory name

# Info
dir.empty <path>                    # Check if empty (exit 0/1)
dir.count <path>                    # Count items in directory
dir.size <path>                     # Get total size in bytes
```

**Examples:**
```bash
# Create and list
dir.create "/tmp/mydir"
dir.list "/tmp/mydir"

# Check if empty
if dir.empty "/tmp/mydir"; then
    echo "Directory is empty"
fi

# Get info
dir.count "/tmp/mydir"
dir.size "/tmp/mydir"
```

### search.sh - Search Operations

Search primitives for finding content and files.

```bash
# Content search
search.grep <pattern> <path>              # Recursive grep
search.count <pattern> <path>             # Count matches
search.case_insensitive <pattern> <path>  # Case-insensitive search
search.exact <pattern> <path>             # Exact word match

# File search
search.find <pattern> <path>              # Find files by name
search.locate <name>                      # Quick locate
search.which <command>                    # Find executable in PATH

# Context search
search.before <pattern> <path> <lines>    # N lines before match
search.after <pattern> <path> <lines>     # N lines after match
search.around <pattern> <path> <lines>    # N lines before & after

# Helpers
search.files_with <pattern> <path>        # List files containing pattern
search.line_numbers <pattern> <path>      # Show line numbers in results
```

**Examples:**
```bash
# Search for pattern
search.grep "TODO" "/home/user/projects"

# Find files by name
search.find "*.rb" "/home/user/projects"

# Search with context
search.around "function.*main" "/home/user/app.rb" 5

# Case-insensitive search
search.case_insensitive "error" "/var/log"
```

### sys.sh - System Operations

System information and process management primitives.

```bash
# Environment
sys.exec <command> [args...]        # Execute arbitrary command
sys.env <var>                       # Get environment variable
sys.setenv <var> <value>           # Set environment variable

# Info
sys.hostname                        # Get hostname
sys.user                           # Get current user
sys.date                           # ISO 8601 date/time
sys.timestamp                      # Unix timestamp
sys.uptime                         # System uptime

# Process
sys.pid                            # Current process ID
sys.ppid                           # Parent process ID
sys.kill <pid> <signal>           # Kill process
sys.ps <pattern>                   # Find processes

# System
sys.memory                         # Total memory in MB
sys.disk <path>                    # Disk usage percentage
sys.cpu_count                      # Number of CPU cores
sys.os                             # Operating system name
sys.arch                           # System architecture

# Paths
sys.shell                          # Current shell
sys.home                           # Home directory
sys.temp                           # Temp directory

# Utilities
sys.sleep <seconds>                # Sleep for N seconds
sys.random <min> <max>            # Random number in range
sys.uuid                           # Generate UUID
```

**Examples:**
```bash
# System info
echo "Running on: $(sys.hostname)"
echo "User: $(sys.user)"
echo "CPUs: $(sys.cpu_count)"
echo "OS: $(sys.os) $(sys.arch)"

# Environment
sys.setenv "MY_VAR" "my_value"
echo $(sys.env "MY_VAR")

# Timestamps
echo "Current time: $(sys.date)"
echo "Unix timestamp: $(sys.timestamp)"

# Random & UUID
echo "Random number: $(sys.random 1 100)"
echo "UUID: $(sys.uuid)"
```

## Usage Patterns

### As Library (Source in Script)

```bash
#!/usr/bin/env bash
source ~/scripts/primitives/file.sh
source ~/scripts/primitives/dir.sh

# Now use primitives directly
file.write "/tmp/output.txt" "Processing..."
dir.create "/tmp/results"
```

### As CLI Tool

```bash
# Call primitives as standalone commands
~/scripts/primitives/file.sh read /tmp/test.txt
~/scripts/primitives/dir.sh list /tmp
~/scripts/primitives/search.sh grep "pattern" /home/user
~/scripts/primitives/sys.sh hostname
```

### In Higher Layers

Utilities and modules should source these primitives:

```bash
# In utilities/file.sh
source "$(dirname "${BASH_SOURCE[0]}")/../primitives/file.sh"

util.backup() {
    local file="$1"
    local timestamp=$(sys.timestamp)
    file.copy "$file" "${file}.backup.${timestamp}"
}
```

## Testing

Run the comprehensive test suite:

```bash
cd ~/scripts/primitives
./test_primitives.sh
```

**Test Coverage:**
- 40 tests across all 4 modules
- 100% success rate
- Tests all core functionality
- Validates error handling
- Verifies edge cases

## Performance

All primitives execute in < 100ms:

| Primitive | Typical Time | Notes |
|-----------|-------------|-------|
| file.read | 1-5ms | Depends on file size |
| file.write | 2-10ms | Includes parent dir creation |
| dir.list | 1-5ms | Scales with item count |
| search.grep | 10-100ms | Depends on search scope |
| sys.hostname | < 1ms | Cached by OS |
| sys.uuid | < 1ms | Uses /dev/urandom |

## Design Principles

### 1. Single Responsibility
Each primitive does ONE thing well.

### 2. No Side Effects
Primitives don't modify global state (except for sys.setenv).

### 3. Predictable Behavior
Same input = same output (except random/uuid).

### 4. Error Handling
Always return meaningful exit codes and error messages to stderr.

### 5. Composability
Primitives combine to create complex operations:

```bash
# Example: Safe copy with verification
safe_copy() {
    local src="$1"
    local dst="$2"

    if ! file.exists "$src"; then
        echo "Source not found" >&2
        return 1
    fi

    file.copy "$src" "$dst"

    if file.exists "$dst"; then
        echo "Copy successful"
    else
        echo "Copy failed" >&2
        return 1
    fi
}
```

## Common Patterns

### Check Then Execute

```bash
if file.exists "$config_file"; then
    content=$(file.read "$config_file")
    # Process content
else
    echo "Config not found, using defaults"
fi
```

### Ensure Directory Exists

```bash
if ! dir.exists "/tmp/output"; then
    dir.create "/tmp/output"
fi
file.write "/tmp/output/result.txt" "$data"
```

### Search and Process

```bash
# Find all Ruby files
ruby_files=$(search.find "*.rb" "$project_dir")

# Process each file
while IFS= read -r file; do
    content=$(file.read "$file")
    # Do something with content
done <<< "$ruby_files"
```

### Timestamp Everything

```bash
timestamp=$(sys.timestamp)
log_file="/var/log/app.${timestamp}.log"
file.write "$log_file" "Started at $(sys.date)"
```

## Error Handling

All primitives follow consistent error handling:

```bash
# Success
file.read "/tmp/exists.txt"
echo $?  # 0

# Failure
file.read "/tmp/nonexistent.txt"
echo $?  # 1 (error message sent to stderr)
```

Wrap in conditionals:

```bash
if file.read "/tmp/test.txt" > /dev/null 2>&1; then
    echo "File readable"
else
    echo "Cannot read file"
fi
```

## Next Steps

The primitives layer is now complete. Next up:

1. **Utilities Layer** - Compose primitives for higher-level ops
2. **Modules Layer** - Build domain-specific workflows
3. **API Integration** - Connect to agent-gateway
4. **TUI Integration** - Add command palette entries

## Contributing

When adding new primitives:

1. Follow existing naming convention (module.operation)
2. Keep functions pure (no side effects)
3. Add error handling (exit codes + stderr)
4. Write tests in test_primitives.sh
5. Update this README
6. Ensure 100% test pass rate

---

**Status**: ✅ Complete - All 40 tests passing
**Version**: 1.0
**Last Updated**: 2025-12-18
