package metrics

import (
	"testing"
)

func TestNewBusinessMetricsRegistry(t *testing.T) {
	registry := NewBusinessMetricsRegistry()

	if registry == nil {
		t.Errorf("Expected non-nil BusinessMetricsRegistry, got nil")
	}

	if registry.GetPrometheusRegistry() == nil {
		t.Errorf("Expected non-nil Prometheus registry, got nil")
	}
}

func TestSessionMetrics(t *testing.T) {
	registry := NewBusinessMetricsRegistry()

	apps := []string{"math", "typing", "reading", "piano"}

	for _, app := range apps {
		// Create sessions
		registry.RecordSessionCreated(app)
		registry.RecordSessionCreated(app)

		// Complete sessions
		registry.RecordSessionCompleted(app)
	}
}

func TestXPMetrics(t *testing.T) {
	registry := NewBusinessMetricsRegistry()

	apps := []string{"math", "typing", "reading", "piano"}
	xpAmounts := []int{10, 50, 100, 500}

	for _, app := range apps {
		for _, xp := range xpAmounts {
			registry.RecordXPEarned(app, xp)
		}
	}
}

func TestAchievementMetrics(t *testing.T) {
	registry := NewBusinessMetricsRegistry()

	apps := []string{"math", "typing", "reading", "piano"}
	achievements := []string{"beginner", "intermediate", "advanced", "master"}

	for _, app := range apps {
		for _, achievement := range achievements {
			registry.RecordAchievementUnlocked(app, achievement)
		}
	}
}

func TestTypingMetrics(t *testing.T) {
	registry := NewBusinessMetricsRegistry()

	for i := 0; i < 100; i++ {
		registry.RecordTypingTestCompleted()
	}
}

func TestMathMetrics(t *testing.T) {
	registry := NewBusinessMetricsRegistry()

	difficulties := []string{"easy", "medium", "hard", "expert"}

	for _, diff := range difficulties {
		for i := 0; i < 50; i++ {
			registry.RecordMathProblemSolved(diff)
		}
	}
}

func TestPianoMetrics(t *testing.T) {
	registry := NewBusinessMetricsRegistry()

	difficulties := []string{"beginner", "intermediate", "advanced"}

	for _, diff := range difficulties {
		for i := 0; i < 30; i++ {
			registry.RecordPianoSongPlayed(diff)
		}
	}
}

func TestReadingMetrics(t *testing.T) {
	registry := NewBusinessMetricsRegistry()

	levels := []string{"level1", "level2", "level3", "level4", "level5"}

	for _, level := range levels {
		for i := 0; i < 20; i++ {
			registry.RecordReadingPassageCompleted(level)
		}
	}
}

func TestBusinessMetricsThreadSafety(t *testing.T) {
	registry := NewBusinessMetricsRegistry()

	done := make(chan bool)

	// Concurrent metric recording
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				registry.RecordSessionCreated("math")
				registry.RecordXPEarned("typing", 10)
				registry.RecordTypingTestCompleted()
				registry.RecordMathProblemSolved("easy")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSessionLifecycle(t *testing.T) {
	registry := NewBusinessMetricsRegistry()

	// Create multiple sessions
	for i := 0; i < 5; i++ {
		registry.RecordSessionCreated("math")
	}

	// Complete some
	for i := 0; i < 3; i++ {
		registry.RecordSessionCompleted("math")
	}

	// Create more
	for i := 0; i < 2; i++ {
		registry.RecordSessionCreated("typing")
	}

	// Complete
	registry.RecordSessionCompleted("typing")
}
