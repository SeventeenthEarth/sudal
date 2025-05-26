.PHONY: help init build test test.prepare test.unit test.unit.only test.int test.int.only test.e2e test.e2e.only clean clean-all lint fmt vet proto-clean buf-generate buf-lint buf-breaking run install-tools generate-mocks mock-clean ginkgo-bootstrap ginkgo-clean tmp-clean wire-gen wire-clean generate migrate-up migrate-down migrate-create migrate-status migrate-force migrate-version migrate-squash migrate-reset migrate-fresh migrate-drop

# Variables
BINARY_NAME=sudal-server
CMD_PATH=./cmd/server
OUTPUT_DIR=./bin
CONFIG_FILE?=./configs/config.yaml # Default config path, can be overridden

# Database Migration Variables
MIGRATIONS_DIR=./db/migrations
# Use POSTGRES_DSN if available, otherwise construct from individual components
DATABASE_URL?=$(POSTGRES_DSN)
ifeq ($(DATABASE_URL),)
    # Set defaults for database connection and strip any comments from environment variables
    DB_HOST_CLEAN=$(shell echo "$(DB_HOST)" | cut -d' ' -f1)
    DB_PORT_CLEAN=$(shell echo "$(DB_PORT)" | cut -d' ' -f1)
    DB_USER_CLEAN=$(shell echo "$(DB_USER)" | cut -d' ' -f1)
    DB_PASSWORD_CLEAN=$(shell echo "$(DB_PASSWORD)" | cut -d' ' -f1)
    DB_NAME_CLEAN=$(shell echo "$(DB_NAME)" | cut -d' ' -f1)
    DB_SSLMODE_CLEAN=$(shell echo "$(DB_SSLMODE)" | cut -d' ' -f1)
    # Use defaults if environment variables are empty
    # Note: Use localhost for local development, even when Docker Compose uses 'db' internally
    DB_HOST_FINAL=$(if $(DB_HOST_CLEAN),$(if $(filter db,$(DB_HOST_CLEAN)),localhost,$(DB_HOST_CLEAN)),localhost)
    DB_PORT_FINAL=$(if $(DB_PORT_CLEAN),$(DB_PORT_CLEAN),5432)
    DB_USER_FINAL=$(if $(DB_USER_CLEAN),$(DB_USER_CLEAN),user)
    DB_PASSWORD_FINAL=$(if $(DB_PASSWORD_CLEAN),$(DB_PASSWORD_CLEAN),password)
    DB_NAME_FINAL=$(if $(DB_NAME_CLEAN),$(DB_NAME_CLEAN),quizapp_db)
    DB_SSLMODE_FINAL=$(if $(DB_SSLMODE_CLEAN),$(DB_SSLMODE_CLEAN),disable)
    # Construct DATABASE_URL from cleaned components
    DATABASE_URL=postgres://$(DB_USER_FINAL):$(DB_PASSWORD_FINAL)@$(DB_HOST_FINAL):$(DB_PORT_FINAL)/$(DB_NAME_FINAL)?sslmode=$(DB_SSLMODE_FINAL)
endif

# Git
DESIRED_ORIGIN_URL=git@github.com-17thearth:SeventeenthEarth/sudal.git
GIT_USER_NAME="17thearth"
GIT_USER_EMAIL="17thearth@gmail.com"

# Tools (ensure they are installed or handle installation in init)
GOLANGCILINT=$(shell command -v golangci-lint 2> /dev/null)
PROTOC_GEN_GO=$(shell command -v protoc-gen-go 2> /dev/null)
PROTOC_GEN_CONNECT_GO=$(shell command -v protoc-gen-connect-go 2> /dev/null)
PROTOC_GEN_OPENAPIV2=$(shell command -v protoc-gen-openapiv2 2> /dev/null)
BUF=$(shell command -v buf 2> /dev/null)
GINKGO=$(shell command -v ginkgo 2> /dev/null)
MOCKGEN=$(shell command -v mockgen 2> /dev/null)
MIGRATE=$(shell command -v migrate 2> /dev/null)

# Default goal
.DEFAULT_GOAL := help

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

init: install-tools ## Initialize development environment (install tools)
	@echo "--- Initializing environment ---"
	go mod download

	# Install protoc-gen-go
	@echo "Checking/Installing protoc-gen-go..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@echo "protoc-gen-go installed/updated."

	# Install protoc-gen-connect-go
	@echo "Checking/Installing protoc-gen-connect-go..."
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	@echo "protoc-gen-connect-go installed/updated."

	# Install protoc-gen-openapiv2
	@echo "Checking/Installing protoc-gen-openapiv2..."
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "protoc-gen-openapiv2 installed/updated."

	# Install buf
	@echo "Checking/Installing buf..."
	@if [ -z "$(BUF)" ]; then \
		echo "buf not found. Installing..."; \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
		@echo "buf installed successfully."; \
	else \
		echo "buf already installed."; \
	fi

	@echo "Checking Git remote 'origin'..."
	@if ! git remote | grep -q '^origin$$'; then \
		echo "Remote 'origin' not found. Adding..."; \
		git remote add origin $(DESIRED_ORIGIN_URL); \
		echo "Remote 'origin' added successfully."; \
	else \
		CURRENT_ORIGIN_URL=$$(git remote get-url origin); \
		echo "Remote 'origin' already exists with URL: $$CURRENT_ORIGIN_URL"; \
		if [ "$$CURRENT_ORIGIN_URL" != "$(DESIRED_ORIGIN_URL)" ]; then \
			echo "Updating remote 'origin' URL to $(DESIRED_ORIGIN_URL)..."; \
			git remote set-url origin $(DESIRED_ORIGIN_URL); \
			echo "Remote 'origin' URL updated."; \
		else \
			echo "Remote 'origin' URL is already correct."; \
		fi \
	fi
	@echo "Configuring local Git user name and email..."
	git config --local user.name "$(GIT_USER_NAME)"
	git config --local user.email "$(GIT_USER_EMAIL)"
	@echo "Local Git user configured: $(GIT_USER_NAME) <$(GIT_USER_EMAIL)>"

	@echo "--- Environment initialized ---"

build: ## Build the application binary
	@echo "--- Building application ---"
	go build -ldflags="-s -w" -o $(OUTPUT_DIR)/$(BINARY_NAME) $(CMD_PATH)/main.go
	@echo "Binary available at $(OUTPUT_DIR)/$(BINARY_NAME)"

# Common test preparation steps
test.prepare: generate fmt vet lint ## Prepare for running tests (format, vet, lint, generate)
	@echo "--- Test preparation completed ---"

# Run all tests (unit and integration)
test: test.prepare test.unit.only test.int.only ## Run all tests (unit and integration)
	@echo "--- All tests completed ---"

# Unit tests with preparation steps
test.unit: test.prepare test.unit.only ## Run unit tests with preparation steps

# Unit tests only (without preparation steps)
test.unit.only: ## Run only unit tests without preparation steps
	@echo "--- Running unit tests ---"
ifeq ($(GINKGO),)
	@echo "Ginkgo not found. Running tests with 'go test'..."
	go test -v -race -coverprofile=coverage.unit.out `go list ./internal/... | grep -v "/mocks"` || { echo "Unit tests failed"; exit 1; }
	go tool cover -func=coverage.unit.out
	go tool cover -html=coverage.unit.out -o coverage.unit.html
else
	@echo "Running unit tests with Ginkgo..."
	$(GINKGO) -v -race -cover --coverprofile=coverage.unit.out --trace --fail-on-pending --randomize-all ./internal/... || { echo "Unit tests failed"; exit 1; }
	go tool cover -func=coverage.unit.out
	go tool cover -html=coverage.unit.out -o coverage.unit.html
endif
	@echo "--- Unit tests finished ---"
	@echo "Unit test coverage report generated at coverage.unit.html"

# Integration tests with preparation steps
test.int: test.prepare test.int.only ## Run integration tests with preparation steps

# Integration tests only (without preparation steps)
test.int.only: ## Run only integration tests without preparation steps
	@echo "--- Running integration tests ---"
ifeq ($(GINKGO),)
	@echo "Ginkgo not found. Running tests with 'go test'..."
	go test -v -race -coverprofile=coverage.int.out -coverpkg=github.com/seventeenthearth/sudal/internal/... ./test/integration || { echo "Integration tests failed"; exit 1; }
	go tool cover -func=coverage.int.out
	go tool cover -html=coverage.int.out -o coverage.int.html
else
	@echo "Running integration tests with Ginkgo..."
	$(GINKGO) -v -race -cover -coverpkg=github.com/seventeenthearth/sudal/internal/... --coverprofile=coverage.int.out --trace --fail-on-pending --randomize-all ./test/integration || { echo "Integration tests failed"; exit 1; }
	go tool cover -func=coverage.int.out
	go tool cover -html=coverage.int.out -o coverage.int.html
endif
	@echo "--- Integration tests finished ---"
	@echo "Integration test coverage report generated at coverage.int.html"

# End-to-end tests with preparation steps
test.e2e: test.prepare test.e2e.only ## Run end-to-end tests with preparation steps

# End-to-end tests only (without preparation steps)
test.e2e.only: ## Run only end-to-end tests without preparation steps
	@echo "--- Running end-to-end tests (Go/testify BDD) ---"
	@echo "Checking if server is running..."
	@if ! curl -s "http://localhost:8080/ping" > /dev/null; then \
		echo "Warning: The server doesn't appear to be running on port 8080."; \
		echo "Run 'make run' in a separate terminal before running e2e tests."; \
	fi
	@echo "Note: Coverage data cannot be collected from E2E tests as they test an external server."
	@echo "      Use unit and integration tests for code coverage measurement."
	@echo "Running Go E2E tests with testify BDD style..."
	@echo "Running tests with 'go test'..."
	go test -v -race ./test/e2e || { echo "End-to-end tests failed"; exit 1; }
	@echo "--- End-to-end tests finished ---"

mock-clean: ## Clean generated mock files
	@echo "--- Cleaning mock files ---"
	rm -rf ./internal/mocks
	@echo "--- Mock files cleaned ---"

ginkgo-clean: ## Clean generated Ginkgo test suite files
	@echo "--- Cleaning Ginkgo test suite files ---"
	find . -name "*_suite_test.go" -delete
	@echo "--- Ginkgo test suite files cleaned ---"

tmp-clean: ## Clean temporary files created by development tools
	@echo "--- Cleaning temporary files ---"
	rm -rf tmp/
	rm -f .air.log .air.toml.tmp
	rm -f .dockerignore.tmp
	rm -f .compiledaemon.*
	@echo "--- Temporary files cleaned ---"

clean: proto-clean mock-clean ginkgo-clean wire-clean tmp-clean ogen-clean ## Clean build artifacts, test files, mocks, wire, and caches
	@echo "--- Cleaning ---"
	rm -rf $(OUTPUT_DIR)
	# Remove test coverage files
	rm -f coverage*.out coverage*.html coverprofile.out
	# Clean Go test cache (but not module cache, which can cause errors)
	go clean -testcache
	@echo "--- Clean finished ---"

clean-all: clean ## Clean everything including Go module cache (may fail if modules are in use)
	@echo "--- Cleaning Go module cache ---"
	go clean -modcache || echo "Warning: Could not clean Go module cache completely. This is normal if modules are in use."
	@echo "--- All cleaning finished ---"

fmt: ## Format Go code
	@echo "--- Formatting code ---"
	go fmt ./...
	@echo "--- Formatting finished ---"

vet: ## Run Go vet
	@echo "--- Running go vet ---"
	go vet ./...
	@echo "--- Go vet finished ---"

lint: ## Run linter checks
	@echo "--- Running linter ---"
ifndef GOLANGCILINT
	@echo "golangci-lint not found. Run 'make install-tools' first."
	@exit 1
endif
	$(GOLANGCILINT) run ./... || echo "Warning: Linter found issues, but continuing..."
	@echo "--- Linter finished ---"

generate: ginkgo-bootstrap buf-generate wire-gen generate-mocks ogen-generate ## Generate all code (test suites, proto, wire, mocks, openapi)
	@echo "--- All code generation completed ---"

proto-clean: ## Clean generated Protocol Buffer files
	@echo "--- Cleaning generated Protocol Buffer files ---"
	rm -rf ./gen
	find ./proto -name "*.pb.go" -delete
	find ./proto -name "*.connect.go" -delete
	find ./proto -path "*/*/healthv1connect" -type d -exec rm -rf {} \; 2>/dev/null || true
	@echo "--- Protocol Buffer files cleaned ---"

buf-generate: proto-clean buf-lint ## Generate code from Protobuf definitions using Buf
	@echo "--- Generating code from Proto definitions using Buf ---"
	@echo "Checking for buf and required plugins..."
	@if [ -z "$(BUF)" ]; then \
		echo "buf not found. Installing..."; \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
	fi
	@if [ -z "$(PROTOC_GEN_GO)" ]; then \
		echo "protoc-gen-go not found. Installing..."; \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	fi
	@if [ -z "$(PROTOC_GEN_CONNECT_GO)" ]; then \
		echo "protoc-gen-connect-go not found. Installing..."; \
		go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest; \
	fi
	@if [ -z "$(PROTOC_GEN_OPENAPIV2)" ]; then \
		echo "protoc-gen-openapiv2 not found. Installing..."; \
		go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest; \
	fi
	@echo "Running buf generate..."
	@if [ ! -f "buf.work.yaml" ]; then \
		echo "Creating buf.work.yaml..."; \
		echo "version: v1\ndirectories:\n  - proto" > buf.work.yaml; \
	fi
	@if [ ! -f "proto/buf.yaml" ]; then \
		echo "Creating proto/buf.yaml..."; \
		mkdir -p proto; \
		echo "version: v1\nname: github.com/seventeenthearth/sudal\ndeps:\n  - buf.build/googleapis/googleapis\n  - buf.build/grpc-ecosystem/grpc-gateway\nlint:\n  use:\n    - STANDARD\nbreaking:\n  use:\n    - FILE" > proto/buf.yaml; \
	fi
	@if [ ! -f "proto/buf.gen.yaml" ]; then \
		echo "Creating proto/buf.gen.yaml..."; \
		echo "version: v1\nplugins:\n  # Generate Go structs from Protocol Buffers\n  - plugin: go\n    out: ../gen/go\n    opt: paths=source_relative\n  \n  # Generate Connect-go service interfaces, clients, and handlers\n  - plugin: connect-go\n    out: ../gen/go\n    opt: paths=source_relative\n  \n  # Generate OpenAPI v2 specifications from Protocol Buffers\n  - plugin: openapiv2\n    out: ../gen/openapi\n    opt:\n      - output_format=yaml\n      - allow_merge=true\n      - merge_file_name=api" > proto/buf.gen.yaml; \
	fi
	@cd proto && buf generate
	@echo "Buf code generation completed."

buf-lint: ## Lint Protocol Buffer definitions using Buf
	@echo "--- Linting Protocol Buffer definitions using Buf ---"
	@if [ -z "$(BUF)" ]; then \
		echo "buf not found. Installing..."; \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
	fi
	@echo "Running buf lint..."
	@if [ ! -f "buf.work.yaml" ]; then \
		echo "Creating buf.work.yaml..."; \
		echo "version: v1\ndirectories:\n  - proto" > buf.work.yaml; \
	fi
	@if [ ! -f "proto/buf.yaml" ]; then \
		echo "Creating proto/buf.yaml..."; \
		mkdir -p proto; \
		echo "version: v1\nname: github.com/seventeenthearth/sudal\ndeps:\n  - buf.build/googleapis/googleapis\n  - buf.build/grpc-ecosystem/grpc-gateway\nlint:\n  use:\n    - STANDARD\nbreaking:\n  use:\n    - FILE" > proto/buf.yaml; \
	fi
	@if [ ! -f "proto/buf.gen.yaml" ]; then \
		echo "Creating proto/buf.gen.yaml..."; \
		echo "version: v1\nplugins:\n  # Generate Go structs from Protocol Buffers\n  - plugin: go\n    out: ../gen/go\n    opt: paths=source_relative\n  \n  # Generate Connect-go service interfaces, clients, and handlers\n  - plugin: connect-go\n    out: ../gen/go\n    opt: paths=source_relative\n  \n  # Generate OpenAPI v2 specifications from Protocol Buffers\n  - plugin: openapiv2\n    out: ../gen/openapi\n    opt:\n      - output_format=yaml\n      - allow_merge=true\n      - merge_file_name=api" > proto/buf.gen.yaml; \
	fi
	@cd proto && buf lint
	@echo "Buf linting completed."

buf-breaking: ## Check for breaking changes in Protocol Buffer definitions using Buf
	@echo "--- Checking for breaking changes in Protocol Buffer definitions using Buf ---"
	@if [ -z "$(BUF)" ]; then \
		echo "buf not found. Installing..."; \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
	fi
	@echo "Running buf breaking check..."
	@if [ ! -d "./.git" ]; then \
		echo "Error: Not a git repository. Cannot check for breaking changes."; \
		exit 0; \
	fi
	@if ! git ls-files --error-unmatch proto > /dev/null 2>&1; then \
		echo "Warning: No proto files tracked in git. Skipping breaking change check."; \
		exit 0; \
	fi
	@if [ ! -f "buf.work.yaml" ]; then \
		echo "Creating buf.work.yaml..."; \
		echo "version: v1\ndirectories:\n  - proto" > buf.work.yaml; \
	fi
	@if [ ! -f "proto/buf.yaml" ]; then \
		echo "Creating proto/buf.yaml..."; \
		mkdir -p proto; \
		echo "version: v1\nname: github.com/seventeenthearth/sudal\ndeps:\n  - buf.build/googleapis/googleapis\n  - buf.build/grpc-ecosystem/grpc-gateway\nlint:\n  use:\n    - STANDARD\nbreaking:\n  use:\n    - FILE" > proto/buf.yaml; \
	fi
	@if [ ! -f "proto/buf.gen.yaml" ]; then \
		echo "Creating proto/buf.gen.yaml..."; \
		echo "version: v1\nplugins:\n  # Generate Go structs from Protocol Buffers\n  - plugin: go\n    out: ../gen/go\n    opt: paths=source_relative\n  \n  # Generate Connect-go service interfaces, clients, and handlers\n  - plugin: connect-go\n    out: ../gen/go\n    opt: paths=source_relative\n  \n  # Generate OpenAPI v2 specifications from Protocol Buffers\n  - plugin: openapiv2\n    out: ../gen/openapi\n    opt:\n      - output_format=yaml\n      - allow_merge=true\n      - merge_file_name=api" > proto/buf.gen.yaml; \
	fi
	@if ! git show-ref --verify --quiet refs/heads/main; then \
		echo "Warning: Main branch not found. Using current branch as reference."; \
		GIT_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
		cd proto && buf breaking --against "../.git#branch=$$GIT_BRANCH"; \
	else \
		echo "Checking against main branch..."; \
		cd proto && buf breaking --against '../.git#branch=main'; \
	fi
	@echo "Buf breaking change check completed."

run: ## Run the application using Docker Compose
	@echo "--- Running application with Docker Compose ---"
	docker-compose up --build

# Add placeholder main.go to make build work initially
$(CMD_PATH)/main.go:
	@mkdir -p $(@D)
	@echo "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Starting Sudal Server...\")\n}" > $@
	go mod tidy # Add fmt dependency

# Ensure build depends on the placeholder being created if it doesn't exist
build: $(CMD_PATH)/main.go

install-tools: ## Install development tools
	@echo "--- Installing development tools ---"
	# Install golangci-lint
	@echo "Checking/Installing golangci-lint..."
ifeq ($(GOLANGCILINT),)
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(eval GOLANGCILINT=$(shell go env GOPATH)/bin/golangci-lint)
	@if [ -z "$(GOLANGCILINT)" ]; then echo "Failed to install golangci-lint"; exit 1; fi
	@echo "golangci-lint installed successfully."
else
	@echo "golangci-lint already installed."
endif

	# Install Ginkgo
	@echo "Checking/Installing Ginkgo..."
ifeq ($(GINKGO),)
	@echo "Installing Ginkgo..."
	go install github.com/onsi/ginkgo/v2/ginkgo@latest
	$(eval GINKGO=$(shell go env GOPATH)/bin/ginkgo)
	@if [ -z "$(GINKGO)" ]; then echo "Failed to install Ginkgo"; exit 1; fi
	@echo "Ginkgo installed successfully."
else
	@echo "Ginkgo already installed."
endif

	# Install mockgen
	@echo "Checking/Installing mockgen..."
ifeq ($(MOCKGEN),)
	@echo "Installing mockgen..."
	go install go.uber.org/mock/mockgen@latest
	$(eval MOCKGEN=$(shell go env GOPATH)/bin/mockgen)
	@if [ -z "$(MOCKGEN)" ]; then echo "Failed to install mockgen"; exit 1; fi
	@echo "mockgen installed successfully."
else
	@echo "mockgen already installed."
endif
	@echo "--- Development tools installed ---"

ginkgo-bootstrap: ginkgo-clean ## Bootstrap Ginkgo test suites in all packages with tests
	@echo "--- Bootstrapping Ginkgo test suites ---"
	./scripts/setup_tests.sh
	@echo "--- Ginkgo test suites bootstrapped ---"

wire-clean: ## Clean generated wire code
	@echo "--- Cleaning wire generated code ---"
	rm -f internal/infrastructure/di/wire_gen.go
	@echo "--- Wire generated code cleaned ---"

wire-gen: wire-clean ## Generate dependency injection code using wire
	@echo "--- Generating wire dependency injection code ---"
	@if ! command -v wire >/dev/null 2>&1; then \
		echo "wire not found. Installing..."; \
		GOPROXY=direct go install github.com/google/wire/cmd/wire@latest; \
	fi
	@cd internal/infrastructure/di && wire
	@echo "--- Wire code generation completed ---"

generate-mocks: mock-clean ## Generate mocks using mockgen
	@echo "--- Generating mocks ---"
ifeq ($(MOCKGEN),)
	@echo "mockgen not found. Run 'make install-tools' first."
	@exit 1
endif
	@echo "Running go generate to create mocks..."
	@go generate ./... || echo "Warning: Some mock generation may have failed, but continuing..."
	@echo "Mock generation completed"
	@echo "--- Mocks generated ---"

ogen-clean: ## Clean generated OpenAPI code
	@echo "--- Cleaning generated OpenAPI code ---"
	rm -rf internal/infrastructure/openapi/oas_*.go
	@echo "--- OpenAPI generated code cleaned ---"

ogen-generate: ogen-clean ## Generate OpenAPI server code using ogen
	@echo "--- Generating OpenAPI server code using ogen ---"
	@echo "Checking for ogen..."
	@if ! command -v ogen >/dev/null 2>&1; then \
		echo "ogen not found. Installing..."; \
		go install github.com/ogen-go/ogen/cmd/ogen@latest; \
	fi
	@echo "Running ogen to generate OpenAPI server code..."
	go run github.com/ogen-go/ogen/cmd/ogen -target internal/infrastructure/openapi -package openapi -clean api/openapi.yaml
	@echo "--- OpenAPI code generation completed ---"

# Database Migration Targets

migrate-up: ## Apply all pending database migrations
	@echo "--- Applying database migrations ---"
	@echo "Database URL: $(DATABASE_URL)"
	@echo "Migrations directory: $(MIGRATIONS_DIR)"
	@if [ -z "$(MIGRATE)" ]; then \
		echo "migrate CLI not found. Installing..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@if [ ! -d "$(MIGRATIONS_DIR)" ]; then \
		echo "Error: Migrations directory $(MIGRATIONS_DIR) does not exist"; \
		echo "Run 'make migrate-create DESC=initial_setup' to create your first migration"; \
		exit 1; \
	fi
	@echo "Running migrations..."
	@if [ -z "$(MIGRATE)" ]; then \
		$$(go env GOPATH)/bin/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up; \
	else \
		$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up; \
	fi
	@echo "--- Database migrations applied successfully ---"

migrate-down: ## Rollback the last database migration
	@echo "--- Rolling back last database migration ---"
	@echo "Database URL: $(DATABASE_URL)"
	@echo "Migrations directory: $(MIGRATIONS_DIR)"
	@if [ -z "$(MIGRATE)" ]; then \
		echo "migrate CLI not found. Installing..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@echo "WARNING: This will rollback the last applied migration!"
	@echo "Press Ctrl+C to cancel, or Enter to continue..."
	@read dummy
	@if [ -z "$(MIGRATE)" ]; then \
		$$(go env GOPATH)/bin/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1; \
	else \
		$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1; \
	fi
	@echo "--- Last migration rolled back successfully ---"

migrate-status: ## Show current migration status
	@echo "--- Database migration status ---"
	@echo "Database URL: $(DATABASE_URL)"
	@echo "Migrations directory: $(MIGRATIONS_DIR)"
	@if [ -z "$(MIGRATE)" ]; then \
		echo "migrate CLI not found. Installing..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@if [ ! -d "$(MIGRATIONS_DIR)" ]; then \
		echo "Migrations directory $(MIGRATIONS_DIR) does not exist"; \
		exit 1; \
	fi
	@echo "Current migration version:"
	@if [ -z "$(MIGRATE)" ]; then \
		$$(go env GOPATH)/bin/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version || echo "No migrations applied yet"; \
	else \
		$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version || echo "No migrations applied yet"; \
	fi
	@echo ""
	@echo "Available migration files:"
	@ls -la $(MIGRATIONS_DIR)/ 2>/dev/null || echo "No migration files found"

migrate-version: ## Show current migration version
	@echo "--- Current migration version ---"
	@if [ -z "$(MIGRATE)" ]; then \
		echo "migrate CLI not found. Installing..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@if [ -z "$(MIGRATE)" ]; then \
		$$(go env GOPATH)/bin/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version; \
	else \
		$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version; \
	fi

migrate-force: ## Force set migration version (use with caution)
	@echo "--- Force set migration version ---"
	@echo "WARNING: This is a dangerous operation that should only be used for recovery!"
	@echo "Current version:"
	@$(MAKE) migrate-version || echo "Could not determine current version"
	@echo ""
	@echo "Enter the version number to force set (or Ctrl+C to cancel):"
	@read version; \
	if [ -z "$$version" ]; then \
		echo "No version provided. Cancelling."; \
		exit 1; \
	fi; \
	echo "Setting migration version to $$version..."; \
	if [ -z "$(MIGRATE)" ]; then \
		echo "migrate CLI not found. Installing..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
		$$(go env GOPATH)/bin/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $$version; \
	else \
		$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $$version; \
	fi
	@echo "--- Migration version forced successfully ---"

migrate-create: ## Create new migration files (usage: make migrate-create DESC=description)
	@echo "--- Creating new migration files ---"
	@if [ -z "$(DESC)" ]; then \
		echo "Error: DESC parameter is required"; \
		echo "Usage: make migrate-create DESC=create_users_table"; \
		exit 1; \
	fi
	@if [ -z "$(MIGRATE)" ]; then \
		echo "migrate CLI not found. Installing..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@mkdir -p $(MIGRATIONS_DIR)
	@echo "Creating migration files for: $(DESC)"
	@if [ -z "$(MIGRATE)" ]; then \
		$$(go env GOPATH)/bin/migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(DESC); \
	else \
		$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(DESC); \
	fi
	@echo "--- Migration files created successfully ---"
	@echo "Edit the generated .up.sql and .down.sql files in $(MIGRATIONS_DIR)/"

migrate-squash: ## Squash multiple migrations into one (DANGEROUS - use only in development)
	@echo "--- Migration Squashing Tool ---"
	@echo "⚠️  WARNING: This is a DANGEROUS operation!"
	@echo "⚠️  Only use this in development environments!"
	@echo "⚠️  Make sure to backup your database first!"
	@echo ""
	@echo "Current migration status:"
	@$(MAKE) migrate-status
	@echo ""
	@echo "This will:"
	@echo "1. Create a backup of existing migrations"
	@echo "2. Create a schema dump of the current database"
	@echo "3. Create a new squashed migration file"
	@echo "4. Test the squashed migration on a fresh database"
	@echo ""
	@echo "Do you want to continue? (y/N):"
	@read confirm; \
	if [ "$$confirm" != "y" ] && [ "$$confirm" != "Y" ]; then \
		echo "Operation cancelled."; \
		exit 0; \
	fi; \
	echo "Creating backup of existing migrations..."; \
	mkdir -p db/migrations_backup_$$(date +%Y%m%d_%H%M%S); \
	cp db/migrations/*.sql db/migrations_backup_$$(date +%Y%m%d_%H%M%S)/; \
	echo "Creating schema dump..."; \
	docker exec sudal-db pg_dump -U user -d quizapp_db --schema-only --no-owner --no-privileges --exclude-table=schema_migrations > /tmp/schema_dump.sql; \
	echo "Creating test database..."; \
	docker exec sudal-db dropdb -U user --if-exists test_squashed_migration; \
	docker exec sudal-db createdb -U user test_squashed_migration; \
	echo ""; \
	echo "✅ Backup created in db/migrations_backup_$$(date +%Y%m%d_%H%M%S)/"; \
	echo "✅ Schema dump created at: /tmp/schema_dump.sql"; \
	echo "✅ Test database 'test_squashed_migration' created"; \
	echo ""; \
	echo "Next steps:"; \
	echo "1. Create your squashed migration files manually"; \
	echo "2. Test: DB_NAME=test_squashed_migration make migrate-up"; \
	echo "3. Compare schemas to ensure they match"; \
	echo "4. If successful, replace old migrations with squashed ones"; \
	echo "5. Reset migration version: make migrate-force"

# Database Reset and Fresh Migration Targets

migrate-drop: ## Drop all database objects using DROP SCHEMA CASCADE (DANGEROUS)
	@echo "--- Dropping all database objects ---"
	@echo "⚠️  WARNING: This will DELETE ALL DATA and ALL OBJECTS in the database!"
	@echo "⚠️  This includes: tables, views, functions, triggers, sequences, types, etc."
	@echo "⚠️  This action is IRREVERSIBLE!"
	@echo ""
	@echo "Database: $(DATABASE_URL)"
	@echo ""
	@echo "Do you want to continue? Type 'DROP ALL DATA' to confirm:"
	@read confirm; \
	if [ "$$confirm" != "DROP ALL DATA" ]; then \
		echo "Operation cancelled. You must type exactly 'DROP ALL DATA' to confirm."; \
		exit 0; \
	fi; \
	echo "Dropping public schema and all objects..."; \
	docker exec sudal-db psql -U user -d quizapp_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO \"user\"; GRANT ALL ON SCHEMA public TO public;"; \
	echo "--- All database objects dropped successfully ---"

migrate-reset: ## Reset database to clean state and reapply all migrations
	@echo "--- Resetting database ---"
	@echo "This will:"
	@echo "1. Drop all database objects (tables, views, functions, etc.)"
	@echo "2. Reapply all migrations from scratch"
	@echo ""
	@echo "Do you want to continue? (y/N):"
	@read confirm; \
	if [ "$$confirm" != "y" ] && [ "$$confirm" != "Y" ]; then \
		echo "Operation cancelled."; \
		exit 0; \
	fi; \
	echo "Dropping all database objects..."; \
	docker exec sudal-db psql -U user -d quizapp_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO \"user\"; GRANT ALL ON SCHEMA public TO public;" || echo "Schema drop failed or already clean"; \
	echo "Reapplying all migrations..."; \
	$(MAKE) migrate-up; \
	echo "--- Database reset completed ---"

migrate-fresh: ## Fresh migration setup - backup old migrations and start clean
	@echo "--- Fresh Migration Setup ---"
	@echo "This will:"
	@echo "1. Create backup of current migration files"
	@echo "2. Drop all database tables"
	@echo "3. Clear migration files directory"
	@echo "4. You can then create new migrations from scratch"
	@echo ""
	@echo "⚠️  WARNING: This will backup and remove all current migration files!"
	@echo ""
	@echo "Do you want to continue? (y/N):"
	@read confirm; \
	if [ "$$confirm" != "y" ] && [ "$$confirm" != "Y" ]; then \
		echo "Operation cancelled."; \
		exit 0; \
	fi; \
	echo "Creating backup of migration files..."; \
	BACKUP_DIR="db/migrations_backup_fresh_$$(date +%Y%m%d_%H%M%S)"; \
	mkdir -p "$$BACKUP_DIR"; \
	if [ -n "$$(ls -A $(MIGRATIONS_DIR) 2>/dev/null)" ]; then \
		cp $(MIGRATIONS_DIR)/*.sql "$$BACKUP_DIR/" 2>/dev/null || echo "No .sql files to backup"; \
	fi; \
	echo "Dropping all database objects..."; \
	docker exec sudal-db psql -U user -d quizapp_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO \"user\"; GRANT ALL ON SCHEMA public TO public;" || echo "Schema drop failed or already clean"; \
	echo "Clearing migration files..."; \
	rm -f $(MIGRATIONS_DIR)/*.sql; \
	echo ""; \
	echo "✅ Fresh setup completed!"; \
	echo "✅ Migration files backed up to: $$BACKUP_DIR"; \
	echo "✅ Database tables dropped"; \
	echo "✅ Migration directory cleared"; \
	echo ""; \
	echo "Next steps:"; \
	echo "1. Create your first migration: make migrate-create DESC=initial_schema"; \
	echo "2. Edit the migration files"; \
	echo "3. Apply migrations: make migrate-up"
