package application

import (
	"fmt"

	"github.com/primadi/lokstra-auth/00_core/domain"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

// CredentialConfigService manages credential provider configuration
// @RouterService name="credential-config-service", prefix="/api/registration/config/credentials", middlewares=["recovery", "request-logger"]
type CredentialConfigService struct {
	// @Inject "tenant-service"
	TenantService *service.Cached[*TenantService]
	// @Inject "app-service"
	AppService *service.Cached[*AppService]
}

// GetTenantConfig retrieves tenant-level default credential config
// @Route "GET /tenants/{tenant_id}"
func (s *CredentialConfigService) GetTenantConfig(ctx *request.Context, req *GetTenantCredentialConfigRequest) (*domain.CredentialConfig, error) {
	tenant, err := s.TenantService.MustGet().GetTenant(ctx, &domain.GetTenantRequest{
		ID: req.TenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	if tenant.Config == nil || tenant.Config.DefaultCredentials == nil {
		// Return default config if not set
		return domain.DefaultCredentialConfig(), nil
	}

	return tenant.Config.DefaultCredentials, nil
}

// UpdateTenantConfig updates tenant-level default credential config
// @Route "PUT /tenants/{tenant_id}"
func (s *CredentialConfigService) UpdateTenantConfig(ctx *request.Context, req *UpdateTenantCredentialConfigRequest) (*domain.CredentialConfig, error) {
	tenant, err := s.TenantService.MustGet().GetTenant(ctx, &domain.GetTenantRequest{
		ID: req.TenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Initialize config if nil
	if tenant.Config == nil {
		tenant.Config = &domain.TenantConfig{}
	}

	// Update default credentials
	tenant.Config.DefaultCredentials = req.Config

	// Update tenant
	if _, err := s.TenantService.MustGet().UpdateTenant(ctx, &domain.UpdateTenantRequest{
		ID:       tenant.ID,
		Name:     tenant.Name,
		DBDsn:    tenant.DBDsn,
		DBSchema: tenant.DBSchema,
		Settings: tenant.Settings,
		Config:   tenant.Config,
	}); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant.Config.DefaultCredentials, nil
}

// GetAppConfig retrieves app-specific credential config
// @Route "GET /tenants/{tenant_id}/apps/{app_id}"
func (s *CredentialConfigService) GetAppConfig(ctx *request.Context, req *GetAppCredentialConfigRequest) (*domain.CredentialConfig, error) {
	app, err := s.AppService.MustGet().GetApp(ctx, &domain.GetAppRequest{
		TenantID: req.TenantID,
		ID:       req.AppID,
	})
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}

	if app.Config == nil || app.Config.Credentials == nil {
		// Return tenant default or global default
		return s.GetTenantConfig(ctx, &GetTenantCredentialConfigRequest{
			TenantID: req.TenantID,
		})
	}

	return app.Config.Credentials, nil
}

// UpdateAppConfig updates app-specific credential config
// @Route "PUT /tenants/{tenant_id}/apps/{app_id}"
func (s *CredentialConfigService) UpdateAppConfig(ctx *request.Context, req *UpdateAppCredentialConfigRequest) (*domain.CredentialConfig, error) {
	app, err := s.AppService.MustGet().GetApp(ctx, &domain.GetAppRequest{
		TenantID: req.TenantID,
		ID:       req.AppID,
	})
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}

	// Initialize config if nil
	if app.Config == nil {
		app.Config = &domain.AppConfig{}
	}

	// Update credentials
	app.Config.Credentials = req.Config

	// Update app
	if err := s.AppService.MustGet().UpdateApp(ctx, &domain.UpdateAppRequest{
		TenantID: app.TenantID,
		ID:       app.ID,
		Name:     app.Name,
		Type:     app.Type,
		Config:   app.Config,
		Status:   app.Status,
	}); err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}

	return app.Config.Credentials, nil
}

// =============================================================================
// DTOs
// =============================================================================

type GetTenantCredentialConfigRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
}

type UpdateTenantCredentialConfigRequest struct {
	TenantID string                   `path:"tenant_id" validate:"required"`
	Config   *domain.CredentialConfig `json:"config" validate:"required"`
}

type GetAppCredentialConfigRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
}

type UpdateAppCredentialConfigRequest struct {
	TenantID string                   `path:"tenant_id" validate:"required"`
	AppID    string                   `path:"app_id" validate:"required"`
	Config   *domain.CredentialConfig `json:"config" validate:"required"`
}
