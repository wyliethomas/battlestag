#!/usr/bin/env bash
# json.sh - JSON utility operations
# Composes primitives with jq for JSON manipulation
# Exit codes: 0 = success, 1 = error

set -euo pipefail

# Get script directory and source primitives
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRIMITIVES_DIR="$(dirname "$SCRIPT_DIR")/primitives"

source "$PRIMITIVES_DIR/file.sh"
source "$PRIMITIVES_DIR/sys.sh"

# Check for jq
if ! command -v jq &>/dev/null; then
    echo "Error: jq is required but not installed" >&2
    echo "Install: sudo apt-get install jq" >&2
    exit 1
fi

# util.json.read <file> <path>
# Extract value from JSON file using jq path
# Example: util.json.read config.json '.database.host'
util.json.read() {
    local file="$1"
    local path="$2"

    if ! file.exists "$file"; then
        echo "Error: JSON file not found: $file" >&2
        return 1
    fi

    jq -r "$path" "$file"
}

# util.json.write <file> <path> <value>
# Update JSON field (creates file if doesn't exist)
util.json.write() {
    local file="$1"
    local path="$2"
    local value="$3"

    # Create empty JSON object if file doesn't exist
    if ! file.exists "$file"; then
        file.write "$file" "{}"
    fi

    # Update JSON
    local temp_file
    temp_file=$(mktemp)
    jq "$path = $value" "$file" > "$temp_file"
    file.copy "$temp_file" "$file"
    rm -f "$temp_file"
}

# util.json.write_string <file> <path> <string_value>
# Update JSON field with string value (handles quotes)
util.json.write_string() {
    local file="$1"
    local path="$2"
    local value="$3"

    # Create empty JSON object if file doesn't exist
    if ! file.exists "$file"; then
        file.write "$file" "{}"
    fi

    # Update JSON with string value
    local temp_file
    temp_file=$(mktemp)
    jq --arg val "$value" "$path = \$val" "$file" > "$temp_file"
    file.copy "$temp_file" "$file"
    rm -f "$temp_file"
}

# util.json.delete <file> <path>
# Remove field from JSON
util.json.delete() {
    local file="$1"
    local path="$2"

    if ! file.exists "$file"; then
        echo "Error: JSON file not found: $file" >&2
        return 1
    fi

    local temp_file
    temp_file=$(mktemp)
    jq "del($path)" "$file" > "$temp_file"
    file.copy "$temp_file" "$file"
    rm -f "$temp_file"
}

# util.json.merge <file1> <file2> <output>
# Merge two JSON files
util.json.merge() {
    local file1="$1"
    local file2="$2"
    local output="$3"

    if ! file.exists "$file1"; then
        echo "Error: JSON file not found: $file1" >&2
        return 1
    fi

    if ! file.exists "$file2"; then
        echo "Error: JSON file not found: $file2" >&2
        return 1
    fi

    jq -s '.[0] * .[1]' "$file1" "$file2" > "$output"
}

# util.json.validate <file>
# Check if file contains valid JSON
util.json.validate() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    if jq empty "$file" 2>/dev/null; then
        return 0
    else
        echo "Error: Invalid JSON in $file" >&2
        return 1
    fi
}

# util.json.query <file> <jq_expression>
# Run complex jq query on JSON file
util.json.query() {
    local file="$1"
    local expression="$2"

    if ! file.exists "$file"; then
        echo "Error: JSON file not found: $file" >&2
        return 1
    fi

    jq "$expression" "$file"
}

# util.json.keys <file> <path>
# List keys at given path
util.json.keys() {
    local file="$1"
    local path="${2:-.}"

    if ! file.exists "$file"; then
        echo "Error: JSON file not found: $file" >&2
        return 1
    fi

    jq -r "${path} | keys[]" "$file"
}

# util.json.array_length <file> <path>
# Get length of array at path
util.json.array_length() {
    local file="$1"
    local path="$2"

    if ! file.exists "$file"; then
        echo "Error: JSON file not found: $file" >&2
        return 1
    fi

    jq "${path} | length" "$file"
}

# util.json.array_append <file> <path> <value>
# Append value to array
util.json.array_append() {
    local file="$1"
    local path="$2"
    local value="$3"

    if ! file.exists "$file"; then
        echo "Error: JSON file not found: $file" >&2
        return 1
    fi

    local temp_file
    temp_file=$(mktemp)
    jq "${path} += [$value]" "$file" > "$temp_file"
    file.copy "$temp_file" "$file"
    rm -f "$temp_file"
}

# util.json.array_append_string <file> <path> <string_value>
# Append string value to array
util.json.array_append_string() {
    local file="$1"
    local path="$2"
    local value="$3"

    if ! file.exists "$file"; then
        echo "Error: JSON file not found: $file" >&2
        return 1
    fi

    local temp_file
    temp_file=$(mktemp)
    jq --arg val "$value" "${path} += [\$val]" "$file" > "$temp_file"
    file.copy "$temp_file" "$file"
    rm -f "$temp_file"
}

# util.json.pretty <file>
# Pretty-print JSON file (in-place)
util.json.pretty() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: JSON file not found: $file" >&2
        return 1
    fi

    local temp_file
    temp_file=$(mktemp)
    jq '.' "$file" > "$temp_file"
    file.copy "$temp_file" "$file"
    rm -f "$temp_file"
}

# util.json.compact <file>
# Compact JSON file (in-place)
util.json.compact() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: JSON file not found: $file" >&2
        return 1
    fi

    local temp_file
    temp_file=$(mktemp)
    jq -c '.' "$file" > "$temp_file"
    file.copy "$temp_file" "$file"
    rm -f "$temp_file"
}

# util.json.from_string <json_string>
# Parse JSON string and output pretty
util.json.from_string() {
    local json_string="$1"
    echo "$json_string" | jq '.'
}

# Main dispatcher
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -lt 1 ]]; then
        echo "Usage: json.sh <operation> [args...]" >&2
        echo "Operations:" >&2
        echo "  read, write, write_string, delete, merge, validate" >&2
        echo "  query, keys, array_length, array_append, array_append_string" >&2
        echo "  pretty, compact, from_string" >&2
        exit 1
    fi

    operation="$1"
    shift

    case "$operation" in
        read|write|write_string|delete|merge|validate|query|keys|array_length|array_append|array_append_string|pretty|compact|from_string)
            "util.json.$operation" "$@"
            ;;
        *)
            echo "Error: Unknown operation: $operation" >&2
            exit 1
            ;;
    esac
fi
