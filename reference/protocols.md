# API Protocols & Boundaries

This repository intentionally splits responsibilities:

- HTTP (OpenAPI) is used only for health and readiness probes.
- gRPC (Connect) carries all business functionality (users, quizzes, etc.).

## HTTP (OpenAPI) — Health Only

- Spec: `api/openapi.yaml` (server code via `make generate-ogen`).
- Endpoints:
  - `GET /api/ping` — liveness
  - `GET /api/healthz` — readiness
  - `GET /api/health/database` — DB connectivity
- Quick checks:
  - `curl -s http://localhost:8080/api/ping`
  - `curl -s http://localhost:8080/api/healthz`

## gRPC (Connect) — Business Logic

- Protos live in `proto/*` (e.g., `proto/user/v1/user.proto`, `proto/health/v1/health.proto`).
- Example grpcurl (local, plaintext):
  - List services: `grpcurl -plaintext localhost:8080 list`
  - Health: `grpcurl -plaintext -d '{}' localhost:8080 health.v1.HealthService/Check`
  - User profile: `grpcurl -plaintext -d '{"user_id":"<UUID>"}' localhost:8080 user.v1.UserService/GetUserProfile`
- Connect-go note: handlers can also accept HTTP/JSON to the Connect endpoints; prefer gRPC for production clients.

See also: [Middleware Documentation](middleware.md) for gRPC-only enforcement at the HTTP layer.

## Development & Generation

- Run stack: `make run` (Docker Compose; exposes `:8080`).
- Generate code: `make generate` (buf, wire, mocks, ogen, ginkgo).
- Lint/format: `make fmt && make vet && make lint`.

## Contribution Rules

- New features default to gRPC. Do not add HTTP routes unless there is a clear browser/3rd‑party need.
- If HTTP is necessary, define in `api/openapi.yaml`, regenerate with `make generate-ogen`, and add E2E coverage.
- Keep business logic in `internal/feature/*`; HTTP/gRPC should be thin adapters calling the same service layer.

## Testing

- Unit/integration: `make test` → coverage reports in `coverage.*.html`.
- E2E: `make test.e2e` (HTTP health scenarios; business scenarios via gRPC).
