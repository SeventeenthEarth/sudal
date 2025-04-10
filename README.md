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

4. Run the application:
   ```bash
   ./bin/server
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
- `make fmt`: Format Go code using `go fmt`.
- `make vet`: Run static analysis with `go vet`.
- `make lint`: Run the `golangci-lint` checks.
- `make test`: Run all unit and integration tests using Ginkgo and generate coverage report (automatically runs fmt, vet, lint, ginkgo-bootstrap, and generate-mocks first).
- `make clean`: Remove build artifacts and caches.
- `make generate`: Run all code generation tasks (mocks, test suites, proto).
- `make generate-mocks`: Generate mock implementations using mockgen.
- `make ginkgo-bootstrap`: Bootstrap Ginkgo test suites in all packages with tests.
- `make proto-gen`: Generate Go code from Protobuf definitions (will be added later).
- `make run`: Build and run the application (convenience target - will be added later).

## Testing

### Running Tests

To run all tests and generate a coverage report:

```bash
make test
```

This will:
1. Format the code with `go fmt`
2. Run static analysis with `go vet`
3. Run linter checks with `golangci-lint`
4. Run all code generation tasks via `make generate`:
   - Generate Ginkgo test suites
   - Generate mock implementations
   - Generate code from Protocol Buffers (when implemented)
5. Run all tests with Ginkgo
6. Generate a coverage report (both console summary and HTML report)

After running tests, you can view the detailed coverage report by opening `coverage.html` in your browser.

To run specific tests:

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

- Unit tests are located alongside the code they test
- Integration tests are in the `/test` directory
- End-to-end tests are in the `/test/e2e` directory

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
