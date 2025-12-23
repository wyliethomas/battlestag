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

# Default configuration - Tier 1 (Local/Small Brain)
LLM_ENDPOINT="${LLM_ENDPOINT:-http://192.168.1.232:11434/v1}"
LLM_MODEL="${LLM_MODEL:-hermes3:8b}"
LLM_TIMEOUT="${LLM_TIMEOUT:-60}"
LLM_MAX_TOKENS="${LLM_MAX_TOKENS:-2000}"
LLM_TEMPERATURE="${LLM_TEMPERATURE:-0.7}"

# Tier 2 configuration (Cloud/Big Brain)
# Load from config file if exists
BSTAG_CONFIG="${BSTAG_HOME:-$HOME/.bstag}/config/llm.conf"
if [[ -f "$BSTAG_CONFIG" ]]; then
    source "$BSTAG_CONFIG"
fi

# Tier 2 provider selection (user chooses in config)
TIER2_PROVIDER="${TIER2_PROVIDER:-anthropic}"

# Anthropic (Claude) configuration
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
ANTHROPIC_ENDPOINT="${ANTHROPIC_ENDPOINT:-https://api.anthropic.com/v1}"
ANTHROPIC_MODEL="${ANTHROPIC_MODEL:-claude-sonnet-4-5-20250929}"

# OpenAI configuration
OPENAI_API_KEY="${OPENAI_API_KEY:-}"
OPENAI_ENDPOINT="${OPENAI_ENDPOINT:-https://api.openai.com/v1}"
OPENAI_MODEL="${OPENAI_MODEL:-gpt-4o}"

# Google (Gemini) configuration
GOOGLE_API_KEY="${GOOGLE_API_KEY:-}"
GOOGLE_ENDPOINT="${GOOGLE_ENDPOINT:-https://generativelanguage.googleapis.com/v1beta}"
GOOGLE_MODEL="${GOOGLE_MODEL:-gemini-1.5-pro}"

# Azure OpenAI configuration
AZURE_API_KEY="${AZURE_API_KEY:-}"
AZURE_ENDPOINT="${AZURE_ENDPOINT:-}"
AZURE_MODEL="${AZURE_MODEL:-gpt-4}"
AZURE_DEPLOYMENT="${AZURE_DEPLOYMENT:-}"

# OpenRouter configuration
OPENROUTER_API_KEY="${OPENROUTER_API_KEY:-}"
OPENROUTER_ENDPOINT="${OPENROUTER_ENDPOINT:-https://openrouter.ai/api/v1}"
OPENROUTER_MODEL="${OPENROUTER_MODEL:-anthropic/claude-sonnet-4-5}"

# Common Tier 2 settings
TIER2_MAX_TOKENS="${TIER2_MAX_TOKENS:-4096}"
TIER2_TEMPERATURE="${TIER2_TEMPERATURE:-0.7}"

# util.llm.query <prompt> [max_tokens]
# Send simple prompt to LLM
util.llm.query() {
    local prompt="$1"
    local max_tokens="${2:-$LLM_MAX_TOKENS}"

    # Build JSON payload with proper escaping using jq
    local json_payload
    json_payload=$(jq -n \
        --arg model "$LLM_MODEL" \
        --arg prompt "$prompt" \
        --argjson max_tokens "$max_tokens" \
        --argjson temperature "$LLM_TEMPERATURE" \
        '{
            model: $model,
            messages: [{role: "user", content: $prompt}],
            max_tokens: $max_tokens,
            temperature: $temperature
        }')

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

# ============================================
# TIER-SPECIFIC FUNCTIONS
# ============================================

# util.llm.query_small <prompt> [max_tokens]
# Tier 1: Use local LLM (free, fast)
util.llm.query_small() {
    # Just an alias for util.llm.query
    util.llm.query "$@"
}

# util.llm.query_big <prompt> [max_tokens]
# Tier 2: Use cloud API (paid, quality) - routes to configured provider
util.llm.query_big() {
    local prompt="$1"
    local max_tokens="${2:-$TIER2_MAX_TOKENS}"

    # Route to appropriate provider
    case "$TIER2_PROVIDER" in
        anthropic)
            util.llm._query_anthropic "$prompt" "$max_tokens"
            ;;
        openai)
            util.llm._query_openai "$prompt" "$max_tokens"
            ;;
        google)
            util.llm._query_google "$prompt" "$max_tokens"
            ;;
        azure)
            util.llm._query_azure "$prompt" "$max_tokens"
            ;;
        openrouter)
            util.llm._query_openrouter "$prompt" "$max_tokens"
            ;;
        *)
            echo "Error: Unknown Tier 2 provider: $TIER2_PROVIDER" >&2
            echo "Valid providers: anthropic, openai, google, azure, openrouter" >&2
            return 1
            ;;
    esac
}

# Provider-specific implementations (internal functions)

# util.llm._query_anthropic <prompt> <max_tokens>
util.llm._query_anthropic() {
    local prompt="$1"
    local max_tokens="$2"

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo "Error: ANTHROPIC_API_KEY not set" >&2
        echo "Set in ~/.bstag/config/llm.conf or environment" >&2
        return 1
    fi

    local json_payload
    json_payload=$(jq -n \
        --arg model "$ANTHROPIC_MODEL" \
        --arg prompt "$prompt" \
        --argjson max_tokens "$max_tokens" \
        '{
            model: $model,
            max_tokens: $max_tokens,
            messages: [
                {role: "user", content: $prompt}
            ]
        }')

    local response
    response=$(curl -s --max-time 60 \
        -H "Content-Type: application/json" \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -d "$json_payload" \
        "${ANTHROPIC_ENDPOINT}/messages")

    echo "$response" | jq -r '.content[0].text // .error.message // "Error: No response"'
}

# util.llm._query_openai <prompt> <max_tokens>
util.llm._query_openai() {
    local prompt="$1"
    local max_tokens="$2"

    if [[ -z "$OPENAI_API_KEY" ]]; then
        echo "Error: OPENAI_API_KEY not set" >&2
        echo "Set in ~/.bstag/config/llm.conf or environment" >&2
        return 1
    fi

    local json_payload
    json_payload=$(jq -n \
        --arg model "$OPENAI_MODEL" \
        --arg prompt "$prompt" \
        --argjson max_tokens "$max_tokens" \
        --argjson temperature "$TIER2_TEMPERATURE" \
        '{
            model: $model,
            messages: [{role: "user", content: $prompt}],
            max_tokens: $max_tokens,
            temperature: $temperature
        }')

    local response
    response=$(curl -s --max-time 60 \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $OPENAI_API_KEY" \
        -d "$json_payload" \
        "${OPENAI_ENDPOINT}/chat/completions")

    echo "$response" | jq -r '.choices[0].message.content // .error.message // "Error: No response"'
}

# util.llm._query_google <prompt> <max_tokens>
util.llm._query_google() {
    local prompt="$1"
    local max_tokens="$2"

    if [[ -z "$GOOGLE_API_KEY" ]]; then
        echo "Error: GOOGLE_API_KEY not set" >&2
        echo "Set in ~/.bstag/config/llm.conf or environment" >&2
        return 1
    fi

    # Google uses a different API format
    local json_payload
    json_payload=$(jq -n \
        --arg prompt "$prompt" \
        --argjson max_tokens "$max_tokens" \
        --argjson temperature "$TIER2_TEMPERATURE" \
        '{
            contents: [{
                parts: [{text: $prompt}]
            }],
            generationConfig: {
                maxOutputTokens: $max_tokens,
                temperature: $temperature
            }
        }')

    local response
    response=$(curl -s --max-time 60 \
        -H "Content-Type: application/json" \
        -d "$json_payload" \
        "${GOOGLE_ENDPOINT}/models/${GOOGLE_MODEL}:generateContent?key=${GOOGLE_API_KEY}")

    echo "$response" | jq -r '.candidates[0].content.parts[0].text // .error.message // "Error: No response"'
}

# util.llm._query_azure <prompt> <max_tokens>
util.llm._query_azure() {
    local prompt="$1"
    local max_tokens="$2"

    if [[ -z "$AZURE_API_KEY" ]]; then
        echo "Error: AZURE_API_KEY not set" >&2
        echo "Set in ~/.bstag/config/llm.conf or environment" >&2
        return 1
    fi

    if [[ -z "$AZURE_ENDPOINT" || -z "$AZURE_DEPLOYMENT" ]]; then
        echo "Error: AZURE_ENDPOINT and AZURE_DEPLOYMENT must be set" >&2
        echo "Set in ~/.bstag/config/llm.conf or environment" >&2
        return 1
    fi

    local json_payload
    json_payload=$(jq -n \
        --arg prompt "$prompt" \
        --argjson max_tokens "$max_tokens" \
        --argjson temperature "$TIER2_TEMPERATURE" \
        '{
            messages: [{role: "user", content: $prompt}],
            max_tokens: $max_tokens,
            temperature: $temperature
        }')

    local response
    response=$(curl -s --max-time 60 \
        -H "Content-Type: application/json" \
        -H "api-key: $AZURE_API_KEY" \
        -d "$json_payload" \
        "${AZURE_ENDPOINT}/openai/deployments/${AZURE_DEPLOYMENT}/chat/completions?api-version=2024-02-15-preview")

    echo "$response" | jq -r '.choices[0].message.content // .error.message // "Error: No response"'
}

# util.llm._query_openrouter <prompt> <max_tokens>
util.llm._query_openrouter() {
    local prompt="$1"
    local max_tokens="$2"

    if [[ -z "$OPENROUTER_API_KEY" ]]; then
        echo "Error: OPENROUTER_API_KEY not set" >&2
        echo "Set in ~/.bstag/config/llm.conf or environment" >&2
        return 1
    fi

    # OpenRouter uses OpenAI-compatible format
    local json_payload
    json_payload=$(jq -n \
        --arg model "$OPENROUTER_MODEL" \
        --arg prompt "$prompt" \
        --argjson max_tokens "$max_tokens" \
        --argjson temperature "$TIER2_TEMPERATURE" \
        '{
            model: $model,
            messages: [{role: "user", content: $prompt}],
            max_tokens: $max_tokens,
            temperature: $temperature
        }')

    local response
    response=$(curl -s --max-time 60 \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $OPENROUTER_API_KEY" \
        -H "HTTP-Referer: https://battlestag.com" \
        -H "X-Title: BATTLESTAG Module" \
        -d "$json_payload" \
        "${OPENROUTER_ENDPOINT}/chat/completions")

    echo "$response" | jq -r '.choices[0].message.content // .error.message // "Error: No response"'
}

# util.llm.with_context_big <file> <prompt>
# Tier 2: Query Claude with file context
util.llm.with_context_big() {
    local file="$1"
    local prompt="$2"

    if ! file.exists "$file"; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    local content
    content=$(file.read "$file")

    local full_prompt="Here is the file content:

$content

$prompt"

    util.llm.query_big "$full_prompt"
}

# util.llm.tier_info
# Show tier configuration
util.llm.tier_info() {
    echo "========================================="
    echo "BATTLESTAG LLM Configuration"
    echo "========================================="
    echo ""
    echo "Tier 1 (Small Brain - Local):"
    echo "  Endpoint:    $LLM_ENDPOINT"
    echo "  Model:       $LLM_MODEL"
    echo "  Max Tokens:  $LLM_MAX_TOKENS"
    echo "  Temperature: $LLM_TEMPERATURE"
    echo "  Cost:        Free"
    echo ""
    echo "Tier 2 (Big Brain - Cloud):"
    echo "  Provider:    $TIER2_PROVIDER"
    echo "  Max Tokens:  $TIER2_MAX_TOKENS"
    echo "  Temperature: $TIER2_TEMPERATURE"
    echo ""

    # Show status for each provider
    echo "  Configured Providers:"

    # Anthropic
    echo -n "    Anthropic (Claude):  "
    if [[ -n "$ANTHROPIC_API_KEY" ]]; then
        echo "✓ $ANTHROPIC_MODEL"
    else
        echo "✗ Not configured"
    fi

    # OpenAI
    echo -n "    OpenAI (GPT):        "
    if [[ -n "$OPENAI_API_KEY" ]]; then
        echo "✓ $OPENAI_MODEL"
    else
        echo "✗ Not configured"
    fi

    # Google
    echo -n "    Google (Gemini):     "
    if [[ -n "$GOOGLE_API_KEY" ]]; then
        echo "✓ $GOOGLE_MODEL"
    else
        echo "✗ Not configured"
    fi

    # Azure
    echo -n "    Azure OpenAI:        "
    if [[ -n "$AZURE_API_KEY" ]]; then
        echo "✓ $AZURE_MODEL (deployment: $AZURE_DEPLOYMENT)"
    else
        echo "✗ Not configured"
    fi

    # OpenRouter
    echo -n "    OpenRouter:          "
    if [[ -n "$OPENROUTER_API_KEY" ]]; then
        echo "✓ $OPENROUTER_MODEL"
    else
        echo "✗ Not configured"
    fi

    echo ""
    echo "  Cost: ~\$0.01-0.10 per call (varies by provider)"
    echo ""
    echo "Config file: $BSTAG_CONFIG"
    echo "========================================="
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
        echo "Tier-specific operations:" >&2
        echo "  query_small <prompt>          - Tier 1: Local LLM (free)" >&2
        echo "  query_big <prompt>            - Tier 2: Claude API (paid)" >&2
        echo "  with_context_big <file> <prompt> - Tier 2 with file context" >&2
        echo "  tier_info                     - Show tier configuration" >&2
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
        query|with_context|multi_file|streaming|summarize|explain|review|test_connection|config|models|query_small|query_big|with_context_big|tier_info)
            "util.llm.$operation" "$@"
            ;;
        *)
            echo "Error: Unknown operation: $operation" >&2
            exit 1
            ;;
    esac
fi
