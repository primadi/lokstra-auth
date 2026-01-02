package application

import (
	"fmt"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra/core/request"
)

// CredentialConfigService manages credential provider configuration
// @RouterService name="credential-config-service", prefix="${api-auth-prefix:/api/auth}/core/tenants/{tenant_id}/config/credentials", middlewares=["recovery", "request_logger", "auth"]
type CredentialConfigService struct {
	// @Inject "tenant-service"
	TenantService *TenantService
	// @Inject "app-service"
	AppService *AppService
}

// GetTenantConfig retrieves tenant-level default credential config
// @Route "GET /"
func (s *CredentialConfigService) GetTenantConfig(ctx *request.Context, req *domain.GetTenantCredentialConfigRequest) (*domain.CredentialConfig, error) {
	tenant, err := s.TenantService.GetTenant(ctx, &domain.GetTenantRequest{
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
// @Route "PUT /"
func (s *CredentialConfigService) UpdateTenantConfig(ctx *request.Context, req *domain.UpdateTenantCredentialConfigRequest) (*domain.CredentialConfig, error) {
	tenant, err := s.TenantService.GetTenant(ctx, &domain.GetTenantRequest{
		ID: req.TenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Initialize config if nil
	if tenant.Config == nil {
		tenant.Config = &domain.TenantConfig{}
	}

	// Initialize default credentials if nil
	if tenant.Config.DefaultCredentials == nil {
		tenant.Config.DefaultCredentials = &domain.CredentialConfig{}
	}

	// MERGE credentials config - smart merge based on what's provided
	existing := tenant.Config.DefaultCredentials

	// Basic auth - update only if config is provided
	if req.Config.BasicConfig != nil {
		existing.EnableBasic = req.Config.EnableBasic
		existing.BasicConfig = req.Config.BasicConfig
	} else if req.Config.EnableBasic != existing.EnableBasic {
		// Only enable flag changed without config
		existing.EnableBasic = req.Config.EnableBasic
	}

	// API Key auth
	if req.Config.APIKeyConfig != nil {
		existing.EnableAPIKey = req.Config.EnableAPIKey
		existing.APIKeyConfig = req.Config.APIKeyConfig
	} else if req.Config.EnableAPIKey != existing.EnableAPIKey {
		existing.EnableAPIKey = req.Config.EnableAPIKey
	}

	// OAuth2 auth
	if req.Config.OAuth2Config != nil {
		existing.EnableOAuth2 = req.Config.EnableOAuth2
		existing.OAuth2Config = req.Config.OAuth2Config
	} else if req.Config.EnableOAuth2 != existing.EnableOAuth2 {
		existing.EnableOAuth2 = req.Config.EnableOAuth2
	}

	// Passwordless auth
	if req.Config.PasswordlessConfig != nil {
		existing.EnablePasswordless = req.Config.EnablePasswordless
		existing.PasswordlessConfig = req.Config.PasswordlessConfig
	} else if req.Config.EnablePasswordless != existing.EnablePasswordless {
		existing.EnablePasswordless = req.Config.EnablePasswordless
	}

	// Passkey auth
	if req.Config.PasskeyConfig != nil {
		existing.EnablePasskey = req.Config.EnablePasskey
		existing.PasskeyConfig = req.Config.PasskeyConfig
	} else if req.Config.EnablePasskey != existing.EnablePasskey {
		existing.EnablePasskey = req.Config.EnablePasskey
	}

	// Update tenant
	if _, err := s.TenantService.UpdateTenant(ctx, &domain.UpdateTenantRequest{
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
// @Route "GET /apps/{app_id}"
func (s *CredentialConfigService) GetAppConfig(ctx *request.Context, req *domain.GetAppCredentialConfigRequest) (*domain.CredentialConfig, error) {
	app, err := s.AppService.GetApp(ctx, &domain.GetAppRequest{
		TenantID: req.TenantID,
		ID:       req.AppID,
	})
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}

	if app.Config == nil || app.Config.Credentials == nil {
		// Return tenant default or global default
		return s.GetTenantConfig(ctx, &domain.GetTenantCredentialConfigRequest{
			TenantID: req.TenantID,
		})
	}

	return app.Config.Credentials, nil
}

// UpdateAppConfig updates app-specific credential config
// @Route "PUT/apps/{app_id}"
func (s *CredentialConfigService) UpdateAppConfig(ctx *request.Context, req *domain.UpdateAppCredentialConfigRequest) (*domain.CredentialConfig, error) {
	app, err := s.AppService.GetApp(ctx, &domain.GetAppRequest{
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

	// Initialize credentials if nil
	if app.Config.Credentials == nil {
		app.Config.Credentials = &domain.CredentialConfig{}
	}

	// MERGE credentials config - smart merge based on what's provided
	existing := app.Config.Credentials

	// Basic auth - update only if config is provided
	if req.Config.BasicConfig != nil {
		existing.EnableBasic = req.Config.EnableBasic
		existing.BasicConfig = req.Config.BasicConfig
	} else if req.Config.EnableBasic != existing.EnableBasic {
		// Only enable flag changed without config
		existing.EnableBasic = req.Config.EnableBasic
	}

	// API Key auth
	if req.Config.APIKeyConfig != nil {
		existing.EnableAPIKey = req.Config.EnableAPIKey
		existing.APIKeyConfig = req.Config.APIKeyConfig
	} else if req.Config.EnableAPIKey != existing.EnableAPIKey {
		existing.EnableAPIKey = req.Config.EnableAPIKey
	}

	// OAuth2 auth
	if req.Config.OAuth2Config != nil {
		existing.EnableOAuth2 = req.Config.EnableOAuth2
		existing.OAuth2Config = req.Config.OAuth2Config
	} else if req.Config.EnableOAuth2 != existing.EnableOAuth2 {
		existing.EnableOAuth2 = req.Config.EnableOAuth2
	}

	// Passwordless auth
	if req.Config.PasswordlessConfig != nil {
		existing.EnablePasswordless = req.Config.EnablePasswordless
		existing.PasswordlessConfig = req.Config.PasswordlessConfig
	} else if req.Config.EnablePasswordless != existing.EnablePasswordless {
		existing.EnablePasswordless = req.Config.EnablePasswordless
	}

	// Passkey auth
	if req.Config.PasskeyConfig != nil {
		existing.EnablePasskey = req.Config.EnablePasskey
		existing.PasskeyConfig = req.Config.PasskeyConfig
	} else if req.Config.EnablePasskey != existing.EnablePasskey {
		existing.EnablePasskey = req.Config.EnablePasskey
	}

	// Update app
	if err := s.AppService.UpdateApp(ctx, &domain.UpdateAppRequest{
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
