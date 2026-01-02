package authenticator

import (
	"strings"

	core_domain "github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/credential/domain"
	"github.com/primadi/lokstra-auth/credential/domain/basic"
	"github.com/primadi/lokstra-auth/credential/infrastructure/hasher"
	core_repository "github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
)

// BasicAuthenticator authenticates basic credentials
// @Service "basic-authenticator"
type BasicAuthenticator struct {
	// @Inject "@store.user-store"
	userStore core_repository.UserStore
	// @Inject "@store.tenant-store"
	tenantStore core_repository.TenantStore
}

var _ domain.Authenticator = (*BasicAuthenticator)(nil)

// NewBasicAuthenticator creates a new basic authenticator
// func NewBasicAuthenticator(userStore core_repository.UserStore, tenantStore core_repository.TenantStore) *BasicAuthenticator {
// 	return &BasicAuthenticator{
// 		userStore:   userStore,
// 		tenantStore: tenantStore,
// 	}
// }

// Authenticate verifies the provided credentials within the authentication context
func (a *BasicAuthenticator) Authenticate(ctx *request.Context,
	authCtx *domain.AuthContext, creds domain.Credentials) (*domain.AuthenticationResult, error) {
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
	// Support login with username OR email
	var user *core_domain.User
	var err error

	// Check if input looks like an email (contains @)
	if strings.Contains(basicCreds.Username, "@") {
		user, err = a.userStore.GetByEmail(ctx, authCtx.TenantID, basicCreds.Username)
	} else {
		user, err = a.userStore.GetByUsername(ctx, authCtx.TenantID, basicCreds.Username)
	}

	if err != nil {
		// Any database error (including "no rows") should be treated as authentication failure
		// Don't expose whether user exists or not
		// Don't expose internal database errors
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    basic.ErrAuthenticationFailed,
		}, nil
	}

	// Verify tenant matches (redundant check but good for security)
	if user.TenantID != authCtx.TenantID {
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    basic.ErrAuthenticationFailed,
		}, nil
	}

	// Check if account is locked
	if user.IsLocked() {
		// Auto-unlock if lock period expired
		if user.ShouldAutoUnlock() {
			user.UnlockAccount()
			if err := a.userStore.Update(ctx, user); err != nil {
				// Log error but continue - return generic failure
				return &domain.AuthenticationResult{
					Success:  false,
					TenantID: authCtx.TenantID,
					AppID:    authCtx.AppID,
					Error:    basic.ErrAuthenticationFailed,
				}, nil
			}
		} else {
			// Account still locked - generic error (don't reveal lock status)
			return &domain.AuthenticationResult{
				Success:  false,
				TenantID: authCtx.TenantID,
				AppID:    authCtx.AppID,
				Error:    basic.ErrAuthenticationFailed,
			}, nil
		}
	}

	// Check if user is disabled/suspended/deleted
	if user.Status != core_domain.UserStatusActive {
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    basic.ErrAuthenticationFailed,
		}, nil
	}

	// Check if user has password hash (basic auth enabled)
	if user.PasswordHash == nil {
		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    basic.ErrAuthenticationFailed,
		}, nil
	}

	// Verify password (dereference pointer)
	if !hasher.VerifyPassword(*user.PasswordHash, basicCreds.Password) {
		// Password verification failed - record failed attempt
		user.RecordFailedLogin()

		// Get tenant lockout policy
		tenant, err := a.tenantStore.Get(ctx, authCtx.TenantID)
		if err == nil && tenant != nil && tenant.Settings != nil {
			lockoutPolicy := tenant.Settings.AccountLockout

			// Check if lockout is enabled and max attempts reached
			if lockoutPolicy.Enabled && user.FailedLoginAttempts >= lockoutPolicy.MaxAttempts {
				// Lock the account
				user.LockAccount(lockoutPolicy.LockoutDuration)
				// TODO: Send lockout notification email if lockoutPolicy.NotifyOnLockout == true
			}
		}

		// Update user with failed attempt counter
		_ = a.userStore.Update(ctx, user)

		return &domain.AuthenticationResult{
			Success:  false,
			TenantID: authCtx.TenantID,
			AppID:    authCtx.AppID,
			Error:    basic.ErrAuthenticationFailed,
		}, nil
	}

	// Password verified successfully!
	// Reset failed login attempts after successful authentication
	if user.FailedLoginAttempts > 0 {
		user.ResetFailedLoginAttempts()
		_ = a.userStore.Update(ctx, user)
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

	// Add user metadata to claims if present
	if user.Metadata != nil {
		claims["metadata"] = *user.Metadata
	}

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
