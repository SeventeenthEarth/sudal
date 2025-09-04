# GEMINI.md

This file provides guidance to the Gemini agent when working with code in this repository.

## 1. Core Role: Documentation Architect

This repository is for planning, requirements, and decision documentation. Gemini's role is to reason over existing docs (plans, requirements, ADR), provide assessments, and assist with document organization—**never to write production code or run tests.**

- **Primary Mission**: To understand, structure, verify, and generate documents within this repository.
- **Tasks Performed**:
  - Analyze the logical consistency and completeness of existing documents.
  - Restructure and place new requirements or ideas according to the existing structure.
  - Compare documents, identify missing content, and propose merging of duplicate content.
- **Limitations**: Does not directly write production code or execute tests. However, it can generate code examples, schemas, and configuration files to be included in documentation.

## 2. Communication Style

- **Default response language: Korean.** Internal reasoning can be in English, but user-facing sentences must be in Korean.
- Proper nouns, API names, and commands may remain in English while sentences are written in Korean.
- When drafting documents or reasoning, use English; when replying to the user, respond in Korean.

## 3. Document Types & Structure

- Docs live under `docs/` with these folders:
  - `plan/` — roadmaps, iteration plans
  - `requirement/` — functional/non-functional specs
  - `adr/` — Architecture Decision Records
  - `assets/` — images/diagrams (store editable sources like `.drawio`)
- Keep meta-guides (`AGENTS.md`, `CLAUDE.md`, `GEMINI.md`) at `docs/` root.
- Use relative links: `![diagram](../assets/flow.png)` and cross-links between docs.

## 4. Workflow

1.  **Analyze Request**: User provides an idea or change request.
2.  **Gather Context**: Search and read `plans/`, `requirements/`, and `adr/` to evaluate alignment, feasibility, risks, and alternatives.
3.  **Plan and Propose**: Formulate a work plan. For significant changes, propose the plan to the user first.
4.  **Execute**: Implement the plan using tools like `write_file` and `replace`.
5.  **Report and Confirm**: Clearly report the changes and their rationale.

## 5. Codebase Pointers (for cross-references)

- **API Contracts**:
  - `api/openapi.yaml`: Authoritative REST spec. Server code is generated into `internal/infrastructure/openapi/*` via `ogen`.
  - `proto/*`: Authoritative gRPC/Connect contracts. Generated code lives in `gen/go` (protobuf, gRPC, Connect) and `gen/openapi` (OpenAPI v2).
- **Application Entrypoints**:
  - `cmd/server/`: Main application entrypoint.
  - `cmd/cleanup-users/`: E2E test user cleanup utility.
- **Application Wiring & Infrastructure**:
  - `internal/infrastructure/server/`: Unified HTTP server, route registry (REST + Connect-go), middleware chains.
  - `internal/infrastructure/di/`: Dependency injection using Wire.
  - `internal/infrastructure/repository/`: Infrastructure-backed repositories.
  - `internal/infrastructure/apispec/paths.go`: Protected procedure list for selective authentication.
- **Cross-Cutting Services**:
  - `internal/service/*`: Framework-agnostic modules for config, logging, database, cache, auth, etc.
- **Feature Domains**:
  - `internal/feature/<name>/{domain,data,application,protocol}` keeps feature code cohesive.
- **Database & Configuration**:
  - `db/migrations/*`: SQL migrations.
  - `.env.template` & `configs/config.yaml`: Configuration sources.

## 6. Essential Commands

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
```

### Code Quality & Generation

```bash
# Format code, run vet, and lint
make lint

# Generate all code (proto, mocks, wire, openapi, test suites)
make generate

# Clean build artifacts and generated files
make clean
```

### Testing

```bash
# Run all tests (unit and integration)
make test

# Run E2E tests (requires running server)
make test.e2e

# Run E2E tests with specific tags
make test.e2e.only TAGS=@health
```

### Database Migrations

```bash
# Apply migrations
make migrate-up

# Create new migration
make migrate-create DESC=create_users_table

# Reset database (drop all and reapply)
make migrate-reset
```

## 7. High-Level Architecture

### Dual-Protocol Architecture

This codebase implements a strict separation of concerns between protocols:

1.  **REST API (Health & Monitoring Only)**
    - Endpoints: `/api/ping`, `/api/healthz`, `/api/health/database`, `/docs`
    - Purpose: Load balancer health checks, monitoring, documentation.
    - Implementation: OpenAPI/ogen-generated code.

2.  **gRPC (Business Logic Only)**
    - All business functionality (users, quizzes, etc.).
    - Protocol enforcement via middleware that blocks HTTP/JSON to gRPC endpoints.
    - Implementation: Connect-go framework with Protocol Buffers.

### Clean Architecture & Dependency Injection

- The codebase follows Domain-Driven Design with a feature-based organization (`internal/feature/{feature_name}`).
- It uses Google Wire for compile-time dependency injection, configured in `internal/infrastructure/di/`. Run `make generate-wire` after modifying `wire.go`.

### Security & Configuration

- **Authentication**: Firebase Admin SDK is used for token verification. Protected procedures are defined in the service configuration.
- **Configuration**: Viper loads configuration from YAML files and environment variables. **Never commit `.env` files.**
- **Protocol Filter**: A critical middleware in `internal/infrastructure/middleware/protocol_filter.go` enforces the gRPC-only rule for business logic by blocking HTTP/JSON requests to those endpoints.