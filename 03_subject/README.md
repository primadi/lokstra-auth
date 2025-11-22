# Layer 03: Subject Resolution & Identity Context

Layer 03 Subject Resolution provides complete implementation for transforming authentication claims into rich identity contexts with roles, permissions, groups, and profile data. This layer bridges the gap between token verification and authorization.

## 📋 Table of Contents

- [Concepts](#concepts)
- [Architecture](#architecture)
- [Core Components](#core-components)
  - [Subject Resolver](#subject-resolver)
  - [Identity Context Builder](#identity-context-builder)
  - [Data Providers](#data-providers)
- [Implementations](#implementations)
  - [Simple Resolver](#simple-resolver)
  - [Enriched Builder](#enriched-builder)
  - [Cached Resolver](#cached-resolver)
- [Identity Store](#identity-store)
- [Use Cases](#use-cases)
- [Examples](#examples)
- [Best Practices](#best-practices)

---

## Concepts

### What is Subject Resolution?

Subject resolution is the process of converting raw authentication claims (from tokens) into a structured `Subject` entity representing an authenticated user, service, or device.

### What is Identity Context?

Identity Context is the complete set of information about an authenticated subject, including:
- **Subject**: Core identity (ID, type, principal)
- **Roles**: Role-based access control assignments
- **Permissions**: Fine-grained permission grants
- **Groups**: Group memberships
- **Profile**: User profile data
- **Session**: Session-specific information
- **Metadata**: Additional contextual data

### The Flow

```
Token Claims → Subject → Identity Context → Authorization
   (Layer 2)   (Layer 3)    (Layer 3)        (Layer 4)
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│          Subject Resolution & Identity Layer            │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────────┐      ┌──────────────────────┐    │
│  │ Subject Resolver │      │ Context Builder      │    │
│  │  - Simple        │      │  - Simple            │    │
│  │  - Cached        │      │  - Enriched          │    │
│  └────────┬─────────┘      └─────────┬────────────┘    │
│           │                           │                  │
│           │                           │                  │
│           │         ┌─────────────────▼────────┐        │
│           │         │  Data Providers          │        │
│           │         │  - RoleProvider          │        │
│           │         │  - PermissionProvider    │        │
│           │         │  - GroupProvider         │        │
│           │         │  - ProfileProvider       │        │
│           │         └──────────────────────────┘        │
│           │                                              │
│           │         ┌──────────────────────────┐        │
│           └────────►│  Identity Store/Cache    │        │
│                     └──────────────────────────┘        │
│                                                           │
├─────────────────────────────────────────────────────────┤
│                  Contract Interfaces                     │
│  • SubjectResolver                                       │
│  • IdentityContextBuilder                               │
│  • RoleProvider, PermissionProvider                     │
│  • GroupProvider, ProfileProvider                       │
│  • IdentityStore, IdentityCache                         │
│  • DataEnricher                                          │
└─────────────────────────────────────────────────────────┘
```

---

## Core Components

### Subject Resolver

Transforms token claims into a `Subject` entity.

**Interface:**
```go
type SubjectResolver interface {
    Resolve(ctx context.Context, claims map[string]any) (*Subject, error)
}
```

**Subject Structure:**
```go
type Subject struct {
    ID         string         // Unique identifier
    Type       string         // Subject type (user, service, device)
    Principal  string         // Primary identifier (username, email)
    Attributes map[string]any // Additional attributes
}
```

### Identity Context Builder

Builds complete identity context from a subject.

**Interface:**
```go
type IdentityContextBuilder interface {
    Build(ctx context.Context, subject *Subject) (*IdentityContext, error)
}
```

**IdentityContext Structure:**
```go
type IdentityContext struct {
    Subject     *Subject              // The subject
    Roles       []string              // User roles
    Permissions []string              // User permissions
    Groups      []string              // Group memberships
    Profile     map[string]any        // Profile data
    Session     *SessionInfo          // Session info
    Metadata    map[string]any        // Additional metadata
}
```

**Helper Methods:**
```go
func (ic *IdentityContext) HasRole(role string) bool
func (ic *IdentityContext) HasPermission(permission string) bool
func (ic *IdentityContext) HasAnyRole(roles ...string) bool
func (ic *IdentityContext) HasAllRoles(roles ...string) bool
```

### Data Providers

Providers supply various aspects of identity context:

**RoleProvider:**
```go
type RoleProvider interface {
    GetRoles(ctx context.Context, subject *Subject) ([]string, error)
}
```

**PermissionProvider:**
```go
type PermissionProvider interface {
    GetPermissions(ctx context.Context, subject *Subject) ([]string, error)
}
```

**GroupProvider:**
```go
type GroupProvider interface {
    GetGroups(ctx context.Context, subject *Subject) ([]string, error)
}
```

**ProfileProvider:**
```go
type ProfileProvider interface {
    GetProfile(ctx context.Context, subject *Subject) (map[string]any, error)
}
```

---

## Implementations

### Simple Resolver

Direct mapping from claims to subject with configurable claim keys.

**Features:**
- ✅ Configurable claim mapping
- ✅ Default fallback values
- ✅ Attribute extraction
- ✅ Type-safe claim parsing

**Configuration:**
```go
type Resolver struct {
    SubjectIDClaim     string // Claim key for subject ID (default: "sub")
    SubjectTypeClaim   string // Claim key for type (default: "type")
    PrincipalClaim     string // Claim key for principal (default: "username")
    DefaultSubjectType string // Default type (default: "user")
}
```

**Basic Usage:**
```go
import "github.com/primadi/lokstra-auth/03_subject/simple"

// Create resolver
resolver := simple.NewResolver()

// Custom configuration
resolver.SubjectIDClaim = "user_id"
resolver.PrincipalClaim = "email"
resolver.DefaultSubjectType = "user"

// Resolve subject from claims
claims := map[string]any{
    "sub":      "user-123",
    "email":    "john@example.com",
    "name":     "John Doe",
    "age":      30,
    "verified": true,
}

subject, err := resolver.Resolve(ctx, claims)
// subject.ID = "user-123"
// subject.Principal = "john@example.com"
// subject.Attributes = {"name": "John Doe", "age": 30, "verified": true}
```

### Simple Context Builder

Builds identity context using static data providers.

**Features:**
- ✅ Static role/permission/group/profile providers
- ✅ Session info support
- ✅ Metadata support
- ✅ In-memory implementations for testing

**Basic Usage:**
```go
import "github.com/primadi/lokstra-auth/03_subject/simple"

// Create static providers
roleProvider := simple.NewStaticRoleProvider(map[string][]string{
    "user-123": {"admin", "user"},
    "user-456": {"user"},
})

permProvider := simple.NewStaticPermissionProvider(map[string][]string{
    "user-123": {"read:users", "write:users", "delete:users"},
    "user-456": {"read:users"},
})

groupProvider := simple.NewStaticGroupProvider(map[string][]string{
    "user-123": {"engineering", "management"},
})

profileProvider := simple.NewStaticProfileProvider(map[string]map[string]any{
    "user-123": {
        "name":       "John Doe",
        "email":      "john@example.com",
        "department": "Engineering",
    },
})

// Create context builder
builder := simple.NewContextBuilder(
    roleProvider,
    permProvider,
    groupProvider,
    profileProvider,
)

// Build identity context
identity, err := builder.Build(ctx, subject)
// identity.Roles = ["admin", "user"]
// identity.Permissions = ["read:users", "write:users", "delete:users"]
// identity.Groups = ["engineering", "management"]
// identity.Profile = {"name": "John Doe", ...}
```

**With Session Info:**
```go
// Create context builder with session
builder := simple.NewContextBuilder(
    roleProvider,
    permProvider,
    groupProvider,
    profileProvider,
)

// Set session info
builder.WithSession(&subject.SessionInfo{
    ID:        "session-abc123",
    CreatedAt: time.Now().Unix(),
    ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
    IPAddress: "192.168.1.100",
    UserAgent: "Mozilla/5.0...",
})

identity, err := builder.Build(ctx, subject)
// identity.Session = {ID: "session-abc123", ...}
```

---

### Enriched Builder

Wraps a base builder and applies data enrichers for external data integration.

**Features:**
- ✅ Chainable enrichers
- ✅ Database integration support
- ✅ API integration support
- ✅ Attribute mapping
- ✅ Custom enrichment logic

**Basic Usage:**
```go
import "github.com/primadi/lokstra-auth/03_subject/enriched"

// Create base builder
baseBuilder := simple.NewContextBuilder(roleProvider, permProvider, nil, nil)

// Create enrichers
attrEnricher := enriched.NewAttributeEnricher()
attrEnricher.AttributeMapping = map[string]string{
    "email":    "user_email",
    "verified": "email_verified",
}

// Create enriched builder
enrichedBuilder := enriched.NewContextBuilder(
    baseBuilder,
    attrEnricher,
    // Add more enrichers...
)

// Build enriched identity
identity, err := enrichedBuilder.Build(ctx, subject)
// identity.Metadata["user_email"] = "john@example.com"
// identity.Metadata["email_verified"] = true
```

**Custom Enricher:**
```go
// Database enricher
type DatabaseEnricher struct {
    db *sql.DB
}

func (e *DatabaseEnricher) Enrich(ctx context.Context, identity *subject.IdentityContext) error {
    if identity.Subject == nil {
        return nil
    }

    // Load additional data from database
    var profile map[string]any
    row := e.db.QueryRowContext(ctx,
        "SELECT department, manager, hire_date FROM employees WHERE user_id = $1",
        identity.Subject.ID,
    )
    
    var dept, manager string
    var hireDate time.Time
    if err := row.Scan(&dept, &manager, &hireDate); err != nil {
        return err
    }

    // Enrich profile
    if identity.Profile == nil {
        identity.Profile = make(map[string]any)
    }
    identity.Profile["department"] = dept
    identity.Profile["manager"] = manager
    identity.Profile["hire_date"] = hireDate.Format("2006-01-02")

    return nil
}

// Use it
enrichedBuilder := enriched.NewContextBuilder(
    baseBuilder,
    &DatabaseEnricher{db: myDB},
)
```

**API Enricher Example:**
```go
type APIEnricher struct {
    apiURL string
    client *http.Client
}

func (e *APIEnricher) Enrich(ctx context.Context, identity *subject.IdentityContext) error {
    // Fetch data from external API
    url := fmt.Sprintf("%s/users/%s/preferences", e.apiURL, identity.Subject.ID)
    resp, err := e.client.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var prefs map[string]any
    if err := json.NewDecoder(resp.Body).Decode(&prefs); err != nil {
        return err
    }

    // Add to metadata
    if identity.Metadata == nil {
        identity.Metadata = make(map[string]any)
    }
    identity.Metadata["preferences"] = prefs

    return nil
}
```

---

### Cached Resolver

Wraps resolver/builder with caching for performance optimization.

**Features:**
- ✅ In-memory caching
- ✅ Configurable TTL
- ✅ Cache invalidation
- ✅ Reduced database/API calls
- ✅ Thread-safe

**Cached Resolver:**
```go
import "github.com/primadi/lokstra-auth/03_subject/cached"

// Create base resolver
baseResolver := simple.NewResolver()

// Create cached resolver
cachedResolver := cached.NewResolver(
    baseResolver,
    nil,              // nil = use default in-memory cache
    5 * time.Minute,  // TTL
)

// First call - cache miss, resolves from base
subject1, err := cachedResolver.Resolve(ctx, claims)

// Second call - cache hit, instant return
subject2, err := cachedResolver.Resolve(ctx, claims)
```

**Cached Context Builder:**
```go
// Create base builder
baseBuilder := simple.NewContextBuilder(
    roleProvider,
    permProvider,
    groupProvider,
    profileProvider,
)

// Create cached builder
cachedBuilder := cached.NewContextBuilder(
    baseBuilder,
    nil,              // nil = use default in-memory cache
    5 * time.Minute,  // TTL
)

// Build identity with caching
identity, err := cachedBuilder.Build(ctx, subject)
```

**Custom Cache Implementation:**
```go
// Redis cache example
type RedisCache struct {
    client *redis.Client
}

func (c *RedisCache) Set(ctx context.Context, key string, identity *subject.IdentityContext, ttl int64) error {
    data, err := json.Marshal(identity)
    if err != nil {
        return err
    }
    
    return c.client.Set(ctx, key, data, time.Duration(ttl)*time.Second).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string) (*subject.IdentityContext, error) {
    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil, err
    }
    
    var identity subject.IdentityContext
    if err := json.Unmarshal(data, &identity); err != nil {
        return nil, err
    }
    
    return &identity, nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) Clear(ctx context.Context) error {
    return c.client.FlushDB(ctx).Err()
}

// Use Redis cache
redisCache := &RedisCache{client: redisClient}
cachedResolver := cached.NewResolver(baseResolver, redisCache, 10*time.Minute)
```

---

## Identity Store

Persistent storage for identity contexts, useful for session management.

**Features:**
- ✅ Store identity contexts by session ID
- ✅ Session expiration support
- ✅ List sessions by subject
- ✅ Bulk delete by subject
- ✅ Automatic cleanup
- ✅ Thread-safe in-memory implementation

**Basic Usage:**
```go
import subject "github.com/primadi/lokstra-auth/03_subject"

// Create identity store
store := subject.NewInMemoryIdentityStore()

// Store identity context
err := store.Store(ctx, "session-123", identity)

// Retrieve identity context
identity, err := store.Get(ctx, "session-123")

// List all sessions for a user
identities, err := store.ListBySubject(ctx, "user-123")
fmt.Printf("User has %d active sessions\n", len(identities))

// Delete specific session
err := store.Delete(ctx, "session-123")

// Delete all sessions for a user
err := store.DeleteBySubject(ctx, "user-123")

// Manual cleanup of expired sessions
err := store.Cleanup(ctx)
```

**With Custom Expiration:**
```go
// Set custom session expiration
identity.Session = &subject.SessionInfo{
    ID:        "session-abc",
    CreatedAt: time.Now().Unix(),
    ExpiresAt: time.Now().Add(12 * time.Hour).Unix(), // 12 hours
    IPAddress: "192.168.1.100",
}

store.Store(ctx, "session-abc", identity)

// Auto-cleanup runs every 5 minutes
// Sessions past ExpiresAt are automatically removed
```

---

## Use Cases

### 1. Basic Authentication Flow

```go
// Step 1: Verify token (Layer 2)
tokenResult, err := tokenManager.Verify(ctx, tokenValue)
claims := tokenResult.Claims

// Step 2: Resolve subject (Layer 3)
resolver := simple.NewResolver()
subject, err := resolver.Resolve(ctx, claims)

// Step 3: Build identity context (Layer 3)
builder := simple.NewContextBuilder(roleProvider, permProvider, nil, nil)
identity, err := builder.Build(ctx, subject)

// Step 4: Check permissions
if identity.HasRole("admin") {
    // Allow admin access
}
```

### 2. Multi-Tenant Application

```go
// Resolve subject with tenant info
claims := map[string]any{
    "sub":       "user-123",
    "email":     "john@example.com",
    "tenant_id": "tenant-abc",
}

subject, _ := resolver.Resolve(ctx, claims)
// subject.Attributes["tenant_id"] = "tenant-abc"

// Load tenant-specific roles
type TenantRoleProvider struct {
    db *sql.DB
}

func (p *TenantRoleProvider) GetRoles(ctx context.Context, sub *subject.Subject) ([]string, error) {
    tenantID := sub.Attributes["tenant_id"].(string)
    
    rows, err := p.db.QueryContext(ctx,
        "SELECT role FROM tenant_user_roles WHERE tenant_id = $1 AND user_id = $2",
        tenantID, sub.ID,
    )
    // ... parse and return roles
}
```

### 3. Session Management

```go
// Login flow
identity, _ := builder.Build(ctx, subject)
identity.Session = &subject.SessionInfo{
    ID:        uuid.New().String(),
    CreatedAt: time.Now().Unix(),
    ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
    IPAddress: request.RemoteAddr,
    UserAgent: request.UserAgent(),
}

// Store session
store.Store(ctx, identity.Session.ID, identity)

// List active sessions
sessions, _ := store.ListBySubject(ctx, subject.ID)
for _, sess := range sessions {
    fmt.Printf("Session: %s, IP: %s, Created: %v\n",
        sess.Session.ID,
        sess.Session.IPAddress,
        time.Unix(sess.Session.CreatedAt, 0),
    )
}

// Logout from all devices
store.DeleteBySubject(ctx, subject.ID)
```

### 4. Performance Optimization

```go
// Use caching for high-traffic applications
cachedResolver := cached.NewResolver(baseResolver, nil, 5*time.Minute)
cachedBuilder := cached.NewContextBuilder(baseBuilder, nil, 5*time.Minute)

// First request - database hit
identity1, _ := cachedBuilder.Build(ctx, subject)

// Subsequent requests within 5 minutes - cache hit
identity2, _ := cachedBuilder.Build(ctx, subject)
```

### 5. Dynamic Permission Loading

```go
type DynamicPermissionProvider struct {
    db *sql.DB
}

func (p *DynamicPermissionProvider) GetPermissions(ctx context.Context, sub *subject.Subject) ([]string, error) {
    // Load user's direct permissions
    directPerms, _ := p.loadDirectPermissions(ctx, sub.ID)
    
    // Load role-based permissions
    rolePerms, _ := p.loadRolePermissions(ctx, sub.ID)
    
    // Load group-based permissions
    groupPerms, _ := p.loadGroupPermissions(ctx, sub.ID)
    
    // Combine and deduplicate
    allPerms := append(directPerms, rolePerms...)
    allPerms = append(allPerms, groupPerms...)
    
    return deduplicateStrings(allPerms), nil
}
```

---

## Examples

See `examples/03_subject/` folder for complete examples:

### 01_simple - Simple Resolution
Basic subject resolution and identity building:
- Simple resolver usage
- Static providers
- Identity context creation

```bash
cd examples/03_subject/01_simple
go run main.go
```

### 02_enriched - Enriched Identity
Identity enrichment with external data:
- Attribute enrichment
- Custom enrichers
- Database integration patterns

```bash
cd examples/03_subject/02_enriched
go run main.go
```

### 03_cached_store - Caching & Storage
Performance optimization with caching:
- Cached resolver
- Cached builder
- Identity store usage
- Session management

```bash
cd examples/03_subject/03_cached_store
go run main.go
```

---

## Best Practices

### 1. Claim Validation

✅ **Good**:
```go
func (r *CustomResolver) Resolve(ctx context.Context, claims map[string]any) (*subject.Subject, error) {
    // Validate required claims
    subID, ok := claims["sub"].(string)
    if !ok || subID == "" {
        return nil, errors.New("missing subject ID")
    }
    
    // Type assertions with validation
    email, ok := claims["email"].(string)
    if !ok {
        email = "" // Use default
    }
    
    // ...
}
```

### 2. Provider Error Handling

✅ **Good**:
```go
func (p *DatabaseRoleProvider) GetRoles(ctx context.Context, sub *subject.Subject) ([]string, error) {
    roles, err := p.loadRoles(ctx, sub.ID)
    if err != nil {
        // Log error but return empty roles instead of failing
        log.Printf("Failed to load roles for %s: %v", sub.ID, err)
        return []string{}, nil
    }
    return roles, nil
}
```

### 3. Caching Strategy

✅ **Good**:
```go
// Cache identity context, not just subject
cachedBuilder := cached.NewContextBuilder(
    builder,
    nil,
    5 * time.Minute, // Reasonable TTL
)

// Invalidate cache on role/permission changes
func (s *Service) UpdateUserRoles(ctx context.Context, userID string, roles []string) error {
    // Update database
    err := s.db.UpdateRoles(ctx, userID, roles)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("identity:%s", userID)
    s.cache.Delete(ctx, cacheKey)
    
    return nil
}
```

### 4. Session Management

✅ **Good**:
```go
// Set reasonable session expiry
identity.Session = &subject.SessionInfo{
    ID:        sessionID,
    CreatedAt: time.Now().Unix(),
    ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // 24 hours
    IPAddress: req.RemoteAddr,
    UserAgent: req.UserAgent(),
    Metadata: map[string]any{
        "login_method": "oauth2",
        "device_type":  "mobile",
    },
}

// Implement session refresh
func RefreshSession(ctx context.Context, sessionID string) error {
    identity, err := store.Get(ctx, sessionID)
    if err != nil {
        return err
    }
    
    // Extend expiration
    identity.Session.ExpiresAt = time.Now().Add(24 * time.Hour).Unix()
    
    return store.Update(ctx, sessionID, identity)
}
```

### 5. Multi-Source Data

✅ **Good**:
```go
// Use enrichers for external data
builder := enriched.NewContextBuilder(
    baseBuilder,
    &DatabaseEnricher{db: db},
    &LDAPEnricher{ldap: ldapClient},
    &APIEnricher{api: externalAPI},
)

// Handle enrichment failures gracefully
func (e *ExternalEnricher) Enrich(ctx context.Context, identity *subject.IdentityContext) error {
    data, err := e.fetchExternalData(ctx, identity.Subject.ID)
    if err != nil {
        // Log but don't fail
        log.Printf("External enrichment failed: %v", err)
        return nil // Continue with other enrichers
    }
    
    identity.Metadata["external_data"] = data
    return nil
}
```

### 6. Testing

✅ **Good**:
```go
func TestIdentityBuilding(t *testing.T) {
    // Use static providers for testing
    roleProvider := simple.NewStaticRoleProvider(map[string][]string{
        "user-1": {"admin"},
    })
    
    builder := simple.NewContextBuilder(roleProvider, nil, nil, nil)
    
    subject := &subject.Subject{
        ID:        "user-1",
        Type:      "user",
        Principal: "admin@test.com",
    }
    
    identity, err := builder.Build(context.Background(), subject)
    assert.NoError(t, err)
    assert.True(t, identity.HasRole("admin"))
}
```

---

## Performance Considerations

### Caching Impact

**Without Cache:**
- Database query per request
- ~50-100ms latency per identity build

**With Cache (5min TTL):**
- First request: ~50-100ms
- Subsequent requests: <1ms
- 99%+ cache hit ratio for active users

### Recommendations

1. **Use caching** for production environments
2. **Set appropriate TTL** based on data freshness requirements
3. **Implement cache invalidation** for critical updates
4. **Monitor cache hit ratio** and adjust TTL accordingly
5. **Use Redis/Memcached** for distributed applications

---

## Security Considerations

1. **Claim Validation**: Always validate claims before use
2. **Sensitive Data**: Don't store passwords or tokens in identity context
3. **Session Security**: Use secure session IDs (UUIDs)
4. **Expiration**: Always set session expiration
5. **Audit Logging**: Log identity access for security monitoring
6. **Cache Security**: Ensure cache is not accessible externally

---

## Migration Guide

### From Basic to Enriched

**Before:**
```go
builder := simple.NewContextBuilder(roleProvider, nil, nil, nil)
identity, _ := builder.Build(ctx, subject)
```

**After:**
```go
enrichedBuilder := enriched.NewContextBuilder(
    simple.NewContextBuilder(roleProvider, nil, nil, nil),
    &AttributeEnricher{},
    &DatabaseEnricher{db: db},
)
identity, _ := enrichedBuilder.Build(ctx, subject)
```

### Adding Caching

**Before:**
```go
resolver := simple.NewResolver()
builder := simple.NewContextBuilder(roleProvider, permProvider, nil, nil)
```

**After:**
```go
resolver := cached.NewResolver(simple.NewResolver(), nil, 5*time.Minute)
builder := cached.NewContextBuilder(
    simple.NewContextBuilder(roleProvider, permProvider, nil, nil),
    nil,
    5*time.Minute,
)
```

---

## References

- [Subject-Based Access Control](https://en.wikipedia.org/wiki/Subject_(computer_security))
- [Identity and Access Management Best Practices](https://www.nist.gov/identity-access-management)
- [Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)

---

**Layer 03 Complete** ✅  
Continue to [Layer 04: Authorization](../04_authz/README.md) for access control.
