#!/usr/bin/env bash
# test_primitives.sh - Test suite for primitive operations
# Tests all primitives to ensure they work correctly

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL=0
PASSED=0
FAILED=0

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source primitives
source "$SCRIPT_DIR/file.sh"
source "$SCRIPT_DIR/dir.sh"
source "$SCRIPT_DIR/search.sh"
source "$SCRIPT_DIR/sys.sh"

# Test output directory
TEST_DIR="/tmp/primitive_tests_$$"

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

# ============================================================================
# FILE.SH TESTS
# ============================================================================

test_file_operations() {
    echo -e "\n${YELLOW}=== Testing file.sh ===${NC}"

    # Test file.write
    test_case "file.write creates file with content"
    file.write "$TEST_DIR/test1.txt" "Hello World"
    if [[ -f "$TEST_DIR/test1.txt" ]] && [[ "$(cat "$TEST_DIR/test1.txt")" == "Hello World" ]]; then
        pass
    else
        fail "File not created or content incorrect"
    fi

    # Test file.read
    test_case "file.read returns file contents"
    local content
    content=$(file.read "$TEST_DIR/test1.txt")
    if [[ "$content" == "Hello World" ]]; then
        pass
    else
        fail "Content mismatch: got '$content'"
    fi

    # Test file.append
    test_case "file.append adds to file"
    file.append "$TEST_DIR/test1.txt" "Second line"
    local lines
    lines=$(wc -l < "$TEST_DIR/test1.txt")
    if [[ "$lines" -eq 2 ]]; then
        pass
    else
        fail "Expected 2 lines, got $lines"
    fi

    # Test file.exists
    test_case "file.exists returns true for existing file"
    if file.exists "$TEST_DIR/test1.txt"; then
        pass
    else
        fail "file.exists returned false for existing file"
    fi

    test_case "file.exists returns false for non-existing file"
    if ! file.exists "$TEST_DIR/nonexistent.txt"; then
        pass
    else
        fail "file.exists returned true for non-existing file"
    fi

    # Test file.size
    test_case "file.size returns correct size"
    local size
    size=$(file.size "$TEST_DIR/test1.txt")
    if [[ "$size" -gt 0 ]]; then
        pass
    else
        fail "Size should be > 0, got $size"
    fi

    # Test file.copy
    test_case "file.copy duplicates file"
    file.copy "$TEST_DIR/test1.txt" "$TEST_DIR/test1_copy.txt"
    if [[ -f "$TEST_DIR/test1_copy.txt" ]] && cmp -s "$TEST_DIR/test1.txt" "$TEST_DIR/test1_copy.txt"; then
        pass
    else
        fail "Copy failed or contents differ"
    fi

    # Test file.move
    test_case "file.move renames file"
    file.move "$TEST_DIR/test1_copy.txt" "$TEST_DIR/test1_moved.txt"
    if [[ -f "$TEST_DIR/test1_moved.txt" ]] && [[ ! -f "$TEST_DIR/test1_copy.txt" ]]; then
        pass
    else
        fail "Move failed"
    fi

    # Test file.permissions
    test_case "file.permissions changes file mode"
    file.permissions "$TEST_DIR/test1.txt" "600"
    local perms
    perms=$(stat -c "%a" "$TEST_DIR/test1.txt" 2>/dev/null || stat -f "%Lp" "$TEST_DIR/test1.txt")
    if [[ "$perms" == "600" ]]; then
        pass
    else
        fail "Expected 600, got $perms"
    fi

    # Test file.delete
    test_case "file.delete removes file"
    file.delete "$TEST_DIR/test1_moved.txt"
    if [[ ! -f "$TEST_DIR/test1_moved.txt" ]]; then
        pass
    else
        fail "File still exists after delete"
    fi
}

# ============================================================================
# DIR.SH TESTS
# ============================================================================

test_dir_operations() {
    echo -e "\n${YELLOW}=== Testing dir.sh ===${NC}"

    # Test dir.create
    test_case "dir.create makes directory"
    dir.create "$TEST_DIR/subdir"
    if [[ -d "$TEST_DIR/subdir" ]]; then
        pass
    else
        fail "Directory not created"
    fi

    # Test dir.exists
    test_case "dir.exists returns true for existing dir"
    if dir.exists "$TEST_DIR/subdir"; then
        pass
    else
        fail "dir.exists returned false"
    fi

    # Test dir.list
    test_case "dir.list shows files"
    touch "$TEST_DIR/subdir/file1.txt"
    touch "$TEST_DIR/subdir/file2.txt"
    local count
    count=$(dir.list "$TEST_DIR/subdir" | wc -l)
    if [[ "$count" -eq 2 ]]; then
        pass
    else
        fail "Expected 2 files, got $count"
    fi

    # Test dir.empty
    test_case "dir.empty detects non-empty directory"
    if ! dir.empty "$TEST_DIR/subdir"; then
        pass
    else
        fail "Should report non-empty"
    fi

    test_case "dir.empty detects empty directory"
    dir.create "$TEST_DIR/emptydir"
    if dir.empty "$TEST_DIR/emptydir"; then
        pass
    else
        fail "Should report empty"
    fi

    # Test dir.count
    test_case "dir.count returns correct count"
    local count
    count=$(dir.count "$TEST_DIR/subdir")
    if [[ "$count" -eq 2 ]]; then
        pass
    else
        fail "Expected 2, got $count"
    fi

    # Test dir.parent
    test_case "dir.parent returns parent path"
    local parent
    parent=$(dir.parent "$TEST_DIR/subdir")
    if [[ "$parent" == "$TEST_DIR" ]]; then
        pass
    else
        fail "Expected $TEST_DIR, got $parent"
    fi

    # Test dir.basename
    test_case "dir.basename returns directory name"
    local name
    name=$(dir.basename "$TEST_DIR/subdir")
    if [[ "$name" == "subdir" ]]; then
        pass
    else
        fail "Expected 'subdir', got '$name'"
    fi

    # Test dir.tree
    test_case "dir.tree lists recursively"
    dir.create "$TEST_DIR/subdir/nested"
    touch "$TEST_DIR/subdir/nested/file3.txt"
    local tree_count
    tree_count=$(dir.tree "$TEST_DIR/subdir" 2 | wc -l)
    if [[ "$tree_count" -gt 3 ]]; then
        pass
    else
        fail "Expected > 3 entries, got $tree_count"
    fi

    # Test dir.delete (last, it's destructive)
    test_case "dir.delete removes directory"
    dir.delete "$TEST_DIR/emptydir"
    if [[ ! -d "$TEST_DIR/emptydir" ]]; then
        pass
    else
        fail "Directory still exists"
    fi
}

# ============================================================================
# SEARCH.SH TESTS
# ============================================================================

test_search_operations() {
    echo -e "\n${YELLOW}=== Testing search.sh ===${NC}"

    # Setup test files
    mkdir -p "$TEST_DIR/search_test"
    echo "The quick brown fox" > "$TEST_DIR/search_test/file1.txt"
    echo "jumps over the lazy dog" > "$TEST_DIR/search_test/file2.txt"
    echo "The FOX is quick" > "$TEST_DIR/search_test/file3.txt"

    # Test search.grep
    test_case "search.grep finds pattern"
    local results
    results=$(search.grep "quick" "$TEST_DIR/search_test" | wc -l)
    if [[ "$results" -ge 1 ]]; then
        pass
    else
        fail "Pattern not found"
    fi

    # Test search.case_insensitive
    test_case "search.case_insensitive finds FOX and fox"
    local results
    results=$(search.case_insensitive "fox" "$TEST_DIR/search_test" | wc -l)
    if [[ "$results" -ge 2 ]]; then
        pass
    else
        fail "Expected 2+ matches, got $results"
    fi

    # Test search.files_with
    test_case "search.files_with lists matching files"
    local file_count
    file_count=$(search.files_with "quick" "$TEST_DIR/search_test" | wc -l)
    if [[ "$file_count" -eq 2 ]]; then
        pass
    else
        fail "Expected 2 files, got $file_count"
    fi

    # Test search.find
    test_case "search.find locates files by name"
    local found
    found=$(search.find "file1.txt" "$TEST_DIR/search_test" | wc -l)
    if [[ "$found" -eq 1 ]]; then
        pass
    else
        fail "Expected 1 file, got $found"
    fi

    # Test search.which
    test_case "search.which finds bash executable"
    local bash_path
    bash_path=$(search.which bash)
    if [[ -x "$bash_path" ]]; then
        pass
    else
        fail "bash not found in PATH"
    fi

    # Test search.line_numbers
    test_case "search.line_numbers includes line numbers"
    local output
    output=$(search.line_numbers "quick" "$TEST_DIR/search_test/file1.txt")
    if [[ "$output" == "1:"* ]]; then
        pass
    else
        fail "Line number not found in output: $output"
    fi
}

# ============================================================================
# SYS.SH TESTS
# ============================================================================

test_sys_operations() {
    echo -e "\n${YELLOW}=== Testing sys.sh ===${NC}"

    # Test sys.hostname
    test_case "sys.hostname returns hostname"
    local hostname
    hostname=$(sys.hostname)
    if [[ -n "$hostname" ]]; then
        pass
    else
        fail "Hostname empty"
    fi

    # Test sys.user
    test_case "sys.user returns username"
    local user
    user=$(sys.user)
    if [[ -n "$user" ]]; then
        pass
    else
        fail "Username empty"
    fi

    # Test sys.date
    test_case "sys.date returns ISO 8601 date"
    local date
    date=$(sys.date)
    if [[ "$date" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T ]]; then
        pass
    else
        fail "Date format incorrect: $date"
    fi

    # Test sys.timestamp
    test_case "sys.timestamp returns Unix timestamp"
    local ts
    ts=$(sys.timestamp)
    if [[ "$ts" =~ ^[0-9]+$ ]] && [[ "$ts" -gt 1700000000 ]]; then
        pass
    else
        fail "Invalid timestamp: $ts"
    fi

    # Test sys.pid
    test_case "sys.pid returns process ID"
    local pid
    pid=$(sys.pid)
    if [[ "$pid" =~ ^[0-9]+$ ]]; then
        pass
    else
        fail "Invalid PID: $pid"
    fi

    # Test sys.os
    test_case "sys.os returns OS name"
    local os
    os=$(sys.os)
    if [[ -n "$os" ]]; then
        pass
    else
        fail "OS name empty"
    fi

    # Test sys.arch
    test_case "sys.arch returns architecture"
    local arch
    arch=$(sys.arch)
    if [[ -n "$arch" ]]; then
        pass
    else
        fail "Architecture empty"
    fi

    # Test sys.home
    test_case "sys.home returns home directory"
    local home
    home=$(sys.home)
    if [[ -d "$home" ]]; then
        pass
    else
        fail "Home not a directory: $home"
    fi

    # Test sys.temp
    test_case "sys.temp returns temp directory"
    local temp
    temp=$(sys.temp)
    if [[ -d "$temp" ]]; then
        pass
    else
        fail "Temp not a directory: $temp"
    fi

    # Test sys.cpu_count
    test_case "sys.cpu_count returns CPU count"
    local cpus
    cpus=$(sys.cpu_count)
    if [[ "$cpus" =~ ^[0-9]+$ ]] && [[ "$cpus" -gt 0 ]]; then
        pass
    else
        fail "Invalid CPU count: $cpus"
    fi

    # Test sys.random
    test_case "sys.random generates number in range"
    local rand
    rand=$(sys.random 1 10)
    if [[ "$rand" -ge 1 ]] && [[ "$rand" -le 10 ]]; then
        pass
    else
        fail "Random out of range: $rand"
    fi

    # Test sys.uuid
    test_case "sys.uuid generates UUID"
    local uuid
    uuid=$(sys.uuid)
    if [[ "$uuid" =~ ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ ]]; then
        pass
    else
        fail "Invalid UUID format: $uuid"
    fi

    # Test sys.setenv and sys.env
    test_case "sys.setenv and sys.env work together"
    sys.setenv "TEST_VAR_123" "test_value"
    local val
    val=$(sys.env "TEST_VAR_123")
    if [[ "$val" == "test_value" ]]; then
        pass
    else
        fail "Expected 'test_value', got '$val'"
    fi

    # Test sys.sleep
    test_case "sys.sleep waits correct time"
    local start
    start=$(sys.timestamp)
    sys.sleep 1
    local end
    end=$(sys.timestamp)
    local diff=$((end - start))
    if [[ "$diff" -ge 1 ]]; then
        pass
    else
        fail "Sleep didn't wait long enough: ${diff}s"
    fi
}

# ============================================================================
# MAIN
# ============================================================================

main() {
    echo -e "${YELLOW}╔════════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║   Primitive Operations Test Suite         ║${NC}"
    echo -e "${YELLOW}╚════════════════════════════════════════════╝${NC}"

    setup

    # Run all test suites
    test_file_operations
    test_dir_operations
    test_search_operations
    test_sys_operations

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
