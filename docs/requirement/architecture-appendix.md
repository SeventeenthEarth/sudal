# Architecture Appendix

본 문서는 기존 문서의 Appendix A~C 내용을 분리해 정리한 것입니다. 상세 다이어그램은 `assets/`의 PlantUML 원본을 참조하세요.

## A. 서버/DB/인프라 세부 구현

### A.1 Server (connect-go)
- Buf로 Protocol Buffers 정의 버전 관리.
- connect-go로 gRPC/gRPC‑Web 제공(REST는 OpenAPI로 health 전용).
- 양방향 스트리밍(방 상태 업데이트): Pub/Sub 이벤트 → 다중 인스턴스 동기 → 클라이언트 방송.
- Circuit Breaker: Sony/gobreaker/Hystrix-go, 실패 시 fallback, 개방/반개방/폐쇄 관리.
- Event Sourcing: 상태 변경 이벤트 스트림 저장, append-only 로그, CQRS 결합.

### A.2 Redis (Cloud Memorystore)
- Key 구조(예시):
  - `room:<roomId>:participants` — 참여자 ID Set
  - `room:<roomId>:candy_pot` — 현재 사탕 수(Int)
  - `user:<userId>:presence` — 접속 상태(온라인/오프라인, 방 ID)
- 방 생성 시 초기화, 입장/이탈 시 업데이트.

### A.3 Pub/Sub 이벤트 흐름
1) 상태 변경 처리 → Redis/DB 업데이트
2) 메시지 발행(예: `room::PARTICIPANT_JOINED`)
3) 다른 인스턴스 구독(Push/Pull)
4) 연결된 클라이언트 스트림으로 브로드캐스트

### A.4 PostgreSQL 스키마 개요
- users, quiz_sets, questions, quiz_results, comparisons, comparison_participants, candy_transactions

### A.5 CI/CD & Infra
- GitHub Actions: PR 빌드/테스트/린트, main 머지 시 이미지 빌드→레지스트리→Cloud Run 배포.
- Terraform: Cloud SQL/Memorystore/Pub/Sub/IAM/Cloud Run 등 IaC. Event Store/CB 리소스 프로비저닝.
- Monitoring: GCP Operations, Alert; CB/ES 메트릭 대시보드.

### A.6 Circuit Breaker 상세 설계
- 상태: 닫힘/열림/반열림.
- 매개변수: 실패 임계값, 열림 유지 시간, 반열림 허용 요청 수.
- 모니터링: Prometheus 메트릭.

### A.7 Event Sourcing 상세 설계
- 이벤트 구조(JSON 예시), 이벤트 저장소(PostgreSQL 등), 스냅샷(예: 100개 후), CQRS 연계(커맨드/쿼리 분리, 프로젝션).

세부 항목은 원문 서술을 유지하며 이 문서의 범위를 벗어나는 다이어그램은 아래 B, C에서 참조합니다.

## B. 구성 요소 간 관계 다이어그램
- assets/system-architecture.puml

## B-1. Circuit Breaker 상세 다이어그램
- assets/circuit-breaker.puml

## B-2. Event Sourcing 상세 다이어그램
- assets/event-sourcing.puml

## C. 예시 시퀀스 다이어그램
- assets/sequence-room-comparison.puml

## Related Documents
- **Architecture**: ../adr/realtime-architecture.md
- **Implementation**: ./implementation-overview.md
- **Status Tracking**: ./implementation-status.md
- **Testing**: ./testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
