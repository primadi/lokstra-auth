package core

import (
	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
)

// CreateProviderRequest request to create a credential provider
type CreateProviderRequest struct {
	TenantID    string                  `path:"tenant_id" validate:"required"`
	AppID       string                  `json:"app_id"` // Optional, NULL = tenant-level
	Type        domaincore.ProviderType `json:"type" validate:"required"`
	Name        string                  `json:"name" validate:"required"`
	Description string                  `json:"description"`
	Config      map[string]any          `json:"config" validate:"required"`
	Metadata    *map[string]any         `json:"metadata,omitempty"`
}

// GetProviderRequest request to get a provider
type GetProviderRequest struct {
	TenantID   string `path:"tenant_id" validate:"required"`
	ProviderID string `path:"provider_id" validate:"required"`
}

// UpdateProviderRequest request to update a provider
type UpdateProviderRequest struct {
	TenantID    string                    `path:"tenant_id" validate:"required"`
	ProviderID  string                    `path:"provider_id" validate:"required"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Config      map[string]any            `json:"config"`
	Status      domaincore.ProviderStatus `json:"status"`
	Metadata    *map[string]any           `json:"metadata"`
}

// DeleteProviderRequest request to delete a provider
type DeleteProviderRequest struct {
	TenantID   string `path:"tenant_id" validate:"required"`
	ProviderID string `path:"provider_id" validate:"required"`
}

// ListProvidersRequest request to list providers
type ListProvidersRequest struct {
	TenantID string                    `path:"tenant_id" validate:"required"`
	AppID    string                    `query:"app_id"` // Filter by app (empty = all)
	Type     domaincore.ProviderType   `query:"type"`   // Filter by type
	Status   domaincore.ProviderStatus `query:"status"` // Filter by status
}

// EnableProviderRequest request to enable a provider
type EnableProviderRequest struct {
	TenantID   string `path:"tenant_id" validate:"required"`
	ProviderID string `path:"provider_id" validate:"required"`
}

// DisableProviderRequest request to disable a provider
type DisableProviderRequest struct {
	TenantID   string `path:"tenant_id" validate:"required"`
	ProviderID string `path:"provider_id" validate:"required"`
}

// GetProvidersByTypeRequest request to get providers by type
type GetProvidersByTypeRequest struct {
	TenantID string                  `path:"tenant_id" validate:"required"`
	AppID    string                  `query:"app_id"`
	Type     domaincore.ProviderType `path:"provider_type" validate:"required"`
}

// GetActiveProviderForAppRequest request to get active provider for an app
type GetActiveProviderForAppRequest struct {
	TenantID string                  `path:"tenant_id" validate:"required"`
	AppID    string                  `query:"app_id"`
	Type     domaincore.ProviderType `query:"type" validate:"required"`
}
