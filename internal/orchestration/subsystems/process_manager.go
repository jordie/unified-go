package subsystems

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ProcessManager manages a pool of subprocess instances with resource limiting
type ProcessManager struct {
	maxProcesses      int64
	activeProcesses   int64
	processes         map[string]*ProcessInstance
	processMutex      sync.RWMutex
	resourceLimits    map[string]*ResourceLimit
	limitsLock        sync.RWMutex
	metrics           *ProcessMetrics
	processSemaphore  chan struct{}
	signalChan        chan os.Signal
	closeChan         chan struct{}
	closeOnce         sync.Once
}

// ProcessInstance represents a single subprocess
type ProcessInstance struct {
	ID              string
	Command         string
	Args            []string
	ProcessID       int
	Process         *exec.Cmd
	State           ProcessState
	CreatedAt       time.Time
	StartedAt       *time.Time
	TerminatedAt    *time.Time
	StatusMutex     sync.RWMutex
	Stdout          string
	Stderr          string
	ExitCode        int
	ResourceUsage   *ResourceUsage
	MemoryMBLimit   int64
	CPULimitPercent float64
}

// ProcessState represents the state of a process
type ProcessState string

const (
	ProcessCreated   ProcessState = "created"
	ProcessRunning   ProcessState = "running"
	ProcessCompleted ProcessState = "completed"
	ProcessFailed    ProcessState = "failed"
	ProcessKilled    ProcessState = "killed"
)

// ResourceLimit specifies resource constraints per process
type ResourceLimit struct {
	MaxMemoryMB   int64
	MaxCPUPercent float64
	TimeoutSecs   int64
}

// ResourceUsage tracks actual resource consumption
type ResourceUsage struct {
	MemoryMB    int64
	CPUPercent  float64
	ElapsedSecs int64
}

// ProcessMetrics tracks process pool performance
type ProcessMetrics struct {
	processInstances    int64
	activeProcesses     int64
	successfulStarts    int64
	failedStarts        int64
	successfulShutdowns int64
	failedShutdowns     int64
	processesKilled     int64
	timeoutOccurrences  int64
	totalProcessTime    int64
	peakConcurrentProcs int64
}

// NewProcessManager creates a new process manager
// maxProcesses: maximum concurrent subprocess instances (typically 200 for GAIA)
func NewProcessManager(maxProcesses int64) *ProcessManager {
	pm := &ProcessManager{
		maxProcesses:     maxProcesses,
		processes:        make(map[string]*ProcessInstance),
		resourceLimits:   make(map[string]*ResourceLimit),
		metrics:          &ProcessMetrics{},
		processSemaphore: make(chan struct{}, maxProcesses),
		signalChan:       make(chan os.Signal, 10),
		closeChan:        make(chan struct{}),
	}

	// Fill semaphore with initial tokens
	for i := int64(0); i < maxProcesses; i++ {
		pm.processSemaphore <- struct{}{}
	}

	// Setup signal handling for graceful shutdown
	signal.Notify(pm.signalChan, syscall.SIGTERM, syscall.SIGINT)

	return pm
}

// SetResourceLimit sets resource constraints for a process type
func (pm *ProcessManager) SetResourceLimit(processType string, limit *ResourceLimit) {
	pm.limitsLock.Lock()
	pm.resourceLimits[processType] = limit
	pm.limitsLock.Unlock()
}

// GetResourceLimit gets resource constraints for a process type
func (pm *ProcessManager) GetResourceLimit(processType string) *ResourceLimit {
	pm.limitsLock.RLock()
	defer pm.limitsLock.RUnlock()

	if limit, exists := pm.resourceLimits[processType]; exists {
		return limit
	}

	// Default limits
	return &ResourceLimit{
		MaxMemoryMB:   256,
		MaxCPUPercent: 50.0,
		TimeoutSecs:   300,
	}
}

// StartProcess starts a new subprocess
func (pm *ProcessManager) StartProcess(ctx context.Context, processType string, command string, args []string) (*ProcessInstance, error) {
	// Acquire process slot
	select {
	case <-pm.processSemaphore:
	case <-ctx.Done():
		atomic.AddInt64(&pm.metrics.failedStarts, 1)
		return nil, fmt.Errorf("process start cancelled: %w", ctx.Err())
	case <-pm.closeChan:
		atomic.AddInt64(&pm.metrics.failedStarts, 1)
		return nil, fmt.Errorf("process manager is closed")
	}

	// Get resource limits
	limits := pm.GetResourceLimit(processType)

	// Create process instance
	processID := fmt.Sprintf("proc-%d", time.Now().UnixNano())
	instance := &ProcessInstance{
		ID:                processID,
		Command:           command,
		Args:              args,
		State:             ProcessCreated,
		CreatedAt:         time.Now(),
		ResourceUsage:     &ResourceUsage{},
		MemoryMBLimit:     limits.MaxMemoryMB,
		CPULimitPercent:   limits.MaxCPUPercent,
	}

	// Create command with context timeout
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(limits.TimeoutSecs)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, command, args...)
	instance.Process = cmd

	// Start process
	err := cmd.Start()
	if err != nil {
		atomic.AddInt64(&pm.metrics.failedStarts, 1)
		pm.processSemaphore <- struct{}{}
		return nil, fmt.Errorf("failed to start process %s: %w", command, err)
	}

	instance.ProcessID = cmd.Process.Pid
	instance.StatusMutex.Lock()
	instance.State = ProcessRunning
	now := time.Now()
	instance.StartedAt = &now
	instance.StatusMutex.Unlock()

	// Register process
	pm.processMutex.Lock()
	pm.processes[processID] = instance
	pm.processMutex.Unlock()

	atomic.AddInt64(&pm.metrics.processInstances, 1)
	atomic.AddInt64(&pm.metrics.activeProcesses, 1)
	atomic.AddInt64(&pm.metrics.successfulStarts, 1)

	current := atomic.LoadInt64(&pm.metrics.activeProcesses)
	peak := atomic.LoadInt64(&pm.metrics.peakConcurrentProcs)
	if current > peak {
		atomic.StoreInt64(&pm.metrics.peakConcurrentProcs, current)
	}

	// Wait for process to complete in background
	go pm.waitForProcessCompletion(processID, instance)

	return instance, nil
}

// waitForProcessCompletion waits for a process to finish and cleans up
func (pm *ProcessManager) waitForProcessCompletion(processID string, instance *ProcessInstance) {
	if instance.Process == nil {
		return
	}

	err := instance.Process.Wait()

	instance.StatusMutex.Lock()
	now := time.Now()
	instance.TerminatedAt = &now

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				instance.ExitCode = status.ExitStatus()
			}
			instance.State = ProcessFailed
		} else {
			instance.State = ProcessFailed
		}
	} else {
		instance.State = ProcessCompleted
		instance.ExitCode = 0
	}

	if instance.StartedAt != nil && instance.TerminatedAt != nil {
		elapsed := instance.TerminatedAt.Sub(*instance.StartedAt).Seconds()
		atomic.AddInt64(&pm.metrics.totalProcessTime, int64(elapsed))
	}

	instance.StatusMutex.Unlock()

	// Resource tracking
	instance.ResourceUsage.ElapsedSecs = int64(time.Since(instance.CreatedAt).Seconds())
}

// TerminateProcess terminates a running process gracefully
func (pm *ProcessManager) TerminateProcess(processID string, gracefulTimeout time.Duration) error {
	pm.processMutex.RLock()
	instance, exists := pm.processes[processID]
	pm.processMutex.RUnlock()

	if !exists {
		return fmt.Errorf("process not found: %s", processID)
	}

	instance.StatusMutex.RLock()
	state := instance.State
	instance.StatusMutex.RUnlock()

	if state != ProcessRunning {
		return fmt.Errorf("process not running: %s", processID)
	}

	// Try graceful termination first
	if instance.Process != nil && instance.Process.Process != nil {
		instance.Process.Process.Signal(syscall.SIGTERM)

		// Wait for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- instance.Process.Wait()
		}()

		select {
		case <-time.After(gracefulTimeout):
			// Force kill if graceful shutdown takes too long
			instance.Process.Process.Kill()
			atomic.AddInt64(&pm.metrics.processesKilled, 1)
		case <-done:
			atomic.AddInt64(&pm.metrics.successfulShutdowns, 1)
		}
	}

	instance.StatusMutex.Lock()
	instance.State = ProcessKilled
	now := time.Now()
	instance.TerminatedAt = &now
	instance.StatusMutex.Unlock()

	return nil
}

// KillProcess immediately kills a process (no graceful shutdown)
func (pm *ProcessManager) KillProcess(processID string) error {
	pm.processMutex.RLock()
	instance, exists := pm.processes[processID]
	pm.processMutex.RUnlock()

	if !exists {
		return fmt.Errorf("process not found: %s", processID)
	}

	if instance.Process != nil && instance.Process.Process != nil {
		err := instance.Process.Process.Kill()
		if err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
		atomic.AddInt64(&pm.metrics.processesKilled, 1)
	}

	instance.StatusMutex.Lock()
	instance.State = ProcessKilled
	now := time.Now()
	instance.TerminatedAt = &now
	instance.StatusMutex.Unlock()

	return nil
}

// GetProcessStatus returns the status of a process
func (pm *ProcessManager) GetProcessStatus(processID string) (*ProcessInstance, error) {
	pm.processMutex.RLock()
	instance, exists := pm.processes[processID]
	pm.processMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("process not found: %s", processID)
	}

	return instance, nil
}

// WaitForProcess waits for a process to complete
func (pm *ProcessManager) WaitForProcess(ctx context.Context, processID string) error {
	pm.processMutex.RLock()
	instance, exists := pm.processes[processID]
	pm.processMutex.RUnlock()

	if !exists {
		return fmt.Errorf("process not found: %s", processID)
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-pm.closeChan:
			return fmt.Errorf("process manager is closed")
		case <-ticker.C:
			instance.StatusMutex.RLock()
			state := instance.State
			instance.StatusMutex.RUnlock()

			if state == ProcessCompleted || state == ProcessFailed || state == ProcessKilled {
				return nil
			}
		}
	}
}

// ListProcesses returns all current processes
func (pm *ProcessManager) ListProcesses() []*ProcessInstance {
	pm.processMutex.RLock()
	defer pm.processMutex.RUnlock()

	processes := make([]*ProcessInstance, 0, len(pm.processes))
	for _, p := range pm.processes {
		processes = append(processes, p)
	}

	return processes
}

// CleanupProcess removes a completed process from tracking
func (pm *ProcessManager) CleanupProcess(processID string) error {
	pm.processMutex.Lock()
	instance, exists := pm.processes[processID]
	if !exists {
		pm.processMutex.Unlock()
		return fmt.Errorf("process not found: %s", processID)
	}
	delete(pm.processes, processID)
	pm.processMutex.Unlock()

	// Release process slot
	select {
	case pm.processSemaphore <- struct{}{}:
	default:
	}

	atomic.AddInt64(&pm.metrics.processInstances, -1)
	atomic.AddInt64(&pm.metrics.activeProcesses, -1)

	instance.StatusMutex.RLock()
	state := instance.State
	instance.StatusMutex.RUnlock()

	if state == ProcessCompleted {
		atomic.AddInt64(&pm.metrics.successfulShutdowns, 1)
	} else if state == ProcessFailed || state == ProcessKilled {
		atomic.AddInt64(&pm.metrics.failedShutdowns, 1)
	}

	return nil
}

// GetMetrics returns current process pool metrics
func (pm *ProcessManager) GetMetrics() map[string]interface{} {
	processInstances := atomic.LoadInt64(&pm.metrics.processInstances)
	activeProcesses := atomic.LoadInt64(&pm.metrics.activeProcesses)
	successfulStarts := atomic.LoadInt64(&pm.metrics.successfulStarts)
	failedStarts := atomic.LoadInt64(&pm.metrics.failedStarts)
	successfulShutdowns := atomic.LoadInt64(&pm.metrics.successfulShutdowns)
	failedShutdowns := atomic.LoadInt64(&pm.metrics.failedShutdowns)
	processesKilled := atomic.LoadInt64(&pm.metrics.processesKilled)
	timeouts := atomic.LoadInt64(&pm.metrics.timeoutOccurrences)
	totalTime := atomic.LoadInt64(&pm.metrics.totalProcessTime)
	peakProcs := atomic.LoadInt64(&pm.metrics.peakConcurrentProcs)

	startSuccessRate := float64(0)
	totalStarts := successfulStarts + failedStarts
	if totalStarts > 0 {
		startSuccessRate = float64(successfulStarts) / float64(totalStarts) * 100
	}

	avgProcessTime := float64(0)
	if successfulStarts+failedStarts > 0 {
		avgProcessTime = float64(totalTime) / float64(successfulStarts+failedStarts)
	}

	// Use pooled map to reduce allocations
	metrics := GetMetricsMap()
	metrics["process_instances"] = processInstances
	metrics["active_processes"] = activeProcesses
	metrics["successful_starts"] = successfulStarts
	metrics["failed_starts"] = failedStarts
	metrics["start_success_rate"] = startSuccessRate
	metrics["successful_shutdowns"] = successfulShutdowns
	metrics["failed_shutdowns"] = failedShutdowns
	metrics["processes_killed"] = processesKilled
	metrics["timeout_occurrences"] = timeouts
	metrics["total_process_time_secs"] = totalTime
	metrics["avg_process_time_secs"] = avgProcessTime
	metrics["peak_concurrent_processes"] = peakProcs
	metrics["max_processes"] = pm.maxProcesses

	return metrics
}

// Close closes the process manager and terminates all processes
func (pm *ProcessManager) Close() error {
	pm.closeOnce.Do(func() {
		close(pm.closeChan)

		// Get all running processes
		pm.processMutex.Lock()
		processIDs := make([]string, 0, len(pm.processes))
		for id := range pm.processes {
			processIDs = append(processIDs, id)
		}
		pm.processMutex.Unlock()

		// Terminate all processes gracefully
		for _, id := range processIDs {
			pm.TerminateProcess(id, 5*time.Second)
		}

		// Clean up all process records
		for _, id := range processIDs {
			pm.CleanupProcess(id)
		}

		signal.Stop(pm.signalChan)
		close(pm.signalChan)
	})

	return nil
}
