package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scttfrdmn/prism/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleStorageTransfer_Routing tests method routing on the collection endpoint.
func TestHandleStorageTransfer_Routing(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	unsupportedMethods := []string{
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}
	for _, method := range unsupportedMethods {
		t.Run("method_"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/storage/transfer", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

// TestHandleListTransfers tests GET /api/v1/storage/transfer.
func TestHandleListTransfers(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/storage/transfer", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should return 200 with an empty (or non-nil) list
	assert.Equal(t, http.StatusOK, w.Code)

	// Response should be valid JSON (array or null)
	body := w.Body.Bytes()
	assert.True(t, len(body) > 0, "response body should not be empty")

	// Should be parseable as an array (possibly nil/empty)
	var transfers interface{}
	err := json.Unmarshal(body, &transfers)
	assert.NoError(t, err, "response should be valid JSON")
}

// TestHandleStartTransfer_InvalidType tests that an unsupported transfer type returns 400.
func TestHandleStartTransfer_InvalidType(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := TransferRequest{
		Type:      "sync", // invalid
		LocalPath: "/tmp/file.txt",
		S3Bucket:  "my-bucket",
		S3Key:     "my-key",
	}
	body, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/storage/transfer", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid transfer type")
}

// TestHandleStartTransfer_MissingLocalPath tests that an empty local_path returns 400.
func TestHandleStartTransfer_MissingLocalPath(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := TransferRequest{
		Type:     string(storage.TransferTypeUpload),
		S3Bucket: "my-bucket",
		S3Key:    "my-key",
		// LocalPath intentionally omitted
	}
	body, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/storage/transfer", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "local_path")
}

// TestHandleStartTransfer_MissingS3Bucket tests that an empty s3_bucket returns 400.
func TestHandleStartTransfer_MissingS3Bucket(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := TransferRequest{
		Type:      string(storage.TransferTypeUpload),
		LocalPath: "/tmp/file.txt",
		S3Key:     "my-key",
		// S3Bucket intentionally omitted
	}
	body, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/storage/transfer", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "s3_bucket")
}

// TestHandleStartTransfer_MissingS3Key tests that an empty s3_key returns 400.
func TestHandleStartTransfer_MissingS3Key(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := TransferRequest{
		Type:      string(storage.TransferTypeDownload),
		LocalPath: "/tmp/file.txt",
		S3Bucket:  "my-bucket",
		// S3Key intentionally omitted
	}
	body, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/storage/transfer", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "s3_key")
}

// TestHandleStartTransfer_InvalidJSON tests that malformed JSON returns 400.
func TestHandleStartTransfer_InvalidJSON(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	r := httptest.NewRequest(http.MethodPost, "/api/v1/storage/transfer",
		bytes.NewBufferString(`{not valid json`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

// TestHandleStorageTransferOperations_Routing tests method routing on individual transfer operations.
func TestHandleStorageTransferOperations_Routing(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	unsupportedMethods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
	}
	for _, method := range unsupportedMethods {
		t.Run("method_"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/storage/transfer/some-id", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

// TestHandleGetTransferStatus_NotFound tests that a nonexistent transfer ID returns 404.
func TestHandleGetTransferStatus_NotFound(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/storage/transfer/nonexistent-id-xyz", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should return 404 for nonexistent transfer (requires transfer manager to initialise OK)
	// In test environments without AWS, getTransferManager may return 500 instead — both are acceptable.
	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError,
		"expected 404 (not found) or 500 (no AWS), got %d", w.Code)
	if w.Code == http.StatusNotFound {
		assert.Contains(t, w.Body.String(), "nonexistent-id-xyz")
	}
}

// TestHandleCancelTransfer_NotFound tests that cancelling a nonexistent transfer returns 404.
func TestHandleCancelTransfer_NotFound(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/storage/transfer/nonexistent-id-xyz", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Same reasoning as TestHandleGetTransferStatus_NotFound.
	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError,
		"expected 404 (not found) or 500 (no AWS), got %d", w.Code)
	if w.Code == http.StatusNotFound {
		assert.Contains(t, w.Body.String(), "nonexistent-id-xyz")
	}
}

// TestTransferRequest_TypeValidation verifies both valid transfer types are accepted by validation logic.
func TestTransferRequest_TypeValidation(t *testing.T) {
	tests := []struct {
		transferType string
		wantBadReq   bool
	}{
		{string(storage.TransferTypeUpload), false},
		{string(storage.TransferTypeDownload), false},
		{"sync", true},
		{"move", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run("type_"+tt.transferType, func(t *testing.T) {
			server := createTestServer(t)
			handler := server.createHTTPHandler()

			req := TransferRequest{
				Type:      tt.transferType,
				LocalPath: "/tmp/file.txt",
				S3Bucket:  "my-bucket",
				S3Key:     "my/key",
			}
			body, err := json.Marshal(req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/api/v1/storage/transfer", bytes.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)

			if tt.wantBadReq {
				assert.Equal(t, http.StatusBadRequest, w.Code,
					"type %q should be rejected with 400", tt.transferType)
			} else {
				// Valid type passes validation — may still fail at AWS layer (500) which is OK for unit tests
				assert.NotEqual(t, http.StatusBadRequest, w.Code,
					"type %q should not be rejected with 400", tt.transferType)
			}
		})
	}
}
