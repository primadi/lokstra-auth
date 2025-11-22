package basic

import (
	"errors"
	"strings"

	"github.com/primadi/lokstra-auth/01_credential/domain"
)

var (
	ErrEmptyUsername        = errors.New("username cannot be empty")
	ErrEmptyPassword        = errors.New("password cannot be empty")
	ErrInvalidCredentials   = errors.New("invalid credentials type")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrUserNotFound         = errors.New("user not found")
)

// Credentials represents username/password credentials
type Credentials struct {
	Username string
	Password string
}

var _ domain.Credentials = (*Credentials)(nil)

// Type returns the credential type
func (c *Credentials) Type() string {
	return "basic"
}

// Validate checks if the credentials are well-formed
func (c *Credentials) Validate() error {
	if strings.TrimSpace(c.Username) == "" {
		return ErrEmptyUsername
	}
	if strings.TrimSpace(c.Password) == "" {
		return ErrEmptyPassword
	}
	return nil
}

// User represents a user in the basic authentication system
type User struct {
	ID           string         // Unique user identifier
	TenantID     string         // Tenant identifier (required for multi-tenant)
	Username     string         // Username for login (unique within tenant)
	PasswordHash string         // Bcrypt hashed password
	Email        string         // User email address
	Disabled     bool           // Whether the user account is disabled
	Metadata     map[string]any // Additional custom user metadata
}
