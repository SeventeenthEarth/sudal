# 기능 요구사항 개요

## 1. 서비스 개요
- 소셜 퀴즈 플랫폼: 방(Room)에서 실시간으로 퀴즈를 함께 풀고 결과를 비교·공유.
- 핵심 실시간 개념: Space(퀴즈 진행 공간) 기반 상태 동기화.

## 2. 주요 기능
- 사용자(User): 가입/로그인(Firebase OIDC), 프로필 조회·수정.
- 퀴즈(Quiz): 문제 템플릿/세트, 태그로 검색, 진행 상태 관리.
- 방(Room): 생성/입장/퇴장, 생명주기(대기→진행→완료), 참여자 수 제한.
- 결과(Result): 제출/저장, 과거 기록 조회, 비교(2인 이상) 및 공유.
- 비교(Comparison): 결과 집계, 유사도/차이 포인트 산출, 알림 브로드캐스트.
- 태그(Tag): 퀴즈/사용자 관심사 분류, 검색 필터.
- 사탕(Currency): 참여·기여 리워드, 포인트 적립·소모 정책.
- 사진(Photos): 결과에 사진 첨부, Cloud Storage 연동.

## 3. 사용자 흐름(요약)
- 온보딩: OIDC 로그인 → 사용자 등록 → 기본 프로필 설정.
- 방 만들기/참여: 방 생성(인원/옵션) → 초대/입장 → 대기 → 시작.
- 진행: 문제 풀이 → 상태 동기화(이벤트 기반) → 제출.
- 결과: 개인/상대 결과 비교 → 사진 첨부(선택) → 공유.

## 4. 도메인별 요구사항
- 사용자
  - Firebase 인증 연동, 고유 user_id 발급, 탈퇴 시 개인 데이터 삭제 정책.
  - 프로필: 닉네임, 아바타, 관심 태그.
  - 로그인 Provider: 현재 Google/Email-Password 중심. Apple/Kakao/Naver는 백로그(별도 ADR/PR로 단계 도입).
- 퀴즈
  - 형식: 단답/객관식/스케일 등. 세트·태그 관리.
  - 진행: 진행 단계/남은 시간/참여자 수가 Space 상태에 반영.
- 방
  - 상태: waiting/active/completed/expired, 최대 인원 제한, 호스트 권한.
  - 이탈 정책: 백그라운드/타임아웃 처리, 재입장 허용 범위.
- 결과/비교
  - 저장: 제출 시 스냅샷 저장, 버전(멱등) 처리.
  - 비교: 상대 선택, 비교 지표 산출, 브로드캐스트.
- 태그/사탕
  - 태그: 생성·연결·검색. 
  - 사탕: 적립 조건/상한, 부정 사용 방지.
- 사진
  - 업로드/메타데이터/URL 저장, 실패 재시도.

## 5. 수용 기준(발췌)
- 2명 이상 동시 참여에서 상태 전파 P95 < 150ms(동일 리전).
- 동일 이벤트 중복 수신 시 멱등 처리로 최종 상태 일치.
- 방 최대 인원 초과 시 거부; 만석 해제 시 입장 허용.
- 결과 비교 시 모든 참여자에게 동일 계산 결과 방송.
 - 프로토콜 경계(REST 차단): 비즈니스 gRPC 경로에 HTTP/JSON 접근 시 404 응답 보장.

## 6. 인터페이스
- gRPC/Connect-go 기반 서비스: User/Quiz/Room/Comparison/Candy/Tag.
- 상세 스펙: ./grpc-spec.md
- 프로토콜 경계: REST는 health/readiness 전용(`/api/*`, `/docs`), 비즈니스는 gRPC/Connect(+ gRPC‑Web).

## Related Documents
- **Architecture**: ../adr/realtime-architecture.md
- **Implementation**: ./implementation-overview.md
- **Status Tracking**: ./implementation-status.md
- **Testing**: ./testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
