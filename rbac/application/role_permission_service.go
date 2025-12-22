package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra-auth/rbac/domain"
	"github.com/primadi/lokstra/core/request"
)

// RolePermissionService handles role-permission assignment operations
// @RouterService name="role-permission-service", prefix="${api-auth-prefix:/api/auth}/rbac/tenants/{tenant_id}/apps/{app_id}/roles/{role_id}/permissions", middlewares=["recovery", "request_logger", "auth"]
type RolePermissionService struct {
	// @Inject "role-permission-store"
	RolePermissionStore repository.RolePermissionStore
	// @Inject "role-store"
	RoleStore repository.RoleStore
	// @Inject "permission-store"
	PermissionStore repository.PermissionStore
}

// AssignPermissionToRole assigns a permission to a role
// @Route "POST /"
func (s *RolePermissionService) AssignPermissionToRole(ctx *request.Context, req *domain.AssignPermissionToRoleRequest) error {
	// Verify role exists
	_, err := s.RoleStore.Get(ctx, req.TenantID, req.AppID, req.RoleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Verify permission exists
	_, err = s.PermissionStore.Get(ctx, req.TenantID, req.AppID, req.PermissionID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	rolePermission := &domain.RolePermission{
		RoleID:       req.RoleID,
		TenantID:     req.TenantID,
		AppID:        req.AppID,
		PermissionID: req.PermissionID,
		GrantedAt:    time.Now(),
		RevokedAt:    nil,
	}

	if err := s.RolePermissionStore.AssignPermission(ctx, rolePermission); err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

// RevokePermissionFromRole revokes a permission from a role
// @Route "DELETE /{permission_id}"
func (s *RolePermissionService) RevokePermissionFromRole(ctx *request.Context, req *domain.RevokePermissionFromRoleRequest) error {
	if err := s.RolePermissionStore.RevokePermission(ctx, req.TenantID, req.AppID, req.RoleID, req.PermissionID); err != nil {
		return fmt.Errorf("failed to revoke permission from role: %w", err)
	}
	return nil
}

// ListRolePermissions lists all permissions for a role
// @Route "GET /"
func (s *RolePermissionService) ListRolePermissions(ctx *request.Context, req *domain.ListRolePermissionsRequest) ([]*domain.Permission, error) {
	permissions, err := s.RolePermissionStore.ListRolePermissions(ctx, req.TenantID, req.AppID, req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list role permissions: %w", err)
	}
	return permissions, nil
}
