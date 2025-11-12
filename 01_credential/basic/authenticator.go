package basic

import (
	"context"
	"crypto/subtle"
	"errors"
	"maps"

	credential "github.com/primadi/lokstra-auth/01_credential"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrUserNotFound         = errors.New("user not found")
)

// UserProvider defines the interface for retrieving user credentials
type UserProvider interface {
	// GetUserByUsername retrieves user information by username
	GetUserByUsername(ctx context.Context, username string) (*User, error)
}

// User represents a user in the system
type User struct {
	ID           string
	Username     string
	PasswordHash string
	Email        string
	Disabled     bool
	Metadata     map[string]any
}

// Authenticator authenticates basic credentials
type Authenticator struct {
	userProvider UserProvider
	validator    credential.CredentialValidator
}

// NewAuthenticator creates a new basic authenticator
func NewAuthenticator(userProvider UserProvider, validator credential.CredentialValidator) *Authenticator {
	return &Authenticator{
		userProvider: userProvider,
		validator:    validator,
	}
}

// Authenticate verifies the provided credentials and returns the result
func (a *Authenticator) Authenticate(ctx context.Context, creds credential.Credentials) (*credential.AuthenticationResult, error) {
	// Validate credentials format
	if a.validator != nil {
		if err := a.validator.Validate(ctx, creds); err != nil {
			return &credential.AuthenticationResult{
				Success: false,
				Error:   err,
			}, nil
		}
	}

	basicCreds, ok := creds.(*BasicCredentials)
	if !ok {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   ErrInvalidCredentials,
		}, nil
	}

	// Retrieve user
	user, err := a.userProvider.GetUserByUsername(ctx, basicCreds.Username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return &credential.AuthenticationResult{
				Success: false,
				Error:   ErrAuthenticationFailed,
			}, nil
		}
		return nil, err
	}

	// Check if user is disabled
	if user.Disabled {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   ErrAuthenticationFailed,
		}, nil
	}

	// Verify password
	if !a.verifyPassword(user.PasswordHash, basicCreds.Password) {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   ErrAuthenticationFailed,
		}, nil
	}

	// Build claims
	claims := map[string]any{
		"sub":      user.ID,
		"username": user.Username,
		"email":    user.Email,
	}

	// Add user metadata to claims
	maps.Copy(claims, user.Metadata)

	return &credential.AuthenticationResult{
		Success: true,
		Subject: user.ID,
		Claims:  claims,
		Metadata: map[string]any{
			"auth_type": "basic",
			"username":  user.Username,
		},
	}, nil
}

// Type returns the type of authenticator
func (a *Authenticator) Type() string {
	return "basic"
}

// verifyPassword compares a hashed password with a plaintext password
func (a *Authenticator) verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// SecureCompare performs a constant-time comparison of two strings
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
