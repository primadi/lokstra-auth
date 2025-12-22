package domain

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserSuspended     = errors.New("user is suspended")
	ErrUserDeleted       = errors.New("user is deleted")
	ErrInvalidUserID     = errors.New("invalid user ID")
	ErrDuplicateUsername = errors.New("username already exists in this tenant")
	ErrDuplicateEmail    = errors.New("email already exists in this tenant")
)

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
	UserStatusLocked    UserStatus = "locked" // Account locked due to failed login attempts
)

// User represents a user within a tenant
type User struct {
	ID            string     `json:"id"`        // Unique identifier (UUID)
	TenantID      string     `json:"tenant_id"` // Belongs to tenant
	Username      string     `json:"username"`  // Unique within tenant
	Email         string     `json:"email"`     // Unique within tenant
	FullName      string     `json:"full_name"`
	IsTenantOwner bool       `json:"is_tenant_owner"` // True if this user owns the tenant (only 1 per tenant)
	PasswordHash  *string    `json:"-"`               // Bcrypt hash (for basic auth only, never exposed in JSON)
	Status        UserStatus `json:"status"`

	// Account lockout tracking
	FailedLoginAttempts int        `json:"failed_login_attempts"` // Counter for failed login attempts
	LastFailedLoginAt   *time.Time `json:"last_failed_login_at"`  // Timestamp of last failed login
	LockedAt            *time.Time `json:"locked_at"`             // When account was locked
	LockedUntil         *time.Time `json:"locked_until"`          // When account will be auto-unlocked
	LockoutCount        int        `json:"lockout_count"`         // How many times account has been locked

	Metadata  *map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at,omitempty"`
}

// UserAppStatus represents the status of user access to an app
type UserAppStatus string

const (
	UserAppStatusActive  UserAppStatus = "active"
	UserAppStatusRevoked UserAppStatus = "revoked"
)

// UserApp represents user's access to an application
// This is a simple access control: user X can access app Y
// Authorization (roles, permissions) is handled separately in authz layer
type UserApp struct {
	UserID    string        `json:"user_id"`
	TenantID  string        `json:"tenant_id"`
	AppID     string        `json:"app_id"`
	Status    UserAppStatus `json:"status"` // active, revoked
	GrantedAt time.Time     `json:"granted_at"`
	RevokedAt *time.Time    `json:"revoked_at,omitempty"`
}

// UserFilters for querying users
type UserFilters struct {
	TenantID string
	Status   UserStatus
	Limit    int
	Offset   int
}

// UserAppFilters for querying user-app associations
type UserAppFilters struct {
	TenantID string
	AppID    string
	UserID   string
	Status   UserAppStatus
	Limit    int
	Offset   int
}

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsSuspended checks if user is suspended
func (u *User) IsSuspended() bool {
	return u.Status == UserStatusSuspended
}

// IsDeleted checks if user is deleted
func (u *User) IsDeleted() bool {
	return u.Status == UserStatusDeleted || u.DeletedAt != nil
}

// IsLocked checks if user account is locked
func (u *User) IsLocked() bool {
	if u.Status == UserStatusLocked {
		return true
	}
	// Check if temporary lock is still active
	if u.LockedUntil != nil && time.Now().Before(*u.LockedUntil) {
		return true
	}
	return false
}

// ShouldAutoUnlock checks if account should be automatically unlocked
func (u *User) ShouldAutoUnlock() bool {
	return u.Status == UserStatusLocked &&
		u.LockedUntil != nil &&
		time.Now().After(*u.LockedUntil)
}

// RecordFailedLogin increments failed login counter
func (u *User) RecordFailedLogin() {
	now := time.Now()
	u.FailedLoginAttempts++
	u.LastFailedLoginAt = &now
	u.UpdatedAt = now
}

// ResetFailedLoginAttempts clears failed login counter after successful login
func (u *User) ResetFailedLoginAttempts() {
	u.FailedLoginAttempts = 0
	u.LastFailedLoginAt = nil
	u.UpdatedAt = time.Now()
}

// LockAccount locks the user account
func (u *User) LockAccount(duration time.Duration) {
	now := time.Now()
	until := now.Add(duration)

	u.Status = UserStatusLocked
	u.LockedAt = &now
	u.LockedUntil = &until
	u.LockoutCount++
	u.UpdatedAt = now
}

// UnlockAccount unlocks the user account
func (u *User) UnlockAccount() {
	u.Status = UserStatusActive
	u.LockedAt = nil
	u.LockedUntil = nil
	u.FailedLoginAttempts = 0
	u.LastFailedLoginAt = nil
	u.UpdatedAt = time.Now()
}

// Validate validates user data
func (u *User) Validate() error {
	if u.ID == "" {
		return ErrInvalidUserID
	}
	if u.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if u.Username == "" {
		return errors.New("username is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Status == "" {
		u.Status = UserStatusActive
	}
	return nil
}

// IsActive checks if user-app access is active
func (ua *UserApp) IsActive() bool {
	return ua.Status == UserAppStatusActive && ua.RevokedAt == nil
}

// IsRevoked checks if user-app access is revoked
func (ua *UserApp) IsRevoked() bool {
	return ua.Status == UserAppStatusRevoked || ua.RevokedAt != nil
}

// Validate validates user-app data
func (ua *UserApp) Validate() error {
	if ua.UserID == "" {
		return ErrInvalidUserID
	}
	if ua.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if ua.AppID == "" {
		return errors.New("app_id is required")
	}
	if ua.Status == "" {
		ua.Status = UserAppStatusActive
	}
	if ua.GrantedAt.IsZero() {
		ua.GrantedAt = time.Now()
	}
	return nil
}

// =============================================================================
// User DTOs
// =============================================================================

// CreateUserRequest request to create a user
type CreateUserRequest struct {
	TenantID string          `path:"tenant_id" validate:"required"`
	ID       string          `path:"id" validate:"required"`
	Username string          `json:"username" validate:"required"`
	Email    string          `json:"email" validate:"required,email"`
	Metadata *map[string]any `json:"metadata,omitempty"`
}

// GetUserRequest request to get a user
type GetUserRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// GetUserByUsernameRequest request to get a user by username
type GetUserByUsernameRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	Username string `path:"username" validate:"required"`
}

// GetUserByEmailRequest request to get a user by email
type GetUserByEmailRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	Email    string `path:"email" validate:"required"`
}

// UpdateUserRequest request to update a user
type UpdateUserRequest struct {
	TenantID string          `path:"tenant_id" validate:"required"`
	ID       string          `path:"id" validate:"required"`
	Username string          `json:"username"`
	Email    string          `json:"email" validate:"email"`
	Status   UserStatus      `json:"status"`
	Metadata *map[string]any `json:"metadata"`
}

// DeleteUserRequest request to delete a user
type DeleteUserRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// ListUsersRequest request to list users
type ListUsersRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
}

// ListUsersByAppRequest request to list users by app
type ListUsersByAppRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
}

// AssignUserToAppRequest request to assign user to app
type AssignUserToAppRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	UserID   string `path:"user_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
}

// RemoveUserFromAppRequest request to remove user from app
type RemoveUserFromAppRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	UserID   string `path:"user_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
}

// ActivateUserRequest request to activate a user
type ActivateUserRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// SuspendUserRequest request to suspend a user
type SuspendUserRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// SetPasswordRequest request to set or update user password (basic auth)
type SetPasswordRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	// OldPassword string `json:"old_password,omitempty"` // Required when updating existing password
}

// RemovePasswordRequest request to remove user password (disable basic auth)
type RemovePasswordRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}
