# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

### Development & Build

```bash
# Initialize development environment
make init

# Install all development tools
make install-tools

# Build the application
make build

# Run the application with Docker Compose
make run

# Run server directly (after building)
./bin/sudal --config=./configs/config.yaml
```

### Code Quality

```bash
# Format code
make fmt

# Run linter (golangci-lint)
make lint

# Run go vet
make vet
```

### Testing

```bash
# Run all tests (unit and integration with preparation)
make test

# Run only unit tests
make test.unit

# Run only integration tests
make test.int

# Run E2E tests (requires running server)
make test.e2e

# Run specific test file
go test -v ./path/to/package -run TestName

# Run Ginkgo tests with specific focus
ginkgo -v -focus="specific test description" ./path/to/package
```

### Code Generation

```bash
# Generate all code (proto, mocks, wire, openapi, test suites)
make generate

# Generate protobuf code
make generate-buf

# Generate dependency injection code (Wire)
make generate-wire

# Generate mocks
make generate-mocks

# Generate OpenAPI server code
make generate-ogen

# Bootstrap Ginkgo test suites
make generate-ginkgo
```

### Database Migrations

```bash
# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down

# Create new migration
make migrate-create DESC=create_users_table

# Check migration status
make migrate-status

# Reset database (drop all and reapply)
make migrate-reset
```

## High-Level Architecture

### Dual-Protocol Architecture

This codebase implements a strict separation of concerns between protocols:

1. **REST API (Health & Monitoring Only)**
   - Endpoints: `/api/ping`, `/api/healthz`, `/api/health/database`, `/docs`
   - Purpose: Load balancer health checks, monitoring, documentation
   - Implementation: OpenAPI/ogen-generated code in `internal/infrastructure/openapi/`

2. **gRPC (Business Logic Only)**
   - All business functionality (users, quizzes, etc.)
   - Protocol enforcement via middleware that blocks HTTP/JSON to gRPC endpoints
   - Implementation: Connect-go framework with Protocol Buffers

### Clean Architecture Structure

The codebase follows Domain-Driven Design with feature-based organization:

```
internal/feature/{feature_name}/
├── domain/        # Core business logic
│   ├── entity/    # Business entities and errors
│   └── repo/      # Repository interfaces
├── application/   # Use cases and business orchestration
├── data/          # Repository implementations
└── interface/     # External interfaces (handlers, services)
    └── connect/   # gRPC service implementations
```

### Dependency Injection

- Uses Google Wire for compile-time dependency injection
- Wire configuration in `internal/infrastructure/di/`
- Run `make generate-wire` after modifying wire.go

### Configuration Management

- Environment-based configuration (dev, canary, production)
- Uses Viper for config loading from YAML and environment variables
- Config precedence: ENV vars > .env file > config.yaml
- Never commit actual .env files

### Database & Caching

- PostgreSQL with connection pooling (configurable)
- Redis for caching
- Database migrations in `db/migrations/`
- Repository pattern with base repository for common operations

### Testing Strategy

1. **Unit Tests**: Ginkgo/Gomega for BDD-style tests
2. **Integration Tests**: Test feature interactions with real DB
3. **E2E Tests**: Go + testify BDD style, requires running server
4. All mocks generated to `internal/mocks/`

### Protocol Filter Middleware

Critical security feature that enforces protocol separation:

- Blocks HTTP/JSON requests to gRPC endpoints (returns 404)
- Implemented in `internal/infrastructure/middleware/protocol_filter.go`
- Ensures business logic is only accessible via gRPC

### Code Generation Workflow

1. Modify .proto files in `proto/`
2. Run `make generate-buf` to generate Go code
3. Implement service interfaces in `internal/feature/{feature}/interface/connect/`
4. Run `make generate-wire` if dependencies changed
5. Generated files are gitignored (*.pb.go,*.connect.go, wire_gen.go)

### Important Patterns

- Always use context-aware logging: `log.InfoContext(ctx, "message")`
- Return domain errors from `internal/feature/{feature}/domain/entity/errors.go`
- Use base repository for common DB operations
- Follow existing code style and patterns in the feature being modified
