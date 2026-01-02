# Lokstra Auth - Credential Providers Summary

## Available Authentication Methods

Lokstra Auth mendukung 5 metode autentikasi yang bisa dikonfigurasi per tenant/app:

### 1. **Basic Authentication** ✅ Implemented
- **Path**: `//api/auth/cred/basic/*`
- **Service**: `BasicAuthService`
- **Methods**: Username/Password
- **Endpoints**:
  - `POST /login` - User login
  - `POST /register` - User registration
  - `POST /change-password` - Change password
- **Configuration**: `CredentialConfig.EnableBasic`
- **Use Cases**: Traditional web apps, mobile apps, internal tools

### 2. **API Key Authentication** ✅ Implemented
- **Path**: `//api/auth/cred/apikey/*`
- **Service**: `APIKeyAuthService`
- **Methods**: API Key in header
- **Endpoints**:
  - `POST /authenticate` - Validate API key
- **Configuration**: `CredentialConfig.EnableAPIKey`
- **Use Cases**: Server-to-server, webhooks, background jobs, CI/CD

### 3. **OAuth2 Authentication** ✅ Implemented
- **Path**: `//api/auth/cred/oauth2/*`
- **Service**: `OAuth2AuthService`
- **Methods**: OAuth2/OIDC with external providers
- **Endpoints**:
  - `GET /authorize` - Start OAuth2 flow
  - `GET /callback` - Handle provider callback
- **Configuration**: `CredentialConfig.EnableOAuth2`
- **Providers**: Google, Azure AD, GitHub, Facebook, etc.
- **Use Cases**: Social login, enterprise SSO (SAML/OIDC)

### 4. **Passwordless Authentication** 🚧 Skeleton Only
- **Path**: `//api/auth/cred/passwordless/*`
- **Service**: `PasswordlessAuthService`
- **Methods**: Email/SMS magic link or OTP
- **Endpoints**:
  - `POST /send-code` - Send verification code via email/SMS
  - `POST /verify-code` - Verify code and authenticate
  - `POST /send-magic-link` - Send magic link via email
  - `GET /verify-magic-link/{token}` - Verify magic link
- **Configuration**: `CredentialConfig.EnablePasswordless`
- **Use Cases**: Consumer apps, mobile-first apps, reduce password friction

**Status**: ⚠️ DTOs and RouterService created, implementation TODO
**Next Steps**:
- Implement code generation (6-digit random)
- Integrate email provider (SendGrid/AWS SES)
- Integrate SMS provider (Twilio/AWS SNS)
- Cache management for codes/tokens
- Rate limiting

### 5. **Passkey Authentication (WebAuthn)** 🚧 Skeleton Only
- **Path**: `//api/auth/cred/passkey/*`
- **Service**: `PasskeyAuthService`
- **Methods**: FIDO2/WebAuthn (Touch ID, Face ID, security keys)
- **Endpoints**:
  - `POST /register/begin` - Start passkey registration
  - `POST /register/finish` - Complete registration
  - `POST /authenticate/begin` - Start authentication
  - `POST /authenticate/finish` - Complete authentication
  - `GET /credentials/{user_id}` - List user's passkeys
  - `DELETE /credentials/{credential_id}` - Delete passkey
- **Configuration**: `CredentialConfig.EnablePasskey`
- **Use Cases**: High-security apps, passwordless future, phishing-resistant

**Status**: ⚠️ DTOs and RouterService created, implementation TODO
**Next Steps**:
- Integrate WebAuthn library (go-webauthn/webauthn)
- Implement challenge generation and verification
- Public key cryptography verification
- Credential storage and management
- Sign counter tracking (replay attack prevention)

## Configuration Matrix

| Provider | Enable Flag | Config Struct | Dependencies |
|----------|-------------|---------------|--------------|
| Basic | `EnableBasic` | `BasicCredentialConfig` | User store, Password hasher |
| API Key | `EnableAPIKey` | `APIKeyCredentialConfig` | AppKey store, Secret hasher |
| OAuth2 | `EnableOAuth2` | `OAuth2CredentialConfig` | HTTP client, Provider configs |
| Passwordless | `EnablePasswordless` | `PasswordlessCredentialConfig` | Email/SMS provider, Cache |
| Passkey | `EnablePasskey` | `PasskeyCredentialConfig` | WebAuthn library, Key store |

## Implementation Status

### ✅ Fully Implemented (3/5)
1. **Basic Auth**: Login, register, password management
2. **API Key Auth**: Key generation, validation, rotation
3. **OAuth2 Auth**: Multi-provider support (Google, Azure, etc.)

### 🚧 Skeleton Only (2/5)
4. **Passwordless**: DTOs created, service skeleton, TODO: implementation
5. **Passkey**: DTOs created, service skeleton, TODO: implementation

## Example: Enable Multiple Providers

```json
{
  "enable_basic": true,
  "basic_config": {
    "min_password_length": 10,
    "max_login_attempts": 5
  },
  "enable_apikey": true,
  "apikey_config": {
    "default_expiry_days": 365
  },
  "enable_oauth2": true,
  "oauth2_config": {
    "providers": [
      {
        "name": "google",
        "enabled": true,
        "client_id": "...",
        "scopes": ["openid", "email", "profile"]
      },
      {
        "name": "azure",
        "enabled": true,
        "client_id": "...",
        "scopes": ["openid", "email", "profile"]
      }
    ]
  },
  "enable_passwordless": true,
  "passwordless_config": {
    "enable_email": true,
    "enable_sms": false,
    "code_length": 6,
    "code_expiry_secs": 300
  },
  "enable_passkey": true,
  "passkey_config": {
    "rp_name": "My App",
    "rp_id": "example.com",
    "user_verification": "preferred"
  }
}
```

## Security Comparison

| Provider | Security Level | User Experience | Implementation Complexity |
|----------|---------------|-----------------|---------------------------|
| Basic | Medium | Familiar | Low ✅ |
| API Key | Medium-High | N/A (machine-to-machine) | Low ✅ |
| OAuth2 | High | Good (social login) | Medium ✅ |
| Passwordless | Medium-High | Excellent (no password) | Medium 🚧 |
| Passkey | Very High | Excellent (biometric) | High 🚧 |

## Roadmap

### Phase 1: Core (✅ Complete)
- [x] Basic authentication
- [x] API key authentication  
- [x] OAuth2 with multi-provider

### Phase 2: Passwordless (🚧 Next)
- [ ] Email code/magic link implementation
- [ ] SMS code implementation
- [ ] Email provider integration (SendGrid/SES)
- [ ] SMS provider integration (Twilio/SNS)
- [ ] Rate limiting and abuse prevention

### Phase 3: Passkey/WebAuthn (🚧 Future)
- [ ] WebAuthn registration flow
- [ ] WebAuthn authentication flow
- [ ] Credential management
- [ ] Multi-device support
- [ ] Attestation verification

## References

- [Credential Configuration Guide](./credential_configuration.md)
- [OAuth2 Setup Example](./examples/tenant01_basic_oauth2_setup.md)
- [Deployment Guide](./deployment.md)
- [Multi-Tenant Architecture](./multi_tenant_architecture.md)
