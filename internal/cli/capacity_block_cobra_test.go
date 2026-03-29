// Package cli — unit tests for capacity-block commands (list/reserve/show/cancel).
// Uses httptest.NewServer because capacityBlockClient() requires a real *HTTPClient.
package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiclient "github.com/scttfrdmn/prism/pkg/api/client"
)

func capBlockTestApp(t *testing.T) (*App, *httptest.Server) {
	t.Helper()
	t.Setenv("PRISM_NO_AUTO_START", "1")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/v1/capacity-blocks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id": "cr-abc123", "instance_type": "p3.8xlarge", "instance_count": 2,
					"availability_zone": "us-west-2a", "start_time": "2026-04-01T09:00:00Z",
					"end_time": "2026-04-01T17:00:00Z", "duration_hours": 8,
					"state": "active", "total_cost": 64.0,
				},
			})
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "cr-new999", "instance_type": "p3.2xlarge", "instance_count": 1,
				"availability_zone": "us-west-2b", "start_time": "2026-05-01T00:00:00Z",
				"end_time": "2026-05-01T04:00:00Z", "duration_hours": 4,
				"state": "payment-pending", "total_cost": 16.0,
			})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/capacity-blocks/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "cr-abc123", "instance_type": "p3.8xlarge", "instance_count": 2,
				"state": "active", "total_cost": 64.0,
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	hc := apiclient.NewClientWithOptions(srv.URL, apiclient.Options{})
	app := NewAppWithClient("1.0.0", hc)
	return app, srv
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestCapacityBlockCobraCommands_List(t *testing.T) {
	app, _ := capBlockTestApp(t)
	cbc := NewCapacityBlockCobraCommands(app)
	cmd := cbc.CreateCapacityBlockCommand()

	listCmd, _, _ := cmd.Find([]string{"list"})
	require.NotNil(t, listCmd)

	out := captureStdout(t, func() {
		err := listCmd.RunE(listCmd, []string{})
		require.NoError(t, err)
	})

	assert.Contains(t, out, "cr-abc123")
	assert.Contains(t, out, "p3.8xlarge")
	assert.Contains(t, out, "active")
}

func TestCapacityBlockCobraCommands_Show(t *testing.T) {
	app, _ := capBlockTestApp(t)
	cbc := NewCapacityBlockCobraCommands(app)
	cmd := cbc.CreateCapacityBlockCommand()

	showCmd, _, _ := cmd.Find([]string{"show"})
	require.NotNil(t, showCmd)

	out := captureStdout(t, func() {
		err := showCmd.RunE(showCmd, []string{"cr-abc123"})
		require.NoError(t, err)
	})

	assert.Contains(t, out, "cr-abc123")
	assert.Contains(t, out, "p3.8xlarge")
}

func TestCapacityBlockCobraCommands_Reserve(t *testing.T) {
	app, _ := capBlockTestApp(t)
	cbc := NewCapacityBlockCobraCommands(app)
	cmd := cbc.CreateCapacityBlockCommand()

	reserveCmd, _, _ := cmd.Find([]string{"reserve"})
	require.NotNil(t, reserveCmd)

	// Set required flags
	reserveCmd.Flags().Set("type", "p3.2xlarge")
	reserveCmd.Flags().Set("count", "1")
	reserveCmd.Flags().Set("start", "2026-05-01T00:00:00Z")
	reserveCmd.Flags().Set("hours", "4")

	out := captureStdout(t, func() {
		err := reserveCmd.RunE(reserveCmd, []string{})
		require.NoError(t, err)
	})

	assert.True(t, strings.Contains(out, "cr-new999") || strings.Contains(out, "payment-pending"))
}

func TestCapacityBlockCobraCommands_Cancel(t *testing.T) {
	app, _ := capBlockTestApp(t)
	cbc := NewCapacityBlockCobraCommands(app)
	cmd := cbc.CreateCapacityBlockCommand()

	cancelCmd, _, _ := cmd.Find([]string{"cancel"})
	require.NotNil(t, cancelCmd)

	out := captureStdout(t, func() {
		err := cancelCmd.RunE(cancelCmd, []string{"cr-abc123"})
		require.NoError(t, err)
	})

	assert.Contains(t, out, "cr-abc123")
}
