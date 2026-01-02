package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/core/infrastructure/idgen"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra-auth/rbac/domain"
	"github.com/primadi/lokstra/core/request"
)

// PermissionService handles permission management operations
// @RouterService name="permission-service", prefix="${api-auth-prefix:/api/auth}/rbac/tenants/{tenant_id}/apps/{app_id}/permissions", middlewares=["recovery", "request_logger", "auth"]
type PermissionService struct {
	// @Inject "permission-store"
	PermissionStore repository.PermissionStore
}

// CreatePermission creates a new permission
// @Route "POST /"
func (s *PermissionService) CreatePermission(ctx *request.Context, req *domain.CreatePermissionRequest) (*domain.Permission, error) {
	// Check if permission name already exists
	existing, _ := s.PermissionStore.GetByName(ctx, req.TenantID, req.AppID, req.Name)
	if existing != nil {
		return nil, domain.ErrDuplicatePermissionName
	}

	permission := &domain.Permission{
		ID:          idgen.GenerateID("perm"),
		TenantID:    req.TenantID,
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
		Status:      domain.PermissionStatusActive,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := permission.Validate(); err != nil {
		return nil, err
	}

	if err := s.PermissionStore.Create(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

// GetPermission retrieves a permission by ID
// @Route "GET /{permission_id}"
func (s *PermissionService) GetPermission(ctx *request.Context, req *domain.GetPermissionRequest) (*domain.Permission, error) {
	permission, err := s.PermissionStore.Get(ctx, req.TenantID, req.AppID, req.PermissionID)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

// UpdatePermission updates an existing permission
// @Route "PUT /{permission_id}"
func (s *PermissionService) UpdatePermission(ctx *request.Context, req *domain.UpdatePermissionRequest) (*domain.Permission, error) {
	permission, err := s.PermissionStore.Get(ctx, req.TenantID, req.AppID, req.PermissionID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		// Check for duplicate name
		existing, _ := s.PermissionStore.GetByName(ctx, req.TenantID, req.AppID, *req.Name)
		if existing != nil && existing.ID != permission.ID {
			return nil, domain.ErrDuplicatePermissionName
		}
		permission.Name = *req.Name
	}
	if req.Description != nil {
		permission.Description = *req.Description
	}
	if req.Resource != nil {
		permission.Resource = *req.Resource
	}
	if req.Action != nil {
		permission.Action = *req.Action
	}
	if req.Status != nil {
		permission.Status = *req.Status
	}
	if req.Metadata != nil {
		permission.Metadata = req.Metadata
	}
	permission.UpdatedAt = time.Now()

	if err := s.PermissionStore.Update(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return permission, nil
}

// DeletePermission deletes a permission
// @Route "DELETE /{permission_id}"
func (s *PermissionService) DeletePermission(ctx *request.Context, req *domain.DeletePermissionRequest) error {
	if err := s.PermissionStore.Delete(ctx, req.TenantID, req.AppID, req.PermissionID); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}

// ListPermissions lists all permissions for tenant+app
// @Route "GET /"
func (s *PermissionService) ListPermissions(ctx *request.Context, req *domain.ListPermissionsRequest) ([]*domain.Permission, error) {
	if req.Limit == 0 {
		req.Limit = 100
	}

	permissions, err := s.PermissionStore.ListWithFilters(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, nil
}
