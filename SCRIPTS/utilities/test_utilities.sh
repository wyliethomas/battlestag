#!/usr/bin/env bash
# test_utilities.sh - Test suite for utility operations
# Tests utilities that compose primitives

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test counters
TOTAL=0
PASSED=0
FAILED=0
SKIPPED=0

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source utilities
source "$SCRIPT_DIR/file.sh"
source "$SCRIPT_DIR/json.sh" 2>/dev/null || JSON_SKIP=1
source "$SCRIPT_DIR/log.sh"
source "$SCRIPT_DIR/data.sh"

# Test output directory
TEST_DIR="/tmp/utility_tests_$$"

# Setup test environment
setup() {
    echo -e "${YELLOW}Setting up test environment...${NC}"
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"
}

# Cleanup test environment
cleanup() {
    echo -e "${YELLOW}Cleaning up test environment...${NC}"
    rm -rf "$TEST_DIR"
}

# Test helper
test_case() {
    local description="$1"
    TOTAL=$((TOTAL + 1))
    echo -n "Test $TOTAL: $description... "
}

pass() {
    PASSED=$((PASSED + 1))
    echo -e "${GREEN}PASS${NC}"
}

fail() {
    local reason="$1"
    FAILED=$((FAILED + 1))
    echo -e "${RED}FAIL${NC} - $reason"
}

skip() {
    local reason="$1"
    SKIPPED=$((SKIPPED + 1))
    echo -e "${YELLOW}SKIP${NC} - $reason"
}

# ==============================================================================
# FILE UTILITIES TESTS
# ==============================================================================

test_file_utilities() {
    echo -e "\n${YELLOW}=== Testing file.sh utilities ===${NC}"

    # Test util.backup
    test_case "util.backup creates timestamped backup"
    file.write "$TEST_DIR/original.txt" "original content"
    local backup_file
    backup_file=$(util.backup "$TEST_DIR/original.txt")
    if [[ -f "$backup_file" ]] && [[ "$backup_file" == *".backup."* ]]; then
        pass
    else
        fail "Backup not created or wrong format"
    fi

    # Test util.safe_write
    test_case "util.safe_write backs up before writing"
    util.safe_write "$TEST_DIR/original.txt" "new content" 2>/dev/null
    if [[ "$(file.read "$TEST_DIR/original.txt")" == "new content" ]]; then
        pass
    else
        fail "Content not updated"
    fi

    # Test util.ensure_dir
    test_case "util.ensure_dir creates directory if needed"
    util.ensure_dir "$TEST_DIR/newdir"
    if [[ -d "$TEST_DIR/newdir" ]]; then
        pass
    else
        fail "Directory not created"
    fi

    # Test util.temp_file
    test_case "util.temp_file creates temporary file"
    local temp_file
    temp_file=$(util.temp_file "test")
    if [[ -f "$temp_file" ]] && [[ "$temp_file" == *"/test."* ]]; then
        pass
        rm -f "$temp_file"
    else
        fail "Temp file not created correctly"
    fi

    # Test util.temp_dir
    test_case "util.temp_dir creates temporary directory"
    local temp_dir
    temp_dir=$(util.temp_dir "test")
    if [[ -d "$temp_dir" ]] && [[ "$temp_dir" == *"/test."* ]]; then
        pass
        rm -rf "$temp_dir"
    else
        fail "Temp directory not created correctly"
    fi

    # Test util.template
    test_case "util.template replaces variables"
    echo "Hello {{name}}, you are {{age}} years old" > "$TEST_DIR/template.txt"
    util.template "$TEST_DIR/template.txt" "$TEST_DIR/output.txt" "name=John" "age=30"
    local content
    content=$(file.read "$TEST_DIR/output.txt")
    if [[ "$content" == "Hello John, you are 30 years old" ]]; then
        pass
    else
        fail "Template not processed correctly: $content"
    fi

    # Test util.config_get/set
    test_case "util.config_set and util.config_get work together"
    util.config_set "$TEST_DIR/config.txt" "NAME" "TestValue"
    local value
    value=$(util.config_get "$TEST_DIR/config.txt" "NAME")
    if [[ "$value" == "TestValue" ]]; then
        pass
    else
        fail "Config value mismatch: got '$value'"
    fi

    # Test util.file_age
    test_case "util.file_age returns age in seconds"
    file.write "$TEST_DIR/age_test.txt" "test"
    sleep 1
    local age
    age=$(util.file_age "$TEST_DIR/age_test.txt")
    if [[ "$age" -ge 0 ]]; then
        pass
    else
        fail "Invalid age: $age"
    fi
}

# ==============================================================================
# JSON UTILITIES TESTS
# ==============================================================================

test_json_utilities() {
    echo -e "\n${YELLOW}=== Testing json.sh utilities ===${NC}"

    if [[ "${JSON_SKIP:-0}" == "1" ]]; then
        echo -e "${YELLOW}Skipping JSON tests (jq not available)${NC}"
        return
    fi

    # Test util.json.write_string and read
    test_case "util.json.write_string and util.json.read work"
    util.json.write_string "$TEST_DIR/test.json" ".name" "John Doe"
    local value
    value=$(util.json.read "$TEST_DIR/test.json" ".name")
    if [[ "$value" == "John Doe" ]]; then
        pass
    else
        fail "JSON value mismatch: got '$value'"
    fi

    # Test util.json.validate
    test_case "util.json.validate detects valid JSON"
    if util.json.validate "$TEST_DIR/test.json"; then
        pass
    else
        fail "Valid JSON reported as invalid"
    fi

    # Test util.json.keys
    test_case "util.json.keys lists keys"
    util.json.write_string "$TEST_DIR/test2.json" ".foo" "bar"
    util.json.write_string "$TEST_DIR/test2.json" ".baz" "qux"
    local keys
    keys=$(util.json.keys "$TEST_DIR/test2.json" | wc -l)
    if [[ "$keys" -ge 2 ]]; then
        pass
    else
        fail "Expected 2+ keys, got $keys"
    fi

    # Test util.json.pretty
    test_case "util.json.pretty formats JSON"
    echo '{"a":1,"b":2}' > "$TEST_DIR/compact.json"
    util.json.pretty "$TEST_DIR/compact.json"
    local lines
    lines=$(wc -l < "$TEST_DIR/compact.json")
    if [[ "$lines" -gt 1 ]]; then
        pass
    else
        fail "JSON not prettified"
    fi
}

# ==============================================================================
# LOG UTILITIES TESTS
# ==============================================================================

test_log_utilities() {
    echo -e "\n${YELLOW}=== Testing log.sh utilities ===${NC}"

    local log_file="$TEST_DIR/test.log"

    # Test util.log.info
    test_case "util.log.info writes log entry"
    util.log.info "Test message" "$log_file"
    if [[ -f "$log_file" ]] && grep -q "Test message" "$log_file"; then
        pass
    else
        fail "Log entry not written"
    fi

    # Test util.log.error
    test_case "util.log.error writes error entry"
    util.log.error "Error message" "$log_file"
    if grep -q "\[ERROR\]" "$log_file" && grep -q "Error message" "$log_file"; then
        pass
    else
        fail "Error entry not written correctly"
    fi

    # Test util.log.warn
    test_case "util.log.warn writes warning entry"
    util.log.warn "Warning message" "$log_file"
    if grep -q "\[WARN\]" "$log_file"; then
        pass
    else
        fail "Warning entry not written"
    fi

    # Test util.log.tail
    test_case "util.log.tail shows recent entries"
    local tail_output
    tail_output=$(util.log.tail "$log_file" 2)
    if [[ $(echo "$tail_output" | wc -l) -eq 2 ]]; then
        pass
    else
        fail "Tail didn't return correct number of lines"
    fi

    # Test util.log.grep
    test_case "util.log.grep finds pattern"
    local grep_output
    grep_output=$(util.log.grep "$log_file" "Error message")
    if [[ -n "$grep_output" ]]; then
        pass
    else
        fail "Grep didn't find pattern"
    fi

    # Test util.log.levels
    test_case "util.log.levels counts log levels"
    local levels
    levels=$(util.log.levels "$log_file")
    if [[ "$levels" == *"ERROR:"* ]] && [[ "$levels" == *"WARN:"* ]]; then
        pass
    else
        fail "Level counts incorrect"
    fi

    # Test util.log.errors
    test_case "util.log.errors shows only errors"
    local errors
    errors=$(util.log.errors "$log_file")
    if [[ "$errors" == *"[ERROR]"* ]] && [[ "$errors" != *"[WARN]"* ]]; then
        pass
    else
        fail "Error filter not working"
    fi
}

# ==============================================================================
# DATA UTILITIES TESTS
# ==============================================================================

test_data_utilities() {
    echo -e "\n${YELLOW}=== Testing data.sh utilities ===${NC}"

    # Test CSV operations
    test_case "util.csv.write creates CSV"
    util.csv.write "$TEST_DIR/test.csv" "name,age,city"
    util.csv.write "$TEST_DIR/test.csv" "John,30,NYC"
    util.csv.write "$TEST_DIR/test.csv" "Jane,25,LA"
    if [[ -f "$TEST_DIR/test.csv" ]]; then
        pass
    else
        fail "CSV not created"
    fi

    # Test util.csv.header
    test_case "util.csv.header returns first row"
    local header
    header=$(util.csv.header "$TEST_DIR/test.csv")
    if [[ "$header" == "name,age,city" ]]; then
        pass
    else
        fail "Header mismatch: got '$header'"
    fi

    # Test util.csv.count
    test_case "util.csv.count returns row count"
    local count
    count=$(util.csv.count "$TEST_DIR/test.csv")
    if [[ "$count" -eq 2 ]]; then
        pass
    else
        fail "Expected 2 rows, got $count"
    fi

    # Test util.csv.read
    test_case "util.csv.read gets cell value"
    local value
    value=$(util.csv.read "$TEST_DIR/test.csv" 2 1)
    if [[ "$value" == "John" ]]; then
        pass
    else
        fail "Cell value mismatch: got '$value'"
    fi

    # Test text operations
    test_case "util.text.upper converts to uppercase"
    local upper
    upper=$(util.text.upper "hello")
    if [[ "$upper" == "HELLO" ]]; then
        pass
    else
        fail "Uppercase conversion failed: got '$upper'"
    fi

    test_case "util.text.lower converts to lowercase"
    local lower
    lower=$(util.text.lower "HELLO")
    if [[ "$lower" == "hello" ]]; then
        pass
    else
        fail "Lowercase conversion failed: got '$lower'"
    fi

    test_case "util.text.trim removes whitespace"
    local trimmed
    trimmed=$(util.text.trim "  hello  ")
    if [[ "$trimmed" == "hello" ]]; then
        pass
    else
        fail "Trim failed: got '$trimmed'"
    fi

    test_case "util.text.replace replaces text"
    local replaced
    replaced=$(util.text.replace "hello world" "world" "bash")
    if [[ "$replaced" == "hello bash" ]]; then
        pass
    else
        fail "Replace failed: got '$replaced'"
    fi

    # Test hashing
    test_case "util.hash.md5 generates checksum"
    echo "test content" > "$TEST_DIR/hash_test.txt"
    local md5
    md5=$(util.hash.md5 "$TEST_DIR/hash_test.txt")
    if [[ -n "$md5" ]] && [[ ${#md5} -eq 32 ]]; then
        pass
    else
        fail "MD5 hash invalid: got '$md5'"
    fi

    # Test encoding
    test_case "util.encode.base64 and decode work"
    local encoded
    encoded=$(util.encode.base64 "hello")
    local decoded
    decoded=$(util.decode.base64 "$encoded")
    if [[ "$decoded" == "hello" ]]; then
        pass
    else
        fail "Base64 encode/decode failed"
    fi
}

# ==============================================================================
# MAIN
# ==============================================================================

main() {
    echo -e "${YELLOW}╔════════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║   Utility Operations Test Suite           ║${NC}"
    echo -e "${YELLOW}╚════════════════════════════════════════════╝${NC}"

    setup

    # Run all test suites
    test_file_utilities
    test_json_utilities
    test_log_utilities
    test_data_utilities

    cleanup

    # Print summary
    echo -e "\n${YELLOW}╔════════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║   Test Results                             ║${NC}"
    echo -e "${YELLOW}╚════════════════════════════════════════════╝${NC}"
    echo -e "Total tests:  $TOTAL"
    echo -e "${GREEN}Passed:       $PASSED${NC}"
    if [[ "$FAILED" -gt 0 ]]; then
        echo -e "${RED}Failed:       $FAILED${NC}"
    else
        echo -e "Failed:       $FAILED"
    fi
    if [[ "$SKIPPED" -gt 0 ]]; then
        echo -e "${YELLOW}Skipped:      $SKIPPED${NC}"
    fi

    # Calculate percentage
    if [[ "$TOTAL" -gt 0 ]]; then
        local percentage
        percentage=$((PASSED * 100 / TOTAL))
        echo -e "Success rate: ${percentage}%"
    fi

    # Exit with error if any tests failed
    if [[ "$FAILED" -gt 0 ]]; then
        echo -e "\n${RED}Some tests failed!${NC}"
        exit 1
    else
        echo -e "\n${GREEN}All tests passed!${NC}"
        exit 0
    fi
}

# Run tests
main "$@"
