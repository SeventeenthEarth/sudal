# Usage Scenarios

## 방 생성부터 종료까지
- 방 생성: 호스트가 인원·퀴즈 세트 선택 → 방 생성 → QR 코드 발급.
- 참여: QR 스캔 → 입장(미풀이자는 풀이로 유도).
- 사탕 모으기: 참여자 수만큼 팟 충족 시 비교 준비 완료.
- 비교 시작: 전원 ‘시작’ 동시 클릭 → 비교 실행.
- 결과/기록: 문항별 일치 시각화, 사진 업로드, 개인 메모 작성.
- 종료: 결과 확인 후 방 소멸; 이력은 DB에 저장되어 재확인 가능.

## 정책/제약
- 백그라운드 이탈: 참여자가 백그라운드 전환 시 즉시 이탈, 호스트 이탈 시 즉시 해체.
- 최대 인원: 2~12명; 초과 시 Join 거부.
 - 결과 시각화: 최대 12명까지 한 화면 요약(히트맵/아이콘) + 상세 드릴다운. 세부 UX는 추가 고려 사항 문서 참조.

## Related Documents
- **Architecture**: ../adr/realtime-architecture.md
- **Implementation**: ./implementation-overview.md
- **Status Tracking**: ./implementation-status.md
- **Testing**: ./testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
