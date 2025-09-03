#!/bin/bash

# Buf Configuration Setup Script
# This script manages buf configuration files for Protocol Buffer code generation
# Extracted from Makefile to improve maintainability and reduce duplication

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration - can be overridden by environment variables
PROJECT_NAME="${PROJECT_NAME:-github.com/seventeenthearth/sudal}"
PROTO_DIR="${PROTO_DIR:-proto}"
GEN_GO_DIR="${GEN_GO_DIR:-../gen/go}"
GEN_OPENAPI_DIR="${GEN_OPENAPI_DIR:-../gen/openapi}"

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

# Function to create buf.work.yaml
create_buf_work_yaml() {
    local file="buf.work.yaml"
    
    if [ -f "$file" ]; then
        print_info "$file already exists, skipping creation."
        return 0
    fi
    
    print_info "Creating $file..."
    
    cat > "$file" << EOF
version: v1
directories:
  - $PROTO_DIR
EOF
    
    print_success "$file created successfully."
}

# Function to create proto/buf.yaml
create_proto_buf_yaml() {
    local file="$PROTO_DIR/buf.yaml"
    
    # Create proto directory if it doesn't exist
    mkdir -p "$PROTO_DIR"
    
    if [ -f "$file" ]; then
        print_info "$file already exists, skipping creation."
        return 0
    fi
    
    print_info "Creating $file..."
    
    cat > "$file" << EOF
version: v1
name: $PROJECT_NAME
deps:
  - buf.build/googleapis/googleapis
  - buf.build/grpc-ecosystem/grpc-gateway
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
EOF
    
    print_success "$file created successfully."
}

# Function to create proto/buf.gen.yaml
create_proto_buf_gen_yaml() {
    local file="$PROTO_DIR/buf.gen.yaml"
    
    # Create proto directory if it doesn't exist
    mkdir -p "$PROTO_DIR"
    
    if [ -f "$file" ]; then
        print_info "$file already exists, skipping creation."
        return 0
    fi
    
    print_info "Creating $file..."
    
    cat > "$file" << EOF
version: v1
plugins:
  # Generate Go structs from Protocol Buffers
  - plugin: go
    out: $GEN_GO_DIR
    opt:
      - paths=source_relative

  # Generate Connect-go service interfaces, clients, and handlers
  - plugin: connect-go
    out: $GEN_GO_DIR
    opt:
      - paths=source_relative

  # Generate gRPC Go service interfaces and clients
  - plugin: go-grpc
    out: $GEN_GO_DIR
    opt:
      - paths=source_relative

  # Generate OpenAPI v2 specifications from Protocol Buffers
  - plugin: openapiv2
    out: $GEN_OPENAPI_DIR
    opt:
      - output_format=yaml
      - allow_merge=true
      - merge_file_name=api
EOF
    
    print_success "$file created successfully."
}

# Function to verify buf installation
verify_buf_installation() {
    print_info "Verifying buf installation..."
    
    if ! command_exists buf; then
        print_warning "buf not found. Installing..."
        go install github.com/bufbuild/buf/cmd/buf@latest || handle_error "Failed to install buf"
        print_success "buf installed successfully."
    else
        print_success "buf is already installed."
        print_info "buf version: $(buf --version)"
    fi
}

# Function to verify required protoc plugins
verify_protoc_plugins() {
    print_info "Verifying protoc plugins..."
    
    local plugins=(
        "protoc-gen-go:google.golang.org/protobuf/cmd/protoc-gen-go@latest"
        "protoc-gen-connect-go:connectrpc.com/connect/cmd/protoc-gen-connect-go@latest"
        "protoc-gen-go-grpc:google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
        "protoc-gen-openapiv2:github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest"
    )
    
    for plugin_info in "${plugins[@]}"; do
        local plugin_name="${plugin_info%%:*}"
        local plugin_package="${plugin_info##*:}"
        
        if ! command_exists "$plugin_name"; then
            print_warning "$plugin_name not found. Installing..."
            go install "$plugin_package" || handle_error "Failed to install $plugin_name"
            print_success "$plugin_name installed successfully."
        else
            print_success "$plugin_name is already installed."
        fi
    done
}

# Function to create necessary directories
create_directories() {
    print_info "Creating necessary directories..."
    
    local dirs=(
        "$PROTO_DIR"
        "gen/go"
        "gen/openapi"
    )
    
    for dir in "${dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            print_info "Creating directory: $dir"
            mkdir -p "$dir" || handle_error "Failed to create directory: $dir"
        fi
    done
    
    print_success "All necessary directories verified/created."
}

# Function to run buf operations
run_buf_operations() {
    local operation="$1"
    
    case "$operation" in
        "generate")
            print_info "Running buf generate..."
            cd "$PROTO_DIR" && buf generate || handle_error "buf generate failed"
            print_success "buf generate completed successfully."
            ;;
        "lint")
            print_info "Running buf lint..."
            cd "$PROTO_DIR" && buf lint || handle_error "buf lint failed"
            print_success "buf lint completed successfully."
            ;;
        "breaking")
            print_info "Running buf breaking change check..."
            
            # Check if we're in a git repository
            if [ ! -d ".git" ]; then
                print_warning "Not a git repository. Skipping breaking change check."
                return 0
            fi
            
            # Check if proto files are tracked in git
            if ! git ls-files --error-unmatch "$PROTO_DIR" > /dev/null 2>&1; then
                print_warning "No proto files tracked in git. Skipping breaking change check."
                return 0
            fi
            
            # Determine the reference branch
            if git show-ref --verify --quiet refs/heads/main; then
                local ref_branch="main"
            else
                local ref_branch=$(git rev-parse --abbrev-ref HEAD)
                print_warning "Main branch not found. Using current branch ($ref_branch) as reference."
            fi
            
            cd "$PROTO_DIR" && buf breaking --against "../.git#branch=$ref_branch" || handle_error "buf breaking check failed"
            print_success "buf breaking change check completed successfully."
            ;;
        *)
            print_error "Unknown operation: $operation"
            print_info "Supported operations: generate, lint, breaking"
            exit 1
            ;;
    esac
}

# Function to display help
show_help() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  setup     Setup buf configuration files (default)"
    echo "  generate  Run buf generate"
    echo "  lint      Run buf lint"
    echo "  breaking  Run buf breaking change check"
    echo "  help      Show this help message"
    echo ""
    echo "Options:"
    echo "  --project-name NAME   Set project name (default: $PROJECT_NAME)"
    echo "  --proto-dir DIR       Set proto directory (default: $PROTO_DIR)"
    echo "  --gen-go-dir DIR      Set Go generation directory (default: $GEN_GO_DIR)"
    echo "  --gen-openapi-dir DIR Set OpenAPI generation directory (default: $GEN_OPENAPI_DIR)"
    echo "  --help, -h            Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  PROJECT_NAME, PROTO_DIR, GEN_GO_DIR, GEN_OPENAPI_DIR"
    echo ""
    echo "Examples:"
    echo "  $0 setup                    # Setup buf configuration files"
    echo "  $0 generate                 # Generate code from proto files"
    echo "  $0 lint                     # Lint proto files"
    echo "  $0 breaking                 # Check for breaking changes"
    echo "  $0 setup --project-name myproject  # Setup with custom project name"
}

# Function to setup buf configuration
setup_buf() {
    print_info "=== Setting up buf configuration ==="
    
    verify_buf_installation
    verify_protoc_plugins
    create_directories
    create_buf_work_yaml
    create_proto_buf_yaml
    create_proto_buf_gen_yaml
    
    print_success "=== Buf configuration setup completed successfully! ==="
    
    print_info "=== Next Steps ==="
    echo "  📝 Add your .proto files to the $PROTO_DIR directory"
    echo "  🔧 Generate code: $0 generate"
    echo "  🧹 Lint proto files: $0 lint"
    echo "  🔍 Check breaking changes: $0 breaking"
}

# Main function
main() {
    local command="setup"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            setup|generate|lint|breaking)
                command="$1"
                shift
                ;;
            --project-name)
                PROJECT_NAME="$2"
                shift 2
                ;;
            --proto-dir)
                PROTO_DIR="$2"
                shift 2
                ;;
            --gen-go-dir)
                GEN_GO_DIR="$2"
                shift 2
                ;;
            --gen-openapi-dir)
                GEN_OPENAPI_DIR="$2"
                shift 2
                ;;
            --help|-h|help)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                print_info "Use --help for usage information."
                exit 1
                ;;
        esac
    done
    
    case "$command" in
        "setup")
            setup_buf
            ;;
        "generate"|"lint"|"breaking")
            # Ensure tools exist when invoking directly
            verify_buf_installation
            verify_protoc_plugins
            run_buf_operations "$command"
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
