package oauth2

// =============================================================================
// Request DTOs
// =============================================================================

// AuthorizeRequest request to initiate OAuth2 flow
type AuthorizeRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	Provider string `json:"provider" validate:"required"` // "google", "azure", etc.
}

// AuthorizeResponse response containing OAuth2 authorization URL
type AuthorizeResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"` // CSRF token
}

// CallbackRequest request from OAuth2 callback
type CallbackRequest struct {
	Code  string `query:"code" validate:"required"`
	State string `query:"state" validate:"required"`
}

// =============================================================================
// Response DTOs
// =============================================================================

// LoginResponse response after successful OAuth2 login
type LoginResponse struct {
	Success     bool      `json:"success"`
	AccessToken string    `json:"access_token,omitempty"`
	TokenType   string    `json:"token_type,omitempty"`
	ExpiresIn   int64     `json:"expires_in,omitempty"`
	User        *UserInfo `json:"user,omitempty"`
	Error       string    `json:"error,omitempty"`
}
