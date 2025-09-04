#!/bin/bash

# Install Development Tools Script
# This script installs all necessary development tools for the Sudal project
# Extracted from Makefile to improve maintainability

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to install a Go tool
install_go_tool() {
    local tool_name="$1"
    local tool_package="$2"
    local tool_binary="$3"
    local additional_flags="$4"

    print_info "Checking/Installing $tool_name..."

    if command_exists "$tool_binary"; then
        print_success "$tool_name already installed."
        return 0
    fi

    print_info "Installing $tool_name..."
    if [ -n "$additional_flags" ]; then
        go install $additional_flags "$tool_package" || handle_error "Failed to install $tool_name"
    else
        go install "$tool_package" || handle_error "Failed to install $tool_name"
    fi

    # Verify installation
    if command_exists "$tool_binary"; then
        print_success "$tool_name installed successfully."
    else
        # Check in GOPATH/bin
        GOPATH_BIN="$(go env GOPATH)/bin/$tool_binary"
        if [ -f "$GOPATH_BIN" ]; then
            print_success "$tool_name installed successfully at $GOPATH_BIN"
        else
            handle_error "Failed to verify $tool_name installation"
        fi
    fi
}

# Main installation function
main() {
    print_info "=== Installing Development Tools ==="

    # Check if Go is installed
    if ! command_exists go; then
        handle_error "Go is not installed. Please install Go first."
    fi

    print_info "Go version: $(go version)"
    print_info "GOPATH: $(go env GOPATH)"

    # Core development tools
    print_info "Installing core development tools..."

    # golangci-lint - Go linter
    install_go_tool "golangci-lint" \
        "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest" \
        "golangci-lint"

    # Ginkgo - BDD testing framework (match version in go.mod to avoid CLI/library mismatch)
    if [ -f "go.mod" ]; then
        GINKGO_VERSION=$(awk '/github.com\/onsi\/ginkgo\/v2[[:space:]]/{print $2; exit}' go.mod)
    fi
    if [ -n "$GINKGO_VERSION" ]; then
        install_go_tool "Ginkgo" \
            "github.com/onsi/ginkgo/v2/ginkgo@${GINKGO_VERSION}" \
            "ginkgo"
    else
        install_go_tool "Ginkgo" \
            "github.com/onsi/ginkgo/v2/ginkgo@latest" \
            "ginkgo"
    fi

    # mockgen - Mock generation tool
    install_go_tool "mockgen" \
        "go.uber.org/mock/mockgen@latest" \
        "mockgen"

    # Wire - Dependency injection code generator
    install_go_tool "Wire" \
        "github.com/google/wire/cmd/wire@latest" \
        "wire" \
        "GOPROXY=direct"

    # Protocol Buffer and gRPC tools
    print_info "Installing Protocol Buffer and gRPC tools..."

    # protoc-gen-go - Protocol Buffer Go plugin
    install_go_tool "protoc-gen-go" \
        "google.golang.org/protobuf/cmd/protoc-gen-go@latest" \
        "protoc-gen-go"

    # protoc-gen-connect-go - Connect-Go plugin
    install_go_tool "protoc-gen-connect-go" \
        "connectrpc.com/connect/cmd/protoc-gen-connect-go@latest" \
        "protoc-gen-connect-go"

    # protoc-gen-openapiv2 - OpenAPI v2 generator
    install_go_tool "protoc-gen-openapiv2" \
        "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest" \
        "protoc-gen-openapiv2"

    # buf - Protocol Buffer build tool
    install_go_tool "buf" \
        "github.com/bufbuild/buf/cmd/buf@latest" \
        "buf"

    # Database migration tool
    print_info "Installing database tools..."

    # golang-migrate - Database migration tool
    install_go_tool "migrate" \
        "github.com/golang-migrate/migrate/v4/cmd/migrate@latest" \
        "migrate" \
        "-tags 'postgres'"

    # OpenAPI tools
    print_info "Installing OpenAPI tools..."

    # ogen - OpenAPI v3 Go generator
    install_go_tool "ogen" \
        "github.com/ogen-go/ogen/cmd/ogen@latest" \
        "ogen"

    print_success "=== All development tools installed successfully ==="

    # Print summary
    print_info "=== Installation Summary ==="
    print_info "The following tools have been installed:"
    echo "  • golangci-lint - Go linter"
    echo "  • ginkgo - BDD testing framework"
    echo "  • mockgen - Mock generation"
    echo "  • wire - Dependency injection"
    echo "  • protoc-gen-go - Protocol Buffer Go plugin"
    echo "  • protoc-gen-connect-go - Connect-Go plugin"
    echo "  • protoc-gen-openapiv2 - OpenAPI v2 generator"
    echo "  • buf - Protocol Buffer build tool"
    echo "  • migrate - Database migration tool"
    echo "  • ogen - OpenAPI v3 Go generator"

    print_info "Tools are installed in: $(go env GOPATH)/bin"
    print_warning "Make sure $(go env GOPATH)/bin is in your PATH environment variable."
}

# Run main function
main "$@"
