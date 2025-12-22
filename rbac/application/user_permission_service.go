package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra-auth/rbac/domain"
	"github.com/primadi/lokstra/core/request"
)

// UserPermissionService handles user-permission assignment operations
// @RouterService name="user-permission-service", prefix="${api-auth-prefix:/api/auth}/rbac/tenants/{tenant_id}/apps/{app_id}/users/{user_id}/permissions", middlewares=["recovery", "request_logger", "auth"]
type UserPermissionService struct {
	// @Inject "user-permission-store"
	UserPermissionStore repository.UserPermissionStore
	// @Inject "permission-store"
	PermissionStore repository.PermissionStore
}

// AssignPermissionToUser assigns a permission directly to a user
// @Route "POST /"
func (s *UserPermissionService) AssignPermissionToUser(ctx *request.Context, req *domain.AssignPermissionToUserRequest) error {
	// Verify permission exists
	_, err := s.PermissionStore.Get(ctx, req.TenantID, req.AppID, req.PermissionID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	userPermission := &domain.UserPermission{
		UserID:       req.UserID,
		TenantID:     req.TenantID,
		AppID:        req.AppID,
		PermissionID: req.PermissionID,
		GrantedAt:    time.Now(),
		RevokedAt:    nil,
	}

	if err := s.UserPermissionStore.AssignPermission(ctx, userPermission); err != nil {
		return fmt.Errorf("failed to assign permission to user: %w", err)
	}

	return nil
}

// RevokePermissionFromUser revokes a permission from a user
// @Route "DELETE /{permission_id}"
func (s *UserPermissionService) RevokePermissionFromUser(ctx *request.Context, req *domain.RevokePermissionFromUserRequest) error {
	if err := s.UserPermissionStore.RevokePermission(ctx, req.TenantID, req.AppID, req.UserID, req.PermissionID); err != nil {
		return fmt.Errorf("failed to revoke permission from user: %w", err)
	}
	return nil
}

// ListUserPermissions lists all permissions for a user (including from roles)
// @Route "GET /"
func (s *UserPermissionService) ListUserPermissions(ctx *request.Context, req *domain.ListUserPermissionsRequest) ([]*domain.Permission, error) {
	permissions, err := s.UserPermissionStore.ListUserPermissionsWithRoles(ctx, req.TenantID, req.AppID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user permissions: %w", err)
	}
	return permissions, nil
}
