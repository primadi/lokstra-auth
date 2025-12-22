package authenticator

import (
	"errors"
	"time"

	"github.com/primadi/lokstra-auth/credential/domain"
	"github.com/primadi/lokstra-auth/credential/domain/apikey"
	"github.com/primadi/lokstra-auth/credential/infrastructure/hasher"
	core_repository "github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
)

var (
	ErrAPIKeyNotFound = errors.New("api key not found")
	ErrAPIKeyExpired  = errors.New("api key has expired")
	ErrAPIKeyRevoked  = errors.New("api key has been revoked")
)

// APIKeyAuthenticator authenticates API key credentials
// @Service "apikey-authenticator"
type APIKeyAuthenticator struct {
	// @Inject "app-key-store"
	keyStore core_repository.AppKeyStore
}

var _ domain.Authenticator = (*APIKeyAuthenticator)(nil)

// NewAPIKeyAuthenticator creates a new API key authenticator
// func NewAPIKeyAuthenticator(keyStore core_repository.AppKeyStore) *APIKeyAuthenticator {
// 	return &APIKeyAuthenticator{
// 		keyStore: keyStore,
// 	}
// }

// Authenticate verifies the provided API key credentials
func (a *APIKeyAuthenticator) Authenticate(ctx *request.Context, authCtx *domain.AuthContext, creds domain.Credentials) (*domain.AuthenticationResult, error) {
	// Validate authentication context
	if err := authCtx.Validate(); err != nil {
		return &domain.AuthenticationResult{
			Success: false,
			Error:   err,
		}, nil
	}

	apikeyCreds, ok := creds.(*apikey.Credentials)
	if !ok {
		return &domain.AuthenticationResult{
			Success: false,
			Error:   apikey.ErrInvalidCredentials,
		}, nil
	}

	// Validate format
	if err := apikeyCreds.Validate(); err != nil {
		return &domain.AuthenticationResult{
			Success: false,
			Error:   err,
		}, nil
	}

	// Parse API key
	prefix, keyID, secret, err := apikeyCreds.ParseAPIKey()
	if err != nil {
		return &domain.AuthenticationResult{
			Success: false,
			Error:   err,
		}, nil
	}

	// Retrieve key from store
	key, err := a.keyStore.GetByKeyID(ctx, authCtx.TenantID, authCtx.AppID, keyID)
	if err != nil {
		if errors.Is(err, ErrAPIKeyNotFound) {
			return &domain.AuthenticationResult{
				Success:  false,
				TenantID: authCtx.TenantID,
				AppID:    authCtx.AppID,
				Error:    domain.ErrAuthenticationFailed,
			}, nil
		}
		return nil, err
	}

	// Verify prefix matches
	if key.Prefix != prefix {
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    domain.ErrAuthenticationFailed,
		}, nil
	}

	// Check if revoked
	if key.Revoked {
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    ErrAPIKeyRevoked,
		}, nil
	}

	// Check if expired
	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    ErrAPIKeyExpired,
		}, nil
	}

	// Verify secret hash
	isValid, err := hasher.VerifySecretHash(secret, key.SecretHash)
	if err != nil || !isValid {
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    domain.ErrAuthenticationFailed,
		}, nil
	}

	// Update last used timestamp (using Update method from AppKeyStore)
	now := time.Now()
	key.LastUsed = &now
	_ = a.keyStore.Update(ctx, key)

	// Build claims
	claims := map[string]any{
		"sub":         key.ID,
		"tenant_id":   key.TenantID,
		"app_id":      key.AppID,
		"key_id":      key.KeyID,
		"scopes":      key.Scopes,
		"environment": key.Environment,
		"auth_type":   "apikey",
	}

	// Add metadata to claims
	for k, v := range key.Metadata {
		claims[k] = v
	}

	return &domain.AuthenticationResult{
		Success:  true,
		Subject:  key.ID,
		TenantID: authCtx.TenantID,
		AppID:    authCtx.AppID,
		Claims:   claims,
	}, nil
}

// Type returns the type of authenticator
func (a *APIKeyAuthenticator) Type() string {
	return "apikey"
}
