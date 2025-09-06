# Testing Strategy

## Goals
- 신뢰도 높은 변경 검증(단위→통합→E2E 순), 프로토콜 경계 준수(REST=health, gRPC=비즈니스), 회귀 방지.

## Layers
- Unit: Ginkgo+Gomega, 인터페이스 Mocking(mockgen) 활용.
- Integration: 실제 DB/Redis 통합 경로 점검, 커넥션 풀/마이그레이션 유효성.
- E2E: Godog v0.14 + Gherkin, 실제 서버에 REST(health)/gRPC(+ gRPC‑Web) 호출.
  - 참고: Connect 프로토콜(JSON/streaming)은 gRPC‑only 경로에서 Protocol Filter에 의해 차단됩니다.

## E2E Conventions
- Tags: @rest/@grpc, @health/@user/@quiz, @protocol_filter, @positive/@negative, @concurrency 등.
- Structure: test/e2e/{features,steps}/…, Background/Scenario/Scenario Outline 사용.
- Auth: Firebase Web API Key로 토큰 획득 → gRPC 호출; RegisterUser는 토큰 제출, 조회/갱신은 미들웨어 인증.

## Commands
- Run all: `make test.e2e`
- Filter by tags: `make test.e2e.only TAGS="@quiz"` (여러 개 가능)
- Heavy concurrency only: `make test.e2e.concurrency`
- Run EXCEPT set only: `make test.e2e.except` (EXCEPT_TAGS는 스크립트에서 관리)
- Server availability: `make run` 후 실행 권장.

## Execution
- CI/CD: `make test.e2e` — 모든 E2E 시나리오 실행(외부 제약 회피용 `@firebase_rate_limit`는 기본 제외)
- EXCEPT 세트만: `make test.e2e.except`
- 로컬(부분 실행):
  - 세분화: `make test.e2e.only TAGS=@health` 또는 `SCENARIO="..."`
  - 인증 관련 예: `make test.e2e.only TAGS=@user_auth`

## Data & Cleanup
- Disposable accounts(랜덤 이메일), 시나리오 후 정리 훅으로 사용자 삭제.
- Redis 키 접두사로 테스트 격리, 패턴 삭제로 청소.

## Observability
- 구조화 로그와 메트릭으로 지연/에러/회로 상태 추적.
- 프로토콜 경계 부정 시나리오(@protocol_filter)로 Connect(JSON/streaming) 404 차단을 상시 검증.

## Related Documents
- **Architecture**: ../adr/realtime-architecture.md
- **Implementation**: ./implementation-overview.md
- **Status Tracking**: ./implementation-status.md
- **Testing**: ./testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
