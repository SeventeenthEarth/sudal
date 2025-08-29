# Protocol Filter Middleware

## Status
- Accepted — 2025-08-29

## Context
- REST는 health/readiness 등 모니터링 전용으로만 허용, 비즈니스 기능은 gRPC/Connect(+ gRPC‑Web) 전용.
- gRPC 엔드포인트에 대한 HTTP/JSON 접근을 차단해 엔드포인트 은닉 및 보안 강화 필요.
- 프로토콜 경계 준수로 API 거버넌스 일관성 유지.

## Decision
- Protocol Filter 미들웨어를 도입해 gRPC 전용 경로에 대한 HTTP/JSON 요청을 404로 응답하여 차단한다.
- REST 접근은 `/api/*`(health), `/docs`(Swagger UI) 만 허용한다.
- gRPC 및 gRPC‑Web 헤더/지표를 기반으로 허용 판별을 수행한다.

## Implementation Details
- HTTP 체인: Protocol Filter → Request Logger → Handler(health)
- gRPC 인터셉터: Public | Protected | Selective 인증 패턴 병행
- 적용 순서: 인증 전(Pre-auth)에서 필터 적용
- 허용 패턴: `/api/ping`, `/api/healthz`, `/api/health/*`, `/docs`
- 보안: HTTP/JSON로 gRPC 경로 접근 시 404로 은닉(엔드포인트 디스커버리 방지)

## Consequences
- 엔드포인트 은닉에 따른 보안성 개선.
- 프로토콜 분리로 API 거버넌스 명확화.
- 미들웨어 체인 복잡도 소폭 증가.

## Implementation Status
- Status: COMPLETED
- Location: HTTP 미들웨어 체인에 적용
- Tests: E2E 프로토콜 경계 시나리오로 검증
- References: adr/realtime-architecture.md, adr/middleware-architecture.md

## Related Documents
- **Architecture**: ./realtime-architecture.md
- **Middleware**: ./middleware-architecture.md
- **Plan**: ../plan/iteration-plan.md
