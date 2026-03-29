package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTransport redirects any HTTP request to the provided handler.
type mockTransport struct{ handler http.Handler }

func (mt *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	mt.handler.ServeHTTP(rec, req)
	return rec.Result(), nil
}

func testChecker(t *testing.T, handler http.Handler) *Checker {
	t.Helper()
	t.Setenv("HOME", t.TempDir()) // isolate cache
	cacheDir := filepath.Join(t.TempDir(), "cache")
	return &Checker{
		cacheDir:   cacheDir,
		cacheTTL:   defaultCacheTTL,
		channel:    ChannelStable,
		httpClient: &http.Client{Transport: &mockTransport{handler: handler}},
	}
}

// githubRelease is the JSON structure the checker reads.
type githubRelease struct {
	TagName     string    `json:"tag_name"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
}

func releaseHandler(tag string, pre, draft bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := githubRelease{
			TagName:     tag,
			HTMLURL:     "https://github.com/example/releases/" + tag,
			Body:        "Release notes for " + tag,
			PublishedAt: time.Now(),
			Prerelease:  pre,
			Draft:       draft,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rel)
	})
}

// ── CachedUpdateCheck ─────────────────────────────────────────────────────

func TestCachedUpdateCheck_NotExpired(t *testing.T) {
	c := &CachedUpdateCheck{
		CheckedAt: time.Now(),
		CacheTTL:  time.Hour,
	}
	assert.False(t, c.IsExpired())
}

func TestCachedUpdateCheck_Expired(t *testing.T) {
	c := &CachedUpdateCheck{
		CheckedAt: time.Now().Add(-2 * time.Hour),
		CacheTTL:  time.Hour,
	}
	assert.True(t, c.IsExpired())
}

// ── FormatUpdateMessage ───────────────────────────────────────────────────

func TestFormatUpdateMessage_UpdateAvailable(t *testing.T) {
	info := &UpdateInfo{
		CurrentVersion:    "0.21.0",
		LatestVersion:     "0.22.0",
		IsUpdateAvailable: true,
		ReleaseURL:        "https://github.com/example/releases/v0.22.0",
		PublishedAt:       time.Now(),
		InstallMethod:     InstallMethodBinary,
		UpdateCommand:     "download binary",
	}
	msg := FormatUpdateMessage(info)
	assert.Contains(t, msg, "0.21.0")
	assert.Contains(t, msg, "0.22.0")
}

func TestFormatUpdateMessage_UpToDate(t *testing.T) {
	info := &UpdateInfo{
		CurrentVersion:    "0.21.0",
		LatestVersion:     "0.21.0",
		IsUpdateAvailable: false,
	}
	msg := FormatUpdateMessage(info)
	assert.Contains(t, msg, "latest version")
	assert.Contains(t, msg, "0.21.0")
}

func TestFormatShortUpdateMessage_UpdateAvailable(t *testing.T) {
	info := &UpdateInfo{
		CurrentVersion:    "0.21.0",
		LatestVersion:     "0.22.0",
		IsUpdateAvailable: true,
		UpdateCommand:     "brew upgrade prism",
	}
	msg := FormatShortUpdateMessage(info)
	assert.Contains(t, msg, "0.21.0")
	assert.Contains(t, msg, "0.22.0")
	assert.Contains(t, msg, "brew upgrade prism")
}

func TestFormatShortUpdateMessage_UpToDate(t *testing.T) {
	info := &UpdateInfo{IsUpdateAvailable: false}
	assert.Empty(t, FormatShortUpdateMessage(info))
}

// ── GetUpdateCommand ──────────────────────────────────────────────────────

func TestGetUpdateCommand_AllMethods(t *testing.T) {
	methods := []InstallMethod{
		InstallMethodHomebrew,
		InstallMethodScoop,
		InstallMethodBinary,
		InstallMethodSource,
		InstallMethodUnknown,
	}
	for _, m := range methods {
		t.Run(string(m), func(t *testing.T) {
			cmd := GetUpdateCommand(m)
			assert.NotEmpty(t, cmd)
		})
	}
}

func TestGetUpdateInstructions_AllMethods(t *testing.T) {
	methods := []InstallMethod{
		InstallMethodHomebrew,
		InstallMethodScoop,
		InstallMethodBinary,
		InstallMethodSource,
		InstallMethodUnknown,
	}
	for _, m := range methods {
		t.Run(string(m), func(t *testing.T) {
			inst := GetUpdateInstructions(m, "v0.22.0")
			assert.NotEmpty(t, inst)
		})
	}
}

// ── Checker.CheckForUpdates ───────────────────────────────────────────────

func TestChecker_CheckForUpdates_NewVersion(t *testing.T) {
	// Mock returns a much higher version than current
	c := testChecker(t, releaseHandler("v99.0.0", false, false))

	info, err := c.CheckForUpdates(context.Background())
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.True(t, info.IsUpdateAvailable)
	assert.Equal(t, "99.0.0", info.LatestVersion)
}

func TestChecker_CheckForUpdates_UpToDate(t *testing.T) {
	// Mock returns the same version as current
	c := testChecker(t, releaseHandler("v0.0.1", false, false))
	// Inject a very low current version to ensure it reads "not available" correctly:
	// compareVersions returns true when latest > current (string comparison)
	// We use v0.0.1 which is lower than real version, so IsUpdateAvailable == false
	info, err := c.CheckForUpdates(context.Background())
	require.NoError(t, err)
	require.NotNil(t, info)
	// 0.0.1 <= current version (e.g., 0.21.0) → no update available
	assert.False(t, info.IsUpdateAvailable)
}

func TestChecker_CheckForUpdates_NetworkError(t *testing.T) {
	// Handler returns 500
	errHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})
	c := testChecker(t, errHandler)

	_, err := c.CheckForUpdates(context.Background())
	assert.Error(t, err)
}

func TestChecker_CheckForUpdates_MalformedJSON(t *testing.T) {
	badHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{not valid json"))
	})
	c := testChecker(t, badHandler)

	_, err := c.CheckForUpdates(context.Background())
	assert.Error(t, err)
}

func TestChecker_CheckForUpdates_Prerelease_Skipped(t *testing.T) {
	// Stable channel skips pre-releases
	c := testChecker(t, releaseHandler("v99.0.0-beta", true, false))

	_, err := c.CheckForUpdates(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prerelease")
}

// ── ClearCache ────────────────────────────────────────────────────────────

func TestChecker_ClearCache(t *testing.T) {
	srv := httptest.NewServer(releaseHandler("v99.0.0", false, false))
	defer srv.Close()

	c := testChecker(t, releaseHandler("v99.0.0", false, false))
	// Fetch once to populate cache
	_, err := c.CheckForUpdates(context.Background())
	require.NoError(t, err)

	// Verify cache file was written
	cachePath := filepath.Join(c.cacheDir, cacheFileName)
	_, statErr := os.Stat(cachePath)
	require.NoError(t, statErr, "cache file should exist after check")

	// Clear cache
	require.NoError(t, c.ClearCache())

	_, statErr = os.Stat(cachePath)
	assert.True(t, os.IsNotExist(statErr), "cache file should be gone after ClearCache")
}
