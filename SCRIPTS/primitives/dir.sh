#!/usr/bin/env bash
# dir.sh - Directory operation primitives
# Pure bash directory operations
# Exit codes: 0 = success, 1 = error
# Output: stdout for data, stderr for errors

set -euo pipefail

# dir.create <path>
# Create directory (mkdir -p)
dir.create() {
    local path="$1"

    if [[ -d "$path" ]]; then
        return 0  # Already exists, success
    fi

    mkdir -p "$path"
}

# dir.list <path>
# List files in directory (one per line)
dir.list() {
    local path="${1:-.}"  # Default to current directory

    if [[ ! -d "$path" ]]; then
        echo "Error: Directory not found: $path" >&2
        return 1
    fi

    ls -1 "$path"
}

# dir.tree <path> <depth>
# Recursive list (tree-like)
dir.tree() {
    local path="${1:-.}"
    local depth="${2:-999}"  # Default: unlimited depth

    if [[ ! -d "$path" ]]; then
        echo "Error: Directory not found: $path" >&2
        return 1
    fi

    find "$path" -maxdepth "$depth" -type f -o -type d | sort
}

# dir.exists <path>
# Check if directory exists
# Returns: 0 if exists, 1 if not
dir.exists() {
    local path="$1"
    [[ -d "$path" ]]
}

# dir.delete <path>
# Delete directory recursively (DANGEROUS!)
dir.delete() {
    local path="$1"

    if [[ ! -d "$path" ]]; then
        echo "Error: Directory not found: $path" >&2
        return 1
    fi

    # Safety check: don't delete root or home
    if [[ "$path" == "/" || "$path" == "$HOME" || "$path" == "~" ]]; then
        echo "Error: Refusing to delete critical directory: $path" >&2
        return 1
    fi

    rm -rf "$path"
}

# dir.current
# Get current working directory (pwd)
dir.current() {
    pwd
}

# dir.parent <path>
# Get parent directory
dir.parent() {
    local path="$1"
    dirname "$path"
}

# dir.basename <path>
# Get directory name (last component)
dir.basename() {
    local path="$1"
    basename "$path"
}

# dir.empty <path>
# Check if directory is empty
# Returns: 0 if empty, 1 if not empty or doesn't exist
dir.empty() {
    local path="$1"

    if [[ ! -d "$path" ]]; then
        echo "Error: Directory not found: $path" >&2
        return 1
    fi

    # Directory is empty if ls returns nothing
    if [[ -z "$(ls -A "$path" 2>/dev/null)" ]]; then
        return 0  # Empty
    else
        return 1  # Not empty
    fi
}

# dir.count <path>
# Count items in directory (files + subdirs)
dir.count() {
    local path="${1:-.}"

    if [[ ! -d "$path" ]]; then
        echo "Error: Directory not found: $path" >&2
        return 1
    fi

    ls -1 "$path" | wc -l
}

# dir.size <path>
# Get directory size in bytes (recursive)
dir.size() {
    local path="${1:-.}"

    if [[ ! -d "$path" ]]; then
        echo "Error: Directory not found: $path" >&2
        return 1
    fi

    # Use du with bytes format
    du -sb "$path" 2>/dev/null | cut -f1
}

# Main dispatcher
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -lt 1 ]]; then
        echo "Usage: dir.sh <operation> [args...]" >&2
        echo "Operations: create, list, tree, exists, delete, current, parent, basename, empty, count, size" >&2
        exit 1
    fi

    operation="$1"
    shift

    case "$operation" in
        create|list|tree|exists|delete|current|parent|basename|empty|count|size)
            "dir.$operation" "$@"
            ;;
        *)
            echo "Error: Unknown operation: $operation" >&2
            exit 1
            ;;
    esac
fi
