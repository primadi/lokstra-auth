# Credential to Token Flow

## Overview

This document explains how authentication credentials are transformed into tokens in Lokstra-Auth.

## The Three-Layer Architecture

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Layer 1    │ → │   Layer 2    │ → │   Layer 3    │
│  Credential  │   │    Token     │   │   Subject    │
└──────────────┘    └──────────────┘    └──────────────┘
```

## Detailed Flow

### Step 1: Credential Authentication

**Input:**
```go
loginRequest := &LoginRequest{
    AuthContext: &credential.AuthContext{
        TenantID:  "acme-corp",
        AppID:     "web-portal",
        IPAddress: "192.168.1.100",
    },
    Credentials: &basic.Credentials{
        Username: "alice",
        Password: "secret123",
    },
}
```

**Processing in `auth.go` Line 155-171:**
```go
// Layer 1: Authenticate credentials
credType := request.Credentials.Type()
authenticator, ok := a.authenticators[credType]
if !ok {
    return nil, fmt.Errorf("%w: %s", ErrNoAuthenticator, credType)
}

authResult, err := authenticator.Authenticate(ctx, request.AuthContext, request.Credentials)
if err != nil {
    return nil, fmt.Errorf("authentication error: %w", err)
}

if !authResult.Success {
    return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, authResult.Error)
}
```

**Output:**
```go
authResult := &credential.AuthenticationResult{
    Success:  true,
    Subject:  "user-123",
    TenantID: "acme-corp",
    AppID:    "web-portal",
    Claims: map[string]any{
        "sub":       "user-123",
        "tenant_id": "acme-corp",
        "app_id":    "web-portal",
        "email":     "alice@acme.com",
        "username":  "alice",
        "roles":     []string{"admin", "editor"},
    },
}
```

### Step 2: Token Generation

**The Magic Happens Here** - `authResult.Claims` is passed to token manager:

**Processing in `auth.go` Line 173-179:**
```go
// Layer 2: Generate tokens
if a.tokenManager == nil {
    return nil, ErrNoTokenManager
}

accessToken, err := a.tokenManager.Generate(ctx, authResult.Claims)
if err != nil {
    return nil, fmt.Errorf("%w: %v", ErrTokenGenerationFailed, err)
}
```

**Inside JWT Manager (`token/jwt/manager.go`):**
```go
func (m *Manager) Generate(ctx context.Context, claims token.Claims) (*token.Token, error) {
    // 1. Validate multi-tenant claims
    tenantID, _ := claims.GetTenantID()  // Gets "acme-corp"
    appID, _ := claims.GetAppID()        // Gets "web-portal"
    
    // 2. Build JWT claims
    jwtClaims := jwt.MapClaims{
        "iat":       now.Unix(),
        "exp":       expiresAt.Unix(),
        "iss":       "lokstra-auth",
        "aud":       []string{"lokstra"},
        "tenant_id": tenantID,           // ← From authResult.Claims
        "app_id":    appID,              // ← From authResult.Claims
        "sub":       "user-123",         // ← From authResult.Claims
        "email":     "alice@acme.com",   // ← From authResult.Claims
        "roles":     []string{"admin"},  // ← From authResult.Claims
    }
    
    // 3. Sign JWT
    jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
    tokenString, _ := jwtToken.SignedString(signingKey)
    
    // 4. Return Token
    return &token.Token{
        Value:     tokenString,  // "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
        Type:      "Bearer",
        TenantID:  tenantID,
        AppID:     appID,
        ExpiresAt: expiresAt,
        IssuedAt:  now,
    }, nil
}
```

**Output:**
```go
accessToken := &token.Token{
    Value:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhcHBfaWQiOiJ3ZWItcG9ydGFsIiwiZW1haWwiOiJhbGljZUBhY21lLmNvbSIsInJvbGVzIjpbImFkbWluIl0sInN1YiI6InVzZXItMTIzIiwidGVuYW50X2lkIjoiYWNtZS1jb3JwIn0.xyz",
    Type:      "Bearer",
    TenantID:  "acme-corp",
    AppID:     "web-portal",
    ExpiresAt: time.Now().Add(15 * time.Minute),
    IssuedAt:  time.Now(),
}
```

### Step 3: Optional - Refresh Token

**Processing in `auth.go` Line 187-197:**
```go
// Generate refresh token if enabled
if a.config.IssueRefreshToken {
    if rtHandler, ok := a.tokenManager.(interface {
        GenerateRefreshToken(ctx context.Context, claims token.Claims) (*token.Token, error)
    }); ok {
        refreshToken, err := rtHandler.GenerateRefreshToken(ctx, authResult.Claims)
        if err == nil {
            response.RefreshToken = refreshToken
        }
    }
}
```

### Step 4: Return to Client

**Final Response:**
```go
response := &LoginResponse{
    AccessToken: &token.Token{
        Value:     "eyJhbGci...xyz",
        Type:      "Bearer",
        TenantID:  "acme-corp",
        AppID:     "web-portal",
        ExpiresAt: ...,
    },
    RefreshToken: &token.Token{
        Value:     "eyJhbGci...abc",
        Type:      "Bearer",
        TenantID:  "acme-corp",
        AppID:     "web-portal",
        ExpiresAt: ...,
    },
    Identity: ..., // Optional from Layer 3
}
```

## Key Points

### 1. **Claims Bridge**
`AuthenticationResult.Claims` adalah **jembatan** antara Layer 1 dan Layer 2:
- Layer 1 (Credential) mengisi claims dengan user info
- Layer 2 (Token) menggunakan claims untuk generate JWT

### 2. **Multi-Tenant Enforcement**
Both layers enforce multi-tenant:
- **Credential Layer**: Requires `AuthContext` with tenant_id + app_id
- **Token Layer**: Validates claims include tenant_id + app_id, embeds in JWT

### 3. **Claims Flow**
```
Authenticator.Authenticate()
    ↓ builds
AuthenticationResult.Claims
    ↓ passed to
TokenManager.Generate(claims)
    ↓ embeds in
JWT Token
    ↓ returned as
LoginResponse.AccessToken
```

### 4. **Token Validation (Reverse Flow)**
When validating a token:
```go
verifyResult, _ := tokenManager.Verify(ctx, tokenString)

// verifyResult.Claims contains:
{
    "sub":       "user-123",
    "tenant_id": "acme-corp",
    "app_id":    "web-portal",
    "email":     "alice@acme.com",
    "roles":     ["admin"],
}

// Use these claims for authorization
tenantID, _ := verifyResult.Claims.GetTenantID()
appID, _ := verifyResult.Claims.GetAppID()
```

## Code Example

### Complete Login Flow

```go
package main

import (
    "context"
    lokstraauth "github.com/primadi/lokstra-auth"
    "github.com/primadi/lokstra-auth/credential"
    "github.com/primadi/lokstra-auth/credential/basic"
    "github.com/primadi/lokstra-auth/token/jwt"
)

func main() {
    ctx := context.Background()
    
    // Setup authenticator
    userStore := basic.NewInMemoryUserStore()
    basicAuth := basic.NewAuthenticator(userStore)
    
    // Setup token manager
    jwtManager := jwt.NewManager(jwt.DefaultConfig("secret-key"))
    
    // Build auth runtime
    auth := lokstraauth.NewBuilder().
        WithAuthenticator("basic", basicAuth).
        WithTokenManager(jwtManager).
        Build()
    
    // User attempts login
    response, err := auth.Login(ctx, &lokstraauth.LoginRequest{
        AuthContext: &credential.AuthContext{
            TenantID: "acme-corp",
            AppID:    "web-portal",
        },
        Credentials: &basic.Credentials{
            Username: "alice",
            Password: "secret123",
        },
    })
    
    if err != nil {
        panic(err)
    }
    
    // Response contains token
    fmt.Println("Access Token:", response.AccessToken.Value)
    fmt.Println("Tenant:", response.AccessToken.TenantID)
    fmt.Println("App:", response.AccessToken.AppID)
    fmt.Println("Expires:", response.AccessToken.ExpiresAt)
    
    // Later: Verify the token
    verifyResult, _ := jwtManager.Verify(ctx, response.AccessToken.Value)
    if verifyResult.Valid {
        fmt.Println("Token valid!")
        fmt.Println("Claims:", verifyResult.Claims)
    }
}
```

## Diagrams

### Sequence Diagram

```
Client          Auth Runtime    Authenticator   TokenManager
  │                 │                │               │
  │─Login Request──→│                │               │
  │ (AuthContext+   │                │               │
  │  Credentials)   │                │               │
  │                 │                │               │
  │                 │─Authenticate──→│               │
  │                 │                │               │
  │                 │←AuthResult─────│               │
  │                 │ (Claims)       │               │
  │                 │                │               │
  │                 │─Generate(Claims)──────────────→│
  │                 │                │               │
  │                 │←Token──────────────────────────│
  │                 │                │               │
  │←LoginResponse───│                │               │
  │ (AccessToken)   │                │               │
  │                 │                │               │
```

### Data Transformation

```
┌─────────────────────────────────────────────────────────────┐
│ INPUT: Credentials                                          │
│ {                                                           │
│   Type: "basic",                                            │
│   Username: "alice",                                        │
│   Password: "secret123"                                     │
│ }                                                           │
└─────────────────────────────────────────────────────────────┘
                         ↓
                   [Authenticate]
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ INTERMEDIATE: AuthenticationResult.Claims                   │
│ {                                                           │
│   "sub":       "user-123",                                  │
│   "tenant_id": "acme-corp",                                 │
│   "app_id":    "web-portal",                                │
│   "email":     "alice@acme.com",                            │
│   "roles":     ["admin", "editor"]                          │
│ }                                                           │
└─────────────────────────────────────────────────────────────┘
                         ↓
                  [Generate Token]
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ OUTPUT: JWT Token                                           │
│ eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.                       │
│ eyJhcHBfaWQiOiJ3ZWItcG9ydGFsIiwiZW1haWwiOiJhbGljZUBhY21lLm │
│ NvbSIsInJvbGVzIjpbImFkbWluIl0sInN1YiI6InVzZXItMTIzIiwid │
│ GVuYW50X2lkIjoiYWNtZS1jb3JwIn0.                              │
│ signature...                                                │
│                                                             │
│ Decoded Payload:                                            │
│ {                                                           │
│   "iat": 1700236800,                                        │
│   "exp": 1700240400,                                        │
│   "iss": "lokstra-auth",                                    │
│   "aud": ["lokstra"],                                       │
│   "sub": "user-123",                                        │
│   "tenant_id": "acme-corp",                                 │
│   "app_id": "web-portal",                                   │
│   "email": "alice@acme.com",                                │
│   "roles": ["admin", "editor"]                              │
│ }                                                           │
└─────────────────────────────────────────────────────────────┘
```

## Summary

**The Answer:**
`AuthenticationResult.Claims` is automatically passed to `TokenManager.Generate()` by the `Auth.Login()` method. The claims become the JWT payload, and the tenant_id/app_id are validated and embedded in the token.

**No manual intervention needed** - the Auth runtime handles the entire flow automatically!
