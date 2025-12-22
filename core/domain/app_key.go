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

// =============================================================================
// AppKey DTOs
// =============================================================================

// GenerateAppKeyRequest request to generate a new app key
type GenerateAppKeyRequest struct {
	TenantID    string         `path:"tenant_id" validate:"required"`
	AppID       string         `path:"app_id" validate:"required"`
	Name        string         `json:"name" validate:"required"`
	Purpose     string         `json:"purpose"`
	Description string         `json:"description"`
	Environment string         `json:"environment"` // live, test
	Scopes      []string       `json:"scopes"`
	ExpiresIn   *time.Duration `json:"expires_in"` // nil = never expires
}

// AppKeyResponse response containing generated app key (with secret - shown only once!)
type AppKeyResponse struct {
	KeyID     string  `json:"key_id"`
	KeyString string  `json:"key_string"` // Full key with secret - ONLY shown once!
	AppKey    *AppKey `json:"app_key"`
}

// AppKeyInfo sanitized response without secret hash (for Get/List operations)
type AppKeyInfo struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"tenant_id"`
	AppID       string         `json:"app_id"`
	KeyID       string         `json:"key_id"` // Public identifier only
	Prefix      string         `json:"prefix"`
	KeyType     string         `json:"key_type"`
	Environment string         `json:"environment"`
	UserID      string         `json:"user_id"`
	Name        string         `json:"name"`
	Scopes      []string       `json:"scopes"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	LastUsed    *time.Time     `json:"last_used,omitempty"`
	Revoked     bool           `json:"revoked"`
	RevokedAt   *time.Time     `json:"revoked_at,omitempty"`
	// ❌ NO SecretHash - never exposed!
}

// ToAppKeyInfo converts AppKey to sanitized AppKeyInfo
func ToAppKeyInfo(key *AppKey) *AppKeyInfo {
	if key == nil {
		return nil
	}
	return &AppKeyInfo{
		ID:          key.ID,
		TenantID:    key.TenantID,
		AppID:       key.AppID,
		KeyID:       key.KeyID,
		Prefix:      key.Prefix,
		KeyType:     key.KeyType,
		Environment: key.Environment,
		UserID:      key.UserID,
		Name:        key.Name,
		Scopes:      key.Scopes,
		Metadata:    key.Metadata,
		CreatedAt:   key.CreatedAt,
		ExpiresAt:   key.ExpiresAt,
		LastUsed:    key.LastUsed,
		Revoked:     key.Revoked,
		RevokedAt:   key.RevokedAt,
	}
}

// ToAppKeyInfoList converts slice of AppKey to slice of AppKeyInfo
func ToAppKeyInfoList(keys []*AppKey) []*AppKeyInfo {
	result := make([]*AppKeyInfo, len(keys))
	for i, key := range keys {
		result[i] = ToAppKeyInfo(key)
	}
	return result
}

// GetAppKeyRequest request to get an app key
type GetAppKeyRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	KeyID    string `path:"key_id" validate:"required"`
}

// ListAppKeysRequest request to list app keys
type ListAppKeysRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
}

// RevokeAppKeyRequest request to revoke an app key
type RevokeAppKeyRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	KeyID    string `path:"key_id" validate:"required"`
}

// DeleteAppKeyRequest request to delete an app key
type DeleteAppKeyRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	KeyID    string `path:"key_id" validate:"required"`
}

// RotateAppKeyRequest request to rotate an app key
type RotateAppKeyRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	KeyID    string `path:"key_id" validate:"required"`
}
