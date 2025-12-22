package middleware

import (
	"errors"
	"fmt"

	"github.com/primadi/lokstra/core/request"
)

// TenantMiddleware validates tenant context from header matches token claims
// This prevents cross-tenant access attacks
type TenantMiddleware struct {
	strict bool // If true, require header to match token tenant_id
}

// TenantMiddlewareConfig holds configuration for tenant middleware
type TenantMiddlewareConfig struct {
	// Strict mode: header X-Tenant-ID must match token tenant_id claim
	Strict bool
}

// NewTenantMiddleware creates a new tenant validation middleware
func NewTenantMiddleware(config TenantMiddlewareConfig) *TenantMiddleware {
	return &TenantMiddleware{
		strict: config.Strict,
	}
}

// Handler returns the middleware handler function
func (m *TenantMiddleware) Handler() func(c *request.Context) error {
	return func(c *request.Context) error {
		// Get tenant from header
		headerTenantID := c.R.Header.Get("X-Tenant-ID")

		// Get identity from context (set by AuthMiddleware)
		identity, ok := GetIdentity(c)
		if !ok {
			// If auth middleware not run yet or optional auth, skip validation
			if headerTenantID != "" {
				// Store tenant ID for later use
				c.Set("tenant_id", headerTenantID)
			}
			return c.Next()
		}

		// Validate tenant matches token claims
		if m.strict && headerTenantID != "" && headerTenantID != identity.TenantID {
			c.Resp.WithStatus(403)
			return c.Resp.Json(map[string]any{
				"error":   "Forbidden",
				"message": fmt.Sprintf("tenant mismatch: header=%s, token=%s", headerTenantID, identity.TenantID),
			})
		}

		// If no header provided, use tenant from token
		if headerTenantID == "" {
			headerTenantID = identity.TenantID
		}

		// Inject tenant ID for convenience in handlers
		c.Set("tenant_id", headerTenantID)

		return c.Next()
	}
}

// GetTenantID retrieves tenant ID from context (set by TenantMiddleware)
func GetTenantID(c *request.Context) (string, bool) {
	tenantIDStr, ok := c.Get("tenant_id").(string)
	if !ok {
		return "", false
	}

	return tenantIDStr, true
}

// MustGetTenantID retrieves tenant ID from context or panics
func MustGetTenantID(c *request.Context) string {
	tenantID, ok := GetTenantID(c)
	if !ok {
		panic(errors.New("tenant_id not found in context"))
	}
	return tenantID
}
