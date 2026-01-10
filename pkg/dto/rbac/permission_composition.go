package rbac

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
