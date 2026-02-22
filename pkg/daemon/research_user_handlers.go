package daemon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/scttfrdmn/prism/pkg/profile"
	"github.com/scttfrdmn/prism/pkg/research"
)

// ResearchUserRequest represents a request to create a research user
type ResearchUserRequest struct {
	Username    string `json:"username"`
	FullName    string `json:"full_name,omitempty"`
	DisplayName string `json:"display_name,omitempty"` // Accept display_name from frontend
	Email       string `json:"email,omitempty"`
}

// ResearchUserSSHKeyRequest represents a request to manage SSH keys
type ResearchUserSSHKeyRequest struct {
	Username string `json:"username"`
	KeyType  string `json:"key_type,omitempty"` // "ed25519" or "rsa"
}

// handleResearchUsers handles research user collection operations
func (s *Server) handleResearchUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListResearchUsers(w, r)
	case http.MethodPost:
		s.handleCreateResearchUser(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleListResearchUsers lists all research users
func (s *Server) handleListResearchUsers(w http.ResponseWriter, r *http.Request) {
	service, err := s.getResearchUserService()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize research user service: %v", err))
		return
	}

	users, err := service.ListResearchUsers()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list research users: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(users)
}

// handleCreateResearchUser creates a new research user
func (s *Server) handleCreateResearchUser(w http.ResponseWriter, r *http.Request) {
	var req ResearchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" {
		s.writeError(w, http.StatusBadRequest, "Username is required")
		return
	}

	service, err := s.getResearchUserService()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize research user service: %v", err))
		return
	}

	// Check if user already exists
	_, err = service.GetResearchUser(req.Username)
	if err == nil {
		// User already exists, return 409 Conflict
		s.writeError(w, http.StatusConflict, fmt.Sprintf("User with username '%s' already exists", req.Username))
		return
	}

	// Use DisplayName if FullName is not provided (for frontend compatibility)
	fullName := req.FullName
	if fullName == "" && req.DisplayName != "" {
		fullName = req.DisplayName
	}

	// Create new user
	user, err := service.CreateResearchUser(req.Username, &research.CreateResearchUserOptions{
		FullName:       fullName,
		Email:          req.Email,
		GenerateSSHKey: true, // Generate SSH key by default
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create research user: %v", err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

// handleResearchUserOperations handles individual research user operations
func (s *Server) handleResearchUserOperations(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/api/v1/users/"):]
	parts := splitPath(path)
	if len(parts) == 0 {
		s.writeError(w, http.StatusBadRequest, "Missing username")
		return
	}

	username := parts[0]

	if len(parts) == 1 {
		// Operations on the user itself: GET /api/v1/users/{username}
		switch r.Method {
		case http.MethodGet:
			s.handleGetResearchUser(w, r, username)
		case http.MethodDelete:
			s.handleDeleteResearchUser(w, r, username)
		default:
			s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	} else if len(parts) == 2 {
		// Sub-operations: /api/v1/users/{username}/{operation}
		operation := parts[1]
		switch operation {
		case "ssh-key":
			s.handleResearchUserSSHKey(w, r, username)
		case "status":
			s.handleResearchUserStatus(w, r, username)
		case "enable":
			s.handleEnableResearchUser(w, r, username)
		case "disable":
			s.handleDisableResearchUser(w, r, username)
		case "provision":
			s.handleProvisionResearchUser(w, r, username)
		default:
			s.writeError(w, http.StatusNotFound, "Unknown operation")
		}
	} else {
		s.writeError(w, http.StatusNotFound, "Invalid path")
	}
}

// handleGetResearchUser gets details for a specific research user
func (s *Server) handleGetResearchUser(w http.ResponseWriter, r *http.Request, username string) {
	service, err := s.getResearchUserService()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize research user service: %v", err))
		return
	}

	user, err := service.GetResearchUser(username)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Research user not found: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(user)
}

// handleDeleteResearchUser deletes a research user
func (s *Server) handleDeleteResearchUser(w http.ResponseWriter, r *http.Request, username string) {
	service, err := s.getResearchUserService()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize research user service: %v", err))
		return
	}

	// Get current profile
	profileID, err := s.getCurrentProfile()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get current profile: %v", err))
		return
	}

	// Delete the research user
	err = service.DeleteResearchUser(profileID, username)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete research user: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleResearchUserSSHKey handles SSH key operations for research users
func (s *Server) handleResearchUserSSHKey(w http.ResponseWriter, r *http.Request, username string) {
	switch r.Method {
	case http.MethodPost:
		s.handleGenerateResearchUserSSHKey(w, r, username)
	case http.MethodGet:
		s.handleListResearchUserSSHKeys(w, r, username)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleGenerateResearchUserSSHKey generates SSH keys for a research user
func (s *Server) handleGenerateResearchUserSSHKey(w http.ResponseWriter, r *http.Request, username string) {
	var req ResearchUserSSHKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set default key type if not specified
	if req.KeyType == "" {
		req.KeyType = "ed25519"
	}

	service, err := s.getResearchUserService()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize research user service: %v", err))
		return
	}

	keyPair, privateKey, err := service.ManageSSHKeys().GenerateKeyPair(username, req.KeyType)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate SSH key: %v", err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"username":    username,
		"key_type":    req.KeyType,
		"public_key":  keyPair.PublicKey,
		"private_key": string(privateKey),
		"fingerprint": keyPair.Fingerprint,
	})
}

// handleListResearchUserSSHKeys lists SSH keys for a research user
func (s *Server) handleListResearchUserSSHKeys(w http.ResponseWriter, r *http.Request, username string) {
	service, err := s.getResearchUserService()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize research user service: %v", err))
		return
	}

	keys, err := service.ManageSSHKeys().ListKeys(username)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list SSH keys: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"username": username,
		"keys":     keys,
	})
}

// handleResearchUserStatus gets detailed status for a research user
func (s *Server) handleResearchUserStatus(w http.ResponseWriter, r *http.Request, username string) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	service, err := s.getResearchUserService()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize research user service: %v", err))
		return
	}

	user, err := service.GetResearchUser(username)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Research user not found: %v", err))
		return
	}

	// Get current profile for context
	currentProfile, err := s.getCurrentProfile()
	if err != nil {
		currentProfile = "default" // Fallback
	}

	// Create detailed status response
	status := map[string]interface{}{
		"user":           user,
		"profile":        currentProfile,
		"ssh_keys_count": len(user.SSHPublicKeys),
		"status":         "active",
		"last_updated":   user.CreatedAt,
	}

	_ = json.NewEncoder(w).Encode(status)
}

// getResearchUserService creates a research user service instance
func (s *Server) getResearchUserService() (*research.ResearchUserService, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".prism")

	// Create profile adapter for daemon
	profileAdapter := &DaemonProfileAdapter{}

	// Create research user service with full functionality
	serviceConfig := &research.ResearchUserServiceConfig{
		ConfigDir:  configDir,
		ProfileMgr: profileAdapter,
	}

	return research.NewResearchUserService(serviceConfig), nil
}

// getCurrentProfile gets the current profile name
func (s *Server) getCurrentProfile() (string, error) {
	profileManager, err := profile.NewManagerEnhanced()
	if err != nil {
		return "default", nil // Fallback to default
	}

	profile, err := profileManager.GetCurrentProfile()
	if err != nil {
		return "default", nil // Fallback to default
	}

	return profile.Name, nil
}

// DaemonProfileAdapter adapts the profile manager for daemon use
type DaemonProfileAdapter struct{}

func (d *DaemonProfileAdapter) GetCurrentProfile() (string, error) {
	profileManager, err := profile.NewManagerEnhanced()
	if err != nil {
		return "default", nil
	}

	profile, err := profileManager.GetCurrentProfile()
	if err != nil {
		return "default", nil
	}

	return profile.Name, nil
}

func (d *DaemonProfileAdapter) GetProfileConfig(profileID string) (interface{}, error) {
	// Get profile using enhanced profile manager
	profileManager, err := profile.NewManagerEnhanced()
	if err != nil {
		return nil, fmt.Errorf("failed to create profile manager: %w", err)
	}

	profileConfig, err := profileManager.GetProfile(profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile config: %w", err)
	}

	return profileConfig, nil
}

func (d *DaemonProfileAdapter) UpdateProfileConfig(profileID string, config interface{}) error {
	// Update profile using enhanced profile manager
	profileManager, err := profile.NewManagerEnhanced()
	if err != nil {
		return fmt.Errorf("failed to create profile manager: %w", err)
	}

	// Convert config to profile.Profile if needed
	if profileConfig, ok := config.(*profile.Profile); ok {
		return profileManager.UpdateProfile(profileID, *profileConfig)
	}

	return fmt.Errorf("invalid profile config type")
}

// handleEnableResearchUser enables a research user account
func (s *Server) handleEnableResearchUser(w http.ResponseWriter, r *http.Request, username string) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	service, err := s.getResearchUserService()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize research user service: %v", err))
		return
	}

	// Get user to verify existence
	user, err := service.GetResearchUser(username)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Research user not found: %v", err))
		return
	}

	// Get current profile ID
	profileID, err := s.getCurrentProfile()
	if err != nil {
		profileID = "default" // Fallback to default
	}

	// Update user enabled status
	user.Enabled = true
	err = service.UpdateResearchUser(profileID, user)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to enable user: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleDisableResearchUser disables a research user account
func (s *Server) handleDisableResearchUser(w http.ResponseWriter, r *http.Request, username string) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	service, err := s.getResearchUserService()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize research user service: %v", err))
		return
	}

	// Get user to verify existence
	user, err := service.GetResearchUser(username)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Research user not found: %v", err))
		return
	}

	// Get current profile ID
	profileID, err := s.getCurrentProfile()
	if err != nil {
		profileID = "default" // Fallback to default
	}

	// Update user enabled status
	user.Enabled = false
	err = service.UpdateResearchUser(profileID, user)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to disable user: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleProvisionResearchUser provisions a research user on a workspace instance
func (s *Server) handleProvisionResearchUser(w http.ResponseWriter, r *http.Request, username string) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Instance string `json:"instance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if req.Instance == "" {
		s.writeError(w, http.StatusBadRequest, "instance is required")
		return
	}

	// In test mode, return success without actually provisioning via SSM
	if s.testMode {
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"success":  true,
			"username": username,
			"instance": req.Instance,
			"message":  fmt.Sprintf("User '%s' provisioned on workspace '%s' (test mode)", username, req.Instance),
		})
		return
	}

	// In production, provisioning requires SSM and a running instance
	s.writeError(w, http.StatusNotImplemented, "Workspace provisioning via SSM is not yet implemented in the daemon.")
}
