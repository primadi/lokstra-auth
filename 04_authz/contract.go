package authz

import (
	"context"

	subject "github.com/primadi/lokstra-auth/03_subject"
)

// Action represents an action to be performed on a resource
type Action string

const (
	ActionRead    Action = "read"
	ActionWrite   Action = "write"
	ActionCreate  Action = "create"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionExecute Action = "execute"
)

// Resource represents a protected resource
type Resource struct {
	// Type is the resource type (e.g., "document", "api", "feature")
	Type string

	// ID is the resource identifier
	ID string

	// TenantID is the tenant this resource belongs to (for multi-tenant isolation)
	TenantID string

	// AppID is the app this resource belongs to (for app-level isolation)
	AppID string

	// BranchID is the branch this resource belongs to (optional, for branch-scoped resources)
	BranchID string

	// Attributes contains resource attributes for ABAC
	Attributes map[string]any
}

// AuthorizationRequest represents an authorization request
type AuthorizationRequest struct {
	// Subject is the identity context of the requester
	Subject *subject.IdentityContext

	// Resource is the resource being accessed
	Resource *Resource

	// Action is the action being performed
	Action Action

	// Context contains additional context for the request
	Context map[string]any
}

// AuthorizationDecision represents the result of an authorization check
type AuthorizationDecision struct {
	// Allowed indicates whether the action is permitted
	Allowed bool

	// Reason provides the reason for the decision
	Reason string

	// Obligations are actions that must be performed if access is granted
	Obligations []string

	// Metadata contains additional decision metadata
	Metadata map[string]any
}

// PolicyEvaluator evaluates policies to make authorization decisions
type PolicyEvaluator interface {
	// Evaluate evaluates policies for an authorization request
	Evaluate(ctx context.Context, request *AuthorizationRequest) (*AuthorizationDecision, error)
}

// PermissionChecker checks if a subject has specific permissions
type PermissionChecker interface {
	// HasPermission checks if the subject has a specific permission
	HasPermission(ctx context.Context, identity *subject.IdentityContext, permission string) (bool, error)

	// HasAnyPermission checks if the subject has any of the specified permissions
	HasAnyPermission(ctx context.Context, identity *subject.IdentityContext, permissions ...string) (bool, error)

	// HasAllPermissions checks if the subject has all of the specified permissions
	HasAllPermissions(ctx context.Context, identity *subject.IdentityContext, permissions ...string) (bool, error)
}

// RoleChecker checks if a subject has specific roles
type RoleChecker interface {
	// HasRole checks if the subject has a specific role
	HasRole(ctx context.Context, identity *subject.IdentityContext, role string) (bool, error)

	// HasAnyRole checks if the subject has any of the specified roles
	HasAnyRole(ctx context.Context, identity *subject.IdentityContext, roles ...string) (bool, error)

	// HasAllRoles checks if the subject has all of the specified roles
	HasAllRoles(ctx context.Context, identity *subject.IdentityContext, roles ...string) (bool, error)
}

// AccessControlList manages ACLs for resources (tenant-scoped)
type AccessControlList interface {
	// Grant grants access to a resource (scoped to tenant+app)
	Grant(ctx context.Context, tenantID, appID, subjectID string, resource *Resource, action Action) error

	// Revoke revokes access to a resource (scoped to tenant+app)
	Revoke(ctx context.Context, tenantID, appID, subjectID string, resource *Resource, action Action) error

	// Check checks if a subject has access to a resource (scoped to tenant+app)
	Check(ctx context.Context, tenantID, appID, subjectID string, resource *Resource, action Action) (bool, error)

	// List lists all permissions for a subject on a resource (scoped to tenant+app)
	List(ctx context.Context, tenantID, appID, subjectID string, resource *Resource) ([]Action, error)
}

// Policy represents an authorization policy (tenant and app scoped)
type Policy struct {
	// ID is the policy identifier
	ID string

	// TenantID is the tenant this policy belongs to (required for isolation)
	TenantID string

	// AppID is the app this policy belongs to (optional, if nil applies to all apps in tenant)
	AppID string

	// Name is the policy name
	Name string

	// Description describes the policy
	Description string

	// Effect is the policy effect ("allow" or "deny")
	Effect string

	// Subjects are the subjects this policy applies to
	Subjects []string

	// Resources are the resources this policy applies to
	Resources []string

	// Actions are the actions this policy applies to
	Actions []Action

	// Conditions are additional conditions for the policy
	Conditions map[string]any
}

// PolicyStore stores and retrieves policies (tenant-scoped)
type PolicyStore interface {
	// Create creates a new policy (must include tenantID)
	Create(ctx context.Context, policy *Policy) error

	// Get retrieves a policy by ID (scoped to tenant)
	Get(ctx context.Context, tenantID, policyID string) (*Policy, error)

	// Update updates an existing policy (scoped to tenant)
	Update(ctx context.Context, policy *Policy) error

	// Delete deletes a policy (scoped to tenant)
	Delete(ctx context.Context, tenantID, policyID string) error

	// List lists all policies for a tenant
	List(ctx context.Context, tenantID string) ([]*Policy, error)

	// ListByApp lists all policies for a specific app within a tenant
	ListByApp(ctx context.Context, tenantID, appID string) ([]*Policy, error)

	// FindBySubject finds policies for a subject within a tenant+app
	FindBySubject(ctx context.Context, tenantID, appID, subjectID string) ([]*Policy, error)

	// FindByResource finds policies for a resource within a tenant+app
	FindByResource(ctx context.Context, tenantID, appID, resourceType, resourceID string) ([]*Policy, error)
}

// Authorizer combines multiple authorization checks
type Authorizer interface {
	PolicyEvaluator
	PermissionChecker
	RoleChecker
}

// AttributeProvider provides attributes for ABAC (tenant-scoped)
type AttributeProvider interface {
	// GetSubjectAttributes retrieves attributes for a subject within a tenant
	GetSubjectAttributes(ctx context.Context, tenantID, subjectID string) (map[string]any, error)

	// GetResourceAttributes retrieves attributes for a resource
	GetResourceAttributes(ctx context.Context, resource *Resource) (map[string]any, error)

	// GetEnvironmentAttributes retrieves environment attributes
	GetEnvironmentAttributes(ctx context.Context) (map[string]any, error)
}
