# Multi-Tenant & Multi-App Architecture

## Overview

Lokstra-Auth dirancang untuk mendukung **Multi-Tenant** dan **Multi-App** architecture dimana:

- **Tenant**: Organisasi/perusahaan yang menggunakan sistem (e.g., "Acme Corp", "Widget Inc")
- **App**: Aplikasi dalam satu tenant (e.g., "Web Portal", "Mobile App", "Admin Dashboard")
- **User**: Pengguna yang bisa akses satu atau lebih app dalam satu atau lebih tenant

## Core Concepts

### Tenant Isolation

Setiap tenant **HARUS** terisolasi sepenuhnya:
- Data tidak boleh bocor antar tenant
- Credentials tidak bisa digunakan cross-tenant (kecuali explicitly allowed)
- Policies dan permissions di-scope per tenant
- Audit logs terpisah per tenant

### App Context

Dalam satu tenant, bisa ada multiple apps:
- Setiap app punya configuration sendiri (OAuth providers, token expiry, etc.)
- User permissions bisa berbeda per app
- Token di-scope per app untuk security

### Hierarchy

```
Tenant (tenant_id: "acme-corp")
├── App 1 (app_id: "web-portal")
│   ├── User 1 (user_id: "alice")
│   ├── User 2 (user_id: "bob")
│   └── Roles: admin, user, viewer
├── App 2 (app_id: "mobile-app")
│   ├── User 1 (user_id: "alice")
│   ├── User 3 (user_id: "charlie")
│   └── Roles: user, guest
└── App 3 (app_id: "admin-dashboard")
    ├── User 1 (user_id: "alice")
    └── Roles: super-admin
```

## Architecture Layers

### Layer 0: Core (NEW)

Package: `core`

#### Tenant Management

```go
type Tenant struct {
    ID          string                 // Unique identifier (e.g., "acme-corp")
    Name        string                 // Display name (e.g., "Acme Corporation")
    Domain      string                 // Optional custom domain
    Status      TenantStatus           // active, suspended, deleted
    Settings    TenantSettings         // Tenant-wide settings
    Metadata    map[string]any         // Custom metadata
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}

type TenantSettings struct {
    MaxUsers            int           // Quota limits
    MaxApps             int
    AllowedAuthMethods  []string      // basic, oauth2, apikey, etc.
    PasswordPolicy      PasswordPolicy
    SessionTimeout      time.Duration
    RequireMFA          bool
}

type TenantService interface {
    Create(ctx context.Context, tenant *Tenant) error
    Get(ctx context.Context, tenantID string) (*Tenant, error)
    Update(ctx context.Context, tenant *Tenant) error
    Delete(ctx context.Context, tenantID string) error
    List(ctx context.Context, filters TenantFilters) ([]*Tenant, error)
    Suspend(ctx context.Context, tenantID string) error
    Activate(ctx context.Context, tenantID string) error
}
```

#### App Management

```go
type App struct {
    ID          string                 // Unique within tenant (e.g., "web-portal")
    TenantID    string                 // Parent tenant
    Name        string                 // Display name
    Type        AppType                // web, mobile, api, desktop
    Status      AppStatus              // active, disabled
    Config      AppConfig              // App-specific configuration
    Metadata    map[string]any
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type AppConfig struct {
    // OAuth2 Configuration (per app)
    OAuth2Providers     map[string]OAuth2Config
    
    // Token Configuration (per app)
    AccessTokenExpiry   time.Duration
    RefreshTokenExpiry  time.Duration
    TokenAlgorithm      string
    
    // Security Settings (per app)
    AllowedOrigins      []string      // CORS
    AllowedCallbacks    []string      // OAuth2 redirects
    RateLimits          RateLimitConfig
    
    // Feature Flags (per app)
    Features            map[string]bool
}

type AppService interface {
    Create(ctx context.Context, app *App) error
    Get(ctx context.Context, tenantID, appID string) (*App, error)
    Update(ctx context.Context, app *App) error
    Delete(ctx context.Context, tenantID, appID string) error
    ListByTenant(ctx context.Context, tenantID string) ([]*App, error)
}
```

#### User Management

```go
type User struct {
    ID          string                 // Unique identifier (UUID)
    TenantID    string                 // Belongs to tenant
    Username    string                 // Unique within tenant
    Email       string                 // Unique within tenant
    FullName    string
    Status      UserStatus             // active, suspended, deleted
    Metadata    map[string]any
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}

// User-App association
type UserApp struct {
    UserID      string
    TenantID    string
    AppID       string
    Status      UserAppStatus          // active, revoked
    Roles       []string               // Roles in this app
    Permissions []string               // Additional permissions
    CreatedAt   time.Time
    RevokedAt   *time.Time
}

type UserService interface {
    Create(ctx context.Context, user *User) error
    Get(ctx context.Context, tenantID, userID string) (*User, error)
    GetByUsername(ctx context.Context, tenantID, username string) (*User, error)
    GetByEmail(ctx context.Context, tenantID, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, tenantID, userID string) error
    
    // App access management
    GrantAppAccess(ctx context.Context, userApp *UserApp) error
    RevokeAppAccess(ctx context.Context, tenantID, appID, userID string) error
    ListUserApps(ctx context.Context, tenantID, userID string) ([]*UserApp, error)
    ListAppUsers(ctx context.Context, tenantID, appID string) ([]*User, error)
}
```

### Layer 1: Credential

Setiap credential **HARUS** include `tenant_id` dan `app_id`:

```go
// Updated AuthenticationResult
type AuthenticationResult struct {
    Success   bool
    Subject   string              // User ID
    TenantID  string              // REQUIRED: Tenant context
    AppID     string              // REQUIRED: App context
    Claims    map[string]any
    Error     error
}

// Context untuk authentication
type AuthContext struct {
    TenantID  string              // REQUIRED
    AppID     string              // REQUIRED
    IPAddress string
    UserAgent string
    SessionID string
}
```

#### Basic Auth dengan Multi-Tenant

```go
type User struct {
    ID          string
    TenantID    string              // REQUIRED
    Username    string              // Unique within tenant
    Email       string
    PasswordHash string
    // ...
}

type UserProvider interface {
    GetByUsername(ctx context.Context, tenantID, username string) (*User, error)
    GetByEmail(ctx context.Context, tenantID, email string) (*User, error)
    GetByID(ctx context.Context, tenantID, userID string) (*User, error)
    Store(ctx context.Context, user *User) error
}

// Authenticate dengan context
result, err := authenticator.AuthenticateWithContext(
    ctx,
    &AuthContext{
        TenantID: "acme-corp",
        AppID:    "web-portal",
    },
    &BasicCredentials{
        Username: "alice",
        Password: "secret",
    },
)
```

#### API Key dengan Multi-Tenant

```go
type APIKey struct {
    ID          string
    TenantID    string              // REQUIRED: Key belongs to tenant
    AppID       string              // REQUIRED: Key scoped to app
    KeyID       string
    SecretHash  string
    UserID      string
    // ...
}

// Generate key untuk specific tenant & app
apiKey, err := auth.GenerateKey(
    ctx,
    "acme-corp",                   // tenant_id
    "web-portal",                  // app_id
    apikey.KeyTypePublic,
    apikey.EnvProduction,
    "user123",
    "Production API Key",
    []string{"read:users"},
    &expiresIn,
)
```

#### OAuth2 dengan Multi-Tenant

```go
// OAuth2 config per app
type OAuth2Config struct {
    TenantID      string
    AppID         string
    Provider      string            // google, github, etc.
    ClientID      string            // Different per app
    ClientSecret  string
    RedirectURL   string
    Scopes        []string
}

// Authenticate dengan tenant & app context
result, err := oauth2Auth.AuthenticateWithContext(
    ctx,
    &AuthContext{
        TenantID: "acme-corp",
        AppID:    "mobile-app",
    },
    &OAuth2Credentials{
        Provider: "google",
        Code:     "auth-code",
    },
)
```

### Layer 2: Token

Token **HARUS** include `tenant_id` dan `app_id` dalam claims:

```go
// JWT Claims structure
type Claims struct {
    Subject     string              `json:"sub"`      // User ID
    TenantID    string              `json:"tenant_id"` // REQUIRED
    AppID       string              `json:"app_id"`    // REQUIRED
    IssuedAt    int64               `json:"iat"`
    ExpiresAt   int64               `json:"exp"`
    Issuer      string              `json:"iss"`
    Audience    []string            `json:"aud"`
    Scopes      []string            `json:"scopes"`
    Roles       []string            `json:"roles"`
    Metadata    map[string]any      `json:"metadata"`
}

// Generate token dengan tenant & app context
token, err := tokenManager.Generate(ctx, &TokenRequest{
    Subject:   "user123",
    TenantID:  "acme-corp",
    AppID:     "web-portal",
    Claims: map[string]any{
        "roles": []string{"admin"},
        "scopes": []string{"read:users", "write:posts"},
    },
})

// Verify token - automatically extracts tenant_id & app_id
claims, err := tokenManager.Verify(ctx, tokenString)
// claims.TenantID = "acme-corp"
// claims.AppID = "web-portal"
```

#### Token Store per Tenant & App

```go
type TokenStore interface {
    Store(ctx context.Context, tenantID, appID, userID, token string, expiresAt time.Time) error
    Get(ctx context.Context, token string) (*StoredToken, error)
    Revoke(ctx context.Context, token string) error
    RevokeUserTokens(ctx context.Context, tenantID, appID, userID string) error
    RevokeAllAppTokens(ctx context.Context, tenantID, appID string) error
    RevokeAllTenantTokens(ctx context.Context, tenantID string) error
}
```

### Layer 3: Subject

Subject resolver **HARUS** scope per tenant:

```go
type Subject struct {
    ID          string
    TenantID    string              // REQUIRED
    Username    string
    Email       string
    FullName    string
    Roles       []string            // Per tenant roles
    Permissions []string            // Per tenant permissions
    Metadata    map[string]any
}

type IdentityResolver interface {
    // Resolve dengan tenant context
    Resolve(ctx context.Context, tenantID, userID string) (*Subject, error)
    
    // Resolve dengan app context (includes app-specific roles)
    ResolveForApp(ctx context.Context, tenantID, appID, userID string) (*Subject, error)
}
```

### Layer 4: Authorization

RBAC, ABAC, ACL **HARUS** scope per tenant & app:

```go
// RBAC
type Role struct {
    ID          string
    TenantID    string              // REQUIRED
    AppID       string              // REQUIRED: Role defined per app
    Name        string              // e.g., "admin", "editor"
    Permissions []string
    Description string
}

type RoleAssignment struct {
    UserID      string
    TenantID    string
    AppID       string
    RoleID      string
    GrantedAt   time.Time
    ExpiresAt   *time.Time
}

// Check permission dengan context
hasPermission, err := rbac.HasPermission(
    ctx,
    "acme-corp",                   // tenant_id
    "web-portal",                  // app_id
    "user123",                     // user_id
    "delete:posts",                // permission
)

// ABAC
type Policy struct {
    ID          string
    TenantID    string              // REQUIRED
    AppID       string              // REQUIRED
    Name        string
    Rules       []PolicyRule
}

// ACL
type ACL struct {
    ResourceID  string
    TenantID    string              // REQUIRED
    AppID       string              // REQUIRED
    UserID      string
    Permissions []string
}
```

## Database Schema

### Core Tables

```sql
-- Tenants
CREATE TABLE tenants (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255),
    status VARCHAR(20) NOT NULL,
    settings JSONB,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    
    INDEX idx_status (status),
    INDEX idx_domain (domain)
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
    revoked_at TIMESTAMP,
    
    PRIMARY KEY (tenant_id, app_id, user_id),
    FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id, app_id) REFERENCES apps(tenant_id, id) ON DELETE CASCADE,
    INDEX idx_user (tenant_id, user_id),
    INDEX idx_app (tenant_id, app_id)
);

-- Basic Credentials
CREATE TABLE user_passwords (
    user_id VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    PRIMARY KEY (tenant_id, user_id),
    FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id) ON DELETE CASCADE
);

-- API Keys
CREATE TABLE api_keys (
    id VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    app_id VARCHAR(100) NOT NULL,
    key_id VARCHAR(100) NOT NULL,
    secret_hash VARCHAR(255) NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    name VARCHAR(255),
    scopes JSONB,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE,
    
    PRIMARY KEY (tenant_id, app_id, id),
    FOREIGN KEY (tenant_id, app_id) REFERENCES apps(tenant_id, id) ON DELETE CASCADE,
    UNIQUE (key_id),  -- key_id is globally unique for fast lookup
    INDEX idx_tenant_app_user (tenant_id, app_id, user_id)
);

-- Roles (RBAC)
CREATE TABLE roles (
    id VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    app_id VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    permissions JSONB NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL,
    
    PRIMARY KEY (tenant_id, app_id, id),
    FOREIGN KEY (tenant_id, app_id) REFERENCES apps(tenant_id, id) ON DELETE CASCADE,
    UNIQUE (tenant_id, app_id, name)
);

-- Role Assignments
CREATE TABLE role_assignments (
    user_id VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    app_id VARCHAR(100) NOT NULL,
    role_id VARCHAR(100) NOT NULL,
    granted_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP,
    
    PRIMARY KEY (tenant_id, app_id, user_id, role_id),
    FOREIGN KEY (tenant_id, app_id, role_id) REFERENCES roles(tenant_id, app_id, id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id) ON DELETE CASCADE
);
```

## Migration Strategy

### Phase 1: Add tenant_id & app_id (Breaking Change)
1. Add `core` package
2. Update all contracts to include tenant_id & app_id
3. Update all examples

### Phase 2: Implement Services
1. TenantService
2. AppService
3. UserService

### Phase 3: Update Each Layer
1. Update credential
2. Update token
3. Update subject
4. Update authz

### Phase 4: Examples & Documentation
1. Complete multi-tenant examples
2. Migration guide
3. Best practices

## Security Considerations

### Tenant Isolation
- **ALWAYS** validate tenant_id in every operation
- Use row-level security in database if possible
- Never allow cross-tenant queries

### App Scoping
- Tokens cannot be used across apps (even within same tenant)
- API keys scoped to specific app
- OAuth2 callbacks validated per app

### User Access Control
- Users must be explicitly granted access to apps
- Revoke access when user leaves tenant
- Audit all cross-app access

## Usage Examples

### Complete Flow

```go
// 1. Register Tenant
tenant := &core.Tenant{
    ID:   "acme-corp",
    Name: "Acme Corporation",
}
tenantService.Create(ctx, tenant)

// 2. Register App
app := &core.App{
    ID:       "web-portal",
    TenantID: "acme-corp",
    Name:     "Web Portal",
    Type:     core.AppTypeWeb,
}
appService.Create(ctx, app)

// 3. Register User
user := &core.User{
    ID:       "user123",
    TenantID: "acme-corp",
    Username: "alice",
    Email:    "alice@acme.com",
}
userService.Create(ctx, user)

// 4. Grant App Access
userService.GrantAppAccess(ctx, &core.UserApp{
    UserID:   "user123",
    TenantID: "acme-corp",
    AppID:    "web-portal",
    Roles:    []string{"admin"},
})

// 5. Authenticate
result, _ := auth.AuthenticateWithContext(
    ctx,
    &AuthContext{
        TenantID: "acme-corp",
        AppID:    "web-portal",
    },
    &BasicCredentials{
        Username: "alice",
        Password: "secret",
    },
)

// 6. Generate Token
token, _ := tokenManager.Generate(ctx, &TokenRequest{
    Subject:  result.Subject,
    TenantID: result.TenantID,
    AppID:    result.AppID,
})

// 7. Verify & Use
claims, _ := tokenManager.Verify(ctx, token)
// claims.TenantID = "acme-corp"
// claims.AppID = "web-portal"
```

## Next Steps

1. ✅ Document architecture (this file)
2. ⏳ Create `core` package structure
3. ⏳ Update credential layer contracts
4. ⏳ Implement TenantService, AppService, UserService
5. ⏳ Update all authenticators
6. ⏳ Update token managers
7. ⏳ Update subject resolvers
8. ⏳ Update authz evaluators
9. ⏳ Create migration examples
10. ⏳ Update all documentation
