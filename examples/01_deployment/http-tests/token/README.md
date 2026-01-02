# Token Service HTTP Tests

This folder contains HTTP test files for the Token Service API endpoints.

## Endpoints Tested

### Token Validation
- `POST /api/auth/token/validate` - Validate JWT access tokens

### Token Refresh
- `POST /api/auth/token/refresh` - Refresh expired access tokens using refresh token

### Token Revocation
- `POST /api/auth/token/revoke` - Revoke refresh tokens (logout)

### Token Introspection
- `POST /api/auth/token/introspect` - Get token metadata and claims

## Test Files

### 01-token-service.http
Complete test suite for token operations including:
- Token validation (valid, expired, invalid, malformed)
- Token refresh (with rotation)
- Token revocation (logout)
- Token introspection
- Complete workflows
- Security testing
- Edge cases

## Prerequisites

Before running these tests:

1. **Have a running server** at `http://localhost:9090`
2. **Create a tenant** using `core/01-tenant-service.http`
3. **Create an app** using `core/02-app-service.http`
4. **Create a user and login** using `credential/01-basic-auth-service.http`
5. **Copy the tokens** from login response to the test file variables

## Quick Start

### 1. Get Tokens from Login

```http
POST http://localhost:9090/api/auth/cred/basic/login
X-Tenant-ID: acme-corp
X-App-ID: main-app
Content-Type: application/json

{
  "username": "john.doe",
  "password": "SecurePass123!"
}
```

**Response:**
```json
{
  "success": true,
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### 2. Update Variables in Test File

Open `01-token-service.http` and update:

```http
@accessToken = eyJhbGci...   # Copy from login response
@refreshToken = eyJhbGci...  # Copy from login response
```

### 3. Run Tests

Click "Send Request" above each test case in VS Code.

## Test Scenarios

### Basic Validation
```http
POST /api/auth/token/validate
Content-Type: application/json

{
  "token": "eyJhbGci..."
}
```

**Success Response:**
```json
{
  "valid": true,
  "claims": {
    "sub": "usr-123",
    "tenant_id": "acme-corp",
    "app_id": "main-app",
    "username": "john.doe",
    "exp": 1732634800
  }
}
```

### Token Refresh
```http
POST /api/auth/token/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGci..."
}
```

**Success Response:**
```json
{
  "success": true,
  "access_token": "eyJhbGci...",  // NEW access token
  "refresh_token": "eyJhbGci...",  // NEW refresh token (rotation)
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### Token Revocation
```http
POST /api/auth/token/revoke
Content-Type: application/json

{
  "refresh_token": "eyJhbGci..."
}
```

**Success Response:**
```json
{
  "success": true,
  "message": "token revoked successfully"
}
```

## Common Workflows

### Workflow 1: Normal Token Usage
1. Login → Get access_token + refresh_token
2. Use access_token for API calls
3. When access_token expires → Refresh using refresh_token
4. Get new access_token + new refresh_token (rotation)
5. Repeat steps 2-4 as needed
6. Logout → Revoke refresh_token

### Workflow 2: Token Rotation (Security Best Practice)
1. Login → Get tokens (RT1, AT1)
2. Access token expires
3. Refresh with RT1 → Get new tokens (RT2, AT2)
4. RT1 is now invalid (single-use)
5. Refresh with RT2 → Get new tokens (RT3, AT3)
6. RT2 is now invalid
7. Continue rotation...

### Workflow 3: Logout and Re-login
1. Have valid refresh_token
2. Revoke refresh_token (logout)
3. Try to refresh with revoked token → Fails
4. Must login again to get new tokens

## Expected Behaviors

### Valid Token
- ✅ Returns `valid: true`
- ✅ Returns token claims
- ✅ Expires check passes

### Expired Token
- ❌ Returns `valid: false`
- ❌ Error: "token expired"
- ❌ No claims returned

### Invalid Token
- ❌ Returns `valid: false`
- ❌ Error: "invalid token signature"
- ❌ No claims returned

### Malformed Token
- ❌ Returns `valid: false`
- ❌ Error: "malformed token"
- ❌ No claims returned

### Refresh Token Rotation
- ✅ New access_token issued
- ✅ New refresh_token issued
- ❌ Old refresh_token becomes invalid

### Revoked Token
- ❌ Cannot be used for refresh
- ❌ Error: "token revoked"
- ✅ Revocation is permanent

## Security Tests

### Cross-Tenant Attack
```http
# Login to tenant A
POST /login with tenant_id=acme-corp

# Try using token for tenant B (should fail)
POST /validate with tenant_id=other-corp
```
Expected: ❌ "tenant mismatch"

### Token Tampering
```http
# Modify token payload (change user_id, exp, etc.)
POST /validate with tampered_token
```
Expected: ❌ "invalid signature"

### Replay Attack (After Revocation)
```http
# Revoke token
POST /revoke with refresh_token

# Try using same token again
POST /refresh with refresh_token
```
Expected: ❌ "token revoked"

## Error Responses

### Missing Token
```json
{
  "success": false,
  "error": "token is required"
}
```

### Expired Token
```json
{
  "valid": false,
  "error": "token expired",
  "expired_at": "2024-11-26T10:00:00Z"
}
```

### Invalid Signature
```json
{
  "valid": false,
  "error": "invalid token signature"
}
```

### Revoked Token
```json
{
  "success": false,
  "error": "token has been revoked"
}
```

## Testing Tips

### Use VS Code REST Client
1. Install "REST Client" extension
2. Click "Send Request" above each `###` section
3. View response inline

### Save Responses
Use `@name` to save response and reference in later tests:
```http
# @name login
POST /login
...

###
# Use response from above
POST /validate
{
  "token": "{{login.response.body.access_token}}"
}
```

### Environment Variables
Create `.env` file for different environments:
```
# Development
@baseUrl = http://localhost:9090/api/auth/token
@tenantId = dev-tenant

# Staging
# @baseUrl = https://staging.api.com/api/auth/token
# @tenantId = staging-tenant

# Production
# @baseUrl = https://api.com/api/auth/token
# @tenantId = prod-tenant
```

## Related Documentation

- [Token Module](../../../02_token/README.md) - Token implementation details
- [JWT Manager](../../../02_token/jwt/manager.go) - JWT token manager
- [Credential Flow](../../../docs/credential_to_token_flow.md) - Auth to token flow

## Troubleshooting

### "Token expired" immediately after login
- Check server time vs client time
- Verify token expiration settings in config
- Default: access_token = 1 hour, refresh_token = 7 days

### "Invalid signature" for valid tokens
- Check JWT secret key matches between services
- Verify token was generated with same secret
- Check for whitespace in token string

### "Token revoked" but wasn't explicitly revoked
- Check if token store was cleared
- Verify refresh token isn't expired
- Check revocation list/blacklist

### Refresh always returns new tokens (no rotation)
- Verify token manager supports rotation
- Check configuration: `rotate_refresh_tokens: true`
- JWT manager should implement `GenerateRefreshToken()` with rotation

## Next Steps

After testing token operations:
1. Test with middleware → `middleware-tests.http`
2. Test authorization → `authz/` tests
3. Test complete flows → `core/00-workflow.http`
