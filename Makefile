.PHONY: help init build test clean lint fmt vet proto-gen run install-tools generate-mocks ginkgo-bootstrap generate

# Variables
BINARY_NAME=sudal-server
CMD_PATH=./cmd/server
OUTPUT_DIR=./bin
CONFIG_FILE?=./configs/config.yaml # Default config path, can be overridden

# Git
DESIRED_ORIGIN_URL=git@github.com-17thearth:SeventeenthEarth/sudal.git
GIT_USER_NAME="17thearth"
GIT_USER_EMAIL="17thearth@gmail.com"

# Tools (ensure they are installed or handle installation in init)
GOLANGCILINT=$(shell command -v golangci-lint 2> /dev/null)
PROTOC_GEN_GO=$(shell command -v protoc-gen-go 2> /dev/null)
PROTOC_GEN_CONNECT_GO=$(shell command -v protoc-gen-connect-go 2> /dev/null)
GINKGO=$(shell command -v ginkgo 2> /dev/null)
MOCKGEN=$(shell command -v mockgen 2> /dev/null)

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

test: fmt vet lint generate ## Run unit tests with Ginkgo and generate coverage report
	@echo "--- Running tests with coverage ---"
ifeq ($(GINKGO),)
	@echo "Ginkgo not found. Running tests with 'go test'..."
	go test -v -race -coverprofile=coverage.out `go list ./... | grep -v "/mocks" | grep -v "^github.com/seventeenthearth/sudal/cmd"` && \
	go tool cover -func=coverage.out && \
	go tool cover -html=coverage.out -o coverage.html
else
	@echo "Running tests with Ginkgo..."
	$(GINKGO) -r -v -race -cover --coverprofile=coverage.out --trace --fail-on-pending --randomize-all ./... && \
	go tool cover -func=coverage.out && \
	go tool cover -html=coverage.out -o coverage.html
endif
	@echo "--- Tests finished ---"
	@echo "Coverage report generated at coverage.html"

clean: ## Clean build artifacts and caches
	@echo "--- Cleaning ---"
	rm -rf $(OUTPUT_DIR)
	go clean -testcache -modcache
	@echo "--- Clean finished ---"

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

generate: ginkgo-bootstrap generate-mocks proto-gen ## Generate all code (mocks, test suites, proto)
	@echo "--- All code generation completed ---"

proto-gen: ## Generate code from Protobuf definitions (implement when needed)
	@echo "--- Generating code from Proto definitions ---"
	# Example command using buf (add buf installation to 'init' if used)
	# buf generate api/protobuf
	@echo "Implement proto generation command here" # Placeholder

run: build ## Build and run the application
	@echo "--- Running application ---"
	$(OUTPUT_DIR)/$(BINARY_NAME) --config=$(CONFIG_FILE)

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

ginkgo-bootstrap: ## Bootstrap Ginkgo test suites in all packages with tests
	@echo "--- Bootstrapping Ginkgo test suites ---"
	./scripts/setup_tests.sh
	@echo "--- Ginkgo test suites bootstrapped ---"

generate-mocks: ## Generate mocks using mockgen
	@echo "--- Generating mocks ---"
ifeq ($(MOCKGEN),)
	@echo "mockgen not found. Run 'make install-tools' first."
	@exit 1
endif
	@echo "Running go generate to create mocks..."
	go generate ./... || echo "Warning: Some mock generation commands failed, but continuing..."
	@echo "--- Mocks generated ---"
