#!/bin/bash

# API Gateway Testing Script
# This script demonstrates how to use the Agent Gateway API

# Configuration
API_KEY="${API_KEY:-your-secret-key}"
BASE_URL="${BASE_URL:-http://localhost:8080}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to make API request
api_request() {
    local endpoint=$1
    local description=$2

    echo -e "${BLUE}Testing: ${description}${NC}"
    echo -e "Endpoint: ${endpoint}"

    response=$(curl -s -w "\n%{http_code}" -H "X-API-Key: ${API_KEY}" "${BASE_URL}${endpoint}")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✓ Success (HTTP ${http_code})${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ Failed (HTTP ${http_code})${NC}"
        echo "$body"
    fi
    echo ""
}

# Function to test endpoint without auth
api_request_no_auth() {
    local endpoint=$1
    local description=$2

    echo -e "${BLUE}Testing: ${description}${NC}"
    echo -e "Endpoint: ${endpoint}"

    response=$(curl -s -w "\n%{http_code}" "${BASE_URL}${endpoint}")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✓ Success (HTTP ${http_code})${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ Failed (HTTP ${http_code})${NC}"
        echo "$body"
    fi
    echo ""
}

echo "=================================="
echo "Agent Gateway API Test Suite"
echo "=================================="
echo ""
echo "Base URL: ${BASE_URL}"
echo "API Key: ${API_KEY}"
echo ""

# Check if server is running
echo "Checking if server is running..."
if ! curl -s "${BASE_URL}/" > /dev/null 2>&1; then
    echo -e "${RED}Error: Server is not running at ${BASE_URL}${NC}"
    echo "Start the server with: ./agent-gateway --config config.yaml"
    exit 1
fi
echo -e "${GREEN}Server is running${NC}"
echo ""

# Meta Endpoints
echo "=================================="
echo "Meta Endpoints"
echo "=================================="
echo ""

api_request_no_auth "/api/health" "Health Check (no auth)"
api_request "/api/stats" "System Statistics"

# Stoic Endpoints
echo "=================================="
echo "Stoic Thought Endpoints"
echo "=================================="
echo ""

api_request "/api/stoic/today" "Get Today's Thought"
api_request "/api/stoic/random" "Get Random Thought"
api_request "/api/stoic/latest/5" "Get Latest 5 Thoughts"
api_request "/api/stoic/all?page=1&page_size=3" "Get All Thoughts (Paginated)"

# Tech Endpoints
echo "=================================="
echo "Tech Tip Endpoints"
echo "=================================="
echo ""

api_request "/api/tech/today" "Get Today's Tip"
api_request "/api/tech/random" "Get Random Tip"
api_request "/api/tech/latest/5" "Get Latest 5 Tips"
api_request "/api/tech/all?page=1&page_size=3" "Get All Tips (Paginated)"

# Test specific date (if exists)
echo "=================================="
echo "Date-specific Queries"
echo "=================================="
echo ""

TODAY=$(date +%Y-%m-%d)
api_request "/api/stoic/date/${TODAY}" "Get Stoic Thought for ${TODAY}"
api_request "/api/tech/date/${TODAY}" "Get Tech Tip for ${TODAY}"

echo "=================================="
echo "Test Suite Complete"
echo "=================================="
