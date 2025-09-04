# Repository Guidelines

## Project Structure & Module Organization
- `cmd/server/`: Application entrypoint (`main.go`); verifies DB/Redis connectivity at startup.
- `cmd/cleanup-users/`: Utility to clean up Firebase E2E test users (used by `make test.e2e.cleanup-users`).
- `internal/feature/`: Domain logic by feature (domain, data, application, protocol per feature).
- `internal/infrastructure/`: Platform adapters and app wiring:
  - `server/`: HTTP server, route registry (REST via OpenAPI + gRPC via Connect-go), middleware chains.
  - `di/`: Dependency injection with Wire (constructs services/handlers/managers).
  - `middleware/`: HTTP/Connect middleware (auth, logging, protocol filters).
  - `repository/`: Infrastructure-backed repositories (e.g., Postgres) that depend on service-level SQL executor.
  - `openapi/`: ogen-generated REST server handlers and Swagger UI helper.
  - `apispec/`: Shared API path/authorization specs for middleware.
- `internal/service/`: Cross-cutting service modules (framework-agnostic):
  - `config/`, `logger/`: Configuration loading (.env/YAML) and zap-based logging.
  - `postgres/`, `sql/postgres/`: Postgres manager and minimal SQL executor/transactor surfaces.
  - `redis/`, `cache/`: Redis manager, KV adapter, and cache utilities.
  - `firebaseauth/`, `authutil/`: Token verification and auth helpers.
  - `health/`: Startup and diagnostics utilities for external dependencies.
- `api/openapi.yaml`: REST API spec; server code generated to `internal/infrastructure/openapi` via ogen.
- `proto/*`: gRPC/Connect-go contracts; generated code lives under `gen/*` (see below).
- `gen/go`: Generated protobuf, gRPC, and Connect-go code (`*.pb.go`, `*_grpc.pb.go`, `*connect.go`).
- `gen/openapi`: OpenAPI v2 spec generated from proto (via grpc-gateway `openapiv2`).
- `db/migrations/`: SQL migrations; manage with Make targets.
- `configs/`, `scripts/`, `docs/`, `test/`: Config, automation, docs, and tests (`test/integration`, `test/e2e`).

## Build, Test, and Development Commands
- `make init`: One‑time dev setup (env, tools, scaffolding).
- `make install-tools`: Installs linters, generators.
- `make generate`: Runs all generators (ginkgo, buf, wire, mocks, ogen).
- `make build`: Builds binary to `bin/sudal` from `cmd/server`.
- `make run`: Runs via Docker Compose (server + DB + Redis).
- `make test` | `make test.unit` | `make test.int`: Unit and integration tests with coverage (`coverage.*.html`).
  - Note: Integration coverage targets feature packages (`internal/feature/...`) and intentionally excludes infrastructure packages.
- `make test.e2e` [`VERBOSE=1`]: Godog E2E tests; use `make test.e2e.only TAGS=@tag SCENARIO="..."` for subsets.
- `make test.e2e.auth`: Firebase auth E2E (requires server on `localhost:8080` and `FIREBASE_WEB_API_KEY`).
- `make test.e2e.cleanup-users`: Cleanup Firebase E2E users via `cmd/cleanup-users`.
- `make fmt` | `make vet` | `make lint`: Format, vet, and lint.
- `make migrate-create DESC=create_users_table` and `make migrate-up`: Manage DB migrations.
- Proto/Buf: `make buf-setup`, `make buf-lint`, `make buf-breaking`, `make generate-buf` (outputs to `gen/go`, `gen/openapi`).
- OpenAPI (ogen): `make generate-ogen` (outputs to `internal/infrastructure/openapi`).
- Docs subtree: `make push-docs`, `make pull-docs`.

## Coding Style & Naming Conventions
- Go 1.x, `gofmt`/`go fmt` enforced; prefer tabs for indentation.
- Package names: short, lowercase; exported identifiers use `CamelCase`.
- Files: tests end with `_test.go`; mocks under `internal/mocks` (generated).
- Linting via `golangci-lint` (`make lint`); fix or justify warnings.

## Testing Guidelines
- Frameworks: `testing`, Ginkgo/Gomega (optional), Godog for E2E.
- Unit tests under `internal/**`; integration in `test/integration`; E2E in `test/e2e`.
- Integration coverage focuses on feature layer; infra coverage is intentionally excluded in integration runs.
- Verify HTML coverage reports (`coverage.unit.html`, `coverage.int.html`).
- Example: `go test ./internal/feature/user -v` for a focused run.

## Commit & Pull Request Guidelines
- Messages: imperative, concise (e.g., “Add User Service”, “Fix e2e tests”).
- Reference issues (`#123`) and scope when helpful.
- PRs: describe change, risks, rollout, and testing notes; link issues; include sample requests/responses for API changes (OpenAPI/Proto updates).

## Communication Style
- Default response language: Korean. Internal reasoning can be in English, but user-facing sentences should be in Korean.
- Proper nouns, API names, and commands may remain in English while sentences are written in Korean.
- Documentation rule: Any content written in AGENTS.md must be in English (this document is the contributor guide).

## Security & Configuration Tips
- Copy `.env.template` to `.env`; never commit secrets (`secrets/`, keys).
- Config loading: `.env` and optional YAML via `--config configs/config.yaml`; environment-specific `.env.<env>` supported.
- For Firebase E2E: set `FIREBASE_WEB_API_KEY` and ensure server on `localhost:8080`.
- Prefer Make targets over direct scripts; see `scripts/*.sh --help` for details.
