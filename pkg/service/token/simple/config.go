package simple

import (
	"time"
)

// Config holds simple token manager configuration
type Config struct {
	// TokenLength is the length of generated tokens in bytes (default: 32)
	TokenLength int `json:"token_length"`

	// TokenDuration is how long tokens are valid
	TokenDuration time.Duration `json:"token_duration"`
}

// DefaultConfig returns a default simple token configuration
func DefaultConfig() *Config {
	return &Config{
		TokenLength:   32,
		TokenDuration: 1 * time.Hour,
	}
}
