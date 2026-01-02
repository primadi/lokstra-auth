# User Storage Unification

## Problem Statement

Previously, programmers had to create users twice:
1. Once in `core` for user profile management
2. Once in `credential` for basic authentication

This duplication led to:
- **Synchronization issues** - user data could become inconsistent
- **Complex API** - two separate endpoints for user creation
- **Poor developer experience** - confusing mental model
- **Maintenance burden** - duplicate code and interfaces

## Solution: Unified User Storage

We unified user storage into a single source of truth in `core/domain.User` with optional password support.

### Key Design Decisions

1. **Single User Entity** (`core/domain.User`)
   - All user data lives in one place
   - `PasswordHash` is an optional field (`*string`)
   - Users without passwords can use OAuth2, Passkey, Passwordless authentication

2. **Password as Optional Capability**
   - `PasswordHash *string` - nullable pointer
   - `json:"-"` tag ensures password hash is NEVER exposed in API responses
   - Separate password management methods in UserStore

3. **Separation of Concerns via API Design**
   - User profile management: `POST /api/core/users`
   - Password management: `POST /api/core/users/{id}/password`
   - Authentication: `POST //api/auth/cred/basic/login`

4. **Repository Pattern**
   - `UserStore.SetPassword(ctx, tenantID, userID, passwordHash)` - Add/update password
   - `UserStore.RemovePassword(ctx, tenantID, userID)` - Disable basic auth for user

## Implementation Details

### Domain Model Changes

**File**: `core/domain/user.go`

```go
type User struct {
    ID           string          `json:"id"`
    TenantID     string          `json:"tenant_id"`
    Username     string          `json:"username"`
    Email        string          `json:"email"`
    PasswordHash *string         `json:"-"` // NEW: Optional, never exposed in JSON
    Status       UserStatus      `json:"status"`
    Metadata     *map[string]any `json:"metadata,omitempty"`
    // ... other fields
}

// DTOs for password management
type SetPasswordRequest struct {
    TenantID string `json:"tenant_id"`
    UserID   string `json:"user_id"`
    Password string `json:"password"`
}

type RemovePasswordRequest struct {
    TenantID string `json:"tenant_id"`
    UserID   string `json:"user_id"`
}
```

### Repository Contract Changes

**File**: `core/infrastructure/repository/contract.go`

```go
type UserStore interface {
    // Standard CRUD operations
    Create(ctx context.Context, user *domain.User) error
    Get(ctx context.Context, tenantID, userID string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, tenantID, userID string) error
    
    // Query methods
    List(ctx context.Context, tenantID string) ([]*domain.User, error)
    GetByUsername(ctx context.Context, tenantID, username string) (*domain.User, error)
    GetByEmail(ctx context.Context, tenantID, email string) (*domain.User, error)
    
    // NEW: Password management
    SetPassword(ctx context.Context, tenantID, userID, passwordHash string) error
    RemovePassword(ctx context.Context, tenantID, userID string) error
}
```

### In-Memory Implementation

**File**: `core/infrastructure/repository/memory/user_store.go`

```go
func (s *InMemoryUserStore) SetPassword(
    ctx context.Context, 
    tenantID, userID, passwordHash string,
) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    user, err := s.getUnlocked(tenantID, userID)
    if err != nil {
        return err
    }

    // Set the password hash pointer
    user.PasswordHash = &passwordHash
    return nil
}

func (s *InMemoryUserStore) RemovePassword(
    ctx context.Context, 
    tenantID, userID string,
) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    user, err := s.getUnlocked(tenantID, userID)
    if err != nil {
        return err
    }

    // Clear the password hash pointer (disables basic auth)
    user.PasswordHash = nil
    return nil
}
```

### Service Layer Changes

**File**: `credential/application/basic_service.go`

#### Before (Dual Storage)
```go
type BasicAuthService struct {
    UserProvider repository.UserProvider // @Inject "user-provider"
    // ...
}

func (s *BasicAuthService) Register(...) (*basic.RegisterResponse, error) {
    // Create user in basic.User format
    user := &basic.User{
        ID:           utils.GenerateID("usr"),
        PasswordHash: passwordHash, // string, always required
        // ...
    }
    s.UserProvider.CreateUser(ctx, user)
}
```

#### After (Unified Storage)
```go
type BasicAuthService struct {
    UserStore core_repository.UserStore // @Inject "user-store"
    // ...
}

func (s *BasicAuthService) Register(...) (*basic.RegisterResponse, error) {
    // Hash password
    passwordHash, _ := hasher.HashPassword(req.Password)
    
    // Create user in domain.User format
    user := &core_domain.User{
        ID:           utils.GenerateID("usr"),
        TenantID:     req.TenantID,
        Username:     req.Username,
        Email:        req.Email,
        PasswordHash: &passwordHash, // *string pointer
        Status:       core_domain.UserStatusActive,
        Metadata:     &metadata,
    }
    
    s.UserStore.Create(ctx, user)
}

func (s *BasicAuthService) ChangePassword(...) error {
    user, _ := s.UserStore.Get(ctx, req.TenantID, req.UserID)
    
    // Check if user has password
    if user.PasswordHash == nil {
        return fmt.Errorf("user does not have a password set")
    }
    
    // Verify old password (dereference pointer)
    if !hasher.VerifyPassword(*user.PasswordHash, req.OldPassword) {
        return fmt.Errorf("invalid old password")
    }
    
    // Hash new password
    newHash, _ := hasher.HashPassword(req.NewPassword)
    
    // Use dedicated method instead of Update
    return s.UserStore.SetPassword(ctx, req.TenantID, req.UserID, newHash)
}
```

### Authenticator Changes

**File**: `credential/infrastructure/authenticator/basic_authenticator.go`

```go
type BasicAuthenticator struct {
    userStore core_repository.UserStore
}

func (a *BasicAuthenticator) Authenticate(...) (*domain.AuthenticationResult, error) {
    // Get user from unified store
    user, err := a.userStore.GetByUsername(ctx, authCtx.TenantID, basicCreds.Username)
    if err != nil {
        return &domain.AuthenticationResult{Success: false, Error: basic.ErrAuthenticationFailed}, nil
    }

    // Check if basic auth is enabled for this user
    if user.PasswordHash == nil {
        return &domain.AuthenticationResult{Success: false, Error: basic.ErrAuthenticationFailed}, nil
    }

    // Check user status
    if user.Status != core_domain.UserStatusActive {
        return &domain.AuthenticationResult{Success: false, Error: basic.ErrAuthenticationFailed}, nil
    }

    // Verify password (dereference pointer)
    if !hasher.VerifyPassword(*user.PasswordHash, basicCreds.Password) {
        return &domain.AuthenticationResult{Success: false, Error: basic.ErrAuthenticationFailed}, nil
    }

    // Build claims with user metadata
    claims := map[string]any{
        "sub":       user.ID,
        "tenant_id": user.TenantID,
        "username":  user.Username,
        "email":     user.Email,
        "auth_type": "basic",
    }
    
    if user.Metadata != nil {
        maps.Copy(claims, *user.Metadata)
    }

    return &domain.AuthenticationResult{Success: true, Subject: user.ID, Claims: claims}, nil
}
```

### DTO Conversion

**File**: `credential/domain/basic/dto.go`

```go
// ToUserInfoFromDomain converts domain.User (from core) to UserInfo
func ToUserInfoFromDomain(user *core_domain.User) *UserInfo {
    if user == nil {
        return nil
    }
    
    metadata := make(map[string]any)
    if user.Metadata != nil {
        metadata = *user.Metadata
    }
    
    disabled := (user.Status == core_domain.UserStatusSuspended || 
                 user.Status == core_domain.UserStatusDeleted)
    
    return &UserInfo{
        ID:       user.ID,
        TenantID: user.TenantID,
        Username: user.Username,
        Email:    user.Email,
        Disabled: disabled,
        Metadata: metadata,
    }
}
```

## API Design Philosophy

### User Management (core)

```http
# Create user profile (NO password required)
POST /api/core/users
{
  "tenant_id": "tenant01",
  "username": "john.doe",
  "email": "john@example.com",
  "metadata": {}
}

# Enable basic authentication for user
POST /api/core/users/{user_id}/password
{
  "tenant_id": "tenant01",
  "password": "SecureP@ss123"
}

# Disable basic authentication
DELETE /api/core/users/{user_id}/password
{
  "tenant_id": "tenant01"
}
```

### Authentication (credential)

```http
# Register with password (creates user + sets password in one call)
POST //api/auth/cred/basic/register
{
  "tenant_id": "tenant01",
  "username": "john.doe",
  "password": "SecureP@ss123",
  "email": "john@example.com"
}

# Login
POST //api/auth/cred/basic/login
{
  "tenant_id": "tenant01",
  "username": "john.doe",
  "password": "SecureP@ss123"
}

# Change password
POST //api/auth/cred/basic/change-password
{
  "tenant_id": "tenant01",
  "user_id": "usr_xxx",
  "old_password": "SecureP@ss123",
  "new_password": "NewSecureP@ss456"
}
```

## Benefits

1. **No Duplication**
   - Single `domain.User` entity across entire system
   - No sync issues between storage layers

2. **Flexible Authentication**
   - Users can have multiple auth methods
   - Password is optional - supports OAuth, Passkey, Passwordless
   - Easy to add/remove basic auth for existing users

3. **Security by Design**
   - `PasswordHash` never exposed in JSON responses (`json:"-"`)
   - Pointer type (`*string`) makes nullable intent explicit
   - Separate password management endpoints

4. **Better Developer Experience**
   - Clear mental model: one user entity
   - Intuitive API: profile vs password management
   - Less code to maintain

5. **Type Safety**
   - Compiler enforces nil checks for optional password
   - Prevents accidental password exposure
   - Clear ownership: core owns user data

## Migration Path

If you have existing code using `UserProvider` from `credential`:

1. **Replace dependency injection**:
   ```go
   // OLD
   UserProvider repository.UserProvider // @Inject "user-provider"
   
   // NEW
   UserStore core_repository.UserStore // @Inject "user-store"
   ```

2. **Update method calls**:
   ```go
   // OLD
   user, err := s.UserProvider.GetUserByUsername(ctx, tenantID, username)
   
   // NEW
   user, err := s.UserStore.GetByUsername(ctx, tenantID, username)
   ```

3. **Handle pointer fields**:
   ```go
   // OLD
   hasher.VerifyPassword(user.PasswordHash, password)
   
   // NEW
   if user.PasswordHash != nil {
       hasher.VerifyPassword(*user.PasswordHash, password)
   }
   ```

4. **Use domain.User instead of basic.User**:
   ```go
   // OLD
   user := &basic.User{PasswordHash: hash}
   
   // NEW
   user := &core_domain.User{PasswordHash: &hash}
   ```

## Future Work

- [ ] Implement `PostgresUserStore` with SetPassword/RemovePassword
- [ ] Add password history tracking (prevent reuse)
- [ ] Add multi-factor authentication support
- [ ] Consider password expiration policies
- [ ] Add audit logging for password changes

## Files Modified

- `core/domain/user.go` - Added PasswordHash field and DTOs
- `core/infrastructure/repository/contract.go` - Added SetPassword/RemovePassword methods
- `core/infrastructure/repository/memory/user_store.go` - Implemented password methods
- `credential/infrastructure/repository/contract.go` - Removed UserProvider interface
- `credential/application/basic_service.go` - Migrated to use UserStore
- `credential/infrastructure/authenticator/basic_authenticator.go` - Updated to use UserStore
- `credential/domain/basic/dto.go` - Added ToUserInfoFromDomain converter

## Conclusion

This unification eliminates a major pain point in the authentication system. Developers now have a single, clear place to manage user data, with password support being a simple optional capability. The design is more maintainable, more secure, and provides a better developer experience.
