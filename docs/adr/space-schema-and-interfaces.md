# ADR: Space Schema & Interfaces

## 상태
- Accepted — 2025-08-29

## 배경
- Space는 실시간 동기화의 기본 단위이며, 앱 도메인(퀴즈/비디오/경매 등)에 따라 의미가 달라집니다.
- 공통 스키마와 인터페이스를 강제해 재사용성과 일관성을 확보합니다.

## 공통 데이터베이스 스키마

```sql
-- Core가 관리하는 필수 테이블
CREATE TABLE spaces (
    -- 필수 컬럼 (Core가 강제)
    space_id          UUID PRIMARY KEY,
    space_type        VARCHAR(50) NOT NULL,  -- 'quiz', 'video', 'auction' 등
    state_version     INTEGER NOT NULL DEFAULT 1,
    participant_count INTEGER NOT NULL DEFAULT 0,
    max_participants  INTEGER,
    host_id          VARCHAR(255),
    status           VARCHAR(50) NOT NULL,  -- 'waiting', 'active', 'completed', 'expired'
    
    -- 상태 저장
    state_data       JSONB NOT NULL,  -- 각 앱의 커스텀 상태
    
    -- 시간 관련
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at       TIMESTAMP,
    
    -- 메타데이터
    metadata         JSONB,  -- 앱별 추가 정보
    
    -- 인덱스
    INDEX idx_space_type (space_type),
    INDEX idx_status (status),
    INDEX idx_expires_at (expires_at)
);

-- 참여자 관리 테이블
CREATE TABLE space_participants (
    space_id         UUID NOT NULL REFERENCES spaces(space_id),
    user_id          VARCHAR(255) NOT NULL,
    joined_at        TIMESTAMP NOT NULL DEFAULT NOW(),
    left_at          TIMESTAMP,
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    role             VARCHAR(50),  -- 'host', 'participant' 등
    
    PRIMARY KEY (space_id, user_id),
    INDEX idx_user_spaces (user_id, is_active)
);

-- 이벤트 소싱을 위한 테이블 (Option 3에서 사용)
CREATE TABLE space_events (
    event_id         UUID PRIMARY KEY,
    space_id         UUID NOT NULL REFERENCES spaces(space_id),
    event_type       VARCHAR(100) NOT NULL,
    event_data       JSONB NOT NULL,
    actor_id         VARCHAR(255),
    occurred_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    event_version    INTEGER NOT NULL,
    
    INDEX idx_space_events (space_id, occurred_at),
    INDEX idx_event_type (event_type)
);
```

## Space 인터페이스(Go)

```go
// core/space.go
type Space interface {
    // 필수 메서드 (Core가 사용)
    GetID() string
    GetType() string            // "quiz", "video", "auction" 등
    GetVersion() int
    GetStatus() SpaceStatus     // "waiting", "active", "completed", "expired"
    GetParticipantCount() int
    GetMaxParticipants() int
    GetHostID() string
    
    // 상태 관리
    GetState() State
    SetState(State)
    UpdateVersion()
    
    // 시간 관리
    GetCreatedAt() time.Time
    GetUpdatedAt() time.Time
    GetExpiresAt() *time.Time
}

// 각 앱에서 구현
type QuizSpace struct {
    BaseSpace    // Core가 제공하는 기본 구현
    QuizState    // 퀴즈 특화 상태
}

type VideoSpace struct {
    BaseSpace
    VideoState   // 비디오 특화 상태
}
```

## 관련 문서
- 옵션 비교: ./realtime-options-comparison.md
- 실시간 아키텍처 결정: ./realtime-architecture.md

## Implementation Status
- Design: COMPLETED
- Database Schema: DESIGNED (not implemented)
- Go Interfaces: DESIGNED (not implemented)
- Phase 1 Target: Basic Space CRUD operations
- Phase 2 Target: Event-driven state management
- Phase 3 Target: Full plugin architecture

## Related Documents
- **Architecture**: ./realtime-architecture.md
- **Implementation**: ../requirement/implementation-overview.md
- **Status Tracking**: ../requirement/implementation-status.md
- **Testing**: ../requirement/testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
