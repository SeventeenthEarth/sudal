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

# Clean build artifacts and generated files
make clean

# Complete cleanup including Go module cache
make clean-all
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

# Run E2E tests with specific tags
make test.e2e.only TAGS=@health
make test.e2e.only TAGS=@user
make test.e2e.only TAGS=@grpc

# Run E2E tests with tags and scenario name
make test.e2e.only TAGS=@health SCENARIO="Basic health check"

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

# Lint protobuf files
make buf-lint

# Check for breaking protobuf changes
make buf-breaking
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

# Show current migration version
make migrate-version

# Reset database (drop all and reapply)
make migrate-reset

# Drop all database objects (DANGEROUS - requires confirmation)
make migrate-drop

# Fresh migration setup (DANGEROUS - removes migration files)
make migrate-fresh

# Force set migration version (recovery only)
make migrate-force VERSION=5
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
   - Supports both gRPC and HTTP/JSON protocols via Connect-go

### Clean Architecture Structure

The codebase follows Domain-Driven Design with feature-based organization:

```text
internal/feature/{feature_name}/
├── domain/        # Core business logic
│   ├── entity/    # Business entities and errors
│   └── repo/      # Repository interfaces
├── application/   # Use cases and business orchestration
├── data/          # Repository implementations
└── protocol/      # External interfaces (handlers, services)
    ├── grpc_manager.go    # gRPC service implementation
    └── rest_handler.go    # REST handlers (if applicable)
```

### Middleware Chain Architecture

The server uses a modular middleware chain pattern for clean separation of concerns:

#### Middleware Chain Types

**HTTP Middleware Chains:**

- `PublicHTTP`: Request logging only (for REST endpoints)
- `ProtectedHTTP`: Request logging + Auth (for future authenticated REST)
- `GRPCOnlyHTTP`: Protocol filter + Request logging (ensures gRPC-only access)

**gRPC Interceptor Chains:**

- `PublicGRPC`: No authentication (e.g., Health service)
- `ProtectedGRPC`: Full authentication required
- `SelectiveGRPC`: Selective authentication based on procedure

#### Service Registry & Route Registration

```go
// Clean separation of concerns
ServiceRegistry → Manages all service handlers
MiddlewareChainBuilder → Configures middleware chains
RouteRegistrar → Registers routes with appropriate chains
```

### Firebase Authentication

The project uses Firebase Admin SDK for authentication with a selective approach:

#### Authentication Architecture

```text
RegisterUser: Direct Firebase verification → Create user (no middleware)
GetUserProfile: Auth middleware → Retrieve user → Use context
UpdateUserProfile: Auth middleware → Retrieve user → Use context
```

#### Configuration

```bash
# Environment variables
GOOGLE_APPLICATION_CREDENTIALS=./secrets/firebase_admin_key.json
APP_ENV=dev

# Config file (configs/config.yaml)
firebase_project_id: sudal-14497
firebase_credentials_json: ./secrets/firebase_admin_key.json
```

#### Protected Procedures

Default protected procedures are configured in the service configuration:

- `/user.v1.UserService/GetUserProfile`
- `/user.v1.UserService/UpdateUserProfile`

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
- Uses `golang-migrate/migrate` for schema management

### Migration Best Practices

#### Migration File Naming

```text
{version}_{description}.{direction}.sql
Example: 000001_create_users_table.up.sql
```

#### Important Migration Commands

- `make migrate-reset`: Drops all objects with CASCADE, reapplies migrations
- `make migrate-drop`: Drops ALL database objects (requires confirmation)
- `make migrate-fresh`: Backs up and removes migration files (development only)

**Note**: These commands use `DROP SCHEMA CASCADE` for complete cleanup including tables, views, functions, triggers, sequences, and custom types.

### Testing Strategy

1. **Unit Tests**: Ginkgo/Gomega for BDD-style tests
2. **Integration Tests**: Test feature interactions with real DB
3. **E2E Tests**: Godog v0.14 with Gherkin syntax
4. All mocks generated to `internal/mocks/`

#### E2E Testing with Godog

E2E tests use Gherkin feature files for human-readable scenarios:

```bash
# Run E2E tests with tags
make test.e2e.only TAGS=@health
make test.e2e.only TAGS=@user
make test.e2e.only TAGS=@grpc,@positive

# Run specific scenario
make test.e2e.only TAGS=@health SCENARIO="Basic health check"
```

**Feature Files Structure:**

```text
test/e2e/
├── features/           # Gherkin feature files
│   ├── health/
│   │   ├── rest.feature
│   │   └── grpc.feature
│   └── user/
│       └── grpc.feature
└── steps/             # Go step definitions
    ├── hooks.go
    ├── health_steps.go
    └── user_steps.go
```

**Available Tags:**

- Protocol: `@rest`, `@grpc`, `@connect`
- Domain: `@health`, `@user`
- Type: `@positive`, `@negative`
- Features: `@concurrency`

### Protocol Buffer Considerations

When working with protobuf and Connect-go:

1. **Enum Serialization**: Enums serialize to their string names in JSON
   ```text
   SERVING_STATUS_SERVING → "SERVING_STATUS_SERVING"
   ```

2. **Field Names**: Use camelCase in JSON (e.g., `user_id` → `userId`)

3. **Empty Messages**: Represented as `{}` in JSON

### Protocol Filter Middleware

Critical security feature that enforces protocol separation:

- Blocks HTTP/JSON requests to gRPC endpoints (returns 404)
- Implemented in `internal/infrastructure/middleware/protocol_filter.go`
- Ensures business logic is only accessible via gRPC

### Code Generation Workflow

1. Modify .proto files in `proto/`
2. Run `make generate-buf` to generate Go code
3. Implement service interfaces in `internal/feature/{feature}/protocol/grpc_manager.go`
4. Run `make generate-wire` if dependencies changed
5. Generated files are gitignored (\*.pb.go, \*.connect.go, wire_gen.go)

### Mock Generation

Mocks are generated using mockgen and stored in `/internal/mocks/`:

```go
//go:generate mockgen -destination=../../internal/mocks/mock_health_repository.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/health/domain/repo HealthRepository
```

### Important Patterns

- Always use context-aware logging: `log.InfoContext(ctx, "message")`
- Return domain errors from `internal/feature/{feature}/domain/entity/errors.go`
- Use base repository for common DB operations
- Follow existing code style and patterns in the feature being modified
- Access authenticated user via: `middleware.GetAuthenticatedUser(ctx)`
- Use BDD-style tests with Ginkgo's Describe/Context/It blocks

### Direct Script Usage

For advanced options, scripts can be run directly:

```bash
# Setup with custom Git config
./scripts/setup-dev-env.sh --git-user "Name" --git-email "email"

# Database operations
./scripts/migrate.sh up
./scripts/migrate.sh down 3
./scripts/migrate.sh create create_users_table

# Run E2E tests with specific test
./scripts/run-e2e-tests.sh TestConnectGoHealthService

# Cleanup with specific types
./scripts/clean-all.sh proto
./scripts/clean-all.sh mocks
```
