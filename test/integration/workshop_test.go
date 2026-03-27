//go:build integration
// +build integration

// Package integration — workshop API integration tests for v0.18.0.
// Requires a running prismd daemon (PRISM_TEST_MODE=true recommended).
package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	apiclient "github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newWorkshopClient returns an HTTPClient pointed at a local daemon.
func newWorkshopClient(t *testing.T) *apiclient.HTTPClient {
	t.Helper()
	c := apiclient.NewClientWithOptions("http://localhost:8947", apiclient.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	hc, ok := c.(*apiclient.HTTPClient)
	require.True(t, ok, "expected *HTTPClient")
	return hc
}

// uniqueTitle returns a short unique workshop title for test isolation.
func uniqueTitle(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixMilli()%100000)
}

// TestWorkshopLifecycle exercises full workshop CRUD plus participant and dashboard operations.
func TestWorkshopLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test — requires running daemon")
	}

	ctx := context.Background()
	apiClient := apiclient.NewClientWithOptions("http://localhost:8947", apiclient.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)
	hc := newWorkshopClient(t)

	title := uniqueTitle("NeurIPS-DL")
	start := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	end := time.Now().Add(30 * time.Hour).UTC().Format(time.RFC3339)

	// ── Create ────────────────────────────────────────────────────────────────

	created, err := hc.CreateWorkshop(ctx, map[string]interface{}{
		"title":                  title,
		"description":            "Integration test workshop",
		"owner":                  "test-organizer",
		"template":               "python-ml",
		"max_participants":       30,
		"budget_per_participant": 5.0,
		"start_time":             start,
		"end_time":               end,
		"early_access_hours":     2,
	})
	require.NoError(t, err, "CreateWorkshop")
	workshopID, _ := created["id"].(string)
	require.NotEmpty(t, workshopID, "expected workshop ID in response")
	registry.RegisterWorkshop(workshopID)
	t.Logf("Created workshop: %s", workshopID)

	// Token should be auto-generated
	joinToken, _ := created["join_token"].(string)
	assert.NotEmpty(t, joinToken, "expected auto-generated join_token")

	// Status should start as draft
	status, _ := created["status"].(string)
	assert.Equal(t, "draft", status, "expected draft status on creation")

	// ── Get ───────────────────────────────────────────────────────────────────

	got, err := hc.GetWorkshop(ctx, workshopID)
	require.NoError(t, err, "GetWorkshop")
	assert.Equal(t, title, got["title"])
	assert.Equal(t, "python-ml", got["template"])

	// ── List ──────────────────────────────────────────────────────────────────

	listed, err := hc.ListWorkshops(ctx, "")
	require.NoError(t, err, "ListWorkshops")
	workshops, _ := listed["workshops"].([]interface{})
	found := false
	for _, raw := range workshops {
		w, _ := raw.(map[string]interface{})
		if w["id"] == workshopID {
			found = true
		}
	}
	assert.True(t, found, "created workshop not found in list")

	// List by owner
	ownerListed, err := hc.ListWorkshops(ctx, "owner=test-organizer")
	require.NoError(t, err, "ListWorkshops by owner")
	ownerWorkshops, _ := ownerListed["workshops"].([]interface{})
	ownerFound := false
	for _, raw := range ownerWorkshops {
		w, _ := raw.(map[string]interface{})
		if w["id"] == workshopID {
			ownerFound = true
		}
	}
	assert.True(t, ownerFound, "workshop not found when filtering by owner")

	// ── Update ────────────────────────────────────────────────────────────────

	updated, err := hc.UpdateWorkshop(ctx, workshopID, map[string]interface{}{
		"title": title + "-updated",
	})
	require.NoError(t, err, "UpdateWorkshop")
	assert.Equal(t, title+"-updated", updated["title"])

	// ── Add Participant ───────────────────────────────────────────────────────

	participant, err := hc.AddWorkshopParticipant(ctx, workshopID, map[string]interface{}{
		"user_id":      "alice-123",
		"display_name": "Alice",
		"email":        "alice@example.com",
	})
	require.NoError(t, err, "AddWorkshopParticipant")
	assert.NotEmpty(t, participant)
	t.Log("Added participant alice-123")

	// Idempotent — add same participant again should succeed or be handled gracefully
	_, err = hc.AddWorkshopParticipant(ctx, workshopID, map[string]interface{}{
		"user_id":      "alice-123",
		"display_name": "Alice",
		"email":        "alice@example.com",
	})
	// Either success (idempotent) or a conflict error — both are acceptable
	t.Logf("Idempotent add result: %v", err)

	// ── Dashboard ─────────────────────────────────────────────────────────────

	dashboard, err := hc.GetWorkshopDashboard(ctx, workshopID)
	require.NoError(t, err, "GetWorkshopDashboard")
	assert.Equal(t, workshopID, dashboard["workshop_id"])
	totalParticipants, _ := dashboard["total_participants"].(float64)
	assert.GreaterOrEqual(t, int(totalParticipants), 1, "expected at least 1 participant in dashboard")
	t.Logf("Dashboard: %v participants, %v active instances", dashboard["total_participants"], dashboard["active_instances"])

	// ── Provision ─────────────────────────────────────────────────────────────

	provResult, err := hc.ProvisionWorkshop(ctx, workshopID)
	require.NoError(t, err, "ProvisionWorkshop")
	assert.NotNil(t, provResult)
	t.Logf("Provision result: %v", provResult)

	// ── Remove Participant ────────────────────────────────────────────────────

	err = hc.RemoveWorkshopParticipant(ctx, workshopID, "alice-123")
	require.NoError(t, err, "RemoveWorkshopParticipant")

	// ── End Workshop ──────────────────────────────────────────────────────────

	endResult, err := hc.EndWorkshop(ctx, workshopID)
	require.NoError(t, err, "EndWorkshop")
	assert.NotNil(t, endResult)
	t.Logf("End result: %v", endResult)

	// Verify status is ended
	afterEnd, err := hc.GetWorkshop(ctx, workshopID)
	require.NoError(t, err, "GetWorkshop after end")
	endStatus, _ := afterEnd["status"].(string)
	assert.Equal(t, "ended", endStatus, "workshop should be ended after EndWorkshop")

	// ── Download Manifest ─────────────────────────────────────────────────────

	download, err := hc.GetWorkshopDownload(ctx, workshopID)
	require.NoError(t, err, "GetWorkshopDownload")
	assert.Equal(t, workshopID, download["workshop_id"])

	// ── Delete ────────────────────────────────────────────────────────────────

	err = hc.DeleteWorkshop(ctx, workshopID)
	require.NoError(t, err, "DeleteWorkshop")

	// Verify it's gone
	_, err = hc.GetWorkshop(ctx, workshopID)
	assert.Error(t, err, "GetWorkshop after delete should fail")
}

// TestWorkshopConfig exercises save/list/create-from-config operations (#183).
func TestWorkshopConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test — requires running daemon")
	}

	ctx := context.Background()
	apiClient := apiclient.NewClientWithOptions("http://localhost:8947", apiclient.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)
	hc := newWorkshopClient(t)

	title := uniqueTitle("Config-WS")
	start := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	end := time.Now().Add(30 * time.Hour).UTC().Format(time.RFC3339)

	// Create a workshop to save config from
	created, err := hc.CreateWorkshop(ctx, map[string]interface{}{
		"title":                  title,
		"owner":                  "config-organizer",
		"template":               "pytorch-ml",
		"max_participants":       20,
		"budget_per_participant": 10.0,
		"start_time":             start,
		"end_time":               end,
	})
	require.NoError(t, err, "CreateWorkshop for config test")
	workshopID, _ := created["id"].(string)
	require.NotEmpty(t, workshopID)
	registry.RegisterWorkshop(workshopID)

	configName := uniqueTitle("ml-config")

	// Save config
	saved, err := hc.SaveWorkshopConfig(ctx, workshopID, configName)
	require.NoError(t, err, "SaveWorkshopConfig")
	assert.Equal(t, configName, saved["name"])
	t.Logf("Saved config: %v", saved["name"])

	// List configs — should include our saved config
	configs, err := hc.ListWorkshopConfigs(ctx)
	require.NoError(t, err, "ListWorkshopConfigs")
	configsList, _ := configs["configs"].([]interface{})
	foundConfig := false
	for _, raw := range configsList {
		c, _ := raw.(map[string]interface{})
		if c["name"] == configName {
			foundConfig = true
		}
	}
	assert.True(t, foundConfig, "saved config not found in list")

	// Create from config
	newTitle := uniqueTitle("From-Config")
	newStart := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	newEnd := time.Now().Add(54 * time.Hour).UTC().Format(time.RFC3339)

	newWorkshop, err := hc.CreateWorkshopFromConfig(ctx, configName, map[string]interface{}{
		"title":      newTitle,
		"owner":      "config-organizer",
		"start_time": newStart,
		"end_time":   newEnd,
	})
	require.NoError(t, err, "CreateWorkshopFromConfig")
	newID, _ := newWorkshop["id"].(string)
	require.NotEmpty(t, newID)
	registry.RegisterWorkshop(newID)
	t.Logf("Created from config: %s", newID)

	// Verify the new workshop inherited config values
	newGot, err := hc.GetWorkshop(ctx, newID)
	require.NoError(t, err, "GetWorkshop for config-derived workshop")
	assert.Equal(t, "pytorch-ml", newGot["template"])
}
