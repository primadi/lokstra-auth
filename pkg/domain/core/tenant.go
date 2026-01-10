package core

import (
	"errors"
	"time"
)

var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantAlreadyExists = errors.New("tenant already exists")
	ErrTenantSuspended     = errors.New("tenant is suspended")
	ErrTenantDeleted       = errors.New("tenant is deleted")
	ErrInvalidTenantID     = errors.New("invalid tenant ID")
)

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// Tenant represents an organization/company using the system
type Tenant struct {
	ID        string          `json:"id"`        // Unique identifier (e.g., "acme-corp")
	Name      string          `json:"name"`      // Display name (e.g., "Acme Corporation")
	OwnerID   *string         `json:"owner_id"`  // User ID of tenant owner (billing, legal owner) - REQUIRED
	Domain    string          `json:"domain"`    // Optional custom domain
	DBDsn     string          `json:"db_dsn"`    // Database connection string (required for multi-database tenancy)
	DBSchema  string          `json:"db_schema"` // Database schema name (required for schema-based tenancy)
	Status    TenantStatus    `json:"status"`    // active, suspended, deleted
	Config    *TenantConfig   `json:"config"`    // Tenant-wide configuration (credentials, tokens, security)
	Settings  *TenantSettings `json:"settings"`  // Tenant-wide settings (legacy, will be merged into Config)
	Metadata  *map[string]any `json:"metadata"`  // Custom metadata
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at,omitempty"`
}

// TenantSettings holds tenant-wide configuration
type TenantSettings struct {
	MaxUsers           int            `json:"max_users"` // Quota limits
	MaxApps            int            `json:"max_apps"`
	AllowedAuthMethods []string       `json:"allowed_auth_methods"` // basic, oauth2, apikey, etc.
	PasswordPolicy     PasswordPolicy `json:"password_policy"`
	SessionTimeout     time.Duration  `json:"session_timeout"`
	RequireMFA         bool           `json:"require_mfa"`
	AccountLockout     AccountLockout `json:"account_lockout"` // Failed login attempt protection
	CustomSettings     map[string]any `json:"custom_settings,omitempty"`
}

// PasswordPolicy defines password requirements for a tenant
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSpecial   bool `json:"require_special"`
	MaxAge           int  `json:"max_age"` // Days before password expires
}

// AccountLockout defines account lockout policy for failed login attempts
type AccountLockout struct {
	Enabled            bool          `json:"enabled"`              // Enable account lockout
	MaxAttempts        int           `json:"max_attempts"`         // Max failed attempts before lockout (default: 5)
	LockoutDuration    time.Duration `json:"lockout_duration"`     // How long account is locked (default: 15 minutes)
	ResetAttemptsAfter time.Duration `json:"reset_attempts_after"` // Reset attempt counter after this duration (default: 1 hour)
	PermanentLockAfter int           `json:"permanent_lock_after"` // Permanent lock after N lockouts (0 = disabled)
	NotifyOnLockout    bool          `json:"notify_on_lockout"`    // Send email notification on lockout
	NotifyOnUnlock     bool          `json:"notify_on_unlock"`     // Send email notification on unlock
}

// IsActive checks if tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive
}

// IsSuspended checks if tenant is suspended
func (t *Tenant) IsSuspended() bool {
	return t.Status == TenantStatusSuspended
}

// IsDeleted checks if tenant is deleted
func (t *Tenant) IsDeleted() bool {
	return t.Status == TenantStatusDeleted || t.DeletedAt != nil
}

// Validate validates tenant data
func (t *Tenant) Validate() error {
	if t.ID == "" {
		return ErrInvalidTenantID
	}
	if t.Name == "" {
		return errors.New("tenant name is required")
	}
	if t.OwnerID == nil || *t.OwnerID == "" {
		return errors.New("owner_id is required - every tenant must have an owner")
	}
	if t.DBDsn == "" {
		return errors.New("database DSN is required")
	}
	if t.DBSchema == "" {
		return errors.New("database schema is required")
	}
	if t.Status == "" {
		t.Status = TenantStatusActive
	}
	return nil
}
