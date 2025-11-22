# Migration Guide: Multi-Tenant Architecture

This guide helps you migrate existing `lokstra-auth` code to the new multi-tenant architecture.

## Table of Contents
1. [Overview of Changes](#overview-of-changes)
2. [Breaking Changes](#breaking-changes)
3. [Layer-by-Layer Migration](#layer-by-layer-migration)
4. [Examples](#examples)
5. [Common Patterns](#common-patterns)

## Overview of Changes

The multi-tenant update adds tenant and app isolation across all layers:

- **Credential Layer**: Authenticators now require `AuthContext`
- **Token Layer**: Token managers validate tenant/app IDs
- **Subject Layer**: Resolvers and builders are tenant-scoped
- **Authorization Layer**: Evaluators enforce tenant/app boundaries
- **Service Layer**: NEW - Manage tenants, apps, and users

## Breaking Changes

### 1. Credential Layer - Authenticate Method

**Before:**
```go
result, err := authenticator.Authenticate(ctx, credentials)
```

**After:**
```go
authCtx := &credential.AuthContext{
    TenantID: "acme-corp",
    AppID:    "web-portal",
}
result, err := authenticator.Authenticate(ctx, authCtx, credentials)
```

**Affected Methods:**
- `basic.Authenticator.Authenticate()`
- `apikey.Authenticator.Authenticate()`
- `oauth2.Authenticator.Authenticate()`
- `passwordless.Authenticator.Authenticate()`
- `passkey.Authenticator.Authenticate()`

---

### 2. API Key Generation

**Before:**
```go
key, err := auth.GenerateKey(
    ctx,
    apikey.KeyTypeProduction,
    apikey.EnvironmentProduction,
    "service-account",
    "Backend Service",
    []string{"read:users", "write:data"},
    nil,
)
```

**After:**
```go
key, err := auth.GenerateKey(
    ctx,
    "acme-corp",           // Tenant ID
    "api-service",         // App ID
    apikey.KeyTypeProduction,
    apikey.EnvironmentProduction,
    "service-account",
    "Backend Service",
    []string{"read:users", "write:data"},
    nil,
)
```

---

### 3. Passwordless Authentication

**Before:**
```go
// UserResolver interface
type UserResolver interface {
    ResolveByEmail(ctx context.Context, email string) (string, map[string]any, error)
    ResolveByPhone(ctx context.Context, phone string) (string, map[string]any, error)
}

// Initiating magic link
err := auth.InitiateMagicLink(ctx, email, userID, redirectURL)

// Initiating OTP
err := auth.InitiateOTP(ctx, email, userID)
```

**After:**
```go
// UserResolver interface - added tenantID parameter
type UserResolver interface {
    ResolveByEmail(ctx context.Context, tenantID string, email string) (string, map[string]any, error)
    ResolveByPhone(ctx context.Context, tenantID string, phone string) (string, map[string]any, error)
}

// Initiating magic link - added tenantID
err := auth.InitiateMagicLink(ctx, "acme-corp", email, userID, redirectURL)

// Initiating OTP - added tenantID
err := auth.InitiateOTP(ctx, "acme-corp", email, userID)
```

---

### 4. Passkey Methods

**Before:**
```go
options, err := auth.BeginRegistration(ctx, user)
options, err := auth.BeginLogin(ctx, username)
```

**After:**
```go
options, err := auth.BeginRegistration(ctx, "acme-corp", user)
options, err := auth.BeginLogin(ctx, "acme-corp", username)
```

---

### 5. Subject Layer - IdentityContext Builder

**Before:**
```go
builder := simple.NewContextBuilder(
    roleProvider,
    groupProvider,
    profileProvider,
)
```

**After:**
```go
builder := simple.NewContextBuilder(
    roleProvider,
    permissionProvider,    // NEW - Required parameter
    groupProvider,
    profileProvider,
)
```

---

### 6. Subject Layer - Cache Invalidation

**Before:**
```go
err := cachedBuilder.Invalidate(ctx, subjectID)
entries, err := store.ListBySubject(ctx, subjectID)
err = store.DeleteBySubject(ctx, subjectID)
```

**After:**
```go
err := cachedBuilder.Invalidate(ctx, tenantID, appID, subjectID)
entries, err := store.ListBySubject(ctx, tenantID, subjectID)
err = store.DeleteBySubject(ctx, tenantID, subjectID)
```

---

### 7. Authorization Layer - ABAC AttributeProvider

**Before:**
```go
type AttributeProvider interface {
    GetSubjectAttributes(ctx context.Context, subjectID string) (map[string]any, error)
    GetResourceAttributes(ctx context.Context, resourceType, resourceID string) (map[string]any, error)
}
```

**After:**
```go
type AttributeProvider interface {
    GetSubjectAttributes(ctx context.Context, tenantID string, subjectID string) (map[string]any, error)
    GetResourceAttributes(ctx context.Context, tenantID string, resourceType, resourceID string) (map[string]any, error)
}
```

---

### 8. Authorization Layer - ACL Manager

**Before:**
```go
// Grant permission
err := manager.Grant(ctx, "alice", "document", "doc-123", authz.ActionRead)

// Get permissions
perms, err := manager.GetPermissions(ctx, "alice", "document", identity)

// Get subjects
subjects, err := manager.GetSubjects(ctx, "document", "doc-123")

// Revoke
err = manager.Revoke(ctx, "alice", "document", "doc-123", authz.ActionRead)

// Copy ACL
err = manager.CopyACL(ctx, "document", "doc-123", "doc-456")
```

**After:**
```go
// Grant permission
err := manager.Grant(
    ctx,
    "acme-corp",    // Tenant ID
    "web-app",      // App ID
    "alice",
    &authz.Resource{Type: "document", ID: "doc-123"},
    authz.ActionRead,
)

// Get permissions
perms, err := manager.GetPermissions(
    ctx,
    "acme-corp",
    "web-app",
    "alice",
    "document",
    "doc-123",
    identity,
)

// Get subjects
subjects, err := manager.GetSubjects(
    ctx,
    "acme-corp",
    "web-app",
    "document",
    "doc-123",
)

// Revoke
err = manager.Revoke(
    ctx,
    "acme-corp",
    "web-app",
    "alice",
    &authz.Resource{Type: "document", ID: "doc-123"},
    authz.ActionRead,
)

// Copy ACL
err = manager.CopyACL(
    ctx,
    "acme-corp",
    "web-app",
    "document",
    "doc-123",
    "doc-456",
)
```

---

## Layer-by-Layer Migration

### Credential Layer Migration

#### Basic Authentication

```go
import (
    credential "github.com/primadi/lokstra-auth/01_credential"
    "github.com/primadi/lokstra-auth/01_credential/basic"
)

// Setup user provider with tenant isolation
userProvider := basic.NewInMemoryUserStore()

// Add users with tenant IDs
userProvider.AddUser(&basic.User{
    ID:           "user-001",
    TenantID:     "acme-corp",      // NEW
    Username:     "alice",
    Email:        "alice@acme.com",
    PasswordHash: passwordHash,
    Disabled:     false,
})

// Authenticate with AuthContext
authCtx := &credential.AuthContext{
    TenantID: "acme-corp",
    AppID:    "web-portal",
}

result, err := basicAuth.Authenticate(ctx, authCtx, &basic.BasicCredentials{
    Username: "alice",
    Password: "password123",
})
```

#### OAuth2 Authentication

```go
import (
    credential "github.com/primadi/lokstra-auth/01_credential"
    "github.com/primadi/lokstra-auth/01_credential/oauth2"
)

auth := oauth2.NewAuthenticator(nil)

authCtx := &credential.AuthContext{
    TenantID: "acme-corp",
    AppID:    "mobile-app",
}

result, err := auth.Authenticate(ctx, authCtx, &oauth2.Credentials{
    Provider:    oauth2.ProviderGoogle,
    AccessToken: "google-oauth-token",
})
```

#### API Key Authentication

```go
import (
    credential "github.com/primadi/lokstra-auth/01_credential"
    "github.com/primadi/lokstra-auth/01_credential/apikey"
)

auth := apikey.NewAuthenticator(store)

// Generate key with tenant/app
key, err := auth.GenerateKey(
    ctx,
    "acme-corp",
    "api-service",
    apikey.KeyTypeProduction,
    apikey.EnvironmentProduction,
    "backend-svc",
    "Backend Service Account",
    []string{"read:data", "write:data"},
    nil,
)

// Authenticate
authCtx := &credential.AuthContext{
    TenantID: "acme-corp",
    AppID:    "api-service",
}

result, err := auth.Authenticate(ctx, authCtx, &apikey.Credentials{
    Key: key.Key,
})
```

### Token Layer Migration

No breaking changes in public API - tenant/app context is automatically embedded in tokens:

```go
// JWT tokens now include tenant_id and app_id claims
tokenManager := jwt.NewManager(jwt.DefaultConfig("secret"))

// No changes to Login/Verify calls
loginResp, err := auth.Login(ctx, &lokstraauth.LoginRequest{
    Credentials: credentials,
})
```

### Subject Layer Migration

```go
import (
    subject "github.com/primadi/lokstra-auth/03_subject"
    "github.com/primadi/lokstra-auth/03_subject/simple"
)

// Build with PermissionProvider (was optional, now required)
builder := simple.NewContextBuilder(
    simple.NewStaticRoleProvider(map[string][]string{
        "user-001": {"admin", "developer"},
    }),
    simple.NewStaticPermissionProvider(map[string][]string{
        "user-001": {"read:code", "write:code"},
    }),
    simple.NewStaticGroupProvider(map[string][]string{
        "user-001": {"engineering"},
    }),
    simple.NewStaticProfileProvider(map[string]map[string]any{
        "user-001": {"name": "Alice"},
    }),
)
```

### Authorization Layer Migration

#### RBAC (No Changes)

```go
// RBAC evaluator API unchanged
evaluator := rbac.NewEvaluator(map[string][]string{
    "admin":     {"*"},
    "developer": {"read:code", "write:code"},
})

decision, err := auth.Authorize(ctx, &authz.AuthorizationRequest{
    Subject:  identity,
    Resource: &authz.Resource{Type: "code", ID: "repo-123"},
    Action:   authz.ActionWrite,
})
```

#### ABAC - Update AttributeProvider

```go
type MyAttributeProvider struct{}

// OLD: func (p *MyAttributeProvider) GetSubjectAttributes(ctx context.Context, subjectID string) (map[string]any, error)
// NEW:
func (p *MyAttributeProvider) GetSubjectAttributes(ctx context.Context, tenantID string, subjectID string) (map[string]any, error) {
    // Fetch attributes scoped to tenant
    return map[string]any{
        "department": "Engineering",
        "level":      5,
        "tenant_id":  tenantID,
    }, nil
}

// OLD: func (p *MyAttributeProvider) GetResourceAttributes(ctx context.Context, resourceType, resourceID string) (map[string]any, error)
// NEW:
func (p *MyAttributeProvider) GetResourceAttributes(ctx context.Context, tenantID string, resourceType, resourceID string) (map[string]any, error) {
    // Fetch resource attributes scoped to tenant
    return map[string]any{
        "owner":     "alice",
        "tenant_id": tenantID,
    }, nil
}
```

#### ACL - Update All Calls

See [Breaking Changes #8](#8-authorization-layer---acl-manager) for complete ACL migration.

---

## Examples

### Complete Multi-Tenant Flow

```go
package main

import (
    "context"
    "fmt"
    "log"

    lokstraauth "github.com/primadi/lokstra-auth"
    credential "github.com/primadi/lokstra-auth/01_credential"
    "github.com/primadi/lokstra-auth/01_credential/basic"
    "github.com/primadi/lokstra-auth/02_token/jwt"
    "github.com/primadi/lokstra-auth/03_subject/simple"
    "github.com/primadi/lokstra-auth/04_authz/rbac"
)

func main() {
    ctx := context.Background()

    // 1. Setup user provider with multi-tenant data
    userProvider := basic.NewInMemoryUserStore()
    
    passwordHash, _ := basic.HashPassword("SecurePass123!")
    userProvider.AddUser(&basic.User{
        ID:           "user-001",
        TenantID:     "acme-corp",
        Username:     "alice",
        Email:        "alice@acme.com",
        PasswordHash: passwordHash,
        Disabled:     false,
    })

    // 2. Build Auth runtime
    auth := lokstraauth.NewBuilder().
        WithAuthenticator("basic", basic.NewAuthenticator(userProvider)).
        WithTokenManager(jwt.NewManager(jwt.DefaultConfig("secret-key"))).
        WithSubjectResolver(simple.NewResolver()).
        WithIdentityContextBuilder(simple.NewContextBuilder(
            simple.NewStaticRoleProvider(map[string][]string{
                "user-001": {"admin"},
            }),
            simple.NewStaticPermissionProvider(map[string][]string{}),
            simple.NewStaticGroupProvider(map[string][]string{}),
            simple.NewStaticProfileProvider(map[string]map[string]any{}),
        )).
        WithAuthorizer(rbac.NewEvaluator(map[string][]string{
            "admin": {"*"},
        })).
        EnableRefreshToken().
        Build()

    // 3. Login (AuthContext is embedded in Login request)
    loginResp, err := auth.Login(ctx, &lokstraauth.LoginRequest{
        Credentials: &basic.BasicCredentials{
            Username: "alice",
            Password: "SecurePass123!",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("✓ Logged in: %s\n", loginResp.Identity.Subject.Principal)
    fmt.Printf("  Access Token: %s...\n", loginResp.AccessToken.Value[:50])

    // 4. Verify token
    verifyResp, err := auth.Verify(ctx, &lokstraauth.VerifyRequest{
        Token:                loginResp.AccessToken.Value,
        BuildIdentityContext: true,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("✓ Token valid for subject: %s\n", verifyResp.Identity.Subject.ID)
}
```

---

## Common Patterns

### Pattern 1: Tenant Determination

```go
// Extract tenant from request (e.g., subdomain, header, JWT)
func getTenantFromRequest(r *http.Request) string {
    // Option 1: Subdomain
    host := r.Host
    if strings.Contains(host, ".") {
        return strings.Split(host, ".")[0]
    }
    
    // Option 2: Header
    if tenant := r.Header.Get("X-Tenant-ID"); tenant != "" {
        return tenant
    }
    
    // Option 3: JWT claim
    token := extractToken(r)
    if claims, err := parseToken(token); err == nil {
        return claims["tenant_id"].(string)
    }
    
    return "default"
}
```

### Pattern 2: Middleware Integration

```go
func TenantAuthMiddleware(auth *lokstraauth.Auth) gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := getTenantFromRequest(c.Request)
        appID := "web-app" // Or extract from config
        
        authHeader := c.GetHeader("Authorization")
        token := strings.TrimPrefix(authHeader, "Bearer ")
        
        verifyResp, err := auth.Verify(c.Request.Context(), &lokstraauth.VerifyRequest{
            Token:                token,
            BuildIdentityContext: true,
        })
        
        if err != nil || !verifyResp.Valid {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        // Verify tenant matches
        if verifyResp.Identity.Subject.Attributes["tenant_id"] != tenantID {
            c.JSON(403, gin.H{"error": "Tenant mismatch"})
            c.Abort()
            return
        }
        
        c.Set("identity", verifyResp.Identity)
        c.Set("tenant_id", tenantID)
        c.Set("app_id", appID)
        c.Next()
    }
}
```

### Pattern 3: Service Layer Integration

```go
import (
    "github.com/primadi/lokstra-auth/00_core/services"
)

// Initialize services
tenantService := services.NewTenantService(tenantStore)
appService := services.NewAppService(appStore, tenantService)
userService := services.NewUserService(userStore, tenantService, appService)

// Create tenant
tenant, err := tenantService.CreateTenant(ctx, "Acme Corporation", "Enterprise customer")

// Create app for tenant
app, err := appService.CreateApp(ctx, tenant.TenantID, "web-portal", core.AppTypeWeb, nil)

// Create user in tenant
user, err := userService.CreateUser(ctx, tenant.TenantID, "alice", "alice@acme.com")

// Use tenant/app IDs for authentication
authCtx := &credential.AuthContext{
    TenantID: tenant.TenantID,
    AppID:    app.AppID,
}
```

---

## Checklist

Use this checklist to ensure complete migration:

### Credential Layer
- [ ] Add `AuthContext` parameter to all `Authenticate()` calls
- [ ] Update `UserResolver` interface implementations (passwordless)
- [ ] Add `tenantID` and `appID` to `GenerateKey()` calls (API key)
- [ ] Add `tenantID` to `BeginRegistration()` and `BeginLogin()` (passkey)
- [ ] Update user stores to include `TenantID` field

### Token Layer
- [ ] No changes required (automatic tenant/app embedding)

### Subject Layer
- [ ] Add `PermissionProvider` to `NewContextBuilder()` (was optional)
- [ ] Update cache methods with `tenantID` and `appID` parameters
- [ ] Update custom resolvers if tenant-specific logic needed

### Authorization Layer
- [ ] Update `AttributeProvider` interface implementations (ABAC)
- [ ] Add `tenantID` and `appID` to all ACL manager calls
- [ ] Update policy stores if using policy-based authz

### Service Layer (New)
- [ ] Integrate `TenantService`, `AppService`, `UserService`
- [ ] Update user management to create users via `UserService`
- [ ] Add tenant/app creation flows

### Application Code
- [ ] Implement tenant determination logic
- [ ] Update middleware to extract and validate tenant/app
- [ ] Update login handlers to pass correct `AuthContext`
- [ ] Add tenant isolation to database queries
- [ ] Update API responses to filter by tenant

---

## Support

For migration assistance:
- See `examples/` directory for updated examples
- Check `MULTI_TENANT_IMPLEMENTATION.md` for architecture details
- Review `examples/services/01_multi_tenant_management/` for service layer usage

