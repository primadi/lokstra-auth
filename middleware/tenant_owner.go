package middleware

import (
	"context"
	"net/http"

	"github.com/primadi/lokstra-auth/identity"
)

// TenantOwnerChecker defines interface for checking if a user is tenant owner
type TenantOwnerChecker interface {
	IsTenantOwner(ctx context.Context, tenantID, userID string) (bool, error)
}

// TenantOwnerMiddleware creates middleware that grants implicit admin permissions to tenant owners
// Tenant owners are identified by metadata.is_tenant_owner = true
func TenantOwnerMiddleware(checker TenantOwnerChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract identity context from context (set by auth middleware)
			idCtx, ok := r.Context().Value("identity").(identity.IdentityContext)
			if !ok {
				// No identity, continue (auth middleware will handle)
				next.ServeHTTP(w, r)
				return
			}

			// Check if user is tenant owner
			isOwner, err := checker.IsTenantOwner(r.Context(), idCtx.TenantID, idCtx.Subject.ID)
			if err != nil {
				// Log error but don't fail - continue normal authorization flow
				// TODO: Add proper logging
				next.ServeHTTP(w, r)
				return
			}

			if isOwner {
				// Grant implicit tenant-admin permissions
				// Add "tenant-owner" and "tenant-admin" roles if not present
				if !idCtx.HasRole("tenant-owner") || !idCtx.HasRole("tenant-admin") {
					// Clone identity context and add implicit roles
					enhancedCtx := idCtx
					if !idCtx.HasRole("tenant-owner") {
						enhancedCtx.Roles = append(enhancedCtx.Roles, "tenant-owner")
					}
					if !idCtx.HasRole("tenant-admin") {
						enhancedCtx.Roles = append(enhancedCtx.Roles, "tenant-admin")
					}

					// Store enhanced identity context
					ctx := context.WithValue(r.Context(), "identity", enhancedCtx)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// InMemoryTenantOwnerChecker checks tenant ownership using in-memory metadata
type InMemoryTenantOwnerChecker struct {
	// Map: tenantID -> userID -> isOwner
	owners map[string]map[string]bool
}

func NewInMemoryTenantOwnerChecker() *InMemoryTenantOwnerChecker {
	return &InMemoryTenantOwnerChecker{
		owners: make(map[string]map[string]bool),
	}
}

func (c *InMemoryTenantOwnerChecker) IsTenantOwner(ctx context.Context, tenantID, userID string) (bool, error) {
	if tenantOwners, ok := c.owners[tenantID]; ok {
		return tenantOwners[userID], nil
	}
	return false, nil
}

func (c *InMemoryTenantOwnerChecker) SetTenantOwner(tenantID, userID string, isOwner bool) {
	if _, ok := c.owners[tenantID]; !ok {
		c.owners[tenantID] = make(map[string]bool)
	}
	c.owners[tenantID][userID] = isOwner
}
