package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/authz/domain"
	"github.com/primadi/lokstra-auth/core/infrastructure/idgen"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
)

// PolicyService handles policy management operations
// @RouterService name="policy-service", prefix="${api-auth-prefix:/api/auth}/authz/tenants/{tenant_id}/apps/{app_id}/policies", middlewares=["recovery", "request_logger", "auth"]
type PolicyService struct {
	// @Inject "@store.policy-store"
	PolicyStore repository.PolicyStore
}

// CreatePolicy creates a new policy
// @Route "POST /"
func (s *PolicyService) CreatePolicy(ctx *request.Context, req *domain.CreatePolicyRequest) (*domain.Policy, error) {
	// Check if policy name already exists
	existing, _ := s.PolicyStore.GetByName(ctx, req.TenantID, req.AppID, req.Name)
	if existing != nil {
		return nil, domain.ErrDuplicatePolicyName
	}

	policy := &domain.Policy{
		ID:          idgen.GenerateID("policy"),
		TenantID:    req.TenantID,
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Effect:      req.Effect,
		Subjects:    req.Subjects,
		Resources:   req.Resources,
		Actions:     req.Actions,
		Conditions:  req.Conditions,
		Status:      domain.PolicyStatusActive,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := policy.Validate(); err != nil {
		return nil, err
	}

	if err := s.PolicyStore.Create(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	return policy, nil
}

// GetPolicy retrieves a policy by ID
// @Route "GET /{policy_id}"
func (s *PolicyService) GetPolicy(ctx *request.Context, req *domain.GetPolicyRequest) (*domain.Policy, error) {
	policy, err := s.PolicyStore.Get(ctx, req.TenantID, req.AppID, req.PolicyID)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// UpdatePolicy updates an existing policy
// @Route "PUT /{policy_id}"
func (s *PolicyService) UpdatePolicy(ctx *request.Context, req *domain.UpdatePolicyRequest) (*domain.Policy, error) {
	policy, err := s.PolicyStore.Get(ctx, req.TenantID, req.AppID, req.PolicyID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		// Check for duplicate name
		existing, _ := s.PolicyStore.GetByName(ctx, req.TenantID, req.AppID, *req.Name)
		if existing != nil && existing.ID != policy.ID {
			return nil, domain.ErrDuplicatePolicyName
		}
		policy.Name = *req.Name
	}
	if req.Description != nil {
		policy.Description = *req.Description
	}
	if req.Effect != nil {
		policy.Effect = *req.Effect
	}
	if req.Subjects != nil {
		policy.Subjects = *req.Subjects
	}
	if req.Resources != nil {
		policy.Resources = *req.Resources
	}
	if req.Actions != nil {
		policy.Actions = *req.Actions
	}
	if req.Conditions != nil {
		policy.Conditions = req.Conditions
	}
	if req.Status != nil {
		policy.Status = *req.Status
	}
	if req.Metadata != nil {
		policy.Metadata = req.Metadata
	}
	policy.UpdatedAt = time.Now()

	if err := s.PolicyStore.Update(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	return policy, nil
}

// DeletePolicy deletes a policy
// @Route "DELETE /{policy_id}"
func (s *PolicyService) DeletePolicy(ctx *request.Context, req *domain.DeletePolicyRequest) error {
	if err := s.PolicyStore.Delete(ctx, req.TenantID, req.AppID, req.PolicyID); err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}
	return nil
}

// ListPolicies lists all policies for tenant+app
// @Route "GET /"
func (s *PolicyService) ListPolicies(ctx *request.Context, req *domain.ListPoliciesRequest) ([]*domain.Policy, error) {
	if req.Limit == 0 {
		req.Limit = 100
	}

	policies, err := s.PolicyStore.ListWithFilters(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	return policies, nil
}

// FindPoliciesBySubject finds policies matching a subject
// @Route "GET /by-subject/{subject_id}"
func (s *PolicyService) FindPoliciesBySubject(ctx *request.Context, req *domain.FindPoliciesBySubjectRequest) ([]*domain.Policy, error) {
	policies, err := s.PolicyStore.FindBySubject(ctx, req.TenantID, req.AppID, req.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find policies by subject: %w", err)
	}
	return policies, nil
}

// FindPoliciesByResource finds policies matching a resource
// @Route "GET /by-resource/{resource_type}"
func (s *PolicyService) FindPoliciesByResource(ctx *request.Context, req *domain.FindPoliciesByResourceRequest) ([]*domain.Policy, error) {
	policies, err := s.PolicyStore.FindByResource(ctx, req.TenantID, req.AppID, req.ResourceType, req.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find policies by resource: %w", err)
	}
	return policies, nil
}
