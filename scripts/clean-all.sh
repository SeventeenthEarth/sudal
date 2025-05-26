#!/bin/bash

# Project Cleanup Script
# This script manages all cleanup operations for the Sudal project
# Extracted from Makefile to improve maintainability and provide flexible cleanup options

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration - can be overridden by environment variables
OUTPUT_DIR="${OUTPUT_DIR:-./bin}"

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

# Function to clean Protocol Buffer generated files
clean_proto() {
    print_info "Cleaning generated Protocol Buffer files..."
    
    # Remove gen directory
    if [ -d "./gen" ]; then
        rm -rf ./gen
        print_success "Removed ./gen directory"
    fi
    
    # Remove .pb.go files
    if find ./proto -name "*.pb.go" -type f 2>/dev/null | grep -q .; then
        find ./proto -name "*.pb.go" -delete
        print_success "Removed *.pb.go files"
    fi
    
    # Remove .connect.go files
    if find ./proto -name "*.connect.go" -type f 2>/dev/null | grep -q .; then
        find ./proto -name "*.connect.go" -delete
        print_success "Removed *.connect.go files"
    fi
    
    # Remove healthv1connect directories
    find ./proto -path "*/*/healthv1connect" -type d -exec rm -rf {} \; 2>/dev/null || true
    
    print_success "Protocol Buffer files cleaned"
}

# Function to clean mock files
clean_mocks() {
    print_info "Cleaning generated mock files..."
    
    if [ -d "./internal/mocks" ]; then
        rm -rf ./internal/mocks
        print_success "Removed ./internal/mocks directory"
    else
        print_info "No mock files to clean"
    fi
    
    print_success "Mock files cleaned"
}

# Function to clean Ginkgo test suite files
clean_ginkgo() {
    print_info "Cleaning generated Ginkgo test suite files..."
    
    local count=0
    while IFS= read -r -d '' file; do
        rm -f "$file"
        count=$((count + 1))
    done < <(find . -name "*_suite_test.go" -type f -print0 2>/dev/null)
    
    if [ $count -gt 0 ]; then
        print_success "Removed $count Ginkgo test suite files"
    else
        print_info "No Ginkgo test suite files to clean"
    fi
    
    print_success "Ginkgo test suite files cleaned"
}

# Function to clean Wire generated code
clean_wire() {
    print_info "Cleaning generated Wire code..."
    
    local wire_file="internal/infrastructure/di/wire_gen.go"
    if [ -f "$wire_file" ]; then
        rm -f "$wire_file"
        print_success "Removed $wire_file"
    else
        print_info "No Wire generated files to clean"
    fi
    
    print_success "Wire generated code cleaned"
}

# Function to clean OpenAPI generated code
clean_ogen() {
    print_info "Cleaning generated OpenAPI code..."
    
    local ogen_dir="internal/infrastructure/openapi"
    if [ -d "$ogen_dir" ]; then
        local count=0
        for file in "$ogen_dir"/oas_*.go; do
            if [ -f "$file" ]; then
                rm -f "$file"
                count=$((count + 1))
            fi
        done
        
        if [ $count -gt 0 ]; then
            print_success "Removed $count OpenAPI generated files"
        else
            print_info "No OpenAPI generated files to clean"
        fi
    else
        print_info "No OpenAPI directory to clean"
    fi
    
    print_success "OpenAPI generated code cleaned"
}

# Function to clean temporary files
clean_tmp() {
    print_info "Cleaning temporary files..."
    
    # Remove tmp directory
    if [ -d "tmp/" ]; then
        rm -rf tmp/
        print_success "Removed tmp/ directory"
    fi
    
    # Remove various temporary files
    local temp_files=(
        ".air.log"
        ".air.toml.tmp"
        ".dockerignore.tmp"
    )
    
    local removed_count=0
    for file in "${temp_files[@]}"; do
        if [ -f "$file" ]; then
            rm -f "$file"
            removed_count=$((removed_count + 1))
        fi
    done
    
    # Remove .compiledaemon.* files
    for file in .compiledaemon.*; do
        if [ -f "$file" ]; then
            rm -f "$file"
            removed_count=$((removed_count + 1))
        fi
    done
    
    if [ $removed_count -gt 0 ]; then
        print_success "Removed $removed_count temporary files"
    else
        print_info "No temporary files to clean"
    fi
    
    print_success "Temporary files cleaned"
}

# Function to clean build artifacts
clean_build() {
    print_info "Cleaning build artifacts..."
    
    # Remove output directory
    if [ -d "$OUTPUT_DIR" ]; then
        rm -rf "$OUTPUT_DIR"
        print_success "Removed $OUTPUT_DIR directory"
    else
        print_info "No build output directory to clean"
    fi
    
    print_success "Build artifacts cleaned"
}

# Function to clean test coverage files
clean_coverage() {
    print_info "Cleaning test coverage files..."
    
    local coverage_files=(
        coverage*.out
        coverage*.html
        coverprofile.out
    )
    
    local removed_count=0
    for pattern in "${coverage_files[@]}"; do
        for file in $pattern; do
            if [ -f "$file" ]; then
                rm -f "$file"
                removed_count=$((removed_count + 1))
            fi
        done
    done
    
    if [ $removed_count -gt 0 ]; then
        print_success "Removed $removed_count coverage files"
    else
        print_info "No coverage files to clean"
    fi
    
    print_success "Test coverage files cleaned"
}

# Function to clean Go test cache
clean_go_cache() {
    print_info "Cleaning Go test cache..."
    
    if command_exists go; then
        go clean -testcache
        print_success "Go test cache cleaned"
    else
        print_warning "Go command not found, skipping test cache cleanup"
    fi
}

# Function to clean Go module cache
clean_go_modules() {
    print_info "Cleaning Go module cache..."
    print_warning "This may fail if modules are in use by other processes"
    
    if command_exists go; then
        if go clean -modcache 2>/dev/null; then
            print_success "Go module cache cleaned"
        else
            print_warning "Could not clean Go module cache completely. This is normal if modules are in use."
        fi
    else
        print_warning "Go command not found, skipping module cache cleanup"
    fi
}

# Function to perform standard cleanup
clean_standard() {
    print_info "=== Performing standard cleanup ==="
    
    clean_proto
    clean_mocks
    clean_ginkgo
    clean_wire
    clean_ogen
    clean_tmp
    clean_build
    clean_coverage
    clean_go_cache
    
    print_success "=== Standard cleanup completed ==="
}

# Function to perform complete cleanup
clean_all() {
    print_info "=== Performing complete cleanup ==="
    
    clean_standard
    clean_go_modules
    
    print_success "=== Complete cleanup finished ==="
}

# Function to display help
show_help() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  standard          Perform standard cleanup (default)"
    echo "  all               Perform complete cleanup including Go module cache"
    echo "  proto             Clean Protocol Buffer generated files"
    echo "  mocks             Clean generated mock files"
    echo "  ginkgo            Clean Ginkgo test suite files"
    echo "  wire              Clean Wire generated code"
    echo "  ogen              Clean OpenAPI generated code"
    echo "  tmp               Clean temporary files"
    echo "  build             Clean build artifacts"
    echo "  coverage          Clean test coverage files"
    echo "  go-cache          Clean Go test cache"
    echo "  go-modules        Clean Go module cache"
    echo "  help              Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  OUTPUT_DIR        Build output directory (default: $OUTPUT_DIR)"
    echo ""
    echo "Examples:"
    echo "  $0                        # Standard cleanup"
    echo "  $0 standard               # Standard cleanup"
    echo "  $0 all                    # Complete cleanup"
    echo "  $0 proto                  # Clean only protobuf files"
    echo "  $0 mocks                  # Clean only mock files"
    echo ""
    echo "Standard cleanup includes:"
    echo "  • Protocol Buffer generated files"
    echo "  • Mock files"
    echo "  • Ginkgo test suite files"
    echo "  • Wire generated code"
    echo "  • OpenAPI generated code"
    echo "  • Temporary files"
    echo "  • Build artifacts"
    echo "  • Test coverage files"
    echo "  • Go test cache"
    echo ""
    echo "Complete cleanup adds:"
    echo "  • Go module cache (may fail if modules are in use)"
}

# Main function
main() {
    local command="${1:-standard}"
    
    case "$command" in
        "standard"|"")
            clean_standard
            ;;
        "all")
            clean_all
            ;;
        "proto")
            clean_proto
            ;;
        "mocks")
            clean_mocks
            ;;
        "ginkgo")
            clean_ginkgo
            ;;
        "wire")
            clean_wire
            ;;
        "ogen")
            clean_ogen
            ;;
        "tmp")
            clean_tmp
            ;;
        "build")
            clean_build
            ;;
        "coverage")
            clean_coverage
            ;;
        "go-cache")
            clean_go_cache
            ;;
        "go-modules")
            clean_go_modules
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            print_info "Use '$0 help' for usage information."
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
