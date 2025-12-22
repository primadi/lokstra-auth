package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/core/infrastructure/idgen"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra-auth/rbac/domain"
	"github.com/primadi/lokstra/core/request"
)

// RoleService handles role management operations
// @RouterService name="role-service", prefix="${api-auth-prefix:/api/auth}/rbac/tenants/{tenant_id}/apps/{app_id}/roles", middlewares=["recovery", "request_logger", "auth"]
type RoleService struct {
	// @Inject "role-store"
	RoleStore repository.RoleStore
}

// CreateRole creates a new role
// @Route "POST /"
func (s *RoleService) CreateRole(ctx *request.Context, req *domain.CreateRoleRequest) (*domain.Role, error) {
	// Check if role name already exists
	existing, _ := s.RoleStore.GetByName(ctx, req.TenantID, req.AppID, req.Name)
	if existing != nil {
		return nil, domain.ErrDuplicateRoleName
	}

	role := &domain.Role{
		ID:          idgen.GenerateID("role"),
		TenantID:    req.TenantID,
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Status:      domain.RoleStatusActive,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := role.Validate(); err != nil {
		return nil, err
	}

	if err := s.RoleStore.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetRole retrieves a role by ID
// @Route "GET /{role_id}"
func (s *RoleService) GetRole(ctx *request.Context, req *domain.GetRoleRequest) (*domain.Role, error) {
	role, err := s.RoleStore.Get(ctx, req.TenantID, req.AppID, req.RoleID)
	if err != nil {
		return nil, err
	}
	return role, nil
}

// UpdateRole updates an existing role
// @Route "PUT /{role_id}"
func (s *RoleService) UpdateRole(ctx *request.Context, req *domain.UpdateRoleRequest) (*domain.Role, error) {
	role, err := s.RoleStore.Get(ctx, req.TenantID, req.AppID, req.RoleID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		// Check for duplicate name
		existing, _ := s.RoleStore.GetByName(ctx, req.TenantID, req.AppID, *req.Name)
		if existing != nil && existing.ID != role.ID {
			return nil, domain.ErrDuplicateRoleName
		}
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Status != nil {
		role.Status = *req.Status
	}
	if req.Metadata != nil {
		role.Metadata = req.Metadata
	}
	role.UpdatedAt = time.Now()

	if err := s.RoleStore.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

// DeleteRole deletes a role
// @Route "DELETE /{role_id}"
func (s *RoleService) DeleteRole(ctx *request.Context, req *domain.DeleteRoleRequest) error {
	if err := s.RoleStore.Delete(ctx, req.TenantID, req.AppID, req.RoleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

// ListRoles lists all roles for tenant+app
// @Route "GET /"
func (s *RoleService) ListRoles(ctx *request.Context, req *domain.ListRolesRequest) ([]*domain.Role, error) {
	if req.Limit == 0 {
		req.Limit = 100
	}

	roles, err := s.RoleStore.ListWithFilters(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}
