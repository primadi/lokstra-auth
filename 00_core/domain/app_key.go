package domain

import "time"

// =============================================================================
// AppKey Domain Model
// =============================================================================

// AppKey represents an API key for application authentication
type AppKey struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"tenant_id"`
	AppID       string         `json:"app_id"`
	KeyID       string         `json:"key_id"`      // Public identifier
	Prefix      string         `json:"prefix"`      // Key prefix
	SecretHash  string         `json:"secret_hash"` // SHA3-256 hashed secret
	KeyType     string         `json:"key_type"`    // "secret", "public"
	Environment string         `json:"environment"` // "live", "test"
	UserID      string         `json:"user_id"`     // Owner user ID
	Name        string         `json:"name"`        // Descriptive name
	Scopes      []string       `json:"scopes"`      // Allowed scopes
	Metadata    map[string]any `json:"metadata"`    // Additional data
	CreatedAt   time.Time      `json:"created_at"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	LastUsed    *time.Time     `json:"last_used,omitempty"`
	Revoked     bool           `json:"revoked"`
	RevokedAt   *time.Time     `json:"revoked_at,omitempty"`
}
