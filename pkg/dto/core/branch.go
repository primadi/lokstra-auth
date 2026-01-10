package core

import (
	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
)

// CreateBranchRequest request to create a branch
type CreateBranchRequest struct {
	TenantID string                     `path:"tenant_id" validate:"required"`
	AppID    string                     `path:"app_id" validate:"required"`
	BranchID string                     `json:"branch_id" validate:"required"`
	Name     string                     `json:"name" validate:"required"`
	Type     domaincore.BranchType      `json:"type" validate:"required"`
	Settings *domaincore.BranchSettings `json:"settings"`
	Metadata *map[string]any            `json:"metadata"`
}

// GetBranchRequest request to get a branch
type GetBranchRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// GetBranchByCodeRequest request to get a branch by code
type GetBranchByCodeRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	Code     string `path:"code" validate:"required"`
}

// UpdateBranchRequest request to update a branch
type UpdateBranchRequest struct {
	TenantID string                     `path:"tenant_id" validate:"required"`
	AppID    string                     `path:"app_id" validate:"required"`
	ID       string                     `path:"id" validate:"required"`
	Name     string                     `json:"name"`
	Type     domaincore.BranchType      `json:"type"`
	Status   domaincore.BranchStatus    `json:"status"`
	Settings *domaincore.BranchSettings `json:"settings"`
	Metadata *map[string]any            `json:"metadata"`
}

// DeleteBranchRequest request to delete a branch
type DeleteBranchRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	ID       string `path:"id" validate:"required"`
}

// ListBranchesRequest request to list branches
type ListBranchesRequest struct {
	TenantID string                `path:"tenant_id" validate:"required"`
	AppID    string                `path:"app_id" validate:"required"`
	Type     domaincore.BranchType `query:"type,omitempty"`
}

// ActivateBranchRequest request to activate a branch
type ActivateBranchRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	ID       string `json:"id" validate:"required"`
}

// DisableBranchRequest request to disable a branch
type DisableBranchRequest struct {
	TenantID string `path:"tenant_id" validate:"required"`
	AppID    string `path:"app_id" validate:"required"`
	ID       string `json:"id" validate:"required"`
}
