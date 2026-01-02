# Middleware Guide

## Overview

Middleware adalah layer yang berjalan **sebelum** handler untuk melakukan verifikasi, validasi, dan enrichment pada request.

## Using Middlewares with @RouterService Annotation

**Recommended approach:** Declare middlewares in `@RouterService` annotation for automatic protection.

```go
// Protected endpoints - require authentication
// @RouterService name="tenant-service", prefix="/api/tenants", middlewares=["recovery", "request_logger", "auth"]
type TenantService struct {
    // ... service implementation
}

// Public endpoints - no authentication needed (login, register, etc.)
// @RouterService name="basic-auth-service", prefix="/api/auth/basic", middlewares=["recovery", "request_logger"]
type BasicAuthService struct {
    // ... authentication implementation
}
```

**Middleware execution order:** Left to right
1. `recovery` - Panic recovery
2. `request_logger` - HTTP request logging
3. `auth` - Token verification (for protected endpoints only)
4. Additional middlewares as needed: `permission`, `role`, `tenant`, etc.

## Middleware Security Guidelines

### ✅ MUST Use Auth Middleware

These services **require authentication** and should include `"auth"` in middlewares:

- **Core Services:**
  - `tenant-service` - Tenant management
  - `user-service` - User management
  - `app-service` - Application management
  - `branch-service` - Branch management
  - `app-key-service` - API key management
  - `credential-config-service` - Credential configuration

- **Subject Services:**
  - `role-service` - Role management
  - `permission-service` - Permission management
  - `user-role-service` - User-role assignment
  - `user-permission-service` - User-permission assignment
  - `role-permission-service` - Role-permission assignment

- **Authz Services:**
  - `policy-service` - Policy management

### ❌ MUST NOT Use Auth Middleware

These services are **public endpoints** and should NOT include `"auth"`:

- **Credential Services (Login Endpoints):**
  - `basic-auth-service` - Username/password login
  - `oauth2-auth-service` - OAuth2 authentication
  - `apikey-auth-service` - API key authentication
  - `passkey-auth-service` - WebAuthn/Passkey authentication
  - `passwordless-auth-service` - Magic link authentication

**Why?** These are authentication endpoints themselves - users cannot be authenticated before they authenticate!

## Example Annotations

### Protected Endpoint Example
```go
package application

import "github.com/primadi/lokstra/core/request"

// @RouterService name="tenant-service", prefix="${api-auth-prefix:/api/auth}/core/tenants", middlewares=["recovery", "request_logger", "auth"]
type TenantService struct {
    repo TenantRepository
}

// @Route method="POST", path=""
func (s *TenantService) CreateTenant(c *request.Context) error {
    // Only authenticated users can reach here
    // Token already verified by auth middleware
    
    // Get identity from context
    identity, ok := middleware.GetIdentity(c)
    if !ok {
        return c.Resp.WithStatus(401).Json(map[string]string{"error": "unauthorized"})
    }
    
    // Business logic...
    return c.Resp.Json(tenant)
}
```

### Public Endpoint Example
```go
package application

import "github.com/primadi/lokstra/core/request"

// @RouterService name="basic-auth-service", prefix="${api-auth-prefix:/api/auth}/cred/basic", middlewares=["recovery", "request_logger"]
type BasicAuthService struct {
    authenticator BasicAuthenticator
}

// @Route method="POST", path="/login"
func (s *BasicAuthService) Login(c *request.Context) error {
    // Public endpoint - anyone can access
    // No authentication required
    
    var req LoginRequest
    if err := c.Bind(&req); err != nil {
        return err
    }
    
    // Verify credentials and generate token
    token, err := s.authenticator.Authenticate(req.Username, req.Password)
    if err != nil {
        return c.Resp.WithStatus(401).Json(map[string]string{"error": "invalid credentials"})
    }
    
    return c.Resp.Json(map[string]string{"token": token})
}
```

## Middleware that Harus Dibuat

### ✅ Already Exists

1. **AuthMiddleware** (`middleware/auth.go`)
   - Verify JWT token
   - Build identity context
   - Inject identity into request

2. **RoleMiddleware** (`middleware/role.go`)
   - Check single role
   - Check any role
   - Check all roles

3. **PermissionMiddleware** (`middleware/permission.go`)
   - Check single permission
   - Check any permission
   - Check all permissions

### 🔲 Need to Create

4. **TenantMiddleware** - Validate tenant context
5. **RateLimitMiddleware** - Rate limiting per tenant/user
6. **AuditMiddleware** - Log all operations
7. **ResourceOwnershipMiddleware** - Check resource ownership

## 1. AuthMiddleware (Already Exists)

**Purpose:** Verify token and inject identity

**Usage:**
```go
// Required authentication
router.Use(authMiddleware.Handler())

// Optional authentication (for public + private routes)
optionalAuth := NewAuthMiddleware(AuthMiddlewareConfig{
    Auth:     authRuntime,
    Optional: true,
})
router.Use(optionalAuth.Handler())
```

**Flow:**
```
Request → Extract token from header
  ↓
Verify token (token)
  ↓
Build identity context (subject)
  ↓
Inject into request context
  ↓
Continue to next handler
```

**Response on failure:**
```json
{
  "error": "Unauthorized",
  "message": "invalid token"
}
```

## 2. RoleMiddleware (Already Exists)

**Purpose:** Check if user has required role(s)

**Usage:**
```go
// Require single role
router.GET("/admin/users",
    authMiddleware.Handler(),
    RequireRole(authRuntime, "admin"),
    handler,
)

// Require any of multiple roles
router.GET("/dashboard",
    authMiddleware.Handler(),
    NewAnyRoleMiddleware(authRuntime, []string{"admin", "manager"}).Handler(),
    handler,
)

// Require all roles
router.POST("/sensitive-operation",
    authMiddleware.Handler(),
    NewAllRolesMiddleware(authRuntime, []string{"admin", "auditor"}).Handler(),
    handler,
)
```

**Response on failure:**
```json
{
  "error": "Forbidden",
  "message": "insufficient privileges"
}
```

## 3. PermissionMiddleware (Already Exists)

**Purpose:** Check if user has required permission(s)

**Usage:**
```go
// Require single permission
router.GET("/api/users",
    authMiddleware.Handler(),
    RequirePermission(authRuntime, "read:users"),
    handler,
)

// Require any permission
router.POST("/api/documents",
    authMiddleware.Handler(),
    NewAnyPermissionMiddleware(authRuntime, []string{"create:documents", "admin:all"}).Handler(),
    handler,
)

// Require all permissions
router.DELETE("/api/critical-data",
    authMiddleware.Handler(),
    NewAllPermissionsMiddleware(authRuntime, []string{"delete:data", "admin:confirm"}).Handler(),
    handler,
)
```

## 4. TenantMiddleware (Need to Create)

**Purpose:** 
- Validate tenant header matches token claims
- Prevent cross-tenant access
- Inject tenant context

**Implementation:**
```go
package middleware

import (
    "errors"
    "github.com/primadi/lokstra/core/request"
    subject "github.com/primadi/lokstra-auth/rbac"
)

type TenantMiddleware struct {
    strict bool // If true, require header to match token
}

func (m *TenantMiddleware) Handler() func(c *request.Context) error {
    return func(c *request.Context) error {
        // Get tenant from header
        headerTenantID := c.R.Header.Get("X-Tenant-ID")
        
        // Get identity from context (set by AuthMiddleware)
        identity, ok := GetIdentity(c)
        if !ok {
            return errors.New("identity not found")
        }
        
        // Validate tenant matches
        if m.strict && headerTenantID != "" && headerTenantID != identity.TenantID {
            c.Resp.WithStatus(403)
            return c.Resp.Json(map[string]any{
                "error": "Forbidden",
                "message": "tenant mismatch",
            })
        }
        
        // Inject tenant ID for convenience
        c.Set("tenant_id", identity.TenantID)
        
        return c.Next()
    }
}
```

**Usage:**
```go
// Strict mode: header must match token
router.Use(
    authMiddleware.Handler(),
    NewTenantMiddleware(true).Handler(),
)
```

## 5. RateLimitMiddleware (Need to Create)

**Purpose:**
- Limit requests per tenant/user
- Prevent abuse
- Different limits for different tiers

**Implementation Strategy:**
```go
package middleware

import (
    "fmt"
    "time"
    "github.com/primadi/lokstra/core/request"
)

type RateLimiter interface {
    Allow(key string) bool
    Remaining(key string) int
}

type RateLimitMiddleware struct {
    limiter  RateLimiter
    keyFunc  func(c *request.Context) string
    limit    int
    window   time.Duration
}

func (m *RateLimitMiddleware) Handler() func(c *request.Context) error {
    return func(c *request.Context) error {
        // Get rate limit key (e.g., "tenant:acme-corp:user:usr_123")
        key := m.keyFunc(c)
        
        // Check limit
        if !m.limiter.Allow(key) {
            c.Resp.WithStatus(429)
            c.Resp.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", m.limit))
            c.Resp.Header().Set("X-RateLimit-Remaining", "0")
            c.Resp.Header().Set("Retry-After", fmt.Sprintf("%d", int(m.window.Seconds())))
            
            return c.Resp.Json(map[string]any{
                "error": "Too Many Requests",
                "message": "rate limit exceeded",
            })
        }
        
        // Add rate limit headers
        remaining := m.limiter.Remaining(key)
        c.Resp.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", m.limit))
        c.Resp.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
        
        return c.Next()
    }
}

// Rate limit key functions
func PerTenantKey(c *request.Context) string {
    identity, _ := GetIdentity(c)
    return fmt.Sprintf("tenant:%s", identity.TenantID)
}

func PerUserKey(c *request.Context) string {
    identity, _ := GetIdentity(c)
    return fmt.Sprintf("tenant:%s:user:%s", identity.TenantID, identity.Subject.ID)
}
```

**Usage:**
```go
// 100 requests per minute per tenant
tenantRateLimit := NewRateLimitMiddleware(RateLimitConfig{
    Limiter: redisLimiter,
    KeyFunc: PerTenantKey,
    Limit:   100,
    Window:  time.Minute,
})

router.Use(
    authMiddleware.Handler(),
    tenantRateLimit.Handler(),
)
```

## 6. AuditMiddleware (Need to Create)

**Purpose:**
- Log all operations with tenant context
- Track who did what, when
- Compliance and security monitoring

**Implementation:**
```go
package middleware

import (
    "time"
    "github.com/primadi/lokstra/core/request"
    subject "github.com/primadi/lokstra-auth/rbac"
)

type AuditLogger interface {
    Log(entry *AuditEntry) error
}

type AuditEntry struct {
    Timestamp  time.Time
    TenantID   string
    AppID      string
    UserID     string
    Username   string
    Action     string // HTTP method + path
    Resource   string
    IPAddress  string
    UserAgent  string
    StatusCode int
    Duration   time.Duration
    Metadata   map[string]any
}

type AuditMiddleware struct {
    logger AuditLogger
    skip   func(c *request.Context) bool
}

func (m *AuditMiddleware) Handler() func(c *request.Context) error {
    return func(c *request.Context) error {
        // Skip if configured
        if m.skip != nil && m.skip(c) {
            return c.Next()
        }
        
        start := time.Now()
        
        // Get identity (if available)
        var tenantID, appID, userID, username string
        if identity, ok := GetIdentity(c); ok {
            tenantID = identity.TenantID
            appID = identity.AppID
            userID = identity.Subject.ID
            username = identity.Subject.Principal
        }
        
        // Continue to handler
        err := c.Next()
        
        // Log after handler completes
        duration := time.Since(start)
        
        entry := &AuditEntry{
            Timestamp:  start,
            TenantID:   tenantID,
            AppID:      appID,
            UserID:     userID,
            Username:   username,
            Action:     fmt.Sprintf("%s %s", c.R.Method, c.R.URL.Path),
            Resource:   c.R.URL.Path,
            IPAddress:  c.R.RemoteAddr,
            UserAgent:  c.R.UserAgent(),
            StatusCode: c.Resp.Status(),
            Duration:   duration,
        }
        
        m.logger.Log(entry)
        
        return err
    }
}
```

**Usage:**
```go
auditMiddleware := NewAuditMiddleware(AuditConfig{
    Logger: postgresAuditLogger,
    Skip: func(c *request.Context) bool {
        // Skip health checks
        return c.R.URL.Path == "/health"
    },
})

router.Use(auditMiddleware.Handler())
```

## 7. ResourceOwnershipMiddleware (Need to Create)

**Purpose:**
- Check if user owns the resource
- Prevent cross-tenant or cross-user access
- Works with path parameters

**Implementation:**
```go
package middleware

import (
    "errors"
    "github.com/primadi/lokstra/core/request"
)

type ResourceOwnerChecker interface {
    // IsOwner checks if user owns the resource
    IsOwner(tenantID, userID, resourceType, resourceID string) (bool, error)
}

type ResourceOwnershipMiddleware struct {
    checker      ResourceOwnerChecker
    resourceType string
    paramName    string // Path param name (e.g., "id", "orderId")
}

func (m *ResourceOwnershipMiddleware) Handler() func(c *request.Context) error {
    return func(c *request.Context) error {
        identity, ok := GetIdentity(c)
        if !ok {
            return errors.New("identity not found")
        }
        
        // Get resource ID from path parameter
        resourceID := c.Param(m.paramName)
        if resourceID == "" {
            return errors.New("resource ID not found")
        }
        
        // Check ownership
        isOwner, err := m.checker.IsOwner(
            identity.TenantID,
            identity.Subject.ID,
            m.resourceType,
            resourceID,
        )
        
        if err != nil {
            return err
        }
        
        if !isOwner {
            c.Resp.WithStatus(403)
            return c.Resp.Json(map[string]any{
                "error": "Forbidden",
                "message": "you do not own this resource",
            })
        }
        
        return c.Next()
    }
}
```

**Usage:**
```go
// Check order ownership
router.GET("/api/orders/{id}",
    authMiddleware.Handler(),
    NewResourceOwnershipMiddleware(ResourceOwnershipConfig{
        Checker:      orderOwnerChecker,
        ResourceType: "order",
        ParamName:    "id",
    }).Handler(),
    getOrderHandler,
)
```

## Middleware Combinations

### Public Route (No Auth)
```go
router.GET("/health", healthHandler)
router.GET("/docs", docsHandler)
```

### Authenticated Route (Any User)
```go
router.GET("/api/profile",
    authMiddleware.Handler(),
    getProfileHandler,
)
```

### Role-Based Route
```go
router.GET("/admin/users",
    authMiddleware.Handler(),
    RequireRole(authRuntime, "admin"),
    listUsersHandler,
)
```

### Permission-Based Route
```go
router.POST("/api/documents",
    authMiddleware.Handler(),
    RequirePermission(authRuntime, "create:documents"),
    createDocumentHandler,
)
```

### Full Stack (Auth + Tenant + Rate Limit + Audit)
```go
router.DELETE("/api/orders/{id}",
    auditMiddleware.Handler(),
    authMiddleware.Handler(),
    tenantMiddleware.Handler(),
    rateLimitMiddleware.Handler(),
    RequirePermission(authRuntime, "delete:orders"),
    resourceOwnershipMiddleware.Handler(),
    deleteOrderHandler,
)
```

### Conditional Auth (Public + Private)
```go
// Route works for both authenticated and anonymous users
router.GET("/api/products",
    optionalAuthMiddleware.Handler(), // Optional: true
    listProductsHandler, // Handler checks if identity exists
)

// Handler implementation:
func listProductsHandler(c *request.Context) error {
    identity, authenticated := GetIdentity(c)
    
    if authenticated {
        // Return personalized product list
        return getPersonalizedProducts(identity)
    } else {
        // Return public product list
        return getPublicProducts()
    }
}
```

## Middleware Order (Very Important!)

**Correct order:**
```go
router.Use(
    recoveryMiddleware,        // 1. Catch panics
    loggingMiddleware,         // 2. Log all requests
    corsMiddleware,            // 3. Handle CORS
    auditMiddleware,           // 4. Audit before auth (logs failed auth too)
    authMiddleware,            // 5. Authenticate user
    tenantMiddleware,          // 6. Validate tenant context
    rateLimitMiddleware,       // 7. Rate limit after auth (per user/tenant)
    roleMiddleware,            // 8. Check roles (uses identity from auth)
    permissionMiddleware,      // 9. Check permissions (uses identity)
)
```

**Why this order?**
1. Recovery first (catch all panics)
2. Logging second (log everything including errors)
3. CORS third (browser preflight before auth)
4. Audit fourth (log auth failures too)
5. Auth fifth (verify who they are)
6. Tenant sixth (validate scope)
7. Rate limit seventh (after knowing who they are)
8-9. Authorization last (check what they can do)

## Summary

| Middleware | Status | Purpose | When to Use |
|-----------|--------|---------|-------------|
| **AuthMiddleware** | ✅ Exists | Verify token, build identity | All protected routes |
| **RoleMiddleware** | ✅ Exists | Check roles | Admin panels, role-specific features |
| **PermissionMiddleware** | ✅ Exists | Check permissions | Fine-grained access control |
| **TenantMiddleware** | 🔲 Create | Validate tenant context | All tenant-scoped operations |
| **RateLimitMiddleware** | 🔲 Create | Rate limiting | Public APIs, prevent abuse |
| **AuditMiddleware** | 🔲 Create | Log operations | Compliance, security monitoring |
| **ResourceOwnershipMiddleware** | 🔲 Create | Check resource ownership | User-specific resources |

## Next Steps

1. ✅ Use existing Auth/Role/Permission middleware
2. 🔲 Create TenantMiddleware
3. 🔲 Create RateLimitMiddleware (with Redis)
4. 🔲 Create AuditMiddleware (with PostgreSQL)
5. 🔲 Create ResourceOwnershipMiddleware
6. 🔲 Add middleware tests
7. 🔲 Document middleware usage in examples
