package domain

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

// TenantFilters for querying tenants
type TenantFilters struct {
	Status TenantStatus
	Domain string
	Limit  int
	Offset int
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

// =============================================================================
// Tenant DTOs
// =============================================================================

// CreateTenantRequest request to create a tenant
type CreateTenantRequest struct {
	ID       string          `json:"id" validate:"required"`
	Name     string          `json:"name" validate:"required"`
	DBDsn    string          `json:"db_dsn" validate:"required"`    // Database connection string
	DBSchema string          `json:"db_schema" validate:"required"` // Database schema name
	Settings *TenantSettings `json:"settings,omitempty"`
	Metadata *map[string]any `json:"metadata,omitempty"`
}

// GetTenantRequest request to get a tenant
type GetTenantRequest struct {
	ID string `path:"id" validate:"required"`
}

// UpdateTenantRequest request to update a tenant
type UpdateTenantRequest struct {
	ID       string          `path:"id" validate:"required"`
	Name     string          `json:"name"`
	DBDsn    string          `json:"db_dsn"`    // Can update database connection
	DBSchema string          `json:"db_schema"` // Can update schema name
	Settings *TenantSettings `json:"settings,omitempty"`
	Config   *TenantConfig   `json:"config,omitempty"`
	Metadata *map[string]any `json:"metadata,omitempty"`
}

// DeleteTenantRequest request to delete a tenant
type DeleteTenantRequest struct {
	ID string `path:"id" validate:"required"`
}

// ListTenantsRequest request to list tenants
type ListTenantsRequest struct {
}

// ActivateTenantRequest request to activate a tenant
type ActivateTenantRequest struct {
	ID string `path:"id" validate:"required"`
}

// SuspendTenantRequest request to suspend a tenant
type SuspendTenantRequest struct {
	ID string `path:"id" validate:"required"`
}
