package middleware

import (
	lokstraauth "github.com/primadi/lokstra-auth"
	authz "github.com/primadi/lokstra-auth/authz"
	"github.com/primadi/lokstra/core/request"
)

// RoleMiddleware creates a middleware that checks if user has required role
type RoleMiddleware struct {
	auth         *lokstraauth.Auth
	role         string
	errorHandler ErrorHandler
}

// RoleMiddlewareConfig holds configuration for role middleware
type RoleMiddlewareConfig struct {
	// Auth is the Auth runtime instance
	Auth *lokstraauth.Auth

	// Role is the required role (e.g., "admin", "developer", "team-lead")
	Role string

	// ErrorHandler handles authorization errors (default: return 403)
	ErrorHandler ErrorHandler
}

// NewRoleMiddleware creates a new role check middleware
func NewRoleMiddleware(config RoleMiddlewareConfig) *RoleMiddleware {
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultForbiddenHandler
	}

	return &RoleMiddleware{
		auth:         config.Auth,
		role:         config.Role,
		errorHandler: config.ErrorHandler,
	}
}

// Handler returns the middleware handler function
func (m *RoleMiddleware) Handler() func(c *request.Context) error {
	return func(c *request.Context) error {
		// Get identity from context (should be set by AuthMiddleware)
		identity, ok := GetIdentity(c)
		if !ok {
			return m.errorHandler(c, lokstraauth.ErrAuthenticationFailed)
		}

		// Check role
		hasRole, err := m.auth.CheckRole(c, identity, m.role)
		if err != nil {
			return m.errorHandler(c, err)
		}

		if !hasRole {
			return m.errorHandler(c, lokstraauth.ErrAuthorizationFailed)
		}

		// Continue to next handler
		return c.Next()
	}
}

// RequireRole creates a role middleware with shorthand
func RequireRole(auth *lokstraauth.Auth, role string) func(c *request.Context) error {
	middleware := NewRoleMiddleware(RoleMiddlewareConfig{
		Auth: auth,
		Role: role,
	})
	return middleware.Handler()
}

// AnyRoleMiddleware checks if user has ANY of the specified roles
type AnyRoleMiddleware struct {
	auth         *lokstraauth.Auth
	roles        []string
	errorHandler ErrorHandler
}

// NewAnyRoleMiddleware creates middleware that requires any of the roles
func NewAnyRoleMiddleware(auth *lokstraauth.Auth, roles []string) *AnyRoleMiddleware {
	return &AnyRoleMiddleware{
		auth:         auth,
		roles:        roles,
		errorHandler: DefaultForbiddenHandler,
	}
}

// Handler returns the middleware handler function
func (m *AnyRoleMiddleware) Handler() func(c *request.Context) error {
	return func(c *request.Context) error {
		identity, ok := GetIdentity(c)
		if !ok {
			return m.errorHandler(c, lokstraauth.ErrAuthenticationFailed)
		}

		// Check if user has any of the roles
		checker, ok := m.auth.GetAuthorizer().(authz.RoleChecker)
		if !ok {
			return m.errorHandler(c, lokstraauth.ErrNoAuthorizer)
		}

		hasRole, err := checker.HasAnyRole(c, identity, m.roles...)
		if err != nil {
			return m.errorHandler(c, err)
		}

		if !hasRole {
			return m.errorHandler(c, lokstraauth.ErrAuthorizationFailed)
		}

		return c.Next()
	}
}

// AllRolesMiddleware checks if user has ALL of the specified roles
type AllRolesMiddleware struct {
	auth         *lokstraauth.Auth
	roles        []string
	errorHandler ErrorHandler
}

// NewAllRolesMiddleware creates middleware that requires all roles
func NewAllRolesMiddleware(auth *lokstraauth.Auth, roles []string) *AllRolesMiddleware {
	return &AllRolesMiddleware{
		auth:         auth,
		roles:        roles,
		errorHandler: DefaultForbiddenHandler,
	}
}

// Handler returns the middleware handler function
func (m *AllRolesMiddleware) Handler() func(c *request.Context) error {
	return func(c *request.Context) error {
		identity, ok := GetIdentity(c)
		if !ok {
			return m.errorHandler(c, lokstraauth.ErrAuthenticationFailed)
		}

		// Check if user has all of the roles
		checker, ok := m.auth.GetAuthorizer().(authz.RoleChecker)
		if !ok {
			return m.errorHandler(c, lokstraauth.ErrNoAuthorizer)
		}

		hasRoles, err := checker.HasAllRoles(c, identity, m.roles...)
		if err != nil {
			return m.errorHandler(c, err)
		}

		if !hasRoles {
			return m.errorHandler(c, lokstraauth.ErrAuthorizationFailed)
		}

		return c.Next()
	}
}
