# Core Package (core)

The `core` package provides the foundation for multi-tenant and multi-app architecture.

## Folder Structure

```
core/
├── application/        # Application services (business logic)
│   ├── tenant_service.go
│   ├── app_service.go
│   ├── branch_service.go
│   ├── user_service.go
│   └── app_key_service.go
├── domain/            # Domain models and interfaces
│   ├── tenant.go
│   ├── app.go
│   ├── branch.go
│   ├── user.go
│   └── app_key.go
├── infrastructure/    # Data access layer
│   └── repository/
│       ├── tenant_repository.go
│       ├── app_repository.go
│       ├── branch_repository.go
│       ├── user_repository.go
│       └── app_key_repository.go
└── README.md
```

## Data Hierarchy

```
Tenant (Organization/Company)
├── DB-DSN (Database Connection String)
├── DB-Schema (Database Schema Name)
├── App (Application within Tenant)
│   ├── Branch        ← Branch belongs to App (not directly to Tenant)
│   ├── AppKey        ← AppKey belongs to App (service-to-service auth)
│   └── UserApp       ← User's access to this App
│
└── User (tenant-level user)
    └── UserApp (relationship to Apps - many-to-many)
```

**Key Points:**
- **Branch belongs to App**, not directly to Tenant
  - Each App can have different branch structures (retail stores vs warehouses)
  - Branch requires both TenantID and AppID
- **AppKey belongs to App** for service authentication
- **User belongs to Tenant**, accesses Apps through UserApp relationship
- **Each Tenant has its own database** (multi-database multi-tenancy)

## Overview

This package is the **FOUNDATION** of the entire lokstra-auth system. Every operation in other layers (credential, token, subject, authz) **MUST** use the tenant and app concepts from this package.

## Core Concepts

### 1. Tenant
A **Tenant** represents an organization/company using the system. Each tenant has its own isolated database.

```go
tenant := &core.Tenant{
    ID:       "acme-corp",
    Name:     "Acme Corporation",
    DBDsn:    "postgres://user:pass@localhost:5432/acme_db",
    DBSchema: "acme_schema",
    Status:   core.TenantStatusActive,
    Settings: core.TenantSettings{
        MaxUsers: 1000,
        MaxApps:  10,
        AllowedAuthMethods: []string{"basic", "oauth2", "apikey"},
        PasswordPolicy: core.PasswordPolicy{
            MinLength:        8,
            RequireUppercase: true,
            RequireNumbers:   true,
        },
    },
}
```

**Key Points:**
- Each tenant is **completely isolated** with its own database
- **DBDsn**: Database connection string (required) - supports multi-database tenancy
- **DBSchema**: Schema name (required) - supports schema-based tenancy within a database
- Data cannot leak between tenants
- Credentials cannot be used cross-tenant
- Quotas and policies are managed per tenant

### 2. App
An **App** is an application within a tenant.

```go
app := &core.App{
    ID:       "web-portal",
    TenantID: "acme-corp",
    Name:     "Web Portal",
    Type:     core.AppTypeWeb,
    Status:   core.AppStatusActive,
    Config: core.AppConfig{
        AccessTokenExpiry:  15 * time.Minute,
        RefreshTokenExpiry: 7 * 24 * time.Hour,
        TokenAlgorithm:     "HS256",
        AllowedOrigins:     []string{"https://portal.acme.com"},
        OAuth2Providers: map[string]core.OAuth2ProviderConfig{
            "google": {
                Provider:     "google",
                ClientID:     "google-client-id",
                ClientSecret: "google-secret",
                RedirectURL:  "https://portal.acme.com/auth/callback",
                Scopes:       []string{"email", "profile"},
                Enabled:      true,
            },
        },
    },
}
```

**Key Points:**
- One tenant can have multiple apps
- Each app has its own configuration (OAuth, token expiry, etc.)
- User permissions differ per app
- Tokens are scoped per app

### 3. User
A **User** is an end-user within a tenant.

```go
user := &core.User{
    ID:       "user123",
    TenantID: "acme-corp",
    Username: "alice",
    Email:    "alice@acme.com",
    FullName: "Alice Johnson",
    Status:   core.UserStatusActive,
}
```

**Key Points:**
- Username and email are **unique per tenant** (not globally unique)
- Users can access multiple apps within a tenant
- Access to apps must be explicitly granted

### 4. User-App Access
Links users to apps with roles and permissions.

```go
userApp := &core.UserApp{
    UserID:   "user123",
    TenantID: "acme-corp",
    AppID:    "web-portal",
    Status:   core.UserAppStatusActive,
    Roles:    []string{"admin", "editor"},
    Permissions: []string{"read:users", "write:posts"},
}
```

**Key Points:**
- Users must be granted access to an app before they can login
- Roles and permissions are per app
- Access can be revoked without deleting the user

### 5. Context
**Context** carries tenant and app information in every operation.

```go
ctx := core.NewContext("acme-corp", "web-portal").
    WithUser("user123").
    WithIP("192.168.1.1").
    WithSession("session-abc")
```

**CRITICAL:**
- **EVERY** auth operation **MUST** include context
- Tenant ID and App ID are **REQUIRED**
- Used for isolation and audit trails

## Authentication Models

### User Authentication (Human → App)
End-users authenticate to applications using:
- **Basic Auth**: Username + password
- **OAuth2**: Google, GitHub, etc.
- **Passwordless**: Magic link or OTP
- **Passkey**: WebAuthn/FIDO2

```go
// User "alice" logs into "web-portal" app
authCtx := core.NewContext("acme-corp", "web-portal").
    WithIP("192.168.1.1")

result, _ := basicAuth.Authenticate(ctx, authCtx, &basic.Credentials{
    Username: "alice",
    Password: "secret123",
})
```

### Service-to-Service Authentication (App → App)
Services/applications authenticate using **API Keys**:

```go
// Service registration creates API key
apiKey, _ := apiKeyAuth.GenerateKey(ctx, "acme-corp", "mobile-app", &apikey.KeyOptions{
    Name:        "Mobile App Backend Service",
    Description: "Backend service for mobile app",
    Scopes:      []string{"read:users", "write:notifications"},
    ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
})

// Returns: "mobile-app_a1b2c3d4.s3cr3tk3yv4lu3"
// Format: {app_id}_{key_id}.{secret}
```

**Service-to-Service Flow:**

1. **App Registration** - Tenant admin registers a new app/service:
```go
// Register backend service as an app
backendService := &core.App{
    ID:       "backend-service",
    TenantID: "acme-corp",
    Name:     "Backend API Service",
    Type:     core.AppTypeService, // New type for services
    Config: core.AppConfig{
        // Service-specific config
        AllowedScopes: []string{"read:users", "write:notifications"},
    },
}
appService.Create(ctx, backendService)
```

2. **API Key Generation** - Generate key for the service:
```go
// Generate API key for backend service
apiKey, _ := apiKeyAuth.GenerateKey(
    ctx,
    "acme-corp",           // tenant_id
    "backend-service",     // app_id (the service's app)
    &apikey.KeyOptions{
        Name:      "Production Key",
        Scopes:    []string{"read:users", "write:notifications"},
        ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
    },
)

// Store the full key securely - it won't be shown again!
// apiKey.FullKey = "backend-service_k3ym4p5d.s3cr3tk3yv4lu3h3r3"
```

3. **Service Authentication** - The service uses the API key:
```go
// Backend service authenticates
authCtx := core.NewContext("acme-corp", "backend-service").
    WithIP("10.0.1.50")

result, _ := apiKeyAuth.Authenticate(ctx, authCtx, &apikey.Credentials{
    Key: "backend-service_k3ym4p5d.s3cr3tk3yv4lu3h3r3",
})

// result.Claims includes:
// - tenant_id: "acme-corp"
// - app_id: "backend-service"
// - scopes: ["read:users", "write:notifications"]
```

**Key Points:**
- API Keys are scoped to tenant + app
- Each service is registered as an App with type "service"
- API Keys include scopes for fine-grained access control
- Keys can expire and be rotated
- Full key (key_id.secret) is only shown once at creation

## Services

### TenantService

```go
type TenantService interface {
    Create(ctx context.Context, tenant *Tenant) error
    Get(ctx context.Context, tenantID string) (*Tenant, error)
    Update(ctx context.Context, tenant *Tenant) error
    Delete(ctx context.Context, tenantID string) error
    List(ctx context.Context, filters TenantFilters) ([]*Tenant, error)
    Suspend(ctx context.Context, tenantID string, reason string) error
    Activate(ctx context.Context, tenantID string) error
    GetByDomain(ctx context.Context, domain string) (*Tenant, error)
}
```

**Use Cases:**
- Platform admin registers new tenant
- Suspend tenant for non-payment
- Update tenant quota/limits
- Custom domain mapping

### AppService

```go
type AppService interface {
    Create(ctx context.Context, app *App) error
    Get(ctx context.Context, tenantID, appID string) (*App, error)
    Update(ctx context.Context, app *App) error
    Delete(ctx context.Context, tenantID, appID string) error
    ListByTenant(ctx context.Context, tenantID string) ([]*App, error)
    Disable(ctx context.Context, tenantID, appID string) error
    Enable(ctx context.Context, tenantID, appID string) error
}
```

**Use Cases:**
- Tenant admin creates new app or service
- Configure OAuth providers per app
- Set token expiry per app
- Register service apps for API key authentication
- Disable app temporarily

### UserService

```go
type UserService interface {
    Create(ctx context.Context, user *User) error
    Get(ctx context.Context, tenantID, userID string) (*User, error)
    GetByUsername(ctx context.Context, tenantID, username string) (*User, error)
    GetByEmail(ctx context.Context, tenantID, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, tenantID, userID string) error
    
    // App Access Management
    GrantAppAccess(ctx context.Context, userApp *UserApp) error
    RevokeAppAccess(ctx context.Context, tenantID, appID, userID string) error
    UpdateAppAccess(ctx context.Context, userApp *UserApp) error
    GetAppAccess(ctx context.Context, tenantID, appID, userID string) (*UserApp, error)
    ListUserApps(ctx context.Context, tenantID, userID string) ([]*UserApp, error)
    ListAppUsers(ctx context.Context, tenantID, appID string) ([]*User, error)
}
```

**Use Cases:**
- Register new user in tenant
- Grant user access to app with roles
- List all apps a user can access
- Revoke app access when user leaves team

## Complete Flow Examples

### Example 1: Multi-App User Flow

```go
package main

import (
    "context"
    "github.com/primadi/lokstra-auth/core"
)

func main() {
    ctx := context.Background()
    
    // 1. Platform Admin: Create Tenant
    tenant := &core.Tenant{
        ID:       "acme-corp",
        Name:     "Acme Corporation",
        DBDsn:    "postgres://user:pass@localhost:5432/acme_db",
        DBSchema: "acme",
        Settings: core.TenantSettings{
            MaxUsers: 1000,
            MaxApps:  10,
            AllowedAuthMethods: []string{"basic", "oauth2"},
        },
    }
    tenantService.Create(ctx, tenant)
    
    // 2. Tenant Admin: Create Apps
    webApp := &core.App{
        ID:       "web-portal",
        TenantID: "acme-corp",
        Name:     "Web Portal",
        Type:     core.AppTypeWeb,
        Config: core.AppConfig{
            AccessTokenExpiry: 15 * time.Minute,
            OAuth2Providers: map[string]core.OAuth2ProviderConfig{
                "google": {
                    ClientID:     "...",
                    ClientSecret: "...",
                    Enabled:      true,
                },
            },
        },
    }
    appService.Create(ctx, webApp)
    
    mobileApp := &core.App{
        ID:       "mobile-app",
        TenantID: "acme-corp",
        Name:     "Mobile App",
        Type:     core.AppTypeMobile,
    }
    appService.Create(ctx, mobileApp)
    
    // 3. Tenant Admin: Create User
    alice := &core.User{
        ID:       "user123",
        TenantID: "acme-corp",
        Username: "alice",
        Email:    "alice@acme.com",
        FullName: "Alice Johnson",
    }
    userService.Create(ctx, alice)
    
    // 4. Tenant Admin: Grant App Access
    // Alice can access Web Portal as admin
    userService.GrantAppAccess(ctx, &core.UserApp{
        UserID:   "user123",
        TenantID: "acme-corp",
        AppID:    "web-portal",
        Roles:    []string{"admin"},
    })
    
    // Alice can access Mobile App as user
    userService.GrantAppAccess(ctx, &core.UserApp{
        UserID:   "user123",
        TenantID: "acme-corp",
        AppID:    "mobile-app",
        Roles:    []string{"user"},
    })
    
    // 5. User Authentication (using credential layer)
    authCtx := core.NewContext("acme-corp", "web-portal").
        WithIP("192.168.1.1")
    
    // This will use credential layer (credential)
    // result, _ := authenticator.Authenticate(
    //     ctx, authCtx, credentials
    // )
    
    // 6. List User's Apps
    apps, _ := userService.ListUserApps(ctx, "acme-corp", "user123")
    for _, app := range apps {
        fmt.Printf("App: %s, Roles: %v\n", app.AppID, app.Roles)
    }
    // Output:
    // App: web-portal, Roles: [admin]
    // App: mobile-app, Roles: [user]
}
```

### Example 2: Service-to-Service Authentication
```

### Example 2: Service-to-Service Authentication

```go
package main

import (
    "context"
    "github.com/primadi/lokstra-auth/core"
    "github.com/primadi/lokstra-auth/credential/apikey"
)

func main() {
    ctx := context.Background()
    
    // 1. Tenant Admin: Register Backend Service as App
    backendService := &core.App{
        ID:       "backend-service",
        TenantID: "acme-corp",
        Name:     "Backend API Service",
        Type:     core.AppTypeService,
        Config: core.AppConfig{
            AllowedScopes: []string{
                "read:users",
                "write:users",
                "read:notifications",
                "write:notifications",
            },
        },
    }
    appService.Create(ctx, backendService)
    
    // 2. Tenant Admin: Generate API Key for Service
    apiKey, _ := apiKeyAuth.GenerateKey(
        ctx,
        "acme-corp",
        "backend-service",
        &apikey.KeyOptions{
            Name:        "Production Backend Key",
            Description: "API key for backend service",
            Scopes:      []string{"read:users", "write:notifications"},
            ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
        },
    )
    
    // IMPORTANT: Save the full key - it won't be shown again!
    // apiKey.FullKey = "backend-service_k3ym4p5d.s3cr3tk3yh3r3"
    fmt.Println("API Key:", apiKey.FullKey)
    fmt.Println("Store this key securely!")
    
    // 3. Backend Service: Use API Key to Authenticate
    authCtx := core.NewContext("acme-corp", "backend-service").
        WithIP("10.0.1.50")
    
    result, _ := apiKeyAuth.Authenticate(ctx, authCtx, &apikey.Credentials{
        Key: apiKey.FullKey,
    })
    
    fmt.Printf("Authenticated: %v\n", result.Success)
    fmt.Printf("Tenant: %s\n", result.TenantID)
    fmt.Printf("App: %s\n", result.AppID)
    fmt.Printf("Scopes: %v\n", result.Claims["scopes"])
    
    // 4. Service calls another service with the token
    // The token (from result.Claims) can be used to make authorized requests
}
```

## Database Schema

```sql
-- Tenants
CREATE TABLE tenants (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255),
    db_dsn TEXT NOT NULL,
    db_schema VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    settings JSONB,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    
    INDEX idx_status (status),
    UNIQUE INDEX idx_domain (domain),
    INDEX idx_schema (db_schema)
);

-- Apps
CREATE TABLE apps (
    id VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    config JSONB NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    PRIMARY KEY (tenant_id, id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    INDEX idx_tenant_status (tenant_id, status)
);

-- Users
CREATE TABLE users (
    id VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    status VARCHAR(20) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    
    PRIMARY KEY (tenant_id, id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE (tenant_id, username),
    UNIQUE (tenant_id, email),
    INDEX idx_tenant_status (tenant_id, status)
);

-- User-App Access
CREATE TABLE user_apps (
    user_id VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    app_id VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    roles JSONB,
    permissions JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    
    PRIMARY KEY (tenant_id, app_id, user_id),
    FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id, app_id) REFERENCES apps(tenant_id, id) ON DELETE CASCADE,
    INDEX idx_user (tenant_id, user_id),
    INDEX idx_app (tenant_id, app_id)
);
```

## Security Best Practices

### 1. Always Validate Tenant Context

```go
// ✅ Good
func Authenticate(ctx *core.Context, creds Credentials) error {
    if err := ctx.Validate(); err != nil {
        return err
    }
    // tenant_id and app_id are guaranteed to exist
}

// ❌ Bad
func Authenticate(creds Credentials) error {
    // No tenant context - could leak data!
}
```

### 2. Scope All Queries by Tenant

```go
// ✅ Good
user, _ := store.GetByUsername(ctx, tenantID, username)

// ❌ Bad
user, _ := store.GetByUsername(ctx, username) // Could return wrong user!
```

### 3. Never Allow Cross-Tenant Operations

```go
// ✅ Good
if user.TenantID != requestTenantID {
    return ErrUnauthorized
}

// ❌ Bad
// Just check userID without tenant - security breach!
```

### 4. Validate App Access

```go
// ✅ Good
userApp, _ := userService.GetAppAccess(ctx, tenantID, appID, userID)
if !userApp.IsActive() {
    return ErrNoAccess
}

// ❌ Bad
// Assume user can access any app in tenant
```

### 5. Secure API Key Storage

```go
// ✅ Good - Service stores full key securely on first generation
apiKey, _ := apiKeyAuth.GenerateKey(ctx, tenantID, appID, options)
vault.Store("api_key", apiKey.FullKey) // Store in secure vault
// Later: use the stored key
storedKey, _ := vault.Get("api_key")

// ❌ Bad - Trying to retrieve key later
// API keys can't be retrieved after generation - only key_id is stored
```

## Architecture Layers

### Domain Layer (`domain/`)
Defines core business entities and their behaviors:
- **Tenant**: Organization with isolated database
- **App**: Application within a tenant
- **Branch**: Physical location/office under an app
- **User**: End-user within a tenant
- **AppKey**: Service authentication key for an app

### Application Layer (`application/`)
Business logic and service orchestration:
- **TenantService**: Tenant lifecycle management
- **AppService**: App creation and configuration
- **BranchService**: Branch management within apps
- **UserService**: User management and app access
- **AppKeyService**: Service key generation and validation

### Infrastructure Layer (`infrastructure/`)
Data access and persistence:
- **Repositories**: Database operations for each entity
- **Stores**: In-memory or external storage implementations

## Next Steps

After `core` is complete, other layers will be updated:

1. **credential**: Add tenant_id & app_id to all authenticators ✅
2. **token**: Include tenant_id & app_id in JWT claims
3. **subject**: Resolve subject with tenant context
4. **authz**: Scope all authz checks by tenant & app

## See Also

- [Multi-Tenant Architecture](../docs/multi_tenant_architecture.md)
- [Credential Layer](../credential/README.md)
- [Token Layer](../token/README.md)
