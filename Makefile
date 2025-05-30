# Sudal - Social Quiz Platform Backend
# This Makefile provides development and build automation for the Sudal project
# Built with Go, Connect-go (gRPC), PostgreSQL, and Redis
#
# Complex operations have been moved to scripts/ directory for better maintainability
# Run './scripts/SCRIPT_NAME.sh --help' for detailed help on each script

# Project Configuration
PROJECT_NAME := sudal
CMD_PATH := ./cmd/server
OUTPUT_DIR := ./bin

# Tool Detection (for backward compatibility)
GOLANGCILINT := $(shell command -v golangci-lint 2> /dev/null)
GINKGO := $(shell command -v ginkgo 2> /dev/null)
MOCKGEN := $(shell command -v mockgen 2> /dev/null)

.PHONY: help init install-tools build test test.prepare test.unit test.int test.e2e fmt vet lint generate clean clean-all clean-proto clean-mocks clean-ginkgo clean-wire clean-ogen clean-tmp clean-build clean-coverage clean-go-cache clean-go-modules run generate-buf generate-wire generate-mocks generate-ogen generate-ginkgo buf-generate buf-lint buf-breaking buf-setup wire-gen ogen-generate ginkgo-bootstrap migrate-up migrate-down migrate-status migrate-version migrate-force migrate-create migrate-reset migrate-drop migrate-fresh

.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "🚀 Sudal - Social Quiz Platform Backend"
	@echo "========================================"
	@echo ""
	@echo "📋 Quick Start:"
	@echo "  make init          # Initialize development environment"
	@echo "  make install-tools # Install development tools"
	@echo "  make generate      # Generate all code"
	@echo "  make test          # Run all tests"
	@echo "  make run           # Run the application"
	@echo ""
	@echo "📋 Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_.-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "� Code Quality:"
	@echo "  make fmt                       # Format Go code"
	@echo "  make vet                       # Run Go vet"
	@echo "  make lint                      # Run linter checks"
	@echo ""
	@echo "�🔧 Cleanup Operations:"
	@echo "  make clean-proto               # Clean Protocol Buffer files"
	@echo "  make clean-mocks               # Clean mock files"
	@echo "  make clean-wire                # Clean Wire generated code"
	@echo "  make clean-ogen                # Clean OpenAPI generated code"
	@echo "  make clean-tmp                 # Clean temporary files"
	@echo "  make clean-coverage            # Clean test coverage files"
	@echo ""
	@echo "🗄️  Database Operations:"
	@echo "  make migrate-create DESC=name  # Create new migration"
	@echo "  make migrate-force VERSION=5   # Force set migration version"
	@echo "  make migrate-reset             # Reset database and reapply migrations"
	@echo "  make migrate-drop              # Drop all database objects (DANGEROUS)"
	@echo "  make migrate-fresh             # Fresh migration setup"
	@echo ""
	@echo "🔧 Code Generation:"
	@echo "  make generate-buf              # Generate protobuf code"
	@echo "  make generate-ogen             # Generate OpenAPI server code"
	@echo "  make generate-wire             # Generate dependency injection code"
	@echo "  make generate-mocks            # Generate mocks"
	@echo "  make generate-ginkgo           # Bootstrap Ginkgo test suites"
	@echo ""
	@echo "📦 Buf Operations:"
	@echo "  make buf-setup                 # Setup buf configuration files"
	@echo "  make buf-lint                  # Lint protobuf files"
	@echo "  make buf-breaking              # Check for breaking changes"
	@echo ""
	@echo "💡 All operations are now available via make commands!"
	@echo "📚 For script help: ./scripts/SCRIPT_NAME.sh --help"

init: ## Initialize development environment (quick setup)
	@echo "🚀 Initializing development environment..."
	@./scripts/setup-dev-env.sh
	@echo ""
	@echo "✅ Development environment initialized!"
	@echo ""
	@echo "💡 Next steps:"
	@echo "  make install-tools  # Install development tools"
	@echo "  make generate       # Generate code"

install-tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	@./scripts/install-tools.sh
	@echo ""
	@echo "✅ Development tools installed!"

build: ## Build the application
	@echo "🏗️  Building application..."
	@mkdir -p $(OUTPUT_DIR)
	@go build -ldflags="-s -w" -o $(OUTPUT_DIR)/$(PROJECT_NAME) $(CMD_PATH)/main.go
	@echo "✅ Build completed: $(OUTPUT_DIR)/$(PROJECT_NAME)"

# Test targets
test.prepare: generate fmt vet lint ## Prepare for running tests (format, vet, lint, generate)
	@echo "✅ Test preparation completed"

test: test.prepare test.unit test.int ## Run all tests (unit and integration)
	@echo "✅ All tests completed"

test.unit: ## Run unit tests
	@echo "🧪 Running unit tests..."
ifeq ($(GINKGO),)
	@go test -v -race -coverprofile=coverage.unit.out `go list ./internal/... | grep -v "/mocks"` || { echo "❌ Unit tests failed"; exit 1; }
else
	@$(GINKGO) -v -race -cover --coverprofile=coverage.unit.out --trace --fail-on-pending --randomize-all ./internal/... || { echo "❌ Unit tests failed"; exit 1; }
endif
	@go tool cover -func=coverage.unit.out
	@go tool cover -html=coverage.unit.out -o coverage.unit.html
	@echo "✅ Unit tests completed - coverage report: coverage.unit.html"

test.int: ## Run integration tests (excludes infrastructure - use test.unit.infra for infrastructure coverage)
	@echo "🧪 Running integration tests..."
ifeq ($(GINKGO),)
	@go test -v -race -coverprofile=coverage.int.out -coverpkg=github.com/seventeenthearth/sudal/internal/feature/... ./test/integration || { echo "❌ Integration tests failed"; exit 1; }
else
	@$(GINKGO) -v -race -cover -coverpkg=github.com/seventeenthearth/sudal/internal/feature/... --coverprofile=coverage.int.out --trace --fail-on-pending --randomize-all ./test/integration || { echo "❌ Integration tests failed"; exit 1; }
endif
	@go tool cover -func=coverage.int.out
	@go tool cover -html=coverage.int.out -o coverage.int.html
	@echo "✅ Integration tests completed - coverage report: coverage.int.html"
	@echo "ℹ️  Note: Infrastructure coverage is measured separately via 'make test.unit.infra'"

test.e2e: ## Run end-to-end tests
	@echo "🧪 Running end-to-end tests..."
	@echo "⚠️  Note: Make sure the server is running (make run) before running E2E tests"
	@if ! curl -s "http://localhost:8080/ping" > /dev/null; then \
		echo "⚠️  Warning: Server doesn't appear to be running on port 8080"; \
		echo "   Run 'make run' in a separate terminal first"; \
	fi
	@go test -v -race ./test/e2e || { echo "❌ E2E tests failed"; exit 1; }
	@echo "✅ End-to-end tests completed"

# Code quality targets
fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatting completed"

vet: ## Run Go vet
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ Go vet completed"

lint: ## Run linter checks
	@echo "🔍 Running linter..."
ifndef GOLANGCILINT
	@echo "❌ golangci-lint not found. Run 'make install-tools' first."
	@exit 1
endif
	@$(GOLANGCILINT) run ./... || echo "⚠️  Warning: Linter found issues, but continuing..."
	@echo "✅ Linter completed"

# Code generation targets (simplified - use scripts for advanced operations)
generate: generate-ginkgo generate-buf generate-wire generate-mocks generate-ogen ## Generate all code (test suites, proto, wire, mocks, openapi)
	@echo "✅ All code generation completed"

buf-setup: ## Setup buf configuration files
	@echo "🔧 Setting up buf configuration..."
	@./scripts/setup-buf.sh setup
	@echo "✅ Buf configuration setup completed"

generate-buf: ## Generate protobuf code
	@echo "🔧 Generating protobuf code..."
	@./scripts/setup-buf.sh generate
	@echo "✅ Protobuf code generation completed"

# Legacy alias for backward compatibility
buf-generate: generate-buf

buf-lint: ## Lint protobuf files
	@echo "🔍 Linting protobuf files..."
	@./scripts/setup-buf.sh lint
	@echo "✅ Protobuf linting completed"

buf-breaking: ## Check for breaking changes in protobuf
	@echo "🔍 Checking for breaking changes..."
	@./scripts/setup-buf.sh breaking
	@echo "✅ Breaking change check completed"

generate-wire: ## Generate dependency injection code
	@echo "🔧 Generating Wire dependency injection code..."
	@if ! command -v wire >/dev/null 2>&1; then \
		echo "Installing wire..."; \
		GOPROXY=direct go install github.com/google/wire/cmd/wire@latest; \
	fi
	@cd internal/infrastructure/di && wire
	@echo "✅ Wire code generation completed"

# Legacy alias for backward compatibility
wire-gen: generate-wire

generate-mocks: ## Generate mocks
	@echo "🔧 Generating mocks..."
ifndef MOCKGEN
	@echo "❌ mockgen not found. Run 'make install-tools' first."
	@exit 1
endif
	@go generate ./... || echo "⚠️  Warning: Some mock generation may have failed, but continuing..."
	@echo "✅ Mock generation completed"

generate-ogen: ## Generate OpenAPI server code from specification
	@echo "🔧 Generating OpenAPI server code..."
	@if ! command -v ogen >/dev/null 2>&1; then \
		echo "Installing ogen..."; \
		GOPROXY=direct go install github.com/ogen-go/ogen/cmd/ogen@latest; \
	fi
	@go run github.com/ogen-go/ogen/cmd/ogen \
		-target internal/infrastructure/openapi \
		-package openapi \
		-clean \
		api/openapi.yaml
	@echo "✅ OpenAPI server code generation completed"

# Legacy alias for backward compatibility
ogen-generate: generate-ogen

generate-ginkgo: ## Bootstrap Ginkgo test suites
	@echo "🔧 Bootstrapping Ginkgo test suites..."
	@./scripts/setup_tests.sh
	@echo "✅ Ginkgo test suites bootstrapped"

# Legacy alias for backward compatibility
ginkgo-bootstrap: generate-ginkgo

# Cleanup targets - all operations available via make
clean: ## Clean build artifacts and generated files (standard cleanup)
	@echo "🧹 Cleaning build artifacts and generated files..."
	@./scripts/clean-all.sh standard
	@echo "✅ Standard cleanup completed"

clean-all: ## Clean everything including Go module cache
	@echo "🧹 Performing complete cleanup..."
	@./scripts/clean-all.sh all
	@echo "✅ Complete cleanup finished"

clean-proto: ## Clean only Protocol Buffer generated files
	@echo "🧹 Cleaning Protocol Buffer files..."
	@./scripts/clean-all.sh proto
	@echo "✅ Protocol Buffer files cleaned"

clean-mocks: ## Clean only generated mock files
	@echo "🧹 Cleaning mock files..."
	@./scripts/clean-all.sh mocks
	@echo "✅ Mock files cleaned"

clean-ginkgo: ## Clean only Ginkgo test suite files
	@echo "🧹 Cleaning Ginkgo test suite files..."
	@./scripts/clean-all.sh ginkgo
	@echo "✅ Ginkgo test suite files cleaned"

clean-wire: ## Clean only Wire generated code
	@echo "🧹 Cleaning Wire generated code..."
	@./scripts/clean-all.sh wire
	@echo "✅ Wire generated code cleaned"

clean-ogen: ## Clean only OpenAPI generated code
	@echo "🧹 Cleaning OpenAPI generated code..."
	@./scripts/clean-all.sh ogen
	@echo "✅ OpenAPI generated code cleaned"

clean-tmp: ## Clean only temporary files
	@echo "🧹 Cleaning temporary files..."
	@./scripts/clean-all.sh tmp
	@echo "✅ Temporary files cleaned"

clean-build: ## Clean only build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@./scripts/clean-all.sh build
	@echo "✅ Build artifacts cleaned"

clean-coverage: ## Clean only test coverage files
	@echo "🧹 Cleaning test coverage files..."
	@./scripts/clean-all.sh coverage
	@echo "✅ Test coverage files cleaned"

clean-go-cache: ## Clean only Go test cache
	@echo "🧹 Cleaning Go test cache..."
	@./scripts/clean-all.sh go-cache
	@echo "✅ Go test cache cleaned"

clean-go-modules: ## Clean only Go module cache
	@echo "🧹 Cleaning Go module cache..."
	@./scripts/clean-all.sh go-modules
	@echo "✅ Go module cache cleaned"

# Database migration targets - all operations available via make
migrate-up: ## Apply database migrations
	@echo "🗄️  Applying database migrations..."
	@./scripts/migrate.sh up
	@echo "✅ Database migrations applied"

migrate-down: ## Rollback last migration
	@echo "🗄️  Rolling back last migration..."
	@./scripts/migrate.sh down
	@echo "✅ Migration rolled back"

migrate-status: ## Show migration status
	@echo "🗄️  Checking migration status..."
	@./scripts/migrate.sh status

migrate-version: ## Show current migration version
	@echo "🗄️  Checking current migration version..."
	@./scripts/migrate.sh version

migrate-force: ## Force set migration version (usage: make migrate-force VERSION=5)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION parameter is required"; \
		echo "Usage: make migrate-force VERSION=5"; \
		exit 1; \
	fi
	@echo "🗄️  Force setting migration version to $(VERSION)..."
	@echo "$(VERSION)" | ./scripts/migrate.sh force
	@echo "✅ Migration version forced"

migrate-create: ## Create new migration (usage: make migrate-create DESC=description)
	@if [ -z "$(DESC)" ]; then \
		echo "❌ DESC parameter is required"; \
		echo "Usage: make migrate-create DESC=create_users_table"; \
		exit 1; \
	fi
	@echo "🗄️  Creating new migration: $(DESC)..."
	@./scripts/migrate.sh create $(DESC)
	@echo "✅ Migration files created"

migrate-reset: ## Reset database to clean state and reapply all migrations
	@echo "🗄️  Resetting database..."
	@echo "⚠️  WARNING: This will drop all database objects and reapply migrations!"
	@echo "Press Ctrl+C to cancel, or Enter to continue..."
	@read dummy
	@./scripts/migrate.sh reset
	@echo "✅ Database reset completed"

migrate-drop: ## Drop all database objects (DANGEROUS)
	@echo "🗄️  Dropping all database objects..."
	@echo "⚠️  WARNING: This will DELETE ALL DATA and ALL OBJECTS!"
	@echo "Type 'DROP ALL DATA' to confirm:"
	@read confirm; \
	if [ "$$confirm" != "DROP ALL DATA" ]; then \
		echo "Operation cancelled."; \
		exit 0; \
	fi
	@./scripts/migrate.sh drop
	@echo "✅ All database objects dropped"

migrate-fresh: ## Fresh migration setup - backup old migrations and start clean
	@echo "🗄️  Setting up fresh migrations..."
	@echo "⚠️  WARNING: This will backup and remove all current migration files!"
	@echo "Press Ctrl+C to cancel, or Enter to continue..."
	@read dummy
	@./scripts/migrate.sh fresh
	@echo "✅ Fresh migration setup completed"

# Application targets
run: ## Run the application using Docker Compose
	@echo "🚀 Running application with Docker Compose..."
	@docker-compose up --build

# Placeholder main.go creation (for initial setup)
$(CMD_PATH)/main.go:
	@mkdir -p $(@D)
	@echo "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Starting Sudal Server...\")\n}" > $@
	@go mod tidy

# Ensure build depends on main.go existing
build: $(CMD_PATH)/main.go
