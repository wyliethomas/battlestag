#!/usr/bin/env bash
# review_code_advanced.sh - Advanced Code Review Module (Tier 2)
#
# MODULE: review_code_advanced
# NAME: Code Review (Advanced)
# CATEGORY: code
# TIER: 2
# DESCRIPTION: Deep code review with architecture analysis, security threats, and refactoring suggestions
# PARAM: code_file:path:Path to code file to review
# PARAM: output_file:path:Path to save review (optional)

set -euo pipefail

# Source utilities from .bstag installation
BSTAG_HOME="${BSTAG_HOME:-$HOME/.bstag}"

source "$BSTAG_HOME/utilities/file.sh"
source "$BSTAG_HOME/utilities/log.sh"
source "$BSTAG_HOME/utilities/llm.sh"

# Module: review_code_advanced
# Deep code review using cloud LLM (Tier 2 - comprehensive analysis)
module.review_code_advanced() {
    local code_file="${1:-}"
    local output_file="${2:-}"

    # Validation
    if [[ -z "$code_file" ]]; then
        echo "Error: Code file required" >&2
        echo "Usage: review_code_advanced.sh <code_file> [output_file]" >&2
        return 1
    fi

    if ! file.exists "$code_file"; then
        echo "Error: File not found: $code_file" >&2
        return 1
    fi

    # Set default output
    if [[ -z "$output_file" ]]; then
        output_file="${code_file}.review-advanced.md"
    fi

    util.log.info "Advanced code review: $code_file"
    util.log.info "Using: Tier 2 (Cloud LLM - Deep Analysis)"

    # Read code
    local code_content
    code_content=$(file.read "$code_file")

    # Detect language from extension
    local file_ext="${code_file##*.}"
    local language="unknown"
    case "$file_ext" in
        sh|bash) language="Bash" ;;
        py) language="Python" ;;
        js) language="JavaScript" ;;
        ts) language="TypeScript" ;;
        go) language="Go" ;;
        rb) language="Ruby" ;;
        java) language="Java" ;;
        c|h) language="C" ;;
        cpp|cc|hpp) language="C++" ;;
        rs) language="Rust" ;;
        php) language="PHP" ;;
        *) language="$file_ext" ;;
    esac

    # Check code length
    local line_count
    line_count=$(echo "$code_content" | wc -l)
    util.log.info "File: $line_count lines of $language"

    # Tier 2 advanced review prompt (Fabric-inspired, Principal Engineer level)
    local prompt="You are a Principal-level software engineer. Review this $language code for:

1. CORRECTNESS - Logic errors, bugs, edge cases
2. SECURITY - OWASP Top 10, injections, data exposure
3. PERFORMANCE - Complexity, memory, optimization
4. READABILITY - Clarity, documentation, maintainability
5. BEST PRACTICES - Idioms, patterns, SOLID/DRY principles
6. ERROR HANDLING - Validation, logging, graceful failure

Format:

## Overall Assessment
[Summary of purpose and main concerns]

## Critical Issues
- **Issue** [Category]
  Problem: [description]
  Impact: [what could fail]
  Fix: [solution]

## High Priority Issues
[Same format]

## Medium Priority
[Same format]

## Positive Aspects
[What works well]

## Recommendations
[Top 3 improvements]

Code:
\`\`\`$language
$code_content
\`\`\`"

    # Query cloud LLM (Tier 2)
    util.log.info "Performing deep analysis with cloud LLM (30-60 seconds)..."
    local review
    review=$(util.llm.query_big "$prompt" 4096)

    # Check if we got a response
    if [[ -z "$review" ]] || [[ "$review" == "Error:"* ]]; then
        echo "Error: Failed to review code" >&2
        echo "Response: $review" >&2
        return 1
    fi

    # Build complete output with metadata
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Count issues by severity
    local critical_count high_count medium_count
    critical_count=$(echo "$review" | grep -c "Critical" 2>/dev/null || echo "0")
    high_count=$(echo "$review" | grep -c "High" 2>/dev/null || echo "0")
    medium_count=$(echo "$review" | grep -c "Medium" 2>/dev/null || echo "0")

    # Remove any whitespace
    critical_count=$(echo "$critical_count" | tr -d ' ')
    high_count=$(echo "$high_count" | tr -d ' ')
    medium_count=$(echo "$medium_count" | tr -d ' ')

    cat > "$output_file" << EOF
# Advanced Code Review Report

**File**: $(basename "$code_file")
**Language**: $language
**Lines**: $line_count
**Reviewed**: $timestamp
**Reviewer**: Principal-Level AI Code Review (Tier 2)
**Method**: Comprehensive 6-Dimension Analysis

---

$review

---

## Review Statistics

- **Critical Issues**: $critical_count
- **High Priority**: $high_count
- **Medium Priority**: $medium_count
- **Total Lines Reviewed**: $line_count

## Action Plan

1. **Immediate**: Address all Critical issues before deployment
2. **Sprint**: Resolve High Priority issues in current sprint
3. **Backlog**: Medium Priority issues for future iterations
4. **Learning**: Review suggested resources to prevent similar issues

## About This Review

This advanced review was performed using a cloud-based LLM with Principal Engineer-level analysis capabilities. The review covers:
- ✓ OWASP Top 10 security vulnerabilities
- ✓ Algorithmic complexity and performance
- ✓ SOLID principles and design patterns
- ✓ Production readiness assessment
- ✓ Architectural recommendations

For basic reviews (faster, free, private), use Tier 1: \`review_code.sh\`

---

*Generated by BATTLESTAG Advanced Code Review Module (Tier 2)*
*Cost: ~\$0.03-0.08 | Privacy: Cloud Provider*
EOF

    util.log.info "✓ Advanced review saved to: $output_file"

    # Show summary
    local total_issues=$((critical_count + high_count + medium_count))

    echo ""
    echo "Summary:"
    echo "  File:      $(basename "$code_file")"
    echo "  Language:  $language"
    echo "  Lines:     $line_count"
    echo "  Critical:  $critical_count issues"
    echo "  High:      $high_count issues"
    echo "  Medium:    $medium_count issues"
    echo "  Total:     $total_issues issues"
    echo "  Saved to:  $output_file"
    echo "  Tier:      2 (Cloud LLM - Advanced)"
    echo "  Cost:      ~\$0.03-0.08"
    echo ""
    echo "✓ Done! View with: cat $output_file"
    echo ""
    echo "Next: Address Critical issues immediately!"
}

# Execute if called directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    module.review_code_advanced "$@"
fi
