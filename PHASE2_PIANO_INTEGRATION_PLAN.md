# Phase 2 - Piano App Integration Plan

**Branch**: `feature/phase2-piano-integration-0220`
**Project**: Unified Educational Platform (Go Edition)
**Target Date**: February 20 - March 3, 2024
**Status**: ✅ PLANNING PHASE

---

## 🎯 Phase 2 Objectives

Integrate the completed Piano app foundation into the unified server, implement frontend UI, and enable core functionality for piano practice and lessons.

### Primary Goals
1. ✅ Wire Piano router into main application server
2. ✅ Create database migrations for Piano tables
3. ✅ Implement frontend UI (HTML/CSS/JS)
4. ✅ Integrate MIDI playback and recording
5. ✅ Connect to authentication system
6. ✅ Complete end-to-end testing

---

## 📋 Subtasks (18 Total)

### Subtask 2.1: Database Integration (2 hours)
**Goal**: Create and verify Piano database tables

**Tasks**:
- [ ] Create `migrations/002_create_piano_tables.sql`
  - [ ] songs table with BLOB MIDI support
  - [ ] piano_lessons table with foreign keys
  - [ ] practice_sessions table with recordings
  - [ ] Create indexes for common queries
  - [ ] Add constraints and validations

- [ ] Add migration runner support to `internal/database/migrations.go`
  - [ ] Load and execute piano migrations
  - [ ] Verify table creation
  - [ ] Add migration status tracking

- [ ] Test migrations against SQLite
  - [ ] Verify table structure
  - [ ] Test BLOB storage
  - [ ] Verify foreign key constraints

**Acceptance Criteria**:
- ✅ All Piano tables created successfully
- ✅ BLOB columns work for MIDI files
- ✅ Foreign keys enforce referential integrity
- ✅ Indexes optimized for common queries

---

### Subtask 2.2: Server Integration (2 hours)
**Goal**: Wire Piano router into main application

**Tasks**:
- [ ] Update `internal/router/router.go`
  - [ ] Import piano package
  - [ ] Create Piano router instance
  - [ ] Mount routes at `/piano` prefix
  - [ ] Test route registration

- [ ] Update `cmd/server/main.go`
  - [ ] Ensure Piano router gets database connection
  - [ ] Verify dependency injection
  - [ ] Add logging for Piano endpoints

- [ ] Update navigation/links
  - [ ] Add Piano link to dashboard
  - [ ] Update header navigation
  - [ ] Add sidebar menu item

- [ ] Integration testing
  - [ ] Test all Piano endpoints
  - [ ] Verify routing works
  - [ ] Check database connections

**Acceptance Criteria**:
- ✅ Piano endpoints respond correctly
- ✅ Database operations work
- ✅ Navigation links functional
- ✅ No conflicts with other apps

---

### Subtask 2.3: Frontend - Homepage (3 hours)
**Goal**: Create Piano app homepage with song listing

**Tasks**:
- [ ] Create `pkg/piano/templates/index.html`
  - [ ] Piano app branding (title, logo)
  - [ ] Welcome section
  - [ ] Quick start instructions
  - [ ] Navigation back to dashboard
  - [ ] Responsive design

- [ ] Create song listing page
  - [ ] Fetch songs from API (`GET /api/songs`)
  - [ ] Display as grid/table
  - [ ] Show difficulty badges
  - [ ] Add "Start Practice" buttons
  - [ ] Pagination support

- [ ] Create styling (`static/css/piano.css`)
  - [ ] Match app theme
  - [ ] Responsive layout
  - [ ] Music-themed design elements
  - [ ] Proper spacing and typography

- [ ] Add JavaScript functionality
  - [ ] Fetch songs from backend
  - [ ] Handle loading states
  - [ ] Error handling
  - [ ] Difficulty filtering

**Acceptance Criteria**:
- ✅ Homepage loads without errors
- ✅ Songs display correctly
- ✅ Responsive on mobile/tablet/desktop
- ✅ Navigation working
- ✅ Styling matches app theme

---

### Subtask 2.4: Frontend - Practice Session UI (4 hours)
**Goal**: Create practice/lesson interface

**Tasks**:
- [ ] Create `pkg/piano/templates/practice.html`
  - [ ] Song details (title, composer, BPM)
  - [ ] MIDI player placeholder
  - [ ] Performance metrics display
  - [ ] Start/Stop controls
  - [ ] Recording status indicator

- [ ] Create score display
  - [ ] Accuracy percentage
  - [ ] Tempo accuracy
  - [ ] Composite score
  - [ ] Visual feedback (progress bar)
  - [ ] Comparison to best score

- [ ] Add timing and tracking
  - [ ] Elapsed time display
  - [ ] Notes counter (correct/total)
  - [ ] BPM tracker
  - [ ] Tempo visual indicator

- [ ] Create lesson completion screen
  - [ ] Final scores display
  - [ ] Practice summary
  - [ ] Save/retry buttons
  - [ ] Next lesson recommendation

**Acceptance Criteria**:
- ✅ Practice UI displays properly
- ✅ All metrics visible and updating
- ✅ Controls responsive and functional
- ✅ Completion screen informative

---

### Subtask 2.5: MIDI Integration (4 hours)
**Goal**: Implement MIDI playback and recording

**Tasks**:
- [ ] Add MIDI library to `go.mod`
  - [ ] `github.com/go-midi/midi/v2` or similar
  - [ ] Audio library for playback
  - [ ] Run `go get` and verify

- [ ] Create MIDI player service
  - [ ] Load MIDI files from database
  - [ ] Parse MIDI data
  - [ ] Convert to audio format
  - [ ] Support pause/resume

- [ ] Implement MIDI recording
  - [ ] Capture keyboard input as MIDI
  - [ ] Store as BLOB in database
  - [ ] Time-align recordings
  - [ ] Support overdub (optional Phase 2.5)

- [ ] Web Audio API integration
  - [ ] JavaScript MIDI.js library
  - [ ] Create player controls
  - [ ] Implement volume control
  - [ ] Add visual feedback (playing/paused)

- [ ] Testing
  - [ ] Test MIDI playback
  - [ ] Verify recording quality
  - [ ] Check timing accuracy
  - [ ] Memory usage optimization

**Acceptance Criteria**:
- ✅ MIDI files play correctly
- ✅ Recording captures input
- ✅ Playback quality acceptable
- ✅ No memory leaks

---

### Subtask 2.6: Authentication Integration (2 hours)
**Goal**: Connect Piano app to auth system

**Tasks**:
- [ ] Update handlers to check authentication
  - [ ] Require login for lessons
  - [ ] Read user_id from session
  - [ ] Validate before database operations
  - [ ] Return 401 Unauthorized if not authenticated

- [ ] Update routes to use auth middleware
  - [ ] Apply `AuthRequired()` middleware
  - [ ] Skip auth for homepage
  - [ ] Skip auth for public songs list
  - [ ] Require auth for lessons

- [ ] Update database operations
  - [ ] Associate lessons with user_id
  - [ ] Filter user's lesson history
  - [ ] Get user stats only for authenticated user

- [ ] Testing
  - [ ] Test authenticated endpoints
  - [ ] Test unauthenticated rejection
  - [ ] Verify session persistence

**Acceptance Criteria**:
- ✅ Unauthenticated users cannot practice
- ✅ User data properly isolated
- ✅ Sessions persist across requests

---

### Subtask 2.7: Data Validation & Error Handling (2 hours)
**Goal**: Add comprehensive validation and error handling

**Tasks**:
- [ ] Frontend validation
  - [ ] Required field checks
  - [ ] Input type validation
  - [ ] File size limits for MIDI
  - [ ] User feedback on errors

- [ ] Backend validation
  - [ ] Validate all API inputs
  - [ ] Check MIDI file format
  - [ ] Verify score ranges (0-100)
  - [ ] Check string lengths

- [ ] Error responses
  - [ ] Standardized error format
  - [ ] Appropriate HTTP status codes
  - [ ] User-friendly error messages
  - [ ] Logging for debugging

- [ ] Testing
  - [ ] Test invalid inputs
  - [ ] Test edge cases
  - [ ] Test malformed MIDI
  - [ ] Test concurrent operations

**Acceptance Criteria**:
- ✅ All inputs validated
- ✅ Proper error codes returned
- ✅ No crashes on bad input
- ✅ Clear error messages

---

### Subtask 2.8: User Progress Tracking (2 hours)
**Goal**: Implement progress tracking and statistics

**Tasks**:
- [ ] Create user stats endpoint
  - [ ] GET /api/users/{userId}/stats
  - [ ] Aggregate lesson data
  - [ ] Calculate averages
  - [ ] Track improvement over time

- [ ] Implement lesson history
  - [ ] Display past lessons
  - [ ] Show scores and dates
  - [ ] Allow lesson replay
  - [ ] Show progress graph

- [ ] Create achievement system
  - [ ] Track milestones (songs completed)
  - [ ] Accuracy achievements
  - [ ] Streak tracking
  - [ ] Badge system (future)

- [ ] Dashboard integration
  - [ ] Show Piano stats on dashboard
  - [ ] Recent lessons
  - [ ] Accuracy trend
  - [ ] Next recommended song

**Acceptance Criteria**:
- ✅ User stats accurately calculated
- ✅ History properly stored
- ✅ Trends visualized
- ✅ Dashboard shows Piano data

---

### Subtask 2.9: Testing & QA (3 hours)
**Goal**: Comprehensive testing of Phase 2 features

**Tasks**:
- [ ] Unit testing
  - [ ] Test all new handlers
  - [ ] Test MIDI processing
  - [ ] Test score calculations
  - [ ] Test validation logic

- [ ] Integration testing
  - [ ] End-to-end practice session
  - [ ] Database operations
  - [ ] Authentication flow
  - [ ] Error scenarios

- [ ] UI/UX testing
  - [ ] Test on multiple browsers
  - [ ] Mobile responsiveness
  - [ ] Performance metrics
  - [ ] Accessibility compliance

- [ ] Performance testing
  - [ ] Load test with multiple users
  - [ ] MIDI file handling
  - [ ] Database query performance
  - [ ] API response times

- [ ] Documentation
  - [ ] Update API documentation
  - [ ] Create user guides
  - [ ] Document database schema
  - [ ] Developer notes

**Acceptance Criteria**:
- ✅ All tests passing
- ✅ >80% code coverage
- ✅ No known bugs
- ✅ Performance acceptable

---

### Subtask 2.10: Documentation (1 hour)
**Goal**: Complete Phase 2 documentation

**Tasks**:
- [ ] Create `PHASE2_COMPLETION_SUMMARY.md`
- [ ] Update `README.md` with Piano features
- [ ] Add API documentation
- [ ] Create deployment guide
- [ ] Document known issues/TODOs

**Acceptance Criteria**:
- ✅ All features documented
- ✅ Clear deployment instructions
- ✅ API reference complete

---

## 📊 Implementation Timeline

| Week | Subtasks | Hours | Status |
|------|----------|-------|--------|
| **W1** (Feb 20-24) | 2.1, 2.2, 2.3 | 7 | Planning |
| **W2** (Feb 27-Mar 3) | 2.4, 2.5, 2.6 | 8 | Planning |
| **W3+** (Mar 4+) | 2.7, 2.8, 2.9, 2.10 | 8 | Planning |

**Total Estimated Hours**: 21 hours
**Estimated Completion**: March 10, 2024

---

## 🔧 Technical Dependencies

### Required Libraries
- ✅ `github.com/go-chi/chi/v5` - HTTP routing (already installed)
- ✅ `github.com/mattn/go-sqlite3` - SQLite (already installed)
- 🔲 MIDI library (need to evaluate options)
- 🔲 Audio processing library
- 🔲 Web Audio API compatibility layer

### Database
- ✅ SQLite schema created
- ✅ Foreign key support enabled
- ✅ BLOB storage verified
- ✅ Connection pooling configured

### Frontend
- ✅ HTML/CSS/JavaScript support
- 🔲 MIDI player library (Tone.js or midi-player-js)
- 🔲 Charting library for progress visualization
- 🔲 Audio API support

---

## 🚀 Success Criteria

### Functionality
- ✅ Users can browse available songs
- ✅ Users can start practice sessions
- ✅ MIDI playback works correctly
- ✅ Scores calculated and saved
- ✅ User progress tracked

### Quality
- ✅ All tests passing
- ✅ Code coverage >80%
- ✅ No security vulnerabilities
- ✅ Performance acceptable

### Usability
- ✅ Responsive design
- ✅ Intuitive navigation
- ✅ Clear instructions
- ✅ Helpful error messages

### Documentation
- ✅ API documented
- ✅ Database schema documented
- ✅ Deployment guide provided
- ✅ Developer notes included

---

## ⚠️ Known Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| MIDI library availability | High | Evaluate 3 options early |
| Audio processing complexity | High | Use existing web APIs |
| Browser compatibility | Medium | Test on Chrome, Firefox, Safari |
| Performance with large MIDI files | Medium | Implement streaming |
| Database migration issues | High | Test thoroughly before production |

---

## 🔄 Rollback Plan

If Phase 2 encounters critical issues:

1. **Revert to Main**: `git reset --hard origin/main`
2. **Pause Integration**: Keep foundation in `pkg/piano/`
3. **Reassess**: Identify blockers and adjust plan
4. **Restart**: Create new feature branch with lessons learned

---

## 📝 Checklist for Phase 2 Start

- [ ] Create feature branch (✅ DONE)
- [ ] Review foundation code
- [ ] Set up development environment
- [ ] Create database migration files
- [ ] Set up MIDI library evaluation
- [ ] Assign subtasks to team members
- [ ] Schedule daily sync meetings
- [ ] Create issue tracking

---

## 🎯 Next Steps

1. **Immediate (Today)**
   - [ ] Review this plan
   - [ ] Discuss with team
   - [ ] Approve timeline

2. **This Week**
   - [ ] Start Subtask 2.1 (Database)
   - [ ] Start Subtask 2.2 (Server Integration)
   - [ ] Begin Subtask 2.3 (Frontend)

3. **Next Week**
   - [ ] Complete core integration
   - [ ] Begin MIDI implementation
   - [ ] Start UI development

---

## 📞 Communication

**Daily standup**: 9:00 AM
**Weekly review**: Friday 4:00 PM
**Issues/blockers**: Slack channel #phase2-piano

---

## 📊 Progress Tracking

Subtask completion will be tracked in this branch with commits:
- `wip: Phase 2.X - <description>`
- `feat: Phase 2.X complete - <details>`

Weekly status updates in `PHASE2_STATUS_UPDATES.md`

---

## ✅ Phase 2 Entry Criteria

- [x] Feature branch created
- [x] Plan documented
- [x] Foundation code reviewed
- [x] Dependencies identified
- [x] Team aligned on approach

**Status**: ✅ READY TO BEGIN PHASE 2

---

**Branch**: `feature/phase2-piano-integration-0220`
**Created**: February 20, 2024
**Status**: ACTIVE DEVELOPMENT

