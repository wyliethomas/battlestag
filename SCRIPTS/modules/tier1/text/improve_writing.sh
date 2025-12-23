#!/usr/bin/env bash
# improve_writing.sh - Text Improvement Module (Tier 1)
#
# MODULE: improve_writing
# NAME: Improve Writing
# CATEGORY: text
# TIER: 1
# DESCRIPTION: Improves text for grammar, clarity, and style using local LLM
# PARAM: input_file:path:Path to text file to improve
# PARAM: output_file:path:Path to save improved text (optional)

set -euo pipefail

# Source utilities from .bstag installation
BSTAG_HOME="${BSTAG_HOME:-$HOME/.bstag}"

source "$BSTAG_HOME/utilities/file.sh"
source "$BSTAG_HOME/utilities/log.sh"
source "$BSTAG_HOME/utilities/llm.sh"

# Module: improve_writing
# Improves text using local LLM (Tier 1 - fast, free)
module.improve_writing() {
    local input_file="${1:-}"
    local output_file="${2:-}"

    # Validation
    if [[ -z "$input_file" ]]; then
        echo "Error: Input file required" >&2
        echo "Usage: improve_writing.sh <input_file> [output_file]" >&2
        return 1
    fi

    if ! file.exists "$input_file"; then
        echo "Error: File not found: $input_file" >&2
        return 1
    fi

    # Set default output
    if [[ -z "$output_file" ]]; then
        output_file="${input_file}.improved.txt"
    fi

    util.log.info "Improving text from: $input_file"
    util.log.info "Using: Tier 1 (Local LLM - Free)"

    # Read input
    local input_text
    input_text=$(file.read "$input_file")

    # Check text length
    local char_count
    char_count=$(echo -n "$input_text" | wc -c)
    util.log.info "Input size: $char_count characters"

    # Build prompt for improvement
    local prompt="Please improve the following text by:
1. Fixing any grammar or spelling errors
2. Improving clarity and flow
3. Making it more concise without losing meaning
4. Keeping the same tone and style

Return ONLY the improved text, no explanations.

Text to improve:
---
$input_text
---

Improved text:"

    # Query local LLM (Tier 1)
    util.log.info "Processing with local LLM..."
    local improved_text
    improved_text=$(util.llm.query_small "$prompt" 3000)

    # Check if we got a response
    if [[ -z "$improved_text" ]] || [[ "$improved_text" == "Error:"* ]]; then
        echo "Error: Failed to improve text" >&2
        echo "Response: $improved_text" >&2
        return 1
    fi

    # Write output
    echo "$improved_text" > "$output_file"

    util.log.info "✓ Improved text saved to: $output_file"

    # Show summary
    local output_chars
    output_chars=$(echo -n "$improved_text" | wc -c)

    echo ""
    echo "Summary:"
    echo "  Input:     $char_count characters"
    echo "  Output:    $output_chars characters"
    echo "  Saved to:  $output_file"
    echo "  Tier:      1 (Local LLM)"
    echo "  Cost:      \$0.00"
    echo ""
    echo "✓ Done!"
}

# Execute if called directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    module.improve_writing "$@"
fi
