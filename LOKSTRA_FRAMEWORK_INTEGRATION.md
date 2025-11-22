# Lokstra Framework Integration Guide

## Overview

Dokumen ini menjelaskan cara mengintegrasikan **lokstra-auth** dengan **Lokstra Framework** menggunakan annotation-driven development pattern (`@RouterService`, `@Inject`, `@Route`).

## Prerequisites

- Go 1.23 atau lebih tinggi
- Lokstra Framework v0.x (https://github.com/primadi/lokstra)
- VS Code dengan extension Go terpasang (recommended)

## Architecture

Integrasi dengan Lokstra Framework menggunakan pattern sebagai berikut:

1. **Service Layer** - Business logic dengan annotations
2. **Auto-Generated Code** - Code generation dari annotations
3. **HTTP Endpoints** - Auto-routing berdasarkan `@Route` annotations

## Annotation System

### 1. @RouterService

Mendefinisikan service sebagai HTTP router service.

**Syntax:**
```go
// @RouterService name="service-name", prefix="/api/path", middlewares=["middleware1", "middleware2"]
type ServiceName struct {
    // fields...
}
```

**Parameters:**
- `name` (required) - Nama service untuk registration
- `prefix` (optional) - URL prefix untuk semua routes
- `middlewares` (optional) - Array middleware names

**Example:**
```go
// @RouterService name="tenant-service", prefix="/api/tenants", middlewares=["recovery", "request-logger"]
type TenantService struct {
    // @Inject "tenant-store"
    Store *service.Cached[TenantStore]
}
```

### 2. @Inject

Auto-wiring dependency injection.

**Syntax:**
```go
type MyService struct {
    // @Inject "dependency-service-name"
    FieldName *service.Cached[DependencyType]
}
```

**Example:**
```go
type TenantService struct {
    // @Inject "tenant-store"
    Store *service.Cached[TenantStore]
}
```

### 3. @Route

Mapping method ke HTTP endpoint.

**Syntax:**
```go
// @Route "METHOD /path/{param}"
func (s *Service) MethodName(ctx context.Context, req *RequestType) (*ResponseType, error) {
    // implementation
}
```

**Supported HTTP Methods:**
- `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`, `HEAD`

**Path Parameters:**
- Use `{paramName}` untuk path variables
- Maps to fields in request struct dengan tag `path:"paramName"`

**Example:**
```go
// @Route "GET /{id}"
func (s *TenantService) GetTenant(ctx context.Context, req *GetTenantRequest) (*core.Tenant, error) {
    if req.ID == "" {
        return nil, fmt.Errorf("tenant ID is required")
    }
    
    tenant, err := s.Store.MustGet().Get(ctx, req.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to get tenant: %w", err)
    }
    
    return tenant, nil
}
```

## Request/Response DTOs

Request structs harus menggunakan struct tags untuk parameter binding:

```go
// Path parameters
type GetTenantRequest struct {
    ID string `path:"id" validate:"required"`
}

// Query parameters
type ListTenantsRequest struct {
    Page int `query:"page"`
    Size int `query:"size"`
}

// JSON body with optional fields
type CreateTenantRequest struct {
    Name     string                 `json:"name" validate:"required"`
    Settings *core.TenantSettings   `json:"settings,omitempty"`    // Optional
    Metadata map[string]any `json:"metadata,omitempty"`     // Optional
}
```

**Field Tags:**
- `validate:"required"` - Required field (validation)
- `json:"...,omitempty"` - Optional field (omitted if empty/nil)
- `path:"id"` - From URL path parameter
- `query:"page"` - From query string parameter


## Service Implementation Example

Berikut contoh lengkap implementasi `TenantService`:

```go
package services

import (
    "context"
    "fmt"
    "time"
    
    core "github.com/primadi/lokstra-auth/00_core"
    "github.com/primadi/lokstra/core/service"
)

// @RouterService name="tenant-service", prefix="/api/tenants", middlewares=["recovery", "request-logger"]
type TenantService struct {
    // @Inject "tenant-store"
    Store *service.Cached[TenantStore]
}

// @Route "POST /"
func (s *TenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*core.Tenant, error) {
    // Check if tenant name already exists
    existing, err := s.Store.MustGet().GetByName(ctx, req.Name)
    if err == nil && existing != nil {
        return nil, fmt.Errorf("tenant with name '%s' already exists", req.Name)
    }
    
    // Initialize settings - use provided or defaults
    settings := core.TenantSettings{}
    if req.Settings != nil {
        settings = *req.Settings
    }
    
    // Initialize metadata - use provided or empty map
    metadata := make(map[string]any)
    if req.Metadata != nil {
        metadata = req.Metadata
    }
    
    // Create tenant
    tenant := &core.Tenant{
        ID:        utils.GenerateID("tenant"),
        Name:      req.Name,
        Status:    core.TenantStatusActive,
        Settings:  settings,
        Metadata:  metadata,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    // Store operation using injected dependency
    if err := s.Store.MustGet().Create(ctx, tenant); err != nil {
        return nil, fmt.Errorf("failed to create tenant: %w", err)
    }
    
    return tenant, nil
}

// @Route "GET /{id}"
func (s *TenantService) GetTenant(ctx context.Context, req *GetTenantRequest) (*core.Tenant, error) {
    tenant, err := s.Store.MustGet().Get(ctx, req.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to get tenant: %w", err)
    }
    return tenant, nil
}

// @Route "PUT /{id}"
func (s *TenantService) UpdateTenant(ctx context.Context, req *UpdateTenantRequest) (*core.Tenant, error) {
    // Implementation...
}

// @Route "DELETE /{id}"
func (s *TenantService) DeleteTenant(ctx context.Context, req *DeleteTenantRequest) error {
    // Implementation...
}

// @Route "GET /"
func (s *TenantService) ListTenants(ctx context.Context, req *ListTenantsRequest) ([]*core.Tenant, error) {
    // Implementation...
}

// @Route "POST /{id}/activate"
func (s *TenantService) ActivateTenant(ctx context.Context, req *ActivateTenantRequest) (*core.Tenant, error) {
    // Implementation...
}

// @Route "POST /{id}/suspend"
func (s *TenantService) SuspendTenant(ctx context.Context, req *SuspendTenantRequest) (*core.Tenant, error) {
    // Implementation...
}

// Register triggers package initialization for code generation
func Register() {
    // This function is called during initialization
}
```

## Code Generation

Ketika annotation processor dijalankan, akan menghasilkan file `zz_generated.lokstra.go`:

```go
// AUTO-GENERATED CODE - DO NOT EDIT
package services

import (
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/core/proxy"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/core/service"
)

// Auto-register on package import
func init() {
    RegisterTenantService()
}

// TenantServiceFactory creates a TenantService instance
func TenantServiceFactory(deps map[string]any, config map[string]any) any {
    return &TenantService{
        Store: service.Cast[TenantStore](deps["tenant-store"]),
    }
}

// TenantServiceRemote implements remote HTTP client
type TenantServiceRemote struct {
    proxyService *proxy.Service
}

// RegisterTenantService registers the service with the registry
func RegisterTenantService() {
    lokstra_registry.RegisterServiceType("tenant-service-factory",
        TenantServiceFactory,
        TenantServiceRemoteFactory,
        deploy.WithRouter(&deploy.ServiceTypeRouter{
            PathPrefix:  "/api/tenants",
            Middlewares: []string{ "recovery", "request-logger" },
            CustomRoutes: map[string]string{
                "CreateTenant":    "POST /",
                "GetTenant":       "GET /{id}",
                "UpdateTenant":    "PUT /{id}",
                "DeleteTenant":    "DELETE /{id}",
                "ListTenants":     "GET /",
                "ActivateTenant":  "POST /{id}/activate",
                "SuspendTenant":   "POST /{id}/suspend",
            },
        }),
    )
    
    lokstra_registry.RegisterLazyService("tenant-service",
        "tenant-service-factory",
        map[string]any{
            "depends-on": []string{ "tenant-store" },
        })
}
```

## Generated Endpoints

Dari annotations di atas, akan menghasilkan endpoints berikut:

| Method | Path | Handler |
|--------|------|---------|
| `POST` | `/api/tenants/` | `CreateTenant` |
| `GET` | `/api/tenants/{id}` | `GetTenant` |
| `PUT` | `/api/tenants/{id}` | `UpdateTenant` |
| `DELETE` | `/api/tenants/{id}` | `DeleteTenant` |
| `GET` | `/api/tenants/` | `ListTenants` |
| `POST` | `/api/tenants/{id}/activate` | `ActivateTenant` |
| `POST` | `/api/tenants/{id}/suspend` | `SuspendTenant` |

## Service Access Patterns

### 1. Using LazyLoad (Recommended)

```go
// Package-level lazy loading
var tenantService = service.LazyLoad[*TenantService]("tenant-service")

func someHandler(ctx *request.Context) error {
    // MustGet() for clear error messages
    tenant, err := tenantService.MustGet().GetTenant(ctx, &GetTenantRequest{ID: "123"})
    if err != nil {
        return err
    }
    return ctx.Api.Ok(tenant)
}
```

### 2. Using Registry Lookup (Not Recommended for Production)

```go
func someHandler(ctx *request.Context) error {
    tenantService := lokstra_registry.GetService[*TenantService]("tenant-service")
    tenant, err := tenantService.GetTenant(ctx, &GetTenantRequest{ID: "123"})
    // ...
}
```

## Best Practices

### 1. Use MustGet() for Clear Errors

```go
// ✅ Good - Clear error message
tenant, err := s.Store.MustGet().Get(ctx, id)
// Panic: "service 'tenant-store' not found or not initialized"

// ❌ Bad - Confusing error
tenant, err := s.Store.Get().Get(ctx, id)
// Panic: "runtime error: invalid memory address"
```

### 2. Package-Level LazyLoad

```go
// ✅ Good - Cached after first access
var tenantService = service.LazyLoad[*TenantService]("tenant-service")

// ❌ Bad - Created every request, cache wasted
func handler() {
    tenantService := service.LazyLoad[*TenantService]("tenant-service")
}
```

### 3. Descriptive Service Names

```go
// ✅ Good - Follows naming convention
// @RouterService name="tenant-service"

// ❌ Bad - Inconsistent naming
// @RouterService name="tnntSvc"
```

### 4. Group Related Routes

```go
// ✅ Good - Consistent prefix
// @RouterService name="tenant-service", prefix="/api/v1/tenants"

// @Route "GET /{id}"
func GetTenant(...) { ... }

// @Route "POST /"
func Create(...) { ... }
```

## Services to Annotate

Berikut daftar services di lokstra-auth yang perlu ditambahkan annotations:

1. ✅ **TenantService** (`00_core/services/tenant_service.go`) - DONE
2. ⏳ **UserService** (`00_core/services/user_service.go`)
3. ⏳ **AppService** (`00_core/services/app_service.go`)
4. ⏳ **BranchService** (`00_core/services/branch_service.go`)
5. ⏳ **AppKeyService** (`00_core/services/app_key_service.go`)

## Running the Application

### 1. Install Dependencies

```bash
cd /path/to/lokstra-auth
go mod tidy
```

### 2. Run Code Generation

Code generation akan otomatis dijalankan ketika aplikasi pertama kali di-run:

```bash
go run .
```

### 3. Check Generated Files

Generated files akan ada di folder yang sama dengan service:

```
00_core/services/
├── tenant_service.go
├── zz_generated.lokstra.go  # Auto-generated
└── zz_cache.lokstra.json   # Cache metadata (add to .gitignore)
```

### 4. Test Endpoints

Gunakan HTTP client atau `curl`:

```bash
# Create tenant (minimal - only required fields)
curl -X POST http://localhost:3000/api/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Acme Corp"}'

# Create tenant (with optional settings and metadata)
curl -X POST http://localhost:3000/api/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "settings": {
      "max_apps": 10,
      "max_users": 100
    },
    "metadata": {
      "industry": "Technology",
      "region": "US-West"
    }
  }'

# Get tenant
curl http://localhost:3000/api/tenants/{tenant-id}

# List tenants
curl http://localhost:3000/api/tenants

# Activate tenant
curl -X POST http://localhost:3000/api/tenants/{tenant-id}/activate
```

## Configuration

Example configuration untuk deployment:

```yaml
# config/auth.yaml
deployments:
  - name: auth-api
    type: server
    port: 3000
    
    services:
      # Stores
      - name: tenant-store
        factory: tenant-store-factory
      
      - name: user-store
        factory: user-store-factory
      
      - name: app-store
        factory: app-store-factory
      
      # Services (auto-registered via annotations)
      - name: tenant-service
        factory: tenant-service-factory
        dependencies:
          tenant-store: tenant-store
      
      - name: user-service
        factory: user-service-factory
        dependencies:
          user-store: user-store
          tenant-service: tenant-service
      
      - name: app-service
        factory: app-service-factory
        dependencies:
          app-store: app-store
          tenant-service: tenant-service
```

## Troubleshooting

### Code Generation Not Running

Check if `Register()` function exists in service file:

```go
// Add this at the end of service file
func Register() {
    // Triggers package initialization
}
```

### Service Not Found Error

```
panic: service 'tenant-service' not found or not initialized
```

**Solution:** Ensure service is registered in `main.go` or config file.

### Annotation Not Detected

Ensure annotations are in correct format:
- Annotations must start with `// @` (note the space after `//`)
- Must be on line immediately before target (struct or method)
- No blank lines between annotation and target

```go
// ✅ Correct
// @RouterService name="tenant-service"
type TenantService struct {}

// ❌ Wrong - blank line between annotation and target
// @RouterService name="tenant-service"

type TenantService struct {}
```

## Migration from Manual Registration

### Before (Manual - 70+ lines)

```go
// 1. Define service
type TenantService struct {
    store TenantStore
}

// 2. Create factory
func TenantServiceFactory(deps map[string]any, config map[string]any) any {
    return &TenantService{
        store: service.Cast[TenantStore](deps["tenant-store"]),
    }
}

// 3. Register factory
func init() {
    lokstra_registry.RegisterServiceType("tenant-service-factory",
        TenantServiceFactory, nil)
    
    lokstra_registry.RegisterLazyService("tenant-service",
        "tenant-service-factory",
        map[string]any{"depends-on": []string{"tenant-store"}})
}

// 4. Create routes manually
func setupRoutes() {
    r := lokstra.NewRouter("tenants")
    r.POST("/", createTenantHandler)
    r.GET("/:id", getTenantHandler)
    // ... more routes
}
```

### After (Annotations - 12 lines)

```go
// @RouterService name="tenant-service", prefix="/api/tenants"
type TenantService struct {
    // @Inject "tenant-store"
    Store *service.Cached[TenantStore]
}

// @Route "POST /"
func (s *TenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*core.Tenant, error) {
    return s.Store.MustGet().Create(ctx, tenant)
}

// @Route "GET /{id}"
func (s *TenantService) GetTenant(ctx context.Context, req *GetTenantRequest) (*core.Tenant, error) {
    return s.Store.MustGet().Get(ctx, req.ID)
}

func Register() {}
```

**Benefits:**
- ✅ **83% less code** (70+ lines → 12 lines)
- ✅ Clear and declarative
- ✅ Auto-generated factory, DI wiring, routes
- ✅ Type-safe with generics

## Learn More

- [Lokstra Framework Documentation](https://primadi.github.io/lokstra/)
- [Lokstra GitHub Repository](https://github.com/primadi/lokstra)
- [Enterprise Router Service Example](https://github.com/primadi/lokstra/tree/main/project_templates/02_app_framework/03_enterprise_router_service)
- [Annotation System Guide](https://github.com/primadi/lokstra/docs/00-introduction/key-features.md)

## License

This integration guide is part of the lokstra-auth project.
