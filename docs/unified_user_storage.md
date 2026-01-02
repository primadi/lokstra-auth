# Unified User Storage - Developer Guide

## Overview

Lokstra-auth now uses a **single UserStore** for all authentication methods. Users are created once in `core`, and credentials (passwords, OAuth tokens, passkeys) are managed separately.

## Key Concept

```
ONE User Profile (core/domain.User)
  ├─ Optional: PasswordHash (basic auth)
  ├─ Optional: OAuth2 tokens
  ├─ Optional: Passkey credentials
  └─ Optional: Passwordless tokens
```

## Domain Model

```go
// core/domain/user.go
type User struct {
    ID           string
    TenantID     string
    Username     string
    Email        string
    FullName     string
    PasswordHash *string        // NEW: Optional password (nil = no basic auth)
    Status       UserStatus     // active, suspended, deleted
    Metadata     *map[string]any
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    *time.Time
}
```

**Important:** `PasswordHash` is:
- A **pointer** (`*string`) - can be nil for users without passwords
- **Never exposed** in JSON responses (`json:"-"` tag)
- Only used for basic authentication

## UserStore Interface

```go
// core/infrastructure/repository/contract.go
type UserStore interface {
    // Core CRUD
    Create(ctx context.Context, user *User) error
    Get(ctx context.Context, tenantID, userID string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, tenantID, userID string) error
    List(ctx context.Context, tenantID string) ([]*User, error)
    
    // Lookup methods
    GetByUsername(ctx context.Context, tenantID, username string) (*User, error)
    GetByEmail(ctx context.Context, tenantID, email string) (*User, error)
    Exists(ctx context.Context, tenantID, userID string) (bool, error)
    ListByApp(ctx context.Context, tenantID, appID string) ([]*User, error)
    
    // Password management (NEW)
    SetPassword(ctx context.Context, tenantID, userID, passwordHash string) error
    RemovePassword(ctx context.Context, tenantID, userID string) error
}
```

## Usage Examples

### 1. Create User Without Password

```go
// For OAuth, Passkey, or Passwordless users
user := &domain.User{
    ID:       utils.GenerateID("usr"),
    TenantID: tenantID,
    Username: "john.doe",
    Email:    "john@example.com",
    FullName: "John Doe",
    Status:   domain.UserStatusActive,
    // PasswordHash: nil (default)
}

err := userStore.Create(ctx, user)
```

### 2. Create User With Password (Basic Auth)

```go
// Hash the password
passwordHash, err := hasher.HashPassword(plainPassword)
if err != nil {
    return err
}

// Create user with password
user := &domain.User{
    ID:           utils.GenerateID("usr"),
    TenantID:     tenantID,
    Username:     "jane.doe",
    Email:        "jane@example.com",
    PasswordHash: &passwordHash,  // Set password pointer
    Status:       domain.UserStatusActive,
}

err = userStore.Create(ctx, user)
```

### 3. Set Password for Existing User

```go
// Enable basic auth for OAuth user
passwordHash, _ := hasher.HashPassword(newPassword)
err := userStore.SetPassword(ctx, tenantID, userID, passwordHash)
```

### 4. Remove Password (Disable Basic Auth)

```go
// Remove password, user can still use OAuth/Passkey
err := userStore.RemovePassword(ctx, tenantID, userID)
```

### 5. Verify Password During Login

```go
user, err := userStore.GetByUsername(ctx, tenantID, username)
if err != nil {
    return ErrUserNotFound
}

// Check if user has password set
if user.PasswordHash == nil {
    return errors.New("basic auth not enabled for this user")
}

// Verify password (dereference pointer)
if !hasher.VerifyPassword(*user.PasswordHash, plainPassword) {
    return ErrInvalidCredentials
}
```

### 6. Change Password

```go
// Get user
user, err := userStore.Get(ctx, tenantID, userID)
if err != nil {
    return err
}

// Verify old password
if user.PasswordHash == nil || !hasher.VerifyPassword(*user.PasswordHash, oldPassword) {
    return ErrInvalidOldPassword
}

// Hash new password
newHash, err := hasher.HashPassword(newPassword)
if err != nil {
    return err
}

// Update using SetPassword (better than Update)
err = userStore.SetPassword(ctx, tenantID, userID, newHash)
```

## Dependency Injection

### Service Registration

```go
// core/application/register.go
func Register(registry *lokstra.ServiceRegistry) {
    // Use InMemory or PostgreSQL implementation
    registry.RegisterFactory("user-store", func(cfg map[string]any) any {
        return memory.NewUserStore(cfg)
        // OR: return postgres.NewUserStore(cfg)
    })
}
```

### Service Injection

```go
// credential/application/basic_service.go
type BasicAuthService struct {
    // @Inject "user-store"
    UserStore core_repository.UserStore
    
    // ... other dependencies
}
```

## API Endpoints

### User Profile Management

```bash
# Create user (without password)
POST /api/core/users
{
  "username": "john.doe",
  "email": "john@example.com",
  "full_name": "John Doe"
}

# Get user
GET /api/core/users/{userId}

# Update profile (NOT password)
PUT /api/core/users/{userId}
{
  "full_name": "John Smith"
}
```

### Password Management

```bash
# Set/update password
POST /api/core/users/{userId}/password
{
  "password": "newSecurePassword123"
}

# Remove password (disable basic auth)
DELETE /api/core/users/{userId}/password

# Change password (requires old password)
POST //api/auth/cred/basic/change-password
{
  "user_id": "usr_123",
  "old_password": "oldPassword",
  "new_password": "newPassword"
}
```

### Authentication

```bash
# Register with password
POST //api/auth/cred/basic/register
{
  "username": "john.doe",
  "password": "securePassword123",
  "email": "john@example.com"
}

# Login with password
POST //api/auth/cred/basic/login
{
  "username": "john.doe",
  "password": "securePassword123"
}
```

## Database Schema

### PostgreSQL

```sql
CREATE TABLE users (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    password_hash TEXT,              -- NEW: Optional password
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    PRIMARY KEY (tenant_id, id),
    UNIQUE (tenant_id, username),
    UNIQUE (tenant_id, email)
);

COMMENT ON COLUMN users.password_hash IS 
  'Optional bcrypt hash for basic auth (NULL if OAuth/passkey only)';
```

### Migration from Separate UserProvider

```sql
-- Add password_hash column
ALTER TABLE users ADD COLUMN password_hash TEXT;

-- Migrate existing passwords (if you had separate table)
UPDATE users u
SET password_hash = (
    SELECT password_hash 
    FROM basic_auth_users b 
    WHERE b.user_id = u.id
)
WHERE EXISTS (
    SELECT 1 FROM basic_auth_users b WHERE b.user_id = u.id
);
```

## Best Practices

### ✅ DO

- Use `UserStore.SetPassword()` to manage passwords
- Check for `nil` before dereferencing `PasswordHash`
- Use separate endpoints for profile vs password management
- Allow users to exist without passwords (OAuth, passkey users)
- Validate password complexity before hashing

### ❌ DON'T

- Don't expose `PasswordHash` in API responses
- Don't update password via `UserStore.Update()` (use `SetPassword()`)
- Don't assume all users have passwords
- Don't require password for user creation
- Don't forget to dereference `*string` when verifying

## Common Patterns

### Enable Basic Auth for OAuth User

```go
// User already exists (created via OAuth)
user, _ := userStore.Get(ctx, tenantID, userID)

// Now enable password login
passwordHash, _ := hasher.HashPassword(password)
err := userStore.SetPassword(ctx, tenantID, userID, passwordHash)
```

### Multi-Factor Authentication

```go
// User can have password AND OAuth AND passkey
user := &domain.User{
    ID:           "usr_123",
    Username:     "john.doe",
    PasswordHash: &passwordHash,  // Basic auth enabled
    // OAuth tokens stored in separate oauth2 module
    // Passkeys stored in separate passkey module
}
```

### Disable Basic Auth

```go
// Remove password, force OAuth/passkey only
err := userStore.RemovePassword(ctx, tenantID, userID)
```

## Error Handling

```go
// Check if user has password
user, _ := userStore.GetByUsername(ctx, tenantID, username)
if user.PasswordHash == nil {
    return &LoginResponse{
        Success: false,
        Error:   "basic authentication not enabled, please use OAuth",
    }
}

// Safe password verification
if user.PasswordHash == nil || !hasher.VerifyPassword(*user.PasswordHash, password) {
    return &LoginResponse{
        Success: false,
        Error:   "invalid credentials",
    }
}
```

## Testing

### Unit Test Example

```go
func TestSetPassword(t *testing.T) {
    store := memory.NewUserStore(nil)
    ctx := context.Background()
    
    // Create user without password
    user := &domain.User{
        ID:       "usr_1",
        TenantID: "tenant_1",
        Username: "test",
    }
    store.Create(ctx, user)
    
    // Set password
    hash := "hashed_password"
    err := store.SetPassword(ctx, "tenant_1", "usr_1", hash)
    assert.NoError(t, err)
    
    // Verify password was set
    retrieved, _ := store.Get(ctx, "tenant_1", "usr_1")
    assert.NotNil(t, retrieved.PasswordHash)
    assert.Equal(t, hash, *retrieved.PasswordHash)
    
    // Remove password
    err = store.RemovePassword(ctx, "tenant_1", "usr_1")
    assert.NoError(t, err)
    
    // Verify password was removed
    retrieved, _ = store.Get(ctx, "tenant_1", "usr_1")
    assert.Nil(t, retrieved.PasswordHash)
}
```

## Troubleshooting

### "cannot use string as *string"
```go
// ❌ Wrong
user.PasswordHash = passwordHash  // passwordHash is string

// ✅ Correct
user.PasswordHash = &passwordHash  // Take address
```

### "invalid memory address or nil pointer dereference"
```go
// ❌ Wrong
if hasher.VerifyPassword(user.PasswordHash, password) {  // Panic if nil!

// ✅ Correct
if user.PasswordHash != nil && hasher.VerifyPassword(*user.PasswordHash, password) {
```

### "user already exists" during migration
- Use `SetPassword()` instead of creating duplicate user
- Update existing user record, don't create new one

---

**Related Documents:**
- [Architecture Overview](./architecture.md)
- [Multi-Tenant Guide](./multi_tenant_architecture.md)
- [Credential Providers](./credential_providers.md)
