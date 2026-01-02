# Refresh Token Rotation → Subject → Authz Flow

## Overview
This document explains the complete flow from refresh token rotation to authorization check, including how identity context is built and used for access control.

## Component Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           CLIENT APPLICATION                             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ 1. Login / Refresh
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      CREDENTIAL (Authentication)                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ BasicAuthService.Login()                                         │   │
│  │  • Authenticate user credentials                                 │   │
│  │  • Generate access token (15 min)                                │   │
│  │  • Generate refresh token (7 days)                               │   │
│  │  • Return: { access_token, refresh_token }                       │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ BasicAuthService.Refresh()  ← REFRESH TOKEN ROTATION             │   │
│  │  • Verify old refresh token → Extract claims                     │   │
│  │  • Generate NEW access token from claims                         │   │
│  │  • Generate NEW refresh token from claims (ROTATION)             │   │
│  │  • Revoke old refresh token (prevent reuse)                      │   │
│  │  • Return: { new_access_token, new_refresh_token }               │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                           │
│  Claims in Tokens:                                                       │
│  {                                                                        │
│    "sub": "user-id",          // Subject ID                              │
│    "tenant_id": "acme-corp",  // Tenant context                          │
│    "app_id": "main-app",      // App context                             │
│    "branch_id": "hq-jakarta", // Branch context (optional)               │
│    "type": "access" | "refresh",                                         │
│    "exp": 1234567890,                                                    │
│    "iat": 1234567890                                                     │
│  }                                                                        │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ 2. Access Protected Resource
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      MIDDLEWARE (Token Verification)                     │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ AuthMiddleware.Handler()                                         │   │
│  │  • Extract Bearer token from Authorization header                │   │
│  │  • Verify token signature & expiration                           │   │
│  │  • Extract claims from token                                     │   │
│  │  • Build IdentityContext (inject to request context)             │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ 3. Build Identity
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      SUBJECT (Identity Context)                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ IdentityContextBuilder.Build()                                   │   │
│  │  • Create Subject from token claims                              │   │
│  │  • Fetch Roles (RoleProvider)                                    │   │
│  │  • Fetch Permissions (PermissionProvider)                        │   │
│  │  • Fetch Groups (GroupProvider)                                  │   │
│  │  • Fetch Profile (ProfileProvider)                               │   │
│  │  • Build complete IdentityContext                                │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                           │
│  IdentityContext:                                                        │
│  {                                                                        │
│    Subject: {                                                             │
│      ID: "user-id",                                                      │
│      TenantID: "acme-corp",                                              │
│      Type: "user",                                                       │
│      Principal: "john.doe",                                              │
│      Attributes: { email: "john@acme.com", ... }                         │
│    },                                                                    │
│    TenantID: "acme-corp",                                                │
│    AppID: "main-app",                                                    │
│    BranchID: "hq-jakarta",                                               │
│    Roles: ["admin", "editor"],       ← From RoleProvider                │
│    Permissions: ["read", "write"],   ← From PermissionProvider           │
│    Groups: ["managers"],             ← From GroupProvider                │
│    Profile: { name: "John Doe" },    ← From ProfileProvider              │
│    Session: { IP: "...", UserAgent: "..." }                              │
│  }                                                                        │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ 4. Authorization Check
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      AUTHZ (Authorization)                               │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ RoleMiddleware / PermissionMiddleware                            │   │
│  │  • Get IdentityContext from request                              │   │
│  │  • Check required roles/permissions                              │   │
│  │  • Allow or Deny access                                          │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ PolicyEvaluator.Evaluate()                                       │   │
│  │  • Get IdentityContext                                           │   │
│  │  • Evaluate RBAC/ABAC/ACL policies                               │   │
│  │  • Return AuthorizationDecision { Allowed: true/false }          │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ 5. Execute Business Logic
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      APPLICATION HANDLER                                 │
│  • Access IdentityContext via middleware.GetIdentity(c)                 │
│  • User is authenticated & authorized                                    │
│  • Execute business logic                                                │
└─────────────────────────────────────────────────────────────────────────┘
```

## Complete Integration Flow

### 1. Login (Initial Authentication)

```http
POST /api/auth/cred/basic/login
Content-Type: application/json
X-Tenant-ID: acme-corp
X-App-ID: main-app

{
  "username": "john.doe",
  "password": "SecurePass123!"
}
```

**Response:**
```json
{
  "success": true,
  "access_token": "eyJhbGc...ACCESS_TOKEN_1",
  "refresh_token": "eyJhbGc...REFRESH_TOKEN_1",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Access Token Claims:**
```json
{
  "sub": "john.doe",
  "tenant_id": "acme-corp",
  "app_id": "main-app",
  "branch_id": "hq-jakarta",
  "type": "access",
  "exp": 1732730100,
  "iat": 1732729200
}
```

### 2. Access Protected Resource (First Time)

```http
GET /api/documents/doc-123
Authorization: Bearer eyJhbGc...ACCESS_TOKEN_1
X-Tenant-ID: acme-corp
X-App-ID: main-app
```

**Server-side Processing:**

```go
// 1. AuthMiddleware extracts and verifies token
func (m *AuthMiddleware) Handler() func(c *request.Context) error {
    return func(c *request.Context) error {
        // Extract: "Bearer eyJhbGc...ACCESS_TOKEN_1"
        token, _ := m.tokenExtractor(c)
        
        // Verify token (signature, expiration, revocation)
        verifyResp, _ := m.auth.Verify(c, &lokstraauth.VerifyRequest{
            Token:                token,
            BuildIdentityContext: true, // ← Build full identity
        })
        
        // Inject identity into request context
        c.Set(IdentityContextKey, verifyResp.Identity)
        
        return c.Next()
    }
}

// 2. RoleMiddleware checks required roles
func (m *AnyRoleMiddleware) Handler() func(c *request.Context) error {
    return func(c *request.Context) error {
        // Get identity from context
        identity, _ := middleware.GetIdentity(c)
        
        // Check if user has any required role
        if !identity.HasAnyRole(m.roles...) {
            return errors.New("insufficient roles")
        }
        
        return c.Next()
    }
}

// 3. Application handler
func GetDocument(c *request.Context) error {
    // Get identity (authenticated & authorized)
    identity := middleware.MustGetIdentity(c)
    
    // Access identity information
    userID := identity.Subject.ID          // "john.doe"
    tenantID := identity.TenantID          // "acme-corp"
    roles := identity.Roles                // ["admin", "editor"]
    permissions := identity.Permissions    // ["read", "write", "delete"]
    
    // Fetch document with tenant isolation
    doc, _ := docService.Get(c, tenantID, "doc-123")
    
    return c.Resp.Json(doc)
}
```

### 3. Access Token Expires (After 15 Minutes)

```http
GET /api/documents/doc-456
Authorization: Bearer eyJhbGc...ACCESS_TOKEN_1
```

**Response:**
```json
{
  "error": "Unauthorized",
  "message": "token expired"
}
```

### 4. Refresh Token (Token Rotation)

Client detects 401 and refreshes:

```http
POST /api/auth/cred/basic/refresh
Content-Type: application/json
X-Tenant-ID: acme-corp
X-App-ID: main-app

{
  "refresh_token": "eyJhbGc...REFRESH_TOKEN_1"
}
```

**Server-side Processing:**

```go
func (s *BasicAuthService) Refresh(ctx *request.Context, req *basic.RefreshRequest) (*basic.RefreshResponse, error) {
    // 1. Verify refresh token → Extract claims
    result, _ := s.TokenManager.Verify(ctx, req.RefreshToken)
    
    // result.Claims = {
    //   "sub": "john.doe",
    //   "tenant_id": "acme-corp",
    //   "app_id": "main-app",
    //   "type": "refresh",
    //   ...
    // }
    
    // 2. Generate NEW access token from claims
    newAccessToken, _ := s.TokenManager.Generate(ctx, result.Claims)
    
    // 3. Generate NEW refresh token (ROTATION)
    newRefreshToken, _ := s.TokenManager.GenerateRefreshToken(ctx, result.Claims)
    
    // 4. Revoke old refresh token
    s.TokenManager.Revoke(ctx, req.RefreshToken) // ← REFRESH_TOKEN_1 is now REVOKED
    
    // 5. Return both new tokens
    return &basic.RefreshResponse{
        AccessToken:  newAccessToken.Value,  // ACCESS_TOKEN_2
        RefreshToken: newRefreshToken.Value, // REFRESH_TOKEN_2 (NEW!)
        ...
    }
}
```

**Response:**
```json
{
  "success": true,
  "access_token": "eyJhbGc...ACCESS_TOKEN_2",
  "refresh_token": "eyJhbGc...REFRESH_TOKEN_2",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Important:** Client MUST store and use `REFRESH_TOKEN_2` for next refresh. `REFRESH_TOKEN_1` cannot be reused.

### 5. Retry Protected Resource (With New Access Token)

```http
GET /api/documents/doc-456
Authorization: Bearer eyJhbGc...ACCESS_TOKEN_2
X-Tenant-ID: acme-corp
X-App-ID: main-app
```

**Success!** Same flow as step 2:
- AuthMiddleware verifies `ACCESS_TOKEN_2`
- Builds IdentityContext from token claims
- RoleMiddleware checks authorization
- Handler accesses document

### 6. Authorization Examples

#### Example A: Role-Based Access Control (RBAC)

```go
// Route definition with role middleware
app.GET("/admin/users", 
    middleware.NewAnyRoleMiddleware(auth, []string{"admin"}).Handler(),
    ListUsers,
)

// Request
GET /admin/users
Authorization: Bearer eyJhbGc...ACCESS_TOKEN_2

// Identity has roles: ["admin", "editor"]
// ✅ ALLOWED (has "admin" role)
```

#### Example B: Permission-Based Access Control

```go
// Route with permission middleware
app.POST("/documents/:id/publish",
    middleware.NewAllPermissionsMiddleware(auth, []string{"write", "publish"}).Handler(),
    PublishDocument,
)

// Request
POST /documents/doc-123/publish
Authorization: Bearer eyJhbGc...ACCESS_TOKEN_2

// Identity has permissions: ["read", "write", "delete"]
// ❌ DENIED (missing "publish" permission)
```

#### Example C: Policy-Based Authorization (ABAC)

```go
func DeleteDocument(c *request.Context) error {
    identity := middleware.MustGetIdentity(c)
    
    // Fetch document
    doc, _ := docService.Get(c, identity.TenantID, c.Param("id"))
    
    // Check authorization with policy evaluator
    decision, _ := policyEvaluator.Evaluate(c, &authz.AuthorizationRequest{
        Subject: identity,
        Resource: &authz.Resource{
            Type:     "document",
            ID:       doc.ID,
            TenantID: doc.TenantID,
            AppID:    doc.AppID,
            Attributes: map[string]any{
                "owner":      doc.OwnerID,
                "visibility": doc.Visibility,
                "status":     doc.Status,
            },
        },
        Action: authz.ActionDelete,
        Context: map[string]any{
            "time": time.Now(),
            "ip":   c.ClientIP(),
        },
    })
    
    if !decision.Allowed {
        return errors.New(decision.Reason)
    }
    
    // Delete document
    docService.Delete(c, identity.TenantID, doc.ID)
    return c.Resp.Json(map[string]any{"success": true})
}
```

**Policy Example:**
```json
{
  "id": "policy-001",
  "tenant_id": "acme-corp",
  "app_id": "main-app",
  "name": "Document Owner Delete Policy",
  "effect": "allow",
  "subjects": ["*"],
  "resources": ["document:*"],
  "actions": ["delete"],
  "conditions": {
    "subject.id == resource.owner": true,
    "resource.status != 'published'": true
  }
}
```

### 7. Identity Context in Business Logic

```go
func CreateInvoice(c *request.Context) error {
    // Get authenticated identity
    identity := middleware.MustGetIdentity(c)
    
    // Access identity fields
    userID := identity.Subject.ID               // "john.doe"
    tenantID := identity.TenantID               // "acme-corp"
    appID := identity.AppID                     // "main-app"
    branchID := identity.BranchID               // "hq-jakarta"
    userEmail := identity.Subject.Attributes["email"] // "john@acme.com"
    userName := identity.Profile["name"]        // "John Doe"
    
    // Check permissions programmatically
    if !identity.HasPermission("invoice:create") {
        return errors.New("missing invoice:create permission")
    }
    
    // Check roles
    if identity.HasRole("finance") {
        // Finance users can approve invoices immediately
        invoice.Status = "approved"
    } else {
        invoice.Status = "pending"
    }
    
    // Create invoice with tenant/app/branch isolation
    invoice := &Invoice{
        ID:         idgen.Generate(),
        TenantID:   tenantID,
        AppID:      appID,
        BranchID:   branchID,
        CreatedBy:  userID,
        CreatorEmail: userEmail.(string),
        Amount:     req.Amount,
        Status:     invoice.Status,
    }
    
    invoiceService.Create(c, invoice)
    
    // Audit log with identity
    auditService.Log(c, &AuditEntry{
        TenantID:  tenantID,
        UserID:    userID,
        Action:    "invoice.create",
        Resource:  invoice.ID,
        IPAddress: identity.Session.IPAddress,
        UserAgent: identity.Session.UserAgent,
    })
    
    return c.Resp.Json(invoice)
}
```

## Complete Flow Diagram

```
┌──────────────┐
│   Client     │
└──────┬───────┘
       │
       │ 1. Login (username + password)
       ├────────────────────────────────────────────────┐
       │                                                │
       ▼                                                ▼
┌──────────────────────────────────┐         ┌──────────────────────┐
│ BasicAuthService.Login()         │────────>│ TokenManager         │
│  • Authenticate credentials      │         │  • Generate tokens   │
│  • Generate access + refresh     │<────────│    with claims       │
└──────────────┬───────────────────┘         └──────────────────────┘
               │
               │ Returns: { access_token, refresh_token }
               │
       ┌───────▼────────┐
       │     Client     │
       │  Stores tokens │
       └───────┬────────┘
               │
               │ 2. Access protected resource (with access_token)
               │
               ▼
┌──────────────────────────────────────────────────────────┐
│ AuthMiddleware                                           │
│  • Extract Bearer token                                  │
│  • Verify signature & expiration ────────┐               │
└───────────────────┬──────────────────────┘               │
                    │                                      │
                    │ Valid token                          ▼
                    │                         ┌──────────────────────┐
                    ▼                         │ TokenManager.Verify()│
┌──────────────────────────────────┐          │  • Check signature   │
│ IdentityContextBuilder           │          │  • Check expiration  │
│  • Extract claims from token     │<─────────│  • Check revocation  │
│  • Load roles (RoleProvider)     │          └──────────────────────┘
│  • Load permissions              │
│  • Load groups                   │
│  • Build IdentityContext         │
└───────────────────┬──────────────┘
                    │
                    │ IdentityContext injected into request context
                    │
                    ▼
┌──────────────────────────────────┐
│ RoleMiddleware /                 │
│ PermissionMiddleware             │
│  • Get IdentityContext           │
│  • Check required roles/perms    │
│  • Allow or Deny                 │
└───────────────────┬──────────────┘
                    │
                    │ Authorized ✓
                    │
                    ▼
┌──────────────────────────────────┐
│ Application Handler              │
│  • Get IdentityContext           │
│  • Execute business logic        │
│  • Access user/tenant info       │
└──────────────────────────────────┘

       │
       │ 3. Access token expires (after 15 min)
       │
       ▼
┌──────────────┐
│   Client     │ Receives 401 Unauthorized
└──────┬───────┘
       │
       │ 4. Refresh token (with refresh_token)
       │
       ▼
┌──────────────────────────────────────────────────────────┐
│ BasicAuthService.Refresh()  (TOKEN ROTATION)             │
│                                                           │
│  ┌──────────────────────────────────────┐                │
│  │ 1. Verify old refresh_token          │                │
│  │    • Extract claims                  │                │
│  └──────────────┬───────────────────────┘                │
│                 │                                         │
│  ┌──────────────▼───────────────────────┐                │
│  │ 2. Generate NEW access token         │                │
│  │    • From extracted claims           │                │
│  └──────────────┬───────────────────────┘                │
│                 │                                         │
│  ┌──────────────▼───────────────────────┐                │
│  │ 3. Generate NEW refresh token        │                │
│  │    • From same claims (ROTATION)     │                │
│  └──────────────┬───────────────────────┘                │
│                 │                                         │
│  ┌──────────────▼───────────────────────┐                │
│  │ 4. Revoke old refresh_token          │                │
│  │    • Prevent reuse (security)        │                │
│  └──────────────┬───────────────────────┘                │
│                 │                                         │
│  ┌──────────────▼───────────────────────┐                │
│  │ 5. Return both NEW tokens            │                │
│  └──────────────────────────────────────┘                │
└──────────────────────────────────────────────────────────┘
       │
       │ Returns: { new_access_token, new_refresh_token }
       │
       ▼
┌──────────────┐
│   Client     │ Updates stored tokens (BOTH!)
└──────┬───────┘
       │
       │ 5. Retry protected resource (with new_access_token)
       │
       └─────> (Back to step 2)
```

## Multi-Layer Security

### Layer 1: Token Rotation (Credential)
- ✅ Refresh token is rotated on every use
- ✅ Old refresh token is immediately revoked
- ✅ Prevents token reuse attacks

### Layer 2: Token Verification (Middleware)
- ✅ Signature verification (JWT)
- ✅ Expiration check
- ✅ Revocation check (blacklist)
- ✅ Tenant/App isolation

### Layer 3: Identity Context (Subject)
- ✅ Rich identity with roles, permissions, groups
- ✅ Profile enrichment
- ✅ Session tracking (IP, User Agent)

### Layer 4: Authorization (Authz)
- ✅ Role-based access control (RBAC)
- ✅ Permission-based access control
- ✅ Policy-based access control (ABAC)
- ✅ Resource ownership checks
- ✅ Tenant isolation enforcement

## Complete Implementation Example

### Server Setup

```go
package main

import (
    lokstraauth "github.com/primadi/lokstra-auth"
    "github.com/primadi/lokstra-auth/middleware"
)

func main() {
    // Initialize auth runtime
    auth := lokstraauth.NewAuth(authConfig)
    
    // Setup routes with layered security
    app := lokstra.New()
    
    // Public routes (no auth)
    app.POST("/auth/login", basicAuthService.Login)
    app.POST("/auth/refresh", basicAuthService.Refresh)
    
    // Protected routes (auth required)
    protected := app.Group("/api")
    protected.Use(middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
        Auth: auth,
    }).Handler())
    
    // Admin routes (auth + admin role required)
    admin := protected.Group("/admin")
    admin.Use(middleware.NewAnyRoleMiddleware(auth, []string{"admin"}).Handler())
    admin.GET("/users", ListUsers)
    admin.POST("/users", CreateUser)
    
    // Document routes (auth + specific permissions)
    docs := protected.Group("/documents")
    docs.GET("/:id", GetDocument) // Any authenticated user
    docs.POST("", 
        middleware.NewAllPermissionsMiddleware(auth, []string{"write"}).Handler(),
        CreateDocument,
    )
    docs.PUT("/:id",
        middleware.NewAllPermissionsMiddleware(auth, []string{"write"}).Handler(),
        UpdateDocument,
    )
    docs.DELETE("/:id",
        middleware.NewAllPermissionsMiddleware(auth, []string{"delete"}).Handler(),
        DeleteDocument,
    )
    
    app.Run(":8080")
}
```

### Client Implementation (JavaScript)

```javascript
class AuthClient {
    constructor(baseUrl, tenantId, appId) {
        this.baseUrl = baseUrl;
        this.tenantId = tenantId;
        this.appId = appId;
        this.accessToken = null;
        this.refreshToken = null;
    }
    
    async login(username, password) {
        const response = await fetch(`${this.baseUrl}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Tenant-ID': this.tenantId,
                'X-App-ID': this.appId,
            },
            body: JSON.stringify({ username, password })
        });
        
        const data = await response.json();
        this.accessToken = data.access_token;
        this.refreshToken = data.refresh_token;
        
        // Store in localStorage
        localStorage.setItem('access_token', this.accessToken);
        localStorage.setItem('refresh_token', this.refreshToken);
    }
    
    async refresh() {
        const response = await fetch(`${this.baseUrl}/auth/refresh`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Tenant-ID': this.tenantId,
                'X-App-ID': this.appId,
            },
            body: JSON.stringify({ refresh_token: this.refreshToken })
        });
        
        if (!response.ok) {
            // Refresh failed, redirect to login
            this.logout();
            window.location.href = '/login';
            return false;
        }
        
        const data = await response.json();
        this.accessToken = data.access_token;
        this.refreshToken = data.refresh_token; // ← UPDATE to NEW refresh token!
        
        // Update localStorage
        localStorage.setItem('access_token', this.accessToken);
        localStorage.setItem('refresh_token', this.refreshToken);
        
        return true;
    }
    
    async request(url, options = {}) {
        // Add Authorization header
        options.headers = {
            ...options.headers,
            'Authorization': `Bearer ${this.accessToken}`,
            'X-Tenant-ID': this.tenantId,
            'X-App-ID': this.appId,
        };
        
        let response = await fetch(`${this.baseUrl}${url}`, options);
        
        // If 401, try to refresh
        if (response.status === 401) {
            const refreshed = await this.refresh();
            if (refreshed) {
                // Retry with new token
                options.headers['Authorization'] = `Bearer ${this.accessToken}`;
                response = await fetch(`${this.baseUrl}${url}`, options);
            }
        }
        
        return response;
    }
    
    logout() {
        this.accessToken = null;
        this.refreshToken = null;
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
    }
}

// Usage
const auth = new AuthClient('http://localhost:8080', 'acme-corp', 'main-app');

// Login
await auth.login('john.doe', 'SecurePass123!');

// Make authenticated requests (auto-refresh on 401)
const response = await auth.request('/api/documents/doc-123');
const document = await response.json();
```

## Summary

### Complete Flow:
1. **Login** → Receive access_token + refresh_token
2. **Access Resource** → Middleware verifies token → Build IdentityContext → Check Authorization → Execute Handler
3. **Token Expires** → Client refreshes
4. **Refresh** → NEW access_token + NEW refresh_token (old token revoked)
5. **Retry Access** → Success with new tokens

### Architecture Benefits:
- 🔒 **Security**: Multi-layer (token rotation, verification, authorization)
- 🏢 **Multi-tenant**: Tenant/App/Branch isolation at every layer
- 🎯 **Flexible**: RBAC, PBAC, ABAC, ACL support
- 📊 **Rich Identity**: Full context (roles, permissions, groups, profile)
- 🔄 **Token Rotation**: Prevent token theft/reuse
- ⚡ **Performance**: Caching at subject layer
