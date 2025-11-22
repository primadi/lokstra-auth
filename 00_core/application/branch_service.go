package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/00_core/domain"
	"github.com/primadi/lokstra-auth/00_core/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

// BranchService manages branch lifecycle and operations within apps
// @RouterService name="branch-service", prefix="/api/registration/tenants/{tenant_id}/apps/{app_id}/branches", middlewares=["recovery", "request-logger"]
type BranchService struct {
	// @Inject "branch-store"
	Store *service.Cached[repository.BranchStore]
	// @Inject "app-service"
	AppService *service.Cached[*AppService]
}

// CreateBranch creates a new branch within an app
// @Route "POST /"
func (s *BranchService) CreateBranch(ctx *request.Context, req *domain.CreateBranchRequest) (*domain.Branch, error) {
	// Validate app exists and is active
	app, err := s.AppService.MustGet().GetApp(ctx, &domain.GetAppRequest{
		ID:       req.AppID,
		TenantID: req.TenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}

	if app.Status != domain.AppStatusActive {
		return nil, fmt.Errorf("app is not active: %s", app.Status)
	}

	// Check if branch code already exists in this app
	existing, err := s.Store.MustGet().Get(ctx, req.TenantID, req.AppID, req.BranchID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("branch with id '%s' already exists in app '%s'", req.BranchID, req.AppID)
	}

	// Create branch
	branch := &domain.Branch{
		TenantID:  req.TenantID,
		AppID:     req.AppID,
		ID:        req.BranchID,
		Name:      req.Name,
		Type:      req.Type,
		Status:    domain.BranchStatusActive,
		Settings:  req.Settings,
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to store
	if err := s.Store.MustGet().Create(ctx, branch); err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	return branch, nil
}

// GetBranch retrieves a branch by ID within an app
// @Route "GET /{id}"
func (s *BranchService) GetBranch(ctx *request.Context, req *domain.GetBranchRequest) (*domain.Branch, error) {
	branch, err := s.Store.MustGet().Get(ctx, req.TenantID, req.AppID, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}

	return branch, nil
}

// UpdateBranch updates an existing branch
// @Route "PUT /{id}"
func (s *BranchService) UpdateBranch(ctx *request.Context, req *domain.UpdateBranchRequest) error {
	// Check if branch exists
	dBranch, err := s.Store.MustGet().Get(ctx, req.TenantID, req.AppID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to get branch: %w", err)
	}

	if req.Name != "" {
		dBranch.Name = req.Name
	}
	if req.Type != "" {
		dBranch.Type = req.Type
	}
	if req.Status != "" {
		dBranch.Status = req.Status
	}
	if req.Settings != nil {
		dBranch.Settings = req.Settings
	}

	if req.Metadata != nil {
		dBranch.Metadata = req.Metadata
	}

	// Update timestamp
	dBranch.UpdatedAt = time.Now()

	// Save to store
	if err := s.Store.MustGet().Update(ctx, dBranch); err != nil {
		return fmt.Errorf("failed to update branch: %w", err)
	}

	return nil
}

// DeleteBranch deletes a branch
// @Route "DELETE /{id}"
func (s *BranchService) DeleteBranch(ctx *request.Context, req *domain.DeleteBranchRequest) error {
	// Check if branch exists
	exists, err := s.Store.MustGet().Exists(ctx, req.TenantID, req.AppID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("branch not found: %s in app %s", req.ID, req.AppID)
	}

	// Delete from store
	if err := s.Store.MustGet().Delete(ctx, req.TenantID, req.AppID, req.ID); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}

// ListBranches lists all branches for an app
// @Route "GET /"
func (s *BranchService) ListBranches(ctx *request.Context, req *domain.ListBranchesRequest) ([]*domain.Branch, error) {
	if req.Type != "" {
		branches, err := s.Store.MustGet().ListByType(ctx, req.TenantID, req.AppID, req.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to list branches by type: %w", err)
		}

		return branches, nil
	}

	branches, err := s.Store.MustGet().List(ctx, req.TenantID, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	return branches, nil
}

// ActivateBranch activates a branch
// @Route "POST /{id}/activate"
func (s *BranchService) ActivateBranch(ctx *request.Context, req *domain.ActivateBranchRequest) error {
	branch, err := s.Store.MustGet().Get(ctx, req.TenantID, req.AppID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to get branch: %w", err)
	}

	branch.Status = domain.BranchStatusActive
	branch.UpdatedAt = time.Now()

	if err := s.Store.MustGet().Update(ctx, branch); err != nil {
		return fmt.Errorf("failed to activate branch: %w", err)
	}

	return nil
}

// DisableBranch disables a branch
// @Route "POST /{id}/disable"
func (s *BranchService) DisableBranch(ctx *request.Context, req *domain.DisableBranchRequest) error {
	branch, err := s.Store.MustGet().Get(ctx, req.TenantID, req.AppID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to get branch: %w", err)
	}

	branch.Status = domain.BranchStatusDisabled
	branch.UpdatedAt = time.Now()

	if err := s.Store.MustGet().Update(ctx, branch); err != nil {
		return fmt.Errorf("failed to disable branch: %w", err)
	}

	return nil
}
