package basic

import (
	"errors"
	"strings"

	"github.com/primadi/lokstra-auth/credential/domain"
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

// UserCredential stores only the password hash (separates credentials from user profile)
type UserCredential struct {
	UserID       string // References core/domain.User.ID
	TenantID     string // References core/domain.User.TenantID
	PasswordHash string // Bcrypt hashed password
}
