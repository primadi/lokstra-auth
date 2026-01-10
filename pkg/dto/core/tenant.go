package core

import (
	"time"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
)

// CreateTenantRequest request to create a tenant with auto-owner creation
type CreateTenantRequest struct {
	ID       string                     `json:"id" validate:"required"`
	Name     string                     `json:"name" validate:"required"`
	AppID    string                     `json:"app_id,omitempty"`              // Optional app ID for the default admin app
	DBDsn    string                     `json:"db_dsn" validate:"required"`    // Database connection string
	DBSchema string                     `json:"db_schema" validate:"required"` // Database schema name
	Settings *domaincore.TenantSettings `json:"settings,omitempty"`
	Metadata *map[string]any            `json:"metadata,omitempty"`

	// Owner information - REQUIRED
	// Will automatically create owner user if doesn't exist
	OwnerEmail    string `json:"owner_email" validate:"required,email"` // Owner's email (will create user if not exists)
	OwnerUsername string `json:"owner_username,omitempty"`              // Optional, defaults to email prefix
	OwnerFullName string `json:"owner_full_name,omitempty"`             // Optional owner full name

	// SendWelcomeEmail sends password reset email to owner after creation
	SendWelcomeEmail bool `json:"send_welcome_email"` // Default: true
}

// GetTenantRequest request to get a tenant
type GetTenantRequest struct {
	ID string `path:"id" validate:"required"`
}

// UpdateTenantRequest request to update a tenant
type UpdateTenantRequest struct {
	ID       string                     `path:"id" validate:"required"`
	Name     string                     `json:"name"`
	DBDsn    string                     `json:"db_dsn"`    // Can update database connection
	DBSchema string                     `json:"db_schema"` // Can update schema name
	Settings *domaincore.TenantSettings `json:"settings,omitempty"`
	Config   *domaincore.TenantConfig   `json:"config,omitempty"`
	Metadata *map[string]any            `json:"metadata,omitempty"`
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

// TransferOwnershipRequest request to transfer tenant ownership
// Only current owner or platform admin can transfer ownership
type TransferOwnershipRequest struct {
	TenantID    string `path:"tenant_id" validate:"required"`
	NewOwnerID  string `json:"new_owner_id" validate:"required"` // User ID of new owner (must exist in tenant)
	CurrentUser string `json:"-"`                                // Injected by middleware (must be current owner or platform admin)
}

// TenantOwnershipHistory tracks ownership transfers for audit
type TenantOwnershipHistory struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	PreviousOwner string    `json:"previous_owner"`
	NewOwner      string    `json:"new_owner"`
	TransferredBy string    `json:"transferred_by"` // User who initiated transfer (owner or admin)
	Reason        string    `json:"reason"`
	TransferredAt time.Time `json:"transferred_at"`
}
