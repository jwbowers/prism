// Package cli — unit tests for file-ops commands (push/pull/list).
// Uses httptest.NewServer because fileOpsClient() requires a real *HTTPClient.
package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiclient "github.com/scttfrdmn/prism/pkg/api/client"
)

func fileOpsTestApp(t *testing.T) (*App, *httptest.Server) {
	t.Helper()
	t.Setenv("PRISM_NO_AUTO_START", "1")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// List files
	mux.HandleFunc("/api/v1/instances/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && contains(r.URL.Path, "/files"):
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"path": "/home/ec2-user/data.csv", "size_bytes": 1024, "is_dir": false, "modified_at": "2026-01-01T00:00:00Z", "permissions": "-rw-r--r--"},
				{"path": "/home/ec2-user/models", "size_bytes": 0, "is_dir": true, "modified_at": "2026-01-01T00:00:00Z", "permissions": "drwxr-xr-x"},
			})
		case r.Method == http.MethodPost && contains(r.URL.Path, "/files/push"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"path": "/home/ec2-user/upload.txt", "status": "ok", "size_bytes": 42,
			})
		case r.Method == http.MethodPost && contains(r.URL.Path, "/files/pull"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"path": "/tmp/download.txt", "status": "ok", "size_bytes": 42,
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	hc := apiclient.NewClientWithOptions(srv.URL, apiclient.Options{})
	app := NewAppWithClient("1.0.0", hc)
	return app, srv
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestFileOpsCobraCommands_List(t *testing.T) {
	app, _ := fileOpsTestApp(t)
	fc := NewFileOpsCobraCommands(app)
	cmd := fc.CreateFilesCommand()

	listCmd, _, err2 := cmd.Find([]string{"list"})
	require.NoError(t, err2)
	require.NotNil(t, listCmd)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listCmd.RunE(listCmd, []string{"my-instance"})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "data.csv")
	assert.Contains(t, out, "models")
}

func TestFileOpsCobraCommands_Push(t *testing.T) {
	app, _ := fileOpsTestApp(t)
	fc := NewFileOpsCobraCommands(app)
	cmd := fc.CreateFilesCommand()

	// Create a temp local file to push
	tmp := filepath.Join(t.TempDir(), "upload.txt")
	require.NoError(t, os.WriteFile(tmp, []byte("hello"), 0644))

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	pushCmd, _, _ := cmd.Find([]string{"push"})
	require.NotNil(t, pushCmd)
	err := pushCmd.RunE(pushCmd, []string{"my-instance", tmp, "/home/ec2-user/upload.txt"})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "ok")
}

func TestFileOpsCobraCommands_Pull(t *testing.T) {
	app, _ := fileOpsTestApp(t)
	fc := NewFileOpsCobraCommands(app)
	cmd := fc.CreateFilesCommand()

	localDest := filepath.Join(t.TempDir(), "download.txt")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	pullCmd, _, _ := cmd.Find([]string{"pull"})
	require.NotNil(t, pullCmd)
	err := pullCmd.RunE(pullCmd, []string{"my-instance", "/home/ec2-user/data.csv", localDest})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "ok")
}
