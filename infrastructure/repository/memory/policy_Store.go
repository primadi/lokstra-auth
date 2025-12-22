package memory

import (
	"context"
	"fmt"
	"sync"

	authz "github.com/primadi/lokstra-auth/authz"
)

// InMemoryStore is an in-memory implementation of PolicyStore
// @Service "in-memory-policy-store"
type InMemoryStore struct {
	mu       sync.RWMutex
	policies map[string]*authz.Policy
}

var _ authz.PolicyStore = (*InMemoryStore)(nil)

func (s *InMemoryStore) Init() error {
	s.policies = make(map[string]*authz.Policy)
	return nil
}

// Create creates a new policy
func (s *InMemoryStore) Create(ctx context.Context, policy *authz.Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[policy.ID]; exists {
		return fmt.Errorf("policy already exists: %s", policy.ID)
	}

	s.policies[policy.ID] = policy
	return nil
}

// Get retrieves a policy by ID (scoped to tenant)
func (s *InMemoryStore) Get(ctx context.Context, tenantID, policyID string) (*authz.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, exists := s.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", policyID)
	}

	// Verify tenant matches
	if policy.TenantID != tenantID {
		return nil, fmt.Errorf("policy not found: %s", policyID)
	}

	return policy, nil
}

// Update updates an existing policy
func (s *InMemoryStore) Update(ctx context.Context, policy *authz.Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[policy.ID]; !exists {
		return fmt.Errorf("policy not found: %s", policy.ID)
	}

	s.policies[policy.ID] = policy
	return nil
}

// Delete deletes a policy
func (s *InMemoryStore) Delete(ctx context.Context, tenantID, policyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.policies[policyID]
	if !exists {
		return fmt.Errorf("policy not found: %s", policyID)
	}

	if policy.TenantID != tenantID {
		return fmt.Errorf("policy not found: %s", policyID)
	}

	delete(s.policies, policyID)
	return nil
}

// List lists all policies for a tenant
func (s *InMemoryStore) List(ctx context.Context, tenantID string) ([]*authz.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]*authz.Policy, 0, len(s.policies))
	for _, policy := range s.policies {
		if policy.TenantID == tenantID {
			policies = append(policies, policy)
		}
	}

	return policies, nil
}

// ListByApp lists all policies for a specific app within a tenant
func (s *InMemoryStore) ListByApp(ctx context.Context, tenantID, appID string) ([]*authz.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]*authz.Policy, 0)
	for _, policy := range s.policies {
		if policy.TenantID == tenantID && (policy.AppID == appID || policy.AppID == "") {
			policies = append(policies, policy)
		}
	}

	return policies, nil
}

// FindBySubject finds policies for a subject within a tenant+app
func (s *InMemoryStore) FindBySubject(ctx context.Context, tenantID, appID, subjectID string) ([]*authz.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]*authz.Policy, 0)
	for _, policy := range s.policies {
		// Check tenant and app scope
		if policy.TenantID != tenantID {
			continue
		}
		if policy.AppID != "" && policy.AppID != appID {
			continue
		}

		// Check if subject matches
		for _, sub := range policy.Subjects {
			if sub == subjectID || sub == "*" {
				policies = append(policies, policy)
				break
			}
		}
	}

	return policies, nil
}

// FindByResource finds policies for a resource within a tenant+app
func (s *InMemoryStore) FindByResource(ctx context.Context, tenantID, appID, resourceType, resourceID string) ([]*authz.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]*authz.Policy, 0)
	resourcePattern := fmt.Sprintf("%s:%s", resourceType, resourceID)

	for _, policy := range s.policies {
		// Check tenant and app scope
		if policy.TenantID != tenantID {
			continue
		}
		if policy.AppID != "" && policy.AppID != appID {
			continue
		}

		// Check if resource matches
		for _, res := range policy.Resources {
			if res == resourcePattern || res == fmt.Sprintf("%s:*", resourceType) || res == "*" {
				policies = append(policies, policy)
				break
			}
		}
	}

	return policies, nil
}
