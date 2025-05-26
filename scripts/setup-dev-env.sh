#!/bin/bash

# Development Environment Setup Script
# This script initializes the development environment for the Sudal project
# Extracted from Makefile to improve maintainability

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration - can be overridden by environment variables
DESIRED_ORIGIN_URL="${DESIRED_ORIGIN_URL:-git@github.com-17thearth:SeventeenthEarth/sudal.git}"
GIT_USER_NAME="${GIT_USER_NAME:-17thearth}"
GIT_USER_EMAIL="${GIT_USER_EMAIL:-17thearth@gmail.com}"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to handle errors
handle_error() {
    print_error "$1"
    exit 1
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to setup Git configuration
setup_git() {
    print_info "=== Setting up Git configuration ==="
    
    # Check if we're in a Git repository
    if [ ! -d ".git" ]; then
        print_warning "Not in a Git repository. Skipping Git setup."
        return 0
    fi
    
    # Setup Git remote origin
    print_info "Checking Git remote 'origin'..."
    
    if ! git remote | grep -q '^origin$'; then
        print_info "Remote 'origin' not found. Adding..."
        git remote add origin "$DESIRED_ORIGIN_URL" || handle_error "Failed to add Git remote origin"
        print_success "Remote 'origin' added successfully."
    else
        CURRENT_ORIGIN_URL=$(git remote get-url origin)
        print_info "Remote 'origin' already exists with URL: $CURRENT_ORIGIN_URL"
        
        if [ "$CURRENT_ORIGIN_URL" != "$DESIRED_ORIGIN_URL" ]; then
            print_info "Updating remote 'origin' URL to $DESIRED_ORIGIN_URL..."
            git remote set-url origin "$DESIRED_ORIGIN_URL" || handle_error "Failed to update Git remote origin"
            print_success "Remote 'origin' URL updated."
        else
            print_success "Remote 'origin' URL is already correct."
        fi
    fi
    
    # Setup Git user configuration (local to this repository)
    print_info "Configuring local Git user name and email..."
    git config --local user.name "$GIT_USER_NAME" || handle_error "Failed to set Git user name"
    git config --local user.email "$GIT_USER_EMAIL" || handle_error "Failed to set Git user email"
    print_success "Local Git user configured: $GIT_USER_NAME <$GIT_USER_EMAIL>"
}

# Function to setup Go environment
setup_go_environment() {
    print_info "=== Setting up Go environment ==="
    
    # Check if Go is installed
    if ! command_exists go; then
        handle_error "Go is not installed. Please install Go first."
    fi
    
    print_info "Go version: $(go version)"
    print_info "GOPATH: $(go env GOPATH)"
    print_info "GOROOT: $(go env GOROOT)"
    
    # Download Go module dependencies
    print_info "Downloading Go module dependencies..."
    go mod download || handle_error "Failed to download Go module dependencies"
    print_success "Go module dependencies downloaded successfully."
    
    # Verify go.mod and go.sum are in sync
    print_info "Verifying Go module integrity..."
    go mod verify || handle_error "Go module verification failed"
    print_success "Go module integrity verified."
    
    # Tidy up go.mod and go.sum
    print_info "Tidying up Go modules..."
    go mod tidy || handle_error "Failed to tidy Go modules"
    print_success "Go modules tidied up."
}

# Function to create necessary directories
setup_directories() {
    print_info "=== Setting up project directories ==="
    
    # List of directories that should exist
    DIRECTORIES=(
        "bin"
        "tmp"
        "db/migrations"
        "gen/go"
        "gen/openapi"
        "internal/mocks"
        "proto"
        "scripts"
        "test/e2e"
        "test/integration"
    )
    
    for dir in "${DIRECTORIES[@]}"; do
        if [ ! -d "$dir" ]; then
            print_info "Creating directory: $dir"
            mkdir -p "$dir" || handle_error "Failed to create directory: $dir"
        fi
    done
    
    print_success "Project directories verified/created."
}

# Function to check environment files
check_environment_files() {
    print_info "=== Checking environment configuration ==="
    
    # Check for .env.template
    if [ ! -f ".env.template" ]; then
        print_warning ".env.template file not found."
    else
        print_success ".env.template file exists."
    fi
    
    # Check for .env file
    if [ ! -f ".env" ]; then
        print_warning ".env file not found."
        if [ -f ".env.template" ]; then
            print_info "You can create .env by copying .env.template:"
            print_info "  cp .env.template .env"
        fi
    else
        print_success ".env file exists."
    fi
    
    # Check for config.yaml
    if [ ! -f "configs/config.yaml" ]; then
        print_warning "configs/config.yaml file not found."
    else
        print_success "configs/config.yaml file exists."
    fi
}

# Function to display helpful information
display_next_steps() {
    print_info "=== Next Steps ==="
    echo ""
    echo "Your development environment is now set up! Here are some useful commands:"
    echo ""
    echo "  📦 Install development tools:"
    echo "    make install-tools"
    echo "    # or: ./scripts/install-tools.sh"
    echo ""
    echo "  🔧 Generate code (protobuf, mocks, etc.):"
    echo "    make generate"
    echo ""
    echo "  🧪 Run tests:"
    echo "    make test          # All tests"
    echo "    make test.unit     # Unit tests only"
    echo "    make test.int      # Integration tests only"
    echo "    make test.e2e      # E2E tests only"
    echo ""
    echo "  🚀 Run the application:"
    echo "    make run           # Using Docker Compose"
    echo ""
    echo "  🧹 Clean up:"
    echo "    make clean         # Clean build artifacts"
    echo "    make clean-all     # Clean everything"
    echo ""
    echo "  📚 Get help:"
    echo "    make help          # Show all available targets"
    echo ""
}

# Main setup function
main() {
    print_info "=== Sudal Development Environment Setup ==="
    print_info "Starting development environment initialization..."
    echo ""
    
    # Run setup functions
    setup_go_environment
    echo ""
    
    setup_git
    echo ""
    
    setup_directories
    echo ""
    
    check_environment_files
    echo ""
    
    print_success "=== Development environment setup completed successfully! ==="
    echo ""
    
    display_next_steps
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --git-url)
            DESIRED_ORIGIN_URL="$2"
            shift 2
            ;;
        --git-user)
            GIT_USER_NAME="$2"
            shift 2
            ;;
        --git-email)
            GIT_USER_EMAIL="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --git-url URL     Set Git remote origin URL (default: $DESIRED_ORIGIN_URL)"
            echo "  --git-user NAME   Set Git user name (default: $GIT_USER_NAME)"
            echo "  --git-email EMAIL Set Git user email (default: $GIT_USER_EMAIL)"
            echo "  --help, -h        Show this help message"
            echo ""
            echo "Environment variables can also be used:"
            echo "  DESIRED_ORIGIN_URL, GIT_USER_NAME, GIT_USER_EMAIL"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            print_info "Use --help for usage information."
            exit 1
            ;;
    esac
done

# Run main function
main "$@"
