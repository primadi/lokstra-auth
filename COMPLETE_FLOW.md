# Complete Authentication & Authorization Flow

## Flow Diagram

```
┌─────────────┐
│   CLIENT    │
└──────┬──────┘
       │
       │ 1. POST /api/auth/cred/basic/login
       │    { username, password }
       │    Headers: X-Tenant-ID, X-App-ID
       ▼
┌──────────────────────────────────────────────────────────────┐
│ CREDENTIAL LAYER                                             │
│ BasicAuthService.Login()                                     │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐    │
│ │ 1. Validate tenant + app                             │    │
│ │ 2. Find user by username                             │    │
│ │ 3. Verify password (bcrypt)                          │    │
│ │ 4. Build AuthenticationResult with claims:           │    │
│ │    {                                                 │    │
│ │      "sub": "alice",                                 │    │
│ │      "tenant_id": "demo-tenant",                     │    │
│ │      "app_id": "demo-app",                           │    │
│ │      "email": "alice@demo.com",                      │    │
│ │      "type": "user"                                  │    │
│ │    }                                                 │    │
│ └──────────────────────────────────────────────────────┘    │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           │ 5. Pass claims to Token Manager
                           ▼
┌──────────────────────────────────────────────────────────────┐
│ TOKEN LAYER                                                  │
│ JWT Manager.Generate(claims)                                 │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐    │
│ │ 1. Add standard JWT claims (iat, exp, iss, aud)     │    │
│ │ 2. Merge with user claims                           │    │
│ │ 3. Sign with secret key (HS256)                     │    │
│ │ 4. Generate access token (15 min expiry)            │    │
│ │ 5. Generate refresh token (7 days expiry)           │    │
│ │ 6. Return Token objects                             │    │
│ └──────────────────────────────────────────────────────┘    │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           │ 6. Return to client
                           ▼
┌─────────────┐
│   CLIENT    │
│ Stores:     │
│ - access_token  : "eyJhbGci..." (use for API calls)        │
│ - refresh_token : "eyJhbGci..." (use for refresh)          │
│ - expires_in    : 900 (15 minutes)                          │
└──────┬──────┘
       │
       │ 7. GET /api/protected/info
       │    Headers: 
       │      Authorization: Bearer eyJhbGci...
       │      X-Tenant-ID: demo-tenant
       │      X-App-ID: demo-app
       ▼
┌──────────────────────────────────────────────────────────────┐
│ MIDDLEWARE LAYER - AuthMiddleware                            │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐    │
│ │ 1. Extract Bearer token from Authorization header   │    │
│ │ 2. Validate tenant + app from headers               │    │
│ └──────────────────────────────────────────────────────┘    │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           │ 8. Verify token
                           ▼
┌──────────────────────────────────────────────────────────────┐
│ TOKEN LAYER - TokenManager.Verify()                          │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐    │
│ │ 1. Verify JWT signature                             │    │
│ │ 2. Check expiration                                 │    │
│ │ 3. Check revocation list                            │    │
│ │ 4. Extract claims from JWT payload                  │    │
│ │ 5. Return VerificationResult with claims            │    │
│ └──────────────────────────────────────────────────────┘    │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           │ 9. Build identity from claims
                           ▼
┌──────────────────────────────────────────────────────────────┐
│ IDENTITY LAYER - IdentityResolver                            │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐    │
│ │ 1. IdentityResolver.Resolve(claims)                 │    │
│ │    → Create Subject:                                │    │
│ │      {                                              │    │
│ │        "id": "alice",                               │    │
│ │        "tenant_id": "demo-tenant",                  │    │
│ │        "type": "user",                              │    │
│ │        "principal": "alice",                        │    │
│ │        "attributes": { "email": "..." }             │    │
│ │      }                                              │    │
│ └──────────────────────────────────────────────────────┘    │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐    │
│ │ 2. IdentityBuilder.Build(subject)                   │    │
│ │    → Load roles (tenant+app scoped):                │    │
│ │      RoleProvider.GetRoles(tenant, app, subject)    │    │
│ │      → ["admin", "editor"]                          │    │
│ │                                                     │    │
│ │    → Load permissions (tenant+app scoped):          │    │
│ │      PermProvider.GetPermissions(tenant, app, sub)  │    │
│ │      → ["read", "write", "delete", ...]             │    │
│ │                                                     │    │
│ │    → Load groups (tenant scoped):                   │    │
│ │      GroupProvider.GetGroups(tenant, subject)       │    │
│ │      → ["admins", "staff"]                          │    │
│ │                                                     │    │
│ │    → Load profile (tenant scoped):                  │    │
│ │      ProfileProvider.GetProfile(tenant, subject)    │    │
│ │      → { "name": "Alice", "dept": "Eng" }           │    │
│ │                                                     │    │
│ │    → Build IdentityContext:                         │    │
│ │      {                                              │    │
│ │        "subject": {...},                            │    │
│ │        "tenant_id": "demo-tenant",                  │    │
│ │        "app_id": "demo-app",                        │    │
│ │        "roles": ["admin", "editor"],                │    │
│ │        "permissions": ["read", "write", ...],       │    │
│ │        "groups": ["admins", "staff"],               │    │
│ │        "profile": {...}                             │    │
│ │      }                                              │    │
│ └──────────────────────────────────────────────────────┘    │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           │ 10. Inject identity into request context
                           │     c.Set("identity", identityContext)
                           ▼
┌──────────────────────────────────────────────────────────────┐
│ MIDDLEWARE LAYER - RoleMiddleware (optional)                 │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐    │
│ │ 1. Get IdentityContext from request context         │    │
│ │ 2. Check if identity has required role              │    │
│ │    identity.HasAnyRole("admin")                     │    │
│ │ 3. If YES → Continue to next middleware/handler     │    │
│ │ 4. If NO  → Return 403 Forbidden                    │    │
│ └──────────────────────────────────────────────────────┘    │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           │ 11. Authorization passed
                           ▼
┌──────────────────────────────────────────────────────────────┐
│ MIDDLEWARE LAYER - PermissionMiddleware (optional)           │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐    │
│ │ 1. Get IdentityContext from request context         │    │
│ │ 2. Check if identity has required permissions       │    │
│ │    identity.HasAllPermissions("document:read")      │    │
│ │ 3. If YES → Continue to handler                     │    │
│ │ 4. If NO  → Return 403 Forbidden                    │    │
│ └──────────────────────────────────────────────────────┘    │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           │ 12. All checks passed
                           ▼
┌──────────────────────────────────────────────────────────────┐
│ APPLICATION HANDLER                                          │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐    │
│ │ func ProtectedHandler(c *request.Context) error {   │    │
│ │     // Get identity from context                    │    │
│ │     identity := middleware.MustGetIdentity(c)       │    │
│ │                                                     │    │
│ │     // Can perform additional checks               │    │
│ │     if !identity.HasPermission("special") {        │    │
│ │         return errors.New("not allowed")           │    │
│ │     }                                              │    │
│ │                                                     │    │
│ │     // Execute business logic                      │    │
│ │     result := doSomething(identity)                │    │
│ │                                                     │    │
│ │     return c.Resp.Json(result)                     │    │
│ │ }                                                  │    │
│ └──────────────────────────────────────────────────────┘    │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           │ 13. Return response
                           ▼
┌─────────────┐
│   CLIENT    │
│ Receives:   │
│ {                                                            │
│   "message": "...",                                          │
│   "user_id": "alice",                                        │
│   "tenant_id": "demo-tenant",                                │
│   "roles": ["admin", "editor"],                              │
│   "permissions": ["read", "write", ...]                      │
│ }                                                            │
└─────────────┘
```

## Key Points

### 1. Token is Minimal
Token hanya berisi **identifier** dan **basic claims**:
- ✅ Small size (200-300 bytes)
- ✅ Fast transmission
- ✅ Can't be too stale (15 min expiry)

### 2. Identity is Built Fresh
Identity context di-build **setiap request** dari database:
- ✅ Always fresh data
- ✅ Immediate revocation (role/permission changes take effect immediately)
- ✅ Can add caching (5-min TTL) for performance

### 3. Multi-Layer Security
```
Layer 1: Authentication (Credential)
         → Who are you? (verify username/password)

Layer 2: Token Verification (Token)
         → Is your token valid? (signature, expiry, revocation)

Layer 3: Identity Building (Identity)
         → What context do you have? (roles, permissions, profile)

Layer 4: Authorization (Authz)
         → What are you allowed to do? (RBAC, PBAC, ABAC)

Layer 5: Business Logic (Handler)
         → Execute the actual operation
```

### 4. Multi-Tenant Isolation
Every layer enforces **tenant + app** isolation:
- Token claims include `tenant_id` + `app_id`
- Identity providers keyed by `{tenant}:{app}:{user}`
- Authorization scoped to tenant + app
- Database queries filtered by tenant_id

## Performance Optimization

### Option 1: No Caching (Always Fresh)
```
Request → Verify Token (1ms) → Load Identity from DB (10ms) → Authz (1ms)
Total: ~12ms per request
Stale window: 0 seconds (always fresh!)
```

### Option 2: With Caching (Recommended)
```
Request 1: Verify Token (1ms) → Load Identity from DB (10ms) → Cache (0.1ms) → Authz (1ms)
           Total: ~12ms

Request 2-N (within 5 min): Verify Token (1ms) → Load from Cache (0.1ms) → Authz (1ms)
                            Total: ~2ms

Stale window: Up to 5 minutes (acceptable for most use cases)
```

### Cache Invalidation
```go
// When role/permission changes, invalidate cache
func RevokeUserRole(tenantID, appID, userID, role string) {
    // 1. Remove from database
    roleProvider.RemoveRole(tenantID, appID, userID, role)
    
    // 2. Invalidate cache
    cacheKey := fmt.Sprintf("identity:%s:%s:%s", tenantID, appID, userID)
    identityCache.Delete(cacheKey)
    
    // Next request will load fresh data from DB
}
```

## Security Benefits

### Immediate Revocation
```
09:00 - User login → Token issued (expires 09:15)
09:05 - Admin revoke user's "admin" role in database
09:06 - User makes request:
        ✅ Token still valid (not expired)
        ✅ Identity loaded from DB → roles = ["editor"] (no "admin")
        ❌ Access to admin endpoint DENIED
```

**With embedded roles in token:**
```
09:00 - User login → Token with roles: ["admin"] (expires 09:15)
09:05 - Admin revoke user's "admin" role in database
09:06 - User makes request:
        ✅ Token still valid with roles: ["admin"]
        ✅ Admin endpoint access ALLOWED (security breach!)
09:15 - Token expires → Finally blocked
❌ Security breach window: 10 minutes
```

## Conclusion

**Credential → Token → Identity Resolver → Authz** provides:

1. ✅ **Security:** Immediate revocation, fresh data
2. ✅ **Performance:** Small tokens, caching support
3. ✅ **Flexibility:** Easy to modify roles/permissions
4. ✅ **Multi-tenant:** Isolation at every layer
5. ✅ **Scalability:** Can add caching, load balancing
6. ✅ **Maintainability:** Clean separation of concerns

This is the **correct** and **production-ready** approach! 🚀
