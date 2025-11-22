package basic

// =============================================================================
// Request DTOs
// =============================================================================

// LoginRequest request to login with username/password
type LoginRequest struct {
	TenantID  string `json:"tenant_id" validate:"required"`
	AppID     string `json:"app_id" validate:"required"`
	BranchID  string `json:"branch_id,omitempty"`
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// RegisterRequest request to register a new user
type RegisterRequest struct {
	TenantID string         `json:"tenant_id" validate:"required"`
	AppID    string         `json:"app_id" validate:"required"`
	Username string         `json:"username" validate:"required"`
	Password string         `json:"password" validate:"required"`
	Email    string         `json:"email" validate:"required,email"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ChangePasswordRequest request to change user password
type ChangePasswordRequest struct {
	TenantID    string `json:"tenant_id" validate:"required"`
	UserID      string `json:"user_id" validate:"required"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

// =============================================================================
// Response DTOs
// =============================================================================

// LoginResponse response after successful login
type LoginResponse struct {
	Success     bool      `json:"success"`
	AccessToken string    `json:"access_token,omitempty"`
	TokenType   string    `json:"token_type,omitempty"` // "Bearer"
	ExpiresIn   int64     `json:"expires_in,omitempty"` // seconds
	User        *UserInfo `json:"user,omitempty"`
	Error       string    `json:"error,omitempty"`
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

// ToUserInfo converts User to sanitized UserInfo
func ToUserInfo(user *User) *UserInfo {
	if user == nil {
		return nil
	}
	return &UserInfo{
		ID:       user.ID,
		TenantID: user.TenantID,
		Username: user.Username,
		Email:    user.Email,
		Disabled: user.Disabled,
		Metadata: user.Metadata,
	}
}

// RegisterResponse response after successful registration
type RegisterResponse struct {
	Success bool      `json:"success"`
	User    *UserInfo `json:"user,omitempty"`
	Error   string    `json:"error,omitempty"`
}
