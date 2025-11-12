package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// MockS3Client provides a mock S3 client for testing
type MockS3Client struct {
	s3.Client
	// Track calls
	HeadObjectCalls   int
	PutObjectCalls    int
	GetObjectCalls    int
	DeleteObjectCalls int

	// Mock responses
	HeadObjectResponse *s3.HeadObjectOutput
	HeadObjectError    error
	DeleteObjectError  error
}

func (m *MockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	m.HeadObjectCalls++
	if m.HeadObjectError != nil {
		return nil, m.HeadObjectError
	}
	if m.HeadObjectResponse != nil {
		return m.HeadObjectResponse, nil
	}
	// Default response
	contentLength := int64(1024)
	return &s3.HeadObjectOutput{
		ContentLength: &contentLength,
	}, nil
}

func (m *MockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.DeleteObjectCalls++
	if m.DeleteObjectError != nil {
		return nil, m.DeleteObjectError
	}
	return &s3.DeleteObjectOutput{}, nil
}

func TestDefaultTransferOptions(t *testing.T) {
	opts := DefaultTransferOptions()

	if opts.PartSize != DefaultPartSize {
		t.Errorf("Expected PartSize %d, got %d", DefaultPartSize, opts.PartSize)
	}

	if opts.Concurrency != DefaultConcurrency {
		t.Errorf("Expected Concurrency %d, got %d", DefaultConcurrency, opts.Concurrency)
	}

	if !opts.Checksum {
		t.Error("Expected Checksum to be enabled by default")
	}

	if !opts.ResumeSupport {
		t.Error("Expected ResumeSupport to be enabled by default")
	}

	if opts.AutoCleanup {
		t.Error("Expected AutoCleanup to be disabled by default")
	}

	if opts.ProgressInterval != 1*time.Second {
		t.Errorf("Expected ProgressInterval 1s, got %v", opts.ProgressInterval)
	}
}

func TestNewTransferManager(t *testing.T) {
	mockClient := &MockS3Client{}
	s3Client := &s3.Client{}

	// Test with default options
	tm := NewTransferManager(s3Client, nil)
	if tm == nil {
		t.Fatal("Expected non-nil TransferManager")
	}

	if tm.s3Client != s3Client {
		t.Error("S3 client not set correctly")
	}

	if tm.uploader == nil {
		t.Error("Uploader not initialized")
	}

	if tm.downloader == nil {
		t.Error("Downloader not initialized")
	}

	if tm.transfers == nil {
		t.Error("Transfers map not initialized")
	}

	if tm.options == nil {
		t.Error("Options not initialized")
	}

	// Test with custom options
	customOpts := &TransferOptions{
		PartSize:    20 * 1024 * 1024,
		Concurrency: 10,
		Checksum:    false,
	}
	tm2 := NewTransferManager(s3Client, customOpts)
	if tm2.options.PartSize != customOpts.PartSize {
		t.Errorf("Expected PartSize %d, got %d", customOpts.PartSize, tm2.options.PartSize)
	}

	_ = mockClient // Use mockClient to avoid unused variable warning
}

func TestTransferProgress(t *testing.T) {
	progress := &TransferProgress{
		TransferID:       "test-transfer-1",
		Type:             TransferTypeUpload,
		Status:           TransferStatusInProgress,
		FilePath:         "/tmp/test.txt",
		S3Bucket:         "test-bucket",
		S3Key:            "test-key",
		TotalBytes:       1000,
		TransferredBytes: 500,
		PercentComplete:  50.0,
		StartTime:        time.Now(),
	}

	if progress.TransferID != "test-transfer-1" {
		t.Errorf("Expected TransferID 'test-transfer-1', got '%s'", progress.TransferID)
	}

	if progress.Type != TransferTypeUpload {
		t.Errorf("Expected Type %s, got %s", TransferTypeUpload, progress.Type)
	}

	if progress.Status != TransferStatusInProgress {
		t.Errorf("Expected Status %s, got %s", TransferStatusInProgress, progress.Status)
	}

	if progress.PercentComplete != 50.0 {
		t.Errorf("Expected PercentComplete 50.0, got %f", progress.PercentComplete)
	}
}

func TestGetTransferProgress(t *testing.T) {
	tm := NewTransferManager(&s3.Client{}, nil)

	// Create and register a transfer
	progress := &TransferProgress{
		TransferID: "transfer-123",
		Status:     TransferStatusInProgress,
	}

	tm.mu.Lock()
	tm.transfers["transfer-123"] = progress
	tm.mu.Unlock()

	// Test getting existing transfer
	retrieved, exists := tm.GetTransferProgress("transfer-123")
	if !exists {
		t.Error("Expected transfer to exist")
	}

	if retrieved.TransferID != "transfer-123" {
		t.Errorf("Expected TransferID 'transfer-123', got '%s'", retrieved.TransferID)
	}

	// Test getting non-existent transfer
	_, exists = tm.GetTransferProgress("nonexistent")
	if exists {
		t.Error("Expected transfer to not exist")
	}
}

func TestListTransfers(t *testing.T) {
	tm := NewTransferManager(&s3.Client{}, nil)

	// Add multiple transfers
	transfers := []*TransferProgress{
		{TransferID: "transfer-1", Status: TransferStatusInProgress},
		{TransferID: "transfer-2", Status: TransferStatusCompleted},
		{TransferID: "transfer-3", Status: TransferStatusPending},
	}

	tm.mu.Lock()
	for _, p := range transfers {
		tm.transfers[p.TransferID] = p
	}
	tm.mu.Unlock()

	// Test listing
	list := tm.ListTransfers()
	if len(list) != 3 {
		t.Errorf("Expected 3 transfers, got %d", len(list))
	}

	// Verify all transfers are in the list
	foundIDs := make(map[string]bool)
	for _, p := range list {
		foundIDs[p.TransferID] = true
	}

	for _, expected := range transfers {
		if !foundIDs[expected.TransferID] {
			t.Errorf("Transfer %s not found in list", expected.TransferID)
		}
	}
}

func TestDeleteObject(t *testing.T) {
	// Note: This test would require full S3 mocking
	// For now, just test the interface exists
	tm := NewTransferManager(&s3.Client{}, nil)

	ctx := context.Background()
	err := tm.DeleteObject(ctx, "test-bucket", "test-key")

	// We expect an error because we're using a nil client config
	// Just verify the method is callable
	_ = err // Error is expected with nil config
}

func TestComputeFileMD5(t *testing.T) {
	// Create a temporary test file
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "test-md5.txt")

	content := []byte("Hello, World!")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Compute checksum
	checksum, err := computeFileMD5(testFile)
	if err != nil {
		t.Fatalf("computeFileMD5 failed: %v", err)
	}

	// Verify checksum is valid hex string
	if len(checksum) != 32 {
		t.Errorf("Expected MD5 checksum length 32, got %d", len(checksum))
	}

	// Known MD5 of "Hello, World!" is 65a8e27d8879283831b664bd8b7f0ad4
	expected := "65a8e27d8879283831b664bd8b7f0ad4"
	if checksum != expected {
		t.Errorf("Expected checksum %s, got %s", expected, checksum)
	}

	// Test with nonexistent file
	_, err = computeFileMD5("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestVerifyChecksum(t *testing.T) {
	tests := []struct {
		name     string
		etag     string
		checksum string
		expected bool
	}{
		{
			name:     "matching checksums",
			etag:     "abc123",
			checksum: "abc123",
			expected: true,
		},
		{
			name:     "matching checksums with quotes",
			etag:     "\"abc123\"",
			checksum: "abc123",
			expected: true,
		},
		{
			name:     "non-matching checksums",
			etag:     "abc123",
			checksum: "xyz789",
			expected: false,
		},
		{
			name:     "multipart upload etag (skip verification)",
			etag:     "d41d8cd98f00b204e9800998ecf8427e-2", // 32 char MD5 + dash + part count = 35 chars
			checksum: "anything",
			expected: true, // Multipart ETags are skipped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifyChecksum(tt.etag, tt.checksum)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for etag=%s checksum=%s",
					tt.expected, result, tt.etag, tt.checksum)
			}
		})
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "\"hello\"", expected: "hello"},
		{input: "hello", expected: "hello"},
		{input: "\"", expected: "\""},
		{input: "", expected: ""},
		{input: "\"hello", expected: "\"hello"},
		{input: "hello\"", expected: "hello\""},
	}

	for _, tt := range tests {
		result := trimQuotes(tt.input)
		if result != tt.expected {
			t.Errorf("trimQuotes(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestGenerateTransferID(t *testing.T) {
	id1 := generateTransferID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamp
	id2 := generateTransferID()

	if id1 == id2 {
		t.Error("Expected unique transfer IDs")
	}

	if len(id1) == 0 {
		t.Error("Expected non-empty transfer ID")
	}

	// Should start with "transfer-"
	if len(id1) < 9 || id1[:9] != "transfer-" {
		t.Errorf("Expected transfer ID to start with 'transfer-', got %s", id1)
	}
}

func TestTransferStatusConstants(t *testing.T) {
	// Verify all status constants are defined correctly
	statuses := []TransferStatus{
		TransferStatusPending,
		TransferStatusInProgress,
		TransferStatusPaused,
		TransferStatusCompleted,
		TransferStatusFailed,
		TransferStatusCancelled,
	}

	for _, status := range statuses {
		if string(status) == "" {
			t.Errorf("Transfer status constant has empty string value: %v", status)
		}
	}
}

func TestTransferTypeConstants(t *testing.T) {
	// Verify transfer type constants
	if TransferTypeUpload != "upload" {
		t.Errorf("Expected TransferTypeUpload to be 'upload', got '%s'", TransferTypeUpload)
	}

	if TransferTypeDownload != "download" {
		t.Errorf("Expected TransferTypeDownload to be 'download', got '%s'", TransferTypeDownload)
	}
}

func TestProgressReader(t *testing.T) {
	// Create a test progress object
	progress := &TransferProgress{
		TotalBytes:       100,
		TransferredBytes: 0,
		StartTime:        time.Now(),
	}

	// Create mock reader with 100 bytes
	mockData := make([]byte, 100)
	for i := range mockData {
		mockData[i] = byte(i)
	}

	// Track callback invocations
	callbackCount := 0
	callback := func(p *TransferProgress) {
		callbackCount++
	}

	pr := &progressReader{
		reader:   &mockReader{data: mockData},
		progress: progress,
		callback: callback,
		interval: 10 * time.Millisecond,
	}

	// Read data in chunks
	buf := make([]byte, 10)
	for i := 0; i < 10; i++ {
		n, err := pr.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		if n != 10 {
			t.Errorf("Expected to read 10 bytes, got %d", n)
		}
		time.Sleep(15 * time.Millisecond) // Ensure callback fires
	}

	// Verify progress tracking
	if progress.TransferredBytes != 100 {
		t.Errorf("Expected TransferredBytes 100, got %d", progress.TransferredBytes)
	}

	if progress.PercentComplete != 100.0 {
		t.Errorf("Expected PercentComplete 100.0, got %f", progress.PercentComplete)
	}

	if progress.BytesPerSecond == 0 {
		t.Error("Expected BytesPerSecond to be calculated")
	}

	if callbackCount == 0 {
		t.Error("Expected callback to be invoked")
	}
}

// mockReader provides a simple mock for io.Reader
type mockReader struct {
	data   []byte
	offset int
}

func (m *mockReader) Read(p []byte) (int, error) {
	if m.offset >= len(m.data) {
		return 0, os.ErrClosed
	}

	n := copy(p, m.data[m.offset:])
	m.offset += n
	return n, nil
}

func TestProgressWriter(t *testing.T) {
	// Create a test progress object
	progress := &TransferProgress{
		TotalBytes:       100,
		TransferredBytes: 0,
		StartTime:        time.Now(),
	}

	// Create mock writer
	mockData := make([]byte, 100)

	// Track callback invocations
	callbackCount := 0
	callback := func(p *TransferProgress) {
		callbackCount++
	}

	pw := &progressWriter{
		writer:   &mockWriterAt{data: mockData},
		progress: progress,
		callback: callback,
		interval: 10 * time.Millisecond,
	}

	// Write data in chunks
	data := make([]byte, 10)
	for i := 0; i < 10; i++ {
		n, err := pw.WriteAt(data, int64(i*10))
		if err != nil {
			t.Fatalf("WriteAt error: %v", err)
		}
		if n != 10 {
			t.Errorf("Expected to write 10 bytes, got %d", n)
		}
		time.Sleep(15 * time.Millisecond) // Ensure callback fires
	}

	// Verify progress tracking
	if progress.TransferredBytes != 100 {
		t.Errorf("Expected TransferredBytes 100, got %d", progress.TransferredBytes)
	}

	if progress.PercentComplete != 100.0 {
		t.Errorf("Expected PercentComplete 100.0, got %f", progress.PercentComplete)
	}

	if progress.BytesPerSecond == 0 {
		t.Error("Expected BytesPerSecond to be calculated")
	}

	if callbackCount == 0 {
		t.Error("Expected callback to be invoked")
	}
}

// mockWriterAt provides a simple mock for io.WriterAt
type mockWriterAt struct {
	data []byte
}

func (m *mockWriterAt) WriteAt(p []byte, off int64) (int, error) {
	if off+int64(len(p)) > int64(len(m.data)) {
		return 0, os.ErrInvalid
	}

	return copy(m.data[off:], p), nil
}

func TestTransferManagerConcurrency(t *testing.T) {
	tm := NewTransferManager(&s3.Client{}, nil)

	// Test concurrent access without race conditions
	// Use unique IDs based on goroutine index to avoid timestamp collisions
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			// Create unique transfer ID using goroutine index
			transferID := fmt.Sprintf("transfer-test-%d", id)

			progress := &TransferProgress{
				TransferID: transferID,
				Status:     TransferStatusPending,
			}

			// Concurrent write to transfers map
			tm.mu.Lock()
			tm.transfers[progress.TransferID] = progress
			tm.mu.Unlock()

			// Concurrent read from transfers map
			_, exists := tm.GetTransferProgress(transferID)
			if !exists {
				t.Errorf("Failed to retrieve transfer %s", transferID)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all transfers were added
	list := tm.ListTransfers()
	if len(list) != 10 {
		t.Errorf("Expected 10 transfers, got %d", len(list))
	}
}

func TestTransferOptionsValidation(t *testing.T) {
	opts := &TransferOptions{
		PartSize:    MinPartSize - 1, // Below minimum
		Concurrency: -1,              // Invalid
	}

	// Create transfer manager (should handle invalid values)
	tm := NewTransferManager(&s3.Client{}, opts)
	if tm == nil {
		t.Fatal("Expected non-nil TransferManager even with invalid options")
	}

	// AWS SDK will handle validation of PartSize and Concurrency
	// Just verify manager was created
}
