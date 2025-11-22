package application

import (
	"fmt"

	"github.com/primadi/lokstra-auth/01_credential/domain"
	"github.com/primadi/lokstra-auth/01_credential/domain/apikey"
	"github.com/primadi/lokstra-auth/01_credential/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

// APIKeyAuthService handles API key authentication via HTTP
// @RouterService name="apikey-auth-service", prefix="/api/cred/apikey", middlewares=["recovery", "request-logger"]
type APIKeyAuthService struct {
	// @Inject "apikey-authenticator"
	Authenticator *service.Cached[domain.Authenticator]
	// @Inject "apikey-store"
	Store *service.Cached[repository.APIKeyStore]
	// @Inject "credential-config-resolver"
	ConfigResolver *service.Cached[*ConfigResolver]
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
	result, err := s.Authenticator.MustGet().Authenticate(ctx, authCtx, creds)
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
	if !s.ConfigResolver.MustGet().IsAPIKeyEnabled(ctx, result.TenantID, result.AppID) {
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
