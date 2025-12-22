# Auth Runtime

The `Auth` runtime is the main entry point for using Lokstra Auth framework. It orchestrates all 4 layers and provides a unified, easy-to-use API.

## Overview

Instead of manually wiring together authenticators, token managers, subject resolvers, and authorizers, the `Auth` runtime provides:

- **Fluent Builder API** for configuration
- **Unified methods** that span multiple layers
- **Simplified error handling**
- **Consistent interface** across all authentication flows

## Basic Usage

### 1. Build Auth Runtime

```go
import (
    lokstraauth "github.com/primadi/lokstra-auth"
    "github.com/primadi/lokstra-auth/credential/basic"
    "github.com/primadi/lokstra-auth/token/jwt"
    "github.com/primadi/lokstra-auth/identity/simple"
    "github.com/primadi/lokstra-auth/authz/rbac"
)

auth := lokstraauth.NewBuilder().
    // Layer 1: Register authenticators
    WithAuthenticator("basic", basicAuthenticator).
    WithAuthenticator("oauth2", oauth2Authenticator).
    
    // Layer 2: Set token manager
    WithTokenManager(jwtManager).
    
    // Layer 3: Set subject resolution
    WithIdentityResolver(identityResolver).
    WithIdentityContextBuilder(contextBuilder).
    
    // Layer 4: Set authorizer
    WithAuthorizer(rbacAuthorizer).
    
    // Configuration
    EnableRefreshToken().
    SetDefaultAuthenticator("basic").
    Build()
```

### 2. Login

The `Login` method executes Layer 1 → 2 → 3:

```go
response, err := auth.Login(ctx, &lokstraauth.LoginRequest{
    Credentials: &basic.BasicCredentials{
        Username: "john.doe",
        Password: "SecurePass123!",
    },
    Metadata: map[string]any{
        "ip_address": "192.168.1.100",
    },
})

if err != nil {
    log.Fatal(err)
}

// Access response data
accessToken := response.AccessToken
refreshToken := response.RefreshToken
identity := response.Identity
```

### 3. Verify Token

The `Verify` method executes Layer 2 → 3:

```go
response, err := auth.Verify(ctx, &lokstraauth.VerifyRequest{
    Token: tokenString,
    BuildIdentityContext: true,
})

if err != nil || !response.Valid {
    // Token is invalid
    return
}

// Use verified identity
identity := response.Identity
```

### 4. Authorization

The `Authorize` method executes Layer 4:

```go
decision, err := auth.Authorize(ctx, &authz.AuthorizationRequest{
    Subject: identity,
    Resource: &authz.Resource{
        Type: "document",
        ID:   "doc-123",
    },
    Action: authz.ActionWrite,
})

if err != nil {
    log.Fatal(err)
}

if decision.Allowed {
    // Grant access
} else {
    // Deny access
    log.Printf("Access denied: %s", decision.Reason)
}
```

### 5. Convenience Methods

Quick permission and role checks:

```go
// Check permission
canWrite, err := auth.CheckPermission(ctx, identity, "write:document")

// Check role
isAdmin, err := auth.CheckRole(ctx, identity, "admin")
```

## Builder API

### Configuration Methods

```go
builder := lokstraauth.NewBuilder()

// Set custom configuration
builder.WithConfig(customConfig)

// Register authenticators
builder.WithAuthenticator("basic", basicAuth)
builder.WithAuthenticator("oauth2", oauth2Auth)

// Set token manager
builder.WithTokenManager(jwtManager)

// Set subject resolution
builder.WithIdentityResolver(resolver)
builder.WithIdentityContextBuilder(builder)

// Set authorizer
builder.WithAuthorizer(rbacAuthorizer)

// Enable/disable features
builder.EnableRefreshToken()
builder.DisableRefreshToken()
builder.EnableSessionManagement()
builder.DisableSessionManagement()

// Set defaults
builder.SetDefaultAuthenticator("basic")

// Add metadata
builder.AddMetadata("app_name", "My App")

// Build
auth := builder.Build()
```

## Configuration

The `Config` struct allows you to customize runtime behavior:

```go
config := &lokstraauth.Config{
    DefaultAuthenticatorType: "basic",
    IssueRefreshToken:        true,
    SessionManagement:        false,
    Metadata: map[string]any{
        "app_name": "My Application",
    },
}

auth := lokstraauth.NewBuilder().
    WithConfig(config).
    // ... other configurations
    Build()
```

## Complete Example

See `/examples/credential/runtime_example.go` for a complete working example.

## Benefits

### 1. Simplified Setup
No need to manually wire components together.

### 2. Consistent API
All operations follow the same request/response pattern.

### 3. Error Handling
Unified error types across all layers.

### 4. Flexibility
Can still access individual layers when needed.

### 5. Type Safety
Full compile-time type checking.

## Advanced Usage

### Multiple Authenticators

Register multiple authenticators for different credential types:

```go
auth := lokstraauth.NewBuilder().
    WithAuthenticator("basic", basicAuth).
    WithAuthenticator("oauth2", oauth2Auth).
    WithAuthenticator("apikey", apikeyAuth).
    WithAuthenticator("passkey", passkeyAuth).
    Build()

// Login with basic auth
auth.Login(ctx, &lokstraauth.LoginRequest{
    Credentials: &basic.BasicCredentials{...},
})

// Login with OAuth2
auth.Login(ctx, &lokstraauth.LoginRequest{
    Credentials: &oauth2.OAuth2Credentials{...},
})
```

### Custom Metadata

Pass metadata through the authentication flow:

```go
response, err := auth.Login(ctx, &lokstraauth.LoginRequest{
    Credentials: credentials,
    Metadata: map[string]any{
        "ip_address": req.RemoteAddr,
        "user_agent": req.UserAgent(),
        "device_id":  req.Header.Get("X-Device-ID"),
    },
})
```

### Conditional Identity Building

Control when to build full identity context:

```go
// Don't build identity context (faster, for token-only checks)
response, err := auth.Verify(ctx, &lokstraauth.VerifyRequest{
    Token: tokenString,
    BuildIdentityContext: false,
})

// Build full identity context (for authorization)
response, err := auth.Verify(ctx, &lokstraauth.VerifyRequest{
    Token: tokenString,
    BuildIdentityContext: true,
})
```

## Best Practices

1. **Create once, use many times**: Build the `Auth` runtime once at application startup
2. **Dependency injection**: Pass the `Auth` instance to handlers/services that need it
3. **Error handling**: Always check errors returned from `Auth` methods
4. **Context usage**: Pass proper context with timeout/cancellation support
5. **Security**: Never log sensitive data (passwords, tokens) from requests/responses
