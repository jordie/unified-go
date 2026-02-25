package metrics

import (
	"testing"
)

func TestNewHTTPMetricsRegistry(t *testing.T) {
	registry := NewHTTPMetricsRegistry()

	if registry == nil {
		t.Errorf("Expected non-nil HTTPMetricsRegistry, got nil")
	}

	if registry.GetPrometheusRegistry() == nil {
		t.Errorf("Expected non-nil Prometheus registry, got nil")
	}
}

func TestRecordRequest(t *testing.T) {
	registry := NewHTTPMetricsRegistry()

	tests := []struct {
		name       string
		method     string
		path       string
		status     int
		app        string
		duration   float64
		reqSize    int64
		respSize   int64
	}{
		{
			name:     "GET request success",
			method:   "GET",
			path:     "/api/math/problems",
			status:   200,
			app:      "math",
			duration: 0.05,
			reqSize:  -1,
			respSize: 1024,
		},
		{
			name:     "POST request with body",
			method:   "POST",
			path:     "/api/typing/test",
			status:   201,
			app:      "typing",
			duration: 0.1,
			reqSize:  512,
			respSize: 256,
		},
		{
			name:     "4xx client error",
			method:   "GET",
			path:     "/api/users/:id",
			status:   404,
			app:      "core",
			duration: 0.01,
			reqSize:  -1,
			respSize: 128,
		},
		{
			name:     "5xx server error",
			method:   "POST",
			path:     "/api/piano/songs",
			status:   500,
			app:      "piano",
			duration: 0.5,
			reqSize:  1024,
			respSize: 256,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			registry.RecordRequest(tt.method, tt.path, tt.status, tt.app, tt.duration, tt.reqSize, tt.respSize)
		})
	}
}

func TestActiveRequests(t *testing.T) {
	registry := NewHTTPMetricsRegistry()

	// Increment
	registry.IncrementActiveRequests()
	registry.IncrementActiveRequests()

	// Decrement
	registry.DecrementActiveRequests()

	// Should not panic
	registry.DecrementActiveRequests()
	registry.DecrementActiveRequests()
}

func TestMetricsWithVariousStatusCodes(t *testing.T) {
	registry := NewHTTPMetricsRegistry()

	statuses := []int{200, 201, 204, 400, 401, 403, 404, 500, 502, 503}

	for _, status := range statuses {
		registry.RecordRequest("GET", "/test", status, "core", 0.05, -1, 256)
	}
}

func TestMetricsWithVariousApps(t *testing.T) {
	registry := NewHTTPMetricsRegistry()

	apps := []string{"math", "typing", "reading", "piano", "core"}

	for _, app := range apps {
		registry.RecordRequest("GET", "/api/"+app+"/test", 200, app, 0.05, -1, 256)
	}
}

func TestMetricsWithLargeBodies(t *testing.T) {
	registry := NewHTTPMetricsRegistry()

	// Large request
	registry.RecordRequest("POST", "/api/upload", 200, "core", 0.2, 5*1024*1024, 1024)

	// Large response
	registry.RecordRequest("GET", "/api/export", 200, "core", 0.3, -1, 10*1024*1024)
}

func TestMetricsThreadSafety(t *testing.T) {
	registry := NewHTTPMetricsRegistry()

	done := make(chan bool)

	// Concurrent requests
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				registry.RecordRequest("GET", "/test", 200, "core", 0.05, -1, 256)
				registry.IncrementActiveRequests()
				registry.DecrementActiveRequests()
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
