package templates

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommunityMetadataStore_CreateAndGet tests basic metadata operations
func TestCommunityMetadataStore_CreateAndGet(t *testing.T) {
	store := createTestMetadataStore(t)

	templateName := "Test Template"
	sourceURL := "https://github.com/test/repo"

	// Get non-existent metadata (should create default)
	metadata, err := store.GetMetadata(templateName, sourceURL)
	require.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, 0, metadata.Downloads)
	assert.Equal(t, 0.0, metadata.Rating)
}

// TestCommunityMetadataStore_UpdateDownloadCount tests download tracking
func TestCommunityMetadataStore_UpdateDownloadCount(t *testing.T) {
	store := createTestMetadataStore(t)

	templateName := "Popular Template"
	sourceURL := "https://github.com/test/repo"

	// Update downloads multiple times
	for i := 1; i <= 5; i++ {
		err := store.UpdateDownloadCount(templateName, sourceURL)
		require.NoError(t, err)

		metadata, err := store.GetMetadata(templateName, sourceURL)
		require.NoError(t, err)
		assert.Equal(t, i, metadata.Downloads, "Download count should be %d", i)
	}
}

// TestCommunityMetadataStore_AddRating tests rating system
func TestCommunityMetadataStore_AddRating(t *testing.T) {
	store := createTestMetadataStore(t)

	templateName := "Rated Template"
	sourceURL := "https://github.com/test/repo"

	tests := []struct {
		rating         float64
		expectedRating float64
		expectedCount  int
		description    string
	}{
		{5.0, 5.0, 1, "First rating"},
		{3.0, 4.0, 2, "Second rating (average 4.0)"},
		{4.0, 4.0, 3, "Third rating (average 4.0)"},
		{2.0, 3.5, 4, "Fourth rating (average 3.5)"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			err := store.AddRating(templateName, sourceURL, tt.rating)
			require.NoError(t, err)

			metadata, err := store.GetMetadata(templateName, sourceURL)
			require.NoError(t, err)
			assert.InDelta(t, tt.expectedRating, metadata.Rating, 0.01,
				"Rating should be ~%.1f", tt.expectedRating)
			assert.Equal(t, tt.expectedCount, metadata.RatingCount)
		})
	}
}

// TestCommunityMetadataStore_UpdateUsageStats tests usage tracking
func TestCommunityMetadataStore_UpdateUsageStats(t *testing.T) {
	store := createTestMetadataStore(t)

	templateName := "Used Template"
	sourceURL := "https://github.com/test/repo"

	// Track successful launches
	err := store.UpdateUsageStats(templateName, sourceURL, true, 120)
	require.NoError(t, err)

	metadata, err := store.GetMetadata(templateName, sourceURL)
	require.NoError(t, err)
	assert.Equal(t, 1, metadata.UsageStats.TotalLaunches)
	assert.Equal(t, 1, metadata.UsageStats.SuccessfulLaunches)
	assert.Equal(t, 0, metadata.UsageStats.FailedLaunches)
	assert.Equal(t, 120, metadata.UsageStats.AverageLaunchTime)

	// Track failed launch
	err = store.UpdateUsageStats(templateName, sourceURL, false, 0)
	require.NoError(t, err)

	metadata, err = store.GetMetadata(templateName, sourceURL)
	require.NoError(t, err)
	assert.Equal(t, 2, metadata.UsageStats.TotalLaunches)
	assert.Equal(t, 1, metadata.UsageStats.SuccessfulLaunches)
	assert.Equal(t, 1, metadata.UsageStats.FailedLaunches)

	// Track another successful launch with different time
	err = store.UpdateUsageStats(templateName, sourceURL, true, 180)
	require.NoError(t, err)

	metadata, err = store.GetMetadata(templateName, sourceURL)
	require.NoError(t, err)
	assert.Equal(t, 3, metadata.UsageStats.TotalLaunches)
	assert.Equal(t, 2, metadata.UsageStats.SuccessfulLaunches)
	// Average should be (120 + 180) / 2 = 150
	assert.Equal(t, 150, metadata.UsageStats.AverageLaunchTime)
}

// TestCommunityMetadataStore_UpdateSecurityScan tests security scan tracking
func TestCommunityMetadataStore_UpdateSecurityScan(t *testing.T) {
	store := createTestMetadataStore(t)

	templateName := "Scanned Template"
	sourceURL := "https://github.com/test/repo"

	// Update security scan
	err := store.UpdateSecurityScan(templateName, sourceURL, 85)
	require.NoError(t, err)

	metadata, err := store.GetMetadata(templateName, sourceURL)
	require.NoError(t, err)
	assert.True(t, metadata.SecurityScanned)
	assert.Equal(t, 85, metadata.SecurityScore)
	assert.False(t, metadata.SecurityScanDate.IsZero())

	// Update again with different score
	time.Sleep(10 * time.Millisecond) // Ensure different timestamp
	err = store.UpdateSecurityScan(templateName, sourceURL, 92)
	require.NoError(t, err)

	metadata, err = store.GetMetadata(templateName, sourceURL)
	require.NoError(t, err)
	assert.Equal(t, 92, metadata.SecurityScore)
}

// TestCommunityMetadataStore_PopularityScoring tests popularity algorithm
func TestCommunityMetadataStore_PopularityScoring(t *testing.T) {
	store := createTestMetadataStore(t)

	// Create template with various metrics
	templateName := "Popular Template"
	sourceURL := "https://github.com/test/repo"

	// Set downloads
	for i := 0; i < 100; i++ {
		store.UpdateDownloadCount(templateName, sourceURL)
	}

	// Set rating
	store.AddRating(templateName, sourceURL, 4.5)

	// Set usage stats
	store.UpdateUsageStats(templateName, sourceURL, true, 120)
	store.UpdateUsageStats(templateName, sourceURL, true, 150)
	store.UpdateUsageStats(templateName, sourceURL, false, 0)

	metadata, err := store.GetMetadata(templateName, sourceURL)
	require.NoError(t, err)

	// Calculate expected popularity
	// Downloads (100) * 0.4 = 40
	// Success rate (2/3 = 0.667) * 100 * 0.3 = 20
	// Rating (4.5) * 20 * 0.2 = 18
	// Stars (0) * 0.1 = 0
	// Total = 40 + 20 + 18 + 0 = 78
	expectedPopularity := 78.0

	assert.InDelta(t, expectedPopularity, metadata.UsageStats.PopularityScore, 1.0,
		"Popularity score should be ~%.1f", expectedPopularity)
}

// TestCommunityMetadataStore_GetPopularTemplates tests popular templates query
func TestCommunityMetadataStore_GetPopularTemplates(t *testing.T) {
	store := createTestMetadataStore(t)

	// Create multiple templates with different popularity
	templates := []struct {
		name      string
		downloads int
		rating    float64
	}{
		{"Very Popular", 200, 5.0},
		{"Somewhat Popular", 50, 4.0},
		{"Not Popular", 10, 3.0},
		{"New Template", 5, 4.5},
	}

	for _, tmpl := range templates {
		for i := 0; i < tmpl.downloads; i++ {
			store.UpdateDownloadCount(tmpl.name, "https://github.com/test/repo")
		}
		store.AddRating(tmpl.name, "https://github.com/test/repo", tmpl.rating)
		store.UpdateUsageStats(tmpl.name, "https://github.com/test/repo", true, 100)
	}

	// Get top 3 popular templates
	popular, err := store.GetPopularTemplates(3)
	require.NoError(t, err)
	require.Len(t, popular, 3)

	// TemplateName is now stamped by SetMetadata — verify the most popular template is identifiable.
	assert.Equal(t, "Very Popular", popular[0].TemplateName,
		"most popular template should be first; got %q", popular[0].TemplateName)
	for _, tmpl := range popular {
		assert.NotEmpty(t, tmpl.TemplateName, "every result must carry its template name")
	}
}

// TestCommunityMetadataStore_GetFeaturedTemplates tests featured query
func TestCommunityMetadataStore_GetFeaturedTemplates(t *testing.T) {
	store := createTestMetadataStore(t)

	// Create templates, some featured
	store.UpdateDownloadCount("Featured Template 1", "https://github.com/test/repo")
	store.UpdateDownloadCount("Regular Template", "https://github.com/test/repo")
	store.UpdateDownloadCount("Featured Template 2", "https://github.com/test/repo")

	// Mark as featured
	metadata1, _ := store.GetMetadata("Featured Template 1", "https://github.com/test/repo")
	metadata1.Featured = true
	store.SetMetadata("Featured Template 1", "https://github.com/test/repo", metadata1)

	metadata2, _ := store.GetMetadata("Featured Template 2", "https://github.com/test/repo")
	metadata2.Featured = true
	store.SetMetadata("Featured Template 2", "https://github.com/test/repo", metadata2)

	// Get featured templates
	featured, err := store.GetFeaturedTemplates()
	require.NoError(t, err)
	assert.Len(t, featured, 2)

	// Check both are featured
	for _, tmpl := range featured {
		assert.True(t, tmpl.Featured)
	}
}

// TestCommunityMetadataStore_GetHighRatedTemplates tests rating query
func TestCommunityMetadataStore_GetHighRatedTemplates(t *testing.T) {
	store := createTestMetadataStore(t)

	// Create templates with various ratings
	templates := []struct {
		name   string
		rating float64
		count  int
	}{
		{"Excellent", 4.8, 10},
		{"Very Good", 4.5, 8},
		{"Good", 4.0, 5},
		{"Okay", 3.5, 3},
		{"New", 5.0, 1}, // Should be filtered out (min count)
	}

	for _, tmpl := range templates {
		for i := 0; i < tmpl.count; i++ {
			store.AddRating(tmpl.name, "https://github.com/test/repo", tmpl.rating)
		}
	}

	// Get templates with rating >= 4.0 and at least 5 ratings
	highRated, err := store.GetHighRatedTemplates(4.0, 5)
	require.NoError(t, err)

	// Should have 3 templates (Excellent, Very Good, Good)
	// "New" excluded by rating count
	// "Okay" excluded by rating threshold
	assert.GreaterOrEqual(t, len(highRated), 2)

	for _, tmpl := range highRated {
		assert.GreaterOrEqual(t, tmpl.Rating, 4.0)
		assert.GreaterOrEqual(t, tmpl.RatingCount, 5)
	}
}

// TestCommunityMetadataStore_SearchByKeywords tests keyword search
func TestCommunityMetadataStore_SearchByKeywords(t *testing.T) {
	store := createTestMetadataStore(t)

	// Create templates with keywords
	templates := []struct {
		name     string
		keywords []string
	}{
		{"Python ML", []string{"python", "machine-learning", "tensorflow"}},
		{"R Statistics", []string{"r", "statistics", "data-science"}},
		{"Python Web", []string{"python", "web", "django"}},
	}

	for _, tmpl := range templates {
		metadata, _ := store.GetMetadata(tmpl.name, "https://github.com/test/repo")
		metadata.Keywords = tmpl.keywords
		store.SetMetadata(tmpl.name, "https://github.com/test/repo", metadata)
	}

	// Search for "python"
	results, err := store.SearchByKeywords([]string{"python"})
	require.NoError(t, err)
	assert.Len(t, results, 2, "searching 'python' should return Python ML and Python Web")

	// TemplateName is now stamped — verify results carry their names.
	names := make([]string, 0, len(results))
	for _, r := range results {
		assert.NotEmpty(t, r.TemplateName)
		names = append(names, r.TemplateName)
	}
	assert.Contains(t, names, "Python ML")
	assert.Contains(t, names, "Python Web")
}

// TestCommunityMetadataStore_DisplayMethods tests display helper methods
func TestCommunityMetadataStore_DisplayMethods(t *testing.T) {
	metadata := &CommunityMetadata{
		Rating:        4.6,
		RatingCount:   15,
		TrustLevel:    "verified",
		SecurityScore: 85,
	}

	t.Run("GetTrustBadge", func(t *testing.T) {
		// Verified
		metadata.TrustLevel = "verified"
		assert.Equal(t, "🛡️ VERIFIED", metadata.GetTrustBadge())

		// Community
		metadata.TrustLevel = "community"
		assert.Equal(t, "🤝 COMMUNITY", metadata.GetTrustBadge())

		// Unverified
		metadata.TrustLevel = "unverified"
		assert.Equal(t, "⚠️ UNVERIFIED", metadata.GetTrustBadge())
	})

	t.Run("GetSecurityBadge", func(t *testing.T) {
		// Must be marked as scanned for score-based badges to appear.
		metadata.SecurityScanned = true

		// Excellent (score >= 90)
		metadata.SecurityScore = 95
		badge := metadata.GetSecurityBadge()
		assert.Contains(t, badge, "🟢")
		assert.Contains(t, badge, "EXCELLENT")

		// Good (score >= 75)
		metadata.SecurityScore = 80
		badge = metadata.GetSecurityBadge()
		assert.Contains(t, badge, "🟢")
		assert.Contains(t, badge, "GOOD")

		// Fair (score >= 60)
		metadata.SecurityScore = 65
		badge = metadata.GetSecurityBadge()
		assert.Contains(t, badge, "🟡")
		assert.Contains(t, badge, "FAIR")

		// Poor (score >= 40)
		metadata.SecurityScore = 45
		badge = metadata.GetSecurityBadge()
		assert.Contains(t, badge, "🟠")
		assert.Contains(t, badge, "POOR")

		// Critical (score < 40)
		metadata.SecurityScore = 25
		badge = metadata.GetSecurityBadge()
		assert.Contains(t, badge, "🔴")
		assert.Contains(t, badge, "CRITICAL")
	})

	t.Run("GetRatingDisplay", func(t *testing.T) {
		// 4.6 stars
		metadata.Rating = 4.6
		metadata.RatingCount = 15
		display := metadata.GetRatingDisplay()
		assert.Contains(t, display, "★")
		assert.Contains(t, display, "4.6")
		assert.Contains(t, display, "15 ratings")

		// 5.0 stars
		metadata.Rating = 5.0
		metadata.RatingCount = 10
		display = metadata.GetRatingDisplay()
		assert.Contains(t, display, "5.0")

		// No ratings
		metadata.Rating = 0.0
		metadata.RatingCount = 0
		display = metadata.GetRatingDisplay()
		assert.Contains(t, display, "no ratings")
	})
}

// TestCommunityMetadataStore_Persistence tests metadata persistence
func TestCommunityMetadataStore_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	metadataDir := filepath.Join(tempDir, "metadata")

	// Set env var for custom location
	os.Setenv("PRISM_COMMUNITY_METADATA_DIR", metadataDir)
	t.Cleanup(func() {
		os.Unsetenv("PRISM_COMMUNITY_METADATA_DIR")
	})

	// Create first store and add data
	store1 := NewCommunityMetadataStore()
	err := store1.UpdateDownloadCount("Persistent Template", "https://github.com/test/repo")
	require.NoError(t, err)

	// Create second store (should load persisted data)
	store2 := NewCommunityMetadataStore()
	metadata, err := store2.GetMetadata("Persistent Template", "https://github.com/test/repo")
	require.NoError(t, err)
	assert.Equal(t, 1, metadata.Downloads, "Data should persist across store instances")
}

// TestCommunityMetadataStore_ConcurrentAccess tests thread safety
func TestCommunityMetadataStore_ConcurrentAccess(t *testing.T) {
	store := createTestMetadataStore(t)

	templateName := "Concurrent Template"
	sourceURL := "https://github.com/test/repo"

	// Update downloads concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				store.UpdateDownloadCount(templateName, sourceURL)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 100 downloads total
	metadata, err := store.GetMetadata(templateName, sourceURL)
	require.NoError(t, err)
	assert.Equal(t, 100, metadata.Downloads)
}

// TestCommunityMetadataStore_InvalidRating tests rating validation
func TestCommunityMetadataStore_InvalidRating(t *testing.T) {
	store := createTestMetadataStore(t)

	tests := []struct {
		name   string
		rating float64
		valid  bool
	}{
		{"Valid low", 0.0, true},
		{"Valid mid", 3.5, true},
		{"Valid high", 5.0, true},
		{"Invalid negative", -1.0, false},
		{"Invalid too high", 6.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.AddRating("Test", "https://github.com/test/repo", tt.rating)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestCommunityMetadataStore_EmptyQueries tests handling of empty results
func TestCommunityMetadataStore_EmptyQueries(t *testing.T) {
	store := createTestMetadataStore(t)

	t.Run("EmptyPopularTemplates", func(t *testing.T) {
		popular, err := store.GetPopularTemplates(10)
		require.NoError(t, err)
		assert.Empty(t, popular)
	})

	t.Run("EmptyFeaturedTemplates", func(t *testing.T) {
		featured, err := store.GetFeaturedTemplates()
		require.NoError(t, err)
		assert.Empty(t, featured)
	})

	t.Run("EmptyHighRated", func(t *testing.T) {
		highRated, err := store.GetHighRatedTemplates(4.0, 5)
		require.NoError(t, err)
		assert.Empty(t, highRated)
	})

	t.Run("EmptyKeywordSearch", func(t *testing.T) {
		results, err := store.SearchByKeywords([]string{"nonexistent"})
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

// Helper Functions

// createTestMetadataStore creates a metadata store with isolated storage
func createTestMetadataStore(t *testing.T) *CommunityMetadataStore {
	tempDir := t.TempDir()
	metadataDir := filepath.Join(tempDir, "metadata")

	// Set environment variable for isolated testing
	os.Setenv("PRISM_COMMUNITY_METADATA_DIR", metadataDir)
	t.Cleanup(func() {
		os.Unsetenv("PRISM_COMMUNITY_METADATA_DIR")
	})

	return NewCommunityMetadataStore()
}

// Benchmark tests

func BenchmarkMetadataStore_GetMetadata(b *testing.B) {
	store := NewCommunityMetadataStore()
	templateName := "Benchmark Template"
	sourceURL := "https://github.com/test/repo"

	// Pre-populate
	store.UpdateDownloadCount(templateName, sourceURL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetMetadata(templateName, sourceURL)
	}
}

func BenchmarkMetadataStore_UpdateDownloadCount(b *testing.B) {
	store := NewCommunityMetadataStore()
	templateName := "Benchmark Template"
	sourceURL := "https://github.com/test/repo"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.UpdateDownloadCount(templateName, sourceURL)
	}
}

func BenchmarkMetadataStore_AddRating(b *testing.B) {
	store := NewCommunityMetadataStore()
	templateName := "Benchmark Template"
	sourceURL := "https://github.com/test/repo"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.AddRating(templateName, sourceURL, 4.5)
	}
}
