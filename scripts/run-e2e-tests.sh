#!/bin/bash

# Godog E2E Test Runner Script
# This script runs the godog-based E2E tests with proper setup and validation

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

# Tag sets managed here
# Rarely-run set (excluded from default run, included in .except run)
EXCEPT_TAGS=("@firebase_rate_limit")
# Temporarily skipped set (empty by default)
SKIP_TAGS=()

echo -e "${BLUE}=== Godog E2E Test Runner ===${NC}"

# Function to get verbose flag for go test
get_verbose_flag() {
    if [ "$VERBOSE" = true ]; then
        echo "-v"
    else
        echo ""
    fi
}

# Function to check if server is running
check_server() {
    echo -e "${YELLOW}Checking if server is running at ${SERVER_URL}...${NC}"

    if curl -s --max-time 5 "${SERVER_URL}/api/ping" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Server is running${NC}"
        return 0
    else
        echo -e "${RED}✗ Server is not running at ${SERVER_URL}${NC}"
        echo -e "${YELLOW}Please start the server with: make run${NC}"
        return 1
    fi
}

# Function to run all godog tests
build_exclude_expr() {
    # Usage: build_exclude_expr "${ARRAY[@]}"
    local expr=""
    local t
    for t in "$@"; do
        if [ -n "$t" ]; then
            if [ -n "$expr" ]; then expr+=" && "; fi
            expr+="~${t}"
        fi
    done
    echo "$expr"
}

build_include_or_expr() {
    # Usage: build_include_or_expr "${ARRAY[@]}"
    local expr=""
    local t
    for t in "$@"; do
        if [ -n "$t" ]; then
            if [ -n "$expr" ]; then expr+=" || "; fi
            expr+="${t}"
        fi
    done
    if [ -n "$expr" ]; then
        echo "(${expr})"
    else
        echo ""
    fi
}

run_all_tests() {
    echo -e "${BLUE}Running all godog E2E tests (filtering EXCEPT and SKIP)...${NC}"

    cd "$(dirname "$0")/.."
    if [ -f ".env" ]; then
        echo -e "${YELLOW}Loading environment variables from .env file...${NC}"
        set -a; source .env; set +a
    fi

    local verbose_flag=$(get_verbose_flag)
    local exclude_except=$(build_exclude_expr "${EXCEPT_TAGS[@]}")
    local exclude_skip=$(build_exclude_expr "${SKIP_TAGS[@]}")
    local tag_expr=""
    if [ -n "$exclude_except" ] && [ -n "$exclude_skip" ]; then
        tag_expr="$exclude_except && $exclude_skip"
    elif [ -n "$exclude_except" ]; then
        tag_expr="$exclude_except"
    elif [ -n "$exclude_skip" ]; then
        tag_expr="$exclude_skip"
    else
        tag_expr=""
    fi

    echo -e "${YELLOW}Tag filter (ALL): ${NC}${tag_expr:-<none>}"

    if [ -n "$tag_expr" ]; then
        go test $verbose_flag -count=1 ./test/e2e -godog.tags="$tag_expr"
    else
        go test $verbose_flag -count=1 ./test/e2e
    fi
}

run_except_set() {
    echo -e "${BLUE}Running EXCEPT set godog E2E tests...${NC}"

    cd "$(dirname "$0")/.."
    if [ -f ".env" ]; then
        echo -e "${YELLOW}Loading environment variables from .env file...${NC}"
        set -a; source .env; set +a
    fi

    local verbose_flag=$(get_verbose_flag)
    local include_except=$(build_include_or_expr "${EXCEPT_TAGS[@]}")
    local exclude_skip=$(build_exclude_expr "${SKIP_TAGS[@]}")
    if [ -z "$include_except" ]; then
        echo -e "${YELLOW}No EXCEPT tags configured. Nothing to run.${NC}"
        return 0
    fi
    local tag_expr="$include_except"
    if [ -n "$exclude_skip" ]; then
        tag_expr="$tag_expr && $exclude_skip"
    fi
    echo -e "${YELLOW}Tag filter (EXCEPT): ${NC}${tag_expr}"
    go test $verbose_flag -count=1 ./test/e2e -godog.tags="$tag_expr"
}

# Function to run specific scenarios
run_specific_scenarios() {
    local tags="$1"
    local scenario="$2"
    echo -e "${BLUE}Running specific godog scenarios...${NC}"

    # Change to project root directory (scripts is one level down from root)
    cd "$(dirname "$0")/.."

    # Load environment variables from .env file if it exists
    if [ -f ".env" ]; then
        echo -e "${YELLOW}Loading environment variables from .env file...${NC}"
        set -a  # automatically export all variables
        source .env
        set +a  # stop automatically exporting
    fi

    local verbose_flag=$(get_verbose_flag)
    local godog_args=""

    if [ -n "$tags" ]; then
        godog_args="$godog_args -godog.tags=\"$tags\""
        echo -e "${YELLOW}Using tags: $tags${NC}"
    fi

    if [ -n "$scenario" ]; then
        godog_args="$godog_args -godog.name=\"$scenario\""
        echo -e "${YELLOW}Using scenario: $scenario${NC}"
    fi

    # Run specific scenarios
    if eval "go test $verbose_flag -count=1 ./test/e2e $godog_args"; then
        echo -e "${GREEN}✓ Specific scenarios passed${NC}"
        return 0
    else
        echo -e "${RED}✗ Specific scenarios failed${NC}"
        return 1
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help           Show this help message"
    echo "  -s, --skip-check     Skip server availability check"
    echo "  -v, --verbose        Run tests with verbose output"
    echo "  -o, --only TAGS [SCENARIO]  Run specific scenarios by tags and/or scenario name"
    echo ""
    echo "Examples:"
    echo "  $0                           # Run all godog E2E tests"
    echo "  $0 -s                        # Run tests without server check"
    echo "  $0 -v                        # Run tests with verbose output"
    echo "  $0 -s -v                     # Run tests without server check and verbose output"
    echo "  $0 --only @health            # Run only health-tagged scenarios"
    echo "  $0 --only @rest              # Run only REST-tagged scenarios"
    echo "  $0 --only @grpc              # Run only gRPC-tagged scenarios"
    echo "  $0 --only @user              # Run only user-tagged scenarios"
    echo "  $0 --only @user_auth         # Run only user auth-tagged scenarios"
    echo "  $0 --only @quiz              # Run only quiz-tagged scenarios"
    echo "  $0 --run-except-set          # Run only scenarios in EXCEPT_TAGS set"
    echo "  $0 --only @health \"Basic health check\"  # Run specific scenario"
    echo ""
}

# Parse command line arguments
SKIP_CHECK=false
ONLY_MODE=false
EXCEPT_MODE=false
VERBOSE=false
TAGS=""
SCENARIO=""

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
            VERBOSE=true
            shift
            ;;
        -o|--only)
            ONLY_MODE=true
            shift
            if [[ $# -gt 0 && ! "$1" =~ ^- ]]; then
                TAGS="$1"
                shift
                # Check if next argument is a scenario name (not starting with -)
                if [[ $# -gt 0 && ! "$1" =~ ^- ]]; then
                    SCENARIO="$1"
                    shift
                fi
            else
                echo -e "${RED}Error: --only requires a tags argument${NC}"
                show_usage
                exit 1
            fi
            ;;
        --run-except-set)
            EXCEPT_MODE=true
            shift
            ;;
        -*)
            echo -e "${RED}Unknown option: $1${NC}"
            show_usage
            exit 1
            ;;
        *)
            echo -e "${RED}Unexpected argument: $1${NC}"
            show_usage
            exit 1
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

    # Run tests based on mode
    if [ "$ONLY_MODE" = true ]; then
        run_specific_scenarios "$TAGS" "$SCENARIO"
    elif [ "$EXCEPT_MODE" = true ]; then
        run_except_set
    else
        run_all_tests
    fi
}

# Run main function
main "$@"
