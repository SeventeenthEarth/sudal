# Scripts Guide - Sudal Project

This document provides a comprehensive guide to all the scripts and make targets available in the Sudal project.

## 🚀 Quick Start

```bash
# Initialize development environment
make init

# Install development tools
make install-tools

# Generate all code
make generate

# Run tests
make test

# Run the application
make run
```

## 📋 Available Make Targets

### Development Environment

- `make init` - Initialize development environment (quick setup)
- `make install-tools` - Install all development tools

### Build & Test

- `make build` - Build the application
- `make test` - Run all tests (unit and integration)
- `make test.unit` - Run unit tests only
- `make test.int` - Run integration tests only
- `make test.e2e` - Run end-to-end tests

### Code Quality

- `make fmt` - Format Go code
- `make vet` - Run Go vet
- `make lint` - Run linter checks

### Code Generation

- `make generate` - Generate all code (test suites, proto, wire, mocks)
- `make buf-setup` - Setup buf configuration files
- `make buf-generate` - Generate protobuf code
- `make buf-lint` - Lint protobuf files
- `make buf-breaking` - Check for breaking changes in protobuf
- `make wire-gen` - Generate dependency injection code
- `make generate-mocks` - Generate mocks
- `make ginkgo-bootstrap` - Bootstrap Ginkgo test suites

### Cleanup Operations

- `make clean` - Standard cleanup (build artifacts and generated files)
- `make clean-all` - Complete cleanup including Go module cache
- `make clean-proto` - Clean only Protocol Buffer generated files
- `make clean-mocks` - Clean only generated mock files
- `make clean-ginkgo` - Clean only Ginkgo test suite files
- `make clean-wire` - Clean only Wire generated code
- `make clean-ogen` - Clean only OpenAPI generated code
- `make clean-tmp` - Clean only temporary files
- `make clean-build` - Clean only build artifacts
- `make clean-coverage` - Clean only test coverage files
- `make clean-go-cache` - Clean only Go test cache
- `make clean-go-modules` - Clean only Go module cache

### Database Operations

- `make migrate-up` - Apply database migrations
- `make migrate-down` - Rollback last migration
- `make migrate-status` - Show migration status
- `make migrate-version` - Show current migration version
- `make migrate-force VERSION=5` - Force set migration version
- `make migrate-create DESC=description` - Create new migration
- `make migrate-reset` - Reset database to clean state and reapply all migrations
- `make migrate-drop` - Drop all database objects (DANGEROUS)
- `make migrate-fresh` - Fresh migration setup - backup old migrations and start clean

### Application

- `make run` - Run the application using Docker Compose

## 🔧 Direct Script Usage

All scripts are also available for direct execution with advanced options:

### ./scripts/setup-dev-env.sh

Advanced environment setup with Git configuration.

```bash
# Basic usage
./scripts/setup-dev-env.sh

# With custom Git settings
./scripts/setup-dev-env.sh --git-user "Your Name" --git-email "your@email.com"

# Help
./scripts/setup-dev-env.sh --help
```

### ./scripts/install-tools.sh

Install all development tools with detailed output.

```bash
# Install all tools
./scripts/install-tools.sh

# The script automatically detects and installs:
# - golangci-lint, ginkgo, mockgen, wire
# - protoc-gen-go, protoc-gen-connect-go, protoc-gen-openapiv2
# - buf, migrate, ogen
```

### ./scripts/setup-buf.sh

Manage buf configuration and operations.

```bash
# Setup buf configuration files
./scripts/setup-buf.sh setup

# Generate protobuf code
./scripts/setup-buf.sh generate

# Lint protobuf files
./scripts/setup-buf.sh lint

# Check for breaking changes
./scripts/setup-buf.sh breaking

# Help
./scripts/setup-buf.sh --help
```

### ./scripts/migrate.sh

Database migration operations.

```bash
# Apply all pending migrations
./scripts/migrate.sh up

# Rollback last migration
./scripts/migrate.sh down

# Rollback last 3 migrations
./scripts/migrate.sh down 3

# Show migration status
./scripts/migrate.sh status

# Create new migration
./scripts/migrate.sh create create_users_table

# Force set migration version (dangerous!)
./scripts/migrate.sh force 5

# Help
./scripts/migrate.sh --help
```

### ./scripts/clean-all.sh

Advanced cleanup operations.

```bash
# Standard cleanup
./scripts/clean-all.sh standard

# Complete cleanup including Go module cache
./scripts/clean-all.sh all

# Clean specific types
./scripts/clean-all.sh proto
./scripts/clean-all.sh mocks
./scripts/clean-all.sh wire

# Help
./scripts/clean-all.sh --help
```

### ./scripts/run-e2e-tests.sh

Run end-to-end tests with server availability check.

```bash
# Run all E2E tests
./scripts/run-e2e-tests.sh

# Run a specific test
./scripts/run-e2e-tests.sh TestConnectGoHealthService

# Skip server availability check
./scripts/run-e2e-tests.sh --skip-check

# Show help
./scripts/run-e2e-tests.sh --help
```

### ./scripts/setup_tests.sh

Set up the test environment by installing required tools and bootstrapping Ginkgo test suites.

```bash
# Set up the test environment
./scripts/setup_tests.sh
```

## 🌟 Best Practices

1. **Use make targets for common operations** - They provide consistent output and error handling
2. **Use direct scripts for advanced options** - When you need specific parameters or custom behavior
3. **Always run `make help`** - To see the most up-to-date list of available targets
4. **Check script help** - Each script has detailed help: `./scripts/SCRIPT_NAME.sh --help`

## 🔍 Troubleshooting

### Common Issues

1. **Permission denied when running scripts**

   ```bash
   chmod +x scripts/*.sh
   ```

2. **Tools not found**

   ```bash
   make install-tools
   ```

3. **Database connection issues**
   - Check your `.env` file
   - Ensure Docker containers are running: `make run`

4. **Migration errors**

   ```bash
   # Check status first
   make migrate-status

   # If needed, force set version
   make migrate-force VERSION=0
   ```

### Getting Help

- `make help` - Show all available make targets
- `./scripts/SCRIPT_NAME.sh --help` - Detailed help for each script
- Check the logs for detailed error messages

## 📚 Additional Resources

- [Project README](../README.md) - Main project documentation
- [Development Setup](./DEVELOPMENT.md) - Detailed development setup guide
- [Database Guide](./DATABASE.md) - Database setup and migration guide
