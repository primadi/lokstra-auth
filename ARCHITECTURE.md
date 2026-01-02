# Lokstra Auth - Architecture & Design

## Layer Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│ CLIENT APPLICATIONS                                                   │
│ Web Apps, Mobile Apps, Services                                      │
└──────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────┐
│ PUBLIC API ENDPOINTS (HTTP/REST)                                     │
├──────────────────────────────────────────────────────────────────────┤
│ core - Management APIs (Tenant, App, User, Keys, Configs)        │
│   • /api/core/tenant/*                                               │
│   • /api/core/app/*                                                  │
│   • /api/core/user/*                                                 │
│   • /api/core/app-key/*                                              │
│   • /api/core/credential-config/*                                    │
│                                                                       │
│ credential - Authentication APIs (Login, Register)                │
│   • //api/auth/cred/basic/*        → Username/Password                     │
│   • //api/auth/cred/oauth2/*       → Google, Azure, GitHub                 │
│   • //api/auth/cred/apikey/*       → API Key validation                    │
│   • //api/auth/cred/passwordless/* → Email/SMS OTP, Magic Link             │
│   • //api/auth/cred/passkey/*      → WebAuthn/FIDO2                        │
└──────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────┐
│ MIDDLEWARE LAYER                                                     │
├──────────────────────────────────────────────────────────────────────┤
│ • AuthMiddleware       → Verify token, inject identity               │
│ • RoleMiddleware       → Check role(s)                               │
│ • PermissionMiddleware → Check permission(s)                         │
│ • TenantMiddleware     → Validate & inject tenant context            │
│ • RateLimitMiddleware  → Rate limiting per tenant/user               │
│ • AuditMiddleware      → Log all operations with tenant context      │
└──────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────┐
│ INTERNAL LIBRARIES (No HTTP endpoints)                               │
├──────────────────────────────────────────────────────────────────────┤
│ token - Token Management                                          │
│   • TokenManager       → Generate, verify, revoke tokens             │
│   • TokenStore         → Store refresh tokens                        │
│   • JWT implementation                                               │
│                                                                       │
│ subject - Identity Context                                        │
│   • IdentityResolver    → Build subject from token claims             │
│   • IdentityBuilder    → Enrich identity with roles/permissions      │
│   • RoleProvider       → Get roles for user                          │
│   • PermissionProvider → Get permissions for user                    │
│                                                                       │
│ authz - Authorization                                             │
│   • PolicyEvaluator    → Evaluate ABAC/RBAC policies                 │
│   • PermissionChecker  → Check permissions                           │
│   • RoleChecker        → Check roles                                 │
│   • ACL Manager        → Manage access control lists                 │
└──────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────┐
│ DATA LAYER                                                           │
├──────────────────────────────────────────────────────────────────────┤
│ PostgreSQL (Multi-tenant with schema isolation)                      │
│   • public schema      → Global metadata, tenant registry            │
│   • tenant_{id} schema → Tenant-specific data                        │
└──────────────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### ✅ Layers with HTTP Endpoints

| Layer | Purpose | HTTP Routes | Used By |
|-------|---------|-------------|---------|
| **core** | Tenant/App/User management | `/api/core/*` | Admin dashboards, setup tools |
| **credential** | Authentication (verify + issue tokens) | `//api/auth/cred/*` | Client apps (login/register) |

### ❌ Layers WITHOUT HTTP Endpoints (Pure Libraries)

| Layer | Purpose | HTTP Routes | Used By |
|-------|---------|-------------|---------|
| **token** | Token encode/decode/validate | ❌ None | credential, middleware |
| **subject** | Identity context building | ❌ None | Middleware, services |
| **authz** | Authorization checks | ❌ None | Middleware, services |

## Why This Design?

### 1. **core = Management APIs** ✅ Has Endpoints

**Purpose:** CRUD operations for system resources
- Create tenants, apps, users
- Manage API keys
- Configure credential providers

**Accessed by:** Admin users, setup scripts, management dashboards

### 2. **credential = Authentication APIs** ✅ Has Endpoints

**Purpose:** Authenticate users and issue tokens
- Login with username/password
- OAuth2 social login
- Passwordless authentication
- Returns JWT tokens directly

**Accessed by:** End users through client apps

### 3. **token = Token Utilities** ❌ No Endpoints

**Purpose:** Technical utilities for token handling
- Generate JWT tokens
- Verify token signatures
- Manage token revocation list

**Why no endpoints?**
- Token generation happens inside credential (login returns token)
- Token validation happens in middleware (before protected routes)
- No direct client interaction needed

### 4. **subject = Identity Context** ❌ No Endpoints

**Purpose:** Build rich identity from token claims
- Extract subject from token
- Load roles/permissions from database
- Build complete identity context

**Why no endpoints?**
- Identity building happens in middleware (after token verification)
- Identity context injected into request
- Services use injected identity

**Flow:**
```
Token → token.Verify() → Claims
  ↓
Claims → identity.IdentityResolver.Resolve() → Subject
  ↓
Subject → subject.IdentityBuilder.Build() → IdentityContext
  ↓
Inject into request context
```

### 5. **authz = Authorization Logic** ❌ No Endpoints

**Purpose:** Make authorization decisions
- Check if user has role
- Check if user has permission
- Evaluate ABAC policies

**Why no endpoints?**
- Authorization checks happen in middleware (before handler execution)
- Policy evaluation happens inline
- Services can also call directly for fine-grained checks

**Flow:**
```
Request → AuthMiddleware → Verify token → Build identity
  ↓
RoleMiddleware → Check identity.HasRole("admin")
  ↓
Handler → Can also call auth.CheckPermission() inline
```

## Request Flow Examples

### Example 1: Public Login (No Auth Required)

```
POST //api/auth/cred/basic/login
{
  "username": "john.doe",
  "password": "SecurePass123!"
}

Flow:
1. Request → BasicAuthService.Login()
2. Verify password (credential/authenticator)
3. Generate tokens (token/jwt/manager)
4. Return response with tokens

Response:
{
  "success": true,
  "access_token": "eyJhbG...",
  "refresh_token": "eyJhbG...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

### Example 2: Protected Resource (Auth Required)

```
GET /api/business/orders
Authorization: Bearer eyJhbG...

Middleware chain:
1. AuthMiddleware → Verify token
   ↓ Extract claims from token (token)
   ↓ Build subject from claims (subject)
   ↓ Load roles/permissions (subject)
   ↓ Inject IdentityContext into request

2. RoleMiddleware → Check identity.HasRole("sales")
   ↓ Use injected identity
   ↓ Call authz.RoleChecker

3. Handler → Process request
   ↓ Can access identity from context
   ↓ Can do inline permission checks
```

### Example 3: Fine-grained Permission Check

```
DELETE /api/business/orders/{id}
Authorization: Bearer eyJhbG...

Middleware chain:
1. AuthMiddleware → Build identity

2. PermissionMiddleware → Check "delete:orders"
   ↓ identity.HasPermission("delete:orders")
   ↓ Or call auth.CheckPermission()

3. Handler → Additional checks
   ↓ Check if order belongs to user's tenant
   ↓ Check if order status allows deletion
   ↓ Can call authz.PolicyEvaluator for complex rules
```

## Multi-tenant Isolation

All layers enforce multi-tenant isolation:

| Layer | Tenant Isolation Method |
|-------|------------------------|
| **core** | Path parameter `/tenant/id/{tenant_id}` |
| **credential** | Header `X-Tenant-ID`, token claims |
| **token** | Claims include `tenant_id`, `app_id` |
| **subject** | Identity includes `tenantID`, `appID` |
| **authz** | All checks scoped to `tenantID`, `appID` |
| **Database** | Schema-per-tenant: `tenant_{id}` |

## Security Layers

```
┌─────────────────────────────────────────────────────────┐
│ 1. Authentication (credential + token)            │
│    Who are you?                                         │
│    → Verify credentials                                 │
│    → Issue signed JWT token                             │
└─────────────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────────┐
│ 2. Identity Resolution (subject)                     │
│    What context do you have?                            │
│    → Extract claims from token                          │
│    → Load roles, permissions, groups                    │
│    → Build complete identity context                    │
└─────────────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────────┐
│ 3. Authorization (authz)                             │
│    What are you allowed to do?                          │
│    → Check roles (RBAC)                                 │
│    → Check permissions                                  │
│    → Evaluate policies (ABAC)                           │
└─────────────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────────┐
│ 4. Business Logic                                       │
│    Execute the actual operation                         │
│    → Can do additional fine-grained checks              │
│    → Access tenant-scoped data                          │
└─────────────────────────────────────────────────────────┘
```

## Summary: Endpoint vs Library Decision

**Has HTTP Endpoints:**
- ✅ **core** - CRUD for tenants/apps/users (management plane)
- ✅ **credential** - Authentication flows (data plane)

**Pure Libraries (No Endpoints):**
- ❌ **token** - Used by credential and middleware
- ❌ **subject** - Used by middleware to build identity
- ❌ **authz** - Used by middleware for access control

**Rationale:**
- Operations users/admins perform directly = endpoints
- Technical utilities used by other layers = libraries
- This keeps API surface clean and focused
- Better separation of concerns
- Easier to test and maintain
