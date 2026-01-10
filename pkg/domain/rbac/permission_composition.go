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
