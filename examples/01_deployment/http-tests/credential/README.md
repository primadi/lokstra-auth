# HTTP Tests for Lokstra Auth Credential Services

This folder contains HTTP test files for testing all credential authentication services (credential) of Lokstra Auth.

## Prerequisites

- Install REST Client extension for VS Code
- Ensure the server is running on `http://localhost:9090`
- Complete core setup first (tenant, app, user must exist)
- Configure credential providers in credential config service

## HTTP Request Headers

All credential authentication requests require the following headers:

```http
X-Tenant-ID: acme-corp
X-App-ID: main-app
X-Branch-ID: hq-jakarta  # Optional for most endpoints
Content-Type: application/json
```

**Why Headers Instead of Body?**
- ✅ **Security**: Separates authentication context from credentials
- ✅ **Industry Standard**: AWS, Stripe, Shopify use custom headers for tenant/account context
- ✅ **Middleware Friendly**: Headers can be extracted and validated before handler execution
- ✅ **Clean Separation**: Headers = context (who/where), Body = business data (what)

### Example Usage

```http
POST http://localhost:9090//api/auth/cred/basic/login
Content-Type: application/json
X-Tenant-ID: acme-corp
X-App-ID: main-app

{
  "username": "john.doe",
  "password": "SecurePass123!"
}
```

## Authentication Methods Overview

Lokstra Auth supports 5 different credential/authentication methods:

1. **Basic Auth** - Traditional username/password
2. **API Key** - Service-to-service authentication
3. **OAuth2** - Social login (Google, Azure, GitHub, Facebook)
4. **Passwordless** - Magic link & OTP via email/SMS
5. **Passkey** - WebAuthn/FIDO2 biometric authentication

## Test Files Overview

### 01-basic-auth-service.http
Tests for **username/password authentication**.

**Endpoints:**
- `POST /api/auth/cred/basic/login` - Authenticate with username/password
- `POST /api/auth/cred/basic/register` - Register new user with password
- `POST /api/auth/cred/basic/change-password` - Change password (user)
- `POST /api/auth/cred/basic/reset-password` - Reset password (admin)

**Features:**
- Login with username or email
- Password strength validation
- Account lockout after failed attempts
- User registration with credentials
- Password change & reset

**Test Scenarios:**
- Valid login
- Invalid password
- User not found
- Disabled user account
- Duplicate username/email
- Weak password validation
- Failed login lockout

**Required Variables:**
- `@tenantId`
- `@appId`
- User must exist in database

---

### 02-apikey-auth-service.http
Tests for **API key authentication** (service-to-service).

**Endpoints:**
- `POST //api/auth/cred/apikey/authenticate` - Validate API key

**Features:**
- API key validation
- Scope-based authorization
- Environment-specific keys (prod/staging/dev)
- Rate limiting per key
- Revoked/expired key detection

**API Key Format:**
```
{prefix}_{key_id}.{secret}
Example: appkey_0dabaa82466f9cc882559e8bea626c98.N8KqPz9vL6wR3mJ4fH2xTbY5gC1sA7nE
```

**Test Scenarios:**
- Valid API key
- Invalid format
- Non-existent key
- Revoked key
- Expired key
- Scope validation
- Rate limiting
- Environment verification

**Required Variables:**
- API keys generated from core/05-app-key-service.http
- Must have valid key with secret

---

### 03-oauth2-auth-service.http
Tests for **OAuth2 social login**.

**Endpoints:**
- `POST //api/auth/cred/oauth2/authorize` - Initiate OAuth2 flow
- `GET //api/auth/cred/oauth2/callback` - Handle provider callback

**Supported Providers:**
- **Google** - Google Workspace, Gmail accounts
- **Azure** - Microsoft/Azure AD accounts
- **GitHub** - GitHub accounts
- **Facebook** - Facebook accounts
- Custom OAuth2 providers

**OAuth2 Flow:**
```
1. Client initiates authorization
   POST /authorize { provider, redirect_uri, scopes, state }

2. Server returns authorization_url
   { authorization_url: "https://provider.com/oauth2/authorize?..." }

3. User is redirected to provider
4. User authorizes the app
5. Provider redirects back to callback
   GET /callback?code=...&state=...

6. Server exchanges code for tokens
7. Server returns user info & access token
```

**Test Scenarios:**
- Google OAuth2 flow
- Azure OAuth2 flow
- GitHub OAuth2 flow
- Invalid provider
- Provider not enabled
- User denied access
- State mismatch (CSRF protection)
- Expired state
- Invalid redirect URI

**Configuration:**
- OAuth2 providers must be configured in credential config
- Each provider needs client_id, client_secret, callback_url

---

### 04-passwordless-auth-service.http
Tests for **passwordless authentication** (magic link & OTP).

**Endpoints:**

**OTP (One-Time Password):**
- `POST //api/auth/cred/passwordless/send-code` - Send OTP via email/SMS
- `POST //api/auth/cred/passwordless/verify-code` - Verify OTP code
- `POST //api/auth/cred/passwordless/resend-code` - Resend OTP

**Magic Link:**
- `POST //api/auth/cred/passwordless/send-magic-link` - Send magic link via email
- `GET //api/auth/cred/passwordless/verify-magic-link/{token}` - Verify magic link

**Features:**

**OTP:**
- 6-digit code (configurable)
- Email or SMS delivery
- Configurable expiration (default: 5 minutes)
- Max attempts limit
- Resend with cooldown period

**Magic Link:**
- Token-based authentication
- Configurable expiration (default: 10 minutes)
- One-time use
- Redirect to specified URL after verification

**Test Scenarios:**
- Send OTP via email
- Send OTP via SMS
- Verify valid code
- Verify invalid code
- Verify expired code
- Resend code
- Send magic link
- Verify magic link
- Rate limiting
- Max attempts exceeded
- Cooldown period

**Security Features:**
- Rate limiting (max sends per minute)
- Max verification attempts
- Cooldown period between sends
- Code expiration
- One-time use enforcement

**Required Variables:**
- `@tenantId`
- `@appId`
- User with valid email/phone
- Email/SMS delivery service configured

---

### 05-passkey-auth-service.http
Tests for **WebAuthn/FIDO2 passkey authentication**.

**Endpoints:**

**Registration:**
- `POST //api/auth/cred/passkey/register/begin` - Start passkey registration
- `POST //api/auth/cred/passkey/register/finish` - Complete passkey registration

**Authentication:**
- `POST //api/auth/cred/passkey/authenticate/begin` - Start authentication
- `POST //api/auth/cred/passkey/authenticate/finish` - Complete authentication

**Management:**
- `GET //api/auth/cred/passkey/credentials/{user_id}` - List user's passkeys
- `DELETE //api/auth/cred/passkey/credentials/{credential_id}` - Remove passkey

**Features:**

**Authenticator Types:**
- **Platform Authenticators**: Touch ID, Face ID, Windows Hello
- **Cross-Platform Authenticators**: USB security keys, Bluetooth devices

**User Verification:**
- `required` - Always require biometric/PIN
- `preferred` - Request if available
- `discouraged` - Skip if possible

**Resident Keys:**
- Discoverable credentials (passwordless)
- Username-less authentication

**Important Notes:**
⚠️ **Passkey authentication requires WebAuthn-capable browser**
⚠️ **Cannot be fully tested via HTTP client alone**
⚠️ **Tests show request/response format for integration**
⚠️ **Use browser DevTools for actual WebAuthn testing**

**Test Scenarios:**
- Register platform authenticator
- Register cross-platform authenticator
- Register resident key (passwordless)
- Authenticate with passkey
- Discoverable credential flow
- List user credentials
- Delete credential
- Multi-device registration

**WebAuthn Flow:**

**Registration:**
```
1. Begin registration
   POST /register/begin
   Response: PublicKeyCredentialCreationOptions

2. Browser calls navigator.credentials.create()
3. User authenticates with biometric/PIN
4. Complete registration
   POST /register/finish (with credential from WebAuthn)
```

**Authentication:**
```
1. Begin authentication
   POST /authenticate/begin
   Response: PublicKeyCredentialRequestOptions

2. Browser calls navigator.credentials.get()
3. User authenticates with biometric/PIN
4. Complete authentication
   POST /authenticate/finish (with assertion from WebAuthn)
```

**Required Variables:**
- `@tenantId`
- `@appId`
- `@userId`
- WebAuthn-capable browser (Chrome 67+, Firefox 60+, Safari 13+)

---

## Quick Start Guide

### Option 1: Test Individual Auth Methods

**Basic Auth:**
```
1. Open 01-basic-auth-service.http
2. Ensure user exists (create from core/04-user-service.http)
3. Test login with username/password
```

**API Key:**
```
1. Generate API key from core/05-app-key-service.http
2. Copy the full key (prefix + secret)
3. Open 02-apikey-auth-service.http
4. Test authentication with API key
```

**OAuth2:**
```
1. Configure OAuth2 provider in credential config
2. Open 03-oauth2-auth-service.http
3. Initiate authorization flow
4. Note: Full flow requires browser interaction
```

**Passwordless:**
```
1. Configure email/SMS service
2. Open 04-passwordless-auth-service.http
3. Send OTP or magic link
4. Verify code/link
```

**Passkey:**
```
1. Open 05-passkey-auth-service.http
2. Note: Requires WebAuthn browser testing
3. Review request/response formats
```

### Option 2: Complete Authentication Workflow

1. **Setup** (from core):
   - Create tenant
   - Create app
   - Configure credential config
   - Create users
   - Generate API keys

2. **Test Each Method**:
   - Basic: Username/password login
   - API Key: Service authentication
   - OAuth2: Social login
   - Passwordless: OTP/Magic link
   - Passkey: Biometric auth

## Common Variables

```http
@baseUrl = http://localhost:9090
@contentType = application/json
@tenantId = acme-corp
@appId = main-app
@branchId = hq-jakarta
@userId = user_abc123
```

## Configuration Requirements

Before testing, ensure credential providers are enabled:

### Enable Basic Auth
```http
PUT /api/auth/core/config/credentials/tenants/{{tenantId}}
{
  "config": {
    "enable_basic": true,
    "basic_config": {
      "min_password_length": 8,
      "require_strong_pwd": true,
      "max_login_attempts": 5,
      "lockout_duration_secs": 300
    }
  }
}
```

### Enable API Key Auth
```http
PUT /api/auth/core/config/credentials/tenants/{{tenantId}}
{
  "config": {
    "enable_apikey": true,
    "apikey_config": {
      "secret_length": 32,
      "hash_algo": "sha3-256",
      "default_expiry_days": 365,
      "rate_limit_per_minute": 60
    }
  }
}
```

### Enable OAuth2
```http
PUT /api/auth/core/config/credentials/tenants/{{tenantId}}
{
  "config": {
    "enable_oauth2": true,
    "oauth2_config": {
      "providers": [
        {
          "name": "google",
          "enabled": true,
          "client_id": "your-client-id",
          "client_secret": "your-secret",
          "scopes": ["openid", "email", "profile"]
        }
      ],
      "callback_url": "http://localhost:9090//api/auth/cred/oauth2/callback"
    }
  }
}
```

### Enable Passwordless
```http
PUT /api/auth/core/config/credentials/tenants/{{tenantId}}
{
  "config": {
    "enable_passwordless": true,
    "passwordless_config": {
      "enable_email": true,
      "enable_sms": false,
      "code_length": 6,
      "code_expiry_secs": 300,
      "max_attempts_per_email": 3
    }
  }
}
```

### Enable Passkey
```http
PUT /api/auth/core/config/credentials/tenants/{{tenantId}}
{
  "config": {
    "enable_passkey": true,
    "passkey_config": {
      "rp_name": "Acme Corporation",
      "rp_id": "localhost",
      "rp_origins": ["http://localhost:9090"],
      "user_verification": "preferred"
    }
  }
}
```

## Authentication Method Comparison

| Method | Security | UX | Use Case |
|--------|----------|----|---------| 
| **Basic** | ⭐⭐⭐ | ⭐⭐⭐⭐ | Traditional apps, internal tools |
| **API Key** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Service-to-service, APIs |
| **OAuth2** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Social login, SSO |
| **Passwordless** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Modern web/mobile apps |
| **Passkey** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Highest security, passwordless |

## Security Best Practices

### Basic Auth
- ✅ Enforce strong password policies
- ✅ Implement account lockout after failed attempts
- ✅ Use secure password hashing (bcrypt)
- ✅ Enable rate limiting
- ⚠️ Never log passwords
- ⚠️ Use HTTPS in production

### API Key
- ✅ Rotate keys periodically
- ✅ Use environment-specific keys
- ✅ Implement scope-based permissions
- ✅ Store secrets securely (shown only once!)
- ✅ Revoke compromised keys immediately
- ⚠️ Never commit keys to version control
- ⚠️ Use short expiration for production keys

### OAuth2
- ✅ Validate state parameter (CSRF protection)
- ✅ Use HTTPS for callback URLs
- ✅ Verify token signatures
- ✅ Store tokens securely
- ⚠️ Never expose client_secret
- ⚠️ Implement token refresh logic

### Passwordless
- ✅ Implement rate limiting
- ✅ Use short code expiration
- ✅ Enforce cooldown between sends
- ✅ Limit max verification attempts
- ⚠️ Secure email/SMS delivery
- ⚠️ One-time use enforcement

### Passkey
- ✅ Require user verification
- ✅ Use platform authenticators when possible
- ✅ Support multiple devices
- ✅ Implement credential management
- ⚠️ Validate attestation in enterprise
- ⚠️ Backup authentication method

## Troubleshooting

### Basic Auth Issues
```
Error: Invalid username or password
→ Check user exists and password is correct
→ Verify user account is not disabled
→ Check if account is locked due to failed attempts
```

### API Key Issues
```
Error: Invalid API key format
→ Ensure format: {prefix}_{key_id}.{secret}
→ Check key was copied completely
→ Verify key hasn't been revoked or expired
```

### OAuth2 Issues
```
Error: Provider not enabled
→ Check provider is enabled in credential config
→ Verify client_id and client_secret are correct
→ Ensure callback URL matches configuration

Error: State mismatch
→ State parameter expired (default: 10 minutes)
→ Check CSRF protection is working
```

### Passwordless Issues
```
Error: Rate limit exceeded
→ Wait for cooldown period
→ Check max attempts configuration

Error: Invalid code
→ Code may have expired (default: 5 minutes)
→ Verify correct code was entered
→ Check max attempts not exceeded
```

### Passkey Issues
```
Error: WebAuthn not supported
→ Use compatible browser (Chrome 67+, Safari 13+)
→ HTTPS required in production
→ Check device has authenticator

Error: No credentials found
→ User must register passkey first
→ Credential may have been deleted
```

## Example Multi-Method Workflow

**Complete User Journey:**

1. **First-Time User Registration**
   ```http
   POST //api/auth/cred/basic/register
   { username, email, password }
   ```

2. **Setup Passwordless (Optional)**
   ```http
   POST //api/auth/cred/passwordless/send-code
   Verify email ownership
   ```

3. **Add Passkey for Convenience**
   ```http
   POST //api/auth/cred/passkey/register/begin
   Complete with WebAuthn
   ```

4. **Generate API Key for Integrations**
   ```http
   POST /api/auth/core/tenants/{tenant}/apps/{app}/keys
   ```

5. **Link Social Accounts**
   ```http
   POST //api/auth/cred/oauth2/authorize { provider: "google" }
   ```

**Future Logins:**
- Quick: Passkey (biometric)
- Fallback 1: Passwordless (email OTP)
- Fallback 2: Basic (username/password)
- Or: OAuth2 (social login)

## Production Deployment

**HTTPS Only:**
- All credential endpoints must use HTTPS
- OAuth2 callbacks require HTTPS
- Passkey requires secure context

**Environment Variables:**
```bash
# OAuth2
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
AZURE_CLIENT_ID=...
AZURE_CLIENT_SECRET=...

# Email/SMS
SMTP_HOST=...
SMTP_USER=...
SMS_API_KEY=...

# Security
JWT_SECRET=...
ENCRYPTION_KEY=...
```

**Rate Limiting:**
- Configure per-endpoint limits
- Implement IP-based throttling
- Monitor for brute force attacks

## Next Steps

After testing credential services, proceed to:
- **token**: Token generation and validation
- **subject**: User context and claims
- **authz**: Authorization policies

---

Now ready for comprehensive authentication testing! 🔐
