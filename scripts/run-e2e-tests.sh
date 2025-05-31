#!/bin/bash

# Go E2E Test Runner Script
# This script runs the Go-based E2E tests with proper setup and validation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8080"
TIMEOUT=30

echo -e "${BLUE}=== Go E2E Test Runner ===${NC}"

# Function to check if server is running
check_server() {
    echo -e "${YELLOW}Checking if server is running at ${SERVER_URL}...${NC}"
    
    if curl -s --max-time 5 "${SERVER_URL}/ping" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Server is running${NC}"
        return 0
    else
        echo -e "${RED}✗ Server is not running at ${SERVER_URL}${NC}"
        echo -e "${YELLOW}Please start the server with: make run${NC}"
        return 1
    fi
}

# Function to run tests
run_tests() {
    echo -e "${BLUE}Running Go E2E tests...${NC}"
    
    # Change to project root directory (scripts is one level down from root)
    cd "$(dirname "$0")/.."
    
    # Run tests with verbose output
    if go test -v -race ./test/e2e; then
        echo -e "${GREEN}✓ All E2E tests passed${NC}"
        return 0
    else
        echo -e "${RED}✗ Some E2E tests failed${NC}"
        return 1
    fi
}

# Function to run specific test
run_specific_test() {
    local test_name="$1"
    echo -e "${BLUE}Running specific test: ${test_name}${NC}"
    
    # Change to project root directory (scripts is one level down from root)
    cd "$(dirname "$0")/.."
    
    # Run specific test
    if go test -v -race ./test/e2e -run "${test_name}"; then
        echo -e "${GREEN}✓ Test ${test_name} passed${NC}"
        return 0
    else
        echo -e "${RED}✗ Test ${test_name} failed${NC}"
        return 1
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS] [TEST_NAME]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -s, --skip-check  Skip server availability check"
    echo "  -v, --verbose  Run with verbose output (default)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run all E2E tests"
    echo "  $0 TestConnectGoHealthService         # Run specific test"
    echo "  $0 -s                                 # Run tests without server check"
    echo ""
}

# Parse command line arguments
SKIP_CHECK=false
TEST_NAME=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -s|--skip-check)
            SKIP_CHECK=true
            shift
            ;;
        -v|--verbose)
            # Verbose is default, so this is a no-op
            shift
            ;;
        -*)
            echo -e "${RED}Unknown option: $1${NC}"
            show_usage
            exit 1
            ;;
        *)
            TEST_NAME="$1"
            shift
            ;;
    esac
done

# Main execution
main() {
    # Check server availability unless skipped
    if [ "$SKIP_CHECK" = false ]; then
        if ! check_server; then
            exit 1
        fi
    else
        echo -e "${YELLOW}Skipping server availability check${NC}"
    fi
    
    # Run tests
    if [ -n "$TEST_NAME" ]; then
        run_specific_test "$TEST_NAME"
    else
        run_tests
    fi
}

# Run main function
main "$@"
