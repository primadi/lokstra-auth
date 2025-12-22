# Layer 02 Token Examples

Examples demonstrating token management features including JWT tokens, opaque tokens, and token storage.

## 📁 Examples

### 01_jwt - JWT Token Management
Comprehensive demonstration of JWT token features:
- ✅ Access token generation
- ✅ Refresh token generation
- ✅ Token verification
- ✅ Refresh mechanism
- ✅ Token revocation
- ✅ Expiry validation
- ✅ Claims helpers
- ✅ Invalid token handling

**Run**:
```bash
cd 01_jwt
go run main.go
```

**Output**:
```
=== JWT Token Manager Example ===

1️⃣  Generating Access Token...
✅ Access Token Generated:
   Value: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   Type: Bearer
   Signing Method: HS256
   Expires At: 2025-11-12T04:32:27+07:00
   Duration: 15m0s
   
2️⃣  Verifying Access Token...
✅ Token Verification Successful!
   Subject: user123
   Email: john@example.com
   ...
```

---

### 02_simple - Simple Opaque Token Management
Demonstration of cryptographically secure opaque tokens:
- ✅ Opaque token generation
- ✅ Token verification
- ✅ Multiple tokens per user
- ✅ Token revocation
- ✅ Expiry validation
- ✅ Configurable token length
- ✅ Invalid token handling

**Run**:
```bash
cd 02_simple
go run main.go
```

**Output**:
```
=== Simple Opaque Token Manager Example ===

1️⃣  Generating Opaque Token...
✅ Opaque Token Generated:
   Value: uOYAvG_3Ri75FiEJz2jyLO9k_dbDVFr6NL6IfG-7Z-0=
   Type: Bearer
   Token Type: opaque
   
3️⃣  Generating Multiple Tokens...
✅ Generated 3 tokens for user456:
   Token 1 (desktop): uOYAvG_3Ri75...
   Token 2 (mobile): hta34rILQ4nG...
   Token 3 (web): N02LvHUbEoRR...
```

---

### 03_store - Token Store Management
Demonstration of token lifecycle management:
- ✅ Store tokens for multiple users
- ✅ Multiple tokens per user
- ✅ Retrieve specific tokens
- ✅ List all user tokens
- ✅ Revoke tokens
- ✅ Check revocation status
- ✅ Delete tokens
- ✅ Cleanup expired tokens

**Run**:
```bash
cd 03_store
go run main.go
```

**Output**:
```
=== Token Store Management Example ===

1️⃣  Storing Tokens for Multiple Users...
   ✅ Stored token for Alice (user1)
   ✅ Stored token for Bob (user2)
   ✅ Stored token for Charlie (user3)

4️⃣  Listing All Tokens for user1...
✅ Found 4 tokens for user1:
   1. unknown (ID: unknown)
   2. desktop (ID: user1-desktop)
   3. mobile (ID: user1-mobile)
   4. tablet (ID: user1-tablet)
```

---

## 🚀 Quick Start

### Prerequisites
```bash
# Install dependencies
go mod download
```

### Run All Examples
```bash
# JWT Token Example
cd 01_jwt && go run main.go

# Simple Token Example
cd 02_simple && go run main.go

# Token Store Example
cd 03_store && go run main.go
```

---

## 💡 Use Cases

### 1. API Authentication with JWT
```go
// From 01_jwt example
config := &jwt.JWTConfig{
    SigningKey:        []byte("your-secret-key"),
    SigningMethod:     "HS256",
    AccessDuration:    15 * time.Minute,
    RefreshDuration:   7 * 24 * time.Hour,
    Issuer:           "my-app",
    EnableRevocation: true,
}

manager := jwt.NewJWTManager(config)
token, _ := manager.Generate(ctx, "user123", metadata)
```

### 2. Session Management with Opaque Tokens
```go
// From 02_simple example
config := &simple.Config{
    TokenLength:      32,
    TokenDuration:    1 * time.Hour,
    EnableRevocation: true,
}

manager := simple.NewManager(config)
sessionToken, _ := manager.Generate(ctx, "user456", metadata)
```

### 3. Multi-Device Token Management
```go
// From 03_store example
store := NewInMemoryTokenStore()

// Store desktop token
store.Store(ctx, desktopToken)

// Store mobile token
store.Store(ctx, mobileToken)

// List all user devices
tokens, _ := store.List(ctx, userID)
```

---

## 🔒 Security Features

### JWT Manager
- ✅ HMAC SHA-256 signing
- ✅ Issuer validation
- ✅ Audience validation
- ✅ Expiry enforcement
- ✅ Revocation list support
- ✅ Refresh token mechanism

### Simple Token Manager
- ✅ Cryptographically secure random generation
- ✅ No information leakage (opaque)
- ✅ Revocation support
- ✅ Automatic cleanup
- ✅ Configurable token length

### Token Store
- ✅ Multi-user isolation
- ✅ Device tracking
- ✅ Revocation management
- ✅ Expiry cleanup
- ✅ Concurrent access safety

---

## 📊 Comparison

| Feature | JWT | Simple Token |
|---------|-----|--------------|
| Token Size | Large (~200 chars) | Small (44 chars) |
| Self-contained | ✅ Yes | ❌ No |
| Database Lookup | ❌ No (unless revoked) | ✅ Yes |
| Revocation | ⚠️ Requires list | ✅ Easy |
| Claims | ✅ Built-in | ❌ Stored separately |
| Performance | ⚡ Fast verification | 🐢 Requires lookup |
| Use Case | API, Microservices | Sessions, API Keys |

---

## 🔧 Configuration Examples

### Short-Lived Tokens
```go
config := &jwt.JWTConfig{
    AccessDuration:  5 * time.Minute,  // Very short
    RefreshDuration: 1 * time.Hour,    // Short
}
```

### Long-Lived Tokens
```go
config := &simple.Config{
    TokenDuration: 30 * 24 * time.Hour,  // 30 days
}
```

### High-Security Tokens
```go
config := &simple.Config{
    TokenLength:      64,  // 512 bits
    TokenDuration:    15 * time.Minute,
    EnableRevocation: true,
}
```

---

## 📚 Learn More

- [Layer 02 Token Documentation](../../token/README.md)
- [JWT Token Manager](../../token/jwt/)
- [Simple Token Manager](../../token/simple/)
- [Token Store](../../token/store.go)

---

## 🐛 Troubleshooting

### Token Verification Failed
```
Error: token signature is invalid
```
**Solution**: Check signing key matches between generation and verification

### Token Expired
```
Error: token has expired
```
**Solution**: Generate new token or use refresh token

### Token Revoked
```
Error: token has been revoked
```
**Solution**: User logged out or admin revoked access - require re-login

---

## License

MIT License - see [LICENSE](../../LICENSE) file for details.
