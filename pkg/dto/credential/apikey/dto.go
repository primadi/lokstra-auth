package apikey

// =============================================================================
// Request DTOs
// =============================================================================

// AuthenticateRequest request to authenticate with API key
type AuthenticateRequest struct {
	APIKey    string   `json:"api_key" validate:"required"`
	Scopes    []string `json:"scopes,omitempty"` // Requested scopes
	IPAddress string   `json:"ip_address,omitempty"`
	UserAgent string   `json:"user_agent,omitempty"`
}

// =============================================================================
// Response DTOs
// =============================================================================

// AuthenticateResponse response after successful API key authentication
type AuthenticateResponse struct {
	Success   bool           `json:"success"`
	Validated bool           `json:"validated"`
	KeyID     string         `json:"key_id,omitempty"`
	TenantID  string         `json:"tenant_id,omitempty"`
	AppID     string         `json:"app_id,omitempty"`
	Scopes    []string       `json:"scopes,omitempty"`
	Claims    map[string]any `json:"claims,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// APIKeyInfo sanitized API key information (no secret hash)
type APIKeyInfo struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"tenant_id"`
	AppID       string         `json:"app_id"`
	KeyID       string         `json:"key_id"`
	Prefix      string         `json:"prefix"`
	Name        string         `json:"name"`
	Environment string         `json:"environment"`
	Scopes      []string       `json:"scopes"`
	Metadata    map[string]any `json:"metadata"`
	ExpiresAt   *string        `json:"expires_at,omitempty"`
	RevokedAt   *string        `json:"revoked_at,omitempty"`
	LastUsedAt  *string        `json:"last_used_at,omitempty"`
	CreatedAt   string         `json:"created_at"`
	// ❌ NO SecretHash - never exposed!
}
