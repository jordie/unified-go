package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// BusinessMetricsRegistry tracks user engagement and app-specific business metrics
type BusinessMetricsRegistry struct {
	// User engagement metrics
	userSessionsCreated   *prometheus.CounterVec
	userSessionsCompleted *prometheus.CounterVec
	userActiveSessions    *prometheus.GaugeVec

	// XP and progression
	xpEarned          *prometheus.CounterVec
	achievementsUnlocked *prometheus.CounterVec

	// App-specific metrics
	typingTestsCompleted    prometheus.Counter
	mathProblemssSolved     *prometheus.CounterVec
	pianoSongsPlayed        *prometheus.CounterVec
	readingPassagesCompleted *prometheus.CounterVec

	registry *prometheus.Registry
	mu       sync.RWMutex
}

// NewBusinessMetricsRegistry creates and registers all business metrics
func NewBusinessMetricsRegistry() *BusinessMetricsRegistry {
	registry := prometheus.NewRegistry()

	b := &BusinessMetricsRegistry{
		registry: registry,
	}

	b.registerMetrics()
	return b
}

// registerMetrics registers all business metric collectors
func (b *BusinessMetricsRegistry) registerMetrics() {
	// User session created counter: tracks sessions by app
	b.userSessionsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gaia_user_sessions_created_total",
			Help: "Total user sessions created by app",
		},
		[]string{"app"},
	)
	b.registry.MustRegister(b.userSessionsCreated)

	// User session completed counter: tracks completed sessions by app
	b.userSessionsCompleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gaia_user_sessions_completed_total",
			Help: "Total user sessions completed by app",
		},
		[]string{"app"},
	)
	b.registry.MustRegister(b.userSessionsCompleted)

	// Active user sessions gauge: tracks concurrent sessions by app
	b.userActiveSessions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gaia_user_active_sessions",
			Help: "Number of active user sessions by app",
		},
		[]string{"app"},
	)
	b.registry.MustRegister(b.userActiveSessions)

	// XP earned counter: tracks total XP earned by app
	b.xpEarned = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gaia_xp_earned_total",
			Help: "Total XP earned by app",
		},
		[]string{"app"},
	)
	b.registry.MustRegister(b.xpEarned)

	// Achievements unlocked counter: tracks unlocked achievements by app and type
	b.achievementsUnlocked = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gaia_achievements_unlocked_total",
			Help: "Total achievements unlocked by app and achievement type",
		},
		[]string{"app", "achievement_type"},
	)
	b.registry.MustRegister(b.achievementsUnlocked)

	// Typing tests completed counter
	b.typingTestsCompleted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gaia_typing_tests_completed_total",
			Help: "Total typing tests completed",
		},
	)
	b.registry.MustRegister(b.typingTestsCompleted)

	// Math problems solved counter: tracks by difficulty level
	b.mathProblemssSolved = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gaia_math_problems_solved_total",
			Help: "Total math problems solved by difficulty level",
		},
		[]string{"difficulty"},
	)
	b.registry.MustRegister(b.mathProblemssSolved)

	// Piano songs played counter: tracks by difficulty level
	b.pianoSongsPlayed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gaia_piano_songs_played_total",
			Help: "Total piano songs played by difficulty level",
		},
		[]string{"difficulty"},
	)
	b.registry.MustRegister(b.pianoSongsPlayed)

	// Reading passages completed counter: tracks by level
	b.readingPassagesCompleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gaia_reading_passages_completed_total",
			Help: "Total reading passages completed by level",
		},
		[]string{"level"},
	)
	b.registry.MustRegister(b.readingPassagesCompleted)
}

// RecordSessionCreated records a new user session creation
func (b *BusinessMetricsRegistry) RecordSessionCreated(app string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.userSessionsCreated.WithLabelValues(app).Inc()
	b.userActiveSessions.WithLabelValues(app).Inc()
}

// RecordSessionCompleted records a completed user session
func (b *BusinessMetricsRegistry) RecordSessionCompleted(app string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.userSessionsCompleted.WithLabelValues(app).Inc()
	b.userActiveSessions.WithLabelValues(app).Dec()
}

// RecordXPEarned records XP earned by a user
func (b *BusinessMetricsRegistry) RecordXPEarned(app string, amount int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.xpEarned.WithLabelValues(app).Add(float64(amount))
}

// RecordAchievementUnlocked records an unlocked achievement
func (b *BusinessMetricsRegistry) RecordAchievementUnlocked(app, achievementType string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.achievementsUnlocked.WithLabelValues(app, achievementType).Inc()
}

// RecordTypingTestCompleted records a completed typing test
func (b *BusinessMetricsRegistry) RecordTypingTestCompleted() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.typingTestsCompleted.Inc()
}

// RecordMathProblemSolved records a solved math problem
func (b *BusinessMetricsRegistry) RecordMathProblemSolved(difficulty string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.mathProblemssSolved.WithLabelValues(difficulty).Inc()
}

// RecordPianoSongPlayed records a played piano song
func (b *BusinessMetricsRegistry) RecordPianoSongPlayed(difficulty string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pianoSongsPlayed.WithLabelValues(difficulty).Inc()
}

// RecordReadingPassageCompleted records a completed reading passage
func (b *BusinessMetricsRegistry) RecordReadingPassageCompleted(level string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.readingPassagesCompleted.WithLabelValues(level).Inc()
}

// GetPrometheusRegistry returns the underlying prometheus.Registry
func (b *BusinessMetricsRegistry) GetPrometheusRegistry() *prometheus.Registry {
	return b.registry
}
