# 구현 로드맵

## 목표
- 기술: 실시간 동기화 엔진을 단계적으로 고도화하여 재사용 가능한 Event 기반 Core로 발전.
- 비즈니스: 소셜 퀴즈 앱을 빠르게 출시하고, 검증 후 다른 실시간 협업 도메인으로 확장.

## 단계별 일정
- Phase 1 (2–3주): Option 1 — 단순 이벤트 기반으로 MVP 가동 (Space 기본/이벤트 상태관리/Redis PubSub/기본 테스트).
- Phase 2 (1–2주): Option 2 — Core 추출 및 비즈니스 로직 분리, 인터페이스 정의, 테스트 개선.
- Phase 3 (2–3주): Option 3 — StateMachine 추상화, Plugin 시스템, SyncEngine, 샘플 플러그인.
- Phase 4 (1–2주): 검증 및 최적화 — 성능·부하 테스트, 문서화, 추가 플러그인 예제.

## 작업 묶음 (연계 PR 라인업 요약)
- 기본 환경: 초기 구조/문서, Health 체크, Docker/Compose, 린팅·테스트, 설정/로깅.
- API 기반: connect-go 통합, buf 기반 Proto 빌드 파이프라인.
- 저장소: PostgreSQL 연동/마이그레이션, Redis 연동/캐싱.
- 도메인: 사용자/퀴즈/방/비교/태그/사탕/사진의 Repository 패턴 및 gRPC 구현.
- 품질/운영: 부하 테스트, CI/CD, GCP 배포, 모니터링·알림.

## 산출물 및 수용 기준
- Phase 1
  - 산출물: Space 테이블/참여자 관리, 이벤트 핸들러, Redis PubSub, 기본 엔드투엔드 흐름.
  - 기준: 다중 사용자의 퀴즈 진행이 실시간 동기화되고, 레이스 없이 주요 상태 전이가 일관적으로 반영된다.
- Phase 2
  - 산출물: Core 인터페이스(예: Space, State, Event)와 앱 로직 분리, 테스트 커버리지 향상.
  - 기준: 비즈니스 로직 교체/추가 시 Core 수정 없이 플러그블하게 확장 가능.
- Phase 3
  - 산출물: StateMachine/Plugin/SyncEngine, 최소 1개(퀴즈) 플러그인.
  - 기준: 플러그인만으로 신규 앱(예: Video/Auction) PoC 가능.
- Phase 4
  - 산출물: 성능·부하 지표, 병목 제거, 문서.
  - 기준: 목표 TPS·지연(기본 상호작용 < 150ms P95) 충족, 장애 시 빠른 복구.

## 리스크 및 완화
- 초기 과도한 추상화: Phase 1에서 단순화 후 점진적 리팩토링.
- 일관성/경쟁상태: 낙관적 락(state_version), 멱등 이벤트 처리, 재시도·Dead-letter 큐.
- 운영 복잡도: 관측성(로깅/트레이싱/메트릭) 표준화, Circuit Breaker·백오프 적용.

## Related Documents
- **Architecture**: ../adr/realtime-architecture.md
- **Implementation**: ../requirement/implementation-overview.md
- **Status Tracking**: ../requirement/implementation-status.md
- **Testing**: ../requirement/testing-strategy.md
- **Plan**: ./iteration-plan.md
