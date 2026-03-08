// Package rbac provides role-based access control for Prism.
// Roles map user identities (profile names) to named permission sets.
package rbac

import "time"

// Role represents a named set of permissions.
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"` // e.g., "instances:launch", "storage:delete", "*"
	CreatedAt   time.Time `json:"created_at"`
}

// UserRole records that a user has been assigned a role.
type UserRole struct {
	UserID     string    `json:"user_id"`
	RoleID     string    `json:"role_id"`
	AssignedBy string    `json:"assigned_by,omitempty"`
	AssignedAt time.Time `json:"assigned_at"`
}

// AssignRoleRequest is the request body for POST /api/v1/rbac/assign.
type AssignRoleRequest struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

// Built-in role IDs.
const (
	RoleAdmin      = "admin"
	RoleResearcher = "researcher"
	RoleStudent    = "student"
	RoleViewer     = "viewer"
)

// Well-known action strings used across the daemon for permission checks.
const (
	ActionInstancesLaunch    = "instances:launch"
	ActionInstancesStop      = "instances:stop"
	ActionInstancesStart     = "instances:start"
	ActionInstancesTerminate = "instances:terminate"
	ActionInstancesView      = "instances:view"
	ActionInstancesConnect   = "instances:connect"
	ActionStorageCreate      = "storage:create"
	ActionStorageDelete      = "storage:delete"
	ActionStorageAttach      = "storage:attach"
	ActionStorageDetach      = "storage:detach"
	ActionStorageView        = "storage:view"
	ActionTemplatesView      = "templates:view"
	ActionProfilesView       = "profiles:view"
	ActionProfilesManage     = "profiles:manage"
	ActionPoliciesManage     = "policies:manage"
	ActionAuditView          = "audit:view"
	ActionRBACManage         = "rbac:manage"
)
