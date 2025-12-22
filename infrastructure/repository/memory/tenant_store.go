package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
)

// InMemoryTenantStore is an in-memory implementation of TenantStore
type InMemoryTenantStore struct {
	tenants map[string]*domain.Tenant
	mu      sync.RWMutex
}

var _ repository.TenantStore = (*InMemoryTenantStore)(nil)

// NewTenantStore creates a new in-memory tenant store
func NewTenantStore() *InMemoryTenantStore {
	return &InMemoryTenantStore{
		tenants: make(map[string]*domain.Tenant),
	}
}

// Create creates a new tenant
func (s *InMemoryTenantStore) Create(ctx context.Context, tenant *domain.Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[tenant.ID]; exists {
		return fmt.Errorf("tenant already exists: %s", tenant.ID)
	}

	s.tenants[tenant.ID] = tenant
	return nil
}

// Get retrieves a tenant by ID
func (s *InMemoryTenantStore) Get(ctx context.Context, tenantID string) (*domain.Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenant, exists := s.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	return tenant, nil
}

// Update updates an existing tenant
func (s *InMemoryTenantStore) Update(ctx context.Context, tenant *domain.Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[tenant.ID]; !exists {
		return fmt.Errorf("tenant not found: %s", tenant.ID)
	}

	s.tenants[tenant.ID] = tenant
	return nil
}

// Delete deletes a tenant
func (s *InMemoryTenantStore) Delete(ctx context.Context, tenantID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tenants, tenantID)
	return nil
}

// List lists all tenants
func (s *InMemoryTenantStore) List(ctx context.Context) ([]*domain.Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenants := make([]*domain.Tenant, 0, len(s.tenants))
	for _, tenant := range s.tenants {
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

// GetByName retrieves a tenant by name
func (s *InMemoryTenantStore) GetByName(ctx context.Context, name string) (*domain.Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, tenant := range s.tenants {
		if tenant.Name == name {
			return tenant, nil
		}
	}

	return nil, fmt.Errorf("tenant not found with name: %s", name)
}

// Exists checks if a tenant exists
func (s *InMemoryTenantStore) Exists(ctx context.Context, tenantID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.tenants[tenantID]
	return exists, nil
}
