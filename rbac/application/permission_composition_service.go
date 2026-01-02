package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra-auth/rbac/domain"
	"github.com/primadi/lokstra/core/request"
)

// PermissionCompositionService handles compound permission management
// @RouterService name="permission-composition-service", prefix="${api-auth-prefix:/api/auth}/rbac/tenants/{tenant_id}/apps/{app_id}/permissions/{permission_id}/compositions", middlewares=["recovery", "request_logger", "auth"]
type PermissionCompositionService struct {
	// @Inject "permission-composition-store"
	CompositionStore repository.PermissionCompositionStore
	// @Inject "permission-store"
	PermissionStore repository.PermissionStore
}

// AddChildPermission adds a child permission to a compound permission
// @Route "POST /"
func (s *PermissionCompositionService) AddChildPermission(ctx *request.Context, req *domain.CreatePermissionCompositionRequest) (*domain.PermissionComposition, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify parent permission exists
	_, err := s.PermissionStore.Get(ctx, req.TenantID, req.AppID, req.ParentPermissionID)
	if err != nil {
		return nil, fmt.Errorf("parent permission not found: %w", err)
	}

	// Verify child permission exists
	_, err = s.PermissionStore.Get(ctx, req.TenantID, req.AppID, req.ChildPermissionID)
	if err != nil {
		return nil, fmt.Errorf("child permission not found: %w", err)
	}

	// Check if composition already exists
	exists, err := s.CompositionStore.Exists(ctx, req.TenantID, req.AppID, req.ParentPermissionID, req.ChildPermissionID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrPermissionCompositionAlreadyExists
	}

	// Check for circular dependencies
	hasCircular, err := s.CompositionStore.HasCircularDependency(ctx, req.TenantID, req.AppID, req.ParentPermissionID, req.ChildPermissionID)
	if err != nil {
		return nil, err
	}
	if hasCircular {
		return nil, domain.ErrCircularDependency
	}

	composition := &domain.PermissionComposition{
		ParentPermissionID: req.ParentPermissionID,
		ChildPermissionID:  req.ChildPermissionID,
		TenantID:           req.TenantID,
		AppID:              req.AppID,
		IsRequired:         req.IsRequired,
		Priority:           req.Priority,
		Metadata:           req.Metadata,
		CreatedAt:          time.Now(),
	}

	if err := s.CompositionStore.Create(ctx, composition); err != nil {
		return nil, fmt.Errorf("failed to create permission composition: %w", err)
	}

	return composition, nil
}

// RemoveChildPermission removes a child permission from a compound permission
// @Route "DELETE /{child_permission_id}"
func (s *PermissionCompositionService) RemoveChildPermission(ctx *request.Context, req *domain.DeletePermissionCompositionRequest) error {
	if err := s.CompositionStore.Delete(ctx, req.TenantID, req.AppID, req.ParentPermissionID, req.ChildPermissionID); err != nil {
		return err
	}
	return nil
}

// ListCompositions lists all child permissions for a compound permission
// @Route "GET /"
func (s *PermissionCompositionService) ListCompositions(ctx *request.Context, req *domain.ListPermissionCompositionsRequest) ([]*domain.PermissionComposition, error) {
	if req.ParentPermissionID != nil {
		return s.CompositionStore.ListByParent(ctx, req.TenantID, req.AppID, *req.ParentPermissionID)
	}
	if req.ChildPermissionID != nil {
		return s.CompositionStore.ListByChild(ctx, req.TenantID, req.AppID, *req.ChildPermissionID)
	}
	return nil, fmt.Errorf("either parent_permission_id or child_permission_id must be provided")
}

// GetEffectivePermissions recursively resolves all permissions for a compound permission
// @Route "GET /effective"
func (s *PermissionCompositionService) GetEffectivePermissions(ctx *request.Context, req *domain.GetEffectivePermissionsRequest) (*EffectivePermissionsResponse, error) {
	permissionIDs, err := s.CompositionStore.GetEffectivePermissions(ctx, req.TenantID, req.AppID, req.PermissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get effective permissions: %w", err)
	}

	// Fetch full permission details
	permissions := []*domain.Permission{}
	for _, permID := range permissionIDs {
		perm, err := s.PermissionStore.Get(ctx, req.TenantID, req.AppID, permID)
		if err != nil {
			// Skip if permission not found (might have been deleted)
			continue
		}
		permissions = append(permissions, perm)
	}

	return &EffectivePermissionsResponse{
		ParentPermissionID:   req.PermissionID,
		EffectivePermissions: permissions,
		TotalCount:           len(permissions),
	}, nil
}

// EffectivePermissionsResponse represents the response for effective permissions
type EffectivePermissionsResponse struct {
	ParentPermissionID   string               `json:"parent_permission_id"`
	EffectivePermissions []*domain.Permission `json:"effective_permissions"`
	TotalCount           int                  `json:"total_count"`
}
