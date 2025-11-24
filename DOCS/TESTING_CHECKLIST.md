# Testing Checklist for Battlestag Programs

**Comprehensive testing guide for CLI programs and API integrations**

This document provides testing procedures, checklists, and templates for validating new programs in the Battlestag ecosystem.

---

## Table of Contents

1. [Overview](#overview)
2. [Testing Levels](#testing-levels)
3. [CLI Program Testing](#cli-program-testing)
4. [API Wrapper Testing](#api-wrapper-testing)
5. [Integration Testing](#integration-testing)
6. [LLM Interaction Testing](#llm-interaction-testing)
7. [Test Script Templates](#test-script-templates)
8. [Deployment Verification](#deployment-verification)
9. [Troubleshooting Tests](#troubleshooting-tests)

---

## Overview

### Testing Philosophy

✅ **Test Early**: Test CLI programs before writing API wrappers
✅ **Test Often**: Test after every significant change
✅ **Test Thoroughly**: Cover happy paths, edge cases, and error conditions
✅ **Test Realistically**: Use actual LLM interactions in final testing

### Test Environment

Always use a separate test database to avoid corrupting production data:

```bash
export DB_PATH=/tmp/test-your-program.db
```

---

## Testing Levels

### Level 1: CLI Program Testing
- Individual executables work correctly
- Flags and arguments parse correctly
- Database operations succeed
- Exit codes are correct
- Output is clean and parseable

### Level 2: API Wrapper Testing
- Program registers successfully
- Info() returns complete metadata
- Execute() handles all commands
- Parameter extraction works
- Errors are handled gracefully

### Level 3: Integration Testing
- API routes to correct program
- CLI executables are found and execute
- Output flows back through API
- Concurrent requests don't conflict
- Database transactions are atomic

### Level 4: LLM Interaction Testing
- LLM can discover program
- LLM understands when to use it
- LLM extracts parameters correctly
- LLM interprets output correctly
- Natural language variations work

---

## CLI Program Testing

### Pre-Installation Checklist

Before running `install.sh`:

- [ ] All Go files compile without errors
- [ ] `go mod tidy` has been run
- [ ] No hardcoded paths in code
- [ ] Database schema is correct
- [ ] Exit codes package exists
- [ ] Config package exists

```bash
# Verify compilation
cd PROGRAMS/your-program
go build ./cmd/command1
go build ./cmd/command2
go clean  # Remove test binaries

# Check dependencies
go mod tidy
go mod verify
```

### Installation Testing

- [ ] `install.sh` runs without errors
- [ ] All executables created in `~/.local/bin`
- [ ] Wrapper scripts created (if applicable)
- [ ] Config directory created
- [ ] Database directory created
- [ ] Default data seeded (if applicable)

```bash
# Test installation
cd PROGRAMS/your-program
./install.sh

# Verify executables
ls -la ~/.local/bin/your_*

# Verify config
cat ~/.config/your-program/.env

# Verify PATH
which your_command1
```

### Command Testing Template

Test each command systematically:

#### 1. Help Messages

```bash
# All commands should support --help
your_command1 --help
your_command2 --help
```

**Expected**:
- Clear usage information
- List of all flags
- Examples (if provided)
- Exit code 0

#### 2. Missing Required Arguments

```bash
# Should fail gracefully with clear error
your_command1
```

**Expected**:
- Clear error message
- Exit code 1 (ArgsError)
- No changes to database

#### 3. Invalid Arguments

```bash
# Test with invalid values
your_command1 --id abc  # String instead of int
your_command1 --status invalid  # Invalid enum value
```

**Expected**:
- Clear validation error
- Exit code 1 (ArgsError)
- No changes to database

#### 4. Database Errors

```bash
# Test with invalid database path
export DB_PATH=/invalid/path/db.sqlite
your_command1 --id 1
```

**Expected**:
- Database error message
- Exit code 2 (DBError)
- No crash

#### 5. Not Found Cases

```bash
# Test with non-existent ID
your_command1 --id 99999
```

**Expected**:
- "Not found" message
- Exit code 3 (NotFound)
- No crash

#### 6. Happy Path

```bash
# Test successful operation
export DB_PATH=/tmp/test-db.db
your_command1 --name "Test" --value "123"
```

**Expected**:
- Success message
- Exit code 0
- Database updated correctly
- Output is clean (parseable by API wrapper)

### Database Testing

Test database operations:

```bash
# Set test database
export DB_PATH=/tmp/test-db.db
rm -f $DB_PATH  # Start fresh

# Test 1: Create operation
your_command_add --name "Item 1"
echo "Exit code: $?"

# Test 2: Verify creation
your_command_list | grep "Item 1"

# Test 3: Update operation
your_command_update --id 1 --name "Updated"

# Test 4: Verify update
your_command_list | grep "Updated"

# Test 5: Delete operation
your_command_delete --id 1

# Test 6: Verify deletion
your_command_list | grep "Item 1" && echo "FAIL: Not deleted" || echo "PASS: Deleted"

# Test 7: Inspect database
sqlite3 $DB_PATH "SELECT * FROM your_table;"

# Cleanup
rm -f $DB_PATH
```

### Concurrent Operation Testing

Test that concurrent operations don't corrupt database:

```bash
export DB_PATH=/tmp/test-concurrent.db
rm -f $DB_PATH

# Run 10 operations in parallel
for i in {1..10}; do
    your_command_add --name "Item $i" &
done
wait

# Verify all 10 created
count=$(your_command_list | wc -l)
if [ "$count" -eq 10 ]; then
    echo "PASS: All items created"
else
    echo "FAIL: Expected 10, got $count"
fi

rm -f $DB_PATH
```

### Edge Case Testing

Test boundary conditions:

```bash
# Empty strings
your_command_add --name ""

# Very long strings
your_command_add --name "$(printf 'A%.0s' {1..1000})"

# Special characters
your_command_add --name "Test's \"quoted\" value & symbols"

# Unicode
your_command_add --name "测试 🎉 café"

# Zero/negative numbers
your_command_update --id 0
your_command_update --id -1

# Large numbers
your_command_update --id 999999999
```

### CLI Testing Checklist

Complete checklist for CLI program:

- [ ] **Compilation**
  - [ ] All commands compile without errors
  - [ ] No unused imports
  - [ ] `go mod tidy` runs clean

- [ ] **Installation**
  - [ ] `install.sh` succeeds
  - [ ] All executables in PATH
  - [ ] Config directory created
  - [ ] Database directory created

- [ ] **Help Messages**
  - [ ] All commands support `--help`
  - [ ] Help text is clear and accurate

- [ ] **Argument Parsing**
  - [ ] Required arguments enforced
  - [ ] Optional arguments have defaults
  - [ ] Invalid arguments rejected with clear errors

- [ ] **Database Operations**
  - [ ] Create operations work
  - [ ] Read operations work
  - [ ] Update operations work
  - [ ] Delete operations work
  - [ ] Schema initializes correctly
  - [ ] Foreign keys enforced (if applicable)

- [ ] **Exit Codes**
  - [ ] Success: 0
  - [ ] Invalid args: 1
  - [ ] Database errors: 2
  - [ ] Not found: 3

- [ ] **Output Format**
  - [ ] Success output to stdout
  - [ ] Errors to stderr
  - [ ] Clean, parseable format
  - [ ] JSON valid (if applicable)

- [ ] **Edge Cases**
  - [ ] Empty strings handled
  - [ ] Special characters handled
  - [ ] Unicode handled
  - [ ] Zero/negative numbers handled
  - [ ] Large values handled

- [ ] **Concurrent Access**
  - [ ] Multiple simultaneous operations work
  - [ ] No database corruption
  - [ ] No race conditions

---

## API Wrapper Testing

### Compilation Testing

```bash
cd ORCHESTRATOR/agent-gateway

# Test compilation
go build -o agent-gateway .

# Check for errors
echo "Exit code: $?"
```

**Expected**: Exit code 0, no errors

### Registration Testing

```bash
# Start API in foreground to see logs
./agent-gateway

# Look for registration message
# Expected: "Registered program: your-program"
```

**Checklist**:
- [ ] No compilation errors
- [ ] Registration log message appears
- [ ] No "already registered" error
- [ ] No panics or crashes

### Info() Testing

Test program metadata:

```bash
# Query programs list
curl -H "X-API-Key: your-key" http://localhost:8080/api/programs/list | jq
```

**Verify**:
- [ ] Program appears in list
- [ ] ID is correct and unique
- [ ] Name is descriptive
- [ ] Description is comprehensive
- [ ] Category is correct
- [ ] All parameters listed
- [ ] Required parameters marked
- [ ] Parameter descriptions are clear

**Example expected output**:
```json
{
  "programs": [
    {
      "id": "your-program",
      "name": "Your Program",
      "description": "Detailed description of what the program does...",
      "category": "category",
      "parameters": [
        {
          "name": "command",
          "type": "string",
          "description": "Command to execute: 'list', 'add', 'update'",
          "required": true
        }
      ]
    }
  ]
}
```

### Execute() Testing

Test each command through the API:

#### Test 1: Missing Command

```bash
curl -X POST \
  -H "X-API-Key: your-key" \
  -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{}}' \
  http://localhost:8080/api/programs/execute
```

**Expected**: Error about missing command parameter

#### Test 2: Invalid Command

```bash
curl -X POST \
  -H "X-API-Key: your-key" \
  -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"invalid"}}' \
  http://localhost:8080/api/programs/execute
```

**Expected**: "unknown command: invalid" error

#### Test 3: Missing Required Parameters

```bash
curl -X POST \
  -H "X-API-Key: your-key" \
  -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"add"}}' \
  http://localhost:8080/api/programs/execute
```

**Expected**: Error about missing required parameter (e.g., "name required")

#### Test 4: Valid Command

```bash
curl -X POST \
  -H "X-API-Key: your-key" \
  -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"add","name":"Test"}}' \
  http://localhost:8080/api/programs/execute
```

**Expected**: Success response with output

### Parameter Extraction Testing

Test different parameter types:

```bash
# String parameter
curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"test","name":"value"}}' \
  http://localhost:8080/api/programs/execute

# Integer parameter
curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"test","id":123}}' \
  http://localhost:8080/api/programs/execute

# Boolean parameter
curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"test","flag":true}}' \
  http://localhost:8080/api/programs/execute

# Float as int (JSON default)
curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"test","id":123.0}}' \
  http://localhost:8080/api/programs/execute
```

### Case Sensitivity Testing

Test case normalization:

```bash
# Lowercase
curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"add","context":"property"}}' \
  http://localhost:8080/api/programs/execute

# Uppercase
curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"add","context":"PROPERTY"}}' \
  http://localhost:8080/api/programs/execute

# Mixed case
curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"add","context":"Property"}}' \
  http://localhost:8080/api/programs/execute
```

**Expected**: All three should work identically (if case normalization applied)

### API Wrapper Testing Checklist

- [ ] **Compilation**
  - [ ] No build errors
  - [ ] No import errors
  - [ ] API starts successfully

- [ ] **Registration**
  - [ ] Program registers on startup
  - [ ] No duplicate registration error
  - [ ] Appears in `/api/programs/list`

- [ ] **Info() Method**
  - [ ] ID is unique
  - [ ] Name is descriptive
  - [ ] Description is comprehensive
  - [ ] Category is correct
  - [ ] All parameters documented
  - [ ] Required parameters marked
  - [ ] Parameter types correct

- [ ] **Execute() Method**
  - [ ] Missing command handled
  - [ ] Invalid command handled
  - [ ] All valid commands work
  - [ ] Required parameters enforced
  - [ ] Optional parameters work

- [ ] **Parameter Extraction**
  - [ ] String parameters work
  - [ ] Integer parameters work
  - [ ] Boolean parameters work
  - [ ] Float64 to int conversion works
  - [ ] Missing parameters return defaults

- [ ] **Case Sensitivity**
  - [ ] Case normalization applied (if needed)
  - [ ] Uppercase input works
  - [ ] Lowercase input works
  - [ ] Mixed case input works

- [ ] **Error Handling**
  - [ ] CLI errors propagate correctly
  - [ ] Error messages are clear
  - [ ] Output included in errors
  - [ ] No panics or crashes

---

## Integration Testing

### Full Workflow Testing

Test complete user workflows:

```bash
# Set test environment
export DB_PATH=/tmp/test-integration.db
rm -f $DB_PATH

# Workflow: Create → List → Update → Query → Delete

# Step 1: Create item
./your_command add --name "Item 1"

# Step 2: Verify via API
curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"list"}}' \
  http://localhost:8080/api/programs/execute | jq

# Step 3: Update via API
curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"update","id":1,"name":"Updated"}}' \
  http://localhost:8080/api/programs/execute

# Step 4: Verify via CLI
./your_command list | grep "Updated"

# Step 5: Delete via CLI
./your_command delete --id 1

# Step 6: Verify via API
curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"list"}}' \
  http://localhost:8080/api/programs/execute | jq

# Cleanup
rm -f $DB_PATH
```

### Concurrent Request Testing

Test multiple simultaneous API requests:

```bash
# Test concurrent API requests
for i in {1..10}; do
    curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
      -d "{\"program_id\":\"your-program\",\"parameters\":{\"command\":\"add\",\"name\":\"Item $i\"}}" \
      http://localhost:8080/api/programs/execute &
done
wait

# Verify all items created
curl -X POST -H "X-API-Key: your-key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"list"}}' \
  http://localhost:8080/api/programs/execute | jq '.output' | grep -c "Item"
```

**Expected**: Count should be 10

### Integration Testing Checklist

- [ ] **CLI ↔ API Consistency**
  - [ ] CLI changes visible via API
  - [ ] API changes visible via CLI
  - [ ] Both use same database

- [ ] **Concurrent Access**
  - [ ] Multiple API requests work
  - [ ] CLI + API work simultaneously
  - [ ] No database locks
  - [ ] No data corruption

- [ ] **Error Propagation**
  - [ ] CLI exit codes translate correctly
  - [ ] Error messages pass through API
  - [ ] Stack traces don't leak

---

## LLM Interaction Testing

### Discovery Testing

Test if LLM can discover the program:

**Prompt**: "What programs are available?"

**Expected Response**: Should list your program with description

### Intent Recognition Testing

Test if LLM knows when to use your program:

**Test Cases**:
```
Prompt: "Add a new task"
Expected: Uses task-manager program

Prompt: "Show me my tasks"
Expected: Uses task-manager program

Prompt: "What's the weather?"
Expected: Does NOT use task-manager (not relevant)
```

### Parameter Extraction Testing

Test if LLM extracts parameters correctly:

**Test Case 1: Simple Command**
```
User: "List all projects"
Expected LLM extraction:
{
  "program_id": "task-manager",
  "parameters": {
    "command": "list-projects"
  }
}
```

**Test Case 2: With Parameters**
```
User: "Add a sawmill project to property with goal of making it operational"
Expected LLM extraction:
{
  "program_id": "task-manager",
  "parameters": {
    "command": "add-project",
    "context": "property",
    "project_name": "sawmill",
    "goal": "making it operational"
  }
}
```

**Test Case 3: Case Variations**
```
User: "Add project to Property" (capitalized)
User: "Add project to PROPERTY" (all caps)
User: "Add project to property" (lowercase)
Expected: All should work (if case normalization implemented)
```

### Natural Language Variations

Test different ways of saying the same thing:

```
"Show me the sawmill project status"
"What's the status of sawmill?"
"Tell me about the sawmill project"
"How is the sawmill doing?"
```

**All should**: Execute query-status command for sawmill

### Error Recovery Testing

Test if LLM handles errors gracefully:

```
User: "Add a project"
API: "missing required parameter: project_name"
Expected: LLM asks for project name
```

### LLM Testing Checklist

- [ ] **Discovery**
  - [ ] Program appears in available programs list
  - [ ] Description helps LLM understand purpose

- [ ] **Intent Recognition**
  - [ ] LLM uses program for relevant requests
  - [ ] LLM doesn't use program for irrelevant requests

- [ ] **Parameter Extraction**
  - [ ] LLM extracts required parameters
  - [ ] LLM extracts optional parameters when mentioned
  - [ ] LLM handles case variations

- [ ] **Natural Language**
  - [ ] Multiple phrasings work
  - [ ] Colloquial language works
  - [ ] Formal language works

- [ ] **Error Handling**
  - [ ] LLM explains errors to user
  - [ ] LLM retries with corrections
  - [ ] LLM asks for missing information

---

## Test Script Templates

### Basic CLI Test Script

```bash
#!/bin/bash
# test_cli.sh - Basic CLI testing script

set -e  # Exit on error

# Configuration
PROGRAM_NAME="your-program"
DB_PATH="/tmp/test-${PROGRAM_NAME}.db"
PASS_COUNT=0
FAIL_COUNT=0

# Setup
echo "=== Testing $PROGRAM_NAME CLI ==="
export DB_PATH
rm -f "$DB_PATH"

# Helper function
test_command() {
    local description="$1"
    local command="$2"
    local expected_exit="$3"

    echo -n "Testing: $description... "

    if eval "$command" > /dev/null 2>&1; then
        actual_exit=0
    else
        actual_exit=$?
    fi

    if [ "$actual_exit" -eq "$expected_exit" ]; then
        echo "PASS"
        ((PASS_COUNT++))
    else
        echo "FAIL (expected exit $expected_exit, got $actual_exit)"
        ((FAIL_COUNT++))
    fi
}

# Tests
test_command "Help message" "your_command --help" 0
test_command "Missing args" "your_command add" 1
test_command "Valid creation" "your_command add --name 'Test'" 0
test_command "List items" "your_command list" 0
test_command "Not found" "your_command get --id 999" 3

# Results
echo ""
echo "=== Test Results ==="
echo "Passed: $PASS_COUNT"
echo "Failed: $FAIL_COUNT"

# Cleanup
rm -f "$DB_PATH"

# Exit
if [ "$FAIL_COUNT" -gt 0 ]; then
    exit 1
fi
exit 0
```

### API Integration Test Script

```bash
#!/bin/bash
# test_api.sh - API integration testing script

API_URL="http://localhost:8080"
API_KEY="your-api-key"
PROGRAM_ID="your-program"

# Helper function
api_test() {
    local description="$1"
    local params="$2"
    local should_succeed="$3"

    echo -n "Testing API: $description... "

    response=$(curl -s -X POST \
        -H "X-API-Key: $API_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"program_id\":\"$PROGRAM_ID\",\"parameters\":$params}" \
        "$API_URL/api/programs/execute")

    success=$(echo "$response" | jq -r '.success')

    if [ "$success" == "$should_succeed" ]; then
        echo "PASS"
    else
        echo "FAIL"
        echo "  Response: $response"
    fi
}

# Tests
api_test "Missing command" '{}' "false"
api_test "Invalid command" '{"command":"invalid"}' "false"
api_test "Valid list" '{"command":"list"}' "true"
api_test "Valid add" '{"command":"add","name":"Test"}' "true"
```

### Full Integration Test Script

```bash
#!/bin/bash
# test_integration.sh - Full integration test

set -e

export DB_PATH="/tmp/test-integration.db"
rm -f "$DB_PATH"

echo "=== Integration Test ==="

# Test 1: CLI create
echo "1. Creating via CLI..."
./your_command add --name "Test Item"

# Test 2: API list
echo "2. Listing via API..."
response=$(curl -s -X POST \
    -H "X-API-Key: your-key" \
    -H "Content-Type: application/json" \
    -d '{"program_id":"your-program","parameters":{"command":"list"}}' \
    http://localhost:8080/api/programs/execute)

if echo "$response" | grep -q "Test Item"; then
    echo "   PASS: Item visible via API"
else
    echo "   FAIL: Item not found"
    exit 1
fi

# Test 3: API update
echo "3. Updating via API..."
curl -s -X POST \
    -H "X-API-Key: your-key" \
    -H "Content-Type: application/json" \
    -d '{"program_id":"your-program","parameters":{"command":"update","id":1,"name":"Updated"}}' \
    http://localhost:8080/api/programs/execute > /dev/null

# Test 4: CLI verify
echo "4. Verifying via CLI..."
if ./your_command list | grep -q "Updated"; then
    echo "   PASS: Update visible via CLI"
else
    echo "   FAIL: Update not found"
    exit 1
fi

echo ""
echo "=== All Integration Tests Passed ==="

rm -f "$DB_PATH"
```

---

## Deployment Verification

### Pre-Deployment Checklist

Before deploying to production:

- [ ] All tests passing locally
- [ ] Code reviewed
- [ ] Documentation updated
- [ ] README updated
- [ ] Git committed with clear message

### Deployment Steps with Verification

```bash
# 1. Push code
git push origin main

# 2. On server: Pull changes
ssh user@server
cd /path/to/project
git pull

# 3. Install/update CLI program
cd PROGRAMS/your-program
./install.sh
# Verify:
which your_command

# 4. Rebuild API
cd ../../ORCHESTRATOR/agent-gateway
go build -o agent-gateway .
# Verify:
./agent-gateway --version  # If version flag exists

# 5. Update service
sudo ./update.sh
# Verify:
sudo systemctl status agent-gateway

# 6. Check logs
sudo journalctl -u agent-gateway -f
# Look for: "Registered program: your-program"

# 7. Test API endpoint
curl -H "X-API-Key: key" http://localhost:8080/api/programs/list | jq | grep your-program
```

### Post-Deployment Testing

```bash
# Test 1: Program registered
curl -H "X-API-Key: key" http://localhost:8080/api/programs/list | jq '.programs[] | select(.id=="your-program")'

# Test 2: Simple command
curl -X POST \
  -H "X-API-Key: key" \
  -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"list"}}' \
  http://localhost:8080/api/programs/execute

# Test 3: LLM interaction (via TUI or chat)
# Send natural language request and verify response
```

### Post-Deployment Checklist

- [ ] Service running without errors
- [ ] Program registered (check logs)
- [ ] Program in `/api/programs/list`
- [ ] Simple API call succeeds
- [ ] LLM can discover program
- [ ] LLM can use program
- [ ] No errors in logs after 10 minutes
- [ ] Database created/accessible
- [ ] Permissions correct

---

## Troubleshooting Tests

### Test Fails: "Executable not found"

**Symptoms**:
```
exec: "your_command": executable file not found in $PATH
```

**Debug**:
```bash
# Check if installed
which your_command

# Check PATH
echo $PATH | grep ".local/bin"

# Check systemd PATH (if service)
systemctl show agent-gateway | grep PATH

# Fix: Add to systemd
sudo systemctl edit agent-gateway
# Add: Environment="PATH=/home/user/.local/bin:..."
sudo systemctl daemon-reload
sudo systemctl restart agent-gateway
```

### Test Fails: "Database error"

**Symptoms**:
```
Error: database is locked
Error: unable to open database file
```

**Debug**:
```bash
# Check database exists
ls -la "$DB_PATH"

# Check permissions
ls -la "$(dirname $DB_PATH)"

# Check for locks
fuser "$DB_PATH"

# Check SQLite
sqlite3 "$DB_PATH" "PRAGMA integrity_check;"

# Fix permissions
chmod 644 "$DB_PATH"
chmod 755 "$(dirname $DB_PATH)"
```

### Test Fails: "Invalid API key"

**Debug**:
```bash
# Check config
cat /etc/agent-gateway/config.yaml | grep api_key

# Check environment
env | grep API

# Check systemd env
systemctl show agent-gateway | grep Environment

# Verify key matches
# Config: api_key: "abc123"
# Request: -H "X-API-Key: abc123"
```

### Test Fails: "Program not registered"

**Debug**:
```bash
# Check logs
sudo journalctl -u agent-gateway -n 100 | grep -i register

# Check for errors
sudo journalctl -u agent-gateway -n 100 | grep -i error

# Verify import in main.go
grep "your_program" main.go

# Rebuild and restart
go build
sudo systemctl restart agent-gateway
```

---

## Summary

### Quick Test Commands

```bash
# CLI quick test
export DB_PATH=/tmp/test.db && rm -f $DB_PATH && \
  your_command add --name "Test" && \
  your_command list && \
  rm -f $DB_PATH

# API quick test
curl -X POST -H "X-API-Key: key" -H "Content-Type: application/json" \
  -d '{"program_id":"your-program","parameters":{"command":"list"}}' \
  http://localhost:8080/api/programs/execute | jq

# Registration check
curl -H "X-API-Key: key" http://localhost:8080/api/programs/list | \
  jq '.programs[] | select(.id=="your-program") | .name'
```

### Testing Priority Order

1. **CLI Program** (must work standalone)
2. **API Wrapper** (must compile and register)
3. **Integration** (CLI + API together)
4. **LLM Interaction** (natural language)
5. **Deployment** (production environment)

### When to Run Tests

- **After writing CLI command**: Test that command
- **After writing API wrapper**: Test registration and execution
- **Before committing**: Run full test suite
- **Before deploying**: Run integration tests
- **After deploying**: Run smoke tests
- **Daily**: Run automated tests (if set up)

---

**Next**: Use these checklists when building new programs to ensure quality and reliability.
