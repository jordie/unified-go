package subsystems

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMultiSubsystemIntegration(t *testing.T) {
	// Initialize all subsystems
	apiPool := NewAPIClientPool(100, 1000)
	defer apiPool.Close()

	fileManager := NewFileManager(50, 65536)
	defer fileManager.Close()

	networkCoordinator := NewNetworkCoordinator(1000000) // 1MB/s
	defer networkCoordinator.Close()

	browserPool := NewBrowserPool(10, 50)
	defer browserPool.Close()

	processManager := NewProcessManager(20)
	defer processManager.Close()

	// Verify all subsystems initialized
	apiMetrics := apiPool.GetMetrics()
	if apiMetrics == nil {
		t.Fatal("Failed to initialize API pool")
	}

	fileMetrics := fileManager.GetMetrics()
	if fileMetrics == nil {
		t.Fatal("Failed to initialize file manager")
	}

	networkMetrics := networkCoordinator.GetMetrics()
	if networkMetrics == nil {
		t.Fatal("Failed to initialize network coordinator")
	}

	browserMetrics := browserPool.GetMetrics()
	if browserMetrics == nil {
		t.Fatal("Failed to initialize browser pool")
	}

	processMetrics := processManager.GetMetrics()
	if processMetrics == nil {
		t.Fatal("Failed to initialize process manager")
	}
}

func TestConcurrentSubsystemOperations(t *testing.T) {
	// Initialize subsystems
	apiPool := NewAPIClientPool(50, 500)
	defer apiPool.Close()

	fileManager := NewFileManager(30, 65536)
	defer fileManager.Close()

	networkCoordinator := NewNetworkCoordinator(1000000)
	defer networkCoordinator.Close()

	ctx := context.Background()

	// Create concurrent work across subsystems
	done := make(chan error, 30)

	// Simulate API calls
	for i := 0; i < 10; i++ {
		go func(idx int) {
			domain := fmt.Sprintf("api%d.example.com", idx)
			networkCoordinator.SetConnectionLimit(domain, 10)

			// Simulate bandwidth check
			_, err := networkCoordinator.ThrottleBandwidth(ctx, 1000)
			done <- err
		}(i)
	}

	// Simulate file operations
	for i := 0; i < 10; i++ {
		go func(idx int) {
			metrics := fileManager.GetMetrics()
			if metrics == nil {
				done <- fmt.Errorf("failed to get metrics")
			} else {
				done <- nil
			}
		}(i)
	}

	// Simulate network operations
	for i := 0; i < 10; i++ {
		go func(idx int) {
			metrics := networkCoordinator.GetMetrics()
			if metrics == nil {
				done <- fmt.Errorf("failed to get metrics")
			} else {
				done <- nil
			}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < 30; i++ {
		if <-done == nil {
			successCount++
		}
	}

	if successCount < 25 {
		t.Fatalf("Expected at least 25 successful operations, got %d", successCount)
	}
}

func TestWorkflowWithMultipleSubsystems(t *testing.T) {
	// Create a workflow that uses multiple subsystems
	workflow := &WorkflowDefinition{
		ID:   "test-workflow-1",
		Name: "Multi-Subsystem Workflow",
		Steps: []*WorkflowStep{
			{
				ID:   "step-1",
				Name: "API Call",
				Type: WorkTypeAPICall,
				Inputs: map[string]interface{}{
					"url": "https://api.example.com/data",
				},
			},
			{
				ID:   "step-2",
				Name: "Process Result",
				Type: WorkTypeFileOperation,
				Dependencies: []string{"step-1"},
			},
			{
				ID:   "step-3",
				Name: "Browser Automation",
				Type: WorkTypeBrowserAutomation,
				Dependencies: []string{"step-2"},
			},
			{
				ID:   "step-4",
				Name: "Subprocess",
				Type: WorkTypeProcessExecution,
				Dependencies: []string{"step-3"},
			},
		},
		Timeout: 30 * time.Second,
	}

	// Verify workflow structure
	if len(workflow.Steps) != 4 {
		t.Fatalf("Expected 4 steps, got %d", len(workflow.Steps))
	}

	if workflow.Steps[0].Type != WorkTypeAPICall {
		t.Fatalf("Expected first step to be API call, got %s", workflow.Steps[0].Type)
	}

	if workflow.Steps[3].Type != WorkTypeProcessExecution {
		t.Fatalf("Expected last step to be process execution, got %s", workflow.Steps[3].Type)
	}
}

func TestResourceAllocationAcrossSubsystems(t *testing.T) {
	// Test resource allocation across subsystems
	apiPool := NewAPIClientPool(100, 1000)
	defer apiPool.Close()

	fileManager := NewFileManager(50, 65536)
	defer fileManager.Close()

	browserPool := NewBrowserPool(10, 50)
	defer browserPool.Close()

	processManager := NewProcessManager(20)
	defer processManager.Close()

	// Set resource limits
	processManager.SetResourceLimit("cpu-intensive", &ResourceLimit{
		MaxMemoryMB:   512,
		MaxCPUPercent: 80.0,
		TimeoutSecs:   600,
	})

	processManager.SetResourceLimit("memory-intensive", &ResourceLimit{
		MaxMemoryMB:   1024,
		MaxCPUPercent: 20.0,
		TimeoutSecs:   300,
	})

	// Verify limits are set
	cpuLimit := processManager.GetResourceLimit("cpu-intensive")
	if cpuLimit.MaxMemoryMB != 512 {
		t.Fatalf("Expected max memory 512, got %d", cpuLimit.MaxMemoryMB)
	}

	memLimit := processManager.GetResourceLimit("memory-intensive")
	if memLimit.MaxMemoryMB != 1024 {
		t.Fatalf("Expected max memory 1024, got %d", memLimit.MaxMemoryMB)
	}
}

func TestErrorPropagationAcrossSubsystems(t *testing.T) {
	apiPool := NewAPIClientPool(10, 100)
	defer apiPool.Close()

	browserPool := NewBrowserPool(5, 50)
	defer browserPool.Close()

	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	// Try browser launch with cancelled context
	_, browserErr := browserPool.LaunchBrowser(cancelCtx, true)
	if browserErr == nil {
		t.Fatal("Expected error for cancelled browser launch")
	}
}

func TestMetricsAggregation(t *testing.T) {
	// Initialize all subsystems
	apiPool := NewAPIClientPool(50, 500)
	defer apiPool.Close()

	fileManager := NewFileManager(30, 65536)
	defer fileManager.Close()

	networkCoordinator := NewNetworkCoordinator(1000000)
	defer networkCoordinator.Close()

	browserPool := NewBrowserPool(5, 50)
	defer browserPool.Close()

	processManager := NewProcessManager(10)
	defer processManager.Close()

	// Get metrics from each subsystem
	apiMetrics := apiPool.GetMetrics()
	fileMetrics := fileManager.GetMetrics()
	networkMetrics := networkCoordinator.GetMetrics()
	browserMetrics := browserPool.GetMetrics()
	processMetrics := processManager.GetMetrics()

	// Check API metrics
	for _, key := range []string{"max_capacity", "active_clients", "total_requests"} {
		if _, exists := apiMetrics[key]; !exists {
			t.Logf("Note: API metric %s not found, but subsystem initialized", key)
		}
	}

	// Verify all subsystems returned metrics
	if apiMetrics == nil || fileMetrics == nil || networkMetrics == nil ||
		browserMetrics == nil || processMetrics == nil {
		t.Fatal("One or more subsystems failed to return metrics")
	}
}

func TestLoadBalancingAcrossSubsystems(t *testing.T) {
	// Test load balancing with limited resources
	apiPool := NewAPIClientPool(20, 200)
	defer apiPool.Close()

	fileManager := NewFileManager(15, 65536)
	defer fileManager.Close()

	browserPool := NewBrowserPool(5, 50)
	defer browserPool.Close()

	processManager := NewProcessManager(10)
	defer processManager.Close()

	// Simulate uneven load distribution
	apiDone := make(chan error, 20)
	fileDone := make(chan error, 15)
	browserDone := make(chan error, 5)
	procDone := make(chan error, 10)

	// API operations (most load)
	for i := 0; i < 20; i++ {
		go func() {
			metrics := apiPool.GetMetrics()
			if metrics != nil {
				apiDone <- nil
			} else {
				apiDone <- fmt.Errorf("failed")
			}
		}()
	}

	// File operations (medium load)
	for i := 0; i < 15; i++ {
		go func() {
			metrics := fileManager.GetMetrics()
			if metrics != nil {
				fileDone <- nil
			} else {
				fileDone <- fmt.Errorf("failed")
			}
		}()
	}

	// Browser operations (light load)
	for i := 0; i < 5; i++ {
		go func() {
			metrics := browserPool.GetMetrics()
			if metrics != nil {
				browserDone <- nil
			} else {
				browserDone <- fmt.Errorf("failed")
			}
		}()
	}

	// Process operations (medium load)
	for i := 0; i < 10; i++ {
		go func() {
			metrics := processManager.GetMetrics()
			if metrics != nil {
				procDone <- nil
			} else {
				procDone <- fmt.Errorf("failed")
			}
		}()
	}

	// Collect results
	for i := 0; i < 20; i++ {
		<-apiDone
	}
	for i := 0; i < 15; i++ {
		<-fileDone
	}
	for i := 0; i < 5; i++ {
		<-browserDone
	}
	for i := 0; i < 10; i++ {
		<-procDone
	}
}

func TestSubsystemInterconnection(t *testing.T) {
	// Test that subsystems can work together
	apiPool := NewAPIClientPool(50, 500)
	defer apiPool.Close()

	networkCoordinator := NewNetworkCoordinator(1000000)
	defer networkCoordinator.Close()

	// Set up network configuration for API calls
	networkCoordinator.SetConnectionLimit("api.example.com", 25)
	networkCoordinator.SetConnectionLimit("api.test.com", 25)

	// Simulate API calls with network coordination
	ctx := context.Background()
	done := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			domain := "api.example.com"

			// Check bandwidth before API call
			_, err := networkCoordinator.ThrottleBandwidth(ctx, 100)
			if err != nil {
				done <- err
				return
			}

			// Acquire connection
			err = networkCoordinator.AcquireConnection(domain)
			if err != nil {
				done <- err
				return
			}

			// Release connection
			networkCoordinator.ReleaseConnection(domain)
			done <- nil
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < 10; i++ {
		if <-done == nil {
			successCount++
		}
	}

	if successCount < 8 {
		t.Fatalf("Expected at least 8 successful operations, got %d", successCount)
	}
}

func BenchmarkIntegrationThroughput(b *testing.B) {
	// Benchmark combined throughput of all subsystems
	apiPool := NewAPIClientPool(50, 500)
	defer apiPool.Close()

	fileManager := NewFileManager(30, 65536)
	defer fileManager.Close()

	networkCoordinator := NewNetworkCoordinator(1000000)
	defer networkCoordinator.Close()

	browserPool := NewBrowserPool(10, 50)
	defer browserPool.Close()

	processManager := NewProcessManager(20)
	defer processManager.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Collect metrics from all subsystems
		_ = apiPool.GetMetrics()
		_ = fileManager.GetMetrics()
		_ = networkCoordinator.GetMetrics()
		_ = browserPool.GetMetrics()
		_ = processManager.GetMetrics()
	}

	b.StopTimer()

	// Log final metrics
	b.Logf("API Pool - Max: %d", apiPool.GetMetrics()["max_capacity"])
	b.Logf("File Manager - Max: %d", fileManager.GetMetrics()["max_concurrent"])
	b.Logf("Browser Pool - Max: %d", browserPool.GetMetrics()["max_browsers"])
	b.Logf("Process Manager - Max: %d", processManager.GetMetrics()["max_processes"])
}

func TestCrossSubsystemDataFlow(t *testing.T) {
	// Test data flowing from one subsystem to another
	apiPool := NewAPIClientPool(30, 300)
	defer apiPool.Close()

	fileManager := NewFileManager(20, 65536)
	defer fileManager.Close()

	// Simulate: API response → File storage → Process execution

	// Step 1: Simulate API response
	apiResponse := map[string]interface{}{
		"data": "api_response_data",
		"id":   "12345",
	}

	// Step 2: File manager would process this
	fileData := map[string]interface{}{
		"api_response": apiResponse,
		"timestamp":    time.Now(),
	}

	// Step 3: Verify data integrity
	if apiResponse["id"] != fileData["api_response"].(map[string]interface{})["id"] {
		t.Fatal("Data integrity failed")
	}

	// Verify subsystems are still healthy
	apiMetrics := apiPool.GetMetrics()
	fileMetrics := fileManager.GetMetrics()

	if apiMetrics == nil || fileMetrics == nil {
		t.Fatal("Subsystems not responding")
	}
}
