package middleware

import (
	lokstraauth "github.com/primadi/lokstra-auth"
	authz "github.com/primadi/lokstra-auth/04_authz"
	"github.com/primadi/lokstra/core/request"
)

// PermissionMiddleware creates a middleware that checks if user has required permission
type PermissionMiddleware struct {
	auth         *lokstraauth.Auth
	permission   string
	errorHandler ErrorHandler
}

// PermissionMiddlewareConfig holds configuration for permission middleware
type PermissionMiddlewareConfig struct {
	// Auth is the Auth runtime instance
	Auth *lokstraauth.Auth

	// Permission is the required permission (e.g., "read:users", "write:posts")
	Permission string

	// ErrorHandler handles authorization errors (default: return 403)
	ErrorHandler ErrorHandler
}

// NewPermissionMiddleware creates a new permission check middleware
func NewPermissionMiddleware(config PermissionMiddlewareConfig) *PermissionMiddleware {
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultForbiddenHandler
	}

	return &PermissionMiddleware{
		auth:         config.Auth,
		permission:   config.Permission,
		errorHandler: config.ErrorHandler,
	}
}

// Handler returns the middleware handler function
func (m *PermissionMiddleware) Handler() func(c *request.Context) error {
	return func(c *request.Context) error {
		// Get identity from context (should be set by AuthMiddleware)
		identity, ok := GetIdentity(c)
		if !ok {
			return m.errorHandler(c, lokstraauth.ErrAuthenticationFailed)
		}

		// Check permission
		hasPermission, err := m.auth.CheckPermission(c, identity, m.permission)
		if err != nil {
			return m.errorHandler(c, err)
		}

		if !hasPermission {
			return m.errorHandler(c, lokstraauth.ErrAuthorizationFailed)
		}

		// Continue to next handler
		return c.Next()
	}
}

// RequirePermission creates a permission middleware with shorthand
func RequirePermission(auth *lokstraauth.Auth, permission string) func(c *request.Context) error {
	middleware := NewPermissionMiddleware(PermissionMiddlewareConfig{
		Auth:       auth,
		Permission: permission,
	})
	return middleware.Handler()
}

// DefaultForbiddenHandler returns 403 Forbidden
func DefaultForbiddenHandler(c *request.Context, err error) error {
	c.Resp.WithStatus(403)
	return c.Resp.Json(map[string]any{
		"error":   "Forbidden",
		"message": err.Error(),
	})
}

// AnyPermissionMiddleware checks if user has ANY of the specified permissions
type AnyPermissionMiddleware struct {
	auth         *lokstraauth.Auth
	permissions  []string
	errorHandler ErrorHandler
}

// NewAnyPermissionMiddleware creates middleware that requires any of the permissions
func NewAnyPermissionMiddleware(auth *lokstraauth.Auth, permissions []string) *AnyPermissionMiddleware {
	return &AnyPermissionMiddleware{
		auth:         auth,
		permissions:  permissions,
		errorHandler: DefaultForbiddenHandler,
	}
}

// Handler returns the middleware handler function
func (m *AnyPermissionMiddleware) Handler() func(c *request.Context) error {
	return func(c *request.Context) error {
		identity, ok := GetIdentity(c)
		if !ok {
			return m.errorHandler(c, lokstraauth.ErrAuthenticationFailed)
		}

		// Check if user has any of the permissions
		checker, ok := m.auth.GetAuthorizer().(authz.PermissionChecker)
		if !ok {
			return m.errorHandler(c, lokstraauth.ErrNoAuthorizer)
		}

		hasPermission, err := checker.HasAnyPermission(c, identity, m.permissions...)
		if err != nil {
			return m.errorHandler(c, err)
		}

		if !hasPermission {
			return m.errorHandler(c, lokstraauth.ErrAuthorizationFailed)
		}

		return c.Next()
	}
}

// AllPermissionsMiddleware checks if user has ALL of the specified permissions
type AllPermissionsMiddleware struct {
	auth         *lokstraauth.Auth
	permissions  []string
	errorHandler ErrorHandler
}

// NewAllPermissionsMiddleware creates middleware that requires all permissions
func NewAllPermissionsMiddleware(auth *lokstraauth.Auth, permissions []string) *AllPermissionsMiddleware {
	return &AllPermissionsMiddleware{
		auth:         auth,
		permissions:  permissions,
		errorHandler: DefaultForbiddenHandler,
	}
}

// Handler returns the middleware handler function
func (m *AllPermissionsMiddleware) Handler() func(c *request.Context) error {
	return func(c *request.Context) error {
		identity, ok := GetIdentity(c)
		if !ok {
			return m.errorHandler(c, lokstraauth.ErrAuthenticationFailed)
		}

		// Check if user has all of the permissions
		checker, ok := m.auth.GetAuthorizer().(authz.PermissionChecker)
		if !ok {
			return m.errorHandler(c, lokstraauth.ErrNoAuthorizer)
		}

		hasPermissions, err := checker.HasAllPermissions(c, identity, m.permissions...)
		if err != nil {
			return m.errorHandler(c, err)
		}

		if !hasPermissions {
			return m.errorHandler(c, lokstraauth.ErrAuthorizationFailed)
		}

		return c.Next()
	}
}
