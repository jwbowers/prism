package daemon

import (
	"net/http"
)

// RegisterCommunityRoutes registers community template API routes
// Simplified for v0.7.2 - full implementation in v0.7.3
func (s *Server) RegisterCommunityRoutes(mux *http.ServeMux, applyMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	// Community template endpoints - stubbed for v0.7.2
	mux.HandleFunc("/api/v1/community/templates", applyMiddleware(s.handleCommunityTemplatesStub))
	mux.HandleFunc("/api/v1/community/sources", applyMiddleware(s.handleCommunitySourcesStub))
}

// handleCommunityTemplatesStub provides a clear message that the feature is coming
func (s *Server) handleCommunityTemplatesStub(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":         "Community Templates feature coming in v0.7.3",
		"status":          "in_development",
		"eta":             "March 2026",
		"current_version": "v0.7.2",
		"features_planned": []string{
			"GitHub repository integration",
			"Template discovery and search",
			"Security scanning (6 automated checks)",
			"Three-tier trust system",
			"Community ratings and reviews",
		},
	})
}

// handleCommunitySourcesStub provides information about source management
func (s *Server) handleCommunitySourcesStub(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Community template sources coming in v0.7.3",
		"status":  "in_development",
		"sources": []interface{}{},
	})
}
