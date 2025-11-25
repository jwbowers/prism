package usermgmt

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// memoryUserStorage provides a real in-memory implementation of UserStorage
// This replaces the stub implementation in types.go
type memoryUserStorage struct {
	mu     sync.RWMutex
	users  map[string]*User
	groups map[string]*Group

	// Track user-group memberships
	userGroups map[string][]string // userID -> []groupID
	groupUsers map[string][]string // groupID -> []userID
}

// NewMemoryUserStorage creates a new in-memory user storage with real implementations
func NewMemoryUserStorage() UserStorage {
	return &memoryUserStorage{
		users:      make(map[string]*User),
		groups:     make(map[string]*Group),
		userGroups: make(map[string][]string),
		groupUsers: make(map[string][]string),
	}
}

// StoreUser stores a user in memory
func (m *memoryUserStorage) StoreUser(user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user.ID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Create a copy to avoid external modifications
	userCopy := *user
	userCopy.CreatedAt = time.Now()
	userCopy.UpdatedAt = time.Now()

	m.users[user.ID] = &userCopy
	return nil
}

// RetrieveUser retrieves a user by ID
func (m *memoryUserStorage) RetrieveUser(id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Return a copy to prevent external modifications
	userCopy := *user
	return &userCopy, nil
}

// UpdateUser updates an existing user
func (m *memoryUserStorage) UpdateUser(user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.ID]; !exists {
		return ErrUserNotFound
	}

	// Create a copy and update timestamp
	userCopy := *user
	userCopy.UpdatedAt = time.Now()

	m.users[user.ID] = &userCopy
	return nil
}

// DeleteUser deletes a user
func (m *memoryUserStorage) DeleteUser(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[id]; !exists {
		return ErrUserNotFound
	}

	delete(m.users, id)

	// Clean up group memberships
	delete(m.userGroups, id)
	for groupID := range m.groupUsers {
		m.removeUserFromGroupList(id, groupID)
	}

	return nil
}

// ListUsers lists users matching the filter
func (m *memoryUserStorage) ListUsers(filter *UserFilter) ([]*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*User

	for _, user := range m.users {
		// Apply filters
		if filter != nil {
			// Filter by username
			if filter.Username != "" && !strings.Contains(strings.ToLower(user.Username), strings.ToLower(filter.Username)) {
				continue
			}

			// Filter by email
			if filter.Email != "" && !strings.Contains(strings.ToLower(user.Email), strings.ToLower(filter.Email)) {
				continue
			}

			// Filter by provider
			if filter.Provider != "" && user.Provider != filter.Provider {
				continue
			}

			// Filter by enabled status
			if filter.EnabledOnly && !user.Enabled {
				continue
			}

			if filter.DisabledOnly && user.Enabled {
				continue
			}
		}

		// Create a copy
		userCopy := *user
		result = append(result, &userCopy)
	}

	return result, nil
}

// StoreGroup stores a group
func (m *memoryUserStorage) StoreGroup(group *Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if group.ID == "" {
		return fmt.Errorf("group ID cannot be empty")
	}

	groupCopy := *group
	m.groups[group.ID] = &groupCopy
	return nil
}

// RetrieveGroup retrieves a group by ID
func (m *memoryUserStorage) RetrieveGroup(id string) (*Group, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group, exists := m.groups[id]
	if !exists {
		return nil, ErrGroupNotFound
	}

	groupCopy := *group
	return &groupCopy, nil
}

// UpdateGroup updates an existing group
func (m *memoryUserStorage) UpdateGroup(group *Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.groups[group.ID]; !exists {
		return ErrGroupNotFound
	}

	groupCopy := *group
	m.groups[group.ID] = &groupCopy
	return nil
}

// DeleteGroup deletes a group
func (m *memoryUserStorage) DeleteGroup(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.groups[id]; !exists {
		return ErrGroupNotFound
	}

	delete(m.groups, id)

	// Clean up memberships
	delete(m.groupUsers, id)
	for userID := range m.userGroups {
		m.removeGroupFromUserList(userID, id)
	}

	return nil
}

// ListGroups lists groups matching the filter
func (m *memoryUserStorage) ListGroups(filter *GroupFilter) ([]*Group, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Group

	for _, group := range m.groups {
		// Apply filters
		if filter != nil {
			// Filter by name
			if filter.Name != "" && !strings.Contains(strings.ToLower(group.Name), strings.ToLower(filter.Name)) {
				continue
			}

			// Filter by provider
			if filter.Provider != "" && group.Provider != filter.Provider {
				continue
			}
		}

		groupCopy := *group
		result = append(result, &groupCopy)
	}

	return result, nil
}

// StoreUserGroupMembership adds a user to a group
func (m *memoryUserStorage) StoreUserGroupMembership(userID, groupID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify user and group exist
	if _, exists := m.users[userID]; !exists {
		return ErrUserNotFound
	}
	if _, exists := m.groups[groupID]; !exists {
		return ErrGroupNotFound
	}

	// Add to userGroups map
	if m.userGroups[userID] == nil {
		m.userGroups[userID] = []string{}
	}
	// Check if already added
	for _, gid := range m.userGroups[userID] {
		if gid == groupID {
			return nil // Already member
		}
	}
	m.userGroups[userID] = append(m.userGroups[userID], groupID)

	// Add to groupUsers map
	if m.groupUsers[groupID] == nil {
		m.groupUsers[groupID] = []string{}
	}
	// Check if already added
	for _, uid := range m.groupUsers[groupID] {
		if uid == userID {
			return nil // Already member
		}
	}
	m.groupUsers[groupID] = append(m.groupUsers[groupID], userID)

	return nil
}

// RemoveUserGroupMembership removes a user from a group
func (m *memoryUserStorage) RemoveUserGroupMembership(userID, groupID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.removeUserFromGroupList(userID, groupID)
	m.removeGroupFromUserList(userID, groupID)

	return nil
}

// GetUserGroups returns the group IDs for a user
func (m *memoryUserStorage) GetUserGroups(userID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groups, exists := m.userGroups[userID]
	if !exists {
		return []string{}, nil
	}

	// Return a copy
	result := make([]string, len(groups))
	copy(result, groups)
	return result, nil
}

// GetGroupUsers returns the user IDs in a group
func (m *memoryUserStorage) GetGroupUsers(groupID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users, exists := m.groupUsers[groupID]
	if !exists {
		return []string{}, nil
	}

	// Return a copy
	result := make([]string, len(users))
	copy(result, users)
	return result, nil
}

// Helper methods (must be called with lock held)

func (m *memoryUserStorage) removeUserFromGroupList(userID, groupID string) {
	users := m.groupUsers[groupID]
	for i, uid := range users {
		if uid == userID {
			m.groupUsers[groupID] = append(users[:i], users[i+1:]...)
			break
		}
	}
}

func (m *memoryUserStorage) removeGroupFromUserList(userID, groupID string) {
	groups := m.userGroups[userID]
	for i, gid := range groups {
		if gid == groupID {
			m.userGroups[userID] = append(groups[:i], groups[i+1:]...)
			break
		}
	}
}
