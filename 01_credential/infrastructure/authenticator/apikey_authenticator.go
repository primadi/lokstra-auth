package authenticator

import (
	"context"
	"errors"
	"time"

	"github.com/primadi/lokstra-auth/01_credential/domain"
	"github.com/primadi/lokstra-auth/01_credential/domain/apikey"
	"github.com/primadi/lokstra-auth/01_credential/infrastructure/hasher"
	"github.com/primadi/lokstra-auth/01_credential/infrastructure/repository"
)

var (
	ErrAPIKeyNotFound = errors.New("api key not found")
	ErrAPIKeyExpired  = errors.New("api key has expired")
	ErrAPIKeyRevoked  = errors.New("api key has been revoked")
)

// APIKeyAuthenticator authenticates API key credentials
type APIKeyAuthenticator struct {
	keyStore repository.APIKeyStore
}

var _ domain.Authenticator = (*APIKeyAuthenticator)(nil)

// NewAPIKeyAuthenticator creates a new API key authenticator
func NewAPIKeyAuthenticator(keyStore repository.APIKeyStore) *APIKeyAuthenticator {
	return &APIKeyAuthenticator{
		keyStore: keyStore,
	}
}

// Authenticate verifies the provided API key credentials
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, authCtx *domain.AuthContext, creds domain.Credentials) (*domain.AuthenticationResult, error) {
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
	if key.RevokedAt != nil {
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    ErrAPIKeyRevoked,
		}, nil
	}

	// Check if expired
	if key.ExpiresAt != nil {
		expiresAt, err := time.Parse(time.RFC3339, *key.ExpiresAt)
		if err == nil && time.Now().After(expiresAt) {
			return &domain.AuthenticationResult{
				Success:  false,
				TenantID: authCtx.TenantID,
				AppID:    authCtx.AppID,
				Error:    ErrAPIKeyExpired,
			}, nil
		}
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

	// Update last used timestamp
	now := time.Now().Format(time.RFC3339)
	key.LastUsedAt = &now
	_ = a.keyStore.UpdateLastUsed(ctx, key.ID, now)

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
