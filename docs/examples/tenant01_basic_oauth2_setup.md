# Contoh Konfigurasi: Tenant01 dengan Basic + OAuth2 (Google & Azure)

## Skenario

**Tenant**: `tenant01`
**App**: `app01` (Web Application)

**Kebutuhan**:
- Enable Basic Authentication (username/password)
- Enable OAuth2 dengan 2 provider:
  - Google
  - Microsoft Azure AD

## 1️⃣ Cara Konfigurasi via REST API

### Step 1: Set Tenant Default Configuration

Buat konfigurasi default di level tenant yang enable Basic + OAuth2:

```http
PUT /api/registration/config/credentials/tenants/tenant01
Content-Type: application/json

{
  "config": {
    "enable_basic": true,
    "basic_config": {
      "min_username_length": 3,
      "max_username_length": 50,
      "min_password_length": 10,
      "require_strong_pwd": true,
      "max_login_attempts": 5,
      "lockout_duration_secs": 900,
      "session_timeout_secs": 3600
    },
    "enable_apikey": false,
    "enable_oauth2": true,
    "oauth2_config": {
      "providers": [
        {
          "name": "google",
          "enabled": true,
          "client_id": "123456789-abc.apps.googleusercontent.com",
          "client_secret": "GOCSPX-your-secret-here",
          "scopes": ["openid", "email", "profile"],
          "auth_url": "https://accounts.google.com/o/oauth2/v2/auth",
          "token_url": "https://oauth2.googleapis.com/token",
          "user_url": "https://www.googleapis.com/oauth2/v3/userinfo"
        },
        {
          "name": "azure",
          "enabled": true,
          "client_id": "azure-app-client-id",
          "client_secret": "azure-app-client-secret",
          "scopes": ["openid", "email", "profile"],
          "auth_url": "https://login.microsoftonline.com/{tenant}/oauth2/v2.0/authorize",
          "token_url": "https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token",
          "user_url": "https://graph.microsoft.com/v1.0/me"
        }
      ],
      "callback_url": "https://app.tenant01.com/api/cred/oauth2/callback",
      "state_expiry_secs": 600,
      "session_timeout_secs": 3600
    },
    "enable_passwordless": false,
    "enable_passkey": false
  }
}
```

### Step 2 (Optional): Override untuk App Tertentu

Jika `app01` butuh konfigurasi berbeda (misal rate limit lebih tinggi):

```http
PUT /api/registration/config/credentials/tenants/tenant01/apps/app01
Content-Type: application/json

{
  "config": {
    "enable_basic": true,
    "basic_config": {
      "min_password_length": 12,
      "max_login_attempts": 3,
      "lockout_duration_secs": 1800
    },
    "enable_oauth2": true,
    "oauth2_config": {
      "providers": [
        {
          "name": "google",
          "enabled": true,
          "client_id": "app01-specific-google-client-id",
          "client_secret": "app01-specific-google-secret",
          "scopes": ["openid", "email", "profile"]
        },
        {
          "name": "azure",
          "enabled": true,
          "client_id": "app01-specific-azure-client-id",
          "client_secret": "app01-specific-azure-secret",
          "scopes": ["openid", "email", "profile"]
        }
      ],
      "callback_url": "https://app01.tenant01.com/auth/callback"
    }
  }
}
```

### Step 3: Verify Configuration

Check apakah konfigurasi sudah benar:

```http
# Get tenant default
GET /api/registration/config/credentials/tenants/tenant01

# Get app-specific (with fallback to tenant)
GET /api/registration/config/credentials/tenants/tenant01/apps/app01
```

Response:
```json
{
  "enable_basic": true,
  "basic_config": {
    "min_password_length": 12,
    "max_login_attempts": 3,
    "lockout_duration_secs": 1800,
    "session_timeout_secs": 3600
  },
  "enable_oauth2": true,
  "oauth2_config": {
    "providers": [
      {
        "name": "google",
        "enabled": true,
        "client_id": "app01-specific-google-client-id",
        "scopes": ["openid", "email", "profile"]
      },
      {
        "name": "azure",
        "enabled": true,
        "client_id": "app01-specific-azure-client-id",
        "scopes": ["openid", "email", "profile"]
      }
    ],
    "callback_url": "https://app01.tenant01.com/auth/callback",
    "state_expiry_secs": 600,
    "session_timeout_secs": 3600
  }
}
```

## 2️⃣ Cara Konfigurasi via Code (Direct Database)

Jika tidak pakai REST API, bisa set langsung di code:

```go
package main

import (
    "context"
    
    coredomain "github.com/primadi/lokstra-auth/00_core/domain"
)

func setupTenant01() {
    ctx := context.Background()
    
    // Create tenant with credential config
    tenant := &coredomain.Tenant{
        ID:       "tenant01",
        Name:     "Tenant 01",
        DBDsn:    "postgres://user:pass@localhost/tenant01_db",
        DBSchema: "tenant01",
        Config: &coredomain.TenantConfig{
            DefaultCredentials: &coredomain.CredentialConfig{
                EnableBasic: true,
                BasicConfig: &coredomain.BasicCredentialConfig{
                    MinUsernameLength:   3,
                    MaxUsernameLength:   50,
                    MinPasswordLength:   10,
                    RequireStrongPwd:    true,
                    MaxLoginAttempts:    5,
                    LockoutDurationSecs: 900,
                    SessionTimeoutSecs:  3600,
                },
                EnableOAuth2: true,
                OAuth2Config: &coredomain.OAuth2CredentialConfig{
                    Providers: []coredomain.OAuth2ProviderConfig{
                        {
                            Name:         "google",
                            Enabled:      true,
                            ClientID:     "123456789-abc.apps.googleusercontent.com",
                            ClientSecret: "GOCSPX-your-secret-here",
                            Scopes:       []string{"openid", "email", "profile"},
                            AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
                            TokenURL:     "https://oauth2.googleapis.com/token",
                            UserURL:      "https://www.googleapis.com/oauth2/v3/userinfo",
                        },
                        {
                            Name:         "azure",
                            Enabled:      true,
                            ClientID:     "azure-app-client-id",
                            ClientSecret: "azure-app-client-secret",
                            Scopes:       []string{"openid", "email", "profile"},
                            AuthURL:      "https://login.microsoftonline.com/{tenant}/oauth2/v2.0/authorize",
                            TokenURL:     "https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token",
                            UserURL:      "https://graph.microsoft.com/v1.0/me",
                        },
                    },
                    CallbackURL:        "https://app.tenant01.com/api/cred/oauth2/callback",
                    StateExpirySecs:    600,
                    SessionTimeoutSecs: 3600,
                },
            },
        },
    }
    
    // Save to database via TenantService
    tenantService.CreateTenant(ctx, tenant)
}
```

## 3️⃣ Flow Otentikasi User

### Basic Auth Flow
```
1. User → POST /api/cred/basic/login
   Body: { "tenant_id": "tenant01", "app_id": "app01", "username": "john", "password": "secret123" }

2. BasicAuthService checks:
   - ConfigResolver.IsBasicEnabled("tenant01", "app01") → true ✅
   - ConfigResolver.GetBasicConfig("tenant01", "app01") → min_password_length=12

3. Validate password length, authenticate, return token
```

### OAuth2 Google Flow
```
1. User clicks "Sign in with Google"

2. Frontend → GET /api/cred/oauth2/authorize?provider=google&tenant_id=tenant01&app_id=app01

3. OAuth2Service checks:
   - ConfigResolver.IsOAuth2Enabled("tenant01", "app01") → true ✅
   - ConfigResolver.GetOAuth2Config("tenant01", "app01")
   - Find provider "google" in config

4. Redirect to Google:
   https://accounts.google.com/o/oauth2/v2/auth?
     client_id=app01-specific-google-client-id&
     redirect_uri=https://app01.tenant01.com/auth/callback&
     scope=openid+email+profile&
     state=random-state

5. Google redirects back:
   https://app01.tenant01.com/auth/callback?code=AUTH_CODE&state=random-state

6. OAuth2Service exchanges code for token, gets user info, creates session
```

### OAuth2 Azure Flow
```
Same as Google, but using Azure endpoints:
1. Redirect to: https://login.microsoftonline.com/{tenant}/oauth2/v2.0/authorize
2. Token exchange: https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token
3. User info: https://graph.microsoft.com/v1.0/me
```

## 4️⃣ Testing Configuration

### Test Basic Auth
```bash
curl -X POST http://localhost:8080/api/cred/basic/login \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant01",
    "app_id": "app01",
    "username": "john",
    "password": "MySecurePass123!"
  }'
```

### Test OAuth2 Google
```bash
# Get authorization URL
curl "http://localhost:8080/api/cred/oauth2/authorize?provider=google&tenant_id=tenant01&app_id=app01"

# User visits URL, signs in, gets redirected back with code
# Then exchange code for token (handled automatically by callback endpoint)
```

### Test Configuration Check
```bash
# Check tenant default config
curl http://localhost:8080/api/registration/config/credentials/tenants/tenant01

# Check app-specific config
curl http://localhost:8080/api/registration/config/credentials/tenants/tenant01/apps/app01
```

## 5️⃣ Tips & Best Practices

### Security
1. **Never commit secrets**: OAuth2 client_secret harus di environment variable
2. **Use HTTPS**: Semua OAuth2 callback URL harus HTTPS
3. **Validate redirect_uri**: Prevent open redirect attacks
4. **Short state timeout**: Default 10 menit untuk OAuth2 state

### Configuration Management
1. **Tenant default first**: Set tenant default dulu, app override kemudian
2. **Minimize overrides**: Hanya override jika benar-benar perlu
3. **Version control**: Backup konfigurasi sebelum update
4. **Test in dev first**: Jangan langsung update production config

### Multi-Provider OAuth2
1. **Provider naming**: Use lowercase: "google", "azure", "github"
2. **Scopes**: Request minimal scopes yang dibutuhkan saja
3. **Provider metadata**: Simpan provider URLs di config untuk flexibility
4. **Fallback**: Selalu ada fallback (basic auth) jika OAuth2 down

## 6️⃣ Troubleshooting

### "basic authentication is not enabled"
```
Cause: ConfigResolver.IsBasicEnabled() returns false
Fix: Set enable_basic: true di tenant atau app config
```

### "OAuth2 provider not found"
```
Cause: Provider name tidak match dengan config
Fix: Check provider name exact match (case-sensitive)
```

### "Invalid OAuth2 callback"
```
Cause: callback_url di config tidak match dengan registered URL
Fix: Update callback_url di config DAN di provider console (Google/Azure)
```

## 📚 Referensi

- [Credential Configuration Docs](./credential_configuration.md)
- [OAuth2 Implementation Guide](./01_credential.md)
- [Multi-Tenant Architecture](./multi_tenant_architecture.md)
