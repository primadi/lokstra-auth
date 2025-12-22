# App API Keys - Service-to-Service Authentication

App API Keys adalah fitur untuk autentikasi **service-to-service**, **webhook callbacks**, dan **background jobs** yang tidak diakses oleh user tetapi oleh aplikasi/server lain.

## 📋 Table of Contents

- [Overview](#overview)
- [Use Cases](#use-cases)
- [Key Concepts](#key-concepts)
- [Architecture](#architecture)
- [Usage Guide](#usage-guide)
- [Security Best Practices](#security-best-practices)
- [API Reference](#api-reference)

## Overview

### User API Keys vs App API Keys

| Aspect | User API Keys | App API Keys |
|--------|--------------|--------------|
| **Owner** | Individual user | Application/Service |
| **Subject** | User ID | `app:{appID}` |
| **Use Case** | User accessing API on behalf of themselves | Service-to-service, webhooks, background jobs |
| **Lifecycle** | Managed by user | Managed by app administrators |
| **Permissions** | User-level permissions | App-level scopes |
| **Scope** | Limited to user's data | Can access broader app resources |

## Use Cases

### 1. Webhook Callbacks
External services need to call your webhook endpoints:

```go
keyString, apiKey, err := appKeyService.GenerateWebhookKey(
    ctx,
    tenantID,
    appID,
    "payment-processor",
    []string{"webhook:receive", "payment:create"},
)
```

**Example**: Stripe webhook calling your `/webhooks/payment` endpoint to notify payment events.

### 2. Background Jobs
Scheduled tasks need to access your API:

```go
keyString, apiKey, err := appKeyService.GenerateBackgroundJobKey(
    ctx,
    tenantID,
    appID,
    "nightly-report",
    []string{"data:read", "report:generate"},
)
```

**Example**: Cron job running every night to generate reports.

### 3. Service-to-Service Authentication
Microservice A needs to call Microservice B:

```go
keyString, apiKey, err := appKeyService.GenerateServiceKey(
    ctx,
    tenantID,
    appID,
    "analytics-service",
    []string{"data:read", "analytics:write"},
)
```

**Example**: Analytics service fetching data from Core API.

### 4. CI/CD Pipelines
Deployment pipelines need to access API:

```go
keyString, apiKey, err := appKeyService.GenerateAppKey(ctx, &services.GenerateAppKeyRequest{
    TenantID:    tenantID,
    AppID:       appID,
    Name:        "github-actions-deploy",
    Purpose:     "ci_cd",
    KeyType:     apikey.KeyTypeSecret,
    Environment: apikey.EnvLive,
    Scopes:      []string{"deploy:execute", "config:read"},
})
```

### 5. Third-Party Integrations
External systems integrating with your platform:

```go
// Generate key for Zapier integration
keyString, apiKey, err := appKeyService.GenerateAppKey(ctx, &services.GenerateAppKeyRequest{
    TenantID:    tenantID,
    AppID:       appID,
    Name:        "zapier-integration",
    Purpose:     "third_party",
    Description: "Zapier workflow automation",
    KeyType:     apikey.KeyTypeRestricted,
    Environment: apikey.EnvLive,
    Scopes:      []string{"task:create", "task:read"},
    ExpiresIn:   &expiry, // 1 year
})
```

## Key Concepts

### Key Format

App API Keys follow the same format as user API keys:

```
{prefix}_{env}_{key_id}.{secret}
```

Examples:
- `sk_live_d1ZRTeuDd9ByP2l9eDbA7Q.gkXMNvOEUf3YguSoRuu_2myQwRRJ5Zq64taYDPazAFU`
- `sv_live_w2laip2Q4X08N8-WIj3ClQ.BVYdRpM7I3Ri5s7zr4jIKxWE8U1fkN0LjZUUHRQGKNA`

### Key Types

```go
type KeyType string

const (
    KeyTypeSecret     KeyType = "sk" // Secret Key (server-side only)
    KeyTypePublic     KeyType = "pk" // Public Key (can be exposed client-side)
    KeyTypeRestricted KeyType = "rk" // Restricted Key (limited scope)
    KeyTypeService    KeyType = "sv" // Service Key (service-to-service only)
)
```

### Environments

```go
type Environment string

const (
    EnvLive Environment = "live" // Production
    EnvTest Environment = "test" // Development/Testing
)
```

### Subject Identification

App API keys use a special subject format:

```go
subject := "app:{appID}"
// Example: "app:app_17d7b98c935ee0697a556a81967967b9"
```

This differentiates app authentication from user authentication.

## Architecture

### Service Structure

```
┌─────────────────────────────────────────────────────────┐
│                    AppKeyService                        │
├─────────────────────────────────────────────────────────┤
│ - GenerateAppKey()                                      │
│ - GenerateWebhookKey()                                  │
│ - GenerateBackgroundJobKey()                            │
│ - GenerateServiceKey()                                  │
│ - RotateAppKey()                                        │
│ - RevokeAppKey()                                        │
│ - ListAppKeys()                                         │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ├──> AppKeyStore (persistence)
                  │
                  ├──> AppService (validation)
                  │
                  └──> apikey.Authenticator (authentication)
```

### Data Flow

#### 1. Key Generation

```
AppKeyService.GenerateWebhookKey()
    ↓
Validate tenant/app
    ↓
Generate key string (prefix + key_id + secret)
    ↓
Hash secret
    ↓
Store APIKey record
    ↓
Return key string (ONLY TIME VISIBLE)
```

#### 2. Authentication

```
Incoming Request with API Key
    ↓
apikey.Authenticator.Authenticate()
    ↓
Parse key_id and secret from key string
    ↓
Lookup APIKey by key_id
    ↓
Verify tenant/app match
    ↓
Compare secret hash
    ↓
Check revocation/expiration
    ↓
Return AuthenticationResult with app:{appID} as subject
```

## Usage Guide

### 1. Setup Services

```go
// Initialize stores
tenantStore := services.NewInMemoryTenantStore()
appStore := services.NewInMemoryAppStore()
appKeyStore := services.NewInMemoryAppKeyStore()

// Create services
tenantService := services.NewTenantService(tenantStore)
appService := services.NewAppService(appStore, tenantService)
appKeyService := services.NewAppKeyService(appKeyStore, appStore)
```

### 2. Create Tenant and App

```go
// Create tenant
tenant, err := tenantService.CreateTenant(ctx, "acme-corp", "ACME Corporation")

// Create service app
app, err := appService.CreateApp(ctx, tenant.ID, "webhook-service", 
    core.AppTypeService, core.AppConfig{
        AllowedScopes: []string{"webhook:receive", "data:write"},
    })
```

### 3. Generate App API Key

```go
// Generate webhook key
keyString, apiKey, err := appKeyService.GenerateWebhookKey(
    ctx,
    tenant.ID,
    app.ID,
    "stripe-webhook",
    []string{"webhook:receive", "payment:create"},
)

// ⚠️ IMPORTANT: Save keyString immediately - it's only returned once!
fmt.Printf("API Key: %s\n", keyString)
```

### 4. Use Key for Authentication

```go
// Setup authenticator
authenticator := apikey.NewAuthenticator(&apikey.Config{
    KeyStore: appKeyStore,
})

// Create auth context
authCtx := &credential.AuthContext{
    TenantID: tenant.ID,
    AppID:    app.ID,
}

// Create credentials with API key from request header
creds := &apikey.Credentials{
    APIKey: r.Header.Get("X-API-Key"),
}

// Authenticate
result, err := authenticator.Authenticate(ctx, authCtx, creds)
if err != nil {
    return fmt.Errorf("auth error: %w", err)
}

if !result.Success {
    return fmt.Errorf("auth failed: %w", result.Error)
}

// Access claims
subject := result.Subject        // "app:app_xxx"
keyName := result.Claims["key_name"]
scopes := result.Claims["scopes"]
```

### 5. List App Keys

```go
keys, err := appKeyService.ListAppKeys(ctx, tenant.ID, app.ID)
for _, key := range keys {
    fmt.Printf("Key: %s, Scopes: %v\n", key.Name, key.Scopes)
}
```

### 6. Rotate Key

```go
// Best practice: Rotate keys every 90 days
newKeyString, newKey, err := appKeyService.RotateAppKey(ctx, oldKey.ID)

// Old key is automatically revoked
// Update the key in external service configuration
```

### 7. Revoke Key

```go
// Immediately revoke compromised key
err := appKeyService.RevokeAppKey(ctx, apiKey.ID)
```

## Security Best Practices

### 1. Regular Rotation

```go
// Rotate keys every 90 days
expiry := 90 * 24 * time.Hour
keyString, apiKey, err := appKeyService.GenerateWebhookKey(
    ctx, tenantID, appID, "webhook", scopes,
)

// Set reminder to rotate before expiration
```

### 2. Principle of Least Privilege

```go
// ❌ BAD: Too broad
scopes := []string{"*:*"}

// ✅ GOOD: Specific scopes only
scopes := []string{"webhook:receive", "payment:create"}
```

### 3. Separate Keys per Integration

```go
// ❌ BAD: One key for everything
singleKey := generateKey("master-key", allScopes)

// ✅ GOOD: Separate keys per service
stripeKey := generateWebhookKey("stripe", stripeScopes)
githubKey := generateWebhookKey("github", githubScopes)
cronKey := generateBackgroundJobKey("cron", cronScopes)
```

### 4. Monitor Usage

```go
keys, _ := appKeyService.ListAppKeys(ctx, tenantID, appID)
for _, key := range keys {
    if key.LastUsed == nil {
        // Key never used - consider removing
        log.Printf("Unused key: %s", key.Name)
    } else if time.Since(*key.LastUsed) > 90*24*time.Hour {
        // Key not used in 90 days
        log.Printf("Inactive key: %s", key.Name)
    }
}
```

### 5. Secure Storage

```go
// ✅ Store key in secure vault (HashiCorp Vault, AWS Secrets Manager)
vault.SetSecret("stripe-webhook-key", keyString)

// ✅ Use environment variables
os.Setenv("WEBHOOK_API_KEY", keyString)

// ❌ NEVER commit to git
// ❌ NEVER log in plain text
// ❌ NEVER store in database unencrypted
```

### 6. Use HTTPS Only

```go
// ✅ Always use HTTPS for API requests
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
        },
    },
}

req.Header.Set("X-API-Key", apiKey)
resp, err := client.Do(req)
```

### 7. Rate Limiting

```go
// Configure rate limits per app
appConfig := core.AppConfig{
    RateLimits: core.RateLimitConfig{
        Enabled:        true,
        RequestsPerMin: 100,
        BurstSize:      20,
    },
}
```

## API Reference

### AppKeyService Methods

#### GenerateAppKey

```go
func (s *AppKeyService) GenerateAppKey(
    ctx context.Context, 
    req *GenerateAppKeyRequest,
) (keyString string, apiKey *APIKey, err error)
```

Generates a new app API key with full customization.

**Parameters:**
- `TenantID` - Required tenant ID
- `AppID` - Required app ID
- `Name` - Required descriptive name
- `Purpose` - Purpose category (webhook, background_job, etc.)
- `Description` - Optional detailed description
- `KeyType` - Key type (sk, pk, rk, sv)
- `Environment` - Environment (live, test)
- `Scopes` - Allowed permissions
- `ExpiresIn` - Optional expiration duration

**Returns:** Full key string (only visible once), APIKey object, error

#### GenerateWebhookKey

```go
func (s *AppKeyService) GenerateWebhookKey(
    ctx context.Context,
    tenantID, appID, webhookName string,
    allowedScopes []string,
) (keyString string, apiKey *APIKey, err error)
```

Convenience method for webhook keys. Sets 90-day expiration.

#### GenerateBackgroundJobKey

```go
func (s *AppKeyService) GenerateBackgroundJobKey(
    ctx context.Context,
    tenantID, appID, jobName string,
    allowedScopes []string,
) (keyString string, apiKey *APIKey, err error)
```

Convenience method for background job keys. No expiration.

#### GenerateServiceKey

```go
func (s *AppKeyService) GenerateServiceKey(
    ctx context.Context,
    tenantID, appID, serviceName string,
    allowedScopes []string,
) (keyString string, apiKey *APIKey, err error)
```

Convenience method for service-to-service keys. Uses `sv` prefix, no expiration.

#### RotateAppKey

```go
func (s *AppKeyService) RotateAppKey(
    ctx context.Context,
    oldKeyID string,
) (newKeyString string, newAPIKey *APIKey, err error)
```

Rotates a key by generating new one and revoking old one.

#### RevokeAppKey

```go
func (s *AppKeyService) RevokeAppKey(
    ctx context.Context,
    keyID string,
) error
```

Immediately revokes an app API key.

#### ListAppKeys

```go
func (s *AppKeyService) ListAppKeys(
    ctx context.Context,
    tenantID, appID string,
) ([]*APIKey, error)
```

Lists all API keys for a specific app.

#### ListTenantAppKeys

```go
func (s *AppKeyService) ListTenantAppKeys(
    ctx context.Context,
    tenantID string,
) ([]*APIKey, error)
```

Lists all app API keys across all apps in a tenant.

## Example: Complete Webhook Integration

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"
    
    "github.com/primadi/lokstra-auth/core/services"
    "github.com/primadi/lokstra-auth/credential/apikey"
)

func main() {
    ctx := context.Background()
    
    // Setup
    appKeyService := setupAppKeyService()
    authenticator := setupAuthenticator()
    
    // Generate webhook key
    keyString, _, err := appKeyService.GenerateWebhookKey(
        ctx, "tenant-1", "app-1", "stripe-webhook",
        []string{"webhook:receive", "payment:write"},
    )
    if err != nil {
        panic(err)
    }
    
    // Give key to external service (Stripe)
    fmt.Printf("Add this key to Stripe webhook settings: %s\n", keyString)
    
    // Setup webhook endpoint
    http.HandleFunc("/webhooks/stripe", func(w http.ResponseWriter, r *http.Request) {
        // Extract API key from header
        apiKey := r.Header.Get("X-API-Key")
        if apiKey == "" {
            http.Error(w, "Missing API key", http.StatusUnauthorized)
            return
        }
        
        // Authenticate
        authCtx := &credential.AuthContext{
            TenantID: "tenant-1",
            AppID:    "app-1",
        }
        
        creds := &apikey.Credentials{APIKey: apiKey}
        
        result, err := authenticator.Authenticate(ctx, authCtx, creds)
        if err != nil || !result.Success {
            http.Error(w, "Invalid API key", http.StatusUnauthorized)
            return
        }
        
        // Check scopes
        scopes, _ := result.Claims["scopes"].([]string)
        if !contains(scopes, "webhook:receive") {
            http.Error(w, "Insufficient permissions", http.StatusForbidden)
            return
        }
        
        // Process webhook payload
        var payload map[string]any
        json.NewDecoder(r.Body).Decode(&payload)
        
        // Handle event
        handleStripeWebhook(payload)
        
        w.WriteHeader(http.StatusOK)
    })
    
    http.ListenAndServe(":8080", nil)
}
```

## See Also

- [User API Keys](../examples/credential/05_apikey/) - For user-level authentication
- [Authentication Guide](./credential.md) - Core authentication concepts
- [Multi-tenant Architecture](./architecture.md) - Tenant and app isolation
- [Example: App Keys](../examples/core/01_app_keys/) - Complete working example
