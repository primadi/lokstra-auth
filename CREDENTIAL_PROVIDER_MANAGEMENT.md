# Credential Provider Configuration

## Overview

The Credential Provider system manages authentication provider configurations (OAuth2, SAML, Email) with support for **multiple configurations per provider type** within a tenant.

This enables scenarios like:
- **Multi-environment**: Separate Google OAuth2 configs for dev, staging, production
- **Multi-app**: Different Google Client IDs for web-portal vs mobile-app
- **Tenant-level defaults**: Fallback configuration when app doesn't have specific override

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Tenant: acme-corp                                           │
├─────────────────────────────────────────────────────────────┤
│ Credential Providers:                                       │
│                                                              │
│ 1. Google OAuth (Tenant Default)                            │
│    - app_id: NULL                                            │
│    - client_id: tenant-default.apps.googleusercontent.com   │
│    - Used by: all apps without specific override            │
│                                                              │
│ 2. Google OAuth (Web Portal)                                │
│    - app_id: web-portal                                     │
│    - client_id: web-portal.apps.googleusercontent.com       │
│    - Used by: web-portal app only                           │
│                                                              │
│ 3. Google OAuth (Mobile App)                                │
│    - app_id: mobile-app                                     │
│    - client_id: mobile-app.apps.googleusercontent.com       │
│    - Used by: mobile-app only                               │
│                                                              │
│ 4. GitHub OAuth (Mobile App)                                │
│    - app_id: mobile-app                                     │
│    - client_id: Iv1.mobile-client-id                        │
│                                                              │
│ 5. SAML (Enterprise SSO - Tenant Default)                   │
│    - app_id: NULL                                            │
│    - entity_id: https://sso.acme-corp.com/saml/metadata     │
└─────────────────────────────────────────────────────────────┘
```

## Domain Model

### CredentialProvider Entity

```go
type CredentialProvider struct {
    ID          string                 // e.g., "google-web-portal"
    TenantID    string                 // e.g., "acme-corp"
    AppID       string                 // "" = tenant-level, "web-portal" = app-specific
    Type        ProviderType           // oauth2_google, oauth2_github, saml, email
    Name        string                 // e.g., "Google OAuth (Web Portal)"
    Description string
    Status      ProviderStatus         // active, disabled
    Config      map[string]any // Provider-specific config
    Metadata    *map[string]any
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### Supported Provider Types

```go
const (
    ProviderTypeOAuth2Google     ProviderType = "oauth2_google"
    ProviderTypeOAuth2GitHub     ProviderType = "oauth2_github"
    ProviderTypeOAuth2Microsoft  ProviderType = "oauth2_microsoft"
    ProviderTypeOAuth2Facebook   ProviderType = "oauth2_facebook"
    ProviderTypeOAuth2Apple      ProviderType = "oauth2_apple"
    ProviderTypeOAuth2LinkedIn   ProviderType = "oauth2_linkedin"
    ProviderTypeOAuth2Twitter    ProviderType = "oauth2_twitter"
    ProviderTypeSAML             ProviderType = "saml"
    ProviderTypeOIDC             ProviderType = "oidc"
    ProviderTypeEmail            ProviderType = "email"
    ProviderTypePasskey          ProviderType = "passkey"
    ProviderTypeMagicLink        ProviderType = "magic_link"
    ProviderTypeOTP              ProviderType = "otp"
)
```

### Configuration Structures

#### OAuth2Config
```go
{
    "client_id": "123456789-abc.apps.googleusercontent.com",
    "client_secret": "GOCSPX-secret-key",
    "redirect_uri": "https://app.example.com/callback/google",
    "scopes": ["openid", "profile", "email"],
    "authorize_url": "https://accounts.google.com/o/oauth2/v2/auth",
    "token_url": "https://oauth2.googleapis.com/token",
    "userinfo_url": "https://www.googleapis.com/oauth2/v3/userinfo"
}
```

#### SAMLConfig
```go
{
    "entity_id": "https://sso.example.com/saml/metadata",
    "sso_url": "https://sso.example.com/saml/sso",
    "slo_url": "https://sso.example.com/saml/logout",
    "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
    "attributes_mapping": {
        "email": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
        "first_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
        "last_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
    }
}
```

#### EmailConfig (SMTP)
```go
{
    "smtp_host": "smtp.sendgrid.net",
    "smtp_port": 587,
    "smtp_username": "apikey",
    "smtp_password": "SG.secret-key",
    "from_email": "noreply@example.com",
    "from_name": "Example App",
    "use_tls": true
}
```

## Service Operations

### CredentialProviderService

```go
type CredentialProviderService struct {
    store repository.CredentialProviderStore
}

// CRUD operations
func (s *CredentialProviderService) CreateProvider(ctx context.Context, req *CreateProviderRequest) (*CredentialProvider, error)
func (s *CredentialProviderService) GetProvider(ctx context.Context, req *GetProviderRequest) (*CredentialProvider, error)
func (s *CredentialProviderService) UpdateProvider(ctx context.Context, req *UpdateProviderRequest) (*CredentialProvider, error)
func (s *CredentialProviderService) DeleteProvider(ctx context.Context, req *DeleteProviderRequest) error
func (s *CredentialProviderService) ListProviders(ctx context.Context, req *ListProvidersRequest) ([]*CredentialProvider, error)

// Status management
func (s *CredentialProviderService) EnableProvider(ctx context.Context, req *EnableProviderRequest) (*CredentialProvider, error)
func (s *CredentialProviderService) DisableProvider(ctx context.Context, req *DisableProviderRequest) (*CredentialProvider, error)

// Query operations
func (s *CredentialProviderService) GetProvidersByType(ctx context.Context, req *GetProvidersByTypeRequest) ([]*CredentialProvider, error)

// Resolution logic with fallback
func (s *CredentialProviderService) GetActiveProviderForApp(ctx context.Context, req *GetActiveProviderForAppRequest) (*CredentialProvider, error)
```

### Resolution Logic: GetActiveProviderForApp

This method implements **app-level override with tenant-level fallback**:

```go
func (s *CredentialProviderService) GetActiveProviderForApp(
    ctx context.Context, 
    req *GetActiveProviderForAppRequest,
) (*CredentialProvider, error) {
    // Step 1: Try to find app-specific provider (active only)
    appProviders, err := s.store.ListByType(ctx, req.TenantID, req.AppID, req.Type)
    if err != nil {
        return nil, err
    }
    
    for _, provider := range appProviders {
        if provider.AppID == req.AppID && provider.Status == ProviderStatusActive {
            return provider, nil // App-level override found
        }
    }
    
    // Step 2: Fallback to tenant-level provider (app_id = "")
    tenantProviders, err := s.store.ListByType(ctx, req.TenantID, "", req.Type)
    if err != nil {
        return nil, err
    }
    
    for _, provider := range tenantProviders {
        if provider.Status == ProviderStatusActive {
            return provider, nil // Tenant-level default found
        }
    }
    
    return nil, ErrProviderNotFound
}
```

**Example Scenarios:**

```
Tenant: acme-corp
Providers:
  - ID: google-default, Type: oauth2_google, AppID: "" (tenant-level)
  - ID: google-web, Type: oauth2_google, AppID: "web-portal"

Query 1: GetActiveProviderForApp(tenant=acme-corp, app=web-portal, type=oauth2_google)
Result: Returns "google-web" (app-level override)

Query 2: GetActiveProviderForApp(tenant=acme-corp, app=mobile-app, type=oauth2_google)
Result: Returns "google-default" (tenant-level fallback)

Query 3: GetActiveProviderForApp(tenant=acme-corp, app=api-service, type=oauth2_google)
Result: Returns "google-default" (tenant-level fallback)
```

## Database Schema

```sql
CREATE TABLE credential_providers (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255),  -- NULL = tenant-level, non-NULL = app-specific
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    config JSONB NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (tenant_id, id),
    
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id, app_id) REFERENCES apps(tenant_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_credential_providers_type ON credential_providers(tenant_id, type);
CREATE INDEX idx_credential_providers_app_id ON credential_providers(tenant_id, app_id) WHERE app_id IS NOT NULL;
CREATE INDEX idx_credential_providers_status ON credential_providers(tenant_id, status);
```

## Use Cases

### UC1: Multi-Environment OAuth2 (Dev/Staging/Prod)

```json
// Development
{
  "id": "google-dev",
  "tenant_id": "acme-corp",
  "app_id": "web-portal",
  "type": "oauth2_google",
  "name": "Google OAuth (Dev)",
  "config": {
    "client_id": "dev-123.apps.googleusercontent.com",
    "redirect_uri": "http://localhost:3000/callback/google"
  },
  "metadata": {
    "environment": "development"
  }
}

// Production
{
  "id": "google-prod",
  "tenant_id": "acme-corp",
  "app_id": "web-portal",
  "type": "oauth2_google",
  "name": "Google OAuth (Prod)",
  "config": {
    "client_id": "prod-456.apps.googleusercontent.com",
    "redirect_uri": "https://app.example.com/callback/google"
  },
  "metadata": {
    "environment": "production"
  }
}
```

**Resolution:**
- Application selects provider based on `metadata.environment`
- Or uses separate provider IDs and switches based on ENV variable

### UC2: Multi-App Different Client IDs

```
Tenant: enterprise-inc

App 1: customer-portal
  Provider: google-customer
    - ClientID: customer-123.apps.googleusercontent.com
    - Scopes: [openid, profile, email]

App 2: admin-dashboard
  Provider: google-admin
    - ClientID: admin-456.apps.googleusercontent.com
    - Scopes: [openid, profile, email, https://www.googleapis.com/auth/admin.directory.user]
```

**Benefit:** Different OAuth2 consent screens and permissions per application.

### UC3: Tenant Default + App Override

```
Tenant: multi-app-org

Tenant-level Google Provider (Default):
  - AppID: NULL
  - ClientID: default-123.apps.googleusercontent.com
  - Used by: app-a, app-b, app-d

App-level Google Provider (Override):
  - AppID: app-c
  - ClientID: appc-456.apps.googleusercontent.com
  - Used by: app-c only
```

**Benefit:** Centralized default config with selective overrides.

### UC4: Multiple Social Providers per App

```
Tenant: social-hub
App: web-portal

Providers:
1. Google OAuth
   - Type: oauth2_google
   - ClientID: google-123.apps.googleusercontent.com

2. GitHub OAuth
   - Type: oauth2_github
   - ClientID: Iv1.github-client-id

3. Microsoft OAuth
   - Type: oauth2_microsoft
   - ClientID: microsoft-app-id
   - Tenant: common

4. SAML (Enterprise)
   - Type: saml
   - EntityID: https://sso.enterprise.com/saml/metadata
```

**Benefit:** Users can choose their preferred login method.

## Integration with Authentication Flow

### OAuth2 Login Flow

```go
// 1. User clicks "Sign in with Google" for app "web-portal"

// 2. Retrieve active Google provider for this app
providerReq := &GetActiveProviderForAppRequest{
    TenantID: "acme-corp",
    AppID:    "web-portal",
    Type:     ProviderTypeOAuth2Google,
}
provider, err := credentialProviderService.GetActiveProviderForApp(ctx, providerReq)

// 3. Extract OAuth2 config
oauth2Config := provider.GetOAuth2Config()
// oauth2Config.ClientID = "web-portal-123.apps.googleusercontent.com"
// oauth2Config.RedirectURI = "https://web.example.com/callback/google"

// 4. Generate authorization URL
authURL := fmt.Sprintf(
    "%s?client_id=%s&redirect_uri=%s&scope=%s&response_type=code&state=%s",
    oauth2Config.AuthorizeURL,
    oauth2Config.ClientID,
    oauth2Config.RedirectURI,
    strings.Join(oauth2Config.Scopes, " "),
    generateState(),
)

// 5. Redirect user to Google
http.Redirect(w, r, authURL, http.StatusFound)

// 6. Handle callback with provider's ClientID + ClientSecret
// (Exchange authorization code for tokens)
```

### SAML Login Flow

```go
// 1. User initiates SAML login

// 2. Retrieve SAML provider
provider, err := credentialProviderService.GetActiveProviderForApp(ctx, &GetActiveProviderForAppRequest{
    TenantID: "enterprise-inc",
    AppID:    "intranet-portal",
    Type:     ProviderTypeSAML,
})

// 3. Extract SAML config
samlConfig := provider.GetSAMLConfig()
// samlConfig.SSOURL = "https://sso.enterprise.com/saml/sso"

// 4. Generate SAML AuthnRequest
authnRequest := generateSAMLAuthnRequest(
    samlConfig.EntityID,
    samlConfig.SSOURL,
)

// 5. Redirect to IdP
http.Redirect(w, r, samlConfig.SSOURL+"?SAMLRequest="+base64Encode(authnRequest), http.StatusFound)
```

## API Examples

See `examples/01_deployment/http-tests/core/08-credential-provider-service.http` for complete HTTP test suite.

### Create Tenant-level Default

```http
POST /credential-providers
Content-Type: application/json

{
  "tenant_id": "acme-corp",
  "app_id": "",
  "type": "oauth2_google",
  "name": "Google OAuth (Default)",
  "config": {
    "client_id": "tenant-default.apps.googleusercontent.com",
    "client_secret": "GOCSPX-tenant-secret",
    "redirect_uri": "https://auth.example.com/callback/google",
    "scopes": ["openid", "profile", "email"]
  }
}
```

### Create App-level Override

```http
POST /credential-providers
Content-Type: application/json

{
  "tenant_id": "acme-corp",
  "app_id": "web-portal",
  "type": "oauth2_google",
  "name": "Google OAuth (Web Portal)",
  "config": {
    "client_id": "web-portal.apps.googleusercontent.com",
    "client_secret": "GOCSPX-web-secret",
    "redirect_uri": "https://web.example.com/callback/google",
    "scopes": ["openid", "profile", "email", "https://www.googleapis.com/auth/drive.readonly"]
  }
}
```

### Get Active Provider (with Fallback)

```http
GET /credential-providers/active?tenant_id=acme-corp&app_id=web-portal&type=oauth2_google
```

Response:
```json
{
  "id": "google-web-portal",
  "tenant_id": "acme-corp",
  "app_id": "web-portal",
  "type": "oauth2_google",
  "name": "Google OAuth (Web Portal)",
  "status": "active",
  "config": {
    "client_id": "web-portal.apps.googleusercontent.com",
    "scopes": ["openid", "profile", "email", "https://www.googleapis.com/auth/drive.readonly"]
  }
}
```

## Security Considerations

1. **Secrets Storage**: Store `client_secret` encrypted in production
2. **Access Control**: Only authorized users can create/modify providers
3. **Audit Logging**: Track all provider configuration changes
4. **Validation**: Validate OAuth2 redirect URIs against whitelist
5. **Certificate Validation**: Verify SAML certificates on creation

## Migration from Legacy System

If you have existing provider configs embedded in `App.config`:

```go
// Old approach (single provider per app)
app.Config = map[string]any{
    "oauth2": map[string]any{
        "google_client_id": "old-client-id",
        "google_client_secret": "old-secret",
    },
}

// New approach (multiple providers per app)
provider := &CredentialProvider{
    ID:       "google-default",
    TenantID: app.TenantID,
    AppID:    app.ID,
    Type:     ProviderTypeOAuth2Google,
    Config: map[string]any{
        "client_id":     "old-client-id",
        "client_secret": "old-secret",
        "redirect_uri":  "https://app.example.com/callback/google",
        "scopes":        []string{"openid", "profile", "email"},
    },
}
credentialProviderService.CreateProvider(ctx, &CreateProviderRequest{Provider: provider})
```

## Best Practices

1. **Naming Convention**: Use descriptive names like `google-web-prod`, `github-mobile-dev`
2. **Tenant Defaults**: Always create tenant-level defaults for commonly used providers
3. **App Overrides**: Use app-level configs only when truly different (scopes, domains)
4. **Environment Separation**: Use `metadata.environment` to track dev/staging/prod configs
5. **Status Management**: Disable instead of delete to maintain audit trail
6. **Config Validation**: Validate OAuth2 redirect URIs, SAML certificates on creation

## Related Documentation

- [User Identity Management](./USER_IDENTITY_MANAGEMENT.md) - Linking provider identities to users
- [OAuth2 Integration](./docs/credential_providers.md) - OAuth2 flow details
- [SAML Integration](./docs/credential_providers.md) - SAML flow details
