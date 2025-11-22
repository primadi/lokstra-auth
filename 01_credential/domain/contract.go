package domain

import "context"

// Credentials represents the input credentials from a user
type Credentials interface {
	// Type returns the type of credentials (e.g., "basic", "oauth2", "apikey")
	Type() string

	// Validate checks if the credentials are well-formed
	Validate() error
}

// Authenticator defines the interface for authenticating credentials
type Authenticator interface {
	// Authenticate verifies the provided credentials within the given context
	// and returns the authentication result.
	//
	// The AuthContext MUST contain valid tenant_id and app_id for proper
	// tenant isolation and app scoping.
	Authenticate(ctx context.Context, authCtx *AuthContext, credentials Credentials) (*AuthenticationResult, error)

	// Type returns the type of authenticator (must match Credentials.Type())
	Type() string
}
