package rbac

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Manager handles role-based access control for Prism.
// User identities are profile names (as returned by profile.GetCurrentProfile().Name).
// Assignments are persisted to ~/.prism/rbac.json.
type Manager struct {
	roles       map[string]*Role
	userRoles   map[string]string // userID -> roleID
	mutex       sync.RWMutex
	storagePath string
}

// NewManager creates a Manager with built-in default roles and loads any persisted assignments.
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	m := &Manager{
		roles:       make(map[string]*Role),
		userRoles:   make(map[string]string),
		storagePath: filepath.Join(homeDir, ".prism", "rbac.json"),
	}

	m.registerDefaultRoles()
	_ = m.load() // best-effort: no file on first run

	return m, nil
}

func (m *Manager) registerDefaultRoles() {
	now := time.Now()

	m.roles[RoleAdmin] = &Role{
		ID:          RoleAdmin,
		Name:        "Administrator",
		Description: "Full access to all resources and operations",
		Permissions: []string{"*"},
		CreatedAt:   now,
	}

	m.roles[RoleResearcher] = &Role{
		ID:          RoleResearcher,
		Name:        "Researcher",
		Description: "Launch and manage instances and storage",
		Permissions: []string{
			ActionInstancesLaunch, ActionInstancesStop, ActionInstancesStart,
			ActionInstancesTerminate, ActionInstancesView, ActionInstancesConnect,
			ActionStorageCreate, ActionStorageAttach, ActionStorageDetach, ActionStorageView,
			ActionTemplatesView, ActionProfilesView, ActionProfilesManage,
		},
		CreatedAt: now,
	}

	m.roles[RoleStudent] = &Role{
		ID:          RoleStudent,
		Name:        "Student",
		Description: "Launch and stop instances using approved templates; no termination or storage deletion",
		Permissions: []string{
			ActionInstancesLaunch, ActionInstancesStop, ActionInstancesStart,
			ActionInstancesView, ActionInstancesConnect,
			ActionStorageAttach, ActionStorageDetach, ActionStorageView,
			ActionTemplatesView,
		},
		CreatedAt: now,
	}

	m.roles[RoleViewer] = &Role{
		ID:          RoleViewer,
		Name:        "Viewer",
		Description: "Read-only access to all resources",
		Permissions: []string{
			ActionInstancesView, ActionStorageView, ActionTemplatesView, ActionProfilesView,
		},
		CreatedAt: now,
	}
}

// AssignRole assigns a named role to a user ID, persisting the assignment.
func (m *Manager) AssignRole(userID, roleID, assignedBy string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.roles[roleID]; !exists {
		return fmt.Errorf("role %q not found", roleID)
	}
	m.userRoles[userID] = roleID
	return m.save()
}

// RemoveRole removes any role assignment for the user, reverting to the default researcher role.
func (m *Manager) RemoveRole(userID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.userRoles, userID)
	return m.save()
}

// GetUserRole returns the role for a user. Defaults to the researcher role if no assignment exists.
func (m *Manager) GetUserRole(userID string) *Role {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	roleID, exists := m.userRoles[userID]
	if !exists {
		return m.roles[RoleResearcher]
	}
	if role, ok := m.roles[roleID]; ok {
		return role
	}
	return m.roles[RoleResearcher]
}

// CanPerformAction returns (allowed, reason) for the given user and action string.
// action format: "instances:launch", "storage:delete", etc.
func (m *Manager) CanPerformAction(userID, action string) (bool, string) {
	role := m.GetUserRole(userID)
	if role == nil {
		return false, "no role assigned"
	}

	for _, perm := range role.Permissions {
		if perm == "*" {
			return true, fmt.Sprintf("allowed by role %q (full access)", role.Name)
		}
		if perm == action {
			return true, fmt.Sprintf("allowed by role %q", role.Name)
		}
		// Wildcard segment: "instances:*" matches "instances:launch"
		if strings.HasSuffix(perm, ":*") {
			prefix := strings.TrimSuffix(perm, ":*")
			if strings.HasPrefix(action, prefix+":") {
				return true, fmt.Sprintf("allowed by role %q", role.Name)
			}
		}
	}

	return false, fmt.Sprintf("action %q not permitted for role %q — contact your administrator", action, role.Name)
}

// ListRoles returns all available roles (built-in and any custom ones).
func (m *Manager) ListRoles() []*Role {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	roles := make([]*Role, 0, len(m.roles))
	for _, r := range m.roles {
		copy := *r
		roles = append(roles, &copy)
	}
	return roles
}

// ListUserAssignments returns all explicit user→role mappings.
func (m *Manager) ListUserAssignments() map[string]string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]string, len(m.userRoles))
	for k, v := range m.userRoles {
		result[k] = v
	}
	return result
}

// persistence

type rbacState struct {
	UserRoles map[string]string `json:"user_roles"`
}

func (m *Manager) save() error {
	state := rbacState{UserRoles: m.userRoles}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(m.storagePath), 0700); err != nil {
		return err
	}
	return os.WriteFile(m.storagePath, data, 0600)
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.storagePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var state rbacState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	if state.UserRoles != nil {
		m.userRoles = state.UserRoles
	}
	return nil
}
