# Layer 1: Credential Input / Authentication Examples

This folder contains examples demonstrating different authentication methods (Layer 1).

## Examples

### 1. `01_basic/` - Basic Runtime Usage
Shows how to use the Auth runtime with basic username/password authentication.

**Run:**
```bash
go run ./examples/01_credential/01_basic/main.go
```

**Key Features:**
- Basic authentication with username/password
- Auth runtime setup with builder pattern
- Single authenticator configuration
- Complete flow through all 4 layers

---

### 2. `02_multi_auth/` - Enterprise Multi-Authenticator
Demonstrates enterprise authentication system with multiple credential types.

**Run:**
```bash
go run ./examples/01_credential/02_multi_auth/main.go
```

**Key Features:**
- **Multiple authentication methods:**
  - Basic (username/password)
  - OAuth2 (Google, GitHub, etc.)
  - Passwordless (OTP, Magic Link)
- **Hybrid RBAC + ABAC authorization:**
  - Role-based permissions
  - Attribute-based conditions (ownership, department, verification)
  - ABAC can override RBAC for resource owners
- All managed through single Auth runtime

**Authentication Methods:**

1. **Basic Auth**
   ```go
   basicResponse, _ := auth.Login(ctx, &lokstraauth.LoginRequest{
       Credentials: &basic.BasicCredentials{
           Username: "john.doe",
           Password: "SecurePass123!",
       },
   })
   ```

2. **OAuth2**
   ```go
   googleResponse, _ := auth.Login(ctx, &lokstraauth.LoginRequest{
       Credentials: &OAuth2Credentials{
           Provider: "google",
           Token:    "oauth-token-from-provider",
       },
   })
   ```

3. **Passwordless**
   ```go
   passwordlessResponse, _ := auth.Login(ctx, &lokstraauth.LoginRequest{
       Credentials: &PasswordlessCredentials{
           Email: "user@example.com",
           Token: "magic-link-token",
       },
   })
   ```

**Authorization Scenarios:**

1. **RBAC allows + ABAC allows**: Same department access
2. **RBAC denies + ABAC overrides**: Resource owner can read
3. **RBAC + ABAC conditions**: Verified users can deploy
4. **ABAC blocks**: Cross-department access denied

---

### 3. `03_oauth2/` - OAuth2 Authentication ✅
Demonstrates OAuth2 authentication with multiple providers.

**Run:**
```bash
go run ./examples/01_credential/03_oauth2/main.go
```

**Key Features:**
- Google, GitHub, Facebook OAuth2 integration
- Authorization flow explanation
- Access token authentication
- Provider-specific claims
- Security best practices
- Production implementation guide

**Supported Providers:**
- Google (OpenID Connect)
- GitHub (OAuth Apps)
- Facebook (Facebook Login)

---

### 4. `04_passwordless/` - Passwordless Authentication ✅
Demonstrates passwordless authentication with Magic Link and OTP.

**Run:**
```bash
go run ./examples/01_credential/04_passwordless/main.go
```

**Key Features:**
- Magic Link authentication (15min expiry)
- OTP authentication (6-digit, 5min expiry)
- One-time token enforcement
- Token lifecycle management
- Email-based user resolution
- MockUserResolver and MockTokenSender

**Authentication Flows:**
- Magic Link: InitiateMagicLink → User clicks link → Authenticate
- OTP: InitiateOTP → User enters code → Authenticate

---

### 5. `05_apikey/` - API Key Authentication ✅
Demonstrates API key authentication for service-to-service communication.

**Run:**
```bash
go run ./examples/01_credential/05_apikey/main.go
```

**Key Features:**
- API key generation with SHA3-256 hashing
- Scope-based permissions
- Key expiration support
- Key revocation
- Last used timestamp tracking
- Metadata support
- Prefix-based key identification

**Security Features:**
- One-time key display
- Constant-time comparison
- Automatic expiry checking
- Soft delete with revocation

---

## Extending with Custom Authenticators

To add a new authentication method:

1. **Define Credentials:**
   ```go
   type MyCredentials struct {
       Token string
   }
   
   func (c *MyCredentials) Type() string {
       return "my-auth-type"
   }
   
   func (c *MyCredentials) Validate() error {
       return nil
   }
   ```

2. **Implement Authenticator:**
   ```go
   type MyAuthenticator struct{}
   
   func (a *MyAuthenticator) Authenticate(ctx context.Context, creds credential.Credentials) (*credential.AuthenticationResult, error) {
       // Verify credentials
       return &credential.AuthenticationResult{
           Success: true,
           Subject: "user-id",
           Claims: map[string]interface{}{
               "sub": "user-id",
           },
       }, nil
   }
   
   func (a *MyAuthenticator) Type() string {
       return "my-auth-type"
   }
   ```

3. **Register with Auth Runtime:**
   ```go
   auth := lokstraauth.NewBuilder().
       WithAuthenticator("my-auth-type", myAuth).
       Build()
   ```

---

## Real-World Use Cases

### Multi-Tenant SaaS
Different tenants can use different authentication methods:
- Enterprise tenant: SSO (OAuth2)
- Small business: Username/password
- Individual users: Passwordless (magic link)

### Enterprise Systems
Support multiple authentication channels:
- Internal users: Active Directory / LDAP
- External partners: OAuth2 / SAML
- API clients: API keys / JWT

### B2C Applications
Flexible login options:
- Social login: Google, Facebook, GitHub
- Traditional: Email/password
- Modern: Passwordless, biometric

### API & Service-to-Service
Programmatic access:
- API keys with scopes
- Machine-to-machine OAuth2
- Service accounts

---

## Best Practices

1. **Fail Secure**: Default to deny when authentication fails
2. **Audit Logging**: Log all authentication attempts
3. **Rate Limiting**: Prevent brute force attacks
4. **Token Security**: Use secure token generation and storage
5. **Multi-Factor**: Consider adding 2FA/MFA support
6. **Password Policy**: Enforce strong password requirements (see `basic/validator.go`)
7. **API Key Rotation**: Regularly rotate API keys
8. **OAuth2 Security**: Always use HTTPS for redirect URLs in production

---

## Next Steps

- See `examples/02_token/` for token management examples
- See `examples/03_subject/` for subject resolution examples
- See `examples/04_authz/` for authorization examples
- See `examples/complete/` for full end-to-end flow
- See `01_credential/README.md` for detailed authenticator documentation
