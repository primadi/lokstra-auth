# Middleware Documentation

Complete middleware infrastructure for Lokstra-Auth with declarative authorization support.

## Overview

The middleware package provides authentication and authorization middlewares that can be used with **@RouterService** annotations for declarative security configuration.

## Quick Reference

| Middleware | Parameters | Use Case | Logic |
|------------|------------|----------|-------|
| `auth` | - | Verify token | Required for all protected endpoints |
| `role` | `role=X` | Single role | User must have role X |
| `any-role` | `roles=X,Y,Z` | Multiple roles (OR) | User must have at least one role |
| `all-roles` | `roles=X,Y` | Multiple roles (AND) | User must have all roles |
| `permission` | `permission=X` | Single permission | User must have permission X |
| `any-permission` | `permissions=X,Y,Z` | Multiple permissions (OR) | User must have at least one permission |
| `all-permissions` | `permissions=X,Y` | Multiple permissions (AND) | User must have all permissions |

---

## Available Middlewares

### 1. Authentication Middleware

#### `auth`
Verifies JWT token and injects identity into request context.

**Usage:**
```go
// @Route "GET /profile", middlewares=["auth"]
```

**Behavior:**
- Extracts token from `Authorization: Bearer <token>` header
- Verifies token using Auth runtime
- Injects `IdentityContext` into request context
- Returns 401 if token is missing or invalid

---

### 2. Role-Based Middlewares

#### `role`
Requires a specific role.

**Usage:**
```go
// @Route "GET /admin/dashboard", middlewares=["auth", "role role=admin"]
```

**Parameters:**
- `role`: The required role (e.g., "admin", "editor")

**Behavior:**
- Checks if user has the specified role
- Returns 403 if user doesn't have the role

---

#### `any-role`
Requires **at least one** of the specified roles (OR logic).

**Usage:**
```go
// @Route "GET /content", middlewares=["auth", "any-role roles=admin,editor,manager"]
```

**Parameters:**
- `roles`: Comma-separated list of roles (e.g., "admin,editor,manager")

**Behavior:**
- Checks if user has any of the specified roles
- Returns 403 if user has none of the roles

---

#### `all-roles`
Requires **all** of the specified roles (AND logic).

**Usage:**
```go
// @Route "GET /sensitive", middlewares=["auth", "all-roles roles=admin,security-officer"]
```

**Parameters:**
- `roles`: Comma-separated list of roles

**Behavior:**
- Checks if user has all of the specified roles
- Returns 403 if user is missing any role

---

### 3. Permission-Based Middlewares

#### `permission`
Requires a specific permission.

**Usage:**
```go
// @Route "GET /documents/{id}", middlewares=["auth", "permission permission=document:read"]
```

**Parameters:**
- `permission`: The required permission (e.g., "document:read", "user:write")

**Behavior:**
- Checks if user has the specified permission
- Uses RBAC evaluator for role→permission mapping
- Returns 403 if user doesn't have the permission

---

#### `any-permission`
Requires **at least one** of the specified permissions (OR logic).

**Usage:**
```go
// @Route "GET /reports", middlewares=["auth", "any-permission permissions=report:read,report:export"]
```

**Parameters:**
- `permissions`: Comma-separated list of permissions

**Behavior:**
- Checks if user has any of the specified permissions
- Returns 403 if user has none of the permissions

---

#### `all-permissions`
Requires **all** of the specified permissions (AND logic).

**Usage:**
```go
// @Route "DELETE /documents/{id}", middlewares=["auth", "all-permissions permissions=document:read,document:delete"]
```

**Parameters:**
- `permissions`: Comma-separated list of permissions

**Behavior:**
- Checks if user has all of the specified permissions
- Returns 403 if user is missing any permission

---

## Installation & Registration

### Step 1: Import Package

```go
import (
    lokstraauth "github.com/primadi/lokstra-auth"
    "github.com/primadi/lokstra-auth/middleware"
)
```

### Step 2: Register Middlewares

**Important:** Call `middleware.Register()` in your application's `main()` before starting the server:

```go
func main() {
    // Register all middlewares with Lokstra framework
    middleware.Register()
    
    // ... rest of initialization
    // Create auth runtime
    // Create app
    // Start server
}
```

### What Gets Registered

The `Register()` function registers all available middlewares:

```go
func Register() {
    // Auth middleware (no params)
    lokstra_registry.RegisterMiddlewareFactory("auth", createAuthMiddleware)

    // Role middlewares
    lokstra_registry.RegisterMiddlewareFactory("role", createRoleMiddleware)
    lokstra_registry.RegisterMiddlewareFactory("any-role", createAnyRoleMiddleware)
    lokstra_registry.RegisterMiddlewareFactory("all-roles", createAllRolesMiddleware)

    // Permission middlewares
    lokstra_registry.RegisterMiddlewareFactory("permission", createPermissionMiddleware)
    lokstra_registry.RegisterMiddlewareFactory("any-permission", createAnyPermissionMiddleware)
    lokstra_registry.RegisterMiddlewareFactory("all-permissions", createAllPermissionsMiddleware)
}
```

---

## Usage with @RouterService

### Basic Example

```go
package handlers

// @RouterService name="demo-service", prefix="/api"
type DemoHandler struct {
    // ... dependencies
}

// Public endpoint - no auth required
// @Route "GET /public"
func (h *DemoHandler) Public(c *request.Context) error {
    return c.Resp.Json(map[string]any{"message": "public"})
}

// Protected endpoint - auth required
// @Route "GET /profile", middlewares=["auth"]
func (h *DemoHandler) Profile(c *request.Context) error {
    identity := middleware.MustGetIdentity(c)
    return c.Resp.Json(identity)
}

// Admin only endpoint
// @Route "GET /admin/dashboard", middlewares=["auth", "role role=admin"]
func (h *DemoHandler) AdminDashboard(c *request.Context) error {
    return c.Resp.Json(map[string]any{"message": "admin dashboard"})
}

// Requires specific permission
// @Route "GET /documents/{id}", middlewares=["auth", "permission permission=document:read"]
func (h *DemoHandler) GetDocument(c *request.Context) error {
    return c.Resp.Json(map[string]any{"message": "document content"})
}
```

### Advanced Examples

#### Multiple Roles (OR Logic)
```go
// @Route "GET /content", middlewares=["auth", "any-role roles=admin,editor,author"]
func (h *DemoHandler) ManageContent(c *request.Context) error {
    // User must be admin OR editor OR author
}
```

#### Multiple Roles (AND Logic)
```go
// @Route "DELETE /critical", middlewares=["auth", "all-roles roles=admin,security-officer"]
func (h *DemoHandler) DeleteCritical(c *request.Context) error {
    // User must be BOTH admin AND security-officer
}
```

#### Multiple Permissions (OR Logic)
```go
// @Route "GET /reports", middlewares=["auth", "any-permission permissions=report:read,report:export,report:share"]
func (h *DemoHandler) AccessReport(c *request.Context) error {
    // User must have at least one of: report:read, report:export, report:share
}
```

#### Multiple Permissions (AND Logic)
```go
// @Route "PUT /settings", middlewares=["auth", "all-permissions permissions=settings:read,settings:write"]
func (h *DemoHandler) UpdateSettings(c *request.Context) error {
    // User must have BOTH settings:read AND settings:write
}
```

#### Combined Role and Permission
```go
// @Route "DELETE /users/{id}", middlewares=["auth", "role role=admin", "permission permission=user:delete"]
func (h *DemoHandler) DeleteUser(c *request.Context) error {
    // User must be admin AND have user:delete permission
}
```

---

## Middleware Factory Pattern

All parametrized middlewares use the **factory pattern** for registration:

### How It Works

1. **Annotation Parsing**: When Lokstra sees `"role role=admin"`:
   - Name: `"role"`
   - Params: `map[string]any{"role": "admin"}`

2. **Factory Lookup**: Lokstra looks up the factory by name (`"role"`)

3. **Middleware Creation**: Factory is called with params to create the actual middleware handler

4. **Request Processing**: The returned handler is used to process requests

### Factory Function Signature

```go
func(params map[string]any) func(*request.Context) error
```

### Example Factory

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

---

## Error Responses

All middlewares return standardized error responses:

### 401 Unauthorized
When authentication fails:
```json
{
  "error": "Unauthorized",
  "message": "missing or invalid token"
}
```

### 403 Forbidden
When authorization fails:

**Missing role:**
```json
{
  "error": "Forbidden",
  "message": "missing required role: admin"
}
```

**Missing permission:**
```json
{
  "error": "Forbidden",
  "message": "missing required permission: document:read"
}
```

**Missing any of roles:**
```json
{
  "error": "Forbidden",
  "message": "missing any of required roles: admin,editor,manager"
}
```

**Missing all roles:**
```json
{
  "error": "Forbidden",
  "message": "missing all required roles: admin,security-officer"
}
```

---

## Middleware Chaining

Middlewares are executed in the order specified in the annotation:

```go
// @Route "POST /sensitive", middlewares=["auth", "role role=admin", "permission permission=system:write"]
```

Execution order:
1. `auth` - Verify token and inject identity
2. `role role=admin` - Check for admin role
3. `permission permission=system:write` - Check for system:write permission
4. Handler - Execute business logic

**Important:** Always place `auth` middleware first, as other middlewares depend on the identity context.

---

## Multi-Tenant Isolation

All middlewares respect **multi-tenant isolation**:

- Identity providers use keys: `{tenantID}:{appID}:{subjectID}`
- RBAC evaluator uses keys: `{tenantID}:{appID}:{role}`
- Users can only access resources within their tenant

**Example:**
```
User: tenant1:app1:alice
Roles: tenant1:app1:admin → [document:read, document:write]

✅ Can access: document:read in tenant1
❌ Cannot access: document:read in tenant2 (different tenant)
```

---

## Best Practices

### 1. Use Role Middlewares for Simple Checks
```go
// ✅ Good for simple role checks
// @Route "GET /admin/dashboard", middlewares=["auth", "role role=admin"]
```

### 2. Use Permission Middlewares for Fine-Grained Control
```go
// ✅ Good for resource-level permissions
// @Route "GET /documents/{id}", middlewares=["auth", "permission permission=document:read"]
```

### 3. Combine Role and Permission for Complex Requirements
```go
// ✅ Good for complex authorization
// @Route "DELETE /users/{id}", middlewares=["auth", "role role=admin", "all-permissions permissions=user:read,user:delete"]
```

### 4. Use any-role for Flexible Access
```go
// ✅ Good for multi-role access
// @Route "GET /content", middlewares=["auth", "any-role roles=editor,author,admin"]
```

### 5. Avoid Over-Specification
```go
// ❌ Bad - too restrictive
// @Route "GET /profile", middlewares=["auth", "all-roles roles=user,member,verified,active"]

// ✅ Good - just verify authentication
// @Route "GET /profile", middlewares=["auth"]
```

---

## Extending Middleware

To add custom middleware:

### 1. Create the Factory Function
```go
func createCustomMiddleware(params map[string]any) func(*request.Context) error {
    someParam, _ := params["param"].(string)
    
    return func(c *request.Context) error {
        // Your custom logic
        return c.Next()
    }
}
```

### 2. Register in `Register()`
```go
func Register() {
    // ... existing registrations
    lokstra_registry.RegisterMiddlewareFactory("custom", createCustomMiddleware)
}
```

### 3. Use in Annotations
```go
// @Route "GET /custom", middlewares=["auth", "custom param=value"]
```

---

## Testing

Test HTTP files are available in `examples/01_deployment/http-tests/`:

- `01-public.http` - Public endpoint (no auth)
- `02-login.http` - Login to get token
- `07-complete-flow.http` - Complete flow testing all middleware types

Run tests using REST Client extension in VS Code.

---

## Helper Functions

### GetIdentity
Retrieves identity from request context:

```go
identity, ok := middleware.GetIdentity(c)
if !ok {
    return errors.New("not authenticated")
}
```

### MustGetIdentity
Retrieves identity or panics (use in middleware handlers):

```go
identity := middleware.MustGetIdentity(c)
// Will panic if identity not found
```

---

## See Also

- [COMPLETE_FLOW.md](../COMPLETE_FLOW.md) - Complete authentication & authorization flow
- [IDENTITY_AUTHZ_IMPLEMENTATION.md](../IDENTITY_AUTHZ_IMPLEMENTATION.md) - Identity & authorization layer implementation
- [examples/01_deployment/](../examples/01_deployment/) - Demo application with all middleware examples
