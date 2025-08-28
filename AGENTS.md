# Repository Guidelines

## Project Structure & Module Organization
- `cmd/server/`: Application entrypoint (`main.go`).
- `internal/feature/`: Domain logic by feature; keep packages cohesive.
- `internal/infrastructure/`: Adapters (DB, DI via Wire, OpenAPI, gRPC, Redis, Firebase).
- `api/openapi.yaml` and `proto/*`: API contracts; generated code lives under `internal/infrastructure/*`.
- `db/migrations/`: SQL migrations; manage with Make targets.
- `configs/`, `scripts/`, `docs/`, `test/`: Config, automation, docs, and tests (`test/integration`, `test/e2e`).

## Build, Test, and Development Commands
- `make init`: One‑time dev setup (env, tools, scaffolding).
- `make install-tools`: Installs linters, generators.
- `make generate`: Runs all generators (buf, wire, mocks, ogen, ginkgo).
- `make build`: Builds binary to `bin/sudal` from `cmd/server`.
- `make run`: Runs via Docker Compose (DB/redis/services).
- `make test` | `make test.unit` | `make test.int`: Unit and integration tests with coverage (`coverage.*.html`).
- `make test.e2e` [`VERBOSE=1`]: Godog E2E tests; use `make test.e2e.only TAGS=@tag` for subsets.
- `make fmt` | `make vet` | `make lint`: Format, vet, and lint.
- `make migrate-create DESC=create_users_table` and `make migrate-up`: Manage DB migrations.

## Coding Style & Naming Conventions
- Go 1.x, `gofmt`/`go fmt` enforced; prefer tabs for indentation.
- Package names: short, lowercase; exported identifiers use `CamelCase`.
- Files: tests end with `_test.go`; mocks under `internal/mocks` (generated).
- Linting via `golangci-lint` (`make lint`); fix or justify warnings.

## Testing Guidelines
- Frameworks: `testing`, Ginkgo/Gomega (optional), Godog for E2E.
- Unit tests under `internal/**`; integration in `test/integration`; E2E in `test/e2e`.
- Aim for meaningful coverage on features; verify HTML coverage reports.
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
- For Firebase E2E: set `FIREBASE_WEB_API_KEY` and ensure server on `localhost:8080`.
- Prefer Make targets over direct scripts; see `scripts/*.sh --help` for details.
