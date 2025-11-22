package oauth2

import (
	"errors"

	"github.com/primadi/lokstra-auth/01_credential/domain"
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

// ProviderInfo represents OAuth2 provider information
type ProviderInfo struct {
	Name         string   `json:"name"` // "google", "azure", etc.
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"user_info_url"`
	Scopes       []string `json:"scopes"`
}

// UserInfo represents user information from OAuth2 provider
type UserInfo struct {
	ID            string         `json:"id"` // Provider user ID
	Email         string         `json:"email"`
	Name          string         `json:"name"`
	Picture       string         `json:"picture"`
	EmailVerified bool           `json:"email_verified"`
	Metadata      map[string]any `json:"metadata"` // Provider-specific data
}
