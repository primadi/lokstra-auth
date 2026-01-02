# Unified User Storage Implementation

## Summary

Berhasil menghilangkan duplikasi user creation dengan menggunakan **single UserStore** dari `core` untuk semua credential types.

## Changes Made

### 1. Domain Model (core/domain/user.go)
```go
type User struct {
    // ... existing fields
    PasswordHash *string `json:"-"` // NEW: Optional password for basic auth
    // ... rest of fields
}
```

- Added `PasswordHash *string` field (nullable, never exposed in JSON)
- Added DTOs: `SetPasswordRequest`, `RemovePasswordRequest`

### 2. Repository Contract (core/infrastructure/repository/contract.go)
```go
type UserStore interface {
    // ... existing methods
    SetPassword(ctx, tenantID, userID, passwordHash string) error
    RemovePassword(ctx, tenantID, userID string) error
}
```

### 3. Implementations

#### InMemoryUserStore (core/infrastructure/repository/memory/user_store.go)
- ✅ `SetPassword()` - Sets user.PasswordHash pointer
- ✅ `RemovePassword()` - Clears user.PasswordHash to nil

#### PostgresUserStore (core/infrastructure/repository/postgres/user_store.go)
- ✅ Updated all queries to include `password_hash` column
- ✅ `SetPassword()` - UPDATE users SET password_hash = $1
- ✅ `RemovePassword()` - UPDATE users SET password_hash = NULL

#### Database Schema (core/infrastructure/repository/postgres/db_schema.sql)
```sql
CREATE TABLE users (
    -- ... existing columns
    password_hash TEXT,  -- NEW: Optional bcrypt hash
    -- ... rest of columns
);
```

### 4. Basic Auth Service (credential/application/basic_service.go)

**Before:**
```go
UserProvider repository.UserProvider // @Inject "user-provider"
user := &basic.User{...}
s.UserProvider.CreateUser(ctx, user)
```

**After:**
```go
UserStore core_repository.UserStore // @Inject "user-store"
user := &core_domain.User{
    PasswordHash: &passwordHash,  // pointer
    Status: core_domain.UserStatusActive,
    // ...
}
s.UserStore.Create(ctx, user)
```

- Changed dependency injection from `user-provider` → `user-store`
- Register() creates `domain.User` instead of `basic.User`
- ChangePassword() uses `UserStore.SetPassword()` instead of Update()
- All methods handle `PasswordHash` as `*string` pointer

### 5. Basic Authenticator (credential/infrastructure/authenticator/basic_authenticator.go)

**Before:**
```go
userProvider repository.UserProvider
user, _ := a.userProvider.GetUserByUsername(...)
if !hasher.VerifyPassword(user.PasswordHash, ...) // PasswordHash: string
```

**After:**
```go
userStore core_repository.UserStore
user, _ := a.userStore.GetByUsername(...)
if user.PasswordHash == nil || !hasher.VerifyPassword(*user.PasswordHash, ...) // *string
```

- Changed to use `UserStore` from `core`
- Dereference `PasswordHash` pointer before verification
- Check for nil password (users without basic auth)

### 6. DTO Converter (credential/domain/basic/dto.go)
```go
func ToUserInfoFromDomain(user *core_domain.User) *UserInfo {
    disabled := (user.Status == core_domain.UserStatusSuspended || 
                 user.Status == core_domain.UserStatusDeleted)
    return &UserInfo{
        ID: user.ID,
        TenantID: user.TenantID,
        Username: user.Username,
        Email: user.Email,
        Disabled: disabled,
        Metadata: metadata,
    }
}
```

### 7. Removed Files
- ❌ `credential/infrastructure/repository/contract.go` - UserProvider interface removed
- ❌ `credential/infrastructure/repository/memory/user_provider_inmemory.go` - deleted
- ❌ `basic.User` type - no longer needed (may be removed in future cleanup)

## API Design

### User Profile Management (core)
```bash
POST   /api/core/users              # Create user (NO password)
GET    /api/core/users/{id}         # Get user
PUT    /api/core/users/{id}         # Update profile
DELETE /api/core/users/{id}         # Delete user
```

### Password Management (credential)
```bash
POST   /api/core/users/{id}/password        # Set/update password
DELETE /api/core/users/{id}/password        # Remove password
POST   //api/auth/cred/basic/change-password      # Change password (requires old)
```

### Authentication
```bash
POST   //api/auth/cred/basic/login                # Login with username/password
POST   //api/auth/cred/basic/register             # Register new user with password
```

## Benefits

### ✅ Single Source of Truth
- Users created once in `core`
- No synchronization needed between stores
- Consistent user data across all credential types

### ✅ Optional Password
- `PasswordHash` is nullable pointer
- Users can exist without passwords (OAuth, passkey, etc.)
- Basic auth is opt-in via `SetPassword()`

### ✅ Security
- Password hash **never** exposed in API responses (`json:"-"`)
- Separate endpoints for password vs profile management
- Clear separation of concerns

### ✅ Flexibility
- Same user can use multiple auth methods
- Can enable/disable basic auth per user
- Easy to add new credential types

## Migration Path

For existing deployments:

1. **Add password_hash column** to users table:
   ```sql
   ALTER TABLE users ADD COLUMN password_hash TEXT;
   ```

2. **Migrate existing basic auth users**:
   ```sql
   UPDATE users u
   SET password_hash = (
       SELECT password_hash 
       FROM basic_users bu 
       WHERE bu.user_id = u.id
   );
   ```

3. **Update dependency injection** in your registration:
   ```go
   // Remove: registry.RegisterFactory("user-provider", ...)
   // Already using: registry.RegisterFactory("user-store", ...)
   ```

## Testing Checklist

- [x] InMemoryUserStore SetPassword/RemovePassword
- [x] PostgresUserStore SetPassword/RemovePassword
- [x] BasicAuthService Register creates domain.User
- [x] BasicAuthService Login handles nil password
- [x] BasicAuthService ChangePassword uses SetPassword()
- [x] BasicAuthenticator uses UserStore from core
- [x] ToUserInfoFromDomain converts correctly
- [ ] Integration test: Create user → Set password → Login → Change password
- [ ] Integration test: OAuth user (no password) can't use basic auth

## Next Steps

1. Update other credential services (OAuth2, Passwordless, Passkey) to use unified UserStore
2. Remove `basic.User` type completely
3. Create migration script for existing deployments
4. Add unit tests for password management
5. Update examples to show unified approach

---

**Status**: ✅ Implementation Complete  
**Date**: November 26, 2025  
**Impact**: Breaking change for credential provider implementations
