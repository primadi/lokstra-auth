package rbac

import (
	rbacdomain "github.com/primadi/lokstra-auth/pkg/domain/rbac"
)

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
	TenantID     string                       `json:"tenant_id" validate:"required"`
	AppID        string                       `json:"app_id" validate:"required"`
	PermissionID string                       `json:"permission_id" validate:"required"`
	Name         *string                      `json:"name,omitempty"`
	Description  *string                      `json:"description,omitempty"`
	Resource     *string                      `json:"resource,omitempty"`
	Action       *string                      `json:"action,omitempty"`
	Status       *rbacdomain.PermissionStatus `json:"status,omitempty"`
	Metadata     *map[string]any              `json:"metadata,omitempty"`
}

// DeletePermissionRequest for deleting a permission
type DeletePermissionRequest struct {
	TenantID     string `json:"tenant_id" validate:"required"`
	AppID        string `json:"app_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
}

// ListPermissionsRequest for listing permissions
type ListPermissionsRequest struct {
	TenantID string                       `json:"tenant_id" validate:"required"`
	AppID    string                       `json:"app_id" validate:"required"`
	Resource *string                      `json:"resource,omitempty"`
	Action   *string                      `json:"action,omitempty"`
	Status   *rbacdomain.PermissionStatus `json:"status,omitempty"`
	Limit    int                          `json:"limit"`
	Offset   int                          `json:"offset"`
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
