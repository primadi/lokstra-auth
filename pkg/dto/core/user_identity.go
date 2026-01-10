package core

import (
	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
)

// LinkIdentityRequest request to link an identity to a user
type LinkIdentityRequest struct {
	TenantID   string                      `path:"tenant_id" validate:"required"`
	UserID     string                      `path:"user_id" validate:"required"`
	Provider   domaincore.IdentityProvider `json:"provider" validate:"required"`
	ProviderID string                      `json:"provider_id" validate:"required"`
	Email      string                      `json:"email" validate:"email"`
	Username   string                      `json:"username"`
	Verified   bool                        `json:"verified"`
	Metadata   *map[string]any             `json:"metadata,omitempty"`
}

// UnlinkIdentityRequest request to unlink an identity from a user
type UnlinkIdentityRequest struct {
	TenantID   string `path:"tenant_id" validate:"required"`
	UserID     string `path:"user_id" validate:"required"`
	IdentityID string `path:"identity_id" validate:"required"`
}

// GetUserIdentityRequest request to get a specific identity
type GetUserIdentityRequest struct {
	TenantID   string `path:"tenant_id" validate:"required"`
	UserID     string `path:"user_id" validate:"required"`
	IdentityID string `path:"identity_id" validate:"required"`
}

// ListUserIdentitiesRequest request to list all identities for a user
type ListUserIdentitiesRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	UserID   string `path:"user_id" validate:"required"`
}

// FindUserByProviderRequest request to find user by provider identity
type FindUserByProviderRequest struct {
	TenantID   string                      `path:"tenant_id" validate:"required"`
	Provider   domaincore.IdentityProvider `path:"provider" validate:"required"`
	ProviderID string                      `path:"provider_id" validate:"required"`
}

// UpdateUserIdentityRequest request to update an identity
type UpdateUserIdentityRequest struct {
	TenantID   string          `path:"tenant_id" validate:"required"`
	UserID     string          `path:"user_id" validate:"required"`
	IdentityID string          `path:"identity_id" validate:"required"`
	Email      string          `json:"email" validate:"email"`
	Username   string          `json:"username"`
	Verified   bool            `json:"verified"`
	Metadata   *map[string]any `json:"metadata"`
}

// GetIdentityRequest request to get a specific identity
type GetIdentityRequest struct {
	TenantID   string `path:"tenant_id" validate:"required"`
	UserID     string `path:"user_id" validate:"required"`
	IdentityID string `path:"identity_id" validate:"required"`
}

// ListIdentitiesRequest request to list all identities for a user
type ListIdentitiesRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	UserID   string `path:"user_id" validate:"required"`
}

// GetOrCreateUserRequest request to get or create user by provider identity
type GetOrCreateUserRequest struct {
	TenantID   string                      `path:"tenant_id" validate:"required"`
	Provider   domaincore.IdentityProvider `json:"provider" validate:"required"`
	ProviderID string                      `json:"provider_id" validate:"required"`
	Email      string                      `json:"email" validate:"email"`
	Username   string                      `json:"username"`
	Verified   bool                        `json:"verified"`
	Metadata   *map[string]any             `json:"metadata,omitempty"`
}

// UserWithIdentity response containing user and their identity
type UserWithIdentity struct {
	User       *domaincore.User         `json:"user"`
	Identity   *domaincore.UserIdentity `json:"identity"`
	WasCreated bool                     `json:"was_created"` // true if user was just created
}
