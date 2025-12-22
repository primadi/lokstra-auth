# Middleware Usage Example

This example demonstrates how to register and use all middleware in your Lokstra application.

## Complete Middleware Stack

```go
package main

import (
	"time"
	
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra-auth/middleware"
	token "github.com/primadi/lokstra-auth/token"
	"github.com/primadi/lokstra-auth/token/jwt"
)

func main() {
	app := lokstra.New()
	
	// =============================================================================
	// Initialize Token Manager
	// =============================================================================
	
	jwtConfig := jwt.DefaultConfig("your-secret-key-min-32-chars-long!")
	jwtConfig.EnableRevocation = true
	tokenManager := jwt.NewManager(jwtConfig)
	
	// =============================================================================
	// Initialize Middleware
	// =============================================================================
	
	// 1. Auth Middleware (verify token, inject identity)
	authMiddleware := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
		TokenManager: tokenManager,
		Optional:     false, // true for public + private routes
	})
	
	// 2. Tenant Middleware (validate tenant context)
	tenantMiddleware := middleware.NewTenantMiddleware(middleware.TenantMiddlewareConfig{
		Strict: true, // header must match token
	})
	
	// 3. Rate Limit Middleware (prevent abuse)
	rateLimiter := middleware.NewInMemoryRateLimiter(100, time.Minute)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(middleware.RateLimitConfig{
		Limiter: rateLimiter,
		KeyFunc: middleware.PerTenantKey, // or PerUserKey, PerIPKey
		Limit:   100,
		Window:  time.Minute,
	})
	
	// 4. Audit Middleware (log all operations)
	auditLogger := middleware.NewConsoleAuditLogger() // or implement PostgresAuditLogger
	auditMiddleware := middleware.NewAuditMiddleware(middleware.AuditConfig{
		Logger: auditLogger,
		Skip: func(c *request.Context) bool {
			// Skip health checks and static files
			return c.R.URL.Path == "/health" || c.R.URL.Path == "/metrics"
		},
	})
	
	// 5. Resource Ownership Middleware (check ownership)
	ownershipChecker := middleware.NewInMemoryResourceOwnerChecker()
	orderOwnershipMiddleware := middleware.NewResourceOwnershipMiddleware(middleware.ResourceOwnershipConfig{
		Checker:      ownershipChecker,
		ResourceType: "order",
		ParamName:    "id",
		AllowAdmin:   true, // admins can access all
	})
	
	// =============================================================================
	// Global Middleware (Applied to All Routes)
	// =============================================================================
	
	router := app.Router()
	
	// Order matters! Apply in this sequence:
	router.Use(
		recoveryMiddleware(),        // 1. Catch panics
		loggingMiddleware(),         // 2. Log all requests
		corsMiddleware(),            // 3. Handle CORS
		auditMiddleware.Handler(),   // 4. Audit (before auth - logs failed auth too)
	)
	
	// =============================================================================
	// Public Routes (No Authentication)
	// =============================================================================
	
	publicAPI := router.Group("/api/public")
	{
		publicAPI.GET("/health", healthHandler)
		publicAPI.GET("/docs", docsHandler)
	}
	
	// =============================================================================
	// Authentication Routes (Special Case - No Auth Required)
	// =============================================================================
	
	authAPI := router.Group("/api/auth")
	authAPI.Use(
		rateLimitMiddleware.Handler(), // Rate limit auth endpoints
	)
	{
		authAPI.POST("/login", loginHandler)
		authAPI.POST("/register", registerHandler)
		authAPI.POST("/refresh", refreshHandler)
		authAPI.POST("/logout", logoutHandler)
	}
	
	// =============================================================================
	// Protected Routes (Authentication Required)
	// =============================================================================
	
	protectedAPI := router.Group("/api")
	protectedAPI.Use(
		authMiddleware.Handler(),      // 5. Verify token
		tenantMiddleware.Handler(),    // 6. Validate tenant
		rateLimitMiddleware.Handler(), // 7. Rate limit
	)
	{
		// User profile routes
		protectedAPI.GET("/profile", getProfileHandler)
		protectedAPI.PUT("/profile", updateProfileHandler)
		
		// Resource routes with ownership check
		ordersAPI := protectedAPI.Group("/orders")
		ordersAPI.Use(orderOwnershipMiddleware.Handler())
		{
			ordersAPI.GET("/{id}", getOrderHandler)
			ordersAPI.PUT("/{id}", updateOrderHandler)
			ordersAPI.DELETE("/{id}", deleteOrderHandler)
		}
	}
	
	// =============================================================================
	// Admin Routes (Role-Based Access)
	// =============================================================================
	
	adminAPI := router.Group("/api/admin")
	adminAPI.Use(
		authMiddleware.Handler(),
		tenantMiddleware.Handler(),
		middleware.RequireRole("admin"), // 8. Check role
	)
	{
		adminAPI.GET("/users", listUsersHandler)
		adminAPI.POST("/users", createUserHandler)
		adminAPI.DELETE("/users/{id}", deleteUserHandler)
	}
	
	// =============================================================================
	// Permission-Based Routes
	// =============================================================================
	
	documentsAPI := router.Group("/api/documents")
	documentsAPI.Use(
		authMiddleware.Handler(),
		tenantMiddleware.Handler(),
	)
	{
		// Read documents - requires read:documents permission
		documentsAPI.GET("", middleware.RequirePermission("read:documents"), listDocumentsHandler)
		documentsAPI.GET("/{id}", middleware.RequirePermission("read:documents"), getDocumentHandler)
		
		// Create documents - requires create:documents OR admin:all permission
		documentsAPI.POST("", 
			middleware.NewAnyPermissionMiddleware([]string{"create:documents", "admin:all"}).Handler(),
			createDocumentHandler,
		)
		
		// Delete documents - requires BOTH delete:documents AND admin:confirm permissions
		documentsAPI.DELETE("/{id}",
			middleware.NewAllPermissionsMiddleware([]string{"delete:documents", "admin:confirm"}).Handler(),
			deleteDocumentHandler,
		)
	}
	
	// =============================================================================
	// Conditional Auth Routes (Public + Private)
	// =============================================================================
	
	productsAPI := router.Group("/api/products")
	
	// Optional auth middleware
	optionalAuth := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
		TokenManager: tokenManager,
		Optional:     true, // Don't fail if no token
	})
	
	productsAPI.Use(optionalAuth.Handler())
	{
		// Returns public catalog for anonymous
		// Returns personalized catalog for authenticated users
		productsAPI.GET("", listProductsHandler)
	}
	
	// =============================================================================
	// Complex Middleware Chain Example
	// =============================================================================
	
	// Full stack: Auth → Tenant → Rate Limit → Audit → Role → Permission → Ownership
	sensitiveAPI := router.Group("/api/sensitive")
	sensitiveAPI.Use(
		authMiddleware.Handler(),                                    // 1. Authenticate
		tenantMiddleware.Handler(),                                  // 2. Validate tenant
		rateLimitMiddleware.Handler(),                               // 3. Rate limit
		middleware.NewAllRolesMiddleware([]string{"admin", "auditor"}).Handler(), // 4. Require both roles
		middleware.RequirePermission("access:sensitive"),            // 5. Require permission
	)
	{
		sensitiveAPI.POST("/operation", sensitiveOperationHandler)
	}
	
	app.Run(":8080")
}

// =============================================================================
// Handler Examples
// =============================================================================

func healthHandler(c *request.Context) error {
	return c.Resp.Json(map[string]string{"status": "healthy"})
}

func getProfileHandler(c *request.Context) error {
	// Get identity injected by AuthMiddleware
	identity, ok := middleware.GetIdentity(c)
	if !ok {
		return c.Resp.Status(401).Json(map[string]string{"error": "unauthorized"})
	}
	
	// Get tenant ID injected by TenantMiddleware
	tenantID, _ := middleware.GetTenantID(c)
	
	return c.Resp.Json(map[string]any{
		"user_id":   identity.Subject.ID,
		"username":  identity.Subject.Principal,
		"tenant_id": tenantID,
		"app_id":    identity.AppID,
		"roles":     identity.Roles,
	})
}

func listProductsHandler(c *request.Context) error {
	// Check if user is authenticated
	identity, authenticated := middleware.GetIdentity(c)
	
	if authenticated {
		// Return personalized product list
		return c.Resp.Json(map[string]any{
			"products":     getPersonalizedProducts(identity),
			"personalized": true,
		})
	} else {
		// Return public product list
		return c.Resp.Json(map[string]any{
			"products":     getPublicProducts(),
			"personalized": false,
		})
	}
}

// =============================================================================
// Helper Middleware (Not from lokstra-auth)
// =============================================================================

func recoveryMiddleware() func(c *request.Context) error {
	return func(c *request.Context) error {
		defer func() {
			if r := recover(); r != nil {
				c.Resp.Status(500).Json(map[string]any{
					"error": "Internal Server Error",
					"panic": fmt.Sprintf("%v", r),
				})
			}
		}()
		return c.Next()
	}
}

func loggingMiddleware() func(c *request.Context) error {
	return func(c *request.Context) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		
		fmt.Printf("[%s] %s %s - %v\n",
			start.Format("2006-01-02 15:04:05"),
			c.R.Method,
			c.R.URL.Path,
			duration,
		)
		
		return err
	}
}

func corsMiddleware() func(c *request.Context) error {
	return func(c *request.Context) error {
		c.Resp.Header().Set("Access-Control-Allow-Origin", "*")
		c.Resp.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Resp.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID, X-App-ID, X-Branch-ID")
		
		// Handle preflight
		if c.R.Method == "OPTIONS" {
			c.Resp.Status(204)
			return nil
		}
		
		return c.Next()
	}
}
```

## Middleware Combinations Cheat Sheet

| Route Type | Middleware Stack | Use Case |
|-----------|-----------------|----------|
| **Public** | None | Health checks, docs, static files |
| **Auth** | Rate Limit only | Login, register endpoints |
| **Protected** | Auth → Tenant → Rate Limit | Standard authenticated routes |
| **Admin** | Auth → Tenant → Rate Limit → Role | Admin panels |
| **Permission** | Auth → Tenant → Rate Limit → Permission | Feature-specific access |
| **Ownership** | Auth → Tenant → Rate Limit → Permission → Ownership | User-specific resources |
| **Sensitive** | Auth → Tenant → Rate Limit → Role → Permission | Critical operations |
| **Optional** | Optional Auth | Public + personalized content |

## Custom Middleware Examples

### IP Whitelist Middleware
```go
func IPWhitelistMiddleware(allowedIPs []string) func(c *request.Context) error {
	return func(c *request.Context) error {
		clientIP := c.R.RemoteAddr
		if forwarded := c.R.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = forwarded
		}
		
		for _, allowedIP := range allowedIPs {
			if clientIP == allowedIP {
				return c.Next()
			}
		}
		
		return c.Resp.Status(403).Json(map[string]string{
			"error": "IP not whitelisted",
		})
	}
}

// Usage
adminAPI.Use(IPWhitelistMiddleware([]string{"192.168.1.100", "10.0.0.1"}))
```

### Request ID Middleware
```go
func RequestIDMiddleware() func(c *request.Context) error {
	return func(c *request.Context) error {
		requestID := c.R.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Set("request_id", requestID)
		c.Resp.Header().Set("X-Request-ID", requestID)
		
		return c.Next()
	}
}
```

### API Version Middleware
```go
func APIVersionMiddleware(supportedVersions []string) func(c *request.Context) error {
	return func(c *request.Context) error {
		version := c.R.Header.Get("X-API-Version")
		if version == "" {
			version = "v1" // default
		}
		
		// Check if version is supported
		supported := false
		for _, v := range supportedVersions {
			if v == version {
				supported = true
				break
			}
		}
		
		if !supported {
			return c.Resp.Status(400).Json(map[string]string{
				"error": "unsupported API version",
			})
		}
		
		c.Set("api_version", version)
		return c.Next()
	}
}
```

## Production Recommendations

1. **Always use HTTPS** in production
2. **Use Redis** for rate limiting and token revocation (not in-memory)
3. **Use PostgreSQL** for audit logging (not console)
4. **Enable audit logging** for compliance
5. **Set strict tenant validation** to prevent cross-tenant access
6. **Use different rate limits** for different endpoint types:
   - Auth endpoints: 5 req/min per IP
   - API endpoints: 100 req/min per user
   - Admin endpoints: 50 req/min per user
7. **Monitor rate limit violations** to detect abuse
8. **Regularly cleanup** expired tokens from revocation list
9. **Set appropriate token durations**:
   - Access token: 15 minutes
   - Refresh token: 7 days (with rotation)
10. **Implement token rotation** on refresh
