package middleware

import (
	"net/http"

	"github.com/primadi/lokstra-auth/identity"
)

// PlatformAdminMiddleware ensures only platform admins can access the endpoint
// Platform admins are users in the "platform" tenant with "platform-admin" role
func PlatformAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract identity context from context (set by auth middleware)
		idCtx, ok := r.Context().Value("identity").(identity.IdentityContext)
		if !ok {
			http.Error(w, "Unauthorized: No identity in context", http.StatusUnauthorized)
			return
		}

		// Check if user is in platform tenant
		if idCtx.TenantID != "platform" {
			http.Error(w, "Forbidden: Platform admin access required", http.StatusForbidden)
			return
		}

		// Check if user has platform-admin role
		if !idCtx.HasRole("platform-admin") {
			http.Error(w, "Forbidden: Platform admin role required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
