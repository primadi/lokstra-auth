package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/00_core/domain"
	"github.com/primadi/lokstra-auth/00_core/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

// AppService manages app lifecycle and operations within tenants
// @RouterService name="app-service", prefix="/api/registration/tenants/{tenant_id}/apps", middlewares=["recovery", "request_logger"]
type AppService struct {
	// @Inject "app-store"
	Store *service.Cached[repository.AppStore]
	// @Inject "user-app-store"
	UserAppStore *service.Cached[repository.UserAppStore]
	// @Inject "tenant-service"
	TenantService *service.Cached[*TenantService]
}

// CreateApp creates a new app within a tenant
// @Route "POST /"
func (s *AppService) CreateApp(ctx *request.Context, req *domain.CreateAppRequest) (*domain.App, error) {
	// Validate tenant exists
	tenant, err := s.TenantService.MustGet().GetTenant(ctx,
		&domain.GetTenantRequest{
			ID: req.TenantID,
		})
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Check tenant status
	if tenant.Status != domain.TenantStatusActive {
		return nil, fmt.Errorf("tenant is not active: %s", tenant.Status)
	}

	// Check if app name already exists in tenant
	existing, err := s.Store.MustGet().GetByName(ctx, req.TenantID, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("app with name '%s' already exists in tenant '%s'",
			req.Name, req.TenantID)
	}

	// Create app
	app := &domain.App{
		TenantID:  req.TenantID,
		ID:        req.ID,
		Name:      req.Name,
		Type:      req.Type,
		Config:    req.Config,
		Status:    domain.AppStatusActive,
		Metadata:  &map[string]any{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to store
	if err := s.Store.MustGet().Create(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to create app: %w", err)
	}

	return app, nil
}

// GetApp retrieves an app by ID within a tenant
// @Route "GET /{id}"
func (s *AppService) GetApp(ctx *request.Context, req *domain.GetAppRequest) (*domain.App, error) {
	app, err := s.Store.MustGet().Get(ctx, req.TenantID, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app: %w", err)
	}

	return app, nil
}

// UpdateApp updates an existing app
// @Route "PUT /{id}"
func (s *AppService) UpdateApp(ctx *request.Context, req *domain.UpdateAppRequest) error {
	// Load existing app
	dApp, err := s.Store.MustGet().Get(ctx, req.TenantID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	if req.Name != "" {
		dApp.Name = req.Name
	}
	if req.Type != "" {
		dApp.Type = req.Type
	}
	if req.Config != nil {
		dApp.Config = req.Config
	}
	if req.Status != "" {
		dApp.Status = req.Status
	}
	if req.Metadata != nil {
		dApp.Metadata = req.Metadata
	}
	dApp.UpdatedAt = time.Now()

	// Save to store
	if err := s.Store.MustGet().Update(ctx, dApp); err != nil {
		return fmt.Errorf("failed to update app: %w", err)
	}

	return nil
}

// DeleteApp deletes an app
// @Route "DELETE /{id}"
func (s *AppService) DeleteApp(ctx *request.Context, req *domain.DeleteAppRequest) error {
	// Check if app exists
	exists, err := s.Store.MustGet().Exists(ctx, req.TenantID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to check app existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("app not found: %s in tenant %s", req.ID, req.TenantID)
	}

	// Delete from store
	if err := s.Store.MustGet().Delete(ctx, req.TenantID, req.ID); err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}

	return nil
}

// ListApps lists all apps for a tenant
// @Route "GET /"
func (s *AppService) ListApps(ctx *request.Context, req *domain.ListAppsRequest) ([]*domain.App, error) {
	if req.Type != "" {
		apps, err := s.Store.MustGet().ListByType(ctx, req.TenantID, req.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to list apps by type: %w", err)
		}

		return apps, nil
	}

	apps, err := s.Store.MustGet().List(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}

	return apps, nil
}

// ActivateApp activates an app
// @Route "POST /{id}/activate"
func (s *AppService) ActivateApp(ctx *request.Context, req *domain.ActivateAppRequest) error {

	app, err := s.Store.MustGet().Get(ctx, req.TenantID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	app.Status = domain.AppStatusActive
	app.UpdatedAt = time.Now()

	if err := s.Store.MustGet().Update(ctx, app); err != nil {
		return fmt.Errorf("failed to activate app: %w", err)
	}

	return nil
}

// SuspendApp suspends an app
// @Route "POST /{id}/suspend"
func (s *AppService) SuspendApp(ctx *request.Context, req *domain.SuspendAppRequest) error {

	app, err := s.Store.MustGet().Get(ctx, req.TenantID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	app.Status = domain.AppStatusDisabled
	app.UpdatedAt = time.Now()

	if err := s.Store.MustGet().Update(ctx, app); err != nil {
		return fmt.Errorf("failed to suspend app: %w", err)
	}

	return nil
}

// ListAppUsers lists all users who have access to an app
// @Route "GET /{app_id}/users"
func (s *AppService) ListAppUsers(ctx *request.Context, req *domain.ListUsersByAppRequest) ([]string, error) {
	userIDs, err := s.UserAppStore.MustGet().ListAppUsers(ctx, req.TenantID, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("failed to list app users: %w", err)
	}

	return userIDs, nil
}
