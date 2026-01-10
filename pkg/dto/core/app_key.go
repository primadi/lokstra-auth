package core

import (
	"time"

	coredomain "github.com/primadi/lokstra-auth/pkg/domain/core"
)

// GenerateAppKeyRequest request to generate a new app key
type GenerateAppKeyRequest struct {
	TenantID    string         `path:"tenant_id" validate:"required"`
	AppID       string         `path:"app_id" validate:"required"`
	Name        string         `json:"name" validate:"required"`
	Purpose     string         `json:"purpose"`
	Description string         `json:"description"`
	Environment string         `json:"environment"` // live, test
	Scopes      []string       `json:"scopes"`
	ExpiresIn   *time.Duration `json:"expires_in"` // nil = never expires
}

// AppKeyResponse response containing generated app key (with secret - shown only once!)
type AppKeyResponse struct {
	KeyID     string             `json:"key_id"`
	KeyString string             `json:"key_string"` // Full key with secret - ONLY shown once!
	AppKey    *coredomain.AppKey `json:"app_key"`
}

// AppKeyInfo sanitized response without secret hash (for Get/List operations)
type AppKeyInfo struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	AppID       string          `json:"app_id"`
	BranchID    *string         `json:"branch_id,omitempty"`
	KeyID       string          `json:"key_id"` // Public identifier only
	KeyType     string          `json:"key_type"`
	Environment string          `json:"environment"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Status      string          `json:"status"`
	Metadata    *map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	ExpiresAt   *time.Time      `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time      `json:"last_used_at,omitempty"`
	CreatedBy   string          `json:"created_by"`
	// ❌ NO SecretHash - never exposed!
}

// ToAppKeyInfo converts AppKey to sanitized AppKeyInfo
func ToAppKeyInfo(key *coredomain.AppKey) *AppKeyInfo {
	if key == nil {
		return nil
	}
	return &AppKeyInfo{
		ID:          key.ID,
		TenantID:    key.TenantID,
		AppID:       key.AppID,
		BranchID:    key.BranchID,
		KeyID:       key.KeyID,
		KeyType:     key.KeyType,
		Environment: key.Environment,
		Name:        key.Name,
		Description: key.Description,
		Status:      key.Status,
		Metadata:    key.Metadata,
		CreatedAt:   key.CreatedAt,
		ExpiresAt:   key.ExpiresAt,
		LastUsedAt:  key.LastUsedAt,
		CreatedBy:   key.CreatedBy,
	}
}

// ToAppKeyInfoList converts slice of AppKey to slice of AppKeyInfo
func ToAppKeyInfoList(keys []*coredomain.AppKey) []*AppKeyInfo {
	result := make([]*AppKeyInfo, len(keys))
	for i, key := range keys {
		result[i] = ToAppKeyInfo(key)
	}
	return result
}

// GetAppKeyRequest request to get an app key
type GetAppKeyRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	KeyID    string `path:"key_id" validate:"required"`
}

// ListAppKeysRequest request to list app keys
type ListAppKeysRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
}

// RevokeAppKeyRequest request to revoke an app key
type RevokeAppKeyRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	KeyID    string `path:"key_id" validate:"required"`
}

// DeleteAppKeyRequest request to delete an app key
type DeleteAppKeyRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	KeyID    string `path:"key_id" validate:"required"`
}

// RotateAppKeyRequest request to rotate an app key
type RotateAppKeyRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	KeyID    string `path:"key_id" validate:"required"`
}
