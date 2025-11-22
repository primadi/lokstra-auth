package memory

import (
	"context"
	"sync"
)

// InMemoryUserAppStore - User-App Access Relationships
type InMemoryUserAppStore struct {
	mu     sync.RWMutex
	access map[string]map[string]bool // key: tenantID:userID -> map[appID]bool
}

// NewUserAppStore creates a new in-memory user-app store
func NewUserAppStore() *InMemoryUserAppStore {
	return &InMemoryUserAppStore{
		access: make(map[string]map[string]bool),
	}
}

// makeKey creates a composite key for tenant+user
func (s *InMemoryUserAppStore) makeKey(tenantID, userID string) string {
	return tenantID + ":" + userID
}

// GrantAccess grants a user access to an app
func (s *InMemoryUserAppStore) GrantAccess(ctx context.Context, tenantID, appID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(tenantID, userID)
	if s.access[key] == nil {
		s.access[key] = make(map[string]bool)
	}
	s.access[key][appID] = true
	return nil
}

// RevokeAccess revokes a user's access to an app
func (s *InMemoryUserAppStore) RevokeAccess(ctx context.Context, tenantID, appID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(tenantID, userID)
	if s.access[key] != nil {
		delete(s.access[key], appID)
	}
	return nil
}

// HasAccess checks if user has access to an app
func (s *InMemoryUserAppStore) HasAccess(ctx context.Context, tenantID, appID, userID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(tenantID, userID)
	if s.access[key] == nil {
		return false, nil
	}
	return s.access[key][appID], nil
}

// ListUserApps lists all app IDs a user has access to
func (s *InMemoryUserAppStore) ListUserApps(ctx context.Context, tenantID, userID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(tenantID, userID)
	if s.access[key] == nil {
		return []string{}, nil
	}

	apps := make([]string, 0, len(s.access[key]))
	for appID := range s.access[key] {
		apps = append(apps, appID)
	}
	return apps, nil
}

// ListAppUsers lists all user IDs who have access to an app
func (s *InMemoryUserAppStore) ListAppUsers(ctx context.Context, tenantID, appID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]string, 0)
	for key, apps := range s.access {
		// Extract userID from key (format: tenantID:userID)
		// Simple string split by ':'
		colonIndex := -1
		for i, r := range key {
			if r == ':' {
				colonIndex = i
				break
			}
		}
		if colonIndex > 0 && key[:colonIndex] == tenantID {
			userID := key[colonIndex+1:]
			if apps[appID] {
				users = append(users, userID)
			}
		}
	}
	return users, nil
}
