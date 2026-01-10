package oauth2

import (
	"errors"

	domain "github.com/primadi/lokstra-auth/pkg/domain/credential"
)

var (
	ErrInvalidProvider = errors.New("invalid oauth2 provider")
	ErrMissingCode     = errors.New("missing authorization code")
	ErrMissingState    = errors.New("missing state parameter")
)

// Credentials represents OAuth2 credentials (authorization code flow)
type Credentials struct {
	Provider string // "google", "azure", "github", etc.
	Code     string // Authorization code
	State    string // CSRF protection token
}

var _ domain.Credentials = (*Credentials)(nil)

// Type returns the credential type
func (c *Credentials) Type() string {
	return "oauth2"
}

// Validate checks if the credentials are well-formed
func (c *Credentials) Validate() error {
	if c.Provider == "" {
		return ErrInvalidProvider
	}
	if c.Code == "" {
		return ErrMissingCode
	}
	if c.State == "" {
		return ErrMissingState
	}
	return nil
}
