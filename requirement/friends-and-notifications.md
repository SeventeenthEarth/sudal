# 친구 및 알림 기능 요구사항

## 목표
- 친구 목록을 관리하고, 친구의 퀴즈 결과가 업데이트되면 새로운 비교 매칭을 유도하는 알림을 제공.

## 범위
- 포함: 친구 추가/수락/삭제, 친구 목록 조회, 친구 결과 업데이트 알림(푸시/인앱), 알림 센터.
- 제외: 추천 친구, 소셜 그래프 분석, 고급 추천 랭킹.

## 사용자 흐름
- 친구 추가: 사용자가 친구를 검색/요청 → 상대가 수락 → 양방향 친구 관계 성립.
- 알림: 친구가 퀴즈 결과 제출/업데이트 → 이벤트 발행 → 관련 친구에게 알림 → 사용자 탭에서 비교 시작 동작 유도.

## 데이터 모델(요약)
- friendships(user_id, friend_id, status[pending|accepted|blocked], created_at, updated_at).
- notification_events(id, type, actor_id, target_user_id, payload, created_at, read_at).

## 인터페이스(gRPC 제안)
- FriendsService
  - RequestFriend(user_id, target_id) → {pending_id}
  - RespondFriend(pending_id, accept: bool) → {status}
  - ListFriends(user_id, cursor) → {friends[]}
- NotificationService
  - ListNotifications(user_id, cursor) → {items[], unread_count}
  - MarkAsRead(user_id, ids[]) → {updated}

## 이벤트/통합
- 이벤트: QuizResultUpdated(user_id, quiz_id, room_id, version).
- 처리: 이벤트 필터링(친구 관계 기준) → 알림 생성 → 푸시 발송(비동기) → 인앱 뱃지 증가.
- 실패: 푸시 실패 시 재시도/백오프, DLQ 적재.

## 수용 기준
- 친구 수락 이후부터 해당 사용자 퀴즈 결과 업데이트 시 P95 < 500ms 내 알림 수신.
- 중복 결과 업데이트에 대한 멱등 알림(같은 version 재발송 금지).
- 차단(blocked) 상태인 관계에는 알림 미발송.
- 알림 센터에서 읽음 처리 시 다른 클라이언트에도 2초 내 동기화.

## 비기능
- 레이트 리밋: 친구 요청/분 단위 상한(스팸 방지).
- 프라이버시: 친구 목록은 본인만 조회 가능, 알림 페이로드 최소화.

## Related Documents
- **Architecture**: ../adr/realtime-architecture.md
- **Implementation**: ./implementation-overview.md
- **Status Tracking**: ./implementation-status.md
- **Testing**: ./testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
