package core

import (
	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
)

// CreateUserRequest request to create a user
type CreateUserRequest struct {
	TenantID string          `header:"X-Tenant-ID" validate:"required"`
	ID       string          `path:"id" validate:"required"`
	Username string          `json:"username" validate:"required"`
	Email    string          `json:"email" validate:"required,email"`
	Metadata *map[string]any `json:"metadata,omitempty"`
}

// GetUserRequest request to get a user
type GetUserRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// GetUserByUsernameRequest request to get a user by username
type GetUserByUsernameRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	Username string `path:"username" validate:"required"`
}

// GetUserByEmailRequest request to get a user by email
type GetUserByEmailRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	Email    string `path:"email" validate:"required"`
}

// UpdateUserRequest request to update a user
type UpdateUserRequest struct {
	TenantID string                `header:"X-Tenant-ID" validate:"required"`
	ID       string                `path:"id" validate:"required"`
	Username string                `json:"username"`
	Email    string                `json:"email" validate:"email"`
	Status   domaincore.UserStatus `json:"status"`
	Metadata *map[string]any       `json:"metadata"`
}

// DeleteUserRequest request to delete a user
type DeleteUserRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// ListUsersRequest request to list users
type ListUsersRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
}

// ListUsersByAppRequest request to list users by app
type ListUsersByAppRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
}

// AssignUserToAppRequest request to assign user to app
type AssignUserToAppRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	UserID   string `path:"user_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
}

// RemoveUserFromAppRequest request to remove user from app
type RemoveUserFromAppRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	UserID   string `path:"user_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
}

// ActivateUserRequest request to activate a user
type ActivateUserRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// SuspendUserRequest request to suspend a user
type SuspendUserRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// SetPasswordRequest request to set or update user password (basic auth)
type SetPasswordRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	ID       string `path:"id" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	// OldPassword string `json:"old_password,omitempty"` // Required when updating existing password
}

// RemovePasswordRequest request to remove user password (disable basic auth)
type RemovePasswordRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	ID       string `path:"id" validate:"required"`
}
