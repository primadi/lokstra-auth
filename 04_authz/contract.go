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

// AccessControlList manages ACLs for resources
type AccessControlList interface {
	// Grant grants access to a resource
	Grant(ctx context.Context, subjectID string, resource *Resource, action Action) error

	// Revoke revokes access to a resource
	Revoke(ctx context.Context, subjectID string, resource *Resource, action Action) error

	// Check checks if a subject has access to a resource
	Check(ctx context.Context, subjectID string, resource *Resource, action Action) (bool, error)

	// List lists all permissions for a subject on a resource
	List(ctx context.Context, subjectID string, resource *Resource) ([]Action, error)
}

// Policy represents an authorization policy
type Policy struct {
	// ID is the policy identifier
	ID string

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

// PolicyStore stores and retrieves policies
type PolicyStore interface {
	// Create creates a new policy
	Create(ctx context.Context, policy *Policy) error

	// Get retrieves a policy by ID
	Get(ctx context.Context, policyID string) (*Policy, error)

	// Update updates an existing policy
	Update(ctx context.Context, policy *Policy) error

	// Delete deletes a policy
	Delete(ctx context.Context, policyID string) error

	// List lists all policies
	List(ctx context.Context) ([]*Policy, error)

	// FindBySubject finds policies for a subject
	FindBySubject(ctx context.Context, subjectID string) ([]*Policy, error)

	// FindByResource finds policies for a resource
	FindByResource(ctx context.Context, resourceType string, resourceID string) ([]*Policy, error)
}

// Authorizer combines multiple authorization checks
type Authorizer interface {
	PolicyEvaluator
	PermissionChecker
	RoleChecker
}

// AttributeProvider provides attributes for ABAC
type AttributeProvider interface {
	// GetSubjectAttributes retrieves attributes for a subject
	GetSubjectAttributes(ctx context.Context, subjectID string) (map[string]any, error)

	// GetResourceAttributes retrieves attributes for a resource
	GetResourceAttributes(ctx context.Context, resource *Resource) (map[string]any, error)

	// GetEnvironmentAttributes retrieves environment attributes
	GetEnvironmentAttributes(ctx context.Context) (map[string]any, error)
}
