package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
)

// InMemoryUserIdentityStore is an in-memory implementation of UserIdentityStore
type InMemoryUserIdentityStore struct {
	// identities: map[tenantID:userID:identityID] -> UserIdentity
	identities map[string]*domain.UserIdentity
	// providerIndex: map[tenantID:provider:providerID] -> identityID
	providerIndex map[string]string
	mu            sync.RWMutex
	userStore     repository.UserStore
}

var _ repository.UserIdentityStore = (*InMemoryUserIdentityStore)(nil)

// NewUserIdentityStore creates a new in-memory user identity store
func NewUserIdentityStore(userStore repository.UserStore) *InMemoryUserIdentityStore {
	return &InMemoryUserIdentityStore{
		identities:    make(map[string]*domain.UserIdentity),
		providerIndex: make(map[string]string),
		userStore:     userStore,
	}
}

// makeKey creates a composite key for tenant+user+identity
func (s *InMemoryUserIdentityStore) makeKey(tenantID, userID, identityID string) string {
	return fmt.Sprintf("%s:%s:%s", tenantID, userID, identityID)
}

// makeProviderKey creates a key for provider index
func (s *InMemoryUserIdentityStore) makeProviderKey(tenantID string, provider domain.IdentityProvider, providerID string) string {
	return fmt.Sprintf("%s:%s:%s", tenantID, provider, providerID)
}

// Create creates a new user identity link
func (s *InMemoryUserIdentityStore) Create(ctx context.Context, identity *domain.UserIdentity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := identity.Validate(); err != nil {
		return err
	}

	key := s.makeKey(identity.TenantID, identity.UserID, identity.ID)
	if _, exists := s.identities[key]; exists {
		return domain.ErrUserIdentityAlreadyExists
	}

	// Check if provider identity already linked
	providerKey := s.makeProviderKey(identity.TenantID, identity.Provider, identity.ProviderID)
	if existingID, exists := s.providerIndex[providerKey]; exists {
		existingIdentity := s.identities[s.findIdentityByID(identity.TenantID, existingID)]
		if existingIdentity != nil && existingIdentity.UserID != identity.UserID {
			return domain.ErrDuplicateProviderIdentity
		}
	}

	s.identities[key] = identity
	s.providerIndex[providerKey] = identity.ID

	return nil
}

// findIdentityByID finds identity key by identity ID
func (s *InMemoryUserIdentityStore) findIdentityByID(tenantID, identityID string) string {
	for key, identity := range s.identities {
		if identity.TenantID == tenantID && identity.ID == identityID {
			return key
		}
	}
	return ""
}

// Get retrieves an identity by ID
func (s *InMemoryUserIdentityStore) Get(ctx context.Context, tenantID, userID, identityID string) (*domain.UserIdentity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(tenantID, userID, identityID)
	identity, exists := s.identities[key]
	if !exists {
		return nil, domain.ErrUserIdentityNotFound
	}

	return identity, nil
}

// GetByProvider retrieves an identity by provider for a user
func (s *InMemoryUserIdentityStore) GetByProvider(ctx context.Context, tenantID, userID string, provider domain.IdentityProvider) (*domain.UserIdentity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, identity := range s.identities {
		if identity.TenantID == tenantID &&
			identity.UserID == userID &&
			identity.Provider == provider {
			return identity, nil
		}
	}

	return nil, domain.ErrUserIdentityNotFound
}

// Update updates an existing identity
func (s *InMemoryUserIdentityStore) Update(ctx context.Context, identity *domain.UserIdentity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := identity.Validate(); err != nil {
		return err
	}

	key := s.makeKey(identity.TenantID, identity.UserID, identity.ID)
	if _, exists := s.identities[key]; !exists {
		return domain.ErrUserIdentityNotFound
	}

	// Update provider index if provider ID changed
	oldIdentity := s.identities[key]
	if oldIdentity.ProviderID != identity.ProviderID {
		// Remove old provider index
		oldProviderKey := s.makeProviderKey(oldIdentity.TenantID, oldIdentity.Provider, oldIdentity.ProviderID)
		delete(s.providerIndex, oldProviderKey)

		// Add new provider index
		newProviderKey := s.makeProviderKey(identity.TenantID, identity.Provider, identity.ProviderID)
		s.providerIndex[newProviderKey] = identity.ID
	}

	s.identities[key] = identity
	return nil
}

// Delete removes an identity link
func (s *InMemoryUserIdentityStore) Delete(ctx context.Context, tenantID, userID, identityID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(tenantID, userID, identityID)
	identity, exists := s.identities[key]
	if !exists {
		return domain.ErrUserIdentityNotFound
	}

	// Remove from provider index
	providerKey := s.makeProviderKey(identity.TenantID, identity.Provider, identity.ProviderID)
	delete(s.providerIndex, providerKey)

	delete(s.identities, key)
	return nil
}

// List lists all identities for a user
func (s *InMemoryUserIdentityStore) List(ctx context.Context, tenantID, userID string) ([]*domain.UserIdentity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.UserIdentity
	for _, identity := range s.identities {
		if identity.TenantID == tenantID && identity.UserID == userID {
			result = append(result, identity)
		}
	}

	return result, nil
}

// FindUserByProvider finds a user by provider identity (for login)
func (s *InMemoryUserIdentityStore) FindUserByProvider(ctx context.Context, tenantID string, provider domain.IdentityProvider, providerID string) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providerKey := s.makeProviderKey(tenantID, provider, providerID)
	identityID, exists := s.providerIndex[providerKey]
	if !exists {
		return nil, domain.ErrUserIdentityNotFound
	}

	// Find the identity to get user ID
	var userID string
	for _, identity := range s.identities {
		if identity.TenantID == tenantID && identity.ID == identityID {
			userID = identity.UserID
			break
		}
	}

	if userID == "" {
		return nil, domain.ErrUserIdentityNotFound
	}

	// Get user from user store
	user, err := s.userStore.Get(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Exists checks if an identity exists
func (s *InMemoryUserIdentityStore) Exists(ctx context.Context, tenantID, userID string, provider domain.IdentityProvider) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, identity := range s.identities {
		if identity.TenantID == tenantID &&
			identity.UserID == userID &&
			identity.Provider == provider {
			return true, nil
		}
	}

	return false, nil
}
