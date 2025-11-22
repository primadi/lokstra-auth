package repository

import (
	"context"

	"github.com/primadi/lokstra-auth/00_core/domain"
)

// TenantStore defines the interface for tenant persistence
type TenantStore interface {
	// Create creates a new tenant
	Create(ctx context.Context, tenant *domain.Tenant) error

	// Get retrieves a tenant by ID
	Get(ctx context.Context, tenantID string) (*domain.Tenant, error)

	// Update updates an existing tenant
	Update(ctx context.Context, tenant *domain.Tenant) error

	// Delete deletes a tenant
	Delete(ctx context.Context, tenantID string) error

	// List lists all tenants
	List(ctx context.Context) ([]*domain.Tenant, error)

	// GetByName retrieves a tenant by name
	GetByName(ctx context.Context, name string) (*domain.Tenant, error)

	// Exists checks if a tenant exists
	Exists(ctx context.Context, tenantID string) (bool, error)
}

// AppStore defines the interface for app persistence
type AppStore interface {
	// Create creates a new app
	Create(ctx context.Context, app *domain.App) error

	// Get retrieves an app by ID within a tenant
	Get(ctx context.Context, tenantID, appID string) (*domain.App, error)

	// Update updates an existing app
	Update(ctx context.Context, app *domain.App) error

	// Delete deletes an app
	Delete(ctx context.Context, tenantID, appID string) error

	// List lists all apps for a tenant
	List(ctx context.Context, tenantID string) ([]*domain.App, error)

	// GetByName retrieves an app by name within a tenant
	GetByName(ctx context.Context, tenantID, name string) (*domain.App, error)

	// Exists checks if an app exists within a tenant
	Exists(ctx context.Context, tenantID, appID string) (bool, error)

	// ListByType lists apps by type within a tenant
	ListByType(ctx context.Context, tenantID string, appType domain.AppType) ([]*domain.App, error)
}

// BranchStore defines the interface for branch persistence
type BranchStore interface {
	// Create creates a new branch
	Create(ctx context.Context, branch *domain.Branch) error

	// Get retrieves a branch by ID within an app
	Get(ctx context.Context, tenantID, appID, branchID string) (*domain.Branch, error)

	// Update updates an existing branch
	Update(ctx context.Context, branch *domain.Branch) error

	// Delete deletes a branch
	Delete(ctx context.Context, tenantID, appID, branchID string) error

	// List lists all branches for an app
	List(ctx context.Context, tenantID, appID string) ([]*domain.Branch, error)

	// Exists checks if a branch exists within an app
	Exists(ctx context.Context, tenantID, appID, branchID string) (bool, error)

	// ListByType lists branches by type within an app
	ListByType(ctx context.Context, tenantID, appID string, branchType domain.BranchType) ([]*domain.Branch, error)
}

// UserStore defines the interface for user persistence
type UserStore interface {
	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error

	// Get retrieves a user by ID within a tenant
	Get(ctx context.Context, tenantID, userID string) (*domain.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *domain.User) error

	// Delete deletes a user
	Delete(ctx context.Context, tenantID, userID string) error

	// List lists all users for a tenant
	List(ctx context.Context, tenantID string) ([]*domain.User, error)

	// GetByUsername retrieves a user by username within a tenant
	GetByUsername(ctx context.Context, tenantID, username string) (*domain.User, error)

	// GetByEmail retrieves a user by email within a tenant
	GetByEmail(ctx context.Context, tenantID, email string) (*domain.User, error)

	// Exists checks if a user exists within a tenant
	Exists(ctx context.Context, tenantID, userID string) (bool, error)

	// ListByApp lists users assigned to an app
	ListByApp(ctx context.Context, tenantID, appID string) ([]*domain.User, error)
}

// UserAppStore defines the interface for user-app access relationship
// This only manages "who can access which app", not authorization (roles/permissions)
// Authorization (RBAC/ABAC/ACL/Policy) is handled by 04_authz layer
type UserAppStore interface {
	// GrantAccess grants a user access to an app
	GrantAccess(ctx context.Context, tenantID, appID, userID string) error

	// RevokeAccess revokes a user's access to an app
	RevokeAccess(ctx context.Context, tenantID, appID, userID string) error

	// HasAccess checks if user has access to an app
	HasAccess(ctx context.Context, tenantID, appID, userID string) (bool, error)

	// ListUserApps lists all app IDs a user has access to
	ListUserApps(ctx context.Context, tenantID, userID string) ([]string, error)

	// ListAppUsers lists all user IDs who have access to an app
	ListAppUsers(ctx context.Context, tenantID, appID string) ([]string, error)
}

// AppKeyStore defines the interface for app key persistence
type AppKeyStore interface {
	// Store saves a new API key
	Store(ctx context.Context, key *domain.AppKey) error

	// GetByID retrieves an API key by its internal ID
	GetByID(ctx context.Context, id string) (*domain.AppKey, error)

	// GetByKeyID retrieves an API key by key ID within tenant/app scope
	GetByKeyID(ctx context.Context, tenantID, appID, keyID string) (*domain.AppKey, error)

	// ListByApp lists all API keys for an app
	ListByApp(ctx context.Context, tenantID, appID string) ([]*domain.AppKey, error)

	// ListByTenant lists all API keys for a tenant (across all apps)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.AppKey, error)

	// Update updates an API key
	Update(ctx context.Context, key *domain.AppKey) error

	// Revoke revokes an API key
	Revoke(ctx context.Context, keyID string) error

	// Delete permanently deletes an API key
	Delete(ctx context.Context, id string) error
}
