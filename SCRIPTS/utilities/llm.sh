#!/usr/bin/env bash
# llm.sh - LLM API utility operations
# Composes primitives for LLM API calls
# Exit codes: 0 = success, 1 = error

set -euo pipefail

# Get script directory and source primitives
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRIMITIVES_DIR="$(dirname "$SCRIPT_DIR")/primitives"

source "$PRIMITIVES_DIR/file.sh"
source "$PRIMITIVES_DIR/sys.sh"

# Check for curl
if ! command -v curl &>/dev/null; then
    echo "Error: curl is required but not installed" >&2
    echo "Install: sudo apt-get install curl" >&2
    exit 1
fi

# Default configuration
LLM_ENDPOINT="${LLM_ENDPOINT:-http://localhost:1234/v1}"
LLM_MODEL="${LLM_MODEL:-hermes-3-llama-3.1-8b}"
LLM_TIMEOUT="${LLM_TIMEOUT:-30}"
LLM_MAX_TOKENS="${LLM_MAX_TOKENS:-2000}"
LLM_TEMPERATURE="${LLM_TEMPERATURE:-0.7}"

# util.llm.query <prompt> [max_tokens]
# Send simple prompt to LLM
util.llm.query() {
    local prompt="$1"
    local max_tokens="${2:-$LLM_MAX_TOKENS}"

    local json_payload
    json_payload=$(cat <<EOF
{
  "model": "$LLM_MODEL",
  "messages": [
    {"role": "user", "content": "$prompt"}
  ],
  "max_tokens": $max_tokens,
  "temperature": $LLM_TEMPERATURE
}
EOF
)

    local response
    response=$(curl -s --max-time "$LLM_TIMEOUT" \
        -H "Content-Type: application/json" \
        -d "$json_payload" \
        "${LLM_ENDPOINT}/chat/completions")

    # Extract content from response
    echo "$response" | jq -r '.choices[0].message.content // .error.message // "Error: No response"'
}

# util.llm.with_context <file> <prompt>
# Query LLM with file content as context
util.llm.with_context() {
    local file="$1"
    local prompt="$2"

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    local content
    content=$(file.read "$file")

    # Escape content for JSON
    local escaped_content
    escaped_content=$(echo "$content" | jq -Rs .)

    local escaped_prompt
    escaped_prompt=$(echo "$prompt" | jq -Rs .)

    local json_payload
    json_payload=$(cat <<EOF
{
  "model": "$LLM_MODEL",
  "messages": [
    {"role": "user", "content": "Here is the file content:\n\n$escaped_content\n\n$escaped_prompt"}
  ],
  "max_tokens": $LLM_MAX_TOKENS,
  "temperature": $LLM_TEMPERATURE
}
EOF
)

    local response
    response=$(curl -s --max-time "$LLM_TIMEOUT" \
        -H "Content-Type: application/json" \
        -d "$json_payload" \
        "${LLM_ENDPOINT}/chat/completions")

    echo "$response" | jq -r '.choices[0].message.content // .error.message // "Error: No response"'
}

# util.llm.multi_file <file1> <file2> ... <prompt>
# Query LLM with multiple files as context
util.llm.multi_file() {
    if [[ $# -lt 2 ]]; then
        echo "Error: Need at least 1 file and a prompt" >&2
        return 1
    fi

    # Last argument is the prompt
    local prompt="${*: -1}"
    local files=("${@:1:$#-1}")

    # Build context from all files
    local context="Here are the files:\n\n"

    for file in "${files[@]}"; do
        if ! file.exists "$file"; then
            echo "Error: File not found: $file" >&2
            return 1
        fi

        local content
        content=$(file.read "$file")
        context="${context}File: ${file}\n---\n${content}\n\n"
    done

    context="${context}${prompt}"

    # Escape for JSON
    local escaped_context
    escaped_context=$(echo "$context" | jq -Rs .)

    local json_payload
    json_payload=$(cat <<EOF
{
  "model": "$LLM_MODEL",
  "messages": [
    {"role": "user", "content": $escaped_context}
  ],
  "max_tokens": $LLM_MAX_TOKENS,
  "temperature": $LLM_TEMPERATURE
}
EOF
)

    local response
    response=$(curl -s --max-time "$LLM_TIMEOUT" \
        -H "Content-Type: application/json" \
        -d "$json_payload" \
        "${LLM_ENDPOINT}/chat/completions")

    echo "$response" | jq -r '.choices[0].message.content // .error.message // "Error: No response"'
}

# util.llm.streaming <prompt>
# Send prompt with streaming response (not yet implemented fully)
util.llm.streaming() {
    local prompt="$1"

    # For now, just use regular query
    # TODO: Implement proper streaming with SSE
    util.llm.query "$prompt"
}

# util.llm.summarize <file>
# Summarize file content
util.llm.summarize() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    util.llm.with_context "$file" "Please provide a concise summary of this content."
}

# util.llm.explain <file>
# Explain code in file
util.llm.explain() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    util.llm.with_context "$file" "Please explain what this code does, how it works, and any important details."
}

# util.llm.review <file>
# Review code for issues
util.llm.review() {
    local file="$1"

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    util.llm.with_context "$file" "Please review this code for: 1) Bugs or logic errors, 2) Security issues, 3) Performance problems, 4) Code quality improvements. Be specific and provide line references where possible."
}

# util.llm.test_connection
# Test if LLM endpoint is reachable
util.llm.test_connection() {
    local response
    response=$(curl -s --max-time 5 "${LLM_ENDPOINT}/models" 2>/dev/null)

    if [[ -n "$response" ]]; then
        echo "✓ LLM endpoint reachable: ${LLM_ENDPOINT}"
        echo "Models available:"
        echo "$response" | jq -r '.data[].id' 2>/dev/null || echo "  (Unable to parse models)"
        return 0
    else
        echo "✗ LLM endpoint not reachable: ${LLM_ENDPOINT}" >&2
        return 1
    fi
}

# util.llm.config
# Show current LLM configuration
util.llm.config() {
    echo "LLM Configuration:"
    echo "  Endpoint:    $LLM_ENDPOINT"
    echo "  Model:       $LLM_MODEL"
    echo "  Timeout:     ${LLM_TIMEOUT}s"
    echo "  Max tokens:  $LLM_MAX_TOKENS"
    echo "  Temperature: $LLM_TEMPERATURE"
}

# util.llm.models
# List available models
util.llm.models() {
    local response
    response=$(curl -s --max-time 5 "${LLM_ENDPOINT}/models")

    if [[ -n "$response" ]]; then
        echo "$response" | jq -r '.data[].id'
    else
        echo "Error: Could not fetch models" >&2
        return 1
    fi
}

# Main dispatcher
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -lt 1 ]]; then
        echo "Usage: llm.sh <operation> [args...]" >&2
        echo "Operations:" >&2
        echo "  query <prompt>                - Simple query" >&2
        echo "  with_context <file> <prompt>  - Query with file context" >&2
        echo "  multi_file <files...> <prompt> - Query with multiple files" >&2
        echo "  summarize <file>              - Summarize file" >&2
        echo "  explain <file>                - Explain code" >&2
        echo "  review <file>                 - Review code" >&2
        echo "  test_connection               - Test LLM endpoint" >&2
        echo "  config                        - Show configuration" >&2
        echo "  models                        - List available models" >&2
        echo "" >&2
        echo "Environment variables:" >&2
        echo "  LLM_ENDPOINT    - API endpoint (default: http://localhost:1234/v1)" >&2
        echo "  LLM_MODEL       - Model name (default: hermes-3-llama-3.1-8b)" >&2
        echo "  LLM_TIMEOUT     - Request timeout (default: 30)" >&2
        echo "  LLM_MAX_TOKENS  - Max response tokens (default: 2000)" >&2
        echo "  LLM_TEMPERATURE - Temperature (default: 0.7)" >&2
        exit 1
    fi

    operation="$1"
    shift

    case "$operation" in
        query|with_context|multi_file|streaming|summarize|explain|review|test_connection|config|models)
            "util.llm.$operation" "$@"
            ;;
        *)
            echo "Error: Unknown operation: $operation" >&2
            exit 1
            ;;
    esac
fi
