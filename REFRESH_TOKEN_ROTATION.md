# Refresh Token Rotation Implementation

## Overview
Refresh token rotation has been implemented in the Basic Auth Service to enhance security by preventing refresh token reuse attacks.

## How It Works

### Traditional Refresh Flow (LESS SECURE)
```
Login → Access Token + Refresh Token
Use Refresh Token → New Access Token (SAME refresh token can be reused)
```
**Problem**: If a refresh token is stolen, it can be used indefinitely until expiration.

### Token Rotation Flow (MORE SECURE)
```
Login → Access Token + Refresh Token #1
Use Refresh Token #1 → New Access Token + NEW Refresh Token #2 (Token #1 is REVOKED)
Use Refresh Token #2 → New Access Token + NEW Refresh Token #3 (Token #2 is REVOKED)
```
**Benefit**: Each refresh operation invalidates the previous refresh token, limiting the window of attack.

## Implementation Details

### Code Changes

**File**: `credential/application/basic_service.go`

**Method**: `BasicAuthService.Refresh()`

```go
func (s *BasicAuthService) Refresh(ctx *request.Context, req *basic.RefreshRequest) (*basic.RefreshResponse, error) {
	// 1. Verify refresh token to extract claims
	result, err := s.TokenManager.Verify(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify refresh token: %w", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("invalid refresh token: %v", result.Error)
	}

	// 2. Generate new access token from claims
	newAccessToken, err := s.TokenManager.Generate(ctx, result.Claims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 3. Generate NEW refresh token (rotation) from same claims
	newRefreshToken, err := s.TokenManager.GenerateRefreshToken(ctx, result.Claims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	// 4. Revoke old refresh token (security: prevent reuse)
	if err := s.TokenManager.Revoke(ctx, req.RefreshToken); err != nil {
		// Log error but don't fail the request
		// Old token might already be expired/revoked
	}

	// 5. Return both new tokens
	return &basic.RefreshResponse{
		Success:      true,
		AccessToken:  newAccessToken.Value,
		RefreshToken: newRefreshToken.Value, // NEW field
		TokenType:    newAccessToken.Type,
		ExpiresIn:    int64(newAccessToken.ExpiresAt.Sub(newAccessToken.IssuedAt).Seconds()),
	}, nil
}
```

### DTO Changes

**File**: `credential/domain/basic/dto.go`

**Added Field**: `RefreshToken` to `RefreshResponse`

```go
type RefreshResponse struct {
	Success      bool   `json:"success"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"` // NEW: rotated refresh token
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}
```

## Security Benefits

1. **Limited Token Lifetime**: Even if a refresh token is stolen, it becomes invalid after a single use
2. **Detection of Token Theft**: If an attacker uses a stolen refresh token, the legitimate user's next refresh will fail, alerting them to the compromise
3. **OAuth 2.0 Best Practices**: Follows [RFC 6749](https://tools.ietf.org/html/rfc6749) recommendations for refresh token rotation

## Usage Example

### 1. Login
```http
POST /api/auth/cred/basic/login
Content-Type: application/json
X-Tenant-ID: acme-corp
X-App-ID: main-app

{
  "username": "john.doe",
  "password": "SecurePass123!"
}
```

**Response**:
```json
{
  "success": true,
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...REFRESH_TOKEN_1",
  "token_type": "Bearer",
  "expires_in": 900
}
```

### 2. Refresh Token (First Time)
```http
POST /api/auth/cred/basic/refresh
Content-Type: application/json
X-Tenant-ID: acme-corp
X-App-ID: main-app

{
  "refresh_token": "eyJhbGc...REFRESH_TOKEN_1"
}
```

**Response**:
```json
{
  "success": true,
  "access_token": "eyJhbGc...NEW_ACCESS_TOKEN",
  "refresh_token": "eyJhbGc...REFRESH_TOKEN_2",  // NEW token
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Important**: `REFRESH_TOKEN_1` is now REVOKED and cannot be used again.

### 3. Refresh Token (Second Time)
```http
POST /api/auth/cred/basic/refresh
Content-Type: application/json
X-Tenant-ID: acme-corp
X-App-ID: main-app

{
  "refresh_token": "eyJhbGc...REFRESH_TOKEN_2"  // Use NEW token from previous refresh
}
```

**Response**:
```json
{
  "success": true,
  "access_token": "eyJhbGc...ANOTHER_NEW_ACCESS_TOKEN",
  "refresh_token": "eyJhbGc...REFRESH_TOKEN_3",  // Another NEW token
  "token_type": "Bearer",
  "expires_in": 900
}
```

### 4. Try Reusing Old Token (FAILS)
```http
POST /api/auth/cred/basic/refresh
Content-Type: application/json
X-Tenant-ID: acme-corp
X-App-ID: main-app

{
  "refresh_token": "eyJhbGc...REFRESH_TOKEN_1"  // Old revoked token
}
```

**Response**:
```json
{
  "error": "failed to verify refresh token: token has been revoked"
}
```

## Client Implementation Guide

### Correct Implementation
```javascript
let accessToken = null;
let refreshToken = null;

// 1. Login
async function login(username, password) {
  const response = await fetch('/api/auth/cred/basic/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Tenant-ID': 'acme-corp',
      'X-App-ID': 'main-app'
    },
    body: JSON.stringify({ username, password })
  });
  
  const data = await response.json();
  accessToken = data.access_token;
  refreshToken = data.refresh_token; // Store initial refresh token
}

// 2. Refresh tokens
async function refreshTokens() {
  const response = await fetch('/api/auth/cred/basic/refresh', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Tenant-ID': 'acme-corp',
      'X-App-ID': 'main-app'
    },
    body: JSON.stringify({ refresh_token: refreshToken })
  });
  
  const data = await response.json();
  accessToken = data.access_token;
  refreshToken = data.refresh_token; // IMPORTANT: Update to NEW refresh token
}

// 3. Make authenticated request
async function makeAuthenticatedRequest() {
  let response = await fetch('/api/protected-resource', {
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });
  
  if (response.status === 401) {
    // Access token expired, refresh it
    await refreshTokens();
    
    // Retry with new access token
    response = await fetch('/api/protected-resource', {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });
  }
  
  return response;
}
```

### Common Mistakes to Avoid

❌ **WRONG**: Storing refresh token once and never updating it
```javascript
const refreshToken = data.refresh_token; // Never updated
```

✅ **CORRECT**: Always update refresh token after each refresh
```javascript
let refreshToken = data.refresh_token; // Let, not const
// After refresh:
refreshToken = newData.refresh_token; // Update to new token
```

## Testing

See `examples/01_deployment/http-tests/credential/01-basic-auth-service.http` for:
- Test #12: Refresh with token rotation
- Test #15: Verify old token is revoked after rotation

## Token Expiration Times

- **Access Token**: 15 minutes (configurable in `deployment.yaml`)
- **Refresh Token**: 7 days (configurable in `deployment.yaml`)

## Migration from Old Clients

Clients that don't update to use the new refresh token will fail after the first refresh. Migration steps:

1. Deploy the new API (backward compatible - returns both access and refresh tokens)
2. Update clients to store and use the new `refresh_token` from refresh responses
3. Monitor logs for failed refresh attempts (indicates clients that need updating)

## Related Files

- `credential/application/basic_service.go` - Refresh token rotation logic
- `credential/domain/basic/dto.go` - RefreshResponse with refresh_token field
- `token/contract.go` - TokenManager interface with Revoke method
- `token/jwt/manager.go` - JWT implementation of token rotation
- `examples/01_deployment/http-tests/credential/01-basic-auth-service.http` - Test cases
