package domain

import (
	"errors"
	"time"
)

var (
	ErrPermissionNotFound      = errors.New("permission not found")
	ErrPermissionAlreadyExists = errors.New("permission already exists")
	ErrInvalidPermissionID     = errors.New("invalid permission ID")
	ErrDuplicatePermissionName = errors.New("permission name already exists in this tenant+app")
)

// PermissionStatus represents the status of a permission
type PermissionStatus string

const (
	PermissionStatusActive   PermissionStatus = "active"
	PermissionStatusInactive PermissionStatus = "inactive"
)

// Permission represents a permission in the system (tenant+app scoped)
type Permission struct {
	ID          string           `json:"id"`          // Unique identifier (UUID)
	TenantID    string           `json:"tenant_id"`   // Belongs to tenant (REQUIRED)
	AppID       string           `json:"app_id"`      // Belongs to app (REQUIRED)
	Name        string           `json:"name"`        // Permission name (e.g., "users:read", "documents:write")
	Description string           `json:"description"` // Permission description
	Resource    string           `json:"resource"`    // Resource type (e.g., "users", "documents")
	Action      string           `json:"action"`      // Action (e.g., "read", "write", "delete")
	Status      PermissionStatus `json:"status"`      // active, inactive
	Metadata    *map[string]any  `json:"metadata,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// RolePermission represents role-to-permission assignment (tenant+app scoped)
type RolePermission struct {
	RoleID       string     `json:"role_id"`
	TenantID     string     `json:"tenant_id"`
	AppID        string     `json:"app_id"`
	PermissionID string     `json:"permission_id"`
	GrantedAt    time.Time  `json:"granted_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

// UserPermission represents direct user-to-permission assignment (tenant+app scoped)
type UserPermission struct {
	UserID       string     `json:"user_id"`
	TenantID     string     `json:"tenant_id"`
	AppID        string     `json:"app_id"`
	PermissionID string     `json:"permission_id"`
	GrantedAt    time.Time  `json:"granted_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

// Validate validates permission data
func (p *Permission) Validate() error {
	if p.ID == "" {
		return ErrInvalidPermissionID
	}
	if p.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if p.AppID == "" {
		return errors.New("app_id is required")
	}
	if p.Name == "" {
		return errors.New("permission name is required")
	}
	return nil
}

// IsActive checks if permission is active
func (p *Permission) IsActive() bool {
	return p.Status == PermissionStatusActive
}

// CreatePermissionRequest for creating a new permission
type CreatePermissionRequest struct {
	TenantID    string          `json:"tenant_id" validate:"required"`
	AppID       string          `json:"app_id" validate:"required"`
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Resource    string          `json:"resource"`
	Action      string          `json:"action"`
	Metadata    *map[string]any `json:"metadata,omitempty"`
}

// GetPermissionRequest for retrieving a permission
type GetPermissionRequest struct {
	TenantID     string `json:"tenant_id" validate:"required"`
	AppID        string `json:"app_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
}

// UpdatePermissionRequest for updating a permission
type UpdatePermissionRequest struct {
	TenantID     string            `json:"tenant_id" validate:"required"`
	AppID        string            `json:"app_id" validate:"required"`
	PermissionID string            `json:"permission_id" validate:"required"`
	Name         *string           `json:"name,omitempty"`
	Description  *string           `json:"description,omitempty"`
	Resource     *string           `json:"resource,omitempty"`
	Action       *string           `json:"action,omitempty"`
	Status       *PermissionStatus `json:"status,omitempty"`
	Metadata     *map[string]any   `json:"metadata,omitempty"`
}

// DeletePermissionRequest for deleting a permission
type DeletePermissionRequest struct {
	TenantID     string `json:"tenant_id" validate:"required"`
	AppID        string `json:"app_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
}

// ListPermissionsRequest for listing permissions
type ListPermissionsRequest struct {
	TenantID string            `json:"tenant_id" validate:"required"`
	AppID    string            `json:"app_id" validate:"required"`
	Resource *string           `json:"resource,omitempty"`
	Action   *string           `json:"action,omitempty"`
	Status   *PermissionStatus `json:"status,omitempty"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// AssignPermissionToRoleRequest for assigning permission to role
type AssignPermissionToRoleRequest struct {
	TenantID     string `json:"tenant_id" validate:"required"`
	AppID        string `json:"app_id" validate:"required"`
	RoleID       string `json:"role_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
}

// RevokePermissionFromRoleRequest for revoking permission from role
type RevokePermissionFromRoleRequest struct {
	TenantID     string `json:"tenant_id" validate:"required"`
	AppID        string `json:"app_id" validate:"required"`
	RoleID       string `json:"role_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
}

// AssignPermissionToUserRequest for assigning permission directly to user
type AssignPermissionToUserRequest struct {
	TenantID     string `json:"tenant_id" validate:"required"`
	AppID        string `json:"app_id" validate:"required"`
	UserID       string `json:"user_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
}

// RevokePermissionFromUserRequest for revoking permission from user
type RevokePermissionFromUserRequest struct {
	TenantID     string `json:"tenant_id" validate:"required"`
	AppID        string `json:"app_id" validate:"required"`
	UserID       string `json:"user_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
}

// ListRolePermissionsRequest for listing role's permissions
type ListRolePermissionsRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	RoleID   string `json:"role_id" validate:"required"`
}

// ListUserPermissionsRequest for listing user's permissions (including from roles)
type ListUserPermissionsRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	UserID   string `json:"user_id" validate:"required"`
}
