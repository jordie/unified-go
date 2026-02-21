# Phase 8: Real-Time Updates & WebSockets

## Overview
Build a real-time notification and streaming system that keeps the dashboard, leaderboards, and user progress synchronized in real-time across all connected clients.

## Goals
- ✅ Live leaderboard updates (rankings change instantly)
- ✅ Real-time session progress (WPM, accuracy, score tracking)
- ✅ Achievement notifications (badges, milestones reached)
- ✅ User activity feed (recent completions, new high scores)
- ✅ Competitive streaks (live rankings during sessions)
- ✅ Cross-app synchronization (changes in one app visible everywhere)

---

## Architecture

### Components to Build

```
WebSocket Server (pkg/realtime/)
├── Hub (manages connections and broadcasts)
├── Client (represents connected user)
├── Messages (event types and payloads)
├── Handlers (message processing)
└── Store (in-memory state management)

Event Bus (pkg/events/)
├── Publisher (emits events from apps)
├── Subscriber (listens for events)
├── Event Types
│   ├── SessionStarted
│   ├── SessionEnded
│   ├── ScoreUpdated
│   ├── AchievementUnlocked
│   ├── RankChanged
│   └── StreakMilestone
└── Middleware (event routing)

Dashboard API Updates (pkg/dashboard/)
├── WebSocket endpoints
├── Live leaderboard streaming
├── Progress channel subscriptions
└── Achievement broadcasting
```

### Real-Time Channels

1. **Leaderboard Channels** (per category)
   - `leaderboard:typing_wpm` - Typing rankings
   - `leaderboard:math_accuracy` - Math rankings
   - `leaderboard:reading_comprehension` - Reading rankings
   - `leaderboard:piano_score` - Piano rankings
   - `leaderboard:overall` - Overall rankings

2. **User Channels** (per user)
   - `user:{userID}:progress` - User's session progress
   - `user:{userID}:achievements` - User's new achievements
   - `user:{userID}:rank-changes` - User's rank movements

3. **Activity Channels**
   - `activity:feed` - Global activity feed
   - `activity:achievements` - New achievements globally
   - `activity:high-scores` - Recent high scores

4. **Session Channels**
   - `session:{sessionID}:live` - Live session metrics
   - `session:{sessionID}:competitors` - Competing users

---

## Implementation Plan

### Subtask 8.1: WebSocket Server Foundation (2 hours)

**Files to create:**
- `pkg/realtime/hub.go` - Central connection manager
- `pkg/realtime/client.go` - Client connection wrapper
- `pkg/realtime/message.go` - Message types and structures
- `pkg/realtime/hub_test.go` - Hub tests

**Key functions:**
```go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan interface{}
    register   chan *Client
    unregister chan *Client
}

type Client struct {
    hub       *Hub
    conn      *websocket.Conn
    userID    uint
    channels  map[string]bool
    send      chan interface{}
}

type Message struct {
    Type    string                 `json:"type"`
    Channel string                 `json:"channel"`
    Data    map[string]interface{} `json:"data"`
    UserID  uint                   `json:"user_id"`
}
```

### Subtask 8.2: Event Bus & Pub/Sub (2 hours)

**Files to create:**
- `pkg/events/bus.go` - Event publisher and subscriber
- `pkg/events/types.go` - Event type definitions
- `pkg/events/handler.go` - Event routing and handling
- `pkg/events/bus_test.go` - Bus tests

**Event types:**
```go
type EventType string

const (
    EventSessionStarted      EventType = "session.started"
    EventSessionEnded        EventType = "session.ended"
    EventScoreUpdated        EventType = "score.updated"
    EventAchievementUnlocked EventType = "achievement.unlocked"
    EventRankChanged         EventType = "rank.changed"
    EventStreakMilestone     EventType = "streak.milestone"
    EventHighScore           EventType = "high.score"
    EventLeaderboardUpdate   EventType = "leaderboard.update"
)

type Event struct {
    Type      EventType
    UserID    uint
    App       string
    Timestamp time.Time
    Data      map[string]interface{}
}
```

### Subtask 8.3: WebSocket Endpoints & Handlers (2 hours)

**Files to modify/create:**
- `pkg/dashboard/router.go` - Add WebSocket route
- `pkg/dashboard/ws_handler.go` - WebSocket connection handler
- `pkg/dashboard/subscriptions.go` - Channel subscription logic

**New endpoints:**
```
GET /api/ws - WebSocket upgrade endpoint
POST /api/ws/subscribe - Subscribe to channels
POST /api/ws/unsubscribe - Unsubscribe from channels
```

### Subtask 8.4: Real-Time Leaderboard Updates (2 hours)

**Files to create:**
- `pkg/dashboard/leaderboard_streaming.go` - Leaderboard event handling
- `pkg/dashboard/rank_tracker.go` - Rank change detection

**Features:**
- Detect rank changes when scores update
- Broadcast rank changes to leaderboard subscribers
- Live rank animation data
- Rank velocity (moving up/down per session)

### Subtask 8.5: Session Progress Streaming (2 hours)

**Files to create:**
- `pkg/dashboard/session_streaming.go` - Live session metrics
- `pkg/dashboard/progress_tracker.go` - Progress updates

**Streaming data:**
- Current WPM (typing)
- Current accuracy (all apps)
- Current score (piano)
- Time elapsed
- Metrics trend

### Subtask 8.6: Achievement & Milestone Notifications (1.5 hours)

**Files to create:**
- `pkg/dashboard/achievement_notifier.go` - Achievement detection
- `pkg/dashboard/milestone_tracker.go` - Milestone events

**Achievement types:**
- Streak milestones (7 days, 30 days, 100 days)
- Score milestones (100, 500, 1000 points)
- Rank milestones (Top 10, Top 5, #1)
- Skill milestones (Level up)
- Consistency milestones (Perfect accuracy)

### Subtask 8.7: Activity Feed & Notifications (1.5 hours)

**Files to create:**
- `pkg/dashboard/activity_feed.go` - Activity log streaming
- `pkg/dashboard/notification_queue.go` - Notification ordering

**Feed events:**
- User completed session
- User improved rank
- User unlocked achievement
- User reached high score
- Leaderboard leaders changed

### Subtask 8.8: Client-Side Streaming (WebSocket Client) (2 hours)

**Files to create:**
- `static/js/realtime.js` - WebSocket client library
- `static/js/leaderboard-stream.js` - Leaderboard updates
- `static/js/progress-stream.js` - Session progress
- `static/js/notifications.js` - Toast notifications

**Client features:**
- Auto-reconnect with exponential backoff
- Channel subscription management
- Message queuing while disconnected
- Smooth DOM updates without flicker
- Notification sounds/toasts

### Subtask 8.9: Dashboard UI Integration (2 hours)

**Files to modify:**
- `pkg/dashboard/templates/unified.html` - Add streaming UI
- Add live leaderboard animations
- Add progress bar animations
- Add achievement notifications
- Add activity feed updates

### Subtask 8.10: Testing & Performance (2 hours)

**Files to create:**
- `pkg/realtime/integration_test.go` - WebSocket integration tests
- `pkg/realtime/load_test.go` - Performance testing
- `pkg/events/bench_test.go` - Event bus benchmarks

**Tests:**
- Multiple concurrent connections
- Message broadcasting
- Channel subscription/unsubscription
- Reconnection handling
- Memory leaks
- Message ordering
- Performance under load (1000+ concurrent connections)

---

## Data Models

### WebSocket Message Format

```json
{
  "type": "leaderboard.update",
  "channel": "leaderboard:typing_wpm",
  "data": {
    "rank": 5,
    "user_id": 123,
    "username": "speedster",
    "metric_value": 145.5,
    "previous_rank": 7,
    "rank_change": 2,
    "timestamp": "2026-02-20T18:30:00Z"
  }
}
```

### Achievement Notification

```json
{
  "type": "achievement.unlocked",
  "channel": "user:123:achievements",
  "data": {
    "achievement": "7-day-streak",
    "title": "Week Warrior",
    "description": "Completed a 7-day practice streak",
    "icon": "🔥",
    "points": 50,
    "timestamp": "2026-02-20T18:30:00Z"
  }
}
```

### Progress Update

```json
{
  "type": "progress.update",
  "channel": "user:123:progress",
  "data": {
    "app": "typing",
    "current_wpm": 125.5,
    "current_accuracy": 96.2,
    "session_duration": 300,
    "timestamp": "2026-02-20T18:30:00Z"
  }
}
```

---

## Performance Targets

| Metric | Target |
|--------|--------|
| WebSocket handshake | < 100ms |
| Message delivery | < 50ms |
| Leaderboard update broadcast | < 100ms |
| Achievement notification | < 200ms |
| Concurrent connections | 10,000+ |
| Memory per connection | < 5KB |
| CPU usage (idle) | < 2% |

---

## File Structure

```
pkg/
├── realtime/
│   ├── hub.go
│   ├── client.go
│   ├── message.go
│   ├── hub_test.go
│   └── integration_test.go
├── events/
│   ├── bus.go
│   ├── types.go
│   ├── handler.go
│   └── bus_test.go
└── dashboard/
    ├── ws_handler.go
    ├── subscriptions.go
    ├── leaderboard_streaming.go
    ├── session_streaming.go
    ├── achievement_notifier.go
    ├── activity_feed.go
    └── notification_queue.go

static/
├── js/
│   ├── realtime.js
│   ├── leaderboard-stream.js
│   ├── progress-stream.js
│   └── notifications.js
└── css/
    └── notifications.css
```

---

## Dependencies

- `github.com/gorilla/websocket` - WebSocket library
- `github.com/google/uuid` - Unique IDs for events
- Existing: chi router, sync.Map for thread safety

---

## Testing Strategy

1. **Unit Tests**
   - Hub registration/unregistration
   - Client subscription management
   - Message routing
   - Event publishing

2. **Integration Tests**
   - WebSocket connection lifecycle
   - Cross-channel messaging
   - Reconnection handling
   - Memory cleanup

3. **Load Tests**
   - 1000+ concurrent connections
   - 100+ messages per second
   - Memory stability
   - CPU usage

4. **Manual Testing**
   - Live leaderboard updates (multiple browsers)
   - Session progress streaming
   - Achievement notifications
   - Network disconnect/reconnect

---

## Success Criteria

- ✅ WebSocket server handles 10,000+ concurrent connections
- ✅ Messages delivered < 50ms latency
- ✅ Leaderboard updates in real-time (< 100ms)
- ✅ No memory leaks with long-lived connections
- ✅ Graceful reconnection handling
- ✅ UI updates smoothly without flicker
- ✅ All tests passing
- ✅ Zero data loss on disconnect

---

## Estimated Timeline

| Subtask | Hours | Cumulative |
|---------|-------|-----------|
| 8.1 WebSocket Foundation | 2 | 2h |
| 8.2 Event Bus | 2 | 4h |
| 8.3 WebSocket Endpoints | 2 | 6h |
| 8.4 Leaderboard Streaming | 2 | 8h |
| 8.5 Session Progress | 2 | 10h |
| 8.6 Achievements | 1.5 | 11.5h |
| 8.7 Activity Feed | 1.5 | 13h |
| 8.8 Client-Side | 2 | 15h |
| 8.9 UI Integration | 2 | 17h |
| 8.10 Testing & Perf | 2 | 19h |
| **Total** | **~19 hours** | **19h** |

---

## Subtask Breakdown

### Priority Order
1. 8.1 - WebSocket Foundation (blocking all others)
2. 8.2 - Event Bus (blocking streaming tasks)
3. 8.3 - WebSocket Endpoints (required for client connection)
4. 8.4 - Leaderboard Streaming (highest user impact)
5. 8.5 - Session Progress (core feature)
6. 8.6 - Achievements (engagement driver)
7. 8.7 - Activity Feed (social feature)
8. 8.8 - Client-Side (UI layer)
9. 8.9 - UI Integration (dashboard updates)
10. 8.10 - Testing & Performance (quality gates)

---

**Phase 8 Status:** READY FOR IMPLEMENTATION
**Next Step:** Begin Subtask 8.1 (WebSocket Foundation)

