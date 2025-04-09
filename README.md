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
   - `golangci-lint` for linting (install via `make init` or manually).
   - `buf` (optional but recommended for Protobuf management).
   - `protoc-gen-go`, `protoc-gen-connect-go` (for code generation from `.proto`).

### Make Commands

The following `make` commands are available for development:

- `make help`: Display available commands and descriptions.
- `make init`: Initialize the development environment (e.g., install tools).
- `make build`: Compile the main server application into the `./bin/` directory.
- `make test`: Run all unit and integration tests.
- `make lint`: Run the `golangci-lint` checks.
- `make clean`: Remove build artifacts and clean caches.
- `make proto-gen`: Generate Go code from Protobuf definitions (will be added later).
- `make run`: Build and run the application (convenience target - will be added later).

## Testing

### Running Tests

To run all tests:

```bash
make test
```

To run specific tests:

```bash
go test ./path/to/package -v
```

Coverage reports can be generated:

```bash
go test -coverprofile=coverage.out ./...
```

### Test Structure

- Unit tests are located alongside the code they test
- Integration tests are in the `/test` directory
- End-to-end tests are in the `/test/e2e` directory
