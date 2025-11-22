package repository

import (
	"context"

	"github.com/primadi/lokstra-auth/01_credential/domain/apikey"
	"github.com/primadi/lokstra-auth/01_credential/domain/basic"
)

// =============================================================================
// Basic Authentication
// =============================================================================

// UserProvider defines the interface for retrieving user credentials.
// Implementations can use in-memory storage, database, or external services.
// All operations are scoped by tenant for proper isolation.
type UserProvider interface {
	// GetUserByUsername retrieves user information by username within a tenant
	GetUserByUsername(ctx context.Context, tenantID, username string) (*basic.User, error)

	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *basic.User) error

	// UpdateUser updates existing user
	UpdateUser(ctx context.Context, user *basic.User) error

	// GetUserByID retrieves user by ID
	GetUserByID(ctx context.Context, tenantID, userID string) (*basic.User, error)
}

// CredentialValidator validates credentials format and complexity requirements.
// This is specific to basic authentication and should be used during USER REGISTRATION/CREATION,
// NOT during login/authentication.
type CredentialValidator interface {
	// ValidateUsername checks if username meets requirements
	ValidateUsername(username string) error

	// ValidatePassword checks if password meets complexity requirements
	ValidatePassword(password string) error
}

// =============================================================================
// API Key Authentication
// =============================================================================

// APIKeyStore defines the interface for API key storage and retrieval
type APIKeyStore interface {
	// GetByKeyID retrieves API key by key ID within tenant and app scope
	GetByKeyID(ctx context.Context, tenantID, appID, keyID string) (*apikey.APIKey, error)

	// Store saves a new API key
	Store(ctx context.Context, key *apikey.APIKey) error

	// UpdateLastUsed updates the last used timestamp
	UpdateLastUsed(ctx context.Context, keyID, timestamp string) error

	// Revoke revokes an API key
	Revoke(ctx context.Context, keyID string) error

	// ListByApp lists all API keys for an app
	ListByApp(ctx context.Context, tenantID, appID string) ([]*apikey.APIKey, error)

	// Delete permanently deletes an API key
	Delete(ctx context.Context, keyID string) error
}
