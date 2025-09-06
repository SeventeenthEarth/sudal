# Implementation Overview

## 3.1 Stack Summary
- Server: Go + connect-go (gRPC/gRPC‑Web; REST는 OpenAPI로 health 전용)
- DB: PostgreSQL (Cloud SQL)
- Cache/RT: Redis (Cloud Memorystore)
- Storage: Firebase Storage or GCS
- Client: Flutter (Dart), BLoC 상태관리
- Auth: Firebase Authentication (OIDC 연동)
- Infra: GCP (Cloud Run, Pub/Sub 등)
- 안정성: Circuit Breaker
- 이벤트: Event Sourcing (+ CQRS)

## 3.2 Server
- 언어/프레임워크: Go + connect-go, 비동기 고성능, 마이크로서비스 확장 여지.
- DB 구성: PostgreSQL(영구), Redis(세션/실시간), Event Store(ES).
- API 설계(프로토콜 경계): REST는 health/readiness 전용(`/api/*`, `/docs`), 비즈니스 기능은 gRPC(+ gRPC‑Web)만 허용.
- OpenAPI(ogen): `/api/openapi.yaml` → 코드 생성(`/api/*`, `/docs`), 비즈니스 REST는 허용하지 않음.
- 인증: Firebase ID Token 서버 검증(Admin SDK), Selective Auth(메서드 단위) 적용.
- 안정성: Circuit Breaker, ES, 재시도/백오프.
  - 참고: Connect 프로토콜(JSON/streaming)은 gRPC‑only 경로에서 Protocol Filter로 차단.

## 3.3 Client
- Flutter + BLoC.
- gRPC/HTTP 통신, 스트리밍 기반 실시간 상태 동기화.

## 3.4 Interface
- 서비스별 gRPC 사양: ./grpc-spec.md (gRPC‑Web 호환)
- 실시간: 스트리밍 API + Pub/Sub-Redis 동기화.

## 3.5 Authentication
- 소셜 로그인(Google/Apple/Kakao/Naver) → Firebase Auth.
- 서버에서 ID Token 검증, 최소 권한 부여.

## 3.6 Infra & Operations
- Cloud Run 배포, Cloud SQL/Memorystore 연결.
- Pub/Sub 이벤트 플로우, CI/CD(GitHub Actions), IaC(Terraform).
- 관측성: 구조화 로깅, 분산 트레이싱, 메트릭·알림.

## 3.7 Configuration & Secrets
- 환경변수 우선, 파일 구성(config.yaml) 보조, 기본값 최후 — 우선순위 합성.
- Cloud Run 환경변수 + Secret Manager로 민감정보 관리(DSN/자격증명, Firebase Admin JSON은 볼륨 마운트).
- 예: `POSTGRES_DSN`, `REDIS_ADDR`, `FIREBASE_CREDENTIALS_JSON`, `FIREBASE_PROJECT_ID` 등.

## 3.8 Database & Cache Details (Enhanced)

### PostgreSQL Connection Pooling
- 구성: MaxOpenConns, MaxIdleConns, ConnMaxLifetime, ConnMaxIdleTime, ConnectTimeout.
- 관측: 풀 통계 메트릭 수집(열림/유휴/대기/닫힘), 헬스체크 엔드포인트.

### PostgreSQL Migration Strategy
- 도구: golang-migrate
- 위치: `db/migrations/`
- 버전 전략: 6자리 시퀀스 번호 + 설명 (`000001_create_users_table.up.sql` / `.down.sql`)

### Migration Operations
| Command | Purpose | Environment |
|---------|---------|-------------|
| `make migrate-up` | Apply pending migrations | All |
| `make migrate-down N` | Rollback N migrations | Dev/Staging |
| `make migrate-reset` | Reset to clean state and reapply all | Dev |
| `make migrate-fresh` | Drop schema + recreate + migrate | Dev |
| `make migrate-force VERSION=NNNNNN` | Force specific version (recovery) | Emergency |

### Migration Safety Practices
- 모든 마이그레이션은 up/down 쌍으로 작성.
- 스테이징에서 롤백 절차 검증, 프로덕션 전 백업 필수.
- 스키마 변경은 PR 리뷰·승인 절차 준수.

### Redis Cache Utility
- 기능: Set/Get/Delete/TTL, `ErrCacheMiss` 구분, 패턴 삭제로 테스트 정리.
- 클라이언트: go-redis 기반, DI 통합, 동시성 안전.

## 3.9 Testing Strategy(요약)
- 단위/통합: Ginkgo+Gomega.
- E2E: Godog v0.14 + Gherkin, 프로토콜 태그(@rest/@grpc), 도메인 태그(@health/@user/@quiz), 차단 검증 태그(@protocol_filter).
- 실행: `make test.e2e`, 시나리오 아웃라인/동시성 시나리오로 커버리지 강화.

## 3.10 Configuration Management Strategy

### Configuration Sources Priority
1. 환경변수 (최상위 우선순위)
2. 구성 파일 (`config.yaml`, `.env`)
3. 기본값 (최후 수단)

### Environment-Specific Strategies
| Environment | Primary Method | Secondary Method | Notes |
|-------------|---------------|------------------|-------|
| Development | .env + Docker Compose | config.yaml | 로컬 개발 단순화 |
| Staging | Cloud Run env vars | - | 배포·테스트 용이 |
| Production | Cloud Run env vars + Secret Manager | - | 민감정보 보안 강화 |

### Configuration Loading Process
- 개발 환경에서 `.env` 자동 로딩, 필수/선택 변수 명시 및 검증.
- `.env.template` 기반 템플릿 제공.
- 누락된 필수 값 검증 및 오류 처리.

### Secret Management
- Google Cloud Secret Manager 통합.
- 서비스 계정 키 파일 관리 및 환경별 격리.
- CI/CD 시크릿 주입 패턴 정립.

## 3.11 Development Workflow & Automation

### Essential Make Targets
| Category | Command | Purpose |
|----------|---------|---------|
| Setup | `make init` | 개발 환경 초기화 |
| Setup | `make install-tools` | 개발 도구 설치 |
| Build | `make build` | 애플리케이션 빌드 |
| Test | `make test` | 단위+통합 테스트 실행 |
| Test | `make test.e2e` | E2E 테스트 실행 |
| Test | `make test.e2e.only` | 태그/시나리오로 부분 실행 |
| Test | `make test.e2e.except` | EXCEPT 태그 세트만 실행 |
| Test | `make test.e2e.concurrency` | 무거운 동시성 시나리오 실행 |
| Quality | `make fmt` | 포맷팅 |
| Quality | `make vet` | 정적 분석 |
| Quality | `make lint` | 린트 |
| Generate | `make generate` | 전체 코드 생성(proto, wire, mocks 등) |
| Generate | `make buf-generate` | 프로토콜 버퍼 코드 생성 |
| Database | `make migrate-up` | 마이그레이션 적용 |
| Run | `make run` | 개발 서버 실행 |

### Code Generation Pipeline
1. Protocol Buffers: `buf generate`
2. Dependency Injection: `wire`
3. Mocks: `mockgen`
4. OpenAPI: `ogen`
5. Test Suites: Ginkgo suite 생성

## Architecture Details
- 자세한 아키텍처 기술 세부는 ./architecture-appendix.md 문서를 참고합니다.

## Related Documents
- **Architecture**: ../adr/realtime-architecture.md
- **Implementation**: ./implementation-overview.md
- **Status Tracking**: ./implementation-status.md
- **Testing**: ./testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
