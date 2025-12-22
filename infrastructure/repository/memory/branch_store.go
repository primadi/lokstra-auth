package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
)

// InMemoryBranchStore is an in-memory implementation of BranchStore
type InMemoryBranchStore struct {
	// branches: map[tenantID:appID:branchID] -> Branch
	branches map[string]*domain.Branch
	mu       sync.RWMutex
}

var _ repository.BranchStore = (*InMemoryBranchStore)(nil)

// NewBranchStore creates a new in-memory branch store
func NewBranchStore() *InMemoryBranchStore {
	return &InMemoryBranchStore{
		branches: make(map[string]*domain.Branch),
	}
}

// makeKey creates a composite key for tenant+app+branch
func (s *InMemoryBranchStore) makeKey(tenantID, appID, branchID string) string {
	return tenantID + ":" + appID + ":" + branchID
}

// Create creates a new branch
func (s *InMemoryBranchStore) Create(ctx context.Context, branch *domain.Branch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(branch.TenantID, branch.AppID, branch.ID)
	if _, exists := s.branches[key]; exists {
		return fmt.Errorf("branch already exists: %s in app %s", branch.ID, branch.AppID)
	}

	s.branches[key] = branch
	return nil
}

// Get retrieves a branch by ID within an app
func (s *InMemoryBranchStore) Get(ctx context.Context, tenantID, appID, branchID string) (*domain.Branch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(tenantID, appID, branchID)
	branch, exists := s.branches[key]
	if !exists {
		return nil, fmt.Errorf("branch not found: %s in app %s", branchID, appID)
	}

	return branch, nil
}

// Update updates an existing branch
func (s *InMemoryBranchStore) Update(ctx context.Context, branch *domain.Branch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(branch.TenantID, branch.AppID, branch.ID)
	if _, exists := s.branches[key]; !exists {
		return fmt.Errorf("branch not found: %s in app %s", branch.ID, branch.AppID)
	}

	s.branches[key] = branch
	return nil
}

// Delete deletes a branch
func (s *InMemoryBranchStore) Delete(ctx context.Context, tenantID, appID, branchID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(tenantID, appID, branchID)
	delete(s.branches, key)
	return nil
}

// List lists all branches for an app
func (s *InMemoryBranchStore) List(ctx context.Context, tenantID, appID string) ([]*domain.Branch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	branches := make([]*domain.Branch, 0)
	for _, branch := range s.branches {
		if branch.TenantID == tenantID && branch.AppID == appID {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

// Exists checks if a branch exists within an app
func (s *InMemoryBranchStore) Exists(ctx context.Context, tenantID, appID, branchID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(tenantID, appID, branchID)
	_, exists := s.branches[key]
	return exists, nil
}

// ListByType lists branches by type within an app
func (s *InMemoryBranchStore) ListByType(ctx context.Context, tenantID, appID string, branchType domain.BranchType) ([]*domain.Branch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	branches := make([]*domain.Branch, 0)
	for _, branch := range s.branches {
		if branch.TenantID == tenantID && branch.AppID == appID && branch.Type == branchType {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}
