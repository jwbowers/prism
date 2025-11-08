// Package daemon provides shared token management API handlers for Prism v0.5.13+
//
// This file implements REST API endpoints for the shared token system,
// enabling project owners to create tokens that multiple users can redeem.
package daemon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
)

// handleSharedTokenOperations routes shared token-related HTTP requests
func (s *Server) handleSharedTokenOperations(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	// Determine which endpoint is being called
	switch {
	// POST /api/v1/projects/{id}/invitations/shared - Create shared token
	case strings.HasPrefix(path, "/api/v1/projects/") && strings.HasSuffix(path, "/invitations/shared") && method == http.MethodPost:
		s.handleCreateSharedToken(w, r)

	// GET /api/v1/projects/{id}/invitations/shared - List project shared tokens
	case strings.HasPrefix(path, "/api/v1/projects/") && strings.HasSuffix(path, "/invitations/shared") && method == http.MethodGet:
		s.handleListSharedTokens(w, r)

	// POST /api/v1/invitations/shared/redeem - Redeem shared token
	case path == "/api/v1/invitations/shared/redeem" && method == http.MethodPost:
		s.handleRedeemSharedToken(w, r)

	// GET /api/v1/invitations/shared/{token} - Get shared token info
	case strings.HasPrefix(path, "/api/v1/invitations/shared/") && !strings.Contains(path, "/qr") && !strings.Contains(path, "/extend") && method == http.MethodGet:
		s.handleGetSharedToken(w, r)

	// GET /api/v1/invitations/shared/{token}/qr - Generate QR code
	case strings.Contains(path, "/qr") && method == http.MethodGet:
		s.handleGenerateQRCode(w, r)

	// PATCH /api/v1/invitations/shared/{token}/extend - Extend expiration
	case strings.Contains(path, "/extend") && method == http.MethodPatch:
		s.handleExtendSharedToken(w, r)

	// DELETE /api/v1/invitations/shared/{token} - Revoke shared token
	case strings.HasPrefix(path, "/api/v1/invitations/shared/") && method == http.MethodDelete:
		s.handleRevokeSharedToken(w, r)

	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleCreateSharedToken creates a new shared invitation token
// POST /api/v1/projects/{id}/invitations/shared
func (s *Server) handleCreateSharedToken(w http.ResponseWriter, r *http.Request) {
	// Parse project ID from path
	path := r.URL.Path[len("/api/v1/projects/"):]
	parts := splitPath(path)
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	projectID := parts[0]

	// Validate project exists
	project, err := s.projectManager.GetProject(r.Context(), projectID)
	if err != nil {
		if err == types.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get project: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse request
	var req types.CreateSharedTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Set project ID
	req.ProjectID = projectID

	// Set default role if not specified
	if req.Role == "" {
		req.Role = types.ProjectRoleMember
	}

	// Set creator (TODO: Get from authenticated user)
	req.CreatedBy = "system"

	// Create shared token
	token, err := s.sharedTokenManager.CreateSharedToken(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create shared token: %v", err), http.StatusBadRequest)
		return
	}

	// Build response
	redemptionURL := fmt.Sprintf("https://prism.dev/join/%s", token.Token)
	qrCodeURL := fmt.Sprintf("/api/v1/invitations/shared/%s/qr", token.Token)

	response := map[string]interface{}{
		"token":            token.Token,
		"id":               token.ID,
		"name":             token.Name,
		"role":             token.Role,
		"redemption_limit": token.RedemptionLimit,
		"redemptions":      token.RedemptionCount,
		"status":           token.Status,
		"expires_at":       token.ExpiresAt,
		"created_at":       token.CreatedAt,
		"redemption_url":   redemptionURL,
		"qr_code_url":      qrCodeURL,
		"project":          project,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleRedeemSharedToken redeems a shared token
// POST /api/v1/invitations/shared/redeem
func (s *Server) handleRedeemSharedToken(w http.ResponseWriter, r *http.Request) {
	var req types.RedeemSharedTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Redeem token
	resp, err := s.sharedTokenManager.RedeemToken(r.Context(), &req)
	if err != nil {
		if err == types.ErrSharedTokenNotFound {
			http.Error(w, "Token not found", http.StatusNotFound)
			return
		}
		if err == types.ErrSharedTokenExpired {
			http.Error(w, "Token has expired", http.StatusBadRequest)
			return
		}
		if err == types.ErrSharedTokenRevoked {
			http.Error(w, "Token has been revoked", http.StatusBadRequest)
			return
		}
		if err == types.ErrSharedTokenExhausted {
			http.Error(w, "Token has reached redemption limit", http.StatusBadRequest)
			return
		}
		if err == types.ErrAlreadyRedeemed {
			http.Error(w, "You have already redeemed this token", http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to redeem token: %v", err), http.StatusBadRequest)
		return
	}

	// TODO: Add user to project with specified role
	// This would integrate with the project membership system

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleGetSharedToken retrieves shared token information
// GET /api/v1/invitations/shared/{token}
func (s *Server) handleGetSharedToken(w http.ResponseWriter, r *http.Request) {
	// Parse token from path
	path := r.URL.Path[len("/api/v1/invitations/shared/"):]
	tokenCode := strings.TrimSuffix(path, "/")

	// Get token
	token, err := s.sharedTokenManager.GetSharedToken(r.Context(), tokenCode)
	if err != nil {
		if err == types.ErrSharedTokenNotFound {
			http.Error(w, "Token not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

// handleListSharedTokens lists all shared tokens for a project
// GET /api/v1/projects/{id}/invitations/shared
func (s *Server) handleListSharedTokens(w http.ResponseWriter, r *http.Request) {
	// Parse project ID from path
	path := r.URL.Path[len("/api/v1/projects/"):]
	parts := splitPath(path)
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	projectID := parts[0]

	// List tokens
	tokens, err := s.sharedTokenManager.ListSharedTokens(r.Context(), projectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list tokens: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tokens": tokens,
		"count":  len(tokens),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleExtendSharedToken extends the expiration of a shared token
// PATCH /api/v1/invitations/shared/{token}/extend
func (s *Server) handleExtendSharedToken(w http.ResponseWriter, r *http.Request) {
	// Parse token from path
	path := r.URL.Path[len("/api/v1/invitations/shared/"):]
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	tokenCode := parts[0]

	// Parse request
	var req struct {
		AddDays int `json:"add_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.AddDays <= 0 {
		http.Error(w, "add_days must be positive", http.StatusBadRequest)
		return
	}

	// Extend expiration
	duration := time.Duration(req.AddDays) * 24 * time.Hour
	err := s.sharedTokenManager.ExtendExpiration(r.Context(), tokenCode, duration)
	if err != nil {
		if err == types.ErrSharedTokenNotFound {
			http.Error(w, "Token not found", http.StatusNotFound)
			return
		}
		if err == types.ErrSharedTokenRevoked {
			http.Error(w, "Cannot extend revoked token", http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to extend token: %v", err), http.StatusInternalServerError)
		return
	}

	// Get updated token
	token, _ := s.sharedTokenManager.GetSharedToken(r.Context(), tokenCode)

	response := map[string]interface{}{
		"success":    true,
		"token":      tokenCode,
		"expires_at": token.ExpiresAt,
		"message":    fmt.Sprintf("Token expiration extended by %d days", req.AddDays),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRevokeSharedToken revokes a shared token
// DELETE /api/v1/invitations/shared/{token}
func (s *Server) handleRevokeSharedToken(w http.ResponseWriter, r *http.Request) {
	// Parse token from path
	path := r.URL.Path[len("/api/v1/invitations/shared/"):]
	tokenCode := strings.TrimSuffix(path, "/")

	// Revoke token
	err := s.sharedTokenManager.RevokeSharedToken(r.Context(), tokenCode)
	if err != nil {
		if err == types.ErrSharedTokenNotFound {
			http.Error(w, "Token not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to revoke token: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"token":   tokenCode,
		"message": "Token revoked successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGenerateQRCode generates a QR code for a shared token
// GET /api/v1/invitations/shared/{token}/qr
func (s *Server) handleGenerateQRCode(w http.ResponseWriter, r *http.Request) {
	// Parse token from path
	path := r.URL.Path[len("/api/v1/invitations/shared/"):]
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	tokenCode := parts[0]

	// Verify token exists
	_, err := s.sharedTokenManager.GetSharedToken(r.Context(), tokenCode)
	if err != nil {
		if err == types.ErrSharedTokenNotFound {
			http.Error(w, "Token not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get token: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: Generate QR code for redemption URL
	// For now, return placeholder response
	// This will be implemented in Phase 4

	response := map[string]interface{}{
		"token":          tokenCode,
		"redemption_url": fmt.Sprintf("https://prism.dev/join/%s", tokenCode),
		"qr_code_data":   "base64_encoded_qr_code_image",
		"message":        "QR code generation will be implemented in Phase 4",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
