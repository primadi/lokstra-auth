package authz

import (
	authzdomain "github.com/primadi/lokstra-auth/pkg/domain/authz"
)

// CreatePolicyRequest for creating a new policy
type CreatePolicyRequest struct {
	TenantID    string                   `json:"tenant_id" validate:"required"`
	AppID       string                   `json:"app_id" validate:"required"`
	Name        string                   `json:"name" validate:"required"`
	Description string                   `json:"description"`
	Effect      authzdomain.PolicyEffect `json:"effect" validate:"required"`
	Subjects    []string                 `json:"subjects" validate:"required"`
	Resources   []string                 `json:"resources" validate:"required"`
	Actions     []string                 `json:"actions" validate:"required"`
	Conditions  *map[string]any          `json:"conditions,omitempty"`
	Metadata    *map[string]any          `json:"metadata,omitempty"`
}

// GetPolicyRequest for retrieving a policy
type GetPolicyRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	PolicyID string `json:"policy_id" validate:"required"`
}

// UpdatePolicyRequest for updating a policy
type UpdatePolicyRequest struct {
	TenantID    string                    `json:"tenant_id" validate:"required"`
	AppID       string                    `json:"app_id" validate:"required"`
	PolicyID    string                    `json:"policy_id" validate:"required"`
	Name        *string                   `json:"name,omitempty"`
	Description *string                   `json:"description,omitempty"`
	Effect      *authzdomain.PolicyEffect `json:"effect,omitempty"`
	Subjects    *[]string                 `json:"subjects,omitempty"`
	Resources   *[]string                 `json:"resources,omitempty"`
	Actions     *[]string                 `json:"actions,omitempty"`
	Conditions  *map[string]any           `json:"conditions,omitempty"`
	Status      *authzdomain.PolicyStatus `json:"status,omitempty"`
	Metadata    *map[string]any           `json:"metadata,omitempty"`
}

// DeletePolicyRequest for deleting a policy
type DeletePolicyRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	PolicyID string `json:"policy_id" validate:"required"`
}

// ListPoliciesRequest for listing policies
type ListPoliciesRequest struct {
	TenantID string                    `json:"tenant_id" validate:"required"`
	AppID    string                    `json:"app_id" validate:"required"`
	Effect   *authzdomain.PolicyEffect `json:"effect,omitempty"`
	Status   *authzdomain.PolicyStatus `json:"status,omitempty"`
	Limit    int                       `json:"limit"`
	Offset   int                       `json:"offset"`
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
