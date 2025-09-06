# Protocol Filter Middleware

## Status
- Accepted — 2025-08-29

## Context
- REST는 health/readiness 등 모니터링 전용으로만 허용, 비즈니스 기능은 gRPC(+ gRPC‑Web) 전용.
- Connect 프로토콜(JSON/streaming)은 gRPC‑only 경로에서 차단 대상.
- gRPC 엔드포인트에 대한 HTTP/JSON 접근을 차단해 엔드포인트 은닉 및 보안 강화 필요.
- 프로토콜 경계 준수로 API 거버넌스 일관성 유지.

## Decision
- Protocol Filter 미들웨어를 도입해 gRPC 전용 경로에 대한 HTTP/JSON 및 Connect 요청을 404로 응답하여 차단한다.
- REST 접근은 `/api/*`(health), `/docs`(Swagger UI) 만 허용한다.
- gRPC 및 gRPC‑Web 헤더/지표를 기반으로 허용 판별을 수행한다.

## Implementation Details
- HTTP 체인: Protocol Filter → Request Logger → Handler(health)
- gRPC 인터셉터: Public | Protected | Selective 인증 패턴 병행
- 적용 순서: 인증 전(Pre-auth)에서 필터 적용
- 허용 패턴: `/api/ping`, `/api/healthz`, `/api/health/*`, `/docs`
- 보안: HTTP/JSON 및 Connect로 gRPC 경로 접근 시 404로 은닉(엔드포인트 디스커버리 방지)
- gRPC‑only 경로: Health/User/Quiz 서비스 경로(`GetGRPCOnlyPaths`) 포함

## Consequences
- 엔드포인트 은닉에 따른 보안성 개선.
- 프로토콜 분리로 API 거버넌스 명확화.
- 미들웨어 체인 복잡도 소폭 증가.

## Implementation Status
- Status: COMPLETED
- Location: HTTP 미들웨어 체인에 적용
- Tests: E2E 프로토콜 경계 시나리오로 검증
- References: adr/realtime-architecture.md, adr/middleware-architecture.md

## Continuous Verification (E2E)
- 모든 PR에서 `make test.e2e` 전체 E2E가 실행되며 다음을 포함합니다:
  - REST→gRPC 경로에 대한 HTTP/JSON 요청은 404로 차단(@rest @negative)
  - Connect(JSON/streaming) 요청은 gRPC‑only 경로에서 404로 차단(@protocol_filter)
  - gRPC(+ gRPC‑Web) 스모크(@grpc)
  - 실패 시 병합 금지. requirement/testing-strategy.md의 실행 정책과 연동됩니다.

## Related Documents
- **Architecture**: ./realtime-architecture.md
- **Middleware**: ./middleware-architecture.md
- **Plan**: ../plan/iteration-plan.md
