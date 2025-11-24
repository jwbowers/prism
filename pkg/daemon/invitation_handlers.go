// Package daemon provides invitation management API handlers for Prism v0.5.11+
//
// This file implements REST API endpoints for the user invitation system,
// enabling project owners to invite collaborators via email.
package daemon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/scttfrdmn/prism/pkg/invitation"
	"github.com/scttfrdmn/prism/pkg/research"
	"github.com/scttfrdmn/prism/pkg/types"
)

// handleInvitationOperations routes invitation-related HTTP requests
// This handles all invitation endpoints with path-based routing
func (s *Server) handleInvitationOperations(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	// Route shared token operations (v0.5.13)
	if strings.Contains(path, "/shared") || (strings.HasPrefix(path, "/api/v1/invitations/shared/")) {
		s.handleSharedTokenOperations(w, r)
		return
	}

	// Determine which endpoint is being called based on path pattern
	switch {
	// POST /api/v1/projects/{id}/invitations/bulk - Bulk invitation
	case strings.HasPrefix(path, "/api/v1/projects/") && strings.HasSuffix(path, "/invitations/bulk") && method == http.MethodPost:
		s.handleBulkInvitation(w, r)

	// POST /api/v1/projects/{id}/invitations - Send invitation
	case strings.HasPrefix(path, "/api/v1/projects/") && strings.HasSuffix(path, "/invitations") && method == http.MethodPost:
		s.handleSendInvitation(w, r)

	// GET /api/v1/projects/{id}/invitations - List project invitations
	case strings.HasPrefix(path, "/api/v1/projects/") && strings.HasSuffix(path, "/invitations") && method == http.MethodGet:
		s.handleListProjectInvitations(w, r)

	// GET /api/v1/invitations/my - List user's received invitations
	case path == "/api/v1/invitations/my" && method == http.MethodGet:
		s.handleListMyInvitations(w, r)

	// POST /api/v1/invitations/{token}/accept - Accept invitation
	case strings.Contains(path, "/accept") && method == http.MethodPost:
		s.handleAcceptInvitation(w, r)

	// POST /api/v1/invitations/{token}/decline - Decline invitation
	case strings.Contains(path, "/decline") && method == http.MethodPost:
		s.handleDeclineInvitation(w, r)

	// POST /api/v1/invitations/{id}/resend - Resend invitation
	case strings.Contains(path, "/resend") && method == http.MethodPost:
		s.handleResendInvitation(w, r)

	// DELETE /api/v1/invitations/{id} - Revoke invitation
	case strings.HasPrefix(path, "/api/v1/invitations/") && method == http.MethodDelete:
		s.handleRevokeInvitation(w, r)

	// POST /api/v1/invitations/quota-check - Check AWS quota for bulk invitations
	case path == "/api/v1/invitations/quota-check" && method == http.MethodPost:
		s.handleQuotaCheck(w, r)

	// GET /api/v1/invitations/{token} - Get invitation by token
	case strings.HasPrefix(path, "/api/v1/invitations/") && method == http.MethodGet:
		s.handleGetInvitation(w, r)

	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleSendInvitation creates and sends a new project invitation
// POST /api/v1/projects/{id}/invitations
func (s *Server) handleSendInvitation(w http.ResponseWriter, r *http.Request) {
	// Parse project ID from path: /api/v1/projects/{id}/invitations
	path := r.URL.Path[len("/api/v1/projects/"):]
	parts := splitPath(path)
	if len(parts) < 2 || parts[1] != "invitations" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	projectID := parts[0]

	var req invitation.CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Set project ID from URL
	req.ProjectID = projectID

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

	// TODO: Check if requester has permission to invite (must be owner/admin)
	// For now, we'll implement basic validation

	// Create invitation
	inv, err := s.invitationManager.CreateInvitation(r.Context(), &req)
	if err != nil {
		if err == types.ErrDuplicateInvitation {
			http.Error(w, "An invitation for this email already exists for this project", http.StatusConflict)
			return
		}
		if err == types.ErrInvalidEmail || err == types.ErrInvalidRole || err == types.ErrInvalidInviter {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to create invitation: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: Send invitation email via EmailSender
	// For now, we'll just return the invitation with the token (for testing)
	// In production, token should only be sent via email

	response := map[string]interface{}{
		"invitation": inv,
		"project":    project,
		"message":    "Invitation created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleListProjectInvitations lists all invitations for a project
// GET /api/v1/projects/{id}/invitations?status=pending&limit=50&offset=0
func (s *Server) handleListProjectInvitations(w http.ResponseWriter, r *http.Request) {
	// Parse project ID from path: /api/v1/projects/{id}/invitations
	path := r.URL.Path[len("/api/v1/projects/"):]
	parts := splitPath(path)
	if len(parts) < 2 || parts[1] != "invitations" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	projectID := parts[0]

	// Validate project exists
	_, err := s.projectManager.GetProject(r.Context(), projectID)
	if err != nil {
		if err == types.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get project: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: Check if requester has permission to view invitations (must be member)

	// Parse query parameters
	filter := &types.InvitationFilter{
		ProjectID: projectID,
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = types.InvitationStatus(status)
	}

	if includeExpired := r.URL.Query().Get("include_expired"); includeExpired == "true" {
		filter.IncludeExpired = true
	}

	// Parse pagination
	if limit := r.URL.Query().Get("limit"); limit != "" {
		fmt.Sscanf(limit, "%d", &filter.Limit)
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		fmt.Sscanf(offset, "%d", &filter.Offset)
	}

	// List invitations
	invitations, err := s.invitationManager.ListInvitations(r.Context(), filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list invitations: %v", err), http.StatusInternalServerError)
		return
	}

	// Get summary
	summary, err := s.invitationManager.GetProjectSummary(r.Context(), projectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get invitation summary: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"invitations": invitations,
		"summary":     summary,
		"filter":      filter,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListMyInvitations lists invitations received by the current user
// GET /api/v1/invitations/my?email=user@example.com&status=pending
func (s *Server) handleListMyInvitations(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter required", http.StatusBadRequest)
		return
	}

	filter := &types.InvitationFilter{
		Email:          email,
		IncludeExpired: false,
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = types.InvitationStatus(status)
	}

	// Parse pagination
	if limit := r.URL.Query().Get("limit"); limit != "" {
		fmt.Sscanf(limit, "%d", &filter.Limit)
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		fmt.Sscanf(offset, "%d", &filter.Offset)
	}

	// List invitations
	invitations, err := s.invitationManager.ListInvitations(r.Context(), filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list invitations: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"invitations": invitations,
		"email":       email,
		"filter":      filter,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetInvitation retrieves an invitation by token (public endpoint)
// GET /api/v1/invitations/{token}
func (s *Server) handleGetInvitation(w http.ResponseWriter, r *http.Request) {
	// Parse token from path: /api/v1/invitations/{token}
	token := r.URL.Path[len("/api/v1/invitations/"):]
	if token == "" {
		http.Error(w, "Missing invitation token", http.StatusBadRequest)
		return
	}

	inv, err := s.invitationManager.GetInvitationByToken(r.Context(), token)
	if err != nil {
		if err == types.ErrInvitationNotFound {
			http.Error(w, "Invitation not found", http.StatusNotFound)
			return
		}
		if err == types.ErrInvitationExpired {
			http.Error(w, "Invitation has expired", http.StatusGone)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get invitation: %v", err), http.StatusInternalServerError)
		return
	}

	// Get project details for context
	project, err := s.projectManager.GetProject(r.Context(), inv.ProjectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get project: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"invitation": inv,
		"project":    project,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAcceptInvitation accepts an invitation and adds user to project
// POST /api/v1/invitations/{token}/accept
func (s *Server) handleAcceptInvitation(w http.ResponseWriter, r *http.Request) {
	// Parse token from path: /api/v1/invitations/{token}/accept
	path := r.URL.Path[len("/api/v1/invitations/"):]
	parts := splitPath(path)
	if len(parts) < 2 || parts[1] != "accept" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	token := parts[0]

	// Accept invitation
	inv, err := s.invitationManager.AcceptInvitation(r.Context(), token)
	if err != nil {
		if err == types.ErrInvitationNotFound {
			http.Error(w, "Invitation not found", http.StatusNotFound)
			return
		}
		if err == types.ErrInvitationExpired {
			http.Error(w, "Invitation has expired", http.StatusGone)
			return
		}
		if err == types.ErrInvitationAlreadyUsed {
			http.Error(w, "Invitation has already been used", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to accept invitation: %v", err), http.StatusInternalServerError)
		return
	}

	// Add user to project members (Issue #102)
	member := &types.ProjectMember{
		UserID:  inv.Email, // Use invitation email as user ID
		Role:    inv.Role,  // Use role from invitation
		AddedBy: inv.InvitedBy,
	}

	if err := s.projectManager.AddProjectMember(r.Context(), inv.ProjectID, member); err != nil {
		// Handle duplicate member case gracefully (user may already be added)
		if !strings.Contains(err.Error(), "already a member") {
			http.Error(w, fmt.Sprintf("Failed to add member to project: %v", err), http.StatusInternalServerError)
			return
		}
		// If already a member, that's fine - continue with acceptance response
	}

	// Auto-provision research user (Issue #106)
	// Extract username from email (e.g., "user@example.com" → "user")
	username := strings.Split(inv.Email, "@")[0]

	var researchUser interface{}
	var provisioningError string

	if researchUserService, err := s.getResearchUserService(); err == nil {
		// Try to create or get existing research user
		user, err := researchUserService.CreateResearchUser(username, &research.CreateResearchUserOptions{
			GenerateSSHKey: true, // Automatically generate SSH keys
		})
		if err != nil {
			// Log error but don't fail invitation acceptance
			provisioningError = fmt.Sprintf("Research user provisioning failed: %v", err)
		} else {
			researchUser = user
		}
	} else {
		provisioningError = fmt.Sprintf("Failed to initialize research user service: %v", err)
	}

	// Get project details
	project, err := s.projectManager.GetProject(r.Context(), inv.ProjectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get project: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: Send acceptance confirmation email via EmailSender

	// Build response with provisioning status
	response := map[string]interface{}{
		"invitation": inv,
		"project":    project,
		"message":    "Invitation accepted successfully",
	}

	// Add provisioning information to response
	if researchUser != nil {
		response["research_user"] = researchUser
		response["provisioning_status"] = "success"
	} else if provisioningError != "" {
		response["provisioning_status"] = "failed"
		response["provisioning_error"] = provisioningError
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDeclineInvitation declines an invitation
// POST /api/v1/invitations/{token}/decline
func (s *Server) handleDeclineInvitation(w http.ResponseWriter, r *http.Request) {
	// Parse token from path: /api/v1/invitations/{token}/decline
	path := r.URL.Path[len("/api/v1/invitations/"):]
	parts := splitPath(path)
	if len(parts) < 2 || parts[1] != "decline" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	token := parts[0]

	// Parse optional decline reason
	var req struct {
		Reason string `json:"reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Ignore decode errors for optional body
	}

	// Decline invitation
	inv, err := s.invitationManager.DeclineInvitation(r.Context(), token, req.Reason)
	if err != nil {
		if err == types.ErrInvitationNotFound {
			http.Error(w, "Invitation not found", http.StatusNotFound)
			return
		}
		if err == types.ErrInvitationExpired {
			http.Error(w, "Invitation has expired", http.StatusGone)
			return
		}
		if err == types.ErrInvitationAlreadyUsed {
			http.Error(w, "Invitation has already been used", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to decline invitation: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"invitation": inv,
		"message":    "Invitation declined",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleResendInvitation resends an invitation email
// POST /api/v1/invitations/{id}/resend
func (s *Server) handleResendInvitation(w http.ResponseWriter, r *http.Request) {
	// Parse invitation ID from path: /api/v1/invitations/{id}/resend
	path := r.URL.Path[len("/api/v1/invitations/"):]
	parts := splitPath(path)
	if len(parts) < 2 || parts[1] != "resend" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	invitationID := parts[0]

	// Get invitation
	inv, err := s.invitationManager.GetInvitation(r.Context(), invitationID)
	if err != nil {
		if err == types.ErrInvitationNotFound {
			http.Error(w, "Invitation not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get invitation: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: Check if requester has permission to resend (must be owner/admin of project)

	// Resend invitation
	if err := s.invitationManager.ResendInvitation(r.Context(), invitationID); err != nil {
		if err == types.ErrInvitationNotPending {
			http.Error(w, "Cannot resend: invitation is not pending", http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to resend invitation: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: Send invitation email via EmailSender

	response := map[string]interface{}{
		"invitation": inv,
		"message":    "Invitation resent successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRevokeInvitation revokes a pending invitation
// DELETE /api/v1/invitations/{id}
func (s *Server) handleRevokeInvitation(w http.ResponseWriter, r *http.Request) {
	// Parse invitation ID from path: /api/v1/invitations/{id}
	invitationID := r.URL.Path[len("/api/v1/invitations/"):]
	if invitationID == "" {
		http.Error(w, "Missing invitation ID", http.StatusBadRequest)
		return
	}

	// Get invitation for project context
	inv, err := s.invitationManager.GetInvitation(r.Context(), invitationID)
	if err != nil {
		if err == types.ErrInvitationNotFound {
			http.Error(w, "Invitation not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get invitation: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: Check if requester has permission to revoke (must be owner/admin of project)

	// Revoke invitation
	if err := s.invitationManager.RevokeInvitation(r.Context(), invitationID); err != nil {
		if err == types.ErrInvitationNotPending {
			http.Error(w, "Cannot revoke: invitation is not pending", http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to revoke invitation: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"invitation": inv,
		"message":    "Invitation revoked successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleBulkInvitation creates multiple invitations from a bulk request
// POST /api/v1/projects/{id}/invitations/bulk
func (s *Server) handleBulkInvitation(w http.ResponseWriter, r *http.Request) {
	// Parse project ID from path: /api/v1/projects/{id}/invitations/bulk
	path := r.URL.Path[len("/api/v1/projects/"):]
	parts := splitPath(path)
	if len(parts) < 2 || parts[1] != "invitations" {
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

	// TODO: Check if requester has permission to invite (must be owner/admin)

	// Parse bulk invitation request
	var bulkReq types.BulkInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&bulkReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if len(bulkReq.Invitations) == 0 {
		http.Error(w, "No invitations provided", http.StatusBadRequest)
		return
	}

	// Set default role if not specified
	if bulkReq.DefaultRole == "" {
		bulkReq.DefaultRole = types.ProjectRoleMember
	}

	// TODO: Get inviter from authenticated user
	// For now, use project owner or first admin
	invitedBy := "system" // Placeholder

	// Create bulk invitations
	bulkResp, err := s.invitationManager.CreateBulkInvitations(r.Context(), projectID, invitedBy, &bulkReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create bulk invitations: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: Send invitation emails via EmailSender for all successful invitations
	// For now, tokens are returned in the response (for testing)

	response := map[string]interface{}{
		"summary": bulkResp.Summary,
		"results": bulkResp.Results,
		"message": bulkResp.Message,
		"project": project,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// QuotaCheckRequest represents a request to check AWS quotas for bulk invitations
type QuotaCheckRequest struct {
	InstanceType string `json:"instance_type"`
	Count        int    `json:"count"` // Number of students/invitations
	TemplateName string `json:"template_name,omitempty"`
}

// handleQuotaCheck checks AWS quotas for bulk invitation launch requirements
// POST /api/v1/invitations/quota-check
func (s *Server) handleQuotaCheck(w http.ResponseWriter, r *http.Request) {
	var req QuotaCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.InstanceType == "" {
		http.Error(w, "instance_type is required", http.StatusBadRequest)
		return
	}
	if req.Count <= 0 {
		http.Error(w, "count must be greater than 0", http.StatusBadRequest)
		return
	}

	// Get AWS manager
	if s.awsManager == nil {
		http.Error(w, "AWS manager not initialized", http.StatusInternalServerError)
		return
	}

	// Check quota
	quotaCheck, err := s.awsManager.CheckQuotaForInvitations(r.Context(), req.InstanceType, req.Count)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check quota: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	response := map[string]interface{}{
		"has_sufficient_quota": quotaCheck.HasSufficientQuota,
		"required_vcpus":       quotaCheck.RequiredVCPUs,
		"current_usage":        quotaCheck.CurrentUsage,
		"quota_limit":          quotaCheck.QuotaLimit,
		"available_vcpus":      quotaCheck.AvailableVCPUs,
		"instance_type":        quotaCheck.InstanceType,
	}

	if !quotaCheck.HasSufficientQuota {
		response["warning"] = quotaCheck.WarningMessage
		w.WriteHeader(http.StatusPreconditionFailed) // 412 status for insufficient quota
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
