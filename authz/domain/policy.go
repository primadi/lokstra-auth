package domain

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

// CreatePolicyRequest for creating a new policy
type CreatePolicyRequest struct {
	TenantID    string          `json:"tenant_id" validate:"required"`
	AppID       string          `json:"app_id" validate:"required"`
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Effect      PolicyEffect    `json:"effect" validate:"required"`
	Subjects    []string        `json:"subjects" validate:"required"`
	Resources   []string        `json:"resources" validate:"required"`
	Actions     []string        `json:"actions" validate:"required"`
	Conditions  *map[string]any `json:"conditions,omitempty"`
	Metadata    *map[string]any `json:"metadata,omitempty"`
}

// GetPolicyRequest for retrieving a policy
type GetPolicyRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	PolicyID string `json:"policy_id" validate:"required"`
}

// UpdatePolicyRequest for updating a policy
type UpdatePolicyRequest struct {
	TenantID    string          `json:"tenant_id" validate:"required"`
	AppID       string          `json:"app_id" validate:"required"`
	PolicyID    string          `json:"policy_id" validate:"required"`
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Effect      *PolicyEffect   `json:"effect,omitempty"`
	Subjects    *[]string       `json:"subjects,omitempty"`
	Resources   *[]string       `json:"resources,omitempty"`
	Actions     *[]string       `json:"actions,omitempty"`
	Conditions  *map[string]any `json:"conditions,omitempty"`
	Status      *PolicyStatus   `json:"status,omitempty"`
	Metadata    *map[string]any `json:"metadata,omitempty"`
}

// DeletePolicyRequest for deleting a policy
type DeletePolicyRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	PolicyID string `json:"policy_id" validate:"required"`
}

// ListPoliciesRequest for listing policies
type ListPoliciesRequest struct {
	TenantID string        `json:"tenant_id" validate:"required"`
	AppID    string        `json:"app_id" validate:"required"`
	Effect   *PolicyEffect `json:"effect,omitempty"`
	Status   *PolicyStatus `json:"status,omitempty"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
}

// FindPoliciesBySubjectRequest for finding policies by subject
type FindPoliciesBySubjectRequest struct {
	TenantID  string `json:"tenant_id" validate:"required"`
	AppID     string `json:"app_id" validate:"required"`
	SubjectID string `json:"subject_id" validate:"required"`
}

// FindPoliciesByResourceRequest for finding policies by resource
type FindPoliciesByResourceRequest struct {
	TenantID     string `json:"tenant_id" validate:"required"`
	AppID        string `json:"app_id" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
	ResourceID   string `json:"resource_id,omitempty"`
}

// EvaluatePolicyRequest for evaluating authorization
type EvaluatePolicyRequest struct {
	TenantID  string          `json:"tenant_id" validate:"required"`
	AppID     string          `json:"app_id" validate:"required"`
	SubjectID string          `json:"subject_id" validate:"required"`
	Resource  string          `json:"resource" validate:"required"`
	Action    string          `json:"action" validate:"required"`
	Context   *map[string]any `json:"context,omitempty"`
}

// PolicyEvaluationResult represents the result of policy evaluation
type PolicyEvaluationResult struct {
	Allowed         bool     `json:"allowed"`
	Reason          string   `json:"reason"`
	MatchedPolicies []string `json:"matched_policies,omitempty"`
	Obligations     []string `json:"obligations,omitempty"`
}
