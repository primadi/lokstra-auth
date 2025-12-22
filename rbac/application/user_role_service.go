package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra-auth/rbac/domain"
	"github.com/primadi/lokstra/core/request"
)

// UserRoleService handles user-role assignment operations
// @RouterService name="user-role-service", prefix="${api-auth-prefix:/api/auth}/rbac/tenants/{tenant_id}/apps/{app_id}/users/{user_id}/roles", middlewares=["recovery", "request_logger", "auth"]
type UserRoleService struct {
	// @Inject "user-role-store"
	UserRoleStore repository.UserRoleStore
	// @Inject "role-store"
	RoleStore repository.RoleStore
}

// AssignRole assigns a role to a user
// @Route "POST /"
func (s *UserRoleService) AssignRole(ctx *request.Context, req *domain.AssignRoleRequest) error {
	// Verify role exists
	_, err := s.RoleStore.Get(ctx, req.TenantID, req.AppID, req.RoleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	userRole := &domain.UserRole{
		UserID:    req.UserID,
		TenantID:  req.TenantID,
		AppID:     req.AppID,
		RoleID:    req.RoleID,
		GrantedAt: time.Now(),
		RevokedAt: nil,
	}

	if err := s.UserRoleStore.AssignRole(ctx, userRole); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RevokeRole revokes a role from a user
// @Route "DELETE /{role_id}"
func (s *UserRoleService) RevokeRole(ctx *request.Context, req *domain.RevokeRoleRequest) error {
	if err := s.UserRoleStore.RevokeRole(ctx, req.TenantID, req.AppID, req.UserID, req.RoleID); err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}
	return nil
}

// ListUserRoles lists all roles for a user
// @Route "GET /"
func (s *UserRoleService) ListUserRoles(ctx *request.Context, req *domain.ListUserRolesRequest) ([]*domain.Role, error) {
	roles, err := s.UserRoleStore.ListUserRoles(ctx, req.TenantID, req.AppID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user roles: %w", err)
	}
	return roles, nil
}
