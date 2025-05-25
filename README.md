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

## API Framework

The backend supports multiple API protocols:

### Connect-go (gRPC/HTTP)
The backend uses Connect-go framework for API development, which supports both gRPC and HTTP/JSON protocols with a single implementation. APIs are defined using Protocol Buffers (`.proto` files) and served via Connect-go handlers.

### OpenAPI (REST)
The backend also provides REST API endpoints with OpenAPI 3.0 specification and Swagger UI documentation. REST APIs are generated using [ogen-go/ogen](https://github.com/ogen-go/ogen) from OpenAPI specifications, ensuring type safety and automatic code generation.

### Protocol Buffers and JSON Serialization

When using Protocol Buffers with Connect-go, be aware of the following JSON serialization behaviors:

- **Enum Values**: Protobuf enum values are serialized to JSON using their full enum name. For example, an enum value defined as `SERVING_STATUS_SERVING` in the `.proto` file will appear as the string `"SERVING_STATUS_SERVING"` in JSON responses, not as a shortened form like `"SERVING"`.
- **Field Names**: Protobuf field names are converted to camelCase in JSON (e.g., `user_id` becomes `userId`).
- **Default Values**: Fields with default values are typically omitted from JSON output unless explicitly configured otherwise.

These behaviors should be considered when writing client code that consumes the API or when writing tests that verify API responses.

## Quick Start

### Prerequisites

- Go 1.24 or higher
- Make

### Installation

1. Initialize the development environment:
   ```bash
   make init
   ```

2. Run the application using Docker Compose:
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
- `make generate`: Run all code generation tasks (mocks, test suites, proto, openapi).
- `make ogen-generate`: Generate OpenAPI server code from specification.
- `make ogen-clean`: Clean generated OpenAPI code.
- `make run`: Run the application using Docker Compose with `docker-compose up --build`.

## Testing

The project has three types of tests:

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test interactions between components
3. **End-to-End Tests**: Test the entire system with a running server

To run all tests:

```bash
make test
```

For detailed information about testing, including:
- Running specific test types
- Test structure and organization
- Testing strategy and tools
- Writing BDD tests with Ginkgo
- End-to-end testing

See the [Testing Documentation](docs/test.md).

## API Documentation

### OpenAPI and Swagger UI

The project provides interactive API documentation through Swagger UI, generated from OpenAPI 3.0 specifications:

- **Swagger UI**: Available at `/docs` when the server is running
- **OpenAPI Spec**: Available at `/api/openapi.yaml`
- **REST Endpoints**: Available under `/api/*`

#### Accessing the Documentation

1. Start the server:
   ```bash
   make run
   ```

2. Open Swagger UI in your browser:
   ```
   http://localhost:8080/docs
   ```

#### Available Endpoints

- `GET /api/ping` - Simple health check
- `GET /api/healthz` - Comprehensive health check
- `GET /api/health/database` - Database health check

#### Code Generation

The OpenAPI server code is automatically generated using ogen-go/ogen:

```bash
# Generate OpenAPI code
make ogen-generate

# Clean generated code
make ogen-clean
```

**Note**: Generated files (`internal/infrastructure/openapi/oas_*.go`) are automatically excluded from version control via `.gitignore`. Only the OpenAPI specification (`api/openapi.yaml`) and custom handler implementations are tracked in git.

For detailed information about OpenAPI implementation, code generation, and development workflow, see the [OpenAPI Documentation](docs/openapi.md).

## Configuration

Sudal uses a flexible configuration system that supports different environments through environment variables and configuration files.

### Environment Setup

The application supports three distinct environments:
- **dev**: Development environment (default)
- **canary**: Canary/staging environment
- **production**: Production environment

#### Setting Up Environment Configuration

1. **Create environment-specific files**:
   - Copy `.env.template` to create your environment-specific files:
     - `.env` - For local development (default)
     - `.env.canary` - For canary/staging environment
     - `.env.production` - For production environment

   ```bash
   # For local development
   cp .env.template .env

   # For canary environment
   cp .env.template .env.canary

   # For production environment
   cp .env.template .env.production
   ```

2. **Configure environment variables**:
   - Edit each file to set the appropriate values for that environment
   - Required variables are marked with `[REQUIRED]` in the template
   - Environment-specific required variables are marked with `[ENV:xxx]`
   - Variables with `[OPTIONAL]` have sensible defaults if not specified

3. **Set the APP_ENV variable**:
   - The `APP_ENV` variable determines which environment configuration to use
   - Default is `dev` if not specified
   - Set to `canary` or `production` to use those environments:

   ```bash
   # For local development (default)
   APP_ENV=dev

   # For canary environment
   APP_ENV=canary

   # For production environment
   APP_ENV=production
   ```

4. **Database Configuration**:
   - You can configure the database connection in two ways:
     - Set `POSTGRES_DSN` directly with a full connection string
     - Or set individual components: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
   - In production, database configuration is required

### Quick Start for Local Development

1. **Create a `.env` file** in the project root:
   ```bash
   cp .env.template .env
   ```

2. **Edit the `.env` file** with your local development settings:
   ```
   APP_ENV=dev
   SERVER_PORT=8080
   LOG_LEVEL=debug
   DB_HOST=localhost
   DB_USER=user
   DB_PASSWORD=password
   DB_NAME=quizapp_db
   ```

3. **Run with Docker Compose** (recommended):
   ```bash
   docker-compose up
   ```

4. **Or run directly** with the configuration file:
   ```bash
   ./bin/server --config=./configs/config.yaml
   ```

### Important Notes

- **Never commit** your actual `.env`, `.env.canary`, or `.env.production` files to version control
- Only the `.env.template` file should be committed
- The application will automatically load the appropriate `.env` file based on the `APP_ENV` value
- Environment variables set in the system take precedence over those in the `.env` files

### Documentation

For detailed configuration instructions, including:
- Environment-specific configuration guides
- Required parameters and their descriptions
- Cloud Run deployment examples
- Secret management

See the [Configuration Documentation](docs/configuration.md).

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
