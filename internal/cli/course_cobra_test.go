// Package cli — unit tests for CourseCobraCommands and TACobraCommands.
// Uses httptest.NewServer because courseClient() requires a real *HTTPClient.
package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiclient "github.com/scttfrdmn/prism/pkg/api/client"
)

// courseTestApp creates an App backed by an httptest.Server that serves
// minimal JSON responses for every course/TA endpoint.
func courseTestApp(t *testing.T) (*App, *httptest.Server) {
	t.Helper()
	t.Setenv("PRISM_NO_AUTO_START", "1")

	mux := http.NewServeMux()

	// Ping — required by Course() and TA() before dispatching
	mux.HandleFunc("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Courses collection — matches /api/v1/courses and /api/v1/courses?... (no trailing slash)
	mux.HandleFunc("/api/v1/courses", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"courses": []map[string]interface{}{
					{"id": "cs101", "code": "CS101", "title": "Intro to CS", "status": "active"},
				},
			})
		case http.MethodPost:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "new-course", "code": "CS201", "title": "Data Structures", "status": "pending",
			})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Individual course
	mux.HandleFunc("/api/v1/courses/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "cs101", "code": "CS101", "title": "Intro to CS", "status": "active",
				"members": []interface{}{},
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	hc := apiclient.NewClientWithOptions(srv.URL, apiclient.Options{})
	app := NewAppWithClient("1.0.0", hc)
	return app, srv
}

// ── course list ───────────────────────────────────────────────────────────────

func TestCourseCobraCommands_List(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"list"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_List_WithFlags(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"list", "--semester", "Fall-2099", "--status", "active"})
	require.NoError(t, cmd.Execute())
}

// ── course create ─────────────────────────────────────────────────────────────

func TestCourseCobraCommands_Create(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"create", "CS201",
		"--title", "Data Structures",
		"--semester", "Fall 2099",
		"--owner", "prof1",
		"--start", "2099-09-01",
		"--end", "2099-12-15",
	})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Create_MissingArgs(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"create"}) // ExactArgs(1) — missing code
	err := cmd.Execute()
	assert.Error(t, err)
}

// ── course show ───────────────────────────────────────────────────────────────

func TestCourseCobraCommands_Show(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"show", "cs101"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Show_MissingID(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"show"})
	err := cmd.Execute()
	assert.Error(t, err)
}

// ── course close ──────────────────────────────────────────────────────────────

func TestCourseCobraCommands_Close(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"close", "cs101"})
	require.NoError(t, cmd.Execute())
}

// ── course delete ─────────────────────────────────────────────────────────────

func TestCourseCobraCommands_Delete(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"delete", "cs101"})
	require.NoError(t, cmd.Execute())
}

// ── course templates ──────────────────────────────────────────────────────────

func TestCourseCobraCommands_Templates_List(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"templates", "list", "cs101"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Templates_Add(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"templates", "add", "cs101", "python-ml"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Templates_Remove(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"templates", "remove", "cs101", "python-ml"})
	require.NoError(t, cmd.Execute())
}

// ── course students ───────────────────────────────────────────────────────────

func TestCourseCobraCommands_Students_List(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"students", "list", "cs101"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Students_Enroll(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"students", "enroll", "cs101",
		"--email", "student@uni.edu",
		"--name", "Alice",
		"--role", "student",
	})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Students_Unenroll(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"students", "unenroll", "cs101", "alice"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Students_Import(t *testing.T) {
	// Write a temp CSV to pass as a file argument
	tmpFile := t.TempDir() + "/roster.csv"
	f := mustCreateFile(t, tmpFile)
	_, err := io.WriteString(f, "user_id,email,display_name,role\nalice,alice@uni.edu,Alice,student\n")
	require.NoError(t, err)

	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"students", "import", "cs101", tmpFile, "--format", "prism"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Students_Provision(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"students", "provision", "cs101", "alice"})
	require.NoError(t, cmd.Execute())
}

// ── course budget ─────────────────────────────────────────────────────────────

func TestCourseCobraCommands_Budget_Show(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"budget", "show", "cs101"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Budget_Distribute(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"budget", "distribute", "cs101", "--amount", "50"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Budget_Distribute_MissingAmount(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"budget", "distribute", "cs101"}) // --amount is required
	err := cmd.Execute()
	assert.Error(t, err)
}

// ── course archive / report / audit ──────────────────────────────────────────

func TestCourseCobraCommands_Archive(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"archive", "cs101"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Report(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"report", "cs101"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Report_CSV(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"report", "cs101", "--format", "csv"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Audit(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"audit", "cs101"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_Audit_WithFilters(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"audit", "cs101", "--student", "alice", "--limit", "10"})
	require.NoError(t, cmd.Execute())
}

// ── TA commands ───────────────────────────────────────────────────────────────

func TestTACobraCommands_Debug(t *testing.T) {
	app, _ := courseTestApp(t)
	tc := NewTACobraCommands(app)
	cmd := tc.CreateTACommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"debug", "cs101", "alice"})
	require.NoError(t, cmd.Execute())
}

func TestTACobraCommands_Debug_MissingArgs(t *testing.T) {
	app, _ := courseTestApp(t)
	tc := NewTACobraCommands(app)
	cmd := tc.CreateTACommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"debug", "cs101"}) // ExactArgs(2) — missing student ID
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestTACobraCommands_Reset(t *testing.T) {
	app, _ := courseTestApp(t)
	tc := NewTACobraCommands(app)
	cmd := tc.CreateTACommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"reset", "cs101", "alice", "--reason", "broken environment"})
	require.NoError(t, cmd.Execute())
}

func TestTACobraCommands_Reset_MissingReason(t *testing.T) {
	app, _ := courseTestApp(t)
	tc := NewTACobraCommands(app)
	cmd := tc.CreateTACommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"reset", "cs101", "alice"}) // --reason required
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestTACobraCommands_Overview(t *testing.T) {
	app, _ := courseTestApp(t)
	tc := NewTACobraCommands(app)
	cmd := tc.CreateTACommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"overview", "cs101"})
	require.NoError(t, cmd.Execute())
}

// ── v0.19.0 course ta-access ─────────────────────────────────────────────────

func TestCourseCobraCommands_TAAccessList(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"ta-access", "list", "cs101"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_TAAccessGrant(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"ta-access", "grant", "cs101", "--email", "ta@uni.edu"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_TAAccessGrant_MissingEmail(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"ta-access", "grant", "cs101"}) // --email required
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCourseCobraCommands_TAAccessRevoke(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"ta-access", "revoke", "cs101", "--email", "ta@uni.edu"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_TAAccessConnect(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"ta-access", "connect", "cs101", "--student", "stu@uni.edu", "--reason", "office hours"})
	require.NoError(t, cmd.Execute())
}

// ── v0.19.0 course materials ──────────────────────────────────────────────────

func TestCourseCobraCommands_MaterialsCreate(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"materials", "create", "cs101", "--size", "50"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_MaterialsList(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"materials", "list", "cs101"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_MaterialsMount(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"materials", "mount", "cs101"})
	require.NoError(t, cmd.Execute())
}

// ── v0.19.0 course reset-workspace ───────────────────────────────────────────

func TestCourseCobraCommands_ResetWorkspace(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"reset-workspace", "cs101", "--student", "stu@uni.edu", "--reason", "broken"})
	require.NoError(t, cmd.Execute())
}

func TestCourseCobraCommands_ResetWorkspace_MissingStudent(t *testing.T) {
	app, _ := courseTestApp(t)
	cc := NewCourseCobraCommands(app)
	cmd := cc.CreateCourseCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"reset-workspace", "cs101", "--reason", "broken"}) // --student required
	err := cmd.Execute()
	assert.Error(t, err)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// mustCreate opens/creates a file for writing and returns it.
func mustCreateFile(t *testing.T, path string) io.WriteCloser {
	t.Helper()
	f, err := os.Create(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	return f
}
