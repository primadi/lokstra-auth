package lokstraauth

import (
	authz "github.com/primadi/lokstra-auth/authz"
	credential "github.com/primadi/lokstra-auth/credential/domain"
	identity "github.com/primadi/lokstra-auth/identity"
	token "github.com/primadi/lokstra-auth/token"
)

// Builder provides a fluent API for building Auth runtime
type Builder struct {
	auth *Auth
}

// NewBuilder creates a new Auth builder
func NewBuilder() *Builder {
	return &Builder{
		auth: New(DefaultConfig()),
	}
}

// WithConfig sets a custom configuration
func (b *Builder) WithConfig(config *Config) *Builder {
	b.auth.config = config
	return b
}

// WithAuthenticator registers an authenticator
func (b *Builder) WithAuthenticator(authType string, authenticator credential.Authenticator) *Builder {
	b.auth.RegisterAuthenticator(authType, authenticator)
	return b
}

// WithTokenManager sets the token manager
func (b *Builder) WithTokenManager(manager token.TokenManager) *Builder {
	b.auth.SetTokenManager(manager)
	return b
}

// WithIdentityResolver sets the identity resolver
func (b *Builder) WithIdentityResolver(resolver identity.IdentityResolver) *Builder {
	b.auth.SetIdentityResolver(resolver)
	return b
}

// WithIdentityContextBuilder sets the identity context builder
func (b *Builder) WithIdentityContextBuilder(builder identity.IdentityContextBuilder) *Builder {
	b.auth.SetIdentityContextBuilder(builder)
	return b
}

// WithAuthorizer sets the authorizer
func (b *Builder) WithAuthorizer(authorizer authz.Authorizer) *Builder {
	b.auth.SetAuthorizer(authorizer)
	return b
}

// EnableRefreshToken enables refresh token generation
func (b *Builder) EnableRefreshToken() *Builder {
	b.auth.config.IssueRefreshToken = true
	return b
}

// DisableRefreshToken disables refresh token generation
func (b *Builder) DisableRefreshToken() *Builder {
	b.auth.config.IssueRefreshToken = false
	return b
}

// EnableSessionManagement enables session management
func (b *Builder) EnableSessionManagement() *Builder {
	b.auth.config.SessionManagement = true
	return b
}

// DisableSessionManagement disables session management
func (b *Builder) DisableSessionManagement() *Builder {
	b.auth.config.SessionManagement = false
	return b
}

// SetDefaultAuthenticator sets the default authenticator type
func (b *Builder) SetDefaultAuthenticator(authType string) *Builder {
	b.auth.config.DefaultAuthenticatorType = authType
	return b
}

// AddMetadata adds metadata to the configuration
func (b *Builder) AddMetadata(key string, value any) *Builder {
	if b.auth.config.Metadata == nil {
		b.auth.config.Metadata = make(map[string]any)
	}
	b.auth.config.Metadata[key] = value
	return b
}

// Build returns the configured Auth instance
func (b *Builder) Build() *Auth {
	return b.auth
}
