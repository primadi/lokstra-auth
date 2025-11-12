package token

import (
	"context"
	"time"
)

// Token represents a security token
type Token struct {
	// Value is the actual token string
	Value string

	// Type is the token type (e.g., "Bearer", "JWT")
	Type string

	// ExpiresAt indicates when the token expires
	ExpiresAt time.Time

	// IssuedAt indicates when the token was issued
	IssuedAt time.Time

	// Metadata contains additional token metadata
	Metadata map[string]any
}

// Claims represents extracted claims from a token
type Claims map[string]any

// GetString retrieves a string claim
func (c Claims) GetString(key string) (string, bool) {
	val, ok := c[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetInt64 retrieves an int64 claim
func (c Claims) GetInt64(key string) (int64, bool) {
	val, ok := c[key]
	if !ok {
		return 0, false
	}

	switch v := val.(type) {
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	default:
		return 0, false
	}
}

// GetBool retrieves a bool claim
func (c Claims) GetBool(key string) (bool, bool) {
	val, ok := c[key]
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

// GetStringSlice retrieves a string slice claim
func (c Claims) GetStringSlice(key string) ([]string, bool) {
	val, ok := c[key]
	if !ok {
		return nil, false
	}

	switch v := val.(type) {
	case []string:
		return v, true
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result, true
	default:
		return nil, false
	}
}

// VerificationResult represents the result of token verification
type VerificationResult struct {
	// Valid indicates whether the token is valid
	Valid bool

	// Claims contains the extracted claims
	Claims Claims

	// Error contains the error if verification failed
	Error error

	// Metadata contains additional verification metadata
	Metadata map[string]any
}

// TokenGenerator generates tokens from claims
type TokenGenerator interface {
	// Generate creates a new token from the provided claims
	Generate(ctx context.Context, claims Claims) (*Token, error)

	// Type returns the type of tokens this generator creates
	Type() string
}

// TokenVerifier verifies and validates tokens
type TokenVerifier interface {
	// Verify validates a token and extracts its claims
	Verify(ctx context.Context, tokenValue string) (*VerificationResult, error)

	// Type returns the type of tokens this verifier handles
	Type() string
}

// ClaimExtractor extracts specific claims from a token
type ClaimExtractor interface {
	// Extract extracts claims from a token
	Extract(ctx context.Context, token *Token) (Claims, error)

	// ExtractClaim extracts a specific claim by key
	ExtractClaim(ctx context.Context, token *Token, key string) (any, error)
}

// TokenManager combines generation and verification
type TokenManager interface {
	TokenGenerator
	TokenVerifier
}

// RefreshTokenHandler handles token refresh operations
type RefreshTokenHandler interface {
	// Refresh generates a new access token from a refresh token
	Refresh(ctx context.Context, refreshToken string) (*Token, error)

	// Revoke invalidates a refresh token
	Revoke(ctx context.Context, refreshToken string) error
}

// TokenStore stores and retrieves tokens
type TokenStore interface {
	// Store saves a token
	Store(ctx context.Context, subject string, token *Token) error

	// Get retrieves a token
	Get(ctx context.Context, subject string, tokenID string) (*Token, error)

	// Delete removes a token
	Delete(ctx context.Context, subject string, tokenID string) error

	// List returns all tokens for a subject
	List(ctx context.Context, subject string) ([]*Token, error)

	// Revoke invalidates a token
	Revoke(ctx context.Context, tokenID string) error

	// IsRevoked checks if a token is revoked
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
}

// TokenRevocationList manages revoked tokens
type TokenRevocationList interface {
	// Add adds a token to the revocation list
	Add(ctx context.Context, tokenID string, expiresAt time.Time) error

	// IsRevoked checks if a token is in the revocation list
	IsRevoked(ctx context.Context, tokenID string) (bool, error)

	// Remove removes a token from the revocation list (typically after expiry)
	Remove(ctx context.Context, tokenID string) error

	// Cleanup removes expired tokens from the revocation list
	Cleanup(ctx context.Context) error
}
