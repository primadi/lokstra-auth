package authenticator

import (
	"context"
	"errors"
	"maps"

	"github.com/primadi/lokstra-auth/01_credential/domain"
	"github.com/primadi/lokstra-auth/01_credential/domain/basic"
	"github.com/primadi/lokstra-auth/01_credential/infrastructure/hasher"
	"github.com/primadi/lokstra-auth/01_credential/infrastructure/repository"
)

// BasicAuthenticator authenticates basic credentials
type BasicAuthenticator struct {
	userProvider repository.UserProvider
}

var _ domain.Authenticator = (*BasicAuthenticator)(nil)

// NewBasicAuthenticator creates a new basic authenticator
func NewBasicAuthenticator(userProvider repository.UserProvider) *BasicAuthenticator {
	return &BasicAuthenticator{
		userProvider: userProvider,
	}
}

// Authenticate verifies the provided credentials within the authentication context
func (a *BasicAuthenticator) Authenticate(ctx context.Context, authCtx *domain.AuthContext, creds domain.Credentials) (*domain.AuthenticationResult, error) {
	// Validate authentication context
	if err := authCtx.Validate(); err != nil {
		return &domain.AuthenticationResult{
			Success: false,
			Error:   err,
		}, nil
	}

	basicCreds, ok := creds.(*basic.Credentials)
	if !ok {
		return &domain.AuthenticationResult{
			Success: false,
			Error:   basic.ErrInvalidCredentials,
		}, nil
	}

	// Only validate that credentials are not empty during login
	// Do NOT validate complexity/patterns - that's for registration only
	if err := basicCreds.Validate(); err != nil {
		return &domain.AuthenticationResult{
			Success: false,
			Error:   err,
		}, nil
	}

	// Retrieve user within tenant scope
	user, err := a.userProvider.GetUserByUsername(ctx, authCtx.TenantID, basicCreds.Username)
	if err != nil {
		if errors.Is(err, basic.ErrUserNotFound) {
			return &domain.AuthenticationResult{
				Success:  false,
				TenantID: authCtx.TenantID,
				AppID:    authCtx.AppID,
				Error:    basic.ErrAuthenticationFailed,
			}, nil
		}
		return nil, err
	}

	// Verify tenant matches
	if user.TenantID != authCtx.TenantID {
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    basic.ErrAuthenticationFailed,
		}, nil
	}

	// Check if user is disabled
	if user.Disabled {
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    basic.ErrAuthenticationFailed,
		}, nil
	}

	// Verify password
	if !hasher.VerifyPassword(user.PasswordHash, basicCreds.Password) {
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    basic.ErrAuthenticationFailed,
		}, nil
	}

	// Build claims with user metadata
	claims := map[string]any{
		"sub":       user.ID,
		"tenant_id": user.TenantID,
		"app_id":    authCtx.AppID,
		"username":  user.Username,
		"email":     user.Email,
		"auth_type": "basic",
	}

	// Add user metadata to claims
	maps.Copy(claims, user.Metadata)

	return &domain.AuthenticationResult{
		Success:  true,
		Subject:  user.ID,
		TenantID: authCtx.TenantID,
		AppID:    authCtx.AppID,
		Claims:   claims,
	}, nil
}

// Type returns the type of authenticator
func (a *BasicAuthenticator) Type() string {
	return "basic"
}
