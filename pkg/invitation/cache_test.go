package invitation

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInvitationCache(t *testing.T) {
	// Create temporary cache directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)
	assert.NotNil(t, cache)
	assert.Equal(t, 0, cache.Size())
}

func TestNewInvitationCache_LoadExisting(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	// Create a cache with pre-existing data
	stateDir := filepath.Join(tmpDir, ".prism")
	require.NoError(t, os.MkdirAll(stateDir, 0755))
	cachePath := filepath.Join(stateDir, "invitation_cache.json")

	// Write existing cache data
	existing := []*CachedInvitation{
		{
			Token:        "token1",
			InvitationID: "inv1",
			ProjectID:    "proj1",
			ProjectName:  "Project One",
			Email:        "user@example.com",
			Role:         types.ProjectRoleMember,
			Status:       types.InvitationPending,
			AddedAt:      time.Now(),
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	require.NoError(t, os.WriteFile(cachePath, data, 0600))

	// Load cache
	cache, err := NewInvitationCache()
	require.NoError(t, err)
	assert.Equal(t, 1, cache.Size())

	// Verify loaded data
	cached, err := cache.Get("Project One")
	require.NoError(t, err)
	assert.Equal(t, "token1", cached.Token)
	assert.Equal(t, "inv1", cached.InvitationID)
}

func TestInvitationCache_Add(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	ctx := context.Background()
	invitation := &types.Invitation{
		ID:        "inv-123",
		ProjectID: "proj-456",
		Email:     "user@example.com",
		Role:      types.ProjectRoleMember,
		InvitedBy: "admin@example.com",
		InvitedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Status:    types.InvitationPending,
		Message:   "Welcome to the project",
	}

	err = cache.Add(ctx, "full-token-string", invitation, "Test Project")
	require.NoError(t, err)
	assert.Equal(t, 1, cache.Size())

	// Retrieve and verify
	cached, err := cache.Get("Test Project")
	require.NoError(t, err)
	assert.Equal(t, "full-token-string", cached.Token)
	assert.Equal(t, "inv-123", cached.InvitationID)
	assert.Equal(t, "proj-456", cached.ProjectID)
	assert.Equal(t, "Test Project", cached.ProjectName)
	assert.Equal(t, "user@example.com", cached.Email)
	assert.Equal(t, types.ProjectRoleMember, cached.Role)
	assert.Equal(t, "admin@example.com", cached.InvitedBy)
	assert.Equal(t, types.InvitationPending, cached.Status)
	assert.Equal(t, "Welcome to the project", cached.Message)
}

func TestInvitationCache_Add_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	ctx := context.Background()
	invitation1 := &types.Invitation{
		ID:        "inv-123",
		ProjectID: "proj-456",
		Email:     "user@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationPending,
		InvitedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// Add first invitation
	err = cache.Add(ctx, "token1", invitation1, "Project X")
	require.NoError(t, err)

	// Add second invitation with same project name (should overwrite)
	invitation2 := &types.Invitation{
		ID:        "inv-789",
		ProjectID: "proj-456",
		Email:     "user@example.com",
		Role:      types.ProjectRoleAdmin,
		Status:    types.InvitationPending,
		InvitedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err = cache.Add(ctx, "token2", invitation2, "Project X")
	require.NoError(t, err)
	assert.Equal(t, 1, cache.Size()) // Still only 1 entry

	// Verify new token overwrote old one
	cached, err := cache.Get("Project X")
	require.NoError(t, err)
	assert.Equal(t, "token2", cached.Token)
	assert.Equal(t, "inv-789", cached.InvitationID)
	assert.Equal(t, types.ProjectRoleAdmin, cached.Role)
}

func TestInvitationCache_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	_, err = cache.Get("Nonexistent Project")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no cached invitation for project")
}

func TestInvitationCache_List(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	ctx := context.Background()

	// Add multiple invitations
	for i := 1; i <= 3; i++ {
		invitation := &types.Invitation{
			ID:        "inv-" + string(rune('0'+i)),
			ProjectID: "proj-" + string(rune('0'+i)),
			Email:     "user@example.com",
			Role:      types.ProjectRoleMember,
			Status:    types.InvitationPending,
			InvitedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		err = cache.Add(ctx, "token"+string(rune('0'+i)), invitation, "Project "+string(rune('0'+i)))
		require.NoError(t, err)
	}

	// List all
	invitations, err := cache.List()
	require.NoError(t, err)
	assert.Len(t, invitations, 3)

	// Verify modifications don't affect cache (returns copies)
	invitations[0].Status = types.InvitationAccepted
	cached, err := cache.Get(invitations[0].ProjectName)
	require.NoError(t, err)
	assert.Equal(t, types.InvitationPending, cached.Status, "List should return copies")
}

func TestInvitationCache_List_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	invitations, err := cache.List()
	require.NoError(t, err)
	assert.Len(t, invitations, 0)
}

func TestInvitationCache_Remove(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	ctx := context.Background()
	invitation := &types.Invitation{
		ID:        "inv-123",
		ProjectID: "proj-456",
		Email:     "user@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationPending,
		InvitedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err = cache.Add(ctx, "token", invitation, "Test Project")
	require.NoError(t, err)
	assert.Equal(t, 1, cache.Size())

	// Remove
	err = cache.Remove("Test Project")
	require.NoError(t, err)
	assert.Equal(t, 0, cache.Size())

	// Verify removed
	_, err = cache.Get("Test Project")
	assert.Error(t, err)
}

func TestInvitationCache_Remove_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	err = cache.Remove("Nonexistent Project")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no cached invitation for project")
}

func TestInvitationCache_CleanupExpired(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now()

	// Add expired invitation
	expired := &types.Invitation{
		ID:        "inv-expired",
		ProjectID: "proj-1",
		Email:     "user1@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationPending,
		InvitedAt: now.Add(-8 * 24 * time.Hour),
		ExpiresAt: now.Add(-1 * time.Hour), // Expired 1 hour ago
	}
	err = cache.Add(ctx, "token-expired", expired, "Expired Project")
	require.NoError(t, err)

	// Add valid invitation
	valid := &types.Invitation{
		ID:        "inv-valid",
		ProjectID: "proj-2",
		Email:     "user2@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationPending,
		InvitedAt: now,
		ExpiresAt: now.Add(7 * 24 * time.Hour), // Expires in 7 days
	}
	err = cache.Add(ctx, "token-valid", valid, "Valid Project")
	require.NoError(t, err)

	assert.Equal(t, 2, cache.Size())

	// Cleanup expired
	removed, err := cache.CleanupExpired()
	require.NoError(t, err)
	assert.Equal(t, 1, removed)
	assert.Equal(t, 1, cache.Size())

	// Verify expired removed
	_, err = cache.Get("Expired Project")
	assert.Error(t, err)

	// Verify valid remains
	_, err = cache.Get("Valid Project")
	assert.NoError(t, err)
}

func TestInvitationCache_CleanupExpired_NoneExpired(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	ctx := context.Background()
	invitation := &types.Invitation{
		ID:        "inv-valid",
		ProjectID: "proj-1",
		Email:     "user@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationPending,
		InvitedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err = cache.Add(ctx, "token", invitation, "Valid Project")
	require.NoError(t, err)

	// Cleanup (nothing expired)
	removed, err := cache.CleanupExpired()
	require.NoError(t, err)
	assert.Equal(t, 0, removed)
	assert.Equal(t, 1, cache.Size())
}

func TestInvitationCache_UpdateStatus(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	ctx := context.Background()
	invitation := &types.Invitation{
		ID:        "inv-123",
		ProjectID: "proj-456",
		Email:     "user@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationPending,
		InvitedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err = cache.Add(ctx, "token", invitation, "Test Project")
	require.NoError(t, err)

	// Update status
	err = cache.UpdateStatus("Test Project", types.InvitationAccepted)
	require.NoError(t, err)

	// Verify updated
	cached, err := cache.Get("Test Project")
	require.NoError(t, err)
	assert.Equal(t, types.InvitationAccepted, cached.Status)
}

func TestInvitationCache_UpdateStatus_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	err = cache.UpdateStatus("Nonexistent Project", types.InvitationAccepted)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no cached invitation for project")
}

func TestInvitationCache_GetSummary(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now()

	// Add pending invitation
	pending := &types.Invitation{
		ID:        "inv-pending",
		ProjectID: "proj-1",
		Email:     "user1@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationPending,
		InvitedAt: now,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
	}
	err = cache.Add(ctx, "token1", pending, "Pending Project")
	require.NoError(t, err)

	// Add accepted invitation
	accepted := &types.Invitation{
		ID:        "inv-accepted",
		ProjectID: "proj-2",
		Email:     "user2@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationAccepted,
		InvitedAt: now,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
	}
	err = cache.Add(ctx, "token2", accepted, "Accepted Project")
	require.NoError(t, err)

	// Add declined invitation
	declined := &types.Invitation{
		ID:        "inv-declined",
		ProjectID: "proj-3",
		Email:     "user3@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationDeclined,
		InvitedAt: now,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
	}
	err = cache.Add(ctx, "token3", declined, "Declined Project")
	require.NoError(t, err)

	// Add expired invitation
	expired := &types.Invitation{
		ID:        "inv-expired",
		ProjectID: "proj-4",
		Email:     "user4@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationPending,
		InvitedAt: now.Add(-8 * 24 * time.Hour),
		ExpiresAt: now.Add(-1 * time.Hour),
	}
	err = cache.Add(ctx, "token4", expired, "Expired Project")
	require.NoError(t, err)

	// Get summary
	total, pendingCount, acceptedCount, declinedCount, expiredCount := cache.GetSummary()
	assert.Equal(t, 4, total)
	assert.Equal(t, 1, pendingCount)
	assert.Equal(t, 1, acceptedCount)
	assert.Equal(t, 1, declinedCount)
	assert.Equal(t, 1, expiredCount)
}

func TestInvitationCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	ctx := context.Background()

	// Add some invitations
	for i := 1; i <= 3; i++ {
		invitation := &types.Invitation{
			ID:        "inv-" + string(rune('0'+i)),
			ProjectID: "proj-" + string(rune('0'+i)),
			Email:     "user@example.com",
			Role:      types.ProjectRoleMember,
			Status:    types.InvitationPending,
			InvitedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		err = cache.Add(ctx, "token"+string(rune('0'+i)), invitation, "Project "+string(rune('0'+i)))
		require.NoError(t, err)
	}

	assert.Equal(t, 3, cache.Size())

	// Clear
	err = cache.Clear()
	require.NoError(t, err)
	assert.Equal(t, 0, cache.Size())

	// Verify all removed
	invitations, err := cache.List()
	require.NoError(t, err)
	assert.Len(t, invitations, 0)
}

func TestInvitationCache_Size(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	assert.Equal(t, 0, cache.Size())

	ctx := context.Background()
	invitation := &types.Invitation{
		ID:        "inv-123",
		ProjectID: "proj-456",
		Email:     "user@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationPending,
		InvitedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err = cache.Add(ctx, "token", invitation, "Test Project")
	require.NoError(t, err)
	assert.Equal(t, 1, cache.Size())

	err = cache.Remove("Test Project")
	require.NoError(t, err)
	assert.Equal(t, 0, cache.Size())
}

func TestInvitationCache_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cache, err := NewInvitationCache()
	require.NoError(t, err)

	ctx := context.Background()
	done := make(chan bool)

	// Concurrent adds
	for i := 0; i < 10; i++ {
		go func(id int) {
			invitation := &types.Invitation{
				ID:        "inv-" + string(rune('0'+id)),
				ProjectID: "proj-" + string(rune('0'+id)),
				Email:     "user@example.com",
				Role:      types.ProjectRoleMember,
				Status:    types.InvitationPending,
				InvitedAt: time.Now(),
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			}
			cache.Add(ctx, "token"+string(rune('0'+id)), invitation, "Project "+string(rune('0'+id)))
			done <- true
		}(i)
	}

	// Wait for all adds
	for i := 0; i < 10; i++ {
		<-done
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func(id int) {
			cache.List()
			done <- true
		}(i)
	}

	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no race conditions (test will fail with -race if issues exist)
	assert.True(t, cache.Size() > 0)
}

func TestInvitationCache_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	// Create cache and add invitation
	cache1, err := NewInvitationCache()
	require.NoError(t, err)

	ctx := context.Background()
	invitation := &types.Invitation{
		ID:        "inv-123",
		ProjectID: "proj-456",
		Email:     "user@example.com",
		Role:      types.ProjectRoleMember,
		Status:    types.InvitationPending,
		InvitedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err = cache1.Add(ctx, "token", invitation, "Test Project")
	require.NoError(t, err)

	// Create new cache instance (should load from disk)
	cache2, err := NewInvitationCache()
	require.NoError(t, err)

	// Verify data persisted
	assert.Equal(t, 1, cache2.Size())
	cached, err := cache2.Get("Test Project")
	require.NoError(t, err)
	assert.Equal(t, "token", cached.Token)
	assert.Equal(t, "inv-123", cached.InvitationID)
}
