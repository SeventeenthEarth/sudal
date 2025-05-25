#!/bin/bash

# Python Code Formatter Script
# This script formats Python code using Black

set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Formatting Python code with Black...${NC}"

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
pip install -r "$SCRIPT_DIR/requirements.txt" > /dev/null

# Change to test directory
cd "$SCRIPT_DIR"

# Check if we should just check formatting or actually format
if [ "$1" = "--check" ]; then
    echo -e "${YELLOW}Checking Python code formatting...${NC}"
    if black --check .; then
        echo -e "${GREEN}All Python code is properly formatted!${NC}"
    else
        echo -e "${RED}Some Python code needs formatting. Run 'make fmt-python' to fix.${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}Formatting Python code...${NC}"
    black .
    echo -e "${GREEN}Python code formatting completed!${NC}"
fi
