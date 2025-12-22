#!/usr/bin/env bash
# data.sh - Data processing utility operations
# Composes primitives for CSV, text, and data manipulation
# Exit codes: 0 = success, 1 = error

set -euo pipefail

# Get script directory and source primitives
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRIMITIVES_DIR="$(dirname "$SCRIPT_DIR")/primitives"

source "$PRIMITIVES_DIR/file.sh"
source "$PRIMITIVES_DIR/sys.sh"

# ==============================================================================
# CSV OPERATIONS
# ==============================================================================

# util.csv.read <file> <row> <col>
# Get cell value from CSV (1-indexed)
util.csv.read() {
    local file="$1"
    local row="$2"
    local col="$3"

    if ! file.exists "$file"; then
        echo "Error: CSV file not found: $file" >&2
        return 1
    fi

    awk -F',' -v row="$row" -v col="$col" 'NR==row {print $col}' "$file"
}

# util.csv.write <file> <data>
# Append row to CSV (comma-separated values)
util.csv.write() {
    local file="$1"
    local data="$2"

    file.append "$file" "$data"
}

# util.csv.header <file>
# Get CSV header (first row)
util.csv.header() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: CSV file not found: $file" >&2
        return 1
    fi

    head -n 1 "$file"
}

# util.csv.count <file>
# Count rows in CSV (excluding header)
util.csv.count() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: CSV file not found: $file" >&2
        return 1
    fi

    # Total lines minus header
    echo $(($(wc -l < "$file") - 1))
}

# util.csv.query <file> <col_name> <value>
# Find rows where column matches value
util.csv.query() {
    local file="$1"
    local col_name="$2"
    local value="$3"

    if ! file.exists "$file"; then
        echo "Error: CSV file not found: $file" >&2
        return 1
    fi

    # Get header to find column index
    local header
    header=$(util.csv.header "$file")

    # Find column index
    local col_index=0
    local i=1
    IFS=',' read -ra COLS <<< "$header"
    for col in "${COLS[@]}"; do
        if [[ "$col" == "$col_name" ]]; then
            col_index=$i
            break
        fi
        ((i++))
    done

    if [[ $col_index -eq 0 ]]; then
        echo "Error: Column not found: $col_name" >&2
        return 1
    fi

    # Search for matching rows
    awk -F',' -v col="$col_index" -v val="$value" '$col==val' "$file"
}

# util.csv.column <file> <col_number>
# Extract entire column (1-indexed)
util.csv.column() {
    local file="$1"
    local col_number="$2"

    if ! file.exists "$file"; then
        echo "Error: CSV file not found: $file" >&2
        return 1
    fi

    awk -F',' -v col="$col_number" '{print $col}' "$file"
}

# ==============================================================================
# TEXT OPERATIONS
# ==============================================================================

# util.text.trim <text>
# Remove leading/trailing whitespace
util.text.trim() {
    local text="$1"
    echo "$text" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
}

# util.text.upper <text>
# Convert to uppercase
util.text.upper() {
    local text="$1"
    echo "$text" | tr '[:lower:]' '[:upper:]'
}

# util.text.lower <text>
# Convert to lowercase
util.text.lower() {
    local text="$1"
    echo "$text" | tr '[:upper:]' '[:lower:]'
}

# util.text.replace <text> <old> <new>
# Replace all occurrences
util.text.replace() {
    local text="$1"
    local old="$2"
    local new="$3"
    echo "$text" | sed "s/${old}/${new}/g"
}

# util.text.lines <text>
# Count lines in text
util.text.lines() {
    local text="$1"
    echo "$text" | wc -l
}

# util.text.words <text>
# Count words in text
util.text.words() {
    local text="$1"
    echo "$text" | wc -w
}

# util.text.chars <text>
# Count characters in text
util.text.chars() {
    local text="$1"
    echo "$text" | wc -c
}

# util.text.substring <text> <start> <length>
# Extract substring (1-indexed)
util.text.substring() {
    local text="$1"
    local start="$2"
    local length="$3"
    echo "$text" | cut -c "${start}-$((start + length - 1))"
}

# util.text.split <text> <delimiter>
# Split text by delimiter (one per line)
util.text.split() {
    local text="$1"
    local delimiter="$2"
    echo "$text" | tr "$delimiter" '\n'
}

# util.text.join <delimiter> <line1> <line2> ...
# Join arguments with delimiter
util.text.join() {
    local delimiter="$1"
    shift
    local IFS="$delimiter"
    echo "$*"
}

# ==============================================================================
# HASHING & ENCODING
# ==============================================================================

# util.hash.md5 <file>
# Calculate MD5 checksum
util.hash.md5() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    # Cross-platform MD5
    if command -v md5sum &>/dev/null; then
        md5sum "$file" | awk '{print $1}'
    elif command -v md5 &>/dev/null; then
        md5 -q "$file"
    else
        echo "Error: No MD5 tool available" >&2
        return 1
    fi
}

# util.hash.sha256 <file>
# Calculate SHA256 checksum
util.hash.sha256() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    # Cross-platform SHA256
    if command -v sha256sum &>/dev/null; then
        sha256sum "$file" | awk '{print $1}'
    elif command -v shasum &>/dev/null; then
        shasum -a 256 "$file" | awk '{print $1}'
    else
        echo "Error: No SHA256 tool available" >&2
        return 1
    fi
}

# util.encode.base64 <text>
# Base64 encode text
util.encode.base64() {
    local text="$1"
    echo "$text" | base64
}

# util.decode.base64 <text>
# Base64 decode text
util.decode.base64() {
    local text="$1"
    echo "$text" | base64 -d
}

# util.encode.url <text>
# URL encode text
util.encode.url() {
    local text="$1"
    echo "$text" | jq -sRr @uri
}

# util.decode.url <text>
# URL decode text
util.decode.url() {
    local text="$1"
    printf '%b' "${text//%/\\x}"
}

# ==============================================================================
# DATA TRANSFORMATION
# ==============================================================================

# util.data.to_json <format> <file>
# Convert CSV to JSON
util.data.to_json() {
    local format="$1"
    local file="$2"

    if [[ "$format" != "csv" ]]; then
        echo "Error: Only CSV format supported currently" >&2
        return 1
    fi

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    # Simple CSV to JSON array conversion
    python3 -c "
import csv, json, sys
with open('$file') as f:
    reader = csv.DictReader(f)
    print(json.dumps(list(reader), indent=2))
" 2>/dev/null || {
        # Fallback if python not available
        echo "Error: Python3 required for CSV to JSON conversion" >&2
        return 1
    }
}

# util.data.stats <file>
# Show file statistics
util.data.stats() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    echo "File: $file"
    echo "Size: $(file.size "$file") bytes"
    echo "Lines: $(wc -l < "$file")"
    echo "Words: $(wc -w < "$file")"
    echo "Characters: $(wc -c < "$file")"
}

# Main dispatcher
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -lt 1 ]]; then
        echo "Usage: data.sh <operation> [args...]" >&2
        echo "" >&2
        echo "CSV Operations:" >&2
        echo "  csv.read, csv.write, csv.header, csv.count" >&2
        echo "  csv.query, csv.column" >&2
        echo "" >&2
        echo "Text Operations:" >&2
        echo "  text.trim, text.upper, text.lower, text.replace" >&2
        echo "  text.lines, text.words, text.chars" >&2
        echo "  text.substring, text.split, text.join" >&2
        echo "" >&2
        echo "Hashing & Encoding:" >&2
        echo "  hash.md5, hash.sha256" >&2
        echo "  encode.base64, decode.base64" >&2
        echo "  encode.url, decode.url" >&2
        echo "" >&2
        echo "Data Transformation:" >&2
        echo "  data.to_json, data.stats" >&2
        exit 1
    fi

    operation="$1"
    shift

    case "$operation" in
        csv.*|text.*|hash.*|encode.*|decode.*|data.*)
            "util.$operation" "$@"
            ;;
        *)
            echo "Error: Unknown operation: $operation" >&2
            exit 1
            ;;
    esac
fi
