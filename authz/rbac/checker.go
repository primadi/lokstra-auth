package rbac

import (
	"context"

	identity "github.com/primadi/lokstra-auth/identity"
)

// RoleChecker implements role-based access control checks
type RoleChecker struct{}

// NewRoleChecker creates a new RBAC role checker
func NewRoleChecker() *RoleChecker {
	return &RoleChecker{}
}

// HasRole checks if the subject has a specific role
func (c *RoleChecker) HasRole(ctx context.Context, identity *identity.IdentityContext, role string) (bool, error) {
	if identity == nil {
		return false, nil
	}
	return identity.HasRole(role), nil
}

// HasAnyRole checks if the subject has any of the specified roles
func (c *RoleChecker) HasAnyRole(ctx context.Context, identity *identity.IdentityContext, roles ...string) (bool, error) {
	if identity == nil {
		return false, nil
	}
	return identity.HasAnyRole(roles...), nil
}

// HasAllRoles checks if the subject has all of the specified roles
func (c *RoleChecker) HasAllRoles(ctx context.Context, identity *identity.IdentityContext, roles ...string) (bool, error) {
	if identity == nil {
		return false, nil
	}
	return identity.HasAllRoles(roles...), nil
}

// PermissionChecker implements permission-based access control checks
type PermissionChecker struct{}

// NewPermissionChecker creates a new RBAC permission checker
func NewPermissionChecker() *PermissionChecker {
	return &PermissionChecker{}
}

// HasPermission checks if the subject has a specific permission
func (c *PermissionChecker) HasPermission(ctx context.Context, identity *identity.IdentityContext, permission string) (bool, error) {
	if identity == nil {
		return false, nil
	}
	return identity.HasPermission(permission), nil
}

// HasAnyPermission checks if the subject has any of the specified permissions
func (c *PermissionChecker) HasAnyPermission(ctx context.Context, identity *identity.IdentityContext, permissions ...string) (bool, error) {
	if identity == nil {
		return false, nil
	}
	for _, perm := range permissions {
		if identity.HasPermission(perm) {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if the subject has all of the specified permissions
func (c *PermissionChecker) HasAllPermissions(ctx context.Context, identity *identity.IdentityContext, permissions ...string) (bool, error) {
	if identity == nil {
		return false, nil
	}
	for _, perm := range permissions {
		if !identity.HasPermission(perm) {
			return false, nil
		}
	}
	return true, nil
}
