package authz

import (
	"errors"
	"time"
)

var (
	ErrPolicyNotFound      = errors.New("policy not found")
	ErrPolicyAlreadyExists = errors.New("policy already exists")
	ErrInvalidPolicyID     = errors.New("invalid policy ID")
	ErrDuplicatePolicyName = errors.New("policy name already exists in this tenant+app")
)

// PolicyEffect represents the effect of a policy
type PolicyEffect string

const (
	PolicyEffectAllow PolicyEffect = "allow"
	PolicyEffectDeny  PolicyEffect = "deny"
)

// PolicyStatus represents the status of a policy
type PolicyStatus string

const (
	PolicyStatusActive   PolicyStatus = "active"
	PolicyStatusInactive PolicyStatus = "inactive"
)

// Policy represents an authorization policy (tenant+app scoped)
type Policy struct {
	ID          string          `json:"id"`                   // Unique identifier (UUID)
	TenantID    string          `json:"tenant_id"`            // Belongs to tenant (REQUIRED)
	AppID       string          `json:"app_id"`               // Belongs to app (REQUIRED)
	Name        string          `json:"name"`                 // Policy name
	Description string          `json:"description"`          // Policy description
	Effect      PolicyEffect    `json:"effect"`               // allow, deny
	Subjects    []string        `json:"subjects"`             // Subject patterns (user IDs, role IDs, "*")
	Resources   []string        `json:"resources"`            // Resource patterns ("document:*", "api:/users/*")
	Actions     []string        `json:"actions"`              // Actions (read, write, delete, etc.)
	Conditions  *map[string]any `json:"conditions,omitempty"` // ABAC conditions
	Status      PolicyStatus    `json:"status"`               // active, inactive
	Metadata    *map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Validate validates policy data
func (p *Policy) Validate() error {
	if p.ID == "" {
		return ErrInvalidPolicyID
	}
	if p.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if p.AppID == "" {
		return errors.New("app_id is required")
	}
	if p.Name == "" {
		return errors.New("policy name is required")
	}
	if p.Effect != PolicyEffectAllow && p.Effect != PolicyEffectDeny {
		return errors.New("policy effect must be 'allow' or 'deny'")
	}
	if len(p.Subjects) == 0 {
		return errors.New("policy must have at least one subject")
	}
	if len(p.Resources) == 0 {
		return errors.New("policy must have at least one resource")
	}
	if len(p.Actions) == 0 {
		return errors.New("policy must have at least one action")
	}
	return nil
}

// IsActive checks if policy is active
func (p *Policy) IsActive() bool {
	return p.Status == PolicyStatusActive
}
