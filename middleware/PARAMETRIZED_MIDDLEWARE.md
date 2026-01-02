# Parametrized Middleware Registration in Lokstra

## Summary

To register middleware with parameters in Lokstra framework, use `RegisterMiddlewareFactory` instead of `RegisterMiddleware`.

## The Pattern

### 1. Create a Middleware Factory Function

A middleware factory accepts `params map[string]any` and returns `func(*request.Context) error`:

```go
func createRoleMiddleware(params map[string]any) func(*request.Context) error {
    role, _ := params["role"].(string)
    if role == "" {
        return forbiddenHandler("missing role parameter")
    }

    return func(c *request.Context) error {
        identity := MustGetIdentity(c)
        if !identity.HasRole(role) {
            c.Resp.WithStatus(403)
            return c.Resp.Json(map[string]any{
                "error":   "Forbidden",
                "message": "missing required role: " + role,
            })
        }
        return c.Next()
    }
}
```

### 2. Register the Factory

Use `RegisterMiddlewareFactory` to register the factory:

```go
func Register() {
    lokstra_registry.RegisterMiddlewareFactory("role", createRoleMiddleware)
    lokstra_registry.RegisterMiddlewareFactory("permission", createPermissionMiddleware)
    lokstra_registry.RegisterMiddlewareFactory("any-role", createAnyRoleMiddleware)
}
```

### 3. Use in Annotations

In your `@Route` or `@RouterService` annotations, specify parameters using the format:

```
middlewareName key=value
```

Examples:

```go
// Single parameter
// @Route "GET /dashboard", middlewares=["auth", "role role=admin"]

// Multiple values (comma-separated)
// @Route "GET /content", middlewares=["auth", "any-role roles=editor,admin"]

// Different middleware types
// @Route "GET /{id}", middlewares=["auth", "permission permission=document:read"]
```

## How It Works

When Lokstra processes the annotation `middlewares=["role role=admin"]`:

1. **Parse**: It splits by space:
   - `"role"` → middleware name (registry lookup key)
   - `"role=admin"` → parameter string

2. **Extract params**: It parses key=value pairs into a map:
   ```go
   map[string]any{"role": "admin"}
   ```

3. **Call factory**: It calls your factory function:
   ```go
   createRoleMiddleware(map[string]any{"role": "admin"})
   ```

4. **Returns handler**: Your factory returns the actual middleware handler that will be used in the chain.

## Common Patterns

### Single Value Parameter

```go
// Factory
func createRoleMiddleware(params map[string]any) func(*request.Context) error {
    role, _ := params["role"].(string)
    // ... use role
}

// Annotation
// @Route "GET /admin", middlewares=["role role=admin"]
```

### Multiple Values (Comma-Separated)

```go
// Factory
func createAnyRoleMiddleware(params map[string]any) func(*request.Context) error {
    rolesStr, _ := params["roles"].(string)
    roles := strings.Split(rolesStr, ",")
    for i := range roles {
        roles[i] = strings.TrimSpace(roles[i])
    }
    // ... use roles
}

// Annotation
// @Route "GET /editor", middlewares=["any-role roles=editor,admin"]
```

### No Parameters

```go
// Factory still needs to accept params even if unused
func createAuthMiddleware(params map[string]any) func(*request.Context) error {
    mw := NewAuthMiddleware(AuthMiddlewareConfig{})
    return mw.Handler()
}

// Annotation - no parameters
// @Route "GET /profile", middlewares=["auth"]
```

## Key Differences

| Aspect | `RegisterMiddleware` | `RegisterMiddlewareFactory` |
|--------|---------------------|----------------------------|
| **Signature** | `(name string, handler request.HandlerFunc)` | `(name string, factory any)` |
| **Purpose** | Register pre-built middleware | Register middleware factory |
| **Parameters** | None (static) | Accepts `map[string]any` from annotation |
| **Use Case** | Simple middlewares (logging, recovery) | Parametrized middlewares (auth, permissions) |

## Complete Example

```go
// middleware/register.go
package middleware

import (
    "strings"
    "github.com/primadi/lokstra/core/request"
    "github.com/primadi/lokstra/lokstra_registry"
)

func Register() {
    // Parametrized middlewares use RegisterMiddlewareFactory
    lokstra_registry.RegisterMiddlewareFactory("role", createRoleMiddleware)
    lokstra_registry.RegisterMiddlewareFactory("permission", createPermissionMiddleware)
}

func createRoleMiddleware(params map[string]any) func(*request.Context) error {
    role, _ := params["role"].(string)
    if role == "" {
        return forbiddenHandler("missing role parameter")
    }

    return func(c *request.Context) error {
        identity := MustGetIdentity(c)
        if !identity.HasRole(role) {
            c.Resp.WithStatus(403)
            return c.Resp.Json(map[string]any{
                "error":   "Forbidden",
                "message": "missing required role: " + role,
            })
        }
        return c.Next()
    }
}

func createPermissionMiddleware(params map[string]any) func(*request.Context) error {
    permission, _ := params["permission"].(string)
    if permission == "" {
        return forbiddenHandler("missing permission parameter")
    }

    return func(c *request.Context) error {
        identity := MustGetIdentity(c)
        
        checker := rbac.NewPermissionChecker()
        hasPermission, _ := checker.HasPermission(c, identity, permission)
        
        if !hasPermission {
            c.Resp.WithStatus(403)
            return c.Resp.Json(map[string]any{
                "error":   "Forbidden",
                "message": "missing required permission: " + permission,
            })
        }
        return c.Next()
    }
}

func forbiddenHandler(message string) func(*request.Context) error {
    return func(c *request.Context) error {
        c.Resp.WithStatus(403)
        return c.Resp.Json(map[string]any{
            "error":   "Forbidden",
            "message": message,
        })
    }
}
```

```go
// handlers/demo.go
package handlers

// AdminService provides admin-only endpoints
// @RouterService name="admin-service", prefix="/api/admin", middlewares=["recovery", "request_logger"]
type AdminService struct{}

// GetDashboard returns admin dashboard information
// @Route "GET /dashboard", middlewares=["auth", "role role=admin"]
func (s *AdminService) GetDashboard(ctx *request.Context) (map[string]any, error) {
    identity := authmiddleware.MustGetIdentity(ctx)
    return map[string]any{
        "message": "Welcome, admin!",
        "user_id": identity.Subject.ID,
        "roles":   identity.Roles,
    }, nil
}

// EditorService provides editor endpoints
// @RouterService name="editor-service", prefix="/api/editor", middlewares=["recovery", "request_logger"]
type EditorService struct{}

// GetContent returns editor content
// @Route "GET /content", middlewares=["auth", "any-role roles=editor,admin"]
func (s *EditorService) GetContent(ctx *request.Context) (map[string]any, error) {
    identity := authmiddleware.MustGetIdentity(ctx)
    return map[string]any{
        "message": "You have editor access",
        "user_id": identity.Subject.ID,
        "roles":   identity.Roles,
    }, nil
}

// DocumentService provides document management endpoints
// @RouterService name="document-service", prefix="/api/documents", middlewares=["recovery", "request_logger"]
type DocumentService struct{}

// GetDocument reads a document (requires document:read permission)
// @Route "GET /{id}", middlewares=["auth", "permission permission=document:read"]
func (s *DocumentService) GetDocument(ctx *request.Context) (map[string]any, error) {
    identity := authmiddleware.MustGetIdentity(ctx)
    documentID := ctx.Req.PathParam("id", "")

    return map[string]any{
        "message":     "Document read access granted",
        "document_id": documentID,
        "user_id":     identity.Subject.ID,
    }, nil
}
```

## Troubleshooting

### Error: "cannot use ... as request.HandlerFunc"

**Cause**: Using `RegisterMiddleware` instead of `RegisterMiddlewareFactory`.

**Solution**: Change to `RegisterMiddlewareFactory`:

```go
// ❌ Wrong
lokstra_registry.RegisterMiddleware("role", createRoleMiddleware)

// ✅ Correct
lokstra_registry.RegisterMiddlewareFactory("role", createRoleMiddleware)
```

### Error: "c.Resp.Status undefined"

**Cause**: Method chaining changed in newer Lokstra versions.

**Solution**: Use `WithStatus()` separately:

```go
// ❌ Wrong
return c.Resp.Status(403).Json(...)

// ✅ Correct
c.Resp.WithStatus(403)
return c.Resp.Json(...)
```

### Missing Parameters

If parameters are missing in your factory:

```go
func createRoleMiddleware(params map[string]any) func(*request.Context) error {
    role, ok := params["role"].(string)
    if !ok || role == "" {
        // Return a middleware that always returns error
        return forbiddenHandler("missing role parameter")
    }
    // ... rest of implementation
}
```

## References

- Lokstra Registry Docs: `go doc github.com/primadi/lokstra/lokstra_registry`
- Middleware Factory Pattern: `go doc github.com/primadi/lokstra/lokstra_registry RegisterMiddlewareFactory`
- Examples: `examples/01_deployment/handlers/demo.go`
