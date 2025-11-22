 # Layer 01: Credential - Authentication Layer

The first layer in Lokstra's authentication architecture that handles various user authentication methods.

## 📋 Table of Contents

- [Architecture](#architecture)
- [Interface Contract](#interface-contract)
- [Implementation](#implementation)
  - [Basic Authenticator](#basic-authenticator)
  - [OAuth2 Authenticator](#oauth2-authenticator)
  - [Passwordless Authenticator](#passwordless-authenticator)
  - [API Key Authenticator](#api-key-authenticator)
- [Usage](#usage)
- [Extensibility](#extensibility)

## 🏗️ Architecture

The Credential layer follows a contract-first pattern with clear interfaces:

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

The credential layer defines **3 core interfaces** that all authenticators must implement:

### 1. Credentials Interface

Represents user credentials input for any authentication method.

```go
type Credentials interface {
    Type() string     // Returns the credential type (e.g., "basic", "oauth2", "apikey")
    Validate() error  // Validates credential format
}
```

### 2. Authenticator Interface

Main interface for authenticating credentials. **This is the only general contract** used by all authentication implementations.

```go
type Authenticator interface {
    Authenticate(ctx context.Context, creds Credentials) (*AuthenticationResult, error)
    Type() string // Returns authenticator type (must match Credentials.Type())
}
```

### 3. AuthenticationResult Struct

Standardized result returned by all authenticators.

```go
type AuthenticationResult struct {
    Success  bool               // Whether authentication succeeded
    Subject  string             // Authenticated user ID
    Claims   map[string]any     // Additional claims for token
    Metadata map[string]any     // Authentication metadata
    Error    error              // Error if authentication failed
}
```

### Additional Contracts (Implementation-Specific)

Some authenticators define their own specific contracts:

**Basic Authenticator:**
- `CredentialValidator` - Validates username/password complexity (see `01_credential/basic/contract.go`)
- `UserProvider` - Retrieves user data from storage

**Passkey Authenticator:**
- `CredentialStore` - Manages WebAuthn credentials (see `01_credential/passkey/store.go`)

**Passwordless Authenticator:**
- `TokenStore` - Manages OTP/Magic Link tokens
- `UserResolver` - Resolves user ID from email
- `TokenSender` - Sends tokens via email

**API Key Authenticator:**
- `KeyStore` - Stores and validates API keys

## 🔐 Implementation

### Basic Authenticator

Username/password authentication with bcrypt hashing and extensible storage.

**Features:**
- Bcrypt password hashing (cost factor 10)
- Username-based login
- Extensible `UserProvider` interface
- Standard `User` model for all implementations
- In-memory store for testing/development
- Database store example for production

**Architecture:**

```
01_credential/basic/
├── contract.go          # User model + UserProvider interface + CredentialValidator
├── authenticator.go     # Authentication logic
├── credential.go        # BasicCredentials struct
├── credential_validator.go  # Password complexity validation
├── store_inmemory.go    # In-memory implementation (testing)
└── store_database.go    # PostgreSQL example (production)
```

**Core Contracts:**

1. **User Model** - Standard model for all UserProvider implementations:

```go
type User struct {
    ID           string         // Unique user identifier
    Username     string         // Username for login
    PasswordHash string         // Bcrypt hashed password
    Email        string         // User email address
    Disabled     bool           // Account status
    Metadata     map[string]any // Custom metadata
}
```

2. **UserProvider Interface** - Storage abstraction:

```go
type UserProvider interface {
    GetUserByUsername(ctx context.Context, username string) (*User, error)
}
```

**Usage:**

```go
import "github.com/primadi/lokstra-auth/01_credential/basic"

// For testing/development - use in-memory store
userStore := basic.NewInMemoryUserStore()

// For production - use database store
// db, _ := sql.Open("postgres", connString)
// userStore := basic.NewDatabaseUserStore(db)

// Setup authenticator
auth := basic.NewAuthenticator(userStore, nil)

// Register user (in-memory example)
hashedPassword, _ := basic.HashPassword("mypassword")
userStore.AddUser(&basic.User{
    ID:       "user123",
    Username: "john",
    Email:    "john@example.com",
    PasswordHash: hashedPassword,
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

**Implementations:**

1. **InMemoryUserStore** - For testing and development:
   - Thread-safe with mutex
   - Helper methods: AddUser, RemoveUser, UpdateUser, ListUsers
   - No persistence

2. **DatabaseUserStore** - For production (PostgreSQL example):
   - SQL database backend
   - User lookup by username or email
   - CRUD operations
   - Can be adapted for MySQL, SQLite, etc.

**Custom Implementation:**

Implement `UserProvider` for your database/storage:

```go
type MongoUserStore struct {
    collection *mongo.Collection
}

func (s *MongoUserStore) GetUserByUsername(ctx context.Context, username string) (*basic.User, error) {
    var user basic.User
    filter := bson.M{"username": username, "disabled": false}
    err := s.collection.FindOne(ctx, filter).Decode(&user)
    if err == mongo.ErrNoDocuments {
        return nil, basic.ErrUserNotFound
    }
    return &user, err
}
```

---

### OAuth2 Authenticator

Authentication using OAuth2 providers (Google, GitHub, Facebook).

**Supported Providers:**
- **Google** - OpenID Connect
- **GitHub** - GitHub OAuth Apps
- **Facebook** - Facebook Login

**Features:**
- Authorization code flow
- Token validation with provider
- User info fetching
- Email verification
- Configurable client credentials per provider

**Usage:**

```go
import "github.com/primadi/lokstra-auth/01_credential/oauth2"

// Setup authenticator with multiple providers
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

// Authenticate with authorization code
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

**Claims from Providers:**

Google:
- `email` - User email
- `name` - Full name
- `picture` - Profile picture URL
- `email_verified` - Email verification status

GitHub:
- `email` - Primary email (verified)
- `login` - GitHub username
- `name` - Full name
- `avatar_url` - Avatar URL

Facebook:
- `email` - User email
- `name` - Full name
- `picture` - Profile picture URL

**Implementation Details:**

Authenticator uses HTTP client to communicate with provider APIs:
- Google: `https://www.googleapis.com/oauth2/v2/userinfo`
- GitHub: `https://api.github.com/user` + `/user/emails`
- Facebook: `https://graph.facebook.com/me`

---

### Passwordless Authenticator

Password-less authentication using Magic Link or OTP.

**Authentication Methods:**
1. **Magic Link** - URL token sent via email (15 minute expiry)
2. **OTP** - 6-digit code sent via email (5 minute expiry)

**Features:**
- Token lifecycle management (generation, validation, expiry)
- One-time use enforcement
- Extensible token storage (in-memory, Redis, database)
- Customizable token generator
- Email sender interface
- User resolver interface

**Magic Link Usage:**

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
// Token will be sent via email using TokenSender

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

**OTP Usage:**

```go
// Request OTP
creds := &passwordless.Credentials{
    Email:      "user@example.com",
    TokenType:  passwordless.TokenTypeOTP,
}

result, err := auth.Authenticate(ctx, creds)
// 6-digit OTP will be sent via email

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
- User must provide: `UserResolver`, `TokenSender`

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

API key-based authentication for service-to-service or programmatic access.

**Features:**
- SHA3-256 key hashing
- Key expiration support
- Scope-based permissions
- Key revocation
- Usage tracking (last used timestamp)
- Metadata support
- Key prefix for identification
- In-memory and custom key store

**Usage:**

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
    Metadata  map[string]any // Custom metadata
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

1. **Key Hashing**: API keys are hashed using SHA3-256 before storage
2. **One-time Display**: Key string only displayed during generation
3. **Constant Time Comparison**: Prevents timing attacks
4. **Automatic Expiry Check**: Validates expiry during authentication
5. **Revocation Support**: Soft delete with timestamp

**Best Practices:**

```go
// 1. Set expiration for production keys
thirtyDays := 30 * 24 * time.Hour
keyString, _, err := auth.GenerateKey(ctx, userID, "Prod", scopes, &thirtyDays)

// 2. Use scopes to limit permissions
scopes := []string{"read:users", "write:posts"}

// 3. Add metadata for tracking
apiKey.Metadata["app_name"] = "Mobile App"
apiKey.Metadata["environment"] = "production"

// 4. Rotate keys periodically
// Revoke old key
auth.RevokeKey(ctx, oldKeyID)
// Generate new key
newKey, _, _ := auth.GenerateKey(ctx, userID, "Rotated Key", scopes, &expiresIn)

// 5. Monitor last used to detect unused keys
if apiKey.LastUsed != nil && time.Since(*apiKey.LastUsed) > 90*24*time.Hour {
    // Key not used in 90 days - consider revoking
}
```

---

## 🚀 Usage

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

Use `MultiAuthenticator` to support various authentication methods:

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

Create custom authenticator by implementing the interface:

```go
type MyCustomAuthenticator struct{}

func (a *MyCustomAuthenticator) Authenticate(ctx context.Context, creds credential.Credentials) (*credential.AuthenticationResult, error) {
    // Your custom authentication logic
    return &credential.AuthenticationResult{
        Success: true,
        Subject: "user123",
        Claims: map[string]any{
            "method": "custom",
        },
    }, nil
}

func (a *MyCustomAuthenticator) Type() string {
    return "custom"
}
```

### Custom Storage Implementation

The basic authenticator includes both in-memory and database examples.

**Using the Database Store:**

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
    "github.com/primadi/lokstra-auth/01_credential/basic"
)

// Setup database connection
db, err := sql.Open("postgres", "postgres://user:pass@localhost/mydb?sslmode=disable")
if err != nil {
    log.Fatal(err)
}

// Create database user store
userStore := basic.NewDatabaseUserStore(db)

// Use with authenticator
auth := basic.NewAuthenticator(userStore, nil)
```

**Custom Implementation (e.g., MongoDB):**

```go
type MongoUserStore struct {
    collection *mongo.Collection
}

func (s *MongoUserStore) GetUserByUsername(ctx context.Context, username string) (*basic.User, error) {
    var user basic.User
    filter := bson.M{"username": username, "disabled": false}
    
    err := s.collection.FindOne(ctx, filter).Decode(&user)
    if err == mongo.ErrNoDocuments {
        return nil, basic.ErrUserNotFound
    }
    return &user, err
}

// Use it
userStore := &MongoUserStore{collection: mongoCollection}
auth := basic.NewAuthenticator(userStore, nil)
```

### Custom Token Store

Redis implementation for passwordless:

```go
type RedisTokenStore struct {
    client *redis.Client
}

func (s *RedisTokenStore) Store(ctx context.Context, email, token string, expiresAt time.Time, tokenType passwordless.TokenType) error {
    key := fmt.Sprintf("token:%s:%s:%s", tokenType, email, token)
    ttl := time.Until(expiresAt)
    
    data := map[string]any{
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

Database implementation for API keys:

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

Each authenticator provides in-memory implementations for testing without external dependencies:

- `basic.InMemoryUserStore` - For basic authentication testing
- `passwordless.InMemoryTokenStore` - For OTP/Magic Link testing
- `apikey.InMemoryKeyStore` - For API key testing

**Testing Basic Authentication:**

```go
func TestAuthentication(t *testing.T) {
    // Use in-memory store for testing
    userStore := basic.NewInMemoryUserStore()
    auth := basic.NewAuthenticator(userStore, nil)
    
    // Add test user
    hashedPassword, _ := basic.HashPassword("testpass")
    userStore.AddUser(&basic.User{
        ID:           "test123",
        Username:     "testuser",
        PasswordHash: hashedPassword,
        Email:        "test@example.com",
    })
    
    // Test authentication
    creds := &basic.Credentials{
        Username: "testuser",
        Password: "testpass",
    }
    
    result, err := auth.Authenticate(context.Background(), creds)
    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, "test123", result.Subject)
}
```

**For production code**, replace in-memory stores with database implementations:

```go
// Testing
userStore := basic.NewInMemoryUserStore()

// Production
db, _ := sql.Open("postgres", connString)
userStore := basic.NewDatabaseUserStore(db)
```

## 🔐 Security Considerations

1. **Password Hashing**: Use bcrypt with minimum cost factor of 10
2. **Token Expiry**: Set reasonable expiry for each token type
3. **One-time Tokens**: Enforce one-time use for magic links and OTP
4. **API Key Storage**: Don't store plain API keys, always hash
5. **HTTPS Only**: OAuth2 redirect URLs must be HTTPS in production
6. **Rate Limiting**: Implement rate limiting to prevent brute force
7. **Audit Logging**: Log all authentication attempts

## 📚 Resources

- [OAuth2 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [bcrypt Paper](https://www.usenix.org/legacy/events/usenix99/provos/provos.pdf)

---

**Layer 01 Complete** ✅  
Continue to [Layer 02: Token](../02_token/README.md) for JWT token management.
