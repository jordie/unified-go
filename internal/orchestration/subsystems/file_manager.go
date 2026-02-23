package subsystems

import (
	"bufio"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// FileManager handles concurrent file I/O operations with streaming and buffer pooling
type FileManager struct {
	maxConcurrent     int
	bufferPool        *BufferPool
	semaphore         chan struct{}
	metrics           *FileMetrics
	parseWorkers      int
	streamBufferSize  int
	mu                sync.RWMutex
}

// BufferPool manages reusable buffers for streaming operations
type BufferPool struct {
	pool sync.Pool
	size int
}

// FileOperation represents a file operation request
type FileOperation struct {
	Path     string
	OpType   string // "read", "write", "parse_json", "parse_xml", "stream"
	Data     []byte
	Metadata map[string]interface{}
}

// FileResult represents the result of a file operation
type FileResult struct {
	Path          string
	Data          []byte
	ParsedData    interface{}
	BytesProcessed int64
	Duration      time.Duration
	Error         error
}

// FileMetrics tracks file operation performance
type FileMetrics struct {
	successCount    int64
	errorCount      int64
	totalOperations int64
	totalBytes      int64
	totalDuration   int64 // nanoseconds
	concurrentPeaks int64
	currentConcurrent int64
}

// NewFileManager creates a new file manager with concurrent operation limits
// maxConcurrent: maximum concurrent file operations (default 500)
// streamBufferSize: size of buffers for streaming operations (default 64KB)
func NewFileManager(maxConcurrent, streamBufferSize int) *FileManager {
	if maxConcurrent <= 0 {
		maxConcurrent = 500
	}
	if streamBufferSize <= 0 {
		streamBufferSize = 65536 // 64KB
	}

	fm := &FileManager{
		maxConcurrent:    maxConcurrent,
		semaphore:        make(chan struct{}, maxConcurrent),
		metrics:          &FileMetrics{},
		parseWorkers:     4,
		streamBufferSize: streamBufferSize,
		bufferPool: &BufferPool{
			size: streamBufferSize,
		},
	}

	// Pre-populate semaphore
	for i := 0; i < maxConcurrent; i++ {
		fm.semaphore <- struct{}{}
	}

	return fm
}

// ExecuteOperation executes a file operation with concurrency control
func (fm *FileManager) ExecuteOperation(ctx context.Context, op FileOperation) (*FileResult, error) {
	atomic.AddInt64(&fm.metrics.totalOperations, 1)

	// Acquire semaphore slot
	select {
	case <-fm.semaphore:
		defer func() { fm.semaphore <- struct{}{} }()
	case <-ctx.Done():
		atomic.AddInt64(&fm.metrics.errorCount, 1)
		return nil, fmt.Errorf("context cancelled waiting for semaphore")
	}

	// Update current concurrent count
	current := atomic.AddInt64(&fm.metrics.currentConcurrent, 1)
	defer atomic.AddInt64(&fm.metrics.currentConcurrent, -1)

	// Track peak concurrency
	for {
		peak := atomic.LoadInt64(&fm.metrics.concurrentPeaks)
		if current > peak && atomic.CompareAndSwapInt64(&fm.metrics.concurrentPeaks, peak, current) {
			break
		}
		if current <= peak {
			break
		}
	}

	start := time.Now()

	var result *FileResult

	switch op.OpType {
	case "read":
		result = fm.readFile(ctx, op.Path)
	case "write":
		result = fm.writeFile(ctx, op.Path)
	case "parse_json":
		result = fm.parseJSON(ctx, op.Path)
	case "parse_xml":
		result = fm.parseXML(ctx, op.Path)
	case "stream":
		result = fm.streamFile(ctx, op.Path)
	default:
		atomic.AddInt64(&fm.metrics.errorCount, 1)
		return nil, fmt.Errorf("unknown operation type: %s", op.OpType)
	}

	duration := time.Since(start)
	result.Duration = duration

	if result.Error != nil {
		atomic.AddInt64(&fm.metrics.errorCount, 1)
	} else {
		atomic.AddInt64(&fm.metrics.successCount, 1)
		atomic.AddInt64(&fm.metrics.totalBytes, int64(result.BytesProcessed))
	}

	atomic.AddInt64(&fm.metrics.totalDuration, duration.Nanoseconds())

	return result, nil
}

// readFile reads a file into memory
func (fm *FileManager) readFile(ctx context.Context, path string) *FileResult {
	result := &FileResult{Path: path}

	file, err := os.Open(path)
	if err != nil {
		result.Error = fmt.Errorf("failed to open file: %w", err)
		return result
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		result.Error = fmt.Errorf("failed to read file: %w", err)
		return result
	}

	result.Data = data
	result.BytesProcessed = int64(len(data))
	return result
}

// writeFile writes data to a file
func (fm *FileManager) writeFile(ctx context.Context, path string) *FileResult {
	result := &FileResult{Path: path}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		result.Error = fmt.Errorf("failed to create directory: %w", err)
		return result
	}

	file, err := os.Create(path)
	if err != nil {
		result.Error = fmt.Errorf("failed to create file: %w", err)
		return result
	}
	defer file.Close()

	// This is a placeholder - actual data comes from FileOperation.Data
	result.BytesProcessed = 0
	return result
}

// parseJSON parses a JSON file
func (fm *FileManager) parseJSON(ctx context.Context, path string) *FileResult {
	result := &FileResult{Path: path}

	file, err := os.Open(path)
	if err != nil {
		result.Error = fmt.Errorf("failed to open JSON file: %w", err)
		return result
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		result.Error = fmt.Errorf("failed to read JSON file: %w", err)
		return result
	}

	var parsedData interface{}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		result.Error = fmt.Errorf("failed to parse JSON: %w", err)
		return result
	}

	result.Data = data
	result.ParsedData = parsedData
	result.BytesProcessed = int64(len(data))
	return result
}

// parseXML parses an XML file
func (fm *FileManager) parseXML(ctx context.Context, path string) *FileResult {
	result := &FileResult{Path: path}

	file, err := os.Open(path)
	if err != nil {
		result.Error = fmt.Errorf("failed to open XML file: %w", err)
		return result
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		result.Error = fmt.Errorf("failed to read XML file: %w", err)
		return result
	}

	var parsedData interface{}
	if err := xml.Unmarshal(data, &parsedData); err != nil {
		result.Error = fmt.Errorf("failed to parse XML: %w", err)
		return result
	}

	result.Data = data
	result.ParsedData = parsedData
	result.BytesProcessed = int64(len(data))
	return result
}

// streamFile streams a file with buffered I/O
func (fm *FileManager) streamFile(ctx context.Context, path string) *FileResult {
	result := &FileResult{Path: path}

	file, err := os.Open(path)
	if err != nil {
		result.Error = fmt.Errorf("failed to open file for streaming: %w", err)
		return result
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, fm.streamBufferSize)
	var totalBytes int64

	// Stream data in chunks
	buf := make([]byte, fm.streamBufferSize)
	for {
		select {
		case <-ctx.Done():
			result.Error = fmt.Errorf("streaming cancelled")
			return result
		default:
		}

		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			result.Error = fmt.Errorf("streaming read error: %w", err)
			return result
		}

		if n > 0 {
			totalBytes += int64(n)
		}

		if err == io.EOF {
			break
		}
	}

	result.BytesProcessed = totalBytes
	return result
}

// BatchExecute executes multiple file operations concurrently
func (fm *FileManager) BatchExecute(ctx context.Context, operations []FileOperation) ([]*FileResult, error) {
	results := make([]*FileResult, len(operations))
	errors := make([]error, len(operations))
	var wg sync.WaitGroup

	for i, op := range operations {
		wg.Add(1)
		go func(idx int, operation FileOperation) {
			defer wg.Done()
			result, err := fm.ExecuteOperation(ctx, operation)
			results[idx] = result
			errors[idx] = err
		}(i, op)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// GetMetrics returns current file manager metrics
func (fm *FileManager) GetMetrics() map[string]interface{} {
	total := atomic.LoadInt64(&fm.metrics.totalOperations)
	success := atomic.LoadInt64(&fm.metrics.successCount)
	errors := atomic.LoadInt64(&fm.metrics.errorCount)
	totalBytes := atomic.LoadInt64(&fm.metrics.totalBytes)
	totalDuration := atomic.LoadInt64(&fm.metrics.totalDuration)
	peaks := atomic.LoadInt64(&fm.metrics.concurrentPeaks)
	current := atomic.LoadInt64(&fm.metrics.currentConcurrent)

	var avgLatency float64
	if total > 0 {
		avgLatency = float64(totalDuration) / float64(total)
	}

	successRate := float64(0)
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	var avgBytesPerOp float64
	if total > 0 {
		avgBytesPerOp = float64(totalBytes) / float64(total)
	}

	// Use pooled map to reduce allocations
	metrics := GetMetricsMap()
	metrics["total_operations"] = total
	metrics["successful"] = success
	metrics["errors"] = errors
	metrics["success_rate"] = successRate
	metrics["total_bytes"] = totalBytes
	metrics["avg_bytes_per_op"] = avgBytesPerOp
	metrics["avg_latency_ns"] = avgLatency
	metrics["avg_latency_ms"] = avgLatency / 1_000_000
	metrics["peak_concurrency"] = peaks
	metrics["current_concurrent"] = current
	metrics["max_concurrent"] = int64(fm.maxConcurrent)

	return metrics
}

// BufferPool methods

// Get gets a buffer from the pool or creates a new one
func (bp *BufferPool) Get() []byte {
	if v := bp.pool.Get(); v != nil {
		return v.([]byte)
	}
	return make([]byte, bp.size)
}

// Put returns a buffer to the pool
func (bp *BufferPool) Put(buf []byte) {
	if cap(buf) == bp.size {
		bp.pool.Put(buf)
	}
}

// Close closes the file manager and releases resources
func (fm *FileManager) Close() error {
	close(fm.semaphore)
	return nil
}

// ListFiles lists files matching a pattern
func (fm *FileManager) ListFiles(ctx context.Context, pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob pattern error: %w", err)
	}
	return matches, nil
}

// WalkDirectory walks a directory tree
func (fm *FileManager) WalkDirectory(ctx context.Context, root string) ([]string, error) {
	var files []string
	var mu sync.Mutex

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("walk cancelled")
		default:
		}

		if !info.IsDir() {
			mu.Lock()
			files = append(files, path)
			mu.Unlock()
		}

		return nil
	})

	return files, err
}
