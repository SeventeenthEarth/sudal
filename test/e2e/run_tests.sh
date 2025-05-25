#!/bin/bash

# E2E Test Runner Script
# This script runs the E2E tests using pytest and pytest-bdd

set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting E2E Tests...${NC}"

# Check if virtual environment exists
if [ ! -d "$PROJECT_ROOT/venv" ]; then
    echo -e "${RED}Error: Virtual environment not found at $PROJECT_ROOT/venv${NC}"
    echo "Please create a virtual environment first:"
    echo "  python3 -m venv venv"
    exit 1
fi

# Activate virtual environment
echo -e "${YELLOW}Activating virtual environment...${NC}"
source "$PROJECT_ROOT/venv/bin/activate"

# Install dependencies if needed
echo -e "${YELLOW}Installing dependencies...${NC}"
pip install -r "$SCRIPT_DIR/requirements.txt"

# Check if server is running
SERVER_PORT=${SERVER_PORT:-8080}
echo -e "${YELLOW}Checking if server is running on port $SERVER_PORT...${NC}"

if ! curl -s "http://localhost:$SERVER_PORT/ping" > /dev/null; then
    echo -e "${RED}Error: Server is not running on port $SERVER_PORT${NC}"
    echo "Please start the server first:"
    echo "  make run"
    echo "Or run with Docker:"
    echo "  docker-compose up"
    exit 1
fi

echo -e "${GREEN}Server is running!${NC}"

# Change to test directory
cd "$SCRIPT_DIR"

# Run tests
echo -e "${YELLOW}Running E2E tests...${NC}"
pytest -v --tb=short "$@"

echo -e "${GREEN}E2E tests completed!${NC}"
