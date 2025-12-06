package usermgmt

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// userManagementService provides a real in-memory implementation of UserManagementService
// This replaces the stub implementation in types.go
type userManagementService struct {
	storage   UserStorage
	providers map[Provider]UserManagementProvider
	mu        sync.Mutex // Protects critical sections like CreateUser to prevent race conditions
}

// NewUserManagementService creates a new user management service with real implementations
func NewUserManagementService(storage UserStorage) UserManagementService {
	return &userManagementService{
		storage:   storage,
		providers: make(map[Provider]UserManagementProvider),
	}
}

// CreateUser creates a new user
func (s *userManagementService) CreateUser(user *User) error {
	// Protect the entire check-then-store operation with a mutex to prevent race conditions
	// This ensures that duplicate checking and user creation are atomic
	log.Printf("[DEBUG SERVICE] CreateUser: Attempting to acquire service lock for username=%s\n", user.Username)
	s.mu.Lock()
	log.Printf("[DEBUG SERVICE] CreateUser: Service lock acquired for username=%s\n", user.Username)
	defer func() {
		log.Printf("[DEBUG SERVICE] CreateUser: Releasing service lock for username=%s\n", user.Username)
		s.mu.Unlock()
	}()

	// Validate required fields
	if user.Username == "" {
		return fmt.Errorf("username is required")
	}
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}

	// Check for duplicate username - get ALL users to avoid issues with substring filtering
	existingUsers, err := s.storage.ListUsers(nil)
	if err != nil {
		return fmt.Errorf("failed to check for duplicate username: %w", err)
	}
	fmt.Printf("[DEBUG] CreateUser: Checking %d existing users for duplicates of username=%s\n", len(existingUsers), user.Username)
	for _, existing := range existingUsers {
		if strings.EqualFold(existing.Username, user.Username) {
			fmt.Printf("[DEBUG] CreateUser: Found duplicate username=%s, returning error\n", user.Username)
			return ErrDuplicateUsername
		}
		if strings.EqualFold(existing.Email, user.Email) {
			return ErrDuplicateEmail
		}
	}
	fmt.Printf("[DEBUG] CreateUser: No duplicates found, proceeding to store username=%s\n", user.Username)

	// Generate ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Set defaults
	if user.Provider == "" {
		user.Provider = ProviderLocal
	}
	if len(user.Roles) == 0 {
		user.Roles = []UserRole{UserRoleUser}
	}
	user.Enabled = true
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Store user
	return s.storage.StoreUser(user)
}

// GetUser retrieves a user by ID
func (s *userManagementService) GetUser(id string) (*User, error) {
	return s.storage.RetrieveUser(id)
}

// GetUserByUsername retrieves a user by username
func (s *userManagementService) GetUserByUsername(username string) (*User, error) {
	users, err := s.storage.ListUsers(&UserFilter{Username: username})
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if strings.EqualFold(user.Username, username) {
			return user, nil
		}
	}

	return nil, ErrUserNotFound
}

// GetUserByEmail retrieves a user by email
func (s *userManagementService) GetUserByEmail(email string) (*User, error) {
	users, err := s.storage.ListUsers(&UserFilter{Email: email})
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if strings.EqualFold(user.Email, email) {
			return user, nil
		}
	}

	return nil, ErrUserNotFound
}

// UpdateUser updates an existing user
func (s *userManagementService) UpdateUser(user *User) error {
	// Verify user exists
	_, err := s.storage.RetrieveUser(user.ID)
	if err != nil {
		return err
	}

	user.UpdatedAt = time.Now()
	return s.storage.UpdateUser(user)
}

// DeleteUser deletes a user
func (s *userManagementService) DeleteUser(id string) error {
	return s.storage.DeleteUser(id)
}

// ListUsers lists users with filtering and pagination
func (s *userManagementService) ListUsers(filter *UserFilter, pagination *PaginationOptions) (*PaginatedUsers, error) {
	users, err := s.storage.ListUsers(filter)
	if err != nil {
		return nil, err
	}

	// Apply pagination
	total := len(users)
	if pagination != nil && pagination.PageSize > 0 {
		// Convert page and pageSize to offset and limit
		page := pagination.Page
		if page < 1 {
			page = 1
		}
		start := (page - 1) * pagination.PageSize
		end := start + pagination.PageSize

		if start > total {
			start = total
		}
		if end > total {
			end = total
		}

		users = users[start:end]
	}

	return &PaginatedUsers{
		Users: users,
		Total: total,
	}, nil
}

// CreateGroup creates a new group
func (s *userManagementService) CreateGroup(group *Group) error {
	// Protect the entire check-then-store operation with a mutex to prevent race conditions
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate required fields
	if group.Name == "" {
		return fmt.Errorf("group name is required")
	}

	// Check for duplicate name
	existingGroups, err := s.storage.ListGroups(&GroupFilter{Name: group.Name})
	if err != nil {
		return fmt.Errorf("failed to check for duplicate group name: %w", err)
	}
	for _, existing := range existingGroups {
		if strings.EqualFold(existing.Name, group.Name) {
			return ErrDuplicateGroup
		}
	}

	// Generate ID if not provided
	if group.ID == "" {
		group.ID = uuid.New().String()
	}

	// Store group
	return s.storage.StoreGroup(group)
}

// GetGroup retrieves a group by ID
func (s *userManagementService) GetGroup(id string) (*Group, error) {
	return s.storage.RetrieveGroup(id)
}

// GetGroupByName retrieves a group by name
func (s *userManagementService) GetGroupByName(name string) (*Group, error) {
	groups, err := s.storage.ListGroups(&GroupFilter{Name: name})
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		if strings.EqualFold(group.Name, name) {
			return group, nil
		}
	}

	return nil, ErrGroupNotFound
}

// UpdateGroup updates an existing group
func (s *userManagementService) UpdateGroup(group *Group) error {
	// Verify group exists
	_, err := s.storage.RetrieveGroup(group.ID)
	if err != nil {
		return err
	}

	return s.storage.UpdateGroup(group)
}

// DeleteGroup deletes a group
func (s *userManagementService) DeleteGroup(id string) error {
	return s.storage.DeleteGroup(id)
}

// ListGroups lists groups with filtering and pagination
func (s *userManagementService) ListGroups(filter *GroupFilter, pagination *PaginationOptions) (*PaginatedGroups, error) {
	groups, err := s.storage.ListGroups(filter)
	if err != nil {
		return nil, err
	}

	// Apply pagination
	total := len(groups)
	if pagination != nil && pagination.PageSize > 0 {
		// Convert page and pageSize to offset and limit
		page := pagination.Page
		if page < 1 {
			page = 1
		}
		start := (page - 1) * pagination.PageSize
		end := start + pagination.PageSize

		if start > total {
			start = total
		}
		if end > total {
			end = total
		}

		groups = groups[start:end]
	}

	return &PaginatedGroups{
		Groups: groups,
		Total:  total,
	}, nil
}

// GetGroups returns all groups
func (s *userManagementService) GetGroups() ([]*Group, error) {
	return s.storage.ListGroups(nil)
}

// AddUserToGroup adds a user to a group
func (s *userManagementService) AddUserToGroup(userID, groupID string) error {
	// Verify user and group exist
	_, err := s.storage.RetrieveUser(userID)
	if err != nil {
		return err
	}

	_, err = s.storage.RetrieveGroup(groupID)
	if err != nil {
		return err
	}

	return s.storage.StoreUserGroupMembership(userID, groupID)
}

// RemoveUserFromGroup removes a user from a group
func (s *userManagementService) RemoveUserFromGroup(userID, groupID string) error {
	return s.storage.RemoveUserGroupMembership(userID, groupID)
}

// GetUserGroups returns all groups a user belongs to
func (s *userManagementService) GetUserGroups(userID string) ([]*Group, error) {
	groupIDs, err := s.storage.GetUserGroups(userID)
	if err != nil {
		return nil, err
	}

	var groups []*Group
	for _, groupID := range groupIDs {
		group, err := s.storage.RetrieveGroup(groupID)
		if err != nil {
			continue // Skip groups that no longer exist
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// GetGroupUsers returns all users in a group
func (s *userManagementService) GetGroupUsers(groupID string) ([]*User, error) {
	userIDs, err := s.storage.GetGroupUsers(groupID)
	if err != nil {
		return nil, err
	}

	var users []*User
	for _, userID := range userIDs {
		user, err := s.storage.RetrieveUser(userID)
		if err != nil {
			continue // Skip users that no longer exist
		}
		users = append(users, user)
	}

	return users, nil
}

// SyncUsers synchronizes users (placeholder for now)
func (s *userManagementService) SyncUsers(options *SyncOptions) (*SyncResult, error) {
	return &SyncResult{}, nil
}

// ProvisionUser provisions a user from a provider (placeholder for now)
func (s *userManagementService) ProvisionUser(providerUser interface{}, options *UserProvisionOptions) (*User, error) {
	return nil, fmt.Errorf("user provisioning not implemented")
}

// RegisterProvider registers a user management provider
func (s *userManagementService) RegisterProvider(provider UserManagementProvider) error {
	providerType := provider.GetProviderType()
	s.providers[providerType] = provider
	return nil
}

// UnregisterProvider unregisters a user management provider
func (s *userManagementService) UnregisterProvider(providerType Provider) error {
	delete(s.providers, providerType)
	return nil
}

// Authenticate authenticates a user (basic implementation)
func (s *userManagementService) Authenticate(username, password string) (*AuthenticationResult, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return &AuthenticationResult{
			Success:      false,
			ErrorMessage: "invalid username or password",
		}, nil
	}

	if !user.Enabled {
		return &AuthenticationResult{
			Success:      false,
			ErrorMessage: "user account is disabled",
		}, nil
	}

	// Note: This is a basic implementation
	// In production, you would check password hash here
	// For now, we just return a simple token
	token := uuid.New().String()

	now := time.Now()
	user.LastLogin = &now
	_ = s.UpdateUser(user)

	expiresAt := time.Now().Add(24 * time.Hour)
	return &AuthenticationResult{
		Success:   true,
		User:      user,
		Token:     token,
		ExpiresAt: &expiresAt,
	}, nil
}

// EnableUser enables a user account
func (s *userManagementService) EnableUser(id string) error {
	user, err := s.storage.RetrieveUser(id)
	if err != nil {
		return err
	}

	user.Enabled = true
	return s.storage.UpdateUser(user)
}

// DisableUser disables a user account
func (s *userManagementService) DisableUser(id string) error {
	user, err := s.storage.RetrieveUser(id)
	if err != nil {
		return err
	}

	user.Enabled = false
	return s.storage.UpdateUser(user)
}

// SynchronizeUsers synchronizes users (placeholder for now)
func (s *userManagementService) SynchronizeUsers(options *SyncOptions) (*SyncResult, error) {
	return &SyncResult{}, nil
}

// SetDefaultProvisionOptions sets default provisioning options (placeholder)
func (s *userManagementService) SetDefaultProvisionOptions(options *UserProvisionOptions) error {
	return nil
}

// GetDefaultProvisionOptions gets default provisioning options (placeholder)
func (s *userManagementService) GetDefaultProvisionOptions() (*UserProvisionOptions, error) {
	return &UserProvisionOptions{}, nil
}

// Close closes the service
func (s *userManagementService) Close() error {
	return nil
}
