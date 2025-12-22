package domain

import (
	"errors"
	"time"
)

var (
	ErrPermissionCompositionNotFound      = errors.New("permission composition not found")
	ErrPermissionCompositionAlreadyExists = errors.New("permission composition already exists")
	ErrCircularDependency                 = errors.New("circular dependency detected in permission composition")
	ErrInvalidComposition                 = errors.New("invalid permission composition")
)

// PermissionComposition represents a compound permission definition
type PermissionComposition struct {
	ParentPermissionID string          `json:"parent_permission_id"`
	ChildPermissionID  string          `json:"child_permission_id"`
	TenantID           string          `json:"tenant_id"`
	AppID              string          `json:"app_id"`
	IsRequired         bool            `json:"is_required"`
	Priority           int             `json:"priority"`
	Metadata           *map[string]any `json:"metadata,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
}

// CreatePermissionCompositionRequest represents a request to add a child permission to a compound permission
type CreatePermissionCompositionRequest struct {
	TenantID           string          `json:"tenant_id" validate:"required"`
	AppID              string          `json:"app_id" validate:"required"`
	ParentPermissionID string          `json:"parent_permission_id" validate:"required"`
	ChildPermissionID  string          `json:"child_permission_id" validate:"required"`
	IsRequired         bool            `json:"is_required"`
	Priority           int             `json:"priority"`
	Metadata           *map[string]any `json:"metadata,omitempty"`
}

// DeletePermissionCompositionRequest represents a request to remove a child permission from a compound permission
type DeletePermissionCompositionRequest struct {
	TenantID           string `json:"tenant_id" validate:"required"`
	AppID              string `json:"app_id" validate:"required"`
	ParentPermissionID string `json:"parent_permission_id" validate:"required"`
	ChildPermissionID  string `json:"child_permission_id" validate:"required"`
}

// ListPermissionCompositionsRequest represents a request to list compositions
type ListPermissionCompositionsRequest struct {
	TenantID           string  `json:"tenant_id" validate:"required"`
	AppID              string  `json:"app_id" validate:"required"`
	ParentPermissionID *string `json:"parent_permission_id,omitempty"`
	ChildPermissionID  *string `json:"child_permission_id,omitempty"`
}

// GetEffectivePermissionsRequest represents a request to get all effective permissions for a compound permission
type GetEffectivePermissionsRequest struct {
	TenantID     string `json:"tenant_id" validate:"required"`
	AppID        string `json:"app_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
}

// Validate validates the create request
func (r *CreatePermissionCompositionRequest) Validate() error {
	if r.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if r.AppID == "" {
		return errors.New("app_id is required")
	}
	if r.ParentPermissionID == "" {
		return errors.New("parent_permission_id is required")
	}
	if r.ChildPermissionID == "" {
		return errors.New("child_permission_id is required")
	}
	if r.ParentPermissionID == r.ChildPermissionID {
		return errors.New("parent and child permission cannot be the same")
	}
	return nil
}
