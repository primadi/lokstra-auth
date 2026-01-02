# credential → token Integration

## Overview

Layer `credential` bertugas untuk **authenticate** (verify credentials), sedangkan layer `token` bertugas untuk **generate & manage tokens**.

Setelah credential berhasil diverifikasi, `credential` akan memanggil `token` untuk generate access token dan refresh token.

## Token Response Format

Semua authentication methods mengembalikan OAuth2-compliant token response:

```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": "usr_123",
    "username": "john.doe",
    "email": "john@example.com"
  }
}
```

### Token Types

1. **Access Token**
   - Short-lived (default: 1 hour)
   - Used for API authorization
   - Sent in `Authorization: Bearer {token}` header
   - Contains user claims and permissions

2. **Refresh Token**
   - Long-lived (default: 30 days)
   - Used to get new access token when expired
   - Stored securely on client (httpOnly cookie or secure storage)
   - Can be revoked/rotated

## Integration Flow

### 1. Basic Authentication Flow

```go
// credential/application/basic_service.go
func (s *BasicAuthService) Login(ctx *request.Context, req *basic.LoginRequest) (*basic.LoginResponse, error) {
    // Step 1: Authenticate credentials
    result, err := s.Authenticator.Authenticate(ctx, authCtx, creds)
    if err != nil {
        return nil, err
    }

    // Step 2: Get user info
    user, err := s.UserProvider.GetUserByID(ctx, result.TenantID, result.Subject)
    if err != nil {
        return nil, err
    }

    // Step 3: Generate tokens using token layer
    // @Inject "token-manager"
    // tokenManager token.Manager
    
    tokenReq := &token.GenerateRequest{
        TenantID: result.TenantID,
        AppID:    result.AppID,
        Subject:  result.Subject,
        Claims: map[string]any{
            "username": user.Username,
            "email":    user.Email,
            "roles":    user.Roles,
        },
    }
    
    tokens, err := s.TokenManager.Generate(ctx, tokenReq)
    if err != nil {
        return nil, err
    }

    // Step 4: Return response with tokens
    return &basic.LoginResponse{
        Success:      true,
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
        TokenType:    "Bearer",
        ExpiresIn:    tokens.ExpiresIn,
        User:         basic.ToUserInfo(user),
    }, nil
}
```

### 2. Token Refresh Flow

```go
// credential/application/token_service.go (new service)
func (s *TokenService) Refresh(ctx *request.Context, req *RefreshRequest) (*RefreshResponse, error) {
    // Step 1: Validate refresh token
    claims, err := s.TokenManager.Validate(ctx, req.RefreshToken)
    if err != nil {
        return nil, err
    }

    // Step 2: Check if refresh token is revoked
    if s.TokenStore.IsRevoked(ctx, req.RefreshToken) {
        return nil, errors.New("refresh token has been revoked")
    }

    // Step 3: Generate new access token (and optionally rotate refresh token)
    tokens, err := s.TokenManager.Refresh(ctx, req.RefreshToken)
    if err != nil {
        return nil, err
    }

    return &RefreshResponse{
        Success:      true,
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken, // New refresh token (rotation)
        TokenType:    "Bearer",
        ExpiresIn:    tokens.ExpiresIn,
    }, nil
}
```

## Required Services from token

### TokenManager Interface

```go
package token

type Manager interface {
    // Generate creates new access and refresh tokens
    Generate(ctx context.Context, req *GenerateRequest) (*TokenPair, error)

    // Refresh validates refresh token and generates new token pair
    Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)

    // Validate validates and parses token claims
    Validate(ctx context.Context, token string) (*Claims, error)

    // Revoke revokes a token (adds to blacklist)
    Revoke(ctx context.Context, token string) error
}

type GenerateRequest struct {
    TenantID string
    AppID    string
    Subject  string         // User ID
    Claims   map[string]any // Additional claims
}

type TokenPair struct {
    AccessToken  string
    RefreshToken string
    TokenType    string // "Bearer"
    ExpiresIn    int64  // seconds
}

type Claims struct {
    TenantID  string
    AppID     string
    Subject   string
    Username  string
    Email     string
    Roles     []string
    ExpiresAt int64
    IssuedAt  int64
}
```

## Authentication Methods → Token Generation

| Authentication Method | Credential Input | Token Generation |
|----------------------|------------------|------------------|
| **Basic** | Username + Password | ✅ After password verification |
| **API Key** | API Key | ❌ No tokens (API key IS the credential) |
| **OAuth2** | Authorization Code | ✅ After code exchange |
| **Passwordless** | Email/SMS Code | ✅ After code verification |
| **Passkey** | WebAuthn Assertion | ✅ After signature verification |

### Note: API Key Authentication

API Key authentication **does NOT generate tokens**. The API key itself serves as the long-lived credential:

```go
// API Key response does NOT include tokens
type AuthenticateResponse struct {
    Success   bool
    Validated bool
    KeyID     string
    TenantID  string
    AppID     string
    Scopes    []string
    // ❌ No access_token or refresh_token
}
```

Clients use the API key directly:
```http
GET /api/resource
Authorization: ApiKey acme-corp.main-app.dev_a1b2c3d4e5f6
```

## Token Security Best Practices

### 1. Access Token
- **Storage**: Memory only (never localStorage)
- **Lifetime**: Short (15 min - 1 hour)
- **Transport**: HTTPS only
- **Validation**: Signature + expiration + revocation check

### 2. Refresh Token
- **Storage**: httpOnly cookie (web) or secure storage (mobile)
- **Lifetime**: Medium (7-30 days)
- **Rotation**: Generate new refresh token on each refresh
- **Revocation**: Store in database for logout/security events

### 3. Token Rotation
```go
// On refresh, generate NEW refresh token and revoke old one
func (m *TokenManager) Refresh(ctx context.Context, oldRefreshToken string) (*TokenPair, error) {
    // 1. Validate old refresh token
    claims, err := m.Validate(ctx, oldRefreshToken)
    
    // 2. Generate new token pair
    newTokens, err := m.Generate(ctx, &GenerateRequest{
        TenantID: claims.TenantID,
        AppID:    claims.AppID,
        Subject:  claims.Subject,
    })
    
    // 3. Revoke old refresh token (prevent reuse)
    m.Revoke(ctx, oldRefreshToken)
    
    return newTokens, nil
}
```

## Implementation Checklist

- [x] Add `refresh_token` field to all authentication response DTOs
- [ ] Inject `TokenManager` into credential services
- [ ] Implement token generation after successful authentication
- [ ] Create `TokenService` for refresh/revoke operations
- [ ] Add token validation middleware
- [ ] Implement token rotation strategy
- [ ] Add token revocation on logout
- [ ] Store refresh tokens in database
- [ ] Add token blacklist/revocation list
- [ ] Implement token cleanup (expired tokens)

## Example: Complete Login Flow

```go
// Client request
POST //api/auth/cred/basic/login
X-Tenant-ID: acme-corp
X-App-ID: main-app
{
  "username": "john.doe",
  "password": "SecurePass123!"
}

// Backend processing
1. BasicAuthService validates username/password
2. BasicAuthenticator checks credentials against database
3. TokenManager generates access + refresh tokens
4. Response sent to client

// Response
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c3JfMTIzIiwidGVuYW50X2lkIjoiYWNtZS1jb3JwIiwiYXBwX2lkIjoibWFpbi1hcHAiLCJ1c2VybmFtZSI6ImpvaG4uZG9lIiwiZXhwIjoxNzAwMDAwMDAwfQ.xyz",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c3JfMTIzIiwidGVuYW50X2lkIjoiYWNtZS1jb3JwIiwiYXBwX2lkIjoibWFpbi1hcHAiLCJ0eXBlIjoicmVmcmVzaCIsImV4cCI6MTcwMjU5MjAwMH0.abc",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": "usr_123",
    "tenant_id": "acme-corp",
    "username": "john.doe",
    "email": "john.doe@acme.com"
  }
}

// Client usage
GET /api/protected-resource
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Next Steps

1. Implement `token` layer with JWT or simple token manager
2. Add `TokenManager` injection to credential services
3. Replace `"TODO_GENERATE_TOKEN"` with actual token generation
4. Create token refresh endpoint
5. Add token validation middleware
6. Implement logout (token revocation)
7. Add token cleanup background job
