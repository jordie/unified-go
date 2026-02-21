package dashboard

import (
	"strings"
	"testing"
)

func TestNewSubscriptionManager(t *testing.T) {
	sm := NewSubscriptionManager()

	if sm == nil {
		t.Fatal("NewSubscriptionManager returned nil")
	}
	if sm.validChannels == nil {
		t.Fatal("validChannels not initialized")
	}
	if len(sm.validChannels) == 0 {
		t.Fatal("validChannels is empty, should be initialized")
	}
}

func TestIsValidChannelExactMatch(t *testing.T) {
	sm := NewSubscriptionManager()

	tests := []struct {
		channel string
		valid   bool
	}{
		{"leaderboard:typing_wpm", true},
		{"leaderboard:math_accuracy", true},
		{"leaderboard:reading_comprehension", true},
		{"leaderboard:piano_score", true},
		{"leaderboard:overall", true},
		{"activity:feed", true},
		{"activity:achievements", true},
		{"activity:high-scores", true},
		{"system:notifications", true},
		{"system:alerts", true},
		{"invalid:channel", false},
		{"fake:data", false},
	}

	for _, tt := range tests {
		result := sm.IsValidChannel(tt.channel)
		if result != tt.valid {
			t.Errorf("IsValidChannel(%q): expected %v, got %v", tt.channel, tt.valid, result)
		}
	}
}

func TestIsValidChannelPatternMatch(t *testing.T) {
	sm := NewSubscriptionManager()

	tests := []struct {
		channel string
		valid   bool
	}{
		{"user:123:progress", true},
		{"user:456:achievements", true},
		{"user:999:rank-changes", true},
		{"user:1:high-scores", true},
		{"session:abc123:live", true},
		{"session:xyz789:competitors", true},
		// Note: "user:invalid:progress" will match "user:*:progress" pattern
		// Invalid channels are those that don't match any valid pattern
		{"user:123:invalid", false},
		{"session:123:invalid", false},
		{"invalid:123:progress", false},
	}

	for _, tt := range tests {
		result := sm.IsValidChannel(tt.channel)
		if result != tt.valid {
			t.Errorf("IsValidChannel(%q): expected %v, got %v", tt.channel, tt.valid, result)
		}
	}
}

func TestMatchesPatternWithSingleWildcard(t *testing.T) {
	sm := NewSubscriptionManager()

	tests := []struct {
		pattern string
		channel string
		matches bool
	}{
		{"user:*:progress", "user:123:progress", true},
		{"user:*:progress", "user:456:progress", true},
		{"user:*:achievements", "user:123:achievements", true},
		{"user:*:rank-changes", "user:999:rank-changes", true},
		{"session:*:live", "session:abc123:live", true},
		{"session:*:competitors", "session:xyz789:competitors", true},
		{"user:*:progress", "user:123:achievements", false},
		{"user:*:progress", "session:123:progress", false},
	}

	for _, tt := range tests {
		result := sm.matchesPattern(tt.pattern, tt.channel)
		if result != tt.matches {
			t.Errorf("matchesPattern(%q, %q): expected %v, got %v",
				tt.pattern, tt.channel, tt.matches, result)
		}
	}
}

func TestMatchesPatternExactMatch(t *testing.T) {
	sm := NewSubscriptionManager()

	pattern := "leaderboard:typing_wpm"
	channel := "leaderboard:typing_wpm"

	result := sm.matchesPattern(pattern, channel)
	if !result {
		t.Errorf("matchesPattern(%q, %q): expected true, got false", pattern, channel)
	}
}

func TestMatchesPatternNoWildcard(t *testing.T) {
	sm := NewSubscriptionManager()

	pattern := "leaderboard:typing_wpm"
	channel := "leaderboard:math_accuracy"

	result := sm.matchesPattern(pattern, channel)
	if result {
		t.Errorf("matchesPattern(%q, %q): expected false, got true", pattern, channel)
	}
}

func TestValidateChannels(t *testing.T) {
	sm := NewSubscriptionManager()

	channels := []string{
		"leaderboard:typing_wpm",
		"user:123:progress",
		"invalid:channel",
		"activity:feed",
		"fake:data",
		"session:abc123:live",
	}

	valid, invalid := sm.ValidateChannels(channels)

	if len(valid) != 4 {
		t.Errorf("expected 4 valid channels, got %d: %v", len(valid), valid)
	}
	if len(invalid) != 2 {
		t.Errorf("expected 2 invalid channels, got %d: %v", len(invalid), invalid)
	}

	// Check that valid channels are correct
	expectedValid := map[string]bool{
		"leaderboard:typing_wpm": true,
		"user:123:progress":      true,
		"activity:feed":          true,
		"session:abc123:live":    true,
	}

	for _, ch := range valid {
		if !expectedValid[ch] {
			t.Errorf("unexpected valid channel: %s", ch)
		}
	}

	// Check that invalid channels are correct
	expectedInvalid := map[string]bool{
		"invalid:channel": true,
		"fake:data":       true,
	}

	for _, ch := range invalid {
		if !expectedInvalid[ch] {
			t.Errorf("unexpected invalid channel: %s", ch)
		}
	}
}

func TestGetChannelsByType(t *testing.T) {
	sm := NewSubscriptionManager()

	tests := []struct {
		channelType string
		expected    []string
	}{
		{
			"leaderboard:",
			[]string{
				"leaderboard:typing_wpm",
				"leaderboard:math_accuracy",
				"leaderboard:reading_comprehension",
				"leaderboard:piano_score",
				"leaderboard:overall",
			},
		},
		{
			"activity:",
			[]string{
				"activity:feed",
				"activity:achievements",
				"activity:high-scores",
			},
		},
		{
			"system:",
			[]string{
				"system:notifications",
				"system:alerts",
			},
		},
	}

	for _, tt := range tests {
		result := sm.GetChannelsByType(tt.channelType)

		if len(result) != len(tt.expected) {
			t.Errorf("GetChannelsByType(%q): expected %d channels, got %d",
				tt.channelType, len(tt.expected), len(result))
			continue
		}

		// Verify all expected channels are present
		resultMap := make(map[string]bool)
		for _, ch := range result {
			resultMap[ch] = true
		}

		for _, expected := range tt.expected {
			if !resultMap[expected] {
				t.Errorf("GetChannelsByType(%q): missing channel %q", tt.channelType, expected)
			}
		}
	}
}

func TestNewLeaderboardChannel(t *testing.T) {
	tests := []struct {
		category string
		expected string
	}{
		{"typing_wpm", "leaderboard:typing_wpm"},
		{"math_accuracy", "leaderboard:math_accuracy"},
		{"reading_comprehension", "leaderboard:reading_comprehension"},
		{"piano_score", "leaderboard:piano_score"},
		{"overall", "leaderboard:overall"},
	}

	for _, tt := range tests {
		result := NewLeaderboardChannel(tt.category)
		if result != tt.expected {
			t.Errorf("NewLeaderboardChannel(%q): expected %q, got %q",
				tt.category, tt.expected, result)
		}
	}
}

func TestNewUserChannels(t *testing.T) {
	userID := uint(123)

	tests := []struct {
		name     string
		fn       func(uint) string
		expected string
	}{
		{"progress", NewUserProgressChannel, "user:123:progress"},
		{"achievement", NewUserAchievementChannel, "user:123:achievements"},
		{"rank_change", NewUserRankChangeChannel, "user:123:rank-changes"},
		{"high_score", NewUserHighScoreChannel, "user:123:high-scores"},
	}

	for _, tt := range tests {
		result := tt.fn(userID)
		if result != tt.expected {
			t.Errorf("%s: expected %q, got %q", tt.name, tt.expected, result)
		}
	}
}

func TestNewSessionChannels(t *testing.T) {
	sessionID := "abc123xyz"

	tests := []struct {
		name     string
		fn       func(string) string
		expected string
	}{
		{"live", NewSessionChannel, "session:abc123xyz:live"},
		{"competitors", NewSessionCompetitorsChannel, "session:abc123xyz:competitors"},
	}

	for _, tt := range tests {
		result := tt.fn(sessionID)
		if result != tt.expected {
			t.Errorf("%s: expected %q, got %q", tt.name, tt.expected, result)
		}
	}
}

func TestNewChannelGroupManager(t *testing.T) {
	cgm := NewChannelGroupManager()

	if cgm == nil {
		t.Fatal("NewChannelGroupManager returned nil")
	}
	if cgm.groups == nil {
		t.Fatal("groups not initialized")
	}
	if len(cgm.groups) == 0 {
		t.Fatal("groups is empty, should be initialized")
	}
}

func TestGetGroup(t *testing.T) {
	cgm := NewChannelGroupManager()

	tests := []struct {
		groupName string
		wantError bool
		expected  []string
	}{
		{
			"leaderboards",
			false,
			[]string{
				"leaderboard:typing_wpm",
				"leaderboard:math_accuracy",
				"leaderboard:reading_comprehension",
				"leaderboard:piano_score",
				"leaderboard:overall",
			},
		},
		{
			"activity",
			false,
			[]string{
				"activity:feed",
				"activity:achievements",
				"activity:high-scores",
			},
		},
		{
			"system",
			false,
			[]string{
				"system:notifications",
				"system:alerts",
			},
		},
		{
			"nonexistent",
			true,
			nil,
		},
	}

	for _, tt := range tests {
		channels, err := cgm.GetGroup(tt.groupName)

		if (err != nil) != tt.wantError {
			t.Errorf("GetGroup(%q): wantError %v, got error: %v", tt.groupName, tt.wantError, err)
			continue
		}

		if !tt.wantError {
			if len(channels) != len(tt.expected) {
				t.Errorf("GetGroup(%q): expected %d channels, got %d",
					tt.groupName, len(tt.expected), len(channels))
				continue
			}

			// Verify all expected channels are present
			channelMap := make(map[string]bool)
			for _, ch := range channels {
				channelMap[ch] = true
			}

			for _, expected := range tt.expected {
				if !channelMap[expected] {
					t.Errorf("GetGroup(%q): missing channel %q", tt.groupName, expected)
				}
			}
		}
	}
}

func TestGetAllGroups(t *testing.T) {
	cgm := NewChannelGroupManager()

	groups := cgm.GetAllGroups()

	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}

	expectedGroups := []string{"leaderboards", "activity", "system"}
	for _, groupName := range expectedGroups {
		if _, exists := groups[groupName]; !exists {
			t.Errorf("missing group: %s", groupName)
		}
	}
}

func TestDefaultSubscriptionStrategy(t *testing.T) {
	userID := uint(123)
	strategy := DefaultSubscriptionStrategy(userID)

	if strategy.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, strategy.UserID)
	}

	if len(strategy.Channels) == 0 {
		t.Error("expected channels to be populated")
	}

	// Verify all leaderboards are included
	leaderboards := []string{
		"leaderboard:typing_wpm",
		"leaderboard:math_accuracy",
		"leaderboard:reading_comprehension",
		"leaderboard:piano_score",
		"leaderboard:overall",
	}

	channelMap := make(map[string]bool)
	for _, ch := range strategy.Channels {
		channelMap[ch] = true
	}

	for _, lb := range leaderboards {
		if !channelMap[lb] {
			t.Errorf("missing leaderboard channel: %s", lb)
		}
	}

	// Verify user-specific channels
	expectedUserChannels := []string{
		NewUserProgressChannel(userID),
		NewUserAchievementChannel(userID),
		NewUserRankChangeChannel(userID),
		NewUserHighScoreChannel(userID),
	}

	for _, ch := range expectedUserChannels {
		if !channelMap[ch] {
			t.Errorf("missing user channel: %s", ch)
		}
	}
}

func TestCompetitiveSubscriptionStrategy(t *testing.T) {
	userID := uint(456)
	strategy := CompetitiveSubscriptionStrategy(userID)

	if strategy.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, strategy.UserID)
	}

	if len(strategy.Channels) == 0 {
		t.Error("expected channels to be populated")
	}

	// Verify leaderboards are included (focus on rankings)
	channelMap := make(map[string]bool)
	for _, ch := range strategy.Channels {
		channelMap[ch] = true
	}

	// Should include competitive-focused channels
	if !channelMap["leaderboard:typing_wpm"] {
		t.Error("missing leaderboard:typing_wpm")
	}
	if !channelMap["activity:high-scores"] {
		t.Error("missing activity:high-scores")
	}
	if !channelMap["activity:achievements"] {
		t.Error("missing activity:achievements")
	}

	// Should NOT include progress updates (not competitive)
	if channelMap["activity:feed"] {
		t.Error("activity:feed should not be in competitive strategy")
	}
}

func TestCasualSubscriptionStrategy(t *testing.T) {
	userID := uint(789)
	strategy := CasualSubscriptionStrategy(userID)

	if strategy.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, strategy.UserID)
	}

	if len(strategy.Channels) == 0 {
		t.Error("expected channels to be populated")
	}

	// Verify personal and motivational channels
	channelMap := make(map[string]bool)
	for _, ch := range strategy.Channels {
		channelMap[ch] = true
	}

	// Should include personal updates
	if !channelMap[NewUserProgressChannel(userID)] {
		t.Error("missing user progress channel")
	}
	if !channelMap[NewUserAchievementChannel(userID)] {
		t.Error("missing user achievement channel")
	}

	// Should include activity feed for motivation
	if !channelMap["activity:feed"] {
		t.Error("missing activity:feed")
	}

	// Should NOT include leaderboards (not casual focus)
	if channelMap["leaderboard:typing_wpm"] {
		t.Error("leaderboard should not be in casual strategy")
	}
}

// TestSubscriptionManagerConcurrency verifies thread-safe access
func TestSubscriptionManagerConcurrency(t *testing.T) {
	sm := NewSubscriptionManager()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			channel := "user:123:progress"
			_ = sm.IsValidChannel(channel)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// If no panic, test passes
}

// TestChannelGroupManagerConcurrency verifies thread-safe access
func TestChannelGroupManagerConcurrency(t *testing.T) {
	cgm := NewChannelGroupManager()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = cgm.GetAllGroups()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// If no panic, test passes
}

// BenchmarkIsValidChannel benchmarks channel validation
func BenchmarkIsValidChannel(b *testing.B) {
	sm := NewSubscriptionManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.IsValidChannel("user:123:progress")
	}
}

// BenchmarkValidateChannels benchmarks batch channel validation
func BenchmarkValidateChannels(b *testing.B) {
	sm := NewSubscriptionManager()

	channels := []string{
		"leaderboard:typing_wpm",
		"user:123:progress",
		"invalid:channel",
		"activity:feed",
		"session:abc:live",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.ValidateChannels(channels)
	}
}

// BenchmarkGetChannelsByType benchmarks type-based retrieval
func BenchmarkGetChannelsByType(b *testing.B) {
	sm := NewSubscriptionManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.GetChannelsByType("leaderboard:")
	}
}

// BenchmarkGetAllGroups benchmarks group retrieval
func BenchmarkGetAllGroups(b *testing.B) {
	cgm := NewChannelGroupManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cgm.GetAllGroups()
	}
}

// TestChannelNameFormatting verifies channel name format consistency
func TestChannelNameFormatting(t *testing.T) {
	// All channel names should follow colon-separated format
	channels := []string{
		NewLeaderboardChannel("test"),
		NewUserProgressChannel(1),
		NewUserAchievementChannel(1),
		NewUserRankChangeChannel(1),
		NewUserHighScoreChannel(1),
		NewSessionChannel("test"),
		NewSessionCompetitorsChannel("test"),
	}

	for _, ch := range channels {
		if !strings.Contains(ch, ":") {
			t.Errorf("channel %q missing colon separator", ch)
		}

		parts := strings.Split(ch, ":")
		if len(parts) < 2 {
			t.Errorf("channel %q has invalid format", ch)
		}
	}
}
