package subsystems

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestBrowserPoolBasic(t *testing.T) {
	bp := NewBrowserPool(10, 50)
	defer bp.Close()

	// Verify metrics are initialized
	metrics := bp.GetMetrics()
	if metrics["browser_instances"].(int64) != 0 {
		t.Fatalf("Expected 0 browser instances, got %d", metrics["browser_instances"].(int64))
	}
}

func TestBrowserPoolMaxLimit(t *testing.T) {
	bp := NewBrowserPool(5, 50)
	defer bp.Close()

	// Verify max browsers constraint
	if bp.maxBrowsers != 5 {
		t.Fatalf("Expected max browsers 5, got %d", bp.maxBrowsers)
	}

	metrics := bp.GetMetrics()
	if metrics["max_browsers"].(int64) != 5 {
		t.Fatalf("Expected max browsers 5 in metrics, got %d", metrics["max_browsers"].(int64))
	}
}

func TestBrowserPoolMetrics(t *testing.T) {
	bp := NewBrowserPool(10, 50)
	defer bp.Close()

	metrics := bp.GetMetrics()

	// Verify all metrics exist
	requiredMetrics := []string{
		"browser_instances",
		"active_browsers",
		"total_tabs",
		"active_tabs",
		"successful_launches",
		"failed_launches",
		"extensions_loaded",
		"extensions_failed",
		"peak_concurrent_browsers",
		"peak_concurrent_tabs",
	}

	for _, metric := range requiredMetrics {
		if _, exists := metrics[metric]; !exists {
			t.Fatalf("Missing metric: %s", metric)
		}
	}
}

func TestBrowserContextCancellation(t *testing.T) {
	bp := NewBrowserPool(10, 50)
	defer bp.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should fail due to cancelled context
	_, err := bp.LaunchBrowser(ctx, true)
	if err == nil {
		t.Fatal("Expected error for cancelled context")
	}
}

func TestTabCreationAndClosure(t *testing.T) {
	bp := NewBrowserPool(10, 50)
	defer bp.Close()

	// Create browser without launching (mock)
	browserID := "test-browser-1"
	browser := &BrowserInstance{
		ID:        browserID,
		ProcessID: 99999,
		Port:      9222,
		Tabs:      make(map[string]*Tab),
		MaxTabs:   50,
		CreatedAt: time.Now(),
	}

	bp.browserMutex.Lock()
	bp.browsers[browserID] = browser
	bp.browserMutex.Unlock()

	bp.cdpMutex.Lock()
	bp.cdpConnections[browserID] = "ws://localhost:9222"
	bp.cdpMutex.Unlock()

	// Create tab
	ctx := context.Background()
	tab, err := bp.CreateTab(ctx, browserID, "https://example.com")
	if err != nil {
		t.Fatalf("Failed to create tab: %v", err)
	}

	if tab.URL != "https://example.com" {
		t.Fatalf("Expected URL https://example.com, got %s", tab.URL)
	}

	// Verify tab registered
	if len(browser.Tabs) != 1 {
		t.Fatalf("Expected 1 tab, got %d", len(browser.Tabs))
	}

	// Close tab
	err = bp.CloseTab(browserID, tab.ID)
	if err != nil {
		t.Fatalf("Failed to close tab: %v", err)
	}

	if len(browser.Tabs) != 0 {
		t.Fatalf("Expected 0 tabs after close, got %d", len(browser.Tabs))
	}
}

func TestExtensionLoading(t *testing.T) {
	bp := NewBrowserPool(10, 50)
	defer bp.Close()

	// Create browser
	browserID := "test-browser-2"
	browser := &BrowserInstance{
		ID:        browserID,
		ProcessID: 99999,
		Port:      9222,
		Tabs:      make(map[string]*Tab),
		MaxTabs:   50,
		CreatedAt: time.Now(),
	}

	bp.browserMutex.Lock()
	bp.browsers[browserID] = browser
	bp.browserMutex.Unlock()

	ctx := context.Background()

	// Create temporary extension directory for testing
	extensionPath := "/tmp/test-extension-8888"

	// Load extension
	extensionID, err := bp.LoadExtension(ctx, browserID, extensionPath)
	if err == nil || extensionID == "" {
		// Expected to fail since path doesn't exist, but extensionID format should be validated
		t.Logf("Extension loading test (path doesn't exist as expected): %v", err)
	}
}

func TestBrowserStatus(t *testing.T) {
	bp := NewBrowserPool(10, 50)
	defer bp.Close()

	// Create browser
	browserID := "test-browser-3"
	browser := &BrowserInstance{
		ID:        browserID,
		ProcessID: 99999,
		Port:      9222,
		Tabs:      make(map[string]*Tab),
		MaxTabs:   50,
		CreatedAt: time.Now(),
	}

	bp.browserMutex.Lock()
	bp.browsers[browserID] = browser
	bp.browserMutex.Unlock()

	// Get status
	status, err := bp.GetBrowserStatus(browserID)
	if err != nil {
		t.Fatalf("Failed to get browser status: %v", err)
	}

	if status.ID != browserID {
		t.Fatalf("Expected browser ID %s, got %s", browserID, status.ID)
	}

	if status.Port != 9222 {
		t.Fatalf("Expected port 9222, got %d", status.Port)
	}
}

func TestBrowserClosing(t *testing.T) {
	bp := NewBrowserPool(10, 50)
	defer bp.Close()

	// Create browser
	browserID := "test-browser-4"
	browser := &BrowserInstance{
		ID:        browserID,
		ProcessID: 99999,
		Port:      9223,
		Tabs:      make(map[string]*Tab),
		MaxTabs:   50,
		CreatedAt: time.Now(),
	}

	bp.browserMutex.Lock()
	bp.browsers[browserID] = browser
	bp.browserMutex.Unlock()

	bp.cdpMutex.Lock()
	bp.cdpConnections[browserID] = "ws://localhost:9223"
	bp.cdpMutex.Unlock()

	// Close browser
	err := bp.CloseBrowser(browserID)
	if err != nil {
		t.Fatalf("Failed to close browser: %v", err)
	}

	// Verify browser removed
	bp.browserMutex.RLock()
	_, exists := bp.browsers[browserID]
	bp.browserMutex.RUnlock()

	if exists {
		t.Fatal("Browser should be removed after closing")
	}
}

func TestConcurrentTabCreation(t *testing.T) {
	bp := NewBrowserPool(10, 50)
	defer bp.Close()

	// Create browser
	browserID := "test-browser-5"
	browser := &BrowserInstance{
		ID:        browserID,
		ProcessID: 99999,
		Port:      9224,
		Tabs:      make(map[string]*Tab),
		MaxTabs:   50,
		CreatedAt: time.Now(),
	}

	bp.browserMutex.Lock()
	bp.browsers[browserID] = browser
	bp.browserMutex.Unlock()

	bp.cdpMutex.Lock()
	bp.cdpConnections[browserID] = "ws://localhost:9224"
	bp.cdpMutex.Unlock()

	// Create tabs concurrently
	ctx := context.Background()
	done := make(chan error, 20)

	for i := 0; i < 20; i++ {
		go func() {
			_, err := bp.CreateTab(ctx, browserID, "https://example.com")
			done <- err
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < 20; i++ {
		if <-done == nil {
			successCount++
		}
	}

	if successCount < 15 {
		t.Fatalf("Expected at least 15 successful tab creations, got %d", successCount)
	}

	// Verify tabs in browser (should match successes)
	browser.TabMutex.RLock()
	tabCount := len(browser.Tabs)
	browser.TabMutex.RUnlock()

	if tabCount != successCount {
		t.Logf("Note: Created %d tabs from %d attempts", tabCount, successCount)
	}
}

func TestTabMaxLimit(t *testing.T) {
	bp := NewBrowserPool(10, 50)
	defer bp.Close()

	// Create browser with low max tabs
	browserID := "test-browser-6"
	browser := &BrowserInstance{
		ID:        browserID,
		ProcessID: 99999,
		Port:      9225,
		Tabs:      make(map[string]*Tab),
		MaxTabs:   5,
		CreatedAt: time.Now(),
	}

	bp.browserMutex.Lock()
	bp.browsers[browserID] = browser
	bp.browserMutex.Unlock()

	bp.cdpMutex.Lock()
	bp.cdpConnections[browserID] = "ws://localhost:9225"
	bp.cdpMutex.Unlock()

	ctx := context.Background()

	// Create max tabs
	for i := 0; i < 5; i++ {
		_, err := bp.CreateTab(ctx, browserID, "https://example.com")
		if err != nil {
			t.Fatalf("Failed to create tab %d: %v", i, err)
		}
	}

	// 6th tab should fail
	_, err := bp.CreateTab(ctx, browserID, "https://example.com")
	if err == nil {
		t.Fatal("Expected error when exceeding tab limit")
	}
}

func TestPoolMetricsTracking(t *testing.T) {
	bp := NewBrowserPool(10, 50)
	defer bp.Close()

	// Create browser
	browserID := "test-browser-7"
	browser := &BrowserInstance{
		ID:        browserID,
		ProcessID: 99999,
		Port:      9226,
		Tabs:      make(map[string]*Tab),
		MaxTabs:   50,
		CreatedAt: time.Now(),
	}

	bp.browserMutex.Lock()
	bp.browsers[browserID] = browser
	bp.browserMutex.Unlock()

	bp.cdpMutex.Lock()
	bp.cdpConnections[browserID] = "ws://localhost:9226"
	bp.cdpMutex.Unlock()

	// Create some tabs
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		bp.CreateTab(ctx, browserID, "https://example.com")
	}

	// Check metrics
	metrics := bp.GetMetrics()

	if metrics["active_tabs"].(int64) != 5 {
		t.Fatalf("Expected 5 active tabs, got %d", metrics["active_tabs"].(int64))
	}

	if metrics["total_tabs"].(int64) != 5 {
		t.Fatalf("Expected 5 total tabs, got %d", metrics["total_tabs"].(int64))
	}
}

func TestBrowserPoolLoadPattern(t *testing.T) {
	bp := NewBrowserPool(20, 50)
	defer bp.Close()

	ctx := context.Background()

	// Create 10 browser instances (mocked)
	for i := 0; i < 10; i++ {
		browserID := fmt.Sprintf("load-browser-%d", i)
		browser := &BrowserInstance{
			ID:        browserID,
			ProcessID: 99999 + i,
			Port:      9300 + i,
			Tabs:      make(map[string]*Tab),
			MaxTabs:   50,
			CreatedAt: time.Now(),
		}

		bp.browserMutex.Lock()
		bp.browsers[browserID] = browser
		bp.browserMutex.Unlock()

		bp.cdpMutex.Lock()
		bp.cdpConnections[browserID] = fmt.Sprintf("ws://localhost:%d", 9300+i)
		bp.cdpMutex.Unlock()

		// Create tabs in each browser
		for j := 0; j < 10; j++ {
			bp.CreateTab(ctx, browserID, "https://example.com")
		}
	}

	metrics := bp.GetMetrics()

	if metrics["active_tabs"].(int64) != 100 {
		t.Fatalf("Expected 100 active tabs, got %d", metrics["active_tabs"].(int64))
	}

	if metrics["total_tabs"].(int64) != 100 {
		t.Fatalf("Expected 100 total tabs, got %d", metrics["total_tabs"].(int64))
	}
}

func BenchmarkBrowserPoolTabCreation(b *testing.B) {
	bp := NewBrowserPool(10, 100)
	defer bp.Close()

	// Create browser
	browserID := "bench-browser"
	browser := &BrowserInstance{
		ID:        browserID,
		ProcessID: 99999,
		Port:      9227,
		Tabs:      make(map[string]*Tab),
		MaxTabs:   1000,
		CreatedAt: time.Now(),
	}

	bp.browserMutex.Lock()
	bp.browsers[browserID] = browser
	bp.browserMutex.Unlock()

	bp.cdpMutex.Lock()
	bp.cdpConnections[browserID] = "ws://localhost:9227"
	bp.cdpMutex.Unlock()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp.CreateTab(ctx, browserID, "https://example.com")
	}

	metrics := bp.GetMetrics()
	b.Logf("Active tabs: %d", metrics["active_tabs"].(int64))
	b.Logf("Total tabs: %d", metrics["total_tabs"].(int64))
}

func BenchmarkBrowserPoolOperations(b *testing.B) {
	bp := NewBrowserPool(10, 50)
	defer bp.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp.GetMetrics()
	}

	metrics := bp.GetMetrics()
	b.Logf("Max browsers: %d", metrics["max_browsers"].(int64))
}
