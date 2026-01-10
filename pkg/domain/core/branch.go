package core

import (
	"errors"
	"time"
)

var (
	ErrBranchNotFound      = errors.New("branch not found")
	ErrBranchAlreadyExists = errors.New("branch already exists")
	ErrBranchDisabled      = errors.New("branch is disabled")
	ErrInvalidBranchID     = errors.New("invalid branch ID")
)

// BranchStatus represents the status of a branch
type BranchStatus string

const (
	BranchStatusActive   BranchStatus = "active"
	BranchStatusDisabled BranchStatus = "disabled"
)

// BranchType represents the type of branch
type BranchType string

const (
	BranchTypeHeadquarters BranchType = "headquarters" // Main office/HQ
	BranchTypeRegional     BranchType = "regional"     // Regional office
	BranchTypeStore        BranchType = "store"        // Retail store
	BranchTypeWarehouse    BranchType = "warehouse"    // Warehouse/Distribution center
	BranchTypeFranchise    BranchType = "franchise"    // Franchise outlet
	BranchTypeOffice       BranchType = "office"       // General office/branch office
	BranchTypeOther        BranchType = "other"        // Other types
)

// Branch represents a branch/location within a tenant's app
// Hierarchy: Tenant → App → Branch
//
// Use cases:
//   - Multi-store retail: Each store is a branch
//   - Multi-branch bank: Each office is a branch
//   - Multi-location company: Each location is a branch
//   - Multi-franchise: Each franchise outlet is a branch
//   - Multi-warehouse: Each warehouse is a branch
type Branch struct {
	ID        string          `json:"id"`        // Unique within tenant+app
	TenantID  string          `json:"tenant_id"` // Parent tenant
	AppID     string          `json:"app_id"`    // Parent app
	Name      string          `json:"name"`      // Display name
	Type      BranchType      `json:"type"`      // headquarters, regional, store, warehouse, etc.
	Status    BranchStatus    `json:"status"`    // active, disabled
	Address   *BranchAddress  `json:"address"`   // Physical address
	Contact   *BranchContact  `json:"contact"`   // Contact information
	Settings  *BranchSettings `json:"settings"`  // Branch-specific settings
	Metadata  *map[string]any `json:"metadata"`  // Custom metadata
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at,omitempty"`
}

// BranchAddress holds branch physical address
type BranchAddress struct {
	Street     string  `json:"street,omitempty"`
	City       string  `json:"city,omitempty"`
	State      string  `json:"state,omitempty"`
	Country    string  `json:"country,omitempty"`
	PostalCode string  `json:"postal_code,omitempty"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
}

// BranchContact holds branch contact information
type BranchContact struct {
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
	Fax   string `json:"fax,omitempty"`
}

// BranchSettings holds branch-specific configuration
type BranchSettings struct {
	Timezone       string            `json:"timezone"`            // e.g., "Asia/Jakarta"
	Currency       string            `json:"currency"`            // e.g., "IDR", "USD"
	Language       string            `json:"language"`            // e.g., "id", "en"
	BusinessHours  map[string]string `json:"business_hours"`      // e.g., {"monday": "09:00-17:00"}
	MaxUsers       int               `json:"max_users,omitempty"` // User quota for this branch
	Features       []string          `json:"features,omitempty"`  // Enabled features for this branch
	CustomSettings map[string]any    `json:"custom_settings,omitempty"`
}

// IsActive checks if branch is active
func (b *Branch) IsActive() bool {
	return b.Status == BranchStatusActive
}

// IsDisabled checks if branch is disabled
func (b *Branch) IsDisabled() bool {
	return b.Status == BranchStatusDisabled
}

// IsHeadquarters checks if branch is headquarters
func (b *Branch) IsHeadquarters() bool {
	return b.Type == BranchTypeHeadquarters
}

// Validate validates branch data
func (b *Branch) Validate() error {
	if b.ID == "" {
		return ErrInvalidBranchID
	}
	if b.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if b.AppID == "" {
		return errors.New("app_id is required")
	}
	if b.Name == "" {
		return errors.New("branch name is required")
	}
	if b.Status == "" {
		b.Status = BranchStatusActive
	}
	if b.Type == "" {
		b.Type = BranchTypeOffice
	}
	return nil
}

// GetFullAddress returns formatted full address
func (b *Branch) GetFullAddress() string {
	addr := b.Address
	result := ""
	if addr.Street != "" {
		result += addr.Street
	}
	if addr.City != "" {
		if result != "" {
			result += ", "
		}
		result += addr.City
	}
	if addr.State != "" {
		if result != "" {
			result += ", "
		}
		result += addr.State
	}
	if addr.Country != "" {
		if result != "" {
			result += ", "
		}
		result += addr.Country
	}
	if addr.PostalCode != "" {
		if result != "" {
			result += " "
		}
		result += addr.PostalCode
	}
	return result
}
