#!/usr/bin/env bash
# file.sh - File operation primitives
# Pure bash file operations - the foundation layer
# Exit codes: 0 = success, 1 = error
# Output: stdout for data, stderr for errors

set -euo pipefail

# file.read <path>
# Read file contents to stdout
file.read() {
    local path="$1"

    if [[ ! -f "$path" ]]; then
        echo "Error: File not found: $path" >&2
        return 1
    fi

    cat "$path"
}

# file.write <path> <content>
# Write content to file (overwrites)
file.write() {
    local path="$1"
    local content="$2"

    # Ensure parent directory exists
    local parent_dir
    parent_dir=$(dirname "$path")
    if [[ ! -d "$parent_dir" ]]; then
        mkdir -p "$parent_dir" || {
            echo "Error: Cannot create directory: $parent_dir" >&2
            return 1
        }
    fi

    echo "$content" > "$path"
}

# file.append <path> <content>
# Append content to file
file.append() {
    local path="$1"
    local content="$2"

    # Create file if it doesn't exist
    if [[ ! -f "$path" ]]; then
        local parent_dir
        parent_dir=$(dirname "$path")
        if [[ ! -d "$parent_dir" ]]; then
            mkdir -p "$parent_dir" || {
                echo "Error: Cannot create directory: $parent_dir" >&2
                return 1
            }
        fi
        touch "$path"
    fi

    echo "$content" >> "$path"
}

# file.exists <path>
# Check if file exists
# Returns: 0 if exists, 1 if not
file.exists() {
    local path="$1"
    [[ -f "$path" ]]
}

# file.size <path>
# Get file size in bytes
file.size() {
    local path="$1"

    if [[ ! -f "$path" ]]; then
        echo "Error: File not found: $path" >&2
        return 1
    fi

    stat -c%s "$path" 2>/dev/null || stat -f%z "$path" 2>/dev/null
}

# file.modified <path>
# Get last modification timestamp (Unix epoch)
file.modified() {
    local path="$1"

    if [[ ! -f "$path" ]]; then
        echo "Error: File not found: $path" >&2
        return 1
    fi

    stat -c%Y "$path" 2>/dev/null || stat -f%m "$path" 2>/dev/null
}

# file.delete <path>
# Delete file
file.delete() {
    local path="$1"

    if [[ ! -f "$path" ]]; then
        echo "Error: File not found: $path" >&2
        return 1
    fi

    rm -f "$path"
}

# file.copy <src> <dst>
# Copy file
file.copy() {
    local src="$1"
    local dst="$2"

    if [[ ! -f "$src" ]]; then
        echo "Error: Source file not found: $src" >&2
        return 1
    fi

    # Ensure destination parent directory exists
    local parent_dir
    parent_dir=$(dirname "$dst")
    if [[ ! -d "$parent_dir" ]]; then
        mkdir -p "$parent_dir" || {
            echo "Error: Cannot create directory: $parent_dir" >&2
            return 1
        }
    fi

    cp "$src" "$dst"
}

# file.move <src> <dst>
# Move/rename file
file.move() {
    local src="$1"
    local dst="$2"

    if [[ ! -f "$src" ]]; then
        echo "Error: Source file not found: $src" >&2
        return 1
    fi

    # Ensure destination parent directory exists
    local parent_dir
    parent_dir=$(dirname "$dst")
    if [[ ! -d "$parent_dir" ]]; then
        mkdir -p "$parent_dir" || {
            echo "Error: Cannot create directory: $parent_dir" >&2
            return 1
        }
    fi

    mv "$src" "$dst"
}

# file.permissions <path> <mode>
# Set file permissions (chmod)
file.permissions() {
    local path="$1"
    local mode="$2"

    if [[ ! -f "$path" ]]; then
        echo "Error: File not found: $path" >&2
        return 1
    fi

    chmod "$mode" "$path"
}

# Main dispatcher - allows calling as: file.sh read /path/to/file
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -lt 1 ]]; then
        echo "Usage: file.sh <operation> [args...]" >&2
        echo "Operations: read, write, append, exists, size, modified, delete, copy, move, permissions" >&2
        exit 1
    fi

    operation="$1"
    shift

    case "$operation" in
        read|write|append|exists|size|modified|delete|copy|move|permissions)
            "file.$operation" "$@"
            ;;
        *)
            echo "Error: Unknown operation: $operation" >&2
            exit 1
            ;;
    esac
fi
