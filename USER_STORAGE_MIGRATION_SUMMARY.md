# User Storage Migration Summary

## What Changed?

**BEFORE**: Users created twice (once in core, once in credential)  
**AFTER**: Single user creation in core with optional password

## Quick Reference

### User Entity
```go
// core/domain/user.go
type User struct {
    ID           string          `json:"id"`
    TenantID     string          `json:"tenant_id"`
    Username     string          `json:"username"`
    Email        string          `json:"email"`
    PasswordHash *string         `json:"-"` // Optional, never exposed
    Status       UserStatus      `json:"status"`
    Metadata     *map[string]any `json:"metadata,omitempty"`
}
```

### Repository Methods (NEW)
```go
// core/infrastructure/repository/contract.go
type UserStore interface {
    // ... existing CRUD methods
    
    // Password management
    SetPassword(ctx, tenantID, userID, passwordHash string) error
    RemovePassword(ctx, tenantID, userID string) error
}
```

### Service Usage

#### Basic Auth Service
```go
// credential/application/basic_service.go

// Dependency injection changed
type BasicAuthService struct {
    UserStore core_repository.UserStore // @Inject "user-store"
    // ... other fields
}

// Register creates user with password
func (s *BasicAuthService) Register(...) {
    passwordHash, _ := hasher.HashPassword(req.Password)
    
    user := &core_domain.User{
        ID:           utils.GenerateID("usr"),
        PasswordHash: &passwordHash, // Pointer!
        Status:       core_domain.UserStatusActive,
        // ...
    }
    
    s.UserStore.Create(ctx, user)
}

// ChangePassword uses SetPassword method
func (s *BasicAuthService) ChangePassword(...) {
    user, _ := s.UserStore.Get(ctx, tenantID, userID)
    
    // Check nil pointer
    if user.PasswordHash == nil {
        return fmt.Errorf("no password set")
    }
    
    // Dereference pointer
    if !hasher.VerifyPassword(*user.PasswordHash, oldPassword) {
        return fmt.Errorf("invalid password")
    }
    
    // Use dedicated method
    newHash, _ := hasher.HashPassword(newPassword)
    return s.UserStore.SetPassword(ctx, tenantID, userID, newHash)
}
```

#### Authenticator
```go
// credential/infrastructure/authenticator/basic_authenticator.go

type BasicAuthenticator struct {
    userStore core_repository.UserStore
}

func (a *BasicAuthenticator) Authenticate(...) {
    user, _ := a.userStore.GetByUsername(ctx, tenantID, username)
    
    // Check if basic auth enabled
    if user.PasswordHash == nil {
        return AuthFailed
    }
    
    // Dereference pointer
    if !hasher.VerifyPassword(*user.PasswordHash, password) {
        return AuthFailed
    }
    
    // ...
}
```

## Key Points

1. **PasswordHash is `*string` (pointer)**
   - `nil` = user has no password (can use OAuth, passkey, etc.)
   - `&hash` = user has password (can use basic auth)

2. **Never exposed in JSON**
   - `json:"-"` tag prevents serialization
   - Security by design

3. **Separate password management**
   - `UserStore.SetPassword()` - add/update password
   - `UserStore.RemovePassword()` - disable basic auth
   - Don't use `UserStore.Update()` for password changes

4. **Type conversion**
   - Use `basic.ToUserInfoFromDomain(user)` to convert `*core_domain.User` to `*basic.UserInfo`

## Migration Checklist

- [x] Add `PasswordHash *string` to `domain.User`
- [x] Add `SetPassword`/`RemovePassword` to `UserStore` contract
- [x] Implement methods in `InMemoryUserStore`
- [x] Update `BasicAuthService` to use `UserStore`
- [x] Update `BasicAuthenticator` to use `UserStore`
- [x] Remove `UserProvider` interface
- [x] Add `ToUserInfoFromDomain` converter
- [ ] Implement methods in `PostgresUserStore` (when needed)

## Files Modified

- `core/domain/user.go`
- `core/infrastructure/repository/contract.go`
- `core/infrastructure/repository/memory/user_store.go`
- `credential/infrastructure/repository/contract.go`
- `credential/application/basic_service.go`
- `credential/infrastructure/authenticator/basic_authenticator.go`
- `credential/domain/basic/dto.go`

## Benefits

✅ No user duplication  
✅ Single source of truth  
✅ Password is optional capability  
✅ Supports multiple auth methods per user  
✅ Better security (json:"-")  
✅ Cleaner API design  

See `USER_STORAGE_UNIFICATION.md` for detailed documentation.
