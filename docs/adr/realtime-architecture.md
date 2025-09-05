# ADR: 실시간 동기화 아키텍처(Event + Core)

## 상태
- Accepted — 2025-08-29

## 문맥
- 소셜 퀴즈 플랫폼은 다수 사용자의 상태를 실시간 동기화해야 함.
- 장기적으로 퀴즈 외의 실시간 협업/엔터테인먼트 앱으로 확장 예정.
- 옵션 비교(Option 1/2/3)와 로드맵, 서비스/인프라/PR 계획을 종합.

## 프로토콜 경계(HTTP vs gRPC)
- 원칙: REST는 health/readiness 등 모니터링 전용, 비즈니스 기능은 gRPC/Connect(+ gRPC‑Web)만 사용.
- 이유: 일관된 스키마·성능·양방향 스트리밍·보안(엔드포인트 은닉) 확보.
- 구현: HTTP/JSON로 gRPC 경로 접근 차단을 위한 Protocol Filter 미들웨어 적용(REST health만 허용).

## 결정
- 목표 아키텍처로 Option 2(Event + Core) 채택.
  - Core: Space/State/Event/SideEffect 등 공통 추상화와 저장·동기화 책임.
  - 앱 로직: 도메인별 비즈니스 전이는 Core 인터페이스 뒤로 분리.
- 실행 전략: Phase 1에서는 Option 1로 빠르게 가동하고, Phase 2에서 Core를 추출하여 마이그레이션.
- 향후: 필요 시 Option 3(StateMachine + Plugin + Core)로 확장.

## 미들웨어 체인(요약)
- HTTP 체인: ProtocolFilter → RequestLogger → Handler
- gRPC 인터셉터: PublicGRPC | ProtectedGRPC | SelectiveGRPC(메서드 단위 인증)
- 선택적 인증: RegisterUser는 핸들러 내 직접 토큰 검증, 조회/갱신류는 미들웨어 인증 사용.

## 근거
- 점진적 전달: MVP 속도(Option 1)와 아키텍처 일관성(Option 2) 균형.
- 재사용성/확장성: 여러 앱 도메인(퀴즈/비디오/경매 등)에 공통 Core 재사용.
- 테스트 용이성: 비즈니스 로직과 동기화 인프라의 관심사 분리.

## 결과
- 장점: 플러그블 도메인 로직, 명확한 책임 경계, 성능·확장성 지향 설계.
- 단점/트레이드오프: 초기에는 이중 전환(Option 1→2) 비용; 추상화로 인한 소폭 오버헤드.
- 운영: Pub/Sub(Redis), 멱등 처리, 낙관적 락, DLQ/리플레이, 관측성 표준화가 필수.
  - 프로토콜 경계 검증은 E2E 전체(`make test.e2e`)로 상시 실행
  - 핵심 SLI: Join→First Broadcast P95, Reconnect Recovery P95, Stream Error/Retry

## 대안 검토
- Option 1(단순 이벤트): 구현 용이·빠른 가동이나 재사용성/유지보수 한계.
- Option 3(완전 플러그인): 장기적으로 이상적이나 초기 복잡도/디버깅 비용이 큼.

## 마이그레이션 계획
1) Phase 1: 단순 이벤트 기반으로 Space/이벤트 처리 파이프라인 가동.
2) Phase 2: Core 인터페이스 도출(Space/State/Event), 앱 로직 분리, 테스트 강화.
3) Phase 3: StateMachine/Plugin 도입, SyncEngine 탑재, 샘플 플러그인 운영.

## Related Documents
- **Architecture**: ./realtime-architecture.md
- **Implementation**: ../requirement/implementation-overview.md
- **Status Tracking**: ../requirement/implementation-status.md
- **Testing**: ../requirement/testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
