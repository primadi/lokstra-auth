package rbac

import (
	rbacdomain "github.com/primadi/lokstra-auth/pkg/domain/rbac"
)

// CreateRoleRequest for creating a new role
type CreateRoleRequest struct {
	TenantID    string          `json:"tenant_id" validate:"required"`
	AppID       string          `json:"app_id" validate:"required"`
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Metadata    *map[string]any `json:"metadata,omitempty"`
}

// GetRoleRequest for retrieving a role
type GetRoleRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	RoleID   string `json:"role_id" validate:"required"`
}

// UpdateRoleRequest for updating a role
type UpdateRoleRequest struct {
	TenantID    string                 `json:"tenant_id" validate:"required"`
	AppID       string                 `json:"app_id" validate:"required"`
	RoleID      string                 `json:"role_id" validate:"required"`
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Status      *rbacdomain.RoleStatus `json:"status,omitempty"`
	Metadata    *map[string]any        `json:"metadata,omitempty"`
}

// DeleteRoleRequest for deleting a role
type DeleteRoleRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	RoleID   string `json:"role_id" validate:"required"`
}

// ListRolesRequest for listing roles
type ListRolesRequest struct {
	TenantID string                 `json:"tenant_id" validate:"required"`
	AppID    string                 `json:"app_id" validate:"required"`
	Status   *rbacdomain.RoleStatus `json:"status,omitempty"`
	Limit    int                    `json:"limit"`
	Offset   int                    `json:"offset"`
}

// AssignRoleRequest for assigning role to user
type AssignRoleRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	UserID   string `json:"user_id" validate:"required"`
	RoleID   string `json:"role_id" validate:"required"`
}

// RevokeRoleRequest for revoking role from user
type RevokeRoleRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	UserID   string `json:"user_id" validate:"required"`
	RoleID   string `json:"role_id" validate:"required"`
}

// ListUserRolesRequest for listing user's roles
type ListUserRolesRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	UserID   string `json:"user_id" validate:"required"`
}
