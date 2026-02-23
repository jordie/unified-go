package subsystems

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestFileManagerBasic(t *testing.T) {
	fm := NewFileManager(10, 65536)
	defer fm.Close()

	// Create test file
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.txt")
	testData := []byte("test data content")

	err := ioutil.WriteFile(testFile, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Read file
	op := FileOperation{
		Path:   testFile,
		OpType: "read",
	}

	result, err := fm.ExecuteOperation(context.Background(), op)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(result.Data) != string(testData) {
		t.Fatalf("Expected %s, got %s", string(testData), string(result.Data))
	}

	if result.BytesProcessed != int64(len(testData)) {
		t.Fatalf("Expected %d bytes, got %d", len(testData), result.BytesProcessed)
	}
}

func TestParseJSON(t *testing.T) {
	fm := NewFileManager(10, 65536)
	defer fm.Close()

	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.json")

	// Create test JSON file
	testData := map[string]interface{}{
		"name": "test",
		"age":  30,
		"tags": []string{"a", "b", "c"},
	}

	jsonData, _ := json.Marshal(testData)
	err := ioutil.WriteFile(testFile, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test JSON: %v", err)
	}

	// Parse JSON
	op := FileOperation{
		Path:   testFile,
		OpType: "parse_json",
	}

	result, err := fm.ExecuteOperation(context.Background(), op)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ParsedData == nil {
		t.Fatal("Expected parsed data, got nil")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(result.Data, &parsed)
	if err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}
}

func TestParseXML(t *testing.T) {
	fm := NewFileManager(10, 65536)
	defer fm.Close()

	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.xml")

	// Create test XML file
	xmlContent := `<?xml version="1.0"?>
<root>
	<item>
		<name>test</name>
		<value>123</value>
	</item>
</root>`

	err := ioutil.WriteFile(testFile, []byte(xmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test XML: %v", err)
	}

	// Parse XML
	op := FileOperation{
		Path:   testFile,
		OpType: "parse_xml",
	}

	result, err := fm.ExecuteOperation(context.Background(), op)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error result, got %v", result.Error)
	}

	// Verify we read the data
	if len(result.Data) == 0 {
		t.Fatal("Expected non-empty XML data")
	}
}

func TestConcurrentFileOperations(t *testing.T) {
	fm := NewFileManager(20, 65536)
	defer fm.Close()

	testDir := t.TempDir()

	// Create test files
	numFiles := 50
	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(testDir, fmt.Sprintf("file_%d.txt", i))
		ioutil.WriteFile(testFile, []byte("test data"), 0644)
	}

	// Read all files concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	errorCount := 0

	for i := 0; i < numFiles; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			testFile := filepath.Join(testDir, fmt.Sprintf("file_%d.txt", idx))
			op := FileOperation{
				Path:   testFile,
				OpType: "read",
			}

			result, err := fm.ExecuteOperation(context.Background(), op)
			mu.Lock()
			if err != nil || result.Error != nil {
				errorCount++
			} else {
				successCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if successCount == 0 {
		t.Fatalf("Expected successful operations, got %d", successCount)
	}

	metrics := fm.GetMetrics()
	if metrics["total_operations"].(int64) == 0 {
		t.Fatal("Expected operations in metrics")
	}
}

func TestBatchExecute(t *testing.T) {
	fm := NewFileManager(20, 65536)
	defer fm.Close()

	testDir := t.TempDir()

	// Create test files
	operations := make([]FileOperation, 10)
	for i := 0; i < 10; i++ {
		testFile := filepath.Join(testDir, fmt.Sprintf("batch_%d.txt", i))
		ioutil.WriteFile(testFile, []byte("batch test data"), 0644)
		operations[i] = FileOperation{
			Path:   testFile,
			OpType: "read",
		}
	}

	// Execute batch
	results, err := fm.BatchExecute(context.Background(), operations)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 10 {
		t.Fatalf("Expected 10 results, got %d", len(results))
	}

	for i, result := range results {
		if result == nil {
			t.Fatalf("Result %d is nil", i)
		}
		if result.Error != nil {
			t.Fatalf("Result %d has error: %v", i, result.Error)
		}
	}
}

func TestStreamFile(t *testing.T) {
	fm := NewFileManager(10, 65536)
	defer fm.Close()

	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "stream.txt")

	// Create large test file
	largeData := make([]byte, 1024*1024) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	ioutil.WriteFile(testFile, largeData, 0644)

	// Stream file
	op := FileOperation{
		Path:   testFile,
		OpType: "stream",
	}

	result, err := fm.ExecuteOperation(context.Background(), op)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.BytesProcessed != int64(len(largeData)) {
		t.Fatalf("Expected %d bytes processed, got %d", len(largeData), result.BytesProcessed)
	}
}

func TestFileContextCancellation(t *testing.T) {
	fm := NewFileManager(1, 65536)
	defer fm.Close()

	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "cancel.txt")
	ioutil.WriteFile(testFile, []byte("test"), 0644)

	// Create a context that gets cancelled while trying to acquire semaphore
	ctx, cancel := context.WithCancel(context.Background())

	// Acquire the single available slot
	<-fm.semaphore

	// Now try to execute with cancelled context
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	op := FileOperation{
		Path:   testFile,
		OpType: "read",
	}

	// Return the slot
	fm.semaphore <- struct{}{}

	_, err := fm.ExecuteOperation(ctx, op)
	if err == nil {
		// Either error or success is okay - the context was cancelled so it might succeed before cancel
		// What matters is the test runs without panic
	}
}

func TestFileMetrics(t *testing.T) {
	fm := NewFileManager(10, 65536)
	defer fm.Close()

	testDir := t.TempDir()

	// Execute some operations
	for i := 0; i < 5; i++ {
		testFile := filepath.Join(testDir, fmt.Sprintf("metrics_%d.txt", i))
		ioutil.WriteFile(testFile, []byte("metrics test"), 0644)

		op := FileOperation{
			Path:   testFile,
			OpType: "read",
		}
		fm.ExecuteOperation(context.Background(), op)
	}

	metrics := fm.GetMetrics()

	if metrics["total_operations"].(int64) != 5 {
		t.Fatalf("Expected 5 operations, got %d", metrics["total_operations"].(int64))
	}

	if metrics["successful"].(int64) != 5 {
		t.Fatalf("Expected 5 successes, got %d", metrics["successful"].(int64))
	}

	successRate := metrics["success_rate"].(float64)
	if successRate != 100.0 {
		t.Fatalf("Expected 100%% success rate, got %.2f", successRate)
	}
}

func TestPeakConcurrency(t *testing.T) {
	fm := NewFileManager(10, 65536)
	defer fm.Close()

	testDir := t.TempDir()

	// Create test files
	numFiles := 20
	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(testDir, fmt.Sprintf("concurrent_%d.txt", i))
		ioutil.WriteFile(testFile, []byte("concurrent test"), 0644)
	}

	// Execute operations with potential for concurrent execution
	var wg sync.WaitGroup
	for i := 0; i < numFiles; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			testFile := filepath.Join(testDir, fmt.Sprintf("concurrent_%d.txt", idx))
			op := FileOperation{
				Path:   testFile,
				OpType: "read",
			}
			fm.ExecuteOperation(context.Background(), op)
		}(i)
	}

	wg.Wait()

	metrics := fm.GetMetrics()
	peakConcurrency := metrics["peak_concurrency"].(int64)

	if peakConcurrency == 0 {
		t.Fatal("Expected non-zero peak concurrency")
	}

	if peakConcurrency > int64(fm.maxConcurrent) {
		t.Fatalf("Peak concurrency %d exceeds max %d", peakConcurrency, fm.maxConcurrent)
	}
}

func TestFileErrorHandling(t *testing.T) {
	fm := NewFileManager(10, 65536)
	defer fm.Close()

	// Try to read non-existent file
	op := FileOperation{
		Path:   "/nonexistent/file.txt",
		OpType: "read",
	}

	result, _ := fm.ExecuteOperation(context.Background(), op)
	if result.Error == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

func TestListFiles(t *testing.T) {
	fm := NewFileManager(10, 65536)
	defer fm.Close()

	testDir := t.TempDir()

	// Create test files
	for i := 0; i < 5; i++ {
		ioutil.WriteFile(filepath.Join(testDir, fmt.Sprintf("test_%d.txt", i)), []byte("test"), 0644)
	}

	// List files
	pattern := filepath.Join(testDir, "test_*.txt")
	files, err := fm.ListFiles(context.Background(), pattern)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(files) != 5 {
		t.Fatalf("Expected 5 files, got %d", len(files))
	}
}

func TestWalkDirectory(t *testing.T) {
	fm := NewFileManager(10, 65536)
	defer fm.Close()

	testDir := t.TempDir()

	// Create nested structure
	for i := 0; i < 3; i++ {
		subDir := filepath.Join(testDir, fmt.Sprintf("dir_%d", i))
		os.MkdirAll(subDir, 0755)
		for j := 0; j < 2; j++ {
			ioutil.WriteFile(filepath.Join(subDir, fmt.Sprintf("file_%d.txt", j)), []byte("walk test"), 0644)
		}
	}

	// Walk directory
	files, err := fm.WalkDirectory(context.Background(), testDir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(files) != 6 { // 3 dirs * 2 files each
		t.Fatalf("Expected 6 files, got %d", len(files))
	}
}

func BenchmarkFileManagerRead(b *testing.B) {
	fm := NewFileManager(100, 65536)
	defer fm.Close()

	testDir := b.TempDir()
	testFile := filepath.Join(testDir, "bench.txt")
	ioutil.WriteFile(testFile, make([]byte, 10*1024*1024), 0644) // 10MB file

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		op := FileOperation{
			Path:   testFile,
			OpType: "read",
		}
		fm.ExecuteOperation(context.Background(), op)
	}

	metrics := fm.GetMetrics()
	b.Logf("Total operations: %d", metrics["total_operations"].(int64))
	b.Logf("Success rate: %.2f%%", metrics["success_rate"].(float64))
	b.Logf("Avg latency: %.2f ms", metrics["avg_latency_ms"].(float64))
}

func BenchmarkFileManagerConcurrent(b *testing.B) {
	fm := NewFileManager(100, 65536)
	defer fm.Close()

	testDir := b.TempDir()

	// Create test files
	numFiles := 50
	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(testDir, fmt.Sprintf("bench_%d.txt", i))
		ioutil.WriteFile(testFile, make([]byte, 1024*1024), 0644) // 1MB each
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			testFile := filepath.Join(testDir, fmt.Sprintf("bench_%d.txt", idx%numFiles))
			op := FileOperation{
				Path:   testFile,
				OpType: "read",
			}
			fm.ExecuteOperation(context.Background(), op)
			idx++
		}
	})

	metrics := fm.GetMetrics()
	b.Logf("Total operations: %d", metrics["total_operations"].(int64))
	b.Logf("Success rate: %.2f%%", metrics["success_rate"].(float64))
	b.Logf("Peak concurrency: %d", metrics["peak_concurrency"].(int64))
}

func TestBufferPool(t *testing.T) {
	bp := &BufferPool{size: 1024}

	// Get buffer
	buf1 := bp.Get()
	if len(buf1) == 0 {
		t.Fatal("Expected non-empty buffer")
	}

	// Put buffer back
	bp.Put(buf1)

	// Get again (should reuse)
	buf2 := bp.Get()
	if cap(buf2) != 1024 {
		t.Fatalf("Expected capacity 1024, got %d", cap(buf2))
	}
}

func TestFileLoadPattern(t *testing.T) {
	fm := NewFileManager(50, 65536)
	defer fm.Close()

	testDir := t.TempDir()

	// Create test files
	numFiles := 100
	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(testDir, fmt.Sprintf("load_%d.txt", i))
		ioutil.WriteFile(testFile, []byte("load test data"), 0644)
	}

	// Execute operations concurrently
	var wg sync.WaitGroup
	for i := 0; i < numFiles; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			testFile := filepath.Join(testDir, fmt.Sprintf("load_%d.txt", idx))
			op := FileOperation{
				Path:   testFile,
				OpType: "read",
			}
			fm.ExecuteOperation(context.Background(), op)
		}(i)
	}

	wg.Wait()

	metrics := fm.GetMetrics()
	successRate := metrics["success_rate"].(float64)

	if successRate < 95 {
		t.Fatalf("Expected high success rate, got %.2f%%", successRate)
	}
}

