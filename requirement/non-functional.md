# 비기능 요구사항

## 성능/확장성
- 지연: 핵심 상호작용 P95 < 150ms(동일 리전), P99 < 300ms.
- 처리량: 방 동시 1k, 이벤트/초 10k 목표(초기). 수평 확장 전제.
- 캐싱: Redis를 읽기 캐시 및 Pub/Sub으로 활용. 만료/일관성 규칙 명시.

## 일관성/신뢰성
- 상태 버전(state_version)으로 낙관적 락, 충돌 시 재시도/병합 전략.
- 멱등 키 적용(이벤트/커맨드), Dead-letter 큐/리플레이.
- 장애 허용: Circuit Breaker, 지수 백오프, 세이프 디그레이드.

## 보안/프라이버시
- 인증: Firebase(OIDC), 토큰 검증·만료 관리, 최소 권한.
- 데이터: PII 최소 수집, 저장 시 암호화(가능한 경우), 로그 민감정보 마스킹.
- 비밀관리: 환경변수/Secret Manager, 저장소 커밋 금지.
- 프로토콜 경계: gRPC 전용 엔드포인트에 HTTP/JSON 접근 시 404(Protocol Filter)로 은닉.
- gRPC‑Web: 웹 클라이언트 호환 허용.

## 관측성/운영
- 로깅: 구조화 로그(zap/zerolog), Trace ID 전파.
- 메트릭: 이벤트 처리량, 지연, 실패율, 회로 상태, 재시도 횟수.
- 트레이싱: 주요 경로(입장/이벤트/브로드캐스트) 분산 추적.
- OpenAPI 운영: `/api/*`(health 계열), `/docs`(Swagger UI) 제공. 비즈니스 REST는 제공하지 않음.

### Real-time SLI & Dashboards
- Join→First Broadcast P95 (목표: < 150ms)
- Reconnect Recovery P95 (목표 설정 및 추적)
- Stream Error Rate / Retry Count

## OpenAPI & Developer Experience

### Code Generation Strategy
- 도구: ogen-go/ogen 기반 자동 서버 코드 생성
- 소스: `api/openapi.yaml` 단일 소스 오브 트루스
- 생성 경로: `internal/infrastructure/openapi/oas_*.go`
- 커스텀 구현: `internal/infrastructure/openapi/handler.go`

### Available Endpoints
- OpenAPI Endpoints: `/api/*` — 생성된 REST 엔드포인트(health 전용)
- Swagger UI: `/docs` — 상호작용형 문서
- OpenAPI Spec: `/api/openapi.yaml` — 원본 스펙 다운로드

### Health Check Integration
| Endpoint | Purpose | Implementation |
|----------|---------|----------------|
| `GET /api/ping` | 서비스 가용성 | 단순 상태 체크 |
| `GET /api/healthz` | 종합 헬스 | 의존성 점검 |
| `GET /api/health/database` | DB 연결성 | 연결/풀 상태 점검 |

### Implementation Benefits
- 코드와 문서 동기화
- REST 엔드포인트(health)의 개발 시간 단축
- Swagger UI를 통한 상호작용 테스트
- (필요 시) 클라이언트 SDK 생성 지원

## 인프라
- 배포: GCP Cloud Run(또는 동급), Cloud SQL, Memorystore.
- CI/CD: 빌드/테스트/린트/컨테이너 이미지, 환경별 배포 분리.
- 데이터 백업/복구: 정기 백업, 스키마 마이그레이션 롤백 전략.

## Database Configuration Requirements

### Connection Pooling
- MaxOpenConns: 25(프로덕션), 10(개발)
- MaxIdleConns: 5(프로덕션), 2(개발)
- ConnMaxLifetime: 5분
- ConnMaxIdleTime: 2분
- ConnectTimeout: 10초

### Monitoring
- 커넥션 풀 메트릭 수집(열린/유휴/대기/닫힘 통계)
- 쿼리 성능 모니터링(지연/에러율)
- 데드 커넥션 감지
- DB 연결 헬스체크 엔드포인트 제공

## 규정/정책
- 데이터 보존 기간, 사용자 삭제 요청에 대한 완전 삭제(정책 문서화).
- 콘텐츠/커뮤니티 가이드라인 위반 대응(신고/차단) 준비.

## Related Documents
- **Architecture**: ../adr/realtime-architecture.md
- **Implementation**: ./implementation-overview.md
- **Status Tracking**: ./implementation-status.md
- **Testing**: ./testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
