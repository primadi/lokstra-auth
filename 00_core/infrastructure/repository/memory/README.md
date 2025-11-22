# In-Memory Store Implementation

This directory contains the in-memory implementation of all core stores for the lokstra-auth framework.

## Purpose

In-memory stores are ideal for:
- **Development & Testing**: Fast and simple setup without database dependencies
- **Unit Tests**: Isolated testing without external dependencies
- **Prototyping**: Quick proof-of-concept implementations
- **Small Applications**: Low-traffic applications that don't require persistence

⚠️ **Not for Production**: Data is lost on application restart. Use PostgreSQL stores for production.

## Structure

The implementation is split into separate files for better maintainability:

```
memory/
├── app_key_store.go    - API key storage and management
├── app_store.go        - Application management
├── branch_store.go     - Branch/location management
├── tenant_store.go     - Tenant/organization management
├── user_app_store.go   - User-app access relationships
└── user_store.go       - User management
```

## Files

### `app_key_store.go`
**Type**: `InMemoryAppKeyStore`

Manages API keys for application authentication:
- Dual-index design: by ID and by KeyID for fast lookups
- Thread-safe with `sync.RWMutex`
- Supports key revocation and expiration
- Tracks last usage timestamps

**Methods**:
- `Store()` - Store new API key
- `GetByKeyID()` - Retrieve by key ID (for authentication)
- `GetByID()` - Retrieve by internal ID
- `GetByPrefix()` - List keys by prefix
- `ListByApp()` - List all keys for an app
- `ListByTenant()` - List all keys for a tenant
- `Update()` - Update key details
- `Revoke()` - Revoke a key
- `Delete()` - Permanently delete a key
- `UpdateLastUsed()` - Update last usage timestamp

### `tenant_store.go`
**Type**: `InMemoryTenantStore`

Manages multi-tenant organizations:
- Simple map-based storage: `map[tenantID]*Tenant`
- Thread-safe operations
- Support for tenant metadata and settings

**Methods**:
- `Create()` - Create new tenant
- `Get()` - Retrieve tenant by ID
- `Update()` - Update tenant details
- `Delete()` - Delete tenant
- `List()` - List all tenants
- `GetByName()` - Find tenant by name
- `Exists()` - Check if tenant exists

### `app_store.go`
**Type**: `InMemoryAppStore`

Manages applications within tenants:
- Composite key: `tenantID:appID`
- Multi-tenant data isolation
- Support for app types and configurations

**Methods**:
- `Create()` - Create new app
- `Get()` - Retrieve app by ID
- `Update()` - Update app details
- `Delete()` - Delete app
- `List()` - List all apps for a tenant
- `GetByName()` - Find app by name
- `Exists()` - Check if app exists
- `ListByType()` - Filter apps by type

### `user_store.go`
**Type**: `InMemoryUserStore`

Manages users within tenants:
- Composite key: `tenantID:userID`
- Support for username and email lookups
- Multi-tenant user isolation

**Methods**:
- `Create()` - Create new user
- `Get()` - Retrieve user by ID
- `Update()` - Update user details
- `Delete()` - Delete user
- `List()` - List all users for a tenant
- `GetByUsername()` - Find user by username
- `GetByEmail()` - Find user by email
- `Exists()` - Check if user exists
- `ListByApp()` - List users assigned to an app (placeholder)

### `branch_store.go`
**Type**: `InMemoryBranchStore`

Manages branches/locations within apps:
- Composite key: `tenantID:appID:branchID`
- Support for different branch types (HQ, store, warehouse, etc.)
- Thread-safe operations

**Methods**:
- `Create()` - Create new branch
- `Get()` - Retrieve branch by ID
- `Update()` - Update branch details
- `Delete()` - Delete branch
- `List()` - List all branches for an app
- `Exists()` - Check if branch exists
- `ListByType()` - Filter branches by type

### `user_app_store.go`
**Type**: `InMemoryUserAppStore`

Manages user access to applications:
- Nested map: `map[tenantID:userID]map[appID]bool`
- Simple boolean flags for access control
- Thread-safe operations

**Methods**:
- `GrantAccess()` - Grant user access to app
- `RevokeAccess()` - Revoke user access
- `HasAccess()` - Check if user has access
- `ListUserApps()` - List all apps user can access
- `ListAppUsers()` - List all users with access to app

## Usage

### Basic Usage

```go
import "github.com/primadi/lokstra-auth/00_core/infrastructure/repository/memory"

// Create stores
tenantStore := memory.NewInMemoryTenantStore()
appStore := memory.NewInMemoryAppStore()
userStore := memory.NewInMemoryUserStore()
branchStore := memory.NewInMemoryBranchStore()
userAppStore := memory.NewInMemoryUserAppStore()
appKeyStore := memory.NewInMemoryAppKeyStore()

// Use stores
ctx := context.Background()

// Create a tenant
tenant := &domain.Tenant{
    ID:   "acme-corp",
    Name: "Acme Corporation",
    // ... other fields
}
tenantStore.Create(ctx, tenant)

// Create an app
app := &domain.App{
    ID:       "web-portal",
    TenantID: "acme-corp",
    Name:     "Web Portal",
    Type:     domain.AppTypeWeb,
    // ... other fields
}
appStore.Create(ctx, app)

// Create a user
user := &domain.User{
    ID:       "user-001",
    TenantID: "acme-corp",
    Username: "john.doe",
    Email:    "john@acme.com",
    // ... other fields
}
userStore.Create(ctx, user)

// Grant user access to app
userAppStore.GrantAccess(ctx, "acme-corp", "web-portal", "user-001")
```

### Complete Example

```go
package main

import (
    "context"
    "log"
    
    "github.com/primadi/lokstra-auth/00_core/domain"
    "github.com/primadi/lokstra-auth/00_core/infrastructure/repository/memory"
)

func main() {
    ctx := context.Background()
    
    // Initialize stores
    stores := struct {
        Tenant  *memory.InMemoryTenantStore
        App     *memory.InMemoryAppStore
        User    *memory.InMemoryUserStore
        Branch  *memory.InMemoryBranchStore
        UserApp *memory.InMemoryUserAppStore
        AppKey  *memory.InMemoryAppKeyStore
    }{
        Tenant:  memory.NewInMemoryTenantStore(),
        App:     memory.NewInMemoryAppStore(),
        User:    memory.NewInMemoryUserStore(),
        Branch:  memory.NewInMemoryBranchStore(),
        UserApp: memory.NewInMemoryUserAppStore(),
        AppKey:  memory.NewInMemoryAppKeyStore(),
    }
    
    // Create and use entities
    tenant := &domain.Tenant{
        ID:   "demo",
        Name: "Demo Corp",
    }
    
    if err := stores.Tenant.Create(ctx, tenant); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Tenant created successfully!")
}
```

## Thread Safety

All stores use `sync.RWMutex` for thread-safe operations:

- **Read operations**: Multiple concurrent reads allowed
- **Write operations**: Exclusive access required
- **Performance**: Optimized for read-heavy workloads

Example of mutex usage:
```go
// Read operation
func (s *InMemoryTenantStore) Get(ctx context.Context, tenantID string) (*domain.Tenant, error) {
    s.mu.RLock()         // Acquire read lock
    defer s.mu.RUnlock() // Release on function exit
    
    // Safe concurrent reads
    tenant, exists := s.tenants[tenantID]
    // ...
}

// Write operation
func (s *InMemoryTenantStore) Create(ctx context.Context, tenant *domain.Tenant) error {
    s.mu.Lock()         // Acquire write lock (exclusive)
    defer s.mu.Unlock() // Release on function exit
    
    // Exclusive write access
    s.tenants[tenant.ID] = tenant
    // ...
}
```

## Data Structure

### Map Keys

The stores use different key strategies:

1. **Single Key**: `TenantStore`, `AppKeyStore` (by ID)
   ```go
   map[string]*Domain  // map[id]*Tenant
   ```

2. **Composite Key**: `AppStore`, `UserStore`, `BranchStore`
   ```go
   map[string]*Domain  // map["tenantID:entityID"]*Entity
   ```

3. **Nested Map**: `UserAppStore`
   ```go
   map[string]map[string]bool  // map["tenantID:userID"]map[appID]bool
   ```

### Composite Key Format

For entities that belong to a tenant, we use composite keys:

```go
key := tenantID + ":" + entityID
// Examples:
// "acme-corp:user-001"
// "acme-corp:web-portal"
// "acme-corp:web-portal:branch-hq"
```

This ensures:
- Multi-tenant data isolation
- Fast lookups without scanning
- Simple implementation

## Performance Characteristics

| Operation | Time Complexity | Space Complexity |
|-----------|----------------|------------------|
| Create    | O(1)           | O(1)             |
| Get       | O(1)           | O(1)             |
| Update    | O(1)           | O(1)             |
| Delete    | O(1)           | O(1)             |
| List      | O(n)           | O(n)             |
| Search    | O(n)           | O(1)             |

**Notes**:
- Direct lookups by ID are O(1) using map access
- List operations require scanning all items: O(n)
- Search by non-indexed fields (name, email) requires full scan: O(n)
- Memory usage grows linearly with data: O(n)

## Limitations

1. **No Persistence**: Data is lost on application restart
2. **Memory Bound**: All data must fit in RAM
3. **Single Instance**: Cannot share data between app instances
4. **No Transactions**: No ACID guarantees across operations
5. **Linear Search**: Non-indexed field lookups are slow for large datasets
6. **No Indexing**: Only primary keys are indexed

## When to Use

✅ **Good For**:
- Development and testing
- Unit tests
- Quick prototypes
- Small applications (< 10,000 entities)
- Read-heavy workloads with small datasets

❌ **Not Good For**:
- Production applications
- Applications requiring data persistence
- Multi-instance deployments
- Large datasets (> 100,000 entities)
- Applications requiring complex queries

## Migration to PostgreSQL

When you need persistence, migrate to PostgreSQL stores:

**Before**:
```go
tenantStore := memory.NewInMemoryTenantStore()
appStore := memory.NewInMemoryAppStore()
```

**After**:
```go
db, _ := postgres.NewConnection(config)
tenantStore := postgres.NewTenantStore(db, "public")
appStore := postgres.NewAppStore(db, "public")
```

See `POSTGRES_MIGRATION_GUIDE.md` for detailed migration steps.

## Testing

Use in-memory stores for fast unit tests:

```go
func TestTenantService(t *testing.T) {
    // Setup
    store := memory.NewInMemoryTenantStore()
    service := application.NewTenantService(store)
    
    // Test
    tenant := &domain.Tenant{
        ID:   "test",
        Name: "Test Corp",
    }
    
    err := service.CreateTenant(context.Background(), tenant)
    assert.NoError(t, err)
    
    // Verify
    retrieved, err := service.GetTenant(context.Background(), "test")
    assert.NoError(t, err)
    assert.Equal(t, "Test Corp", retrieved.Name)
}
```

## Comparison with PostgreSQL

| Feature | In-Memory | PostgreSQL |
|---------|-----------|------------|
| Persistence | ❌ Volatile | ✅ Persistent |
| Setup Complexity | ✅ None | ⚠️ Requires DB |
| Performance | ✅ Very fast | ✅ Fast |
| Memory Usage | ⚠️ All in RAM | ✅ Disk-based |
| Scalability | ❌ Limited | ✅ High |
| Concurrent Access | ⚠️ Single instance | ✅ Multi-instance |
| Transactions | ❌ None | ✅ Full ACID |
| Complex Queries | ❌ Limited | ✅ Full SQL |
| Production Ready | ❌ No | ✅ Yes |

## See Also

- `../postgres/` - PostgreSQL implementation
- `../contract.go` - Store interface definitions
- `../../application/` - Service layer using stores
- `POSTGRES_MIGRATION_GUIDE.md` - Migration guide
