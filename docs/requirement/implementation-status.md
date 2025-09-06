# Implementation Status Tracking

## Core Infrastructure ✅
- [x] Clean Architecture (Domain/UseCase/Repository) — COMPLETED
- [x] PostgreSQL Integration — COMPLETED
- [x] Firebase Authentication — COMPLETED
- [x] Protocol Filter Middleware — COMPLETED
- [x] Testing Framework (Ginkgo/Godog) — COMPLETED

## Database & Persistence 🔄
- [x] Connection Pooling — COMPLETED
- [x] Migration System — COMPLETED
- [ ] Redis Cache System — PARTIAL (utility available; active usage TBD)
- [ ] Event Store Schema — PLANNED

## Real‑time Synchronization ❌
- [ ] Space Concept Implementation — DESIGNED ONLY (see adr/space-schema-and-interfaces.md)
- [ ] Event Sourcing — PLANNED (see PR-88~90)
- [ ] Circuit Breaker — PLANNED (see PR-86~87)
- [ ] Pub/Sub Integration — PLANNED (see PR-52~58)

## Business Services ❌
- [x] User Service — Register/Profile/Update COMPLETED
- [ ] Quiz Service — Scaffolding(Connect‑go 핸들러; 미구현은 CodeUnimplemented), Selective Auth 범위 지정, E2E(@quiz) 추가
- [ ] Room Service — PLANNED
- [ ] Comparison Service — PLANNED
- [ ] Candy (Virtual Currency) Service — PLANNED

## Implementation Phases
- Phase 1 (MVP): Core infrastructure ready; domain services pending
- Phase 2 (Core Architecture): Not started — depends on Phase 1 completion
- Phase 3 (Plugin System): Not started — future evolution

## Additional Implemented Features
- OpenAPI/Swagger UI for health endpoints
- E2E testing with Godog (Gherkin) and tag conventions
- Redis cache utility (CRUD/TTL, ErrCacheMiss, pattern cleanup)
- Middleware chain with Protocol Filter, Request Logger, Selective Auth
 - QuizService 스캐폴딩(DI/Wire/ServiceRegistry 연결 포함)
 - `apispec.ProtectedProcedures()`에 Quiz 보호 RPC 반영(SubmitQuizResult, GetUserQuizHistory)
 - gRPC‑only 필터 경로에 Quiz 서비스 포함, Connect 차단(@protocol_filter) E2E로 검증

## Notes
- Status reflects current implementation intentions and available designs documented in this repo.
- PR roadmap status should remain synchronized with this file.

## Reference Documentation Status
| Reference File | Status | Integrated Into |
|----------------|--------|-----------------|
| protocols.md | INTEGRATED | requirement/functional-overview.md |
| firebase-authentication.md | INTEGRATED | requirement/grpc-spec.md, adr/realtime-architecture.md |
| middleware.md, middleware-chain-refactoring.md | INTEGRATED | adr/realtime-architecture.md, adr/middleware-architecture.md |
| database-connection-pooling.md | INTEGRATED | requirement/non-functional.md |
| test.md, e2e-testing-guide.md | INTEGRATED | requirement/testing-strategy.md |
| cache_utility.md | PARTIAL | requirement/non-functional.md (요구), implementation-overview.md(개요) |
| configuration.md | INTEGRATED | requirement/implementation-overview.md (3.7, 3.10) |
| database-migrations.md | INTEGRATED | requirement/implementation-overview.md (3.8) |
| openapi.md | INTEGRATED | requirement/non-functional.md (OpenAPI & DX) |
| script_guide.md | INTEGRATED | requirement/implementation-overview.md (3.11) |
| e2e-authentication-tests.md | PARTIAL | requirement/testing-strategy.md (정책/흐름) |

## Related Documents
- **Architecture**: ../adr/realtime-architecture.md
- **Implementation**: ./implementation-overview.md
- **Testing**: ./testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
