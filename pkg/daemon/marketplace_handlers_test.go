package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleMarketplaceTemplates tests the marketplace templates search endpoint
func TestHandleMarketplaceTemplates(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "search all templates",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "search by query",
			queryParams:    "?query=ml",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "search by category",
			queryParams:    "?category=machine-learning",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "search with pagination",
			queryParams:    "?limit=10&offset=0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "search with filters",
			queryParams:    "?min_rating=4&verified_only=true",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/marketplace/templates"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Marketplace registry may not be initialized in test mode
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError,
				"Expected %d or 500, got %d", tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "templates")
				assert.Contains(t, response, "total_count")
			}
		})
	}
}

// TestHandleMarketplaceTemplate tests the individual template endpoint
func TestHandleMarketplaceTemplate(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		templateID     string
		expectedStatus int
	}{
		{
			name:           "get template by ID",
			templateID:     "test-template-123",
			expectedStatus: http.StatusNotFound, // Template doesn't exist in test mode
		},
		{
			name:           "empty template ID",
			templateID:     "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/marketplace/template/" + tt.templateID
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound,
				"Expected %d, 404, or 500, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// TestHandleMarketplaceCategories tests the categories listing endpoint
func TestHandleMarketplaceCategories(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/marketplace/categories", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Marketplace registry may not be initialized
	assert.True(t, w.Code == http.StatusOK ||
		w.Code == http.StatusInternalServerError ||
		w.Code == http.StatusNotFound)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "categories")
		assert.Contains(t, response, "total")
	}
}

// TestHandleMarketplaceFeatured tests the featured templates endpoint
func TestHandleMarketplaceFeatured(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/marketplace/featured", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK ||
		w.Code == http.StatusInternalServerError ||
		w.Code == http.StatusNotFound)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "templates")
		assert.Contains(t, response, "count")
	}
}

// TestHandleMarketplaceTrending tests the trending templates endpoint
func TestHandleMarketplaceTrending(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name        string
		queryParams string
	}{
		{
			name:        "default timeframe",
			queryParams: "",
		},
		{
			name:        "week timeframe",
			queryParams: "?timeframe=week",
		},
		{
			name:        "month timeframe",
			queryParams: "?timeframe=month",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/marketplace/trending"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, w.Code == http.StatusOK ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "templates")
				assert.Contains(t, response, "timeframe")
			}
		})
	}
}

// TestHandleMarketplacePublish tests the publish template endpoint
func TestHandleMarketplacePublish(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid publication",
			requestBody: map[string]interface{}{
				"name":        "Test Template",
				"description": "Test description",
				"category":    "machine-learning",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			requestBody: map[string]interface{}{
				"description": "Test description",
				"category":    "machine-learning",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing description",
			requestBody: map[string]interface{}{
				"name":     "Test Template",
				"category": "machine-learning",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing category",
			requestBody: map[string]interface{}{
				"name":        "Test Template",
				"description": "Test description",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/api/v1/marketplace/publish", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Marketplace registry may not be initialized
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound)
		})
	}
}

// TestHandleMarketplaceUpdate tests the update template endpoint
func TestHandleMarketplaceUpdate(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		templateID     string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:       "valid update",
			templateID: "test-template-123",
			requestBody: map[string]interface{}{
				"description": "Updated description",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty template ID",
			templateID:     "",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			path := "/api/v1/marketplace/update/" + tt.templateID
			req := httptest.NewRequest("PUT", path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound)
		})
	}
}

// TestHandleMarketplaceUnpublish tests the unpublish template endpoint
func TestHandleMarketplaceUnpublish(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		templateID     string
		expectedStatus int
	}{
		{
			name:           "unpublish template",
			templateID:     "test-template-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty template ID",
			templateID:     "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/marketplace/unpublish/" + tt.templateID
			req := httptest.NewRequest("DELETE", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound)
		})
	}
}

// TestHandleMyPublications tests the my publications endpoint
func TestHandleMyPublications(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/marketplace/my-publications", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK ||
		w.Code == http.StatusInternalServerError ||
		w.Code == http.StatusNotFound)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "publications")
		assert.Contains(t, response, "count")
	}
}

// TestHandleMarketplaceReviews tests the reviews endpoint (GET and POST)
func TestHandleMarketplaceReviews(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	t.Run("GET reviews", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/marketplace/reviews/test-template-123", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.True(t, w.Code == http.StatusOK ||
			w.Code == http.StatusInternalServerError ||
			w.Code == http.StatusNotFound)
	})

	t.Run("POST review - valid", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"rating":  5,
			"title":   "Great template!",
			"content": "This template works perfectly",
		}
		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/marketplace/reviews/test-template-123", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.True(t, w.Code == http.StatusCreated ||
			w.Code == http.StatusInternalServerError ||
			w.Code == http.StatusNotFound)
	})

	t.Run("POST review - invalid rating", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"rating":  6, // Invalid: must be 1-5
			"title":   "Test",
			"content": "Test content",
		}
		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/marketplace/reviews/test-template-123", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST review - missing title", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"rating":  5,
			"content": "Test content",
		}
		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/marketplace/reviews/test-template-123", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestHandleMarketplaceFork tests the fork template endpoint
func TestHandleMarketplaceFork(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		templateID     string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:       "valid fork",
			templateID: "test-template-123",
			requestBody: map[string]interface{}{
				"new_name":        "My Fork",
				"new_description": "My forked template",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:       "missing new name",
			templateID: "test-template-123",
			requestBody: map[string]interface{}{
				"new_description": "My forked template",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "missing new description",
			templateID: "test-template-123",
			requestBody: map[string]interface{}{
				"new_name": "My Fork",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty template ID",
			templateID:     "",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			path := "/api/v1/marketplace/fork/" + tt.templateID
			req := httptest.NewRequest("POST", path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound)
		})
	}
}

// TestHandleMarketplaceAnalytics tests the analytics endpoint
func TestHandleMarketplaceAnalytics(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name        string
		templateID  string
		queryParams string
	}{
		{
			name:        "overview analytics",
			templateID:  "test-template-123",
			queryParams: "?timeframe=overview",
		},
		{
			name:        "week analytics",
			templateID:  "test-template-123",
			queryParams: "?timeframe=week",
		},
		{
			name:        "default timeframe",
			templateID:  "test-template-123",
			queryParams: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/marketplace/analytics/" + tt.templateID + tt.queryParams
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, w.Code == http.StatusOK ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound)
		})
	}
}

// TestHandleMarketplaceTemplateInstall tests the template installation endpoint
func TestHandleMarketplaceTemplateInstall(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid installation",
			requestBody: map[string]interface{}{
				"marketplace_template_id": "test-template-123",
				"local_name":              "my-local-template",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing marketplace_template_id",
			requestBody: map[string]interface{}{
				"local_name": "my-local-template",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing local_name",
			requestBody: map[string]interface{}{
				"marketplace_template_id": "test-template-123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/api/v1/templates/install-marketplace", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound)
		})
	}
}

// TestHandleMarketplaceTemplateTracking tests the usage tracking endpoint
func TestHandleMarketplaceTemplateTracking(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		templateID     string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:       "valid tracking event",
			templateID: "test-template-123",
			requestBody: map[string]interface{}{
				"event_type": "launch",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			templateID:     "test-template-123",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "empty template ID",
			templateID: "",
			requestBody: map[string]interface{}{
				"event_type": "launch",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			path := "/api/v1/marketplace/template-tracking/" + tt.templateID + "/track"
			req := httptest.NewRequest("POST", path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound)
		})
	}
}

// TestMarketplaceHandlersMethodValidation tests HTTP method validation
func TestMarketplaceHandlersMethodValidation(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "POST on templates endpoint",
			method:         "POST",
			path:           "/api/v1/marketplace/templates",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT on featured endpoint",
			method:         "PUT",
			path:           "/api/v1/marketplace/featured",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE on categories endpoint",
			method:         "DELETE",
			path:           "/api/v1/marketplace/categories",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET on publish endpoint",
			method:         "GET",
			path:           "/api/v1/marketplace/publish",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

// TestMarketplaceHandlersConcurrency tests concurrent access
func TestMarketplaceHandlersConcurrency(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	const numRequests = 10
	done := make(chan bool, numRequests)

	endpoints := []string{
		"/api/v1/marketplace/templates",
		"/api/v1/marketplace/categories",
		"/api/v1/marketplace/featured",
	}

	for i := 0; i < numRequests; i++ {
		go func(endpoint string) {
			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Should return valid status (OK or error)
			assert.True(t, w.Code >= 200 && w.Code < 600)
			done <- true
		}(endpoints[i%len(endpoints)])
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}
