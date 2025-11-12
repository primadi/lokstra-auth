package policy

import (
	"context"
	"fmt"
	"sync"

	authz "github.com/primadi/lokstra-auth/04_authz"
)

// InMemoryStore is an in-memory implementation of PolicyStore
type InMemoryStore struct {
	mu       sync.RWMutex
	policies map[string]*authz.Policy
}

// NewInMemoryStore creates a new in-memory policy store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		policies: make(map[string]*authz.Policy),
	}
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

// Get retrieves a policy by ID
func (s *InMemoryStore) Get(ctx context.Context, policyID string) (*authz.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, exists := s.policies[policyID]
	if !exists {
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
func (s *InMemoryStore) Delete(ctx context.Context, policyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[policyID]; !exists {
		return fmt.Errorf("policy not found: %s", policyID)
	}

	delete(s.policies, policyID)
	return nil
}

// List lists all policies
func (s *InMemoryStore) List(ctx context.Context) ([]*authz.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]*authz.Policy, 0, len(s.policies))
	for _, policy := range s.policies {
		policies = append(policies, policy)
	}

	return policies, nil
}

// FindBySubject finds policies for a subject
func (s *InMemoryStore) FindBySubject(ctx context.Context, subjectID string) ([]*authz.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]*authz.Policy, 0)
	for _, policy := range s.policies {
		for _, sub := range policy.Subjects {
			if sub == subjectID || sub == "*" {
				policies = append(policies, policy)
				break
			}
		}
	}

	return policies, nil
}

// FindByResource finds policies for a resource
func (s *InMemoryStore) FindByResource(ctx context.Context, resourceType string, resourceID string) ([]*authz.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]*authz.Policy, 0)
	resourcePattern := fmt.Sprintf("%s:%s", resourceType, resourceID)

	for _, policy := range s.policies {
		for _, res := range policy.Resources {
			if res == resourcePattern || res == fmt.Sprintf("%s:*", resourceType) || res == "*" {
				policies = append(policies, policy)
				break
			}
		}
	}

	return policies, nil
}
