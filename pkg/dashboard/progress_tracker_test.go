package dashboard

import (
	"testing"
	"time"
)

func TestNewProgressTracker(t *testing.T) {
	pt := NewProgressTracker()

	if pt == nil {
		t.Fatal("NewProgressTracker returned nil")
	}
	if pt.activeSessions == nil {
		t.Fatal("activeSessions not initialized")
	}
	if pt.sessionMetrics == nil {
		t.Fatal("sessionMetrics not initialized")
	}
}

func TestStartSession(t *testing.T) {
	pt := NewProgressTracker()

	err := pt.StartSession("session1", 123, "typing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify session was created
	progress := pt.GetSessionProgress("session1")
	if progress == nil {
		t.Fatal("session not created")
	}
	if progress.SessionID != "session1" {
		t.Errorf("expected session1, got %s", progress.SessionID)
	}
	if progress.UserID != 123 {
		t.Errorf("expected userID 123, got %d", progress.UserID)
	}
	if progress.App != "typing" {
		t.Errorf("expected app typing, got %s", progress.App)
	}
	if !progress.IsActive {
		t.Error("expected session to be active")
	}
}

func TestStartSessionDuplicate(t *testing.T) {
	pt := NewProgressTracker()

	_ = pt.StartSession("session1", 123, "typing")
	err := pt.StartSession("session1", 123, "typing")

	if err == nil {
		t.Error("expected error when starting duplicate session")
	}
}

func TestEndSession(t *testing.T) {
	pt := NewProgressTracker()

	_ = pt.StartSession("session1", 123, "typing")
	pt.EndSession("session1")

	// Verify session is marked inactive
	progress := pt.GetSessionProgress("session1")
	if progress != nil && progress.IsActive {
		t.Error("expected session to be inactive")
	}
}

func TestRecordMetric(t *testing.T) {
	pt := NewProgressTracker()

	_ = pt.StartSession("session1", 123, "typing")

	err := pt.RecordMetric("session1", "wpm", 150.5, "150.5 WPM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify metric was recorded
	progress := pt.GetSessionProgress("session1")
	if progress.CurrentMetric != 150.5 {
		t.Errorf("expected metric 150.5, got %f", progress.CurrentMetric)
	}
}

func TestRecordMetricNotFound(t *testing.T) {
	pt := NewProgressTracker()

	err := pt.RecordMetric("nonexistent", "wpm", 150.0, "test")
	if err == nil {
		t.Error("expected error for non-existent session")
	}
}

func TestRecordMultipleMetrics(t *testing.T) {
	pt := NewProgressTracker()

	_ = pt.StartSession("session1", 123, "typing")

	_ = pt.RecordMetric("session1", "wpm", 100.0, "100 WPM")
	time.Sleep(5 * time.Millisecond)
	_ = pt.RecordMetric("session1", "accuracy", 95.0, "95%")
	time.Sleep(5 * time.Millisecond)
	_ = pt.RecordMetric("session1", "wpm", 120.0, "120 WPM")

	// Verify metrics were recorded
	history := pt.GetMetricHistory("session1", 0)
	if len(history) != 3 {
		t.Errorf("expected 3 metrics, got %d", len(history))
	}

	// Verify metrics have correct values
	if history[0].Value != 100.0 {
		t.Errorf("expected first metric 100.0, got %f", history[0].Value)
	}
	if history[1].Value != 95.0 {
		t.Errorf("expected second metric 95.0, got %f", history[1].Value)
	}
	if history[2].Value != 120.0 {
		t.Errorf("expected third metric 120.0, got %f", history[2].Value)
	}
}

func TestTrackerGetSessionProgress(t *testing.T) {
	pt := NewProgressTracker()

	// Non-existent session
	progress := pt.GetSessionProgress("nonexistent")
	if progress != nil {
		t.Error("expected nil for non-existent session")
	}

	// Existing session
	_ = pt.StartSession("session1", 123, "typing")
	progress = pt.GetSessionProgress("session1")
	if progress == nil {
		t.Fatal("expected session to exist")
	}
}

func TestGetMetricHistory(t *testing.T) {
	pt := NewProgressTracker()

	_ = pt.StartSession("session1", 123, "typing")

	for i := 1; i <= 5; i++ {
		_ = pt.RecordMetric("session1", "wpm", float64(100*i), "")
		time.Sleep(2 * time.Millisecond)
	}

	// Get full history
	history := pt.GetMetricHistory("session1", 0)
	if len(history) != 5 {
		t.Errorf("expected 5 metrics, got %d", len(history))
	}

	// Get limited history
	limited := pt.GetMetricHistory("session1", 2)
	if len(limited) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(limited))
	}
}

func TestTrackerGetActiveSessions(t *testing.T) {
	pt := NewProgressTracker()

	// No active sessions
	sessions := pt.GetActiveSessions()
	if len(sessions) != 0 {
		t.Errorf("expected 0 active sessions, got %d", len(sessions))
	}

	// Start multiple sessions
	_ = pt.StartSession("session1", 1, "typing")
	_ = pt.StartSession("session2", 2, "math")

	sessions = pt.GetActiveSessions()
	if len(sessions) != 2 {
		t.Errorf("expected 2 active sessions, got %d", len(sessions))
	}

	// End one session
	pt.EndSession("session1")

	sessions = pt.GetActiveSessions()
	if len(sessions) != 1 {
		t.Errorf("expected 1 active session after ending one, got %d", len(sessions))
	}
}

func TestNewMetricsAnalyzer(t *testing.T) {
	pt := NewProgressTracker()
	analyzer := NewMetricsAnalyzer(pt)

	if analyzer == nil {
		t.Fatal("NewMetricsAnalyzer returned nil")
	}
}

func TestAnalyzeMetric(t *testing.T) {
	pt := NewProgressTracker()
	analyzer := NewMetricsAnalyzer(pt)

	_ = pt.StartSession("session1", 123, "typing")

	// Record metrics
	_ = pt.RecordMetric("session1", "wpm", 100.0, "100 WPM")
	time.Sleep(5 * time.Millisecond)
	_ = pt.RecordMetric("session1", "wpm", 120.0, "120 WPM")
	time.Sleep(5 * time.Millisecond)
	_ = pt.RecordMetric("session1", "wpm", 115.0, "115 WPM")

	// Analyze metric
	stats := analyzer.AnalyzeMetric("session1", "wpm")

	if stats == nil {
		t.Fatal("AnalyzeMetric returned nil")
	}
	if stats.Current != 115.0 {
		t.Errorf("expected current 115.0, got %f", stats.Current)
	}
	if stats.Min != 100.0 {
		t.Errorf("expected min 100.0, got %f", stats.Min)
	}
	if stats.Max != 120.0 {
		t.Errorf("expected max 120.0, got %f", stats.Max)
	}
	if stats.SampleSize != 3 {
		t.Errorf("expected 3 samples, got %d", stats.SampleSize)
	}
}

func TestAnalyzeMetricNoData(t *testing.T) {
	pt := NewProgressTracker()
	analyzer := NewMetricsAnalyzer(pt)

	_ = pt.StartSession("session1", 123, "typing")

	// Analyze metric without recording any
	stats := analyzer.AnalyzeMetric("session1", "wpm")

	if stats != nil {
		t.Error("expected nil for empty metrics")
	}
}

func TestAnalyzeAllMetrics(t *testing.T) {
	pt := NewProgressTracker()
	analyzer := NewMetricsAnalyzer(pt)

	_ = pt.StartSession("session1", 123, "typing")

	// Record different metric types
	_ = pt.RecordMetric("session1", "wpm", 100.0, "100 WPM")
	time.Sleep(2 * time.Millisecond)
	_ = pt.RecordMetric("session1", "accuracy", 95.0, "95%")
	time.Sleep(2 * time.Millisecond)
	_ = pt.RecordMetric("session1", "wpm", 120.0, "120 WPM")

	// Analyze all metrics
	allStats := analyzer.AnalyzeAllMetrics("session1")

	if len(allStats) != 2 {
		t.Errorf("expected 2 metric types, got %d", len(allStats))
	}

	if wpmStats, exists := allStats["wpm"]; exists {
		if wpmStats.SampleSize != 2 {
			t.Errorf("expected 2 WPM samples, got %d", wpmStats.SampleSize)
		}
	} else {
		t.Error("WPM stats not found")
	}

	if accuStats, exists := allStats["accuracy"]; exists {
		if accuStats.SampleSize != 1 {
			t.Errorf("expected 1 accuracy sample, got %d", accuStats.SampleSize)
		}
	} else {
		t.Error("accuracy stats not found")
	}
}

func TestCompareWithPrevious(t *testing.T) {
	pt := NewProgressTracker()
	analyzer := NewMetricsAnalyzer(pt)

	// First session
	_ = pt.StartSession("session1", 123, "typing")
	_ = pt.RecordMetric("session1", "wpm", 100.0, "100 WPM")

	// Second session
	_ = pt.StartSession("session2", 123, "typing")
	_ = pt.RecordMetric("session2", "wpm", 120.0, "120 WPM")

	// Compare
	comparison := analyzer.CompareWithPrevious("session2", "session1", "wpm")

	if comparison == nil {
		t.Fatal("CompareWithPrevious returned nil")
	}
	if !comparison.IsBetter {
		t.Error("expected comparison to show improvement")
	}
	if comparison.Improvement <= 0 {
		t.Error("expected positive improvement")
	}
}

func TestSummarizeSession(t *testing.T) {
	pt := NewProgressTracker()
	analyzer := NewMetricsAnalyzer(pt)

	_ = pt.StartSession("session1", 123, "typing")
	_ = pt.RecordMetric("session1", "wpm", 100.0, "100 WPM")
	time.Sleep(5 * time.Millisecond)
	_ = pt.RecordMetric("session1", "accuracy", 95.0, "95%")

	pt.EndSession("session1")

	summary := analyzer.SummarizeSession("session1")

	if summary == nil {
		t.Fatal("SummarizeSession returned nil")
	}
	if summary.SessionID != "session1" {
		t.Errorf("expected session1, got %s", summary.SessionID)
	}
	if summary.UserID != 123 {
		t.Errorf("expected userID 123, got %d", summary.UserID)
	}
	if summary.App != "typing" {
		t.Errorf("expected app typing, got %s", summary.App)
	}
}

// BenchmarkRecordMetric benchmarks metric recording
func BenchmarkRecordMetric(b *testing.B) {
	pt := NewProgressTracker()
	_ = pt.StartSession("session1", 123, "typing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pt.RecordMetric("session1", "wpm", float64(100+i%50), "")
	}
}

// BenchmarkAnalyzeMetric benchmarks metric analysis
func BenchmarkAnalyzeMetric(b *testing.B) {
	pt := NewProgressTracker()
	analyzer := NewMetricsAnalyzer(pt)

	_ = pt.StartSession("session1", 123, "typing")
	for i := 0; i < 100; i++ {
		_ = pt.RecordMetric("session1", "wpm", float64(100+i), "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.AnalyzeMetric("session1", "wpm")
	}
}

// TestProgressTrackerConcurrency verifies thread-safe access
func TestProgressTrackerConcurrency(t *testing.T) {
	pt := NewProgressTracker()

	for i := 0; i < 5; i++ {
		_ = pt.StartSession("session_"+string(rune(i)), uint(i), "typing")
	}

	done := make(chan bool)
	for _, sessionID := range []string{"session_0", "session_1", "session_2", "session_3", "session_4"} {
		go func(sid string) {
			for j := 0; j < 20; j++ {
				_ = pt.RecordMetric(sid, "wpm", float64(100+j), "")
				_ = pt.GetSessionProgress(sid)
			}
			done <- true
		}(sessionID)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	// If no panic, test passes
}
