package core

import (
	coredomain "github.com/primadi/lokstra-auth/pkg/domain/core"
)

// GetTenantCredentialConfigRequest request to get tenant credential config
type GetTenantCredentialConfigRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
}

// UpdateTenantCredentialConfigRequest request to update tenant credential config
type UpdateTenantCredentialConfigRequest struct {
	TenantID string                       `path:"tenant_id" validate:"required"`
	Config   *coredomain.CredentialConfig `json:"config" validate:"required"`
}

// GetAppCredentialConfigRequest request to get app credential config
type GetAppCredentialConfigRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
}

// UpdateAppCredentialConfigRequest request to update app credential config
type UpdateAppCredentialConfigRequest struct {
	TenantID string                       `path:"tenant_id" validate:"required"`
	AppID    string                       `path:"app_id" validate:"required"`
	Config   *coredomain.CredentialConfig `json:"config" validate:"required"`
}
