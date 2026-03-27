// Package cli — unit tests for WorkshopCobraCommands.
// Uses httptest.NewServer because workshopClient() requires a real *HTTPClient.
package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	apiclient "github.com/scttfrdmn/prism/pkg/api/client"
)

// workshopTestApp creates an App backed by an httptest.Server that serves
// minimal JSON responses for every workshop endpoint.
func workshopTestApp(t *testing.T) (*App, *httptest.Server) {
	t.Helper()
	t.Setenv("PRISM_NO_AUTO_START", "1")

	mux := http.NewServeMux()

	// Ping — required before dispatching workshop commands
	mux.HandleFunc("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Workshops collection
	mux.HandleFunc("/api/v1/workshops", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"workshops": []map[string]interface{}{
					{
						"id":           "ws-001",
						"title":        "NeurIPS Tutorial",
						"status":       "draft",
						"start_time":   "2026-12-08T09:00:00Z",
						"end_time":     "2026-12-08T15:00:00Z",
						"participants": []interface{}{},
					},
				},
			})
		case http.MethodPost:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         "ws-new",
				"title":      "New Workshop",
				"status":     "draft",
				"join_token": "WS-TESTTOKEN",
			})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Workshop configs collection
	mux.HandleFunc("/api/v1/workshops/configs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configs": []map[string]interface{}{
					{
						"name":                   "ml-6h",
						"template":               "pytorch-ml",
						"duration_hours":         6,
						"max_participants":       30,
						"budget_per_participant": 5.0,
					},
				},
			})
		case http.MethodPost:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name":             "saved-config",
				"template":         "pytorch-ml",
				"duration_hours":   6,
				"max_participants": 30,
			})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Create-from-config route
	mux.HandleFunc("/api/v1/workshops/from-config/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         "ws-from-cfg",
			"title":      "Workshop From Config",
			"join_token": "WS-CFGTOKEN",
		})
	})

	// Individual workshop operations
	mux.HandleFunc("/api/v1/workshops/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":     "ws-001",
				"title":  "NeurIPS Tutorial",
				"status": "draft",
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPost:
			// provision / end / etc.
			path := r.URL.Path
			switch {
			case len(path) > 0 && path[len(path)-9:] == "provision":
				json.NewEncoder(w).Encode(map[string]interface{}{
					"provisioned": 3,
					"skipped":     0,
					"errors":      []interface{}{},
				})
			case len(path) > 0 && path[len(path)-3:] == "end":
				json.NewEncoder(w).Encode(map[string]interface{}{
					"stopped": 3,
					"errors":  []interface{}{},
				})
			case len(path) > 0 && path[len(path)-6:] == "config":
				json.NewEncoder(w).Encode(map[string]interface{}{
					"name":             "saved-config",
					"template":         "pytorch-ml",
					"duration_hours":   6,
					"max_participants": 30,
				})
			default:
				json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
			}
		default:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"workshop_id":           "ws-001",
				"title":                 "NeurIPS Tutorial",
				"status":                "draft",
				"total_participants":    3,
				"active_instances":      2,
				"stopped_instances":     1,
				"pending_instances":     0,
				"time_remaining":        "2h 30m",
				"total_spent":           12.50,
				"participants":          []interface{}{},
				"participants_download": []interface{}{},
				"workshop_id_dl":        "ws-001",
			})
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	hc := apiclient.NewClientWithOptions(srv.URL, apiclient.Options{})
	app := NewAppWithClient("1.0.0", hc)
	return app, srv
}

// ── workshop list ─────────────────────────────────────────────────────────────

func TestWorkshopCobraCommands_List(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"list"})
	require.NoError(t, cmd.Execute())
}

func TestWorkshopCobraCommands_List_WithFlags(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"list", "--owner", "organizer1", "--status", "active"})
	require.NoError(t, cmd.Execute())
}

// ── workshop create ───────────────────────────────────────────────────────────

func TestWorkshopCobraCommands_Create(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"create",
		"--title", "NeurIPS DL Tutorial",
		"--template", "pytorch-ml",
		"--owner", "organizer1",
		"--start", "2026-12-08T09:00:00",
		"--end", "2026-12-08T15:00:00",
		"--max-participants", "60",
		"--budget-per-participant", "5.00",
		"--early-access", "24",
		"--description", "A test workshop",
	})
	require.NoError(t, cmd.Execute())
}

func TestWorkshopCobraCommands_Create_MissingRequired(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// Missing --template, --owner, --start, --end
	cmd.SetArgs([]string{"create", "--title", "No Required Flags"})
	require.Error(t, cmd.Execute())
}

// ── workshop show ─────────────────────────────────────────────────────────────

func TestWorkshopCobraCommands_Show(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"show", "ws-001"})
	require.NoError(t, cmd.Execute())
}

// ── workshop delete ───────────────────────────────────────────────────────────

func TestWorkshopCobraCommands_Delete(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"delete", "ws-001"})
	require.NoError(t, cmd.Execute())
}

// ── workshop provision ────────────────────────────────────────────────────────

func TestWorkshopCobraCommands_Provision(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"provision", "ws-001"})
	require.NoError(t, cmd.Execute())
}

// ── workshop dashboard ────────────────────────────────────────────────────────

func TestWorkshopCobraCommands_Dashboard(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"dashboard", "ws-001"})
	require.NoError(t, cmd.Execute())
}

// ── workshop end ──────────────────────────────────────────────────────────────

func TestWorkshopCobraCommands_End(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"end", "ws-001"})
	require.NoError(t, cmd.Execute())
}

// ── workshop download ─────────────────────────────────────────────────────────

func TestWorkshopCobraCommands_Download(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"download", "ws-001"})
	require.NoError(t, cmd.Execute())
}

// ── workshop config ───────────────────────────────────────────────────────────

func TestWorkshopCobraCommands_ConfigList(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"config", "list"})
	require.NoError(t, cmd.Execute())
}

func TestWorkshopCobraCommands_ConfigSave(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"config", "save", "ws-001", "my-config"})
	require.NoError(t, cmd.Execute())
}

func TestWorkshopCobraCommands_ConfigUse(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"config", "use", "ml-6h",
		"--title", "New Run",
		"--start", "2027-01-10T09:00:00",
		"--owner", "organizer1",
	})
	require.NoError(t, cmd.Execute())
}

func TestWorkshopCobraCommands_ConfigUse_MissingRequired(t *testing.T) {
	app, _ := workshopTestApp(t)
	wc := NewWorkshopCobraCommands(app)
	cmd := wc.CreateWorkshopCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// Missing --title and --start
	cmd.SetArgs([]string{"config", "use", "ml-6h"})
	require.Error(t, cmd.Execute())
}
