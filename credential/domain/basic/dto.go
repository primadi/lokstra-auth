package basic

import (
	core_domain "github.com/primadi/lokstra-auth/core/domain"
)

// =============================================================================
// Request DTOs
// =============================================================================

// LoginRequest request to login with username/password
type LoginRequest struct {
	TenantID  string `header:"X-Tenant-ID" validate:"required"`
	AppID     string `header:"X-App-ID" validate:"required"`
	BranchID  string `header:"X-Branch-ID"`
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// ChangePasswordRequest request to change user password
type ChangePasswordRequest struct {
	TenantID    string `header:"X-Tenant-ID" validate:"required"`
	UserID      string `json:"user_id" validate:"required"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

// RefreshRequest request to refresh access token using refresh token
type RefreshRequest struct {
	TenantID     string `header:"X-Tenant-ID" validate:"required"`
	AppID        string `header:"X-App-ID" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest request to logout and revoke refresh token
type LogoutRequest struct {
	TenantID     string `header:"X-Tenant-ID" validate:"required"`
	AppID        string `header:"X-App-ID" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ForgotPasswordRequest request to initiate password reset flow
type ForgotPasswordRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	AppID    string `header:"X-App-ID" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest request to reset password using token from email
type ResetPasswordRequest struct {
	TenantID    string `header:"X-Tenant-ID" validate:"required"`
	AppID       string `header:"X-App-ID" validate:"required"`
	ResetToken  string `json:"reset_token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

// =============================================================================
// Response DTOs
// =============================================================================

// TokenResponse unified response for login and token refresh operations
type TokenResponse struct {
	Success      bool      `json:"success"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"` // "Bearer"
	ExpiresIn    int64     `json:"expires_in,omitempty"` // seconds
	User         *UserInfo `json:"user,omitempty"`       // Present on login, optional on refresh
	Error        string    `json:"error,omitempty"`
}

// UserInfo sanitized user information (no password hash)
type UserInfo struct {
	ID       string         `json:"id"`
	TenantID string         `json:"tenant_id"`
	Username string         `json:"username"`
	Email    string         `json:"email"`
	Disabled bool           `json:"disabled"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ToUserInfoFromDomain converts domain.User (from core) to UserInfo
func ToUserInfoFromDomain(user *core_domain.User) *UserInfo {
	if user == nil {
		return nil
	}
	metadata := make(map[string]any)
	if user.Metadata != nil {
		metadata = *user.Metadata
	}
	disabled := (user.Status == core_domain.UserStatusSuspended || user.Status == core_domain.UserStatusDeleted)
	return &UserInfo{
		ID:       user.ID,
		TenantID: user.TenantID,
		Username: user.Username,
		Email:    user.Email,
		Disabled: disabled,
		Metadata: metadata,
	}
}

// LogoutResponse response after successful logout
type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ForgotPasswordResponse response after initiating password reset
type ForgotPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"` // "Password reset link sent to your email"
	Error   string `json:"error,omitempty"`
}

// ResetPasswordResponse response after successful password reset
type ResetPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"` // "Password reset successful"
	Error   string `json:"error,omitempty"`
}
