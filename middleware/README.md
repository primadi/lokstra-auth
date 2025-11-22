# Lokstra Auth Middlewares

HTTP middlewares for Lokstra framework to integrate authentication and authorization.

## Available Middlewares

### 1. Authentication Middleware (`auth.go`)
Verifies JWT tokens and injects identity into request context.

### 2. Permission Middleware (`permission.go`)
Checks if user has required permission(s).

### 3. Role Middleware (`role.go`)
Checks if user has required role(s).

---

## Installation

Middlewares are built using [Lokstra Framework](https://github.com/primadi/lokstra).

```go
import (
    lokstraauth "github.com/primadi/lokstra-auth"
    "github.com/primadi/lokstra-auth/middleware"
    "github.com/primadi/lokstra/core/app"
    "github.com/primadi/lokstra/core/request"
)
```

---

## Quick Start

### 1. Setup Auth Runtime

```go
auth := lokstraauth.NewBuilder().
    WithAuthenticator("basic", basicAuth).
    WithTokenManager(jwtManager).
    WithSubjectResolver(resolver).
    WithIdentityContextBuilder(contextBuilder).
    WithAuthorizer(rbacEvaluator).
    Build()
```

### 2. Create Application

```go
app := app.New()
```

### 3. Add Authentication Middleware

```go
authMiddleware := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
    Auth: auth,
})

// Protected route
app.GET("/profile",
    authMiddleware.Handler(),
    func(c *request.Context) error {
        identity, _ := middleware.GetIdentity(c)
        return c.Resp.Json(identity)
    },
)
```

---

## Middleware Reference

### Authentication Middleware

**Verifies token and injects identity into context.**

```go
// Basic usage
authMiddleware := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
    Auth: auth,
})

app.GET("/protected", authMiddleware.Handler(), handler)
```

**Optional Authentication:**

```go
// Don't fail if no token provided
optionalAuth := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
    Auth:     auth,
    Optional: true, // Allow unauthenticated users
})

app.GET("/public", optionalAuth.Handler(), handler)
```

**Custom Token Extractor:**

```go
// Extract token from cookie instead of header
authMiddleware := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
    Auth: auth,
    TokenExtractor: func(c *request.Context) (string, error) {
        cookie := c.Req.Cookie("access_token")
        if cookie == "" {
            return "", errors.New("no token cookie")
        }
        return cookie, nil
    },
})
```

**Custom Error Handler:**

```go
authMiddleware := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
    Auth: auth,
    ErrorHandler: func(c *request.Context, err error) error {
        c.Resp.WithStatus(401)
        return c.Resp.Json(map[string]any{
            "error": "Authentication failed",
            "details": err.Error(),
        })
    },
})
```

---

### Permission Middleware

**Checks if user has required permission.**

```go
// Single permission
app.GET("/api/users",
    authMiddleware.Handler(),
    middleware.RequirePermission(auth, "read:users"),
    handler,
)

// Or with full config
permMiddleware := middleware.NewPermissionMiddleware(middleware.PermissionMiddlewareConfig{
    Auth:       auth,
    Permission: "delete:users",
})

app.DELETE("/api/users/:id",
    authMiddleware.Handler(),
    permMiddleware.Handler(),
    handler,
)
```

**Any Permission (OR logic):**

```go
// User needs read:code OR write:code
anyPerm := middleware.NewAnyPermissionMiddleware(auth, []string{
    "read:code",
    "write:code",
})

app.GET("/api/code",
    authMiddleware.Handler(),
    anyPerm.Handler(),
    handler,
)
```

**All Permissions (AND logic):**

```go
// User needs BOTH read:users AND write:users
allPerms := middleware.NewAllPermissionsMiddleware(auth, []string{
    "read:users",
    "write:users",
})

app.POST("/api/users",
    authMiddleware.Handler(),
    allPerms.Handler(),
    handler,
)
```

---

### Role Middleware

**Checks if user has required role.**

```go
// Single role
app.GET("/admin/dashboard",
    authMiddleware.Handler(),
    middleware.RequireRole(auth, "admin"),
    handler,
)

// Or with full config
roleMiddleware := middleware.NewRoleMiddleware(middleware.RoleMiddlewareConfig{
    Auth: auth,
    Role: "team-lead",
})

app.POST("/api/deploy",
    authMiddleware.Handler(),
    roleMiddleware.Handler(),
    handler,
)
```

**Any Role (OR logic):**

```go
// User needs admin OR team-lead
anyRole := middleware.NewAnyRoleMiddleware(auth, []string{
    "admin",
    "team-lead",
})

app.POST("/api/deploy",
    authMiddleware.Handler(),
    anyRole.Handler(),
    handler,
)
```

**All Roles (AND logic):**

```go
// User needs BOTH admin AND developer roles
allRoles := middleware.NewAllRolesMiddleware(auth, []string{
    "admin",
    "developer",
})

app.GET("/api/special",
    authMiddleware.Handler(),
    allRoles.Handler(),
    handler,
)
```

---

## Helper Functions

### Get Identity from Context

```go
func handler(c *request.Context) error {
    // Get identity (returns nil if not authenticated)
    identity, ok := middleware.GetIdentity(c)
    if !ok {
        return c.Resp.Json(map[string]any{
            "error": "Not authenticated",
        })
    }

    // Use identity
    userID := identity.Subject.ID
    email := identity.Subject.Principal
    roles := identity.Roles
    permissions := identity.Permissions

    return c.Resp.Json(map[string]any{
        "user_id": userID,
        "email": email,
        "roles": roles,
    })
}
```

### Must Get Identity (Panics if not found)

```go
func handler(c *request.Context) error {
    // Use this ONLY after authentication middleware
    identity := middleware.MustGetIdentity(c)
    
    return c.Resp.Json(identity)
}
```

---

## Middleware Chaining

Middlewares are applied in order:

```go
app.POST("/api/admin/users",
    // 1. Verify token
    authMiddleware.Handler(),
    
    // 2. Check admin role
    middleware.RequireRole(auth, "admin"),
    
    // 3. Check delete permission
    middleware.RequirePermission(auth, "delete:users"),
    
    // 4. Handler
    func(c *request.Context) error {
        // User is authenticated, has admin role, and delete:users permission
        return c.Resp.Json(map[string]any{
            "message": "User deleted",
        })
    },
)
```

---

## Complete Example

See `examples/middleware/main.go` for a complete working example.

**Run the example:**

```bash
go run ./examples/middleware/main.go
```

**Test with curl:**

```bash
# 1. Login
curl -X POST http://localhost:3000/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# Response:
# {
#   "access_token": "eyJhbGc...",
#   "refresh_token": "eyJhbGc...",
#   "user": {...}
# }

# 2. Access protected endpoint
TOKEN="eyJhbGc..."

curl http://localhost:3000/profile \
  -H "Authorization: Bearer $TOKEN"

# 3. Access admin-only endpoint
curl http://localhost:3000/admin/dashboard \
  -H "Authorization: Bearer $TOKEN"

# 4. Access permission-based endpoint
curl http://localhost:3000/api/users \
  -H "Authorization: Bearer $TOKEN"

# 5. Public content (no auth required)
curl http://localhost:3000/public/content
```

---

## Error Responses

### 401 Unauthorized
Missing or invalid token.

```json
{
  "error": "Unauthorized",
  "message": "missing authentication token"
}
```

### 403 Forbidden
User authenticated but lacks required permission/role.

```json
{
  "error": "Forbidden",
  "message": "authorization check failed"
}
```

---

## Best Practices

1. **Always use AuthMiddleware first**
   ```go
   app.GET("/protected",
       authMiddleware.Handler(),  // ✓ First
       permissionMiddleware.Handler(),
       handler,
   )
   ```

2. **Use Optional auth for mixed endpoints**
   ```go
   optionalAuth := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
       Auth:     auth,
       Optional: true,
   })
   ```

3. **Combine role and permission checks**
   ```go
   app.POST("/critical",
       authMiddleware.Handler(),
       middleware.RequireRole(auth, "admin"),
       middleware.RequirePermission(auth, "critical:action"),
       handler,
   )
   ```

4. **Use helpers to access identity**
   ```go
   identity, ok := middleware.GetIdentity(c)
   if !ok {
       return errorResponse
   }
   ```

5. **Custom error handlers for better UX**
   ```go
   authMiddleware := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
       Auth: auth,
       ErrorHandler: customErrorHandler,
   })
   ```

---

## Architecture

```
HTTP Request
    ↓
AuthMiddleware (Layer 2 + 3)
    ├─ Extract token from header
    ├─ Verify token (JWT signature, expiry)
    ├─ Build identity context
    └─ Inject into request context
    ↓
PermissionMiddleware (Layer 4)
    ├─ Get identity from context
    ├─ Check permission via Auth.CheckPermission()
    └─ Allow/Deny
    ↓
RoleMiddleware (Layer 4)
    ├─ Get identity from context
    ├─ Check role via Auth.CheckRole()
    └─ Allow/Deny
    ↓
Handler
    ├─ Get identity via middleware.GetIdentity()
    └─ Process request
```

---

## See Also

- [Layer 1: Credential Examples](../01_credential/)
- [Layer 2: Token Management](../02_token/)
- [Layer 3: Subject Resolution](../03_subject/)
- [Layer 4: Authorization](../04_authz/)
- [Complete Flow Example](../complete/)
- [Lokstra Framework](https://github.com/primadi/lokstra)
