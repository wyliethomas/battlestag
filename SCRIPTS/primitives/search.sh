#!/usr/bin/env bash
# search.sh - Search operation primitives
# Pure bash search operations using grep, find
# Exit codes: 0 = success, 1 = error
# Output: stdout for data, stderr for errors

set -euo pipefail

# search.grep <pattern> <path>
# Recursive grep for pattern
search.grep() {
    local pattern="$1"
    local path="${2:-.}"

    if [[ ! -e "$path" ]]; then
        echo "Error: Path not found: $path" >&2
        return 1
    fi

    grep -r "$pattern" "$path" 2>/dev/null || true
}

# search.count <pattern> <path>
# Count matches of pattern
search.count() {
    local pattern="$1"
    local path="${2:-.}"

    if [[ ! -e "$path" ]]; then
        echo "Error: Path not found: $path" >&2
        return 1
    fi

    grep -rc "$pattern" "$path" 2>/dev/null | awk -F: '{sum+=$2} END {print sum}'
}

# search.find <pattern> <path>
# Find files by name pattern
search.find() {
    local pattern="$1"
    local path="${2:-.}"

    if [[ ! -d "$path" ]]; then
        echo "Error: Directory not found: $path" >&2
        return 1
    fi

    find "$path" -name "$pattern" -type f 2>/dev/null | sort
}

# search.locate <name>
# Quick locate file (if locate is available)
search.locate() {
    local name="$1"

    if ! command -v locate &>/dev/null; then
        echo "Error: locate command not available" >&2
        return 1
    fi

    locate "$name" 2>/dev/null || true
}

# search.which <command>
# Find executable in PATH
search.which() {
    local cmd="$1"
    which "$cmd" 2>/dev/null
}

# search.before <pattern> <path> <lines>
# Grep with N lines before match
search.before() {
    local pattern="$1"
    local path="${2:-.}"
    local lines="${3:-3}"

    if [[ ! -e "$path" ]]; then
        echo "Error: Path not found: $path" >&2
        return 1
    fi

    grep -B "$lines" -r "$pattern" "$path" 2>/dev/null || true
}

# search.after <pattern> <path> <lines>
# Grep with N lines after match
search.after() {
    local pattern="$1"
    local path="${2:-.}"
    local lines="${3:-3}"

    if [[ ! -e "$path" ]]; then
        echo "Error: Path not found: $path" >&2
        return 1
    fi

    grep -A "$lines" -r "$pattern" "$path" 2>/dev/null || true
}

# search.around <pattern> <path> <lines>
# Grep with N lines before and after match
search.around() {
    local pattern="$1"
    local path="${2:-.}"
    local lines="${3:-3}"

    if [[ ! -e "$path" ]]; then
        echo "Error: Path not found: $path" >&2
        return 1
    fi

    grep -C "$lines" -r "$pattern" "$path" 2>/dev/null || true
}

# search.files_with <pattern> <path>
# List only filenames containing pattern (no match lines)
search.files_with() {
    local pattern="$1"
    local path="${2:-.}"

    if [[ ! -e "$path" ]]; then
        echo "Error: Path not found: $path" >&2
        return 1
    fi

    grep -rl "$pattern" "$path" 2>/dev/null | sort
}

# search.case_insensitive <pattern> <path>
# Case-insensitive grep
search.case_insensitive() {
    local pattern="$1"
    local path="${2:-.}"

    if [[ ! -e "$path" ]]; then
        echo "Error: Path not found: $path" >&2
        return 1
    fi

    grep -ri "$pattern" "$path" 2>/dev/null || true
}

# search.exact <pattern> <path>
# Exact word match (grep -w)
search.exact() {
    local pattern="$1"
    local path="${2:-.}"

    if [[ ! -e "$path" ]]; then
        echo "Error: Path not found: $path" >&2
        return 1
    fi

    grep -rw "$pattern" "$path" 2>/dev/null || true
}

# search.line_numbers <pattern> <path>
# Grep with line numbers
search.line_numbers() {
    local pattern="$1"
    local path="${2:-.}"

    if [[ ! -e "$path" ]]; then
        echo "Error: Path not found: $path" >&2
        return 1
    fi

    grep -rn "$pattern" "$path" 2>/dev/null || true
}

# Main dispatcher
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -lt 1 ]]; then
        echo "Usage: search.sh <operation> [args...]" >&2
        echo "Operations:" >&2
        echo "  grep, count, find, locate, which" >&2
        echo "  before, after, around" >&2
        echo "  files_with, case_insensitive, exact, line_numbers" >&2
        exit 1
    fi

    operation="$1"
    shift

    case "$operation" in
        grep|count|find|locate|which|before|after|around|files_with|case_insensitive|exact|line_numbers)
            "search.$operation" "$@"
            ;;
        *)
            echo "Error: Unknown operation: $operation" >&2
            exit 1
            ;;
    esac
fi
