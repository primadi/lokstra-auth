package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/primadi/lokstra-auth/00_core/domain"
)

// InMemoryAppStore is an in-memory implementation of AppStore
type InMemoryAppStore struct {
	// apps: map[tenantID:appID] -> App
	apps map[string]*domain.App
	mu   sync.RWMutex
}

// NewAppStore creates a new in-memory app store
func NewAppStore() *InMemoryAppStore {
	return &InMemoryAppStore{
		apps: make(map[string]*domain.App),
	}
}

// makeKey creates a composite key for tenant+app
func (s *InMemoryAppStore) makeKey(tenantID, appID string) string {
	return tenantID + ":" + appID
}

// Create creates a new app
func (s *InMemoryAppStore) Create(ctx context.Context, app *domain.App) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(app.TenantID, app.ID)
	if _, exists := s.apps[key]; exists {
		return fmt.Errorf("app already exists: %s in tenant %s", app.ID, app.TenantID)
	}

	s.apps[key] = app
	return nil
}

// Get retrieves an app by ID within a tenant
func (s *InMemoryAppStore) Get(ctx context.Context, tenantID, appID string) (*domain.App, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(tenantID, appID)
	app, exists := s.apps[key]
	if !exists {
		return nil, fmt.Errorf("app not found: %s in tenant %s", appID, tenantID)
	}

	return app, nil
}

// Update updates an existing app
func (s *InMemoryAppStore) Update(ctx context.Context, app *domain.App) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(app.TenantID, app.ID)
	if _, exists := s.apps[key]; !exists {
		return fmt.Errorf("app not found: %s in tenant %s", app.ID, app.TenantID)
	}

	s.apps[key] = app
	return nil
}

// Delete deletes an app
func (s *InMemoryAppStore) Delete(ctx context.Context, tenantID, appID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(tenantID, appID)
	delete(s.apps, key)
	return nil
}

// List lists all apps for a tenant
func (s *InMemoryAppStore) List(ctx context.Context, tenantID string) ([]*domain.App, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apps := make([]*domain.App, 0)
	for _, app := range s.apps {
		if app.TenantID == tenantID {
			apps = append(apps, app)
		}
	}

	return apps, nil
}

// GetByName retrieves an app by name within a tenant
func (s *InMemoryAppStore) GetByName(ctx context.Context, tenantID, name string) (*domain.App, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, app := range s.apps {
		if app.TenantID == tenantID && app.Name == name {
			return app, nil
		}
	}

	return nil, fmt.Errorf("app not found with name: %s in tenant %s", name, tenantID)
}

// Exists checks if an app exists within a tenant
func (s *InMemoryAppStore) Exists(ctx context.Context, tenantID, appID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(tenantID, appID)
	_, exists := s.apps[key]
	return exists, nil
}

// ListByType lists apps by type within a tenant
func (s *InMemoryAppStore) ListByType(ctx context.Context, tenantID string, appType domain.AppType) ([]*domain.App, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apps := make([]*domain.App, 0)
	for _, app := range s.apps {
		if app.TenantID == tenantID && app.Type == appType {
			apps = append(apps, app)
		}
	}

	return apps, nil
}
