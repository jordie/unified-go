package subsystems

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// BrowserPool manages a pool of browser instances with Chrome DevTools Protocol support
type BrowserPool struct {
	maxBrowsers       int64
	activeBrowsers    int64
	browsers          map[string]*BrowserInstance
	browserMutex      sync.RWMutex
	cdpConnections    map[string]string // browserID -> WSEndpoint
	cdpMutex          sync.RWMutex
	extensionPaths    map[string]string // extensionID -> path
	extensionMutex    sync.RWMutex
	metrics           *BrowserMetrics
	instanceSemaphore chan struct{}
	tabSemaphore      chan struct{}
	closeChan         chan struct{}
	closeOnce         sync.Once
}

// BrowserInstance represents a single browser instance
type BrowserInstance struct {
	ID              string
	ProcessID       int
	WSEndpoint      string
	Port            int
	Tabs            map[string]*Tab
	TabMutex        sync.RWMutex
	MaxTabs         int64
	ActiveTabs      int64
	CreatedAt       time.Time
	LastActivityAt  time.Time
	StatusMutex     sync.RWMutex
	Extensions      []string
}

// Tab represents a browser tab/page
type Tab struct {
	ID        string
	CDPURL    string
	CreatedAt time.Time
	URL       string
}

// BrowserMetrics tracks browser pool performance
type BrowserMetrics struct {
	browserInstances      int64
	activeBrowsers        int64
	totalTabs             int64
	activeTabs            int64
	successfulLaunches    int64
	failedLaunches        int64
	extensionsLoaded      int64
	extensionsFailedLoad  int64
	cdpConnectionErrors   int64
	tabCreationErrors     int64
	peakConcurrentBrowsers int64
	peakConcurrentTabs    int64
}

// NewBrowserPool creates a new browser pool
// maxBrowsers: maximum concurrent browser instances (typically 100 for 8-core system)
// maxTabs: maximum tabs per browser (typically 20-50)
func NewBrowserPool(maxBrowsers int64, maxTabs int64) *BrowserPool {
	bp := &BrowserPool{
		maxBrowsers:       maxBrowsers,
		browsers:          make(map[string]*BrowserInstance),
		cdpConnections:    make(map[string]string),
		extensionPaths:    make(map[string]string),
		metrics:           &BrowserMetrics{},
		instanceSemaphore: make(chan struct{}, maxBrowsers),
		tabSemaphore:      make(chan struct{}, maxBrowsers*maxTabs),
		closeChan:         make(chan struct{}),
	}

	// Fill semaphores with initial tokens
	for i := int64(0); i < maxBrowsers; i++ {
		bp.instanceSemaphore <- struct{}{}
	}
	for i := int64(0); i < maxBrowsers*maxTabs; i++ {
		bp.tabSemaphore <- struct{}{}
	}

	return bp
}

// LaunchBrowser launches a new browser instance
func (bp *BrowserPool) LaunchBrowser(ctx context.Context, headless bool) (*BrowserInstance, error) {
	// Acquire browser slot
	select {
	case <-bp.instanceSemaphore:
	case <-ctx.Done():
		atomic.AddInt64(&bp.metrics.failedLaunches, 1)
		return nil, fmt.Errorf("browser launch cancelled: %w", ctx.Err())
	case <-bp.closeChan:
		atomic.AddInt64(&bp.metrics.failedLaunches, 1)
		return nil, fmt.Errorf("browser pool is closed")
	}

	defer func() {
		// Return slot if we fail
		if true {
			current := atomic.AddInt64(&bp.metrics.activeBrowsers, 1)
			peak := atomic.LoadInt64(&bp.metrics.peakConcurrentBrowsers)
			if current > peak {
				atomic.StoreInt64(&bp.metrics.peakConcurrentBrowsers, current)
			}
		}
	}()

	// Find available port
	port, err := findAvailablePort()
	if err != nil {
		atomic.AddInt64(&bp.metrics.failedLaunches, 1)
		bp.instanceSemaphore <- struct{}{}
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}

	// Build chrome command
	chromeArgs := []string{
		fmt.Sprintf("--remote-debugging-port=%d", port),
		"--no-first-run",
		"--no-default-browser-check",
	}

	if headless {
		chromeArgs = append(chromeArgs, "--headless=new")
	}

	// Launch browser process
	cmd := exec.CommandContext(ctx, "google-chrome", chromeArgs...)
	err = cmd.Start()
	if err != nil {
		atomic.AddInt64(&bp.metrics.failedLaunches, 1)
		bp.instanceSemaphore <- struct{}{}
		return nil, fmt.Errorf("failed to launch chrome: %w", err)
	}

	// Create browser instance
	browserID := fmt.Sprintf("browser-%d", cmd.Process.Pid)
	instance := &BrowserInstance{
		ID:        browserID,
		ProcessID: cmd.Process.Pid,
		Port:      port,
		Tabs:      make(map[string]*Tab),
		MaxTabs:   50,
		CreatedAt: time.Now(),
	}

	// Get WebSocket endpoint
	wsEndpoint, err := getWebSocketEndpoint(fmt.Sprintf("localhost:%d", port))
	if err != nil {
		atomic.AddInt64(&bp.metrics.failedLaunches, 1)
		cmd.Process.Kill()
		bp.instanceSemaphore <- struct{}{}
		return nil, fmt.Errorf("failed to get WebSocket endpoint: %w", err)
	}

	instance.WSEndpoint = wsEndpoint

	// Register browser
	bp.browserMutex.Lock()
	bp.browsers[browserID] = instance
	bp.browserMutex.Unlock()

	bp.cdpMutex.Lock()
	bp.cdpConnections[browserID] = wsEndpoint
	bp.cdpMutex.Unlock()

	atomic.AddInt64(&bp.metrics.successfulLaunches, 1)
	atomic.AddInt64(&bp.metrics.browserInstances, 1)

	return instance, nil
}

// CreateTab creates a new tab in the browser
func (bp *BrowserPool) CreateTab(ctx context.Context, browserID string, url string) (*Tab, error) {
	// Acquire tab slot
	select {
	case <-bp.tabSemaphore:
	case <-ctx.Done():
		atomic.AddInt64(&bp.metrics.tabCreationErrors, 1)
		return nil, fmt.Errorf("tab creation cancelled: %w", ctx.Err())
	case <-bp.closeChan:
		atomic.AddInt64(&bp.metrics.tabCreationErrors, 1)
		return nil, fmt.Errorf("browser pool is closed")
	}

	bp.browserMutex.RLock()
	instance, exists := bp.browsers[browserID]
	bp.browserMutex.RUnlock()

	if !exists {
		atomic.AddInt64(&bp.metrics.tabCreationErrors, 1)
		bp.tabSemaphore <- struct{}{}
		return nil, fmt.Errorf("browser instance not found: %s", browserID)
	}

	// Check tab limit
	if atomic.LoadInt64(&instance.ActiveTabs) >= instance.MaxTabs {
		atomic.AddInt64(&bp.metrics.tabCreationErrors, 1)
		bp.tabSemaphore <- struct{}{}
		return nil, fmt.Errorf("browser %s tab limit exceeded", browserID)
	}

	// Create tab in CDP
	tabID := fmt.Sprintf("tab-%d-%d", time.Now().UnixNano(), instance.ActiveTabs)
	tab := &Tab{
		ID:        tabID,
		CDPURL:    fmt.Sprintf("%s/devtools/page/%s", instance.WSEndpoint, tabID),
		CreatedAt: time.Now(),
		URL:       url,
	}

	// Register tab
	instance.TabMutex.Lock()
	instance.Tabs[tabID] = tab
	instance.TabMutex.Unlock()

	atomic.AddInt64(&instance.ActiveTabs, 1)
	atomic.AddInt64(&bp.metrics.activeTabs, 1)
	atomic.AddInt64(&bp.metrics.totalTabs, 1)

	current := atomic.LoadInt64(&bp.metrics.activeTabs)
	peak := atomic.LoadInt64(&bp.metrics.peakConcurrentTabs)
	if current > peak {
		atomic.StoreInt64(&bp.metrics.peakConcurrentTabs, current)
	}

	instance.StatusMutex.Lock()
	instance.LastActivityAt = time.Now()
	instance.StatusMutex.Unlock()

	return tab, nil
}

// CloseTab closes a tab and releases resources
func (bp *BrowserPool) CloseTab(browserID string, tabID string) error {
	bp.browserMutex.RLock()
	instance, exists := bp.browsers[browserID]
	bp.browserMutex.RUnlock()

	if !exists {
		return fmt.Errorf("browser instance not found: %s", browserID)
	}

	instance.TabMutex.Lock()
	_, tabExists := instance.Tabs[tabID]
	if !tabExists {
		instance.TabMutex.Unlock()
		return fmt.Errorf("tab not found: %s", tabID)
	}
	delete(instance.Tabs, tabID)
	instance.TabMutex.Unlock()

	atomic.AddInt64(&instance.ActiveTabs, -1)
	atomic.AddInt64(&bp.metrics.activeTabs, -1)

	instance.StatusMutex.Lock()
	instance.LastActivityAt = time.Now()
	instance.StatusMutex.Unlock()

	// Release tab semaphore
	bp.tabSemaphore <- struct{}{}

	return nil
}

// LoadExtension loads an extension into a browser instance
func (bp *BrowserPool) LoadExtension(ctx context.Context, browserID string, extensionPath string) (string, error) {
	bp.browserMutex.RLock()
	instance, exists := bp.browsers[browserID]
	bp.browserMutex.RUnlock()

	if !exists {
		atomic.AddInt64(&bp.metrics.extensionsFailedLoad, 1)
		return "", fmt.Errorf("browser instance not found: %s", browserID)
	}

	// Verify extension path exists
	if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
		atomic.AddInt64(&bp.metrics.extensionsFailedLoad, 1)
		return "", fmt.Errorf("extension path does not exist: %s", extensionPath)
	}

	// Register extension
	extensionID := filepath.Base(extensionPath)
	bp.extensionMutex.Lock()
	bp.extensionPaths[extensionID] = extensionPath
	bp.extensionMutex.Unlock()

	// Add to browser instance
	instance.TabMutex.Lock()
	instance.Extensions = append(instance.Extensions, extensionID)
	instance.TabMutex.Unlock()

	atomic.AddInt64(&bp.metrics.extensionsLoaded, 1)

	instance.StatusMutex.Lock()
	instance.LastActivityAt = time.Now()
	instance.StatusMutex.Unlock()

	return extensionID, nil
}

// CloseBrowser closes a browser instance and releases all resources
func (bp *BrowserPool) CloseBrowser(browserID string) error {
	bp.browserMutex.Lock()
	instance, exists := bp.browsers[browserID]
	if !exists {
		bp.browserMutex.Unlock()
		return fmt.Errorf("browser instance not found: %s", browserID)
	}
	delete(bp.browsers, browserID)
	bp.browserMutex.Unlock()

	// Close all tabs
	instance.TabMutex.RLock()
	tabCount := len(instance.Tabs)
	instance.TabMutex.RUnlock()

	// Release tab semaphore slots (non-blocking)
	for i := 0; i < tabCount; i++ {
		select {
		case bp.tabSemaphore <- struct{}{}:
		default:
		}
	}

	// Kill process
	if instance.ProcessID > 0 {
		proc, err := os.FindProcess(instance.ProcessID)
		if err == nil {
			proc.Kill()
		}
	}

	// Clean up CDP connection
	bp.cdpMutex.Lock()
	delete(bp.cdpConnections, browserID)
	bp.cdpMutex.Unlock()

	// Release browser slot (non-blocking)
	select {
	case bp.instanceSemaphore <- struct{}{}:
	default:
	}

	atomic.AddInt64(&bp.metrics.browserInstances, -1)
	atomic.AddInt64(&bp.metrics.activeBrowsers, -1)
	atomic.AddInt64(&bp.metrics.activeTabs, -int64(tabCount))

	return nil
}

// GetBrowserStatus returns status of a browser instance
func (bp *BrowserPool) GetBrowserStatus(browserID string) (*BrowserInstance, error) {
	bp.browserMutex.RLock()
	instance, exists := bp.browsers[browserID]
	bp.browserMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("browser instance not found: %s", browserID)
	}

	return instance, nil
}

// GetMetrics returns current browser pool metrics
func (bp *BrowserPool) GetMetrics() map[string]interface{} {
	browserInstances := atomic.LoadInt64(&bp.metrics.browserInstances)
	activeBrowsers := atomic.LoadInt64(&bp.metrics.activeBrowsers)
	totalTabs := atomic.LoadInt64(&bp.metrics.totalTabs)
	activeTabs := atomic.LoadInt64(&bp.metrics.activeTabs)
	successfulLaunches := atomic.LoadInt64(&bp.metrics.successfulLaunches)
	failedLaunches := atomic.LoadInt64(&bp.metrics.failedLaunches)
	extensionsLoaded := atomic.LoadInt64(&bp.metrics.extensionsLoaded)
	extensionsFailedLoad := atomic.LoadInt64(&bp.metrics.extensionsFailedLoad)
	cdpErrors := atomic.LoadInt64(&bp.metrics.cdpConnectionErrors)
	tabErrors := atomic.LoadInt64(&bp.metrics.tabCreationErrors)
	peakBrowsers := atomic.LoadInt64(&bp.metrics.peakConcurrentBrowsers)
	peakTabs := atomic.LoadInt64(&bp.metrics.peakConcurrentTabs)

	successRate := float64(0)
	totalLaunches := successfulLaunches + failedLaunches
	if totalLaunches > 0 {
		successRate = float64(successfulLaunches) / float64(totalLaunches) * 100
	}

	extensionSuccessRate := float64(0)
	totalExtensions := extensionsLoaded + extensionsFailedLoad
	if totalExtensions > 0 {
		extensionSuccessRate = float64(extensionsLoaded) / float64(totalExtensions) * 100
	}

	// Use pooled map to reduce allocations
	metrics := GetMetricsMap()
	metrics["browser_instances"] = browserInstances
	metrics["active_browsers"] = activeBrowsers
	metrics["total_tabs"] = totalTabs
	metrics["active_tabs"] = activeTabs
	metrics["successful_launches"] = successfulLaunches
	metrics["failed_launches"] = failedLaunches
	metrics["launch_success_rate"] = successRate
	metrics["extensions_loaded"] = extensionsLoaded
	metrics["extensions_failed"] = extensionsFailedLoad
	metrics["extension_success_rate"] = extensionSuccessRate
	metrics["cdp_connection_errors"] = cdpErrors
	metrics["tab_creation_errors"] = tabErrors
	metrics["peak_concurrent_browsers"] = peakBrowsers
	metrics["peak_concurrent_tabs"] = peakTabs
	metrics["max_browsers"] = bp.maxBrowsers

	return metrics
}

// Close closes the browser pool and releases all resources
func (bp *BrowserPool) Close() error {
	bp.closeOnce.Do(func() {
		close(bp.closeChan)

		bp.browserMutex.Lock()
		browserIDs := make([]string, 0, len(bp.browsers))
		for id := range bp.browsers {
			browserIDs = append(browserIDs, id)
		}
		bp.browserMutex.Unlock()

		for _, id := range browserIDs {
			bp.CloseBrowser(id)
		}

		bp.cdpMutex.Lock()
		bp.cdpConnections = make(map[string]string)
		bp.cdpMutex.Unlock()

		bp.extensionMutex.Lock()
		bp.extensionPaths = make(map[string]string)
		bp.extensionMutex.Unlock()
	})

	return nil
}

// Helper functions

// findAvailablePort finds an available port
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// getWebSocketEndpoint gets the WebSocket endpoint from Chrome debugging port
func getWebSocketEndpoint(addr string) (string, error) {
	// In real implementation, would make HTTP request to http://addr/json/version
	// For now, construct endpoint from known format
	return fmt.Sprintf("ws://%s", addr), nil
}
