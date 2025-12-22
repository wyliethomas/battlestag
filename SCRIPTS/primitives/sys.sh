#!/usr/bin/env bash
# sys.sh - System operation primitives
# Pure bash system operations
# Exit codes: 0 = success, 1 = error
# Output: stdout for data, stderr for errors

set -euo pipefail

# sys.exec <command> [args...]
# Execute arbitrary command
sys.exec() {
    if [[ $# -lt 1 ]]; then
        echo "Error: No command specified" >&2
        return 1
    fi

    "$@"
}

# sys.env <var>
# Get environment variable value
sys.env() {
    local var="$1"

    if [[ -z "${!var:-}" ]]; then
        echo "Error: Environment variable not set: $var" >&2
        return 1
    fi

    echo "${!var}"
}

# sys.setenv <var> <value>
# Set environment variable (for current session)
sys.setenv() {
    local var="$1"
    local value="$2"

    export "$var=$value"
}

# sys.hostname
# Get system hostname
sys.hostname() {
    hostname
}

# sys.user
# Get current user
sys.user() {
    whoami
}

# sys.date
# Get current date/time (ISO 8601 format)
sys.date() {
    date -u +"%Y-%m-%dT%H:%M:%SZ"
}

# sys.timestamp
# Get Unix timestamp (seconds since epoch)
sys.timestamp() {
    date +%s
}

# sys.uptime
# Get system uptime
sys.uptime() {
    uptime -p 2>/dev/null || uptime
}

# sys.pid
# Get current process ID
sys.pid() {
    echo $$
}

# sys.ppid
# Get parent process ID
sys.ppid() {
    echo $PPID
}

# sys.kill <pid> <signal>
# Kill process with signal (default: TERM)
sys.kill() {
    local pid="$1"
    local signal="${2:-TERM}"

    if ! kill -0 "$pid" 2>/dev/null; then
        echo "Error: Process not found: $pid" >&2
        return 1
    fi

    kill -s "$signal" "$pid"
}

# sys.ps <pattern>
# Find processes matching pattern
sys.ps() {
    local pattern="${1:-.}"
    ps aux | grep -v grep | grep "$pattern"
}

# sys.memory
# Get total memory in MB
sys.memory() {
    free -m 2>/dev/null | awk '/^Mem:/ {print $2}' || {
        # macOS fallback
        sysctl -n hw.memsize 2>/dev/null | awk '{print int($1/1024/1024)}'
    }
}

# sys.disk <path>
# Get disk usage for path (percentage)
sys.disk() {
    local path="${1:-.}"

    if [[ ! -e "$path" ]]; then
        echo "Error: Path not found: $path" >&2
        return 1
    fi

    df -h "$path" | awk 'NR==2 {print $5}' | tr -d '%'
}

# sys.cpu_count
# Get number of CPU cores
sys.cpu_count() {
    nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "1"
}

# sys.os
# Get operating system name
sys.os() {
    uname -s
}

# sys.arch
# Get system architecture
sys.arch() {
    uname -m
}

# sys.shell
# Get current shell
sys.shell() {
    echo "$SHELL"
}

# sys.home
# Get user home directory
sys.home() {
    echo "$HOME"
}

# sys.temp
# Get temp directory
sys.temp() {
    echo "${TMPDIR:-/tmp}"
}

# sys.sleep <seconds>
# Sleep for N seconds
sys.sleep() {
    local seconds="$1"
    sleep "$seconds"
}

# sys.random <min> <max>
# Generate random number between min and max
sys.random() {
    local min="${1:-0}"
    local max="${2:-100}"

    if [[ "$min" -gt "$max" ]]; then
        echo "Error: min must be <= max" >&2
        return 1
    fi

    local range=$((max - min + 1))
    echo $((RANDOM % range + min))
}

# sys.uuid
# Generate UUID
sys.uuid() {
    if command -v uuidgen &>/dev/null; then
        uuidgen | tr '[:upper:]' '[:lower:]'
    else
        cat /proc/sys/kernel/random/uuid 2>/dev/null || {
            echo "Error: UUID generation not available" >&2
            return 1
        }
    fi
}

# Main dispatcher
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -lt 1 ]]; then
        echo "Usage: sys.sh <operation> [args...]" >&2
        echo "Operations:" >&2
        echo "  exec, env, setenv, hostname, user, date, timestamp, uptime" >&2
        echo "  pid, ppid, kill, ps" >&2
        echo "  memory, disk, cpu_count, os, arch" >&2
        echo "  shell, home, temp, sleep, random, uuid" >&2
        exit 1
    fi

    operation="$1"
    shift

    case "$operation" in
        exec|env|setenv|hostname|user|date|timestamp|uptime|pid|ppid|kill|ps|memory|disk|cpu_count|os|arch|shell|home|temp|sleep|random|uuid)
            "sys.$operation" "$@"
            ;;
        *)
            echo "Error: Unknown operation: $operation" >&2
            exit 1
            ;;
    esac
fi
