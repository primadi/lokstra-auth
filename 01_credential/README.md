# Layer 01: Credential - Authentication Layer

Layer pertama dalam arsitektur autentikasi Lokstra yang menangani berbagai metode autentikasi pengguna.

## 📋 Daftar Isi

- [Arsitektur](#arsitektur)
- [Interface Contract](#interface-contract)
- [Implementasi](#implementasi)
  - [Basic Authenticator](#basic-authenticator)
  - [OAuth2 Authenticator](#oauth2-authenticator)
  - [Passwordless Authenticator](#passwordless-authenticator)
  - [API Key Authenticator](#api-key-authenticator)
- [Penggunaan](#penggunaan)
- [Extensibility](#extensibility)

## 🏗️ Arsitektur

Layer Credential mengikuti pattern contract-first dengan interface yang jelas:

```
01_credential/
├── contract.go              # Core interfaces
├── basic/
│   └── authenticator.go     # Password-based authentication
├── oauth2/
│   └── authenticator.go     # OAuth2 (Google, GitHub, Facebook)
├── passwordless/
│   └── authenticator.go     # Magic Link & OTP
└── apikey/
    └── authenticator.go     # API Key authentication
```

## 🔌 Interface Contract

### Credentials Interface

```go
type Credentials interface {
    Type() string     // Jenis kredensial
    Validate() error  // Validasi kredensial
}
```

### Authenticator Interface

```go
type Authenticator interface {
    Authenticate(ctx context.Context, creds Credentials) (*AuthenticationResult, error)
    Type() string
}
```

### AuthenticationResult

```go
type AuthenticationResult struct {
    Success bool
    Subject string                 // User ID
    Claims  map[string]interface{} // Metadata tambahan
    Error   error
}
```

## 🔐 Implementasi

### Basic Authenticator

Autentikasi berbasis username/email dan password dengan bcrypt hashing.

**Features:**
- Bcrypt password hashing (cost factor 10)
- Username atau email login
- User store interface untuk extensibility
- In-memory implementation untuk testing

**Penggunaan:**

```go
import "github.com/primadi/lokstra-auth/01_credential/basic"

// Setup authenticator
userStore := basic.NewInMemoryUserStore()
auth := basic.NewAuthenticator(&basic.Config{
    UserStore: userStore,
})

// Register user
hashedPassword, _ := basic.HashPassword("mypassword")
userStore.AddUser(&basic.User{
    ID:       "user123",
    Username: "john",
    Email:    "john@example.com",
    Password: hashedPassword,
})

// Authenticate
creds := &basic.Credentials{
    Username: "john",
    Password: "mypassword",
}

result, err := auth.Authenticate(ctx, creds)
if result.Success {
    fmt.Println("User ID:", result.Subject)
}
```

**UserStore Interface:**

```go
type UserStore interface {
    GetByUsername(ctx context.Context, username string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    GetByID(ctx context.Context, id string) (*User, error)
}
```

Implementasi custom dapat menggunakan database (PostgreSQL, MongoDB, dll).

---

### OAuth2 Authenticator

Autentikasi menggunakan OAuth2 providers (Google, GitHub, Facebook).

**Supported Providers:**
- **Google** - OpenID Connect
- **GitHub** - GitHub OAuth Apps
- **Facebook** - Facebook Login

**Features:**
- Authorization code flow
- Token validation dengan provider
- User info fetching
- Email verification
- Configurable client credentials per provider

**Penggunaan:**

```go
import "github.com/primadi/lokstra-auth/01_credential/oauth2"

// Setup authenticator dengan multiple providers
auth := oauth2.NewAuthenticator(&oauth2.Config{
    Providers: map[string]*oauth2.ProviderConfig{
        "google": {
            ClientID:     "your-google-client-id",
            ClientSecret: "your-google-secret",
            RedirectURL:  "https://yourapp.com/auth/google/callback",
        },
        "github": {
            ClientID:     "your-github-client-id",
            ClientSecret: "your-github-secret",
            RedirectURL:  "https://yourapp.com/auth/github/callback",
        },
    },
})

// Authenticate dengan authorization code
creds := &oauth2.Credentials{
    Provider:          "google",
    AuthorizationCode: "code-from-oauth-callback",
}

result, err := auth.Authenticate(ctx, creds)
if result.Success {
    claims := result.Claims
    email := claims["email"].(string)
    name := claims["name"].(string)
    picture := claims["picture"].(string)
}
```

**Provider Configuration:**

```go
type ProviderConfig struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
}
```

**Claims dari Providers:**

Google:
- `email` - Email pengguna
- `name` - Nama lengkap
- `picture` - URL foto profil
- `email_verified` - Status verifikasi email

GitHub:
- `email` - Email primary (verified)
- `login` - Username GitHub
- `name` - Nama lengkap
- `avatar_url` - URL avatar

Facebook:
- `email` - Email pengguna
- `name` - Nama lengkap
- `picture` - URL foto profil

**Implementation Details:**

Authenticator menggunakan HTTP client untuk berkomunikasi dengan provider APIs:
- Google: `https://www.googleapis.com/oauth2/v2/userinfo`
- GitHub: `https://api.github.com/user` + `/user/emails`
- Facebook: `https://graph.facebook.com/me`

---

### Passwordless Authenticator

Autentikasi tanpa password menggunakan Magic Link atau OTP.

**Authentication Methods:**
1. **Magic Link** - URL token dikirim via email (expiry 15 menit)
2. **OTP** - 6-digit code dikirim via email (expiry 5 menit)

**Features:**
- Token lifecycle management (generation, validation, expiry)
- One-time use enforcement
- Extensible token storage (in-memory, Redis, database)
- Customizable token generator
- Email sender interface
- User resolver interface

**Penggunaan Magic Link:**

```go
import "github.com/primadi/lokstra-auth/01_credential/passwordless"

// Setup authenticator
tokenStore := passwordless.NewInMemoryTokenStore()
auth := passwordless.NewAuthenticator(&passwordless.Config{
    TokenStore: tokenStore,
    UserResolver: &MyUserResolver{},
    TokenSender:  &MyEmailSender{},
})

// Request magic link
creds := &passwordless.Credentials{
    Email:      "user@example.com",
    TokenType:  passwordless.TokenTypeMagicLink,
}

result, err := auth.Authenticate(ctx, creds)
// Token akan dikirim via email menggunakan TokenSender

// Verify magic link
creds = &passwordless.Credentials{
    Email:      "user@example.com",
    Token:      "token-from-email",
    TokenType:  passwordless.TokenTypeMagicLink,
}

result, err = auth.Authenticate(ctx, creds)
if result.Success {
    fmt.Println("User ID:", result.Subject)
}
```

**Penggunaan OTP:**

```go
// Request OTP
creds := &passwordless.Credentials{
    Email:      "user@example.com",
    TokenType:  passwordless.TokenTypeOTP,
}

result, err := auth.Authenticate(ctx, creds)
// OTP 6-digit akan dikirim via email

// Verify OTP
creds = &passwordless.Credentials{
    Email:      "user@example.com",
    Token:      "123456",
    TokenType:  passwordless.TokenTypeOTP,
}

result, err = auth.Authenticate(ctx, creds)
```

**Core Interfaces:**

```go
// TokenStore manages token lifecycle
type TokenStore interface {
    Store(ctx context.Context, email, token string, expiresAt time.Time, tokenType TokenType) error
    Get(ctx context.Context, email, token string, tokenType TokenType) (*StoredToken, error)
    MarkUsed(ctx context.Context, email, token string, tokenType TokenType) error
    Delete(ctx context.Context, email, token string, tokenType TokenType) error
    Cleanup(ctx context.Context) error
}

// UserResolver resolves user ID from email
type UserResolver interface {
    ResolveUser(ctx context.Context, email string) (userID string, err error)
}

// TokenGenerator generates random tokens
type TokenGenerator interface {
    GenerateMagicLinkToken() (string, error)
    GenerateOTP() (string, error)
}

// TokenSender sends tokens via email
type TokenSender interface {
    SendMagicLink(ctx context.Context, email, token string) error
    SendOTP(ctx context.Context, email, otp string) error
}
```

**Default Implementations:**

- `InMemoryTokenStore` - Mutex-protected in-memory map
- `DefaultTokenGenerator` - 32-byte random magic link, 6-digit OTP
- User harus menyediakan: `UserResolver`, `TokenSender`

**Token Cleanup:**

InMemoryTokenStore memiliki background cleanup setiap 1 jam:

```go
// Cleanup expired tokens manually
tokenStore.Cleanup(ctx)

// Or use background cleanup (default)
tokenStore := passwordless.NewInMemoryTokenStore()
// Cleanup runs every 1 hour automatically
```

---

### API Key Authenticator

Autentikasi berbasis API key untuk service-to-service atau programmatic access.

**Features:**
- SHA3-256 key hashing
- Key expiration support
- Scope-based permissions
- Key revocation
- Usage tracking (last used timestamp)
- Metadata support
- Key prefix untuk identification
- In-memory dan custom key store

**Penggunaan:**

```go
import "github.com/primadi/lokstra-auth/01_credential/apikey"

// Setup authenticator
keyStore := apikey.NewInMemoryKeyStore()
auth := apikey.NewAuthenticator(&apikey.Config{
    KeyStore: keyStore,
})

// Generate API key
expiresIn := 30 * 24 * time.Hour // 30 days
keyString, apiKey, err := auth.GenerateKey(
    ctx,
    "user123",                    // User ID
    "Production API Key",         // Key name
    []string{"read", "write"},    // Scopes
    &expiresIn,
)

// keyString: "dGVzdC1rZXktZXhhbXBsZS0xMjM0NTY3ODk"
// Save this to give to user (only shown once!)

fmt.Println("API Key:", keyString)
fmt.Println("Key ID:", apiKey.ID)
fmt.Println("Prefix:", apiKey.Prefix)

// Authenticate with API key
creds := &apikey.Credentials{
    APIKey: keyString,
}

result, err := auth.Authenticate(ctx, creds)
if result.Success {
    fmt.Println("User ID:", result.Subject)
    fmt.Println("Scopes:", result.Claims["scopes"])
    fmt.Println("Key Name:", result.Claims["key_name"])
}
```

**Key Management:**

```go
// Revoke API key
err := auth.RevokeKey(ctx, apiKey.ID)

// Key store operations
keyStore.GetByHash(ctx, hash)
keyStore.GetByPrefix(ctx, "dGVzdC1r") // First 8 chars
keyStore.UpdateLastUsed(ctx, keyID, time.Now())
keyStore.Delete(ctx, keyID)
```

**APIKey Structure:**

```go
type APIKey struct {
    ID        string
    KeyHash   string                 // SHA3-256 hash
    Prefix    string                 // For identification
    UserID    string                 // Owner
    Name      string                 // Descriptive name
    Scopes    []string               // Permissions
    Metadata  map[string]interface{} // Custom metadata
    CreatedAt time.Time
    ExpiresAt *time.Time             // nil = never expires
    LastUsed  *time.Time
    Revoked   bool
    RevokedAt *time.Time
}
```

**KeyStore Interface:**

```go
type KeyStore interface {
    GetByHash(ctx context.Context, hash string) (*APIKey, error)
    GetByPrefix(ctx context.Context, prefix string) ([]*APIKey, error)
    Store(ctx context.Context, key *APIKey) error
    UpdateLastUsed(ctx context.Context, keyID string, timestamp time.Time) error
    Revoke(ctx context.Context, keyID string) error
    Delete(ctx context.Context, keyID string) error
}
```

**Security Features:**

1. **Key Hashing**: API keys di-hash menggunakan SHA3-256 sebelum disimpan
2. **One-time Display**: Key string hanya ditampilkan saat generation
3. **Constant Time Comparison**: Mencegah timing attacks
4. **Automatic Expiry Check**: Validasi expiry saat authentication
5. **Revocation Support**: Soft delete dengan timestamp

**Best Practices:**

```go
// 1. Set expiration untuk production keys
thirtyDays := 30 * 24 * time.Hour
keyString, _, err := auth.GenerateKey(ctx, userID, "Prod", scopes, &thirtyDays)

// 2. Use scopes untuk limit permissions
scopes := []string{"read:users", "write:posts"}

// 3. Add metadata untuk tracking
apiKey.Metadata["app_name"] = "Mobile App"
apiKey.Metadata["environment"] = "production"

// 4. Rotate keys periodically
// Revoke old key
auth.RevokeKey(ctx, oldKeyID)
// Generate new key
newKey, _, _ := auth.GenerateKey(ctx, userID, "Rotated Key", scopes, &expiresIn)

// 5. Monitor last used untuk detect unused keys
if apiKey.LastUsed != nil && time.Since(*apiKey.LastUsed) > 90*24*time.Hour {
    // Key not used in 90 days - consider revoking
}
```

---

## 🚀 Penggunaan

### Single Authenticator

```go
import (
    credential "github.com/primadi/lokstra-auth/01_credential"
    "github.com/primadi/lokstra-auth/01_credential/basic"
)

func main() {
    // Create authenticator
    auth := basic.NewAuthenticator(nil)
    
    // Authenticate
    creds := &basic.Credentials{
        Username: "john",
        Password: "secret",
    }
    
    result, err := auth.Authenticate(ctx, creds)
    if err != nil {
        log.Fatal(err)
    }
    
    if result.Success {
        fmt.Println("Welcome", result.Subject)
    }
}
```

### Multi-Authenticator

Gunakan `MultiAuthenticator` untuk mendukung berbagai metode autentikasi:

```go
import (
    credential "github.com/primadi/lokstra-auth/01_credential"
    "github.com/primadi/lokstra-auth/01_credential/basic"
    "github.com/primadi/lokstra-auth/01_credential/oauth2"
    "github.com/primadi/lokstra-auth/01_credential/apikey"
)

func main() {
    // Create multi-authenticator
    multi := credential.NewMultiAuthenticator()
    
    // Add authenticators
    multi.AddAuthenticator(basic.NewAuthenticator(nil))
    multi.AddAuthenticator(oauth2.NewAuthenticator(&oauth2.Config{
        Providers: map[string]*oauth2.ProviderConfig{
            "google": {
                ClientID:     "...",
                ClientSecret: "...",
                RedirectURL:  "...",
            },
        },
    }))
    multi.AddAuthenticator(apikey.NewAuthenticator(nil))
    
    // Authenticate - will try appropriate authenticator based on credentials type
    result, err := multi.Authenticate(ctx, creds)
}
```

## 🔧 Extensibility

### Custom Authenticator

Buat authenticator custom dengan mengimplementasi interface:

```go
type MyCustomAuthenticator struct{}

func (a *MyCustomAuthenticator) Authenticate(ctx context.Context, creds credential.Credentials) (*credential.AuthenticationResult, error) {
    // Your custom authentication logic
    return &credential.AuthenticationResult{
        Success: true,
        Subject: "user123",
        Claims: map[string]interface{}{
            "method": "custom",
        },
    }, nil
}

func (a *MyCustomAuthenticator) Type() string {
    return "custom"
}
```

### Custom User Store

Implementasi custom untuk database:

```go
type PostgresUserStore struct {
    db *sql.DB
}

func (s *PostgresUserStore) GetByUsername(ctx context.Context, username string) (*basic.User, error) {
    var user basic.User
    err := s.db.QueryRowContext(ctx,
        "SELECT id, username, email, password FROM users WHERE username = $1",
        username,
    ).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
    
    if err == sql.ErrNoRows {
        return nil, basic.ErrUserNotFound
    }
    return &user, err
}
```

### Custom Token Store

Implementasi Redis untuk passwordless:

```go
type RedisTokenStore struct {
    client *redis.Client
}

func (s *RedisTokenStore) Store(ctx context.Context, email, token string, expiresAt time.Time, tokenType passwordless.TokenType) error {
    key := fmt.Sprintf("token:%s:%s:%s", tokenType, email, token)
    ttl := time.Until(expiresAt)
    
    data := map[string]interface{}{
        "token":      token,
        "email":      email,
        "type":       tokenType,
        "expires_at": expiresAt.Unix(),
        "used":       false,
    }
    
    return s.client.HMSet(ctx, key, data).Err()
}
```

### Custom Key Store

Implementasi database untuk API keys:

```go
type DatabaseKeyStore struct {
    db *sql.DB
}

func (s *DatabaseKeyStore) GetByHash(ctx context.Context, hash string) (*apikey.APIKey, error) {
    var key apikey.APIKey
    // Query database
    err := s.db.QueryRowContext(ctx,
        "SELECT id, key_hash, prefix, user_id, name, scopes, created_at, expires_at, revoked FROM api_keys WHERE key_hash = $1",
        hash,
    ).Scan(&key.ID, &key.KeyHash, &key.Prefix, &key.UserID, &key.Name, &key.Scopes, &key.CreatedAt, &key.ExpiresAt, &key.Revoked)
    
    return &key, err
}
```

## 📦 Dependencies

```
github.com/golang-jwt/jwt/v5
golang.org/x/crypto/bcrypt
golang.org/x/crypto/sha3
```

## ✅ Testing

Setiap authenticator memiliki in-memory implementation untuk testing:

- `basic.InMemoryUserStore`
- `passwordless.InMemoryTokenStore`
- `apikey.InMemoryKeyStore`

Gunakan untuk unit testing tanpa database:

```go
func TestAuthentication(t *testing.T) {
    userStore := basic.NewInMemoryUserStore()
    auth := basic.NewAuthenticator(&basic.Config{
        UserStore: userStore,
    })
    
    // Add test user
    userStore.AddUser(&basic.User{
        ID:       "test",
        Username: "testuser",
        Password: "$2a$10$...",
    })
    
    // Test authentication
    result, err := auth.Authenticate(ctx, creds)
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

## 🔐 Security Considerations

1. **Password Hashing**: Gunakan bcrypt dengan cost factor minimal 10
2. **Token Expiry**: Set expiry yang reasonable untuk setiap token type
3. **One-time Tokens**: Enforce one-time use untuk magic links dan OTP
4. **API Key Storage**: Jangan simpan plain API keys, selalu hash
5. **HTTPS Only**: OAuth2 redirect URLs harus HTTPS di production
6. **Rate Limiting**: Implementasi rate limiting untuk prevent brute force
7. **Audit Logging**: Log semua authentication attempts

## 📚 Resources

- [OAuth2 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [bcrypt Paper](https://www.usenix.org/legacy/events/usenix99/provos/provos.pdf)

---

**Layer 01 Complete** ✅  
Lanjut ke [Layer 02: Token](../02_token/README.md) untuk JWT token management.
