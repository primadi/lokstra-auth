package simple

import (
	"time"

	token "github.com/primadi/lokstra-auth/token"
)

// Config holds simple token manager configuration
type Config struct {
	// TokenLength is the length of generated tokens in bytes (default: 32)
	TokenLength int `json:"token_length"`

	// TokenDuration is how long tokens are valid
	TokenDuration time.Duration `json:"token_duration"`

	// Store is the token store (optional, uses in-memory if not provided)
	Store token.TokenStore `json:"-"`

	// EnableRevocation enables token revocation support
	EnableRevocation bool `json:"enable_revocation"`
}

// DefaultConfig returns a default simple token configuration
func DefaultConfig() *Config {
	return &Config{
		TokenLength:      32,
		TokenDuration:    1 * time.Hour,
		EnableRevocation: false,
	}
}
