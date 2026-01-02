package application

import (
	"fmt"
	"time"

	core_domain "github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/credential/domain"
	"github.com/primadi/lokstra-auth/credential/domain/basic"
	"github.com/primadi/lokstra-auth/credential/infrastructure/hasher"
	"github.com/primadi/lokstra-auth/credential/infrastructure/validator"
	core_repository "github.com/primadi/lokstra-auth/infrastructure/repository"
	token "github.com/primadi/lokstra-auth/token"
	"github.com/primadi/lokstra/core/request"
)

// BasicAuthService handles basic authentication (username/password) via HTTP.
// @RouterService name="basic-auth-service", prefix="${api-auth-prefix:/api/auth}/cred/basic", middlewares=["recovery", "request_logger"]
type BasicAuthService struct {
	// @Inject "basic-authenticator"
	Authenticator domain.Authenticator
	// @Inject "@store.user-store"
	UserStore core_repository.UserStore
	// @Inject "credential-validator"
	Validator validator.CredentialValidator
	// @Inject "@store.tenant-store"
	TenantStore core_repository.TenantStore
	// @Inject "@store.app-store"
	AppStore core_repository.AppStore
	// @Inject "token-manager"
	TokenManager token.TokenManager
}

// Login authenticates user with username/password
// @Route "POST /login"
func (s *BasicAuthService) Login(ctx *request.Context, req *basic.LoginRequest) (*basic.TokenResponse, error) {
	// Check if basic auth is enabled
	if !s.isBasicEnabled(ctx, req.TenantID, req.AppID) {
		return &basic.TokenResponse{
			Success: false,
			Error:   "basic authentication is not enabled for this application",
		}, nil
	}

	// Build auth context
	authCtx := &domain.AuthContext{
		TenantID:  req.TenantID,
		AppID:     req.AppID,
		BranchID:  req.BranchID,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}

	// Build credentials
	creds := &basic.Credentials{
		Username: req.Username,
		Password: req.Password,
	}

	// Authenticate
	result, err := s.Authenticator.Authenticate(ctx, authCtx, creds)
	if err != nil {
		// Generic error - don't expose internal details
		// Log the real error internally for debugging
		// TODO: Add proper logging: log.Error("authentication failed", "error", err)
		return nil, fmt.Errorf("authentication service unavailable")
	}

	if !result.Success {
		// Return 200 OK with success:false for failed login attempts
		// This prevents attackers from distinguishing between:
		// - User doesn't exist
		// - Wrong password
		// - Account locked/suspended
		return &basic.TokenResponse{
			Success: false,
			Error:   "invalid credentials", // Generic message
		}, nil
	}

	// Use claims from authentication result (includes metadata)
	accessToken, err := s.TokenManager.Generate(ctx, result.Claims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.TokenManager.GenerateRefreshToken(ctx, result.Claims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Get user info for response
	user, err := s.UserStore.Get(ctx, result.TenantID, result.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &basic.TokenResponse{
		Success:      true,
		AccessToken:  accessToken.Value,
		RefreshToken: refreshToken.Value,
		TokenType:    accessToken.Type,
		ExpiresIn:    int64(accessToken.ExpiresAt.Sub(accessToken.IssuedAt).Seconds()),
		User:         basic.ToUserInfoFromDomain(user),
	}, nil
}

// ChangePassword changes user password
// @Route "POST /change-password", ["auth"]
func (s *BasicAuthService) ChangePassword(ctx *request.Context, req *basic.ChangePasswordRequest) error {
	// Get user
	user, err := s.UserStore.Get(ctx, req.TenantID, req.UserID)
	if err != nil {
		// Generic error - don't reveal if user exists
		return fmt.Errorf("failed to change password")
	}

	// Check if user has password hash
	if user.PasswordHash == nil {
		// Generic error - don't reveal password state
		return fmt.Errorf("failed to change password")
	}

	// Verify old password (dereference pointer)
	if !hasher.VerifyPassword(*user.PasswordHash, req.OldPassword) {
		// Generic error - don't reveal which field is wrong
		// This prevents attackers from knowing if old password is correct
		return fmt.Errorf("failed to change password")
	}

	// Validate new password
	if err := s.Validator.ValidatePassword(req.NewPassword); err != nil {
		// Only for new password validation, we can be specific to help user
		return fmt.Errorf("invalid new password: %w", err)
	}

	// Hash new password
	newHash, err := hasher.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password using UserStore.SetPassword
	if err := s.UserStore.SetPassword(ctx, req.TenantID, req.UserID, newHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ForgotPassword initiates password reset flow by sending reset token via email
// @Route "POST /forgot-password"
func (s *BasicAuthService) ForgotPassword(ctx *request.Context, req *basic.ForgotPasswordRequest) (*basic.ForgotPasswordResponse, error) {
	// Get user by email
	user, err := s.UserStore.GetByEmail(ctx, req.TenantID, req.Email)
	if err != nil {
		// Don't reveal if user exists - always return success for security
		return &basic.ForgotPasswordResponse{
			Success: true,
			Message: "If the email exists, a password reset link has been sent",
		}, nil
	}

	// Check if user is active
	if user.Status != core_domain.UserStatusActive {
		// Don't reveal user status - return generic message
		return &basic.ForgotPasswordResponse{
			Success: true,
			Message: "If the email exists, a password reset link has been sent",
		}, nil
	}

	// Generate reset token (short-lived, e.g., 15 minutes)
	resetClaims := token.Claims{
		"sub":        user.ID,
		"tenant_id":  req.TenantID,
		"app_id":     req.AppID,
		"email":      user.Email,
		"type":       "password_reset",
		"expires_at": time.Now().Add(15 * time.Minute).Unix(),
	}

	resetToken, err := s.TokenManager.Generate(ctx, resetClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}

	// TODO: Send email with reset link
	// Email should contain: https://your-app.com/reset-password?token={resetToken.Value}
	// For now, return token in response (in production, only send via email)
	_ = resetToken // Placeholder until email service is implemented

	return &basic.ForgotPasswordResponse{
		Success: true,
		Message: "If the email exists, a password reset link has been sent",
	}, nil
}

// ResetPassword resets password using token from forgot password email
// @Route "POST /reset-password"
func (s *BasicAuthService) ResetPassword(ctx *request.Context, req *basic.ResetPasswordRequest) (*basic.ResetPasswordResponse, error) {
	// Verify reset token
	result, err := s.TokenManager.Verify(ctx, req.ResetToken)
	if err != nil {
		return &basic.ResetPasswordResponse{
			Success: false,
			Error:   "Invalid or expired reset token",
		}, nil
	}

	if !result.Valid {
		return &basic.ResetPasswordResponse{
			Success: false,
			Error:   "Invalid or expired reset token",
		}, nil
	}

	// Verify token type
	tokenType, ok := result.Claims["type"].(string)
	if !ok || tokenType != "password_reset" {
		return &basic.ResetPasswordResponse{
			Success: false,
			Error:   "Invalid token type",
		}, nil
	}

	// Extract user ID from token
	userID, ok := result.Claims["sub"].(string)
	if !ok {
		return &basic.ResetPasswordResponse{
			Success: false,
			Error:   "Invalid token claims",
		}, nil
	}

	tenantID, ok := result.Claims["tenant_id"].(string)
	if !ok || tenantID != req.TenantID {
		return &basic.ResetPasswordResponse{
			Success: false,
			Error:   "Invalid token claims",
		}, nil
	}

	// Validate new password
	if err := s.Validator.ValidatePassword(req.NewPassword); err != nil {
		return &basic.ResetPasswordResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid password: %v", err),
		}, nil
	}

	// Hash new password
	newHash, err := hasher.HashPassword(req.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.UserStore.SetPassword(ctx, tenantID, userID, newHash); err != nil {
		return nil, fmt.Errorf("failed to reset password: %w", err)
	}

	// Revoke the reset token to prevent reuse
	if err := s.TokenManager.Revoke(ctx, req.ResetToken); err != nil {
		// Log but don't fail - password already reset
	}

	return &basic.ResetPasswordResponse{
		Success: true,
		Message: "Password reset successful. You can now login with your new password.",
	}, nil
}

// Helper methods for config resolution

func (s *BasicAuthService) getEffectiveConfig(ctx *request.Context, tenantID, appID string) *core_domain.CredentialConfig {
	// Try to get app config
	if appID != "" {
		app, err := s.AppStore.Get(ctx, tenantID, appID)
		if err == nil && app.Config != nil && app.Config.Credentials != nil {
			return app.Config.Credentials
		}
	}

	// Try to get tenant default config
	tenant, err := s.TenantStore.Get(ctx, tenantID)
	if err == nil && tenant.Config != nil && tenant.Config.DefaultCredentials != nil {
		return tenant.Config.DefaultCredentials
	}

	// Return global default
	return core_domain.DefaultCredentialConfig()
}

func (s *BasicAuthService) isBasicEnabled(ctx *request.Context, tenantID, appID string) bool {
	config := s.getEffectiveConfig(ctx, tenantID, appID)
	return config.EnableBasic
}
