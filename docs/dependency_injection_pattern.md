# Dependency Injection Pattern

## Philosophy

Lokstra Auth menggunakan dua strategi injection yang berbeda berdasarkan **layer architecture**:

### 1. Repository Layer → EAGER Loading
**Pattern:**
```go
type MyService struct {
    // @Inject "my-store"
    Store repository.MyStore
}
```

**Usage:**
```go
s.Store.Get(ctx, id)  // Direct access, no wrapper
```

**Alasan:**
- ✅ **No Circular Dependencies** - Repository tidak pernah depend ke Application Service
- ✅ **Better Performance** - Tidak ada overhead `MustGet()` di setiap akses
- ✅ **Always Available** - Repository selalu bisa di-load saat startup
- ✅ **Cleaner Code** - Lebih sedikit boilerplate

### 2. Application Service → LAZY Loading
**Pattern:**
```go
type MyService struct {
    // @Inject "other-service"
    OtherService *service.Cached[*OtherService]
}
```

**Usage:**
```go
s.OtherService.MustGet().DoSomething(ctx, req)  // Lazy access with wrapper
```

**Alasan:**
- ✅ **Solve Circular Dependencies** - Service A bisa depend ke Service B dan sebaliknya
- ✅ **Lazy Initialization** - Service di-load hanya saat pertama kali diakses
- ✅ **Prevent Load Order Issues** - Tidak perlu memikirkan urutan registrasi service

## Examples

### ❌ WRONG - Service as Eager
```go
type UserService struct {
    Store repository.UserStore        // ✅ Correct
    TenantService *TenantService       // ❌ Wrong! Could cause circular dependency
}
```

### ✅ CORRECT - Mixed Strategy
```go
type UserService struct {
    // Repository - Eager (fast, no circular deps)
    Store repository.UserStore
    
    // Service - Lazy (handles circular deps)
    TenantService *service.Cached[*TenantService]
}

func (s *UserService) CreateUser(ctx *request.Context, req *domain.CreateUserRequest) (*domain.User, error) {
    // Repository access - direct
    existing, err := s.Store.GetByEmail(ctx, req.Email)
    
    // Service access - lazy with MustGet()
    tenant, err := s.TenantService.MustGet().GetTenant(ctx, &domain.GetTenantRequest{
        ID: req.TenantID,
    })
    
    // ...
}
```

## Benefits

| Aspect | Repository (Eager) | Service (Lazy) |
|--------|-------------------|----------------|
| Performance | ⚡ Faster (direct access) | 🐢 Small overhead (MustGet) |
| Circular Deps | ✅ Not possible | ✅ Handled |
| Load Order | ✅ Any order | ✅ Any order |
| Code Clarity | ✅ Clean syntax | ⚠️ Extra MustGet() |
| Use Case | Data access | Business logic coordination |

## When to Use Each

### Use EAGER (Direct Injection) for:
- ✅ Repository/Store implementations
- ✅ Infrastructure services (cache, logger, etc.)
- ✅ External clients (HTTP, gRPC, etc.)
- ✅ Any dependency that doesn't create circular references

### Use LAZY (service.Cached) for:
- ✅ Application Services calling other Application Services
- ✅ Any potential circular dependency scenario
- ✅ Services that might not always be needed

## Summary

**Rule of Thumb:**
- **Repository → Application Service** = EAGER (direct)
- **Application Service → Application Service** = LAZY (service.Cached)

This pattern gives us the best of both worlds:
- Performance where it matters (repository access)
- Flexibility where we need it (service composition)
