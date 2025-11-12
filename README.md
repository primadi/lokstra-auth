# Lokstra Auth

Modular Authentication and Authorization framework for [Lokstra](https://github.com/primadi/lokstra).

## 📖 Overview

Lokstra Auth is a modular authentication and authorization framework built on top of Lokstra Framework. Designed with a 4-layer architecture that enables high flexibility and composability.

## 🏗️ Architecture: 4 Layers

Lokstra Auth divides the authentication and authorization process into 4 independent layers:

### 1. **Credential Layer** (`01_credential/`) ✅ COMPLETE
The first layer is responsible for receiving and validating credentials from various sources:
- ✅ **Basic Auth** - Username/password with bcrypt
- ✅ **OAuth2** - Google, GitHub, Facebook integration
- ✅ **Passwordless** - Magic Link and OTP via email
- ✅ **API Key** - Key-based authentication with SHA3-256 hashing
- ✅ **Passkey** - WebAuthn/FIDO2 support

**Status**: Production ready with 5 authenticator types
**Documentation**: [01_credential/README.md](./01_credential/README.md)

### 2. **Token Layer** (`02_token/`) ✅ COMPLETE
The second layer manages token lifecycle and data extraction:
- ✅ **JWT Manager** - Access + Refresh token with rotation
- ✅ **Simple Token** - Opaque token management
- ✅ **Token Store** - In-memory token storage for testing
- ✅ Claim extraction and validation
- ✅ Custom token formats

**Status**: Production ready with 2 token manager types
**Documentation**: [02_token/README.md](./02_token/README.md)

### 3. **Subject Layer** (`03_subject/`) ✅ COMPLETE
The third layer transforms claims into complete identity context:
- ✅ **Simple Resolver** - Direct claim to identity mapping
- ✅ **Enriched Resolver** - Identity enrichment with external data
- ✅ **Cached Resolver** - Performance optimization with caching
- ✅ **Identity Store** - In-memory user data storage
- ✅ Role and permission loading
- ✅ Multi-source data aggregation

**Status**: Production ready with 3 resolver types
**Documentation**: [03_subject/README.md](./03_subject/README.md)

### 4. **Authorization Layer** (`04_authz/`) ✅ COMPLETE
The fourth layer performs access evaluation and policy enforcement:
- ✅ **RBAC** - Role-Based Access Control with wildcard support
- ✅ **ABAC** - Attribute-Based Access Control with rules
- ✅ **ACL** - Resource-level Access Control Lists
- ✅ **Policy-Based** - Flexible policy evaluation with combining algorithms
- ✅ Resource-level permissions
- ✅ Thread-safe implementations

**Status**: Production ready with 4 authorization models
**Documentation**: [04_authz/README.md](./04_authz/README.md)

## 🚀 Quick Start

### Installation

```bash
go get github.com/primadi/lokstra-auth
```

### Simple Example: Complete Authentication Flow

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/primadi/lokstra-auth/01_credential/basic"
    "github.com/primadi/lokstra-auth/02_token/jwt"
    "github.com/primadi/lokstra-auth/03_subject/simple"
    authz "github.com/primadi/lokstra-auth/04_authz"
    "github.com/primadi/lokstra-auth/04_authz/rbac"
)

func main() {
    ctx := context.Background()

    // Layer 1: Setup authentication
    userProvider := basic.NewInMemoryUserProvider()
    passwordHash, _ := basic.HashPassword("MySecure@Pass123")
    userProvider.AddUser(&basic.User{
        ID:           "user-001",
        Username:     "john.doe",
        PasswordHash: passwordHash,
    })
    
    validator := basic.NewValidator(basic.DefaultValidatorConfig())
    authenticator := basic.NewAuthenticator(userProvider, validator)

    // Layer 2: Setup token management
    jwtConfig := jwt.DefaultConfig("your-secret-key")
    tokenManager := jwt.NewManager(jwtConfig)

    // Layer 3: Setup subject resolution
    resolver := simple.NewResolver()
    roleProvider := simple.NewStaticRoleProvider(map[string][]string{
        "user-001": {"admin", "user"},
    })
    contextBuilder := simple.NewContextBuilder(
        roleProvider,
        simple.NewStaticPermissionProvider(map[string][]string{}),
        simple.NewStaticGroupProvider(map[string][]string{}),
        simple.NewStaticProfileProvider(map[string]map[string]any{}),
    )

    // Layer 4: Setup authorization
    rbacEvaluator := rbac.NewEvaluator(map[string][]string{
        "admin": {"*"}, // Admin has all permissions
        "user":  {"read:posts", "create:posts"},
    })

    // Use it!
    // 1. Authenticate
    authResult, _ := authenticator.Authenticate(ctx, &basic.BasicCredentials{
        Username: "john.doe",
        Password: "MySecure@Pass123",
    })
    fmt.Println("✅ Authenticated:", authResult.Subject)

    // 2. Generate token
    token, _ := tokenManager.Generate(ctx, authResult.Claims)
    fmt.Println("✅ Token generated")

    // 3. Verify token
    verifyResult, _ := tokenManager.Verify(ctx, token.Value)
    fmt.Println("✅ Token verified")

    // 4. Build identity
    subject, _ := resolver.Resolve(ctx, verifyResult.Claims)
    identity, _ := contextBuilder.Build(ctx, subject)
    fmt.Println("✅ Identity built, Roles:", identity.Roles)

    // 5. Check authorization
    decision, _ := rbacEvaluator.Evaluate(ctx, &authz.AuthorizationRequest{
        Subject:  identity,
        Resource: &authz.Resource{Type: "posts", ID: "123"},
        Action:   authz.ActionRead,
    })
    fmt.Println("✅ Authorization:", decision.Allowed)
}
```

### More Examples

See [examples/](./examples/) directory for complete working examples:
- **Basic Flow**: [examples/complete/01_basic_flow/](./examples/complete/01_basic_flow/)
- **Multi-Credential**: [examples/complete/02_multi_auth/](./examples/complete/02_multi_auth/)

## 📝 Detailed Examples

### Basic Authentication Example

```go
import (
    "github.com/primadi/lokstra-auth/01_credential/basic"
)

// Create authenticator
userStore := basic.NewInMemoryUserStore()
auth := basic.NewAuthenticator(&basic.Config{
    UserStore: userStore,
})

// Register user
hashedPassword, _ := basic.HashPassword("mypassword")
userStore.AddUser(&basic.User{
    ID:       "user123",
    Username: "john",
    Password: hashedPassword,
})

// Authenticate
creds := &basic.Credentials{
    Username: "john",
    Password: "mypassword",
}

result, err := auth.Authenticate(ctx, creds)
if result.Success {
    fmt.Println("Logged in as:", result.Subject)
}
```

### OAuth2 Authentication Example

```go
import (
    "github.com/primadi/lokstra-auth/01_credential/oauth2"
)

// Create OAuth2 authenticator
auth := oauth2.NewAuthenticator(nil) // Uses default providers

// Authenticate with Google access token
creds := &oauth2.Credentials{
    Provider:    oauth2.ProviderGoogle,
    AccessToken: "ya29.a0AfH6SMBxxxxx...",
}

result, err := auth.Authenticate(ctx, creds)
if result.Success {
    email := result.Claims["email"].(string)
    name := result.Claims["name"].(string)
    // ...
}
```

### Passwordless Authentication Example

```go
import (
    "github.com/primadi/lokstra-auth/01_credential/passwordless"
)

// Create passwordless authenticator
auth := passwordless.NewAuthenticator(&passwordless.Config{
    TokenStore:   passwordless.NewInMemoryTokenStore(),
    UserResolver: myUserResolver,
    TokenSender:  myEmailSender,
})

// Request magic link
err := auth.InitiateMagicLink(ctx, "user@example.com", "user123", "https://myapp.com")
// Email sent with magic link

// Verify magic link token
creds := &passwordless.Credentials{
    Email:     "user@example.com",
    Token:     "token-from-email",
    TokenType: passwordless.TokenTypeMagicLink,
}

result, err := auth.Authenticate(ctx, creds)
```

### API Key Authentication Example

```go
import (
    "github.com/primadi/lokstra-auth/01_credential/apikey"
)

// Create API key authenticator
keyStore := apikey.NewInMemoryKeyStore()
auth := apikey.NewAuthenticator(&apikey.Config{
    KeyStore: keyStore,
})

// Generate API key
expiresIn := 30 * 24 * time.Hour
keyString, apiKey, err := auth.GenerateKey(
    ctx,
    "user123",                    // User ID
    "Production API Key",         // Key name
    []string{"read", "write"},    // Scopes
    &expiresIn,
)

// Authenticate with API key
creds := &apikey.Credentials{
    APIKey: keyString,
}

result, err := auth.Authenticate(ctx, creds)
if result.Success {
    scopes := result.Claims["scopes"]
    // ...
}
```

## 📦 Project Structure

```
lokstra-auth/
├── 01_credential/      # ✅ Layer 1: Credential Input (COMPLETE)
│   ├── contract.go     # Core interfaces
│   ├── basic/          # Username/password
│   ├── oauth2/         # OAuth2 (Google, GitHub, Facebook)
│   ├── passwordless/   # Magic Link & OTP
│   ├── apikey/         # API key authentication
│   └── README.md       # ✅ Complete documentation
├── 02_token/           # ✅ Layer 2: Token Verification (COMPLETE)
│   ├── contract.go     # Core interfaces
│   ├── jwt/            # JWT with access+refresh tokens
│   ├── simple/         # Simple token manager
│   └── README.md       # ✅ Complete documentation
├── 03_subject/         # ✅ Layer 3: Subject Resolution (COMPLETE)
│   ├── contract.go     # Interface definitions
│   ├── simple/         # Simple resolver
│   ├── enriched/       # Enriched resolver with external data
│   ├── cached/         # Cached resolver for performance
│   └── README.md       # ✅ Complete documentation
├── 04_authz/           # ✅ Layer 4: Authorization (COMPLETE)
│   ├── contract.go     # Interface definitions
│   ├── rbac/           # Role-based access control
│   ├── abac/           # Attribute-based access control
│   ├── acl/            # Access control lists
│   ├── policy/         # Policy-based authorization
│   └── README.md       # ✅ Complete documentation
├── middleware/         # ✅ Lokstra Framework Integration
│   ├── auth.go         # Token verification middleware
│   ├── permission.go   # Permission check middleware
│   └── role.go         # Role check middleware
├── examples/           # ✅ Working Examples
│   ├── 01_credential/  # Credential layer examples
│   │   ├── 01_basic/       # Basic auth flow
│   │   ├── 02_multi_auth/  # Multi-authenticator
│   │   ├── 03_oauth2/      # ✅ OAuth2 example
│   │   ├── 04_passwordless/# ✅ Passwordless example
│   │   └── 05_apikey/      # ✅ API Key example
│   ├── 02_token/       # ✅ Token layer examples
│   ├── 03_subject/     # ✅ Subject layer examples
│   ├── 04_authz/       # ✅ Authorization layer examples
│   │   ├── 01_rbac/        # RBAC examples
│   │   ├── 02_abac/        # ABAC examples
│   │   └── 03_acl/         # ACL examples
│   └── complete/       # Complete 4-layer integration
│       ├── 01_basic_flow/  # Basic authentication flow
│       └── 02_multi_auth/  # Multi-credential demo
└── README.md           # This file
```

## 📚 Documentation

### Layer Documentation
- ✅ [Layer 1: Credential](./01_credential/README.md) - **Complete** - Basic, OAuth2, Passwordless, API Key
- ✅ [Layer 2: Token](./02_token/README.md) - **Complete** - JWT (Access+Refresh), Simple, Store
- ✅ [Layer 3: Subject](./03_subject/README.md) - **Complete** - Simple, Enriched, Cached resolvers
- ✅ [Layer 4: Authorization](./04_authz/README.md) - **Complete** - RBAC, ABAC, ACL, Policy-based

### Examples
- ✅ [Basic Authentication](./examples/01_credential/01_basic/) - Username/password flow
- ✅ [Multi-Authenticator](./examples/01_credential/02_multi_auth/) - Multiple auth methods
- ✅ [OAuth2 Auth](./examples/01_credential/03_oauth2/) - Provider integration guide
- ✅ [Passwordless Auth](./examples/01_credential/04_passwordless/) - Magic Link & OTP
- ✅ [API Key Auth](./examples/01_credential/05_apikey/) - Full API key lifecycle
- ✅ [JWT Token Management](./examples/02_token/) - Access & refresh tokens
- ✅ [Subject Resolution](./examples/03_subject/) - Identity enrichment & caching
- ✅ [Authorization Examples](./examples/04_authz/) - RBAC, ABAC, ACL examples
- ✅ [Complete Flow](./examples/complete/01_basic_flow/) - All 4 layers integrated
- ✅ [Multi-Credential Demo](./examples/complete/02_multi_auth/) - Multiple auth methods with RBAC

## ✨ Features

### Credential Layer (01_credential/)
- ✅ **5 Authenticator Types**: Basic, OAuth2, Passwordless, API Key, Passkey
- ✅ **Provider Support**: Google, GitHub, Facebook OAuth2
- ✅ **Passwordless Methods**: Magic Link (15min TTL), OTP (5min TTL)
- ✅ **API Key Features**: SHA3-256 hashing, scopes, expiry, revocation
- ✅ **Passkey Support**: WebAuthn/FIDO2 authentication
- ✅ **Multi-Authenticator**: Handle multiple auth methods simultaneously
- ✅ **Extensible**: Custom authenticators via interface
- ✅ **In-Memory Stores**: Testing-ready implementations

### Token Layer (02_token/)
- ✅ JWT generation with access + refresh tokens
- ✅ Automatic token rotation
- ✅ Token verification and validation
- ✅ Simple opaque token management
- ✅ Token store for testing
- ✅ Configurable token expiry
- ✅ Custom claims support

### Subject Layer (03_subject/)
- ✅ Simple subject resolver (direct mapping)
- ✅ Enriched resolver (external data integration)
- ✅ Cached resolver (performance optimization)
- ✅ Identity store for user data
- ✅ User/subject resolution from tokens
- ✅ Identity context building
- ✅ Claims enrichment with roles, permissions, profile
- ✅ Multi-source data aggregation

### Authorization Layer (04_authz/)
- ✅ Role-Based Access Control (RBAC) with wildcards
- ✅ Attribute-Based Access Control (ABAC) with conditional rules
- ✅ Access Control Lists (ACL) for fine-grained permissions
- ✅ Policy-based authorization with multiple combining algorithms
- ✅ Permission and role checking helpers
- ✅ Resource-level access control
- ✅ Thread-safe implementations
- ✅ Flexible policy evaluation

### Integration
- ✅ Modular design - use any layer independently
- ✅ Composable - combine layers as needed
- ✅ Production-ready implementations
- ✅ Comprehensive examples
- ✅ Complete documentation

## 🔐 Security Features

1. **Password Security**
   - Bcrypt hashing (cost factor 10)
   - Constant-time comparison

2. **Token Security**
   - JWT with HS256/RS256
   - Configurable expiration
   - Refresh token rotation

3. **API Key Security**
   - SHA3-256 hashing
   - One-time display
   - Constant-time comparison
   - Automatic expiry checking

4. **Passwordless Security**
   - One-time use tokens
   - Time-based expiration
   - Cryptographically secure random generation
   - Automatic cleanup

5. **OAuth2 Security**
   - Token validation with provider
   - Email verification checking
   - HTTPS-only in production

6. **Passkey Security**
   - WebAuthn/FIDO2 standard compliance
   - Public key cryptography
   - Phishing-resistant authentication

## 🧪 Testing

Each layer comes with in-memory implementations for testing:

```go
// Basic Auth testing
userStore := basic.NewInMemoryUserStore()
userStore.AddUser(&basic.User{...})

// Passwordless testing
tokenStore := passwordless.NewInMemoryTokenStore()

// API Key testing
keyStore := apikey.NewInMemoryKeyStore()

// Run examples
go run examples/01_credential/05_apikey/main.go
go run examples/01_credential/04_passwordless/main.go
go run examples/complete/02_multi_auth/main.go
```

## 🎯 Design Principles

1. **Modularity** - Each layer can be used independently
2. **Composability** - Layers can be combined as needed
3. **Extensibility** - Easy to add new providers or strategies
4. **Type Safety** - Leveraging Go interfaces for type-safe operations
5. **Lokstra Integration** - Built on top of Lokstra Framework
6. **Production Ready** - Following security best practices
7. **Developer Friendly** - Clear APIs and comprehensive documentation

## 📋 Requirements

- Go 1.21 or higher
- [Lokstra Framework](https://github.com/primadi/lokstra) v0.3.4+

## 🗺️ Roadmap

### Layer 1: Credential ✅
- [x] Basic authenticator
- [x] OAuth2 authenticator (Google, GitHub, Facebook)
- [x] Passwordless authenticator (Magic Link, OTP)
- [x] API Key authenticator
- [x] Passkey/WebAuthn authenticator
- [x] Multi-authenticator support
- [x] Complete documentation
- [x] Working examples

### Layer 2: Token ✅
- [x] JWT token manager
- [x] Access + Refresh token support
- [x] Simple token manager
- [x] Token store implementation
- [x] Complete documentation
- [x] Working examples

### Layer 3: Subject ✅
- [x] Simple subject resolver
- [x] Enriched resolver with external data
- [x] Cached resolver for performance
- [x] Identity store implementation
- [x] Identity context builder
- [x] Complete documentation
- [x] Working examples

### Layer 4: Authorization ✅
- [x] RBAC authorizer with wildcards
- [x] ABAC authorizer with rules
- [x] ACL manager for resource permissions
- [x] Policy-based authorization
- [x] Policy store implementation
- [x] Multiple combining algorithms
- [x] Complete documentation
- [x] Working examples

### Integration ✅
- [x] Complete 4-layer examples
- [x] Multi-credential demo
- [x] Comprehensive documentation
- [ ] Auth runtime orchestrator
- [ ] Builder API
- [ ] Lokstra middleware
- [ ] Testing utilities
- [ ] Benchmark suite

## 📄 License

See [LICENSE](./LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

**Built with ❤️ using [Lokstra Framework](https://github.com/primadi/lokstra)**
