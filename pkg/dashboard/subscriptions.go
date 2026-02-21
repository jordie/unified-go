package dashboard

import (
	"fmt"
	"strings"
	"sync"
)

// SubscriptionManager manages channel subscriptions and validation
type SubscriptionManager struct {
	validChannels map[string]bool
	mu            sync.RWMutex
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager() *SubscriptionManager {
	sm := &SubscriptionManager{
		validChannels: make(map[string]bool),
	}
	sm.initializeValidChannels()
	return sm
}

// initializeValidChannels initializes the list of valid channels
func (sm *SubscriptionManager) initializeValidChannels() {
	validChannels := []string{
		// Leaderboard channels
		"leaderboard:typing_wpm",
		"leaderboard:math_accuracy",
		"leaderboard:reading_comprehension",
		"leaderboard:piano_score",
		"leaderboard:overall",

		// User-specific channels (with {userID} placeholder)
		"user:*:progress",
		"user:*:achievements",
		"user:*:rank-changes",
		"user:*:high-scores",

		// Activity channels
		"activity:feed",
		"activity:achievements",
		"activity:high-scores",

		// Session channels (with {sessionID} placeholder)
		"session:*:live",
		"session:*:competitors",

		// System channels
		"system:notifications",
		"system:alerts",
	}

	for _, channel := range validChannels {
		sm.validChannels[channel] = true
	}
}

// IsValidChannel checks if a channel name is valid
func (sm *SubscriptionManager) IsValidChannel(channel string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check exact match
	if sm.validChannels[channel] {
		return true
	}

	// Check pattern match (e.g., user:123:progress matches user:*/progress)
	for validChannel := range sm.validChannels {
		if sm.matchesPattern(validChannel, channel) {
			return true
		}
	}

	return false
}

// matchesPattern checks if a channel matches a pattern
func (sm *SubscriptionManager) matchesPattern(pattern, channel string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == channel
	}

	// Simple pattern matching: replace * with regex-like matching
	parts := strings.Split(pattern, "*")
	if len(parts) != 2 {
		return false
	}

	return strings.HasPrefix(channel, parts[0]) && strings.HasSuffix(channel, parts[1])
}

// ValidateChannels validates a list of channels
func (sm *SubscriptionManager) ValidateChannels(channels []string) ([]string, []string) {
	valid := []string{}
	invalid := []string{}

	for _, channel := range channels {
		if sm.IsValidChannel(channel) {
			valid = append(valid, channel)
		} else {
			invalid = append(invalid, channel)
		}
	}

	return valid, invalid
}

// GetChannelsByType returns all channels of a specific type
func (sm *SubscriptionManager) GetChannelsByType(channelType string) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	channels := []string{}
	for channel := range sm.validChannels {
		if strings.HasPrefix(channel, channelType) {
			channels = append(channels, channel)
		}
	}
	return channels
}

// ChannelBuilder helps build channel names dynamically
type ChannelBuilder struct {
	baseChannel string
}

// NewLeaderboardChannel creates a leaderboard channel name
func NewLeaderboardChannel(category string) string {
	return fmt.Sprintf("leaderboard:%s", category)
}

// NewUserProgressChannel creates a user progress channel name
func NewUserProgressChannel(userID uint) string {
	return fmt.Sprintf("user:%d:progress", userID)
}

// NewUserAchievementChannel creates a user achievement channel name
func NewUserAchievementChannel(userID uint) string {
	return fmt.Sprintf("user:%d:achievements", userID)
}

// NewUserRankChangeChannel creates a user rank change channel name
func NewUserRankChangeChannel(userID uint) string {
	return fmt.Sprintf("user:%d:rank-changes", userID)
}

// NewUserHighScoreChannel creates a user high score channel name
func NewUserHighScoreChannel(userID uint) string {
	return fmt.Sprintf("user:%d:high-scores", userID)
}

// NewSessionChannel creates a session channel name
func NewSessionChannel(sessionID string) string {
	return fmt.Sprintf("session:%s:live", sessionID)
}

// NewSessionCompetitorsChannel creates a session competitors channel name
func NewSessionCompetitorsChannel(sessionID string) string {
	return fmt.Sprintf("session:%s:competitors", sessionID)
}

// ChannelGroup represents a group of channels
type ChannelGroup struct {
	name     string
	channels []string
}

// ChannelGroupManager manages predefined channel groups
type ChannelGroupManager struct {
	groups map[string][]string
	mu     sync.RWMutex
}

// NewChannelGroupManager creates a new channel group manager
func NewChannelGroupManager() *ChannelGroupManager {
	cgm := &ChannelGroupManager{
		groups: make(map[string][]string),
	}
	cgm.initializeGroups()
	return cgm
}

// initializeGroups initializes predefined channel groups
func (cgm *ChannelGroupManager) initializeGroups() {
	cgm.groups["leaderboards"] = []string{
		"leaderboard:typing_wpm",
		"leaderboard:math_accuracy",
		"leaderboard:reading_comprehension",
		"leaderboard:piano_score",
		"leaderboard:overall",
	}

	cgm.groups["activity"] = []string{
		"activity:feed",
		"activity:achievements",
		"activity:high-scores",
	}

	cgm.groups["system"] = []string{
		"system:notifications",
		"system:alerts",
	}
}

// GetGroup returns a channel group by name
func (cgm *ChannelGroupManager) GetGroup(groupName string) ([]string, error) {
	cgm.mu.RLock()
	defer cgm.mu.RUnlock()

	if channels, ok := cgm.groups[groupName]; ok {
		return channels, nil
	}

	return nil, fmt.Errorf("group %s not found", groupName)
}

// GetAllGroups returns all channel groups
func (cgm *ChannelGroupManager) GetAllGroups() map[string][]string {
	cgm.mu.RLock()
	defer cgm.mu.RUnlock()

	groups := make(map[string][]string)
	for name, channels := range cgm.groups {
		groupCopy := make([]string, len(channels))
		copy(groupCopy, channels)
		groups[name] = groupCopy
	}
	return groups
}

// SubscriptionStrategy represents a subscription strategy for a user
type SubscriptionStrategy struct {
	UserID   uint
	Channels []string
}

// DefaultSubscriptionStrategy returns the default channels for a user
func DefaultSubscriptionStrategy(userID uint) *SubscriptionStrategy {
	return &SubscriptionStrategy{
		UserID: userID,
		Channels: []string{
			// Subscribe to all leaderboards
			"leaderboard:typing_wpm",
			"leaderboard:math_accuracy",
			"leaderboard:reading_comprehension",
			"leaderboard:piano_score",
			"leaderboard:overall",

			// Subscribe to user-specific channels
			NewUserProgressChannel(userID),
			NewUserAchievementChannel(userID),
			NewUserRankChangeChannel(userID),
			NewUserHighScoreChannel(userID),

			// Subscribe to activity and system
			"activity:feed",
			"system:notifications",
		},
	}
}

// CompetitiveSubscriptionStrategy returns channels for a competitive user
func CompetitiveSubscriptionStrategy(userID uint) *SubscriptionStrategy {
	return &SubscriptionStrategy{
		UserID: userID,
		Channels: []string{
			// Focus on leaderboards
			"leaderboard:typing_wpm",
			"leaderboard:math_accuracy",
			"leaderboard:reading_comprehension",
			"leaderboard:piano_score",
			"leaderboard:overall",

			// Personal rank tracking
			NewUserRankChangeChannel(userID),
			NewUserHighScoreChannel(userID),

			// Competitive activity
			"activity:high-scores",
			"activity:achievements",
		},
	}
}

// CasualSubscriptionStrategy returns channels for a casual user
func CasualSubscriptionStrategy(userID uint) *SubscriptionStrategy {
	return &SubscriptionStrategy{
		UserID: userID,
		Channels: []string{
			// Personal updates only
			NewUserProgressChannel(userID),
			NewUserAchievementChannel(userID),

			// Activity for motivation
			"activity:feed",
			"system:notifications",
		},
	}
}
