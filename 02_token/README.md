# Layer 02: Token Management

Layer 02 Token Management menyediakan implementasi lengkap untuk token generation, verification, storage, dan lifecycle management dalam sistem autentikasi. Layer ini mendukung berbagai jenis token termasuk JWT (JSON Web Token) dan opaque tokens.

## 📋 Daftar Isi

- [Konsep](#konsep)
- [Arsitektur](#arsitektur)
- [Token Managers](#token-managers)
  - [JWT Manager](#jwt-manager)
  - [Simple Token Manager](#simple-token-manager)
- [Token Store](#token-store)
- [Use Cases](#use-cases)
- [Examples](#examples)
- [Security Best Practices](#security-best-practices)

---

## Konsep

### Apa itu Token?

Token adalah kredensial yang digunakan untuk mengakses resource tanpa perlu mengirim username dan password setiap kali. Token memiliki karakteristik:

- **Stateless**: Tidak memerlukan session storage di server (untuk JWT)
- **Self-contained**: Membawa informasi (claims) di dalamnya
- **Expirable**: Memiliki masa berlaku
- **Revocable**: Dapat dibatalkan sebelum expired

### Jenis Token

**1. JWT (JSON Web Token)**
- Self-contained dengan claims
- Signed untuk validasi
- Dapat diverifikasi tanpa database lookup
- Support untuk refresh tokens

**2. Opaque Token**
- Random string tanpa informasi
- Memerlukan database lookup untuk verifikasi
- Lebih aman untuk sensitive data
- Mudah untuk revoke

---

## Arsitektur

```
┌─────────────────────────────────────────────────────┐
│              Token Management Layer                  │
├─────────────────────────────────────────────────────┤
│                                                       │
│  ┌──────────────┐      ┌───────────────────────┐   │
│  │ JWT Manager  │      │ Simple Token Manager  │   │
│  └──────────────┘      └───────────────────────┘   │
│         │                         │                  │
│         │                         │                  │
│         └──────────┬──────────────┘                 │
│                    │                                 │
│              ┌─────▼──────┐                         │
│              │ TokenStore │                         │
│              └────────────┘                         │
│                                                       │
├─────────────────────────────────────────────────────┤
│              Contract Interfaces                     │
│  • TokenManager                                      │
│  • TokenStore                                        │
│  • RevocationList                                    │
└─────────────────────────────────────────────────────┘
```

### Core Interfaces

#### TokenManager
```go
type TokenManager interface {
    Generate(ctx context.Context, subject string, metadata map[string]any) (*Token, error)
    Verify(ctx context.Context, tokenValue string) (*Token, error)
}
```

#### TokenStore
```go
type TokenStore interface {
    Store(ctx context.Context, token *Token) error
    Get(ctx context.Context, subject, tokenID string) (*Token, error)
    Delete(ctx context.Context, subject, tokenID string) error
    List(ctx context.Context, subject string) ([]*Token, error)
    Revoke(ctx context.Context, tokenID string) error
    IsRevoked(ctx context.Context, tokenID string) (bool, error)
    Cleanup(ctx context.Context) error
}
```

#### RevocationList
```go
type RevocationList interface {
    Revoke(ctx context.Context, tokenID string) error
    IsRevoked(ctx context.Context, tokenID string) (bool, error)
}
```

---

## Token Managers

### JWT Manager

JWT Manager mengimplementasikan token management menggunakan JSON Web Tokens (RFC 7519).

#### Features

- ✅ **HS256 Signing**: HMAC SHA-256 algorithm
- ✅ **Access Tokens**: Short-lived tokens (15 minutes default)
- ✅ **Refresh Tokens**: Long-lived tokens (7 days default)
- ✅ **Token Revocation**: Revocation list support
- ✅ **Claims Helpers**: Easy access to token claims
- ✅ **Validation**: Issuer, audience, expiry checks

#### Configuration

```go
type JWTConfig struct {
    SigningKey        []byte        // Secret key untuk signing
    SigningMethod     string        // "HS256", "HS384", "HS512"
    AccessDuration    time.Duration // Durasi access token
    RefreshDuration   time.Duration // Durasi refresh token
    Issuer           string        // Token issuer
    Audience         []string      // Token audience
    EnableRevocation bool          // Enable revocation list
}
```

#### Basic Usage

```go
// Create JWT manager
config := &jwt.JWTConfig{
    SigningKey:        []byte("your-secret-key"),
    SigningMethod:     "HS256",
    AccessDuration:    15 * time.Minute,
    RefreshDuration:   7 * 24 * time.Hour,
    Issuer:           "my-app",
    Audience:         []string{"my-app-users"},
    EnableRevocation: true,
}

manager := jwt.NewJWTManager(config)

// Generate access token
metadata := map[string]any{
    "email": "user@example.com",
    "name":  "John Doe",
    "role":  "admin",
}

token, err := manager.Generate(ctx, "user123", metadata)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Access Token:", token.Value)
fmt.Println("Expires At:", token.ExpiresAt)
```

#### Verify Token

```go
// Verify and extract claims
verifiedToken, err := manager.Verify(ctx, tokenValue)
if err != nil {
    log.Fatal(err)
}

// Access claims
subject := verifiedToken.Subject
email := verifiedToken.Metadata["email"]
role := verifiedToken.Metadata["role"]
```

#### Refresh Token

```go
// Generate refresh token
refreshToken, err := manager.GenerateRefreshToken(ctx, "user123", metadata)
if err != nil {
    log.Fatal(err)
}

// Later, refresh access token
newAccessToken, err := manager.Refresh(ctx, refreshToken.Value)
if err != nil {
    log.Fatal(err)
}
```

#### Revoke Token

```go
// Revoke a token
err := manager.Revoke(ctx, tokenValue)
if err != nil {
    log.Fatal(err)
}

// Verify will fail for revoked token
_, err = manager.Verify(ctx, tokenValue)
// err: "token has been revoked"
```

#### Claims Helpers

```go
// Helper methods untuk extract claims
email := token.GetStringClaim("email")
age := token.GetInt64Claim("age")
isAdmin := token.GetBoolClaim("is_admin")
roles := token.GetStringSliceClaim("roles")
```

### Simple Token Manager

Simple Token Manager menggunakan cryptographically secure random tokens (opaque tokens).

#### Features

- ✅ **Secure Random**: Crypto-grade random generation
- ✅ **Opaque Tokens**: No information leakage
- ✅ **Configurable Length**: 16-64 bytes
- ✅ **Revocation Support**: Built-in revocation list
- ✅ **Auto Cleanup**: Expired tokens cleanup
- ✅ **Base64 Encoding**: URL-safe encoding

#### Configuration

```go
type Config struct {
    TokenLength      int           // Token length in bytes (default: 32)
    TokenDuration    time.Duration // Token lifetime (default: 1 hour)
    EnableRevocation bool          // Enable revocation list
}
```

#### Basic Usage

```go
// Create simple token manager
config := &simple.Config{
    TokenLength:      32,
    TokenDuration:    1 * time.Hour,
    EnableRevocation: true,
}

manager := simple.NewManager(config)

// Generate token
metadata := map[string]any{
    "device": "mobile",
    "ip":     "192.168.1.100",
}

token, err := manager.Generate(ctx, "user456", metadata)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Token:", token.Value)
// Output: uOYAvG_3Ri75FiEJz2jyLO9k_dbDVFr6NL6IfG-7Z-0=
```

#### Multiple Tokens per User

```go
// Generate multiple tokens for different devices
desktopToken, _ := manager.Generate(ctx, "user456", map[string]any{
    "device": "desktop",
})

mobileToken, _ := manager.Generate(ctx, "user456", map[string]any{
    "device": "mobile",
})

webToken, _ := manager.Generate(ctx, "user456", map[string]any{
    "device": "web",
})
```

#### Token Verification

```go
// Verify token
verifiedToken, err := manager.Verify(ctx, tokenValue)
if err != nil {
    log.Fatal(err)
}

// Access metadata
device := verifiedToken.Metadata["device"]
```

#### Revocation

```go
// Revoke token
err := manager.Revoke(ctx, tokenValue)

// Verify will fail
_, err = manager.Verify(ctx, tokenValue)
// err: "token has been revoked"
```

---

## Token Store

Token Store menyediakan persistent storage untuk token lifecycle management.

### Features

- ✅ **Multi-User Support**: Store tokens for multiple users
- ✅ **Multi-Device**: Multiple tokens per user
- ✅ **Device Tracking**: Track device information
- ✅ **Revocation Management**: Revoke individual tokens
- ✅ **Expiry Cleanup**: Automatic cleanup of expired tokens
- ✅ **CRUD Operations**: Full create, read, update, delete

### Implementation

```go
type InMemoryTokenStore struct {
    mu              sync.RWMutex
    tokens          map[string]map[string]*Token // subject -> tokenID -> Token
    revokedTokens   map[string]bool              // tokenID -> revoked
}
```

### Basic Usage

#### Store Token

```go
store := NewInMemoryTokenStore()

// Store token
token := &Token{
    Value:     "token-value",
    Type:      "Bearer",
    Subject:   "user1",
    IssuedAt:  time.Now(),
    ExpiresAt: time.Now().Add(1 * time.Hour),
    Metadata: map[string]any{
        "token_id": "user1-desktop",
        "device":   "desktop",
    },
}

err := store.Store(ctx, token)
```

#### Retrieve Token

```go
// Get specific token
token, err := store.Get(ctx, "user1", "user1-desktop")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Device:", token.Metadata["device"])
fmt.Println("Expires:", token.ExpiresAt)
```

#### List User Tokens

```go
// List all tokens for a user
tokens, err := store.List(ctx, "user1")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d tokens\n", len(tokens))
for _, tk := range tokens {
    device := tk.Metadata["device"]
    tokenID := tk.Metadata["token_id"]
    fmt.Printf("- %s (%s)\n", device, tokenID)
}
```

#### Revoke Token

```go
// Revoke token
err := store.Revoke(ctx, "user1-mobile")
if err != nil {
    log.Fatal(err)
}

// Check revocation status
isRevoked, _ := store.IsRevoked(ctx, "user1-mobile")
fmt.Println("Revoked:", isRevoked) // true
```

#### Delete Token

```go
// Delete token permanently
err := store.Delete(ctx, "user1", "user1-tablet")
if err != nil {
    log.Fatal(err)
}
```

#### Cleanup Expired Tokens

```go
// Manual cleanup
err := store.Cleanup(ctx)
if err != nil {
    log.Fatal(err)
}

// Or setup automatic cleanup
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        if err := store.Cleanup(ctx); err != nil {
            log.Printf("Cleanup error: %v", err)
        }
    }
}()
```

---

## Use Cases

### 1. API Authentication

```go
// Generate API token
apiToken, err := simpleManager.Generate(ctx, "api-client-123", map[string]any{
    "client_name": "Mobile App",
    "scopes":      []string{"read", "write"},
})

// Store in token store
store.Store(ctx, apiToken)

// Verify on API request
verified, err := simpleManager.Verify(ctx, apiToken.Value)
if err != nil {
    return http.StatusUnauthorized
}
```

### 2. Session Management

```go
// Login - generate session token
sessionToken, err := jwtManager.Generate(ctx, userID, map[string]any{
    "email":      user.Email,
    "session_id": uuid.New().String(),
})

// Store session
store.Store(ctx, sessionToken)

// Logout - revoke session
store.Revoke(ctx, sessionID)
```

### 3. Multi-Device Access

```go
// User login from desktop
desktopToken, _ := manager.Generate(ctx, userID, map[string]any{
    "device":    "desktop",
    "device_id": "device-uuid-1",
})

// User login from mobile
mobileToken, _ := manager.Generate(ctx, userID, map[string]any{
    "device":    "mobile",
    "device_id": "device-uuid-2",
})

// List all user sessions
tokens, _ := store.List(ctx, userID)
// Returns: desktop + mobile tokens
```

### 4. Refresh Token Flow

```go
// Login - generate access + refresh
accessToken, _ := jwtManager.Generate(ctx, userID, metadata)
refreshToken, _ := jwtManager.GenerateRefreshToken(ctx, userID, metadata)

// Access token expired, use refresh token
newAccessToken, err := jwtManager.Refresh(ctx, refreshToken.Value)
if err != nil {
    // Refresh token invalid/expired - require re-login
    return errors.New("please login again")
}

// Continue with new access token
```

### 5. Token Revocation (Forced Logout)

```go
// Admin forces user logout
tokens, _ := store.List(ctx, userID)
for _, token := range tokens {
    tokenID := token.Metadata["token_id"].(string)
    store.Revoke(ctx, tokenID)
}

// User's next request will fail
_, err := manager.Verify(ctx, userToken)
// err: "token has been revoked"
```

### 6. Device Management

```go
// User views their devices
tokens, _ := store.List(ctx, userID)
for _, token := range tokens {
    device := token.Metadata["device"]
    lastSeen := token.IssuedAt
    
    isRevoked, _ := store.IsRevoked(ctx, token.Metadata["token_id"].(string))
    status := "Active"
    if isRevoked {
        status = "Revoked"
    }
    
    fmt.Printf("Device: %s, Last Seen: %s, Status: %s\n", 
        device, lastSeen, status)
}

// User revokes specific device
store.Revoke(ctx, "user1-mobile")
```

---

## Examples

Lihat folder `examples/02_token/` untuk contoh lengkap:

### 01_jwt - JWT Token Management
Demonstrasi lengkap JWT token features:
- Generate access tokens
- Generate refresh tokens
- Verify tokens
- Refresh mechanism
- Token revocation
- Claims helpers
- Expiry validation

```bash
cd examples/02_token/01_jwt
go run main.go
```

### 02_simple - Simple Opaque Tokens
Demonstrasi opaque token management:
- Generate secure random tokens
- Multiple tokens per user
- Token verification
- Revocation
- Different token lengths

```bash
cd examples/02_token/02_simple
go run main.go
```

### 03_store - Token Store Management
Demonstrasi token lifecycle management:
- Store tokens for multiple users
- Multi-device support
- List user tokens
- Revoke tokens
- Cleanup expired tokens

```bash
cd examples/02_token/03_store
go run main.go
```

---

## Security Best Practices

### 1. Secret Key Management

❌ **Bad**:
```go
config := &jwt.JWTConfig{
    SigningKey: []byte("mysecret"),  // Weak secret
}
```

✅ **Good**:
```go
// Load from environment
signingKey := []byte(os.Getenv("JWT_SIGNING_KEY"))

// Or use crypto random
signingKey := make([]byte, 32)
if _, err := rand.Read(signingKey); err != nil {
    log.Fatal(err)
}

config := &jwt.JWTConfig{
    SigningKey: signingKey,
}
```

### 2. Token Duration

❌ **Bad**:
```go
AccessDuration: 24 * time.Hour,  // Too long
RefreshDuration: 365 * 24 * time.Hour,  // Way too long
```

✅ **Good**:
```go
AccessDuration: 15 * time.Minute,  // Short-lived
RefreshDuration: 7 * 24 * time.Hour,  // Reasonable
```

### 3. Token Storage

❌ **Bad**:
```go
// Store in localStorage (vulnerable to XSS)
localStorage.setItem('token', token)
```

✅ **Good**:
```go
// Use httpOnly cookies
http.SetCookie(w, &http.Cookie{
    Name:     "access_token",
    Value:    token.Value,
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteStrictMode,
})
```

### 4. Revocation

❌ **Bad**:
```go
// No revocation support
manager := jwt.NewJWTManager(&jwt.JWTConfig{
    EnableRevocation: false,
})
```

✅ **Good**:
```go
// Enable revocation
manager := jwt.NewJWTManager(&jwt.JWTConfig{
    EnableRevocation: true,
})

// Implement logout
func Logout(ctx context.Context, token string) error {
    return manager.Revoke(ctx, token)
}
```

### 5. Sensitive Data

❌ **Bad**:
```go
// Don't put sensitive data in JWT
metadata := map[string]any{
    "password": "secret123",
    "ssn": "123-45-6789",
}
```

✅ **Good**:
```go
// Use opaque tokens for sensitive data
metadata := map[string]any{
    "user_id": "user123",
    "role": "admin",
}

// Store sensitive data in database, lookup by token
```

### 6. Token Validation

❌ **Bad**:
```go
// Skip validation
token, _ := manager.Verify(ctx, tokenValue)
// Use token without checking error
```

✅ **Good**:
```go
// Always validate
token, err := manager.Verify(ctx, tokenValue)
if err != nil {
    return nil, fmt.Errorf("invalid token: %w", err)
}

// Check expiry
if time.Now().After(token.ExpiresAt) {
    return nil, errors.New("token expired")
}

// Check revocation
isRevoked, _ := store.IsRevoked(ctx, tokenID)
if isRevoked {
    return nil, errors.New("token revoked")
}
```

### 7. HTTPS Only

❌ **Bad**:
```go
// Send tokens over HTTP
http.Get("http://api.example.com/data?token=" + token)
```

✅ **Good**:
```go
// Always use HTTPS
req, _ := http.NewRequest("GET", "https://api.example.com/data", nil)
req.Header.Set("Authorization", "Bearer " + token)
```

### 8. Cleanup

❌ **Bad**:
```go
// Never cleanup expired tokens
// Memory leak over time
```

✅ **Good**:
```go
// Regular cleanup
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        store.Cleanup(ctx)
    }
}()
```

---

## Migration Guide

### From Session-Based to Token-Based

**Before (Session)**:
```go
// Login
sessionID := generateSessionID()
sessions[sessionID] = UserSession{
    UserID: userID,
    Data:   userData,
}

// Verify
session := sessions[sessionID]
if session == nil {
    return errors.New("invalid session")
}
```

**After (Token)**:
```go
// Login
token, err := manager.Generate(ctx, userID, metadata)

// Verify
verifiedToken, err := manager.Verify(ctx, tokenValue)
```

**Benefits**:
- No server-side session storage
- Stateless - easier to scale
- Works across multiple servers
- Mobile-friendly

---

## Performance Considerations

### JWT vs Simple Tokens

**JWT**:
- ✅ No database lookup for verification
- ✅ Self-contained claims
- ❌ Larger token size
- ❌ Can't revoke without revocation list

**Simple Tokens**:
- ✅ Smaller token size
- ✅ Easy to revoke
- ❌ Requires database lookup
- ❌ Metadata stored separately

### Recommendations

- Use **JWT** for:
  - Microservices communication
  - Stateless APIs
  - Mobile apps
  
- Use **Simple Tokens** for:
  - Web sessions
  - API keys
  - Sensitive applications

---

## Testing

### Unit Tests

```go
func TestTokenGeneration(t *testing.T) {
    manager := jwt.NewJWTManager(testConfig)
    
    token, err := manager.Generate(ctx, "user123", metadata)
    assert.NoError(t, err)
    assert.NotEmpty(t, token.Value)
    assert.Equal(t, "Bearer", token.Type)
}

func TestTokenVerification(t *testing.T) {
    manager := jwt.NewJWTManager(testConfig)
    
    token, _ := manager.Generate(ctx, "user123", metadata)
    verified, err := manager.Verify(ctx, token.Value)
    
    assert.NoError(t, err)
    assert.Equal(t, "user123", verified.Subject)
}

func TestTokenRevocation(t *testing.T) {
    manager := jwt.NewJWTManager(testConfig)
    
    token, _ := manager.Generate(ctx, "user123", metadata)
    err := manager.Revoke(ctx, token.Value)
    assert.NoError(t, err)
    
    _, err = manager.Verify(ctx, token.Value)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "revoked")
}
```

---

## Troubleshooting

### Common Issues

**1. "Token has expired"**
- Check `AccessDuration` configuration
- Implement refresh token flow
- Use appropriate token lifetimes

**2. "Token signature is invalid"**
- Verify `SigningKey` is correct
- Check `SigningMethod` matches
- Ensure key hasn't changed

**3. "Token has been revoked"**
- Check revocation list
- User may have logged out
- Admin may have forced logout

**4. "Invalid token"**
- Token may be malformed
- Check token format
- Verify token wasn't modified

---

## References

- [JWT RFC 7519](https://tools.ietf.org/html/rfc7519)
- [OAuth 2.0 Token Usage](https://tools.ietf.org/html/rfc6750)
- [OWASP Token Best Practices](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)

---

## License

MIT License - see [LICENSE](../LICENSE) file for details.
