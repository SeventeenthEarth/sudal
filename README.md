# Sudal - Social Quiz Platform

[![Go Report Card](https://goreportcard.com/badge/github.com/seventeenthearth/sudal)](https://goreportcard.com/report/github.com/seventeenthearth/sudal)
[![Go Version](https://img.shields.io/github/go-mod/go-version/SeventeenthEarth/sudal)](https://golang.org/dl/)

## Project Overview

Sudal is a modern social quiz platform that allows users to create, share, and participate in quizzes on various topics. The platform is designed to be scalable, maintainable, and follows clean architecture principles.

## Architecture Summary

The project follows a clean architecture approach with a feature-centric structure. Each feature is organized with the following components:

- **Domain:** Core business entities and rules.
- **Application:** Use cases orchestrating the flow of data.
- **Interfaces:** Adapters connecting the application to external elements (Handlers, Repository Interfaces).

The following components are independent from features:

- **Infrastructure:** Concrete implementations for databases, external services, logging, configuration, etc.
- **API:** Defined using Protocol Buffers (`.proto`) and served via connect-go.

The codebase is organized into the following directories:

- `/cmd`: Application entry points
- `/internal`: Private application code
- `/pkg`: Public library code
- `/api`: Protocol definition files (Connect/gRPC)
- `/configs`: Configuration files
- `/scripts`: Build and utility scripts
- `/docs`: Documentation files
- `/test`: Additional test utilities

## Quick Start

### Prerequisites

- Go 1.24 or higher
- Make

### Installation

1. Clone the repository:
   ```bash
   git clone git@github.com:SeventeenthEarth/sudal.git
   cd sudal
   ```

2. Initialize the development environment:
   ```bash
   make init
   ```

3. Build the application:
   ```bash
   make build
   ```

4. Run the application using Docker Compose:
   ```bash
   make run
   ```

## Development Setup

### Environment Setup

1. **Go:** Version 1.22 or higher required.
2. **Dependencies:** Managed using Go Modules. Run `make init` or `go mod download`.
3. **Tools:**
   - `golangci-lint` for linting (install via `make install-tools` or manually).
   - `ginkgo` for BDD-style testing (install via `make install-tools` or manually).
   - `mockgen` for generating mocks (install via `make install-tools` or manually).
   - `buf` (optional but recommended for Protobuf management).
   - `protoc-gen-go`, `protoc-gen-connect-go` (for code generation from `.proto`).

### Make Commands

The following `make` commands are available for development:

- `make help`: Display available commands and descriptions.
- `make init`: Initialize the development environment (e.g., install tools).
- `make install-tools`: Install development tools (golangci-lint, ginkgo, mockgen).
- `make build`: Compile the main server application into the `./bin/` directory.
- `make lint`: Run the `golangci-lint` checks.
- `make test`: Run all unit and integration tests (runs preparation steps only once).
- `make test.prepare`: Prepare for running tests (format, vet, lint, generate code).
- `make test.unit`: Run unit tests with preparation steps.
- `make test.int`: Run integration tests with preparation steps.
- `make test.e2e`: Run end-to-end tests with preparation steps.
- `make clean`: Remove build artifacts and caches.
- `make generate`: Run all code generation tasks (mocks, test suites, proto).
- `make run`: Run the application using Docker Compose with `docker-compose up --build`.

## Testing

### Running Tests

The project has three types of tests:

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test interactions between components
3. **End-to-End Tests**: Test the entire system with a running server

#### Running All Tests

To run both unit and integration tests (with preparation steps run only once):

```bash
make test
```

#### Running Specific Test Types

To run only unit tests (with preparation steps):

```bash
make test.unit
```

To run only unit tests (without preparation steps):

```bash
make test.unit.only
```

To run only integration tests (with preparation steps):

```bash
make test.int
```

To run only integration tests (without preparation steps):

```bash
make test.int.only
```

To run end-to-end tests (with preparation steps):

```bash
make test.e2e
```

To run end-to-end tests (without preparation steps):

```bash
make test.e2e.only
```

#### What Each Test Command Does

Each test command first runs `make test.prepare`, which:
1. Formats the code with `go fmt`
2. Runs static analysis with `go vet`
3. Runs linter checks with `golangci-lint`
4. Runs all code generation tasks via `make generate`

Then, the test command:
1. Runs the specified tests with Ginkgo
2. Generates a coverage report (both console summary and HTML report)

After running tests, you can view the detailed coverage reports:
- Unit tests: `coverage.unit.html`
- Integration tests: `coverage.int.html`
- End-to-end tests: `coverage.e2e.html`

Note: Integration and E2E tests measure coverage of the internal packages using the `-coverpkg` flag.

To run specific tests manually:

```bash
go test ./path/to/package -v
```

### Testing Strategy

The project follows a Behavior-Driven Development (BDD) approach to testing using the following tools:

- **Ginkgo**: A BDD-style testing framework for Go that provides a more expressive and readable syntax for writing tests.
- **Gomega**: An assertion library that pairs with Ginkgo to provide a rich set of matchers for making assertions in tests.
- **mockgen**: Used to generate mock implementations of interfaces for testing.
- **httptest**: Standard library package for testing HTTP handlers and servers.
- **go-sqlmock**: Used for mocking database interactions in tests.

### Test Structure

- **Unit Tests**: Located alongside the code they test in the `/internal` directory
- **Integration Tests**: Located in the `/test/integration` directory
- **End-to-End Tests**: Located in the `/test/e2e` directory

### Writing BDD Tests

Tests follow the BDD style using Ginkgo's `Describe`, `Context`, and `It` blocks to organize test cases:

```go
var _ = Describe("Handler", func() {
    Context("when handling a valid request", func() {
        It("should return a successful response", func() {
            // Test code here
            Expect(response.StatusCode).To(Equal(http.StatusOK))
        })
    })
})
```

### Mocking Dependencies

Interfaces are mocked using `mockgen` to isolate the component being tested:

```go
//go:generate mockgen -destination=../mocks/mock_service.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/example Service
```

## Configuration

Sudal uses a flexible configuration system that supports different environments through environment variables and configuration files.

### Quick Start for Local Development

1. **Create a `.env` file** in the project root (copy from `.env.example`):
   ```
   SERVER_PORT=8080
   LOG_LEVEL=debug
   ENVIRONMENT=development
   # See .env.example for all available options
   ```

2. **Run with Docker Compose** (recommended):
   ```bash
   docker-compose up
   ```

3. **Or run directly** with the configuration file:
   ```bash
   ./bin/server --config=./configs/config.yaml
   ```

### Documentation

For detailed configuration instructions, including:
- Environment-specific configuration guides
- Required parameters and their descriptions
- Cloud Run deployment examples
- Secret management

See the [Configuration Documentation](docs/configuration/configuration.md).

## Logging

Sudal uses structured JSON logging via the `zap` library for optimal performance and observability.

### Log Format

All logs are output in JSON format with the following standard fields:

- `timestamp`: ISO8601-formatted timestamp of the log event
- `level`: Log severity level (`debug`, `info`, `warn`, `error`)
- `message`: The log message
- `caller`: File and line number where the log was called
- `trace_id`: Unique identifier for request tracing (automatically generated for each request)
- `user_id`: User identifier (when available in authenticated contexts)

For error-level logs, a `stacktrace` field is automatically included with the full call stack.

### Configuration

The log level can be configured using the `LOG_LEVEL` environment variable or in the configuration file:

```yaml
log_level: debug  # Options: debug, info, warn, error
```

The default log level is `info` if not specified.

### Usage in Code

The logging package provides context-aware logging functions:

```go
// Standard logging
log.Debug("Debug message", zap.String("key", "value"))
log.Info("Info message", zap.Int("count", 42))
log.Error("Error occurred", zap.Error(err))

// Context-aware logging (preferred)
log.InfoContext(ctx, "Processing request", zap.String("item_id", id))
log.ErrorContext(ctx, "Failed to process request", zap.Error(err))
```
