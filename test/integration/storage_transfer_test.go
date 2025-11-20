//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/scttfrdmn/prism/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestS3TransferIntegration tests S3 file transfers against real AWS
// Run with: go test -tags=integration -v ./test/integration -run TestS3TransferIntegration
func TestS3TransferIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check for AWS credentials
	if os.Getenv("AWS_PROFILE") == "" && os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Skip("Skipping S3 transfer integration test - no AWS credentials configured")
	}

	// Get test bucket from environment or use default
	testBucket := os.Getenv("PRISM_TEST_S3_BUCKET")
	if testBucket == "" {
		t.Skip("Skipping S3 transfer integration test - PRISM_TEST_S3_BUCKET not set")
	}

	ctx := context.Background()

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	require.NoError(t, err, "Failed to load AWS config")

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Create transfer manager
	transferMgr := storage.NewTransferManager(s3Client, storage.DefaultTransferOptions())
	require.NotNil(t, transferMgr, "Failed to create transfer manager")

	t.Run("UploadSmallFile", func(t *testing.T) {
		testUploadSmallFile(t, ctx, transferMgr, testBucket)
	})

	t.Run("DownloadSmallFile", func(t *testing.T) {
		testDownloadSmallFile(t, ctx, transferMgr, testBucket)
	})

	t.Run("UploadLargeFile", func(t *testing.T) {
		testUploadLargeFile(t, ctx, transferMgr, testBucket)
	})

	t.Run("DownloadLargeFile", func(t *testing.T) {
		testDownloadLargeFile(t, ctx, transferMgr, testBucket)
	})

	t.Run("ChecksumVerification", func(t *testing.T) {
		testChecksumVerification(t, ctx, transferMgr, testBucket)
	})

	t.Run("ProgressTracking", func(t *testing.T) {
		testProgressTracking(t, ctx, transferMgr, testBucket)
	})

	t.Run("ConcurrentTransfers", func(t *testing.T) {
		testConcurrentTransfers(t, ctx, transferMgr, testBucket)
	})
}

// testUploadSmallFile tests uploading a small file (<5MB)
func testUploadSmallFile(t *testing.T, ctx context.Context, tm *storage.TransferManager, bucket string) {
	// Create test file (1MB)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-small.dat")
	data := make([]byte, 1*1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 256)
	}
	err := os.WriteFile(testFile, data, 0644)
	require.NoError(t, err)

	// Upload file
	key := fmt.Sprintf("prism-test/small-%d.dat", time.Now().Unix())
	progress, err := tm.UploadFile(ctx, testFile, bucket, key)
	require.NoError(t, err, "Upload failed")
	require.NotNil(t, progress)

	// Verify upload completed
	assert.Equal(t, storage.TransferStatusCompleted, progress.Status)
	assert.Equal(t, int64(1*1024*1024), progress.TotalBytes)
	assert.Equal(t, progress.TotalBytes, progress.TransferredBytes)
	assert.Equal(t, 100.0, progress.PercentComplete)
	assert.NotEmpty(t, progress.Checksum)

	// Cleanup
	defer tm.DeleteObject(ctx, bucket, key)

	t.Logf("✅ Small file upload successful: %d bytes in %v", progress.TransferredBytes, time.Since(progress.StartTime))
}

// testDownloadSmallFile tests downloading a small file
func testDownloadSmallFile(t *testing.T, ctx context.Context, tm *storage.TransferManager, bucket string) {
	// Create and upload test file
	tmpDir := t.TempDir()
	uploadFile := filepath.Join(tmpDir, "upload-small.dat")
	data := make([]byte, 1*1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 256)
	}
	err := os.WriteFile(uploadFile, data, 0644)
	require.NoError(t, err)

	key := fmt.Sprintf("prism-test/download-small-%d.dat", time.Now().Unix())
	_, err = tm.UploadFile(ctx, uploadFile, bucket, key)
	require.NoError(t, err)
	defer tm.DeleteObject(ctx, bucket, key)

	// Download file
	downloadFile := filepath.Join(tmpDir, "download-small.dat")
	progress, err := tm.DownloadFile(ctx, bucket, key, downloadFile)
	require.NoError(t, err, "Download failed")
	require.NotNil(t, progress)

	// Verify download completed
	assert.Equal(t, storage.TransferStatusCompleted, progress.Status)
	assert.Equal(t, int64(1*1024*1024), progress.TotalBytes)
	assert.Equal(t, progress.TotalBytes, progress.TransferredBytes)
	assert.Equal(t, 100.0, progress.PercentComplete)

	// Verify file contents match
	downloadedData, err := os.ReadFile(downloadFile)
	require.NoError(t, err)
	assert.Equal(t, data, downloadedData, "Downloaded file content mismatch")

	t.Logf("✅ Small file download successful: %d bytes in %v", progress.TransferredBytes, time.Since(progress.StartTime))
}

// testUploadLargeFile tests uploading a large file (>10MB) to verify multipart
func testUploadLargeFile(t *testing.T, ctx context.Context, tm *storage.TransferManager, bucket string) {
	// Create test file (20MB to trigger multipart)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-large.dat")
	data := make([]byte, 20*1024*1024) // 20MB
	for i := range data {
		data[i] = byte(i % 256)
	}
	err := os.WriteFile(testFile, data, 0644)
	require.NoError(t, err)

	// Upload file
	key := fmt.Sprintf("prism-test/large-%d.dat", time.Now().Unix())
	progress, err := tm.UploadFile(ctx, testFile, bucket, key)
	require.NoError(t, err, "Large file upload failed")
	require.NotNil(t, progress)

	// Verify multipart upload occurred (TotalParts > 1)
	assert.Greater(t, progress.TotalParts, 1, "Expected multipart upload for 20MB file")
	assert.Equal(t, storage.TransferStatusCompleted, progress.Status)
	assert.Equal(t, int64(20*1024*1024), progress.TotalBytes)
	assert.Equal(t, progress.TotalBytes, progress.TransferredBytes)

	// Cleanup
	defer tm.DeleteObject(ctx, bucket, key)

	t.Logf("✅ Large file upload successful: %d bytes, %d parts in %v",
		progress.TransferredBytes, progress.TotalParts, time.Since(progress.StartTime))
}

// testDownloadLargeFile tests downloading a large file
func testDownloadLargeFile(t *testing.T, ctx context.Context, tm *storage.TransferManager, bucket string) {
	// Create and upload large test file
	tmpDir := t.TempDir()
	uploadFile := filepath.Join(tmpDir, "upload-large.dat")
	data := make([]byte, 20*1024*1024) // 20MB
	for i := range data {
		data[i] = byte(i % 256)
	}
	err := os.WriteFile(uploadFile, data, 0644)
	require.NoError(t, err)

	key := fmt.Sprintf("prism-test/download-large-%d.dat", time.Now().Unix())
	_, err = tm.UploadFile(ctx, uploadFile, bucket, key)
	require.NoError(t, err)
	defer tm.DeleteObject(ctx, bucket, key)

	// Download file
	downloadFile := filepath.Join(tmpDir, "download-large.dat")
	progress, err := tm.DownloadFile(ctx, bucket, key, downloadFile)
	require.NoError(t, err, "Large file download failed")

	// Verify download
	assert.Equal(t, storage.TransferStatusCompleted, progress.Status)
	assert.Equal(t, int64(20*1024*1024), progress.TotalBytes)
	assert.Greater(t, progress.TotalParts, 1, "Expected multipart download")

	// Verify file size (full content check would be slow)
	stat, err := os.Stat(downloadFile)
	require.NoError(t, err)
	assert.Equal(t, int64(20*1024*1024), stat.Size())

	t.Logf("✅ Large file download successful: %d bytes, %d parts in %v",
		progress.TransferredBytes, progress.TotalParts, time.Since(progress.StartTime))
}

// testChecksumVerification tests MD5 checksum verification
func testChecksumVerification(t *testing.T, ctx context.Context, tm *storage.TransferManager, bucket string) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-checksum.dat")
	data := []byte("Test data for checksum verification")
	err := os.WriteFile(testFile, data, 0644)
	require.NoError(t, err)

	// Upload with checksum
	key := fmt.Sprintf("prism-test/checksum-%d.dat", time.Now().Unix())
	uploadProgress, err := tm.UploadFile(ctx, testFile, bucket, key)
	require.NoError(t, err)
	require.NotEmpty(t, uploadProgress.Checksum, "Upload should compute checksum")
	defer tm.DeleteObject(ctx, bucket, key)

	// Download with checksum verification
	downloadFile := filepath.Join(tmpDir, "download-checksum.dat")
	downloadProgress, err := tm.DownloadFile(ctx, bucket, key, downloadFile)
	require.NoError(t, err)
	require.NotEmpty(t, downloadProgress.Checksum, "Download should compute checksum")

	// Checksums should match
	assert.Equal(t, uploadProgress.Checksum, downloadProgress.Checksum,
		"Upload and download checksums should match")

	t.Logf("✅ Checksum verification successful: %s", uploadProgress.Checksum)
}

// testProgressTracking tests progress callback functionality
func testProgressTracking(t *testing.T, ctx context.Context, tm *storage.TransferManager, bucket string) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-progress.dat")
	data := make([]byte, 10*1024*1024) // 10MB
	err := os.WriteFile(testFile, data, 0644)
	require.NoError(t, err)

	// Track progress updates
	progressUpdates := make([]*storage.TransferProgress, 0)
	options := storage.DefaultTransferOptions()
	options.ProgressCallback = func(p *storage.TransferProgress) {
		progressUpdates = append(progressUpdates, p)
	}
	options.ProgressInterval = 100 * time.Millisecond

	// Create new AWS config for progress tracking transfer manager
	cfg, err := config.LoadDefaultConfig(ctx)
	require.NoError(t, err)
	s3Client := s3.NewFromConfig(cfg)
	tmWithProgress := storage.NewTransferManager(s3Client, options)

	// Upload with progress tracking
	key := fmt.Sprintf("prism-test/progress-%d.dat", time.Now().Unix())
	_, err = tmWithProgress.UploadFile(ctx, testFile, bucket, key)
	require.NoError(t, err)
	defer tm.DeleteObject(ctx, bucket, key)

	// Verify progress updates occurred
	assert.Greater(t, len(progressUpdates), 0, "Should have progress updates")

	// Verify progress increases monotonically
	for i := 1; i < len(progressUpdates); i++ {
		assert.GreaterOrEqual(t, progressUpdates[i].PercentComplete,
			progressUpdates[i-1].PercentComplete,
			"Progress should increase monotonically")
	}

	t.Logf("✅ Progress tracking successful: %d updates", len(progressUpdates))
}

// testConcurrentTransfers tests multiple simultaneous transfers
func testConcurrentTransfers(t *testing.T, ctx context.Context, tm *storage.TransferManager, bucket string) {
	tmpDir := t.TempDir()

	// Create 5 test files
	numFiles := 5
	files := make([]string, numFiles)
	keys := make([]string, numFiles)

	for i := 0; i < numFiles; i++ {
		files[i] = filepath.Join(tmpDir, fmt.Sprintf("concurrent-%d.dat", i))
		data := make([]byte, 2*1024*1024) // 2MB each
		for j := range data {
			data[j] = byte((i + j) % 256)
		}
		err := os.WriteFile(files[i], data, 0644)
		require.NoError(t, err)
		keys[i] = fmt.Sprintf("prism-test/concurrent-%d-%d.dat", time.Now().Unix(), i)
	}

	// Upload concurrently
	startTime := time.Now()
	results := make(chan error, numFiles)

	for i := 0; i < numFiles; i++ {
		go func(idx int) {
			_, err := tm.UploadFile(ctx, files[idx], bucket, keys[idx])
			results <- err
		}(i)
	}

	// Wait for all uploads
	for i := 0; i < numFiles; i++ {
		err := <-results
		require.NoError(t, err, "Concurrent upload %d failed", i)
	}

	duration := time.Since(startTime)

	// Cleanup
	for _, key := range keys {
		defer tm.DeleteObject(ctx, bucket, key)
	}

	// List transfers to verify tracking
	transfers := tm.ListTransfers()
	assert.GreaterOrEqual(t, len(transfers), numFiles,
		"Should track at least %d transfers", numFiles)

	t.Logf("✅ Concurrent transfers successful: %d files in %v", numFiles, duration)
}

// TestStorageTransferAPI tests the daemon API handlers for storage transfers
func TestStorageTransferAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API integration test in short mode")
	}

	// TODO: Implement daemon API integration tests
	// This should:
	// 1. Start daemon
	// 2. Make HTTP requests to /api/v1/storage/transfer endpoints
	// 3. Verify proper error handling, authentication, etc.
	// 4. Test GET/POST/DELETE operations

	t.Skip("Daemon API integration tests not yet implemented")
}

// TestCLIStorageTransfer tests CLI integration with storage transfer
func TestCLIStorageTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// TODO: Implement CLI integration tests
	// This should:
	// 1. Execute prism CLI commands for file transfers
	// 2. Verify output and error handling
	// 3. Test various command-line flags

	t.Skip("CLI storage transfer integration tests not yet implemented")
}
