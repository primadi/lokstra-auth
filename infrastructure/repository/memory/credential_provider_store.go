package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
)

// InMemoryCredentialProviderStore is an in-memory implementation of CredentialProviderStore
type InMemoryCredentialProviderStore struct {
	// providers: map[tenantID:providerID] -> CredentialProvider
	providers map[string]*domain.CredentialProvider
	mu        sync.RWMutex
}

var _ repository.CredentialProviderStore = (*InMemoryCredentialProviderStore)(nil)

// NewCredentialProviderStore creates a new in-memory credential provider store
func NewCredentialProviderStore() *InMemoryCredentialProviderStore {
	return &InMemoryCredentialProviderStore{
		providers: make(map[string]*domain.CredentialProvider),
	}
}

// makeKey creates a composite key for tenant+provider
func (s *InMemoryCredentialProviderStore) makeKey(tenantID, providerID string) string {
	return fmt.Sprintf("%s:%s", tenantID, providerID)
}

// Create creates a new credential provider
func (s *InMemoryCredentialProviderStore) Create(ctx context.Context, provider *domain.CredentialProvider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := provider.Validate(); err != nil {
		return err
	}

	key := s.makeKey(provider.TenantID, provider.ID)
	if _, exists := s.providers[key]; exists {
		return domain.ErrProviderAlreadyExists
	}

	s.providers[key] = provider
	return nil
}

// Get retrieves a provider by ID
func (s *InMemoryCredentialProviderStore) Get(ctx context.Context, tenantID, providerID string) (*domain.CredentialProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(tenantID, providerID)
	provider, exists := s.providers[key]
	if !exists {
		return nil, domain.ErrProviderNotFound
	}

	return provider, nil
}

// Update updates an existing provider
func (s *InMemoryCredentialProviderStore) Update(ctx context.Context, provider *domain.CredentialProvider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := provider.Validate(); err != nil {
		return err
	}

	key := s.makeKey(provider.TenantID, provider.ID)
	if _, exists := s.providers[key]; !exists {
		return domain.ErrProviderNotFound
	}

	s.providers[key] = provider
	return nil
}

// Delete deletes a provider
func (s *InMemoryCredentialProviderStore) Delete(ctx context.Context, tenantID, providerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(tenantID, providerID)
	if _, exists := s.providers[key]; !exists {
		return domain.ErrProviderNotFound
	}

	delete(s.providers, key)
	return nil
}

// List lists all providers for tenant (optionally filtered by app)
func (s *InMemoryCredentialProviderStore) List(ctx context.Context, tenantID, appID string) ([]*domain.CredentialProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.CredentialProvider
	for _, provider := range s.providers {
		if provider.TenantID != tenantID {
			continue
		}

		// Filter by app
		if appID != "" {
			// If appID specified, return only providers for that app OR tenant-level providers
			if provider.AppID != appID && provider.AppID != "" {
				continue
			}
		}

		result = append(result, provider)
	}

	return result, nil
}

// ListByType lists all providers of a specific type for tenant+app
func (s *InMemoryCredentialProviderStore) ListByType(ctx context.Context, tenantID, appID string, providerType domain.ProviderType) ([]*domain.CredentialProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.CredentialProvider
	for _, provider := range s.providers {
		if provider.TenantID != tenantID {
			continue
		}

		if provider.Type != providerType {
			continue
		}

		// Filter by app
		if appID != "" {
			// If appID specified, return only providers for that app OR tenant-level providers
			if provider.AppID != appID && provider.AppID != "" {
				continue
			}
		} else {
			// If appID not specified (empty), return only tenant-level providers
			if provider.AppID != "" {
				continue
			}
		}

		result = append(result, provider)
	}

	return result, nil
}

// Exists checks if a provider exists
func (s *InMemoryCredentialProviderStore) Exists(ctx context.Context, tenantID, providerID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(tenantID, providerID)
	_, exists := s.providers[key]
	return exists, nil
}
