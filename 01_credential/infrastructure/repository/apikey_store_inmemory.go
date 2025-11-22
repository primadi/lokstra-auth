package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/primadi/lokstra-auth/01_credential/domain/apikey"
)

var (
	ErrAPIKeyNotFound      = errors.New("api key not found")
	ErrAPIKeyAlreadyExists = errors.New("api key already exists")
)

// InMemoryAPIKeyStore provides in-memory API key storage
type InMemoryAPIKeyStore struct {
	mu   sync.RWMutex
	keys map[string]*apikey.APIKey // key: "tenantID:appID:keyID"
}

var _ APIKeyStore = (*InMemoryAPIKeyStore)(nil)

// NewInMemoryAPIKeyStore creates a new in-memory API key store
func NewInMemoryAPIKeyStore() *InMemoryAPIKeyStore {
	return &InMemoryAPIKeyStore{
		keys: make(map[string]*apikey.APIKey),
	}
}

// GetByKeyID retrieves API key by key ID
func (s *InMemoryAPIKeyStore) GetByKeyID(ctx context.Context, tenantID, appID, keyID string) (*apikey.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := tenantID + ":" + appID + ":" + keyID
	apiKey, exists := s.keys[key]
	if !exists {
		return nil, ErrAPIKeyNotFound
	}

	return apiKey, nil
}

// Store saves a new API key
func (s *InMemoryAPIKeyStore) Store(ctx context.Context, key *apikey.APIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storeKey := key.TenantID + ":" + key.AppID + ":" + key.KeyID
	if _, exists := s.keys[storeKey]; exists {
		return ErrAPIKeyAlreadyExists
	}

	s.keys[storeKey] = key
	return nil
}

// UpdateLastUsed updates the last used timestamp
func (s *InMemoryAPIKeyStore) UpdateLastUsed(ctx context.Context, keyID, timestamp string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find key by ID
	for _, key := range s.keys {
		if key.ID == keyID {
			key.LastUsedAt = &timestamp
			return nil
		}
	}

	return ErrAPIKeyNotFound
}

// Revoke revokes an API key
func (s *InMemoryAPIKeyStore) Revoke(ctx context.Context, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find key by ID
	for _, key := range s.keys {
		if key.ID == keyID {
			now := "now" // In real impl, use time.Now().Format(time.RFC3339)
			key.RevokedAt = &now
			return nil
		}
	}

	return ErrAPIKeyNotFound
}

// ListByApp lists all API keys for an app
func (s *InMemoryAPIKeyStore) ListByApp(ctx context.Context, tenantID, appID string) ([]*apikey.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*apikey.APIKey
	prefix := tenantID + ":" + appID + ":"

	for key, apiKey := range s.keys {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			result = append(result, apiKey)
		}
	}

	return result, nil
}

// Delete permanently deletes an API key
func (s *InMemoryAPIKeyStore) Delete(ctx context.Context, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and delete key by ID
	for key, apiKey := range s.keys {
		if apiKey.ID == keyID {
			delete(s.keys, key)
			return nil
		}
	}

	return ErrAPIKeyNotFound
}
