# ADR: Realtime Options Comparison (Option 1/2/3)

## 상태
- Accepted — 2025-08-29 (목표: Option 2 → 3 단계적 진화)

## 개요
- Option 1: Simple Event-Driven — 빠른 구현, 낮은 추상화.
- Option 2: Event + Core — 공통 Core와 도메인 로직 분리.
- Option 3: Event + Plugin + Core — StateMachine/Plugin으로 완전 추상화.

## Option 1: Simple Event-Driven
- 개념: 이벤트 기반으로 상태 갱신/방송. 단일 앱에 최적화.
- 장점: 구현 간단, 빠른 출시, 학습 비용 낮음.
- 단점: 재사용성/유지보수 한계, 앱 확장 시 중복 증가.
- 적용: 초기 MVP, 실사용 확보 후 개선 단계로 전환.

## Option 2: Event + Core (선택)
- 개념: Space/State/Event 추상화, Core가 동기화/저장/브로드캐스트 담당.
- 장점: 플러그블 도메인 로직, 테스트 용이성, 재사용성/확장성 균형.
- 단점: Option 1 대비 초기 구조화 비용.
- 마이그레이션(1→2): Core 추출, 인터페이스 정의, 로직 분리, 테스트 강화.

## Option 3: Event + Plugin + Core
- 개념: StateMachine + SideEffect 패턴으로 각 앱 로직을 플러그인화.
- 장점: 완전 재사용, 앱 추가가 간단, 독립 테스트, 확장성 극대화.
- 단점: 초기 복잡도/오버헤드, 디버깅 레이어 증가.
- 마이그레이션(2→3): StateMachine 인터페이스, SyncEngine 구현, 플러그인 등록.

## 예시(Plugin 기반 일부)

```go
// plugins/video_space.go
type VideoSpaceStateMachine struct {
    state VideoSpaceState
}

func (m *VideoSpaceStateMachine) HandleEvent(event Event) (State, []SideEffect) {
    switch e := event.(type) {
    case PlayEvent:
        return VideoSpaceState{
            Phase:       "playing",
            CurrentTime: e.Timestamp,
        }, []SideEffect{
            SyncPlaybackEffect{Time: e.Timestamp},
        }
    }
}

// plugins/auction_space.go
type AuctionSpaceStateMachine struct {
    state AuctionSpaceState
}

func (m *AuctionSpaceStateMachine) HandleEvent(event Event) (State, []SideEffect) {
    switch e := event.(type) {
    case BidEvent:
        if e.Amount <= m.state.CurrentBid {
            return m.state, []SideEffect{ErrorEffect{Msg: "bid too low"}}
        }
        return AuctionSpaceState{
            CurrentBid: e.Amount,
            TopBidder:  e.UserID,
        }, []SideEffect{
            BroadcastBidEffect{Amount: e.Amount},
            ExtendTimerEffect{Seconds: 10},
        }
    }
}
```

또한 퀴즈 플러그인(참여자/사탕 팟/준비 상태) 예시:

```go
// plugins/quiz_space.go
type QuizSpaceStateMachine struct {
    state QuizSpaceState
}

func (m *QuizSpaceStateMachine) HandleEvent(event Event) (State, []SideEffect) {
    newState := m.state.Clone().(QuizSpaceState)
    effects := []SideEffect{}
    
    switch e := event.(type) {
    case UserJoinedEvent:
        if len(newState.Participants) >= 12 {
            return m.state, []SideEffect{ErrorEffect{Msg: "space full"}}
        }
        newState.Participants[e.UserID] = e.UserInfo
        effects = append(effects, BroadcastEffect{Type: "USER_JOINED", Data: e.UserInfo})
    case CandyContributedEvent:
        newState.CandyPot += e.Amount
        if newState.CandyPot >= len(newState.Participants) {
            newState.Phase = "ready"
            effects = append(effects, NotificationEffect{Type: "SPACE_READY"})
        }
    }
    
    m.state = newState
    return newState, effects
}
```

## 결론
- 현재는 Option 2를 표준으로 채택하고(ADR: realtime-architecture), 제품 성장과 함께 Option 3로의 진화를 준비합니다.

## Related Documents
- **Architecture**: ./realtime-architecture.md
- **Implementation**: ../requirement/implementation-overview.md
- **Status Tracking**: ../requirement/implementation-status.md
- **Testing**: ../requirement/testing-strategy.md
- **Plan**: ../plan/iteration-plan.md
