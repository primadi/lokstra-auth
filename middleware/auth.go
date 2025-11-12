package middleware

import (
	"errors"
	"strings"

	lokstraauth "github.com/primadi/lokstra-auth"
	subject "github.com/primadi/lokstra-auth/03_subject"
	"github.com/primadi/lokstra/core/request"
)

// IdentityContextKey is the key used to store identity in context
const IdentityContextKey = "lokstra_auth_identity"

// AuthMiddleware creates a middleware that verifies JWT tokens
// and injects identity into the request context
type AuthMiddleware struct {
	auth           *lokstraauth.Auth
	tokenExtractor TokenExtractor
	errorHandler   ErrorHandler
	optional       bool
}

// TokenExtractor extracts token from request
type TokenExtractor func(c *request.Context) (string, error)

// ErrorHandler handles authentication errors
type ErrorHandler func(c *request.Context, err error) error

// AuthMiddlewareConfig holds configuration for auth middleware
type AuthMiddlewareConfig struct {
	// Auth is the Auth runtime instance
	Auth *lokstraauth.Auth

	// TokenExtractor extracts token from request (default: from Authorization header)
	TokenExtractor TokenExtractor

	// ErrorHandler handles auth errors (default: return 401)
	ErrorHandler ErrorHandler

	// Optional indicates if authentication is optional (default: false)
	// If true, requests without token are allowed to proceed
	Optional bool
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config AuthMiddlewareConfig) *AuthMiddleware {
	if config.TokenExtractor == nil {
		config.TokenExtractor = DefaultTokenExtractor
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultErrorHandler
	}

	return &AuthMiddleware{
		auth:           config.Auth,
		tokenExtractor: config.TokenExtractor,
		errorHandler:   config.ErrorHandler,
		optional:       config.Optional,
	}
}

// Handler returns the middleware handler function
func (m *AuthMiddleware) Handler() func(c *request.Context) error {
	return func(c *request.Context) error {
		// Extract token from request
		token, err := m.tokenExtractor(c)
		if err != nil {
			if m.optional {
				// No token, but optional - continue without identity
				return c.Next()
			}
			return m.errorHandler(c, err)
		}

		// Verify token and build identity context
		verifyResp, err := m.auth.Verify(c, &lokstraauth.VerifyRequest{
			Token:                token,
			BuildIdentityContext: true,
		})
		if err != nil {
			return m.errorHandler(c, err)
		}

		if !verifyResp.Valid {
			return m.errorHandler(c, lokstraauth.ErrAuthenticationFailed)
		}

		// Inject identity into context
		if verifyResp.Identity != nil {
			c.Set(IdentityContextKey, verifyResp.Identity)
		}

		// Continue to next handler
		return c.Next()
	}
}

// DefaultTokenExtractor extracts token from Authorization header
// Format: "Bearer <token>"
func DefaultTokenExtractor(c *request.Context) (string, error) {
	auth := c.R.Header.Get("Authorization")
	if auth == "" {
		return "", ErrMissingToken
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", ErrInvalidTokenFormat
	}

	return parts[1], nil
}

// DefaultErrorHandler returns 401 Unauthorized
func DefaultErrorHandler(c *request.Context, err error) error {
	c.Resp.WithStatus(401)
	return c.Resp.Json(map[string]interface{}{
		"error":   "Unauthorized",
		"message": err.Error(),
	})
}

// GetIdentity retrieves identity from request context
func GetIdentity(c *request.Context) (*subject.IdentityContext, bool) {
	identity, ok := c.Get(IdentityContextKey).(*subject.IdentityContext)
	return identity, ok
}

// MustGetIdentity retrieves identity from context or panics
func MustGetIdentity(c *request.Context) *subject.IdentityContext {
	identity, ok := GetIdentity(c)
	if !ok {
		panic("identity not found in context")
	}
	return identity
}

var (
	ErrMissingToken       = errors.New("missing authentication token")
	ErrInvalidTokenFormat = errors.New("invalid token format, expected 'Bearer <token>'")
)
