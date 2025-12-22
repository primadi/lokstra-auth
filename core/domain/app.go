package domain

import (
	"errors"
	"time"
)

var (
	ErrAppNotFound      = errors.New("app not found")
	ErrAppAlreadyExists = errors.New("app already exists")
	ErrAppDisabled      = errors.New("app is disabled")
	ErrInvalidAppID     = errors.New("invalid app ID")
)

// AppType represents the type of application
type AppType string

const (
	AppTypeWeb     AppType = "web"
	AppTypeMobile  AppType = "mobile"
	AppTypeAPI     AppType = "api"
	AppTypeDesktop AppType = "desktop"
	AppTypeService AppType = "service" // For service-to-service authentication
)

// AppStatus represents the status of an app
type AppStatus string

const (
	AppStatusActive   AppStatus = "active"
	AppStatusDisabled AppStatus = "disabled"
)

// App represents an application within a tenant
type App struct {
	ID        string          `json:"id"`        // Unique within tenant (e.g., "web-portal")
	TenantID  string          `json:"tenant_id"` // Parent tenant
	Name      string          `json:"name"`      // Display name
	Type      AppType         `json:"type"`      // web, mobile, api, desktop
	Status    AppStatus       `json:"status"`    // active, disabled
	Config    *AppConfig      `json:"config"`    // App-specific configuration
	Metadata  *map[string]any `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// AppConfig holds app-specific configuration
type AppConfig struct {
	// Credential Configuration (which auth methods are enabled)
	Credentials *CredentialConfig `json:"credentials"`

	// Token Configuration (per app)
	AccessTokenExpiry  time.Duration `json:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `json:"refresh_token_expiry"`
	TokenAlgorithm     string        `json:"token_algorithm"` // HS256, RS256, etc.
	TokenSecret        string        `json:"token_secret,omitempty"`

	// Service Authentication (for AppTypeService)
	AllowedScopes []string `json:"allowed_scopes,omitempty"` // Scopes this service can request

	// Security Settings (per app)
	AllowedOrigins   []string        `json:"allowed_origins"`   // CORS
	AllowedCallbacks []string        `json:"allowed_callbacks"` // OAuth2 redirects
	RateLimits       RateLimitConfig `json:"rate_limits"`

	// Feature Flags (per app)
	Features map[string]bool `json:"features,omitempty"`

	// Custom settings
	CustomSettings map[string]any `json:"custom_settings,omitempty"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool `json:"enabled"`
	RequestsPerMin int  `json:"requests_per_min"`
	BurstSize      int  `json:"burst_size"`
}

// AppFilters for querying apps
type AppFilters struct {
	TenantID string
	Type     AppType
	Status   AppStatus
	Limit    int
	Offset   int
}

// IsActive checks if app is active
func (a *App) IsActive() bool {
	return a.Status == AppStatusActive
}

// IsDisabled checks if app is disabled
func (a *App) IsDisabled() bool {
	return a.Status == AppStatusDisabled
}

// Validate validates app data
func (a *App) Validate() error {
	if a.ID == "" {
		return ErrInvalidAppID
	}
	if a.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if a.Name == "" {
		return errors.New("app name is required")
	}
	if a.Type == "" {
		a.Type = AppTypeWeb
	}
	if a.Status == "" {
		a.Status = AppStatusActive
	}
	return nil
}

// =============================================================================
// App DTOs
// =============================================================================

// CreateAppRequest request to create an app
type CreateAppRequest struct {
	TenantID string     `path:"tenant_id" validate:"required"`
	ID       string     `json:"id" validate:"required"`
	Name     string     `json:"name" validate:"required"`
	Type     AppType    `json:"type" validate:"required"`
	Config   *AppConfig `json:"config"`
}

// GetAppRequest request to get an app
type GetAppRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// UpdateAppRequest request to update an app
type UpdateAppRequest struct {
	TenantID string          `path:"tenant_id" validate:"required"`
	ID       string          `path:"id" validate:"required"`
	Name     string          `json:"name,omitempty"`
	Type     AppType         `json:"type,omitempty"`
	Config   *AppConfig      `json:"config,omitempty"`
	Status   AppStatus       `json:"status,omitempty"`
	Metadata *map[string]any `json:"metadata,omitempty"`
}

// DeleteAppRequest request to delete an app
type DeleteAppRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// ListAppsRequest request to list apps
type ListAppsRequest struct {
	TenantID string  `path:"tenant_id" validate:"required"`
	Type     AppType `query:"type"`
}

// ActivateAppRequest request to activate an app
type ActivateAppRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// SuspendAppRequest request to suspend an app
type SuspendAppRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}
