# Multi-Tab Refresh Token Problem & Solution

## 🔴 Problem: Shared Refresh Token with Rotation

When multiple tabs share the same refresh token and rotation is enabled, the second tab gets logged out.

### Timeline Without Grace Period:

```
Time: 00:00
├─ Tab 1: refresh_token_A (active) ✅
└─ Tab 2: refresh_token_A (active) ✅

Time: 00:05 - Tab 1 access token expires
├─ Tab 1: POST /refresh with refresh_token_A
├─ Backend: Mark refresh_token_A as "used"
├─ Backend: Generate refresh_token_B (new)
├─ Tab 1: Receives refresh_token_B ✅
└─ Tab 2: Still has refresh_token_A (now "used")

Time: 00:10 - Tab 2 access token expires
├─ Tab 2: POST /refresh with refresh_token_A
├─ Backend: Check refresh_token_A → status = "used" ❌
├─ Backend: Error "Token already used"
└─ Tab 2: LOGGED OUT! 💥
```

## ✅ Solution: Grace Period

Allow recently-used tokens to still be valid for a short period (e.g., 30 seconds).

### Timeline With Grace Period (30 seconds):

```
Time: 00:00
├─ Tab 1: refresh_token_A (active) ✅
└─ Tab 2: refresh_token_A (active) ✅

Time: 00:05 - Tab 1 access token expires
├─ Tab 1: POST /refresh with refresh_token_A
├─ Backend: Mark refresh_token_A as "used" (used_at = 00:05)
├─ Backend: Generate refresh_token_B (new)
├─ Backend: Store refresh_token_B with rotated_from_id = refresh_token_A.id
├─ Tab 1: Receives refresh_token_B ✅
└─ Tab 2: Still has refresh_token_A (now "used")

Time: 00:10 - Tab 2 access token expires (5 seconds later)
├─ Tab 2: POST /refresh with refresh_token_A
├─ Backend: Check refresh_token_A → status = "used", used_at = 00:05
├─ Backend: Calculate: now - used_at = 5 seconds
├─ Backend: 5 seconds < 30 seconds grace period ✅
├─ Backend: Allow refresh! Generate refresh_token_C
├─ Backend: Store refresh_token_C with rotated_from_id = refresh_token_A.id
└─ Tab 2: Receives refresh_token_C ✅

Time: 00:40 - Attacker tries to use stolen refresh_token_A (35 seconds later)
├─ Attacker: POST /refresh with refresh_token_A
├─ Backend: Check refresh_token_A → status = "used", used_at = 00:05
├─ Backend: Calculate: now - used_at = 35 seconds
├─ Backend: 35 seconds > 30 seconds grace period ❌
├─ Backend: REUSE DETECTED! Revoke entire token family
└─ Result: All tabs (Tab 1 & Tab 2) logged out (security measure)
```

## Configuration

### Tenant Config Example:

```json
{
  "security": {
    "enable_refresh_token_rotation": true,
    "refresh_token_rotation_grace_period": 30,
    "detect_refresh_token_reuse": true
  }
}
```

### Parameters:

| Parameter | Type | Description |
|-----------|------|-------------|
| `enable_refresh_token_rotation` | bool | Enable automatic refresh token rotation on each use |
| `refresh_token_rotation_grace_period` | int | Seconds to allow reuse of recently-used token (for multi-tab) |
| `detect_refresh_token_reuse` | bool | Revoke entire token family if reuse detected outside grace period |

## Recommended Settings

### High Security (Banking, Payment):
```json
{
  "enable_refresh_token_rotation": true,
  "refresh_token_rotation_grace_period": 10,
  "detect_refresh_token_reuse": true
}
```
- Tight grace period (10s)
- Aggressive revocation on reuse
- ⚠️ May affect multi-tab UX slightly

### Balanced (Most Apps):
```json
{
  "enable_refresh_token_rotation": true,
  "refresh_token_rotation_grace_period": 30,
  "detect_refresh_token_reuse": true
}
```
- Comfortable grace period (30s)
- Detect and revoke on suspicious reuse
- ✅ Good balance of security and UX

### Developer Friendly (Development):
```json
{
  "enable_refresh_token_rotation": false,
  "refresh_token_rotation_grace_period": 0,
  "detect_refresh_token_reuse": false
}
```
- No rotation (simpler debugging)
- No grace period needed
- ⚠️ Less secure - not for production

## Database Schema

```sql
-- refresh_tokens table
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    token_family UUID,                  -- For grouping rotated tokens
    rotated_from_id UUID,               -- Previous token in rotation chain
    status VARCHAR(50) NOT NULL,        -- active, used, revoked, expired
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,                  -- When token was used (for grace period)
    revoked_at TIMESTAMP,
    
    CONSTRAINT fk_rotated_from FOREIGN KEY (rotated_from_id) 
        REFERENCES refresh_tokens(id) ON DELETE SET NULL
);

CREATE INDEX idx_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_token_family ON refresh_tokens(token_family);
CREATE INDEX idx_rotated_from ON refresh_tokens(rotated_from_id);
```

## Flow Diagram

```
┌──────────────────────────────────────────────────────────────┐
│               Refresh Token Flow with Grace Period           │
└──────────────────────────────────────────────────────────────┘

1. Client Request:
   POST /auth/refresh
   { "refresh_token": "abc123..." }
   
2. Lookup Token:
   hash = SHA256(token_value)
   token = GetByToken(hash)
   
3. Check Status:
   ├─ Status = "active"
   │  ├─ Mark as "used" (used_at = NOW)
   │  ├─ Generate new token (rotation)
   │  └─ Return new token ✅
   │
   ├─ Status = "used"
   │  ├─ Calculate: age = NOW - used_at
   │  ├─ If age <= grace_period:
   │  │  ├─ Allow (multi-tab scenario)
   │  │  ├─ Generate new token
   │  │  └─ Return new token ✅
   │  └─ If age > grace_period:
   │     ├─ REUSE DETECTED! ❌
   │     ├─ Revoke token family
   │     └─ Return error (force re-login)
   │
   ├─ Status = "revoked"
   │  └─ Return error ❌
   │
   └─ Status = "expired"
      └─ Return error ❌
```

## Security Benefits

1. **Token Rotation**: New token on each use → old tokens become invalid
2. **Grace Period**: Allows legitimate multi-tab usage without UX issues
3. **Reuse Detection**: Catches stolen tokens used after grace period
4. **Family Revocation**: Immediately revokes all related tokens on breach
5. **Audit Trail**: Track token usage via `used_at`, `rotated_from_id`

## Implementation Checklist

- [x] Add `rotated_from_id` column to `refresh_tokens` table
- [x] Add `used_at` timestamp tracking
- [x] Implement `IsWithinGracePeriod()` helper method
- [x] Add grace period config to `TenantSecurityConfig`
- [x] Implement rotation logic with grace period check
- [x] Add reuse detection and family revocation
- [ ] Add logging for security events (token reuse detected)
- [ ] Add metrics for monitoring (grace period usage, reuse attempts)
- [ ] Document for frontend developers (handle token rotation)

## Frontend Integration

```javascript
// Frontend needs to update stored refresh token after each refresh
async function refreshAccessToken() {
  const currentRefreshToken = localStorage.getItem('refresh_token');
  
  const response = await fetch('/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: currentRefreshToken })
  });
  
  if (!response.ok) {
    // Token invalid → logout
    logout();
    return null;
  }
  
  const data = await response.json();
  
  // ⚠️ IMPORTANT: Update stored refresh token!
  localStorage.setItem('refresh_token', data.refresh_token);
  localStorage.setItem('access_token', data.access_token);
  
  return data.access_token;
}
```

## Testing Scenarios

### Test 1: Multi-Tab Normal Usage
1. Open 2 tabs with same login
2. Wait for access token expiry in Tab 1
3. Tab 1 refreshes → gets new refresh token
4. Within 30s, Tab 2 refreshes → should work ✅
5. Both tabs should remain logged in

### Test 2: Grace Period Expiry
1. Login and get refresh token
2. Use token to refresh (mark as "used")
3. Wait > grace period (e.g., 35 seconds)
4. Try to use same token again
5. Should fail with "reuse detected" ❌

### Test 3: Token Theft Detection
1. Login and get refresh token
2. Refresh normally → get new token
3. Attacker uses old token after grace period
4. Should revoke entire token family ❌
5. Legitimate user should be logged out (security)

## Monitoring & Alerts

Log and monitor these events:
- `refresh_token.reuse_detected` - Alert! Possible token theft
- `refresh_token.grace_period_used` - Normal multi-tab usage
- `refresh_token.family_revoked` - Security breach response
- `refresh_token.rotation_success` - Normal operation

## FAQ

**Q: Why not just disable rotation?**  
A: Rotation provides security. If token is stolen, it becomes invalid after next refresh.

**Q: What if user has slow internet?**  
A: Grace period accounts for this. 30s is usually enough for slow connections.

**Q: Should grace period be longer?**  
A: Longer = better UX, but less secure. 30s is a good balance.

**Q: What about single-page apps (SPA)?**  
A: SPAs typically only need 1 refresh token per browser, so grace period helps.

**Q: Can I disable rotation in development?**  
A: Yes, set `enable_refresh_token_rotation: false` for simpler debugging.
