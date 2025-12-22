#!/usr/bin/env bash
# log.sh - Logging utility operations
# Composes primitives for structured logging
# Exit codes: 0 = success, 1 = error

set -euo pipefail

# Get script directory and source primitives
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRIMITIVES_DIR="$(dirname "$SCRIPT_DIR")/primitives"

source "$PRIMITIVES_DIR/file.sh"
source "$PRIMITIVES_DIR/dir.sh"
source "$PRIMITIVES_DIR/sys.sh"

# Default log directory
LOG_DIR="${LOG_DIR:-${HOME}/.local/share/battlestag/logs}"

# util.log.init <log_file>
# Initialize log file (create directory if needed)
util.log.init() {
    local log_file="$1"
    local log_dir
    log_dir=$(dir.parent "$log_file")

    if ! dir.exists "$log_dir"; then
        dir.create "$log_dir"
    fi

    if ! file.exists "$log_file"; then
        file.write "$log_file" ""
    fi
}

# util.log.write <log_file> <level> <message>
# Write log entry with timestamp and level
util.log.write() {
    local log_file="$1"
    local level="$2"
    local message="$3"

    util.log.init "$log_file"

    local timestamp
    timestamp=$(sys.date)
    local hostname
    hostname=$(sys.hostname)
    local user
    user=$(sys.user)

    local log_entry="[${timestamp}] [${hostname}] [${user}] [${level}] ${message}"
    file.append "$log_file" "$log_entry"
}

# util.log.info <message> [log_file]
# Log info message
util.log.info() {
    local message="$1"
    local log_file="${2:-${LOG_DIR}/app.log}"

    util.log.write "$log_file" "INFO" "$message"
}

# util.log.error <message> [log_file]
# Log error message
util.log.error() {
    local message="$1"
    local log_file="${2:-${LOG_DIR}/app.log}"

    util.log.write "$log_file" "ERROR" "$message"
}

# util.log.warn <message> [log_file]
# Log warning message
util.log.warn() {
    local message="$1"
    local log_file="${2:-${LOG_DIR}/app.log}"

    util.log.write "$log_file" "WARN" "$message"
}

# util.log.debug <message> [log_file]
# Log debug message (only if DEBUG=1)
util.log.debug() {
    local message="$1"
    local log_file="${2:-${LOG_DIR}/app.log}"

    if [[ "${DEBUG:-0}" == "1" ]]; then
        util.log.write "$log_file" "DEBUG" "$message"
    fi
}

# util.log.tail <log_file> [lines]
# Show last N lines of log file
util.log.tail() {
    local log_file="$1"
    local lines="${2:-20}"

    if ! file.exists "$log_file"; then
        echo "Error: Log file not found: $log_file" >&2
        return 1
    fi

    tail -n "$lines" "$log_file"
}

# util.log.grep <log_file> <pattern>
# Search log file for pattern
util.log.grep() {
    local log_file="$1"
    local pattern="$2"

    if ! file.exists "$log_file"; then
        echo "Error: Log file not found: $log_file" >&2
        return 1
    fi

    grep "$pattern" "$log_file" || true
}

# util.log.rotate <log_file> [keep_count]
# Rotate log file (keep last N versions)
util.log.rotate() {
    local log_file="$1"
    local keep_count="${2:-10}"

    if ! file.exists "$log_file"; then
        return 0  # Nothing to rotate
    fi

    local timestamp
    timestamp=$(sys.timestamp)
    local rotated_file="${log_file}.${timestamp}"

    # Rotate current log
    file.move "$log_file" "$rotated_file"

    # Create new empty log
    file.write "$log_file" ""

    # Clean old rotated logs (keep only last N)
    local log_dir
    log_dir=$(dir.parent "$log_file")
    local log_name
    log_name=$(dir.basename "$log_file")

    # Find and remove old rotated logs beyond keep_count
    find "$log_dir" -name "${log_name}.*" -type f | sort -r | tail -n +$((keep_count + 1)) | xargs rm -f 2>/dev/null || true
}

# util.log.clean <log_dir> <days>
# Remove log files older than N days
util.log.clean() {
    local log_dir="$1"
    local days="${2:-30}"

    if ! dir.exists "$log_dir"; then
        echo "Error: Log directory not found: $log_dir" >&2
        return 1
    fi

    find "$log_dir" -name "*.log" -type f -mtime +"$days" -delete 2>/dev/null || true
    find "$log_dir" -name "*.log.*" -type f -mtime +"$days" -delete 2>/dev/null || true
}

# util.log.size <log_file>
# Get log file size in human-readable format
util.log.size() {
    local log_file="$1"

    if ! file.exists "$log_file"; then
        echo "Error: Log file not found: $log_file" >&2
        return 1
    fi

    local size_bytes
    size_bytes=$(file.size "$log_file")

    # Convert to human-readable
    if [[ $size_bytes -lt 1024 ]]; then
        echo "${size_bytes}B"
    elif [[ $size_bytes -lt 1048576 ]]; then
        echo "$((size_bytes / 1024))KB"
    elif [[ $size_bytes -lt 1073741824 ]]; then
        echo "$((size_bytes / 1048576))MB"
    else
        echo "$((size_bytes / 1073741824))GB"
    fi
}

# util.log.stats <log_file>
# Show log file statistics
util.log.stats() {
    local log_file="$1"

    if ! file.exists "$log_file"; then
        echo "Error: Log file not found: $log_file" >&2
        return 1
    fi

    echo "Log file: $log_file"
    echo "Size: $(util.log.size "$log_file")"
    echo "Lines: $(wc -l < "$log_file")"
    echo "Errors: $(grep -c "\[ERROR\]" "$log_file" || echo "0")"
    echo "Warnings: $(grep -c "\[WARN\]" "$log_file" || echo "0")"
    echo "Info: $(grep -c "\[INFO\]" "$log_file" || echo "0")"
}

# util.log.levels <log_file>
# Show count of each log level
util.log.levels() {
    local log_file="$1"

    if ! file.exists "$log_file"; then
        echo "Error: Log file not found: $log_file" >&2
        return 1
    fi

    echo "ERROR: $(grep -c "\[ERROR\]" "$log_file" || echo "0")"
    echo "WARN:  $(grep -c "\[WARN\]" "$log_file" || echo "0")"
    echo "INFO:  $(grep -c "\[INFO\]" "$log_file" || echo "0")"
    echo "DEBUG: $(grep -c "\[DEBUG\]" "$log_file" || echo "0")"
}

# util.log.today <log_file>
# Show only today's log entries
util.log.today() {
    local log_file="$1"

    if ! file.exists "$log_file"; then
        echo "Error: Log file not found: $log_file" >&2
        return 1
    fi

    local today
    today=$(date -u +"%Y-%m-%d")
    grep "$today" "$log_file" || true
}

# util.log.errors <log_file>
# Show only error entries
util.log.errors() {
    local log_file="$1"

    if ! file.exists "$log_file"; then
        echo "Error: Log file not found: $log_file" >&2
        return 1
    fi

    grep "\[ERROR\]" "$log_file" || true
}

# Main dispatcher
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -lt 1 ]]; then
        echo "Usage: log.sh <operation> [args...]" >&2
        echo "Operations:" >&2
        echo "  init, write, info, error, warn, debug" >&2
        echo "  tail, grep, rotate, clean, size, stats" >&2
        echo "  levels, today, errors" >&2
        exit 1
    fi

    operation="$1"
    shift

    case "$operation" in
        init|write|info|error|warn|debug|tail|grep|rotate|clean|size|stats|levels|today|errors)
            "util.log.$operation" "$@"
            ;;
        *)
            echo "Error: Unknown operation: $operation" >&2
            exit 1
            ;;
    esac
fi
