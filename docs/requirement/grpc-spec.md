# gRPC Spec for Core Functions

본 문서는 gRPC 사양을 분리·정리한 문서입니다. 서비스별 요청/응답 구조와 핵심 동작을 정의합니다.

## 1. 사용자 서비스 (UserService)

### 1.1. 사용자 등록 (RegisterUser)
- 기능: Firebase ID Token 검증 후 사용자 등록(토큰-우선 설계)
- 입력: `RegisterUserRequest`
  - `id_token`: string - Firebase ID Token(필수)
  - `display_name`: string (선택)
- 출력: `RegisterUserResponse`
  - `user_id`: string
  - `auth_provider`: string - 서버가 토큰에서 판별한 Provider(e.g., google, email)
  - `success`: bool
  
주:
- 서버가 `id_token`을 검증해 `firebase_uid`/`auth_provider`를 식별·저장합니다(클라이언트 입력 아님).

### 1.2. 사용자 정보 가져오기 (GetUserProfile)
- 입력: `GetUserProfileRequest` (인증 필요; 미들웨어 인증)
  - `user_id`: string
- 출력: `UserProfile`
  - `user_id`: string
  - `display_name`: string
  - `avatar_url`: string
  - `candy_balance`: int32
  - `created_at`: timestamp

### 1.3. 사용자 프로필 업데이트 (UpdateUserProfile)
- 입력: `UpdateUserProfileRequest` (인증 필요; 미들웨어 인증)
  - `user_id`: string
  - `display_name`: string (opt)
  - `avatar_url`: string (opt)
- 출력: `UpdateUserProfileResponse`
  - `success`: bool

## 2. 퀴즈 서비스 (QuizService)

### 2.1. 퀴즈 세트 목록 (ListQuizSets)
- 입력: `ListQuizSetsRequest { page, page_size, tags[] }`
- 출력: `ListQuizSetsResponse { quiz_sets[], total_count, total_pages }`

### 2.2. 퀴즈 세트 상세 (GetQuizSet)
- 입력: `GetQuizSetRequest { quiz_set_id }`
- 출력: `QuizSet { quiz_set_id, title, description, tags[], questions[] }`

### 2.3. 퀴즈 결과 제출 (SubmitQuizResult)
- 입력: `SubmitQuizResultRequest { user_id, quiz_set_id, answers[]bool }`
- 출력: `SubmitQuizResultResponse { result_id, timestamp, candy_earned }`

### 2.4. 사용자 퀴즈 이력 (GetUserQuizHistory)
- 입력: `GetUserQuizHistoryRequest { user_id, page, page_size, quiz_set_id? }`
- 출력: `GetUserQuizHistoryResponse { history[], total_count, total_pages }`

## 3. 방 서비스 (RoomService)

### 3.1. 방 생성 (CreateRoom)
- 입력: `CreateRoomRequest { host_user_id, quiz_set_id, max_participants }`
- 출력: `CreateRoomResponse { room_id, qr_code_url, room_code }`

### 3.2. 방 참여 (JoinRoom)
- 입력: `JoinRoomRequest { user_id, room_id }`
- 출력: `JoinRoomResponse { success, room_info }`

### 3.3. 방 상태 스트리밍 (StreamRoomState)
- 입력: stream `RoomStateRequest { room_id, user_id, action, action_data }`
- 출력: stream `RoomStateUpdate { room_id, update_type, room_state }`

### 3.4. 방 나가기 (LeaveRoom)
- 입력: `LeaveRoomRequest { user_id, room_id }`
- 출력: `LeaveRoomResponse { success }`

## 4. 비교 서비스 (ComparisonService)

### 4.1. 비교 시작 (StartComparison)
- 입력: `StartComparisonRequest { room_id, initiator_user_id }`
- 출력: `StartComparisonResponse { comparison_id, success }`

### 4.2. 비교 결과 (GetComparisonResults)
- 입력: `GetComparisonResultsRequest { comparison_id }`
- 출력: `ComparisonResults { comparison_id, quiz_set_id, quiz_set_title, participants[], results[], timestamp }`

### 4.3. 사진 추가 (AddComparisonPhotos)
- 입력: `AddComparisonPhotosRequest { comparison_id, photo_data[] }`
- 출력: `AddComparisonPhotosResponse { photo_urls[], success }`

### 4.4. 개인 메모 (UpdatePersonalMemo)
- 입력: `UpdatePersonalMemoRequest { comparison_id, user_id, memo }`
- 출력: `UpdatePersonalMemoResponse { success }`

### 4.5. 사용자 비교 이력 (GetUserComparisonHistory)
- 입력: `GetUserComparisonHistoryRequest { user_id, page, page_size, favorites_only }`
- 출력: `GetUserComparisonHistoryResponse { comparisons[], total_count, total_pages }`

### 4.6. 즐겨찾기 토글 (ToggleComparisonFavorite)
- 입력: `ToggleComparisonFavoriteRequest { comparison_id, user_id, is_favorite }`
- 출력: `ToggleComparisonFavoriteResponse { success }`

## 5. 가상 화폐(사탕) 서비스 (CandyService)

### 5.1. 잔액 조회 (GetCandyBalance)
- 입력: `GetCandyBalanceRequest { user_id }`
- 출력: `GetCandyBalanceResponse { balance }`

### 5.2. 사탕 구매 (PurchaseCandy)
- 입력: `PurchaseCandyRequest { user_id, product_id, receipt_data, store_type }`
- 출력: `PurchaseCandyResponse { transaction_id, amount_purchased, new_balance, success }`

### 5.3. 사탕 기부 (ContributeCandy)
- 입력: `ContributeCandyRequest { user_id, room_id, amount }`
- 출력: `ContributeCandyResponse { success, new_balance, pot_total }`

### 5.4. 거래 이력 (GetCandyTransactionHistory)
- 입력: `GetCandyTransactionHistoryRequest { user_id, page, page_size }`
- 출력: `GetCandyTransactionHistoryResponse { transactions[], total_count, total_pages }`

## 6. 태그 서비스 (TagService)

### 6.1. 태그 목록 (ListTags)
- 입력: `ListTagsRequest {}`
- 출력: `ListTagsResponse { tags[] }`

## Related Documents
- **Architecture**: ../adr/realtime-architecture.md
- **Implementation**: ./implementation-overview.md
- **Status Tracking**: ./implementation-status.md
- **Testing**: ./testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
