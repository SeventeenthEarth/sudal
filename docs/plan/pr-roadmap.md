# PR 로드맵 (Phase & Area)

본 문서는 기존 PR 계획을 Phase/영역별로 재구성하고 `Status` 컬럼을 추가한 정리본입니다. 상세 작업/주의사항은 각 행에 포함되어 있습니다. 상태는 requirement/implementation-status.md와 동기화합니다.

참고: PR-1부터 PR-90까지 총 90개가 정의되어 있습니다. (안정성/확장성 패턴 섹션: PR-86~90)

## Phase 1 — Setup & MVP

### 1. 초기 환경 & 도구 설정

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-1  | 프로젝트 기본 구조 & 문서화 | Completed | - Go 모듈 초기화- 최소 디렉토리 구조</br>- .gitignore, README, LICENSE | N/A |
| PR-2  | 기본 Ping/Health Check 서버 | Completed | - 간단한 HTTP 핑/헬스체크 엔드포인트</br>- 포트·라우팅 기본 설정 | N/A |
| PR-3  | Docker 환경 구성 | Completed | - Dockerfile 작성</br>- docker-compose.yml 기본 설정</br>- (옵션) hot reload(air 등) 적용 | N/A |
| PR-4  | 개발 도구 & 린팅/테스트 환경 설정 | Completed | - golangci-lint 적용</br>- 단위 테스트 프레임워크 도입</br>- Makefile로 빌드/테스트/린트 일관화 | N/A |
| PR-5  | 기본 설정 관리 유틸리티 | Completed | - 환경 변수 로딩 (dotenv, Viper 등)</br>- 설정 구조체 정의 & 검증 로직 | N/A |
| PR-6  | 구조화 로깅 시스템 도입 | Completed | - zap/zerolog 통합- 로깅 레벨 & 포맷 설정</br>- Trace ID, 컨텍스트 로깅 | N/A |

### 2. Connect-go & Proto 빌드 파이프라인

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-7  | Connect-go 통합 (기본 서비스 예시) | Completed | - Connect-go 의존성 설정</br>- health check용 Proto 작성</br>- connect-go 라우터 연결 | N/A |
| PR-8  | Protocol Buffers 빌드 시스템 | Completed | - buf.yaml, buf.gen.yaml 설정</br>- Proto 컴파일 자동화 스크립트</br>- Makefile 통합 | N/A |

### 3. PostgreSQL 연동 & 마이그레이션

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-9  | PostgreSQL 도커 환경 구성 | Completed | - docker-compose.yml에 PostgreSQL 추가</br>- DB 포트, 볼륨, 네트워크 설정 | N/A |
| PR-10 | DB Driver & Env Config | Completed | - Go DB 드라이버(github.com/lib/pq 등) 의존성 추가</br>- DB 연결용 환경 변수(호스트·포트·계정 등) 로드 로직 구현</br>- .env 또는 Viper 설정 통합 | - .env 파일을 버전관리할지 여부 결정</br>- 프로덕션/개발 환경 분리 (별도 ENV) 방안 |
| PR-11 | Connection Pool & Basic Test | Completed | - DB 풀 설정(MaxOpenConns 등) 구성</br>- 연결 성공/실패 간단 테스트 코드 작성</br>- 초기 로깅(zap/zerolog)으로 DB 연결 시 에러 추적 | - Pool 크기, 타임아웃, SSL 설정(TLS) 등 확정</br>- 로깅 포맷(JSON vs 텍스트) 결정 |
| PR-12 | Migration Tool Integration | Completed | - golang-migrate(또는 goose) 의존성 추가</br>- 마이그레이션 스크립트 위치(예: db/migrations) 및 버전 관리</br>- 간단한 샘플 마이그레이션 적용/롤백 테스트 | - 마이그레이션 버전 관리 전략(폴더 구조, 네이밍 룰)</br>- DB 변경 시 수동 롤백 프로세스 정의 |
| PR-13 | Initial Schema Migration | Completed | - users 테이블 등 최소 스키마 예시 마이그레이션 파일 작성</br>- 서버 구동 시 자동 마이그레이션 옵션(make migrate 등) 적용 | - users 테이블 컬럼 구체화(ex. 닉네임 최대 길이)</br>- timezone(UTC vs local) 처리 방안 |

### 4. Redis 연동 & 기본 캐싱

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-14 | Redis Docker & Compose Setup | Completed | - docker-compose.yml에 Redis 추가</br>- Redis 포트/볼륨 설정, 기본 환경 세팅정 | - 로컬 개발 시 Redis 데이터 볼륨 유지/초기화 여부 결정</br>- 개발/테스트용 separate Redis 스펙 고려 |
| PR-15 | Redis Basic Ping Test | Completed | - go-redis(또는 redigo) 클라이언트 의존성 추가</br>- Redis PING 응답 테스트, 초간단 캐싱 예제</br>- go-redis 등 클라이언트 설정</br>- 연결 상태 확인 & 기본 캐싱 유틸 | - 로컬에서 Redis 연결이 안 될 때 대체 옵션(모의 테스트 등) |
| PR-16 | Redis Client Configuration | Completed | - Redis 연결 풀(ConnPool) 설정</br>- Redis 관련 에러 처리 로직, timeouts, retry 정책 수립 | - 연결 풀 크기, 만료시간 등 설정값 확정</br>- 클라이언트 라이브러리 버전(major update 여부) 확인 |
| PR-17 | Cache Utility & Integration Test | Completed | - 간단한 key-value 캐싱 함수 작성</br>- 통합테스트(서버 기동 후 Redis 접근)로 read/write 확인 | - 통합테스트 시 실제 Redis에 write하는 시점(테스트가 병렬일 경우 key 충돌 주의)</br>- 캐시 만료 시간 설정할지 여부 |

### 5-1. 저장소 레이어 패턴 도입

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-18 | Repository Interfaces (User) | Completed | - UserRepository 인터페이스 정의</br>- (추후 DB, Redis 등 구현 분리 가능토록 인터페이스화) | - interfaces 패키지 구조 논의(도메인별 분리 vs 단일화</br>- 다중 저장소(DB+Redis) 혼합 사용 시 Read/Write 우선순위 결정 |
| PR-19 | Repository Interfaces (Quiz, Comparison, Candy) | Completed | - QuizRepository, ComparisonRepository, CandyRepository 등 도메인별 인터페이스 구조 설계</br>- 의존성/패키지 구조 점검 | - 파일/패키지 분할 전략(퀴즈, 비교, 사탕 등 각각 모듈화) |
| PR-20 | Base Repo Implementation & Tests | Completed | - PostgreSQL 구현체 초안 + 단위 테스트</br>- 추후 세부 도메인 별로 확장 (Transaction, JOIN 등) | - 트랜잭션 단위(메서드별 트랜잭션 vs 상위에서 컨트롤) 결정</br>- Repo 레벨에서 에러 정의(에러 wrapping) |

### 5-2. 사용자(User) 도메인

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-21 | user.proto & Code Generation | Completed | - user.proto에 메시지/서비스 정의 (Register, GetProfile 등)</br>- buf.yaml/buf.gen.yaml 설정 & 코드 생성</br>- 간단 서버/클라이언트 스텁 검증 | - user.proto에서 user_id 타입(string vs int64) 확정</br>- display_name, avatar_url 등 필드 형식/크기 |
| PR-22 | UserService Scaffolding | Completed | - UserService 골격 작성</br>- connect-go 라우팅 설정</br>- health check 수준의 서비스 동작 확인 | - 스캐폴딩 시 헬스체크 엔드포인트 분리 여부</br>- 인증/인가 미들웨어 위치(전역 vs 개별 서비스) 결정 |
| PR-23 | User Table Schema & Migration | Completed | - users 테이블 확정 스키마(닉네임·아바타·firebase_uid·생성일시 등) 작성</br>- 마이그레이션 스크립트 추가 & 기존 유저 테이블 초안 수정(있다면) | - firebase_uid 유니크 인덱스 여부 결정</br>- nickname not null 제약조건(혹은 default 값) 고려 |
| PR-24 | UserRepository Implementation | Completed | - PostgreSQL 기반 CRUD 구현</br>- firebase_uid로 유저 조회/생성/업데이트 로직 | - firebase_uid 중복 등록 시 에러 처리</br>- Soft delete(사용자 비활성화) 필요 여부 |
| PR-25 | UserService Logic & Unit Test | Completed | - RegisterUser, GetUserProfile 등 실제 비즈니스 로직</br>- 단위테스트(메모리 모크 or 테스트 DB 활용) | - RegisterUser 시 프로필 중복 처리(이미 존재하는 UID면 업데이트? 에러?)</br>- 닉네임 유효성 검증(금칙어 등) |
| PR-26 | Integration Test (UserService) | Completed | - 서버 구동 후 connect-go 클라이언트로 E2E 흐름 테스트</br>- edge case(중복 등록, 필드 누락 등) 처리 | - 인증 토큰 없는 경우 등 오류 케이스 테스트</br>- 생성/수정/조회 순서 종합 시나리오 |
| PR-27 | Firebase Admin SDK & Token Verification | Completed | - Firebase Admin SDK 설정</br>- ID Token 검증 로직 & 미들웨어 작성</br>- 인증 실패 시 에러 응답 처리 | - 테스트 환경에서 Firebase 인증 모킹 방법</br>- 구글/애플/카카오/네이버 등 소셜 로그인 Federation 시나리오 |

### 5-3. 퀴즈(Quiz) 도메인

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-28 | quiz.proto & Code Generation | Planned | - 퀴즈 관련 메시지(QuizSet, Question 등) 정의</br>- buf generate 후 서버 스텁 생성/검증 | - Question 구성이 2지선다에만 한정되는지, 향후 확장성(3~4지선다)에 대한 고려</br>- quiz_set_id를 string vs UUID 결정 |
| PR-29 | QuizService Scaffolding | Planned | - QuizService 골격 / connect</br>- go 라우팅</br>- ListQuizSets, GetQuizSet 등 기본 함수 시그니처 | - 태그 기반 필터링 로직을 어떻게 적용할지(스캐폴딩 단계에선 단순히 무시할지) |
| PR-30 | QuizService API Contract Tests | Planned | - quiz.proto 메시지 구조에 맞춰 요청/응답 샘플 작성</br>- 간단한 contract test, JSON 예시 검증 | - pagination(page/page_size) 기본값 처리</br>- 없는 quiz_set_id 요청 시 HTTP Status 결정(404 vs 400 등) |
| PR-31 | Quiz & Question Tables Migration | Planned | - quiz_sets, questions 테이블 마이그레이션 스크립트</br>- FK 연계(QuizSet - Question) | - question 테이블에 option_a, option_b 길이 제한</br>- 퀴즈 세트 삭제 시 질문도 함께 삭제 Cascade 여부 결정 |
| PR-32 | QuizRepository Implementation | Planned | - QuizRepository(PostgreSQL) CRUD</br>- Tag 연관관계(있다면)도 대비 | - quiz_sets insert 시 중복 title 허용 여부</br>- 태그 매핑 시 중간테이블 사용 or 테이블 내 배열 컬럼 사용 여부 |

### 5-5. 방(Room) 도메인 + 실시간 기본

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-41 | room.proto & Code Generation | Planned | - 방 생성, 참여, 상태 스트리밍 등에 필요한 메시지 정의</br>- buf generate 후 서버 스텁 | - 방 식별자(room_id)는 string vs UUID</br>- max_participants 상한(12명) 초과 요청 시 처리 방식 |
| PR-42 | RoomService Scaffolding | Planned | - RoomService 라우팅 설정</br>- CreateRoom, JoinRoom 시그니처 | - QR 코드 URL을 어떻게 생성/저장할지(proto 상에 포함 vs 별도 API) |
| PR-43 | RoomService Message Flow | Planned | - 참여자 목록, 호스트(방장) 식별, QR코드 URL 공유 로직 설계</br>- error case(방 full 등) | - 호스트가 이탈했을 때 방 자동 해체 처리 시점</br>- 앱 백그라운드 이동 시 즉시 이탈 처리 구현 범위 |
| PR-44 | QR Code Library Integration | Planned | - QR 코드 생성 라이브러리 선택(zxing, qrcode 등)</br>- 이미지/URL 생성 로직 & 간단 테스트 | - 생성된 QR 이미지를 어디에 저장할지(파일 vs 메모리)</br>- URL encoding 여부 결정 |
| PR-45 | Room in Redis (Key Structure) | Planned | - room::participants, room::candy_pot 등 키 설계</br>- Room 저장/조회/삭제 로직 | - TTL(유효기간)을 둘지 여부</br>- 방이 종료된 후 Redis 기록 삭제 시점(즉시 vs 일정 시간 후) |
| PR-46 | RoomRepository & Tests | Planned | - RoomRepository(Redis) 기본 CRUD</br>- 간단한 유닛테스트 + Redis flush 모드에서 동작 확인 | - 방 ID 중복 시 어떤 동작? (overwrite vs 에러)</br>- 예외처리(이미 없는 방 제거 등) |
| PR-47 | RoomService Implementation | Planned | - CreateRoom, JoinRoom, LeaveRoom 로직</br>- 방 인원수 제한, 중복 참여 방지, 호스트/게스트 분기 | - join 시 중복 참여자 확인(동일 유저가 이미 있는지) 로직</br>- 방장의 권한(호스트만 강제 종료 가능?) |
| PR-48 | Integration Test (RoomService) | Planned | - 여러 명이 순차적으로 Join/Leave 하는 시나리오</br>- 만석/방장 이탈 등 edge case | - 최대인원(12명) 도달 테스트</br>- 방장 이탈로 인한 방 종료 시 동시성 문제(참여자 입장과 충돌) |
| PR-49 | Connect-go Streaming Basics | Planned | - connect-go 스트리밍(ServeStream 등) 예제</br>- 단순 Echo/Chat 예제로 테스트 | - gRPC 웹 클라이언트 지원 여부(브라우저 호환성 등) |
| PR-50 | Real-time Room Updates | Planned | - 방 내 참여자 변동, 사탕 기부, 비교 준비 상태 등을 스트리밍으로 알림</br>- 서버 측 브로드캐스트 & 클라이언트 구독 로직 | - 브로드캐스트 시 메시지 순서 보장(또는 무작위) 고려</br>- 대기 상태를 UI에서 어떻게 업데이트할지(주기적 vs 이벤트 기반) |
| PR-51 | Streaming Error Handling & Recovery | Planned | - 네트워크 끊김/타임아웃 시 재연결</br>- 백그라운드 이탈 시 즉시 방에서 제외 로직 | - 재연결 간격, 최대 시도 횟수 등 정책</br>- 클라이언트 재접속 시 Redis 상 기존 기록(중복) 어떻게 처리할지 |

## Phase 2 — Domain Build-out & Pub/Sub

### 5-4. 비교(Comparison) 도메인

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-33 | comparison.proto & Basic Definition | Planned | - ComparisonService용 프로토콜(비교 시작, 결과 조회 등) 정의</br>- buf generate 후 서버/클라이언트 스텁 | - 비교 알고리즘(문항별 선택을 어떻게 묶어 보관할지) 프로토콜 정의 명세</br>- 결과 응답 시 참여자별 선택 항목을 배열 vs map 중 무엇으로 구현할지 |
| PR-34 | ComparisonService Scaffolding | Planned | - ComparisonService 골격</br>- connect-go 라우팅 | - 시작/중간/완료 상태를 어디에서 관리할지(방 vs 비교) |
| PR-35 | Comparison Message Patterns | Planned | - 각 문항별 선택/결과를 어떻게 담을지 설계</br>- 중복 비교, 시간, 사진 등 부가정보 포함 여부 | - 사진 첨부를 어느 시점에 할지(proto 구조 내 별도 필드 분리 여부)</br>- 중복 비교 방어 로직(이미 비교 진행된 퀴즈 세트 재비교 시 허용?) |
| PR-36 | Comparisons Tables Migration | Planned | - comparisons, comparison_participants 등 테이블</br>- quiz_results와의 관계 설정 | - comparisons PK는 UUID vs auto-increment 선택</br>- 비교 시점에 사용된 quiz_result 버전 관리(사용자 재풀이 시점 혼동) |
| PR-37 | ComparisonRepository Implementation | Planned | - PostgreSQL CRUD</br>- 비교 시 필요한 JOIN/집계 로직(퀴즈 결과와 매핑) | - 대규모 비교(최대 12명) JOIN 성능 고려</br>- DB 트랜잭션 범위(비교 시작-종료) |
| PR-38 | Comparison Repo Test & Edge Cases | Planned | - 동시 비교(트랜잭션), invalid room 등 에러 처리</br>- 단위테스트, 인수테스트 | - 한 명이 중간에 퀴즈 변경했을 때 어떤 예외 발생?</br>- 중복 참여(동일 사용자 여러 번 참여)에 대한 방어 로직 |
| PR-39 | ComparisonService Logic | Planned | - StartComparison, GetComparisonResults 등 구현</br>- 퀴즈 결과를 불러와 비교 알고리즘 적용 | - 문항별 일치도 계산 시 어떤 수식/표현으로 반환?</br>- 실시간 모드와 비교 완료 상태 분리 |
| PR-40 | Integration Test (Comparison) | Planned | - 실제 DB, 실제 gRPC 호출로 E2E 시나리오 검증</br>- 사진, 개인메모 등 필드 처리를 Mock 또는 Stub 로 처리 | - 여러 참여자가 순차적으로 들어왔다 나갈 때도 정상 동작하는지 시나리오 점검</br>- 비교 후 방이 자동 종료되는지 확실히 체크 |

### Pub/Sub 통합 (GCP)

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-52 | Pub/Sub Setup (Terraform) | Planned | - Pub/Sub 토픽/구독 생성 TF 스크립트</br>- Service Account 권한 할당(Cloud Run → Pub/Sub) | - Dev/Prod 분리된 GCP 프로젝트인지 여부</br>- Pub/Sub 요금 고려(메시지 용량) |
| PR-53 | Pub/Sub Client Integration | Planned | - Go용 Pub/Sub 클라이언트 설정</br>- Topic/Subscription 초기화 코드</br>- 배포 환경 변수(GCP Project ID 등) | - Push vs Pull 방식 결정</br>- 대규모 트래픽 시 구독 확장(서버리스 Autoscaling) 고려 |
| PR-54 | Publish & Subscribe Test | Planned | - 간단 메시지 발행/수신 단위테스트</br>- Push vs Pull 구독 방식 결정 및 샘플 코드 | - 메시지 포맷(proto vs JSON) 최종 확정</br>- 메시지 타입별 토픽/구독 분할 전략 |
| PR-55 | Room Event Publishing | Planned | - Room 상태(입장/퇴장/비교시작 등) 변경 시 Pub/Sub 토픽에 메시지 발행</br>- 메시지 포맷(proto or JSON) 정의 | - 어떤 시점에 발행할지(Join 직후? 트랜잭션 커밋 직후?)</br>- 중복 이벤트 방어(같은 상태를 여러 번 발행하지 않는지) |
| PR-56 | Broadcast Room Events to Clients | Planned | - 다른 Cloud Run 인스턴스에서 메시지 수신 → 스트리밍 통해 클라이언트에게 재전송</br>- Redis/GCP Pub/Sub 간 상태 동기화 | - 메시지 수신 지연(수 초 정도)에 따른 UI 동기화 문제</br>- Pub/Sub 메시지 수신 시점과 Redis 상태 불일치 가능성 (re-check 필요) |
| PR-57 | Consistency & Duplicate Event Handling | Planned | - 동일 이벤트 중복으로 수신 시 처리 방안</br>- 메시지 Ack/Nack 전략 | - 메시지 유실시 보강 전략(재시도 vs 보류)</br>- 이벤트 중복 감지(이벤트 ID 기반 deduplication) |
| PR-58 | Pub/Sub Subscriber Implementation | Planned | - Push/Pull 구독 엔드포인트 구현</br>- 수신 메시지 구조 파싱, Redis 업데이트 | - Pull 모델 시 polling 주기</br>- Push 모델 시 인증 토큰 관리, 보안 이슈 |

### 5-7. 태그(Tag) 시스템

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-65 | Tag Tables & Migration | Planned | - tags 테이블, QuizSet-Tag 다대다 구조(있다면) 정의</br>- 마이그레이션 스크립트 | - 다대다 관계(MTM) 테이블명(quiz_set_tags 등) 확정</br>- 태그 길이 제한, 중복 태그명 처리 |
| PR-66 | TagRepository & Basic TagService | Planned | - 태그 목록 조회, 태그별 퀴즈세트 필터</br>- 단위테스트, 인수테스트 | - 태그를 수동으로만 등록할지(관리자 UI) vs 사용자 생성 허용할지</br>- 필터링 시 다중 태그 검색(AND/OR) 정책 |

## Phase 3 — Currency & Photos

### 5-6. 사탕(Candy) 도메인 (가상 화폐)

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-59 | candy.proto & Basic Definition | Planned | - CandyService(잔액 조회, 거래 등) 프로토콜 작성</br>- buf generate 후 코드 검증 | - 사탕 단위(int vs int64), 최대 보유량 제한 여부</br>- 환불/취소 시나리오 반영 여부 |
| PR-60 | CandyService Scaffolding | Planned | - CandyService 라우팅, 메서드 시그니처</br>- 초기 Health check 등 | - In-App 결제(애플, 구글) 연동 범위</br>- 테스트 결제 수단 시뮬레이션 방안 |
| PR-61 | Candy Message Patterns | Planned | - 거래(Transaction) 구조 설계(PURCHASE, EARN 등)</br>- In-App 결제 영수증 필드 등 | - purchase 시 영수증 검증 로직(서드파티 vs 자체 검증) 결정</br>- 무료지급(퀴즈 첫 풀이) 이벤트 트리거 시점 |
| PR-62 | Candy Transactions Table Migration | Planned | - candy_transactions 테이블 스키마</br>- 무결성 보장(사탕 잔액이 음수가 되면 안 됨)</br>- DB 트랜잭션 고려 | - 사용자별 잔액 실시간 업데이트 방안(SELECT FOR UPDATE?)</br>- 중복 결제나 롤백 시나리오 |
| PR-63 | CandyRepository Implementation | Planned | - PostgreSQL 기반 잔액조회, 거래내역 삽입, 트랜잭션 처리</br>- 충전/사용(기부) 로직 | - DB 트랜잭션 범위 지정(INSERT + UPDATE, 에러 처리 시점)</br>- 부정 결제/환불 시 CandyTransaction 정정 로직 |
| PR-64 | CandyService Logic & Unit Test | Planned | - GetCandyBalance, PurchaseCandy, ContributeCandy 등 비즈니스 로직</br>- 단위테스트(실제 DB or 모크) | - 하나의 방에 여러 사람이 기부하는 경우 동시성 처리</br>- 구글/애플 구매 영수증 검증 Mock 테스트 |

### 6. 사진(Firebase Storage) 연동

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-67 | Firebase Storage Config | Planned | - Firebase Storage(GCS) SDK 설정</br>- 인증 키 파일(서비스 계정) 관리 & CI 시크릿 | - 업로드 파일 크기 제한(사진 용량) 결정</br>- 보안 규칙(파일 접근 권한) 설정(퍼블릭 vs 인증 사용자 전용) |
| PR-68 | Photo Upload & URL Handling | Planned | - 업로드 API 구현</br>- 사진 URL DB 매핑(Comparison 결과 등) | - 업로드 실패 시 재시도 로직</br>- 이미지 파일 형식(JPEG/PNG) 이외 처리 |
| PR-69 | Photo Attachment in Comparison | Planned | - 비교 완료 후 사진 업로드 & 저장</br>- 등록 후 수정 불가 정책 적용 | - 사진 여러 장 업로드 시 순서 관리 필요?</br>- 한 장이라도 업로드 실패 시 전부 실패로 볼지(트랜잭션) |
| PR-70 | Personal Memo & Access Control | Planned | - 개인 메모(CRUD), 본인만 접근 가능하게 권한 체크</br>- 즐겨찾기(Favorite) 기능 여부도 함께 처리 | - 메모의 최대 길이, 이모티콘 처리 여부</br>- 즐겨찾기 표시를 DB에서 bool로 저장 vs 별도 테이블로 관리 |
| PR-80 | 사진 업로드 Circuit Breaker 적용 | Planned | - 사진 업로드 시 Circuit Breaker 패턴 적용</br>- 실패 시 로컬 캐싱 및 재시도 전략 | - 실패한 업로드 대기열 관리</br>- 네트워크 불안정 시 부분 성공 처리 방안 |
| PR-81 | 이미지 처리 이벤트 소싱 구현 | Planned | - 이미지 업로드/수정/삭제를 이벤트로 기록</br>- 이벤트 기반 이미지 처리 파이프라인 | - 이미지 메타데이터 버전 관리</br>- 이벤트 소싱을 통한 이미지 상태 복원 메커니즘 |

## Phase 4 — CI/CD, 배포, 안정성

### 7. 부하 테스트 & CI/CD & GCP 배포

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-71 | 부하 테스트 스크립트 작성 | Planned | - k6, vegeta 등 도구 선정- 주요 시나리오별 테스트(사용자 가입, 퀴즈 제출 등)- 성능 지표 & 임계값 설정 | N/A |
| PR-72 | CI/CD 파이프라인 구성 | Planned | - GitHub Actions(or 다른 CI)로 빌드/테스트/린트 자동화</br>- Docker 이미지 빌드 & 레지스트리 푸시</br>- (옵션) 커버리지 연동 | - 서버 기동/환경 변수 로딩, 시크릿 주입 |
| PR-73 | GCP 배포 설정 (Cloud Run 등) | Planned | - Cloud Run 배포 스크립트- Cloud SQL, Memorystore 연결</br>- 비밀(Secrets) 관리 및 보안 설정 | N/A |
| PR-74 | 운영 모니터링 & 알림 시스템 | Planned | - GCP Operations Suite(Logging/Monitoring/Alert) 설정</br>- 에러 리포팅, 대시보드 구성</br>- 알림(Slack 등) 규칙 | N/A |
| PR-82 | Circuit Breaker 부하 테스트 | Planned | - 서비스 장애 상황 시뮬레이션</br>- Circuit Breaker 임계값 최적화</br>- 회로 상태 변환 모니터링 | - 실제 서비스 장애를 안전하게 시뮬레이션하는 방법</br>- 부하 테스트와 실제 운영 환경의 차이점 |
| PR-83 | Event Sourcing 성능 & 일관성 테스트 | Planned | - 이벤트 처리량 측정</br>- 높은 부하에서 이벤트 소싱 테스트</br>- 이벤트 재생 성능 분석 | - 대용량 이벤트 처리 시 병목 지점 식별</br>- 이벤트 일관성 검증 방법론 |
| PR-84 | 마이크로서비스 통합 안정성 테스트 | Planned | - 서비스 간 통신 장애 시나리오</br>- Circuit Breaker와 Event Sourcing 통합 테스트</br>- 복구 프로세스 검증 | - 실제 환경과 유사한 테스트 환경 구성</br>- 분산 시스템 장애 시뮬레이션 기법 |
| PR-85 | 확장성 패턴 모니터링 및 알림 시스템 통합 | Planned | - Circuit Breaker 상태 모니터링 대시보드</br>- 이벤트 소싱 메트릭 수집</br>- 장애 감지 및 알림 규칙 설정 | - 중요 메트릭 선별</br>- 알림 노이즈 최소화(false positive 방지) |

### 8. 안정성 및 확장성 패턴

| PR 번호 | 제목 | 상태 | 주요 작업 | 구현시 주의 사항 |
| ----- | ----- | ----- | ----- | ----- |
| PR-86 | Circuit Breaker 패턴 도입 | Planned | - Circuit Breaker 라이브러리(Sony/gobreaker 등) 통합</br>- 서비스 호출 래퍼 구현</br>- 임계값 및 타임아웃 설정 | - 임계값 설정 최적화(너무 민감하거나 둔감하면 안 됨)</br>- Fallback 메커니즘 구현 여부 결정 |
| PR-87 | Circuit Breaker 모니터링 및 메트릭 | Planned | - 회로 상태(열림/닫힘) 모니터링</br>- 실패율, 응답 시간 등 메트릭 수집</br>- 대시보드 연동 | - 알림 임계값 설정</br>- 회로 상태 변경 시 로깅 상세도 |
| PR-88 | Event Sourcing 기본 구조 구현 | Planned | - 이벤트 스토어 스키마 설계</br>- 이벤트 저장 및 조회 기능</br>- 도메인 이벤트 정의 | - 이벤트 스키마 버전 관리 방안</br>- 이벤트 저장소 성능 고려(인덱싱, 샤딩 등) |
| PR-89 | CQRS 패턴과 Event Sourcing 통합 | Planned | - Command와 Query 모델 분리</br>- 이벤트 기반 상태 재구성</br>- 읽기 모델 프로젝션 구현 | - 읽기/쓰기 모델 간 일관성 보장 방안</br>- 이벤트 소싱 기반 스냅샷 전략 |
| PR-90 | Event Sourcing 통합 테스트 및 성능 최적화 | Planned | - 이벤트 재생(Replay) 시나리오 테스트</br>- 이벤트 스토어 인덱싱 최적화</br>- 대규모 이벤트 처리 성능 테스트 | - 이벤트 수가 많을 때 재생 성능</br>- 장기 운영 시 이벤트 저장소 크기 관리(아카이빙, 압축 등) |

## Related Documents
- **Architecture**: ../adr/realtime-architecture.md
- **Implementation**: ../requirement/implementation-overview.md
- **Status Tracking**: ../requirement/implementation-status.md
- **Testing**: ../requirement/testing-strategy.md
- **Plan**: ./iteration-plan.md

## 비고
- 본 정리본은 `Status` 초기값을 `Planned`로 일괄 설정했습니다. 진행 중/완료 여부에 따라 `In Progress`/`Done` 등으로 업데이트 가능합니다.
