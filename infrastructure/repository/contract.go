package repository

import (
	"context"

	authzdomain "github.com/primadi/lokstra-auth/authz/domain"
	coredomain "github.com/primadi/lokstra-auth/core/domain"
	rbacdomain "github.com/primadi/lokstra-auth/rbac/domain"
	"github.com/primadi/lokstra/core/request"
)

// ============================================================================
// CORE ENTITY STORES
// ============================================================================

// TenantStore defines the interface for tenant persistence
type TenantStore interface {
	// Create creates a new tenant
	Create(ctx context.Context, tenant *coredomain.Tenant) error

	// Get retrieves a tenant by ID
	Get(ctx context.Context, tenantID string) (*coredomain.Tenant, error)

	// Update updates an existing tenant
	Update(ctx context.Context, tenant *coredomain.Tenant) error

	// Delete deletes a tenant
	Delete(ctx context.Context, tenantID string) error

	// List lists all tenants
	List(ctx context.Context) ([]*coredomain.Tenant, error)

	// GetByName retrieves a tenant by name
	GetByName(ctx context.Context, name string) (*coredomain.Tenant, error)

	// Exists checks if a tenant exists
	Exists(ctx context.Context, tenantID string) (bool, error)
}

// AppStore defines the interface for app persistence
type AppStore interface {
	// Create creates a new app
	Create(ctx context.Context, app *coredomain.App) error

	// Get retrieves an app by ID within a tenant
	Get(ctx context.Context, tenantID, appID string) (*coredomain.App, error)

	// Update updates an existing app
	Update(ctx context.Context, app *coredomain.App) error

	// Delete deletes an app
	Delete(ctx context.Context, tenantID, appID string) error

	// List lists all apps for a tenant
	List(ctx context.Context, tenantID string) ([]*coredomain.App, error)

	// GetByName retrieves an app by name within a tenant
	GetByName(ctx context.Context, tenantID, name string) (*coredomain.App, error)

	// Exists checks if an app exists within a tenant
	Exists(ctx context.Context, tenantID, appID string) (bool, error)

	// ListByType lists apps by type within a tenant
	ListByType(ctx context.Context, tenantID string, appType coredomain.AppType) ([]*coredomain.App, error)
}

// BranchStore defines the interface for branch persistence
type BranchStore interface {
	// Create creates a new branch
	Create(ctx context.Context, branch *coredomain.Branch) error

	// Get retrieves a branch by ID within an app
	Get(ctx context.Context, tenantID, appID, branchID string) (*coredomain.Branch, error)

	// Update updates an existing branch
	Update(ctx context.Context, branch *coredomain.Branch) error

	// Delete deletes a branch
	Delete(ctx context.Context, tenantID, appID, branchID string) error

	// List lists all branches for an app
	List(ctx context.Context, tenantID, appID string) ([]*coredomain.Branch, error)

	// Exists checks if a branch exists within an app
	Exists(ctx context.Context, tenantID, appID, branchID string) (bool, error)

	// ListByType lists branches by type within an app
	ListByType(ctx context.Context, tenantID, appID string, branchType coredomain.BranchType) ([]*coredomain.Branch, error)
}

// UserStore defines the interface for user persistence
type UserStore interface {
	// Create creates a new user
	Create(ctx context.Context, user *coredomain.User) error

	// Get retrieves a user by ID within a tenant
	Get(ctx context.Context, tenantID, userID string) (*coredomain.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *coredomain.User) error

	// Delete deletes a user
	Delete(ctx context.Context, tenantID, userID string) error

	// List lists all users for a tenant
	List(ctx context.Context, tenantID string) ([]*coredomain.User, error)

	// GetByUsername retrieves a user by username within a tenant
	GetByUsername(ctx context.Context, tenantID, username string) (*coredomain.User, error)

	// GetByEmail retrieves a user by email within a tenant
	GetByEmail(ctx context.Context, tenantID, email string) (*coredomain.User, error)

	// Exists checks if a user exists within a tenant
	Exists(ctx context.Context, tenantID, userID string) (bool, error)

	// ListByApp lists users assigned to an app
	ListByApp(ctx context.Context, tenantID, appID string) ([]*coredomain.User, error)

	// SetPassword sets or updates user password hash (for basic auth)
	SetPassword(ctx context.Context, tenantID, userID, passwordHash string) error

	// RemovePassword removes user password hash (disables basic auth)
	RemovePassword(ctx context.Context, tenantID, userID string) error
}

// UserIdentityStore defines the interface for user identity persistence
// Manages linked identities from various authentication providers (OAuth2, SAML, etc.)
type UserIdentityStore interface {
	// Create creates a new user identity link
	Create(ctx context.Context, identity *coredomain.UserIdentity) error

	// Get retrieves an identity by ID
	Get(ctx context.Context, tenantID, userID, identityID string) (*coredomain.UserIdentity, error)

	// GetByProvider retrieves an identity by provider for a user
	GetByProvider(ctx context.Context, tenantID, userID string, provider coredomain.IdentityProvider) (*coredomain.UserIdentity, error)

	// Update updates an existing identity
	Update(ctx context.Context, identity *coredomain.UserIdentity) error

	// Delete removes an identity link
	Delete(ctx context.Context, tenantID, userID, identityID string) error

	// List lists all identities for a user
	List(ctx context.Context, tenantID, userID string) ([]*coredomain.UserIdentity, error)

	// FindUserByProvider finds a user by provider identity (for login)
	FindUserByProvider(ctx context.Context, tenantID string, provider coredomain.IdentityProvider, providerID string) (*coredomain.User, error)

	// Exists checks if an identity exists
	Exists(ctx context.Context, tenantID, userID string, provider coredomain.IdentityProvider) (bool, error)
}

// CredentialProviderStore defines the interface for credential provider persistence
type CredentialProviderStore interface {
	// Create creates a new credential provider
	Create(ctx context.Context, provider *coredomain.CredentialProvider) error

	// Get retrieves a provider by ID
	Get(ctx context.Context, tenantID, providerID string) (*coredomain.CredentialProvider, error)

	// Update updates an existing provider
	Update(ctx context.Context, provider *coredomain.CredentialProvider) error

	// Delete deletes a provider
	Delete(ctx context.Context, tenantID, providerID string) error

	// List lists all providers for tenant (optionally filtered by app)
	// appID can be empty string to get all providers (both tenant-level and app-level)
	// appID = "" returns only tenant-level providers (where AppID IS NULL)
	List(ctx context.Context, tenantID, appID string) ([]*coredomain.CredentialProvider, error)

	// ListByType lists all providers of a specific type for tenant+app
	ListByType(ctx context.Context, tenantID, appID string, providerType coredomain.ProviderType) ([]*coredomain.CredentialProvider, error)

	// Exists checks if a provider exists
	Exists(ctx context.Context, tenantID, providerID string) (bool, error)
}

// UserAppStore defines the interface for user-app access relationship
// This only manages "who can access which app", not authorization (roles/permissions)
// Authorization (RBAC/ABAC/ACL/Policy) is handled by authz layer
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
	Store(ctx context.Context, key *coredomain.AppKey) error

	// GetByID retrieves an API key by its internal ID
	GetByID(ctx context.Context, id string) (*coredomain.AppKey, error)

	// GetByKeyID retrieves an API key by key ID within tenant/app scope
	GetByKeyID(ctx context.Context, tenantID, appID, keyID string) (*coredomain.AppKey, error)

	// ListByApp lists all API keys for an app
	ListByApp(ctx context.Context, tenantID, appID string) ([]*coredomain.AppKey, error)

	// ListByTenant lists all API keys for a tenant (across all apps)
	ListByTenant(ctx context.Context, tenantID string) ([]*coredomain.AppKey, error)

	// Update updates an API key
	Update(ctx context.Context, key *coredomain.AppKey) error

	// Revoke revokes an API key
	Revoke(ctx context.Context, tenantID, appID, keyID string) error

	// Delete permanently deletes an API key
	Delete(ctx context.Context, tenantID, appID, keyID string) error
}

// ============================================================================
// RBAC STORES (Roles & Permissions)
// ============================================================================

// RoleStore manages role persistence (tenant+app scoped)
type RoleStore interface {
	// Create creates a new role
	Create(ctx context.Context, role *rbacdomain.Role) error

	// Get retrieves a role by ID
	Get(ctx context.Context, tenantID, appID, roleID string) (*rbacdomain.Role, error)

	// GetByName retrieves a role by name
	GetByName(ctx context.Context, tenantID, appID, name string) (*rbacdomain.Role, error)

	// Update updates an existing role
	Update(ctx context.Context, role *rbacdomain.Role) error

	// Delete deletes a role
	Delete(ctx context.Context, tenantID, appID, roleID string) error

	// List lists all roles for tenant+app
	List(ctx context.Context, tenantID, appID string) ([]*rbacdomain.Role, error)

	// ListWithFilters lists roles with filters
	ListWithFilters(ctx context.Context, filters *rbacdomain.ListRolesRequest) ([]*rbacdomain.Role, error)
}

// UserRoleStore manages user-role assignments (tenant+app scoped)
type UserRoleStore interface {
	// AssignRole assigns a role to a user
	AssignRole(ctx context.Context, userRole *rbacdomain.UserRole) error

	// RevokeRole revokes a role from a user
	RevokeRole(ctx context.Context, tenantID, appID, userID, roleID string) error

	// ListUserRoles lists all roles for a user
	ListUserRoles(ctx context.Context, tenantID, appID, userID string) ([]*rbacdomain.Role, error)

	// ListRoleUsers lists all users with a specific role
	ListRoleUsers(ctx context.Context, tenantID, appID, roleID string) ([]string, error)

	// HasRole checks if user has a specific role
	HasRole(ctx context.Context, tenantID, appID, userID, roleID string) (bool, error)
}

// PermissionStore manages permission persistence (tenant+app scoped)
type PermissionStore interface {
	// Create creates a new permission
	Create(ctx context.Context, permission *rbacdomain.Permission) error

	// Get retrieves a permission by ID
	Get(ctx context.Context, tenantID, appID, permissionID string) (*rbacdomain.Permission, error)

	// GetByName retrieves a permission by name
	GetByName(ctx context.Context, tenantID, appID, name string) (*rbacdomain.Permission, error)

	// Update updates an existing permission
	Update(ctx context.Context, permission *rbacdomain.Permission) error

	// Delete deletes a permission
	Delete(ctx context.Context, tenantID, appID, permissionID string) error

	// List lists all permissions for tenant+app
	List(ctx context.Context, tenantID, appID string) ([]*rbacdomain.Permission, error)

	// ListWithFilters lists permissions with filters
	ListWithFilters(ctx context.Context, filters *rbacdomain.ListPermissionsRequest) ([]*rbacdomain.Permission, error)
}

// RolePermissionStore manages role-permission assignments (tenant+app scoped)
type RolePermissionStore interface {
	// AssignPermission assigns a permission to a role
	AssignPermission(ctx context.Context, rolePermission *rbacdomain.RolePermission) error

	// RevokePermission revokes a permission from a role
	RevokePermission(ctx context.Context, tenantID, appID, roleID, permissionID string) error

	// ListRolePermissions lists all permissions for a role
	ListRolePermissions(ctx context.Context, tenantID, appID, roleID string) ([]*rbacdomain.Permission, error)

	// ListPermissionRoles lists all roles with a specific permission
	ListPermissionRoles(ctx context.Context, tenantID, appID, permissionID string) ([]string, error)

	// HasPermission checks if role has a specific permission
	HasPermission(ctx context.Context, tenantID, appID, roleID, permissionID string) (bool, error)
}

// UserPermissionStore manages direct user-permission assignments (tenant+app scoped)
type UserPermissionStore interface {
	// AssignPermission assigns a permission directly to a user
	AssignPermission(ctx context.Context, userPermission *rbacdomain.UserPermission) error

	// RevokePermission revokes a permission from a user
	RevokePermission(ctx context.Context, tenantID, appID, userID, permissionID string) error

	// ListUserPermissions lists all direct permissions for a user (not including role permissions)
	ListUserPermissions(ctx context.Context, tenantID, appID, userID string) ([]*rbacdomain.Permission, error)

	// ListUserPermissionsWithRoles lists all permissions for a user (including from roles)
	ListUserPermissionsWithRoles(ctx context.Context, tenantID, appID, userID string) ([]*rbacdomain.Permission, error)

	// ListPermissionUsers lists all users with a specific direct permission
	ListPermissionUsers(ctx context.Context, tenantID, appID, permissionID string) ([]string, error)

	// HasPermission checks if user has a specific permission (direct or via role)
	HasPermission(ctx context.Context, tenantID, appID, userID, permissionID string) (bool, error)
}

// PermissionCompositionStore manages compound permission compositions
type PermissionCompositionStore interface {
	// Create creates a new permission composition (adds child to compound permission)
	Create(ctx context.Context, composition *rbacdomain.PermissionComposition) error

	// Delete removes a child permission from compound permission
	Delete(ctx context.Context, tenantID, appID, parentPermissionID, childPermissionID string) error

	// ListByParent lists all child permissions for a compound permission
	ListByParent(ctx context.Context, tenantID, appID, parentPermissionID string) ([]*rbacdomain.PermissionComposition, error)

	// ListByChild finds all compound permissions that include a specific permission
	ListByChild(ctx context.Context, tenantID, appID, childPermissionID string) ([]*rbacdomain.PermissionComposition, error)

	// GetEffectivePermissions recursively resolves all permissions (including nested compounds)
	GetEffectivePermissions(ctx context.Context, tenantID, appID, permissionID string) ([]string, error)

	// Exists checks if a composition exists
	Exists(ctx context.Context, tenantID, appID, parentPermissionID, childPermissionID string) (bool, error)

	// HasCircularDependency checks if adding this composition would create a circular dependency
	HasCircularDependency(ctx context.Context, tenantID, appID, parentPermissionID, childPermissionID string) (bool, error)
}

// ============================================================================
// AUTHORIZATION STORES (Policy-based)
// ============================================================================

// PolicyStore defines the interface for policy persistence
type PolicyStore interface {
	// Create creates a new policy
	Create(ctx *request.Context, policy *authzdomain.Policy) error

	// Get retrieves a policy by ID
	Get(ctx *request.Context, tenantID, appID, policyID string) (*authzdomain.Policy, error)

	// GetByName retrieves a policy by name within tenant+app scope
	GetByName(ctx *request.Context, tenantID, appID, name string) (*authzdomain.Policy, error)

	// Update updates an existing policy
	Update(ctx *request.Context, policy *authzdomain.Policy) error

	// Delete deletes a policy (soft delete)
	Delete(ctx *request.Context, tenantID, appID, policyID string) error

	// List lists all policies for tenant+app
	List(ctx *request.Context, tenantID, appID string) ([]*authzdomain.Policy, error)

	// ListWithFilters lists policies with filters
	ListWithFilters(ctx *request.Context, req *authzdomain.ListPoliciesRequest) ([]*authzdomain.Policy, error)

	// FindBySubject finds policies that match a subject
	FindBySubject(ctx *request.Context, tenantID, appID, subjectID string) ([]*authzdomain.Policy, error)

	// FindByResource finds policies that match a resource pattern
	FindByResource(ctx *request.Context, tenantID, appID, resourceType, resourceID string) ([]*authzdomain.Policy, error)

	// Exists checks if a policy exists
	Exists(ctx *request.Context, tenantID, appID, policyID string) (bool, error)
}

// ============================================================================
// AUDIT LOG STORE
// ============================================================================

// AuditLogStore defines the interface for audit log persistence
type AuditLogStore interface {
	// Create creates a new audit log entry
	Create(ctx context.Context, log *coredomain.AuditLog) error

	// Get retrieves an audit log by ID
	Get(ctx context.Context, id int64) (*coredomain.AuditLog, error)

	// List lists audit logs with filters
	List(ctx context.Context, req *coredomain.ListAuditLogsRequest) ([]*coredomain.AuditLog, error)

	// Count counts audit logs matching filters
	Count(ctx context.Context, req *coredomain.ListAuditLogsRequest) (int64, error)

	// CleanupOld deletes audit logs older than specified days
	CleanupOld(ctx context.Context, daysToKeep int) (int, error)
}
