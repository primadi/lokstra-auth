package passwordless

import "time"

// Credentials represents passwordless authentication credentials
type Credentials struct {
	// Delivery method: "email" or "sms"
	Method string `json:"method"`

	// Email or phone number
	Identifier string `json:"identifier"`

	// Code or token sent to user
	Code string `json:"code,omitempty"`

	// Magic link token (alternative to code)
	Token string `json:"token,omitempty"`
}

// User represents a user in passwordless authentication
type User struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Email       string    `json:"email,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Verified    bool      `json:"verified"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// =============================================================================
// DTOs
// =============================================================================

// SendCodeRequest request to send verification code
type SendCodeRequest struct {
	TenantID   string `header:"X-Tenant-ID" validate:"required"`
	AppID      string `header:"X-App-ID" validate:"required"`
	Method     string `json:"method" validate:"required,oneof=email sms"` // email or sms
	Identifier string `json:"identifier" validate:"required"`             // email or phone
	IPAddress  string `json:"ip_address,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
}

// SendCodeResponse response after sending code
type SendCodeResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"` // For tracking delivery
	ExpiresIn int    `json:"expires_in"`           // Seconds until code expires
	Error     string `json:"error,omitempty"`
}

// VerifyCodeRequest request to verify code
type VerifyCodeRequest struct {
	TenantID   string `header:"X-Tenant-ID" validate:"required"`
	AppID      string `header:"X-App-ID" validate:"required"`
	Method     string `json:"method" validate:"required"`
	Identifier string `json:"identifier" validate:"required"`
	Code       string `json:"code" validate:"required"`
	IPAddress  string `json:"ip_address,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
}

// VerifyCodeResponse response after verifying code
type VerifyCodeResponse struct {
	Success      bool   `json:"success"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	Error        string `json:"error,omitempty"`
}

// SendMagicLinkRequest request to send magic link
type SendMagicLinkRequest struct {
	TenantID   string `header:"X-Tenant-ID" validate:"required"`
	AppID      string `header:"X-App-ID" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	RedirectTo string `json:"redirect_to,omitempty"` // URL to redirect after verification
	IPAddress  string `json:"ip_address,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
}

// SendMagicLinkResponse response after sending magic link
type SendMagicLinkResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	ExpiresIn int    `json:"expires_in"`
	Error     string `json:"error,omitempty"`
}

// VerifyMagicLinkRequest request to verify magic link token
type VerifyMagicLinkRequest struct {
	Token     string `path:"token" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// VerifyMagicLinkResponse response after verifying magic link
type VerifyMagicLinkResponse struct {
	Success      bool   `json:"success"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	RedirectTo   string `json:"redirect_to,omitempty"`
	Error        string `json:"error,omitempty"`
}
