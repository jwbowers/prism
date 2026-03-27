package workshop

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testManager creates a Manager backed by a temporary directory.
func testManager(t *testing.T) *Manager {
	t.Helper()
	dir := t.TempDir()
	orig := os.Getenv("HOME")
	t.Cleanup(func() { _ = os.Setenv("HOME", orig) })
	_ = os.Setenv("HOME", dir)

	m, err := NewManager()
	require.NoError(t, err)
	return m
}

// testCreateReq returns a valid CreateWorkshopRequest in the future.
func testCreateReq() *CreateWorkshopRequest {
	start := time.Now().Add(24 * time.Hour)
	end := start.Add(6 * time.Hour)
	return &CreateWorkshopRequest{
		Title:     "Test Workshop",
		Owner:     "organizer1",
		Template:  "python-ml",
		StartTime: start,
		EndTime:   end,
	}
}

// ── CreateWorkshop ──────────────────────────────────────────────────────────

func TestCreateWorkshop_Success(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)
	assert.NotEmpty(t, w.ID)
	assert.Equal(t, "Test Workshop", w.Title)
	assert.Equal(t, "organizer1", w.Owner)
	assert.Equal(t, "python-ml", w.Template)
	assert.Equal(t, WorkshopStatusDraft, w.Status)
	assert.True(t, strings.HasPrefix(w.JoinToken, "WS-"))
	assert.NotZero(t, w.CreatedAt)
}

func TestCreateWorkshop_ActiveStatus(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req := testCreateReq()
	req.StartTime = time.Now().Add(-1 * time.Hour) // started an hour ago
	req.EndTime = time.Now().Add(5 * time.Hour)

	w, err := m.CreateWorkshop(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, WorkshopStatusActive, w.Status)
}

func TestCreateWorkshop_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	tests := []struct {
		name   string
		mutate func(*CreateWorkshopRequest)
		errMsg string
	}{
		{"missing title", func(r *CreateWorkshopRequest) { r.Title = "" }, "title is required"},
		{"missing owner", func(r *CreateWorkshopRequest) { r.Owner = "" }, "owner is required"},
		{"missing template", func(r *CreateWorkshopRequest) { r.Template = "" }, "template is required"},
		{"missing start", func(r *CreateWorkshopRequest) { r.StartTime = time.Time{} }, "start_time is required"},
		{"missing end", func(r *CreateWorkshopRequest) { r.EndTime = time.Time{} }, "end_time is required"},
		{"end before start", func(r *CreateWorkshopRequest) { r.EndTime = r.StartTime.Add(-time.Hour) }, "end_time must be after start_time"},
		{"negative capacity", func(r *CreateWorkshopRequest) { r.MaxParticipants = -1 }, "max_participants cannot be negative"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := testCreateReq()
			tt.mutate(req)
			_, err := m.CreateWorkshop(ctx, req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// ── GetWorkshop / ListWorkshops ─────────────────────────────────────────────

func TestGetWorkshop(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	got, err := m.GetWorkshop(ctx, w.ID)
	require.NoError(t, err)
	assert.Equal(t, w.ID, got.ID)
	assert.Equal(t, "Test Workshop", got.Title)
}

func TestGetWorkshop_NotFound(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	_, err := m.GetWorkshop(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrWorkshopNotFound)
}

func TestListWorkshops_NoFilter(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req1 := testCreateReq()
	req2 := testCreateReq()
	req2.Title = "Second Workshop"
	req2.Owner = "organizer2"

	_, err := m.CreateWorkshop(ctx, req1)
	require.NoError(t, err)
	_, err = m.CreateWorkshop(ctx, req2)
	require.NoError(t, err)

	list, err := m.ListWorkshops(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestListWorkshops_FilterByOwner(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req1 := testCreateReq()
	req2 := testCreateReq()
	req2.Title = "Second Workshop"
	req2.Owner = "other-owner"

	_, err := m.CreateWorkshop(ctx, req1)
	require.NoError(t, err)
	_, err = m.CreateWorkshop(ctx, req2)
	require.NoError(t, err)

	list, err := m.ListWorkshops(ctx, &WorkshopFilter{Owner: "organizer1"})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Test Workshop", list[0].Title)
}

// ── UpdateWorkshop ──────────────────────────────────────────────────────────

func TestUpdateWorkshop(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	newTitle := "Updated Workshop"
	updated, err := m.UpdateWorkshop(ctx, w.ID, &UpdateWorkshopRequest{
		Title: &newTitle,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Workshop", updated.Title)
	assert.True(t, updated.UpdatedAt.After(w.UpdatedAt) || updated.UpdatedAt.Equal(w.UpdatedAt))
}

func TestUpdateWorkshop_NotFound(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	title := "X"
	_, err := m.UpdateWorkshop(ctx, "no-such-id", &UpdateWorkshopRequest{Title: &title})
	assert.ErrorIs(t, err, ErrWorkshopNotFound)
}

// ── DeleteWorkshop ──────────────────────────────────────────────────────────

func TestDeleteWorkshop(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	require.NoError(t, m.DeleteWorkshop(ctx, w.ID))

	_, err = m.GetWorkshop(ctx, w.ID)
	assert.ErrorIs(t, err, ErrWorkshopNotFound)
}

func TestDeleteWorkshop_NotFound(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	assert.ErrorIs(t, m.DeleteWorkshop(ctx, "ghost"), ErrWorkshopNotFound)
}

// ── AddParticipant / RemoveParticipant ──────────────────────────────────────

func TestAddParticipant_Success(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	p := WorkshopParticipant{UserID: "student1", Email: "s1@example.com"}
	updated, err := m.AddParticipant(ctx, w.ID, p)
	require.NoError(t, err)
	assert.Len(t, updated.Participants, 1)
	assert.Equal(t, "pending", updated.Participants[0].Status)
	assert.False(t, updated.Participants[0].JoinedAt.IsZero())
}

func TestAddParticipant_Idempotent(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	p := WorkshopParticipant{UserID: "student1", Email: "s1@example.com"}
	_, err = m.AddParticipant(ctx, w.ID, p)
	require.NoError(t, err)
	updated, err := m.AddParticipant(ctx, w.ID, p) // second add
	require.NoError(t, err)
	assert.Len(t, updated.Participants, 1) // still 1
}

func TestAddParticipant_CapacityEnforced(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req := testCreateReq()
	req.MaxParticipants = 1
	w, err := m.CreateWorkshop(ctx, req)
	require.NoError(t, err)

	_, err = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p1"})
	require.NoError(t, err)

	_, err = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p2"})
	assert.ErrorIs(t, err, ErrWorkshopFull)
}

func TestAddParticipant_EndedWorkshopRejected(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	_, _, err = m.EndWorkshop(ctx, w.ID)
	require.NoError(t, err)

	_, err = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "late"})
	assert.ErrorIs(t, err, ErrWorkshopEnded)
}

func TestRemoveParticipant(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	_, err = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p1"})
	require.NoError(t, err)

	require.NoError(t, m.RemoveParticipant(ctx, w.ID, "p1"))

	got, err := m.GetWorkshop(ctx, w.ID)
	require.NoError(t, err)
	assert.Len(t, got.Participants, 0)
}

func TestRemoveParticipant_NotFound(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	assert.ErrorIs(t, m.RemoveParticipant(ctx, w.ID, "ghost"), ErrParticipantNotFound)
}

// ── ProvisionWorkshop ───────────────────────────────────────────────────────

func TestProvisionWorkshop_ReturnsPendingParticipants(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	_, _ = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p1"})
	_, _ = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p2"})

	result, toProvision, err := m.ProvisionWorkshop(ctx, w.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Skipped)
	assert.Len(t, toProvision, 2)
}

func TestProvisionWorkshop_SkipsAlreadyProvisioned(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	_, _ = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p1", Status: "running"})
	_, _ = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p2"})

	result, toProvision, err := m.ProvisionWorkshop(ctx, w.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Skipped) // p1 already running
	assert.Len(t, toProvision, 1)      // only p2 needs provisioning
}

// ── GetDashboard ────────────────────────────────────────────────────────────

func TestGetDashboard(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	_, _ = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p1", Status: "running"})
	_, _ = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p2", Status: "stopped"})
	_, _ = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p3"}) // pending

	dash, err := m.GetDashboard(ctx, w.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, dash.TotalParticipants)
	assert.Equal(t, 1, dash.ActiveInstances)
	assert.Equal(t, 1, dash.StoppedInstances)
	assert.Equal(t, 1, dash.PendingInstances)
	assert.NotEmpty(t, dash.TimeRemaining)
}

// ── EndWorkshop ─────────────────────────────────────────────────────────────

func TestEndWorkshop(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	_, _ = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p1", InstanceName: "ws-p1"})
	_, _ = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p2"}) // no instance

	result, withInstances, err := m.EndWorkshop(ctx, w.ID)
	require.NoError(t, err)
	assert.Equal(t, w.ID, result.WorkshopID)
	assert.Len(t, withInstances, 1) // only p1 has an instance

	// Verify status updated
	got, err := m.GetWorkshop(ctx, w.ID)
	require.NoError(t, err)
	assert.Equal(t, WorkshopStatusEnded, got.Status)
}

// ── GetDownloadManifest ─────────────────────────────────────────────────────

func TestGetDownloadManifest(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	_, _ = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p1", InstanceName: "ws-p1", DisplayName: "Alice"})
	_, _ = m.AddParticipant(ctx, w.ID, WorkshopParticipant{UserID: "p2", DisplayName: "Bob"})

	manifest, err := m.GetDownloadManifest(ctx, w.ID)
	require.NoError(t, err)
	assert.Equal(t, w.ID, manifest.WorkshopID)
	assert.Len(t, manifest.Participants, 2)

	// p1 has an instance — download note should mention the instance name
	for _, pw := range manifest.Participants {
		if pw.UserID == "p1" {
			assert.Contains(t, pw.DownloadNote, "ws-p1")
		} else {
			assert.Contains(t, pw.DownloadNote, "no workspace")
		}
	}
}

// ── Config Templates ────────────────────────────────────────────────────────

func TestSaveAndListConfigs(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)

	cfg, err := m.SaveConfig(ctx, w.ID, "ml-workshop-6h")
	require.NoError(t, err)
	assert.Equal(t, "ml-workshop-6h", cfg.Name)
	assert.Equal(t, "python-ml", cfg.Template)
	assert.Equal(t, 6, cfg.DurationHours)

	configs, err := m.ListConfigs(ctx)
	require.NoError(t, err)
	assert.Len(t, configs, 1)
	assert.Equal(t, "ml-workshop-6h", configs[0].Name)
}

func TestSaveConfig_NotFound(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	_, err := m.SaveConfig(ctx, "no-such-workshop", "my-config")
	assert.ErrorIs(t, err, ErrWorkshopNotFound)
}

func TestCreateFromConfig(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	// Create source workshop + save config
	w, err := m.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)
	_, err = m.SaveConfig(ctx, w.ID, "ml-template")
	require.NoError(t, err)

	// Create from config
	start := time.Now().Add(48 * time.Hour)
	newReq := &CreateWorkshopRequest{
		Title:     "New Workshop",
		Owner:     "organizer2",
		StartTime: start,
	}
	newW, err := m.CreateFromConfig(ctx, "ml-template", newReq)
	require.NoError(t, err)
	assert.Equal(t, "New Workshop", newW.Title)
	assert.Equal(t, "python-ml", newW.Template) // filled from config
	assert.NotZero(t, newW.EndTime)             // derived from duration
}

func TestCreateFromConfig_NotFound(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	_, err := m.CreateFromConfig(ctx, "nonexistent", &CreateWorkshopRequest{
		Title:     "X",
		Owner:     "o",
		Template:  "t",
		StartTime: time.Now().Add(time.Hour),
		EndTime:   time.Now().Add(2 * time.Hour),
	})
	assert.ErrorIs(t, err, ErrConfigNotFound)
}

// ── Persistence (save/load round-trip) ─────────────────────────────────────

func TestPersistenceRoundTrip(t *testing.T) {
	dir := t.TempDir()
	orig := os.Getenv("HOME")
	t.Cleanup(func() { _ = os.Setenv("HOME", orig) })
	_ = os.Setenv("HOME", dir)

	ctx := context.Background()
	m1, err := NewManager()
	require.NoError(t, err)

	w, err := m1.CreateWorkshop(ctx, testCreateReq())
	require.NoError(t, err)
	_, _ = m1.SaveConfig(ctx, w.ID, "saved-cfg")

	// Load a fresh manager from the same directory
	m2, err := NewManager()
	require.NoError(t, err)

	got, err := m2.GetWorkshop(ctx, w.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test Workshop", got.Title)

	configs, err := m2.ListConfigs(ctx)
	require.NoError(t, err)
	assert.Len(t, configs, 1)
}
