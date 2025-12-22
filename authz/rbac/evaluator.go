package rbac

import (
	"context"
	"fmt"
	"strings"

	authz "github.com/primadi/lokstra-auth/authz"
	identity "github.com/primadi/lokstra-auth/identity"
)

// Evaluator is an RBAC policy evaluator (multi-tenant aware)
type Evaluator struct {
	// rolePermissions: map[tenantID:appID:role] -> []permissions
	// Composite key ensures tenant+app isolation
	rolePermissions map[string][]string
}

// NewEvaluator creates a new RBAC evaluator
func NewEvaluator(rolePermissions map[string][]string) *Evaluator {
	return &Evaluator{
		rolePermissions: rolePermissions,
	}
}

// makeRoleKey creates a composite key for tenant+app+role
func (e *Evaluator) makeRoleKey(tenantID, appID, role string) string {
	return tenantID + ":" + appID + ":" + role
}

// Evaluate evaluates policies for an authorization request
func (e *Evaluator) Evaluate(ctx context.Context, request *authz.AuthorizationRequest) (*authz.AuthorizationDecision, error) {
	// Extract tenant and app from identity context
	tenantID := request.Subject.TenantID
	appID := request.Subject.AppID

	// Validate tenant+app match resource
	if request.Resource.TenantID != "" && request.Resource.TenantID != tenantID {
		return &authz.AuthorizationDecision{
			Allowed: false,
			Reason:  "resource tenant mismatch",
		}, nil
	}
	if request.Resource.AppID != "" && request.Resource.AppID != appID {
		return &authz.AuthorizationDecision{
			Allowed: false,
			Reason:  "resource app mismatch",
		}, nil
	}

	// Build required permission from resource and action
	requiredPermission := fmt.Sprintf("%s:%s:%s",
		request.Resource.Type,
		request.Resource.ID,
		request.Action)

	// Also check simple permission format (action:type)
	simplePermission := fmt.Sprintf("%s:%s", string(request.Action), request.Resource.Type)

	// Check if any of the subject's roles have the required permission
	for _, role := range request.Subject.Roles {
		// Use composite key for tenant+app+role lookup
		roleKey := e.makeRoleKey(tenantID, appID, role)
		permissions, ok := e.rolePermissions[roleKey]
		if !ok {
			continue
		}

		for _, permission := range permissions {
			if e.matchPermission(permission, requiredPermission) ||
				e.matchPermission(permission, simplePermission) {
				return &authz.AuthorizationDecision{
					Allowed: true,
					Reason:  fmt.Sprintf("role %s has permission %s", role, permission),
				}, nil
			}
		}
	}

	return &authz.AuthorizationDecision{
		Allowed: false,
		Reason:  "no matching role permissions found",
	}, nil
}

// matchPermission checks if a permission pattern matches a required permission
// Supports wildcard matching (e.g., "document:*" matches "document:read", "document:write")
func (e *Evaluator) matchPermission(pattern, required string) bool {
	// Exact match
	if pattern == required {
		return true
	}

	// Wildcard match - full wildcard
	if pattern == "*" {
		return true
	}

	// Wildcard match - prefix matching (e.g., "document:*" matches "document:read")
	if strings.HasSuffix(pattern, ":*") {
		prefix := strings.TrimSuffix(pattern, ":*")
		if strings.HasPrefix(required, prefix+":") || required == prefix {
			return true
		}
	}

	// Wildcard match - pattern with wildcards in segments
	patternParts := strings.Split(pattern, ":")
	requiredParts := strings.Split(required, ":")

	// If pattern has more parts than required, can't match
	if len(patternParts) > len(requiredParts) {
		return false
	}

	// Check each part
	for i, part := range patternParts {
		if part == "*" {
			continue // Wildcard matches anything
		}
		if i >= len(requiredParts) || part != requiredParts[i] {
			return false
		}
	}

	return true
}

// HasPermission checks if the subject has a specific permission (tenant+app scoped)
func (e *Evaluator) HasPermission(ctx context.Context, identity *identity.IdentityContext, permission string) (bool, error) {
	tenantID := identity.TenantID
	appID := identity.AppID

	for _, role := range identity.Roles {
		roleKey := e.makeRoleKey(tenantID, appID, role)
		permissions, ok := e.rolePermissions[roleKey]
		if !ok {
			continue
		}

		for _, p := range permissions {
			if p == permission || p == "*" {
				return true, nil
			}
		}
	}

	return false, nil
}

// HasAnyPermission checks if the subject has any of the specified permissions
func (e *Evaluator) HasAnyPermission(ctx context.Context, identity *identity.IdentityContext, permissions ...string) (bool, error) {
	for _, permission := range permissions {
		has, err := e.HasPermission(ctx, identity, permission)
		if err != nil {
			return false, err
		}
		if has {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if the subject has all of the specified permissions
func (e *Evaluator) HasAllPermissions(ctx context.Context, identity *identity.IdentityContext, permissions ...string) (bool, error) {
	for _, permission := range permissions {
		has, err := e.HasPermission(ctx, identity, permission)
		if err != nil {
			return false, err
		}
		if !has {
			return false, nil
		}
	}
	return true, nil
}

// HasRole checks if the subject has a specific role
func (e *Evaluator) HasRole(ctx context.Context, identity *identity.IdentityContext, role string) (bool, error) {
	return identity.HasRole(role), nil
}

// HasAnyRole checks if the subject has any of the specified roles
func (e *Evaluator) HasAnyRole(ctx context.Context, identity *identity.IdentityContext, roles ...string) (bool, error) {
	return identity.HasAnyRole(roles...), nil
}

// HasAllRoles checks if the subject has all of the specified roles
func (e *Evaluator) HasAllRoles(ctx context.Context, identity *identity.IdentityContext, roles ...string) (bool, error) {
	return identity.HasAllRoles(roles...), nil
}

// AddRolePermission adds a permission to a role (tenant+app scoped)
func (e *Evaluator) AddRolePermission(tenantID, appID, role, permission string) {
	if e.rolePermissions == nil {
		e.rolePermissions = make(map[string][]string)
	}

	roleKey := e.makeRoleKey(tenantID, appID, role)
	permissions, ok := e.rolePermissions[roleKey]
	if !ok {
		e.rolePermissions[roleKey] = []string{permission}
		return
	}

	// Check if permission already exists
	for _, p := range permissions {
		if p == permission {
			return
		}
	}

	e.rolePermissions[roleKey] = append(permissions, permission)
}

// RemoveRolePermission removes a permission from a role (tenant+app scoped)
func (e *Evaluator) RemoveRolePermission(tenantID, appID, role, permission string) {
	roleKey := e.makeRoleKey(tenantID, appID, role)
	permissions, ok := e.rolePermissions[roleKey]
	if !ok {
		return
	}

	for i, p := range permissions {
		if p == permission {
			e.rolePermissions[roleKey] = append(permissions[:i], permissions[i+1:]...)
			return
		}
	}
}

// GetRolePermissions returns all permissions for a role (tenant+app scoped)
func (e *Evaluator) GetRolePermissions(tenantID, appID, role string) []string {
	roleKey := e.makeRoleKey(tenantID, appID, role)
	permissions, ok := e.rolePermissions[roleKey]
	if !ok {
		return []string{}
	}

	// Return a copy to prevent external modifications
	result := make([]string, len(permissions))
	copy(result, permissions)
	return result
}
