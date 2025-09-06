# Middleware Chain Architecture

## Status
- Accepted — 2025-08-29

## HTTP Chain
1. Protocol Filter — gRPC 전용 경로 HTTP/JSON+Connect 차단(404)
2. Request Logger — 구조화 로깅
3. Handler — health/readiness 전용(`/api/*`)

## gRPC Interceptors
1. Public gRPC — 인증 없음(헬스 등)
2. Protected gRPC — 전역 인증 미들웨어 적용
3. Selective gRPC — 메서드 단위 인증(선택적 보호)

참고:
- Selective 인증의 보호 대상은 `internal/infrastructure/apispec/paths.go`의 `ProtectedProcedures()`에서 관리됩니다.
- Quiz: `SubmitQuizResult`, `GetUserQuizHistory`는 보호 대상이며, `ListQuizSets`, `GetQuizSet`은 Public입니다.

## Authentication Strategy
- RegisterUser: 핸들러 내 Firebase 토큰 직접 검증(id_token)
- Get/Update Profile: 미들웨어 인증 + 컨텍스트 주입 기반 처리

## Benefits
- 관심사 분리와 보안 강화(프로토콜 필터)
- 유연한 인증 패턴(전역/선택)
- 명확한 체인 구성으로 유지보수 용이

## Related Documents
- **Architecture**: ./realtime-architecture.md
- **Protocol Filter**: ./protocol-filter-middleware.md
- **Plan**: ../plan/iteration-plan.md
