package credential

import (
	"context"
)

// Credentials represents the input credentials from a user
type Credentials interface {
	// Type returns the type of credentials (e.g., "basic", "oauth2", "apikey")
	Type() string

	// Validate checks if the credentials are well-formed
	Validate() error
}

// AuthenticationResult represents the result of an authentication attempt
type AuthenticationResult struct {
	// Success indicates whether authentication was successful
	Success bool

	// Subject is the authenticated subject identifier (e.g., user ID, email)
	Subject string

	// Claims contains additional claims to be included in the token
	Claims map[string]any

	// Metadata contains additional metadata about the authentication
	Metadata map[string]any

	// Error contains the error if authentication failed
	Error error
}

// Authenticator defines the interface for authenticating credentials
type Authenticator interface {
	// Authenticate verifies the provided credentials and returns the result
	Authenticate(ctx context.Context, credentials Credentials) (*AuthenticationResult, error)

	// Type returns the type of authenticator (must match Credentials.Type())
	Type() string
}

// CredentialValidator validates credentials format and basic rules
type CredentialValidator interface {
	// Validate checks if the credentials meet the required format and rules
	Validate(ctx context.Context, credentials Credentials) error

	// Type returns the type of credentials this validator handles
	Type() string
}

// LoginFlowHandler manages the complete login flow
type LoginFlowHandler interface {
	// HandleLogin processes the complete login flow and returns authentication result
	HandleLogin(ctx context.Context, credentials Credentials) (*AuthenticationResult, error)

	// SupportedTypes returns the list of credential types this handler supports
	SupportedTypes() []string
}

// AuthenticationMiddleware can be used to add additional logic to authentication
type AuthenticationMiddleware interface {
	// Process is called during authentication to add additional logic
	// (e.g., rate limiting, logging, 2FA checks)
	Process(ctx context.Context, credentials Credentials, next Authenticator) (*AuthenticationResult, error)
}

// CredentialStore defines the interface for storing and retrieving credentials
type CredentialStore interface {
	// Get retrieves stored credentials for a subject
	Get(ctx context.Context, subject string, credType string) (any, error)

	// Store saves credentials for a subject
	Store(ctx context.Context, subject string, credType string, data any) error

	// Delete removes credentials for a subject
	Delete(ctx context.Context, subject string, credType string) error

	// Verify checks if the provided credentials match the stored ones
	Verify(ctx context.Context, subject string, credentials Credentials) (bool, error)
}
