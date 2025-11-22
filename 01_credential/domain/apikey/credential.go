package apikey

import (
	"errors"
	"strings"

	"github.com/primadi/lokstra-auth/01_credential/domain"
)

var (
	ErrEmptyAPIKey         = errors.New("api key cannot be empty")
	ErrInvalidAPIKeyFormat = errors.New("invalid api key format")
	ErrInvalidCredentials  = errors.New("invalid credentials type")
)

// Credentials represents API key credentials
type Credentials struct {
	APIKey string // Format: prefix_keyid.secret
}

var _ domain.Credentials = (*Credentials)(nil)

// Type returns the credential type
func (c *Credentials) Type() string {
	return "apikey"
}

// Validate checks if the credentials are well-formed
func (c *Credentials) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return ErrEmptyAPIKey
	}

	// Basic format validation: should contain underscore and dot
	if !strings.Contains(c.APIKey, "_") || !strings.Contains(c.APIKey, ".") {
		return ErrInvalidAPIKeyFormat
	}

	return nil
}

// ParseAPIKey parses API key string into components
// Format: {prefix}_{keyid}.{secret}
func (c *Credentials) ParseAPIKey() (prefix, keyID, secret string, err error) {
	parts := strings.SplitN(c.APIKey, ".", 2)
	if len(parts) != 2 {
		return "", "", "", ErrInvalidAPIKeyFormat
	}

	secret = parts[1]
	prefixAndKeyID := parts[0]

	lastUnderscore := strings.LastIndex(prefixAndKeyID, "_")
	if lastUnderscore == -1 {
		return "", "", "", ErrInvalidAPIKeyFormat
	}

	prefix = prefixAndKeyID[:lastUnderscore]
	keyID = prefixAndKeyID[lastUnderscore+1:]

	return prefix, keyID, secret, nil
}

// APIKey represents a stored API key
type APIKey struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"tenant_id"`
	AppID       string         `json:"app_id"`
	KeyID       string         `json:"key_id"`       // Public identifier
	Prefix      string         `json:"prefix"`       // Key prefix for identification
	SecretHash  string         `json:"secret_hash"`  // SHA3-256 hash of secret
	Name        string         `json:"name"`         // Descriptive name
	Environment string         `json:"environment"`  // "live", "test"
	Scopes      []string       `json:"scopes"`       // Allowed scopes
	Metadata    map[string]any `json:"metadata"`     // Additional metadata
	ExpiresAt   *string        `json:"expires_at"`   // ISO8601 timestamp
	RevokedAt   *string        `json:"revoked_at"`   // ISO8601 timestamp
	LastUsedAt  *string        `json:"last_used_at"` // ISO8601 timestamp
	CreatedAt   string         `json:"created_at"`   // ISO8601 timestamp
}
