package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
)

// InMemoryAppKeyStore is an in-memory implementation of AppKeyStore
type InMemoryAppKeyStore struct {
	// keys: map[internal_id] -> AppKey
	keys map[string]*domain.AppKey
	// keysByKeyID: map[key_id] -> AppKey (for fast lookup during auth)
	keysByKeyID map[string]*domain.AppKey
	mu          sync.RWMutex
}

var _ repository.AppKeyStore = (*InMemoryAppKeyStore)(nil)

// NewAppKeyStore creates a new in-memory app key store
func NewAppKeyStore() *InMemoryAppKeyStore {
	return &InMemoryAppKeyStore{
		keys:        make(map[string]*domain.AppKey),
		keysByKeyID: make(map[string]*domain.AppKey),
	}
}

// Store stores a new app API key
func (s *InMemoryAppKeyStore) Store(ctx context.Context, key *domain.AppKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.keys[key.ID]; exists {
		return fmt.Errorf("key already exists: %s", key.ID)
	}

	s.keys[key.ID] = key
	s.keysByKeyID[key.KeyID] = key
	return nil
}

// GetByKeyID retrieves an API key by its key ID within tenant/app scope
func (s *InMemoryAppKeyStore) GetByKeyID(ctx context.Context, tenantID, appID, keyID string) (*domain.AppKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, exists := s.keysByKeyID[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	// Verify tenant and app match
	if key.TenantID != tenantID || key.AppID != appID {
		return nil, fmt.Errorf("key not found")
	}

	return key, nil
}

// GetByID retrieves an API key by its internal ID
func (s *InMemoryAppKeyStore) GetByID(ctx context.Context, id string) (*domain.AppKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, exists := s.keys[id]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", id)
	}

	return key, nil
}

// GetByPrefix retrieves all API keys with the given prefix
func (s *InMemoryAppKeyStore) GetByPrefix(ctx context.Context, prefix string) ([]*domain.AppKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]*domain.AppKey, 0)
	for _, key := range s.keys {
		if key.Prefix == prefix {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// ListByApp lists all API keys for an app
func (s *InMemoryAppKeyStore) ListByApp(ctx context.Context, tenantID, appID string) ([]*domain.AppKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]*domain.AppKey, 0)
	for _, key := range s.keys {
		if key.TenantID == tenantID && key.AppID == appID {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// ListByTenant lists all API keys for a tenant (across all apps)
func (s *InMemoryAppKeyStore) ListByTenant(ctx context.Context, tenantID string) ([]*domain.AppKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]*domain.AppKey, 0)
	for _, key := range s.keys {
		if key.TenantID == tenantID {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// Update updates an API key
func (s *InMemoryAppKeyStore) Update(ctx context.Context, key *domain.AppKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.keys[key.ID]; !exists {
		return fmt.Errorf("key not found: %s", key.ID)
	}

	s.keys[key.ID] = key
	s.keysByKeyID[key.KeyID] = key
	return nil
}

// Revoke revokes an API key
func (s *InMemoryAppKeyStore) Revoke(ctx context.Context, tenantID, appID, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	appKey, err := s.GetByKeyID(ctx, tenantID, appID, keyID)
	if err != nil {
		return fmt.Errorf("key not found: %s", appKey.KeyID)
	}

	now := time.Now()
	appKey.Revoked = true
	appKey.RevokedAt = &now
	return nil
}

// Delete permanently deletes an API key
func (s *InMemoryAppKeyStore) Delete(ctx context.Context, tenantID, appID, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	appKey, err := s.GetByKeyID(ctx, tenantID, appID, keyID)
	if err != nil {
		return fmt.Errorf("key not found: %s", appKey.KeyID)
	}

	delete(s.keys, appKey.ID)
	delete(s.keysByKeyID, appKey.KeyID)
	return nil
}

// UpdateLastUsed updates the last used timestamp
func (s *InMemoryAppKeyStore) UpdateLastUsed(ctx context.Context, id string, lastUsed time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exists := s.keys[id]
	if !exists {
		return fmt.Errorf("key not found: %s", id)
	}

	key.LastUsed = &lastUsed
	return nil
}
