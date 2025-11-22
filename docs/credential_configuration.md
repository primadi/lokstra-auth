# Credential Configuration Management

Sistem konfigurasi credential di lokstra-auth menggunakan **hierarchical configuration** dengan 3 level:

```
Global Default → Tenant Default → App Override
```

## 📋 Arsitektur

### 1. **Configuration Hierarchy**

```go
// Level 1: Global Default (di-hardcode)
domain.DefaultCredentialConfig()

// Level 2: Tenant Default (di database)
Tenant.Config.DefaultCredentials

// Level 3: App Override (di database)
App.Config.Credentials
```

### 2. **Configuration Resolution**

`ConfigResolver` di `01_credential/application/config_resolver.go` bertanggung jawab untuk menentukan konfigurasi mana yang digunakan:

```go
func (r *ConfigResolver) GetEffectiveConfig(ctx, tenantID, appID) *CredentialConfig {
    // 1. Cek App Config (highest priority)
    if app.Config != nil && app.Config.Credentials != nil {
        return app.Config.Credentials
    }
    
    // 2. Cek Tenant Config
    if tenant.Config != nil && tenant.Config.DefaultCredentials != nil {
        return tenant.Config.DefaultCredentials
    }
    
    // 3. Return Global Default
    return domain.DefaultCredentialConfig()
}
```

## 🎯 Use Cases

### Use Case 1: SaaS Multi-Tenant

**Scenario**: Platform SaaS dengan berbagai tenant yang memiliki kebutuhan security berbeda.

- **Tenant A** (Enterprise): Hanya basic auth dengan password policy ketat
- **Tenant B** (Startup): OAuth2 + API Key untuk fleksibilitas
- **Tenant C** (Corporate): Basic + Passkey untuk security maksimal

**Implementation**:

```json
// Tenant A Default Config
{
  "enable_basic": true,
  "basic_config": {
    "min_password_length": 16,
    "max_login_attempts": 3,
    "lockout_duration_secs": 1800
  },
  "enable_apikey": false,
  "enable_oauth2": false
}

// Tenant B Default Config
{
  "enable_basic": false,
  "enable_apikey": true,
  "enable_oauth2": true,
  "oauth2_config": {
    "providers": [
      {
        "name": "google",
        "enabled": true,
        "client_id": "...",
        "scopes": ["openid", "email", "profile"]
      }
    ]
  }
}
```

### Use Case 2: Per-App Override

**Scenario**: Dalam satu tenant, ada berbagai aplikasi dengan kebutuhan berbeda.

- **Web App**: Basic auth untuk user login
- **Mobile App**: Basic auth + API Key untuk background sync
- **Partner API**: Hanya API Key

**Implementation**:

```json
// Tenant Default: Basic only
{
  "enable_basic": true,
  "enable_apikey": false
}

// Mobile App Override: Basic + API Key
{
  "enable_basic": true,
  "enable_apikey": true,
  "apikey_config": {
    "default_expiry_days": 365,
    "allow_never_expire": true
  }
}

// Partner API Override: API Key only
{
  "enable_basic": false,
  "enable_apikey": true,
  "apikey_config": {
    "default_expiry_days": 180,
    "allow_never_expire": false,
    "rate_limit_per_minute": 1000
  }
}
```

## 🔧 REST API

### Tenant-Level Configuration

#### Get Tenant Default Config
```http
GET /api/registration/config/credentials/tenants/{tenant_id}
```

Response:
```json
{
  "tenant_id": "tenant_abc",
  "default_credentials": {
    "enable_basic": true,
    "basic_config": {...},
    "enable_apikey": true,
    "apikey_config": {...}
  }
}
```

#### Update Tenant Default Config
```http
PUT /api/registration/config/credentials/tenants/{tenant_id}
Content-Type: application/json

{
  "default_credentials": {
    "enable_basic": true,
    "basic_config": {
      "min_password_length": 12,
      "max_login_attempts": 5
    }
  }
}
```

### App-Level Configuration

#### Get App Config
```http
GET /api/registration/config/credentials/tenants/{tenant_id}/apps/{app_id}
```

Response:
```json
{
  "tenant_id": "tenant_abc",
  "app_id": "app_mobile",
  "credentials": {
    "enable_basic": true,
    "enable_apikey": true,
    "apikey_config": {...}
  },
  "source": "app"  // atau "tenant" jika fallback ke tenant default
}
```

#### Update App Config
```http
PUT /api/registration/config/credentials/tenants/{tenant_id}/apps/{app_id}
Content-Type: application/json

{
  "credentials": {
    "enable_basic": true,
    "enable_apikey": true,
    "apikey_config": {
      "default_expiry_days": 180
    }
  }
}
```

## 💻 Penggunaan di Service Layer

### Basic Auth Service

```go
func (s *BasicAuthService) Login(ctx *request.Context, req *basic.LoginRequest) (*basic.LoginResponse, error) {
    // 1. Check if basic auth is enabled
    if !s.ConfigResolver.MustGet().IsBasicEnabled(ctx, req.TenantID, req.AppID) {
        return &basic.LoginResponse{
            Success: false,
            Error:   "basic authentication is not enabled for this application",
        }, nil
    }

    // 2. Get effective config for validation
    config := s.ConfigResolver.MustGet().GetBasicConfig(ctx, req.TenantID, req.AppID)
    
    // 3. Apply config rules
    if len(req.Password) < config.MinPasswordLength {
        return &basic.LoginResponse{
            Success: false,
            Error:   fmt.Sprintf("password must be at least %d characters", config.MinPasswordLength),
        }, nil
    }

    // ... continue with authentication
}
```

### API Key Service

```go
func (s *APIKeyAuthService) Authenticate(ctx *request.Context, req *apikey.AuthenticateRequest) (*apikey.AuthenticateResponse, error) {
    // 1. Authenticate first to get tenant/app from key
    result, err := s.Authenticator.MustGet().Authenticate(ctx, authCtx, creds)
    if err != nil {
        return nil, err
    }

    // 2. Check if API key auth is enabled for this tenant/app
    if !s.ConfigResolver.MustGet().IsAPIKeyEnabled(ctx, result.TenantID, result.AppID) {
        return &apikey.AuthenticateResponse{
            Success: false,
            Error:   "API key authentication is not enabled for this application",
        }, nil
    }

    // ... continue
}
```

## 📊 Configuration Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Request                           │
│                (Login/Register/Authenticate)                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  Router Service                             │
│          (BasicAuthService, APIKeyAuthService)              │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              ConfigResolver.IsXXXEnabled()                  │
│         GetEffectiveConfig(tenantID, appID)                 │
└────────────────────────┬────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
   ┌─────────┐    ┌──────────┐    ┌──────────┐
   │   App   │    │  Tenant  │    │  Global  │
   │ Config  │    │ Default  │    │ Default  │
   └─────────┘    └──────────┘    └──────────┘
   Priority 1     Priority 2      Priority 3
```

## 🔐 Security Best Practices

1. **Tenant Isolation**: Setiap tenant memiliki konfigurasi independen
2. **Principle of Least Privilege**: Default configuration restrictive, override untuk relaksasi
3. **Audit Trail**: Semua perubahan konfigurasi dicatat (TODO: implement audit log)
4. **Validation**: Validasi konfigurasi sebelum disimpan untuk mencegah misconfiguration
5. **Fallback**: Selalu ada fallback ke default yang aman

## 📝 Contoh Lengkap

Lihat `examples/00_core/02_credential_config/main.go` untuk contoh lengkap penggunaan sistem konfigurasi credential.

## 🎓 Key Takeaways

1. **Konfigurasi di 00_core**, bukan di 01_credential
   - 00_core owns the configuration model
   - 01_credential reads configuration via ConfigResolver

2. **3-Level Hierarchy**: Global → Tenant → App
   - Global: Default aman untuk semua tenant
   - Tenant: Customize per tenant
   - App: Fine-tune per aplikasi dalam tenant

3. **ConfigResolver** adalah single source of truth
   - Semua service menggunakan ConfigResolver
   - Konsisten di seluruh aplikasi
   - Easy to test dan mock

4. **REST API** untuk management
   - Admin dapat mengatur konfigurasi via API
   - Tidak perlu restart aplikasi
   - Real-time configuration update
