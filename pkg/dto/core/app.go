package core

import (
	coredomain "github.com/primadi/lokstra-auth/pkg/domain/core"
)

// CreateAppRequest request to create a new app
type CreateAppRequest struct {
	TenantID string                `path:"tenant_id" validate:"required"`
	ID       string                `json:"id" validate:"required"`
	Name     string                `json:"name" validate:"required"`
	Type     coredomain.AppType    `json:"type" validate:"required"`
	Config   *coredomain.AppConfig `json:"config"`
}

// GetAppRequest request to get an app
type GetAppRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// UpdateAppRequest request to update an app
type UpdateAppRequest struct {
	TenantID string                `path:"tenant_id" validate:"required"`
	ID       string                `path:"id" validate:"required"`
	Name     string                `json:"name,omitempty"`
	Type     coredomain.AppType    `json:"type,omitempty"`
	Config   *coredomain.AppConfig `json:"config,omitempty"`
	Status   coredomain.AppStatus  `json:"status,omitempty"`
	Metadata *map[string]any       `json:"metadata,omitempty"`
}

// DeleteAppRequest request to delete an app
type DeleteAppRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// ListAppsRequest request to list apps
type ListAppsRequest struct {
	TenantID string             `path:"tenant_id" validate:"required"`
	Type     coredomain.AppType `query:"type"`
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
