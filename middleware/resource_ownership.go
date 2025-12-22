package middleware

import (
	"context"
	"errors"

	"github.com/primadi/lokstra/core/request"
)

// ResourceOwnerChecker is the interface for checking resource ownership
type ResourceOwnerChecker interface {
	// IsOwner checks if user owns the resource
	IsOwner(ctx context.Context, tenantID, userID, resourceType, resourceID string) (bool, error)
}

// ResourceOwnershipMiddleware checks if user owns the resource they're accessing
// This prevents cross-user access within the same tenant
type ResourceOwnershipMiddleware struct {
	checker      ResourceOwnerChecker
	resourceType string
	paramName    string // Path parameter name (e.g., "id", "orderId", "documentId")
	allowAdmin   bool   // Allow admins to access all resources
}

// ResourceOwnershipConfig holds configuration for resource ownership middleware
type ResourceOwnershipConfig struct {
	// Checker is the ownership verification implementation
	Checker ResourceOwnerChecker

	// ResourceType is the type of resource being checked (e.g., "order", "document", "profile")
	ResourceType string

	// ParamName is the path parameter containing resource ID
	ParamName string

	// AllowAdmin allows users with "admin" role to bypass ownership check
	AllowAdmin bool
}

// NewResourceOwnershipMiddleware creates a new resource ownership middleware
func NewResourceOwnershipMiddleware(config ResourceOwnershipConfig) *ResourceOwnershipMiddleware {
	return &ResourceOwnershipMiddleware{
		checker:      config.Checker,
		resourceType: config.ResourceType,
		paramName:    config.ParamName,
		allowAdmin:   config.AllowAdmin,
	}
}

// Handler returns the middleware handler function
func (m *ResourceOwnershipMiddleware) Handler() func(c *request.Context) error {
	return func(c *request.Context) error {
		// Get identity from context (requires AuthMiddleware to run first)
		identity, ok := GetIdentity(c)
		if !ok {
			c.Resp.WithStatus(401)
			return c.Resp.Json(map[string]any{
				"error":   "Unauthorized",
				"message": "authentication required for resource ownership check",
			})
		}

		// If admin bypass is enabled, check for admin role
		if m.allowAdmin {
			// Check if user has admin role
			for _, role := range identity.Roles {
				if role == "admin" || role == "super_admin" {
					// Admin can access all resources
					return c.Next()
				}
			}
		}

		// Get resource ID from path parameter
		resourceID := c.Req.PathParam(m.paramName, "")
		if resourceID == "" {
			c.Resp.WithStatus(400)
			return c.Resp.Json(map[string]any{
				"error":   "Bad Request",
				"message": "resource ID not found in path",
			})
		}

		// Check ownership
		isOwner, err := m.checker.IsOwner(
			c,
			identity.TenantID,
			identity.Subject.ID,
			m.resourceType,
			resourceID,
		)

		if err != nil {
			c.Resp.WithStatus(500)
			return c.Resp.Json(map[string]any{
				"error":   "Internal Server Error",
				"message": "failed to verify resource ownership",
			})
		}

		if !isOwner {
			c.Resp.WithStatus(403)
			return c.Resp.Json(map[string]any{
				"error":   "Forbidden",
				"message": "you do not own this resource",
			})
		}

		// Ownership verified, continue to handler
		return c.Next()
	}
}

// InMemoryResourceOwnerChecker is a simple in-memory ownership checker (for testing)
type InMemoryResourceOwnerChecker struct {
	// Map of resourceType -> resourceID -> ownerUserID
	ownership map[string]map[string]string
}

// NewInMemoryResourceOwnerChecker creates a new in-memory ownership checker
func NewInMemoryResourceOwnerChecker() *InMemoryResourceOwnerChecker {
	return &InMemoryResourceOwnerChecker{
		ownership: make(map[string]map[string]string),
	}
}

// SetOwner sets the owner of a resource
func (c *InMemoryResourceOwnerChecker) SetOwner(resourceType, resourceID, ownerUserID string) {
	if c.ownership[resourceType] == nil {
		c.ownership[resourceType] = make(map[string]string)
	}
	c.ownership[resourceType][resourceID] = ownerUserID
}

func (c *InMemoryResourceOwnerChecker) IsOwner(ctx context.Context, tenantID, userID, resourceType, resourceID string) (bool, error) {
	if c.ownership[resourceType] == nil {
		return false, errors.New("resource type not found")
	}

	ownerID, exists := c.ownership[resourceType][resourceID]
	if !exists {
		return false, errors.New("resource not found")
	}

	return ownerID == userID, nil
}
