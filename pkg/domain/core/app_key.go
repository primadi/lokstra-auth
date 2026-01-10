package core

import "time"

// =============================================================================
// AppKey Domain Model
// =============================================================================

// AppKey represents an API key for application authentication
type AppKey struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	AppID       string          `json:"app_id"`
	BranchID    *string         `json:"branch_id,omitempty"`
	KeyID       string          `json:"key_id"`      // Public identifier
	Prefix      string          `json:"prefix"`      // Key prefix
	SecretHash  string          `json:"secret_hash"` // SHA3-256 hashed secret
	KeyType     string          `json:"key_type"`    // "secret", "public"
	Environment string          `json:"environment"` // "live", "test"
	Name        string          `json:"name"`        // Descriptive name
	Description *string         `json:"description,omitempty"`
	Status      string          `json:"status"` // "active", "revoked"
	Scopes      []string        `json:"scopes"` // Allowed scopes
	ExpiresAt   *time.Time      `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time      `json:"last_used_at,omitempty"`
	CreatedBy   string          `json:"created_by"`
	Metadata    *map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
}
