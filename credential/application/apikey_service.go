package application

import (
	"fmt"

	core_domain "github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/credential/domain"
	"github.com/primadi/lokstra-auth/credential/domain/apikey"
	core_repository "github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
)

// APIKeyAuthService handles API key authentication via HTTP
// @RouterService name="apikey-auth-service", prefix="${api-auth-prefix:/api/auth}/cred/apikey", middlewares=["recovery", "request_logger"]
type APIKeyAuthService struct {
	// @Inject "apikey-authenticator"
	Authenticator domain.Authenticator
	// @Inject "@store.app-key-store"
	Store core_repository.AppKeyStore
	// @Inject "@store.tenant-store"
	TenantStore core_repository.TenantStore
	// @Inject "@store.app-store"
	AppStore core_repository.AppStore
}

// Authenticate validates an API key
// @Route "POST /authenticate"
func (s *APIKeyAuthService) Authenticate(ctx *request.Context, req *apikey.AuthenticateRequest) (*apikey.AuthenticateResponse, error) {
	// For API key auth, we need to extract tenant/app from the key first
	// Parse the key to get keyID, then look it up
	creds := &apikey.Credentials{
		APIKey: req.APIKey,
	}

	_, keyID, _, err := creds.ParseAPIKey()
	if err != nil {
		return &apikey.AuthenticateResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Try to find the key (we'll need to search across all tenants/apps)
	// For now, this is a simplified implementation
	// In production, you'd want to index by keyID for faster lookup

	// Build auth context (for API keys, tenant and app come from the key itself)
	// We'll use placeholder values and let the authenticator fill them from the key
	authCtx := &domain.AuthContext{
		TenantID:  "placeholder", // Will be filled from key
		AppID:     "placeholder", // Will be filled from key
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}

	// Authenticate
	result, err := s.Authenticator.Authenticate(ctx, authCtx, creds)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}

	if !result.Success {
		return &apikey.AuthenticateResponse{
			Success:   false,
			Validated: false,
			Error:     result.Error.Error(),
		}, nil
	}

	// Check if API key auth is enabled for this tenant/app
	// (Do this AFTER authentication so we know which tenant/app to check)
	if !s.isAPIKeyEnabled(ctx, result.TenantID, result.AppID) {
		return &apikey.AuthenticateResponse{
			Success: false,
			Error:   "API key authentication is not enabled for this application",
		}, nil
	}

	return &apikey.AuthenticateResponse{
		Success:   true,
		Validated: true,
		KeyID:     keyID,
		TenantID:  result.TenantID,
		AppID:     result.AppID,
		Scopes:    result.Claims["scopes"].([]string),
		Claims:    result.Claims,
	}, nil
}

// Helper methods for config resolution

func (s *APIKeyAuthService) getEffectiveConfig(ctx *request.Context, tenantID, appID string) *core_domain.CredentialConfig {
	// Check app config first
	if appID != "" {
		app, err := s.AppStore.Get(ctx, tenantID, appID)
		if err == nil && app.Config != nil && app.Config.Credentials != nil {
			return app.Config.Credentials
		}
	}

	// Then tenant config
	tenant, err := s.TenantStore.Get(ctx, tenantID)
	if err == nil && tenant.Config != nil && tenant.Config.DefaultCredentials != nil {
		return tenant.Config.DefaultCredentials
	}

	// Finally global default
	return core_domain.DefaultCredentialConfig()
}

func (s *APIKeyAuthService) isAPIKeyEnabled(ctx *request.Context, tenantID, appID string) bool {
	config := s.getEffectiveConfig(ctx, tenantID, appID)
	return config != nil && config.EnableAPIKey && config.APIKeyConfig != nil
}
