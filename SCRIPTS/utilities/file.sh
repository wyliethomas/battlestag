#!/usr/bin/env bash
# file.sh - File utility operations
# Composes primitives for higher-level file operations
# Exit codes: 0 = success, 1 = error

set -euo pipefail

# Get script directory and source primitives
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRIMITIVES_DIR="$(dirname "$SCRIPT_DIR")/primitives"

source "$PRIMITIVES_DIR/file.sh"
source "$PRIMITIVES_DIR/dir.sh"
source "$PRIMITIVES_DIR/sys.sh"

# util.backup <file>
# Create timestamped backup of file
util.backup() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    local timestamp
    timestamp=$(sys.timestamp)
    local backup_file="${file}.backup.${timestamp}"

    file.copy "$file" "$backup_file"
    echo "$backup_file"
}

# util.safe_write <file> <content>
# Backup then write (atomic-ish operation)
util.safe_write() {
    local file="$1"
    local content="$2"

    # Backup if file exists
    if file.exists "$file"; then
        local backup
        backup=$(util.backup "$file")
        echo "Backup created: $backup" >&2
    fi

    # Write new content
    file.write "$file" "$content"
}

# util.restore <file> <timestamp>
# Restore file from backup
util.restore() {
    local file="$1"
    local timestamp="$2"
    local backup_file="${file}.backup.${timestamp}"

    if ! file.exists "$backup_file"; then
        echo "Error: Backup not found: $backup_file" >&2
        return 1
    fi

    file.copy "$backup_file" "$file"
    echo "Restored from: $backup_file" >&2
}

# util.ensure_dir <path>
# Create directory if it doesn't exist
util.ensure_dir() {
    local path="$1"

    if ! dir.exists "$path"; then
        dir.create "$path"
    fi
}

# util.clean_old_files <directory> <days>
# Remove files older than N days
util.clean_old_files() {
    local directory="$1"
    local days="${2:-30}"

    if ! dir.exists "$directory"; then
        echo "Error: Directory not found: $directory" >&2
        return 1
    fi

    # Find and delete files older than N days
    find "$directory" -type f -mtime +"$days" -delete
}

# util.archive <directory> <output_file>
# Create tar.gz archive of directory
util.archive() {
    local directory="$1"
    local output_file="$2"

    if ! dir.exists "$directory"; then
        echo "Error: Directory not found: $directory" >&2
        return 1
    fi

    # Ensure output directory exists
    local output_dir
    output_dir=$(dir.parent "$output_file")
    util.ensure_dir "$output_dir"

    # Create archive
    tar -czf "$output_file" -C "$(dir.parent "$directory")" "$(dir.basename "$directory")"
    echo "Archive created: $output_file" >&2
}

# util.unarchive <archive_file> <destination>
# Extract tar.gz archive
util.unarchive() {
    local archive_file="$1"
    local destination="${2:-.}"

    if ! file.exists "$archive_file"; then
        echo "Error: Archive not found: $archive_file" >&2
        return 1
    fi

    util.ensure_dir "$destination"
    tar -xzf "$archive_file" -C "$destination"
    echo "Extracted to: $destination" >&2
}

# util.template <template_file> <output_file> <var1=value1> <var2=value2> ...
# Replace {{var}} placeholders in template
util.template() {
    local template_file="$1"
    local output_file="$2"
    shift 2

    if ! file.exists "$template_file"; then
        echo "Error: Template not found: $template_file" >&2
        return 1
    fi

    # Read template
    local content
    content=$(file.read "$template_file")

    # Replace variables
    for var_pair in "$@"; do
        local var_name="${var_pair%%=*}"
        local var_value="${var_pair#*=}"
        # Replace {{var_name}} with value
        content="${content//\{\{$var_name\}\}/$var_value}"
    done

    # Write output
    file.write "$output_file" "$content"
}

# util.config_get <config_file> <key>
# Get value from KEY=VALUE config file
util.config_get() {
    local config_file="$1"
    local key="$2"

    if ! file.exists "$config_file"; then
        echo "Error: Config file not found: $config_file" >&2
        return 1
    fi

    grep "^${key}=" "$config_file" | cut -d'=' -f2- | head -1
}

# util.config_set <config_file> <key> <value>
# Set value in KEY=VALUE config file (updates or adds)
util.config_set() {
    local config_file="$1"
    local key="$2"
    local value="$3"

    # Create file if doesn't exist
    if ! file.exists "$config_file"; then
        file.write "$config_file" ""
    fi

    # Check if key exists
    if grep -q "^${key}=" "$config_file" 2>/dev/null; then
        # Update existing key (cross-platform sed)
        if [[ "$(uname)" == "Darwin" ]]; then
            sed -i '' "s|^${key}=.*|${key}=${value}|" "$config_file"
        else
            sed -i "s|^${key}=.*|${key}=${value}|" "$config_file"
        fi
    else
        # Add new key
        file.append "$config_file" "${key}=${value}"
    fi
}

# util.file_age <file>
# Get file age in seconds
util.file_age() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    local modified
    modified=$(file.modified "$file")
    local now
    now=$(sys.timestamp)
    echo $((now - modified))
}

# util.temp_file <prefix>
# Create temporary file with prefix
util.temp_file() {
    local prefix="${1:-temp}"
    local temp_dir
    temp_dir=$(sys.temp)
    local timestamp
    timestamp=$(sys.timestamp)
    local random
    random=$(sys.random 1000 9999)

    local temp_file="${temp_dir}/${prefix}.${timestamp}.${random}"
    file.write "$temp_file" ""
    echo "$temp_file"
}

# util.temp_dir <prefix>
# Create temporary directory with prefix
util.temp_dir() {
    local prefix="${1:-temp}"
    local temp_base
    temp_base=$(sys.temp)
    local timestamp
    timestamp=$(sys.timestamp)
    local random
    random=$(sys.random 1000 9999)

    local temp_dir="${temp_base}/${prefix}.${timestamp}.${random}"
    dir.create "$temp_dir"
    echo "$temp_dir"
}

# Main dispatcher
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -lt 1 ]]; then
        echo "Usage: file.sh <operation> [args...]" >&2
        echo "Operations:" >&2
        echo "  backup, safe_write, restore, ensure_dir, clean_old_files" >&2
        echo "  archive, unarchive, template, config_get, config_set" >&2
        echo "  file_age, temp_file, temp_dir" >&2
        exit 1
    fi

    operation="$1"
    shift

    case "$operation" in
        backup|safe_write|restore|ensure_dir|clean_old_files|archive|unarchive|template|config_get|config_set|file_age|temp_file|temp_dir)
            "util.$operation" "$@"
            ;;
        *)
            echo "Error: Unknown operation: $operation" >&2
            exit 1
            ;;
    esac
fi
