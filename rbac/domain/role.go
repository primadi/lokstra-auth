package domain

import (
	"errors"
	"time"
)

var (
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleAlreadyExists = errors.New("role already exists")
	ErrInvalidRoleID     = errors.New("invalid role ID")
	ErrDuplicateRoleName = errors.New("role name already exists in this tenant+app")
)

// RoleStatus represents the status of a role
type RoleStatus string

const (
	RoleStatusActive   RoleStatus = "active"
	RoleStatusInactive RoleStatus = "inactive"
)

// Role represents a role in the system (tenant+app scoped)
type Role struct {
	ID          string          `json:"id"`          // Unique identifier (UUID)
	TenantID    string          `json:"tenant_id"`   // Belongs to tenant (REQUIRED)
	AppID       string          `json:"app_id"`      // Belongs to app (REQUIRED)
	Name        string          `json:"name"`        // Role name (unique within tenant+app)
	Description string          `json:"description"` // Role description
	Status      RoleStatus      `json:"status"`      // active, inactive
	Metadata    *map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// UserRole represents user-to-role assignment (tenant+app scoped)
type UserRole struct {
	UserID    string     `json:"user_id"`
	TenantID  string     `json:"tenant_id"`
	AppID     string     `json:"app_id"`
	RoleID    string     `json:"role_id"`
	GrantedAt time.Time  `json:"granted_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// Validate validates role data
func (r *Role) Validate() error {
	if r.ID == "" {
		return ErrInvalidRoleID
	}
	if r.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if r.AppID == "" {
		return errors.New("app_id is required")
	}
	if r.Name == "" {
		return errors.New("role name is required")
	}
	return nil
}

// IsActive checks if role is active
func (r *Role) IsActive() bool {
	return r.Status == RoleStatusActive
}

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
	TenantID    string          `json:"tenant_id" validate:"required"`
	AppID       string          `json:"app_id" validate:"required"`
	RoleID      string          `json:"role_id" validate:"required"`
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Status      *RoleStatus     `json:"status,omitempty"`
	Metadata    *map[string]any `json:"metadata,omitempty"`
}

// DeleteRoleRequest for deleting a role
type DeleteRoleRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	RoleID   string `json:"role_id" validate:"required"`
}

// ListRolesRequest for listing roles
type ListRolesRequest struct {
	TenantID string      `json:"tenant_id" validate:"required"`
	AppID    string      `json:"app_id" validate:"required"`
	Status   *RoleStatus `json:"status,omitempty"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
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
