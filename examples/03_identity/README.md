# Subject Resolution Examples

This directory contains examples demonstrating the subject resolution and identity context building capabilities of Layer 03.

## Overview

Layer 03 (Subject) transforms authentication tokens and claims into rich identity contexts that include:
- Subject information (ID, type, principal)
- Roles and permissions
- Group memberships
- Profile data
- Session information
- Metadata

## Examples

### 1. Simple Subject Resolution (`01_simple/`)

Demonstrates basic subject resolution from claims:
- Resolving subject from JWT claims
- Subject with custom types (user, service, bot)
- Subject with additional attributes
- Basic identity context building

**Run**:
```bash
go run examples/rbac/01_simple/main.go
```

**Key Concepts**:
- `SimpleResolver` - Direct claim-to-subject mapping
- Subject types: `user`, `service`, `bot`
- Subject attributes from claims
- Basic identity context creation

**Output Shows**:
- Subject ID, type, and principal extraction
- Custom subject types (service accounts, bots)
- Attribute mapping from claims
- Identity context with roles and permissions

### 2. Enriched Identity Context (`02_enriched/`)

Demonstrates identity enrichment with external data:
- Role and permission loading
- Group membership resolution
- Profile data enrichment
- Session information
- Metadata attachment

**Run**:
```bash
go run examples/rbac/02_enriched/main.go
```

**Key Concepts**:
- `EnrichedContextBuilder` - Identity enrichment
- Static providers for roles, permissions, groups, profiles
- Identity context helper methods
- Role and permission checking

**Features Demonstrated**:
- Admin with wildcard permissions (`users:*`, `projects:*`)
- Regular user with limited permissions
- Multiple roles per user
- Group memberships
- Profile data (department, level, full name)
- Permission checking methods:
  - `HasPermission()`
  - `HasAnyPermission()`
  - `HasAllPermissions()`
- Role checking methods:
  - `HasRole()`
  - `HasAnyRole()`
  - `HasAllRoles()`

**Output Shows**:
- Complete identity context with all enrichments
- Admin user with full permissions
- Regular user with limited access
- Helper method results

### 3. Cached & Stored Identity (`03_cached_store/`)

Demonstrates performance optimization and persistence:
- Cached subject resolution
- Identity store (user database)
- Cache hit/miss metrics
- Performance comparison
- Identity lifecycle management

**Run**:
```bash
go run examples/rbac/03_cached_store/main.go
```

**Key Concepts**:
- `CachedResolver` - Performance optimization with TTL
- `InMemoryIdentityStore` - User data persistence
- Cache statistics and metrics
- Performance timing

**Features Demonstrated**:
- Cache miss on first access
- Cache hit on subsequent access
- TTL-based cache expiration
- Identity store CRUD operations:
  - Create/Store identity
  - Get by ID
  - Update identity
  - Delete identity
  - List all identities
- Performance metrics

**Output Shows**:
- Cache performance (miss vs hit)
- Resolution time comparison
- Identity store operations
- Complete identity management lifecycle

## Running Examples

Each example is a standalone Go program in its own subfolder:

```bash
# Run simple example
go run examples/rbac/01_simple/main.go

# Run enriched example
go run examples/rbac/02_enriched/main.go

# Run cached & stored example
go run examples/rbac/03_cached_store/main.go
```

Or run all examples:

```bash
# PowerShell
Get-ChildItem examples/rbac/*/main.go | ForEach-Object { 
    Write-Host "Running $_"
    go run $_.FullName
    Write-Host ""
}

# Bash
for dir in examples/rbac/*/; do
    echo "Running ${dir}main.go"
    go run "${dir}main.go"
    echo ""
done
```

## Example Structure

Each example follows this pattern:

1. **Setup**: Create resolvers and providers
2. **Test Cases**: Multiple scenarios demonstrating features
3. **Output**: Clear display of:
   - Subject information
   - Identity context
   - Enriched data (roles, permissions, groups, profile)
   - Helper method results
   - Performance metrics (where applicable)

## Common Patterns

### Basic Subject Resolution

```go
import (
    "github.com/primadi/lokstra-auth/identity/simple"
)

resolver := simple.NewResolver()

claims := map[string]any{
    "sub":      "user123",
    "username": "john_doe",
    "email":    "john@example.com",
    "type":     "user",
}

subject, err := resolver.Resolve(ctx, claims)
```

### Enriched Identity Context

```go
import (
    "github.com/primadi/lokstra-auth/identity/enriched"
    "github.com/primadi/lokstra-auth/identity/simple"
)

// Setup providers
roleProvider := simple.NewStaticRoleProvider(roles)
permProvider := simple.NewStaticPermissionProvider(permissions)
groupProvider := simple.NewStaticGroupProvider(groups)
profileProvider := simple.NewStaticProfileProvider(profiles)

// Create enriched builder
builder := enriched.NewBuilder(
    resolver,
    roleProvider,
    permProvider,
    groupProvider,
    profileProvider,
)

// Build identity context
identity, err := builder.BuildIdentityContext(ctx, claims, nil)
```

### Cached Resolution

```go
import (
    "github.com/primadi/lokstra-auth/identity/cached"
    "time"
)

cache := cached.NewInMemoryCache()
cachedResolver := cached.NewResolver(
    baseResolver,
    cache,
    5*time.Minute, // TTL
)

subject, err := cachedResolver.Resolve(ctx, claims)
```

### Identity Store

```go
import (
    "github.com/primadi/lokstra-auth/identity/simple"
)

store := simple.NewInMemoryIdentityStore()

// Create identity
identity := &subject.IdentityContext{
    Subject: &subject.Subject{ID: "user123"},
    Roles:   []string{"admin"},
}
store.Store(ctx, identity)

// Retrieve identity
retrieved, err := store.Get(ctx, "user123")

// List all identities
all, err := store.List(ctx)
```

## Key Components Used

### Resolvers
- **SimpleResolver**: Direct claim-to-subject mapping
- **CachedResolver**: Adds caching layer with TTL

### Builders
- **SimpleContextBuilder**: Basic identity context creation
- **EnrichedContextBuilder**: Full identity enrichment

### Providers
- **StaticRoleProvider**: In-memory role assignments
- **StaticPermissionProvider**: In-memory permission assignments
- **StaticGroupProvider**: In-memory group memberships
- **StaticProfileProvider**: In-memory profile data

### Storage
- **InMemoryCache**: Temporary subject caching
- **InMemoryIdentityStore**: Persistent identity storage

## Testing Different Scenarios

Each example demonstrates:

### Subject Types
- **User**: Regular user accounts
- **Service**: Service accounts for APIs
- **Bot**: Automated bot accounts

### Permission Patterns
- **Wildcard**: `users:*`, `projects:*` (admin access)
- **Specific**: `projects:read`, `users:create`
- **Resource-specific**: `project:123:edit`

### Role Hierarchies
- **Admin**: Full access with wildcard permissions
- **Developer**: Code and project access
- **User**: Read-only access

### Performance Optimization
- **Cache Hit**: ~0ms (instant)
- **Cache Miss**: ~0-1ms (first time)
- **TTL Expiration**: Automatic cache invalidation

## Helper Methods

All examples demonstrate these helper methods:

### Permission Checking
```go
identity.HasPermission("users:read")              // Single permission
identity.HasAnyPermission("users:read", "admin")  // Any of (OR)
identity.HasAllPermissions("read", "write")       // All of (AND)
```

### Role Checking
```go
identity.HasRole("admin")                         // Single role
identity.HasAnyRole("admin", "developer")         // Any of (OR)
identity.HasAllRoles("user", "verified")          // All of (AND)
```

### Group Checking
```go
identity.HasGroup("engineering")                  // Single group
identity.HasAnyGroup("eng", "product")            // Any of (OR)
identity.HasAllGroups("eng", "team-a")            // All of (AND)
```

## Integration with Other Layers

### From Layer 02 (Token)
```go
// Token verification produces claims
claims, err := tokenManager.Verify(ctx, token)

// Layer 03 converts claims to identity
identity, err := builder.BuildIdentityContext(ctx, claims, nil)
```

### To Layer 04 (Authorization)
```go
// Identity context used in authorization
decision, err := authorizer.Evaluate(ctx, &authz.AuthorizationRequest{
    Subject:  identity,  // From Layer 03
    Action:   "read",
    Resource: resource,
})
```

## Best Practices Demonstrated

1. **Separation of Concerns**: Resolvers handle subject extraction, builders handle enrichment
2. **Provider Pattern**: Modular data sources for roles, permissions, groups, profiles
3. **Caching**: Performance optimization for frequently accessed identities
4. **Helper Methods**: Convenient permission and role checking
5. **Type Safety**: Strong typing for subjects and identity contexts
6. **Error Handling**: Proper error handling in all examples
7. **Context Usage**: All operations use context.Context

## Performance Considerations

### Simple Resolution
- **Speed**: Very fast (~0-1ms)
- **Memory**: Minimal (no caching)
- **Use Case**: Simple apps, infrequent access

### Enriched Resolution
- **Speed**: Fast (~1-2ms with static providers)
- **Memory**: Moderate (multiple providers)
- **Use Case**: Apps needing rich identity data

### Cached Resolution
- **Speed**: Instant on cache hit (~0ms)
- **Memory**: Higher (caches identities)
- **Use Case**: High-traffic apps, frequent access

### Identity Store
- **Speed**: Fast for in-memory (~0-1ms)
- **Memory**: Stores all identities
- **Use Case**: Persistent user data, CRUD operations

## Understanding Output

Example output format:

```
=== Example Name ===

1️⃣  Test Case 1...
✅ Result:
   Field: Value
   Field: Value

2️⃣  Test Case 2...
✅ Result:
   Field: Value
```

Each test shows:
- **Subject Info**: ID, type, principal, attributes
- **Identity Context**: Roles, permissions, groups, profile
- **Helper Results**: Permission/role check outcomes
- **Performance**: Timing and cache statistics (where applicable)

## Troubleshooting

If an example doesn't compile:

1. **Check imports**: Ensure all packages are imported correctly
2. **Run go mod tidy**: Update dependencies
3. **Check Go version**: Requires Go 1.21+

If identity resolution doesn't work as expected:

1. **Check claim format**: Ensure claims contain required fields (`sub`)
2. **Verify providers**: Check that static providers have data for the subject ID
3. **Review output**: Look for error messages or unexpected values

## Next Steps

After understanding these examples:

1. **Custom Providers**: Implement database-backed providers
2. **Dynamic Enrichment**: Load data from external APIs
3. **Advanced Caching**: Use Redis or other distributed caches
4. **Session Management**: Add session tracking and expiration
5. **Audit Logging**: Log all identity resolutions
6. **Multi-tenancy**: Add tenant isolation to identity contexts

## Additional Examples

You can create additional examples for:

- External API-based enrichment (LDAP, Active Directory)
- Database-backed identity store (PostgreSQL, MongoDB)
- Redis-backed caching
- Dynamic permission loading
- Hierarchical role systems
- Custom claim mapping
- Multi-tenant identity isolation
